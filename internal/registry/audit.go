// Package audit implements the Hexagonal Gates for mutation verification.
// It serves as the final barrier before State 5 (Sequencing).
// Deterministic Status: SEQUENCED (VERACITY-ANCHORED)
package audit

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"

	"genesis/internal/identity" // Allowed L1 Import
	"genesis/internal/scanner"  // Allowed L4 Import
)

// Verdict represents the outcome of an audit gate.
type Verdict struct {
	Gate    string
	Passed  bool
	Message string
}

// Auditor coordinates the verification of Mutation Worksets.
type Auditor struct {
	scanner *scanner.FileScanner
}

// New initializes a new Conscience for the engine.
func New() *Auditor {
	return &Auditor{
		scanner: scanner.New(),
	}
}

// AuditNode executes verified hexagonal gates.
// Enforces Chapter 6: Verifies NodeID and L-ID with strict singularity.
func (a *Auditor) AuditNode(quad identity.IdentityQuad, content []byte) []Verdict {
	var results []Verdict

	// --- GATE A: PHYSICS (Full Syntax Verification) ---
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, "", content, parser.AllErrors)
	
	physicsPassed := err == nil
	results = append(results, Verdict{
		Gate:    "A: Physics",
		Passed:  physicsPassed,
		Message: func() string { if err != nil { return err.Error() }; return "" }(),
	})

	if !physicsPassed {
		return results // Fail fast
	}

	// --- GATE B: IDENTITY (Structural & Logic Anchor) ---
	// Reduced Claim: This auditor verifies what is physically extractable at L7.
	
	// 1. L-ID Verification (The Logic Anchor)
	actualLID := identity.CalculateLID(content)
	lidMatch := actualLID == quad.LogicHash

	// 2. NodeID Singularity (The Structural Anchor)
	scanRes, scanErr := a.scanner.ScanBody(quad.NodeID.Module, quad.NodeID.Package, bytes.NewReader(content))
	
	var idPassed bool
	var idMsg string

	if scanErr != nil {
		idPassed = false
		idMsg = fmt.Sprintf("Scanner failure: %v", scanErr)
	} else if len(scanRes.Nodes) != 1 {
		idPassed = false
		idMsg = fmt.Sprintf("Singularity Failure: found %d nodes, expected 1", len(scanRes.Nodes))
	} else {
		foundNode := scanRes.Nodes[0]
		idMatch := foundNode.NodeID.String() == quad.NodeID.String()
		
		if idMatch && lidMatch {
			idPassed = true
		} else {
			idMsg = fmt.Sprintf("Anchor Mismatch [NodeID: %t, L-ID: %t]", idMatch, lidMatch)
		}
	}

	results = append(results, Verdict{
		Gate:    "B: Identity",
		Passed:  idPassed,
		Message: idMsg,
	})

	return results
}

// VerifyReplay (Gate E) ensures the sequence is bit-identical to the plan.
func (a *Auditor) VerifyReplay(plannedLID string, content []byte) Verdict {
	actualLID := identity.CalculateLID(content)
	passed := actualLID == plannedLID
	
	v := Verdict{Gate: "E: Replay", Passed: passed}
	if !passed {
		v.Message = fmt.Sprintf("REPLAY FAILURE: Hash %s != Expected %s", actualLID, plannedLID)
	}
	return v
}
