package tools

import (
	"context"
	"database/sql"
)

// ListPatternsInput is the input schema for the list_patterns tool.
type ListPatternsInput struct {
	Domain    string `json:"domain,omitempty"`
	CSSMode   string `json:"css_mode,omitempty"`
	Framework string `json:"framework,omitempty"`
	Query     string `json:"query,omitempty"`
	Mode      string `json:"mode,omitempty"` // fts | semantic | filter
	Limit     int    `json:"limit,omitempty"`
}

// PatternResult is a single pattern returned from list_patterns.
type PatternResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	Category    string `json:"category"`
	Framework   string `json:"framework"`
	CSSMode     string `json:"css_mode"`
	Tags        string `json:"tags"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

// ListPatternsOutput is the output schema for the list_patterns tool.
type ListPatternsOutput struct {
	Patterns []PatternResult `json:"patterns"`
	Total    int             `json:"total"`
}

// ListPatterns implements the list_patterns MCP tool.
func (s *Server) ListPatterns(ctx context.Context, input ListPatternsInput) string {
	if s.DB == nil {
		return errorJSON("database not available")
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}

	// Auto-detect mode
	mode := input.Mode
	if mode == "" {
		if input.Query != "" {
			mode = "fts"
		} else {
			mode = "filter"
		}
	}

	var patterns []PatternResult
	var err error

	switch mode {
	case "semantic":
		patterns, err = s.listPatternsSemanticOrFTS(ctx, input, limit)
	case "fts":
		patterns, err = s.listPatternsFTS(ctx, input, limit)
	default:
		patterns, err = s.listPatternsFilter(ctx, input, limit)
	}

	if err != nil {
		// Fall back to filter on error
		patterns, _ = s.listPatternsFilter(ctx, input, limit)
	}

	if patterns == nil {
		patterns = []PatternResult{}
	}

	return mustJSON(ListPatternsOutput{
		Patterns: patterns,
		Total:    len(patterns),
	})
}

func (s *Server) listPatternsSemanticOrFTS(ctx context.Context, input ListPatternsInput, limit int) ([]PatternResult, error) {
	if s.Embedder != nil && input.Query != "" {
		vec, err := s.Embedder.Embed(ctx, input.Query)
		if err == nil && vec != nil {
			results, err := s.listPatternsVector(ctx, input, vec, limit)
			if err == nil {
				return results, nil
			}
		}
	}
	// Fall back to FTS
	if input.Query != "" {
		return s.listPatternsFTS(ctx, input, limit)
	}
	return s.listPatternsFilter(ctx, input, limit)
}

func (s *Server) listPatternsVector(ctx context.Context, input ListPatternsInput, vec []float32, limit int) ([]PatternResult, error) {
	encoded := encodeFloat32(vec)
	query := `
		SELECT p.id, p.name, p.domain, COALESCE(p.category,''), p.framework, p.css_mode, COALESCE(p.tags,''), COALESCE(p.description,''), p.created_at
		FROM vector_top_k('patterns_vec_idx', ?, ?) AS results
		JOIN patterns p ON p.rowid = results.id
		WHERE (? = '' OR p.domain = ?)
		  AND (? = '' OR p.framework = ?)
		  AND (? = '' OR p.css_mode = ?)
		ORDER BY results.distance`

	rows, err := s.DB.QueryContext(ctx, query,
		encoded, limit,
		input.Domain, input.Domain,
		input.Framework, input.Framework,
		input.CSSMode, input.CSSMode,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPatterns(rows)
}

func (s *Server) listPatternsFTS(ctx context.Context, input ListPatternsInput, limit int) ([]PatternResult, error) {
	if input.Query == "" {
		return s.listPatternsFilter(ctx, input, limit)
	}
	query := `
		SELECT p.id, p.name, p.domain, COALESCE(p.category,''), COALESCE(p.framework,''), COALESCE(p.css_mode,''), COALESCE(p.tags,''), COALESCE(p.description,''), p.created_at
		FROM patterns_fts f
		JOIN patterns p ON p.rowid = f.rowid
		WHERE patterns_fts MATCH ?
		  AND (? = '' OR p.domain = ?)
		  AND (? = '' OR p.framework = ?)
		  AND (? = '' OR p.css_mode = ?)
		ORDER BY rank
		LIMIT ?`

	rows, err := s.DB.QueryContext(ctx, query,
		input.Query,
		input.Domain, input.Domain,
		input.Framework, input.Framework,
		input.CSSMode, input.CSSMode,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPatterns(rows)
}

func (s *Server) listPatternsFilter(ctx context.Context, input ListPatternsInput, limit int) ([]PatternResult, error) {
	query := `
		SELECT id, name, domain, COALESCE(category,''), COALESCE(framework,''), COALESCE(css_mode,''), COALESCE(tags,''), COALESCE(description,''), created_at
		FROM patterns
		WHERE (? = '' OR domain = ?)
		  AND (? = '' OR framework = ?)
		  AND (? = '' OR css_mode = ?)
		ORDER BY created_at DESC
		LIMIT ?`

	rows, err := s.DB.QueryContext(ctx, query,
		input.Domain, input.Domain,
		input.Framework, input.Framework,
		input.CSSMode, input.CSSMode,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPatterns(rows)
}

func scanPatterns(rows *sql.Rows) ([]PatternResult, error) {
	var results []PatternResult
	for rows.Next() {
		var p PatternResult
		if err := rows.Scan(&p.ID, &p.Name, &p.Domain, &p.Category, &p.Framework, &p.CSSMode, &p.Tags, &p.Description, &p.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, p)
	}
	return results, rows.Err()
}
