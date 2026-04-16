package spec

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"saayn/internal/identity"
)

// Specbook represents the normative and executable authority.
type Specbook struct {
	Version     string              `yaml:"version" json:"version"`
	Purpose     string              `yaml:"purpose" json:"purpose"`
	Contracts   map[string]Contract `yaml:"contracts" json:"contracts"`
	Fingerprint string              `yaml:"-" json:"-"`
}

// Contract defines the strict boundaries and required symbols for a single package or domain.
type Contract struct {
	PackagePath string   `yaml:"package_path" json:"package_path"`
	Signatures  []string `yaml:"signatures" json:"signatures"`
}

// Loader is responsible for reading, hashing, and validating the normative Specbook.
type Loader struct {
	workspaceRoot string
}

// NewLoader initializes a Loader tied to the staged workspace boundary.
func NewLoader(workspaceRoot string) (*Loader, error) {
	if workspaceRoot == "" {
		return nil, errors.New("validation failed: workspace root cannot be empty")
	}
	return &Loader{
		workspaceRoot: workspaceRoot,
	}, nil
}

// Load executes the file materialization, schema validation, and Identity Triad resolution.
func (l *Loader) Load() (*Specbook, error) {
	specPath := filepath.Join(l.workspaceRoot, "specbook.yaml")

	data, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("file read error: %w", err)
	}

	// 1. Strict Duplicate Key & Scalar Validation via AST
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return nil, fmt.Errorf("yaml parse error: %w", err)
	}
	if err := checkDuplicateKeys(&node); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 2. Strict Decoding (Rejects Unknown Fields)
	var book Specbook
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&book); err != nil {
		return nil, fmt.Errorf("strict decode error: %w", err)
	}

	// 3. Rigorous Structural Validation
	if book.Version == "" {
		return nil, errors.New("validation failed: version is required")
	}
	if book.Purpose == "" {
		return nil, errors.New("validation failed: purpose is required")
	}
	if len(book.Contracts) == 0 {
		return nil, errors.New("validation failed: at least one contract is required")
	}

	if err := l.validateContracts(&book); err != nil {
		return nil, err
	}

	// 4. Deterministic Fingerprint
	canonicalBytes, err := json.Marshal(book)
	if err != nil {
		return nil, fmt.Errorf("canonical serialization error: %w", err)
	}
	
	hash := sha256.Sum256(canonicalBytes)
	book.Fingerprint = hex.EncodeToString(hash[:])

	return &book, nil
}

// validateContracts ensures all required boundaries and Identity grammar rules are met.
func (l *Loader) validateContracts(book *Specbook) error {
	for name, contract := range book.Contracts {
		if contract.PackagePath == "" {
			return fmt.Errorf("validation failed: contract '%s' missing package path", name)
		}
		if len(contract.Signatures) == 0 {
			return fmt.Errorf("validation failed: contract '%s' requires at least one signature", name)
		}

		for _, sigRaw := range contract.Signatures {
			parsedID, err := identity.ParsePublicID(sigRaw)
			if err != nil {
				return fmt.Errorf("validation failed: contract '%s' signature '%s' invalid: %w", name, sigRaw, err)
			}

			if parsedID.PkgPath != contract.PackagePath {
				return fmt.Errorf("validation failed: signature '%s' package path '%s' does not match contract package '%s'", 
					sigRaw, parsedID.PkgPath, contract.PackagePath)
			}
		}
	}
	return nil
}

// checkDuplicateKeys recursively inspects a yaml.Node to ensure no duplicate keys exist in mappings,
// and strictly enforces that all mapping keys are scalars. Explicit switch guarantees deterministic branching.
func checkDuplicateKeys(node *yaml.Node) error {
	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			if err := checkDuplicateKeys(child); err != nil {
				return err
			}
		}
	case yaml.MappingNode:
		seen := make(map[string]bool)
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			
			if keyNode.Kind != yaml.ScalarNode {
				return errors.New("non-scalar key detected in yaml mapping")
			}
			
			if seen[keyNode.Value] {
				return fmt.Errorf("duplicate key detected in yaml mapping: %s", keyNode.Value)
			}
			seen[keyNode.Value] = true
			if err := checkDuplicateKeys(node.Content[i+1]); err != nil {
				return err
			}
		}
	case yaml.SequenceNode:
		for _, child := range node.Content {
			if err := checkDuplicateKeys(child); err != nil {
				return err
			}
		}
	}
	return nil
}

// ValidateFingerprint asserts that a given hash matches the strictly serialized Specbook fingerprint.
func (s *Specbook) ValidateFingerprint(proposedHash string) error {
	if len(proposedHash) != 64 {
		return fmt.Errorf("validation failed: invalid fingerprint length %d, expected 64", len(proposedHash))
	}
	if _, err := hex.DecodeString(proposedHash); err != nil {
		return fmt.Errorf("validation failed: fingerprint must be a valid hex string: %w", err)
	}
	if s.Fingerprint != proposedHash {
		return fmt.Errorf("validation failed: fingerprint mismatch, expected %s, got %s", s.Fingerprint, proposedHash)
	}
	return nil
}
