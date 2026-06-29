package tabs

import (
	"fmt"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ai-manager/internal/monitor"
	"ai-manager/internal/tui/styles"
	"ai-manager/internal/tui/types"
	"ai-manager/internal/tui/util"
)

type ResourcesTab struct {
	ctx         *types.AppContext
	current     *monitor.ResourceUsage
	selectedIdx int
	loading     bool
}

func NewResourcesTab(ctx *types.AppContext) ResourcesTab {
	return ResourcesTab{
		ctx:     ctx,
		loading: true,
	}
}

func (t *ResourcesTab) Init() tea.Cmd {
	return t.refreshResources()
}

func (t *ResourcesTab) Update(msg tea.Msg) (TabInterface, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if t.selectedIdx > 0 {
				t.selectedIdx--
			}
		case "down", "j":
			// No history to navigate
		case "r":
			t.loading = true
			return t, t.refreshResources()
		}
	case refreshResourcesMsg:
		t.current = msg.current
		t.loading = false
	}
	return t, nil
}

func (t *ResourcesTab) View() string {
	if t.loading {
		return styles.HelpText.Render("Loading resources...")
	}

	var lines []string
	lines = append(lines, "╭─ Resource Monitor ─────────────────────────────────────╮")

	if t.current == nil {
		lines = append(lines, "│  No resource data available                              │")
	} else {
		lines = append(lines, fmt.Sprintf("│  CPU:   %s                                        │",
			util.FormatPercentage(t.current.CPU)))
		lines = append(lines, fmt.Sprintf("│  Memory: %s                                      │",
			util.FormatMemoryMB(t.current.MemoryMB)))
		if t.current.GPU != nil {
			lines = append(lines, fmt.Sprintf("│  GPU:   %s                                        │",
				t.current.GPU.Name))
			lines = append(lines, fmt.Sprintf("│        %s (%.0f%% util)                    │",
				util.FormatMemoryMB(t.current.GPU.MemoryUsedMB),
				t.current.GPU.Utilization))
			lines = append(lines, fmt.Sprintf("│        VRAM: %s / %s                           │",
				util.FormatMemoryMB(t.current.GPU.MemoryUsedMB),
				util.FormatMemoryMB(t.current.GPU.MemoryTotalMB)))
		}
	}

	lines = append(lines, "╰────────────────────────────────────────────────────────╯")
	lines = append(lines, "")
	lines = append(lines, "  ↑/↓ Navigate  r: Refresh")

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (t *ResourcesTab) refreshResources() tea.Cmd {
	return func() tea.Msg {
		usage, err := t.ctx.Monitor.GetUsage()
		if err != nil {
			return refreshResourcesMsg{current: nil}
		}
		return refreshResourcesMsg{current: usage}
	}
}

type refreshResourcesMsg struct {
	current *monitor.ResourceUsage
}
