package sync_test

import (
	"context"
	"testing"
	"time"

	"neabrain/internal/domain"
	neasync "neabrain/internal/sync"
)

// --- in-memory transport ---

type memTransport struct {
	manifest neasync.Manifest
	state    neasync.State
	chunks   map[string][]byte
}

func newMemTransport() *memTransport {
	return &memTransport{
		manifest: neasync.Manifest{Version: 1},
		chunks:   map[string][]byte{},
	}
}

func (m *memTransport) ReadManifest(_ context.Context) (neasync.Manifest, error) { return m.manifest, nil }
func (m *memTransport) WriteManifest(_ context.Context, man neasync.Manifest) error {
	m.manifest = man
	return nil
}
func (m *memTransport) WriteChunk(_ context.Context, id string, data []byte) error {
	m.chunks[id] = data
	return nil
}
func (m *memTransport) ReadChunk(_ context.Context, id string) ([]byte, error) { return m.chunks[id], nil }
func (m *memTransport) ChunkExists(_ context.Context, id string) (bool, error) {
	_, ok := m.chunks[id]
	return ok, nil
}
func (m *memTransport) ReadState(_ context.Context) (neasync.State, error) { return m.state, nil }
func (m *memTransport) WriteState(_ context.Context, s neasync.State) error {
	m.state = s
	return nil
}

// --- in-memory observation repo ---

type memRepo struct {
	items map[string]domain.Observation
}

func newMemRepo(obs ...domain.Observation) *memRepo {
	r := &memRepo{items: map[string]domain.Observation{}}
	for _, o := range obs {
		r.items[o.ID] = o
	}
	return r
}

func (r *memRepo) List(_ context.Context, _ domain.ObservationListFilter) ([]domain.Observation, error) {
	out := make([]domain.Observation, 0, len(r.items))
	for _, o := range r.items {
		out = append(out, o)
	}
	return out, nil
}

func (r *memRepo) ImportObservation(_ context.Context, obs domain.Observation) (domain.Observation, error) {
	if _, exists := r.items[obs.ID]; exists {
		return domain.Observation{}, domain.NewConflict("observation already exists")
	}
	r.items[obs.ID] = obs
	return obs, nil
}

// --- tests ---

func obs(id, content string) domain.Observation {
	return domain.Observation{
		ID:        id,
		Content:   content,
		CreatedAt: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		Project:   "test",
	}
}

func TestExportCreatesChunkAndManifest(t *testing.T) {
	tr := newMemTransport()
	repo := newMemRepo(obs("o1", "hello"), obs("o2", "world"))
	s := neasync.New(tr)

	result, err := s.Export(context.Background(), repo, domain.ObservationListFilter{})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if result.Count != 2 {
		t.Fatalf("expected 2 observations, got %d", result.Count)
	}
	if result.ChunkID == "" {
		t.Fatal("expected non-empty ChunkID")
	}
	if len(tr.manifest.Chunks) != 1 {
		t.Fatalf("expected 1 manifest entry, got %d", len(tr.manifest.Chunks))
	}
	if _, ok := tr.chunks[result.ChunkID]; !ok {
		t.Fatal("chunk not stored in transport")
	}
}

func TestExportIdempotentOnIdenticalContent(t *testing.T) {
	tr := newMemTransport()
	repo := newMemRepo(obs("o1", "hello"))
	s := neasync.New(tr)

	r1, _ := s.Export(context.Background(), repo, domain.ObservationListFilter{})
	r2, err := s.Export(context.Background(), repo, domain.ObservationListFilter{})
	if err != nil {
		t.Fatalf("second Export failed: %v", err)
	}
	if r1.ChunkID != r2.ChunkID {
		t.Fatal("expected same chunk id for identical content")
	}
	if !r2.AlreadyExists {
		t.Fatal("expected AlreadyExists on second export")
	}
	if len(tr.manifest.Chunks) != 1 {
		t.Fatalf("expected manifest to stay at 1 chunk, got %d", len(tr.manifest.Chunks))
	}
}

func TestImportCreatesObservations(t *testing.T) {
	tr := newMemTransport()
	exportRepo := newMemRepo(obs("o1", "hello"), obs("o2", "world"))
	s := neasync.New(tr)
	if _, err := s.Export(context.Background(), exportRepo, domain.ObservationListFilter{}); err != nil {
		t.Fatal(err)
	}

	importRepo := newMemRepo()
	result, err := s.Import(context.Background(), importRepo)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}
	if result.Created != 2 {
		t.Fatalf("expected 2 created, got %d", result.Created)
	}
	if result.ChunksProcessed != 1 {
		t.Fatalf("expected 1 chunk processed, got %d", result.ChunksProcessed)
	}
	if len(importRepo.items) != 2 {
		t.Fatalf("expected 2 observations in import repo, got %d", len(importRepo.items))
	}
}

func TestImportIdempotent(t *testing.T) {
	tr := newMemTransport()
	exportRepo := newMemRepo(obs("o1", "hello"))
	s := neasync.New(tr)
	if _, err := s.Export(context.Background(), exportRepo, domain.ObservationListFilter{}); err != nil {
		t.Fatal(err)
	}

	importRepo := newMemRepo()
	s.Import(context.Background(), importRepo) //nolint
	result, err := s.Import(context.Background(), importRepo)
	if err != nil {
		t.Fatalf("second Import failed: %v", err)
	}
	if result.ChunksProcessed != 0 {
		t.Fatalf("expected 0 chunks processed on second import, got %d", result.ChunksProcessed)
	}
	if len(importRepo.items) != 1 {
		t.Fatalf("expected 1 observation, got %d", len(importRepo.items))
	}
}

func TestImportSkipsDuplicates(t *testing.T) {
	tr := newMemTransport()
	exportRepo := newMemRepo(obs("o1", "hello"))
	s := neasync.New(tr)
	s.Export(context.Background(), exportRepo, domain.ObservationListFilter{}) //nolint

	// Pre-populate the import repo with the same observation
	importRepo := newMemRepo(obs("o1", "hello"))
	result, err := s.Import(context.Background(), importRepo)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}
	if result.Skipped != 1 {
		t.Fatalf("expected 1 skipped, got %d", result.Skipped)
	}
	if result.Created != 0 {
		t.Fatalf("expected 0 created, got %d", result.Created)
	}
}

func TestStatus(t *testing.T) {
	tr := newMemTransport()
	repo := newMemRepo(obs("o1", "a"), obs("o2", "b"))
	s := neasync.New(tr)
	s.Export(context.Background(), repo, domain.ObservationListFilter{}) //nolint

	status, err := s.Status(context.Background())
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if status.TotalChunks != 1 {
		t.Fatalf("expected 1 total chunk, got %d", status.TotalChunks)
	}
	if status.PendingChunks != 1 {
		t.Fatalf("expected 1 pending, got %d", status.PendingChunks)
	}
	if status.TotalExported != 2 {
		t.Fatalf("expected 2 total exported, got %d", status.TotalExported)
	}
}

func TestFSTransportRoundTrip(t *testing.T) {
	dir := t.TempDir()
	tr := neasync.NewFSTransport(dir)
	repo := newMemRepo(obs("o1", "hello"), obs("o2", "world"))
	s := neasync.New(tr)

	r1, err := s.Export(context.Background(), repo, domain.ObservationListFilter{})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	importRepo := newMemRepo()
	imp, err := s.Import(context.Background(), importRepo)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}
	if imp.Created != 2 {
		t.Fatalf("expected 2 created, got %d", imp.Created)
	}
	_ = r1
}
