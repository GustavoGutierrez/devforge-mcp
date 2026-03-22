// Package tui provides the Bubble Tea TUI for dev-forge.
package tui

import (
	"database/sql"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"dev-forge-mcp/internal/config"
	"dev-forge-mcp/internal/tools"
)

// View identifies which TUI view is active.
type View int

const (
	ViewHome View = iota
	ViewBrowsePatterns
	ViewBrowseArchitectures
	ViewAnalyzeLayout
	ViewGenerateLayout
	ViewGenerateImages
	ViewOptimizeImages
	ViewGenerateFavicon
	ViewColorPalettes
	ViewSettings
)

// NavigateTo is a message that triggers view navigation.
type NavigateTo struct{ View View }

// Model is the root Bubble Tea model.
type Model struct {
	currentView View
	width       int
	height      int

	// Shared dependencies
	db     *sql.DB
	config *config.Config
	srv    *tools.Server

	// Sub-models
	home                homeModel
	browsePatterns      browsePatternsModel
	browseArchitectures browseArchitecturesModel
	analyzeLayout       analyzeLayoutModel
	generateLayout      generateLayoutModel
	generateImages      generateImagesModel
	optimizeImages      optimizeImagesModel
	generateFavicon     generateFaviconModel
	colorPalettes       colorPalettesModel
	settings            settingsModel

	// Detected stack
	detectedFramework string
	detectedCSSMode   string
}

// New creates the root model with all dependencies.
func New(database *sql.DB, cfg *config.Config, srv *tools.Server, framework, cssMode string) Model {
	m := Model{
		currentView:       ViewHome,
		db:                database,
		config:            cfg,
		srv:               srv,
		detectedFramework: framework,
		detectedCSSMode:   cssMode,
	}
	m.home = newHomeModel()
	m.browsePatterns = newBrowsePatternsModel(srv)
	m.browseArchitectures = newBrowseArchitecturesModel(srv)
	m.analyzeLayout = newAnalyzeLayoutModel(srv, framework, cssMode)
	m.generateLayout = newGenerateLayoutModel(srv, framework, cssMode)
	m.generateImages = newGenerateImagesModel(srv, cfg)
	m.optimizeImages = newOptimizeImagesModel(srv)
	m.generateFavicon = newGenerateFaviconModel(srv)
	m.colorPalettes = newColorPalettesModel(srv)
	m.settings = newSettingsModel(cfg)
	return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return m.home.Init()
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case NavigateTo:
		m.currentView = msg.View
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	// Delegate to sub-model
	switch m.currentView {
	case ViewHome:
		updated, cmd := m.home.Update(msg)
		m.home = updated
		// Check if home requested navigation
		if m.home.selected >= 0 {
			view := homeItemToView(m.home.selected)
			m.home.selected = -1
			m.currentView = view
			return m, nil
		}
		return m, cmd

	case ViewBrowsePatterns:
		updated, cmd := m.browsePatterns.Update(msg)
		m.browsePatterns = updated
		if m.browsePatterns.goHome {
			m.browsePatterns.goHome = false
			m.currentView = ViewHome
		}
		return m, cmd

	case ViewBrowseArchitectures:
		updated, cmd := m.browseArchitectures.Update(msg)
		m.browseArchitectures = updated
		if m.browseArchitectures.goHome {
			m.browseArchitectures.goHome = false
			m.currentView = ViewHome
		}
		return m, cmd

	case ViewAnalyzeLayout:
		updated, cmd := m.analyzeLayout.Update(msg)
		m.analyzeLayout = updated
		if m.analyzeLayout.goHome {
			m.analyzeLayout.goHome = false
			m.currentView = ViewHome
		}
		return m, cmd

	case ViewGenerateLayout:
		updated, cmd := m.generateLayout.Update(msg)
		m.generateLayout = updated
		if m.generateLayout.goHome {
			m.generateLayout.goHome = false
			m.currentView = ViewHome
		}
		return m, cmd

	case ViewGenerateImages:
		updated, cmd := m.generateImages.Update(msg)
		m.generateImages = updated
		if m.generateImages.goHome {
			m.generateImages.goHome = false
			m.currentView = ViewHome
		}
		if m.generateImages.goSettings {
			m.generateImages.goSettings = false
			m.currentView = ViewSettings
		}
		return m, cmd

	case ViewOptimizeImages:
		updated, cmd := m.optimizeImages.Update(msg)
		m.optimizeImages = updated
		if m.optimizeImages.goHome {
			m.optimizeImages.goHome = false
			m.currentView = ViewHome
		}
		return m, cmd

	case ViewGenerateFavicon:
		updated, cmd := m.generateFavicon.Update(msg)
		m.generateFavicon = updated
		if m.generateFavicon.goHome {
			m.generateFavicon.goHome = false
			m.currentView = ViewHome
		}
		return m, cmd

	case ViewColorPalettes:
		updated, cmd := m.colorPalettes.Update(msg)
		m.colorPalettes = updated
		if m.colorPalettes.goHome {
			m.colorPalettes.goHome = false
			m.currentView = ViewHome
		}
		return m, cmd

	case ViewSettings:
		updated, cmd := m.settings.Update(msg)
		m.settings = updated
		if m.settings.goHome {
			m.settings.goHome = false
			// Update config reference if API key changed
			if m.settings.saved {
				m.config = m.settings.cfg
				m.generateImages.cfg = m.config
				m.settings.saved = false
			}
			m.currentView = ViewHome
		}
		return m, cmd
	}

	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	switch m.currentView {
	case ViewHome:
		return m.home.View()
	case ViewBrowsePatterns:
		return m.browsePatterns.View()
	case ViewBrowseArchitectures:
		return m.browseArchitectures.View()
	case ViewAnalyzeLayout:
		return m.analyzeLayout.View()
	case ViewGenerateLayout:
		return m.generateLayout.View()
	case ViewGenerateImages:
		return m.generateImages.View()
	case ViewOptimizeImages:
		return m.optimizeImages.View()
	case ViewGenerateFavicon:
		return m.generateFavicon.View()
	case ViewColorPalettes:
		return m.colorPalettes.View()
	case ViewSettings:
		return m.settings.View()
	}
	return "Unknown view"
}

// homeItemToView maps menu item index to a View.
func homeItemToView(idx int) View {
	switch idx {
	case 0:
		return ViewBrowsePatterns
	case 1:
		return ViewBrowseArchitectures
	case 2:
		return ViewAnalyzeLayout
	case 3:
		return ViewGenerateLayout
	case 4:
		return ViewGenerateImages
	case 5:
		return ViewOptimizeImages
	case 6:
		return ViewGenerateFavicon
	case 7:
		return ViewColorPalettes
	case 8:
		return ViewSettings
	}
	return ViewHome
}

// Shared styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			Background(lipgloss.Color("18"))

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			MarginTop(1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12")).
			Padding(1, 2)
)
