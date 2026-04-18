package registry

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"genesis/internal/identity"

	_ "github.com/mattn/go-sqlite3"
)

var ErrNodeNotFound = errors.New("registry: node not found")

type Registry struct {
	db *sql.DB
}

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
		return nil, fmt.Errorf("registry: failed to connect to genome.db: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	return &Registry{db: db}, nil
}

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
		last_audit_timestamp=excluded.last_audit_timestamp;`

	id := quad.NodeID
	_, err := r.db.Exec(
		query,
		id.String(),
		string(id.Kind),
		string(id.Visibility),
		id.Module,
		id.Package,
		string(id.Receiver),
		id.Symbol,
		id.Arity,
		quad.ContractID,
		quad.LogicHash,
		quad.DependencyHash,
		maturity,
		class,
		auditUnix,
	)
	if err != nil {
		return fmt.Errorf("registry: failed to persist node %s: %w", id.String(), err)
	}
	return nil
}

func (r *Registry) GetNode(nodeID string) (identity.IdentityQuad, string, error) {
	var q identity.IdentityQuad
	var maturity string
	var k, v, rec string
	var mod, pkg, sym string
	var arity int

	query := `SELECT kind, visibility, module_path, package_path, receiver,
	                 symbol_name, arity, contract_id, logic_hash,
	                 dependency_hash, maturity FROM nodes WHERE node_id = ?`

	err := r.db.QueryRow(query, nodeID).Scan(
		&k, &v, &mod, &pkg, &rec, &sym, &arity,
		&q.ContractID, &q.LogicHash, &q.DependencyHash, &maturity,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return q, "", fmt.Errorf("%w: %s", ErrNodeNotFound, nodeID)
		}
		return q, "", fmt.Errorf("registry: retrieval failure: %w", err)
	}

	rawID := fmt.Sprintf("%s.%s.%s.%s.%s.%s.%d", k, v, mod, pkg, rec, sym, arity)
	parsedID, err := identity.ParseNodeID(rawID)
	if err != nil {
		return q, "", fmt.Errorf("registry: stored identity corruption for %s: %w", nodeID, err)
	}

	q.NodeID = parsedID
	return q, maturity, nil
}

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

	rawID := fmt.Sprintf("%s.%s.%s.%s.%s.%s.%d", k, v, mod, pkg, rec, sym, arity)
	parsedID, err := identity.ParseNodeID(rawID)
	if err != nil {
		return q, "", 0, fmt.Errorf("registry: stored identity corruption for %s: %w", nodeID, err)
	}

	q.NodeID = parsedID
	return q, maturity, authClass, nil
}

func (r *Registry) ListAllNodeIDs() ([]string, error) {
	rows, err := r.db.Query("SELECT node_id FROM nodes")
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
