package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/genai"
)

// GenerateUIImageInput is the input schema for the generate_ui_image tool.
type GenerateUIImageInput struct {
	Prompt     string `json:"prompt"`
	Style      string `json:"style"`       // wireframe | mockup | illustration
	Width      int    `json:"width"`       // default 1280
	Height     int    `json:"height"`      // default 720
	OutputPath string `json:"output_path"` // file path to save
}

// GenerateUIImageOutput is the output schema for the generate_ui_image tool.
type GenerateUIImageOutput struct {
	Path       string `json:"path"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	PromptUsed string `json:"prompt_used"`
}

// GenerateUIImage implements the generate_ui_image MCP tool.
func (s *Server) GenerateUIImage(ctx context.Context, input GenerateUIImageInput, geminiAPIKey string) string {
	if geminiAPIKey == "" {
		return errorJSON("Gemini API key not configured. Use configure_gemini to set it.")
	}
	if strings.TrimSpace(input.Prompt) == "" {
		return errorJSON("prompt is required")
	}
	if strings.TrimSpace(input.OutputPath) == "" {
		return errorJSON("output_path is required")
	}

	width := input.Width
	if width <= 0 {
		width = 1280
	}
	height := input.Height
	if height <= 0 {
		height = 720
	}
	style := input.Style
	if style == "" {
		style = "mockup"
	}

	// Build styled prompt
	stylePrefix := stylePromptPrefix(style)
	fullPrompt := fmt.Sprintf("%s %s. Size: %dx%d pixels. Clean UI design.", stylePrefix, input.Prompt, width, height)

	// Call Gemini API
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  geminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return errorJSON("failed to create Gemini client: " + err.Error())
	}

	result, err := client.Models.GenerateContent(ctx,
		"gemini-2.0-flash-preview-image-generation",
		genai.Text(fullPrompt),
		&genai.GenerateContentConfig{
			ResponseModalities: []string{"IMAGE", "TEXT"},
		},
	)
	if err != nil {
		return errorJSON("Gemini API error: " + err.Error())
	}

	// Extract image data — InlineData.Data is already raw bytes (not base64)
	var imageData []byte
	for _, candidate := range result.Candidates {
		if candidate.Content == nil {
			continue
		}
		for _, part := range candidate.Content.Parts {
			if part.InlineData != nil && strings.HasPrefix(part.InlineData.MIMEType, "image/") {
				if len(part.InlineData.Data) > 0 {
					imageData = part.InlineData.Data
					break
				}
			}
		}
		if len(imageData) > 0 {
			break
		}
	}

	if len(imageData) == 0 {
		return errorJSON("no image data in Gemini response")
	}

	// Ensure output directory exists
	outDir := filepath.Dir(input.OutputPath)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return errorJSON("failed to create output directory: " + err.Error())
	}

	// Save image
	if err := os.WriteFile(input.OutputPath, imageData, 0644); err != nil {
		return errorJSON("failed to save image: " + err.Error())
	}

	return mustJSON(GenerateUIImageOutput{
		Path:       input.OutputPath,
		Width:      width,
		Height:     height,
		PromptUsed: fullPrompt,
	})
}

func stylePromptPrefix(style string) string {
	switch style {
	case "wireframe":
		return "Create a clean wireframe UI sketch of:"
	case "illustration":
		return "Create a colorful UI illustration of:"
	default: // mockup
		return "Create a high-fidelity UI mockup of:"
	}
}
