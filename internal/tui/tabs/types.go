package tabs

import "github.com/charmbracelet/bubbletea"

type TabInterface interface {
	Init() tea.Cmd
	Update(tea.Msg) (TabInterface, tea.Cmd)
	View() string
}
