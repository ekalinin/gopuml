// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbolUsecase.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package symbol

import (
	"math"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// usymbolUsecase is the use case symbol: the text in the smallest
// surrounding ellipse, with a slanted line for business use cases. It has
// no container form.
type usymbolUsecase struct {
	usymbolBase
	isBusiness bool
}

// SNames returns usecase, and business for a business use case.
func (s *usymbolUsecase) SNames() []klimt.SName {
	if s.isBusiness {
		return []klimt.SName{klimt.SNameUsecase, klimt.SNameBusiness}
	}
	return []klimt.SName{klimt.SNameUsecase}
}

func usecaseSpecialBusiness(ug klimt.UGraphic, frontier *klimt.UEllipse) {
	rotatedEllipse := klimt.NewRotatedEllipse(frontier, math.Pi/4)

	// A variable, so that the product is rounded before the division, as
	// in Java.
	pi := math.Pi
	theta1 := 20.0 * pi / 180
	theta2 := rotatedEllipse.OtherTheta(theta1)

	frontier2 := frontier.Scale(0.99)
	p1 := frontier2.PointAtAngle(-theta1)
	p2 := frontier2.PointAtAngle(-theta2)
	usecaseDrawLine(ug, p1, p2)
}

func usecaseDrawLine(ug klimt.UGraphic, p1, p2 geom.XPoint2D) {
	ug = ug.Apply(klimt.UTranslatePoint(p1))
	ug.Draw(klimt.NewULine(p2.X-p1.X, p2.Y-p1.Y))
}

// usecaseSmall is the anonymous text block of upstream's asSmall.
type usecaseSmall struct {
	klimt.TextBlockDefaults
	s             *usymbolUsecase
	desc          klimt.TextBlock
	symbolContext *klimt.Fashion
}

func (b *usecaseSmall) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	ug = b.symbolContext.Apply(ug)

	ellipse := klimt.NewTextBlockInContainingEllipse(b.desc, stringBounder)

	ug2 := newUGraphicEllipse(ug, 0, 0, ellipse.UEllipse())

	ellipse.DrawU(ug2)
	if b.s.isBusiness {
		usecaseSpecialBusiness(ug, ellipse.UEllipse())
	}
}

func (b *usecaseSmall) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return klimt.NewTextBlockInContainingEllipse(b.desc, stringBounder).CalculateDimension(stringBounder)
}

// AsSmall returns the stereotype and the label in an ellipse.
func (s *usymbolUsecase) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	_ geom.HorizontalAlignment) klimt.TextBlock {
	tmp := klimt.TextBlockUtilsMergeTB(stereotype, label, geom.HorizontalAlignmentCenter)
	desc := tmp
	if s.isBusiness {
		desc = klimt.TextBlockUtilsWithMargin(tmp, 7, 0)
	}
	return &usecaseSmall{s: s, desc: desc, symbolContext: symbolContext}
}

// AsBig panics, as upstream throws UnsupportedOperationException.
func (s *usymbolUsecase) AsBig(klimt.TextBlock, geom.HorizontalAlignment, klimt.TextBlock, float64, float64,
	*klimt.Fashion, geom.HorizontalAlignment) klimt.TextBlock {
	throwUnsupported()
	return nil
}

// ugraphicEllipse is upstream's MyUGraphicEllipse: separators of the text
// stop at the ellipse.
type ugraphicEllipse struct {
	klimt.AbstractUGraphicHorizontalLine
	startingX          float64
	yTheoricalPosition float64
	ellipse            *klimt.UEllipse
}

func newUGraphicEllipse(ug klimt.UGraphic, startingX, yTheoricalPosition float64,
	ellipse *klimt.UEllipse) *ugraphicEllipse {
	g := &ugraphicEllipse{AbstractUGraphicHorizontalLine: klimt.NewAbstractUGraphicHorizontalLine(ug),
		startingX: startingX, yTheoricalPosition: yTheoricalPosition, ellipse: ellipse}
	g.SetSelf(g)
	return g
}

// CopyHorizontalLine returns the graphic over ug.
func (g *ugraphicEllipse) CopyHorizontalLine(ug klimt.UGraphic) klimt.HorizontalLineGraphic {
	return newUGraphicEllipse(ug, g.startingX, g.yTheoricalPosition, g.ellipse)
}

func (g *ugraphicEllipse) normalized(y float64) float64 {
	if y < g.yTheoricalPosition {
		panic(illegalArgument())
	}
	y = y - g.yTheoricalPosition
	if y > g.ellipse.Height() {
		panic(illegalArgument())
	}
	return y
}

func (g *ugraphicEllipse) startingXInternal(y float64) float64 {
	return g.startingX + g.ellipse.StartingX(g.normalized(y))
}

func (g *ugraphicEllipse) endingXInternal(y float64) float64 {
	return g.startingX + g.ellipse.EndingX(g.normalized(y))
}

// ellipseStencil is the anonymous stencil of upstream's getStencil2.
type ellipseStencil struct {
	g  *ugraphicEllipse
	dy float64
}

func (s ellipseStencil) StartingX(_ font.StringBounder, y float64) float64 {
	return s.g.startingXInternal(y + s.dy)
}

func (s ellipseStencil) EndingX(_ font.StringBounder, y float64) float64 {
	return s.g.endingXInternal(y + s.dy)
}

// DrawHline draws the separator inside the ellipse with a 1.5 stroke.
func (g *ugraphicEllipse) DrawHline(ug klimt.UGraphic, line *klimt.UHorizontalLine, translate klimt.UTranslate) {
	stroke := klimt.UStrokeWithThickness(1.5)
	line.DrawLineInternal(ug.Apply(translate), ellipseStencil{g: g, dy: translate.Dy()}, 0, stroke)
}
