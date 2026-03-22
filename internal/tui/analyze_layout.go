package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"dev-forge-mcp/internal/tools"
)

type analyzeLayoutModel struct {
	srv       *tools.Server
	filePath  string
	framework string
	cssMode   string
	field     int // 0=path, 1=framework, 2=css_mode
	result    string
	err       string
	goHome    bool
	running   bool
}

func newAnalyzeLayoutModel(srv *tools.Server, framework, cssMode string) analyzeLayoutModel {
	if framework == "" {
		framework = "vanilla"
	}
	if cssMode == "" {
		cssMode = "plain-css"
	}
	return analyzeLayoutModel{
		srv:       srv,
		framework: framework,
		cssMode:   cssMode,
	}
}

func (m analyzeLayoutModel) Init() tea.Cmd { return nil }

func (m analyzeLayoutModel) Update(msg tea.Msg) (analyzeLayoutModel, tea.Cmd) {
	switch msg := msg.(type) {
	case analyzeResultMsg:
		m.result = string(msg)
		m.running = false
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.result != "" || m.err != "" {
				m.result = ""
				m.err = ""
			} else {
				m.goHome = true
			}
		case "tab":
			m.field = (m.field + 1) % 3
		case "shift+tab":
			m.field = (m.field + 2) % 3
		case "enter":
			if m.field == 2 {
				m.running = true
				return m, m.runAnalysis()
			}
		case "backspace":
			switch m.field {
			case 0:
				if len(m.filePath) > 0 {
					m.filePath = m.filePath[:len(m.filePath)-1]
				}
			}
		case "up", "k":
			if m.field == 1 {
				m.framework = prevFramework(m.framework)
			}
			if m.field == 2 {
				m.cssMode = prevCSSMode(m.cssMode)
			}
		case "down", "j":
			if m.field == 1 {
				m.framework = nextFramework(m.framework)
			}
			if m.field == 2 {
				m.cssMode = nextCSSMode(m.cssMode)
			}
		default:
			if m.field == 0 && len(msg.String()) == 1 {
				m.filePath += msg.String()
			}
		}
	}
	return m, nil
}

type analyzeResultMsg string

func (m analyzeLayoutModel) runAnalysis() tea.Cmd {
	return func() tea.Msg {
		data, err := os.ReadFile(m.filePath)
		if err != nil {
			return analyzeResultMsg(`{"error":"` + err.Error() + `"}`)
		}
		input := tools.AnalyzeLayoutInput{
			Markup: string(data),
			Stack: tools.StackMeta{
				Framework: m.framework,
				CSSMode:   m.cssMode,
			},
		}
		return analyzeResultMsg(m.srv.AnalyzeLayout(context.Background(), input))
	}
}

func (m analyzeLayoutModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Analyze Layout File") + "\n\n")

	fields := []struct{ label, value string }{
		{"File path", m.filePath},
		{"Framework", m.framework},
		{"CSS mode", m.cssMode},
	}
	for i, f := range fields {
		cursor := "  "
		if m.field == i {
			cursor = "> "
		}
		val := f.value
		if m.field == i && i == 0 {
			val += "_"
		}
		if m.field == i {
			b.WriteString(selectedStyle.Render(fmt.Sprintf("%s%-12s %s", cursor, f.label+":", val)) + "\n")
		} else {
			b.WriteString(normalStyle.Render(fmt.Sprintf("%s%-12s %s", cursor, f.label+":", val)) + "\n")
		}
	}

	b.WriteString("\n" + helpStyle.Render("Tab/Shift+Tab move field • ↑/↓ cycle options • Enter (on CSS mode) to analyze • Esc back"))

	if m.result != "" {
		var out tools.AnalyzeLayoutOutput
		if json.Unmarshal([]byte(m.result), &out) == nil {
			b.WriteString("\n\n" + titleStyle.Render(fmt.Sprintf("Score: %d/100 — %s", out.Score, out.Summary)))
			for _, issue := range out.Issues {
				sev := ""
				switch issue.Severity {
				case "error":
					sev = errorStyle.Render("[ERR]")
				case "warning":
					sev = "  [WRN]"
				default:
					sev = dimStyle.Render("  [SUG]")
				}
				b.WriteString(fmt.Sprintf("\n%s %s: %s", sev, issue.Category, issue.Description))
				if issue.Suggestion != "" {
					b.WriteString("\n       " + dimStyle.Render("→ "+issue.Suggestion))
				}
			}
		} else {
			b.WriteString("\n\n" + m.result)
		}
	}
	if m.err != "" {
		b.WriteString("\n\n" + errorStyle.Render(m.err))
	}

	return b.String()
}

var frameworks = []string{"vanilla", "astro", "next", "sveltekit", "nuxt", "spa-vite"}
var cssModes = []string{"plain-css", "tailwind-v4"}

func nextFramework(current string) string {
	for i, f := range frameworks {
		if f == current {
			return frameworks[(i+1)%len(frameworks)]
		}
	}
	return frameworks[0]
}

func prevFramework(current string) string {
	for i, f := range frameworks {
		if f == current {
			return frameworks[(i+len(frameworks)-1)%len(frameworks)]
		}
	}
	return frameworks[0]
}

func nextCSSMode(current string) string {
	for i, f := range cssModes {
		if f == current {
			return cssModes[(i+1)%len(cssModes)]
		}
	}
	return cssModes[0]
}

func prevCSSMode(current string) string {
	for i, f := range cssModes {
		if f == current {
			return cssModes[(i+len(cssModes)-1)%len(cssModes)]
		}
	}
	return cssModes[0]
}
