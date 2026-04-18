// Package spec handles the ingestion and normalization of the Specbook.
// It acts as the gatekeeper for Architectural Intent.
// Deterministic Status: SEQUENCED (HARDENED)
package spec

import (
	"fmt"
	"io"

	"genesis/internal/identity" // Allowed L1 Import

	"gopkg.in/yaml.v3"
)

// Specbook represents the raw architectural intent.
type Specbook struct {
	Version string     `yaml:"version"`
	Nodes   []NodeSpec `yaml:"nodes"`
}

// NodeSpec defines the desired state of a single node in the genome.
type NodeSpec struct {
	NodeID      string   `yaml:"node_id"`
	Gene        string   `yaml:"gene"`
	Purpose     string   `yaml:"purpose"`
	Authority   int      `yaml:"authority_class"`
	AllowedDeps []string `yaml:"allowed_imports"`
}

// Ingest consumes raw bytes to ensure environment-stable ingestion.
// Enforces Chapter 5: Decouples logic from local filesystem paths.
func Ingest(r io.Reader) (*Specbook, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("spec: read failure: %w", err)
	}

	var book Specbook
	if err := yaml.Unmarshal(data, &book); err != nil {
		return nil, fmt.Errorf("spec: malformed yaml: %w", err)
	}

	// Normalization & Validation Gate
	seenNodes := make(map[string]struct{})
	for i, n := range book.Nodes {
		// 1. Uniqueness Guard
		if _, exists := seenNodes[n.NodeID]; exists {
			return nil, fmt.Errorf("spec: duplicate node_id detected: %s", n.NodeID)
		}

		// 2. Identity Grammar Enforcement (L1 Delegate)
		if _, err := identity.ParseNodeID(n.NodeID); err != nil {
			return nil, fmt.Errorf("spec: node %d fails identity physics: %w", i, err)
		}

		// 3. Authority Bound Check
		if n.Authority < 0 || n.Authority > 2 {
			return nil, fmt.Errorf("spec: invalid authority class for %s: %d", n.NodeID, n.Authority)
		}

		// 4. Required Field Validation (The Soul)
		if n.Gene == "" || n.Purpose == "" {
			return nil, fmt.Errorf("spec: node %s is missing required Gene or Purpose", n.NodeID)
		}

		seenNodes[n.NodeID] = struct{}{}
	}

	return &book, nil
}

// MapToConceptualQuads converts Specs into State 1 Quads.
// Enforces Chapter 1.4: Explicitly defines empty anchors for early-stage nodes.
func (s *Specbook) MapToConceptualQuads() ([]identity.IdentityQuad, error) {
	quads := make([]identity.IdentityQuad, 0, len(s.Nodes))
	
	for _, n := range s.Nodes {
		id, err := identity.ParseNodeID(n.NodeID)
		if err != nil {
			return nil, err
		}

		// Materializing with explicit "Empty Anchor" markers.
		// These signify State 1 (Conceptual) maturity.
		quads = append(quads, identity.IdentityQuad{
			NodeID:         id,
			ContractID:     "pending:hollow",
			LogicHash:      "pending:hydrating",
			DependencyHash: "pending:sequencing",
		})
	}
	
	return quads, nil
}
