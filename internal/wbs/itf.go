// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/wbs/ITF.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/wbs/ITFComposed.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/wbs/ITFLeaf.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package wbs

import (
	"github.com/ekalinin/gopuml/internal/activitydiagram3"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/mindmap"
)

// itf is the drawing of a node below the root and of its children: t1 and
// t2 are the middles of the top and the bottom of its box, f1 and f2 the
// middles of its left and right sides.
type itf interface {
	klimt.TextBlock
	t1(stringBounder font.StringBounder) geom.XPoint2D
	t2(stringBounder font.StringBounder) geom.XPoint2D
	f1(stringBounder font.StringBounder) geom.XPoint2D
	f2(stringBounder font.StringBounder) geom.XPoint2D
	// mainBoxWidth is the width of the box of the node itself.
	mainBoxWidth(stringBounder font.StringBounder) float64
	// setForcedMinWidth widens the box of the node to at least width.
	setForcedMinWidth(width float64)
}

// itfLeaf is the drawing of a node without children.
type itfLeaf struct {
	klimt.TextBlockDefaults
	box            klimt.TextBlock
	fbox           *activitydiagram3.FtileBoxOld
	idea           *wElement
	forcedMinWidth float64
}

func newITFLeaf(idea *wElement, skinParam klimt.ISkinParam) *itfLeaf {
	shape := idea.shape
	style := idea.style()
	label := idea.label
	l := &itfLeaf{idea: idea}
	switch shape {
	case mindmap.IdeaShapePseudo:
		l.box = klimt.TextBlockUtilsEmptyTextBlock()
	case mindmap.IdeaShapeBox:
		l.fbox = activitydiagram3.FtileBoxOldCreateWbs(style, skinParam, label)
		l.box = l.fbox
	default:
		text := label.Create0(style.FontConfiguration(skinParam.IHtmlColorSet(), nil), style.HorizontalAlignment(),
			skinParam, style.WrapWidth(), klimt.CreoleModeFull, nil, nil)
		l.box = klimt.TextBlockUtilsWithMargin4(text, 0, 3, 1, 1)
	}
	return l
}

func (l *itfLeaf) setForcedMinWidth(width float64) {
	l.forcedMinWidth = width
	if l.fbox != nil {
		l.fbox.SetMinimumWidth(width)
	}
}

func (l *itfLeaf) mainBoxWidth(stringBounder font.StringBounder) float64 {
	return l.CalculateDimension(stringBounder).Width()
}

// CalculateDimension implements klimt.TextBlock: the box, widened to the
// forced width for a node without a box.
func (l *itfLeaf) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dim := l.box.CalculateDimension(stringBounder)
	if l.fbox == nil && l.forcedMinWidth > dim.Width() {
		return geom.NewXDimension2D(l.forcedMinWidth, dim.Height())
	}
	return dim
}

// DrawU implements klimt.TextBlock: the box, with a text without a box
// centered in the forced width. It records the geometry of the node.
func (l *itfLeaf) DrawU(ug klimt.UGraphic) {
	if translate, ok := translateOf(ug); ok {
		l.idea.setGeometry(translate, l.CalculateDimension(ug.StringBounder()))
	}
	if l.fbox == nil && l.forcedMinWidth > 0 {
		boxDim := l.box.CalculateDimension(ug.StringBounder())
		if l.forcedMinWidth > boxDim.Width() {
			dx := (l.forcedMinWidth - boxDim.Width()) / 2
			l.box.DrawU(ug.Apply(klimt.UTranslateDx(dx)))
			return
		}
	}
	l.box.DrawU(ug)
}

func (l *itfLeaf) t1(stringBounder font.StringBounder) geom.XPoint2D {
	dim := l.CalculateDimension(stringBounder)
	return geom.NewXPoint2D(dim.Width()/2, 0)
}

func (l *itfLeaf) t2(stringBounder font.StringBounder) geom.XPoint2D {
	dim := l.CalculateDimension(stringBounder)
	return geom.NewXPoint2D(dim.Width()/2, dim.Height())
}

func (l *itfLeaf) f1(stringBounder font.StringBounder) geom.XPoint2D {
	dim := l.CalculateDimension(stringBounder)
	return geom.NewXPoint2D(0, dim.Height()/2)
}

func (l *itfLeaf) f2(stringBounder font.StringBounder) geom.XPoint2D {
	dim := l.CalculateDimension(stringBounder)
	return geom.NewXPoint2D(dim.Width(), dim.Height()/2)
}

// itfComposed is the drawing of a node with children: its box, and below
// it the left children to the left of a vertical line from its middle and
// the right children to its right, one below the other.
type itfComposed struct {
	wbsTextBlock
	left             []itf
	right            []itf
	main             klimt.TextBlock
	marginBottom     float64
	autoWidthApplied bool
}

// itfComposedDelta1x is the space between the vertical line and the
// children.
const itfComposedDelta1x = 10

func newITFComposed(skinParam klimt.ISkinParam, idea *wElement, left, right []itf) *itfComposed {
	c := &itfComposed{wbsTextBlock: newWBSTextBlock(skinParam, idea), left: left, right: right}
	c.main = c.buildMain(idea)
	style := idea.style()
	if idea.shape != mindmap.IdeaShapePseudo {
		c.marginBottom = style.Margin().Bottom()
	}
	return c
}

func (c *itfComposed) mainBoxWidth(stringBounder font.StringBounder) float64 {
	return c.main.CalculateDimension(stringBounder).Width()
}

func (c *itfComposed) setForcedMinWidth(width float64) {
	if box, ok := c.main.(*activitydiagram3.FtileBoxOld); ok {
		box.SetMinimumWidth(width)
	}
}

// applyAutoWidth gives the boxes of the children the width of the widest
// one, with "Width auto". The dimension and the drawing call it, whichever
// comes first; it runs once.
func (c *itfComposed) applyAutoWidth(stringBounder font.StringBounder) {
	if c.autoWidthApplied {
		return
	}
	c.autoWidthApplied = true
	if !c.isAutoWidth() {
		return
	}
	all := append(append([]itf(nil), c.left...), c.right...)
	maxWidth := 0.0
	for _, child := range all {
		maxWidth = max(maxWidth, child.mainBoxWidth(stringBounder))
	}
	for _, child := range all {
		child.setForcedMinWidth(maxWidth)
	}
	c.memo.InvalidateDimensionCache()
}

// build2 returns the drawing of a node below the root and of its children.
func build2(skinParam klimt.ISkinParam, idea *wElement) itf {
	if idea.isLeaf() {
		return newITFLeaf(idea, idea.withBackColor(skinParam))
	}
	var left, right []itf
	for _, child := range idea.children(klimt.DirectionLeft) {
		left = append(left, build2(skinParam, child))
	}
	for _, child := range idea.children(klimt.DirectionRight) {
		right = append(right, build2(skinParam, child))
	}
	return newITFComposed(skinParam, idea, left, right)
}

// w1 is the x of the vertical line: the middle of the box, or further right
// for the left children.
func (c *itfComposed) w1(stringBounder font.StringBounder) float64 {
	mainWidth := c.main.CalculateDimension(stringBounder).Width()
	return max(mainWidth/2, itfComposedDelta1x+collWidth(stringBounder, c.left))
}

func (c *itfComposed) t1(stringBounder font.StringBounder) geom.XPoint2D {
	return geom.NewXPoint2D(c.w1(stringBounder), 0)
}

func (c *itfComposed) t2(stringBounder font.StringBounder) geom.XPoint2D {
	mainDim := c.main.CalculateDimension(stringBounder)
	return geom.NewXPoint2D(c.w1(stringBounder), mainDim.Height())
}

func (c *itfComposed) f1(stringBounder font.StringBounder) geom.XPoint2D {
	mainDim := c.main.CalculateDimension(stringBounder)
	return geom.NewXPoint2D(c.w1(stringBounder)-mainDim.Width()/2, mainDim.Height()/2)
}

func (c *itfComposed) f2(stringBounder font.StringBounder) geom.XPoint2D {
	mainDim := c.main.CalculateDimension(stringBounder)
	return geom.NewXPoint2D(c.w1(stringBounder)+mainDim.Width()/2, mainDim.Height()/2)
}

// CalculateDimension implements klimt.TextBlock: upstream's
// calculateDimensionSlow, memoized.
func (c *itfComposed) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return c.memo.Dimension(stringBounder, c.calculateDimensionSlow)
}

func (c *itfComposed) calculateDimensionSlow(stringBounder font.StringBounder) geom.XDimension2D {
	c.applyAutoWidth(stringBounder)
	mainDim := c.main.CalculateDimension(stringBounder)
	mainWidth := mainDim.Width()
	height := mainDim.Height() + max(collHeight(stringBounder, c.left, c.marginBottom),
		collHeight(stringBounder, c.right, c.marginBottom))
	width := max(mainWidth/2, itfComposedDelta1x+collWidth(stringBounder, c.left)) +
		max(mainWidth/2, itfComposedDelta1x+collWidth(stringBounder, c.right))
	return geom.NewXDimension2D(width, height)
}

// DrawU implements klimt.TextBlock: the box of the node, the children and
// the lines to them. It records the geometry of the node.
func (c *itfComposed) DrawU(ug klimt.UGraphic) {
	c.applyAutoWidth(ug.StringBounder())
	stringBounder := ug.StringBounder()
	mainDim := c.main.CalculateDimension(stringBounder)
	if translate, ok := translateOf(ug); ok {
		c.idea.setGeometry(translate, mainDim)
	}
	wx := c.w1(stringBounder) - mainDim.Width()/2
	c.main.DrawU(ug.Apply(klimt.UTranslateDx(wx)))
	x := c.w1(stringBounder)
	y := mainDim.Height()
	lastY1 := y
	lineColor := c.linkColor()
	for _, child := range c.left {
		y += c.marginBottom
		childDim := child.CalculateDimension(stringBounder)
		lastY1 = y + child.f2(stringBounder).Y
		c.drawLine(ug.Apply(lineColor), x-childDim.Width()-itfComposedDelta1x+child.f2(stringBounder).X, lastY1,
			x, lastY1)
		child.DrawU(ug.Apply(klimt.NewUTranslate(x-childDim.Width()-itfComposedDelta1x, y)))
		y += childDim.Height()
	}
	y = mainDim.Height()
	lastY2 := y
	for _, child := range c.right {
		y += c.marginBottom
		childDim := child.CalculateDimension(stringBounder)
		lastY2 = y + child.f1(stringBounder).Y
		c.drawLine(ug.Apply(lineColor), x, lastY2, x+itfComposedDelta1x+child.f1(stringBounder).X, lastY2)
		child.DrawU(ug.Apply(klimt.NewUTranslate(x+itfComposedDelta1x, y)))
		y += childDim.Height()
	}
	c.drawLine(ug.Apply(lineColor), x, mainDim.Height(), x, max(lastY1, lastY2))
}

// linkColor is the line color of the style of the node with its
// stereotype.
func (c *itfComposed) linkColor() color.HColor {
	query := c.idea.style().Signature().AddStereotypeOf(c.idea.stereotype)
	styleArrow := query.MergedStyle(c.idea.styleBuilder)
	return styleArrow.Value(klimt.PNameLineColor).AsColor(c.skinParam.IHtmlColorSet())
}

func collWidth(stringBounder font.StringBounder, all []itf) float64 {
	result := 0.0
	for _, child := range all {
		result = max(result, child.CalculateDimension(stringBounder).Width())
	}
	return result
}

func collHeight(stringBounder font.StringBounder, all []itf, deltay float64) float64 {
	result := 0.0
	for _, child := range all {
		result += deltay + child.CalculateDimension(stringBounder).Height()
	}
	return result
}
