package db

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"time"

	ollamaapi "github.com/ollama/ollama/api"
)

// EmbeddingClient wraps Ollama's embed endpoint.
// All methods degrade gracefully — if Ollama is unavailable, they return nil, nil.
type EmbeddingClient struct {
	url   string // e.g. http://localhost:11434
	model string // e.g. nomic-embed-text
}

// NewEmbeddingClient creates a new EmbeddingClient.
// If ollamaURL is empty, all Embed() calls will return nil, nil gracefully.
func NewEmbeddingClient(ollamaURL, model string) *EmbeddingClient {
	return &EmbeddingClient{
		url:   ollamaURL,
		model: model,
	}
}

// Embed returns a 768-dim float32 slice, or nil if Ollama is unreachable or URL is empty.
func (e *EmbeddingClient) Embed(ctx context.Context, text string) ([]float32, error) {
	if e == nil || e.url == "" {
		return nil, nil
	}

	// Test Ollama availability with a short timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	client := createOllamaClient(e.url)
	if client == nil {
		return nil, nil
	}

	req := &ollamaapi.EmbeddingRequest{
		Model:  e.model,
		Prompt: text,
	}

	resp, err := client.Embeddings(timeoutCtx, req)
	if err != nil {
		// Graceful degradation: return nil, nil (not an error)
		return nil, nil
	}

	if len(resp.Embedding) == 0 {
		return nil, nil
	}

	result := make([]float32, len(resp.Embedding))
	for i, v := range resp.Embedding {
		result[i] = float32(v)
	}
	return result, nil
}

// createOllamaClient creates an Ollama API client pointing at the given URL.
func createOllamaClient(rawURL string) *ollamaapi.Client {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil
	}
	httpClient := &http.Client{Timeout: 2 * time.Second}
	return ollamaapi.NewClient(u, httpClient)
}

// EncodeForDB encodes a float32 slice to a little-endian IEEE 754 byte slice
// suitable for libSQL F32_BLOB columns.
func EncodeForDB(vec []float32) []byte {
	if len(vec) == 0 {
		return nil
	}
	buf := make([]byte, len(vec)*4)
	for i, f := range vec {
		bits := math.Float32bits(f)
		binary.LittleEndian.PutUint32(buf[i*4:], bits)
	}
	return buf
}

// DecodeFromDB decodes a little-endian IEEE 754 byte slice from a libSQL F32_BLOB column
// back to a float32 slice.
func DecodeFromDB(blob []byte) []float32 {
	if len(blob) == 0 || len(blob)%4 != 0 {
		return nil
	}
	result := make([]float32, len(blob)/4)
	for i := range result {
		bits := binary.LittleEndian.Uint32(blob[i*4:])
		result[i] = math.Float32frombits(bits)
	}
	return result
}

// CheckAvailability pings Ollama with a 1-second timeout.
// Returns true if Ollama is reachable.
func CheckAvailability(ollamaURL string) bool {
	if ollamaURL == "" {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ollamaURL+"/api/tags", nil)
	if err != nil {
		return false
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// Validate checks if the embedding client configuration is usable.
func (e *EmbeddingClient) Validate() error {
	if e == nil || e.url == "" {
		return fmt.Errorf("ollama URL not configured")
	}
	if !CheckAvailability(e.url) {
		return fmt.Errorf("ollama not reachable at %s", e.url)
	}
	return nil
}
