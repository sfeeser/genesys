// Package identity provides the foundational types and logic for the Genesis 
// Identity Quad. It is the L1 root of the Package Topology.
// Deterministic Status: SEQUENCED (REPAIRED)
package identity

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
)

// Kind represents the architectural nature of a node.
type Kind string

// Visibility defines the access scope of the symbol.
type Visibility string

const (
	Exported Visibility = "exported"
	Private  Visibility = "private"
)

// ReceiverShape defines the method attachment physics.
type ReceiverShape string

const (
	None    ReceiverShape = "none"
	Value   ReceiverShape = "value"
	Pointer ReceiverShape = "pointer"
)

// NodeID represents the environment-stable determinant.
type NodeID struct {
	Kind       Kind
	Visibility Visibility
	Module     string
	Package    string
	Receiver   ReceiverShape
	Symbol     string
	Arity      int
}

// String renders the NodeID into its canonical grammar form.
func (n NodeID) String() string {
	return fmt.Sprintf("%s.%s.%s.%s.%s.%s.%d",
		n.Kind,
		n.Visibility,
		n.Module,
		n.Package,
		n.Receiver,
		n.Symbol,
		n.Arity,
	)
}

// ParseNodeID deconstructs a canonical string into a NodeID struct.
// Enforces Chapter 12.3: Identity Grammar Enforcement.
func ParseNodeID(raw string) (NodeID, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 7 {
		return NodeID{}, fmt.Errorf("invalid identity grammar: expected 7 parts, got %d", len(parts))
	}

	// 1. Receiver Validation (Chapter 12.3 Enforcement)
	rec := ReceiverShape(parts[4])
	switch rec {
	case None, Value, Pointer:
		// Valid
	default:
		return NodeID{}, fmt.Errorf("invalid receiver shape: %s", rec)
	}

	// 2. Arity Parsing (Deterministic Round-Trip)
	arity, err := strconv.Atoi(parts[6])
	if err != nil {
		return NodeID{}, fmt.Errorf("invalid arity: %s", parts[6])
	}

	return NodeID{
		Kind:       Kind(parts[0]),
		Visibility: Visibility(parts[1]),
		Module:     parts[2],
		Package:    parts[3],
		Receiver:   rec,
		Symbol:     parts[5],
		Arity:      arity,
	}, nil
}

// IdentityQuad represents the four immutable dimensions of a node.
type IdentityQuad struct {
	NodeID         NodeID
	ContractID     string // C-ID
	LogicHash      string // L-ID
	DependencyHash string // D-ID
}

// CalculateLID generates the Logic-ID for a given source fragment.
// Note: Future iterations will include AST-normalization for order-independence.
func CalculateLID(source []byte) string {
	h := sha256.New()
	h.Write(source)
	return fmt.Sprintf("%x", h.Sum(nil))
}
