# NeaBrain

NeaBrain is a single-user memory system that delivers consistent behavior across CLI, HTTP, MCP, and TUI interfaces. The architecture isolates core domain logic from adapter and storage concerns so all interfaces share identical semantics.

## Goals
- Provide consistent CRUD, search, dedupe, topic upsert, and session behavior across adapters.
- Keep domain rules stable while allowing storage, FTS, and interface implementations to evolve.
- Support local-first operation with clear configuration and override precedence.

## Architecture
NeaBrain follows a hexagonal (ports and adapters) architecture:
- Core entities: Observation, Topic, Session, Duplicate.
- Inbound ports: ObservationService, SearchService, TopicService, SessionService, ConfigService.
- Outbound ports: ObservationRepository, TopicRepository, SessionRepository, DuplicateRepository, SearchIndex, Clock.
- Adapters: CLI, HTTP, MCP, TUI; plus local storage, FTS, config, and clock implementations.

Diagrams (Mermaid specs):
- `docs/diagrams/hexagonal-architecture.md`
- `docs/diagrams/data-flow.md`
- `docs/diagrams/storage-schema.md`

## Install

### From source
Prereqs: Go 1.22 or newer.

```bash
git clone <repo-url>
cd MotorBD
go build -o neabrain ./cmd/neabrain
```

Run directly:

```bash
go run ./cmd/neabrain --help
```

### Binary (optional)
Prebuilt binaries are not published yet. Placeholder for future releases:

```text
# TODO: add release URLs once published
```

## Quick start

Create an observation:

```bash
./neabrain observation create --content "hello" --project "demo" --topic "onboarding" --tags "cli"
```

Search observations:

```bash
./neabrain search --query "hello" --project "demo"
```

Run HTTP server:

```bash
./neabrain serve --addr 127.0.0.1:8080
```

Run MCP server:

```bash
./neabrain mcp
```

Run TUI:

```bash
./neabrain tui
```

## Configuration
Defaults, overrides, and environment variables are documented in `docs/operations.md`.

Summary:
- Config directory: os.UserConfigDir()/neabrain
- Config file: <config dir>/config.json
- Storage path: <config dir>/neabrain.db
- FTS path: defaults to storage path when unset
- Precedence: CLI overrides > environment variables > config file > defaults

Environment variables:
- NEABRAIN_STORAGE_PATH
- NEABRAIN_FTS_PATH
- NEABRAIN_DEFAULT_PROJECT
- NEABRAIN_DEDUPE_POLICY
- NEABRAIN_CONFIG_FILE

## CLI

Top-level commands:
- `observation <create|read|update|list|delete>`
- `search`
- `topic upsert`
- `session <open|resume|update-disclosure>`
- `config show`
- `serve`
- `mcp`
- `tui`

Config override flags (available on most commands):
- `--storage-path`
- `--fts-path`
- `--default-project`
- `--dedupe-policy`
- `--config-file`

Example:

```bash
./neabrain observation list --project "demo" --tags "cli" --storage-path ./data/neabrain.db
```

## HTTP API
All endpoints are served by `serve`.

Observations:
- `POST /observations`
- `GET /observations`
- `GET /observations/{id}`
- `PATCH /observations/{id}`
- `DELETE /observations/{id}`

Search:
- `GET /search?query=...&project=...&topic_key=...&tags=tag1,tag2&include_deleted=true`

Topics:
- `PUT /topics/{topic_key}`

Sessions:
- `POST /sessions`
- `POST /sessions/{id}/resume`
- `PATCH /sessions/{id}`

## MCP tools
MCP server exposes the following tools via JSON-RPC:
- `observation.create`
- `observation.read`
- `observation.update`
- `observation.list`
- `observation.delete`
- `search`
- `topic.upsert`
- `session.open`
- `session.resume`
- `session.update_disclosure`
- `config.show`

## OpenCode MCP plugin
This repo includes an OpenCode MCP plugin package for NeaBrain:
- Package: `plugins/opencode-mcp`
- Install and usage: `docs/opencode-mcp.md`
- Adapter: `plugins/opencode-mcp/adapter.ts` (registers `nbn_*` tool aliases and compaction hooks)

## Verification

Tests:

```bash
go test ./...
```

End-to-end smoke test (CLI, HTTP, MCP):

```powershell
./scripts/e2e_smoke.ps1
```
