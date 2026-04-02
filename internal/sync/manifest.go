package sync

import "time"

// Manifest is the shared index of all known sync chunks.
// It is safe to commit to git — new chunks are only appended, never modified.
type Manifest struct {
	Version int           `json:"version"`
	Chunks  []ChunkRecord `json:"chunks"`
}

// ChunkRecord describes a single exported chunk.
type ChunkRecord struct {
	ID        string    `json:"id"`         // SHA-256 hex digest of the compressed data
	CreatedAt time.Time `json:"created_at"` // Wall-clock time of the export
	Count     int       `json:"count"`      // Number of observations in the chunk
	Size      int64     `json:"size"`       // Compressed size in bytes
}

// State is the per-machine record of which chunks have been imported.
// It lives alongside the manifest but is not shared between machines.
type State struct {
	ImportedChunks []string `json:"imported_chunks"`
}

// HasChunk reports whether the manifest already contains a chunk with the given id.
func (m Manifest) HasChunk(id string) bool {
	for _, c := range m.Chunks {
		if c.ID == id {
			return true
		}
	}
	return false
}

// IsImported reports whether a chunk has been imported in the local state.
func (s State) IsImported(id string) bool {
	for _, c := range s.ImportedChunks {
		if c == id {
			return true
		}
	}
	return false
}

// MarkImported adds id to the imported chunks list if not already present.
func (s *State) MarkImported(id string) {
	if !s.IsImported(id) {
		s.ImportedChunks = append(s.ImportedChunks, id)
	}
}
