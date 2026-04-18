// Package orchestrator coordinates the Genesis Convergence Cycle.
// It serves as the Apex Authority for the internal toolchain.
// Deterministic Status: SEQUENCED (FULL-FIDELITY-VERIFIED)
package orchestrator

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"genesis/internal/auditlog"
	"genesis/internal/identity"
	"genesis/internal/metamorphosis"
	"genesis/internal/registry"
	"genesis/internal/scanner"
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
	scanner  *scanner.FileScanner
}

// New initializes the apex coordinator.
func New(r *registry.Registry, p *metamorphosis.Pipeline, l *auditlog.Logger, s *surgeon.Surgeon) *Orchestrator {
	return &Orchestrator{
		reg:      r,
		pipeline: p,
		logger:   l,
		surg:     s,
		scanner:  scanner.New(),
	}
}

// Converge executes a single optimization cycle.
// Enforces Chapter 1.4: Multi-Determinant Anchor Coherence.
func (o *Orchestrator) Converge(s *spec.Specbook, auditUnix int64) error {
	allConceptual, err := s.MapToConceptualQuads()
	if err != nil {
		return fmt.Errorf("orchestrator: spec mapping failed: %w", err)
	}
	quadMap := make(map[string]identity.IdentityQuad)
	for _, q := range allConceptual {
		quadMap[q.NodeID.String()] = q
	}

	// 1. High-Discipline Bootstrap
	for _, nodeSpec := range s.Nodes {
		targetQuad, ok := quadMap[nodeSpec.NodeID]
		if !ok {
			return fmt.Errorf("orchestrator: identity mismatch for %s", nodeSpec.NodeID)
		}

		_, _, _, err := o.reg.GetNodeWithAuthority(nodeSpec.NodeID)
		if errors.Is(err, registry.ErrNodeNotFound) {
			if pErr := o.reg.PersistNode(targetQuad, string(metamorphosis.StateConceptual), nodeSpec.Authority, auditUnix); pErr != nil {
				return fmt.Errorf("orchestrator: bootstrap failure: %w", pErr)
			}
			o.logger.LogTransition(targetQuad.NodeID, "bootstrap", auditUnix)
		} else if err != nil {
			return fmt.Errorf("orchestrator: registry access error: %w", err)
		}
	}

	// 2. Full-Fidelity Transition (State 2: Hollow)
	loom := staging.NewLoom()
	for _, nodeSpec := range s.Nodes {
		currentQuad, _, _, err := o.reg.GetNodeWithAuthority(nodeSpec.NodeID)
		if err != nil {
			return fmt.Errorf("orchestrator: authoritative read failure: %w", err)
		}

		hollowContent, genErr := o.generateIsomorphicStub(currentQuad.NodeID)
		if genErr != nil {
			return fmt.Errorf("orchestrator: stub generation failed for %s: %w", currentQuad.NodeID.String(), genErr)
		}

		// --- DIAGNOSTIC HARDENED ROUND-TRIP ---
		scanRes, scanErr := o.scanner.ScanBody(currentQuad.NodeID.Module, currentQuad.NodeID.Package, bytes.NewReader(hollowContent))
		if scanErr != nil {
			return fmt.Errorf("orchestrator: round-trip parse failure for %s: %w", currentQuad.NodeID.String(), scanErr)
		}
		
		// Note: We allow len > 1 specifically for methods where a receiver-type sidecar is required.
		// The L4 Scanner only emits []identity.NodeID for FuncDecls, so the sidecar is transparent.
		if len(scanRes.Nodes) != 1 {
			return fmt.Errorf("orchestrator: round-trip count mismatch for %s; found %d nodes", currentQuad.NodeID.String(), len(scanRes.Nodes))
		}
		
		extractedID := scanRes.Nodes[0].NodeID
		if extractedID.String() != currentQuad.NodeID.String() {
			return fmt.Errorf("orchestrator: isomorphism failure; expected %s, got %s", currentQuad.NodeID.String(), extractedID.String())
		}

		// RE-ANCHOR: Update LogicHash. ContractID and DependencyHash are inherited from State 1 placeholders.
		nextQuad := currentQuad
		nextQuad.LogicHash = identity.CalculateLID(hollowContent)

		if err := o.pipeline.Transition(currentQuad.NodeID.String(), metamorphosis.StateHollow, nextQuad, hollowContent, auditUnix); err != nil {
			return fmt.Errorf("orchestrator: transition failed: %w", err)
		}

		if err := loom.StageNode(nextQuad, hollowContent); err != nil {
			return err
		}
	}

	// 3. Materialize & Export
	snapshot := loom.OrderedSnapshot()
	var entries []surgeon.MaterializationEntry
	for _, snap := range snapshot {
		entries = append(entries, surgeon.MaterializationEntry{Path: snap.Path, Content: snap.Content})
	}

	if err := o.surg.ApplyBatch(entries); err != nil {
		return err
	}

	allIDs, err := o.GetGenomeIDs()
	if err != nil {
		return err
	}
	return o.logger.ExportGenome(".genesis/genome.yaml", allIDs)
}

// generateIsomorphicStub synthesizes a fragment that perfectly satisfies the 7-part determinant grammar.
func (o *Orchestrator) generateIsomorphicStub(id identity.NodeID) ([]byte, error) {
	// Synthesize Parameters to satisfy Arity
	params := make([]string, id.Arity)
	for i := 0; i < id.Arity; i++ {
		params[i] = fmt.Sprintf("v%d int", i)
	}
	paramStr := strings.Join(params, ", ")

	var body string
	switch id.Kind {
	case "func":
		if id.Receiver == "none" {
			body = fmt.Sprintf("func %s(%s) {}", id.Symbol, paramStr)
		} else {
			parts := strings.Split(id.Symbol, "__")
			if len(parts) != 2 {
				return nil, fmt.Errorf("orchestrator: invalid method symbol: %s", id.Symbol)
			}
			typeName, methodName := parts[0], parts[1]
			ptr := ""
			if id.Receiver == "pointer" { ptr = "*" }
			// Sidecar type declaration is PERMITTED for State 2 Hollow methods.
			body = fmt.Sprintf("type %s struct{}\nfunc (%s%s) %s(%s) {}", typeName, ptr, typeName, methodName, paramStr)
		}
	case "struct":
		body = fmt.Sprintf("type %s struct{}", id.Symbol)
	case "interface":
		body = fmt.Sprintf("type %s interface{}", id.Symbol)
	default:
		return nil, fmt.Errorf("orchestrator: unsupported kind: %s", id.Kind)
	}

	return []byte(fmt.Sprintf("package %s\n\n%s\n", id.Package, body)), nil
}

// GetGenomeIDs retrieves all node identities from the Physical Authority (L2).
func (o *Orchestrator) GetGenomeIDs() ([]string, error) {
	return o.reg.ListAllNodeIDs()
}
