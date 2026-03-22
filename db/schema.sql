-- db/schema.sql — dev-forge-mcp libSQL schema
-- Extracted from internal/db/schema.go
--
-- NOTE: This file uses libSQL-specific syntax:
--   F32_BLOB(768)            — 768-dimensional float32 vector column
--   libsql_vector_idx(col)   — vector similarity index
--
-- Apply via the Go app (preferred — db.Open() runs migrations):
--   make db-init
--
-- Or via libsql-shell if available:
--   libsql-shell db/ui_patterns.db < db/schema.sql
--
-- Standard sqlite3 CLI will fail on F32_BLOB and libsql_vector_idx statements.

PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;

CREATE TABLE IF NOT EXISTS patterns (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    domain      TEXT NOT NULL DEFAULT 'frontend',
    category    TEXT,
    framework   TEXT,
    css_mode    TEXT,
    tags        TEXT,
    snippet     TEXT,
    css_snippet TEXT,
    description TEXT,
    embedding   F32_BLOB(768),
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE VIRTUAL TABLE IF NOT EXISTS patterns_fts USING fts5(
    name, category, tags, description,
    content='patterns', content_rowid='rowid'
);

CREATE INDEX IF NOT EXISTS patterns_vec_idx
    ON patterns(libsql_vector_idx(embedding));

CREATE TABLE IF NOT EXISTS architectures (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    domain      TEXT NOT NULL DEFAULT 'fullstack',
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
