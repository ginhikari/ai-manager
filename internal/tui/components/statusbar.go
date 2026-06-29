package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"ai-manager/internal/tui/styles"
)

type StatusBar struct {
	TabName   string
	LastRefresh string
	Connection string
	RefreshRate string
}

func NewStatusBar(tabName, lastRefresh, connection, refreshRate string) StatusBar {
	return StatusBar{
		TabName:     tabName,
		LastRefresh: lastRefresh,
		Connection:  connection,
		RefreshRate: refreshRate,
	}
}

func (sb StatusBar) Render() string {
	tab := styles.KeyStyle.Render(fmt.Sprintf("Tab: %s", sb.TabName))
	refreshInfo := styles.HelpText.Render(fmt.Sprintf("Refresh: %s", sb.RefreshRate))
	conn := styles.HelpText.Render(fmt.Sprintf("Ollama: %s", sb.Connection))

	return lipgloss.JoinHorizontal(lipgloss.Top,
		tab,
		"  |  ",
		conn,
		"  |  ",
		refreshInfo,
	)
}
