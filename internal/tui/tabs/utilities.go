package tabs

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ai-manager/internal/process"
	"ai-manager/internal/tui/styles"
	"ai-manager/internal/tui/types"
)

type UtilitiesTab struct {
	ctx         *types.AppContext
	subView     string
	selectedIdx int
	logEntries  []string
	logScroll   int
	healthResults []HealthCheckResult
	loading     bool
}

type HealthCheckResult struct {
	ProcessID string
	Healthy   bool
	Latency   time.Duration
	LastCheck time.Time
}

func NewUtilitiesTab(ctx *types.AppContext) UtilitiesTab {
	return UtilitiesTab{
		ctx:     ctx,
		subView: "logs",
		loading: true,
	}
}

func (t *UtilitiesTab) Init() tea.Cmd {
	return t.refreshLogs()
}

func (t *UtilitiesTab) Update(msg tea.Msg) (TabInterface, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			t.subView = "logs"
			return t, t.refreshLogs()
		case "right":
			t.subView = "health"
			return t, t.refreshHealth()
		case "up", "k":
			if t.subView == "logs" && t.logScroll > 0 {
				t.logScroll--
			}
		case "down", "j":
			if t.subView == "logs" && t.logScroll < len(t.logEntries)-1 {
				t.logScroll++
			}
		case "g":
			t.logScroll = 0
		case "G":
			t.logScroll = len(t.logEntries) - 1
		case "r":
			t.loading = true
			return t, t.refreshLogs()
		}
	case refreshLogsMsg:
		t.logEntries = msg.entries
		t.loading = false
	case refreshHealthMsg:
		t.healthResults = msg.results
		t.loading = false
	}
	return t, nil
}

func (t *UtilitiesTab) View() string {
	if t.loading {
		return styles.HelpText.Render("Loading...")
	}

	var lines []string
	lines = append(lines, "╭─ Utilities ────────────────────────────────────────────╮")

	if t.subView == "logs" {
		lines = append(lines, "│  Log Viewer                                            │")
		lines = append(lines, "│  ────────────────────────────────────────────────────  │")
		if len(t.logEntries) == 0 {
			lines = append(lines, "│  No log entries                                        │")
		} else {
			start := t.logScroll
			if start < 0 {
				start = 0
			}
			end := start + 10
			if end > len(t.logEntries) {
				end = len(t.logEntries)
			}
			for _, entry := range t.logEntries[start:end] {
				lines = append(lines, fmt.Sprintf("│  %s", truncate(entry, 52)))
			}
			if t.logScroll > 0 {
				lines = append(lines, "│  ↑ Scroll up                                           │")
			}
			if t.logScroll+len(t.logEntries[start:end]) < len(t.logEntries) {
				lines = append(lines, "│  ↓ Scroll down                                         │")
			}
		}
	} else {
		lines = append(lines, "│  Health Checks                                           │")
		lines = append(lines, "│  ────────────────────────────────────────────────────  │")
		if len(t.healthResults) == 0 {
			lines = append(lines, "│  No processes to check                                   │")
		} else {
			for _, r := range t.healthResults {
				status := styles.StatusStopped.Render("✗")
				if r.Healthy {
					status = styles.StatusRunning.Render("✓")
				}
				lines = append(lines, fmt.Sprintf("│  %s %s (%v)                                     │",
					status, r.ProcessID, r.Latency.Round(time.Millisecond)))
			}
		}
	}

	lines = append(lines, "╰────────────────────────────────────────────────────────╯")
	lines = append(lines, "")
	lines = append(lines, "  ←/→ Switch view  ↑/↓ Navigate  g/G Top/Bottom  r: Refresh")

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (t *UtilitiesTab) refreshLogs() tea.Cmd {
	return func() tea.Msg {
		return refreshLogsMsg{entries: []string{"Log viewer - select a process to view logs"}}
	}
}

func (t *UtilitiesTab) refreshHealth() tea.Cmd {
	return func() tea.Msg {
		processes := t.ctx.ProcMgr.List()
		var results []HealthCheckResult
		for _, p := range processes {
			if p.State == process.StateRunning {
				healthy, _ := t.ctx.ProcMgr.HealthCheck(p.ID)
				results = append(results, HealthCheckResult{
					ProcessID: p.ID,
					Healthy:   healthy,
					LastCheck: time.Now(),
				})
			}
		}
		return refreshHealthMsg{results: results}
	}
}

type refreshLogsMsg struct {
	entries []string
	err     error
}

type refreshHealthMsg struct {
	results []HealthCheckResult
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
