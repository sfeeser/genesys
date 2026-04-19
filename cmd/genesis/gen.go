// Package main handles the Cobra CLI surface.
// Path: genesis/cmd/genesis/gen.go
// Deterministic Status: SEQUENCED (CERTIFIED-STABLE)
package main

import (
	"fmt"
	"os"

	"genesis/internal/audit"
	"genesis/internal/auditlog"
	"genesis/internal/metamorphosis"
	"genesis/internal/orchestrator"
	"genesis/internal/registry"
	"genesis/internal/spec"
	"genesis/internal/surgeon"
	"github.com/spf13/cobra"
)

func init() {
	// CHAPTER 14.2b: Promotion of 'gen' to Certified status.
	// We reuse the 'specPath' flag defined in root.go.
	rootCmd.AddCommand(genCmd)
}

var genCmd = &cobra.Command{
	Use:   "gen [target-directory]",
	Short: "Composite: apply spec logic to the project",
	Long: `The 'gen' command executes the orchestrator's top-down 
convergence workflow, mapping the normative Specbook to physical Go stubs.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := args[0]

		fmt.Printf("Genesis Gen (Composite) | Audit ID: %d\n", auditUnix)

		// 1. CHAPTER 12.3: Ingest the Normative Spec (L3)
		// We open the file at the shell boundary and pass the reader to the internal API.
		sf, err := os.Open(specPath)
		if err != nil {
			return fmt.Errorf("gen: failed to open specbook at %s: %w", specPath, err)
		}
		defer sf.Close()

		s, err := spec.Ingest(sf)
		if err != nil {
			return fmt.Errorf("gen: spec ingest failure: %w", err)
		}

		// 2. CHAPTER 2.1: Open Physical Authority (L2)
		reg, err := registry.Open(fmt.Sprintf("file:%s", genomePath))
		if err != nil {
			return fmt.Errorf("gen: registry failure: %w", err)
		}
		defer reg.Close()

		// 3. Initialize Conducting Toolchain via Audited Constructors
		auditor := audit.New()
		logger := auditlog.New(reg)
		pipe := metamorphosis.New(reg, auditor)
		surg := surgeon.New(targetDir)
		
		orch := orchestrator.New(reg, pipe, logger, surg)

		// 4. CHAPTER 10: Execute Top-Down Convergence Cycle
		if err := orch.Converge(s, auditUnix); err != nil {
			return err // Already wrapped by orchestrator
		}

		fmt.Println("--------------------------------------------------")
		fmt.Println("✅ Generation Complete. Spec enforced.")
		
		return nil
	},
}
