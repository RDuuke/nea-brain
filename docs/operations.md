# NeaBrain Operations

## Configuration

### Defaults
- Config directory: os.UserConfigDir()/neabrain
- Config file: <config dir>/config.json
- Storage path: <config dir>/neabrain.db
- FTS path: defaults to storage path when unset
- Default project: empty
- Dedupe policy: exact
 - Startup creates the config directory and storage parent directories if missing

### Precedence (highest first)
1) CLI overrides
2) Environment variables
3) Config file
4) Defaults

### Environment Variables
- NEABRAIN_STORAGE_PATH
- NEABRAIN_FTS_PATH
- NEABRAIN_DEFAULT_PROJECT
- NEABRAIN_DEDUPE_POLICY
- NEABRAIN_CONFIG_FILE

### Config File Example
```json
{
  "storage_path": "./data/neabrain.db",
  "fts_path": "./data/neabrain.db",
  "default_project": "personal",
  "dedupe_policy": "exact"
}
```

### Relative Paths
- Config file paths are resolved relative to the current working directory.
- Storage/FTS paths from the config file are resolved relative to the config file location.
- Storage/FTS paths from CLI overrides or environment variables are resolved relative to the current working directory.
- Storage/FTS paths are normalized (cleaned separators and dot segments) before use; parent directories are created if missing.

### CLI Overrides
```bash
go run ./cmd/neabrain observation list --storage-path ./data/neabrain.db --default-project personal
```

## Verification

### Tests
```bash
go test ./...
```

### End-to-End Smoke (CLI, HTTP, MCP)
```powershell
./scripts/e2e_smoke.ps1
```

### Manual TUI Check
```bash
go run ./cmd/neabrain tui
```

Example commands to run inside the prompt:
```text
observation create --content "hello" --project "smoke" --topic "onboarding" --tags "tui"
search --query "hello" --project "smoke"
exit
```
