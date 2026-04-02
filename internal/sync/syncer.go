package sync

import (
	"context"
	"errors"
	"fmt"
	"time"

	"neabrain/internal/domain"
)

// ObservationReader is the read-only surface the Syncer needs from the repo.
type ObservationReader interface {
	List(ctx context.Context, filter domain.ObservationListFilter) ([]domain.Observation, error)
}

// ObservationWriter is the write surface for importing observations.
// The implementation must preserve the observation's original ID and timestamps.
type ObservationWriter interface {
	ImportObservation(ctx context.Context, obs domain.Observation) (domain.Observation, error)
}

// StatusResult summarises the current sync state.
type StatusResult struct {
	SyncDir        string
	TotalChunks    int
	ImportedChunks int
	PendingChunks  int
	TotalExported  int // sum of observation counts across all chunks
}

// ExportResult describes a completed export.
type ExportResult struct {
	ChunkID   string
	Count     int
	SizeBytes int64
	AlreadyExists bool // chunk was already present (content-identical export)
}

// ImportResult describes a completed import.
type ImportResult struct {
	ChunksProcessed int
	Created         int
	Skipped         int // duplicates or already present
}

// Syncer orchestrates export and import of observations via a Transport.
type Syncer struct {
	transport Transport
	clock     func() time.Time
}

// New creates a Syncer backed by the given transport.
func New(t Transport) *Syncer {
	return &Syncer{transport: t, clock: time.Now}
}

// Export reads observations matching filter, encodes them as a new JSONL.gz
// chunk, and appends the chunk record to the manifest.
func (s *Syncer) Export(ctx context.Context, repo ObservationReader, filter domain.ObservationListFilter) (ExportResult, error) {
	observations, err := repo.List(ctx, filter)
	if err != nil {
		return ExportResult{}, fmt.Errorf("sync export: list: %w", err)
	}
	if len(observations) == 0 {
		return ExportResult{}, fmt.Errorf("sync export: no observations to export")
	}

	data, id, err := encodeChunk(observations)
	if err != nil {
		return ExportResult{}, err
	}

	// Check for content-identical chunk (same hash = same data).
	manifest, err := s.transport.ReadManifest(ctx)
	if err != nil {
		return ExportResult{}, err
	}
	if manifest.HasChunk(id) {
		return ExportResult{
			ChunkID:       id,
			Count:         len(observations),
			SizeBytes:     int64(len(data)),
			AlreadyExists: true,
		}, nil
	}

	if err := s.transport.WriteChunk(ctx, id, data); err != nil {
		return ExportResult{}, err
	}

	manifest.Chunks = append(manifest.Chunks, ChunkRecord{
		ID:        id,
		CreatedAt: s.clock(),
		Count:     len(observations),
		Size:      int64(len(data)),
	})
	if err := s.transport.WriteManifest(ctx, manifest); err != nil {
		return ExportResult{}, err
	}

	return ExportResult{
		ChunkID:   id,
		Count:     len(observations),
		SizeBytes: int64(len(data)),
	}, nil
}

// Import reads all chunks not yet marked as imported, creates observations,
// and records each successfully processed chunk in the local state.
//
// Observations that already exist (duplicate content) are silently skipped.
func (s *Syncer) Import(ctx context.Context, repo ObservationWriter) (ImportResult, error) {
	manifest, err := s.transport.ReadManifest(ctx)
	if err != nil {
		return ImportResult{}, fmt.Errorf("sync import: read manifest: %w", err)
	}

	state, err := s.transport.ReadState(ctx)
	if err != nil {
		return ImportResult{}, fmt.Errorf("sync import: read state: %w", err)
	}

	var result ImportResult
	for _, rec := range manifest.Chunks {
		if state.IsImported(rec.ID) {
			continue
		}

		data, err := s.transport.ReadChunk(ctx, rec.ID)
		if err != nil {
			return result, fmt.Errorf("sync import: read chunk %s: %w", rec.ID, err)
		}

		observations, err := decodeChunk(data)
		if err != nil {
			return result, fmt.Errorf("sync import: decode chunk %s: %w", rec.ID, err)
		}

		for _, obs := range observations {
			_, err := repo.ImportObservation(ctx, obs)
			if err != nil {
				if isDomainConflict(err) {
					result.Skipped++
					continue
				}
				return result, fmt.Errorf("sync import: create observation %s: %w", obs.ID, err)
			}
			result.Created++
		}

		state.MarkImported(rec.ID)
		result.ChunksProcessed++
	}

	if result.ChunksProcessed > 0 {
		if err := s.transport.WriteState(ctx, state); err != nil {
			return result, fmt.Errorf("sync import: write state: %w", err)
		}
	}

	return result, nil
}

// Status returns a summary of the current sync state without modifying anything.
func (s *Syncer) Status(ctx context.Context) (StatusResult, error) {
	manifest, err := s.transport.ReadManifest(ctx)
	if err != nil {
		return StatusResult{}, fmt.Errorf("sync status: %w", err)
	}

	state, err := s.transport.ReadState(ctx)
	if err != nil {
		return StatusResult{}, fmt.Errorf("sync status: %w", err)
	}

	imported := len(state.ImportedChunks)
	total := len(manifest.Chunks)
	pending := total - imported
	if pending < 0 {
		pending = 0
	}

	var totalExported int
	for _, c := range manifest.Chunks {
		totalExported += c.Count
	}

	return StatusResult{
		TotalChunks:    total,
		ImportedChunks: imported,
		PendingChunks:  pending,
		TotalExported:  totalExported,
	}, nil
}

// isDomainConflict checks for a domain conflict error (duplicate observation).
func isDomainConflict(err error) bool {
	var de domain.DomainError
	return errors.As(err, &de) && de.Code == domain.ErrorConflict
}
