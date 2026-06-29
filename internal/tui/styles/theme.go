package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	PrimaryColor   = lipgloss.Color("#7C3AED")
	SecondaryColor = lipgloss.Color("#06B6D4")
	SuccessColor   = lipgloss.Color("#10B981")
	WarningColor   = lipgloss.Color("#F59E0B")
	ErrorColor     = lipgloss.Color("#EF4444")
	MutedColor     = lipgloss.Color("#6B7280")
	BgColor        = lipgloss.Color("#1E1E2E")
	FgColor        = lipgloss.Color("#CDD6F4")
	BorderActive   = lipgloss.Color("#7C3AED")
	BorderInactive = lipgloss.Color("#313244")
)

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(FgColor).
			Background(PrimaryColor).
			Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderInactive).
			Padding(1, 2).
			Width(0)

	ActiveBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderActive).
			Padding(1, 2).
			Width(0)

	TabStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			Padding(0, 2)

	ActiveTabStyle = lipgloss.NewStyle().
			Foreground(FgColor).
			Bold(true).
			Padding(0, 2).
			Background(PrimaryColor)

	StatusRunning = lipgloss.NewStyle().Foreground(SuccessColor).Bold(true)
	StatusStopped = lipgloss.NewStyle().Foreground(MutedColor)
	StatusError   = lipgloss.NewStyle().Foreground(ErrorColor).Bold(true)

	GaugeFull = lipgloss.NewStyle().Foreground(PrimaryColor)
	GaugeEmpty = lipgloss.NewStyle().Foreground(BorderInactive)

	TableHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(FgColor).
			Underline(true)

	TableRowSelected = lipgloss.NewStyle().
			Foreground(FgColor).
			Background(BorderInactive)

	ConfirmDialog = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(WarningColor).
			Padding(1, 2).
			Width(50)

	HelpText = lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true)

	KeyStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Bold(true)

	ValueStyle = lipgloss.NewStyle().
			Foreground(FgColor)
)

func BoxWithTitle(title string, width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderActive).
		Padding(1, 2).
		Width(width).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true)
}
