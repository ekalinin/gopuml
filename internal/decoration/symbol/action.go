// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbolAction.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbolProcess.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package symbol

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// usymbolAction is the action symbol: a pentagon pointing right. Upstream
// ignores the shadow and the corners of the fashion.
type usymbolAction struct {
	usymbolBase
	sname klimt.SName
}

// SNames returns the style name of the symbol.
func (s *usymbolAction) SNames() []klimt.SName { return []klimt.SName{s.sname} }

func drawAction(ug klimt.UGraphic, width, height float64) {
	shape := klimt.NewUPolygon()
	shape.AddPoint(0, 0)
	shape.AddPoint(width-10, 0)
	shape.AddPoint(width, height/2)
	shape.AddPoint(width-10, height)
	shape.AddPoint(0, height)
	ug.Draw(shape)
}

// drawProcess draws the process symbol: an arrow-shaped hexagon.
func drawProcess(ug klimt.UGraphic, width, height float64) {
	shape := klimt.NewUPolygon()
	shape.AddPoint(0, 0)
	shape.AddPoint(width-10, 0)
	shape.AddPoint(width, height/2)
	shape.AddPoint(width-10, height)
	shape.AddPoint(0, height)
	shape.AddPoint(10, height/2)
	ug.Draw(shape)
}

// actionHTitle is the vertical space of the title of a big action or
// process.
func actionHTitle(dimTitle geom.XDimension2D) float64 {
	if dimTitle.Width() == 0 {
		return 10
	}
	return dimTitle.Height()
}

// arrowSmall is the anonymous text block of upstream's asSmall, for the
// action and the process symbols.
type arrowSmall struct {
	klimt.TextBlockDefaults
	label, stereotype klimt.TextBlock
	symbolContext     *klimt.Fashion
	stereoAlignment   geom.HorizontalAlignment
	draw              func(ug klimt.UGraphic, width, height float64)
	margin            margin
}

func (b *arrowSmall) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = klimt.UGraphicStencilCreateDim(ug, dim)
	ug = b.symbolContext.Apply(ug)
	b.draw(ug, dim.Width(), dim.Height())
	tb := klimt.TextBlockUtilsMergeTB(b.stereotype, b.label, b.stereoAlignment)
	tb.DrawU(ug.Apply(klimt.NewUTranslate(b.margin.x1, b.margin.y1)))
}

func (b *arrowSmall) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimLabel := b.label.CalculateDimension(stringBounder)
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	return b.margin.addDimension(dimStereo.MergeTB(dimLabel))
}

// arrowBig is the anonymous text block of upstream's asBig, for the action
// and the process symbols: the title first, the stereotype below it.
type arrowBig struct {
	klimt.TextBlockDefaults
	title, stereotype klimt.TextBlock
	width, height     float64
	symbolContext     *klimt.Fashion
	draw              func(ug klimt.UGraphic, width, height float64)
}

func (b *arrowBig) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dim := b.CalculateDimension(stringBounder)
	ug = b.symbolContext.Apply(ug)
	dimTitle := b.title.CalculateDimension(stringBounder)
	b.draw(ug, dim.Width(), dim.Height())
	posTitle := (b.width - dimTitle.Width()) / 2
	b.title.DrawU(ug.Apply(klimt.NewUTranslate(posTitle, 2)))
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	posStereo := (b.width - dimStereo.Width()) / 2

	b.stereotype.DrawU(ug.Apply(klimt.NewUTranslate(4+posStereo, 2+actionHTitle(dimTitle))))
}

func (b *arrowBig) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(b.width, b.height)
}

// AsSmall returns the action around the stereotype and the label.
func (s *usymbolAction) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	stereoAlignment geom.HorizontalAlignment) klimt.TextBlock {
	return &arrowSmall{label: label, stereotype: stereotype, symbolContext: symbolContext,
		stereoAlignment: stereoAlignment, draw: drawAction, margin: newMargin(10, 20, 10, 10)}
}

// AsBig returns the action of a size with the title and the stereotype.
func (s *usymbolAction) AsBig(title klimt.TextBlock, _ geom.HorizontalAlignment, stereotype klimt.TextBlock,
	width, height float64, symbolContext *klimt.Fashion, _ geom.HorizontalAlignment) klimt.TextBlock {
	return &arrowBig{title: title, stereotype: stereotype, width: width, height: height,
		symbolContext: symbolContext, draw: drawAction}
}

// usymbolProcess is the process symbol: an arrow-shaped hexagon. Upstream
// ignores the shadow and the corners of the fashion.
type usymbolProcess struct {
	usymbolBase
	sname klimt.SName
}

// SNames returns the style name of the symbol.
func (s *usymbolProcess) SNames() []klimt.SName { return []klimt.SName{s.sname} }

// AsSmall returns the process around the stereotype and the label.
func (s *usymbolProcess) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	stereoAlignment geom.HorizontalAlignment) klimt.TextBlock {
	return &arrowSmall{label: label, stereotype: stereotype, symbolContext: symbolContext,
		stereoAlignment: stereoAlignment, draw: drawProcess, margin: newMargin(20, 20, 10, 10)}
}

// AsBig returns the process of a size with the title and the stereotype.
func (s *usymbolProcess) AsBig(title klimt.TextBlock, _ geom.HorizontalAlignment, stereotype klimt.TextBlock,
	width, height float64, symbolContext *klimt.Fashion, _ geom.HorizontalAlignment) klimt.TextBlock {
	return &arrowBig{title: title, stereotype: stereotype, width: width, height: height,
		symbolContext: symbolContext, draw: drawProcess}
}
