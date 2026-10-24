// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/shape/TextBlockBordered.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/shape/TextBlockLineBefore.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// TextBlockBordered draws a block in a box with padding.
type TextBlockBordered struct {
	TextBlockDefaults
	cornersize               float64
	backgroundColor          color.HColor
	borderColor              color.HColor
	top, right, bottom, left float64
	stroke                   *UStroke
	withShadow               bool
	id                       string
	textBlock                TextBlock
}

func newTextBlockBordered(textBlock TextBlock, stroke *UStroke, borderColor, backgroundColor color.HColor,
	cornersize float64, margins ClockwiseTopRightBottomLeft, id string) *TextBlockBordered {
	return &TextBlockBordered{top: margins.Top(), right: margins.Right(), bottom: margins.Bottom(),
		left: margins.Left(), cornersize: cornersize, textBlock: textBlock, stroke: stroke,
		borderColor: borderColor, backgroundColor: backgroundColor, id: id}
}

func (t *TextBlockBordered) textHeight(stringBounder font.StringBounder) float64 {
	size := t.textBlock.CalculateDimension(stringBounder)
	return size.Height() + t.top + t.bottom
}

func (t *TextBlockBordered) pureTextWidth(stringBounder font.StringBounder) float64 {
	return t.textBlock.CalculateDimension(stringBounder).Width()
}

func (t *TextBlockBordered) textWidth(stringBounder font.StringBounder) float64 {
	return t.pureTextWidth(stringBounder) + t.left + t.right
}

// CalculateDimension returns the size of the box plus one.
func (t *TextBlockBordered) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	height := t.textHeight(stringBounder)
	width := t.textWidth(stringBounder)
	return geom.NewXDimension2D(width+1, height+1)
}

func (t *TextBlockBordered) applyStroke(ug UGraphic) UGraphic {
	if t.stroke == nil {
		return ug
	}
	return ug.Apply(*t.stroke)
}

func (t *TextBlockBordered) noBorder() bool {
	if t.stroke == nil {
		return false
	}
	return t.stroke.Thickness() == 0
}

// DrawU draws the box, unless it is invisible, then the block.
func (t *TextBlockBordered) DrawU(ug UGraphic) {
	stringBounder := ug.StringBounder()
	polygon := t.polygonNormal(stringBounder)
	ugOriginal := ug
	if t.withShadow {
		polygon.SetDeltaShadow(4)
	}
	var back color.HColor
	if t.backgroundColor == nil || t.backgroundColor.IsTransparent() ||
		t.backgroundColor.Equals(ug.DefaultBackground()) {
		back = color.None()
	} else {
		back = t.backgroundColor
	}
	c := t.borderColor
	if t.noBorder() {
		c = back
	}
	if c == nil {
		c = color.None()
	}
	if !back.IsTransparent() || !c.IsTransparent() {
		ug = ug.Apply(Bg(back))
		ug = ug.Apply(c)
		ug = t.applyStroke(ug)
		ug.Draw(polygon)
	}
	toDraw := t.textBlock
	if sb, ok := t.textBlock.(*SheetBlock2); ok {
		toDraw = sb.EnlargeMe(t.left, t.right)
	}
	toDraw.DrawU(ugOriginal.Apply(c).Apply(NewUTranslate(t.left, t.top)))
}

func (t *TextBlockBordered) polygonNormal(stringBounder font.StringBounder) Shadowable {
	height := t.textHeight(stringBounder)
	width := t.textWidth(stringBounder)
	return URectangleBuild(width, height).Rounded(t.cornersize).WithCommentAndCodeLine(t.id, "")
}

// TextBlockLineBefore draws a horizontal separator line, possibly with a
// title, before a block.
type TextBlockLineBefore struct {
	TextBlockDefaults
	textBlock        TextBlock
	separator        rune
	title            TextBlock
	defaultThickness float64
}

var _ WithPorts = (*TextBlockLineBefore)(nil)

// NewTextBlockLineBefore returns the block after a separator; separator is
// the line style ('=', '.', '-', '_', or 0 for the default) and title may be
// nil.
func NewTextBlockLineBefore(defaultThickness float64, textBlock TextBlock, separator rune,
	title TextBlock) *TextBlockLineBefore {
	return &TextBlockLineBefore{textBlock: textBlock, separator: separator, title: title,
		defaultThickness: defaultThickness}
}

// CalculateDimension returns the size of the block, wide and tall enough for
// the title.
func (t *TextBlockLineBefore) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	d := t.textBlock.CalculateDimension(stringBounder)
	if t.title != nil {
		dimTitle := t.title.CalculateDimension(stringBounder)
		return d.AtLeast(dimTitle.Width()+8, dimTitle.Height())
	}
	return d
}

// DrawU draws the separator and the block.
func (t *TextBlockLineBefore) DrawU(ug UGraphic) {
	c := ug.Param().Color()
	if t.title == nil {
		UHorizontalLineInfinite(t.defaultThickness, 1, 1, t.separator).DrawMe(ug)
	}
	t.textBlock.DrawU(ug)
	if c == nil {
		ug = ug.Apply(color.None())
	} else {
		ug = ug.Apply(c)
	}
	if t.title != nil {
		UHorizontalLineInfiniteTitle(t.defaultThickness, 1, 1, t.title, t.separator).DrawMe(ug)
	}
}

// InnerPosition returns the position of member in the block.
func (t *TextBlockLineBefore) InnerPosition(member string, stringBounder font.StringBounder) *geom.XRectangle2D {
	return t.textBlock.InnerPosition(member, stringBounder)
}

// Ports returns the ports of the block, if it has any.
func (t *TextBlockLineBefore) Ports(stringBounder font.StringBounder) *Ports {
	if wp, ok := t.textBlock.(WithPorts); ok {
		return wp.Ports(stringBounder)
	}
	return NewPorts()
}
