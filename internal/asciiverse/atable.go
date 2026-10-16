// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/asciiverse/ATable.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/asciiverse/DataTable.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package asciiverse

import "github.com/ekalinin/plantuml-in-go/internal/klimt"

// ATable is a table of displays drawn with ASCII borders (Gantt and WBS
// text output upstream). A cell shows the first line of its display; cells
// may be merged.
type ATable struct {
	cells           *dataTable[*klimt.Display]
	mergedCells     *dataTable[*cellSpan]
	columnMinWidths []int
	rowMinHeights   []int
	leftPadding     int
	rightPadding    int
}

type cellSpan struct {
	startRow, startCol int
	rowSpan, colSpan   int
}

// NewATable returns a table of at least minRows rows and minCols columns,
// each one character wide and high.
func NewATable(minRows, minCols int) *ATable {
	t := &ATable{cells: newDataTable(klimt.DisplayNULL()), mergedCells: newDataTable[*cellSpan](nil)}
	for range minCols {
		t.columnMinWidths = append(t.columnMinWidths, 1)
	}
	for range minRows {
		t.rowMinHeights = append(t.rowMinHeights, 1)
	}
	return t
}

// WithLeftPadding sets the blank columns before each cell text.
func (t *ATable) WithLeftPadding(leftPadding int) *ATable {
	t.leftPadding = max(leftPadding, 0)
	return t
}

// WithRightPadding sets the blank columns after each cell text.
func (t *ATable) WithRightPadding(rightPadding int) *ATable {
	t.rightPadding = max(rightPadding, 0)
	return t
}

// WithColumnMinWidth sets the width of a column, at least 1.
func (t *ATable) WithColumnMinWidth(col, width int) *ATable {
	t.columnMinWidths[col] = max(width, 1)
	return t
}

// WithRowMinHeight sets the height of a row, at least 1.
func (t *ATable) WithRowMinHeight(row, height int) *ATable {
	t.rowMinHeights[row] = max(height, 1)
	return t
}

// AddRow adds a row.
func (t *ATable) AddRow() *ATable {
	t.rowMinHeights = append(t.rowMinHeights, 1)
	return t
}

// AddColumn adds a column.
func (t *ATable) AddColumn() *ATable {
	t.columnMinWidths = append(t.columnMinWidths, 1)
	return t
}

// SetCellContent sets the display of a cell, growing the table and the
// cell's column and row to fit it.
func (t *ATable) SetCellContent(row, col int, content *klimt.Display) {
	t.cells.set(row, col, content)
	for col >= len(t.columnMinWidths) {
		t.columnMinWidths = append(t.columnMinWidths, 1)
	}
	for row >= len(t.rowMinHeights) {
		t.rowMinHeights = append(t.rowMinHeights, 1)
	}
	t.WithColumnMinWidth(col, max(t.columnMinWidths[col], DisplayContentWidth(content)))
	t.WithRowMinHeight(row, max(t.rowMinHeights[row], content.Size()))
}

// CellContent returns the display of a cell, Display.NULL for none.
func (t *ATable) CellContent(row, col int) *klimt.Display { return t.cells.get(row, col) }

// MergeCells merges a block of cells into its top left one.
func (t *ATable) MergeCells(startRow, startCol, rowSpan, colSpan int) {
	span := &cellSpan{startRow: startRow, startCol: startCol, rowSpan: rowSpan, colSpan: colSpan}
	for i := startRow; i < startRow+rowSpan; i++ {
		for j := startCol; j < startCol+colSpan; j++ {
			t.mergedCells.set(i, j, span)
		}
	}
}

func (t *ATable) columnRenderWidth(col int) int {
	return t.columnMinWidths[col] + t.leftPadding + t.rightPadding
}

// verticalBorderOffsets returns the x offsets of the vertical borders.
func (t *ATable) verticalBorderOffsets() []int {
	result := make([]int, len(t.columnMinWidths)+1)
	x := 0
	result[0] = x
	for col := range t.columnMinWidths {
		x += t.columnRenderWidth(col) + 1
		result[col+1] = x
	}
	return result
}

// drawHorizontalBorder draws a border row with its junctions; position is
// -1 for the top edge, 0 for an inner separator, 1 for the bottom edge.
func (t *ATable) drawHorizontalBorder(plan *InfinitePlan, startX, currentY, position int) {
	borders := t.verticalBorderOffsets()
	right := borders[len(borders)-1]
	for x := 0; x <= right; x++ {
		plan.Move(startX+x, currentY).DrawChar('-')
	}
	for b := range borders {
		isLeft := b == 0
		isRight := b == len(borders)-1
		var junction uint16 = '+'
		if position < 0 {
			switch {
			case isLeft:
				junction = ','
			case isRight:
				junction = '.'
			}
		} else if position > 0 {
			switch {
			case isLeft:
				junction = '`'
			case isRight:
				junction = '\''
			}
		}
		plan.Move(startX+borders[b], currentY).DrawChar(junction)
	}
}

// TotalWidth returns the width of the table, borders included.
func (t *ATable) TotalWidth() int {
	totalWidth := 0
	for col := range t.columnMinWidths {
		totalWidth += t.columnRenderWidth(col)
	}
	return totalWidth + len(t.columnMinWidths) + 1
}

// TotalHeight returns the sum of the row heights, without borders.
func (t *ATable) TotalHeight() int {
	totalHeight := 0
	for _, height := range t.rowMinHeights {
		totalHeight += height
	}
	return totalHeight
}

// AsciiDimension returns the size of the table, borders included.
func (t *ATable) AsciiDimension() ADimension2D {
	return NewADimension2D(t.TotalWidth(), t.TotalHeight()+2+len(t.rowMinHeights)-1)
}

// AsciiDraw draws the table.
func (t *ATable) AsciiDraw(plan *InfinitePlan) {
	currentY := 0
	t.drawHorizontalBorder(plan, 0, currentY, -1)
	currentY++
	for row := 0; row < len(t.rowMinHeights); row++ {
		currentX := 0
		rowHeight := t.rowMinHeights[row]
		for col := 0; col < len(t.columnMinWidths); col++ {
			span := t.mergedCells.get(row, col)
			if span != nil && span.startRow == row && span.startCol == col {
				content := t.cells.get(row, col)
				cellWidth := 0
				for c := col; c < col+span.colSpan; c++ {
					cellWidth += t.columnRenderWidth(c)
				}
				cellWidth += span.colSpan - 1
				plan.Move(currentX, currentY).DrawChar('|')
				currentX++
				line0 := firstLine(content)
				for i := range cellWidth {
					plan.Move(currentX+i, currentY).DrawChar(charAt(content, line0, i-t.leftPadding))
				}
				currentX += cellWidth
				col += span.colSpan - 1
			} else if span == nil {
				content := t.cells.get(row, col)
				cellWidth := t.columnRenderWidth(col)
				plan.Move(currentX, currentY).DrawChar('|')
				currentX++
				line0 := firstLine(content)
				for i := range cellWidth {
					plan.Move(currentX+i, currentY).DrawChar(charAt(content, line0, i-t.leftPadding))
				}
				currentX += cellWidth
			}
		}
		plan.Move(currentX, currentY).DrawChar('|')
		for h := 1; h < rowHeight; h++ {
			currentY++
			currentX = 0
			for col := 0; col < len(t.columnMinWidths); col++ {
				span := t.mergedCells.get(row, col)
				if span != nil && span.startRow == row && span.startCol == col {
					cellWidth := 0
					for c := col; c < col+span.colSpan; c++ {
						cellWidth += t.columnRenderWidth(c)
					}
					cellWidth += span.colSpan - 1
					plan.Move(currentX, currentY).DrawChar('|')
					currentX++
					for i := range cellWidth {
						plan.Move(currentX+i, currentY).DrawChar(' ')
					}
					currentX += cellWidth
					col += span.colSpan - 1
				} else if span == nil {
					cellWidth := t.columnRenderWidth(col)
					plan.Move(currentX, currentY).DrawChar('|')
					currentX++
					for i := range cellWidth {
						plan.Move(currentX+i, currentY).DrawChar(' ')
					}
					currentX += cellWidth
				}
			}
			plan.Move(currentX, currentY).DrawChar('|')
		}
		currentY++
		position := 0
		if row == len(t.rowMinHeights)-1 {
			position = 1
		}
		t.drawHorizontalBorder(plan, 0, currentY, position)
		currentY++
	}
}

// firstLine returns the chars of the first line of a display, nil for an
// empty one.
func firstLine(content *klimt.Display) []uint16 {
	if content.Size() == 0 {
		return nil
	}
	return javaChars(displayLine(content, 0))
}

// charAt is the character i of the first line, a space outside it.
func charAt(content *klimt.Display, line0 []uint16, i int) uint16 {
	if i < 0 {
		return ' '
	}
	if content.Size() == 0 {
		return ' '
	}
	if i < len(line0) {
		return line0[i]
	}
	return ' '
}

// dataTable is a sparse table of values with a default value (upstream's
// DataTable, a HashMap: its order never shows).
type dataTable[E any] struct {
	cells        map[cellKey]E
	defaultValue E
}

type cellKey struct{ x, y int }

func newDataTable[E any](defaultValue E) *dataTable[E] {
	return &dataTable[E]{cells: map[cellKey]E{}, defaultValue: defaultValue}
}

func (t *dataTable[E]) set(x, y int, value E) { t.cells[cellKey{x, y}] = value }

func (t *dataTable[E]) get(x, y int) E {
	if v, ok := t.cells[cellKey{x, y}]; ok {
		return v
	}
	return t.defaultValue
}

func (t *dataTable[E]) contains(x, y int) bool {
	_, ok := t.cells[cellKey{x, y}]
	return ok
}
