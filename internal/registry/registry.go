// Package registry handles the physical storage of the Code Genome.
// Path: internal/registry/registry.go
// Deterministic Status: SEQUENCED (CERTIFIED-STABLE)
package registry

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"genesis/internal/identity"

	// SQLite driver is required for L2 physical authority
	_ "github.com/mattn/go-sqlite3"
)

const bootstrapSchema = `
-- ... (existing nodes, inference_profiles, semantic_records)

-- CHAPTER 2.2: THE SEMANTIC INDEX
-- This virtual table handles high-dimensional vector similarity.
-- Dimensions: 3072 (Standard for EMBED-tier models)
CREATE VIRTUAL TABLE IF NOT EXISTS semantic_index USING vss0(
  vector(3072)
);
`


var ErrNodeNotFound = errors.New("registry: node not found")

type Registry struct {
	db *sql.DB
}

// Open initializes the connection with mandatory WAL and Foreign Key enforcement.
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

	// CHAPTER 2.1: Enforce single-writer serialization
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Reliability: Ensure physical accessibility
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("registry: database unreachable: %w", err)
	}

	// Relational Enforcement
	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		return nil, fmt.Errorf("registry: failed to enable foreign keys: %w", err)
	}

	r := &Registry{db: db}
	if err := r.Migrate(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Registry) Migrate() error {
	if _, err := r.db.Exec(SchemaSQL); err != nil {
		return fmt.Errorf("registry: schema migration failed: %w", err)
	}
	return nil
}

// PersistSemanticRecord validates and stores cognitive outputs ATOMICALLY.
func (r *Registry) PersistSemanticRecord(nodeID, profileID, hash string) error {
	if nodeID == "" || profileID == "" || hash == "" {
		return fmt.Errorf("registry: semantic record requires nodeID, profileID, and hash")
	}

	// CHAPTER 2.3: Transactional Enforcement for SCC and Semantic safety
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("registry: failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Mark existing records for this node/profile as stale
	staleQuery := `UPDATE semantic_records SET is_stale = 1 WHERE node_id = ? AND profile_id = ?`
	if _, err := tx.Exec(staleQuery, nodeID, profileID); err != nil {
		return fmt.Errorf("registry: failed to cycle staleness: %w", err)
	}

	// 2. Insert the new active record
	insertQuery := `INSERT INTO semantic_records (node_id, profile_id, summary_hash, is_stale)
	                VALUES (?, ?, ?, 0)`
	if _, err := tx.Exec(insertQuery, nodeID, profileID, hash); err != nil {
		return fmt.Errorf("registry: failed to persist semantic record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("registry: failed to commit semantic update: %w", err)
	}
	return nil
}

// PersistNode handles the Identity Quad storage with ON CONFLICT resolution.
func (r *Registry) PersistNode(quad identity.IdentityQuad, maturity string, class int, auditUnix int64) error {
	query := `
	INSERT INTO nodes (
		node_id, kind, visibility, module_path, package_path,
		receiver, symbol_name, arity, contract_id, logic_hash,
		dependency_hash, maturity, authority_class, last_audit_timestamp
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(node_id) DO UPDATE SET
		contract_id=excluded.contract_id,
		logic_hash=excluded.logic_hash,
		dependency_hash=excluded.dependency_hash,
		maturity=excluded.maturity,
		authority_class=excluded.authority_class,
		last_audit_timestamp=excluded.last_audit_timestamp;`

	id := quad.NodeID
	_, err := r.db.Exec(
		query,
		id.String(), string(id.Kind), string(id.Visibility), id.Module, id.Package,
		string(id.Receiver), id.Symbol, id.Arity, quad.ContractID, quad.LogicHash,
		quad.DependencyHash, maturity, class, auditUnix,
	)
	if err != nil {
		return fmt.Errorf("registry: failed to persist node %s: %w", id.String(), err)
	}
	return nil
}

// GetNode retrieves a specific node.
func (r *Registry) GetNode(nodeID string) (identity.IdentityQuad, string, error) {
	q, mat, _, err := r.GetNodeWithAuthority(nodeID)
	return q, mat, err
}

// GetNodeWithAuthority is the required interface for L10 and L12 logic.
func (r *Registry) GetNodeWithAuthority(nodeID string) (identity.IdentityQuad, string, int, error) {
	var q identity.IdentityQuad
	var maturity string
	var authClass int
	var k, v, rec, mod, pkg, sym string
	var arity int

	query := `SELECT kind, visibility, module_path, package_path, receiver,
	                 symbol_name, arity, contract_id, logic_hash,
	                 dependency_hash, maturity, authority_class
	          FROM nodes WHERE node_id = ?`

	err := r.db.QueryRow(query, nodeID).Scan(
		&k, &v, &mod, &pkg, &rec, &sym, &arity,
		&q.ContractID, &q.LogicHash, &q.DependencyHash, &maturity, &authClass,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return q, "", 0, fmt.Errorf("%w: %s", ErrNodeNotFound, nodeID)
		}
		return q, "", 0, fmt.Errorf("registry: retrieval failure: %w", err)
	}

	// Grammar rehydration (Chapter 12.3)
	rawID := fmt.Sprintf("%s.%s.%s.%s.%s.%s.%d", k, v, mod, pkg, rec, sym, arity)
	parsedID, err := identity.ParseNodeID(rawID)
	if err != nil {
		return q, "", 0, fmt.Errorf("registry: stored identity corruption for %s: %w", nodeID, err)
	}

	q.NodeID = parsedID
	return q, maturity, authClass, nil
}

// ListAllNodeIDs returns a deterministic, sorted slice of identities.
func (r *Registry) ListAllNodeIDs() ([]string, error) {
	rows, err := r.db.Query("SELECT node_id FROM nodes ORDER BY node_id ASC")
	if err != nil {
		return nil, fmt.Errorf("registry: query failure: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("registry: scan failure: %w", err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("registry: row iteration failure: %w", err)
	}

	return ids, nil
}

// Close gracefully shuts down the database connection.
func (r *Registry) Close() error {
	if err := r.db.Close(); err != nil {
		return fmt.Errorf("registry: close failed: %w", err)
	}
	return nil
}

type VectorMatch struct {
	NodeID   string
	Distance float32
}

// QueryNearestNeighbors executes a cosine similarity search in the VSS virtual table.
func (r *Registry) QueryNearestNeighbors(vector []float32, limit int) ([]VectorMatch, error) {
	// CHAPTER 2: The Physical Semantic Index lookup
	query := `
		SELECT s.node_id, v.distance
		FROM semantic_index v
		JOIN semantic_records s ON v.record_id = s.record_id
		WHERE v.vector MATCH ? 
		AND v.k = ?
		AND s.is_stale = 0`
	
	// Implementation note: MATCH requires a serialized blob of the float32 array.
	rows, err := r.db.Query(query, vector, limit)
	if err != nil {
		return nil, fmt.Errorf("registry: vss lookup failure: %w", err)
	}
	defer rows.Close()

	var matches []VectorMatch
	for rows.Next() {
		var m VectorMatch
		if err := rows.Scan(&m.NodeID, &m.Distance); err != nil {
			return nil, err
		}
		matches = append(matches, m)
	}
	return matches, nil
}
