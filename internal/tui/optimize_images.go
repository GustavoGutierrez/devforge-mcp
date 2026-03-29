package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"dev-forge-mcp/internal/tools"
)

type optimizeImagesModel struct {
	srv     *tools.Server
	paths   string
	quality string
	formats string
	field   int
	result  string
	err     string
	goHome  bool
}

func newOptimizeImagesModel(srv *tools.Server) optimizeImagesModel {
	return optimizeImagesModel{srv: srv, quality: "85", formats: "webp,avif"}
}

func (m optimizeImagesModel) Init() tea.Cmd { return nil }

func (m optimizeImagesModel) Update(msg tea.Msg) (optimizeImagesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case optimizeResultMsg:
		m.result = string(msg)
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
				return m, m.optimize()
			}
		case "backspace":
			switch m.field {
			case 0:
				if len(m.paths) > 0 {
					m.paths = m.paths[:len(m.paths)-1]
				}
			case 1:
				if len(m.quality) > 0 {
					m.quality = m.quality[:len(m.quality)-1]
				}
			case 2:
				if len(m.formats) > 0 {
					m.formats = m.formats[:len(m.formats)-1]
				}
			}
		default:
			if msg.Paste {
				switch m.field {
				case 0:
					m.paths += string(msg.Runes)
				case 1:
					m.quality += string(msg.Runes)
				case 2:
					m.formats += string(msg.Runes)
				}
			} else if len(msg.String()) == 1 {
				switch m.field {
				case 0:
					m.paths += msg.String()
				case 1:
					m.quality += msg.String()
				case 2:
					m.formats += msg.String()
				}
			}
		}
	}
	return m, nil
}

type optimizeResultMsg string

func (m optimizeImagesModel) optimize() tea.Cmd {
	return func() tea.Msg {
		paths := strings.Split(m.paths, ",")
		var inputs []tools.OptimizeInput
		fmts := strings.Split(m.formats, ",")
		for _, p := range paths {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			inputs = append(inputs, tools.OptimizeInput{
				Path:    p,
				Formats: fmts,
			})
		}
		input := tools.OptimizeImagesInput{Inputs: inputs, Parallelism: 4}
		result := m.srv.OptimizeImages(context.Background(), input)
		return optimizeResultMsg(result)
	}
}

func (m optimizeImagesModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Optimize Images") + "\n\n")

	fieldNames := []string{"Image paths (comma-sep)", "Quality (1-100)", "Formats (comma-sep)"}
	fieldVals := []string{m.paths, m.quality, m.formats}

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

	b.WriteString("\n" + helpStyle.Render("Tab move field • Enter (on Formats) optimize • Esc back"))

	if m.result != "" {
		var out tools.OptimizeImagesOutput
		if json.Unmarshal([]byte(m.result), &out) == nil {
			for _, res := range out.Results {
				b.WriteString("\n\n" + titleStyle.Render(res.SourcePath))
				for _, o := range res.Outputs {
					b.WriteString(fmt.Sprintf("\n  %s  %dx%d  %dKB  → %s",
						o.Format, o.Width, o.Height, o.ApproxSizeKB, o.Path))
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
