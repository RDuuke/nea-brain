package sync

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"neabrain/internal/domain"
)

// encodeChunk serialises observations as newline-delimited JSON and compresses
// the result with gzip. Returns the compressed bytes and their SHA-256 digest.
func encodeChunk(observations []domain.Observation) (data []byte, id string, err error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	enc := json.NewEncoder(gz)
	for _, obs := range observations {
		if err := enc.Encode(obs); err != nil {
			return nil, "", fmt.Errorf("sync: encode observation %s: %w", obs.ID, err)
		}
	}
	if err := gz.Close(); err != nil {
		return nil, "", fmt.Errorf("sync: gzip close: %w", err)
	}

	compressed := buf.Bytes()
	hash := sha256.Sum256(compressed)
	return compressed, hex.EncodeToString(hash[:]), nil
}

// decodeChunk decompresses and deserialises a JSONL.gz chunk into observations.
func decodeChunk(data []byte) ([]domain.Observation, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("sync: gzip open: %w", err)
	}
	defer gz.Close()

	var observations []domain.Observation
	dec := json.NewDecoder(gz)
	for {
		var obs domain.Observation
		if err := dec.Decode(&obs); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("sync: decode observation: %w", err)
		}
		observations = append(observations, obs)
	}
	return observations, nil
}
