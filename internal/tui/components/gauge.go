package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"ai-manager/internal/tui/styles"
)

type Gauge struct {
	Label     string
	Percentage float64
	Width     int
	Color     lipgloss.Style
	EmptyStyle lipgloss.Style
}

func NewGauge(label string, percentage float64, width int) Gauge {
	return Gauge{
		Label:      label,
		Percentage: percentage,
		Width:      width,
		Color:      styles.GaugeFull,
		EmptyStyle: styles.GaugeEmpty,
	}
}

func (g Gauge) Render() string {
	filled := int(g.Percentage / 100.0 * float64(g.Width))
	empty := g.Width - filled

	filledBar := strings.Repeat("█", filled)
	emptyBar := strings.Repeat("░", empty)

	label := fmt.Sprintf("%s %6.1f%%", g.Label, g.Percentage)
	return lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Foreground(styles.FgColor).Render(label),
		" ",
		g.Color.Render(filledBar)+g.EmptyStyle.Render(emptyBar),
	)
}
