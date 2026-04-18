// Package surgeon executes the Physical Grafting of synthesized code.
// It implements the Durable Atomic Swap and Sibling-Isolated Surgery Protocol.
// Deterministic Status: SEQUENCED (DURABLE)
package surgeon

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"genesis/internal/identity" // Allowed L1 Import
)

// MaterializationEntry decouples the surgeon from internal/staging.
// This preserves Sibling Isolation (Chapter 12.1).
type MaterializationEntry struct {
	Path    string
	Content []byte
}

// Surgeon handles the materialization of code to the physical disk.
type Surgeon struct {
	targetRoot string
}

// New initializes a Surgeon anchored to a specific project root.
func New(root string) *Surgeon {
	return &Surgeon{targetRoot: root}
}

// ApplyBatch materializes a set of files. 
// Note: Transactional atomicity across multiple files requires a 
// filesystem-level 'commit' or 'rollback' which is handled by L10 Orchestration.
func (s *Surgeon) ApplyBatch(entries []MaterializationEntry) error {
	for _, entry := range entries {
		if err := s.DurableGraft(entry.Path, entry.Content); err != nil {
			return fmt.Errorf("surgeon: durable graft failed for %s: %w", entry.Path, err)
		}
	}
	return nil
}

// DurableGraft implements the Hardened Atomic Swap Pattern.
// Enforces Chapter 5.3: Collision-safe, Durable, and Atomic.
func (s *Surgeon) DurableGraft(relativePath string, content []byte) error {
	fullPath := filepath.Join(s.targetRoot, relativePath)
	dir := filepath.Dir(fullPath)

	// 1. Ensure Directory Topology
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("surgeon: mkdir failure: %w", err)
	}

	// 2. Create Collision-Safe Temp File
	// Using os.CreateTemp prevents concurrent or stale-file collisions.
	tmpFile, err := os.CreateTemp(dir, ".genesis-swap-*.tmp")
	if err != nil {
		return fmt.Errorf("surgeon: failed to create safe temp file: %w", err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName) // Cleanup if rename doesn't happen

	// 3. Write and Sync (Durable Barrier)
	if _, err := tmpFile.Write(content); err != nil {
		tmpFile.Close()
		return fmt.Errorf("surgeon: write failure: %w", err)
	}
	
	// Ensure data is physically on the platter/NAND before the swap
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("surgeon: fsync failure: %w", err)
	}
	tmpFile.Close()

	// 4. Atomic Rename
	if err := os.Rename(tmpName, fullPath); err != nil {
		return fmt.Errorf("surgeon: rename failure: %w", err)
	}

	// 5. Directory Sync (Durability Barrier for Metadata)
	// On many filesystems, the rename isn't durable until the parent dir is synced.
	d, err := os.Open(dir)
	if err == nil {
		d.Sync()
		d.Close()
	}

	return nil
}

// VerifyLID performs a Gate B Check (Identity Coherence).
// Validates that the physical content matches the expected Logic Hash (L-ID).
func (s *Surgeon) VerifyLID(expectedLID string, actualContent []byte) bool {
	actualHash := identity.CalculateLID(actualContent)
	return actualHash == expectedLID
}
