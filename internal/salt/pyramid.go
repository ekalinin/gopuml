// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/ElementPyramid.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/ElementPyramidScrolled.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/TableStrategy.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/LeftFirst.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/TopFirst.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/factory/ScrollStrategy.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package salt

import (
	"sort"
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// tableStrategy is the lines that a group draws, from the character after
// its "{".
type tableStrategy int

const (
	drawNone tableStrategy = iota
	drawOutside
	drawOutsideWithTitle
	drawHorizontal
	drawVertical
	drawAll
	// noStrategy is upstream's null, for a character that is not a strategy.
	noStrategy
)

// tableStrategyFromChar is TableStrategy.fromChar.
func tableStrategyFromChar(c byte) tableStrategy {
	switch c {
	case ' ':
		return drawNone
	case '+':
		return drawOutside
	case '^':
		return drawOutsideWithTitle
	case '-':
		return drawHorizontal
	case '!':
		return drawVertical
	case '#':
		return drawAll
	}
	return noStrategy
}

// elementPyramid is a group: its elements in rows and columns, each column
// as wide as its widest cell and each row as tall as its tallest one, with
// the lines of its strategy and an optional title.
type elementPyramid struct {
	rows, cols    int
	title         klimt.TextBlock
	tableStrategy tableStrategy
	positions     []positioned
	rowsStart     []float64
	colsStart     []float64
}

func newElementPyramid(p *positionner2, strategy tableStrategy, title *string,
	spriteContainer klimt.ISkinSimple) *elementPyramid {
	e := &elementPyramid{positions: p.all(), tableStrategy: strategy}
	if title != nil {
		fontConfiguration := blackBlueTrue(font.ByDefault(10))
		e.title = klimt.DisplayWithNewlines(klimt.PragmaCreateEmpty(), *title).
			Create(fontConfiguration, geom.HorizontalAlignmentLeft, spriteContainer)
	} else {
		e.title = klimt.TextBlockUtilsEmpty(0, 0)
	}
	e.rows = p.nbRows()
	e.cols = p.nbCols()
	for _, pos := range e.positions {
		e.rows = max(e.rows, pos.cell.maxRow)
		e.cols = max(e.cols, pos.cell.maxCol)
	}
	return e
}

func (e *elementPyramid) preferredDimension(stringBounder font.StringBounder, _, _ float64) geom.XDimension2D {
	e.init(stringBounder)
	return geom.NewXDimension2D(e.colsStart[len(e.colsStart)-1],
		e.rowsStart[len(e.rowsStart)-1]+e.title.CalculateDimension(stringBounder).Height())
}

func (e *elementPyramid) drawU(ug klimt.UGraphic, zIndex int, _ *geom.XDimension2D) {
	ug = ug.Apply(saltBlack())
	e.init(ug.StringBounder())
	titleHeight := e.title.CalculateDimension(ug.StringBounder()).Height()
	grid := newGrid(e.rowsStart, e.colsStart, e.tableStrategy, e.title)
	for _, pos := range e.positions {
		c := pos.cell
		xcell := e.colsStart[c.minCol]
		supY := 0.0
		if c.minRow == 0 {
			supY = titleHeight / 2
		}
		ycell := e.rowsStart[c.minRow] + supY
		width := e.colsStart[c.maxCol+1] - e.colsStart[c.minCol] - 1
		height := e.rowsStart[c.maxRow+1] - e.rowsStart[c.minRow] - 1
		grid.addCell(c)
		dim := geom.NewXDimension2D(width, height)
		pos.element.drawU(ug.Apply(klimt.NewUTranslate(xcell+1, ycell+1)), zIndex, &dim)
	}
	if zIndex == 0 {
		grid.drawU(ug, 0, 0, saltWhite())
	}
}

func (e *elementPyramid) init(stringBounder font.StringBounder) {
	if e.rowsStart != nil {
		return
	}
	titleHeight := e.title.CalculateDimension(stringBounder).Height()
	e.rowsStart = make([]float64, e.rows+1)
	for i := range e.rowsStart {
		e.rowsStart[i] = titleHeight / 2
	}
	e.colsStart = make([]float64, e.cols+1)
	all := append([]positioned(nil), e.positions...)
	// Collections.sort is stable.
	sort.SliceStable(all, func(i, j int) bool { return leftFirst(all[i].cell, all[j].cell) < 0 })
	for _, pos := range all {
		dim := pos.element.preferredDimension(stringBounder, 0, 0)
		e.ensureColWidth(pos.cell.minCol, pos.cell.maxCol+1, dim.Width()+2)
	}
	sort.SliceStable(all, func(i, j int) bool { return topFirst(all[i].cell, all[j].cell) < 0 })
	for _, pos := range all {
		supY := 0.0
		if pos.cell.minRow == 0 {
			supY = titleHeight / 2
		}
		dim := pos.element.preferredDimension(stringBounder, 0, 0)
		e.ensureRowHeight(pos.cell.minRow, pos.cell.maxRow+1, dim.Height()+supY+2)
	}
}

func (e *elementPyramid) ensureColWidth(first, last int, width float64) {
	actual := e.colsStart[last] - e.colsStart[first]
	if missing := width - actual; missing > 0 {
		for i := last; i < len(e.colsStart); i++ {
			e.colsStart[i] += missing
		}
	}
}

func (e *elementPyramid) ensureRowHeight(first, last int, height float64) {
	actual := e.rowsStart[last] - e.rowsStart[first]
	if missing := height - actual; missing > 0 {
		for i := last; i < len(e.rowsStart); i++ {
			e.rowsStart[i] += missing
		}
	}
}

func (e *elementPyramid) mayDrawBeyondPreferredDimension() bool {
	for _, pos := range e.positions {
		if pos.element.mayDrawBeyondPreferredDimension() {
			return true
		}
	}
	return false
}

// leftFirst is the LeftFirst comparator: the narrower cells first, then the
// leftmost.
func leftFirst(c0, c1 *cell) int {
	if diffNb := c0.nbCols() - c1.nbCols(); diffNb != 0 {
		return diffNb
	}
	return c0.minCol - c1.minCol
}

// topFirst is the TopFirst comparator: the shorter cells first, then the
// topmost.
func topFirst(c0, c1 *cell) int {
	if diffNb := c0.nbRows() - c1.nbRows(); diffNb != 0 {
		return diffNb
	}
	return c0.minRow - c1.minRow
}

// scrollStrategy is the scroll bars of "{S" (both), "{SI" (vertical only)
// and "{S-" (horizontal only).
type scrollStrategy int

const (
	scrollBoth scrollStrategy = iota
	scrollVerticalOnly
	scrollHorizontalOnly
)

func scrollStrategyFromDesc(s string) scrollStrategy {
	if strings.HasSuffix(s, "-") {
		return scrollHorizontalOnly
	}
	if strings.HasSuffix(s, "I") {
		return scrollVerticalOnly
	}
	return scrollBoth
}

// elementPyramidScrolled is a scroll area: an outlined group with its
// scroll bars. The bars are drawn in both passes, as upstream draws them.
type elementPyramidScrolled struct {
	*elementPyramid
	scrollStrategy scrollStrategy
}

const (
	scrollV1 = 15
	scrollV2 = 12
)

func newElementPyramidScrolled(p *positionner2, spriteContainer klimt.ISkinSimple,
	strategy scrollStrategy) *elementPyramidScrolled {
	return &elementPyramidScrolled{elementPyramid: newElementPyramid(p, drawOutside, nil, spriteContainer),
		scrollStrategy: strategy}
}

func (e *elementPyramidScrolled) preferredDimension(stringBounder font.StringBounder, x, y float64) geom.XDimension2D {
	result := e.elementPyramid.preferredDimension(stringBounder, x, y)
	switch e.scrollStrategy {
	case scrollHorizontalOnly:
		return result.DeltaWH(0, 30)
	case scrollVerticalOnly:
		return result.DeltaWH(30, 0)
	}
	return result.Delta(30)
}

func (e *elementPyramidScrolled) drawU(ug klimt.UGraphic, zIndex int, dimToUse *geom.XDimension2D) {
	e.elementPyramid.drawU(ug, zIndex, dimToUse)
	ug = ug.Apply(saltBlack())
	dim := e.elementPyramid.preferredDimension(ug.StringBounder(), 0, 0)
	if e.scrollStrategy == scrollBoth || e.scrollStrategy == scrollVerticalOnly {
		e.drawV(ug.Apply(klimt.UTranslateDx(dim.Width()+4)), scrollV1, dim.Height())
	}
	if e.scrollStrategy == scrollBoth || e.scrollStrategy == scrollHorizontalOnly {
		e.drawH(ug.Apply(klimt.UTranslateDy(dim.Height()+4)), dim.Width(), scrollV1)
	}
}

func scrollTriangle(points ...float64) *klimt.UPath {
	poly := klimt.UPathNone()
	poly.MoveTo(points[0], points[1])
	for i := 2; i < len(points); i += 2 {
		poly.LineTo(points[i], points[i+1])
	}
	poly.ClosePath()
	return poly
}

func tr0() *klimt.UPath   { return scrollTriangle(3, 0, 6, 5, 0, 5, 3, 0) }
func tr180() *klimt.UPath { return scrollTriangle(3, 5, 6, 0, 0, 0, 3, 5) }
func tr90() *klimt.UPath  { return scrollTriangle(0, 3, 5, 6, 5, 0, 0, 3) }
func tr270() *klimt.UPath { return scrollTriangle(5, 3, 0, 6, 0, 0, 5, 3) }

func (e *elementPyramidScrolled) drawV(ug klimt.UGraphic, width, height float64) {
	ug.Draw(klimt.URectangleBuild(width, height))
	ug.Apply(klimt.UTranslateDy(scrollV2)).Draw(klimt.ULineH(width))
	ug.Apply(klimt.UTranslateDy(height - scrollV2)).Draw(klimt.ULineH(width))
	ug.Apply(klimt.NewUTranslate(4, 4)).Apply(klimt.Bg(saltBlack())).Draw(tr0())
	ug.Apply(klimt.NewUTranslate(4, height-scrollV2+4)).Apply(klimt.Bg(saltBlack())).Draw(tr180())
}

func (e *elementPyramidScrolled) drawH(ug klimt.UGraphic, width, height float64) {
	ug.Draw(klimt.URectangleBuild(width, height))
	ug.Apply(klimt.UTranslateDx(scrollV2)).Draw(klimt.ULineV(height))
	ug.Apply(klimt.UTranslateDx(width - scrollV2)).Draw(klimt.ULineV(height))
	ug.Apply(klimt.NewUTranslate(4, 4)).Apply(klimt.Bg(saltBlack())).Draw(tr90())
	ug.Apply(klimt.NewUTranslate(width-scrollV2+4, 4)).Apply(klimt.Bg(saltBlack())).Draw(tr270())
}
