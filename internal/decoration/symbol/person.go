// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbolPerson.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbolLabel.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbolCollections.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package symbol

import (
	"math"

	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// usymbolPerson is the person symbol: a head over a rounded body that holds
// the text. It has no container form.
type usymbolPerson struct{ usymbolBase }

// SNames returns person.
func (usymbolPerson) SNames() []klimt.SName { return []klimt.SName{klimt.SNamePerson} }

func drawHeadAndBody(ug klimt.UGraphic, shadowing float64, dimBody geom.XDimension2D, headSize float64) {
	head := klimt.UEllipseBuild(headSize, headSize)
	body := klimt.URectangleBuildDim(dimBody).Rounded(headSize)

	body.SetDeltaShadow(shadowing)
	head.SetDeltaShadow(shadowing)

	posx := (dimBody.Width() - headSize) / 2
	ug.Apply(klimt.UTranslateDx(posx)).Draw(head)
	ug.Apply(klimt.UTranslateDy(headSize)).Draw(body)
}

func personHeadSize(dimBody geom.XDimension2D) float64 {
	surface := dimBody.Width() * dimBody.Height()
	return float64(math.Sqrt(surface) * .42)
}

func personMargin() margin { return newMargin(10, 10, 10, 10) }

// personSmall is the anonymous text block of upstream's asSmall.
type personSmall struct {
	klimt.TextBlockDefaults
	label, stereotype klimt.TextBlock
	symbolContext     *klimt.Fashion
	stereoAlignment   geom.HorizontalAlignment
}

func (b *personSmall) DrawU(ug klimt.UGraphic) {
	dimFull := b.CalculateDimension(ug.StringBounder())
	dimBody := b.bodyDimension(ug.StringBounder())
	ug = klimt.UGraphicStencilCreateDim(ug, dimFull)
	ug = b.symbolContext.Apply(ug)
	headSize := personHeadSize(dimBody)
	drawHeadAndBody(ug, b.symbolContext.DeltaShadow(), dimBody, headSize)
	tb := klimt.TextBlockUtilsMergeTB(b.stereotype, b.label, b.stereoAlignment)
	m := personMargin()
	tb.DrawU(ug.Apply(klimt.NewUTranslate(m.x1, m.y1+headSize)))
}

func (b *personSmall) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	body := b.bodyDimension(stringBounder)
	return body.DeltaWH(0, personHeadSize(body))
}

func (b *personSmall) bodyDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimLabel := b.label.CalculateDimension(stringBounder)
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	return personMargin().addDimension(dimStereo.MergeTB(dimLabel))
}

// AsSmall returns the person with the stereotype and the label in its body.
func (usymbolPerson) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	stereoAlignment geom.HorizontalAlignment) klimt.TextBlock {
	return &personSmall{label: label, stereotype: stereotype, symbolContext: symbolContext,
		stereoAlignment: stereoAlignment}
}

// AsBig panics, as upstream throws UnsupportedOperationException.
func (usymbolPerson) AsBig(klimt.TextBlock, geom.HorizontalAlignment, klimt.TextBlock, float64, float64,
	*klimt.Fashion, geom.HorizontalAlignment) klimt.TextBlock {
	throwUnsupported()
	return nil
}

// usymbolLabel is the label symbol: the text alone.
type usymbolLabel struct{ usymbolBase }

// SNames returns label.
func (usymbolLabel) SNames() []klimt.SName { return []klimt.SName{klimt.SNameLabel} }

func labelMargin() margin { return newMargin(10, 10, 10, 10) }

// labelSmall is the anonymous text block of upstream's asSmall.
type labelSmall struct {
	klimt.TextBlockDefaults
	label, stereotype klimt.TextBlock
	symbolContext     *klimt.Fashion
	stereoAlignment   geom.HorizontalAlignment
}

func (b *labelSmall) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = klimt.UGraphicStencilCreateDim(ug, dim)
	ug = b.symbolContext.Apply(ug)
	m := labelMargin()
	tb := klimt.TextBlockUtilsMergeTB(b.stereotype, b.label, b.stereoAlignment)
	tb.DrawU(ug.Apply(klimt.NewUTranslate(m.x1, m.y1)))
}

func (b *labelSmall) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimLabel := b.label.CalculateDimension(stringBounder)
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	return labelMargin().addDimension(dimStereo.MergeTB(dimLabel))
}

// AsSmall returns the stereotype and the label with a margin.
func (usymbolLabel) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	stereoAlignment geom.HorizontalAlignment) klimt.TextBlock {
	return &labelSmall{label: label, stereotype: stereotype, symbolContext: symbolContext,
		stereoAlignment: stereoAlignment}
}

// labelBig is the anonymous text block of upstream's asBig.
type labelBig struct {
	klimt.TextBlockDefaults
	title, stereotype klimt.TextBlock
	labelAlignment    geom.HorizontalAlignment
	width, height     float64
	symbolContext     *klimt.Fashion
	stereoAlignment   geom.HorizontalAlignment
}

func (b *labelBig) DrawU(ug klimt.UGraphic) {
	ug = b.symbolContext.Apply(ug)
	drawBigTitle(ug, b.title, b.stereotype, b.labelAlignment, b.stereoAlignment, b.width, labelMargin())
}

func (b *labelBig) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(b.width, b.height)
}

// AsBig returns the stereotype and the title in an area of a size, without
// a shape.
func (usymbolLabel) AsBig(title klimt.TextBlock, labelAlignment geom.HorizontalAlignment,
	stereotype klimt.TextBlock, width, height float64, symbolContext *klimt.Fashion,
	stereoAlignment geom.HorizontalAlignment) klimt.TextBlock {
	return &labelBig{title: title, labelAlignment: labelAlignment, stereotype: stereotype, width: width,
		height: height, symbolContext: symbolContext, stereoAlignment: stereoAlignment}
}

// collectionsDelta is the offset of the second rectangle of collections.
const collectionsDelta = 4

// usymbolCollections is the collections symbol: two stacked rectangles.
type usymbolCollections struct{ usymbolBase }

// SNames returns collections.
func (usymbolCollections) SNames() []klimt.SName { return []klimt.SName{klimt.SNameCollections} }

func drawCollections(ug klimt.UGraphic, width, height, shadowing, roundCorner float64) {
	small := klimt.URectangleBuild(width-collectionsDelta, height-collectionsDelta).Rounded(roundCorner)
	small.SetDeltaShadow(shadowing)

	ug.Apply(klimt.NewUTranslate(collectionsDelta, collectionsDelta)).Draw(small)
	small.SetDeltaShadow(0)
	ug.Apply(klimt.UTranslateDy(0)).Draw(small)
}

func collectionsMargin() margin { return newMargin(10, 10, 10, 10) }

// collectionsSmall is the anonymous text block of upstream's asSmall.
type collectionsSmall struct {
	klimt.TextBlockDefaults
	label, stereotype klimt.TextBlock
	symbolContext     *klimt.Fashion
	stereoAlignment   geom.HorizontalAlignment
}

func (b *collectionsSmall) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = klimt.UGraphicStencilCreateDim(ug, dim)
	ug = b.symbolContext.Apply(ug)
	drawCollections(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow(), b.symbolContext.RoundCorner())
	m := collectionsMargin()
	tb := klimt.TextBlockUtilsMergeTB(b.stereotype, b.label, b.stereoAlignment)
	tb.DrawU(ug.Apply(klimt.NewUTranslate(m.x1-collectionsDelta/2, m.y1-collectionsDelta/2)))
}

func (b *collectionsSmall) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimLabel := b.label.CalculateDimension(stringBounder)
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	return collectionsMargin().addDimension(dimStereo.MergeTB(dimLabel))
}

// AsSmall returns the collections around the stereotype and the label.
func (usymbolCollections) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	stereoAlignment geom.HorizontalAlignment) klimt.TextBlock {
	return &collectionsSmall{label: label, stereotype: stereotype, symbolContext: symbolContext,
		stereoAlignment: stereoAlignment}
}

// collectionsBig is the anonymous text block of upstream's asBig.
type collectionsBig struct {
	klimt.TextBlockDefaults
	title, stereotype klimt.TextBlock
	width, height     float64
	symbolContext     *klimt.Fashion
	stereoAlignment   geom.HorizontalAlignment
}

func (b *collectionsBig) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = b.symbolContext.Apply(ug)
	drawCollections(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow(), b.symbolContext.RoundCorner())
	// The title is always centered, whatever the label alignment.
	drawBigTitle(ug, b.title, b.stereotype, geom.HorizontalAlignmentCenter, b.stereoAlignment, b.width,
		collectionsMargin())
}

func (b *collectionsBig) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(b.width, b.height)
}

// AsBig returns the collections of a size with the stereotype and the title.
func (usymbolCollections) AsBig(title klimt.TextBlock, _ geom.HorizontalAlignment, stereotype klimt.TextBlock,
	width, height float64, symbolContext *klimt.Fashion, stereoAlignment geom.HorizontalAlignment) klimt.TextBlock {
	return &collectionsBig{title: title, stereotype: stereotype, width: width, height: height,
		symbolContext: symbolContext, stereoAlignment: stereoAlignment}
}
