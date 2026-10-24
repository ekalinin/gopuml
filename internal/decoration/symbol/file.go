// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/decoration/symbol/USymbolFile.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/decoration/symbol/USymbolStack.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/decoration/symbol/USymbolHexagon.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package symbol

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// usymbolFile is the file symbol: a sheet with a folded corner.
type usymbolFile struct{ usymbolBase }

// SNames returns file.
func (usymbolFile) SNames() []klimt.SName { return []klimt.SName{klimt.SNameFile} }

func drawFile(ug klimt.UGraphic, width, height, shadowing, roundCorner float64) {
	const cornersize = 10
	var out klimt.Shadowable
	if roundCorner == 0 {
		polygon := klimt.NewUPolygon()
		polygon.AddPoint(0, 0)
		polygon.AddPoint(0, height)
		polygon.AddPoint(width, height)
		polygon.AddPoint(width, cornersize)
		polygon.AddPoint(width-cornersize, 0)
		polygon.AddPoint(0, 0)
		out = polygon
	} else {
		path := klimt.UPathNone()
		path.MoveTo(0, roundCorner/2)
		path.LineTo(0, height-roundCorner/2)
		path.ArcToPoint(geom.NewXPoint2D(roundCorner/2, height), roundCorner/2, 0, 0)
		path.LineTo(width-roundCorner/2, height)
		path.ArcToPoint(geom.NewXPoint2D(width, height-roundCorner/2), roundCorner/2, 0, 0)
		path.LineTo(width, cornersize)
		path.LineTo(width-cornersize, 0)
		path.LineTo(roundCorner/2, 0)
		path.ArcToPoint(geom.NewXPoint2D(0, roundCorner/2), roundCorner/2, 0, 0)
		out = path
	}

	out.SetDeltaShadow(shadowing)
	ug.Draw(out)

	path := klimt.UPathNone()
	path.MoveTo(width-cornersize, 0)
	if roundCorner == 0 {
		path.LineTo(width-cornersize, cornersize)
	} else {
		path.LineTo(width-cornersize, cornersize-roundCorner/2)
		path.ArcToPoint(geom.NewXPoint2D(width-cornersize+roundCorner/2, cornersize), roundCorner/2, 0, 0)
	}
	path.LineTo(width, cornersize)
	ug.Draw(path)
}

func fileMargin() margin { return newMargin(10, 10, 10, 10) }

// fileSmall is the anonymous text block of upstream's asSmall.
type fileSmall struct {
	klimt.TextBlockDefaults
	label, stereotype klimt.TextBlock
	symbolContext     *klimt.Fashion
}

func (b *fileSmall) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = klimt.UGraphicStencilCreateDim(ug, dim)
	ug = b.symbolContext.Apply(ug)
	drawFile(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow(), b.symbolContext.RoundCorner())
	m := fileMargin()
	tb := klimt.TextBlockUtilsMergeTB(b.stereotype, b.label, geom.HorizontalAlignmentCenter)
	tb.DrawU(ug.Apply(klimt.NewUTranslate(m.x1, m.y1)))
}

func (b *fileSmall) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimLabel := b.label.CalculateDimension(stringBounder)
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	return fileMargin().addDimension(dimStereo.MergeTB(dimLabel))
}

// AsSmall returns the file around the stereotype and the label.
func (usymbolFile) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	_ geom.HorizontalAlignment) klimt.TextBlock {
	return &fileSmall{label: label, stereotype: stereotype, symbolContext: symbolContext}
}

// fileBig is the anonymous text block of upstream's asBig.
type fileBig struct {
	klimt.TextBlockDefaults
	title, stereotype klimt.TextBlock
	width, height     float64
	symbolContext     *klimt.Fashion
}

func (b *fileBig) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = b.symbolContext.Apply(ug)
	drawFile(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow(), b.symbolContext.RoundCorner())
	// The stereotype alignment of upstream's file is always CENTER.
	drawBigCentered(ug, b.title, b.stereotype, b.width, 2, 2)
}

func (b *fileBig) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(b.width, b.height)
}

// AsBig returns the file of a size with the stereotype and the title.
func (usymbolFile) AsBig(title klimt.TextBlock, _ geom.HorizontalAlignment, stereotype klimt.TextBlock,
	width, height float64, symbolContext *klimt.Fashion, _ geom.HorizontalAlignment) klimt.TextBlock {
	return &fileBig{title: title, stereotype: stereotype, width: width, height: height, symbolContext: symbolContext}
}

// usymbolStack is the stack symbol: a box open at the top with two rims.
type usymbolStack struct{ usymbolBase }

// SNames returns stack.
func (usymbolStack) SNames() []klimt.SName { return []klimt.SName{klimt.SNameStack} }

func drawStack(ug klimt.UGraphic, width, height, shadowing, roundCorner float64) {
	const border = 15

	rect := klimt.URectangleBuild(width-2*border, height).Rounded(roundCorner)
	ug.Apply(color.None()).Apply(klimt.UTranslateDx(border)).Draw(rect)

	path := klimt.UPathNone()
	if roundCorner == 0 {
		path.MoveTo(0, 0)
		path.LineTo(border, 0)
		path.LineTo(border, height)
		path.LineTo(width-border, height)
		path.LineTo(width-border, 0)
		path.LineTo(width, 0)
	} else {
		path.MoveTo(0, 0)
		path.LineTo(border-roundCorner/2, 0)
		path.ArcToPoint(geom.NewXPoint2D(border, roundCorner/2), roundCorner/2, 0, 1)
		path.LineTo(border, height-roundCorner/2)
		path.ArcToPoint(geom.NewXPoint2D(border+roundCorner/2, height), roundCorner/2, 0, 0)
		path.LineTo(width-border-roundCorner/2, height)
		path.ArcToPoint(geom.NewXPoint2D(width-border, height-roundCorner/2), roundCorner/2, 0, 0)
		path.LineTo(width-border, roundCorner/2)
		path.ArcToPoint(geom.NewXPoint2D(width-border+roundCorner/2, 0), roundCorner/2, 0, 1)
		path.LineTo(width, 0)
	}
	path.SetDeltaShadow(shadowing)
	ug.Apply(klimt.Bg(color.None())).Draw(path)
}

func stackMargin() margin { return newMargin(25, 25, 10, 10) }

// stackSmall is the anonymous text block of upstream's asSmall.
type stackSmall struct {
	klimt.TextBlockDefaults
	label, stereotype klimt.TextBlock
	symbolContext     *klimt.Fashion
}

func (b *stackSmall) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = klimt.UGraphicStencilCreateDim(ug, dim)
	ug = b.symbolContext.Apply(ug)
	drawStack(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow(), b.symbolContext.RoundCorner())
	m := stackMargin()
	tb := klimt.TextBlockUtilsMergeTB(b.stereotype, b.label, geom.HorizontalAlignmentCenter)
	tb.DrawU(ug.Apply(klimt.NewUTranslate(m.x1, m.y1)))
}

func (b *stackSmall) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimLabel := b.label.CalculateDimension(stringBounder)
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	return stackMargin().addDimension(dimStereo.MergeTB(dimLabel))
}

// AsSmall returns the stack around the stereotype and the label.
func (usymbolStack) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	_ geom.HorizontalAlignment) klimt.TextBlock {
	return &stackSmall{label: label, stereotype: stereotype, symbolContext: symbolContext}
}

// stackBig is the anonymous text block of upstream's asBig.
type stackBig struct {
	klimt.TextBlockDefaults
	title, stereotype klimt.TextBlock
	width, height     float64
	symbolContext     *klimt.Fashion
}

func (b *stackBig) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = b.symbolContext.Apply(ug)
	drawStack(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow(), b.symbolContext.RoundCorner())
	drawBigCentered(ug, b.title, b.stereotype, b.width, 13, 13)
}

func (b *stackBig) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(b.width, b.height)
}

// AsBig returns the stack of a size with the stereotype and the title.
func (usymbolStack) AsBig(title klimt.TextBlock, _ geom.HorizontalAlignment, stereotype klimt.TextBlock,
	width, height float64, symbolContext *klimt.Fashion, _ geom.HorizontalAlignment) klimt.TextBlock {
	return &stackBig{title: title, stereotype: stereotype, width: width, height: height, symbolContext: symbolContext}
}

// hexagonMarginY is the vertical margin of the small hexagon.
const hexagonMarginY = 5

// usymbolHexagon is the hexagon symbol. As a small symbol it only places the
// text (the hexagon itself is drawn by the entity image).
type usymbolHexagon struct{ usymbolBase }

// SNames returns hexagon.
func (usymbolHexagon) SNames() []klimt.SName { return []klimt.SName{klimt.SNameHexagon} }

// hexagonSmall is the anonymous text block of upstream's asSmall.
type hexagonSmall struct {
	klimt.TextBlockDefaults
	label, stereotype klimt.TextBlock
	stereoAlignment   geom.HorizontalAlignment
}

func (b *hexagonSmall) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	tb := klimt.TextBlockUtilsMergeTB(b.stereotype, b.label, b.stereoAlignment)
	deltaX := dim.Width() / 4
	tb.DrawU(ug.Apply(klimt.NewUTranslate(deltaX, hexagonMarginY)))
}

func (b *hexagonSmall) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimLabel := b.label.CalculateDimension(stringBounder)
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	full := dimStereo.MergeTB(dimLabel)
	return geom.NewXDimension2D(full.Width()*2, full.Height()+2*hexagonMarginY)
}

// AsSmall returns the stereotype and the label laid out for a hexagon.
func (usymbolHexagon) AsSmall(_, label, stereotype klimt.TextBlock, _ *klimt.Fashion,
	stereoAlignment geom.HorizontalAlignment) klimt.TextBlock {
	return &hexagonSmall{label: label, stereotype: stereotype, stereoAlignment: stereoAlignment}
}

func drawHexagon(ug klimt.UGraphic, width, height, shadowing float64) {
	shape := klimt.UPathNone()
	dx := width / 8
	shape.MoveTo(0, height/2)
	shape.LineTo(dx, 0)
	shape.LineTo(width-dx, 0)
	shape.LineTo(width, height/2)
	shape.LineTo(width-dx, height)
	shape.LineTo(dx, height)
	shape.LineTo(0, height/2)
	shape.ClosePath()

	shape.SetDeltaShadow(shadowing)

	ug.Draw(shape)
}

func hexagonMargin() margin { return newMargin(10, 10, 10, 10) }

// hexagonBig is the anonymous text block of upstream's asBig.
type hexagonBig struct {
	klimt.TextBlockDefaults
	title, stereotype klimt.TextBlock
	labelAlignment    geom.HorizontalAlignment
	width, height     float64
	symbolContext     *klimt.Fashion
	stereoAlignment   geom.HorizontalAlignment
}

func (b *hexagonBig) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = b.symbolContext.Apply(ug)
	drawHexagon(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow())
	drawBigTitle(ug, b.title, b.stereotype, b.labelAlignment, b.stereoAlignment, b.width, hexagonMargin())
}

func (b *hexagonBig) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(b.width, b.height)
}

// AsBig returns the hexagon of a size with the stereotype and the title.
func (usymbolHexagon) AsBig(title klimt.TextBlock, labelAlignment geom.HorizontalAlignment,
	stereotype klimt.TextBlock, width, height float64, symbolContext *klimt.Fashion,
	stereoAlignment geom.HorizontalAlignment) klimt.TextBlock {
	return &hexagonBig{title: title, labelAlignment: labelAlignment, stereotype: stereotype, width: width,
		height: height, symbolContext: symbolContext, stereoAlignment: stereoAlignment}
}
