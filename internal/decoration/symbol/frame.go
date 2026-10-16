// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbolFrame.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package symbol

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// usymbolFrame is the frame symbol: a rectangle with a title in a corner
// tab (frame, group and partition).
type usymbolFrame struct {
	usymbolBase
	sname klimt.SName
}

// SNames returns the style name of the symbol.
func (s *usymbolFrame) SNames() []klimt.SName { return []klimt.SName{s.sname} }

func drawFrame(ug klimt.UGraphic, width, height float64, dimTitle geom.XDimension2D, shadowing,
	roundCorner float64) {
	rectangle := klimt.URectangleBuild(width, height).Rounded(roundCorner).IgnoreForCompressionOnX().
		IgnoreForCompressionOnY()
	rectangle.SetDeltaShadow(shadowing)

	ug.Draw(rectangle)

	var textWidth float64
	var cornersize float64
	if dimTitle.Width() == 0 {
		textWidth = width / 3
		cornersize = 7
	} else {
		textWidth = dimTitle.Width() + 10
		cornersize = 10
	}
	textHeight := frameYpos(dimTitle)

	line := klimt.UPathNone()
	line.SetIgnoreForCompressionOnX()
	line.MoveTo(textWidth, 0)

	line.LineTo(textWidth, textHeight-cornersize)
	line.LineTo(textWidth-cornersize, textHeight)

	line.LineTo(0, textHeight)
	ug.Apply(klimt.Bg(color.None())).Draw(line)
}

func frameYpos(dimTitle geom.XDimension2D) float64 {
	if dimTitle.Width() == 0 {
		return 12
	}
	return dimTitle.Height() + 3
}

func frameMargin() margin { return newMargin(10+5, 20+5, 15+5, 5+5) }

// frameSmall is the anonymous text block of upstream's asSmall.
type frameSmall struct {
	klimt.TextBlockDefaults
	label, stereotype klimt.TextBlock
	symbolContext     *klimt.Fashion
}

func (b *frameSmall) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = klimt.UGraphicStencilCreateDim(ug, dim)
	ug = b.symbolContext.Apply(ug)
	drawFrame(ug, dim.Width(), dim.Height(), geom.NewXDimension2D(0, 0), b.symbolContext.DeltaShadow(),
		b.symbolContext.RoundCorner())
	m := frameMargin()
	tb := klimt.TextBlockUtilsMergeTB(b.stereotype, b.label, geom.HorizontalAlignmentCenter)
	tb.DrawU(ug.Apply(klimt.NewUTranslate(m.x1, m.y1)))
}

func (b *frameSmall) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimLabel := b.label.CalculateDimension(stringBounder)
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	return frameMargin().addDimension(dimStereo.MergeTB(dimLabel))
}

// AsSmall returns the frame around the stereotype and the label.
func (s *usymbolFrame) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	_ geom.HorizontalAlignment) klimt.TextBlock {
	return &frameSmall{label: label, stereotype: stereotype, symbolContext: symbolContext}
}

// frameBig is the anonymous text block of upstream's asBig.
type frameBig struct {
	klimt.TextBlockDefaults
	title, stereotype klimt.TextBlock
	width, height     float64
	symbolContext     *klimt.Fashion
}

func (b *frameBig) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dim := b.CalculateDimension(stringBounder)
	ug = b.symbolContext.Apply(ug)
	dimTitle := b.title.CalculateDimension(stringBounder)
	widthFull := dim.Width()
	drawFrame(ug, widthFull, dim.Height(), dimTitle, b.symbolContext.DeltaShadow(), b.symbolContext.RoundCorner())
	widthTitle := b.title.CalculateDimension(stringBounder).Width()

	// Temporary hack...
	if widthFull-widthTitle < 25 {
		b.title.DrawU(ug.Apply(klimt.NewUTranslate(3, 1)))
	} else {
		ug.Apply(klimt.NewUTranslate(3, 1)).Draw(klimt.NewSpecialText(b.title))
	}

	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	posStereo := (b.width - dimStereo.Width()) / 2

	b.stereotype.DrawU(ug.Apply(klimt.NewUTranslate(4+posStereo, 2+frameYpos(dimTitle))))
}

func (b *frameBig) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(b.width, b.height)
}

// AsBig returns the frame of a size with the title in its corner tab and the
// stereotype below it.
func (s *usymbolFrame) AsBig(title klimt.TextBlock, _ geom.HorizontalAlignment, stereotype klimt.TextBlock,
	width, height float64, symbolContext *klimt.Fashion, _ geom.HorizontalAlignment) klimt.TextBlock {
	return &frameBig{title: title, stereotype: stereotype, width: width, height: height,
		symbolContext: symbolContext}
}
