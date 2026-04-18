// Package scanner performs structural analysis of existing source code.
// It maps physical Go files back to the Identity Grammar.
// Deterministic Status: SEQUENCED (GRAMMAR-STABLE)
package scanner

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"

	"genesis/internal/identity" // Allowed L1 Import
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

// ScanBody analyzes a raw source stream to extract Identity anchors.
func (s *FileScanner) ScanBody(modulePath, packagePath string, r io.Reader) (*ScanResult, error) {
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
			if err == nil {
				// 1. Round-Trip Enforcement Gate (Chapter 12.3)
				// Verify the ID survives the canonical L1 parser before emission.
				raw := id.String()
				if _, parseErr := identity.ParseNodeID(raw); parseErr != nil {
					continue // Silently skip IDs that break grammar physics
				}

				result.Nodes = append(result.Nodes, identity.IdentityQuad{
					NodeID:         id,
					ContractID:     "pending:scanned",
					LogicHash:      "pending:scanned",
					DependencyHash: "pending:scanned",
				})
			}
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
			// Enforce unique symbol without breaking Dot-Grammar.
			// Format: ReceiverType__MethodName (e.g., Registry__PersistNode)
			symbolName = fmt.Sprintf("%s__%s", typeName, f.Name.Name)
		}
	}

	// Correct Arity Calculation
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
