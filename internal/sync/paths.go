package sync

import (
	"fmt"
	"os"
	"path/filepath"
)

// DefaultSyncDir returns the default sync directory for the current user.
// On all platforms: $HOME/.config/neabrain/sync
func DefaultSyncDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("sync: cannot resolve home directory: %w", err)
	}
	return filepath.Join(home, ".config", "neabrain", "sync"), nil
}
