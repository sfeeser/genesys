// Path: internal/registry/schema.go
package registry

const SchemaSQL = `
CREATE TABLE IF NOT EXISTS nodes (
    node_id TEXT PRIMARY KEY,
    kind TEXT NOT NULL,
    visibility TEXT NOT NULL,
    module_path TEXT NOT NULL,
    package_path TEXT NOT NULL,
    receiver TEXT NOT NULL CHECK (receiver IN ('none', 'pointer', 'value')),
    symbol_name TEXT NOT NULL,
    arity INTEGER NOT NULL,
    contract_id TEXT,
    logic_hash TEXT,
    dependency_hash TEXT,
    maturity TEXT NOT NULL CHECK (maturity IN ('draft', 'hollow', 'anchored', 'hydrated', 'sequenced')),
    authority_class INTEGER NOT NULL CHECK (authority_class IN (0,1,2)),
    last_audit_timestamp INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS inference_profiles (
    profile_id TEXT PRIMARY KEY,
    tier TEXT NOT NULL CHECK (tier IN ('FAST', 'DEEP', 'EMBED')),
    model_name TEXT NOT NULL,
    model_revision TEXT NOT NULL,
    dimensions INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS semantic_records (
    record_id INTEGER PRIMARY KEY,
    node_id TEXT NOT NULL,
    profile_id TEXT NOT NULL,
    summary_hash TEXT NOT NULL,
    is_stale INTEGER NOT NULL CHECK (is_stale IN (0,1)),
    FOREIGN KEY (node_id) REFERENCES nodes(node_id),
    FOREIGN KEY (profile_id) REFERENCES inference_profiles(profile_id)
);

CREATE INDEX IF NOT EXISTS idx_semantic_node ON semantic_records(node_id);
CREATE INDEX IF NOT EXISTS idx_semantic_profile ON semantic_records(profile_id);
`
