// Package orchestrator coordinates the Genesis Convergence Cycle.
// It serves as the Apex Authority for the internal toolchain.
// Deterministic Status: SEQUENCED (FULL-ALIGNMENT)
package orchestrator

import (
	"errors"
	"fmt"

	"genesis/internal/auditlog"
	"genesis/internal/identity"
	"genesis/internal/metamorphosis"
	"genesis/internal/registry"
	"genesis/internal/spec"
	"genesis/internal/staging"
	"genesis/internal/surgeon"
)

// Orchestrator manages the end-to-end materialization lifecycle.
type Orchestrator struct {
	reg      *registry.Registry
	pipeline *metamorphosis.Pipeline
	logger   *auditlog.Logger
	surg     *surgeon.Surgeon
}

// New initializes the apex coordinator.
func New(r *registry.Registry, p *metamorphosis.Pipeline, l *auditlog.Logger, s *surgeon.Surgeon) *Orchestrator {
	return &Orchestrator{reg: r, pipeline: p, logger: l, surg: s}
}

// Converge executes a single optimization cycle.
// Enforces Chapter 1.4 & 1.5: Truthful state transitions and SCC-aware materialization.
func (o *Orchestrator) Converge(s *spec.Specbook, auditUnix int64) error {
	// 1. Precise Ingestion & Identity Mapping
	allConceptual, err := s.MapToConceptualQuads()
	if err != nil {
		return fmt.Errorf("orchestrator: spec mapping failed: %w", err)
	}

	quadMap := make(map[string]identity.IdentityQuad)
	for _, q := range allConceptual {
		quadMap[q.NodeID.String()] = q
	}

	// 2. High-Discipline Bootstrap (State 1: Conceptual)
	// Uses the L2 sentinel contract to distinguish missing nodes from DB errors.
	for _, nodeSpec := range s.Nodes {
		targetQuad, ok := quadMap[nodeSpec.NodeID]
		if !ok {
			return fmt.Errorf("orchestrator: identity mismatch for %s", nodeSpec.NodeID)
		}

		_, _, _, err := o.reg.GetNodeWithAuthority(nodeSpec.NodeID)
		if errors.Is(err, registry.ErrNodeNotFound) {
			if pErr := o.reg.PersistNode(targetQuad, string(metamorphosis.StateConceptual), nodeSpec.Authority, auditUnix); pErr != nil {
				return fmt.Errorf("orchestrator: bootstrap failure for %s: %w", nodeSpec.NodeID, pErr)
			}
			o.logger.LogTransition(targetQuad.NodeID, "bootstrap", auditUnix)
		} else if err != nil {
			return fmt.Errorf("orchestrator: registry access error for %s: %w", nodeSpec.NodeID, err)
		}
	}

	// 3. Kind-Aware Hollow Transition (State 2: Hollow)
	// Enforces Chapter 12.3: Stubs must preserve the authoritative identity grammar.
	loom := staging.NewLoom()
	for _, nodeSpec := range s.Nodes {
		currentQuad, _, _, err := o.reg.GetNodeWithAuthority(nodeSpec.NodeID)
		if err != nil {
			return fmt.Errorf("orchestrator: authoritative read failure: %w", err)
		}

		hollowContent := o.generateIsomorphicStub(currentQuad.NodeID)

		// RE-ANCHOR: Update LogicHash to match the generated physical reality
		nextQuad := currentQuad
		nextQuad.LogicHash = identity.CalculateLID(hollowContent)

		// Governed Transition: Moves the Registry from Conceptual -> Hollow
		if err := o.pipeline.Transition(currentQuad.NodeID.String(), metamorphosis.StateHollow, nextQuad, hollowContent, auditUnix); err != nil {
			return fmt.Errorf("orchestrator: transition to hollow failed for %s: %w", currentQuad.NodeID.String(), err)
		}

		if err := loom.StageNode(nextQuad, hollowContent); err != nil {
			return fmt.Errorf("orchestrator: loom staging failure: %w", err)
		}
	}

	// 4. Physical Materialization & Audit Export
	snapshot := loom.OrderedSnapshot()
	var entries []surgeon.MaterializationEntry
	for _, snap := range snapshot {
		entries = append(entries, surgeon.MaterializationEntry{Path: snap.Path, Content: snap.Content})
	}

	if err := o.surg.ApplyBatch(entries); err != nil {
		return fmt.Errorf("orchestrator: surgery failure: %w", err)
	}

	// Final Step: Export the truthful genome state for Git auditing
	allIDs, err := o.GetGenomeIDs()
	if err != nil {
		return fmt.Errorf("orchestrator: genome discovery failed for export: %w", err)
	}
	return o.logger.ExportGenome(".genesis/genome.yaml", allIDs)
}

// generateIsomorphicStub creates a Go source fragment that truthfully represents the NodeID.
func (o *Orchestrator) generateIsomorphicStub(id identity.NodeID) []byte {
	var stub string
	switch id.Kind {
	case "func":
		if id.Receiver == "none" {
			stub = fmt.Sprintf("func %s() {}\n", id.Symbol)
		} else {
			ptr := ""
			if id.Receiver == "pointer" {
				ptr = "*"
			}
			// Note: For methods, we assume a generic stub type that preserves receiver structure
			stub = fmt.Sprintf("func (%sGENESIS_STUB) %s() {}\n", ptr, id.Symbol)
		}
	case "struct":
		stub = fmt.Sprintf("type %s struct {}\n", id.Symbol)
	case "interface":
		stub = fmt.Sprintf("type %s interface {}\n", id.Symbol)
	default:
		stub = fmt.Sprintf("// genesis: unknown kind %s for %s\n", id.Kind, id.Symbol)
	}

	return []byte(fmt.Sprintf("package %s\n\n%s", id.Package, stub))
}

// GetGenomeIDs retrieves all node identities from the Physical Authority (L2).
func (o *Orchestrator) GetGenomeIDs() ([]string, error) {
	return o.reg.ListAllNodeIDs()
}
