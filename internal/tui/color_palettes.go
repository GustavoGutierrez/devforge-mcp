package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"dev-forge-mcp/internal/tools"
)

type colorPalettesModel struct {
	srv      *tools.Server
	useCase  string
	mood     string
	keywords string
	field    int
	palettes []tools.ColorPalette
	err      string
	goHome   bool
}

func newColorPalettesModel(srv *tools.Server) colorPalettesModel {
	return colorPalettesModel{srv: srv, mood: "calm"}
}

func (m colorPalettesModel) Init() tea.Cmd { return nil }

func (m colorPalettesModel) Update(msg tea.Msg) (colorPalettesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case []tools.ColorPalette:
		m.palettes = msg
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if len(m.palettes) > 0 {
				m.palettes = nil
			} else {
				m.goHome = true
			}
		case "tab":
			m.field = (m.field + 1) % 3
		case "shift+tab":
			m.field = (m.field + 2) % 3
		case "enter":
			if m.field == 2 {
				return m, m.suggest()
			}
		case "backspace":
			switch m.field {
			case 0:
				if len(m.useCase) > 0 {
					m.useCase = m.useCase[:len(m.useCase)-1]
				}
			case 1:
				if len(m.mood) > 0 {
					m.mood = m.mood[:len(m.mood)-1]
				}
			case 2:
				if len(m.keywords) > 0 {
					m.keywords = m.keywords[:len(m.keywords)-1]
				}
			}
		default:
			if len(msg.String()) == 1 {
				switch m.field {
				case 0:
					m.useCase += msg.String()
				case 1:
					m.mood += msg.String()
				case 2:
					m.keywords += msg.String()
				}
			}
		}
	}
	return m, nil
}

func (m colorPalettesModel) suggest() tea.Cmd {
	return func() tea.Msg {
		kws := strings.Split(m.keywords, ",")
		for i, k := range kws {
			kws[i] = strings.TrimSpace(k)
		}
		input := tools.SuggestColorPalettesInput{
			UseCase:       m.useCase,
			Mood:          m.mood,
			BrandKeywords: kws,
			Count:         3,
		}
		result := m.srv.SuggestColorPalettes(context.Background(), input)
		var out tools.SuggestColorPalettesOutput
		if json.Unmarshal([]byte(result), &out) == nil {
			return out.Palettes
		}
		return []tools.ColorPalette{}
	}
}

func (m colorPalettesModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Explore Color Palettes") + "\n\n")

	fieldNames := []string{"Use case", "Mood", "Brand keywords (comma-sep)"}
	fieldVals := []string{m.useCase, m.mood, m.keywords}

	for i, name := range fieldNames {
		cursor := "  "
		if m.field == i {
			cursor = "> "
		}
		val := fieldVals[i]
		if m.field == i {
			val += "_"
		}
		line := fmt.Sprintf("%s%-30s %s", cursor, name+":", val)
		if m.field == i {
			b.WriteString(selectedStyle.Render(line) + "\n")
		} else {
			b.WriteString(normalStyle.Render(line) + "\n")
		}
	}

	b.WriteString("\n" + helpStyle.Render("Tab move field • Enter (on Keywords) suggest • Esc back"))

	for _, palette := range m.palettes {
		b.WriteString("\n\n")
		b.WriteString(titleStyle.Render(palette.Name) + "\n")
		b.WriteString(dimStyle.Render(palette.Description) + "\n")
		b.WriteString(renderSwatch(palette.Tokens))
	}

	if m.err != "" {
		b.WriteString("\n\n" + errorStyle.Render(m.err))
	}

	return b.String()
}

// renderSwatch renders token colors as lipgloss-styled color blocks.
func renderSwatch(tokens tools.PaletteTokens) string {
	type colorEntry struct{ label, hex string }
	entries := []colorEntry{
		{"background", tokens.Background},
		{"surface", tokens.Surface},
		{"primary", tokens.Primary},
		{"primary-soft", tokens.PrimarySoft},
		{"accent", tokens.Accent},
		{"text", tokens.Text},
		{"muted", tokens.Muted},
	}

	var parts []string
	for _, e := range entries {
		swatch := lipgloss.NewStyle().
			Background(lipgloss.Color(e.hex)).
			Foreground(contrastColor(e.hex)).
			Padding(0, 1).
			Render("  ")
		parts = append(parts, fmt.Sprintf("%s %s %s", swatch, e.label, dimStyle.Render(e.hex)))
	}
	return strings.Join(parts, "\n")
}

// contrastColor returns black or white for text on a given hex background.
func contrastColor(hex string) lipgloss.Color {
	// Simple luminance heuristic based on hex value
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return lipgloss.Color("#000000")
	}
	r := hexByte(hex[0:2])
	g := hexByte(hex[2:4])
	b := hexByte(hex[4:6])
	lum := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
	if lum > 128 {
		return lipgloss.Color("#000000")
	}
	return lipgloss.Color("#ffffff")
}

func hexByte(s string) float64 {
	var v float64
	for _, c := range s {
		v *= 16
		switch {
		case c >= '0' && c <= '9':
			v += float64(c - '0')
		case c >= 'a' && c <= 'f':
			v += float64(c-'a') + 10
		case c >= 'A' && c <= 'F':
			v += float64(c-'A') + 10
		}
	}
	return v
}
