// Package staging implements the Virtual Loom (VFS).
// It stages code fragments for Hexagonal Gate verification.
// Deterministic Status: SEQUENCED (CANONICAL-ORDER)
package staging

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"sync"

	"genesis/internal/identity" // Allowed L1 Import
)

// StagedNode pairs the raw content with its full Identity Quad.
type StagedNode struct {
	Quad    identity.IdentityQuad
	Content []byte
}

// VirtualFile represents a stable, anchored file in the Loom.
type VirtualFile struct {
	Path  string
	Nodes map[string]StagedNode // Keyed by NodeID.String() for O(1) deduplication
}

// SnapshotEntry provides a stable, ordered view of Loom state.
type SnapshotEntry struct {
	Path    string
	Content []byte
}

// NewLoom initializes a fresh staging environment.
func NewLoom() *Loom {
	return &Loom{
		files: make(map[string]*VirtualFile),
	}
}

// Loom is the Virtual File System where Hydration (State 4) occurs.
type Loom struct {
	mu    sync.RWMutex
	files map[string]*VirtualFile
}

// StageNode injects a node into the Loom with Isomorphic Path Projection.
// Enforces Chapter 5: Arrival order is ignored; content assembly is identity-sorted.
func (l *Loom) StageNode(quad identity.IdentityQuad, content []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	id := quad.NodeID
	// Canonical Projection: module/package.go
	// We adopt a "Package-Level Aggregation" model with stable inner-file sorting.
	path := fmt.Sprintf("%s/%s.go", id.Module, id.Package)

	f, exists := l.files[path]
	if !exists {
		f = &VirtualFile{
			Path:  path,
			Nodes: make(map[string]StagedNode),
		}
		l.files[path] = f
	}

	nodeKey := id.String()
	if existing, exists := f.Nodes[nodeKey]; exists {
		if !bytes.Equal(existing.Content, content) {
			return fmt.Errorf("staging: deterministic conflict for node %s", nodeKey)
		}
		return nil
	}

	f.Nodes[nodeKey] = StagedNode{
		Quad:    quad,
		Content: content,
	}
	return nil
}

// assembleContent produces bit-identical bytes regardless of insertion order.
// Enforces Chapter 5: Lexical sort by NodeID.
func (f *VirtualFile) assembleContent() []byte {
	keys := make([]string, 0, len(f.Nodes))
	for k := range f.Nodes {
		keys = append(keys, k)
	}
	sort.Strings(keys) // Stable identity-based sort

	var buf bytes.Buffer
	for _, k := range keys {
		buf.Write(f.Nodes[k].Content)
		// Ensure deterministic spacing between staged fragments
		if !bytes.HasSuffix(f.Nodes[k].Content, []byte("\n")) {
			buf.WriteByte('\n')
		}
	}
	return buf.Bytes()
}

// Load reads a staged file's canonical content.
func (l *Loom) Load(path string) (io.Reader, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	f, exists := l.files[path]
	if !exists {
		return nil, fmt.Errorf("staging: path not found: %s", path)
	}

	return bytes.NewReader(f.assembleContent()), nil
}

// OrderedSnapshot returns a stable, sorted slice of Loom state.
func (l *Loom) OrderedSnapshot() []SnapshotEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	paths := make([]string, 0, len(l.files))
	for p := range l.files {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	snapshot := make([]SnapshotEntry, 0, len(paths))
	for _, p := range paths {
		snapshot = append(snapshot, SnapshotEntry{
			Path:    p,
			Content: l.files[p].assembleContent(),
		})
	}
	return snapshot
}
