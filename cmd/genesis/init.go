// Package main handles the Cobra CLI surface.
// Path: genesis/cmd/genesis/init.go
// Deterministic Status: SEQUENCED (CERTIFIED-STABLE)
package main

import (
	"errors"
	"fmt"
	"os"

	"genesis/internal/registry"
	"github.com/spf13/cobra"
)

var forceInit bool

func init() {
	initCmd.Flags().BoolVar(&forceInit, "force", false, "Overwrite existing genome database")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Code Genome registry",
	Long: `Bootstrap the Genesis Engine for the current project. 
This creates the SQLite registry and initializes the physical schema.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Genesis Init | Audit ID: %d\n", auditUnix)

		// 1. CHAPTER 14.1 Safeguard: Exhaustive Path Inspection
		info, err := os.Stat(genomePath)
		switch {
		case err == nil:
			// Path exists. We must verify it's not a directory.
			if info.IsDir() {
				return fmt.Errorf("%w: genome path is a directory: %s", 
					ErrBoundaryViolation, genomePath)
			}
			// If not forced, we refuse to mutate existing state.
			if !forceInit {
				return fmt.Errorf("%w: genome already exists at %s (use --force to overwrite)", 
					ErrBoundaryViolation, genomePath)
			}
			// DESTRUCTIVE PHYSICS: Forced removal MUST be successful.
			fmt.Printf("⚠️  Removing existing genome: %s\n", genomePath)
			if err := os.Remove(genomePath); err != nil {
				return fmt.Errorf("%w: failed to remove existing genome: %v", 
					ErrBoundaryViolation, err)
			}
		case errors.Is(err, os.ErrNotExist):
			// Target is clear. Proceeding to bootstrap.
		default:
			// Unexpected IO error (Permission, bad path, etc.)
			return fmt.Errorf("%w: unable to inspect path %s: %v", 
				ErrBoundaryViolation, genomePath, err)
		}

		// 2. CHAPTER 2.1: Bootstrap Physical Authority
		// We build a simple file DSN; registry.Open handles the query-param hardening.
		dsn := fmt.Sprintf("file:%s", genomePath)
		
		reg, err := registry.Open(dsn)
		if err != nil {
			return fmt.Errorf("initialization failure: %w", err)
		}
		defer reg.Close()

		fmt.Printf("✅ Registry Initialized: %s\n", genomePath)
		fmt.Println("--------------------------------------------------")
		fmt.Println("Project is now ready for enrichment.")
		
		return nil
	},
}
