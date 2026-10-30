// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/Grid.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/Grid2.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/Segment.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package salt

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// segment is a line of a group between two rows or two columns, at a row
// and a column.
type segment struct {
	row, col int
}

func (s segment) hashCode() int32 { return int32(s.row*47 + s.col) }

// segmentSet is a HashSet of segments: the segments in insertion order,
// without duplicates, drawn in the iteration order of upstream's HashSet.
type segmentSet struct {
	all []segment
}

func (s *segmentSet) add(seg segment) {
	for _, e := range s.all {
		if e == seg {
			return
		}
	}
	s.all = append(s.all, seg)
}

func (s *segmentSet) iterationOrder() []segment {
	hashes := make([]int32, len(s.all))
	for i, seg := range s.all {
		hashes[i] = seg.hashCode()
	}
	result := make([]segment, 0, len(s.all))
	for _, i := range klimt.JavaHashOrder(hashes) {
		result = append(result, s.all[i])
	}
	return result
}

// grid draws the lines of a group and its title.
type grid struct {
	rowsStart, colsStart []float64
	strategy             tableStrategy
	title                klimt.TextBlock
	horizontals          segmentSet
	verticals            segmentSet
}

func newGrid(rowsStart, colsStart []float64, strategy tableStrategy, title klimt.TextBlock) *grid {
	g := &grid{rowsStart: rowsStart, colsStart: colsStart, strategy: strategy, title: title}
	if strategy == drawOutside || strategy == drawOutsideWithTitle || strategy == drawAll {
		g.addOutside()
	}
	return g
}

func (g *grid) addOutside() {
	nbRow := len(g.rowsStart)
	nbCol := len(g.colsStart)
	for c := 0; c < nbCol-1; c++ {
		g.horizontals.add(segment{0, c})
		g.horizontals.add(segment{nbRow - 1, c})
	}
	for r := 0; r < nbRow-1; r++ {
		g.verticals.add(segment{r, 0})
		g.verticals.add(segment{r, nbCol - 1})
	}
}

func (g *grid) drawU(ug klimt.UGraphic, x, y float64, white color.HColor) {
	for _, seg := range g.horizontals.iterationOrder() {
		width := g.colsStart[seg.col+1] - g.colsStart[seg.col]
		ug.Apply(klimt.NewUTranslate(x+g.colsStart[seg.col], y+g.rowsStart[seg.row])).Draw(klimt.ULineH(width))
	}
	for _, seg := range g.verticals.iterationOrder() {
		height := g.rowsStart[seg.row+1] - g.rowsStart[seg.row]
		ug.Apply(klimt.NewUTranslate(x+g.colsStart[seg.col], y+g.rowsStart[seg.row])).Draw(klimt.ULineV(height))
	}
	dim := g.title.CalculateDimension(ug.StringBounder())
	if dim.Width() > 0 && dim.Height() > 0 {
		ug2 := ug.Apply(klimt.NewUTranslate(x+6, y-dim.Height()*0))
		ug2.Apply(klimt.Bg(white)).Apply(white).Draw(klimt.URectangleBuildDim(dim))
		g.title.DrawU(ug2)
	}
}

func (g *grid) addCell(c *cell) {
	if g.strategy == drawNone || g.strategy == drawOutside || g.strategy == drawOutsideWithTitle {
		return
	}
	if g.strategy == drawHorizontal || g.strategy == drawAll {
		for col := c.minCol; col <= c.maxCol; col++ {
			g.horizontals.add(segment{c.minRow, col})
			g.horizontals.add(segment{c.maxRow + 1, col})
		}
	}
	if g.strategy == drawVertical || g.strategy == drawAll {
		for r := c.minRow; r <= c.maxRow; r++ {
			g.verticals.add(segment{r, c.minCol})
			g.verticals.add(segment{r, c.maxCol + 1})
		}
	}
}

// grid2 draws the lines of a tree table.
type grid2 struct {
	rowsStart, colsStart []float64
	strategy             tableStrategy
}

func (g *grid2) drawU(ug klimt.UGraphic) {
	xmin := g.colsStart[0]
	xmax := g.colsStart[len(g.colsStart)-1]
	ymin := g.rowsStart[0]
	ymax := g.rowsStart[len(g.rowsStart)-1]
	if g.strategy == drawOutside || g.strategy == drawOutsideWithTitle {
		ug.Apply(klimt.NewUTranslate(xmin, ymin)).Draw(klimt.ULineH(xmax - xmin))
		ug.Apply(klimt.NewUTranslate(xmin, ymax)).Draw(klimt.ULineH(xmax - xmin))
		ug.Apply(klimt.NewUTranslate(xmin, ymin)).Draw(klimt.ULineV(ymax - ymin))
		ug.Apply(klimt.NewUTranslate(xmax, ymin)).Draw(klimt.ULineV(ymax - ymin))
	}
	if g.strategy == drawHorizontal || g.strategy == drawAll {
		for _, y := range g.rowsStart {
			ug.Apply(klimt.NewUTranslate(xmin, y)).Draw(klimt.ULineH(xmax - xmin))
		}
	}
	if g.strategy == drawVertical || g.strategy == drawAll {
		for _, x := range g.colsStart {
			ug.Apply(klimt.NewUTranslate(x, ymin)).Draw(klimt.ULineV(ymax - ymin))
		}
	}
}
