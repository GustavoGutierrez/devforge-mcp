# Esquema de base de datos: ui_patterns.db

SQLite con soporte de FTS5 para búsqueda full-text eficiente.

## Tablas principales

### patterns

```sql
CREATE TABLE patterns (
  id            TEXT PRIMARY KEY,
  name          TEXT NOT NULL,
  category      TEXT NOT NULL, -- landing, dashboard, form, component
  framework     TEXT NOT NULL, -- spa-vite, astro, next, sveltekit, nuxt, vanilla
  css_mode      TEXT NOT NULL, -- tailwind-v4, plain-css
  tags          TEXT,          -- lista separada por comas
  snippet       TEXT NOT NULL, -- código de layout
  css_snippet   TEXT,          -- CSS asociado (Tailwind tokens o CSS moderno)
  description   TEXT,
  created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### architectures

```sql
CREATE TABLE architectures (
  id            TEXT PRIMARY KEY,
  name          TEXT NOT NULL,
  framework     TEXT NOT NULL, -- astro, next, sveltekit, nuxt, spa-vite, vanilla
  css_mode      TEXT NOT NULL,
  description   TEXT NOT NULL,
  decisions     TEXT,          -- ADRs resumidos
  tags          TEXT,
  created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### tokens, audits, assets, palettes

```sql
CREATE TABLE palettes (
  id            TEXT PRIMARY KEY,
  name          TEXT NOT NULL,
  use_case      TEXT,
  mood          TEXT,
  tokens_json   TEXT NOT NULL, -- JSON con colores semánticos
  created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## FTS5

Para búsqueda eficiente sobre patrones y arquitecturas:

```sql
CREATE VIRTUAL TABLE patterns_fts USING fts5(
  name,
  category,
  framework,
  css_mode,
  tags,
  description,
  content='patterns',
  content_rowid='rowid'
);

CREATE VIRTUAL TABLE architectures_fts USING fts5(
  name,
  framework,
  css_mode,
  tags,
  description,
  decisions,
  content='architectures',
  content_rowid='rowid'
);
```

## Opción: libSQL con vector search

Si se requiere búsqueda semántica (por similitud) sobre descripciones,
puede añadirse una tabla con embeddings:

```sql
CREATE TABLE patterns_vectors (
  pattern_id    TEXT PRIMARY KEY,
  embedding     BLOB NOT NULL, -- o tipo vector nativo en libSQL
  FOREIGN KEY(pattern_id) REFERENCES patterns(id)
);
```
