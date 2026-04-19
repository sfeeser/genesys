// Package access defines the security policies and semantic retrieval for the engine.
// It serves as the Final Guardrail for mutation and the Librarian for intent.
// Deterministic Status: SEQUENCED (HARDENED-INTEGRATED)
package access

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"genesis/internal/cognition"
	"genesis/internal/identity"
	"genesis/internal/registry"
)

// Permission defines the level of mutation allowed on a node.
type Permission int

const (
	Deny Permission = iota
	Read
	Mutate
	Admin
)

// SearchResult pairs a NodeID with its semantic relevance score.
type SearchResult struct {
	NodeID string
	Score  float32
}

// Guard coordinates both security evaluation and semantic memory lookups.
type Guard struct {
	reg   *registry.Registry
	embed *cognition.Client
}

// New initializes the policy and semantic gateway.
func New(r *registry.Registry, e *cognition.Client) *Guard {
	return &Guard{
		reg:   r,
		embed: e,
	}
}

// =============================================================================
// POLICY HEMISPHERE (Security & Authority)
// =============================================================================

// Authorize checks if a mutation on the target NodeID is permitted.
// Enforces Chapter 1.2: Authority Partitioning and Field-Aware Scoping.
func (g *Guard) Authorize(nodeID string, agentScope string, requested Permission) (bool, error) {
	// 1. Structural Identity Validation
	id, err := identity.ParseNodeID(nodeID)
	if err != nil {
		return false, fmt.Errorf("access: invalid target identity: %w", err)
	}

	// 2. Field-Aware Scope Check
	// Policy: Agent scope must strictly cover the Module and Package of the target.
	if !strings.HasPrefix(id.Module, agentScope) && !strings.HasPrefix(id.Package, agentScope) {
		return false, fmt.Errorf("access: agent scope %s does not cover namespace %s/%s", agentScope, id.Module, id.Package)
	}

	// 3. Authority Retrieval
	_, _, class, err := g.reg.GetNodeWithAuthority(nodeID)
	if err != nil {
		if errors.Is(err, registry.ErrNodeNotFound) {
			// New nodes enter Class 0. Requires Mutate permission to bootstrap.
			return requested >= Mutate, nil
		}
		return false, fmt.Errorf("access: registry failure; authorization aborted: %w", err)
	}

	// 4. Authority Hierarchy Check (Chapter 1.2)
	switch class {
	case 0: // CONCEPTUAL: Unprotected nodes
		return requested >= Mutate, nil
	case 1: // PROTECTED: Standard genome nodes
		return requested >= Mutate, nil
	case 2: // CORE: L1-L4 Internal nodes
		return requested >= Admin, nil
	default:
		return false, fmt.Errorf("access: undefined authority class %d", class)
	}
}

// IsNamespaceLocked prevents physical surgery in forbidden genesis paths.
func (g *Guard) IsNamespaceLocked(path string) bool {
	forbidden := []string{".git", "vendor", "internal/identity", "internal/registry"}
	for _, f := range forbidden {
		if strings.HasPrefix(path, f) || strings.Contains(path, "/"+f) {
			return true
		}
	}
	return false
}

// =============================================================================
// SEMANTIC HEMISPHERE (Retrieval & Discovery)
// =============================================================================

// SearchIntent finds the most relevant nodes for a given natural language query.
// It leverages the EMBED tier for vectorization and the Registry for VSS lookup.
func (g *Guard) SearchIntent(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if g.embed == nil {
		return nil, fmt.Errorf("access: semantic tier (EMBED) not initialized")
	}

	// 1. CHAPTER 13: Vectorize the query intent
	vector, err := g.embed.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("access: embedding failure: %w", err)
	}

	// 2. CHAPTER 2.2: Physical VSS lookup via Registry
	// We assume registry has implemented QueryNearestNeighbors as defined in L2 extension.
	matches, err := g.reg.QueryNearestNeighbors(vector, limit)
	if err != nil {
		return nil, fmt.Errorf("access: semantic lookup failed: %w", err)
	}

	// 3. Mapping to Semantic Results
	var results []SearchResult
	for _, m := range matches {
		results = append(results, SearchResult{
			NodeID: m.NodeID,
			Score:  m.Distance,
		})
	}

	return results, nil
}
