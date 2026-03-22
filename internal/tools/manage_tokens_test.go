package tools_test

import (
	"context"
	"encoding/json"
	"testing"

	"dev-forge-mcp/internal/testutil"
	"dev-forge-mcp/internal/tools"
)

func TestManageTokens_ReadMode_ReturnsCurrentTokens(t *testing.T) {
	database := testutil.NewTestDB(t)
	srv := &tools.Server{DB: database}

	input := tools.ManageTokensInput{
		Mode:    "read",
		CSSMode: "tailwind-v4",
		Scope:   "colors",
	}
	result := srv.ManageTokens(context.Background(), input)

	var out tools.ManageTokensOutput
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		t.Fatalf("invalid JSON: %v — got: %s", err, result)
	}
	if len(out.CurrentTokens) == 0 {
		t.Error("expected non-empty current tokens in read mode")
	}
	if out.Instructions == "" {
		t.Error("expected non-empty instructions")
	}
}

func TestManageTokens_ApplyUpdate_WritesToDB(t *testing.T) {
	database := testutil.NewTestDB(t)
	srv := &tools.Server{DB: database}

	input := tools.ManageTokensInput{
		Mode:    "apply-update",
		CSSMode: "tailwind-v4",
		Scope:   "colors",
		Proposal: map[string]interface{}{
			"--color-primary": "#ff0000",
		},
	}
	result := srv.ManageTokens(context.Background(), input)

	var out tools.ManageTokensOutput
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Check that the token is in the DB
	var count int
	database.QueryRow(`SELECT COUNT(*) FROM tokens WHERE key = '--color-primary'`).Scan(&count)
	if count == 0 {
		t.Error("expected token record in DB after apply-update")
	}
}

func TestManageTokens_PlanUpdate_DoesNotWriteToDB(t *testing.T) {
	database := testutil.NewTestDB(t)
	srv := &tools.Server{DB: database}

	input := tools.ManageTokensInput{
		Mode:    "plan-update",
		CSSMode: "plain-css",
		Scope:   "colors",
		Proposal: map[string]interface{}{
			"--color-accent": "#00ff00",
		},
	}
	result := srv.ManageTokens(context.Background(), input)

	var out tools.ManageTokensOutput
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// No record should be in DB for plan-update
	var count int
	database.QueryRow(`SELECT COUNT(*) FROM tokens WHERE key = '--color-accent'`).Scan(&count)
	if count != 0 {
		t.Error("plan-update should not write to DB")
	}
	if len(out.Diff) == 0 {
		t.Error("expected non-empty diff for plan-update")
	}
}

func TestManageTokens_InvalidMode_ReturnsError(t *testing.T) {
	database := testutil.NewTestDB(t)
	srv := &tools.Server{DB: database}

	input := tools.ManageTokensInput{
		Mode:    "invalid-mode",
		CSSMode: "tailwind-v4",
		Scope:   "colors",
	}
	result := srv.ManageTokens(context.Background(), input)

	var errOut map[string]string
	if err := json.Unmarshal([]byte(result), &errOut); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := errOut["error"]; !ok {
		t.Errorf("expected error response for invalid mode, got: %s", result)
	}
}

func TestManageTokens_Tailwind_InstructionsContainAtTheme(t *testing.T) {
	database := testutil.NewTestDB(t)
	srv := &tools.Server{DB: database}

	input := tools.ManageTokensInput{
		Mode:    "read",
		CSSMode: "tailwind-v4",
		Scope:   "all",
	}
	result := srv.ManageTokens(context.Background(), input)

	var out tools.ManageTokensOutput
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if out.Instructions == "" {
		t.Error("expected instructions")
	}
	// Tailwind v4 should mention @theme
	found := false
	for _, c := range []string{"@theme", "@layer"} {
		if contains(out.Instructions, c) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected @theme or @layer in tailwind-v4 instructions, got: %s", out.Instructions)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
