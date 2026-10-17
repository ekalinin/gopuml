// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbolCard.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbolStorage.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package symbol

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// usymbolCard is the card symbol: a rectangle with a line under the title
// when it is a container.
type usymbolCard struct{ usymbolBase }

// SNames returns card.
func (usymbolCard) SNames() []klimt.SName { return []klimt.SName{klimt.SNameCard} }

func drawCard(ug klimt.UGraphic, width, height, shadowing, top, roundCorner float64) {
	shape := klimt.URectangleBuild(width, height).Rounded(roundCorner)
	shape.SetDeltaShadow(shadowing)

	ug.Draw(shape)
	if top != 0 {
		ug.Apply(klimt.UTranslateDy(top)).Draw(klimt.ULineH(width))
	}
}

func cardMargin() margin { return newMargin(10, 10, 3, 3) }

// cardSmall is the anonymous text block of upstream's asSmall.
type cardSmall struct {
	klimt.TextBlockDefaults
	label, stereotype klimt.TextBlock
	symbolContext     *klimt.Fashion
}

func (b *cardSmall) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = klimt.UGraphicStencilCreateDim(ug, dim)
	ug = b.symbolContext.Apply(ug)
	drawCard(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow(), 0, b.symbolContext.RoundCorner())
	m := cardMargin()
	tb := klimt.TextBlockUtilsMergeTB(b.stereotype, b.label, geom.HorizontalAlignmentCenter)
	tb.DrawU(ug.Apply(klimt.NewUTranslate(m.x1, m.y1)))
}

func (b *cardSmall) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimLabel := b.label.CalculateDimension(stringBounder)
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	return cardMargin().addDimension(dimStereo.MergeTB(dimLabel))
}

// AsSmall returns the card around the stereotype and the label.
func (usymbolCard) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	_ geom.HorizontalAlignment) klimt.TextBlock {
	return &cardSmall{label: label, stereotype: stereotype, symbolContext: symbolContext}
}

// cardBig is the anonymous text block of upstream's asBig.
type cardBig struct {
	klimt.TextBlockDefaults
	title, stereotype klimt.TextBlock
	width, height     float64
	symbolContext     *klimt.Fashion
}

func (b *cardBig) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = b.symbolContext.Apply(ug)
	dimStereo := b.stereotype.CalculateDimension(ug.StringBounder())
	dimTitle := b.title.CalculateDimension(ug.StringBounder())
	drawCard(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow(), dimTitle.Height()+dimStereo.Height()+4,
		b.symbolContext.RoundCorner())
	posStereo := (b.width - dimStereo.Width()) / 2
	b.stereotype.DrawU(ug.Apply(klimt.NewUTranslate(posStereo, 2)))
	posTitle := (b.width - dimTitle.Width()) / 2
	b.title.DrawU(ug.Apply(klimt.NewUTranslate(posTitle, 2+dimStereo.Height())))
}

func (b *cardBig) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(b.width, b.height)
}

// AsBig returns the card of a size with the stereotype and the title above
// a line.
func (usymbolCard) AsBig(title klimt.TextBlock, _ geom.HorizontalAlignment, stereotype klimt.TextBlock,
	width, height float64, symbolContext *klimt.Fashion, _ geom.HorizontalAlignment) klimt.TextBlock {
	return &cardBig{title: title, stereotype: stereotype, width: width, height: height, symbolContext: symbolContext}
}

// usymbolStorage is the storage symbol: a rectangle with very rounded
// corners.
type usymbolStorage struct{ usymbolBase }

// SNames returns storage.
func (usymbolStorage) SNames() []klimt.SName { return []klimt.SName{klimt.SNameStorage} }

func drawStorage(ug klimt.UGraphic, width, height, shadowing float64) {
	shape := klimt.URectangleBuild(width, height).Rounded(70)
	shape.SetDeltaShadow(shadowing)

	ug.Draw(shape)
}

func storageMargin() margin { return newMargin(10, 10, 10, 10) }

// storageSmall is the anonymous text block of upstream's asSmall.
type storageSmall struct {
	klimt.TextBlockDefaults
	label, stereotype klimt.TextBlock
	symbolContext     *klimt.Fashion
}

func (b *storageSmall) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = klimt.UGraphicStencilCreateDim(ug, dim)
	ug = b.symbolContext.Apply(ug)
	drawStorage(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow())
	m := storageMargin()
	tb := klimt.TextBlockUtilsMergeTB(b.stereotype, b.label, geom.HorizontalAlignmentCenter)
	tb.DrawU(ug.Apply(klimt.NewUTranslate(m.x1, m.y1)))
}

func (b *storageSmall) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimLabel := b.label.CalculateDimension(stringBounder)
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	return storageMargin().addDimension(dimStereo.MergeTB(dimLabel))
}

// AsSmall returns the storage around the stereotype and the label.
func (usymbolStorage) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	_ geom.HorizontalAlignment) klimt.TextBlock {
	return &storageSmall{label: label, stereotype: stereotype, symbolContext: symbolContext}
}

// storageBig is the anonymous text block of upstream's asBig.
type storageBig struct {
	klimt.TextBlockDefaults
	title, stereotype klimt.TextBlock
	width, height     float64
	symbolContext     *klimt.Fashion
}

func (b *storageBig) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = b.symbolContext.Apply(ug)
	drawStorage(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow())
	drawBigCentered(ug, b.title, b.stereotype, b.width, 5, 7)
}

func (b *storageBig) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(b.width, b.height)
}

// AsBig returns the storage of a size with the stereotype and the title.
func (usymbolStorage) AsBig(title klimt.TextBlock, _ geom.HorizontalAlignment, stereotype klimt.TextBlock,
	width, height float64, symbolContext *klimt.Fashion, _ geom.HorizontalAlignment) klimt.TextBlock {
	return &storageBig{title: title, stereotype: stereotype, width: width, height: height,
		symbolContext: symbolContext}
}
