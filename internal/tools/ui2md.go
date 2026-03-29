package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/genai"
	"gopkg.in/yaml.v3"
)

// UI2MDInput is the input schema for the ui2md tool.
type UI2MDInput struct {
	ImagePath string `json:"image_path"`
	OutputDir string `json:"output_dir,omitempty"`
}

// UI2MDOutput is the output schema for the ui2md tool.
type UI2MDOutput struct {
	FilePath string `json:"file_path"`
	Name     string `json:"name"`
}

const ui2mdPrompt = `Actúa como un ingeniero frontend experto en UI/UX. Analiza la imagen de la interfaz de usuario adjunta y proporciona una descripción técnica detallada para replicar su diseño de forma exacta (estructurado para ser implementado fácilmente con CSS moderno o frameworks de utilidades como Tailwind).
La descripción debe ser concisa, estructurada en Markdown y utilizando terminología estándar (colores HEX, flexbox/grid, medidas relativas, sombras, border-radius) para optimizar el consumo de tokens sin perder fidelidad visual.

**Regla crítica:** Tu respuesta DEBE ser única y exclusivamente texto plano en formato YAML válido. No incluyas saludos, explicaciones, formato de bloque de código markdown (como ` + "```yaml" + ` o ` + "```" + `), ni ningún texto adicional antes o después de las claves.

Utiliza exactamente esta estructura, reemplazando los corchetes y su contenido con los datos generados:

filename: [nombre-del-diseño-en-kebab-case]-[unix-timestamp-actual].md
name: [Nombre descriptivo y legible del diseño]
description: |
  # Descripción Técnica del Diseño UI
  - [Breve descripción general del diseño, su estilo y propósito]
  
  # Layout y Estructura
  - [Define contenedores, grid/flex, alineaciones y espaciados principales]

  # Paleta de Colores
  - [Fondos, textos, colores primarios/secundarios y bordes en HEX]

  # Tipografía
  - [Jerarquía de tamaños, pesos (font-weight) y familias inferidas]

  # Componentes y Estilos
  - [Detalles de inputs, botones, cards: border-radius, box-shadow, estados hover/focus]`

// ui2mdResponse is used to parse the YAML response from Gemini.
type ui2mdResponse struct {
	Filename    string `yaml:"filename"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// UI2MD implements the ui2md MCP tool.
func (s *Server) UI2MD(ctx context.Context, input UI2MDInput, geminiAPIKey string, imageModel string) string {
	if geminiAPIKey == "" {
		return errorJSON("Gemini API key not configured. Use configure_gemini to set it.")
	}
	if imageModel == "" {
		imageModel = "gemini-2.5-flash-preview-05-20"
	}
	if strings.TrimSpace(input.ImagePath) == "" {
		return errorJSON("image_path is required")
	}

	// Read image bytes
	imageBytes, err := os.ReadFile(input.ImagePath)
	if err != nil {
		return errorJSON(fmt.Sprintf("failed to read image file: %s", err.Error()))
	}

	mimeType := detectMIMEType(input.ImagePath)

	// Determine output directory
	outputDir := input.OutputDir
	if strings.TrimSpace(outputDir) == "" {
		outputDir = filepath.Dir(input.ImagePath)
	}

	// Create Gemini client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  geminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return errorJSON("failed to create Gemini client: " + err.Error())
	}

	// Build content parts with both text and image
	contents := []*genai.Content{
		{
			Parts: []*genai.Part{
				{Text: ui2mdPrompt},
				{InlineData: &genai.Blob{MIMEType: mimeType, Data: imageBytes}},
			},
		},
	}

	result, err := client.Models.GenerateContent(ctx, imageModel, contents, nil)
	if err != nil {
		return errorJSON("Gemini API error: " + err.Error())
	}

	if len(result.Candidates) == 0 || result.Candidates[0].Content == nil || len(result.Candidates[0].Content.Parts) == 0 {
		return errorJSON("no text response from Gemini")
	}

	responseText := result.Candidates[0].Content.Parts[0].Text
	if strings.TrimSpace(responseText) == "" {
		return errorJSON("empty response from Gemini")
	}

	// Parse YAML response
	var parsed ui2mdResponse
	if err := yaml.Unmarshal([]byte(responseText), &parsed); err != nil {
		return errorJSON("failed to parse Gemini YAML response: " + err.Error())
	}

	if strings.TrimSpace(parsed.Filename) == "" {
		return errorJSON("Gemini response missing 'filename' field")
	}
	if strings.TrimSpace(parsed.Description) == "" {
		return errorJSON("Gemini response missing 'description' field")
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return errorJSON("failed to create output directory: " + err.Error())
	}

	outputPath := filepath.Join(outputDir, parsed.Filename)
	if err := os.WriteFile(outputPath, []byte(parsed.Description), 0644); err != nil {
		return errorJSON("failed to write markdown file: " + err.Error())
	}

	return mustJSON(UI2MDOutput{
		FilePath: outputPath,
		Name:     parsed.Name,
	})
}

// detectMIMEType returns the MIME type for the given image file path.
func detectMIMEType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "image/png"
	}
}
