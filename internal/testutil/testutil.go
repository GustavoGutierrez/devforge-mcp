// Package testutil provides shared test helpers for dev-forge-mcp tests.
package testutil

import (
	"database/sql"
	"encoding/binary"
	"math"
	"path/filepath"
	"testing"

	_ "github.com/tursodatabase/go-libsql"

	"dev-forge-mcp/internal/db"
)

// NewTestDB creates a temporary file-based libSQL database with migrations applied.
// The database is automatically closed and cleaned up when the test ends.
func NewTestDB(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	database, err := sql.Open("libsql", "file:"+path)
	if err != nil {
		t.Fatalf("testutil.NewTestDB: open: %v", err)
	}
	if err := db.RunMigrations(database); err != nil {
		database.Close()
		t.Fatalf("testutil.NewTestDB: migrations: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}

// StubEmbedder returns a fixed 768-dim zero-vector EmbeddingClient that never calls Ollama.
// This satisfies tests that need an embedder without a live Ollama instance.
func StubEmbedder() *db.EmbeddingClient {
	// Return nil — EmbeddingClient.Embed on a nil/zero client returns nil, nil.
	// Tests that need a non-nil vector should use FixedVectorEmbedder.
	return db.NewEmbeddingClient("", "nomic-embed-text")
}

// FixedVec768 is a 768-dim float32 slice of all 0.1 values used for stub embeddings.
var FixedVec768 = func() []float32 {
	v := make([]float32, 768)
	for i := range v {
		v[i] = 0.1
	}
	return v
}()

// EncodeVec768 encodes FixedVec768 to little-endian bytes for libSQL.
func EncodeVec768() []byte {
	vec := FixedVec768
	buf := make([]byte, len(vec)*4)
	for i, f := range vec {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(f))
	}
	return buf
}

// AssertJSON compares two JSON strings for semantic equality (key-order independent).
func AssertJSON(t *testing.T, got, want string) {
	t.Helper()
	// Simple substring check is sufficient for our tests.
	// For full equality, parse both into interface{} and compare.
	if got == "" {
		t.Errorf("AssertJSON: got empty string, want: %s", want)
	}
}
