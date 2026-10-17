// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbolRectangle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package symbol

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// usymbolRectangle is the rectangle symbol, with rounded or diagonal
// corners (rectangle, agent, archimate and the rectangle component).
type usymbolRectangle struct {
	usymbolBase
	sname klimt.SName
}

// SNames returns the style name of the symbol.
func (s *usymbolRectangle) SNames() []klimt.SName { return []klimt.SName{s.sname} }

func drawRectangleSymbol(ug klimt.UGraphic, width, height, shadowing, roundCorner, diagonalCorner float64) {
	rect := klimt.URectangleBuild(width, height)
	var shape klimt.Shadowable
	if diagonalCorner > 0 {
		shape = rect.DiagonalCorner(diagonalCorner)
	} else {
		shape = rect.Rounded(roundCorner)
	}
	shape.SetDeltaShadow(shadowing)
	ug.Draw(shape)
}

func rectangleMargin() margin { return newMargin(10, 10, 10, 10) }

// rectangleSmall is the anonymous text block of upstream's asSmall.
type rectangleSmall struct {
	klimt.TextBlockDefaults
	label, stereotype klimt.TextBlock
	symbolContext     *klimt.Fashion
	stereoAlignment   geom.HorizontalAlignment
}

func (b *rectangleSmall) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = klimt.UGraphicStencilCreateDim(ug, dim)
	ug = b.symbolContext.Apply(ug)
	drawRectangleSymbol(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow(), b.symbolContext.RoundCorner(),
		b.symbolContext.DiagonalCorner())
	m := rectangleMargin()
	tb := klimt.TextBlockUtilsMergeTB(b.stereotype, b.label, b.stereoAlignment)
	tb.DrawU(ug.Apply(klimt.NewUTranslate(m.x1, m.y1)))
}

func (b *rectangleSmall) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimLabel := b.label.CalculateDimension(stringBounder)
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	return rectangleMargin().addDimension(dimStereo.MergeTB(dimLabel))
}

// AsSmall returns the rectangle around the stereotype and the label.
func (s *usymbolRectangle) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	stereoAlignment geom.HorizontalAlignment) klimt.TextBlock {
	return &rectangleSmall{label: label, stereotype: stereotype, symbolContext: symbolContext,
		stereoAlignment: stereoAlignment}
}

// rectangleBig is the anonymous text block of upstream's asBig.
type rectangleBig struct {
	klimt.TextBlockDefaults
	title, stereotype klimt.TextBlock
	labelAlignment    geom.HorizontalAlignment
	width, height     float64
	symbolContext     *klimt.Fashion
	stereoAlignment   geom.HorizontalAlignment
}

func (b *rectangleBig) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = b.symbolContext.Apply(ug)
	drawRectangleSymbol(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow(), b.symbolContext.RoundCorner(),
		b.symbolContext.DiagonalCorner())
	drawBigTitle(ug, b.title, b.stereotype, b.labelAlignment, b.stereoAlignment, b.width, rectangleMargin())
}

// drawBigTitle draws the stereotype and the title at the top of a big
// symbol, aligned as rectangle, hexagon and label do.
func drawBigTitle(ug klimt.UGraphic, title, stereotype klimt.TextBlock, labelAlignment,
	stereoAlignment geom.HorizontalAlignment, width float64, m margin) {
	dimStereo := stereotype.CalculateDimension(ug.StringBounder())
	var posStereoX, posStereoY float64
	if stereoAlignment == geom.HorizontalAlignmentRight {
		posStereoX = width - dimStereo.Width() - m.x1/2
		posStereoY = m.y1 / 2
	} else {
		posStereoX = (width - dimStereo.Width()) / 2
		posStereoY = 2
	}
	stereotype.DrawU(ug.Apply(klimt.NewUTranslate(posStereoX, posStereoY)))
	dimTitle := title.CalculateDimension(ug.StringBounder())
	var posTitle float64
	switch labelAlignment {
	case geom.HorizontalAlignmentLeft:
		posTitle = 3
	case geom.HorizontalAlignmentRight:
		posTitle = width - dimTitle.Width() - 3
	default:
		posTitle = (width - dimTitle.Width()) / 2
	}
	title.DrawU(ug.Apply(klimt.NewUTranslate(posTitle, 2+dimStereo.Height())))
}

func (b *rectangleBig) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(b.width, b.height)
}

// AsBig returns the rectangle of a size with the stereotype and the title
// at its top.
func (s *usymbolRectangle) AsBig(title klimt.TextBlock, labelAlignment geom.HorizontalAlignment,
	stereotype klimt.TextBlock, width, height float64, symbolContext *klimt.Fashion,
	stereoAlignment geom.HorizontalAlignment) klimt.TextBlock {
	return &rectangleBig{title: title, labelAlignment: labelAlignment, stereotype: stereotype, width: width,
		height: height, symbolContext: symbolContext, stereoAlignment: stereoAlignment}
}
