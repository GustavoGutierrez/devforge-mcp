package tools_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"dev-forge-mcp/internal/testutil"
	"dev-forge-mcp/internal/tools"
)

// seedPattern inserts a pattern row and its FTS entry.
func seedPattern(t *testing.T, db *sql.DB, id, name, domain, framework, cssMode, tags, description string) {
	t.Helper()
	db.Exec(`INSERT INTO patterns (id, name, domain, framework, css_mode, tags, description) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, name, domain, framework, cssMode, tags, description)
	db.Exec(`INSERT INTO patterns_fts (rowid, name, category, tags, description)
		SELECT rowid, name, category, tags, description FROM patterns WHERE id = ?`, id)
}

func TestListPatterns_EmptyDB_ReturnsEmptySlice(t *testing.T) {
	database := testutil.NewTestDB(t)
	srv := &tools.Server{DB: database}

	result := srv.ListPatterns(context.Background(), tools.ListPatternsInput{Limit: 20})

	var out tools.ListPatternsOutput
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		t.Fatalf("invalid JSON: %v — got: %s", err, result)
	}
	if out.Patterns == nil {
		t.Error("expected empty slice (not nil) for empty DB")
	}
	if len(out.Patterns) != 0 {
		t.Errorf("expected 0 patterns, got %d", len(out.Patterns))
	}
}

func TestListPatterns_FilterByDomain(t *testing.T) {
	database := testutil.NewTestDB(t)
	seedPattern(t, database, "p1", "Frontend Pattern", "frontend", "astro", "tailwind-v4", "hero", "A frontend layout")
	seedPattern(t, database, "p2", "Backend Pattern", "backend", "vanilla", "plain-css", "api", "A backend structure")

	srv := &tools.Server{DB: database}

	result := srv.ListPatterns(context.Background(), tools.ListPatternsInput{
		Domain: "frontend",
		Limit:  20,
	})

	var out tools.ListPatternsOutput
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out.Patterns) == 0 {
		t.Error("expected at least one frontend pattern")
	}
	for _, p := range out.Patterns {
		if p.Domain != "frontend" {
			t.Errorf("expected domain=frontend, got %q for pattern %q", p.Domain, p.Name)
		}
	}
}

func TestListPatterns_FTSKeywordMatch(t *testing.T) {
	database := testutil.NewTestDB(t)
	seedPattern(t, database, "p1", "Pricing Table", "frontend", "next", "tailwind-v4", "pricing,table", "A beautiful pricing table component")
	seedPattern(t, database, "p2", "Hero Section", "frontend", "astro", "tailwind-v4", "hero,landing", "Hero for landing pages")

	srv := &tools.Server{DB: database}

	result := srv.ListPatterns(context.Background(), tools.ListPatternsInput{
		Query: "pricing",
		Mode:  "fts",
		Limit: 20,
	})

	var out tools.ListPatternsOutput
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out.Patterns) == 0 {
		t.Error("FTS search for 'pricing' returned 0 results")
	}
}

func TestListPatterns_FilterByFramework(t *testing.T) {
	database := testutil.NewTestDB(t)
	seedPattern(t, database, "p1", "Astro Component", "frontend", "astro", "tailwind-v4", "", "")
	seedPattern(t, database, "p2", "Next Component", "frontend", "next", "tailwind-v4", "", "")
	seedPattern(t, database, "p3", "Astro Layout", "frontend", "astro", "plain-css", "", "")

	srv := &tools.Server{DB: database}

	result := srv.ListPatterns(context.Background(), tools.ListPatternsInput{
		Framework: "astro",
		Limit:     20,
	})

	var out tools.ListPatternsOutput
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out.Patterns) != 2 {
		t.Errorf("expected 2 astro patterns, got %d", len(out.Patterns))
	}
}

func TestListPatterns_SemanticMode_FallsBackToFTSWithoutOllama(t *testing.T) {
	database := testutil.NewTestDB(t)
	seedPattern(t, database, "p1", "Dashboard Layout", "frontend", "next", "tailwind-v4", "dashboard,grid", "A grid-based dashboard layout")

	// No embedder — should fall back to FTS
	srv := &tools.Server{DB: database, Embedder: nil}

	result := srv.ListPatterns(context.Background(), tools.ListPatternsInput{
		Query: "dashboard",
		Mode:  "semantic",
		Limit: 20,
	})

	var out tools.ListPatternsOutput
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out.Patterns) == 0 {
		t.Error("expected results when falling back from semantic to FTS")
	}
}
