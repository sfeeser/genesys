// Package scanner performs structural analysis of existing source code.
// It maps physical Go files back to the Identity Grammar.
// Deterministic Status: SEQUENCED (HARDENED-STABLE)
package scanner

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"

	"genesis/internal/identity"
)

// ScanResult represents the collection of identities found in a single file.
type ScanResult struct {
	PackageName string
	Nodes       []identity.IdentityQuad
}

// FileScanner performs AST-based discovery of node identities.
type FileScanner struct {
	fset *token.FileSet
}

// New initializes a new structural scanner.
func New() *FileScanner {
	return &FileScanner{
		fset: token.NewFileSet(),
	}
}

// CalculateLogicHash generates a deterministic SHA-256 fingerprint of an AST node.
// It hashes the formatted AST rendering, which effectively normalizes whitespace
// and excludes comments not anchored within the specific body subtree.
func CalculateLogicHash(node ast.Node) string {
	if node == nil {
		return "" // Return empty for declarations without bodies
	}

	var buf bytes.Buffer
	// RawFormat ensures we are hashing the structural intent, not developer styling.
	conf := printer.Config{Mode: printer.RawFormat, Tabwidth: 8}
	_ = conf.Fprint(&buf, token.NewFileSet(), node)

	sum := sha256.Sum256(buf.Bytes())
	return fmt.Sprintf("%x", sum[:])
}

// ScanBody analyzes a raw source stream to extract Identity anchors and fingerprints.
func (s *FileScanner) ScanBody(modulePath, packagePath string, r io.Reader) (*ScanResult, error) {
	// CHAPTER 4.1: Source Ingestion via AST
	f, err := parser.ParseFile(s.fset, "", r, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("scanner: failed to parse source: %w", err)
	}

	result := &ScanResult{
		PackageName: f.Name.Name,
	}

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			id, err := s.mapFuncToIdentity(modulePath, packagePath, d)
			if err != nil {
				continue
			}

			// CHAPTER 12.3: Round-Trip Enforcement Gate
			// Verify the ID survives the canonical L1 parser before emission.
			if _, parseErr := identity.ParseNodeID(id.String()); parseErr != nil {
				continue 
			}

			// Generate the Logic Hash from the function body
			lHash := CalculateLogicHash(d.Body)

			result.Nodes = append(result.Nodes, identity.IdentityQuad{
				NodeID:         id,
				ContractID:     "pending:contract",
				LogicHash:      lHash,
				DependencyHash: "pending:dependency",
			})
		}
	}

	return result, nil
}

// mapFuncToIdentity converts a function declaration into a formal NodeID.
func (s *FileScanner) mapFuncToIdentity(mod, pkg string, f *ast.FuncDecl) (identity.NodeID, error) {
	kind := "func"
	visibility := identity.Private
	if f.Name.IsExported() {
		visibility = identity.Exported
	}

	receiverShape := identity.None
	symbolName := f.Name.Name

	if f.Recv != nil && len(f.Recv.List) > 0 {
		field := f.Recv.List[0]
		var typeName string
		switch t := field.Type.(type) {
		case *ast.StarExpr:
			receiverShape = identity.Pointer
			if ident, ok := t.X.(*ast.Ident); ok {
				typeName = ident.Name
			}
		case *ast.Ident:
			receiverShape = identity.Value
			typeName = t.Name
		}
		
		if typeName != "" {
			// Flattening: ReceiverType__MethodName
			symbolName = fmt.Sprintf("%s__%s", typeName, f.Name.Name)
		}
	}

	// Calculate Arity based on actual parameters
	arity := 0
	if f.Type.Params != nil {
		for _, field := range f.Type.Params.List {
			namesCount := len(field.Names)
			if namesCount == 0 {
				arity++ 
			} else {
				arity += namesCount
			}
		}
	}

	return identity.NodeID{
		Kind:       identity.Kind(kind),
		Visibility: visibility,
		Module:     mod,
		Package:    pkg,
		Receiver:   receiverShape,
		Symbol:     symbolName,
		Arity:      arity,
	}, nil
}
