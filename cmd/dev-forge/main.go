// cmd/dev-forge is the CLI/TUI entry point for DevForge.
// It launches an interactive Bubble Tea interface with auto-detected stack settings.
package main

import (
	"encoding/json"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"dev-forge-mcp/internal/config"
	"dev-forge-mcp/internal/db"
	"dev-forge-mcp/internal/imgproc"
	"dev-forge-mcp/internal/tools"
	"dev-forge-mcp/internal/tui"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Printf("warning: could not load config: %v", err)
		cfg = &config.Config{
			OllamaURL:      "http://localhost:11434",
			EmbeddingModel: "nomic-embed-text",
		}
	}

	// Open DB
	if err := os.MkdirAll("db", 0755); err != nil {
		log.Fatalf("failed to create db directory: %v", err)
	}
	database, err := db.Open("file:./db/ui_patterns.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// Initialize embedding client (optional)
	var embedder *db.EmbeddingClient
	if cfg.OllamaURL != "" && db.CheckAvailability(cfg.OllamaURL) {
		embedder = db.NewEmbeddingClient(cfg.OllamaURL, cfg.EmbeddingModel)
	}

	// Initialize imgproc (optional, non-fatal)
	var sc *imgproc.StreamClient
	sc, err = imgproc.NewStreamClient("./bin/devforge-imgproc")
	if err != nil {
		sc = nil
	} else {
		defer sc.Close()
	}

	// Build tools server
	srv := &tools.Server{
		DB:       database,
		Imgproc:  sc,
		Embedder: embedder,
	}

	// Auto-detect stack
	framework, cssMode := detectStack()

	// Launch TUI
	m := tui.New(database, cfg, srv, framework, cssMode)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("TUI error: %v", err)
	}
}

// detectStack scans the current working directory for framework and CSS mode indicators.
func detectStack() (framework, cssMode string) {
	framework = "vanilla"
	cssMode = "plain-css"

	// Check for framework config files
	if fileExists("astro.config.mjs") || fileExists("astro.config.ts") || fileExists("astro.config.js") {
		framework = "astro"
	} else if fileExists("next.config.js") || fileExists("next.config.mjs") || fileExists("next.config.ts") {
		framework = "next"
	} else if fileExists("svelte.config.js") || fileExists("svelte.config.ts") {
		framework = "sveltekit"
	} else if fileExists("nuxt.config.js") || fileExists("nuxt.config.ts") || fileExists("nuxt.config.mjs") {
		framework = "nuxt"
	} else if fileExists("vite.config.js") || fileExists("vite.config.ts") || fileExists("vite.config.mjs") {
		framework = "spa-vite"
	}

	// Check for Tailwind v4
	if fileExists("package.json") {
		if hasTailwindDep("package.json") {
			cssMode = "tailwind-v4"
		}
	}

	return
}

// fileExists reports whether a file exists in the current directory.
func fileExists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

// hasTailwindDep checks if package.json contains a tailwindcss dependency.
func hasTailwindDep(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if json.Unmarshal(data, &pkg) != nil {
		return false
	}
	_, inDeps := pkg.Dependencies["tailwindcss"]
	_, inDevDeps := pkg.DevDependencies["tailwindcss"]
	return inDeps || inDevDeps
}
