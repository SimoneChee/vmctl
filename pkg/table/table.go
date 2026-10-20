// Package table provides formatted table output for vmctl
package table

import (
	"fmt"
	"io"
	"strings"

	"github.com/olekukonez/tablewriter"
)

// Writer wraps tablewriter for consistent formatting
type Writer struct {
	table *tablewriter.Table
}

// NewWriter creates a new table writer with vmctl default styling
func NewWriter(w io.Writer) *Writer {
	t := tablewriter.NewWriter(w)
	t.SetHeaderAlignment(tablewriter.ALIGN_CENTER)
	t.SetAlignment(tablewriter.ALIGN_LEFT)
	t.SetCenterSeparator("─")
	t.SetColumnSeparator("│")
	t.SetRowSeparator("")
	t.SetHeaderLine(true)
	t.SetBorder(false)
	t.SetTablePadding("\t")
	t.SetNoWhiteSpace(true)
	t.SetAutoWrapText(false)
	return &Writer{table: t}
}

// SetHeader sets the table header
func (w *Writer) SetHeader(headers []string) {
	w.table.SetHeader(headers)
}

// AddRow adds a row to the table
func (w *Writer) AddRow(row []string) {
	w.table.Append(row)
}

// Render renders the table to the output
func (w *Writer) Render() {
	w.table.Render()
}

// SetColWidth sets the minimum width for a column
func (w *Writer) SetColWidth(col int, width int) {
	widths := w.table.GetColWidth()
	if col < len(widths) {
		widths[col] = width
		w.table.SetColWidth(widths)
	}
}

// SimpleTable renders a simple data table without external dependencies
type SimpleTable struct {
	headers []string
	rows    [][]string
	colWidths []int
}

// NewSimpleTable creates a simple text-based table
func NewSimpleTable(headers []string) *SimpleTable {
	return &SimpleTable{
		headers:   headers,
		rows:      make([][]string, 0),
		colWidths: make([]int, len(headers)),
	}
}

// AddRow adds a row of data
func (t *SimpleTable) AddRow(row []string) {
	t.rows = append(t.rows, row)
	for i, cell := range row {
		if len(cell) > t.colWidths[i] {
			t.colWidths[i] = len(cell)
		}
	}
}

// String returns the formatted table as a string
func (t *SimpleTable) String() string {
	var sb strings.Builder

	// Calculate column widths based on headers too
	for i, h := range t.headers {
		if len(h) > t.colWidths[i] {
			t.colWidths[i] = len(h)
		}
	}

	// Separator line
	sep := "─"
	rowSep := "┼"
	topBorder := "┌"
	bottomBorder := "└"

	var sepParts []string
	for i, w := range t.colWidths[i:] {
		sepParts = append(sepParts, strings.Repeat(sep, w+2))
	}
	
	topLine := topBorder + strings.Join(sepParts, "┬") + "┐"
	midLine := "├" + strings.Join(sepParts, rowSep) + "┤"
	botLine := bottomBorder + strings.Join(sepParts, "┴") + "┘"

	// Header
	sb.WriteString(topLine + "\n")
	sb.WriteString("│")
	for i, h := range t.headers {
		sb.WriteString(fmt.Sprintf(" %-*s │", t.colWidths[i], h))
	}
	sb.WriteString("\n")
	sb.WriteString(midLine + "\n")

	// Rows
	for _, row := range t.rows {
		sb.WriteString("│")
		for i, cell := range row {
			if i < len(t.colWidths) {
				sb.WriteString(fmt.Sprintf(" %-*s │", t.colWidths[i], cell))
			}
		}
		sb.WriteString("\n")
	}
	sb.WriteString(botLine + "\n")

	return sb.String()
}

// Render writes the table to an io.Writer
func (t *SimpleTable) Render(w io.Writer) {
	io.WriteString(w, t.String())
}
