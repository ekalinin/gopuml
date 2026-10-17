// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbolFolder.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package symbol

import (
	"math"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// The margins of the tab of a folder.
const (
	folderMarginTitleX1 = 3
	folderMarginTitleX2 = 3
	folderMarginTitleX3 = 7
	folderMarginTitleY1 = 3
	folderMarginTitleY2 = 3
)

// usymbolFolder is the folder symbol: a rectangle with a tab, which shows
// the title of a package (showTitle) or stays empty.
type usymbolFolder struct {
	usymbolBase
	sname     klimt.SName
	showTitle bool
}

// SNames returns the style name of the symbol.
func (s *usymbolFolder) SNames() []klimt.SName { return []klimt.SName{s.sname} }

func (s *usymbolFolder) drawFolder(ug klimt.UGraphic, width, height float64, dimName geom.XDimension2D,
	shadowing, roundCorner float64) {
	wtitle := folderWTitle(width, dimName)
	htitle := folderHTitle(dimName)

	var shape klimt.Shadowable
	if roundCorner == 0 {
		poly := klimt.NewUPolygon()
		poly.AddPoint(0, 0)
		poly.AddPoint(wtitle, 0)

		poly.AddPoint(wtitle+folderMarginTitleX3, htitle)
		poly.AddPoint(width, htitle)
		poly.AddPoint(width, height)
		poly.AddPoint(0, height)
		poly.AddPoint(0, 0)
		shape = poly
	} else {
		path := klimt.UPathNone()
		path.MoveTo(roundCorner/2, 0)
		path.LineTo(wtitle-roundCorner/2, 0)
		path.ArcToPoint(geom.NewXPoint2D(wtitle, roundCorner/2), roundCorner/2*1.5, 0, 1)
		path.LineTo(wtitle+folderMarginTitleX3, htitle)
		path.LineTo(width-roundCorner/2, htitle)
		path.ArcToPoint(geom.NewXPoint2D(width, htitle+roundCorner/2), roundCorner/2, 0, 1)
		path.LineTo(width, height-roundCorner/2)
		path.ArcToPoint(geom.NewXPoint2D(width-roundCorner/2, height), roundCorner/2, 0, 1)
		path.LineTo(roundCorner/2, height)
		path.ArcToPoint(geom.NewXPoint2D(0, height-roundCorner/2), roundCorner/2, 0, 1)
		path.LineTo(0, roundCorner/2)
		path.ArcToPoint(geom.NewXPoint2D(roundCorner/2, 0), roundCorner/2, 0, 1)
		path.ClosePath()
		shape = path
	}
	shape.SetDeltaShadow(shadowing)

	ug.Draw(shape)
	ug.Apply(klimt.UTranslateDy(htitle)).Draw(klimt.ULineH(wtitle + folderMarginTitleX3))
}

func folderWTitle(width float64, dimTitle geom.XDimension2D) float64 {
	if dimTitle.Width() == 0 {
		return math.Max(30, width/4)
	}
	return dimTitle.Width() + folderMarginTitleX1 + folderMarginTitleX2
}

func folderHTitle(dimTitle geom.XDimension2D) float64 {
	if dimTitle.Width() == 0 {
		return 10
	}
	return dimTitle.Height() + folderMarginTitleY1 + folderMarginTitleY2
}

func folderMargin() margin { return newMargin(10, 10+10, 10+3, 10) }

// folderForce is the magnetic border of a folder: links that end on the tab
// line are pushed down to the body.
func folderForce(position geom.XPoint2D, dim, dimTitle geom.XDimension2D) klimt.UTranslate {
	wtitle := folderWTitle(dim.Width(), dimTitle)
	htitle := folderHTitle(dimTitle)

	if position.X >= wtitle && position.Y >= 0 && position.Y <= htitle {
		return klimt.NewUTranslate(0, htitle)
	}
	if position.Y <= 0 && position.X >= wtitle+folderMarginTitleX3 {
		return klimt.NewUTranslate(0, htitle)
	}
	if position.Y <= 0 && position.X >= wtitle-folderMarginTitleX3 {
		delta := position.X - (wtitle - folderMarginTitleX3)
		how := delta / (2 * folderMarginTitleX3)
		return klimt.NewUTranslate(0, float64(htitle*how))
	}
	return klimt.UTranslateNone()
}

// folderSmall is the anonymous text block of upstream's asSmall.
type folderSmall struct {
	klimt.TextBlockDefaults
	s                        *usymbolFolder
	title, label, stereotype klimt.TextBlock
	symbolContext            *klimt.Fashion
}

func (b *folderSmall) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = klimt.UGraphicStencilCreateDim(ug, dim)
	ug = b.symbolContext.Apply(ug)
	dimTitle := b.dimTitle(ug.StringBounder())
	b.s.drawFolder(ug, dim.Width(), dim.Height(), dimTitle, b.symbolContext.DeltaShadow(),
		b.symbolContext.RoundCorner())
	m := folderMargin()
	tb := klimt.TextBlockUtilsMergeTB(b.stereotype, b.label, geom.HorizontalAlignmentCenter)
	if b.s.showTitle {
		b.title.DrawU(ug.Apply(klimt.NewUTranslate(4, 3)))
	}
	tb.DrawU(ug.Apply(klimt.NewUTranslate(m.x1, m.y1+dimTitle.Height())))
}

func (b *folderSmall) dimTitle(stringBounder font.StringBounder) geom.XDimension2D {
	if b.s.showTitle {
		return b.title.CalculateDimension(stringBounder)
	}
	return geom.NewXDimension2D(40, 15)
}

func (b *folderSmall) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimName := b.dimTitle(stringBounder)
	dimLabel := b.label.CalculateDimension(stringBounder)
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	return folderMargin().addDimension(dimName.MergeTB3(dimStereo, dimLabel))
}

func (b *folderSmall) MagneticBorder() klimt.MagneticBorder {
	return klimt.MagneticBorderFunc(func(stringBounder font.StringBounder, position geom.XPoint2D) klimt.UTranslate {
		return folderForce(position, b.CalculateDimension(stringBounder), b.dimTitle(stringBounder))
	})
}

// AsSmall returns the folder around the stereotype and the label, with the
// title in the tab when the symbol shows it. A nil title panics, as
// upstream's requireNonNull.
func (s *usymbolFolder) AsSmall(title, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	_ geom.HorizontalAlignment) klimt.TextBlock {
	if title == nil {
		throwNPE()
	}
	return &folderSmall{s: s, title: title, label: label, stereotype: stereotype, symbolContext: symbolContext}
}

// folderBig is the anonymous text block of upstream's asBig.
type folderBig struct {
	klimt.TextBlockDefaults
	s                 *usymbolFolder
	title, stereotype klimt.TextBlock
	width, height     float64
	symbolContext     *klimt.Fashion
}

func (b *folderBig) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dim := b.CalculateDimension(stringBounder)
	ug = b.symbolContext.Apply(ug)
	dimTitle := b.title.CalculateDimension(stringBounder)
	b.s.drawFolder(ug, dim.Width(), dim.Height(), dimTitle, b.symbolContext.DeltaShadow(),
		b.symbolContext.RoundCorner())
	b.title.DrawU(ug.Apply(klimt.NewUTranslate(4, 2)))
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	posStereo := (b.width - dimStereo.Width()) / 2

	b.stereotype.DrawU(ug.Apply(klimt.NewUTranslate(4+posStereo, 2+folderHTitle(dimTitle))))
}

func (b *folderBig) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(b.width, b.height)
}

func (b *folderBig) MagneticBorder() klimt.MagneticBorder {
	return klimt.MagneticBorderFunc(func(stringBounder font.StringBounder, position geom.XPoint2D) klimt.UTranslate {
		return folderForce(position, b.CalculateDimension(stringBounder), b.title.CalculateDimension(stringBounder))
	})
}

// AsBig returns the folder of a size with the title in the tab and the
// stereotype below it.
func (s *usymbolFolder) AsBig(title klimt.TextBlock, _ geom.HorizontalAlignment, stereotype klimt.TextBlock,
	width, height float64, symbolContext *klimt.Fashion, _ geom.HorizontalAlignment) klimt.TextBlock {
	return &folderBig{s: s, title: title, stereotype: stereotype, width: width, height: height,
		symbolContext: symbolContext}
}
