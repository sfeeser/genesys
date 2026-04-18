// Package main is the entry point for the Genesis Engine CLI.
// It assembles the internal toolchain and initiates the convergence cycle.
// Deterministic Status: SEQUENCED (SHELL-ALIGNMENT)
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"genesis/internal/audit"
	"genesis/internal/auditlog"
	"genesis/internal/metamorphosis"
	"genesis/internal/orchestrator"
	"genesis/internal/registry"
	"genesis/internal/spec"
	"genesis/internal/surgeon"
)

func main() {
	// 1. CLI Parameter Definition
	specPath := flag.String("spec", "genesis.yaml", "Path to the normative Specbook")
	dbPath := flag.String("db", "genome.db", "Path to the Registry database")
	targetPath := flag.String("out", ".", "Target root for materialization")
	flag.Parse()

	// 2. Assembly: Physical Authority (L2)
	// SQLite DSN remains consistent with the L2 registry.Open contract.
	reg, err := registry.Open(fmt.Sprintf("file:%s", *dbPath))
	if err != nil {
		log.Fatalf("GENESIS CRITICAL: Failed to open Registry: %v", err)
	}

	// 3. Assembly: Evaluators & Historians (L7, L8)
	auditor := audit.New()
	logger := auditlog.New(reg)

	// 4. Assembly: State Machine & Surgery (L9, L6)
	pipeline := metamorphosis.New(reg, auditor)
	surg := surgeon.New(*targetPath)

	// 5. Assembly: Apex Orchestrator (L10)
	orch := orchestrator.New(reg, pipeline, logger, surg)

	// 6. Ingestion: Normative Spec (L3 Interface Alignment)
	// Enforces the io.Reader contract established in L3.
	f, err := os.Open(*specPath)
	if err != nil {
		log.Fatalf("GENESIS CRITICAL: Failed to open Specbook file: %v", err)
	}
	defer f.Close()

	specbook, err := spec.Ingest(f)
	if err != nil {
		log.Fatalf("GENESIS CRITICAL: Failed to ingest Specbook: %v", err)
	}

	// 7. Execution: The Convergence Cycle
	// CAPTURE DETERMINANT: time.Now() is allowed ONLY at this shell boundary.
	auditUnix := time.Now().UTC().Unix()

	fmt.Printf("GENESIS: Initiating Convergence (Audit ID: %d)\n", auditUnix)
	
	if err := orch.Converge(specbook, auditUnix); err != nil {
		fmt.Fprintf(os.Stderr, "\nGENESIS CONVERGENCE FAILED:\n%v\n", err)
		os.Exit(1)
	}

	fmt.Println("GENESIS: Convergence Successful. Genome Sequenced.")
}
