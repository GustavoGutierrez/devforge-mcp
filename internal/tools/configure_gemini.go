package tools

import (
	"context"
	"strings"

	"dev-forge-mcp/internal/config"
)

// ConfigureGeminiInput is the input schema for the configure_gemini tool.
type ConfigureGeminiInput struct {
	APIKey string `json:"api_key"`
}

// ConfigureGeminiOutput is the output schema for the configure_gemini tool.
type ConfigureGeminiOutput struct {
	ConfigPath string `json:"config_path"`
	Status     string `json:"status"`
}

// ConfigureGemini implements the configure_gemini MCP tool.
// It saves the Gemini API key to config file and hot-reloads it in the running server.
func (s *Server) ConfigureGemini(ctx context.Context, input ConfigureGeminiInput, updateConfig func(key string)) string {
	if strings.TrimSpace(input.APIKey) == "" {
		return errorJSON("api_key is required")
	}

	cfg, err := config.Load()
	if err != nil {
		return errorJSON("failed to load config: " + err.Error())
	}

	cfg.GeminiAPIKey = input.APIKey

	if err := config.Save(cfg); err != nil {
		return errorJSON("failed to save config: " + err.Error())
	}

	// Hot-reload — update the in-memory config without restart
	if updateConfig != nil {
		updateConfig(input.APIKey)
	}

	return mustJSON(ConfigureGeminiOutput{
		ConfigPath: config.Path(),
		Status:     "saved",
	})
}
