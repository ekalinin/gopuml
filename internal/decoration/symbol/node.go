// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbolNode.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package symbol

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// usymbolNode is the node symbol: a box in perspective.
type usymbolNode struct{}

// SNames returns node.
func (*usymbolNode) SNames() []klimt.SName { return []klimt.SName{klimt.SNameNode} }

// SuppHeightBecauseOfShape returns 5.
func (*usymbolNode) SuppHeightBecauseOfShape() int { return 5 }

// SuppWidthBecauseOfShape returns 60.
func (*usymbolNode) SuppWidthBecauseOfShape() int { return 60 }

func drawNode(ug klimt.UGraphic, width, height, shadowing float64) {
	shape := klimt.NewUPolygon()
	shape.AddPoint(0, 10)
	shape.AddPoint(10, 0)
	shape.AddPoint(width, 0)
	shape.AddPoint(width, height-10)
	shape.AddPoint(width-10, height)
	shape.AddPoint(0, height)
	shape.AddPoint(0, 10)

	shape.SetDeltaShadow(shadowing)

	ug.Draw(shape)

	ug.Apply(klimt.NewUTranslate(width-10, 10)).Draw(klimt.NewULine(10, -10))

	ug.Apply(klimt.UTranslateDy(10)).Draw(klimt.ULineH(width - 10))
	ug.Apply(klimt.NewUTranslate(width-10, 10)).Draw(klimt.ULineV(height - 10))

	ug.Apply(klimt.NewUTranslate(0, height)).Draw(klimt.NewUEmpty(10, 10))
}

// ugraphicNode is upstream's MyUGraphicNode: separators of the text follow
// the top face of the node.
type ugraphicNode struct {
	klimt.AbstractUGraphicHorizontalLine
	endingX float64
}

func newUGraphicNode(ug klimt.UGraphic, endingX float64) *ugraphicNode {
	g := &ugraphicNode{AbstractUGraphicHorizontalLine: klimt.NewAbstractUGraphicHorizontalLine(ug), endingX: endingX}
	g.SetSelf(g)
	return g
}

// CopyHorizontalLine returns the graphic over ug.
func (g *ugraphicNode) CopyHorizontalLine(ug klimt.UGraphic) klimt.HorizontalLineGraphic {
	return newUGraphicNode(ug, g.endingX)
}

// DrawHline draws the separator with the perspective edge and its title.
func (g *ugraphicNode) DrawHline(ug klimt.UGraphic, line *klimt.UHorizontalLine, translate klimt.UTranslate) {
	ug = ug.Apply(translate)

	g.drawHlineInternal(ug, line)
	if line.IsDouble() {
		g.drawHlineInternal(ug.Apply(klimt.UTranslateDy(2)), line)
	}
	line.DrawTitleInternal(ug, 0, g.endingX-10, 0, true)
}

func (g *ugraphicNode) drawHlineInternal(ug klimt.UGraphic, line *klimt.UHorizontalLine) {
	ug = ug.Apply(line.Stroke()).Apply(klimt.Bg(color.None()))
	ug.Draw(klimt.ULineH(g.endingX - 10))
	ug.Apply(klimt.UTranslateDx(g.endingX - 10)).Draw(klimt.NewULine(10, -10))
}

func nodeMargin() margin { return newMargin(10+5, 20+5, 15+5, 5+5) }

// nodeSmall is the anonymous text block of upstream's asSmall.
type nodeSmall struct {
	klimt.TextBlockDefaults
	label, stereotype klimt.TextBlock
	symbolContext     *klimt.Fashion
	stereoAlignment   geom.HorizontalAlignment
}

func (b *nodeSmall) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = b.symbolContext.Apply(ug)
	drawNode(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow())
	m := nodeMargin()
	tb := klimt.TextBlockUtilsMergeTB(b.stereotype, b.label, b.stereoAlignment)
	ug2 := newUGraphicNode(ug, dim.Width())
	tb.DrawU(ug2.Apply(klimt.NewUTranslate(m.x1, m.y1)))
}

func (b *nodeSmall) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimLabel := b.label.CalculateDimension(stringBounder)
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	return nodeMargin().addDimension(dimStereo.MergeTB(dimLabel))
}

// AsSmall returns the node around the stereotype and the label.
func (*usymbolNode) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	stereoAlignment geom.HorizontalAlignment) klimt.TextBlock {
	return &nodeSmall{label: label, stereotype: stereotype, symbolContext: symbolContext,
		stereoAlignment: stereoAlignment}
}

// nodeBig is the anonymous text block of upstream's asBig.
type nodeBig struct {
	klimt.TextBlockDefaults
	title, stereotype klimt.TextBlock
	width, height     float64
	symbolContext     *klimt.Fashion
	stereoAlignment   geom.HorizontalAlignment
}

func (b *nodeBig) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = b.symbolContext.Apply(ug)
	drawNode(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow())
	ug = ug.Apply(klimt.NewUTranslate(-4, 11))

	dimStereo := b.stereotype.CalculateDimension(ug.StringBounder())
	var posStereoX float64
	const posStereoY = 2
	if b.stereoAlignment == geom.HorizontalAlignmentRight {
		posStereoX = b.width - dimStereo.Width() - nodeMargin().x1
	} else {
		posStereoX = (b.width - dimStereo.Width()) / 2
	}
	b.stereotype.DrawU(ug.Apply(klimt.NewUTranslate(posStereoX, posStereoY)))
	dimTitle := b.title.CalculateDimension(ug.StringBounder())
	posTitle := (b.width - dimTitle.Width()) / 2
	b.title.DrawU(ug.Apply(klimt.NewUTranslate(posTitle, 2+dimStereo.Height())))
}

func (b *nodeBig) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(b.width, b.height)
}

// AsBig returns the node of a size with the stereotype and the title.
func (*usymbolNode) AsBig(title klimt.TextBlock, _ geom.HorizontalAlignment, stereotype klimt.TextBlock,
	width, height float64, symbolContext *klimt.Fashion, stereoAlignment geom.HorizontalAlignment) klimt.TextBlock {
	return &nodeBig{title: title, stereotype: stereotype, width: width, height: height, symbolContext: symbolContext,
		stereoAlignment: stereoAlignment}
}
