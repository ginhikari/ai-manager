package util

import (
	"fmt"
	"math"
	"strings"
	"time"
)

func FormatSize(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

func FormatPercentage(pct float64) string {
	return fmt.Sprintf("%.1f%%", pct)
}

func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func FormatTime(t time.Time) string {
	return t.Format("15:04:05")
}

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func FormatMemoryMB(mb uint64) string {
	if mb < 1024 {
		return fmt.Sprintf("%d MB", mb)
	}
	return fmt.Sprintf("%.1f GB", float64(mb)/1024.0)
}

func RepeatChar(char byte, count int) string {
	return strings.Repeat(string(char), count)
}

func Progressbar(percentage float64, width int) string {
	filled := int(math.Round(percentage / 100.0 * float64(width)))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}
