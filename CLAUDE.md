# DevForge MCP

For full project context, conventions, build commands, and agent roles, see [AGENTS.md](AGENTS.md).

## Claude Code notes

- MCP server binary: `./dev-forge-mcp` (stdio transport — attach via `mcpServers` in your MCP client config)
- Skills are in `.agents/skills/` (symlinked from `.claude/skills/`)
- PRPs live in `PRPs/` — read the relevant PRP before starting any feature implementation
