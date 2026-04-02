package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// FSTransport implements Transport using the local filesystem.
//
// Directory layout:
//
//	<root>/
//	  manifest.json
//	  sync_state.json
//	  chunks/
//	    <sha256>.jsonl.gz
type FSTransport struct {
	root string
}

// NewFSTransport creates a Transport rooted at dir (created on demand).
func NewFSTransport(dir string) *FSTransport {
	return &FSTransport{root: dir}
}

func (t *FSTransport) ReadManifest(_ context.Context) (Manifest, error) {
	data, err := os.ReadFile(t.manifestPath())
	if os.IsNotExist(err) {
		return Manifest{Version: 1}, nil
	}
	if err != nil {
		return Manifest{}, fmt.Errorf("sync: read manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("sync: parse manifest: %w", err)
	}
	if m.Version == 0 {
		m.Version = 1
	}
	return m, nil
}

func (t *FSTransport) WriteManifest(_ context.Context, m Manifest) error {
	if err := t.ensureDirs(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("sync: marshal manifest: %w", err)
	}
	return os.WriteFile(t.manifestPath(), append(data, '\n'), 0o644)
}

func (t *FSTransport) WriteChunk(_ context.Context, id string, data []byte) error {
	if err := t.ensureDirs(); err != nil {
		return err
	}
	return os.WriteFile(t.chunkPath(id), data, 0o644)
}

func (t *FSTransport) ReadChunk(_ context.Context, id string) ([]byte, error) {
	data, err := os.ReadFile(t.chunkPath(id))
	if err != nil {
		return nil, fmt.Errorf("sync: read chunk %s: %w", id, err)
	}
	return data, nil
}

func (t *FSTransport) ChunkExists(_ context.Context, id string) (bool, error) {
	_, err := os.Stat(t.chunkPath(id))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func (t *FSTransport) ReadState(_ context.Context) (State, error) {
	data, err := os.ReadFile(t.statePath())
	if os.IsNotExist(err) {
		return State{}, nil
	}
	if err != nil {
		return State{}, fmt.Errorf("sync: read state: %w", err)
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return State{}, fmt.Errorf("sync: parse state: %w", err)
	}
	return s, nil
}

func (t *FSTransport) WriteState(_ context.Context, s State) error {
	if err := t.ensureDirs(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("sync: marshal state: %w", err)
	}
	return os.WriteFile(t.statePath(), append(data, '\n'), 0o644)
}

func (t *FSTransport) ensureDirs() error {
	if err := os.MkdirAll(filepath.Join(t.root, "chunks"), 0o755); err != nil {
		return fmt.Errorf("sync: mkdir: %w", err)
	}
	return nil
}

func (t *FSTransport) manifestPath() string { return filepath.Join(t.root, "manifest.json") }
func (t *FSTransport) statePath() string    { return filepath.Join(t.root, "sync_state.json") }
func (t *FSTransport) chunkPath(id string) string {
	return filepath.Join(t.root, "chunks", id+".jsonl.gz")
}
