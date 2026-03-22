package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"dev-forge-mcp/internal/config"
)

type settingsModel struct {
	cfg    *config.Config
	apiKey string // masked input buffer
	field  int
	status string
	err    string
	goHome bool
	saved  bool
}

func newSettingsModel(cfg *config.Config) settingsModel {
	key := ""
	if cfg != nil {
		key = cfg.GeminiAPIKey
	}
	return settingsModel{cfg: cfg, apiKey: key}
}

func (m settingsModel) Init() tea.Cmd { return nil }

func (m settingsModel) Update(msg tea.Msg) (settingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case settingsSavedMsg:
		if msg.err != nil {
			m.err = msg.err.Error()
		} else {
			m.status = "Saved successfully."
			m.saved = true
			if m.cfg != nil {
				m.cfg.GeminiAPIKey = m.apiKey
			}
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.goHome = true
		case "tab":
			m.field = (m.field + 1) % 3
		case "shift+tab":
			m.field = (m.field + 2) % 3
		case "enter":
			switch m.field {
			case 1: // Save
				return m, m.save()
			case 2: // Delete key
				m.apiKey = ""
				return m, m.save()
			}
		case "backspace":
			if m.field == 0 && len(m.apiKey) > 0 {
				m.apiKey = m.apiKey[:len(m.apiKey)-1]
			}
		default:
			if m.field == 0 && len(msg.String()) == 1 {
				m.apiKey += msg.String()
			}
		}
	}
	return m, nil
}

type settingsSavedMsg struct{ err error }

func (m settingsModel) save() tea.Cmd {
	return func() tea.Msg {
		cfg := m.cfg
		if cfg == nil {
			cfg = &config.Config{
				OllamaURL:      "http://localhost:11434",
				EmbeddingModel: "nomic-embed-text",
			}
		}
		cfg.GeminiAPIKey = m.apiKey
		err := config.Save(cfg)
		return settingsSavedMsg{err: err}
	}
}

func (m settingsModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Settings") + "\n\n")

	// Config path
	b.WriteString(dimStyle.Render("Config file: "+config.Path()) + "\n\n")

	// API key status
	statusIcon := "✗"
	statusColor := errorStyle
	if m.cfg != nil && m.cfg.GeminiAPIKey != "" {
		statusIcon = "✓"
		statusColor = successStyle
	}
	b.WriteString(fmt.Sprintf("Gemini API key: %s\n\n", statusColor.Render(statusIcon)))

	// Fields
	cursor0, cursor1, cursor2 := "  ", "  ", "  "
	switch m.field {
	case 0:
		cursor0 = "> "
	case 1:
		cursor1 = "> "
	case 2:
		cursor2 = "> "
	}

	// Masked key input
	maskedKey := strings.Repeat("*", len(m.apiKey))
	if m.field == 0 {
		maskedKey += "_"
		b.WriteString(selectedStyle.Render(cursor0+"API Key: "+maskedKey) + "\n")
	} else {
		b.WriteString(normalStyle.Render(cursor0+"API Key: "+maskedKey) + "\n")
	}

	// Save button
	if m.field == 1 {
		b.WriteString(selectedStyle.Render(cursor1+"[Save]") + "\n")
	} else {
		b.WriteString(normalStyle.Render(cursor1+"[Save]") + "\n")
	}

	// Delete button
	if m.field == 2 {
		b.WriteString(selectedStyle.Render(cursor2+"[Delete key]") + "\n")
	} else {
		b.WriteString(normalStyle.Render(cursor2+"[Delete key]") + "\n")
	}

	b.WriteString("\n" + helpStyle.Render("Tab move • Enter activate • Esc back"))

	if m.status != "" {
		b.WriteString("\n\n" + successStyle.Render(m.status))
	}
	if m.err != "" {
		b.WriteString("\n\n" + errorStyle.Render(m.err))
	}

	return b.String()
}
