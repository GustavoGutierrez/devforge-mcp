// Package db provides libSQL database setup, schema migrations, and query helpers.
package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/tursodatabase/go-libsql"
)

// Open opens a libSQL embedded database at the given file path and runs migrations.
// path should be like "file:./db/ui_patterns.db" or a temp file path for tests.
func Open(path string) (*sql.DB, error) {
	database, err := sql.Open("libsql", path)
	if err != nil {
		return nil, fmt.Errorf("db: open: %w", err)
	}
	if err := RunMigrations(database); err != nil {
		database.Close()
		return nil, fmt.Errorf("db: migrations: %w", err)
	}
	return database, nil
}

// RunMigrations applies the schema (idempotent) to the given database.
func RunMigrations(database *sql.DB) error {
	// Run PRAGMAs via QueryRow (they return result rows)
	for _, pragma := range pragmas {
		row := database.QueryRow(pragma)
		var result interface{}
		_ = row.Scan(&result) // ignore result value, just execute
	}

	// Run DDL statements via Exec
	stmts := splitStatements(schema)
	for _, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := database.Exec(stmt); err != nil {
			return fmt.Errorf("db: migration error in %q: %w", truncate(stmt, 60), err)
		}
	}
	return nil
}

// splitStatements splits an SQL string into individual statements by semicolon.
func splitStatements(s string) []string {
	parts := strings.Split(s, ";")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p+";")
		}
	}
	return result
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// Pattern represents a stored UI pattern.
type Pattern struct {
	ID          string
	Name        string
	Domain      string
	Category    string
	Framework   string
	CSSMode     string
	Tags        string
	Snippet     string
	CSSSnippet  string
	Description string
	CreatedAt   string
}

// Architecture represents a stored architecture.
type Architecture struct {
	ID          string
	Name        string
	Domain      string
	Framework   string
	CSSMode     string
	Description string
	Decisions   string
	Tags        string
	CreatedAt   string
}

// Token represents a stored design token.
type Token struct {
	ID        string
	CSSMode   string
	Scope     string
	Key       string
	Value     string
	CreatedAt string
}

// Audit represents a stored layout audit report.
type Audit struct {
	ID         string
	PageType   string
	Framework  string
	CSSMode    string
	ReportJSON string
	CreatedAt  string
}

// Palette represents a stored color palette.
type Palette struct {
	ID         string
	Name       string
	UseCase    string
	Mood       string
	TokensJSON string
	CreatedAt  string
}
