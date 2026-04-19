// Package registry handles the Physical Genome Authority.
// Path: internal/registry/search.go
// Deterministic Status: SEQUENCED (HARDENED-STABLE)
package registry

import (
	"encoding/binary"
	"fmt"
	"math"
)

type VectorMatch struct {
	NodeID   string
	Distance float32
}

// serializeVector32 converts a float32 slice into the Little-Endian BLOB 
// required by the VSS extension.
func serializeVector32(v []float32) ([]byte, error) {
	buf := make([]byte, len(v)*4)
	for i, f := range v {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(f))
	}
	return buf, nil
}

// QueryNearestNeighbors executes a truthful high-dimensional similarity lookup.
func (r *Registry) QueryNearestNeighbors(vector []float32, limit int) ([]VectorMatch, error) {
	// 1. Input Validation
	if len(vector) == 0 {
		return nil, fmt.Errorf("registry: nearest-neighbor query requires non-empty vector")
	}
	if limit <= 0 {
		return nil, fmt.Errorf("registry: nearest-neighbor query requires positive limit")
	}

	// 2. Truthful Physical Encoding
	serialized, err := serializeVector32(vector)
	if err != nil {
		return nil, fmt.Errorf("registry: vector serialization failure: %w", err)
	}

	// 3. The Physical VSS lookup
	// Note: We join semantic_index.rowid with semantic_records.record_id.
	query := `
		SELECT s.node_id, v.distance
		FROM semantic_index v
		JOIN semantic_records s ON v.rowid = s.record_id
		WHERE vss_search(v.vector, vss_search_params(?, ?))
		  AND s.is_stale = 0`

	rows, err := r.db.Query(query, serialized, limit)
	if err != nil {
		return nil, fmt.Errorf("registry: vss lookup failure: %w", err)
	}
	defer rows.Close()

	var matches []VectorMatch
	for rows.Next() {
		var m VectorMatch
		if err := rows.Scan(&m.NodeID, &m.Distance); err != nil {
			return nil, fmt.Errorf("registry: vss scan failure: %w", err)
		}
		matches = append(matches, m)
	}

	// 4. Reliability Check
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("registry: vss row iteration failure: %w", err)
	}

	return matches, nil
}
