# Codex integration

NeaBrain exposes its tools via MCP. Codex connects to it through `~/.codex/config.toml`.

## Quick setup

Print the config snippet:

```bash
neabrain setup codex
```

Install it directly:

```bash
neabrain setup codex --install
```

This writes:

```toml
[mcp_servers.neabrain]
command = "/path/to/neabrain"
args = ["mcp"]
```

Replace `/path/to/neabrain` with the actual path, or keep `neabrain` if it is in `PATH`.

## Remove

```bash
neabrain setup codex --uninstall
```

