package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"dev-forge-mcp/internal/tools"
)

type browsePatternsModel struct {
	srv       *tools.Server
	cursor    int
	query     string
	framework string
	cssMode   string
	patterns  []tools.PatternResult
	detail    string
	err       string
	goHome    bool
	inputMode bool // true when typing in search
}

func newBrowsePatternsModel(srv *tools.Server) browsePatternsModel {
	m := browsePatternsModel{srv: srv}
	m.refresh()
	return m
}

func (m *browsePatternsModel) refresh() {
	input := tools.ListPatternsInput{
		Query:     m.query,
		Framework: m.framework,
		CSSMode:   m.cssMode,
		Limit:     50,
	}
	if m.query != "" {
		input.Mode = "fts"
	}
	result := m.srv.ListPatterns(context.Background(), input)
	var out tools.ListPatternsOutput
	if err := json.Unmarshal([]byte(result), &out); err == nil {
		m.patterns = out.Patterns
	}
}

func (m browsePatternsModel) Init() tea.Cmd { return nil }

func (m browsePatternsModel) Update(msg tea.Msg) (browsePatternsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.inputMode {
			switch msg.String() {
			case "enter":
				m.inputMode = false
				m.refresh()
			case "esc":
				m.inputMode = false
			case "backspace":
				if len(m.query) > 0 {
					m.query = m.query[:len(m.query)-1]
				}
			default:
				if len(msg.String()) == 1 {
					m.query += msg.String()
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "esc":
			m.goHome = true
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.patterns)-1 {
				m.cursor++
			}
		case "/":
			m.inputMode = true
		case "enter":
			if len(m.patterns) > 0 && m.cursor < len(m.patterns) {
				p := m.patterns[m.cursor]
				m.detail = fmt.Sprintf("Name: %s\nDomain: %s\nFramework: %s\nCSS Mode: %s\nTags: %s\nDescription: %s",
					p.Name, p.Domain, p.Framework, p.CSSMode, p.Tags, p.Description)
			}
		case "r":
			m.query = ""
			m.refresh()
		}
	}
	return m, nil
}

func (m browsePatternsModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Browse Patterns"))
	b.WriteString("\n")

	// Search bar
	searchLabel := "Search: "
	if m.inputMode {
		searchLabel = "Search (enter to search, esc to cancel): "
	}
	b.WriteString(dimStyle.Render(searchLabel) + m.query)
	if m.inputMode {
		b.WriteString("_")
	}
	b.WriteString("\n\n")

	if len(m.patterns) == 0 {
		b.WriteString(dimStyle.Render("No patterns found. Press 'r' to reset or '/' to search."))
	} else {
		for i, p := range m.patterns {
			line := fmt.Sprintf("%-30s %-12s %-10s %s", truncStr(p.Name, 30), p.Framework, p.CSSMode, dimStyle.Render(p.Domain))
			if i == m.cursor {
				b.WriteString(selectedStyle.Render("> "+line) + "\n")
			} else {
				b.WriteString(normalStyle.Render("  "+line) + "\n")
			}
		}
	}

	if m.detail != "" {
		b.WriteString("\n" + boxStyle.Render(m.detail))
	}

	b.WriteString("\n" + helpStyle.Render("↑/↓ navigate • Enter view detail • / search • r reset • Esc back"))
	return b.String()
}

func truncStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}
