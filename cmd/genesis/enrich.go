// Package main handles the Cobra CLI surface.
// Path: genesis/cmd/genesis/enrich.go
// Deterministic Status: SEQUENCED (CERTIFIED-STABLE)
package main

import (
	"fmt"

	"genesis/internal/audit"
	"genesis/internal/auditlog"
	"genesis/internal/metamorphosis"
	"genesis/internal/orchestrator"
	"genesis/internal/registry"
	"genesis/internal/surgeon"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(enrichCmd)
}

var enrichCmd = &cobra.Command{
	Use:   "enrich [module-path] [target-directory]",
	Short: "Scan and hydrate the registry from existing source code",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		modulePath := args[0]
		targetDir := args[1]

		fmt.Printf("Genesis Enrichment | Audit ID: %d\n", auditUnix)
		fmt.Printf("Target: %s (%s)\n", targetDir, modulePath)

		// 1. CHAPTER 2.1: Open Physical Authority (L2)
		reg, err := registry.Open(fmt.Sprintf("file:%s", genomePath))
		if err != nil {
			return fmt.Errorf("enrich: registry failure: %w", err)
		}
		defer reg.Close()

		// 2. CHAPTER 12: Wire Internal Conductors via Audited APIs
		// Initialize the Auditor (L8) and Logger (L7)
		auditor := audit.New()
		logger := auditlog.New(reg)
		
		// Initialize the Pipeline (L9) and Surgeon (L6)
		pipe := metamorphosis.New(reg, auditor)
		surg := surgeon.New(targetDir)
		
		// 3. CHAPTER 10: Instantiate Orchestrator (The Conductor)
		orch := orchestrator.New(reg, pipe, logger, surg)

		// 4. CHAPTER 14.1: Execute Discovery Cycle (Disk -> Registry)
		if err := orch.Enrich(modulePath, targetDir, auditUnix); err != nil {
			return err // Error is already wrapped by orchestrator with context
		}

		fmt.Println("--------------------------------------------------")
		fmt.Println("✅ Enrichment Complete. Genome hydrated.")
		
		return nil
	},
}
