package tools_test

import (
	"context"
	"encoding/json"
	"testing"

	"dev-forge-mcp/internal/testutil"
	"dev-forge-mcp/internal/tools"
)

func TestAnalyzeLayout_ValidMarkup_ReturnsIssues(t *testing.T) {
	database := testutil.NewTestDB(t)
	srv := &tools.Server{DB: database}

	markup := `<div><img src="logo.png"><p style="font-size: 14px">Hello</p></div>`
	input := tools.AnalyzeLayoutInput{
		Markup:   markup,
		Stack:    tools.StackMeta{CSSMode: "tailwind-v4", Framework: "astro"},
		PageType: "landing",
	}

	result := srv.AnalyzeLayout(context.Background(), input)

	var out tools.AnalyzeLayoutOutput
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		t.Fatalf("invalid JSON response: %v — got: %s", err, result)
	}
	if out.Score < 0 || out.Score > 100 {
		t.Errorf("score out of range: %d", out.Score)
	}
	if out.Summary == "" {
		t.Error("expected non-empty summary")
	}
}

func TestAnalyzeLayout_MissingAltTag_ReportsError(t *testing.T) {
	database := testutil.NewTestDB(t)
	srv := &tools.Server{DB: database}

	markup := `<section><img src="banner.png"><h1>Welcome</h1></section>`
	input := tools.AnalyzeLayoutInput{
		Markup: markup,
		Stack:  tools.StackMeta{CSSMode: "plain-css", Framework: "next"},
	}

	result := srv.AnalyzeLayout(context.Background(), input)

	var out tools.AnalyzeLayoutOutput
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	found := false
	for _, issue := range out.Issues {
		if issue.Severity == "error" && issue.Category == "accessibility" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected accessibility error for missing alt tag")
	}
}

func TestAnalyzeLayout_MissingMarkup_ReturnsError(t *testing.T) {
	database := testutil.NewTestDB(t)
	srv := &tools.Server{DB: database}

	input := tools.AnalyzeLayoutInput{
		Markup: "",
		Stack:  tools.StackMeta{CSSMode: "tailwind-v4", Framework: "astro"},
	}

	result := srv.AnalyzeLayout(context.Background(), input)

	var errOut map[string]string
	if err := json.Unmarshal([]byte(result), &errOut); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := errOut["error"]; !ok {
		t.Errorf("expected error response, got: %s", result)
	}
}

func TestAnalyzeLayout_StoresAuditInDB(t *testing.T) {
	database := testutil.NewTestDB(t)
	srv := &tools.Server{DB: database}

	input := tools.AnalyzeLayoutInput{
		Markup:   `<main><h1>Hello</h1></main>`,
		Stack:    tools.StackMeta{CSSMode: "plain-css", Framework: "vanilla"},
		PageType: "landing",
	}
	srv.AnalyzeLayout(context.Background(), input)

	var count int
	database.QueryRow("SELECT COUNT(*) FROM audits").Scan(&count)
	if count == 0 {
		t.Error("expected audit record in DB after analyze_layout")
	}
}
