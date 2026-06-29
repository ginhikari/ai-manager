package components

import (
	"github.com/charmbracelet/lipgloss"
	"ai-manager/internal/tui/styles"
)

type ConfirmDialog struct {
	Title   string
	Message string
	Width   int
	Visible bool
	YesLabel string
	NoLabel  string
}

func NewConfirmDialog(title, message string) ConfirmDialog {
	return ConfirmDialog{
		Title:    title,
		Message:  message,
		Width:    50,
		Visible:  false,
		YesLabel: "Yes",
		NoLabel:  "No",
	}
}

func (d ConfirmDialog) Render() string {
	if !d.Visible {
		return ""
	}

	title := lipgloss.NewStyle().Foreground(styles.WarningColor).Render("⚠ " + d.Title)
	msg := d.Message
	footer := "  [" + styles.KeyStyle.Render(d.YesLabel) + "] Yes  [" + styles.KeyStyle.Render(d.NoLabel) + "] No  "

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.WarningColor).
		Padding(1, 2).
		Width(d.Width)

	return box.Render(lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		msg,
		"",
		footer,
	))
}
