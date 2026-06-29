package components

import (
	"strings"

	"ai-manager/internal/tui/styles"
)

type TableColumn struct {
	Header string
	Width  int
}

type Table struct {
	Columns []TableColumn
	Rows    [][]string
	Selected int
	Width    int
}

func NewTable(columns []TableColumn, rows [][]string, width int) Table {
	return Table{
		Columns: columns,
		Rows:    rows,
		Selected: 0,
		Width:   width,
	}
}

func (t *Table) SetSelected(idx int) {
	if idx < 0 {
		idx = 0
	}
	if idx >= len(t.Rows) {
		idx = len(t.Rows) - 1
	}
	t.Selected = idx
}

func (t *Table) Next() {
	if t.Selected < len(t.Rows)-1 {
		t.Selected++
	}
}

func (t *Table) Prev() {
	if t.Selected > 0 {
		t.Selected--
	}
}

func (t Table) Render() string {
	if len(t.Columns) == 0 {
		return ""
	}

	var lines []string

	header := ""
	for i, col := range t.Columns {
		if i > 0 {
			header += "  "
		}
		header += styles.TableHeader.Render(strings.TrimSpace(col.Header))
	}
	lines = append(lines, header)

	for i, row := range t.Rows {
		line := ""
		for j, cell := range row {
			if j > 0 {
				line += "  "
			}
			if i == t.Selected {
				line += styles.TableRowSelected.Render(strings.TrimSpace(cell))
			} else {
				line += styles.ValueStyle.Render(strings.TrimSpace(cell))
			}
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}
