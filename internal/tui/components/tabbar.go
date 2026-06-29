package components

import (
	"github.com/charmbracelet/lipgloss"
	"ai-manager/internal/tui/styles"
)

type Tab struct {
	Label     string
	Shortcut  string
	Active    bool
}

type TabBar struct {
	Tabs []Tab
}

func NewTabBar(tabs []Tab) TabBar {
	return TabBar{Tabs: tabs}
}

func (tb TabBar) Render() string {
	var tabs []string
	for _, tab := range tb.Tabs {
		if tab.Active {
			tabs = append(tabs, styles.ActiveTabStyle.Render(tab.Label))
		} else {
			tabs = append(tabs, styles.TabStyle.Render(tab.Label))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}
