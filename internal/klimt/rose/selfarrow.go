// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/rose/ComponentRoseSelfArrow.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package rose

import (
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// The size of a self message loop.
const (
	selfArrowWidth  = 45
	selfArrowXRight = selfArrowWidth - 3
)

// ComponentRoseSelfArrow is the arrow of a message from a participant to
// itself: a loop on the right of the lifeline, or on its left when the
// message is defined reversed.
type ComponentRoseSelfArrow struct {
	AbstractComponentRoseArrow
	niceArrow bool
}

var _ klimt.ArrowComponent = (*ComponentRoseSelfArrow)(nil)

// NewComponentRoseSelfArrow returns the arrow of a self message.
func NewComponentRoseSelfArrow(style *klimt.Style, stringsToDisplay *klimt.Display,
	arrowConfiguration *klimt.ArrowConfiguration, skinParam klimt.ISkinParam, maxMessageSize klimt.LineBreakStrategy,
	niceArrow bool) *ComponentRoseSelfArrow {
	c := &ComponentRoseSelfArrow{}
	c.initAbstractComponentRoseArrow(c, style, stringsToDisplay, arrowConfiguration, skinParam, maxMessageSize)
	c.niceArrow = niceArrow
	return c
}

// DrawInternalU draws the loop, its heads and decorations, and the label.
func (c *ComponentRoseSelfArrow) DrawInternalU(ug klimt.UGraphic, area *klimt.Area) {
	if c.ArrowConfiguration().IsHidden() {
		return
	}
	stringBounder := ug.StringBounder()
	textHeight := c.TextHeight(stringBounder)
	prefTextWidth := c.PreferredWidth(stringBounder)

	arrowHeight := c.arrowOnlyHeight(stringBounder)

	ug = ug.Apply(c.ForegroundColor())
	ug2 := c.ArrowConfiguration().ApplyStrokeLegacy(ug)

	if c.ArrowConfiguration().IsReverseDefine() {
		c.drawLeftSide(ug, ug2, area, arrowHeight, textHeight, prefTextWidth)
	} else {
		c.drawRightSide(ug, ug2, area, arrowHeight, textHeight)
	}

	c.TextBlock().DrawU(ug.Apply(klimt.UTranslateDx(c.OldPaddingX1())))
}

func (c *ComponentRoseSelfArrow) drawRightSide(ug, ug2 klimt.UGraphic, area *klimt.Area, arrowHeight,
	textHeight float64) {
	config := c.ArrowConfiguration()
	x1 := 0.0
	if area.DeltaX1() < 0 {
		x1 = area.DeltaX1()
	}
	x2 := 1.0
	if area.DeltaX1() > 0 {
		x2 = -area.DeltaX1()
	}
	circle := klimt.UEllipseBuild(DiamCircle, DiamCircle)
	if config.Decoration1() == klimt.ArrowDecorationCircle {
		ug2.Apply(klimt.UStrokeWithThickness(ThinCircle)).Apply(c.ForegroundColor()).
			Apply(klimt.Bg(c.BackgroundColor())).
			Apply(klimt.NewUTranslate(x1+1-DiamCircle/2-ThinCircle, textHeight-DiamCircle/2-ThinCircle/2)).
			Draw(circle)
		x1 += DiamCircle/2 + ThinCircle + 1
		if config.Dressing1().Head() == klimt.ArrowHeadNone {
			x1 -= ThinCircle + 1
		} else {
			x1 -= 0
		}
	}
	if config.Decoration2() == klimt.ArrowDecorationCircle {
		ug2.Apply(klimt.UStrokeWithThickness(ThinCircle)).Apply(c.ForegroundColor()).
			Apply(klimt.Bg(c.BackgroundColor())).
			Apply(klimt.NewUTranslate(x2-DiamCircle/2-ThinCircle,
				textHeight+arrowHeight-DiamCircle/2-ThinCircle/2)).
			Draw(circle)
		x2 += DiamCircle/2 + ThinCircle
	}
	hasStartingCrossX := config.Dressing1().Head() == klimt.ArrowHeadCrossx
	if hasStartingCrossX {
		x1 += 2 * SpaceCrossX
	}

	hasFinalCrossX := config.Dressing2().Head() == klimt.ArrowHeadCrossx
	if hasFinalCrossX {
		x2 += 2 * SpaceCrossX
	}

	if config.Dressing2().Head() == klimt.ArrowHeadAsync {
		x2 -= 1
	}

	ug2.Apply(klimt.NewUTranslate(x1, textHeight)).Draw(klimt.ULineH(selfArrowXRight - x1))
	ug2.Apply(klimt.NewUTranslate(selfArrowXRight, textHeight)).Draw(klimt.ULineV(arrowHeight))
	ug2.Apply(klimt.NewUTranslate(x2, textHeight+arrowHeight)).Draw(klimt.ULineH(selfArrowXRight - x2))

	if config.Dressing2().Head() == klimt.ArrowHeadAsync {
		x2 += 1
	}

	switch {
	case hasStartingCrossX:
		ug.Apply(klimt.UStrokeWithThickness(2)).
			Apply(klimt.NewUTranslate(SpaceCrossX, textHeight-arrowDeltaX/2.0)).
			Draw(klimt.NewULine(arrowDeltaX, arrowDeltaX))
		ug.Apply(klimt.UStrokeWithThickness(2)).
			Apply(klimt.NewUTranslate(SpaceCrossX, textHeight+arrowDeltaX/2.0)).
			Draw(klimt.NewULine(arrowDeltaX, -arrowDeltaX))
	case config.IsAsync1():
		if config.Part() != klimt.ArrowPartBottomPart {
			config.ApplyThicknessOnly(ug).Apply(klimt.NewUTranslate(x1, textHeight)).
				Draw(klimt.NewULine(arrowDeltaX, arrowDeltaY))
		}
		if config.Part() != klimt.ArrowPartTopPart {
			config.ApplyThicknessOnly(ug).Apply(klimt.NewUTranslate(x1, textHeight)).
				Draw(klimt.NewULine(arrowDeltaX, -arrowDeltaY))
		}
	case config.Dressing1().Head() == klimt.ArrowHeadNormal:
		polygon := c.polygon().Translate(0, textHeight)
		ug.Apply(klimt.Bg(c.ForegroundColor())).Apply(klimt.UTranslateDx(x1)).Draw(polygon)
	}

	switch {
	case hasFinalCrossX:
		ug.Apply(klimt.UStrokeWithThickness(2)).
			Apply(klimt.NewUTranslate(SpaceCrossX, textHeight-arrowDeltaX/2.0+arrowHeight)).
			Draw(klimt.NewULine(arrowDeltaX, arrowDeltaX))
		ug.Apply(klimt.UStrokeWithThickness(2)).
			Apply(klimt.NewUTranslate(SpaceCrossX, textHeight+arrowDeltaX/2.0+arrowHeight)).
			Draw(klimt.NewULine(arrowDeltaX, -arrowDeltaX))
	case config.IsAsync2():
		if config.Part() != klimt.ArrowPartBottomPart {
			config.ApplyThicknessOnly(ug).Apply(klimt.NewUTranslate(x2, textHeight+arrowHeight)).
				Draw(klimt.NewULine(arrowDeltaX, -arrowDeltaY))
		}
		if config.Part() != klimt.ArrowPartTopPart {
			config.ApplyThicknessOnly(ug).Apply(klimt.NewUTranslate(x2, textHeight+arrowHeight)).
				Draw(klimt.NewULine(arrowDeltaX, arrowDeltaY))
		}
	case config.Dressing2().Head() == klimt.ArrowHeadNormal:
		polygon := c.polygon().Translate(0, textHeight+arrowHeight)
		ug.Apply(klimt.Bg(c.ForegroundColor())).Apply(klimt.UTranslateDx(x2)).Draw(polygon)
	}
}

func (c *ComponentRoseSelfArrow) drawLeftSide(ug, ug2 klimt.UGraphic, area *klimt.Area, arrowHeight, textHeight,
	prefTextWidth float64) {
	config := c.ArrowConfiguration()
	deltaSize := area.LiveDeltaSize()
	dx := area.DeltaX1()
	level := area.Level()

	x1 := 0.0
	x2 := 1.0
	extraLiveDeltaIndent := float64(float64(level) * deltaSize)
	switch {
	case dx < 0:
		x1 += 0
		if level > 0 {
			x2 += -extraLiveDeltaIndent
		} else {
			x2 += deltaSize
		}
	case dx > 0:
		if level > 1 {
			x1 += deltaSize - extraLiveDeltaIndent
		} else {
			x1 += 0
		}
		if level == 1 {
			x2 += -deltaSize
		} else {
			x2 += 0
		}
	case level > 1:
		x1 -= extraLiveDeltaIndent - deltaSize
		x2 -= extraLiveDeltaIndent - deltaSize
	}

	circle := klimt.UEllipseBuild(DiamCircle, DiamCircle)
	if config.Decoration1() == klimt.ArrowDecorationCircle {
		ug2.Apply(klimt.UStrokeWithThickness(ThinCircle)).Apply(c.ForegroundColor()).
			Apply(klimt.Bg(c.BackgroundColor())).
			Apply(klimt.NewUTranslate(prefTextWidth-x1+1-DiamCircle/2-ThinCircle,
				textHeight-DiamCircle/2-ThinCircle/2)).
			Draw(circle)
		x1 += DiamCircle/2.0 - ThinCircle
		if config.Dressing1().Head() == klimt.ArrowHeadNone {
			x1 += ThinCircle
		} else {
			x1 += 0
		}
	}
	if config.Decoration2() == klimt.ArrowDecorationCircle {
		ug2.Apply(klimt.UStrokeWithThickness(ThinCircle)).Apply(c.ForegroundColor()).
			Apply(klimt.Bg(c.BackgroundColor())).
			Apply(klimt.NewUTranslate(prefTextWidth-x2+2-DiamCircle/2-ThinCircle,
				textHeight+arrowHeight-DiamCircle/2-ThinCircle/2)).
			Draw(circle)
		x2 += DiamCircle/2.0 + ThinCircle
	}
	hasStartingCrossX := config.Dressing1().Head() == klimt.ArrowHeadCrossx
	if hasStartingCrossX {
		if level > 0 {
			x1 += SpaceCrossX + extraLiveDeltaIndent
		} else {
			x1 += SpaceCrossX + SpaceCrossX
		}
	}

	hasFinalCrossX := config.Dressing2().Head() == klimt.ArrowHeadCrossx
	if hasFinalCrossX {
		if level > 0 {
			x2 += SpaceCrossX + extraLiveDeltaIndent
		} else {
			x2 += SpaceCrossX + SpaceCrossX
		}
	}

	extraline := 0.0
	if config.Dressing2().Part() == klimt.ArrowPartFull && config.Dressing2().Head() == klimt.ArrowHeadNormal {
		extraline = 1
	}

	x1 += 1

	ug2.Apply(klimt.NewUTranslate(prefTextWidth-selfArrowXRight, textHeight)).
		Draw(klimt.ULineH(selfArrowXRight - x1))
	ug2.Apply(klimt.NewUTranslate(prefTextWidth-selfArrowXRight, textHeight)).Draw(klimt.ULineV(arrowHeight))
	ug2.Apply(klimt.NewUTranslate(prefTextWidth-selfArrowXRight, textHeight+arrowHeight)).
		Draw(klimt.ULineH(selfArrowXRight - x2 - extraline))

	if config.Dressing2().Part() == klimt.ArrowPartFull && config.Dressing2().Head() == klimt.ArrowHeadNormal {
		x2 += 1
	}
	if config.Dressing2().Head() == klimt.ArrowHeadAsync {
		x2 += 1
	}

	switch {
	case hasStartingCrossX:
		ug.Apply(klimt.UStrokeWithThickness(2)).
			Apply(klimt.NewUTranslate(prefTextWidth-x1-SpaceCrossX/2, textHeight-arrowDeltaX/2.0)).
			Draw(klimt.NewULine(arrowDeltaX, arrowDeltaX))
		ug.Apply(klimt.UStrokeWithThickness(2)).
			Apply(klimt.NewUTranslate(prefTextWidth-x1-SpaceCrossX/2, textHeight+arrowDeltaX/2.0)).
			Draw(klimt.NewULine(arrowDeltaX, -arrowDeltaX))
	case config.Dressing1().Head() == klimt.ArrowHeadNormal:
		polygon := c.polygon().Translate(prefTextWidth-x1, textHeight)
		ug.Apply(klimt.Bg(c.ForegroundColor())).Apply(klimt.UTranslateDx(x1)).Draw(polygon)
	}

	switch {
	case hasFinalCrossX:
		ug.Apply(klimt.UStrokeWithThickness(2)).
			Apply(klimt.NewUTranslate(prefTextWidth-x2-SpaceCrossX/2, textHeight-arrowDeltaX/2.0+arrowHeight)).
			Draw(klimt.NewULine(arrowDeltaX, arrowDeltaX))
		ug.Apply(klimt.UStrokeWithThickness(2)).
			Apply(klimt.NewUTranslate(prefTextWidth-x2-SpaceCrossX/2, textHeight+arrowDeltaX/2.0+arrowHeight)).
			Draw(klimt.NewULine(arrowDeltaX, -arrowDeltaX))
	case config.IsAsync2():
		if config.Part() != klimt.ArrowPartBottomPart {
			config.ApplyThicknessOnly(ug).Apply(klimt.NewUTranslate(prefTextWidth-x2, textHeight+arrowHeight)).
				Draw(klimt.NewULine(-arrowDeltaX, -arrowDeltaY))
		}
		if config.Part() != klimt.ArrowPartTopPart {
			config.ApplyThicknessOnly(ug).Apply(klimt.NewUTranslate(prefTextWidth-x2, textHeight+arrowHeight)).
				Draw(klimt.NewULine(-arrowDeltaX, arrowDeltaY))
		}
	case config.Dressing2().Head() == klimt.ArrowHeadNormal:
		polygon := c.polygon().Translate(0, textHeight+arrowHeight)
		ug.Apply(klimt.Bg(c.ForegroundColor())).Apply(klimt.UTranslateDx(prefTextWidth - x2)).Draw(polygon)
	}
}

func (c *ComponentRoseSelfArrow) polygon() *klimt.UPolygon {
	polygon := klimt.NewUPolygon()
	direction := 1.0
	if c.ArrowConfiguration().IsReverseDefine() {
		direction = -1
	}
	x := float64(direction * arrowDeltaX)
	switch c.ArrowConfiguration().Part() {
	case klimt.ArrowPartTopPart:
		polygon.AddPoint(x-1, -arrowDeltaY)
		polygon.AddPoint(-1, 0)
		polygon.AddPoint(x-1, 0)
	case klimt.ArrowPartBottomPart:
		polygon.AddPoint(x-1, 0)
		polygon.AddPoint(-1, 0)
		polygon.AddPoint(x-1, arrowDeltaY)
	default:
		polygon.AddPoint(x, -arrowDeltaY)
		polygon.AddPoint(0, 0)
		polygon.AddPoint(x, arrowDeltaY)
		if c.niceArrow {
			polygon.AddPoint(x-float64(direction*4), 0)
		}
	}
	return polygon
}

// StartPoint returns the point where the loop leaves the lifeline.
func (c *ComponentRoseSelfArrow) StartPoint(stringBounder font.StringBounder, _ geom.XDimension2D) geom.XPoint2D {
	textHeight := c.TextHeight(stringBounder)
	return geom.NewXPoint2D(c.PaddingX(), textHeight+c.PaddingY())
}

// EndPoint returns the point where the loop comes back to the lifeline.
func (c *ComponentRoseSelfArrow) EndPoint(stringBounder font.StringBounder, _ geom.XDimension2D) geom.XPoint2D {
	textHeight := c.TextHeight(stringBounder)
	textAndArrowHeight := textHeight + c.arrowOnlyHeight(stringBounder)
	return geom.NewXPoint2D(c.PaddingX(), textAndArrowHeight+c.PaddingY())
}

// YPoint returns the middle of the loop. As upstream, it adds the
// horizontal padding.
func (c *ComponentRoseSelfArrow) YPoint(stringBounder font.StringBounder) float64 {
	textHeight := c.TextHeight(stringBounder)
	textAndArrowHeight := textHeight + c.arrowOnlyHeight(stringBounder)
	return (textHeight+textAndArrowHeight)/2 + c.PaddingX()
}

// PreferredHeight returns the height of the label, the loop, a head and the
// padding.
func (c *ComponentRoseSelfArrow) PreferredHeight(stringBounder font.StringBounder) float64 {
	return c.TextHeight(stringBounder) + arrowDeltaY + c.arrowOnlyHeight(stringBounder) + 2*c.PaddingY()
}

func (c *ComponentRoseSelfArrow) arrowOnlyHeight(font.StringBounder) float64 { return 13 }

// PreferredWidth returns the width of the label, at least 50.
func (c *ComponentRoseSelfArrow) PreferredWidth(stringBounder font.StringBounder) float64 {
	return jmath.Max(c.TextWidth(stringBounder), selfArrowWidth+5)
}

// DrawnWidth returns how far the loop and its label reach from the
// lifeline: the loop width, or the padded label when it is longer. It is
// narrower than the preferred width when the label is short or missing.
func (c *ComponentRoseSelfArrow) DrawnWidth(stringBounder font.StringBounder) float64 {
	pureTextWidth := c.PureTextWidth(stringBounder)
	labelReach := 0.0
	if pureTextWidth != 0 {
		labelReach = c.OldPaddingX1() + pureTextWidth
	}
	return jmath.Max(selfArrowXRight, labelReach)
}

// PosArrow panics, as upstream throws UnsupportedOperationException.
func (c *ComponentRoseSelfArrow) PosArrow(font.StringBounder) float64 {
	throw(unsupportedOperation, "")
	return 0
}
