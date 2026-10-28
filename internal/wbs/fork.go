// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/wbs/Fork.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package wbs

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// fork is the drawing of the root: its box, and below it its children side
// by side, joined to it by a horizontal line.
type fork struct {
	wbsTextBlock
	main             klimt.TextBlock
	right            []itf
	autoWidthApplied bool
}

// The space between the children of the root, and between the root and its
// children.
const (
	forkDelta1x = 20
	forkDeltay  = 40
)

// newFork returns the drawing of the root; a nil root throws
// NullPointerException and a node below the root IllegalArgumentException,
// as upstream.
func newFork(skinParam klimt.ISkinParam, idea *wElement) *fork {
	if idea == nil {
		uml.ThrowNPE(`Cannot invoke "net.sourceforge.plantuml.wbs.WElement.withBackColor(net.sourceforge.plantuml.style.ISkinParam)" because "idea" is null`)
	}
	f := &fork{wbsTextBlock: newWBSTextBlock(idea.withBackColor(skinParam), idea)}
	if idea.level != 0 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	f.main = f.buildMain(idea)
	for _, child := range idea.children(klimt.DirectionRight) {
		f.right = append(f.right, build2(skinParam, child))
	}
	return f
}

// applyAutoWidth gives the boxes of the children of the root the width of
// the widest one, with "Width auto". The dimension and the drawing call it,
// whichever comes first; it runs once.
func (f *fork) applyAutoWidth(stringBounder font.StringBounder) {
	if f.autoWidthApplied {
		return
	}
	f.autoWidthApplied = true
	if !f.isAutoWidth() {
		return
	}
	maxWidth := 0.0
	for _, child := range f.right {
		maxWidth = max(maxWidth, child.mainBoxWidth(stringBounder))
	}
	for _, child := range f.right {
		child.setForcedMinWidth(maxWidth)
	}
	f.memo.InvalidateDimensionCache()
}

// DrawU implements klimt.TextBlock: the children side by side, the root
// centered over the first and the last, and the lines between them.
func (f *fork) DrawU(ug klimt.UGraphic) {
	f.applyAutoWidth(ug.StringBounder())
	stringBounder := ug.StringBounder()
	mainDim := f.main.CalculateDimension(stringBounder)
	y0 := mainDim.Height()
	y1 := y0 + forkDeltay/2
	y2 := y0 + forkDeltay
	mainWidth := mainDim.Width()
	if len(f.right) == 0 {
		f.main.DrawU(ug)
		f.drawLine(ug, mainWidth/2, y0, mainWidth/2, y1)
		return
	}
	x := 0.0
	firstX := f.right[0].t1(stringBounder).X
	lastX := firstX
	for _, child := range f.right {
		lastX = x + child.t1(stringBounder).X
		f.drawLine(ug, lastX, y1, lastX, y2)
		child.DrawU(ug.Apply(klimt.NewUTranslate(x, y2)))
		x += child.CalculateDimension(stringBounder).Width() + forkDelta1x
	}
	var posMain float64
	if lastX > firstX {
		f.drawLine(ug, firstX, y1, lastX, y1)
		posMain = firstX + (lastX-firstX-mainWidth)/2
	} else {
		fullDim := f.CalculateDimension(stringBounder)
		posMain = (fullDim.Width() - mainWidth) / 2
		f.drawLine(ug, firstX, y1, posMain+mainWidth/2, y1)
	}
	f.main.DrawU(ug.Apply(klimt.UTranslateDx(posMain)))
	f.drawLine(ug, posMain+mainWidth/2, y0, posMain+mainWidth/2, y1)
}

// CalculateDimension implements klimt.TextBlock: upstream's
// calculateDimensionSlow, memoized.
func (f *fork) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return f.memo.Dimension(stringBounder, f.calculateDimensionSlow)
}

func (f *fork) calculateDimensionSlow(stringBounder font.StringBounder) geom.XDimension2D {
	f.applyAutoWidth(stringBounder)
	width := 0.0
	height := 0.0
	for _, child := range f.right {
		childDim := child.CalculateDimension(stringBounder)
		height = max(height, childDim.Height())
		width += childDim.Width()
	}
	if len(f.right) > 1 {
		width += float64(len(f.right)-1) * forkDelta1x
	}
	mainDim := f.main.CalculateDimension(stringBounder)
	height += mainDim.Height()
	height += forkDeltay
	width = max(width, mainDim.Width())
	return geom.NewXDimension2D(width, height)
}
