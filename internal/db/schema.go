package db

// pragmas are executed with QueryRow (returns a result row).
var pragmas = []string{
	"PRAGMA journal_mode=WAL",
	"PRAGMA foreign_keys=ON",
}

// schema contains DDL statements executed with Exec (no return rows).
const schema = `
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
`
