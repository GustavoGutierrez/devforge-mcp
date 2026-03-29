package tools

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

// StoreArchitectureInput is the input schema for the store_architecture tool.
type StoreArchitectureInput struct {
	Name        string `json:"name"`
	Domain      string `json:"domain,omitempty"`      // default "fullstack"
	Framework   string `json:"framework,omitempty"`
	CSSMode     string `json:"css_mode,omitempty"`
	Description string `json:"description,omitempty"`
	Decisions   string `json:"decisions,omitempty"`
	Tags        string `json:"tags,omitempty"`
}

// StoreArchitectureOutput is the output schema for the store_architecture tool.
type StoreArchitectureOutput struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// StoreArchitecture implements the store_architecture MCP tool.
func (s *Server) StoreArchitecture(ctx context.Context, input StoreArchitectureInput) string {
	if strings.TrimSpace(input.Name) == "" {
		return errorJSON("name is required")
	}

	id := uuid.New().String()
	createdAt := time.Now().UTC().Format(time.RFC3339)

	domain := input.Domain
	if domain == "" {
		domain = "fullstack"
	}

	if s.DB == nil {
		return errorJSON("database not available")
	}

	// Insert architecture
	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO architectures (id, name, domain, framework, css_mode, description, decisions, tags)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, input.Name, domain, input.Framework, input.CSSMode,
		input.Description, input.Decisions, input.Tags,
	)
	if err != nil {
		return errorJSON("failed to insert architecture: " + err.Error())
	}

	// Insert into FTS5 index
	s.DB.ExecContext(ctx,
		`INSERT INTO architectures_fts (rowid, name, description, tags, decisions)
		 SELECT rowid, name, description, tags, decisions FROM architectures WHERE id = ?`,
		id,
	)

	// Generate embedding in background if available
	if s.Embedder != nil {
		textToEmbed := input.Name + " " + input.Description + " " + input.Tags + " " + input.Decisions
		go func() {
			vec, _ := s.Embedder.Embed(context.Background(), textToEmbed)
			if vec != nil {
				encoded := encodeFloat32(vec)
				s.DB.Exec(`UPDATE architectures SET embedding = ? WHERE id = ?`, encoded, id)
			}
		}()
	}

	return mustJSON(StoreArchitectureOutput{
		ID:        id,
		Name:      input.Name,
		CreatedAt: createdAt,
	})
}
