package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"dev-forge-mcp/internal/tools"
)

// ArchitectureResult mirrors pattern result shape for architectures.
type ArchitectureResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	Framework   string `json:"framework"`
	CSSMode     string `json:"css_mode"`
	Description string `json:"description"`
	Decisions   string `json:"decisions"`
	Tags        string `json:"tags"`
	CreatedAt   string `json:"created_at"`
}

type browseArchitecturesModel struct {
	srv           *tools.Server
	cursor        int
	query         string
	architectures []ArchitectureResult
	detail        string
	goHome        bool
	inputMode     bool
}

func newBrowseArchitecturesModel(srv *tools.Server) browseArchitecturesModel {
	m := browseArchitecturesModel{srv: srv}
	m.refresh()
	return m
}

func (m *browseArchitecturesModel) refresh() {
	// Query architectures from DB directly since we don't have a dedicated list_architectures tool
	if m.srv.DB == nil {
		return
	}
	query := `SELECT id, name, domain, COALESCE(framework,''), COALESCE(css_mode,''), COALESCE(description,''), COALESCE(decisions,''), COALESCE(tags,''), created_at
		FROM architectures ORDER BY created_at DESC LIMIT 50`
	rows, err := m.srv.DB.QueryContext(context.Background(), query)
	if err != nil {
		return
	}
	defer rows.Close()
	m.architectures = nil
	for rows.Next() {
		var a ArchitectureResult
		rows.Scan(&a.ID, &a.Name, &a.Domain, &a.Framework, &a.CSSMode, &a.Description, &a.Decisions, &a.Tags, &a.CreatedAt)
		m.architectures = append(m.architectures, a)
	}
}

func (m browseArchitecturesModel) Init() tea.Cmd { return nil }

func (m browseArchitecturesModel) Update(msg tea.Msg) (browseArchitecturesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.inputMode {
			switch msg.String() {
			case "enter", "esc":
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
			if m.cursor < len(m.architectures)-1 {
				m.cursor++
			}
		case "/":
			m.inputMode = true
		case "enter":
			if len(m.architectures) > 0 && m.cursor < len(m.architectures) {
				a := m.architectures[m.cursor]
				m.detail = fmt.Sprintf("Name: %s\nDomain: %s\nFramework: %s\nDescription: %s\nDecisions: %s\nTags: %s",
					a.Name, a.Domain, a.Framework, a.Description, a.Decisions, a.Tags)
			}
		case "r":
			m.query = ""
			m.refresh()
		}
	}
	return m, nil
}

func (m browseArchitecturesModel) View() string {
	// Suppress unused import warning
	_ = json.Marshal

	var b strings.Builder
	b.WriteString(titleStyle.Render("Browse Architectures") + "\n\n")

	if len(m.architectures) == 0 {
		b.WriteString(dimStyle.Render("No architectures found. Seed the database to get started."))
	} else {
		for i, a := range m.architectures {
			line := fmt.Sprintf("%-30s %-12s %s", truncStr(a.Name, 30), a.Domain, dimStyle.Render(a.Framework))
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

	b.WriteString("\n" + helpStyle.Render("↑/↓ navigate • Enter view detail • r refresh • Esc back"))
	return b.String()
}
