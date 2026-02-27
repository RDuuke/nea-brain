package outbound

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	ports "neabrain/internal/ports/outbound"
)

func TestNormalizePath_TrimsAndCleans(t *testing.T) {
	input := "  data/./nested/../db.sqlite  "
	want := filepath.Clean(filepath.FromSlash("data/./nested/../db.sqlite"))

	if got := normalizePath(input); got != want {
		t.Fatalf("normalizePath(%q) = %q, want %q", input, got, want)
	}

	if got := normalizePath("   "); got != "" {
		t.Fatalf("normalizePath(blank) = %q, want empty string", got)
	}

	if runtime.GOOS == "windows" && strings.Contains(normalizePath("data/one/two"), "/") {
		t.Fatalf("normalizePath should convert slashes on Windows")
	}
}

func TestResolve_NormalizesPaths(t *testing.T) {
	tempDir := t.TempDir()
	workingDir := mustGetwd(t)
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(workingDir)
	})

	storage := " data/./nested/../db.sqlite "
	fts := " fts//index.db "
	overrides := ports.ConfigOverrides{
		StoragePath: strPtr(storage),
		FTSPath:     strPtr(fts),
	}

	adapter := NewConfigAdapter()
	resolved, err := adapter.Resolve(context.Background(), overrides)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	wantStorage := filepath.Clean(filepath.Join(tempDir, strings.TrimSpace(storage)))
	wantFTS := filepath.Clean(filepath.Join(tempDir, strings.TrimSpace(fts)))

	if resolved.StoragePath != wantStorage {
		t.Fatalf("Resolve storage path = %q, want %q", resolved.StoragePath, wantStorage)
	}
	if resolved.FTSPath != wantFTS {
		t.Fatalf("Resolve fts path = %q, want %q", resolved.FTSPath, wantFTS)
	}
}

func TestLoadFileConfig_NormalizesRelativePaths(t *testing.T) {
	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "config.json")
	data := []byte(`{"storage_path":"data/../store/db.sqlite","fts_path":"fts/./index.db"}`)
	if err := os.WriteFile(cfgPath, data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := loadFileConfig(cfgPath)
	if err != nil {
		t.Fatalf("loadFileConfig: %v", err)
	}

	wantStorage := filepath.Clean(filepath.Join(tempDir, "data/../store/db.sqlite"))
	wantFTS := filepath.Clean(filepath.Join(tempDir, "fts/./index.db"))

	if cfg.StoragePath != wantStorage {
		t.Fatalf("storage path = %q, want %q", cfg.StoragePath, wantStorage)
	}
	if cfg.FTSPath != wantFTS {
		t.Fatalf("fts path = %q, want %q", cfg.FTSPath, wantFTS)
	}
}

func TestLoad_CreatesParentDirs(t *testing.T) {
	tempDir := t.TempDir()
	storagePath := filepath.Join(tempDir, "new", "dir", "neabrain.db")
	overrides := ports.ConfigOverrides{
		StoragePath: strPtr(storagePath),
		FTSPath:     strPtr(storagePath),
	}

	adapter := NewConfigAdapter()
	cfg, err := adapter.Load(context.Background(), overrides)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	parentDir := filepath.Dir(storagePath)
	info, err := os.Stat(parentDir)
	if err != nil {
		t.Fatalf("stat parent dir: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("parent path is not a dir: %s", parentDir)
	}
	if cfg.StoragePath != filepath.Clean(storagePath) {
		t.Fatalf("storage path = %q, want %q", cfg.StoragePath, filepath.Clean(storagePath))
	}
	if cfg.FTSPath != filepath.Clean(storagePath) {
		t.Fatalf("fts path = %q, want %q", cfg.FTSPath, filepath.Clean(storagePath))
	}
}

func strPtr(value string) *string {
	return &value
}

func mustGetwd(t *testing.T) string {
	t.Helper()
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return workingDir
}
