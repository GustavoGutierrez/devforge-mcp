package tools_test

import (
	"context"
	"encoding/json"
	"testing"

	"dev-forge-mcp/internal/tools"
)

func TestSuggestColorPalettes_ReturnsAllSevenTokenKeys(t *testing.T) {
	srv := &tools.Server{}

	result := srv.SuggestColorPalettes(context.Background(), tools.SuggestColorPalettesInput{
		UseCase: "SaaS dashboard",
		Count:   1,
	})

	var out tools.SuggestColorPalettesOutput
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		t.Fatalf("invalid JSON: %v — got: %s", err, result)
	}
	if len(out.Palettes) == 0 {
		t.Fatal("expected at least 1 palette")
	}

	p := out.Palettes[0]
	requiredKeys := map[string]string{
		"background":   p.Tokens.Background,
		"surface":      p.Tokens.Surface,
		"primary":      p.Tokens.Primary,
		"primary-soft": p.Tokens.PrimarySoft,
		"accent":       p.Tokens.Accent,
		"text":         p.Tokens.Text,
		"muted":        p.Tokens.Muted,
	}
	for key, val := range requiredKeys {
		if val == "" {
			t.Errorf("expected non-empty token %q in palette", key)
		}
	}
}

func TestSuggestColorPalettes_CountParamRespected(t *testing.T) {
	srv := &tools.Server{}

	for _, count := range []int{1, 2, 3} {
		result := srv.SuggestColorPalettes(context.Background(), tools.SuggestColorPalettesInput{
			UseCase: "marketing site",
			Count:   count,
		})

		var out tools.SuggestColorPalettesOutput
		if err := json.Unmarshal([]byte(result), &out); err != nil {
			t.Fatalf("invalid JSON for count=%d: %v", count, err)
		}
		if len(out.Palettes) != count {
			t.Errorf("count=%d: expected %d palettes, got %d", count, count, len(out.Palettes))
		}
	}
}

func TestSuggestColorPalettes_MissingUseCase_ReturnsError(t *testing.T) {
	srv := &tools.Server{}

	result := srv.SuggestColorPalettes(context.Background(), tools.SuggestColorPalettesInput{
		UseCase: "",
	})

	var errOut map[string]string
	if err := json.Unmarshal([]byte(result), &errOut); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := errOut["error"]; !ok {
		t.Errorf("expected error for empty use_case, got: %s", result)
	}
}

func TestSuggestColorPalettes_MoodFiltering(t *testing.T) {
	srv := &tools.Server{}

	result := srv.SuggestColorPalettes(context.Background(), tools.SuggestColorPalettesInput{
		UseCase: "developer tools",
		Mood:    "dark",
		Count:   2,
	})

	var out tools.SuggestColorPalettesOutput
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out.Palettes) == 0 {
		t.Error("expected at least 1 palette for dark mood")
	}
}
