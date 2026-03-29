# PRP-001: DevForge MCP — Full Application Build

**Status**: Draft
**Created**: 2026-03-22
**Module**: `devforge-mcp`

---

## Layer 1 — Purpose and Context

### What and Why

DevForge MCP is a Go-based MCP (Model Context Protocol) server that acts as an acceleration core for the software development cycle. It exposes specialized tools for UI layout analysis, design token management, image processing, favicon generation, color palette suggestion, and AI-assisted UI image generation. It ships alongside a CLI/TUI companion app built with Bubble Tea.

The system is designed to be consumed by AI agents (see `AGENTS.md`) and human developers alike. Agents invoke MCP tools to audit, synthesize, and manage frontend design artifacts. The CLI/TUI provides an interactive interface for the same capabilities.

### Goals

- Provide a stable, well-typed MCP server exposing all design and image tools.
- Persist patterns, architectures, tokens, and palettes in a local libSQL database. Use FTS5 for keyword and tag search. Use vector ANN search (libsql_vector_idx) for semantic similarity queries. Both run fully locally — no Turso cloud dependency.
- Offload heavy image processing to a Rust binary via `internal/dpf`.
- Support optional Gemini API integration for AI image generation.
- Ship a Bubble Tea TUI that wraps all server capabilities in a navigable interface.
- Allow AI agents to consume tools with structured stack metadata (`css_mode`, `framework`).
- Ship comprehensive test coverage for all packages: unit tests for pure functions, integration tests using an in-memory libSQL DB, and stub-based tests for dpf and embedding dependencies.
- Provide an operations guide covering build, installation, MCP client configuration, and system-level auto-start via systemd (Linux) and launchd (macOS).

### Non-Goals

- This PRP does not cover the Rust `bin/dpf` binary itself — it is a pre-built artifact.
- Production deployment, containerization, or cloud hosting are out of scope.
- The sqld server binary is optional — all functionality runs in embedded mode via go-libsql.

---

## Layer 2 — Architecture and File Structure

### Go Module

```
module devforge-mcp

go 1.22
```

### Directory Layout

```
devforge-mcp/
├── cmd/
│   ├── devforge-mcp/       # MCP server entry point
│   │   └── main.go
│   └── devforge/           # CLI/TUI entry point
│       └── main.go
├── internal/
│   ├── db/                  # libSQL setup, migrations, queries, embeddings
│   │   ├── db.go
│   │   ├── db_test.go           # schema migration, WAL, FTS5 smoke test
│   │   ├── schema.go
│   │   ├── embed.go        # embedding client (ollama wrapper, optional — graceful skip if unavailable)
│   │   └── embed_test.go        # embedding client graceful-skip test
│   ├── config/              # Config file read/write
│   │   └── config.go
│   ├── dpf/             # Go bridge to Rust binary
│   │   ├── client.go        # One-shot Client
│   │   ├── stream.go        # StreamClient (persistent process)
│   │   └── jobs.go          # Job types: ResizeJob, OptimizeJob, ConvertJob, FaviconJob, SpriteJob, PlaceholderJob, BatchJob
│   ├── tools/               # One file per MCP tool implementation
│   │   ├── analyze_layout.go
│   │   ├── analyze_layout_test.go
│   │   ├── suggest_layout.go
│   │   ├── manage_tokens.go
│   │   ├── manage_tokens_test.go
│   │   ├── store_pattern.go
│   │   ├── list_patterns.go
│   │   ├── list_patterns_test.go
│   │   ├── generate_ui_image.go
│   │   ├── configure_gemini.go
│   │   ├── optimize_images.go
│   │   ├── generate_favicon.go
│   │   └── suggest_color_palettes.go
│   ├── testutil/
│   │   └── testutil.go          # shared test helpers
│   └── tui/                 # Bubble Tea views
│       ├── model.go         # Root model, navigation state
│       ├── home.go
│       ├── browse_patterns.go
│       ├── browse_architectures.go
│       ├── analyze_layout.go
│       ├── generate_layout.go
│       ├── generate_images.go
│       ├── optimize_images.go
│       ├── generate_favicon.go
│       ├── color_palettes.go
│       └── settings.go
├── db/
│   ├── ui_patterns.db       # SQLite database file (created at runtime)
│   └── seeds/
├── testdata/
│   ├── layouts/
│   │   └── hero.html            # fixture for analyze_layout tests
│   └── images/
│       └── logo.png             # fixture for dpf tests
├── deploy/
│   ├── devforge-mcp.service   # systemd user service unit
│   └── com.devforge.mcp.plist  # macOS launchd plist
├── scripts/
│   ├── link-skills.sh          # symlink .agents/skills → .claude/skills
│   └── install.sh              # copy binaries to ~/.local/bin
├── bin/
│   └── devforge-dpf     # Pre-built Rust binary (chmod +x required)
├── docs/
│   ├── mcp-server-dev-forge.md
│   ├── cli-tui.md
│   └── schema.md
├── PRPs/
│   └── 001-devforge-mcp-app.md
├── AGENTS.md
├── CLAUDE.md
├── README.md
└── go.mod
```

### Key Dependencies

```
github.com/mark3labs/mcp-go          # MCP server SDK for Go
github.com/charmbracelet/bubbletea   # TUI framework
github.com/charmbracelet/bubbles     # TUI components (list, textinput, spinner, viewport)
github.com/charmbracelet/lipgloss    # TUI styling
github.com/tursodatabase/go-libsql          # libSQL embedded driver (CGO) — local vector + FTS5
github.com/ryanskidmore/libsql-vector-go    # float32 vector encoding/decoding for libSQL blobs
github.com/ollama/ollama/api                # local embedding generation via Ollama HTTP API
google.golang.org/genai              # Google Gen AI Go SDK (Gemini)
github.com/google/uuid               # UUID generation for DB primary keys
```

---

## Layer 3 — Detailed Specifications

### 3.1 Configuration

**File**: `internal/config/config.go`

Config path: `~/.config/devforge/config.json`
Override: `DEV_FORGE_CONFIG` environment variable.

```go
type Config struct {
    GeminiAPIKey   string `json:"gemini_api_key"`
    OllamaURL      string `json:"ollama_url"`       // default: http://localhost:11434
    EmbeddingModel string `json:"embedding_model"`  // default: nomic-embed-text (768-dim)
}

func Load() (*Config, error)        // reads and parses config file; returns empty Config if not found
func Save(cfg *Config) error        // writes with os.WriteFile(..., 0600)
func Path() string                  // resolves path from env or default
```

The config directory must be created with `os.MkdirAll` before writing. The file must be written with `0600` permissions.

Hot-reload: after `configure_gemini` saves the key, the MCP server's in-memory `Config` pointer is updated without restart. Use a mutex-protected global or pass a pointer to `MCPServer`.

---

### 3.2 Database

**File**: `internal/db/`
**Driver**: `github.com/tursodatabase/go-libsql` (CGO required — `CGO_ENABLED=1`)
**Database path**: `db/ui_patterns.db` (relative to binary; overridable via config)

#### Connection Setup

```go
import (
    "database/sql"
    "github.com/tursodatabase/go-libsql"
)

connector, err := libsql.NewConnector("file:./db/ui_patterns.db")
if err != nil { return err }
db := sql.OpenDB(connector)

// Required pragmas on startup
db.Exec("PRAGMA journal_mode=WAL;")
db.Exec("PRAGMA foreign_keys=ON;")
```

#### Why FTS5 vs Vector — Decision Guide

| Use case | Mechanism | Reason |
|----------|-----------|--------|
| Search by keyword in name/tags/description | **FTS5** | Exact/prefix match, boolean operators |
| Filter by framework, css_mode, domain | **SQL WHERE** | Structured equality filter |
| "Find patterns similar to this design brief" | **Vector ANN** | Semantic similarity, not keywords |
| "Find architectures for this use case" | **Vector ANN** | Semantic intent matching |
| Filter palettes by mood/use_case keywords | **FTS5** | Exact value matching |

#### Full Schema

```sql
-- ─────────────────────────────────────────────
--  patterns
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS patterns (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    domain      TEXT NOT NULL DEFAULT 'frontend',   -- frontend | backend | fullstack | devops | any
    category    TEXT,                                -- landing | dashboard | form | component | other
    framework   TEXT,                                -- spa-vite | astro | next | sveltekit | nuxt | vanilla
    css_mode    TEXT,                                -- tailwind-v4 | plain-css
    tags        TEXT,                                -- comma-separated
    snippet     TEXT,
    css_snippet TEXT,
    description TEXT,
    embedding   F32_BLOB(768),                       -- nomic-embed-text vector (NULL if ollama unavailable)
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- FTS5: keyword + tag search
CREATE VIRTUAL TABLE IF NOT EXISTS patterns_fts USING fts5(
    name, category, tags, description,
    content='patterns', content_rowid='rowid'
);

-- Vector ANN: semantic similarity (requires non-NULL embedding)
CREATE INDEX IF NOT EXISTS patterns_vec_idx
    ON patterns(libsql_vector_idx(embedding));

-- ─────────────────────────────────────────────
--  architectures
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS architectures (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    domain      TEXT NOT NULL DEFAULT 'fullstack',   -- frontend | backend | fullstack | devops | microservices | any
    framework   TEXT,
    css_mode    TEXT,
    description TEXT,
    decisions   TEXT,
    tags        TEXT,
    embedding   F32_BLOB(768),
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE VIRTUAL TABLE IF NOT EXISTS architectures_fts USING fts5(
    name, description, tags, decisions,
    content='architectures', content_rowid='rowid'
);

CREATE INDEX IF NOT EXISTS architectures_vec_idx
    ON architectures(libsql_vector_idx(embedding));

-- ─────────────────────────────────────────────
--  tokens, audits, assets, palettes
--  (no vector search — keyword + structured filter sufficient)
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS tokens (
    id         TEXT PRIMARY KEY,
    css_mode   TEXT,
    scope      TEXT,
    key        TEXT NOT NULL,
    value      TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS audits (
    id          TEXT PRIMARY KEY,
    page_type   TEXT,
    framework   TEXT,
    css_mode    TEXT,
    report_json TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS assets (
    id          TEXT PRIMARY KEY,
    type        TEXT,
    source_path TEXT,
    output_path TEXT,
    meta_json   TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS palettes (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    use_case    TEXT,
    mood        TEXT,
    tokens_json TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### Embedding Client (internal/db/embed.go)

```go
// EmbeddingClient wraps Ollama's embed endpoint.
// All methods degrade gracefully — if Ollama is unavailable, they return nil, nil.
type EmbeddingClient struct {
    url   string  // e.g. http://localhost:11434
    model string  // e.g. nomic-embed-text
}

func NewEmbeddingClient(ollamaURL, model string) *EmbeddingClient

// Embed returns a 768-dim float32 slice, or nil if Ollama is unreachable.
func (e *EmbeddingClient) Embed(ctx context.Context, text string) ([]float32, error)

// EncodeForDB encodes float32 slice → []byte for libSQL F32_BLOB column.
// Uses github.com/ryanskidmore/libsql-vector-go Encode().
func EncodeForDB(vec []float32) []byte
func DecodeFromDB(blob []byte) []float32
```

Embedding is **optional and non-blocking**: if `Embed()` returns nil, the row is stored with `embedding = NULL` and falls back to FTS5 for search. Vector index automatically excludes NULL rows.

#### Query Patterns

```go
// Keyword search — FTS5
const queryFTS = `
    SELECT p.id, p.name, p.domain, p.category, p.framework, p.css_mode, p.tags, p.description
    FROM patterns_fts f
    JOIN patterns p ON p.rowid = f.rowid
    WHERE patterns_fts MATCH ?
      AND (? = '' OR p.domain = ?)
      AND (? = '' OR p.framework = ?)
      AND (? = '' OR p.css_mode = ?)
    ORDER BY rank
    LIMIT ?`

// Semantic search — Vector ANN
const queryVec = `
    SELECT p.id, p.name, p.domain, p.category, p.framework, p.css_mode, p.tags, p.description,
           results.distance
    FROM vector_top_k('patterns_vec_idx', ?, ?) AS results
    JOIN patterns p ON p.rowid = results.id
    WHERE (? = '' OR p.domain = ?)
      AND (? = '' OR p.framework = ?)
    ORDER BY results.distance`

// Structured filter (no query or vector)
const queryFilter = `
    SELECT id, name, domain, category, framework, css_mode, tags, description
    FROM patterns
    WHERE (? = '' OR domain = ?)
      AND (? = '' OR framework = ?)
      AND (? = '' OR css_mode = ?)
    ORDER BY created_at DESC
    LIMIT ?`
```

The same pattern applies to `architectures`.

#### Seed Data Instructions

Seed files live in `db/seeds/`. Run them once on a fresh database to populate useful initial content. Structure:

**`db/seeds/001_patterns.sql`** — Example layout patterns for each framework/domain combination:
```sql
-- Insert pattern + FTS5 entry together
INSERT INTO patterns (id, name, domain, category, framework, css_mode, tags, snippet, description)
VALUES (
    'pat-001',
    'Hero Section — Tailwind v4 Centered',
    'frontend',
    'landing',
    'astro',
    'tailwind-v4',
    'hero,landing,centered,tailwind',
    '<section class="flex flex-col items-center py-24 gap-6">...</section>',
    'Full-width hero with centered headline, subtext, and CTA button. Tailwind v4 tokens.'
);
INSERT INTO patterns_fts (rowid, name, category, tags, description)
SELECT rowid, name, category, tags, description FROM patterns WHERE id = 'pat-001';
```

Include at minimum:
- 3 frontend patterns per framework (astro, next, sveltekit) × 2 css_mode = 18 entries
- 2 dashboard patterns (frontend)
- 2 backend patterns (API endpoint structure, middleware chain)
- 2 fullstack patterns (auth flow, data fetching pattern)

**`db/seeds/002_architectures.sql`** — Example architectures per domain:
```sql
INSERT INTO architectures (id, name, domain, framework, css_mode, description, decisions, tags)
VALUES (
    'arch-001',
    'Astro Islands Architecture',
    'frontend',
    'astro',
    'tailwind-v4',
    'Static-first Astro site with selective hydration for interactive islands.',
    'Use Astro components for static content. Hydrate only interactive widgets with client:load or client:visible. Shared state via nanostores.',
    'astro,islands,partial-hydration,performance'
);
```

Include at minimum: 2 per domain (frontend, backend, fullstack, devops).

**`db/seeds/003_palettes.sql`** — 6–8 starting palettes covering common use cases:
```sql
INSERT INTO palettes (id, name, use_case, mood, tokens_json) VALUES (
    'pal-001', 'Fintech Calm Blue', 'saas-dashboard', 'serious',
    '{"background":"#0b1220","surface":"#020617","primary":"#22d3ee","primary-soft":"#0f172a","accent":"#facc15","text":"#e2e8f0","muted":"#475569"}'
);
```

#### Startup Embedding Backfill

On server startup, if Ollama is available and any rows have `embedding IS NULL`, optionally backfill embeddings in a background goroutine:

```go
go func() {
    // Select rows with NULL embedding, batch-embed, update
    // Use a semaphore to limit concurrency (max 4 parallel)
    // Log progress — do not block server startup
}()
```

---

### 3.3 MCP Server

**Entry point**: `cmd/devforge-mcp/main.go`

```go
type MCPServer struct {
    db      *sql.DB
    dpf *dpf.StreamClient
    config  *config.Config
    embedder *db.EmbeddingClient   // nil if Ollama unavailable (graceful degradation)
    mu      sync.RWMutex           // protects config for hot-reload
}
```

Startup sequence:
1. Load config.
2. Open libSQL DB and run migrations.
3. Initialize `EmbeddingClient` (test Ollama availability with 1-second timeout; set to nil if unavailable).
4. Initialize `StreamClient` for dpf (`defer sc.Close()`).
5. Register all tools with the MCP SDK.
6. Serve via stdio transport.
7. Launch embedding backfill goroutine if Ollama is available.

The MCP server communicates over stdin/stdout using the `mcp-go` SDK. Each tool is registered with its JSON schema and a handler function.

---

### 3.4 Tool Specifications

#### `analyze_layout`

**Purpose**: Audit HTML/JSX markup for layout issues.

**Input schema**:
```json
{
  "markup":       "string (required) — HTML or JSX string to analyze",
  "stack": {
    "css_mode":   "string (required) — 'tailwind-v4' | 'plain-css'",
    "framework":  "string (required) — 'spa-vite' | 'astro' | 'next' | 'sveltekit' | 'nuxt' | 'vanilla'"
  },
  "page_type":    "string (optional) — e.g. 'landing', 'dashboard', 'form'",
  "device_focus": "string (optional) — 'mobile' | 'desktop' | 'both'"
}
```

**Output**:
```json
{
  "summary":  "string",
  "issues": [
    {
      "severity":    "error | warning | suggestion",
      "category":    "string — e.g. 'spacing', 'typography', 'accessibility'",
      "description": "string",
      "suggestion":  "string — Tailwind class or CSS custom property fix"
    }
  ],
  "score": "number 0-100"
}
```

Implementation: analyze the markup string for common patterns. For `tailwind-v4`, reference `@layer`, `@theme`, and utility class conventions. Store audit in `audits` table.

---

#### `suggest_layout`

**Purpose**: Generate a layout scaffold based on a description.

**Input schema**:
```json
{
  "description":    "string (required)",
  "stack": {
    "css_mode":     "string (required)",
    "framework":    "string (required)"
  },
  "fidelity":       "string (required) — 'wireframe' | 'mid' | 'production'",
  "tokens_profile": "object (optional) — existing token values to incorporate"
}
```

**Output**:
```json
{
  "layout_name": "string",
  "files": [
    { "path": "string", "snippet": "string" }
  ],
  "css_snippets": [
    { "path": "string", "snippet": "string" }
  ],
  "rationale": "string"
}
```

---

#### `manage_tokens`

**Purpose**: Read or update design tokens.

**Input schema**:
```json
{
  "mode":     "string (required) — 'read' | 'plan-update' | 'apply-update'",
  "css_mode": "string (required) — 'tailwind-v4' | 'plain-css'",
  "scope":    "string (required) — 'colors' | 'spacing' | 'typography' | 'all'",
  "proposal": "object (optional) — token key/value pairs to apply (required for apply-update)"
}
```

**Output**:
```json
{
  "current_tokens": "object",
  "diff":           "object — keys changed with old/new values",
  "instructions":   "string — how to apply in Tailwind v4 @theme or CSS :root"
}
```

For `tailwind-v4`: tokens map to `@theme { --color-primary: ...; }` inside a CSS layer.
For `plain-css`: tokens map to `:root { --color-primary: ...; }`.

---

#### `store_pattern`

**Purpose**: Persist a layout pattern to the database.

**Input schema**:
```json
{
  "name":        "string (required)",
  "category":    "string (optional)",
  "framework":   "string (required)",
  "css_mode":    "string (required)",
  "tags":        "string (optional) — comma-separated",
  "snippet":     "string (required) — HTML/JSX snippet",
  "css_snippet": "string (optional)",
  "description": "string (optional)"
}
```

**Output**:
```json
{ "id": "string (UUID)", "name": "string", "created_at": "string" }
```

After INSERT, also INSERT into `patterns_fts`.

---

#### `list_patterns`

**Purpose**: Query stored patterns with optional filters, full-text search, and semantic similarity.

**Input schema**:
```json
{
  "domain":    "string (optional) — 'frontend' | 'backend' | 'fullstack' | 'devops' | 'any'",
  "css_mode":  "string (optional)",
  "framework": "string (optional)",
  "query":     "string (optional) — keyword query (FTS5) OR natural language description (semantic)",
  "mode":      "string (optional) — 'fts' | 'semantic' | 'filter' — default: auto-detect",
  "limit":     "number (optional, default 20)"
}
```

**Output**:
```json
{
  "patterns": [
    {
      "id": "string", "name": "string", "domain": "string", "category": "string",
      "framework": "string", "css_mode": "string", "tags": "string",
      "description": "string", "created_at": "string"
    }
  ],
  "total": "number"
}
```

Mode selection logic:
- `mode=semantic` or `mode` is omitted and Ollama is available: embed `query` and use vector ANN
- `mode=fts`: use FTS5 MATCH
- `mode=filter` or no `query`: SQL WHERE filters only

Apply the same `domain` filter to `list_patterns` and similar spec to architectures browsing.

---

#### `generate_ui_image`

**Purpose**: Generate a UI image via Gemini API.

**Requires**: `gemini_api_key` in config. Return a structured error if not configured.

**Input schema**:
```json
{
  "prompt":      "string (required)",
  "style":       "string (required) — 'wireframe' | 'mockup' | 'illustration'",
  "width":       "number (optional, default 1280)",
  "height":      "number (optional, default 720)",
  "output_path": "string (required) — file path to save the image"
}
```

**Output**:
```json
{
  "path":        "string",
  "width":       "number",
  "height":      "number",
  "prompt_used": "string"
}
```

Implementation: use `google.golang.org/genai` SDK. Model: `gemini-2.0-flash-preview-image-generation` or equivalent image generation model. Send prompt with style prefix. Save base64-decoded image bytes to `output_path`.

---

#### `configure_gemini`

**Purpose**: Save Gemini API key to config file and hot-reload.

**Input schema**:
```json
{ "api_key": "string (required)" }
```

**Output**:
```json
{ "config_path": "string", "status": "saved" }
```

Saves to `config.json` with `0600` permissions. Updates in-memory config on the running server via mutex-protected pointer.

---

#### `optimize_images`

**Purpose**: Optimize and convert images using the Rust dpf binary.

**Input schema**:
```json
{
  "inputs": [
    {
      "path":       "string (required)",
      "max_width":  "number (optional)",
      "max_height": "number (optional)",
      "formats":    "array of string — e.g. ['webp', 'avif', 'jpg']",
      "quality":    "number (optional, 1-100)"
    }
  ],
  "parallelism": "number (optional, default 4)"
}
```

**Output**:
```json
{
  "results": [
    {
      "source_path": "string",
      "outputs": [
        {
          "format":          "string",
          "path":            "string",
          "width":           "number",
          "height":          "number",
          "approx_size_kb":  "number"
        }
      ]
    }
  ]
}
```

Implementation: use `StreamClient.Process(OptimizeJob{...})` from `internal/dpf`. The `StreamClient` is initialized once at server startup and reused.

---

#### `generate_favicon`

**Purpose**: Generate favicon variants from a source image.

**Input schema**:
```json
{
  "source_path":       "string (required)",
  "background_color":  "string (optional, default '#ffffff') — hex color",
  "sizes":             "array of number (optional, default [16,32,48,180,192,512])",
  "formats":           "array of string (optional, default ['ico','png','svg'])"
}
```

**Output**:
```json
{
  "icons": [
    { "size": "number", "format": "string", "path": "string" }
  ],
  "html_snippets": [
    "string — e.g. <link rel='icon' href='/favicon.ico'>"
  ]
}
```

Implementation: use `dpf.FaviconJob{...}` via the `StreamClient`.

---

#### `suggest_color_palettes`

**Purpose**: Generate named color palette proposals.

**Input schema**:
```json
{
  "use_case":       "string (required) — e.g. 'SaaS dashboard', 'marketing site'",
  "brand_keywords": "array of string (optional)",
  "mood":           "string (optional) — e.g. 'calm', 'bold', 'minimal'",
  "count":          "number (optional, default 3)"
}
```

**Output**:
```json
{
  "palettes": [
    {
      "name":        "string",
      "description": "string",
      "tokens": {
        "background":    "string — hex",
        "surface":       "string — hex",
        "primary":       "string — hex",
        "primary-soft":  "string — hex",
        "accent":        "string — hex",
        "text":          "string — hex",
        "muted":         "string — hex"
      }
    }
  ]
}
```

---

### 3.5 dpf Integration

**Package**: `internal/dpf`

**Binary**: `bin/dpf` — pre-built Rust binary. Must be executable (`chmod +x`).

#### Client Types

```go
// One-shot: spawns a new process per call
type Client struct { binPath string }
func NewClient(binPath string) *Client
func (c *Client) Process(job interface{}) ([]byte, error)

// Recommended for servers: persistent subprocess, thread-safe, ~5ms less overhead
type StreamClient struct { /* internal */ }
func NewStreamClient(binPath string) (*StreamClient, error)
func (sc *StreamClient) Process(job interface{}) ([]byte, error)
func (sc *StreamClient) Close() error
```

#### Job Types

```go
type ResizeJob struct {
    InputPath  string `json:"input_path"`
    OutputPath string `json:"output_path"`
    Width      int    `json:"width"`
    Height     int    `json:"height"`
}

type OptimizeJob struct {
    InputPath  string   `json:"input_path"`
    OutputPath string   `json:"output_path"`
    Formats    []string `json:"formats"`
    Quality    int      `json:"quality"`
    MaxWidth   int      `json:"max_width,omitempty"`
    MaxHeight  int      `json:"max_height,omitempty"`
}

type ConvertJob struct {
    InputPath  string `json:"input_path"`
    OutputPath string `json:"output_path"`
    Format     string `json:"format"`
}

type FaviconJob struct {
    SourcePath      string   `json:"source_path"`
    BackgroundColor string   `json:"background_color"`
    Sizes           []int    `json:"sizes"`
    Formats         []string `json:"formats"`
    OutputDir       string   `json:"output_dir"`
}

type BatchJob struct {
    Jobs []interface{} `json:"jobs"`
}
```

**Communication protocol**: The Rust binary reads JSON job requests from stdin (one per line, newline-delimited) and writes JSON responses to stdout. `StreamClient` maintains the subprocess with a goroutine-safe queue.

**Server startup pattern**:
```go
sc, err := dpf.NewStreamClient("./bin/dpf")
if err != nil { log.Fatal(err) }
defer sc.Close()
server := &MCPServer{dpf: sc, ...}
```

---

### 3.6 CLI/TUI Application

**Entry point**: `cmd/devforge/main.go`
**Framework**: Bubble Tea (`charmbracelet/bubbletea`)

#### Navigation Model

```go
type View int
const (
    ViewHome View = iota
    ViewBrowsePatterns
    ViewBrowseArchitectures
    ViewAnalyzeLayout
    ViewGenerateLayout
    ViewGenerateImages
    ViewOptimizeImages
    ViewGenerateFavicon
    ViewColorPalettes
    ViewSettings
)

type Model struct {
    currentView View
    // per-view sub-models
    home                homeModel
    browsePatterns      browsePatternsModel
    browseArchitectures browseArchitecturesModel
    analyzeLayout       analyzeLayoutModel
    generateLayout      generateLayoutModel
    generateImages      generateImagesModel
    optimizeImages      optimizeImagesModel
    generateFavicon     generateFaviconModel
    colorPalettes       colorPalettesModel
    settings            settingsModel
    // shared
    db     *sql.DB
    config *config.Config
    width  int
    height int
}
```

#### Views

**Home** — menu list with items:
1. Browse patterns
2. Browse architectures
3. Analyze layout file
4. Generate layout
5. Generate UI images
6. Optimize images
7. Generate favicon
8. Explore color palettes
9. Settings
10. Quit

Use `bubbles/list` component. Arrow keys navigate, Enter selects, `q` quits.

**Browse patterns / Browse architectures** — shared pattern:
- Filter bar: framework (dropdown/select), css_mode, free-text (FTS5)
- Results list with `bubbles/list`
- Selected item shows detail in right pane or viewport
- `Esc` returns to home

**Analyze layout file**:
- Text inputs: file path, framework (select), css_mode (select)
- On submit: read file, call `analyze_layout` tool logic directly (not via MCP)
- Display structured report with issue list

**Generate layout**:
- Form: description (textarea), framework (select), css_mode (select), fidelity (select)
- On submit: call `suggest_layout` logic
- Show output files with syntax-highlighted snippets
- Option to save as pattern (calls `store_pattern` logic)

**Generate UI images**:
- If `gemini_api_key` not configured: display warning with link to Settings view, offer navigation shortcut
- Else: form with prompt, style (select), width, height, output path
- On submit: call `generate_ui_image` logic
- Show result path and success message

**Optimize images**:
- Multi-path input (comma-separated or one per line)
- Format checkboxes: webp, avif, jpg, png
- Dimension inputs (max_width, max_height), quality slider/input
- On submit: call `optimize_images` logic
- Show results table with sizes

**Generate favicon**:
- File path input, background color input (hex)
- On submit: call `generate_favicon` logic
- Show generated icon list and HTML snippet block

**Explore color palettes**:
- Form: use_case (text), mood (select/text), brand keywords (comma list), count (number)
- On submit: call `suggest_color_palettes` logic
- Display palette cards with color swatches (lipgloss colored blocks)
- Option to save selected palette to `palettes` table

**Settings**:
- Show current config file path
- Gemini API key field: masked input (`bubbles/textinput` with `EchoPassword` mode)
- Status indicator: ✓ (configured) or ✗ (not configured)
- Save button: writes config with `0600`, hot-reloads
- Delete key option: clears key from config
- `Esc` returns to home

#### Stack Auto-Detection

At startup, scan the working directory for:
- `package.json` + presence of `"tailwindcss"` dependency → `tailwind-v4`
- `astro.config.*` → `framework: "astro"`
- `next.config.*` → `framework: "next"`
- `svelte.config.*` → `framework: "sveltekit"`
- `nuxt.config.*` → `framework: "nuxt"`
- `vite.config.*` (no framework match) → `framework: "spa-vite"`
- None of the above → `framework: "vanilla"`, `css_mode: "plain-css"`

Store detected defaults in the root `Model` and pre-fill form fields.

---

### 3.7 Agent Roles (from AGENTS.md)

The server supports the following AI agent personas. Each agent uses a defined subset of tools and must pass `stack.css_mode` and `stack.framework` in all relevant calls.

| Agent | Tools Used |
|-------|-----------|
| `frontend-ux-auditor` | `analyze_layout`, `list_patterns`, `manage_tokens`, `suggest_color_palettes` |
| `layout-synthesizer` | `suggest_layout`, `store_pattern`, `suggest_color_palettes` |
| `design-systemizer` | `manage_tokens`, `list_patterns`, `suggest_color_palettes` |
| `visual-ideation-agent` | `generate_ui_image`, `optimize_images`, `generate_favicon` (requires Gemini key) |
| `asset-optimizer-agent` | `optimize_images`, `generate_favicon` |

Required stack metadata in every call:
```json
{
  "stack": {
    "css_mode":  "tailwind-v4 | plain-css",
    "framework": "spa-vite | astro | next | sveltekit | nuxt | vanilla"
  }
}
```

---

### 3.8 Testing Strategy

Every package must have a corresponding `_test.go` file. Tests run with `go test ./...` and must pass with `CGO_ENABLED=1`.

#### Test categories

| Category | Package | Pattern |
|----------|---------|---------|
| Unit | `internal/config` | Pure function tests, no I/O |
| Unit | `internal/dpf` | Job struct serialization tests |
| Integration | `internal/db` | Real libSQL in-memory or temp file DB |
| Integration | `internal/tools` | Tool handlers with test DB |
| Integration | `internal/tools` | dpf tools with stub binary |
| E2E smoke | `cmd/devforge-mcp` | Start server, invoke one tool via MCP protocol |

#### DB tests (`internal/db/db_test.go`)

Use a temporary file-based libSQL database per test:
```go
func newTestDB(t *testing.T) *sql.DB {
    t.Helper()
    path := filepath.Join(t.TempDir(), "test.db")
    connector, _ := libsql.NewConnector("file:" + path)
    db := sql.OpenDB(connector)
    RunMigrations(db)   // creates all tables + indexes
    t.Cleanup(func() { db.Close() })
    return db
}
```

Must test:
- All 6 tables created by `RunMigrations`
- FTS5 index returns results after INSERT + FTS update trigger
- Vector index created without error (presence check via `sqlite_master`)
- WAL mode pragma returns `"wal"`

#### Tool tests (`internal/tools/*_test.go`)

Each tool test receives a `*sql.DB` from `newTestDB` and a stub `EmbeddingClient` that returns a fixed 768-dim vector.

Required tests per tool:

| Tool | Test cases |
|------|-----------|
| `store_pattern` | valid insert returns UUID; missing required fields return error JSON |
| `list_patterns` | filter by domain; filter by framework; FTS keyword match; semantic mode with stub embedder; empty result returns `[]` not null |
| `manage_tokens` | read mode returns current tokens; apply-update writes to tokens table; plan-update returns diff without writing |
| `analyze_layout` | valid markup returns issues array; stores audit row in DB |
| `suggest_layout` | returns files + css_snippets; tailwind-v4 css_snippet contains `@theme` |
| `configure_gemini` | writes config file with 0600; subsequent Load() returns the key |
| `suggest_color_palettes` | returns palettes with all 7 token keys; count param respected |

For tools using dpf (`optimize_images`, `generate_favicon`): use a stub `StreamClient` that returns a hardcoded JSON response. Do not require the real Rust binary in tests.

For `generate_ui_image`: test that missing API key returns structured error without calling Gemini.

#### Embedding client test (`internal/db/embed_test.go`)

- When Ollama URL is empty string: `Embed()` returns `nil, nil` (graceful skip)
- When Ollama URL is unreachable (use `http://127.0.0.1:1`): `Embed()` returns `nil, nil` within 1.5 seconds (timeout respected)

#### Test helpers (`internal/testutil/testutil.go`)

Create a shared test helpers package:
```go
package testutil

func NewTestDB(t *testing.T) *sql.DB           // temp libSQL DB with migrations applied
func StubEmbedder() *db.EmbeddingClient         // always returns a fixed zero vector
func StubImgprocClient() *dpf.StreamClient  // returns fixture JSON responses
func AssertJSON(t *testing.T, got, want string) // JSON equality ignoring key order
```

#### Running tests

```bash
# All tests
CGO_ENABLED=1 go test ./...

# Verbose with coverage
CGO_ENABLED=1 go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Single package
CGO_ENABLED=1 go test -v ./internal/tools/...

# Skip integration tests (if needed)
CGO_ENABLED=1 go test -short ./...
```

Tests that require a real Ollama instance or Gemini key must be skipped when `OLLAMA_URL` or `GEMINI_API_KEY` environment variables are not set:
```go
if os.Getenv("OLLAMA_URL") == "" {
    t.Skip("OLLAMA_URL not set — skipping live embedding test")
}
```

---

## Layer 4 — Implementation Constraints and Conventions

### Go Conventions

- All errors returned to MCP callers must be structured JSON (`{"error": "message"}`), not panics.
- Use `sync.RWMutex` for any shared state (config, DB connection pool).
- DB access via `database/sql` with prepared statements. No ORM.
- Use `github.com/google/uuid` for generating `id` values (UUID v4).
- All file writes use explicit permission modes: config files `0600`, directories `0755`.
- No global mutable state except the `MCPServer` struct fields (protected by mutex).

### libSQL

- Driver: `github.com/tursodatabase/go-libsql` — requires `CGO_ENABLED=1`.
- Connection: `file:./db/ui_patterns.db` — local embedded mode, no sqld server required.
- The sqld server (`libsql-server`) is optional and only needed for multi-process access.
- Vector index is created with `libsql_vector_idx(embedding)`. Rows with `embedding IS NULL` are excluded from ANN results automatically.
- Use `github.com/ryanskidmore/libsql-vector-go` for encoding `[]float32 ↔ []byte` (little-endian IEEE 754).
- FTS5 is built into libSQL — no extra build tags needed (unlike go-sqlite3).
- Run schema migrations on startup before serving. Migrations are idempotent (`CREATE TABLE IF NOT EXISTS`, `CREATE INDEX IF NOT EXISTS`).
- Enable WAL mode: `PRAGMA journal_mode=WAL;`
- Enable foreign keys: `PRAGMA foreign_keys=ON;`

### Embedding (Ollama)

- Default model: `nomic-embed-text` (768 dimensions). Can be changed via `config.json`.
- Embedding is **optional**: if `ollama_url` is empty or Ollama is not running, all `embedding` columns stay NULL and the system falls back to FTS5 for search.
- Never block server startup waiting for Ollama. Test availability with a 1-second timeout.
- Backfill embeddings for NULL rows in a background goroutine after startup.

### dpf Binary

- Binary path is `./bin/dpf` relative to the server binary.
- If binary is missing or not executable, log a warning but do not crash. Tools that require dpf return a structured error.
- `StreamClient` must be goroutine-safe. Use an internal mutex or channel-based queue.

### Gemini Integration

- If `gemini_api_key` is empty, `generate_ui_image` returns `{"error": "Gemini API key not configured. Use configure_gemini to set it."}`.
- Do not log or expose the API key in any output.
- Use `google.golang.org/genai` SDK, not direct HTTP calls.

### TUI

- All Bubble Tea models implement `tea.Model` interface: `Init() tea.Cmd`, `Update(tea.Msg) (tea.Model, tea.Cmd)`, `View() string`.
- Navigation is managed by the root model; sub-models emit navigation messages.
- Use `lipgloss` for all styling — no raw ANSI escape codes.
- Respect terminal width/height from `tea.WindowSizeMsg`.

---

## Layer 5 — Acceptance Criteria

### Tests

- [ ] `CGO_ENABLED=1 go test ./...` passes with zero failures.
- [ ] `internal/db` tests: all 6 tables + 2 FTS5 + 2 vector indexes created; WAL confirmed.
- [ ] `internal/tools` tests: each tool has at least 2 test cases (happy path + error path).
- [ ] `list_patterns` test covers: domain filter, FTS keyword, semantic mode with stub embedder.
- [ ] `configure_gemini` test verifies 0600 file permissions on written config.
- [ ] Embedding client test: unreachable Ollama returns nil within 1.5 seconds.
- [ ] dpf tests use stub client — no real binary required to run tests.
- [ ] Tests requiring live Ollama or Gemini are skipped when env vars absent.
- [ ] Coverage report generated: `go tool cover -html=coverage.out`.

### Operations

- [ ] `CGO_ENABLED=1 go build -o bin/devforge-mcp ./cmd/devforge-mcp/` succeeds.
- [ ] `CGO_ENABLED=1 go build -o bin/devforge ./cmd/devforge/` succeeds.
- [ ] MCP server responds to `initialize` JSON-RPC call via stdin.
- [ ] MCP server registered as `devforge` in `~/.claude/settings.json`; tools appear in Claude Code.
- [ ] systemd user service file provided at `deploy/devforge-mcp.service`.
- [ ] launchd plist file provided at `deploy/com.devforge.mcp.plist`.
- [ ] `scripts/install.sh` copies binaries to `~/.local/bin/` and sets permissions.

### MCP Server

- [ ] Server starts and registers all 10 tools without error.
- [ ] All tools return valid JSON on both success and error paths.
- [ ] `analyze_layout` stores audit in `audits` table and returns structured issues.
- [ ] `suggest_layout` returns `files` and `css_snippets` arrays appropriate to `css_mode`.
- [ ] `manage_tokens` reads from and writes to the `tokens` table; `apply-update` produces correct Tailwind v4 / plain-CSS output.
- [ ] `store_pattern` inserts into `patterns` and `patterns_fts`; returns UUID.
- [ ] `list_patterns` with `query` uses FTS5; without `query` uses direct filter.
- [ ] `list_patterns` `domain` filter returns only matching domain rows.
- [ ] `list_patterns` `mode=semantic` uses vector ANN when Ollama is available, FTS5 otherwise.
- [ ] `generate_ui_image` returns error when key absent; returns `{path, width, height, prompt_used}` when key is set and Gemini call succeeds.
- [ ] `configure_gemini` writes config with `0600` and updates in-memory config without restart.
- [ ] `optimize_images` delegates to `StreamClient` and returns per-output size/format data.
- [ ] `generate_favicon` returns icon list and valid HTML `<link>` snippets.
- [ ] `suggest_color_palettes` returns palettes with all 7 required token keys.

### Database

- [ ] All 6 tables, 2 FTS5 virtual tables, and 2 vector indexes are created on first run.
- [ ] FTS5 keyword search on `patterns_fts` and `architectures_fts` returns matching rows.
- [ ] Vector ANN query via `vector_top_k('patterns_vec_idx', ?, 10)` returns ranked results when embeddings are present.
- [ ] `list_patterns` with `mode=semantic` embeds the query and uses vector ANN; falls back to FTS5 if Ollama unavailable.
- [ ] `list_patterns` with `domain=backend` filters correctly.
- [ ] Rows inserted without embeddings (Ollama unavailable) are still retrievable via FTS5 and filter queries.
- [ ] WAL mode is enabled.
- [ ] Seed data loads without errors from `db/seeds/*.sql`.
- [ ] Embedding backfill goroutine populates NULL embeddings when Ollama is available post-startup.

### dpf Integration

- [ ] `StreamClient` initializes and keeps process alive across multiple calls.
- [ ] Missing binary produces a non-fatal warning, not a panic.
- [ ] `FaviconJob` and `OptimizeJob` round-trip correctly through the subprocess.

### CLI/TUI

- [ ] Home menu displays all 10 items and navigates to each view.
- [ ] Settings view masks API key input and shows ✓/✗ status.
- [ ] Generate UI images view shows warning and Settings navigation shortcut when key is absent.
- [ ] Browse patterns view supports FTS5 text filter and framework/css_mode filters.
- [ ] Stack auto-detection correctly identifies at least: astro, next, sveltekit, tailwind-v4.
- [ ] Color palettes view renders token colors as lipgloss-styled swatches.
- [ ] All views return to Home on `Esc`.

### Config

- [ ] Config is loaded from `DEV_FORGE_CONFIG` env var path when set.
- [ ] Config file is created with `0600` on first save.
- [ ] Missing config file returns empty `Config{}` without error.

---

## Layer 6 — Installation, Configuration, and Operations

### Prerequisites

| Tool | Version | Required | Notes |
|------|---------|----------|-------|
| Go | ≥ 1.22 | Yes | With CGO support |
| GCC / Clang | any | Yes | Required by go-libsql CGO |
| `bin/dpf` | pre-built | Yes | Rust binary, already in repo |
| Ollama | ≥ 0.3 | No | For vector search. `ollama pull nomic-embed-text` |
| Gemini API key | — | No | Only for `generate_ui_image` |

On Debian/Ubuntu: `sudo apt install gcc`
On macOS: `xcode-select --install`

---

### Build

```bash
# Build both binaries
CGO_ENABLED=1 go build -o bin/devforge-mcp ./cmd/devforge-mcp/
CGO_ENABLED=1 go build -o bin/devforge     ./cmd/devforge/

# Make dpf executable (first time only)
chmod +x bin/dpf

# Run all tests
CGO_ENABLED=1 go test ./...
```

Output binaries: `bin/devforge-mcp` (MCP server) and `bin/devforge` (CLI/TUI).

---

### Install System-Wide (optional)

```bash
# Copy binaries to user PATH
cp bin/devforge-mcp  ~/.local/bin/devforge-mcp
cp bin/devforge      ~/.local/bin/devforge
cp bin/dpf ~/.local/bin/

# Ensure ~/.local/bin is in PATH (add to ~/.bashrc or ~/.zshrc if needed)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
```

The MCP server looks for `devforge-dpf` relative to its own binary. If you install system-wide, all three binaries must be in the same directory, or configure the dpf path explicitly.

---

### Initial Configuration

Config file path: `~/.config/devforge/config.json`

```bash
mkdir -p ~/.config/devforge
cat > ~/.config/devforge/config.json << 'EOF'
{
  "gemini_api_key": "",
  "ollama_url": "http://localhost:11434",
  "embedding_model": "nomic-embed-text"
}
EOF
chmod 600 ~/.config/devforge/config.json
```

Leave `gemini_api_key` empty if you don't have one — all other features work without it.
Leave `ollama_url` as-is if Ollama is running locally on default port.

#### Optional: Ollama setup for vector search

```bash
# Install Ollama (Linux)
curl -fsSL https://ollama.com/install.sh | sh

# Pull the embedding model (274 MB, one-time download)
ollama pull nomic-embed-text

# Verify it works
ollama run nomic-embed-text "test"
```

---

### Integrating as MCP Tool

#### Claude Desktop (`~/.config/claude/claude_desktop_config.json`)

```json
{
  "mcpServers": {
    "devforge": {
      "command": "/home/<user>/.local/bin/devforge-mcp",
      "args": [],
      "env": {
        "DEV_FORGE_CONFIG": "/home/<user>/.config/devforge/config.json"
      }
    }
  }
}
```

#### Claude Code (`.claude/settings.json` in any project)

```json
{
  "mcpServers": {
    "devforge": {
      "command": "devforge-mcp",
      "args": [],
      "type": "stdio"
    }
  }
}
```

Or globally in `~/.claude/settings.json` to make it available in all projects.

After saving the config, restart Claude Desktop / reload Claude Code. The MCP tools will appear in the tools panel.

---

### Running the MCP Server

The MCP server uses **stdio transport** — it is launched on demand by the MCP client (Claude Desktop or Claude Code) and does not need to be kept running manually.

```bash
# Test it works standalone (send a JSON-RPC ping via stdin)
echo '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{}},"id":1}' \
  | devforge-mcp
```

The server starts, responds, and exits when stdin closes. The MCP client manages its lifecycle automatically.

---

### Running the CLI/TUI

```bash
# Interactive TUI (from any project directory)
devforge

# The TUI auto-detects the project stack from the current directory
cd ~/my-next-project && devforge
```

---

### Background / Persistent Mode (advanced)

For use cases where you want the server pre-warmed (e.g. to avoid cold-start latency for embeddings):

#### systemd user service (Linux)

Create `~/.config/systemd/user/devforge-mcp.service`:

```ini
[Unit]
Description=DevForge MCP Server
After=default.target

[Service]
Type=simple
ExecStart=%h/.local/bin/devforge-mcp --mode=http --port=7070
Restart=on-failure
RestartSec=5
Environment=DEV_FORGE_CONFIG=%h/.config/devforge/config.json

[Install]
WantedBy=default.target
```

```bash
# Enable and start
systemctl --user daemon-reload
systemctl --user enable devforge-mcp
systemctl --user start  devforge-mcp

# Check status
systemctl --user status devforge-mcp

# Stop
systemctl --user stop devforge-mcp

# Restart
systemctl --user restart devforge-mcp

# View logs
journalctl --user -u devforge-mcp -f

# Enable auto-start on login
loginctl enable-linger $USER
```

#### launchd plist (macOS)

Create `~/Library/LaunchAgents/com.devforge.mcp.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>             <string>com.devforge.mcp</string>
  <key>ProgramArguments</key>
  <array>
    <string>/Users/<user>/.local/bin/devforge-mcp</string>
    <string>--mode=http</string>
    <string>--port=7070</string>
  </array>
  <key>EnvironmentVariables</key>
  <dict>
    <key>DEV_FORGE_CONFIG</key>
    <string>/Users/<user>/.config/devforge/config.json</string>
  </dict>
  <key>RunAtLoad</key>         <true/>
  <key>KeepAlive</key>         <true/>
  <key>StandardOutPath</key>   <string>/tmp/devforge-mcp.log</string>
  <key>StandardErrorPath</key> <string>/tmp/devforge-mcp.err</string>
</dict>
</plist>
```

```bash
# Load (starts immediately + auto-starts on login)
launchctl load ~/Library/LaunchAgents/com.devforge.mcp.plist

# Stop
launchctl unload ~/Library/LaunchAgents/com.devforge.mcp.plist

# Restart
launchctl unload ~/Library/LaunchAgents/com.devforge.mcp.plist
launchctl load   ~/Library/LaunchAgents/com.devforge.mcp.plist
```

> **Note**: `--mode=http --port=7070` is for a future HTTP transport mode. The current stdio transport is managed by the MCP client. Add this flag only once HTTP mode is implemented.

---

### Stopping and Restarting

| Scenario | Command |
|----------|---------|
| MCP client manages process | Nothing needed — client stops it automatically |
| Running via systemd | `systemctl --user stop devforge-mcp` |
| Running via launchd | `launchctl unload ~/Library/LaunchAgents/com.devforge.mcp.plist` |
| Running manually in terminal | `Ctrl+C` |
| Full restart after binary update | systemd: `systemctl --user restart devforge-mcp` / launchd: unload + load |

---

### Updating the Binary

```bash
# Rebuild
CGO_ENABLED=1 go build -o bin/devforge-mcp ./cmd/devforge-mcp/

# Copy to system path
cp bin/devforge-mcp ~/.local/bin/devforge-mcp

# Restart service if running
systemctl --user restart devforge-mcp   # Linux
# or
launchctl unload ~/Library/LaunchAgents/com.devforge.mcp.plist
launchctl load   ~/Library/LaunchAgents/com.devforge.mcp.plist  # macOS
```
