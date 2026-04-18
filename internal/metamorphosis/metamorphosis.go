// Package metamorphosis implements the 5-State Pipeline for node evolution.
// It serves as the governing state machine for the Genesis Engine.
// Deterministic Status: SEQUENCED (HARDENED-ADJACENCY)
package metamorphosis

import (
	"fmt"

	"genesis/internal/audit"    // Allowed L7 Import
	"genesis/internal/identity" // Allowed L1 Import
	"genesis/internal/registry" // Allowed L2 Import
)

// State represents the maturity of a node in the Metamorphosis Pipeline.
type State string

const (
	StateConceptual State = "draft"     // State 1
	StateHollow     State = "hollow"    // State 2
	StateAnchored   State = "anchored"  // State 3
	StateHydrating  State = "hydrated"  // State 4
	StateSequenced  State = "sequenced" // State 5
)

// Pipeline governs the transition logic for the genome.
type Pipeline struct {
	reg     *registry.Registry
	auditor *audit.Auditor
}

// New initializes the governing state machine.
func New(r *registry.Registry, a *audit.Auditor) *Pipeline {
	return &Pipeline{
		reg:     r,
		auditor: a,
	}
}

// Transition attempts to move a node to its next maturity level.
// Enforces Chapter 3: Unidirectional and Adjacency-Governed evolution.
func (p *Pipeline) Transition(id string, target State, quad identity.IdentityQuad, content []byte, auditUnix int64) error {
	// 1. Identity Binding & Authority Check (Chapter 1.4)
	if id != quad.NodeID.String() {
		return fmt.Errorf("metamorphosis: node_id mismatch between request (%s) and quad (%s)", id, quad.NodeID.String())
	}

	currentQuad, currentMaturity, authClass, err := p.reg.GetNodeWithAuthority(id)
	if err != nil {
		return fmt.Errorf("metamorphosis: node not found in registry: %w", err)
	}

	// 2. Strict Adjacency Guardrail (Chapter 3)
	// Enforces that nodes do not "skip" states like the Virtual Loom or the Conscience.
	if !p.isLegalNextState(State(currentMaturity), target) {
		return fmt.Errorf("metamorphosis: illegal jump from %s to %s; transitions must be adjacent", currentMaturity, target)
	}

	// 3. Authority Consistency (Physics Check)
	// Ensure we are not accidentally shifting the module/package context during state change.
	if currentQuad.NodeID.Module != quad.NodeID.Module || currentQuad.NodeID.Package != quad.NodeID.Package {
		return fmt.Errorf("metamorphosis: forbidden topological drift for %s", id)
	}

	// 4. Final Gate Verification (State 5)
	if target == StateSequenced {
		verdicts := p.auditor.AuditNode(quad, content)
		for _, v := range verdicts {
			if !v.Passed {
				return fmt.Errorf("metamorphosis: gate failure [%s]: %s", v.Gate, v.Message)
			}
		}
	}

	// 5. Persistence (Truthful Determinant)
	return p.reg.PersistNode(quad, string(target), authClass, auditUnix)
}

// isLegalNextState enforces the "Step-Wise Metamorphosis" policy.
func (p *Pipeline) isLegalNextState(current, target State) bool {
	weights := map[State]int{
		StateConceptual: 1,
		StateHollow:     2,
		StateAnchored:   3,
		StateHydrating:  4,
		StateSequenced:  5,
	}

	// Enforce strict adjacency (Current -> Current+1)
	return weights[target] == weights[current]+1
}
