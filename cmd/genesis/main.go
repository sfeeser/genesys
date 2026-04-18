// Package main is the entry point for the Genesis Engine CLI.
// It initiates the Cobra command tree and injects the global audit clock.
// Deterministic Status: SEQUENCED (APEX-ALIGNMENT)
package main

import (
	"time"
)

func main() {
	// CHAPTER 15: CLOCK INJECTION
	// The only legal point for non-deterministic time capture.
	// This value is shared across the entire command execution.
	auditUnix = time.Now().UTC().Unix()

	// CHAPTER 14.1: STRUCTURAL INVARIANT
	// The Apex MUST NOT contain business logic; it triggers the Root command.
	Execute()
}
