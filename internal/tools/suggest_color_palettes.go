package tools

import (
	"context"
	"fmt"
	"strings"
)

// SuggestColorPalettesInput is the input schema for the suggest_color_palettes tool.
type SuggestColorPalettesInput struct {
	UseCase      string   `json:"use_case"`
	BrandKeywords []string `json:"brand_keywords,omitempty"`
	Mood         string   `json:"mood,omitempty"`
	Count        int      `json:"count,omitempty"`
}

// PaletteTokens is the token map for a palette.
type PaletteTokens struct {
	Background  string `json:"background"`
	Surface     string `json:"surface"`
	Primary     string `json:"primary"`
	PrimarySoft string `json:"primary-soft"`
	Accent      string `json:"accent"`
	Text        string `json:"text"`
	Muted       string `json:"muted"`
}

// ColorPalette represents a single palette suggestion.
type ColorPalette struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Tokens      PaletteTokens `json:"tokens"`
}

// SuggestColorPalettesOutput is the output schema for the suggest_color_palettes tool.
type SuggestColorPalettesOutput struct {
	Palettes []ColorPalette `json:"palettes"`
}

// SuggestColorPalettes implements the suggest_color_palettes MCP tool.
func (s *Server) SuggestColorPalettes(ctx context.Context, input SuggestColorPalettesInput) string {
	if strings.TrimSpace(input.UseCase) == "" {
		return errorJSON("use_case is required")
	}

	count := input.Count
	if count <= 0 {
		count = 3
	}

	palettes := generatePalettes(input, count)

	return mustJSON(SuggestColorPalettesOutput{
		Palettes: palettes,
	})
}

// paletteTemplate is a predefined palette configuration.
type paletteTemplate struct {
	name        string
	description string
	mood        string
	useCaseTags []string
	tokens      PaletteTokens
}

// predefinedPalettes is the library of curated palettes.
var predefinedPalettes = []paletteTemplate{
	{
		name:        "Fintech Calm Blue",
		description: "Professional and trustworthy deep blue palette for financial applications.",
		mood:        "serious",
		useCaseTags: []string{"saas", "dashboard", "finance", "banking", "professional"},
		tokens: PaletteTokens{
			Background:  "#0b1220",
			Surface:     "#020617",
			Primary:     "#22d3ee",
			PrimarySoft: "#0f172a",
			Accent:      "#facc15",
			Text:        "#e2e8f0",
			Muted:       "#475569",
		},
	},
	{
		name:        "SaaS Indigo",
		description: "Modern indigo palette for SaaS dashboards and productivity tools.",
		mood:        "professional",
		useCaseTags: []string{"saas", "dashboard", "productivity", "app"},
		tokens: PaletteTokens{
			Background:  "#ffffff",
			Surface:     "#f8fafc",
			Primary:     "#6366f1",
			PrimarySoft: "#eef2ff",
			Accent:      "#f59e0b",
			Text:        "#1e293b",
			Muted:       "#64748b",
		},
	},
	{
		name:        "Marketing Vibrant",
		description: "Bold and energetic palette for marketing sites and landing pages.",
		mood:        "bold",
		useCaseTags: []string{"marketing", "landing", "startup", "bold"},
		tokens: PaletteTokens{
			Background:  "#fafafa",
			Surface:     "#f4f4f5",
			Primary:     "#ef4444",
			PrimarySoft: "#fef2f2",
			Accent:      "#8b5cf6",
			Text:        "#18181b",
			Muted:       "#71717a",
		},
	},
	{
		name:        "Minimal Light",
		description: "Clean and minimal light palette for content-focused applications.",
		mood:        "minimal",
		useCaseTags: []string{"blog", "content", "minimal", "clean", "editorial"},
		tokens: PaletteTokens{
			Background:  "#ffffff",
			Surface:     "#f9fafb",
			Primary:     "#111827",
			PrimarySoft: "#f3f4f6",
			Accent:      "#3b82f6",
			Text:        "#111827",
			Muted:       "#9ca3af",
		},
	},
	{
		name:        "Dark Premium",
		description: "Elegant dark mode palette for premium products and tools.",
		mood:        "premium",
		useCaseTags: []string{"dark", "premium", "tool", "developer", "ide"},
		tokens: PaletteTokens{
			Background:  "#09090b",
			Surface:     "#18181b",
			Primary:     "#a78bfa",
			PrimarySoft: "#1c1c27",
			Accent:      "#34d399",
			Text:        "#fafafa",
			Muted:       "#71717a",
		},
	},
	{
		name:        "E-commerce Warm",
		description: "Warm and inviting palette for e-commerce and retail applications.",
		mood:        "warm",
		useCaseTags: []string{"ecommerce", "retail", "shop", "marketplace"},
		tokens: PaletteTokens{
			Background:  "#fffbf5",
			Surface:     "#fef3c7",
			Primary:     "#d97706",
			PrimarySoft: "#fef9ee",
			Accent:      "#dc2626",
			Text:        "#292524",
			Muted:       "#78716c",
		},
	},
	{
		name:        "Health & Wellness",
		description: "Calming green palette for health, wellness, and nature-focused apps.",
		mood:        "calm",
		useCaseTags: []string{"health", "wellness", "nature", "fitness", "medical"},
		tokens: PaletteTokens{
			Background:  "#f0fdf4",
			Surface:     "#dcfce7",
			Primary:     "#16a34a",
			PrimarySoft: "#f0fdf4",
			Accent:      "#0891b2",
			Text:        "#14532d",
			Muted:       "#6b7280",
		},
	},
	{
		name:        "Developer Tools",
		description: "High contrast dark palette optimized for developer tools and coding interfaces.",
		mood:        "focused",
		useCaseTags: []string{"developer", "tool", "coding", "technical", "api"},
		tokens: PaletteTokens{
			Background:  "#1a1b26",
			Surface:     "#24283b",
			Primary:     "#7aa2f7",
			PrimarySoft: "#2d3f6c",
			Accent:      "#9ece6a",
			Text:        "#c0caf5",
			Muted:       "#565f89",
		},
	},
}

func generatePalettes(input SuggestColorPalettesInput, count int) []ColorPalette {
	useCase := strings.ToLower(input.UseCase)
	mood := strings.ToLower(input.Mood)

	// Score each palette by relevance
	type scored struct {
		p     paletteTemplate
		score int
	}

	var candidates []scored
	for _, p := range predefinedPalettes {
		score := 0

		// Mood match
		if mood != "" && strings.Contains(p.mood, mood) {
			score += 3
		}

		// Use case keyword match
		for _, tag := range p.useCaseTags {
			if strings.Contains(useCase, tag) || strings.Contains(tag, useCase) {
				score += 2
			}
		}

		// Brand keywords match
		for _, kw := range input.BrandKeywords {
			kw = strings.ToLower(kw)
			for _, tag := range p.useCaseTags {
				if strings.Contains(tag, kw) || strings.Contains(kw, tag) {
					score += 1
				}
			}
		}

		candidates = append(candidates, scored{p: p, score: score})
	}

	// Sort by score (simple bubble sort — small list)
	for i := 0; i < len(candidates)-1; i++ {
		for j := 0; j < len(candidates)-1-i; j++ {
			if candidates[j].score < candidates[j+1].score {
				candidates[j], candidates[j+1] = candidates[j+1], candidates[j]
			}
		}
	}

	var palettes []ColorPalette
	for i, c := range candidates {
		if i >= count {
			break
		}
		desc := c.p.description
		if mood != "" {
			desc = fmt.Sprintf("%s Optimized for %s mood.", desc, input.Mood)
		}
		palettes = append(palettes, ColorPalette{
			Name:        c.p.name,
			Description: desc,
			Tokens:      c.p.tokens,
		})
	}

	// If no good matches, return top-scored ones
	if len(palettes) == 0 {
		for i, c := range predefinedPalettes {
			if i >= count {
				break
			}
			palettes = append(palettes, ColorPalette{
				Name:        c.name,
				Description: c.description,
				Tokens:      c.tokens,
			})
		}
	}

	return palettes
}
