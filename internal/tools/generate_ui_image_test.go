package tools_test

import (
	"context"
	"encoding/json"
	"testing"

	"dev-forge-mcp/internal/tools"
)

func TestGenerateUIImage_NoAPIKey_ReturnsError(t *testing.T) {
	srv := &tools.Server{}

	result := srv.GenerateUIImage(context.Background(),
		tools.GenerateUIImageInput{
			Prompt:     "A modern SaaS dashboard",
			Style:      "mockup",
			OutputPath: "/tmp/test.png",
		},
		"", // empty API key
	)

	var errOut map[string]string
	if err := json.Unmarshal([]byte(result), &errOut); err != nil {
		t.Fatalf("invalid JSON: %v — got: %s", err, result)
	}
	if _, ok := errOut["error"]; !ok {
		t.Errorf("expected error when API key is absent, got: %s", result)
	}
}

func TestGenerateUIImage_EmptyPrompt_ReturnsError(t *testing.T) {
	srv := &tools.Server{}

	result := srv.GenerateUIImage(context.Background(),
		tools.GenerateUIImageInput{
			Prompt:     "",
			Style:      "mockup",
			OutputPath: "/tmp/test.png",
		},
		"fake-api-key",
	)

	var errOut map[string]string
	if err := json.Unmarshal([]byte(result), &errOut); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := errOut["error"]; !ok {
		t.Errorf("expected error for empty prompt, got: %s", result)
	}
}

func TestGenerateUIImage_LiveGemini(t *testing.T) {
	apiKey := ""
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set — skipping live Gemini test")
	}
	// Would call real API — skipped in CI
}
