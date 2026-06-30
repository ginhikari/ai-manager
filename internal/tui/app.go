package tui

import (
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ai-manager/internal/tui/components"
	"ai-manager/internal/tui/tabs"
	"ai-manager/internal/tui/types"
)

const (
	tabDashboard = iota
	tabModels
	tabUtilities
	tabResources
	tabSettings
)

type App struct {
	ctx        *types.AppContext
	currentTab int
	tabs       []tabs.TabInterface
	width      int
	height     int
	confirm    components.ConfirmDialog
	refreshCmd tea.Cmd
}



func NewApp(ctx *types.AppContext) *App {
	dashboardTab := tabs.NewDashboardTab(ctx)
	modelsTab := tabs.NewModelsTab(ctx)
	utilitiesTab := tabs.NewUtilitiesTab(ctx)
	resourcesTab := tabs.NewResourcesTab(ctx)
	settingsTab := tabs.NewSettingsTab(ctx)

	tabsList := []tabs.TabInterface{
		&dashboardTab,
		&modelsTab,
		&utilitiesTab,
		&resourcesTab,
		&settingsTab,
	}

	return &App{
		ctx:        ctx,
		currentTab: tabDashboard,
		tabs:       tabsList,
		confirm:    components.NewConfirmDialog("Confirm", "Are you sure?"),
	}
}

func (a *App) Init() tea.Cmd {
	a.ctx.StartTime = time.Now()
	a.refreshCmd = func() tea.Msg {
		return refreshAppMsg{}
	}

	// Initialize each tab and collect their commands
	var cmds []tea.Cmd
	for _, tab := range a.tabs {
		cmd := tab.Init()
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return a.refreshCmd
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, nil
		case "1":
			a.currentTab = tabDashboard
		case "2":
			a.currentTab = tabModels
		case "3":
			a.currentTab = tabUtilities
		case "4":
			a.currentTab = tabResources
		case "5":
			a.currentTab = tabSettings
		case "r":
			a.refreshCmd = func() tea.Msg {
				return refreshAppMsg{}
			}
			return a, a.refreshCmd
		}
	case refreshAppMsg:
		a.ctx.LastRefresh = time.Now()
		return a, nil
	}

	if a.currentTab >= 0 && a.currentTab < len(a.tabs) {
		tab := a.tabs[a.currentTab]
		updatedTab, cmd := tab.Update(msg)
		a.tabs[a.currentTab] = updatedTab
		return a, cmd
	}

	return a, nil
}

func (a *App) View() string {
	if a.currentTab < 0 || a.currentTab >= len(a.tabs) {
		return "Invalid tab"
	}

	tab := a.tabs[a.currentTab]
	content := tab.View()

	tabBar := a.renderTabBar()
	statusBar := a.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left,
		tabBar,
		"",
		content,
		"",
		statusBar,
	)
}

func (a *App) renderTabBar() string {
	tabs := []components.Tab{
		{Label: "Dashboard", Shortcut: "1", Active: a.currentTab == tabDashboard},
		{Label: "Models", Shortcut: "2", Active: a.currentTab == tabModels},
		{Label: "Utilities", Shortcut: "3", Active: a.currentTab == tabUtilities},
		{Label: "Resources", Shortcut: "4", Active: a.currentTab == tabResources},
		{Label: "Settings", Shortcut: "5", Active: a.currentTab == tabSettings},
	}

	return components.NewTabBar(tabs).Render()
}

func (a *App) renderStatusBar() string {
	connection := "Disconnected"
	if a.ctx.Ollama != nil && a.ctx.Ollama.IsServerRunning() {
		connection = "Connected"
	}

	tabNames := []string{"Dashboard", "Models", "Utilities", "Resources", "Settings"}
	tabName := tabNames[a.currentTab]

	return components.NewStatusBar(tabName, a.ctx.LastRefresh.Format("15:04:05"), connection, "1s").Render()
}

type refreshAppMsg struct{}
