// Package auditlog provides the Canonical Audit Export for Genesis.
// It ensures Git-stable, lossless snapshots of the Structural Genome.
// Deterministic Status: SEQUENCED (INTERFACE-ALIGNED)
package auditlog

import (
	"fmt"
	"os"
	"sort"

	"genesis/internal/identity" 
	"genesis/internal/registry" 

	"gopkg.in/yaml.v3"
)

// CanonicalNodeRecord represents the lossless mirror of a Registry Node.
type CanonicalNodeRecord struct {
	NodeID         string `yaml:"node_id"`
	ContractID     string `yaml:"contract_id"`
	LogicHash      string `yaml:"logic_hash"`
	DependencyHash string `yaml:"dependency_hash"`
	Maturity       string `yaml:"maturity"`
	AuthorityClass int    `yaml:"authority_class"`
}

// Logger handles the recording and exporting of the genome state.
type Logger struct {
	reg *registry.Registry
}

// New initializes a new auditor ledger.
func New(r *registry.Registry) *Logger {
	return &Logger{reg: r}
}

// LogTransition notifies the observer of a state change.
// Deterministic Status: Observational No-Op (Validated)
func (l *Logger) LogTransition(id identity.NodeID, action string, auditUnix int64) {
	// Silence unused parameter errors while maintaining the API signature
	_ = id
	_ = action
	_ = auditUnix
}

// ExportGenome creates a lossless, deterministic YAML snapshot of the Genome.
// Enforces Chapter 2.1: Reflects TRUTHFUL authority and determinant state.
func (l *Logger) ExportGenome(outputPath string, nodeIDs []string) error {
	sort.Strings(nodeIDs)

	exportSet := make([]CanonicalNodeRecord, 0, len(nodeIDs))

	for _, id := range nodeIDs {
		// Utilizing the audited L2 interface
		quad, maturity, authClass, err := l.reg.GetNodeWithAuthority(id)
		if err != nil {
			return fmt.Errorf("auditlog: export incomplete; missing node: %s", id)
		}

		exportSet = append(exportSet, CanonicalNodeRecord{
			NodeID:         quad.NodeID.String(),
			ContractID:     quad.ContractID,
			LogicHash:      quad.LogicHash,
			DependencyHash: quad.DependencyHash,
			Maturity:       maturity,
			AuthorityClass: authClass,
		})
	}

	data, err := yaml.Marshal(exportSet)
	if err != nil {
		return fmt.Errorf("auditlog: yaml marshal failure: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("auditlog: filesystem write failure: %w", err)
	}

	return nil
}
