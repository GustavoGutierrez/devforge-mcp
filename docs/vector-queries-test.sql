-- =============================================================================
-- vector-queries-test.sql
-- Manual validation queries for libSQL vector/embedding functionality
-- in dev-forge-mcp (db/ui_patterns.db)
--
-- Run these queries in a libSQL-compatible SQL GUI (e.g. Outerbase, Turso shell,
-- or any tool with libSQL support).
--
-- IMPORTANT: Standard sqlite3 CLI does NOT support F32_BLOB or vector_top_k().
--            You must use a libSQL client (go-libsql, Turso shell, etc.)
--
-- Prerequisites:
--   1. Run the app at least once so migrations are applied:   make db-init
--   2. Load seed data:                                        make db-seed
--   3. Embed patterns (requires Ollama running nomic-embed-text):
--      ./dev-forge-mcp --embed-seeds   (or trigger via MCP store_pattern)
--
-- Embedding model: nomic-embed-text (768-dimensional F32_BLOB)
-- Index name:      patterns_vec_idx  (on patterns.embedding)
--                  architectures_vec_idx (on architectures.embedding)
-- =============================================================================


-- ## SECTION 1: Schema Inspection
-- Verify that all expected tables, indexes, and FTS virtual tables exist,
-- and understand the distribution of embedded vs. non-embedded rows.
-- -----------------------------------------------------------------------------

-- 1.1  List all tables (including virtual/FTS tables)
--      Expected: patterns, patterns_fts, architectures, architectures_fts,
--                tokens, audits, assets, palettes
SELECT name, type
FROM sqlite_master
WHERE type IN ('table', 'index', 'shadow')
ORDER BY type, name;


-- 1.2  Confirm the vector indexes exist
--      Expected: two rows — patterns_vec_idx and architectures_vec_idx
SELECT name, tbl_name, sql
FROM sqlite_master
WHERE name LIKE '%vec_idx%';


-- 1.3  Row counts for every table
--      Expected after seeding: patterns ≥ 18, architectures ≥ 8, palettes ≥ 8
SELECT 'patterns'      AS tbl, COUNT(*) AS rows FROM patterns
UNION ALL
SELECT 'architectures' AS tbl, COUNT(*) AS rows FROM architectures
UNION ALL
SELECT 'palettes'      AS tbl, COUNT(*) AS rows FROM palettes
UNION ALL
SELECT 'tokens'        AS tbl, COUNT(*) AS rows FROM tokens
UNION ALL
SELECT 'audits'        AS tbl, COUNT(*) AS rows FROM audits
UNION ALL
SELECT 'assets'        AS tbl, COUNT(*) AS rows FROM assets;


-- 1.4  Embedding coverage in patterns
--      Expected after embedding: all rows should have embedding IS NOT NULL
SELECT
    COUNT(*)                                          AS total,
    COUNT(embedding)                                  AS with_embedding,
    COUNT(*) - COUNT(embedding)                       AS missing_embedding,
    ROUND(COUNT(embedding) * 100.0 / COUNT(*), 1)    AS pct_embedded
FROM patterns;


-- 1.5  Embedding coverage in architectures
SELECT
    COUNT(*)                                          AS total,
    COUNT(embedding)                                  AS with_embedding,
    COUNT(*) - COUNT(embedding)                       AS missing_embedding,
    ROUND(COUNT(embedding) * 100.0 / COUNT(*), 1)    AS pct_embedded
FROM architectures;


-- 1.6  List all patterns with their embedding status
--      Useful for spotting which specific rows still need embedding
SELECT
    id,
    name,
    domain,
    category,
    framework,
    CASE WHEN embedding IS NULL THEN 'MISSING' ELSE 'OK' END AS emb_status
FROM patterns
ORDER BY emb_status DESC, id;


-- =============================================================================
-- ## SECTION 2: Basic Vector Queries
-- Inspect raw embedding blobs and validate F32_BLOB encoding.
-- The embedding is stored as little-endian IEEE 754 float32 bytes (768 dims).
-- Each dimension occupies 4 bytes, so a full embedding = 3072 bytes.
-- -----------------------------------------------------------------------------

-- 2.1  Show the raw embedding blob for a known pattern (hex preview)
--      Expected: a non-NULL blob, displayed as hex by the GUI
--      Length should be exactly 3072 bytes (768 * 4)
SELECT
    id,
    name,
    embedding,
    LENGTH(embedding) AS blob_bytes
FROM patterns
WHERE id = 'pat-001'   -- Hero Section — Tailwind v4 Centered
  AND embedding IS NOT NULL;


-- 2.2  Verify blob length for all embedded patterns
--      Expected: every row returns blob_bytes = 3072
--      Rows with blob_bytes != 3072 indicate a corrupted or truncated embedding
SELECT
    id,
    name,
    LENGTH(embedding) AS blob_bytes,
    CASE
        WHEN LENGTH(embedding) = 3072 THEN 'OK (768-dim)'
        WHEN LENGTH(embedding) IS NULL THEN 'NULL — not yet embedded'
        ELSE 'WRONG LENGTH — check encoder'
    END AS validation
FROM patterns
ORDER BY validation;


-- 2.3  Use libSQL vector() to decode and inspect the first few dimensions
--      vector_extract() converts the F32_BLOB into a JSON array of floats
--      Expected: a JSON array with 768 float values
SELECT
    id,
    name,
    vector_extract(embedding) AS embedding_json
FROM patterns
WHERE id = 'pat-001';


-- 2.4  Check dimension count using JSON_ARRAY_LENGTH on the extracted vector
--      Expected: 768 for every row (nomic-embed-text output dimension)
SELECT
    id,
    name,
    JSON_ARRAY_LENGTH(vector_extract(embedding)) AS dimensions
FROM patterns
WHERE embedding IS NOT NULL
LIMIT 10;


-- 2.5  Validate F32_BLOB re-encoding round-trip via vector()
--      vector(embedding) normalises the blob; if this succeeds without error
--      the encoding is valid for ANN queries.
--      Expected: same blob length, no errors
SELECT
    id,
    name,
    LENGTH(vector(embedding))  AS normalised_blob_bytes
FROM patterns
WHERE embedding IS NOT NULL
LIMIT 5;


-- =============================================================================
-- ## SECTION 3: ANN Vector Search (the main validation)
-- Uses libSQL's vector_top_k() for approximate nearest-neighbour search.
-- Syntax:
--   SELECT rowid, distance
--   FROM vector_top_k('index_name', vector(query_blob), k)
--
-- Self-similarity trick: query with the embedding of a known seeded pattern.
-- The pattern must appear as the #1 result with distance ≈ 0.0.
--
-- NOTE: vector_top_k() returns only one column: `id` (the rowid of the match).
--       It does NOT return a `distance` column. To obtain the cosine distance,
--       compute it explicitly using:
--         vector_distance_cos(p.embedding, (SELECT embedding FROM patterns WHERE id = 'XXX'))
--       where 'XXX' is the same pattern ID used as the query vector.
--       For architectures queries, use a.embedding instead of p.embedding.
-- -----------------------------------------------------------------------------

-- 3.1  Self-similarity check — Hero Section (pat-001)
--      Query the index with pat-001's own embedding.
--      Expected: pat-001 is result #1 with distance ≈ 0.0
--      Other hero/landing patterns should appear nearby.
SELECT
    p.id,
    p.name,
    p.category,
    p.framework,
    vector_distance_cos(p.embedding, (SELECT embedding FROM patterns WHERE id = 'pat-001')) AS distance
FROM vector_top_k(
    'patterns_vec_idx',
    (SELECT embedding FROM patterns WHERE id = 'pat-001'),
    10
) AS results
JOIN patterns p ON p.rowid = results.id
ORDER BY distance;


-- 3.2  Self-similarity check — Dashboard Stats (pat-006, Next.js)
--      Expected: pat-006 is result #1; other dashboard patterns rank highly
--      (pat-008 Sidebar Nav, pat-010 Dashboard Layout, pat-016 Analytics)
SELECT
    p.id,
    p.name,
    p.category,
    p.framework,
    p.domain,
    vector_distance_cos(p.embedding, (SELECT embedding FROM patterns WHERE id = 'pat-006')) AS distance
FROM vector_top_k(
    'patterns_vec_idx',
    (SELECT embedding FROM patterns WHERE id = 'pat-006'),
    10
) AS results
JOIN patterns p ON p.rowid = results.id
ORDER BY distance;


-- 3.3  Cross-framework similarity — SvelteKit Hero vs all patterns
--      Query with pat-009 (Hero — SvelteKit Tailwind v4).
--      Expected: other hero/landing patterns rank high regardless of framework,
--      demonstrating semantic similarity across frameworks.
SELECT
    p.id,
    p.name,
    p.category,
    p.framework,
    vector_distance_cos(p.embedding, (SELECT embedding FROM patterns WHERE id = 'pat-009')) AS distance
FROM vector_top_k(
    'patterns_vec_idx',
    (SELECT embedding FROM patterns WHERE id = 'pat-009'),
    8
) AS results
JOIN patterns p ON p.rowid = results.id
ORDER BY distance;


-- 3.4  ANN search filtered by framework
--      Find the most similar patterns to the Login Form (pat-007)
--      but restrict results to Next.js patterns only.
--      Expected: pat-007 at distance ≈ 0.0, other Next.js patterns follow.
SELECT
    p.id,
    p.name,
    p.category,
    p.css_mode,
    vector_distance_cos(p.embedding, (SELECT embedding FROM patterns WHERE id = 'pat-007')) AS distance
FROM vector_top_k(
    'patterns_vec_idx',
    (SELECT embedding FROM patterns WHERE id = 'pat-007'),
    10
) AS results
JOIN patterns p ON p.rowid = results.id
WHERE p.framework = 'next'
ORDER BY distance;


-- 3.5  ANN search filtered by domain = 'backend'
--      Use pat-012 (REST API Endpoint Structure) as query vector.
--      Expected: both backend patterns (pat-012, pat-013) at the top.
SELECT
    p.id,
    p.name,
    p.domain,
    p.category,
    vector_distance_cos(p.embedding, (SELECT embedding FROM patterns WHERE id = 'pat-012')) AS distance
FROM vector_top_k(
    'patterns_vec_idx',
    (SELECT embedding FROM patterns WHERE id = 'pat-012'),
    10
) AS results
JOIN patterns p ON p.rowid = results.id
WHERE p.domain = 'backend'
ORDER BY distance;


-- 3.6  ANN search filtered by category = 'landing'
--      Use pat-004 (Hero Section — Plain CSS) as query vector.
--      Expected: all landing/hero patterns cluster at low distance.
SELECT
    p.id,
    p.name,
    p.framework,
    p.css_mode,
    vector_distance_cos(p.embedding, (SELECT embedding FROM patterns WHERE id = 'pat-004')) AS distance
FROM vector_top_k(
    'patterns_vec_idx',
    (SELECT embedding FROM patterns WHERE id = 'pat-004'),
    10
) AS results
JOIN patterns p ON p.rowid = results.id
WHERE p.category = 'landing'
ORDER BY distance;


-- 3.7  Self-similarity check on architectures table
--      Use arch-001 (Astro Islands Architecture) as query vector.
--      Expected: arch-001 at distance ≈ 0.0; other frontend architectures nearby.
SELECT
    a.id,
    a.name,
    a.domain,
    a.framework,
    vector_distance_cos(a.embedding, (SELECT embedding FROM architectures WHERE id = 'arch-001')) AS distance
FROM vector_top_k(
    'architectures_vec_idx',
    (SELECT embedding FROM architectures WHERE id = 'arch-001'),
    8
) AS results
JOIN architectures a ON a.rowid = results.id
ORDER BY distance;


-- 3.8  ANN on architectures filtered by domain = 'backend'
--      Use arch-003 (Hexagonal Architecture) as query vector.
--      Expected: arch-003 and arch-004 (Event-Driven Microservices) at top.
SELECT
    a.id,
    a.name,
    a.domain,
    vector_distance_cos(a.embedding, (SELECT embedding FROM architectures WHERE id = 'arch-003')) AS distance
FROM vector_top_k(
    'architectures_vec_idx',
    (SELECT embedding FROM architectures WHERE id = 'arch-003'),
    8
) AS results
JOIN architectures a ON a.rowid = results.id
WHERE a.domain = 'backend'
ORDER BY distance;


-- =============================================================================
-- ## SECTION 4: FTS5 Full-Text Search
-- patterns_fts indexes: name, category, tags, description
-- architectures_fts indexes: name, description, tags, decisions
--
-- FTS5 'rank' column sorts by BM25 relevance (lower = more relevant in libSQL).
-- -----------------------------------------------------------------------------

-- 4.1  Basic FTS5 match on patterns — search for 'hero'
--      Expected: all hero/landing patterns returned, ranked by relevance
SELECT
    p.id,
    p.name,
    p.category,
    p.framework,
    f.rank
FROM patterns_fts f
JOIN patterns p ON p.rowid = f.rowid
WHERE patterns_fts MATCH 'hero'
ORDER BY f.rank
LIMIT 10;


-- 4.2  FTS5 match on multiple terms — 'dashboard tailwind'
--      Expected: dashboard patterns using Tailwind CSS rank highest
SELECT
    p.id,
    p.name,
    p.category,
    p.framework,
    p.css_mode,
    f.rank
FROM patterns_fts f
JOIN patterns p ON p.rowid = f.rowid
WHERE patterns_fts MATCH 'dashboard tailwind'
ORDER BY f.rank
LIMIT 10;


-- 4.3  FTS5 match with snippet() for highlighted context
--      snippet(table, col_index, start_tag, end_tag, ellipsis, num_tokens)
--      Expected: description column snippets with matched terms highlighted
SELECT
    p.id,
    p.name,
    snippet(patterns_fts, 3, '>>>', '<<<', '...', 12) AS description_snippet
FROM patterns_fts f
JOIN patterns p ON p.rowid = f.rowid
WHERE patterns_fts MATCH 'sticky navigation'
ORDER BY f.rank
LIMIT 5;


-- 4.4  FTS5 phrase search — exact multi-word phrase
--      Expected: patterns that contain 'custom properties' in any indexed column
SELECT
    p.id,
    p.name,
    p.category,
    f.rank
FROM patterns_fts f
JOIN patterns p ON p.rowid = f.rowid
WHERE patterns_fts MATCH '"custom properties"'
ORDER BY f.rank
LIMIT 10;


-- 4.5  FTS5 with column filter — search only in the tags column
--      Expected: patterns that have 'auth' in their tags field
SELECT
    p.id,
    p.name,
    p.tags,
    f.rank
FROM patterns_fts f
JOIN patterns p ON p.rowid = f.rowid
WHERE patterns_fts MATCH 'tags:auth'
ORDER BY f.rank
LIMIT 10;


-- 4.6  FTS5 match combined with SQL filter on framework
--      Expected: only SvelteKit patterns matching 'data' in FTS
SELECT
    p.id,
    p.name,
    p.framework,
    f.rank
FROM patterns_fts f
JOIN patterns p ON p.rowid = f.rowid
WHERE patterns_fts MATCH 'data'
  AND p.framework = 'sveltekit'
ORDER BY f.rank
LIMIT 10;


-- 4.7  FTS5 match on architectures table — search for 'microservices'
--      Expected: arch-004 (Event-Driven Microservices) at top
SELECT
    a.id,
    a.name,
    a.domain,
    f.rank
FROM architectures_fts f
JOIN architectures a ON a.rowid = f.rowid
WHERE architectures_fts MATCH 'microservices'
ORDER BY f.rank
LIMIT 5;


-- 4.8  FTS5 highlight() on architectures decisions column
--      highlight(table, col_index, start_tag, end_tag)
--      col indexes: 0=name, 1=description, 2=tags, 3=decisions
SELECT
    a.id,
    a.name,
    highlight(architectures_fts, 3, '[', ']') AS decisions_highlighted
FROM architectures_fts f
JOIN architectures a ON a.rowid = f.rowid
WHERE architectures_fts MATCH 'server components'
ORDER BY f.rank
LIMIT 5;


-- =============================================================================
-- ## SECTION 5: Hybrid Search
-- Combine FTS5 keyword relevance with ANN vector similarity scores.
-- Approach: run both independently, UNION results, deduplicate, and rank.
--
-- Score normalization note:
--   - FTS rank is negative BM25 (more negative = more relevant); invert to get
--     a positive relevance score.
--   - ANN distance is a non-negative float (0.0 = identical); invert to get
--     a similarity score.
--   - Combine with weighted sum: hybrid_score = α * fts_score + β * vec_score
-- -----------------------------------------------------------------------------

-- 5.1  Hybrid search: FTS UNION ANN, deduplicated
--      Query: 'sidebar navigation'
--      ANN probe: use pat-003 (Nav Header — Tailwind v4 Sticky) embedding
--      Expected: nav/sidebar patterns appear from both legs; union deduplicates
SELECT id, name, category, framework, source
FROM (
    -- FTS leg
    SELECT
        p.id,
        p.name,
        p.category,
        p.framework,
        'fts' AS source,
        f.rank AS sort_key
    FROM patterns_fts f
    JOIN patterns p ON p.rowid = f.rowid
    WHERE patterns_fts MATCH 'sidebar navigation'

    UNION

    -- ANN leg
    SELECT
        p.id,
        p.name,
        p.category,
        p.framework,
        'ann' AS source,
        vector_distance_cos(p.embedding, (SELECT embedding FROM patterns WHERE id = 'pat-003')) AS sort_key
    FROM vector_top_k(
        'patterns_vec_idx',
        (SELECT embedding FROM patterns WHERE id = 'pat-003'),
        10
    ) AS results
    JOIN patterns p ON p.rowid = results.id
)
ORDER BY source, sort_key
LIMIT 20;


-- 5.2  Hybrid search with score fusion (reciprocal rank fusion style)
--      Both legs return rank position; fused score = 1/(k+fts_rank) + 1/(k+ann_rank)
--      where k is a smoothing constant (commonly 60).
--      Expected: patterns that rank well in BOTH legs bubble to the top.
WITH fts_results AS (
    SELECT
        p.id,
        p.name,
        p.category,
        p.framework,
        ROW_NUMBER() OVER (ORDER BY f.rank) AS fts_rank
    FROM patterns_fts f
    JOIN patterns p ON p.rowid = f.rowid
    WHERE patterns_fts MATCH 'dashboard layout'
    LIMIT 10
),
ann_results AS (
    SELECT
        p.id,
        p.name,
        p.category,
        p.framework,
        ROW_NUMBER() OVER (ORDER BY vector_distance_cos(p.embedding, (SELECT embedding FROM patterns WHERE id = 'pat-010'))) AS ann_rank
    FROM vector_top_k(
        'patterns_vec_idx',
        (SELECT embedding FROM patterns WHERE id = 'pat-010'),  -- Dashboard Layout — SvelteKit
        10
    ) AS results
    JOIN patterns p ON p.rowid = results.id
),
fused AS (
    SELECT
        COALESCE(f.id, a.id)               AS id,
        COALESCE(f.name, a.name)           AS name,
        COALESCE(f.category, a.category)   AS category,
        COALESCE(f.framework, a.framework) AS framework,
        COALESCE(1.0 / (60 + f.fts_rank), 0.0) +
        COALESCE(1.0 / (60 + a.ann_rank), 0.0) AS rrf_score
    FROM fts_results f
    FULL OUTER JOIN ann_results a ON f.id = a.id
)
SELECT id, name, category, framework, ROUND(rrf_score, 6) AS rrf_score
FROM fused
ORDER BY rrf_score DESC;


-- 5.3  Hybrid search filtered by css_mode = 'tailwind-v4'
--      FTS term: 'form', ANN probe: pat-007 (Login Form — Next.js Tailwind v4)
--      Expected: form/auth patterns using Tailwind v4 only
SELECT
    p.id,
    p.name,
    p.framework,
    p.css_mode,
    'fts'          AS source,
    NULL           AS ann_distance
FROM patterns_fts f
JOIN patterns p ON p.rowid = f.rowid
WHERE patterns_fts MATCH 'form'
  AND p.css_mode = 'tailwind-v4'

UNION ALL

SELECT
    p.id,
    p.name,
    p.framework,
    p.css_mode,
    'ann'          AS source,
    vector_distance_cos(p.embedding, (SELECT embedding FROM patterns WHERE id = 'pat-007')) AS ann_distance
FROM vector_top_k(
    'patterns_vec_idx',
    (SELECT embedding FROM patterns WHERE id = 'pat-007'),
    8
) AS results
JOIN patterns p ON p.rowid = results.id
WHERE p.css_mode = 'tailwind-v4'

ORDER BY source, ann_distance NULLS LAST;


-- =============================================================================
-- ## SECTION 6: Diagnostic Queries
-- These queries help identify data quality issues before running vector search.
-- Run these if ANN queries return unexpected results.
-- -----------------------------------------------------------------------------

-- 6.1  Patterns with NULL embeddings (not yet processed by Ollama)
--      Expected: empty result set once all patterns are embedded
--      Action if non-empty: run embedding pipeline or check Ollama connectivity
SELECT
    id,
    name,
    domain,
    framework,
    created_at
FROM patterns
WHERE embedding IS NULL
ORDER BY created_at DESC;


-- 6.2  Architectures with NULL embeddings
SELECT
    id,
    name,
    domain,
    created_at
FROM architectures
WHERE embedding IS NULL
ORDER BY created_at DESC;


-- 6.3  Pattern count by category
--      Expected distribution after seeding:
--        landing   ~5  (pat-001,004,009,018, ...)
--        dashboard ~6  (pat-006,008,010,011,016,017)
--        component ~4  (pat-003,005,007,012,013,014,015)
--        form      ~1  (pat-007)
SELECT
    COALESCE(category, '(null)') AS category,
    COUNT(*)                      AS count,
    GROUP_CONCAT(id, ', ')        AS ids
FROM patterns
GROUP BY category
ORDER BY count DESC;


-- 6.4  Pattern count by framework
--      Expected after seeding: astro=5, next=5, sveltekit=3, nuxt=2, vanilla=2
SELECT
    COALESCE(framework, '(null)') AS framework,
    COUNT(*)                       AS count
FROM patterns
GROUP BY framework
ORDER BY count DESC;


-- 6.5  Pattern count by domain
--      Expected: frontend=~13, backend=2, fullstack=2
SELECT
    domain,
    COUNT(*) AS count
FROM patterns
GROUP BY domain
ORDER BY count DESC;


-- 6.6  Architecture count by domain
--      Expected: frontend=2, backend=2, fullstack=2, devops=2
SELECT
    domain,
    COUNT(*) AS count
FROM architectures
GROUP BY domain
ORDER BY count DESC;


-- 6.7  Check for duplicate pattern names (should return empty)
--      If non-empty, OR IGNORE in the seed skipped the duplicate silently.
SELECT
    name,
    COUNT(*) AS occurrences,
    GROUP_CONCAT(id, ', ') AS ids
FROM patterns
GROUP BY name
HAVING COUNT(*) > 1;


-- 6.8  Check for mismatched embedding blob lengths (should return empty)
--      Any row here has a corrupted embedding that will cause vector_top_k errors.
SELECT
    id,
    name,
    LENGTH(embedding) AS blob_bytes,
    LENGTH(embedding) / 4 AS apparent_dimensions
FROM patterns
WHERE embedding IS NOT NULL
  AND LENGTH(embedding) != 3072   -- 768 dims * 4 bytes/float32
ORDER BY blob_bytes;


-- 6.9  Verify FTS content table is in sync with base table
--      If counts differ, the FTS index needs a rebuild.
--      Expected: both counts should match.
SELECT
    (SELECT COUNT(*) FROM patterns)     AS patterns_base_count,
    (SELECT COUNT(*) FROM patterns_fts) AS patterns_fts_count,
    (SELECT COUNT(*) FROM architectures)     AS arch_base_count,
    (SELECT COUNT(*) FROM architectures_fts) AS arch_fts_count;


-- 6.10 Rebuild FTS index if out of sync (run only if 6.9 shows a mismatch)
--      Uncomment and run as needed:
-- INSERT INTO patterns_fts(patterns_fts) VALUES('rebuild');
-- INSERT INTO architectures_fts(architectures_fts) VALUES('rebuild');
