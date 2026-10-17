// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbolDatabase.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package symbol

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// usymbolDatabase is the database symbol: a cylinder.
type usymbolDatabase struct{ usymbolBase }

// SNames returns database.
func (usymbolDatabase) SNames() []klimt.SName { return []klimt.SName{klimt.SNameDatabase} }

func drawDatabase(ug klimt.UGraphic, width, height, shadowing float64) {
	shape := klimt.UPathNone()
	shape.SetDeltaShadow(shadowing)

	shape.MoveTo(0, 10)
	shape.CubicTo(0, 0, width/2, 0, width/2, 0)
	shape.CubicTo(width/2, 0, width, 0, width, 10)
	shape.LineTo(width, height-10)
	shape.CubicTo(width, height, width/2, height, width/2, height)
	shape.CubicTo(width/2, height, 0, height, 0, height-10)
	shape.LineTo(0, 10)

	ug.Draw(shape)

	closing := databaseClosingPath(width)
	ug.Apply(klimt.Bg(color.None())).Draw(closing)
	ug.Apply(klimt.NewUTranslate(width, height)).Draw(klimt.NewUEmpty(10, 10))
}

func databaseClosingPath(width float64) *klimt.UPath {
	closing := klimt.UPathNone()
	closing.MoveTo(0, 10)
	closing.CubicTo(0, 20, width/2, 20, width/2, 20)
	closing.CubicTo(width/2, 20, width, 20, width, 10)
	return closing
}

// ugraphicDatabase is upstream's MyUGraphicDatabase: separators of the text
// are drawn as the closing curve of the cylinder.
type ugraphicDatabase struct {
	klimt.AbstractUGraphicHorizontalLine
	endingX float64
}

func newUGraphicDatabase(ug klimt.UGraphic, endingX float64) *ugraphicDatabase {
	g := &ugraphicDatabase{AbstractUGraphicHorizontalLine: klimt.NewAbstractUGraphicHorizontalLine(ug),
		endingX: endingX}
	g.SetSelf(g)
	return g
}

// CopyHorizontalLine returns the graphic over ug.
func (g *ugraphicDatabase) CopyHorizontalLine(ug klimt.UGraphic) klimt.HorizontalLineGraphic {
	return newUGraphicDatabase(ug, g.endingX)
}

// DrawHline draws the separator as a closing curve and its title.
func (g *ugraphicDatabase) DrawHline(ug klimt.UGraphic, line *klimt.UHorizontalLine, translate klimt.UTranslate) {
	closing := databaseClosingPath(g.endingX)
	ug = ug.Apply(translate)
	ug.Apply(line.Stroke()).Apply(klimt.Bg(color.None())).Apply(klimt.UTranslateDy(-15)).Draw(closing)
	if line.IsDouble() {
		ug.Apply(line.Stroke()).Apply(klimt.Bg(color.None())).Apply(klimt.UTranslateDy(-15 + 2)).Draw(closing)
	}
	line.DrawTitleInternal(ug, 0, g.endingX, 0, true)
}

func databaseMargin() margin { return newMargin(10, 10, 24, 5) }

// databaseSmall is the anonymous text block of upstream's asSmall.
type databaseSmall struct {
	klimt.TextBlockDefaults
	label, stereotype klimt.TextBlock
	symbolContext     *klimt.Fashion
}

func (b *databaseSmall) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = b.symbolContext.Apply(ug)
	drawDatabase(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow())
	m := databaseMargin()
	tb := klimt.TextBlockUtilsMergeTB(b.stereotype, b.label, geom.HorizontalAlignmentCenter)
	ug2 := newUGraphicDatabase(ug, dim.Width())
	tb.DrawU(ug2.Apply(klimt.NewUTranslate(m.x1, m.y1)))
}

func (b *databaseSmall) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimLabel := b.label.CalculateDimension(stringBounder)
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	return databaseMargin().addDimension(dimStereo.MergeTB(dimLabel))
}

// AsSmall returns the cylinder around the stereotype and the label.
func (usymbolDatabase) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	_ geom.HorizontalAlignment) klimt.TextBlock {
	return &databaseSmall{label: label, stereotype: stereotype, symbolContext: symbolContext}
}

// databaseBig is the anonymous text block of upstream's asBig.
type databaseBig struct {
	klimt.TextBlockDefaults
	title, stereotype klimt.TextBlock
	width, height     float64
	symbolContext     *klimt.Fashion
}

func (b *databaseBig) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = b.symbolContext.Apply(ug)
	drawDatabase(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow())
	dimStereo := b.stereotype.CalculateDimension(ug.StringBounder())
	posStereo := (b.width - dimStereo.Width()) / 2
	b.stereotype.DrawU(ug.Apply(klimt.NewUTranslate(posStereo, 2+20)))

	dimTitle := b.title.CalculateDimension(ug.StringBounder())
	posTitle := (b.width - dimTitle.Width()) / 2
	b.title.DrawU(ug.Apply(klimt.NewUTranslate(posTitle, 2+20+dimStereo.Height())))
}

func (b *databaseBig) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(b.width, b.height)
}

// AsBig returns the cylinder of a size with the stereotype and the title.
func (usymbolDatabase) AsBig(title klimt.TextBlock, _ geom.HorizontalAlignment, stereotype klimt.TextBlock,
	width, height float64, symbolContext *klimt.Fashion, _ geom.HorizontalAlignment) klimt.TextBlock {
	return &databaseBig{title: title, stereotype: stereotype, width: width, height: height,
		symbolContext: symbolContext}
}

// SuppHeightBecauseOfShape returns 15.
func (usymbolDatabase) SuppHeightBecauseOfShape() int { return 15 }
