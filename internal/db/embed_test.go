package db_test

import (
	"context"
	"testing"
	"time"

	"dev-forge-mcp/internal/db"
)

func TestEmbeddingClient_EmptyURL_ReturnsNil(t *testing.T) {
	client := db.NewEmbeddingClient("", "nomic-embed-text")
	vec, err := client.Embed(context.Background(), "test text")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if vec != nil {
		t.Errorf("expected nil vector when URL is empty, got %v", vec)
	}
}

func TestEmbeddingClient_UnreachableURL_ReturnsNilWithinTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}

	client := db.NewEmbeddingClient("http://127.0.0.1:1", "nomic-embed-text")

	start := time.Now()
	vec, err := client.Embed(context.Background(), "test text")
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("expected nil error (graceful skip), got %v", err)
	}
	if vec != nil {
		t.Errorf("expected nil vector for unreachable Ollama, got %v", vec)
	}
	if elapsed > 1500*time.Millisecond {
		t.Errorf("expected timeout within 1.5s, took %v", elapsed)
	}
}

func TestEmbeddingClient_LiveOllama(t *testing.T) {
	ollamaURL := "http://localhost:11434"
	if !db.CheckAvailability(ollamaURL) {
		t.Skip("OLLAMA not available — skipping live embedding test")
	}

	client := db.NewEmbeddingClient(ollamaURL, "nomic-embed-text")
	vec, err := client.Embed(context.Background(), "hello world")
	if err != nil {
		t.Fatalf("Embed returned error: %v", err)
	}
	if len(vec) == 0 {
		t.Error("expected non-empty vector from live Ollama")
	}
	if len(vec) != 768 {
		t.Errorf("expected 768-dim vector, got %d", len(vec))
	}
}

func TestEncodeDecodeForDB_RoundTrip(t *testing.T) {
	original := []float32{0.1, 0.2, 0.3, 1.0, -1.0, 0.0}
	encoded := db.EncodeForDB(original)
	decoded := db.DecodeFromDB(encoded)

	if len(decoded) != len(original) {
		t.Fatalf("length mismatch: got %d, want %d", len(decoded), len(original))
	}
	for i, v := range original {
		if decoded[i] != v {
			t.Errorf("mismatch at index %d: got %v, want %v", i, decoded[i], v)
		}
	}
}

func TestEncodeForDB_Nil(t *testing.T) {
	if b := db.EncodeForDB(nil); b != nil {
		t.Errorf("expected nil for nil input, got %v", b)
	}
}

func TestDecodeFromDB_Empty(t *testing.T) {
	if v := db.DecodeFromDB(nil); v != nil {
		t.Errorf("expected nil for nil input, got %v", v)
	}
	if v := db.DecodeFromDB([]byte{1, 2, 3}); v != nil {
		t.Errorf("expected nil for non-4-byte-aligned input, got %v", v)
	}
}
