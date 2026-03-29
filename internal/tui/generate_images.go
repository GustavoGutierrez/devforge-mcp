package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"dev-forge-mcp/internal/config"
	"dev-forge-mcp/internal/tools"
)

type generateImagesModel struct {
	srv        *tools.Server
	cfg        *config.Config
	prompt     string
	style      string
	outputPath string
	field      int
	result     string
	err        string
	generating bool
	goHome     bool
	goSettings bool
}

var imageStyles = []string{"mockup", "wireframe", "illustration"}

func newGenerateImagesModel(srv *tools.Server, cfg *config.Config) generateImagesModel {
	return generateImagesModel{srv: srv, cfg: cfg, style: "mockup", outputPath: "./output.png"}
}

func (m generateImagesModel) Init() tea.Cmd { return nil }

func (m generateImagesModel) Update(msg tea.Msg) (generateImagesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case generateImageResultMsg:
		m.generating = false
		raw := string(msg)
		var out tools.GenerateUIImageOutput
		if json.Unmarshal([]byte(raw), &out) == nil {
			m.result = raw
			m.err = ""
		} else {
			m.err = raw
			m.result = ""
		}
		return m, nil
	case tea.KeyMsg:
		if m.generating {
			return m, nil
		}
		switch msg.String() {
		case "esc":
			if m.result != "" || m.err != "" {
				m.result = ""
				m.err = ""
			} else {
				m.goHome = true
			}
		case "s":
			if m.cfg.GeminiAPIKey == "" {
				m.goSettings = true
			}
		case "tab":
			m.field = (m.field + 1) % 3
		case "shift+tab":
			m.field = (m.field + 2) % 3
		case "up", "k":
			if m.field == 1 {
				m.style = prevImageStyle(m.style)
			}
		case "down", "j":
			if m.field == 1 {
				m.style = nextImageStyle(m.style)
			}
		case "enter":
			m.generating = true
			m.result = ""
			m.err = ""
			return m, m.generate()
		case "backspace":
			switch m.field {
			case 0:
				if len(m.prompt) > 0 {
					m.prompt = m.prompt[:len(m.prompt)-1]
				}
			case 2:
				if len(m.outputPath) > 0 {
					m.outputPath = m.outputPath[:len(m.outputPath)-1]
				}
			}
		default:
			if msg.Paste {
				switch m.field {
				case 0:
					m.prompt += string(msg.Runes)
				case 2:
					m.outputPath += string(msg.Runes)
				}
			} else if len(msg.String()) == 1 {
				switch m.field {
				case 0:
					m.prompt += msg.String()
				case 2:
					m.outputPath += msg.String()
				}
			}
		}
	}
	return m, nil
}

type generateImageResultMsg string

func (m generateImagesModel) generate() tea.Cmd {
	return func() tea.Msg {
		input := tools.GenerateUIImageInput{
			Prompt:     m.prompt,
			Style:      m.style,
			Width:      1280,
			Height:     720,
			OutputPath: m.outputPath,
		}
		key := ""
		imageModel := ""
		if m.cfg != nil {
			key = m.cfg.GeminiAPIKey
			imageModel = m.cfg.ImageModel
		}
		return generateImageResultMsg(m.srv.GenerateUIImage(context.Background(), input, key, imageModel))
	}
}

func (m generateImagesModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Generate UI Images") + "\n\n")

	if m.cfg.GeminiAPIKey == "" {
		b.WriteString(errorStyle.Render("Gemini API key not configured.") + "\n")
		b.WriteString(dimStyle.Render("Press 's' to go to Settings and configure it, or Esc to go back.") + "\n\n")
	}

	fieldNames := []string{"Prompt", "Style", "Output path"}
	fieldVals := []string{m.prompt, m.style, m.outputPath}

	for i, name := range fieldNames {
		cursor := "  "
		if m.field == i {
			cursor = "> "
		}
		val := fieldVals[i]
		if m.field == i && (i == 0 || i == 2) {
			val += "_"
		}
		line := fmt.Sprintf("%s%-12s %s", cursor, name+":", val)
		if m.field == i {
			b.WriteString(selectedStyle.Render(line) + "\n")
		} else {
			b.WriteString(normalStyle.Render(line) + "\n")
		}
	}

	if m.generating {
		b.WriteString("\n" + helpStyle.Render("Please wait..."))
	} else {
		b.WriteString("\n" + helpStyle.Render("Tab move field • ↑/↓ cycle style • Enter generate • s Settings • Esc back"))
	}

	if m.generating {
		b.WriteString("\n\n" + dimStyle.Render("Generating image..."))
	}

	if m.result != "" {
		var out tools.GenerateUIImageOutput
		if json.Unmarshal([]byte(m.result), &out) == nil {
			b.WriteString("\n\n" + successStyle.Render(fmt.Sprintf("✓ Image generated: %s (%dx%d)", out.Path, out.Width, out.Height)))
		} else {
			b.WriteString("\n\n" + successStyle.Render("✓ Image generated: "+m.result))
		}
	}
	if m.err != "" {
		b.WriteString("\n\n" + errorStyle.Render("✗ Error: "+m.err))
	}

	return b.String()
}

func nextImageStyle(current string) string {
	for i, s := range imageStyles {
		if s == current {
			return imageStyles[(i+1)%len(imageStyles)]
		}
	}
	return imageStyles[0]
}

func prevImageStyle(current string) string {
	for i, s := range imageStyles {
		if s == current {
			return imageStyles[(i+len(imageStyles)-1)%len(imageStyles)]
		}
	}
	return imageStyles[0]
}
