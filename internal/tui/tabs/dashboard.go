package tabs

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ai-manager/internal/process"
	"ai-manager/internal/tui/styles"
	"ai-manager/internal/tui/types"
	"ai-manager/internal/tui/util"
)

type DashboardTab struct {
	ctx         *types.AppContext
	data        types.DashboardData
	loading     bool
	errMsg      string
	refreshCmd  tea.Cmd
}

func NewDashboardTab(ctx *types.AppContext) DashboardTab {
	return DashboardTab{
		ctx:     ctx,
		loading: true,
	}
}

func (t *DashboardTab) Init() tea.Cmd {
	t.refreshCmd = RefreshDashboard(t.ctx)
	return t.refreshCmd
}

func (t *DashboardTab) Update(msg tea.Msg) (TabInterface, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "r" {
			t.loading = true
			t.errMsg = ""
			return t, t.refreshCmd
		}
	case refreshDashboardMsg:
		t.data = msg.data
		t.loading = false
	}
	return t, nil
}

func (t *DashboardTab) View() string {
	if t.loading {
		return styles.HelpText.Render("Loading...")
	}

	if t.errMsg != "" {
		return lipgloss.NewStyle().Foreground(styles.ErrorColor).Render(t.errMsg)
	}

	var lines []string

	lines = append(lines, "╭─ System Resources ─────────────────────────────────────╮")
	if t.data.System != nil {
		lines = append(lines, fmt.Sprintf("│  CPU:   %s                                        │",
			util.FormatPercentage(t.data.System.CPU)))
		lines = append(lines, fmt.Sprintf("│  Memory: %s                                    │",
			util.FormatMemoryMB(t.data.System.MemoryMB)))
		if t.data.System.GPU != nil {
			lines = append(lines, fmt.Sprintf("│  GPU:   %s (%.0f%% util, %d/%d MB VRAM)    │",
				t.data.System.GPU.Name,
				t.data.System.GPU.Utilization,
				t.data.System.GPU.MemoryUsedMB,
				t.data.System.GPU.MemoryTotalMB))
		}
	}
	lines = append(lines, "╰────────────────────────────────────────────────────────╯")
	lines = append(lines, "")

	lines = append(lines, "╭─ Ollama Models ────────────────────────────────────────╮")
	if len(t.data.OllamaModels) == 0 {
		lines = append(lines, "│  No models found                                      │")
	} else {
		for _, m := range t.data.OllamaModels {
			lines = append(lines, fmt.Sprintf("│  %-50s │", m.Name))
		}
	}
	lines = append(lines, "╰────────────────────────────────────────────────────────╯")
	lines = append(lines, "")

	lines = append(lines, "╭─ Running Services ─────────────────────────────────────╮")
	if len(t.data.Processes) == 0 {
		lines = append(lines, "│  No services running                                  │")
	} else {
		for _, p := range t.data.Processes {
			stateColor := styles.StatusStopped
			if p.State == process.StateRunning {
				stateColor = styles.StatusRunning
			}
			lines = append(lines, fmt.Sprintf("│  %-20s %s                              │", p.Name, stateColor.Render(string(p.State))))
		}
	}
	lines = append(lines, "╰────────────────────────────────────────────────────────╯")

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

type refreshDashboardMsg struct {
	data types.DashboardData
}

func RefreshDashboard(ctx *types.AppContext) tea.Cmd {
	return func() tea.Msg {
		data := types.DashboardData{
			Uptime: time.Since(ctx.StartTime),
		}

		usage, err := ctx.Monitor.GetUsage()
		if err == nil {
			data.System = usage
		}

		data.Processes = ctx.ProcMgr.List()

		if ctx.Ollama != nil && ctx.Ollama.IsServerRunning() {
			models, err := ctx.Ollama.ListModels()
			if err == nil {
				data.OllamaModels = models
			}
		}

		ctx.LastRefresh = time.Now()
		return refreshDashboardMsg{data: data}
	}
}
