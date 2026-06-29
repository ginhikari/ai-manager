package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

type Sparkline struct {
	Data    []float64
	Width   int
	Height  int
	Color   lipgloss.Style
	Label   string
}

func NewSparkline(data []float64, width, height int, color lipgloss.Style, label string) Sparkline {
	return Sparkline{
		Data:   data,
		Width:  width,
		Height: height,
		Color:  color,
		Label:  label,
	}
}

func (s Sparkline) Render() string {
	if len(s.Data) == 0 {
		return ""
	}

	var lines []string
	for row := s.Height - 1; row >= 0; row-- {
		line := ""
		for i := 0; i < s.Width; i++ {
			idx := i * (len(s.Data) - 1) / (s.Width - 1)
			if idx >= len(s.Data) {
				idx = len(s.Data) - 1
			}

			val := s.Data[idx]
			maxVal := 100.0
			for _, v := range s.Data {
				if v > maxVal {
					maxVal = v
				}
			}
			if maxVal == 0 {
				maxVal = 1
			}

			normalized := val / maxVal * float64(s.Height-1)
			if int(normalized) == row {
				line += s.Color.Render("█")
			} else {
				line += " "
			}
		}
		lines = append(lines, line)
	}

	label := ""
	if s.Label != "" {
		label = fmt.Sprintf("%s: ", s.Label)
	}

	result := label
	for _, line := range lines {
		result += "\n" + line
	}
	return result
}
