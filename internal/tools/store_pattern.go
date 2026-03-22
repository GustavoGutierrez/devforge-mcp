package tools

import (
	"context"
	"encoding/binary"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

// StorePatternInput is the input schema for the store_pattern tool.
type StorePatternInput struct {
	Name        string `json:"name"`
	Category    string `json:"category,omitempty"`
	Domain      string `json:"domain,omitempty"`
	Framework   string `json:"framework"`
	CSSMode     string `json:"css_mode"`
	Tags        string `json:"tags,omitempty"`
	Snippet     string `json:"snippet"`
	CSSSnippet  string `json:"css_snippet,omitempty"`
	Description string `json:"description,omitempty"`
}

// StorePatternOutput is the output schema for the store_pattern tool.
type StorePatternOutput struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// StorePattern implements the store_pattern MCP tool.
func (s *Server) StorePattern(ctx context.Context, input StorePatternInput) string {
	if strings.TrimSpace(input.Name) == "" {
		return errorJSON("name is required")
	}
	if strings.TrimSpace(input.Framework) == "" {
		return errorJSON("framework is required")
	}
	if strings.TrimSpace(input.CSSMode) == "" {
		return errorJSON("css_mode is required")
	}
	if strings.TrimSpace(input.Snippet) == "" {
		return errorJSON("snippet is required")
	}

	id := uuid.New().String()
	createdAt := time.Now().UTC().Format(time.RFC3339)

	domain := input.Domain
	if domain == "" {
		domain = "frontend"
	}

	if s.DB == nil {
		return errorJSON("database not available")
	}

	// Insert pattern
	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO patterns (id, name, domain, category, framework, css_mode, tags, snippet, css_snippet, description)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, input.Name, domain, input.Category, input.Framework, input.CSSMode,
		input.Tags, input.Snippet, input.CSSSnippet, input.Description,
	)
	if err != nil {
		return errorJSON("failed to insert pattern: " + err.Error())
	}

	// Insert into FTS5 index
	s.DB.ExecContext(ctx,
		`INSERT INTO patterns_fts (rowid, name, category, tags, description)
		 SELECT rowid, name, category, tags, description FROM patterns WHERE id = ?`,
		id,
	)

	// Generate embedding in background if available
	if s.Embedder != nil {
		textToEmbed := input.Name + " " + input.Description + " " + input.Tags
		go func() {
			vec, _ := s.Embedder.Embed(context.Background(), textToEmbed)
			if vec != nil {
				encoded := encodeFloat32(vec)
				s.DB.Exec(`UPDATE patterns SET embedding = ? WHERE id = ?`, encoded, id)
			}
		}()
	}

	return mustJSON(StorePatternOutput{
		ID:        id,
		Name:      input.Name,
		CreatedAt: createdAt,
	})
}

// encodeFloat32 encodes a float32 slice to little-endian bytes for libSQL F32_BLOB.
func encodeFloat32(vec []float32) []byte {
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
