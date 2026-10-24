// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/decoration/symbol/USymbolComponent1.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/decoration/symbol/USymbolComponent2.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/decoration/symbol/USymbolArtifact.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package symbol

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// usymbolComponent1 is the UML 1 component symbol: a rectangle with two
// small rectangles on its left side. As a container it is drawn as the UML 2
// component.
type usymbolComponent1 struct{ usymbolBase }

// SNames returns component.
func (usymbolComponent1) SNames() []klimt.SName { return []klimt.SName{klimt.SNameComponent} }

func drawComponent1(ug klimt.UGraphic, widthTotal, heightTotal, shadowing, roundCorner float64) {
	form := klimt.URectangleBuild(widthTotal, heightTotal).Rounded(roundCorner)
	form.SetDeltaShadow(shadowing)

	ug.Draw(form)
	small := klimt.URectangleBuild(10, 5)

	// UML 1 Component Notation
	ug.Apply(klimt.NewUTranslate(-5, 5)).Draw(small)
	ug.Apply(klimt.NewUTranslate(-5, heightTotal-10)).Draw(small)
}

func component1Margin() margin { return newMargin(10, 10, 10, 10) }

// component1Small is the anonymous text block of upstream's asSmall.
type component1Small struct {
	klimt.TextBlockDefaults
	label, stereotype klimt.TextBlock
	symbolContext     *klimt.Fashion
}

func (b *component1Small) DrawU(ug klimt.UGraphic) {
	dimTotal := b.CalculateDimension(ug.StringBounder())
	ug = klimt.UGraphicStencilCreateDim(ug, dimTotal)
	ug = b.symbolContext.Apply(ug)
	drawComponent1(ug, dimTotal.Width(), dimTotal.Height(), b.symbolContext.DeltaShadow(),
		b.symbolContext.RoundCorner())
	m := component1Margin()
	tb := klimt.TextBlockUtilsMergeTB(b.stereotype, b.label, geom.HorizontalAlignmentCenter)
	tb.DrawU(ug.Apply(klimt.NewUTranslate(m.x1, m.y1)))
}

func (b *component1Small) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimLabel := b.label.CalculateDimension(stringBounder)
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	return component1Margin().addDimension(dimStereo.MergeTB(dimLabel))
}

// AsSmall returns the component around the stereotype and the label.
func (usymbolComponent1) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	_ geom.HorizontalAlignment) klimt.TextBlock {
	return &component1Small{label: label, stereotype: stereotype, symbolContext: symbolContext}
}

// AsBig returns the UML 2 component container, as upstream.
func (usymbolComponent1) AsBig(title klimt.TextBlock, labelAlignment geom.HorizontalAlignment,
	stereotype klimt.TextBlock, width, height float64, symbolContext *klimt.Fashion,
	stereoAlignment geom.HorizontalAlignment) klimt.TextBlock {
	return Component2.AsBig(title, labelAlignment, stereotype, width, height, symbolContext, stereoAlignment)
}

// usymbolComponent2 is the UML 2 component symbol: a rectangle with the
// component icon in its top right corner.
type usymbolComponent2 struct{ usymbolBase }

// SNames returns component.
func (usymbolComponent2) SNames() []klimt.SName { return []klimt.SName{klimt.SNameComponent} }

func drawComponent2(ug klimt.UGraphic, widthTotal, heightTotal, shadowing, roundCorner float64) {
	form := klimt.URectangleBuild(widthTotal, heightTotal).Rounded(roundCorner)
	form.SetDeltaShadow(shadowing)

	small := klimt.URectangleBuild(15, 10)
	tiny := klimt.URectangleBuild(4, 2)

	ug.Draw(form)

	// UML 2 Component Notation
	ug.Apply(klimt.NewUTranslate(widthTotal-20, 5)).Draw(small)
	ug.Apply(klimt.NewUTranslate(widthTotal-22, 7)).Draw(tiny)
	ug.Apply(klimt.NewUTranslate(widthTotal-22, 11)).Draw(tiny)
}

func component2Margin() margin { return newMargin(10+5, 20+5, 15+5, 5+5) }

// component2Small is the anonymous text block of upstream's asSmall.
type component2Small struct {
	klimt.TextBlockDefaults
	label, stereotype klimt.TextBlock
	symbolContext     *klimt.Fashion
}

func (b *component2Small) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = klimt.UGraphicStencilCreateDim(ug, dim)
	ug = b.symbolContext.Apply(ug)
	drawComponent2(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow(), b.symbolContext.RoundCorner())
	m := component2Margin()

	tb := klimt.TextBlockUtilsMergeTB(b.stereotype, b.label, geom.HorizontalAlignmentCenter)
	tb.DrawU(ug.Apply(klimt.NewUTranslate(m.x1, m.y1)))
}

func (b *component2Small) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimLabel := b.label.CalculateDimension(stringBounder)
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	return component2Margin().addDimension(dimStereo.MergeTB(dimLabel))
}

// AsSmall returns the component around the stereotype and the label.
func (usymbolComponent2) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	_ geom.HorizontalAlignment) klimt.TextBlock {
	return &component2Small{label: label, stereotype: stereotype, symbolContext: symbolContext}
}

// component2Big is the anonymous text block of upstream's asBig.
type component2Big struct {
	klimt.TextBlockDefaults
	title, stereotype klimt.TextBlock
	width, height     float64
	symbolContext     *klimt.Fashion
}

func (b *component2Big) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = b.symbolContext.Apply(ug)
	drawComponent2(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow(), b.symbolContext.RoundCorner())
	drawBigCentered(ug, b.title, b.stereotype, b.width, 13, 13)
}

func (b *component2Big) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(b.width, b.height)
}

// AsBig returns the component of a size with the stereotype and the title.
func (usymbolComponent2) AsBig(title klimt.TextBlock, _ geom.HorizontalAlignment, stereotype klimt.TextBlock,
	width, height float64, symbolContext *klimt.Fashion, _ geom.HorizontalAlignment) klimt.TextBlock {
	return &component2Big{title: title, stereotype: stereotype, width: width, height: height,
		symbolContext: symbolContext}
}

// usymbolArtifact is the artifact symbol: a rectangle with a document icon
// in its top right corner.
type usymbolArtifact struct{ usymbolBase }

// SNames returns artifact.
func (usymbolArtifact) SNames() []klimt.SName { return []klimt.SName{klimt.SNameArtifact} }

func drawArtifact(ug klimt.UGraphic, widthTotal, heightTotal, shadowing, roundCorner float64) {
	form := klimt.URectangleBuild(widthTotal, heightTotal).Rounded(roundCorner)
	form.SetDeltaShadow(shadowing)

	ug.Draw(form)

	polygon := klimt.NewUPolygon()
	polygon.AddPoint(0, 0)
	const heightSymbol = 14
	polygon.AddPoint(0, heightSymbol)
	const widthSymbol = 12
	polygon.AddPoint(widthSymbol, heightSymbol)
	const cornersize = 6
	polygon.AddPoint(widthSymbol, cornersize)
	polygon.AddPoint(widthSymbol-cornersize, 0)
	polygon.AddPoint(0, 0)

	xSymbol := widthTotal - widthSymbol - 5
	const ySymbol = 5

	ug.Apply(klimt.NewUTranslate(xSymbol, ySymbol)).Draw(polygon)
	ug.Apply(klimt.NewUTranslate(xSymbol+widthSymbol-cornersize, ySymbol)).Draw(klimt.ULineV(cornersize))
	ug.Apply(klimt.NewUTranslate(xSymbol+widthSymbol, ySymbol+cornersize)).Draw(klimt.ULineH(-cornersize))
}

func artifactMargin() margin { return newMargin(10, 10+10, 10+3, 10) }

// artifactSmall is the anonymous text block of upstream's asSmall.
type artifactSmall struct {
	klimt.TextBlockDefaults
	label, stereotype klimt.TextBlock
	symbolContext     *klimt.Fashion
}

func (b *artifactSmall) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = klimt.UGraphicStencilCreateDim(ug, dim)
	ug = b.symbolContext.Apply(ug)
	drawArtifact(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow(), b.symbolContext.RoundCorner())
	m := artifactMargin()
	tb := klimt.TextBlockUtilsMergeTB(b.stereotype, b.label, geom.HorizontalAlignmentCenter)
	tb.DrawU(ug.Apply(klimt.NewUTranslate(m.x1, m.y1)))
}

func (b *artifactSmall) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimLabel := b.label.CalculateDimension(stringBounder)
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	return artifactMargin().addDimension(dimStereo.MergeTB(dimLabel))
}

// AsSmall returns the artifact around the stereotype and the label.
func (usymbolArtifact) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	_ geom.HorizontalAlignment) klimt.TextBlock {
	return &artifactSmall{label: label, stereotype: stereotype, symbolContext: symbolContext}
}

// artifactBig is the anonymous text block of upstream's asBig.
type artifactBig struct {
	klimt.TextBlockDefaults
	title, stereotype klimt.TextBlock
	width, height     float64
	symbolContext     *klimt.Fashion
}

func (b *artifactBig) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = b.symbolContext.Apply(ug)
	drawArtifact(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow(), b.symbolContext.RoundCorner())
	drawBigCentered(ug, b.title, b.stereotype, b.width, 2, 2)
}

func (b *artifactBig) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(b.width, b.height)
}

// AsBig returns the artifact of a size with the stereotype and the title.
func (usymbolArtifact) AsBig(title klimt.TextBlock, _ geom.HorizontalAlignment, stereotype klimt.TextBlock,
	width, height float64, symbolContext *klimt.Fashion, _ geom.HorizontalAlignment) klimt.TextBlock {
	return &artifactBig{title: title, stereotype: stereotype, width: width, height: height,
		symbolContext: symbolContext}
}
