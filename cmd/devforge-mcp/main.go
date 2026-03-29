// cmd/dev-forge-mcp is the MCP server entry point.
// It exposes design, image, and pattern tools via the MCP stdio transport.
package main

import (
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"log"
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"dev-forge-mcp/internal/config"
	"dev-forge-mcp/internal/db"
	"dev-forge-mcp/internal/dpf"
	"dev-forge-mcp/internal/tools"
)

// mcpApp holds all server dependencies with hot-reload support.
type mcpApp struct {
	srv        *tools.Server
	mu         sync.RWMutex
	geminiKey  string
	imageModel string
}

func (a *mcpApp) getGeminiKey() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.geminiKey
}

func (a *mcpApp) setGeminiKey(key string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.geminiKey = key
}

func (a *mcpApp) getImageModel() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.imageModel
}

func main() {
	// 1. Load config
	cfg, err := config.Load()
	if err != nil {
		log.Printf("warning: could not load config: %v", err)
		cfg = &config.Config{
			OllamaURL:      "http://localhost:11434",
			EmbeddingModel: "nomic-embed-text",
		}
	}

	// Resolve paths relative to the executable so the server works regardless of CWD.
	exeDir, err := executableDir()
	if err != nil {
		log.Fatalf("failed to resolve executable directory: %v", err)
	}

	// 2. Open DB
	dbPath := "file:" + filepath.Join(exeDir, "dev-forge.db")
	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// 3. Initialize embedding client (test availability with 1s timeout)
	var embedder *db.EmbeddingClient
	if cfg.OllamaURL != "" && db.CheckAvailability(cfg.OllamaURL) {
		embedder = db.NewEmbeddingClient(cfg.OllamaURL, cfg.EmbeddingModel)
		log.Printf("ollama available at %s, embedding enabled", cfg.OllamaURL)
	} else {
		log.Printf("ollama not available — embedding disabled, falling back to FTS5")
	}

	// 4. Initialize StreamClient for dpf (DevPixelForge)
	dpfPath := filepath.Join(exeDir, "dpf")
	var sc *dpf.StreamClient
	sc, err = dpf.NewStreamClient(dpfPath)
	if err != nil {
		log.Printf("warning: dpf binary not available at %s: %v", dpfPath, err)
		log.Printf("optimize_images, generate_favicon, and media tools will return errors")
		sc = nil
	} else {
		defer sc.Close()
	}

	// 5. Build app state
	app := &mcpApp{
		srv: &tools.Server{
			DB:       database,
			DPF:      sc,
			Embedder: embedder,
		},
		geminiKey:  cfg.GeminiAPIKey,
		imageModel: cfg.ImageModel,
	}

	// 6. Launch embedding backfill if Ollama available
	if embedder != nil {
		go backfillEmbeddings(database, embedder)
	}

	// 7. Build MCP server and register all tools
	s := mcpserver.NewMCPServer("dev-forge", "1.0.0",
		mcpserver.WithToolCapabilities(true),
	)

	registerTools(s, app)

	// 8. Serve via stdio transport
	if err := mcpserver.ServeStdio(s); err != nil {
		log.Fatalf("mcp server error: %v", err)
	}
}

// executableDir returns the directory that contains the running binary,
// resolving symlinks so the path is always the real location.
func executableDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return "", err
	}
	return filepath.Dir(resolved), nil
}

func registerTools(s *mcpserver.MCPServer, app *mcpApp) {
	// ── analyze_layout ──────────────────────────────────────────
	s.AddTool(mcp.NewTool("analyze_layout",
		mcp.WithDescription("Audit HTML/JSX markup for layout issues including accessibility, spacing, typography, and framework-specific conventions."),
		mcp.WithString("markup", mcp.Required(), mcp.Description("HTML or JSX string to analyze")),
		mcp.WithObject("stack", mcp.Required(), mcp.Description("CSS stack metadata")),
		mcp.WithString("page_type", mcp.Description("Page type: landing | dashboard | form")),
		mcp.WithString("device_focus", mcp.Description("Device focus: mobile | desktop | both")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := argsMap(req)
		input := tools.AnalyzeLayoutInput{
			Markup:      mcp.ParseString(req, "markup", ""),
			PageType:    mcp.ParseString(req, "page_type", ""),
			DeviceFocus: mcp.ParseString(req, "device_focus", ""),
		}
		if stackMap, ok := args["stack"].(map[string]interface{}); ok {
			input.Stack.CSSMode = strVal(stackMap, "css_mode")
			input.Stack.Framework = strVal(stackMap, "framework")
		}
		return mcp.NewToolResultText(app.srv.AnalyzeLayout(ctx, input)), nil
	})

	// ── suggest_layout ──────────────────────────────────────────
	s.AddTool(mcp.NewTool("suggest_layout",
		mcp.WithDescription("Generate a layout scaffold based on a description and stack metadata."),
		mcp.WithString("description", mcp.Required(), mcp.Description("Layout description")),
		mcp.WithObject("stack", mcp.Required(), mcp.Description("CSS stack metadata")),
		mcp.WithString("fidelity", mcp.Required(), mcp.Description("wireframe | mid | production")),
		mcp.WithObject("tokens_profile", mcp.Description("Existing token values to incorporate")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := argsMap(req)
		input := tools.SuggestLayoutInput{
			Description: mcp.ParseString(req, "description", ""),
			Fidelity:    mcp.ParseString(req, "fidelity", "mid"),
		}
		if stackMap, ok := args["stack"].(map[string]interface{}); ok {
			input.Stack.CSSMode = strVal(stackMap, "css_mode")
			input.Stack.Framework = strVal(stackMap, "framework")
		}
		return mcp.NewToolResultText(app.srv.SuggestLayout(ctx, input)), nil
	})

	// ── manage_tokens ───────────────────────────────────────────
	s.AddTool(mcp.NewTool("manage_tokens",
		mcp.WithDescription("Read or update design tokens (colors, spacing, typography)."),
		mcp.WithString("mode", mcp.Required(), mcp.Description("read | plan-update | apply-update")),
		mcp.WithString("css_mode", mcp.Required(), mcp.Description("tailwind-v4 | plain-css")),
		mcp.WithString("scope", mcp.Required(), mcp.Description("colors | spacing | typography | all")),
		mcp.WithObject("proposal", mcp.Description("Token key/value pairs to apply (required for apply-update)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := argsMap(req)
		input := tools.ManageTokensInput{
			Mode:    mcp.ParseString(req, "mode", ""),
			CSSMode: mcp.ParseString(req, "css_mode", ""),
			Scope:   mcp.ParseString(req, "scope", "all"),
		}
		if proposal, ok := args["proposal"].(map[string]interface{}); ok {
			input.Proposal = make(map[string]interface{})
			for k, v := range proposal {
				input.Proposal[k] = v
			}
		}
		return mcp.NewToolResultText(app.srv.ManageTokens(ctx, input)), nil
	})

	// ── store_pattern ───────────────────────────────────────────
	s.AddTool(mcp.NewTool("store_pattern",
		mcp.WithDescription("Persist a UI layout pattern to the database."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Pattern name")),
		mcp.WithString("framework", mcp.Required(), mcp.Description("Framework: spa-vite | astro | next | sveltekit | nuxt | vanilla")),
		mcp.WithString("css_mode", mcp.Required(), mcp.Description("tailwind-v4 | plain-css")),
		mcp.WithString("snippet", mcp.Required(), mcp.Description("HTML/JSX snippet")),
		mcp.WithString("category", mcp.Description("landing | dashboard | form | component | other")),
		mcp.WithString("domain", mcp.Description("frontend | backend | fullstack | devops | any")),
		mcp.WithString("tags", mcp.Description("Comma-separated tags")),
		mcp.WithString("css_snippet", mcp.Description("Optional CSS snippet")),
		mcp.WithString("description", mcp.Description("Pattern description")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		input := tools.StorePatternInput{
			Name:        mcp.ParseString(req, "name", ""),
			Framework:   mcp.ParseString(req, "framework", ""),
			CSSMode:     mcp.ParseString(req, "css_mode", ""),
			Snippet:     mcp.ParseString(req, "snippet", ""),
			Category:    mcp.ParseString(req, "category", ""),
			Domain:      mcp.ParseString(req, "domain", ""),
			Tags:        mcp.ParseString(req, "tags", ""),
			CSSSnippet:  mcp.ParseString(req, "css_snippet", ""),
			Description: mcp.ParseString(req, "description", ""),
		}
		return mcp.NewToolResultText(app.srv.StorePattern(ctx, input)), nil
	})

	// ── list_patterns ───────────────────────────────────────────
	s.AddTool(mcp.NewTool("list_patterns",
		mcp.WithDescription("Query stored patterns with optional filters, full-text search, and semantic similarity."),
		mcp.WithString("domain", mcp.Description("frontend | backend | fullstack | devops | any")),
		mcp.WithString("css_mode", mcp.Description("tailwind-v4 | plain-css")),
		mcp.WithString("framework", mcp.Description("Framework filter")),
		mcp.WithString("query", mcp.Description("Keyword or natural language query")),
		mcp.WithString("mode", mcp.Description("fts | semantic | filter (default: auto)")),
		mcp.WithNumber("limit", mcp.Description("Max results (default 20)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		input := tools.ListPatternsInput{
			Domain:    mcp.ParseString(req, "domain", ""),
			CSSMode:   mcp.ParseString(req, "css_mode", ""),
			Framework: mcp.ParseString(req, "framework", ""),
			Query:     mcp.ParseString(req, "query", ""),
			Mode:      mcp.ParseString(req, "mode", ""),
			Limit:     mcp.ParseInt(req, "limit", 20),
		}
		return mcp.NewToolResultText(app.srv.ListPatterns(ctx, input)), nil
	})

	// ── generate_ui_image ────────────────────────────────────────
	s.AddTool(mcp.NewTool("generate_ui_image",
		mcp.WithDescription("Generate a UI image via Gemini API. Requires gemini_api_key to be configured."),
		mcp.WithString("prompt", mcp.Required(), mcp.Description("Image generation prompt")),
		mcp.WithString("style", mcp.Required(), mcp.Description("wireframe | mockup | illustration")),
		mcp.WithNumber("width", mcp.Description("Image width in pixels (default 1280)")),
		mcp.WithNumber("height", mcp.Description("Image height in pixels (default 720)")),
		mcp.WithString("output_path", mcp.Required(), mcp.Description("File path to save the generated image")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		input := tools.GenerateUIImageInput{
			Prompt:     mcp.ParseString(req, "prompt", ""),
			Style:      mcp.ParseString(req, "style", "mockup"),
			Width:      mcp.ParseInt(req, "width", 1280),
			Height:     mcp.ParseInt(req, "height", 720),
			OutputPath: mcp.ParseString(req, "output_path", ""),
		}
		return mcp.NewToolResultText(app.srv.GenerateUIImage(ctx, input, app.getGeminiKey(), app.getImageModel())), nil
	})

	// ── ui2md ────────────────────────────────────────────────────
	s.AddTool(mcp.NewTool("ui2md",
		mcp.WithDescription("Analyze a UI screenshot and generate a Markdown design spec using Gemini vision."),
		mcp.WithString("image_path", mcp.Required(), mcp.Description("Path to the UI image to analyze")),
		mcp.WithString("output_dir", mcp.Description("Directory to save the generated markdown (default: same as image)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		input := tools.UI2MDInput{
			ImagePath: mcp.ParseString(req, "image_path", ""),
			OutputDir: mcp.ParseString(req, "output_dir", ""),
		}
		return mcp.NewToolResultText(app.srv.UI2MD(ctx, input, app.getGeminiKey(), app.getImageModel())), nil
	})

	// ── configure_gemini ────────────────────────────────────────
	s.AddTool(mcp.NewTool("configure_gemini",
		mcp.WithDescription("Save Gemini API key to config file and hot-reload without restart."),
		mcp.WithString("api_key", mcp.Required(), mcp.Description("Gemini API key")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		input := tools.ConfigureGeminiInput{
			APIKey: mcp.ParseString(req, "api_key", ""),
		}
		result := app.srv.ConfigureGemini(ctx, input, app.setGeminiKey)
		return mcp.NewToolResultText(result), nil
	})

	// ── optimize_images ─────────────────────────────────────────
	s.AddTool(mcp.NewTool("optimize_images",
		mcp.WithDescription("Optimize and convert images using the Rust dpf (DevPixelForge) engine."),
		mcp.WithArray("inputs", mcp.Required(), mcp.Description("Array of image optimization requests"),
			mcp.Items(map[string]any{"type": "object"})),
		mcp.WithNumber("parallelism", mcp.Description("Max parallel operations (default 4)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := argsMap(req)
		var optInputs []tools.OptimizeInput
		if inputsRaw, ok := args["inputs"]; ok {
			data, _ := json.Marshal(inputsRaw)
			json.Unmarshal(data, &optInputs)
		}
		input := tools.OptimizeImagesInput{
			Inputs:      optInputs,
			Parallelism: mcp.ParseInt(req, "parallelism", 4),
		}
		return mcp.NewToolResultText(app.srv.OptimizeImages(ctx, input)), nil
	})

	// ── generate_favicon ────────────────────────────────────────
	s.AddTool(mcp.NewTool("generate_favicon",
		mcp.WithDescription("Generate favicon variants (ico, png, svg) from a source image."),
		mcp.WithString("source_path", mcp.Required(), mcp.Description("Path to source image (PNG or SVG)")),
		mcp.WithString("background_color", mcp.Description("Hex background color (default #ffffff)")),
		mcp.WithArray("sizes", mcp.Description("Icon sizes (default [16,32,48,180,192,512])"), mcp.WithNumberItems()),
		mcp.WithArray("formats", mcp.Description("Output formats: ico | png | svg"), mcp.WithStringItems()),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := argsMap(req)
		input := tools.GenerateFaviconInput{
			SourcePath:      mcp.ParseString(req, "source_path", ""),
			BackgroundColor: mcp.ParseString(req, "background_color", "#ffffff"),
		}
		if sizesRaw, ok := args["sizes"]; ok {
			data, _ := json.Marshal(sizesRaw)
			var sizesF []float64
			if json.Unmarshal(data, &sizesF) == nil {
				for _, f := range sizesF {
					input.Sizes = append(input.Sizes, int(f))
				}
			}
		}
		if formatsRaw, ok := args["formats"]; ok {
			data, _ := json.Marshal(formatsRaw)
			json.Unmarshal(data, &input.Formats)
		}
		return mcp.NewToolResultText(app.srv.GenerateFavicon(ctx, input)), nil
	})

	// ── suggest_color_palettes ──────────────────────────────────
	s.AddTool(mcp.NewTool("suggest_color_palettes",
		mcp.WithDescription("Generate named color palette proposals for a given use case and mood."),
		mcp.WithString("use_case", mcp.Required(), mcp.Description("e.g. 'SaaS dashboard', 'marketing site'")),
		mcp.WithArray("brand_keywords", mcp.Description("Brand keyword list"), mcp.WithStringItems()),
		mcp.WithString("mood", mcp.Description("e.g. calm | bold | minimal | professional")),
		mcp.WithNumber("count", mcp.Description("Number of palettes to generate (default 3)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := argsMap(req)
		input := tools.SuggestColorPalettesInput{
			UseCase: mcp.ParseString(req, "use_case", ""),
			Mood:    mcp.ParseString(req, "mood", ""),
			Count:   mcp.ParseInt(req, "count", 3),
		}
		if kwRaw, ok := args["brand_keywords"]; ok {
			data, _ := json.Marshal(kwRaw)
			json.Unmarshal(data, &input.BrandKeywords)
		}
		return mcp.NewToolResultText(app.srv.SuggestColorPalettes(ctx, input)), nil
	})

	// ── Video Tools ────────────────────────────────────────────────

	// video_transcode
	s.AddTool(mcp.NewTool("video_transcode",
		mcp.WithDescription("Transcode video to different codec (h264, h265, vp8, vp9, av1)."),
		mcp.WithString("input", mcp.Required(), mcp.Description("Input video path")),
		mcp.WithString("output", mcp.Required(), mcp.Description("Output video path")),
		mcp.WithString("codec", mcp.Required(), mcp.Description("Target codec: h264, h265, vp8, vp9, av1")),
		mcp.WithString("bitrate", mcp.Description("Target bitrate (e.g., '2M', '5000k')")),
		mcp.WithString("preset", mcp.Description("Encoding preset: ultrafast, fast, medium, slow, veryslow")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		input := tools.VideoTranscodeInput{
			Input:   mcp.ParseString(req, "input", ""),
			Output:  mcp.ParseString(req, "output", ""),
			Codec:   mcp.ParseString(req, "codec", ""),
			Bitrate: mcp.ParseString(req, "bitrate", ""),
			Preset:  mcp.ParseString(req, "preset", ""),
		}
		return mcp.NewToolResultText(app.srv.VideoTranscode(ctx, input)), nil
	})

	// video_resize
	s.AddTool(mcp.NewTool("video_resize",
		mcp.WithDescription("Resize video while maintaining aspect ratio."),
		mcp.WithString("input", mcp.Required(), mcp.Description("Input video path")),
		mcp.WithString("output", mcp.Required(), mcp.Description("Output video path")),
		mcp.WithNumber("width", mcp.Description("Target width")),
		mcp.WithNumber("height", mcp.Description("Target height")),
		mcp.WithBoolean("maintain_aspect", mcp.Description("Maintain aspect ratio (default true)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		input := tools.VideoResizeInput{
			Input:          mcp.ParseString(req, "input", ""),
			Output:         mcp.ParseString(req, "output", ""),
			Width:          uint32(mcp.ParseInt(req, "width", 0)),
			Height:         uint32(mcp.ParseInt(req, "height", 0)),
			MaintainAspect: mcp.ParseBoolean(req, "maintain_aspect", true),
		}
		return mcp.NewToolResultText(app.srv.VideoResize(ctx, input)), nil
	})

	// video_trim
	s.AddTool(mcp.NewTool("video_trim",
		mcp.WithDescription("Extract a time range from video."),
		mcp.WithString("input", mcp.Required(), mcp.Description("Input video path")),
		mcp.WithString("output", mcp.Required(), mcp.Description("Output video path")),
		mcp.WithNumber("start", mcp.Required(), mcp.Description("Start time in seconds")),
		mcp.WithNumber("end", mcp.Required(), mcp.Description("End time in seconds")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := argsMap(req)
		input := tools.VideoTrimInput{
			Input:  mcp.ParseString(req, "input", ""),
			Output: mcp.ParseString(req, "output", ""),
			Start:  numVal(args, "start", 0),
			End:    numVal(args, "end", 0),
		}
		return mcp.NewToolResultText(app.srv.VideoTrim(ctx, input)), nil
	})

	// video_thumbnail
	s.AddTool(mcp.NewTool("video_thumbnail",
		mcp.WithDescription("Extract a frame as image from video."),
		mcp.WithString("input", mcp.Required(), mcp.Description("Input video path")),
		mcp.WithString("output", mcp.Required(), mcp.Description("Output image path")),
		mcp.WithString("timestamp", mcp.Required(), mcp.Description("Timestamp: '25%' or seconds like '30.5'")),
		mcp.WithString("format", mcp.Description("Output format: jpeg, png, webp (default jpeg)")),
		mcp.WithNumber("quality", mcp.Description("Image quality 1-100 (default 85)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		input := tools.VideoThumbnailInput{
			Input:     mcp.ParseString(req, "input", ""),
			Output:    mcp.ParseString(req, "output", ""),
			Timestamp: mcp.ParseString(req, "timestamp", ""),
			Format:    mcp.ParseString(req, "format", "jpeg"),
			Quality:   mcp.ParseInt(req, "quality", 85),
		}
		return mcp.NewToolResultText(app.srv.VideoThumbnail(ctx, input)), nil
	})

	// video_profile
	s.AddTool(mcp.NewTool("video_profile",
		mcp.WithDescription("Apply web-optimized encoding profile."),
		mcp.WithString("input", mcp.Required(), mcp.Description("Input video path")),
		mcp.WithString("output", mcp.Required(), mcp.Description("Output video path")),
		mcp.WithString("profile", mcp.Required(), mcp.Description("Profile: web-low (480p/1M), web-mid (720p/2.5M), web-high (1080p/5M)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		input := tools.VideoProfileInput{
			Input:   mcp.ParseString(req, "input", ""),
			Output:  mcp.ParseString(req, "output", ""),
			Profile: mcp.ParseString(req, "profile", ""),
		}
		return mcp.NewToolResultText(app.srv.VideoProfile(ctx, input)), nil
	})

	// ── Audio Tools ────────────────────────────────────────────────

	// audio_transcode
	s.AddTool(mcp.NewTool("audio_transcode",
		mcp.WithDescription("Convert between audio formats (mp3, aac, opus, vorbis, flac, wav)."),
		mcp.WithString("input", mcp.Required(), mcp.Description("Input audio path")),
		mcp.WithString("output", mcp.Required(), mcp.Description("Output audio path")),
		mcp.WithString("codec", mcp.Required(), mcp.Description("Target codec: mp3, aac, opus, vorbis, flac, wav")),
		mcp.WithString("bitrate", mcp.Description("Target bitrate (e.g., '192k')")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		input := tools.AudioTranscodeInput{
			Input:   mcp.ParseString(req, "input", ""),
			Output:  mcp.ParseString(req, "output", ""),
			Codec:   mcp.ParseString(req, "codec", ""),
			Bitrate: mcp.ParseString(req, "bitrate", ""),
		}
		return mcp.NewToolResultText(app.srv.AudioTranscode(ctx, input)), nil
	})

	// audio_trim
	s.AddTool(mcp.NewTool("audio_trim",
		mcp.WithDescription("Trim audio by timestamps."),
		mcp.WithString("input", mcp.Required(), mcp.Description("Input audio path")),
		mcp.WithString("output", mcp.Required(), mcp.Description("Output audio path")),
		mcp.WithNumber("start", mcp.Required(), mcp.Description("Start time in seconds")),
		mcp.WithNumber("end", mcp.Required(), mcp.Description("End time in seconds")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := argsMap(req)
		input := tools.AudioTrimInput{
			Input:  mcp.ParseString(req, "input", ""),
			Output: mcp.ParseString(req, "output", ""),
			Start:  numVal(args, "start", 0),
			End:    numVal(args, "end", 0),
		}
		return mcp.NewToolResultText(app.srv.AudioTrim(ctx, input)), nil
	})

	// audio_normalize
	s.AddTool(mcp.NewTool("audio_normalize",
		mcp.WithDescription("Normalize loudness to target LUFS."),
		mcp.WithString("input", mcp.Required(), mcp.Description("Input audio path")),
		mcp.WithString("output", mcp.Required(), mcp.Description("Output audio path")),
		mcp.WithNumber("target_lufs", mcp.Required(), mcp.Description("Target LUFS (-14 for YouTube, -16 for Spotify, -23 for EBU R128)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := argsMap(req)
		input := tools.AudioNormalizeInput{
			Input:      mcp.ParseString(req, "input", ""),
			Output:     mcp.ParseString(req, "output", ""),
			TargetLUFS: numVal(args, "target_lufs", -14),
		}
		return mcp.NewToolResultText(app.srv.AudioNormalize(ctx, input)), nil
	})

	// audio_silence_trim
	s.AddTool(mcp.NewTool("audio_silence_trim",
		mcp.WithDescription("Remove leading/trailing silence from audio."),
		mcp.WithString("input", mcp.Required(), mcp.Description("Input audio path")),
		mcp.WithString("output", mcp.Required(), mcp.Description("Output audio path")),
		mcp.WithNumber("threshold_db", mcp.Description("Silence threshold in dB (default -40)")),
		mcp.WithNumber("min_duration", mcp.Description("Minimum silence duration in seconds (default 0.5)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := argsMap(req)
		input := tools.AudioSilenceTrimInput{
			Input:       mcp.ParseString(req, "input", ""),
			Output:      mcp.ParseString(req, "output", ""),
			ThresholdDB: numVal(args, "threshold_db", -40),
			MinDuration: numVal(args, "min_duration", 0.5),
		}
		return mcp.NewToolResultText(app.srv.AudioSilenceTrim(ctx, input)), nil
	})
}

// backfillEmbeddings populates NULL embeddings for patterns and architectures in background.
func backfillEmbeddings(database *sql.DB, embedder *db.EmbeddingClient) {
	tables := []struct {
		table string
		text  string
	}{
		{"patterns", "name || ' ' || COALESCE(description,'') || ' ' || COALESCE(tags,'')"},
		{"architectures", "name || ' ' || COALESCE(description,'') || ' ' || COALESCE(tags,'')"},
	}

	sem := make(chan struct{}, 4) // max 4 parallel

	for _, t := range tables {
		rows, err := database.Query(
			"SELECT id, " + t.text + " AS text FROM " + t.table + " WHERE embedding IS NULL LIMIT 100",
		)
		if err != nil {
			continue
		}
		type row struct{ id, text string }
		var pending []row
		for rows.Next() {
			var r row
			if rows.Scan(&r.id, &r.text) == nil {
				pending = append(pending, r)
			}
		}
		rows.Close()

		for _, r := range pending {
			r := r
			tbl := t.table
			sem <- struct{}{}
			go func() {
				defer func() { <-sem }()
				vec, err := embedder.Embed(context.Background(), r.text)
				if err != nil || vec == nil {
					return
				}
				// Encode float32 slice to little-endian bytes
				buf := make([]byte, len(vec)*4)
				for i, f := range vec {
					binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(f))
				}
				database.Exec("UPDATE "+tbl+" SET embedding = ? WHERE id = ?", buf, r.id)
			}()
		}
	}
}

// argsMap safely extracts the arguments map from a CallToolRequest.
func argsMap(req mcp.CallToolRequest) map[string]interface{} {
	if m, ok := req.Params.Arguments.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}

// strVal extracts a string from a map[string]any.
func strVal(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// numVal extracts a number as float64 from a map[string]any.
func numVal(m map[string]interface{}, key string, fallback float64) float64 {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			return n
		case float32:
			return float64(n)
		case int:
			return float64(n)
		case int64:
			return float64(n)
		}
	}
	return fallback
}
