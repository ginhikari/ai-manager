package tabs

import (
	"fmt"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ai-manager/internal/config"
	"ai-manager/internal/tui/styles"
	"ai-manager/internal/tui/types"
)

type SettingsTab struct {
	ctx         *types.AppContext
	config      *config.AppConfig
	selectedIdx int
	loading     bool
}

func NewSettingsTab(ctx *types.AppContext) SettingsTab {
	return SettingsTab{
		ctx:     ctx,
		config:  ctx.ConfigMgr.GetConfig(),
		loading: true,
	}
}

func (t *SettingsTab) Init() tea.Cmd {
	return nil
}

func (t *SettingsTab) Update(msg tea.Msg) (TabInterface, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if t.selectedIdx > 0 {
				t.selectedIdx--
			}
		case "down", "j":
			if t.selectedIdx < len(t.config.Processes)-1 {
				t.selectedIdx++
			}
		case "enter":
			if len(t.config.Processes) > 0 {
				p := t.config.Processes[t.selectedIdx]
				t.ctx.ProcMgr.Stop(p.ID)
			}
		}
	}
	return t, nil
}

func (t *SettingsTab) View() string {
	var lines []string
	lines = append(lines, "╭─ Settings ─────────────────────────────────────────────╮")

	lines = append(lines, "│  General                                                │")
	lines = append(lines, fmt.Sprintf("│    Config path:   %s                                │",
		config.DefaultConfigPath()))
	lines = append(lines, "")

	lines = append(lines, "│  Services                                               │")
	if len(t.config.Processes) == 0 {
		lines = append(lines, "│    No services configured                               │")
	} else {
		for i, p := range t.config.Processes {
			selected := "  "
			if i == t.selectedIdx {
				selected = "▸ "
			}
			lines = append(lines, fmt.Sprintf("│    %s%-30s %s                                │",
				selected,
				p.Name,
				p.Command))
		}
	}

	lines = append(lines, "╰────────────────────────────────────────────────────────╯")
	lines = append(lines, "")
	lines = append(lines, "  ↑/↓ Navigate  Enter: Stop service")

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func boolStr(b bool) string {
	if b {
		return lipgloss.NewStyle().Foreground(styles.SuccessColor).Render("Enabled")
	}
	return lipgloss.NewStyle().Foreground(styles.MutedColor).Render("Disabled")
}
