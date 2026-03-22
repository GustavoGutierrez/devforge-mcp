package db_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/tursodatabase/go-libsql"

	"dev-forge-mcp/internal/db"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	database, err := sql.Open("libsql", "file:"+path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.RunMigrations(database); err != nil {
		database.Close()
		t.Fatalf("migrations: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}

func TestRunMigrations_AllTablesCreated(t *testing.T) {
	database := newTestDB(t)

	tables := []string{
		"patterns", "architectures", "tokens", "audits", "assets", "palettes",
	}
	for _, table := range tables {
		var name string
		err := database.QueryRow(
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table,
		).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}

func TestRunMigrations_FTS5VirtualTablesCreated(t *testing.T) {
	database := newTestDB(t)

	vtables := []string{"patterns_fts", "architectures_fts"}
	for _, vt := range vtables {
		var name string
		err := database.QueryRow(
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, vt,
		).Scan(&name)
		if err != nil {
			t.Errorf("FTS5 virtual table %q not found: %v", vt, err)
		}
	}
}

func TestRunMigrations_VectorIndexCreated(t *testing.T) {
	database := newTestDB(t)

	indexes := []string{"patterns_vec_idx", "architectures_vec_idx"}
	for _, idx := range indexes {
		var name string
		err := database.QueryRow(
			`SELECT name FROM sqlite_master WHERE type='index' AND name=?`, idx,
		).Scan(&name)
		if err != nil {
			t.Errorf("vector index %q not found: %v", idx, err)
		}
	}
}

func TestRunMigrations_WALMode(t *testing.T) {
	database := newTestDB(t)

	var mode string
	err := database.QueryRow("PRAGMA journal_mode").Scan(&mode)
	if err != nil {
		t.Fatalf("query journal_mode: %v", err)
	}
	if mode != "wal" {
		t.Errorf("expected WAL mode, got %q", mode)
	}
}

func TestRunMigrations_Idempotent(t *testing.T) {
	database := newTestDB(t)

	// Run migrations again — should not fail
	if err := db.RunMigrations(database); err != nil {
		t.Errorf("second migration run failed: %v", err)
	}
}

func TestRunMigrations_FTS5InsertSearch(t *testing.T) {
	database := newTestDB(t)

	// Insert a pattern
	_, err := database.Exec(
		`INSERT INTO patterns (id, name, domain, framework, css_mode, tags, description)
		 VALUES ('test-1', 'Hero Section Tailwind', 'frontend', 'astro', 'tailwind-v4', 'hero,landing', 'A hero section for landing pages')`,
	)
	if err != nil {
		t.Fatalf("insert pattern: %v", err)
	}

	// Insert into FTS5
	_, err = database.Exec(
		`INSERT INTO patterns_fts (rowid, name, category, tags, description)
		 SELECT rowid, name, category, tags, description FROM patterns WHERE id = 'test-1'`,
	)
	if err != nil {
		t.Fatalf("insert fts: %v", err)
	}

	// FTS5 search
	var count int
	err = database.QueryRow(
		`SELECT COUNT(*) FROM patterns_fts WHERE patterns_fts MATCH 'hero'`,
	).Scan(&count)
	if err != nil {
		t.Fatalf("fts search: %v", err)
	}
	if count == 0 {
		t.Error("FTS5 search returned 0 results after insert")
	}
}
