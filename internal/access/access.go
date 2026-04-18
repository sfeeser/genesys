// Package access defines the security and authority policies for the engine.
// It serves as the Final Guardrail for mutation authorization.
// Deterministic Status: SEQUENCED (GRAMMAR-AWARE)
package access

import (
	"errors"
	"fmt"
	"strings"

	"genesis/internal/identity" // Now consumed for field-aware scope
	"genesis/internal/registry" // Allowed L2 Import
)

// Permission defines the level of mutation allowed on a node.
type Permission int

const (
	Deny Permission = iota
	Read
	Mutate
	Admin
)

// Guard evaluates mutation requests against the Structural Genome's authority.
type Guard struct {
	reg *registry.Registry
}

// New initializes the policy evaluator.
func New(r *registry.Registry) *Guard {
	return &Guard{reg: r}
}

// Authorize checks if a mutation on the target NodeID is permitted.
// Enforces Chapter 1.2: Authority Partitioning and Field-Aware Scoping.
func (g *Guard) Authorize(nodeID string, agentScope string, requested Permission) (bool, error) {
	// 1. Structural Identity Validation
	// Enforces Chapter 12.3: We parse first to prevent "Identity Squatting" or prefix attacks.
	id, err := identity.ParseNodeID(nodeID)
	if err != nil {
		return false, fmt.Errorf("access: invalid target identity: %w", err)
	}

	// 2. Field-Aware Scope Check
	// Policy: Agent scope must strictly cover the Module and Package of the target.
	if !strings.HasPrefix(id.Module, agentScope) && !strings.HasPrefix(id.Package, agentScope) {
		return false, fmt.Errorf("access: agent scope %s does not cover namespace %s/%s", agentScope, id.Module, id.Package)
	}

	// 3. Authority Retrieval with Hardened Error Branching
	_, _, class, err := g.reg.GetNodeWithAuthority(nodeID)
	if err != nil {
		if errors.Is(err, registry.ErrNodeNotFound) {
			// New nodes enter Class 0 (Conceptual). 
			// Policy: New nodes require at least Mutate permission to bootstrap.
			return requested >= Mutate, nil
		}
		// Failure of the Physical Authority (L2) results in immediate Denial.
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
	// Truthful Genesis Protected Paths
	forbidden := []string{".git", "vendor", "internal/identity", "internal/registry"}
	for _, f := range forbidden {
		if strings.HasPrefix(path, f) || strings.Contains(path, "/"+f) {
			return true
		}
	}
	return false
}
