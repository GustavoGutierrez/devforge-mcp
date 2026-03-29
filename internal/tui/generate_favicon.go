package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"dev-forge-mcp/internal/tools"
)

type generateFaviconModel struct {
	srv        *tools.Server
	source     string
	bgColor    string
	field      int
	result     string
	err        string
	generating bool
	goHome     bool
}

func newGenerateFaviconModel(srv *tools.Server) generateFaviconModel {
	return generateFaviconModel{srv: srv, bgColor: "#ffffff"}
}

func (m generateFaviconModel) Init() tea.Cmd { return nil }

func (m generateFaviconModel) Update(msg tea.Msg) (generateFaviconModel, tea.Cmd) {
	switch msg := msg.(type) {
	case faviconResultMsg:
		m.generating = false
		raw := string(msg)
		var out tools.GenerateFaviconOutput
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
			m.generating = true
			m.result = ""
			m.err = ""
			return m, m.generate()
		case "backspace":
			switch m.field {
			case 0:
				if len(m.source) > 0 {
					m.source = m.source[:len(m.source)-1]
				}
			case 1:
				if len(m.bgColor) > 0 {
					m.bgColor = m.bgColor[:len(m.bgColor)-1]
				}
			}
		default:
			if msg.Paste {
				switch m.field {
				case 0:
					m.source += string(msg.Runes)
				case 1:
					m.bgColor += string(msg.Runes)
				}
			} else if len(msg.String()) == 1 {
				switch m.field {
				case 0:
					m.source += msg.String()
				case 1:
					m.bgColor += msg.String()
				}
			}
		}
	}
	return m, nil
}

type faviconResultMsg string

func (m generateFaviconModel) generate() tea.Cmd {
	return func() tea.Msg {
		input := tools.GenerateFaviconInput{
			SourcePath:      m.source,
			BackgroundColor: m.bgColor,
			Sizes:           []int{16, 32, 48, 180, 192, 512},
			Formats:         []string{"ico", "png"},
		}
		result := m.srv.GenerateFavicon(context.Background(), input)
		return faviconResultMsg(result)
	}
}

func (m generateFaviconModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Generate Favicon") + "\n\n")

	fieldNames := []string{"Source image path", "Background color (hex)"}
	fieldVals := []string{m.source, m.bgColor}

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
		b.WriteString("\n" + helpStyle.Render("Tab move field • Enter generate • Esc back"))
	}

	if m.generating {
		b.WriteString("\n\n" + dimStyle.Render("Generating favicon..."))
	}

	if m.result != "" {
		var out tools.GenerateFaviconOutput
		if json.Unmarshal([]byte(m.result), &out) == nil {
			b.WriteString("\n\n" + successStyle.Render(fmt.Sprintf("✓ Favicon generated: %d icons", len(out.Icons))))
			for _, icon := range out.Icons {
				b.WriteString(fmt.Sprintf("\n  %s %dx%d → %s", icon.Format, icon.Size, icon.Size, icon.Path))
			}
			if len(out.HTMLSnippets) > 0 {
				b.WriteString("\n\nHTML snippets:\n")
				for _, s := range out.HTMLSnippets {
					b.WriteString(dimStyle.Render("  "+s) + "\n")
				}
			}
		} else {
			b.WriteString("\n\n" + successStyle.Render("✓ Favicon generated: "+m.result))
		}
	}
	if m.err != "" {
		b.WriteString("\n\n" + errorStyle.Render("✗ Error: "+m.err))
	}

	return b.String()
}
