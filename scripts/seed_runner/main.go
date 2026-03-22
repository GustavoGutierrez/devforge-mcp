// scripts/seed_runner.go — SQL seed runner for make db-seed.
//
// Applies one or more SQL files to a libSQL database using the go-libsql
// driver (which understands F32_BLOB and libsql_vector_idx).  Plain sqlite3
// CLI cannot be used because those are libSQL-specific types.
//
// The SQL splitter correctly handles semicolons inside single-quoted string
// literals (e.g. CSS snippets in INSERT statements) so it does not break
// multi-line values.
//
// Usage:
//
//	CGO_ENABLED=1 go run ./scripts/seed_runner.go -db ./dist/dev-forge.db -sql db/seeds/001_patterns.sql
//	CGO_ENABLED=1 go run ./scripts/seed_runner.go -db ./dist/dev-forge.db -sql db/seeds/001_patterns.sql -sql db/seeds/002_architectures.sql
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/tursodatabase/go-libsql"
)

// multiFlag allows -sql to be repeated multiple times on the command line.
type multiFlag []string

func (m *multiFlag) String() string        { return strings.Join(*m, ", ") }
func (m *multiFlag) Set(v string) error    { *m = append(*m, v); return nil }

func main() {
	dbPath := flag.String("db", "", "Path to the libSQL database file")
	var sqlFiles multiFlag
	flag.Var(&sqlFiles, "sql", "SQL file to apply (may be repeated)")
	flag.Parse()

	if *dbPath == "" || len(sqlFiles) == 0 {
		fmt.Fprintln(os.Stderr, "usage: seed_runner -db <path> -sql <file> [-sql <file> ...]")
		os.Exit(1)
	}

	database, err := sql.Open("libsql", "file:"+*dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer database.Close()

	totalApplied := 0
	for _, sqlFile := range sqlFiles {
		n, err := applyFile(database, sqlFile)
		if err != nil {
			log.Fatalf("apply %s: %v", sqlFile, err)
		}
		fmt.Printf("  applied %d statement(s) from %s\n", n, sqlFile)
		totalApplied += n
	}
	fmt.Printf("Seed complete — %d total statement(s) applied\n", totalApplied)
}

func applyFile(database *sql.DB, path string) (int, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("read %s: %w", path, err)
	}

	stmts := splitSQL(string(content))
	applied := 0
	for _, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := database.Exec(stmt); err != nil {
			// Print a warning but continue — INSERT OR IGNORE conflicts are not fatal.
			fmt.Fprintf(os.Stderr, "  warning: %v\n    stmt: %s\n", err, truncate(stmt, 80))
		} else {
			applied++
		}
	}
	return applied, nil
}

// splitSQL splits an SQL string into individual statements, correctly handling:
//   - Single-quoted string literals (including escaped single quotes via '')
//   - Line comments (-- ...)
//   - Statement terminator is ; outside of any string or comment
func splitSQL(s string) []string {
	var stmts []string
	var buf strings.Builder
	inString := false
	i := 0
	runes := []rune(s)
	n := len(runes)

	for i < n {
		ch := runes[i]

		if inString {
			buf.WriteRune(ch)
			if ch == '\'' {
				// Check for escaped quote: '' inside a string literal
				if i+1 < n && runes[i+1] == '\'' {
					// Write the second quote and advance
					i++
					buf.WriteRune(runes[i])
				} else {
					// End of string literal
					inString = false
				}
			}
			i++
			continue
		}

		// Not in a string
		switch {
		case ch == '\'':
			// Start of string literal
			inString = true
			buf.WriteRune(ch)
			i++

		case ch == '-' && i+1 < n && runes[i+1] == '-':
			// Line comment: skip to end of line
			for i < n && runes[i] != '\n' {
				i++
			}

		case ch == ';':
			// Statement terminator
			stmt := strings.TrimSpace(buf.String())
			if stmt != "" {
				stmts = append(stmts, stmt+";")
			}
			buf.Reset()
			i++

		default:
			buf.WriteRune(ch)
			i++
		}
	}

	// Handle any trailing statement without a final semicolon
	if stmt := strings.TrimSpace(buf.String()); stmt != "" {
		stmts = append(stmts, stmt+";")
	}

	return stmts
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
