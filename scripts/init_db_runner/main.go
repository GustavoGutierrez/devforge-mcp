// scripts/init_db_runner.go — DB initializer helper for make db-init.
//
// This program opens (or creates) a libSQL database at the given path and
// runs all schema migrations using the same db.Open() / RunMigrations() code
// that the MCP server uses at runtime.  It is intentionally kept simple: no
// MCP server, no embeddings, no stdio transport — just DB init.
//
// Usage:
//
//	CGO_ENABLED=1 go run ./scripts/init_db_runner.go -db ./dist/dev-forge.db
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"dev-forge-mcp/internal/db"
)

func main() {
	dbPath := flag.String("db", "", "Path to the libSQL database file to create/migrate")
	flag.Parse()

	if *dbPath == "" {
		fmt.Fprintln(os.Stderr, "usage: init_db_runner -db <path>")
		os.Exit(1)
	}

	// Ensure the parent directory exists
	dir := filepath.Dir(*dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatalf("failed to create directory %s: %v", dir, err)
	}

	// Open (creates if missing) and run migrations via the project's own db package.
	// db.Open() calls RunMigrations() internally — idempotent, safe to re-run.
	database, err := db.Open("file:" + *dbPath)
	if err != nil {
		log.Fatalf("failed to open/migrate database at %s: %v", *dbPath, err)
	}
	database.Close()

	abs, _ := filepath.Abs(*dbPath)
	fmt.Printf("Database initialized: %s\n", abs)
}
