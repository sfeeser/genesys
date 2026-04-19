// Package orchestrator coordinates the Genesis Convergence and Enrichment cycles.
// Path: internal/orchestrator/orchestrator.go
// Deterministic Status: SEQUENCED (CERTIFIED-STABLE)
package orchestrator

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
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

type Orchestrator struct {
	reg      *registry.Registry
	pipeline *metamorphosis.Pipeline
	logger   *auditlog.Logger
	surg     *surgeon.Surgeon
	scanner  *scanner.FileScanner
}

func New(r *registry.Registry, p *metamorphosis.Pipeline, l *auditlog.Logger, s *surgeon.Surgeon) *Orchestrator {
	return &Orchestrator{
		reg:      r,
		pipeline: p,
		logger:   l,
		surg:     s,
		scanner:  scanner.New(),
	}
}

// =============================================================================
// ENRICHMENT HEMISPHERE (Bottom-Up: Disk -> Registry)
// =============================================================================

func (o *Orchestrator) Enrich(modulePath, targetRoot string, auditUnix int64) error {
	err := filepath.WalkDir(targetRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil { return err }
		if d.IsDir() {
			if d.Name() == "vendor" || d.Name() == "node_modules" || strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		return o.enrichFile(modulePath, targetRoot, path, auditUnix)
	})
	if err != nil {
		return fmt.Errorf("orchestrator: enrichment failed: %w", err)
	}
	return nil
}

func (o *Orchestrator) enrichFile(modulePath, targetRoot, filePath string, auditUnix int64) error {
	f, err := os.Open(filePath)
	if err != nil { return err }
	defer f.Close()

	relPath, err := filepath.Rel(targetRoot, filepath.Dir(filePath))
	if err != nil {
		return fmt.Errorf("orchestrator: path derivation failure for %s: %w", filePath, err)
	}
	packagePath := strings.ReplaceAll(relPath, string(os.PathSeparator), "/")
	if packagePath == "." { packagePath = "" }

	res, err := o.scanner.ScanBody(modulePath, packagePath, f)
	if err != nil {
		return fmt.Errorf("orchestrator: scan failure on %s: %w", filePath, err)
	}

	for _, quad := range res.Nodes {
		if err := o.reg.PersistNode(quad, "hydrated", 0, auditUnix); err != nil {
			return fmt.Errorf("orchestrator: hydration failure for %s: %w", quad.NodeID.String(), err)
		}
	}
	return nil
}

// =============================================================================
// CONVERGENCE HEMISPHERE (Top-Down: Spec -> Disk)
// =============================================================================

func (o *Orchestrator) Converge(s *spec.Specbook, auditUnix int64) error {
	allConceptual, err := s.MapToConceptualQuads()
	if err != nil { return err }
	
	quadMap := make(map[string]identity.IdentityQuad)
	for _, q := range allConceptual {
		quadMap[q.NodeID.String()] = q
	}

	// 1. Bootstrap (State 1)
	for _, nodeSpec := range s.Nodes {
		targetQuad := quadMap[nodeSpec.NodeID]
		_, _, _, err := o.reg.GetNodeWithAuthority(nodeSpec.NodeID)
		if errors.Is(err, registry.ErrNodeNotFound) {
			if err := o.reg.PersistNode(targetQuad, "draft", nodeSpec.Authority, auditUnix); err != nil {
				return fmt.Errorf("orchestrator: bootstrap failure for %s: %w", nodeSpec.NodeID, err)
			}
			o.logger.LogTransition(targetQuad.NodeID, "bootstrap", auditUnix)
		} else if err != nil {
			return fmt.Errorf("orchestrator: registry read failure: %w", err)
		}
	}

	// 2. Hollow Transition (State 2)
	loom := staging.NewLoom()
	for _, nodeSpec := range s.Nodes {
		currentQuad, _, _, err := o.reg.GetNodeWithAuthority(nodeSpec.NodeID)
		if err != nil {
			return fmt.Errorf("orchestrator: authoritative read failure: %w", err)
		}

		if currentQuad.NodeID.Kind != "func" { continue }

		hollowContent, genErr := o.generateIsomorphicStub(currentQuad.NodeID)
		if genErr != nil { return genErr }

		// --- CHAPTER 12.3: STRUCTURAL ISOMORPHISM GATE ---
		scanRes, scanErr := o.scanner.ScanBody(currentQuad.NodeID.Module, currentQuad.NodeID.Package, bytes.NewReader(hollowContent))
		if scanErr != nil {
			return fmt.Errorf("orchestrator: round-trip parse failure for %s: %w", currentQuad.NodeID.String(), scanErr)
		}
		if len(scanRes.Nodes) != 1 {
			return fmt.Errorf("orchestrator: round-trip count mismatch for %s; found %d", currentQuad.NodeID.String(), len(scanRes.Nodes))
		}
		if scanRes.Nodes[0].NodeID.String() != currentQuad.NodeID.String() {
			return fmt.Errorf("orchestrator: isomorphism mismatch; expected %s, got %s", currentQuad.NodeID.String(), scanRes.Nodes[0].NodeID.String())
		}

		// TRUTHFUL ANCHORING: Extract LogicHash from generated AST body
		fset := token.NewFileSet()
		f, _ := parser.ParseFile(fset, "", hollowContent, 0)
		var bodyNode ast.Node
		for _, d := range f.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok { bodyNode = fn.Body; break }
		}

		nextQuad := currentQuad
		nextQuad.LogicHash = scanner.CalculateLogicHash(bodyNode)

		if err := o.pipeline.Transition(currentQuad.NodeID.String(), metamorphosis.StateHollow, nextQuad, hollowContent, auditUnix); err != nil {
			return fmt.Errorf("orchestrator: transition failure: %w", err)
		}
		if err := loom.StageNode(nextQuad, hollowContent); err != nil {
			return fmt.Errorf("orchestrator: staging failure: %w", err)
		}
	}

	// 3. Materialize & Export
	snapshot := loom.OrderedSnapshot()
	var entries []surgeon.MaterializationEntry
	for _, snap := range snapshot {
		entries = append(entries, surgeon.MaterializationEntry{Path: snap.Path, Content: snap.Content})
	}
	if err := o.surg.ApplyBatch(entries); err != nil {
		return fmt.Errorf("orchestrator: materialization failure: %w", err)
	}

	allIDs, err := o.GetGenomeIDs()
	if err != nil {
		return fmt.Errorf("orchestrator: identity discovery failure: %w", err)
	}
	return o.logger.ExportGenome(".genesis/genome.yaml", allIDs)
}

func (o *Orchestrator) generateIsomorphicStub(id identity.NodeID) ([]byte, error) {
	pkgName := "main"
	if id.Package != "" {
		parts := strings.Split(id.Package, "/")
		pkgName = parts[len(parts)-1]
	}

	params := make([]string, id.Arity)
	for i := 0; i < id.Arity; i++ {
		params[i] = fmt.Sprintf("v%d int", i)
	}
	paramStr := strings.Join(params, ", ")

	var body string
	if id.Receiver == "none" {
		body = fmt.Sprintf("func %s(%s) {}", id.Symbol, paramStr)
	} else {
		parts := strings.Split(id.Symbol, "__")
		if len(parts) != 2 {
			return nil, fmt.Errorf("orchestrator: invalid method symbol grammar: %s", id.Symbol)
		}
		typeName, methodName := parts[0], parts[1]
		ptr := ""
		if id.Receiver == "pointer" { ptr = "*" }
		body = fmt.Sprintf("type %s struct{}\nfunc (%s%s) %s(%s) {}", typeName, ptr, typeName, methodName, paramStr)
	}
	return []byte(fmt.Sprintf("package %s\n\n%s\n", pkgName, body)), nil
}

func (o *Orchestrator) GetGenomeIDs() ([]string, error) {
	return o.reg.ListAllNodeIDs()
}
