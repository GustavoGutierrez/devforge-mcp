package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var menuItems = []string{
	"Browse patterns",
	"Browse architectures",
	"Analyze layout file",
	"Generate layout",
	"Generate Image",
	"Optimize images",
	"Generate favicon",
	"Process video",
	"Process audio",
	"UI to Markdown",
	"Explore color palettes",
	"Settings",
	"Add Record",
	"Setup MCP Clients",
	"About",
	"Quit",
}

const asciiLogo = `██████╗ ███████╗██╗   ██╗███████╗ ██████╗ ██████╗  ██████╗ ███████╗
██╔══██╗██╔════╝██║   ██║██╔════╝██╔═══██╗██╔══██╗██╔════╝ ██╔════╝
██║  ██║█████╗  ██║   ██║█████╗  ██║   ██║██████╔╝██║  ███╗█████╗
██║  ██║██╔══╝  ╚██╗ ██╔╝██╔══╝  ██║   ██║██╔══██╗██║   ██║██╔══╝
██████╔╝███████╗ ╚████╔╝ ██║     ╚██████╔╝██║  ██║╚██████╔╝███████╗
╚═════╝ ╚══════╝  ╚═══╝  ╚═╝      ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚══════╝`

type homeModel struct {
	cursor   int
	selected int // -1 = none selected
}

func newHomeModel() homeModel {
	return homeModel{cursor: 0, selected: -1}
}

func (m homeModel) Init() tea.Cmd { return nil }

func (m homeModel) Update(msg tea.Msg) (homeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(menuItems)-1 {
				m.cursor++
			}
		case "enter", " ":
			if m.cursor == len(menuItems)-1 {
				return m, tea.Quit
			}
			m.selected = m.cursor
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m homeModel) View() string {
	var b strings.Builder

	logo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Render(asciiLogo)
	b.WriteString(logo + "\n\n")

	b.WriteString(dimStyle.Render("AI-powered toolkit for design consistency, UI generation, and code acceleration across your stack.") + "\n\n")
	b.WriteString(helpStyle.Render("↑ ↓  move   Enter  select   Tab  switch field   Esc  go back   q  quit") + "\n\n")

	for i, item := range menuItems {
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
		}

		var line string
		if m.cursor == i {
			line = selectedStyle.Render(fmt.Sprintf("%s%s", cursor, item))
		} else {
			line = normalStyle.Render(fmt.Sprintf("%s%s", cursor, item))
		}
		b.WriteString(line + "\n")
	}

	return boxStyle.Render(b.String())
}
