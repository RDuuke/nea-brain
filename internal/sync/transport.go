// Package sync provides portable, conflict-free observation sync via immutable chunks.
//
// Each export creates a new JSONL.gz chunk file identified by its SHA-256 hash.
// A manifest.json tracks all known chunks. A local sync_state.json records which
// chunks have already been imported on this machine, enabling idempotent imports
// across multiple machines without merge conflicts.
package sync

import "context"

// Transport abstracts chunk and manifest storage.
// The default implementation uses the local filesystem; future implementations
// can target S3, Dropbox, or any other backend.
type Transport interface {
	// ReadManifest returns the current manifest, or an empty manifest if none exists.
	ReadManifest(ctx context.Context) (Manifest, error)
	// WriteManifest persists the manifest atomically.
	WriteManifest(ctx context.Context, m Manifest) error
	// WriteChunk stores a compressed JSONL chunk. id is its SHA-256 hex digest.
	WriteChunk(ctx context.Context, id string, data []byte) error
	// ReadChunk retrieves a chunk by its SHA-256 id.
	ReadChunk(ctx context.Context, id string) ([]byte, error)
	// ChunkExists reports whether a chunk with the given id already exists.
	ChunkExists(ctx context.Context, id string) (bool, error)
	// ReadState returns the local sync state (imported chunk IDs).
	ReadState(ctx context.Context) (State, error)
	// WriteState persists the local sync state.
	WriteState(ctx context.Context, s State) error
}
