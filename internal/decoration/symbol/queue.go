// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/decoration/symbol/USymbolQueue.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package symbol

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// queueDx is the depth of the rounded ends of the queue.
const queueDx = 5

// usymbolQueue is the queue symbol: a horizontal cylinder.
type usymbolQueue struct{ usymbolBase }

// SNames returns queue.
func (usymbolQueue) SNames() []klimt.SName { return []klimt.SName{klimt.SNameQueue} }

func drawQueue(ug klimt.UGraphic, width, height, shadowing float64) {
	shape := klimt.UPathNone()
	shape.SetDeltaShadow(shadowing)

	shape.MoveTo(queueDx, 0)
	shape.LineTo(width-queueDx, 0)
	shape.CubicTo(width, 0, width, height/2, width, height/2)
	shape.CubicTo(width, height/2, width, height, width-queueDx, height)
	shape.LineTo(queueDx, height)

	shape.CubicTo(0, height, 0, height/2, 0, height/2)
	shape.CubicTo(0, height/2, 0, 0, queueDx, 0)

	ug.Draw(shape)

	closing := queueClosingPath(width, height)
	ug.Apply(klimt.Bg(color.None())).Draw(closing)
}

func queueClosingPath(width, height float64) *klimt.UPath {
	closing := klimt.UPathNone()
	closing.MoveTo(width-queueDx, 0)
	closing.CubicTo(width-queueDx*2, 0, width-queueDx*2, height/2, width-queueDx*2, height/2)
	closing.CubicTo(width-queueDx*2, height, width-queueDx, height, width-queueDx, height)
	return closing
}

// ugraphicQueue is upstream's MyUGraphicQueue: separators of the text stop
// at the closing curve of the queue.
type ugraphicQueue struct {
	klimt.AbstractUGraphicHorizontalLine
	x1, x2, fullHeight float64
}

func newUGraphicQueue(ug klimt.UGraphic, x1, x2, fullHeight float64) *ugraphicQueue {
	g := &ugraphicQueue{AbstractUGraphicHorizontalLine: klimt.NewAbstractUGraphicHorizontalLine(ug), x1: x1, x2: x2,
		fullHeight: fullHeight}
	g.SetSelf(g)
	return g
}

// CopyHorizontalLine returns the graphic over ug.
func (g *ugraphicQueue) CopyHorizontalLine(ug klimt.UGraphic) klimt.HorizontalLineGraphic {
	return newUGraphicQueue(ug, g.x1, g.x2, g.fullHeight)
}

// DrawHline draws the separator up to the closing curve.
func (g *ugraphicQueue) DrawHline(ug klimt.UGraphic, line *klimt.UHorizontalLine, translate klimt.UTranslate) {
	line.DrawLineInternal(ug, g, translate.Dy(), line.Stroke())
}

// StartingX returns 0.
func (g *ugraphicQueue) StartingX(font.StringBounder, float64) float64 { return 0 }

// EndingX returns the position of the closing curve at y.
func (g *ugraphicQueue) EndingX(_ font.StringBounder, y float64) float64 {
	var factor2 float64
	halfHeight := g.fullHeight / 2
	if y <= halfHeight {
		factor2 = 1 - (y / halfHeight)
	} else {
		factor2 = (y - halfHeight) / halfHeight
	}
	return float64((g.x2-g.x1)*factor2) + g.x1
}

func queueMargin() margin { return newMargin(5, 15, 5, 5) }

// queueSmall is the anonymous text block of upstream's asSmall.
type queueSmall struct {
	klimt.TextBlockDefaults
	label, stereotype klimt.TextBlock
	symbolContext     *klimt.Fashion
}

func (b *queueSmall) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = b.symbolContext.Apply(ug)
	drawQueue(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow())
	m := queueMargin()
	tb := klimt.TextBlockUtilsMergeTB(b.stereotype, b.label, geom.HorizontalAlignmentCenter)
	ug2 := newUGraphicQueue(ug, dim.Width()-2*queueDx, dim.Width()-queueDx, dim.Height())
	tb.DrawU(ug2.Apply(klimt.NewUTranslate(m.x1, m.y1)))
}

func (b *queueSmall) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimLabel := b.label.CalculateDimension(stringBounder)
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	return queueMargin().addDimension(dimStereo.MergeTB(dimLabel))
}

// AsSmall returns the queue around the stereotype and the label.
func (usymbolQueue) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	_ geom.HorizontalAlignment) klimt.TextBlock {
	return &queueSmall{label: label, stereotype: stereotype, symbolContext: symbolContext}
}

// queueBig is the anonymous text block of upstream's asBig.
type queueBig struct {
	klimt.TextBlockDefaults
	title, stereotype klimt.TextBlock
	width, height     float64
	symbolContext     *klimt.Fashion
}

func (b *queueBig) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = b.symbolContext.Apply(ug)
	drawQueue(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow())
	dimStereo := b.stereotype.CalculateDimension(ug.StringBounder())
	posStereo := (b.width - dimStereo.Width()) / 2
	b.stereotype.DrawU(ug.Apply(klimt.NewUTranslate(posStereo, 2)))

	dimTitle := b.title.CalculateDimension(ug.StringBounder())
	posTitle := (b.width - dimTitle.Width()) / 2
	b.title.DrawU(ug.Apply(klimt.NewUTranslate(posTitle, 2+dimStereo.Height())))
}

func (b *queueBig) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(b.width, b.height)
}

// AsBig returns the queue of a size with the stereotype and the title.
func (usymbolQueue) AsBig(title klimt.TextBlock, _ geom.HorizontalAlignment, stereotype klimt.TextBlock,
	width, height float64, symbolContext *klimt.Fashion, _ geom.HorizontalAlignment) klimt.TextBlock {
	return &queueBig{title: title, stereotype: stereotype, width: width, height: height,
		symbolContext: symbolContext}
}
