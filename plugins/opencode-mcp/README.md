# OpenCode MCP Plugin: NeaBrain

This package lets OpenCode start the NeaBrain MCP server via `neabrain mcp` and
loads a lightweight OpenCode plugin adapter.

## Install
Copy this folder into your OpenCode plugins directory and register it in `~/.opencode/opencode.json`.
Ensure `@opencode-ai/plugin` is installed (via `~/.config/opencode/package.json` or a project
`.opencode/package.json`).

See `docs/opencode-mcp.md` for full install and usage instructions.

## Adapter behavior
- Registers NeaBrain MCP tool aliases prefixed with `nbn_`.
- Injects brief memory instructions into the system prompt.
- On compaction, stores a short session summary and refreshes context.
