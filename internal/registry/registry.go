// Package registry handles the Physical Genome Authority.
// Path: internal/registry/registry.go
// Deterministic Status: SEQUENCED (CERTIFIED-STABLE)
package registry

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net/url"

	"genesis/internal/identity"

	_ "github.com/mattn/go-sqlite3"
)

const (
	// EmbedDimensions is the required vector size for the semantic index.
	EmbedDimensions = 3072
)

// ErrNodeNotFound is returned when a requested node does not exist.
var ErrNodeNotFound = errors.New("registry: node not found")

// Registry is the authoritative SQLite-backed storage layer.
type Registry struct {
	db *sql.DB
}

// VectorMatch represents a nearest-neighbor search result.
type VectorMatch struct {
	NodeID   string
	Distance float32
}

// Open initializes a hardened SQLite connection.
//
// Expected DSN shape from the shell is typically:
//   file:/path/to/genome.db
func Open(dsn string) (*Registry, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("registry: invalid dsn format: %w", err)
	}

	q := u.Query()
	q.Set("_journal", "WAL")
	q.Set("_busy_timeout", "5000")
	u.RawQuery = q.Encode()

	db, err := sql.Open("sqlite3", u.String())
	if err != nil {
		return nil, fmt.Errorf("registry: connection failed: %w", err)
	}

	// Chapter 2.1: Single-writer serialization for SQLite.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("registry: database unreachable: %w", err)
	}

	// Enforce relational integrity explicitly.
	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("registry: failed to enable foreign keys: %w", err)
	}

	r := &Registry{db: db}
	if err := r.Migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return r, nil
}

// Migrate materializes the relational and semantic index schema.
//
// Requires SchemaSQL and VSSSchemaSQL to be defined in schema.go.
func (r *Registry) Migrate() error {
	if _, err := r.db.Exec(SchemaSQL); err != nil {
		return fmt.Errorf("registry: relational migration failed: %w", err)
	}
	if _, err := r.db.Exec(VSSSchemaSQL); err != nil {
		return fmt.Errorf("registry: semantic index migration failed: %w", err)
	}
	return nil
}

// Close gracefully shuts down the database connection.
func (r *Registry) Close() error {
	if err := r.db.Close(); err != nil {
		return fmt.Errorf("registry: close failed: %w", err)
	}
	return nil
}

// PersistNode stores or updates a node in the physical registry.
//
// Identity-shaping fields remain immutable after first insert.
// Mutable fields are updated via ON CONFLICT.
func (r *Registry) PersistNode(q identity.IdentityQuad, maturity string, class int, auditUnix int64) error {
	query := `
	INSERT INTO nodes (
		node_id, module_path, package_path, symbol_name, kind, visibility, receiver_shape, arity,
		logic_hash, contract_id, dependency_hash, maturity, authority_class, last_audit_timestamp
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(node_id) DO UPDATE SET
		logic_hash = excluded.logic_hash,
		contract_id = excluded.contract_id,
		dependency_hash = excluded.dependency_hash,
		maturity = excluded.maturity,
		authority_class = excluded.authority_class,
		last_audit_timestamp = excluded.last_audit_timestamp`

	id := q.NodeID
	_, err := r.db.Exec(
		query,
		id.String(),
		id.Module,
		id.Package,
		id.Symbol,
		string(id.Kind),
		string(id.Visibility),
		string(id.Receiver),
		id.Arity,
		q.LogicHash,
		q.ContractID,
		q.DependencyHash,
		maturity,
		class,
		auditUnix,
	)
	if err != nil {
		return fmt.Errorf("registry: failed to persist node %s: %w", id.String(), err)
	}
	return nil
}

// GetNode retrieves a node without authority metadata.
func (r *Registry) GetNode(nodeID string) (identity.IdentityQuad, string, error) {
	q, maturity, _, err := r.GetNodeWithAuthority(nodeID)
	return q, maturity, err
}

// GetNodeWithAuthority retrieves a node and its authority metadata.
//
// Chapter 12.3: identity is reconstructed and revalidated through ParseNodeID.
func (r *Registry) GetNodeWithAuthority(nodeID string) (identity.IdentityQuad, string, int, error) {
	query := `
	SELECT
		module_path, package_path, symbol_name, kind, visibility, receiver_shape, arity,
		logic_hash, contract_id, dependency_hash, maturity, authority_class
	FROM nodes
	WHERE node_id = ?`

	var q identity.IdentityQuad
	var id identity.NodeID
	var kind, visibility, receiverShape, maturity string
	var class int

	err := r.db.QueryRow(query, nodeID).Scan(
		&id.Module,
		&id.Package,
		&id.Symbol,
		&kind,
		&visibility,
		&receiverShape,
		&id.Arity,
		&q.LogicHash,
		&q.ContractID,
		&q.DependencyHash,
		&maturity,
		&class,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return q, "", 0, fmt.Errorf("%w: %s", ErrNodeNotFound, nodeID)
		}
		return q, "", 0, fmt.Errorf("registry: read failure for %s: %w", nodeID, err)
	}

	id.Kind = identity.Kind(kind)
	id.Visibility = identity.Visibility(visibility)
	id.Receiver = identity.ReceiverShape(receiverShape)

	if _, err := identity.ParseNodeID(id.String()); err != nil {
		return q, "", 0, fmt.Errorf("registry: stored identity corruption for %s: %w", nodeID, err)
	}

	q.NodeID = id
	return q, maturity, class, nil
}

// ListAllNodeIDs returns a deterministic, sorted slice of all node identities.
func (r *Registry) ListAllNodeIDs() ([]string, error) {
	rows, err := r.db.Query(`SELECT node_id FROM nodes ORDER BY node_id ASC`)
	if err != nil {
		return nil, fmt.Errorf("registry: list failure: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("registry: list scan failure: %w", err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("registry: list iteration failure: %w", err)
	}

	return ids, nil
}

// PersistSemanticRecord atomically updates the semantic record for a node/profile
// and keeps the VSS rowid aligned with semantic_records.record_id.
//
// Behavior:
//   - validates inputs
//   - marks prior records for (nodeID, profileID) stale
//   - inserts a new semantic_records row
//   - inserts the corresponding vector into semantic_index using the same rowid
func (r *Registry) PersistSemanticRecord(
	ctx context.Context,
	nodeID string,
	profileID string,
	summary string,
	vector []float32,
) error {
	if nodeID == "" {
		return fmt.Errorf("registry: semantic record requires nodeID")
	}
	if profileID == "" {
		return fmt.Errorf("registry: semantic record requires profileID")
	}
	if summary == "" {
		return fmt.Errorf("registry: semantic record requires summary")
	}
	if len(vector) == 0 {
		return fmt.Errorf("registry: semantic record requires non-empty vector")
	}
	if len(vector) != EmbedDimensions {
		return fmt.Errorf(
			"registry: semantic record vector dimension mismatch: got %d want %d",
			len(vector), EmbedDimensions,
		)
	}

	serialized, err := serializeVector32(vector)
	if err != nil {
		return fmt.Errorf("registry: vector serialization failure: %w", err)
	}

	summaryHash := hashSummary(summary)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("registry: failed to begin semantic transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// 1. Mark older records stale for this node/profile.
	staleQuery := `
		UPDATE semantic_records
		SET is_stale = 1
		WHERE node_id = ? AND profile_id = ?`
	if _, err := tx.ExecContext(ctx, staleQuery, nodeID, profileID); err != nil {
		return fmt.Errorf("registry: failed to cycle semantic staleness: %w", err)
	}

	// 2. Insert the new relational semantic record.
	insertRecord := `
		INSERT INTO semantic_records (node_id, profile_id, summary_hash, summary_text, is_stale)
		VALUES (?, ?, ?, ?, 0)`
	res, err := tx.ExecContext(ctx, insertRecord, nodeID, profileID, summaryHash, summary)
	if err != nil {
		return fmt.Errorf("registry: failed to persist semantic record: %w", err)
	}

	recordID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("registry: failed to read semantic record id: %w", err)
	}

	// 3. Insert aligned vector row into the semantic index.
	insertVector := `INSERT INTO semantic_index (rowid, vector) VALUES (?, ?)`
	if _, err := tx.ExecContext(ctx, insertVector, recordID, serialized); err != nil {
		return fmt.Errorf("registry: failed to persist semantic vector: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("registry: failed to commit semantic record: %w", err)
	}

	return nil
}

// QueryNearestNeighbors executes a nearest-neighbor lookup against the VSS index.
func (r *Registry) QueryNearestNeighbors(vector []float32, limit int) ([]VectorMatch, error) {
	if len(vector) == 0 {
		return nil, fmt.Errorf("registry: nearest-neighbor query requires non-empty vector")
	}
	if len(vector) != EmbedDimensions {
		return nil, fmt.Errorf(
			"registry: nearest-neighbor vector dimension mismatch: got %d want %d",
			len(vector), EmbedDimensions,
		)
	}
	if limit <= 0 {
		return nil, fmt.Errorf("registry: nearest-neighbor query requires positive limit")
	}

	serialized, err := serializeVector32(vector)
	if err != nil {
		return nil, fmt.Errorf("registry: vector serialization failure: %w", err)
	}

	query := `
		SELECT s.node_id, v.distance
		FROM semantic_index v
		JOIN semantic_records s ON v.rowid = s.record_id
		WHERE vss_search(v.vector, vss_search_params(?, ?))
		  AND s.is_stale = 0`

	rows, err := r.db.Query(query, serialized, limit)
	if err != nil {
		return nil, fmt.Errorf("registry: nearest-neighbor search failure: %w", err)
	}
	defer rows.Close()

	var matches []VectorMatch
	for rows.Next() {
		var m VectorMatch
		if err := rows.Scan(&m.NodeID, &m.Distance); err != nil {
			return nil, fmt.Errorf("registry: nearest-neighbor scan failure: %w", err)
		}
		matches = append(matches, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("registry: nearest-neighbor iteration failure: %w", err)
	}

	return matches, nil
}

// hashSummary returns a deterministic SHA-256 hex digest of the semantic summary.
func hashSummary(summary string) string {
	sum := sha256.Sum256([]byte(summary))
	return fmt.Sprintf("%x", sum[:])
}

// serializeVector32 converts a float32 slice to a little-endian byte blob.
func serializeVector32(vector []float32) ([]byte, error) {
	if len(vector) == 0 {
		return nil, fmt.Errorf("empty vector")
	}

	out := make([]byte, len(vector)*4)
	for i, f := range vector {
		binary.LittleEndian.PutUint32(out[i*4:], math.Float32bits(f))
	}
	return out, nil
}
