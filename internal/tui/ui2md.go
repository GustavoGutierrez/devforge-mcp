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

type ui2mdModel struct {
	srv        *tools.Server
	cfg        *config.Config
	imagePath  string
	outputDir  string
	field      int
	result     string
	err        string
	generating bool
	goHome     bool
}

func newUI2MDModel(srv *tools.Server, cfg *config.Config) ui2mdModel {
	return ui2mdModel{srv: srv, cfg: cfg}
}

func (m ui2mdModel) Init() tea.Cmd { return nil }

func (m ui2mdModel) Update(msg tea.Msg) (ui2mdModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ui2mdResultMsg:
		m.generating = false
		raw := string(msg)
		var out tools.UI2MDOutput
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
		case "tab":
			m.field = (m.field + 1) % 2
		case "shift+tab":
			m.field = (m.field - 1 + 2) % 2
		case "enter":
			if strings.TrimSpace(m.imagePath) == "" {
				return m, nil
			}
			m.generating = true
			m.result = ""
			m.err = ""
			return m, m.analyze()
		case "backspace":
			switch m.field {
			case 0:
				if len(m.imagePath) > 0 {
					m.imagePath = m.imagePath[:len(m.imagePath)-1]
				}
			case 1:
				if len(m.outputDir) > 0 {
					m.outputDir = m.outputDir[:len(m.outputDir)-1]
				}
			}
		default:
			if msg.Paste {
				switch m.field {
				case 0:
					m.imagePath += string(msg.Runes)
				case 1:
					m.outputDir += string(msg.Runes)
				}
			} else if len(msg.String()) == 1 {
				switch m.field {
				case 0:
					m.imagePath += msg.String()
				case 1:
					m.outputDir += msg.String()
				}
			}
		}
	}
	return m, nil
}

type ui2mdResultMsg string

func (m ui2mdModel) analyze() tea.Cmd {
	return func() tea.Msg {
		input := tools.UI2MDInput{
			ImagePath: m.imagePath,
			OutputDir: m.outputDir,
		}
		key := ""
		imageModel := ""
		if m.cfg != nil {
			key = m.cfg.GeminiAPIKey
			imageModel = m.cfg.ImageModel
		}
		result := m.srv.UI2MD(context.Background(), input, key, imageModel)
		return ui2mdResultMsg(result)
	}
}

func (m ui2mdModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("UI to Markdown") + "\n\n")

	fieldNames := []string{"Image path", "Output dir (optional)"}
	fieldVals := []string{m.imagePath, m.outputDir}

	for i, name := range fieldNames {
		cursor := "  "
		if m.field == i {
			cursor = "> "
		}
		val := fieldVals[i]
		if m.field == i {
			val += "_"
		}
		line := fmt.Sprintf("%s%-24s %s", cursor, name+":", val)
		if m.field == i {
			b.WriteString(selectedStyle.Render(line) + "\n")
		} else {
			b.WriteString(normalStyle.Render(line) + "\n")
		}
	}

	if m.generating {
		b.WriteString("\n" + helpStyle.Render("Please wait..."))
	} else {
		b.WriteString("\n" + helpStyle.Render("Tab move field • Enter analyze • Esc back"))
	}

	if m.generating {
		b.WriteString("\n\n" + dimStyle.Render("Analyzing image..."))
	}

	if m.result != "" {
		var out tools.UI2MDOutput
		if json.Unmarshal([]byte(m.result), &out) == nil {
			b.WriteString("\n\n" + successStyle.Render("✓ Spec generated: "+out.FilePath))
			if out.Name != "" {
				b.WriteString("\n" + dimStyle.Render("  Name: "+out.Name))
			}
		} else {
			b.WriteString("\n\n" + successStyle.Render("✓ Spec generated: "+m.result))
		}
	}
	if m.err != "" {
		b.WriteString("\n\n" + errorStyle.Render("✗ Error: "+m.err))
	}

	return b.String()
}
