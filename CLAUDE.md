# DevForge MCP

For full project context, conventions, build commands, and agent roles, see [AGENTS.md](AGENTS.md).

## Claude Code notes

- MCP server binary: `./dist/devforge-mcp` (stdio transport — attach via `mcpServers` in your MCP client config)
- CLI/TUI binary: `./dist/devforge`
- Media processing: `bin/dpf` (DevPixelForge) — requires FFmpeg for video/audio
- Skills are in `.agents/skills/` (symlinked from `.claude/skills/`)
- PRPs live in `PRPs/` — read the relevant PRP before starting any feature implementation

## Build commands

```bash
# Build Go binaries
CGO_ENABLED=1 go build ./...

# Build Rust binary (requires Rust toolchain)
make build-rust

# Full distribution build
make dist
```
