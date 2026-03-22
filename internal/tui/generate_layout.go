package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"dev-forge-mcp/internal/tools"
)

type generateLayoutModel struct {
	srv         *tools.Server
	description string
	framework   string
	cssMode     string
	fidelity    string
	field       int
	result      *tools.SuggestLayoutOutput
	err         string
	goHome      bool
}

func newGenerateLayoutModel(srv *tools.Server, framework, cssMode string) generateLayoutModel {
	if framework == "" {
		framework = "vanilla"
	}
	if cssMode == "" {
		cssMode = "plain-css"
	}
	return generateLayoutModel{srv: srv, framework: framework, cssMode: cssMode, fidelity: "mid"}
}

var fidelities = []string{"wireframe", "mid", "production"}

func (m generateLayoutModel) Init() tea.Cmd { return nil }

func (m generateLayoutModel) Update(msg tea.Msg) (generateLayoutModel, tea.Cmd) {
	switch msg := msg.(type) {
	case generateLayoutResultMsg:
		var out tools.SuggestLayoutOutput
		if json.Unmarshal([]byte(msg), &out) == nil {
			m.result = &out
		} else {
			m.err = string(msg)
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.result != nil || m.err != "" {
				m.result = nil
				m.err = ""
			} else {
				m.goHome = true
			}
		case "tab":
			m.field = (m.field + 1) % 4
		case "shift+tab":
			m.field = (m.field + 3) % 4
		case "up", "k":
			switch m.field {
			case 1:
				m.framework = prevFramework(m.framework)
			case 2:
				m.cssMode = prevCSSMode(m.cssMode)
			case 3:
				m.fidelity = prevFidelity(m.fidelity)
			}
		case "down", "j":
			switch m.field {
			case 1:
				m.framework = nextFramework(m.framework)
			case 2:
				m.cssMode = nextCSSMode(m.cssMode)
			case 3:
				m.fidelity = nextFidelity(m.fidelity)
			}
		case "enter":
			if m.field == 3 {
				return m, m.generate()
			}
		case "backspace":
			if m.field == 0 && len(m.description) > 0 {
				m.description = m.description[:len(m.description)-1]
			}
		default:
			if m.field == 0 && len(msg.String()) == 1 {
				m.description += msg.String()
			}
		}
	}
	return m, nil
}

type generateLayoutResultMsg string

func (m generateLayoutModel) generate() tea.Cmd {
	return func() tea.Msg {
		input := tools.SuggestLayoutInput{
			Description: m.description,
			Stack:       tools.StackMeta{Framework: m.framework, CSSMode: m.cssMode},
			Fidelity:    m.fidelity,
		}
		return generateLayoutResultMsg(m.srv.SuggestLayout(context.Background(), input))
	}
}

func (m generateLayoutModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Generate Layout") + "\n\n")

	fieldNames := []string{"Description", "Framework", "CSS mode", "Fidelity"}
	fieldVals := []string{m.description, m.framework, m.cssMode, m.fidelity}

	for i, name := range fieldNames {
		cursor := "  "
		if m.field == i {
			cursor = "> "
		}
		val := fieldVals[i]
		if m.field == i && i == 0 {
			val += "_"
		}
		line := fmt.Sprintf("%s%-12s %s", cursor, name+":", val)
		if m.field == i {
			b.WriteString(selectedStyle.Render(line) + "\n")
		} else {
			b.WriteString(normalStyle.Render(line) + "\n")
		}
	}

	b.WriteString("\n" + helpStyle.Render("Tab move field • ↑/↓ cycle options • Enter (on Fidelity) generate • Esc back"))

	if m.result != nil {
		b.WriteString("\n\n" + successStyle.Render("Generated: "+m.result.LayoutName))
		b.WriteString("\n" + dimStyle.Render(m.result.Rationale))
		for _, f := range m.result.Files {
			b.WriteString("\n\n" + titleStyle.Render(f.Path) + "\n")
			b.WriteString(dimStyle.Render(truncStr(f.Snippet, 200)))
		}
	}
	if m.err != "" {
		b.WriteString("\n\n" + errorStyle.Render(m.err))
	}

	_ = json.Marshal // keep import
	return b.String()
}

func nextFidelity(current string) string {
	for i, f := range fidelities {
		if f == current {
			return fidelities[(i+1)%len(fidelities)]
		}
	}
	return fidelities[0]
}

func prevFidelity(current string) string {
	for i, f := range fidelities {
		if f == current {
			return fidelities[(i+len(fidelities)-1)%len(fidelities)]
		}
	}
	return fidelities[0]
}
