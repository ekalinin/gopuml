// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/ElementTree.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/ElementTreeEntry.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/Skeleton.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/element/ListWidth.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package salt

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/text"
)

// elementTree is a tree or a tree table: each entry at its level, linked to
// its parent by gray lines, with the widgets of its other cells in columns.
// As upstream, it draws in the pass 0 only, and gives null as the size of
// the cells.
type elementTree struct {
	abstractElement
	entries         []*elementTreeEntry
	font            font.UFont
	spriteContainer klimt.ISkinSimple
	strategy        tableStrategy
}

const treeMargin = 10

func newElementTree(f font.UFont, spriteContainer klimt.ISkinSimple, strategy tableStrategy) *elementTree {
	return &elementTree{font: f, spriteContainer: spriteContainer, strategy: strategy}
}

func (e *elementTree) addEntry(s string) {
	level := 0
	for strings.HasPrefix(s, "+") {
		level++
		s = s[1:]
	}
	elmt := newElementText([]string{text.Trin(s)}, e.font, e.spriteContainer)
	e.entries = append(e.entries, &elementTreeEntry{level: level, firstElement: elmt})
}

func (e *elementTree) addCellToEntry(elmt element) {
	if size := len(e.entries); size > 0 {
		e.entries[size-1].addCell(elmt)
	}
}

func (e *elementTree) preferredDimension(stringBounder font.StringBounder, _, _ float64) geom.XDimension2D {
	w1 := 0.0
	h := 0.0
	for _, entry := range e.entries {
		dim1 := entry.preferredDimensionFirstCell(stringBounder)
		w1 = max(w1, dim1.Width())
		h += entry.rowHeight(stringBounder)
	}
	w2 := e.widthOther(stringBounder).totalWidthWithMargin(treeMargin)
	if w2 > 0 {
		w2 += treeMargin
	}
	return geom.NewXDimension2D(w1+w2+2, h)
}

func (e *elementTree) widthOther(stringBounder font.StringBounder) *listWidth {
	merge := &listWidth{}
	for _, entry := range e.entries {
		merge = merge.mergeMax(entry.preferredDimensionOtherCell(stringBounder))
	}
	return merge
}

func (e *elementTree) width1(stringBounder font.StringBounder) float64 {
	w1 := 0.0
	for _, entry := range e.entries {
		w1 = max(w1, entry.preferredDimensionFirstCell(stringBounder).Width())
	}
	return w1
}

func (e *elementTree) drawU(ug klimt.UGraphic, zIndex int, _ *geom.XDimension2D) {
	if zIndex != 0 {
		return
	}
	ug = ug.Apply(saltBlack())
	stringBounder := ug.StringBounder()
	w1 := e.width1(stringBounder)
	otherWidth := e.widthOther(stringBounder)
	skeleton := &skeleton{}
	yvar := 0.0
	rows := []float64{yvar}
	xvar := 0.0
	cols := []float64{xvar}
	xvar += w1 + treeMargin/2
	cols = append(cols, xvar)
	for _, w := range otherWidth.allWidth {
		xvar += w + treeMargin
		cols = append(cols, xvar)
	}
	for _, entry := range e.entries {
		entry.drawFirstCell(ug, 0, yvar)
		entry.drawSecondCell(ug, w1+treeMargin, yvar, otherWidth, treeMargin)
		h := entry.rowHeight(stringBounder)
		skeleton.add(entry.xDelta()-7, yvar+h/2-1)
		yvar += h
		rows = append(rows, yvar)
	}
	ug = ug.Apply(saltColor88())
	skeleton.draw(ug, 0, 0)
	if e.strategy != drawNone {
		(&grid2{rowsStart: rows, colsStart: cols, strategy: e.strategy}).drawU(ug)
	}
}

// elementTreeEntry is an entry of a tree: its label and the widgets of its
// other cells.
type elementTreeEntry struct {
	firstElement  element
	level         int
	otherElements []element
}

func (t *elementTreeEntry) addCell(elmt element) { t.otherElements = append(t.otherElements, elmt) }

func (t *elementTreeEntry) preferredDimensionFirstCell(stringBounder font.StringBounder) geom.XDimension2D {
	return t.firstElement.preferredDimension(stringBounder, 0, 0).DeltaWH(t.xDelta(), 0)
}

func (t *elementTreeEntry) preferredDimensionOtherCell(stringBounder font.StringBounder) *listWidth {
	result := &listWidth{}
	for _, elmt := range t.otherElements {
		result.add(elmt.preferredDimension(stringBounder, 0, 0).Width())
	}
	return result
}

func (t *elementTreeEntry) rowHeight(stringBounder font.StringBounder) float64 {
	h := t.preferredDimensionFirstCell(stringBounder).Height()
	for _, elmt := range t.otherElements {
		h = max(h, elmt.preferredDimension(stringBounder, 0, 0).Height())
	}
	return h
}

func (t *elementTreeEntry) xDelta() float64 { return float64(t.level * 10) }

func (t *elementTreeEntry) drawFirstCell(ug klimt.UGraphic, x, y float64) {
	t.firstElement.drawU(ug.Apply(klimt.NewUTranslate(x+t.xDelta(), y)), 0, nil)
}

func (t *elementTreeEntry) drawSecondCell(ug klimt.UGraphic, x, y float64, otherWidth *listWidth, margin float64) {
	for i, elmt := range t.otherElements {
		elmt.drawU(ug.Apply(klimt.NewUTranslate(x, y)), 0, nil)
		x += otherWidth.allWidth[i] + margin
	}
}

// skeleton draws the lines of a tree: a small square before an entry
// followed by a deeper one, and lines from each entry to its parent.
type skeleton struct {
	entries []skeletonEntry
}

type skeletonEntry struct {
	xpos, ypos float64
}

func (e skeletonEntry) drawRectangle(ug klimt.UGraphic) {
	ug.Apply(klimt.NewUTranslate(e.xpos, e.ypos)).Draw(klimt.URectangleBuild(2, 2))
}

func (s *skeleton) add(x, y float64) { s.entries = append(s.entries, skeletonEntry{x, y}) }

func (s *skeleton) draw(ug klimt.UGraphic, _, _ float64) {
	for i, en := range s.entries {
		if i+1 < len(s.entries) && s.entries[i+1].xpos > en.xpos {
			en.drawRectangle(ug)
		}
		var parent *skeletonEntry
		for j := 0; j < i; j++ {
			if s.entries[j].xpos < en.xpos {
				parent = &s.entries[j]
			}
		}
		if parent != nil {
			drawSkeletonChild(ug, *parent, en)
		}
	}
}

func drawSkeletonChild(ug klimt.UGraphic, parent, child skeletonEntry) {
	dy := child.ypos - parent.ypos - 2
	ug.Apply(klimt.NewUTranslate(parent.xpos+1, parent.ypos+3)).Draw(klimt.ULineV(dy))
	dx := child.xpos - parent.xpos - 2
	ug.Apply(klimt.NewUTranslate(parent.xpos+1, child.ypos+1)).Draw(klimt.ULineH(dx))
}

// listWidth is the widths of the columns of a tree table.
type listWidth struct {
	allWidth []float64
}

func (l *listWidth) add(width float64) { l.allWidth = append(l.allWidth, width) }

func (l *listWidth) mergeMax(other *listWidth) *listWidth {
	result := &listWidth{}
	for i := 0; i < len(l.allWidth) || i < len(other.allWidth); i++ {
		result.add(max(l.widthSafe(i), other.widthSafe(i)))
	}
	return result
}

func (l *listWidth) widthSafe(i int) float64 {
	if i < len(l.allWidth) {
		return l.allWidth[i]
	}
	return 0
}

func (l *listWidth) totalWidthWithMargin(margin float64) float64 {
	result := 0.0
	for _, w := range l.allWidth {
		if result > 0 {
			result += margin
		}
		result += w
	}
	return result
}
