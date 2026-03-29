package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"

	"dev-forge-mcp/internal/tools"
)

// recordSavedMsg carries the result of an async save operation.
type recordSavedMsg struct {
	output string
	err    error
}

// fieldDef describes a single form field in the wizard.
type fieldDef struct {
	key       string
	label     string
	required  bool
	hint      string
	multiline bool
}

var patternFields = []fieldDef{
	{"name", "Name", true, "Pattern name", false},
	{"category", "Category", false, "e.g. layout, form, card", false},
	{"domain", "Domain", false, "default: frontend", false},
	{"framework", "Framework", true, "e.g. next, astro, svelte", false},
	{"css_mode", "CSS Mode", true, "e.g. tailwind-v4, plain-css", false},
	{"tags", "Tags", false, "comma-separated", false},
	{"snippet", "HTML Snippet", true, "paste or type code", true},
	{"css_snippet", "CSS Snippet", false, "optional CSS", true},
	{"description", "Description", false, "short description", true},
}

var architectureFields = []fieldDef{
	{"name", "Name", true, "Architecture name", false},
	{"domain", "Domain", false, "default: fullstack", false},
	{"framework", "Framework", false, "e.g. next+trpc, go+htmx", false},
	{"css_mode", "CSS Mode", false, "e.g. tailwind-v4", false},
	{"description", "Description", false, "what this architecture does", true},
	{"decisions", "Decisions", false, "key technical decisions", true},
	{"tags", "Tags", false, "comma-separated", false},
}

var recordTypes = []string{"Pattern", "Architecture"}

// addRecordModel is the sub-model for the Add Record wizard.
type addRecordModel struct {
	phase    int // 0=type select, 1=fields, 2=result
	typeIdx  int // 0=Pattern, 1=Architecture
	fieldIdx int
	inputBuf string
	fieldErr string
	fieldValues map[string]string
	result      string
	isError     bool
	saving      bool
	embedderAvailable bool
	goHome bool
	srv    *tools.Server
}

func newAddRecordModel(srv *tools.Server) addRecordModel {
	m := addRecordModel{
		srv:         srv,
		fieldValues: make(map[string]string),
	}
	if srv != nil && srv.Embedder != nil {
		m.embedderAvailable = true
	}
	return m
}

func (m addRecordModel) currentFields() []fieldDef {
	if m.typeIdx == 0 {
		return patternFields
	}
	return architectureFields
}

func (m addRecordModel) Init() tea.Cmd { return nil }

func (m addRecordModel) Update(msg tea.Msg) (addRecordModel, tea.Cmd) {
	switch msg := msg.(type) {
	case recordSavedMsg:
		m.saving = false
		if msg.err != nil {
			m.result = msg.err.Error()
			m.isError = true
		} else {
			m.result = msg.output
			m.isError = false
		}
		m.phase = 2
		return m, nil

	case tea.KeyMsg:
		switch m.phase {
		case 0:
			return m.updatePhase0(msg)
		case 1:
			return m.updatePhase1(msg)
		case 2:
			return m.updatePhase2(msg)
		}
	}
	return m, nil
}

func (m addRecordModel) updatePhase0(msg tea.KeyMsg) (addRecordModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.typeIdx > 0 {
			m.typeIdx--
		}
	case "down", "j":
		if m.typeIdx < len(recordTypes)-1 {
			m.typeIdx++
		}
	case "enter", " ":
		// Reset field state
		m.fieldIdx = 0
		m.inputBuf = ""
		m.fieldErr = ""
		m.fieldValues = make(map[string]string)
		m.phase = 1
	case "esc", "q":
		m.goHome = true
	}
	return m, nil
}

func (m addRecordModel) updatePhase1(msg tea.KeyMsg) (addRecordModel, tea.Cmd) {
	fields := m.currentFields()

	switch msg.String() {
	case "esc":
		if m.fieldIdx == 0 {
			// Go back to type selection
			m.phase = 0
			m.inputBuf = ""
			m.fieldErr = ""
		} else {
			// Go back one field, restore previous value
			m.fieldIdx--
			m.inputBuf = m.fieldValues[fields[m.fieldIdx].key]
			m.fieldErr = ""
		}

	case "enter":
		field := fields[m.fieldIdx]
		val := strings.TrimSpace(m.inputBuf)
		if field.required && val == "" {
			m.fieldErr = "This field is required"
			return m, nil
		}
		m.fieldErr = ""
		m.fieldValues[field.key] = m.inputBuf

		if m.fieldIdx == len(fields)-1 {
			// Last field — trigger save
			m.saving = true
			return m, m.saveCmd()
		}
		m.fieldIdx++
		m.inputBuf = m.fieldValues[fields[m.fieldIdx].key]

	case "ctrl+j":
		// Newline for multiline fields
		m.inputBuf += "\n"

	case "backspace", "ctrl+h":
		if len(m.inputBuf) > 0 {
			// Trim last rune (UTF-8 safe)
			runes := []rune(m.inputBuf)
			m.inputBuf = string(runes[:len(runes)-1])
		}

	default:
		// Append printable runes
		for _, r := range msg.Runes {
			if unicode.IsPrint(r) {
				m.inputBuf += string(r)
			}
		}
	}
	return m, nil
}

func (m addRecordModel) updatePhase2(msg tea.KeyMsg) (addRecordModel, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc", "q":
		m.goHome = true
	}
	return m, nil
}

func (m addRecordModel) saveCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		var out string
		if m.typeIdx == 0 { // Pattern
			input := tools.StorePatternInput{
				Name:        m.fieldValues["name"],
				Category:    m.fieldValues["category"],
				Domain:      m.fieldValues["domain"],
				Framework:   m.fieldValues["framework"],
				CSSMode:     m.fieldValues["css_mode"],
				Tags:        m.fieldValues["tags"],
				Snippet:     m.fieldValues["snippet"],
				CSSSnippet:  m.fieldValues["css_snippet"],
				Description: m.fieldValues["description"],
			}
			out = m.srv.StorePattern(ctx, input)
		} else {
			input := tools.StoreArchitectureInput{
				Name:        m.fieldValues["name"],
				Domain:      m.fieldValues["domain"],
				Framework:   m.fieldValues["framework"],
				CSSMode:     m.fieldValues["css_mode"],
				Description: m.fieldValues["description"],
				Decisions:   m.fieldValues["decisions"],
				Tags:        m.fieldValues["tags"],
			}
			out = m.srv.StoreArchitecture(ctx, input)
		}
		if strings.Contains(out, `"error"`) {
			return recordSavedMsg{err: fmt.Errorf("%s", out)}
		}
		return recordSavedMsg{output: out}
	}
}

func (m addRecordModel) View() string {
	switch m.phase {
	case 0:
		return m.viewPhase0()
	case 1:
		return m.viewPhase1()
	case 2:
		return m.viewPhase2()
	}
	return ""
}

func (m addRecordModel) viewPhase0() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("◆ Add New Record") + "\n\n")
	b.WriteString(normalStyle.Render("Select record type:") + "\n")

	for i, t := range recordTypes {
		if i == m.typeIdx {
			b.WriteString(selectedStyle.Render(fmt.Sprintf("> %s", t)) + "\n")
		} else {
			b.WriteString(normalStyle.Render(fmt.Sprintf("  %s", t)) + "\n")
		}
	}

	b.WriteString("\n" + helpStyle.Render("↑/↓ navigate • Enter select • Esc back"))
	return b.String()
}

func (m addRecordModel) viewPhase1() string {
	fields := m.currentFields()
	field := fields[m.fieldIdx]
	typeName := recordTypes[m.typeIdx]
	total := len(fields)
	current := m.fieldIdx + 1

	var b strings.Builder

	title := fmt.Sprintf("◆ Add %s (%d/%d)", typeName, current, total)
	b.WriteString(titleStyle.Render(title) + "\n\n")

	// Field label + required indicator
	labelLine := field.label
	if field.required {
		labelLine += "  " + dimStyle.Render("[required]")
	}
	b.WriteString(normalStyle.Render(labelLine) + "\n")
	if field.hint != "" {
		b.WriteString(dimStyle.Render(field.hint) + "\n")
	}
	b.WriteString("\n")

	// Input display
	if field.multiline {
		lines := strings.Split(m.inputBuf, "\n")
		// Show last 5 lines
		start := 0
		if len(lines) > 5 {
			start = len(lines) - 5
		}
		for _, line := range lines[start:] {
			b.WriteString(boxStyle.Render(line) + "\n")
		}
		if len(m.inputBuf) == 0 {
			b.WriteString(boxStyle.Render("_") + "\n")
		}
		b.WriteString(dimStyle.Render("(ctrl+j for newline)") + "\n")
	} else {
		display := m.inputBuf + "_"
		b.WriteString(boxStyle.Render(display) + "\n")
	}

	// Inline validation error
	if m.fieldErr != "" {
		b.WriteString("\n" + errorStyle.Render("⚠ "+m.fieldErr) + "\n")
	}

	b.WriteString("\n")
	if m.saving {
		b.WriteString(dimStyle.Render("Saving...") + "\n")
	}

	helpLine := "Esc back • Enter confirm"
	if field.multiline {
		helpLine += " • ctrl+j newline"
	}
	b.WriteString(helpStyle.Render(helpLine))
	return b.String()
}

func (m addRecordModel) viewPhase2() string {
	var b strings.Builder

	if m.isError {
		b.WriteString(errorStyle.Render("✗ Error saving record") + "\n\n")
		b.WriteString(errorStyle.Render(m.result) + "\n")
	} else {
		// Parse result JSON
		typeName := recordTypes[m.typeIdx]
		var parsed struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			CreatedAt string `json:"created_at"`
		}
		if err := json.Unmarshal([]byte(m.result), &parsed); err == nil {
			b.WriteString(successStyle.Render(fmt.Sprintf("✓ %s saved: %q", typeName, parsed.Name)) + "\n")
			b.WriteString(dimStyle.Render(fmt.Sprintf("ID: %s", parsed.ID)) + "\n")
			if parsed.CreatedAt != "" {
				b.WriteString(dimStyle.Render(fmt.Sprintf("Created: %s", parsed.CreatedAt)) + "\n")
			}
		} else {
			b.WriteString(successStyle.Render(fmt.Sprintf("✓ %s saved", typeName)) + "\n")
			b.WriteString(dimStyle.Render(m.result) + "\n")
		}

		if m.embedderAvailable {
			b.WriteString("\n" + dimStyle.Render("Embedding generating in background...") + "\n")
		}
	}

	b.WriteString("\n" + helpStyle.Render("Enter/Esc/q → back to menu"))
	return b.String()
}
