# AGENTS.md — Dev Forge MCP

Single source of truth for AI agents working in this repository. Covers build, test, conventions, constraints, tool inventory, and agent roles.

---

## Project Overview

DevForge MCP is a Go MCP server that acts as a transversal intelligence layer across the software development lifecycle. It exposes tools for UI/design, image processing, and design-system management through the MCP stdio transport protocol, and ships a companion CLI/TUI built with Bubble Tea.

Current tool surface covers UI and design for:
- **CSS modes**: Tailwind CSS v4+ (no `tailwind.config.js`; tokens in CSS), plain CSS with custom properties.
- **Frameworks**: SPA Vite 8, Astro, Next.js, SvelteKit, Nuxt.js, vanilla.

Go module: `dev-forge-mcp`

---

## File Structure

```
cmd/dev-forge-mcp/        MCP server entry point (stdio transport)
cmd/dev-forge/            CLI/TUI entry point (Bubble Tea)
internal/config/          Config read/write (~/.config/dev-forge/config.json)
internal/db/              SQLite setup, schema migrations, queries
internal/imgproc/         Go bridge to Rust image-processing binary
internal/tools/           One file per MCP tool implementation
internal/tui/             Bubble Tea views
db/ui_patterns.db         SQLite database (created at runtime)
bin/devforge-imgproc      Pre-built Rust binary (must be chmod +x)
docs/                     Tool and schema documentation
PRPs/                     Product Requirement Prompts — read before implementing a feature
scripts/link-skills.sh    Creates .claude/skills/ → .agents/skills/ symlinks
```

Skills live in `.agents/skills/` and are symlinked to `.claude/skills/`. Run `./scripts/link-skills.sh` to (re)create symlinks.

---

## Build & Run

SQLite requires CGO. Always build with `CGO_ENABLED=1`.

```bash
# Build MCP server
CGO_ENABLED=1 go build ./cmd/dev-forge-mcp/

# Build CLI/TUI
CGO_ENABLED=1 go build ./cmd/dev-forge/

# Build everything
CGO_ENABLED=1 go build ./...

# Run MCP server (stdio transport)
./dev-forge-mcp

# Run CLI/TUI
./dev-forge
```

Attach the MCP server to an MCP client by adding it to the `mcpServers` section of your client config (transport: `stdio`, command: path to `./dev-forge-mcp`).

### Rust image-processing binary

The pre-built binary lives at `bin/devforge-imgproc`. Ensure it is executable:

```bash
chmod +x bin/devforge-imgproc
```

To recompile from source (requires Rust toolchain):

```bash
# Dynamic binary
make build-rust

# Static binary (no system deps)
make build-rust-static
# Output: rust-imgproc/target/x86_64-unknown-linux-musl/release/devforge-imgproc
# Copy to bin/devforge-imgproc manually
```

---

## Testing

```bash
CGO_ENABLED=1 go test ./...
```

FTS5 must be available in the SQLite build for tests to pass (see Constraints below).

---

## Development Conventions

- **Errors**: All MCP tool errors return structured JSON `{"error": "message"}`. Never panic in a tool handler.
- **DB access**: `database/sql` with prepared statements. No ORM. WAL mode enabled at startup.
- **Config**: Read/write `~/.config/dev-forge/config.json` with `0600` permissions. Override path with `DEV_FORGE_CONFIG` env var. Config is shared between the MCP server and the CLI.
- **Concurrency**: Use `sync.RWMutex` for shared state. `StreamClient` for imgproc is goroutine-safe.
- **TUI**: All Bubble Tea models implement `tea.Model`. Use `lipgloss` for styling. No raw ANSI escape codes.
- **IDs**: UUID v4 (`github.com/google/uuid`) for all database primary keys.
- **New tools**: Add one file per tool under `internal/tools/`. Register the tool in the MCP server entry point.
- **PRPs**: Read the relevant PRP in `PRPs/` before implementing any new feature.

---

## Key Constraints & Gotchas

- **`CGO_ENABLED=1` is required.** Plain `go build` will fail because of `go-sqlite3`.
- **FTS5 must be available.** The server exits at startup with a clear error if the SQLite build lacks FTS5. Use the `sqlite_fts5` build tag with `go-sqlite3` if your system SQLite does not include it.
- **`bin/devforge-imgproc` must exist and be executable.** Tools that require it (`optimize_images`, `generate_favicon`) return a structured error if it is missing — the server does not crash.
- **`generate_ui_image` requires `gemini_api_key`** to be present in config. Returns a clear error if absent.
- **`configure_gemini` hot-reloads** the Gemini API key into the running server without restart.
- **MCP transport is stdio only.** The server reads from stdin and writes to stdout; do not add HTTP transport unless explicitly planned.
- **Binary path for imgproc** must be relative to the process CWD (default: `./bin/devforge-imgproc`).
- **`StreamClient` lifecycle**: initialize once at server startup, close with `defer sc.Close()`. It keeps the Rust process alive across requests (~5 ms overhead saved per operation vs. one-shot client).

---

## Stack Metadata (Tool Call Convention)

Every tool call that involves layout or design must include:

```json
{
  "stack": {
    "css_mode": "tailwind-v4" | "plain-css",
    "framework": "spa-vite" | "astro" | "next" | "sveltekit" | "nuxt" | "vanilla"
  }
}
```

Agents must adapt layout and token suggestions to the declared stack.

### Tailwind v4 specifics
- Import via `@import "tailwindcss";` in a single CSS file. No `tailwind.config.js`.
- Design tokens are CSS-native: `@property`, `:root`, `@layer theme`, `@layer base`.
- Do not generate `tailwind.config.js` output for v4 projects.

### Plain CSS specifics
- Tokens as CSS custom properties (`--color-primary`, `--spacing-md`, etc.).
- Output HTML/JSX/Svelte with class names + custom properties, no Tailwind utility classes.

---

## MCP Tools Reference

| Tool | Description | Requires |
|------|-------------|----------|
| `analyze_layout` | Audit an existing layout for UX issues and token/pattern consistency | — |
| `suggest_layout` | Generate a new layout variant for the given stack | — |
| `manage_tokens` | Read or write design tokens (colors, spacing, typography) | — |
| `store_pattern` | Save a UI pattern to the SQLite database | — |
| `list_patterns` | Query stored patterns with FTS5 full-text search | FTS5 |
| `suggest_color_palettes` | Generate cohesive color palette options | — |
| `generate_ui_image` | Generate a UI image via Gemini Vision | `gemini_api_key` in config |
| `optimize_images` | Compress PNG/JPEG and generate WebP variants | `bin/devforge-imgproc` |
| `generate_favicon` | Generate a full favicon pack from SVG/PNG | `bin/devforge-imgproc` |
| `configure_gemini` | Set or update Gemini API key (hot-reload, no restart) | — |

---

## Database Schema (SQLite)

Tables: `patterns`, `architectures`, `tokens`, `audits`, `assets`, `palettes`.
FTS5 virtual tables provide full-text search over pattern and architecture descriptions.
Database file: `db/ui_patterns.db` (created at runtime if absent).

Optional future path: migrate to libSQL for vector/semantic search while keeping 100% local execution with standard SQLite as the default.

---

## imgproc Bridge — Job Types

Use `StreamClient` in the MCP server (goroutine-safe, single persistent Rust process):

```go
sc, err := imgproc.NewStreamClient("./bin/devforge-imgproc")
defer sc.Close()

result, err := sc.Execute(&imgproc.ResizeJob{
    Operation: "resize",
    Input:     "uploads/photo.jpg",
    OutputDir: "public/img",
    Widths:    []uint32{320, 640, 1280},
})
```

| Job type | `operation` value | Use case |
|----------|------------------|----------|
| `ResizeJob` | `"resize"` | Responsive image variants |
| `OptimizeJob` | `"optimize"` | Compress PNG/JPEG + generate WebP |
| `ConvertJob` | `"convert"` | Format conversion (SVG→WebP, PNG→AVIF, etc.) |
| `FaviconJob` | `"favicon"` | Full favicon pack from SVG/PNG |
| `SpriteJob` | `"sprite"` | Sprite sheet + CSS |
| `PlaceholderJob` | `"placeholder"` | LQIP, dominant color, CSS gradient |
| `BatchJob` | `"batch"` | Multiple operations in parallel |

Import path: `"dev-forge-mcp/internal/imgproc"`

---

## Agent Roles

### 1. frontend-ux-auditor

Audits existing layouts (Tailwind v4 or plain CSS) and proposes UI/UX and component-architecture improvements.

**Uses:** `analyze_layout`, `list_patterns`, `manage_tokens` (read), `suggest_color_palettes` (when the issue is chromatic/branding).

**Steps:**
1. Detect whether the layout uses Tailwind v4 or plain CSS (check for `@import "tailwindcss"` or utility classes).
2. Analyze visual hierarchy, accessibility, token coherence, and pattern consistency.
3. Propose specific, actionable changes. Do not impose Tailwind on a plain-CSS project.

---

### 2. layout-synthesizer

Generates new layout variants for any supported stack.

**Uses:** `suggest_layout`, `store_pattern`, `suggest_color_palettes` (for initial color sets).

**Steps:**
1. Receive a screen description and the stack (`framework` + `css_mode`).
2. Generate output:
   - **Tailwind v4**: markup with utility classes and token references.
   - **Plain CSS**: HTML/JSX/Svelte with class names + custom properties.
3. Optionally propose file organization for Astro/Next/SvelteKit/Nuxt projects.

---

### 3. design-systemizer

Unifies design-token management for both Tailwind v4 and plain CSS.

**Uses:** `manage_tokens`, `list_patterns`, `suggest_color_palettes`.

**Steps:**
1. Read existing tokens:
   - **Tailwind v4**: CSS token layers (`@property`, `:root`, `@layer theme`).
   - **Plain CSS**: custom properties (`--color-primary`, etc.).
2. Propose coherent color, spacing, and typography scales.
3. Never generate `tailwind.config.js` output for v4 projects (legacy format).

---

### 4. visual-ideation-agent

Generates visual ideas and translates them into production-ready components for any supported stack.

**Uses:** `generate_ui_image`, `optimize_images`, `generate_favicon`.

**Prerequisite:** `gemini_api_key` must be set via `configure_gemini`. If absent, skip `generate_ui_image` and work with optimization and favicon tools only.

---

### 5. asset-optimizer-agent

Optimizes images for the web, including favicon packs.

**Uses:** `optimize_images`, `generate_favicon`.

**Steps:**
1. Minimize file size while preserving adequate quality.
2. Generate WebP/AVIF variants for modern browsers.
3. Generate favicons without distorting the source image — use letterboxing or intelligent cropping, never stretching.
