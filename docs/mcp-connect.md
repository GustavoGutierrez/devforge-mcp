# Connecting to devforge-mcp

How to attach an MCP client to the server and how to send tool calls from the terminal.

---

## Transport: stdio — no host, no port

`devforge-mcp` uses the **MCP stdio transport**. There is no TCP socket, no
HTTP endpoint, and no port to connect to. The protocol works as follows:

1. The MCP client **spawns** `devforge-mcp` as a child process.
2. JSON-RPC 2.0 messages are exchanged over the process's **stdin / stdout**.
3. The server exits when the client closes the pipe.

This means you cannot point `curl` at a host/port — you write JSON-RPC messages
to stdin and read responses from stdout.

---

## Connecting from an MCP client

### Claude Desktop (macOS / Windows)

Edit the Claude Desktop config file:

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "devforge": {
      "command": "/home/YOUR_USERNAME/.local/bin/devforge-mcp",
      "args": [],
      "env": {
        "DEV_FORGE_CONFIG": "/home/YOUR_USERNAME/.config/devforge/config.json"
      }
    }
  }
}
```

Replace `YOUR_USERNAME` with your actual username. After saving, restart Claude
Desktop. The `devforge` tools will appear in the tool palette.

> **Working directory note**: The server must run from a directory that contains
> `./bin/dpf`. If you installed via `make install`, copy or symlink
> the imgproc binary so it is reachable from wherever the client launches the
> process:
>
> ```bash
> mkdir -p ~/.local/bin/bin
> cp bin/dpf ~/.local/bin/bin/dpf
> ```
>
> Or set the `workingDirectory` key if your client supports it.

---

### VS Code — GitHub Copilot MCP (`.vscode/mcp.json`)

Create `.vscode/mcp.json` in any workspace:

```json
{
  "servers": {
    "devforge": {
      "type": "stdio",
      "command": "/home/YOUR_USERNAME/.local/bin/devforge-mcp",
      "args": [],
      "env": {
        "DEV_FORGE_CONFIG": "/home/YOUR_USERNAME/.config/devforge/config.json"
      }
    }
  }
}
```

---

### Cursor

Open **Settings → MCP → Add server**:

| Field | Value |
|-------|-------|
| Name | `devforge` |
| Type | `stdio` |
| Command | `/home/YOUR_USERNAME/.local/bin/devforge-mcp` |

---

### OpenCode

[OpenCode](https://opencode.ai) reads MCP server config from `~/.config/opencode/config.json`
(global) or from an `opencode.json` file at the root of any project (project-local).

**Global config** `~/.config/opencode/config.json`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "devforge": {
      "type": "local",
      "command": ["/home/YOUR_USERNAME/.local/bin/devforge-mcp"],
      "environment": {
        "DEV_FORGE_CONFIG": "/home/YOUR_USERNAME/.config/devforge/config.json"
      }
    }
  }
}
```

**Project-local config** `opencode.json` (at workspace root):

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "devforge": {
      "type": "local",
      "command": ["/home/YOUR_USERNAME/.local/bin/devforge-mcp"],
      "environment": {
        "DEV_FORGE_CONFIG": "/home/YOUR_USERNAME/.config/devforge/config.json"
      }
    }
  }
}
```

After saving, run `opencode` in the terminal — the `devforge` tools will be available
in the session automatically.

> **Working directory**: OpenCode launches the server from the current project
> directory. Make sure `./bin/dpf` is reachable from there, or set
> an absolute `DEVFORGE_IMGPROC_PATH` in `environment` if you add support for it.
> The simplest approach is to keep a `bin/dpf` symlink in any
> project that uses image tools:
>
> ```bash
> mkdir -p bin
> ln -sf ~/.local/bin/dpf bin/dpf
> ```

---

### Claude Code

[Claude Code](https://docs.anthropic.com/en/docs/claude-code) supports MCP servers
through a global user config or a project-level file.

**Option A — global config** (applies to every project):

```bash
claude mcp add devforge /home/YOUR_USERNAME/.local/bin/devforge-mcp \
  -e DEV_FORGE_CONFIG=/home/YOUR_USERNAME/.config/devforge/config.json
```

This writes the entry to `~/.claude.json` automatically. To verify:

```bash
claude mcp list
```

**Option B — project-level `.mcp.json`** (checked into the repo, shared with the team):

Create `.mcp.json` at the workspace root:

```json
{
  "mcpServers": {
    "devforge": {
      "command": "/home/YOUR_USERNAME/.local/bin/devforge-mcp",
      "args": [],
      "env": {
        "DEV_FORGE_CONFIG": "/home/YOUR_USERNAME/.config/devforge/config.json"
      }
    }
  }
}
```

**Option C — manual edit of `~/.claude.json`**:

```json
{
  "mcpServers": {
    "devforge": {
      "command": "/home/YOUR_USERNAME/.local/bin/devforge-mcp",
      "args": [],
      "env": {
        "DEV_FORGE_CONFIG": "/home/YOUR_USERNAME/.config/devforge/config.json"
      }
    }
  }
}
```

After any of the above, start a new `claude` session — the `devforge` tools are
available immediately, no restart needed.

---

### Claude Desktop on Ubuntu (Linux)

Claude Desktop on Linux stores its config at:

```
~/.config/Claude/claude_desktop_config.json
```

Create the file if it does not exist:

```bash
mkdir -p ~/.config/Claude
cat > ~/.config/Claude/claude_desktop_config.json <<'EOF'
{
  "mcpServers": {
    "devforge": {
      "command": "/home/YOUR_USERNAME/.local/bin/devforge-mcp",
      "args": [],
      "env": {
        "DEV_FORGE_CONFIG": "/home/YOUR_USERNAME/.config/devforge/config.json"
      }
    }
  }
}
EOF
```

Replace `YOUR_USERNAME` with your actual username (`echo $USER`). Restart Claude Desktop
after saving. The tools appear in the tool palette on the next session.

---

### Any MCP-compatible client

All MCP clients that support the stdio transport follow the same pattern:

| Setting | Value |
|---------|-------|
| Transport | `stdio` |
| Command | Absolute path to `devforge-mcp` binary |
| Args | _(none)_ |
| Working dir | Directory containing `./bin/dpf` |
| Env | `DEV_FORGE_CONFIG` (optional — only if overriding the default path) |

---

## Testing from the terminal

Because the transport is stdio, "curl-style" testing means **piping JSON-RPC
messages to the binary**. Each message is a single-line JSON object followed by
a newline.

### Smoke test — initialize

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1"}}}' \
  | ~/.local/bin/devforge-mcp
```

Expected response:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": { "tools": {} },
    "serverInfo": { "name": "devforge", "version": "1.0.0" }
  }
}
```

---

### List available tools

```bash
printf '%s\n%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1"}}}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' \
  | ~/.local/bin/devforge-mcp
```

---

### Tool call: `suggest_color_palettes`

```bash
printf '%s\n%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1"}}}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"suggest_color_palettes","arguments":{"mood":"calm and professional","count":3}}}' \
  | ~/.local/bin/devforge-mcp
```

---

### Tool call: `analyze_layout`

```bash
LAYOUT=$(cat testdata/layouts/hero.html | jq -Rs .)

printf '%s\n%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1"}}}' \
  "{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/call\",\"params\":{\"name\":\"analyze_layout\",\"arguments\":{\"markup\":$LAYOUT,\"stack\":{\"css_mode\":\"tailwind-v4\",\"framework\":\"next\"},\"page_type\":\"landing\",\"device_focus\":\"responsive\"}}}" \
  | ~/.local/bin/devforge-mcp
```

---

### Tool call: `suggest_layout`

```bash
printf '%s\n%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1"}}}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"suggest_layout","arguments":{"description":"SaaS landing page with hero, features grid, and pricing table","stack":{"css_mode":"tailwind-v4","framework":"next"},"fidelity":"mid"}}}' \
  | ~/.local/bin/devforge-mcp
```

---

### Tool call: `manage_tokens` (read)

```bash
printf '%s\n%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1"}}}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"manage_tokens","arguments":{"mode":"read","css_mode":"tailwind-v4","scope":"colors"}}}' \
  | ~/.local/bin/devforge-mcp
```

---

### Tool call: `list_patterns`

```bash
printf '%s\n%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1"}}}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_patterns","arguments":{"query":"hero section","framework":"next","limit":5}}}' \
  | ~/.local/bin/devforge-mcp
```

---

### Tool call: `configure_gemini`

```bash
printf '%s\n%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1"}}}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"configure_gemini","arguments":{"api_key":"AIzaXXXXXXXXXXX"}}}' \
  | ~/.local/bin/devforge-mcp
```

---

### Tool call: `store_pattern`

```bash
printf '%s\n%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1"}}}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"store_pattern","arguments":{"name":"Hero Split","description":"Two-column hero: copy on the left, image on the right","category":"layout","framework":"next","css_mode":"tailwind-v4","snippet":"<section class=\"grid grid-cols-2 gap-8\">...</section>"}}}' \
  | ~/.local/bin/devforge-mcp
```

---

## JSON-RPC message format reference

Every message sent to stdin must follow the MCP JSON-RPC 2.0 envelope:

```json
{
  "jsonrpc": "2.0",
  "id": <integer>,
  "method": "<method-name>",
  "params": { ... }
}
```

Key methods:

| Method | Purpose |
|--------|---------|
| `initialize` | Handshake — must be the first message in every session |
| `tools/list` | Returns the full list of registered tools with their schemas |
| `tools/call` | Invokes a tool; `params.name` is the tool name, `params.arguments` is the input |

All tool errors are returned as valid JSON-RPC responses with the error payload
inside `result.content[0].text` as `{"error":"message"}` — the server never
crashes on tool-level errors.

---

## Tips for multi-message sessions

The server keeps state (open DB, imgproc process) for the entire lifetime of the
process. For automated testing, pipe multiple newline-separated JSON-RPC messages
in a single stdin stream:

```bash
cat <<'EOF' | ~/.local/bin/devforge-mcp | jq .
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1"}}}
{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"suggest_color_palettes","arguments":{"mood":"bold and energetic","count":2}}}
EOF
```
