// Package tools contains MCP tool handler implementations.
// Each tool is implemented in its own file following the naming convention
// of the tool name (e.g., analyze_layout.go, manage_tokens.go).
package tools

import (
	"database/sql"
	"encoding/json"

	"dev-forge-mcp/internal/db"
	"dev-forge-mcp/internal/dpf"
)

// Server holds shared dependencies for all tool handlers.
type Server struct {
	DB       *sql.DB
	DPF      *dpf.StreamClient
	Embedder *db.EmbeddingClient
	// GetConfig returns the current config (hot-reloadable).
	GetConfig func() interface{ GetGeminiAPIKey() string }
}

// errorJSON returns a JSON-encoded error response.
func errorJSON(msg string) string {
	b, _ := json.Marshal(map[string]string{"error": msg})
	return string(b)
}

// mustJSON marshals v to JSON or returns an error JSON.
func mustJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return errorJSON("failed to marshal response: " + err.Error())
	}
	return string(b)
}
