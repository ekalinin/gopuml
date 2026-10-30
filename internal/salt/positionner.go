// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/Cell.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/Positionner2.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package salt

import (
	"fmt"

	"github.com/ekalinin/gopuml/internal/uml"
)

// cell is the rows and columns that an element of a group spans.
type cell struct {
	minRow, maxRow int
	minCol, maxCol int
}

func newCell(row, col int) *cell { return &cell{minRow: row, maxRow: row, minCol: col, maxCol: col} }

func (c *cell) mergeLeft()  { c.maxCol++ }
func (c *cell) nbRows() int { return c.maxRow - c.minRow + 1 }
func (c *cell) nbCols() int { return c.maxCol - c.minCol + 1 }

func (c *cell) String() string {
	return fmt.Sprintf("(%d,%d)-(%d,%d)", c.minRow, c.minCol, c.maxRow, c.maxCol)
}

// positioned is an element of a group with its cell.
type positioned struct {
	element element
	cell    *cell
}

// positionner2 is upstream's Positionner2: it places the elements of a group
// in rows and columns, in the order of their tokens. The elements are kept in
// insertion order, as upstream's LinkedHashMap.
type positionner2 struct {
	row, col       int
	maxRow, maxCol int
	positions      []positioned
	last           *cell
}

func (p *positionner2) add(e terminated[element]) {
	p.addWithoutMove(e.element)
	if e.terminator == newCol {
		p.moveNextColumn()
	} else {
		p.moveNextRow()
	}
}

func (p *positionner2) moveNextColumn() { p.col++ }

func (p *positionner2) moveNextRow() {
	p.row++
	p.col = 0
}

// addWithoutMove places the element in a new cell. Upstream's map is keyed
// by the element, and every token makes a new element, so each one is added
// once.
func (p *positionner2) addWithoutMove(e element) {
	p.last = newCell(p.row, p.col)
	p.positions = append(p.positions, positioned{element: e, cell: p.last})
	p.updateMax()
}

// mergeLeft extends the last cell over the next column, for a "*" cell. A
// group that starts with "*" has no last cell: upstream throws a
// NullPointerException.
func (p *positionner2) mergeLeft(t terminator) {
	p.updateMax()
	if t == newCol {
		p.col++
	} else {
		p.row++
		p.col = 0
	}
	if p.last == nil {
		uml.ThrowNPE(`Cannot invoke "net.sourceforge.plantuml.salt.Cell.mergeLeft()" because "this.last" is null`)
	}
	p.last.mergeLeft()
}

func (p *positionner2) updateMax() {
	if p.row > p.maxRow {
		p.maxRow = p.row
	}
	if p.col > p.maxCol {
		p.maxCol = p.col
	}
}

func (p *positionner2) all() []positioned { return p.positions }
func (p *positionner2) nbRows() int       { return p.maxRow + 1 }
func (p *positionner2) nbCols() int       { return p.maxCol + 1 }
