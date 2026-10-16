// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/rose/AbstractComponentRoseArrow.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/rose/ComponentRoseArrow.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package rose

import (
	"github.com/ekalinin/plantuml-in-go/internal/jmath"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// The size of the arrow heads.
const (
	arrowDeltaX = 10
	arrowDeltaY = 4
)

// AbstractComponentRoseArrow is the base of the message arrows: a label
// padded as the sequence diagram padder says, the colors of the style and
// the arrow configuration with the thickness of the style.
type AbstractComponentRoseArrow struct {
	klimt.AbstractTextualComponent
	foregroundColor    color.HColor
	backgroundColor    color.HColor
	arrowConfiguration *klimt.ArrowConfiguration
}

func (c *AbstractComponentRoseArrow) initAbstractComponentRoseArrow(self klimt.AbstractComponentHooks,
	style *klimt.Style, stringsToDisplay *klimt.Display, arrowConfiguration *klimt.ArrowConfiguration,
	skinParam klimt.ISkinParam, maxMessageSize klimt.LineBreakStrategy) {
	c.InitAbstractTextualComponent(self, style, style, maxMessageSize, klimt.ClockwiseTopRightBottomLeftOf(1, 7, 1, 7),
		skinParam, stringsToDisplay, false)
	c.foregroundColor = c.ColorLine()
	c.backgroundColor = c.ColorBackGround()
	stroke := c.Stroke()
	c.arrowConfiguration = arrowConfiguration.WithThickness(stroke.Thickness())
}

// TextBlock returns the label with the padding of the sequence diagram
// padder. Upstream overrides getTextBlock, so the text sizes below use it.
func (c *AbstractComponentRoseArrow) TextBlock() klimt.TextBlock {
	padder := c.SkinParam().SequenceDiagramPadder()
	return padder.Apply(c.AbstractTextualComponent.TextBlock())
}

// PureTextWidth returns the width of the padded label.
func (c *AbstractComponentRoseArrow) PureTextWidth(stringBounder font.StringBounder) float64 {
	return c.TextBlock().CalculateDimension(stringBounder).Width()
}

// TextWidth returns the width of the padded label with the horizontal
// padding.
func (c *AbstractComponentRoseArrow) TextWidth(stringBounder font.StringBounder) float64 {
	return c.PureTextWidth(stringBounder) + c.OldPaddingX1() + c.OldPaddingX2()
}

// TextHeight returns the height of the padded label with the vertical
// padding.
func (c *AbstractComponentRoseArrow) TextHeight(stringBounder font.StringBounder) float64 {
	size := c.TextBlock().CalculateDimension(stringBounder)
	return size.Height() + c.Padding().Top() + c.Padding().Bottom()
}

// ForegroundColor returns the line color of the style.
func (c *AbstractComponentRoseArrow) ForegroundColor() color.HColor { return c.foregroundColor }

// BackgroundColor returns the background color of the style.
func (c *AbstractComponentRoseArrow) BackgroundColor() color.HColor { return c.backgroundColor }

// ArrowDeltaX returns the length of the arrow heads.
func (c *AbstractComponentRoseArrow) ArrowDeltaX() int { return arrowDeltaX }

// ArrowDeltaY returns the half width of the arrow heads.
func (c *AbstractComponentRoseArrow) ArrowDeltaY() int { return arrowDeltaY }

// PaddingY returns 4.
func (c *AbstractComponentRoseArrow) PaddingY() float64 { return 4 }

// ArrowConfiguration returns the arrow configuration with the thickness of
// the style.
func (c *AbstractComponentRoseArrow) ArrowConfiguration() *klimt.ArrowConfiguration {
	return c.arrowConfiguration
}

// The sizes of the arrow decorations.
const (
	// SpaceCrossX is the space before the cross of a lost message.
	SpaceCrossX = 6
	// DiamCircle is the diameter of the circle decoration.
	DiamCircle = 8
	// ThinCircle is the thickness of the circle decoration.
	ThinCircle = 1.5
)

// ComponentRoseArrow is the arrow of a message between two participants.
type ComponentRoseArrow struct {
	AbstractComponentRoseArrow
	messagePosition  geom.HorizontalAlignment
	niceArrow        bool
	belowForResponse bool
	inclination1     int
	inclination2     int
}

var _ klimt.ArrowComponent = (*ComponentRoseArrow)(nil)

// NewComponentRoseArrow returns the arrow of a message; messagePosition is
// the alignment of the label (HorizontalAlignmentNull for the default).
func NewComponentRoseArrow(style *klimt.Style, stringsToDisplay *klimt.Display,
	arrowConfiguration *klimt.ArrowConfiguration, messagePosition geom.HorizontalAlignment,
	skinParam klimt.ISkinParam, maxMessageSize klimt.LineBreakStrategy, niceArrow, belowForResponse bool,
) *ComponentRoseArrow {
	c := &ComponentRoseArrow{}
	c.initAbstractComponentRoseArrow(c, style, stringsToDisplay, arrowConfiguration, skinParam, maxMessageSize)
	c.messagePosition = messagePosition
	c.niceArrow = niceArrow
	c.belowForResponse = belowForResponse
	c.inclination1 = arrowConfiguration.Inclination1()
	c.inclination2 = arrowConfiguration.Inclination2()
	return c
}

// DrawInternalU draws the line, the heads and decorations, and the label.
func (c *ComponentRoseArrow) DrawInternalU(ug klimt.UGraphic, area *klimt.Area) {
	arrowConfiguration := c.ArrowConfiguration()
	if arrowConfiguration.IsHidden() {
		return
	}

	dimensionToUse := area.DimensionToUse()
	stringBounder := ug.StringBounder()
	ug = ug.Apply(c.ForegroundColor())

	dressing1 := arrowConfiguration.Dressing1()
	dressing2 := arrowConfiguration.Dressing2()

	start := 0.0
	length := dimensionToUse.Width() - 1
	lenFull := dimensionToUse.Width()

	pos1 := start + 1
	pos2 := length - 1

	if arrowConfiguration.Decoration2() == klimt.ArrowDecorationCircle {
		if dressing2.Head() == klimt.ArrowHeadNone {
			length -= DiamCircle / 2
		}
		if dressing2.Head() != klimt.ArrowHeadNone {
			length -= DiamCircle/2 + ThinCircle
		}
	}

	if arrowConfiguration.Decoration1() == klimt.ArrowDecorationCircle {
		if dressing1.Head() == klimt.ArrowHeadNone {
			start += DiamCircle / 2
			length -= DiamCircle / 2
		}
		if dressing1.Head() == klimt.ArrowHeadAsync {
			start += DiamCircle/2 + ThinCircle
			length -= DiamCircle/2 + ThinCircle
		}
		if dressing1.Head() == klimt.ArrowHeadNormal {
			start += DiamCircle/2 + ThinCircle
			length -= DiamCircle/2 + ThinCircle
		}
	}

	if dressing2.Part() == klimt.ArrowPartFull && dressing2.Head() == klimt.ArrowHeadNormal {
		length -= arrowDeltaX / 2
	}

	if dressing1.Part() == klimt.ArrowPartFull && dressing1.Head() == klimt.ArrowHeadNormal {
		start += arrowDeltaX / 2
		length -= arrowDeltaX / 2
	}

	if dressing2.Head() == klimt.ArrowHeadCrossx {
		length -= 2 * SpaceCrossX
	}

	if dressing1.Head() == klimt.ArrowHeadCrossx {
		start += 2 * SpaceCrossX
		length -= 2 * SpaceCrossX
	}

	var posArrow, yText float64
	if c.isBelowForResponse() {
		posArrow = 0
		yText = c.OldPaddingY()
	} else {
		posArrow = c.TextHeight(stringBounder)
		yText = 0
	}

	c.drawDressing1(ug.Apply(klimt.NewUTranslate(pos1, posArrow+float64(c.inclination1))), dressing1,
		arrowConfiguration.Decoration1(), lenFull)
	c.drawDressing2(ug.Apply(klimt.NewUTranslate(pos2, posArrow+float64(c.inclination2))), dressing2,
		arrowConfiguration.Decoration2(), lenFull)

	switch {
	case c.inclination1 == 0 && c.inclination2 == 0:
		arrowConfiguration.ApplyStroke(ug, c.Style()).Apply(klimt.NewUTranslate(start, posArrow)).
			Draw(klimt.NewULine(length, 0))
	case c.inclination1 != 0:
		drawLine(arrowConfiguration.ApplyStroke(ug, c.Style()), start+length, posArrow, 0,
			posArrow+float64(c.inclination1))
	case c.inclination2 != 0:
		drawLine(arrowConfiguration.ApplyStroke(ug, c.Style()), start, posArrow, pos2,
			posArrow+float64(c.inclination2))
	}

	direction2 := c.direction2()
	var textPos float64
	switch c.messagePosition {
	case geom.HorizontalAlignmentCenter:
		textWidth := c.TextBlock().CalculateDimension(stringBounder).Width()
		textPos = (dimensionToUse.Width() - jmath.Abs(area.TextDeltaX()) - textWidth) / 2
	case geom.HorizontalAlignmentRight:
		textWidth := c.TextBlock().CalculateDimension(stringBounder).Width()
		extra := 0.0
		if direction2 == klimt.ArrowDirectionLeftToRightNormal {
			extra = arrowDeltaX
		}
		textPos = dimensionToUse.Width() - jmath.Abs(area.TextDeltaX()) - textWidth - c.OldPaddingX2() - extra
	default:
		extra := 0.0
		if direction2 == klimt.ArrowDirectionRightToLeftReverse || direction2 == klimt.ArrowDirectionBothDirection {
			extra = arrowDeltaX
		}
		textPos = c.OldPaddingX1() + extra
	}
	c.TextBlock().DrawU(ug.Apply(klimt.NewUTranslate(textPos+jmath.Max(0, area.TextDeltaX()), yText)))
}

func drawLine(ug klimt.UGraphic, x1, y1, x2, y2 float64) {
	ug = ug.Apply(klimt.NewUTranslate(x1, y1))
	ug.Draw(klimt.NewULine(x2-x1, y2-y1))
}

// PosArrow returns the vertical position of the line in the label area.
func (c *ComponentRoseArrow) PosArrow(stringBounder font.StringBounder) float64 {
	if c.isBelowForResponse() {
		return 0
	}
	return c.TextHeight(stringBounder) - 2*c.OldPaddingY()
}

func (c *ComponentRoseArrow) isBelowForResponse() bool {
	return c.belowForResponse && c.ArrowConfiguration().IsReverseDefine()
}

func (c *ComponentRoseArrow) drawDressing1(ug klimt.UGraphic, dressing klimt.ArrowDressing,
	decoration klimt.ArrowDecoration, lenFull float64) {
	if decoration == klimt.ArrowDecorationCircle {
		circle := klimt.UEllipseBuild(DiamCircle, DiamCircle)
		ug.Apply(klimt.UStrokeWithThickness(ThinCircle)).Apply(c.ForegroundColor()).Apply(klimt.Bg(c.BackgroundColor())).
			Apply(klimt.NewUTranslate(-DiamCircle/2-ThinCircle, -DiamCircle/2-ThinCircle/2)).Draw(circle)
		if dressing.Head() != klimt.ArrowHeadCrossx {
			ug = ug.Apply(klimt.UTranslateDx(DiamCircle/2 + ThinCircle))
		}
	}

	switch dressing.Head() {
	case klimt.ArrowHeadAsync:
		if dressing.Part() != klimt.ArrowPartBottomPart {
			c.ArrowConfiguration().ApplyThicknessOnly(ug).Draw(
				klimt.NewULine(arrowDeltaX, -arrowDeltaY).Rotate(jmath.Atan2(float64(-c.inclination1), lenFull)))
		}
		if dressing.Part() != klimt.ArrowPartTopPart {
			c.ArrowConfiguration().ApplyThicknessOnly(ug).Draw(
				klimt.NewULine(arrowDeltaX, arrowDeltaY).Rotate(jmath.Atan2(float64(-c.inclination1), lenFull)))
		}
	case klimt.ArrowHeadCrossx:
		ug = ug.Apply(klimt.UStrokeWithThickness(2))
		ug.Apply(klimt.NewUTranslate(SpaceCrossX, -arrowDeltaX/2)).Draw(klimt.NewULine(arrowDeltaX, arrowDeltaX))
		ug.Apply(klimt.NewUTranslate(SpaceCrossX, arrowDeltaX/2)).Draw(klimt.NewULine(arrowDeltaX, -arrowDeltaX))
	case klimt.ArrowHeadNormal:
		polygon := c.polygonReverse(dressing.Part())
		if c.inclination1 != 0 {
			polygon.Rotate(jmath.Atan2(float64(-c.inclination1), lenFull))
		}
		ug.Apply(klimt.Bg(c.ForegroundColor())).Draw(polygon)
	}
}

func (c *ComponentRoseArrow) drawDressing2(ug klimt.UGraphic, dressing klimt.ArrowDressing,
	decoration klimt.ArrowDecoration, lenFull float64) {
	if decoration == klimt.ArrowDecorationCircle {
		ug = ug.Apply(klimt.UStrokeWithThickness(ThinCircle)).Apply(c.ForegroundColor()).
			Apply(klimt.Bg(c.BackgroundColor()))
		circle := klimt.UEllipseBuild(DiamCircle, DiamCircle)
		ug.Apply(klimt.NewUTranslate(-DiamCircle/2+ThinCircle, -DiamCircle/2-ThinCircle/2)).Draw(circle)
		ug = ug.Apply(klimt.UStrokeSimple())
		ug = ug.Apply(klimt.UTranslateDx(-DiamCircle/2 - ThinCircle))
	}

	switch dressing.Head() {
	case klimt.ArrowHeadAsync:
		if dressing.Part() != klimt.ArrowPartBottomPart {
			c.ArrowConfiguration().ApplyThicknessOnly(ug).Draw(
				klimt.NewULine(-arrowDeltaX, -arrowDeltaY).Rotate(jmath.Atan2(float64(c.inclination2), lenFull)))
		}
		if dressing.Part() != klimt.ArrowPartTopPart {
			c.ArrowConfiguration().ApplyThicknessOnly(ug).Draw(
				klimt.NewULine(-arrowDeltaX, arrowDeltaY).Rotate(jmath.Atan2(float64(c.inclination2), lenFull)))
		}
	case klimt.ArrowHeadCrossx:
		ug = ug.Apply(klimt.UStrokeWithThickness(2))
		ug.Apply(klimt.NewUTranslate(-SpaceCrossX-arrowDeltaX, -arrowDeltaX/2)).
			Draw(klimt.NewULine(arrowDeltaX, arrowDeltaX))
		ug.Apply(klimt.NewUTranslate(-SpaceCrossX-arrowDeltaX, arrowDeltaX/2)).
			Draw(klimt.NewULine(arrowDeltaX, -arrowDeltaX))
	case klimt.ArrowHeadNormal:
		polygon := c.polygonNormal(dressing.Part())
		if c.inclination2 != 0 {
			polygon.Rotate(jmath.Atan2(float64(c.inclination2), lenFull))
		}
		ug.Apply(klimt.Bg(c.ForegroundColor())).Draw(polygon)
	}
}

func (c *ComponentRoseArrow) polygonNormal(part klimt.ArrowPart) *klimt.UPolygon {
	polygon := klimt.NewUPolygon()
	switch part {
	case klimt.ArrowPartTopPart:
		polygon.AddPoint(-arrowDeltaX, -arrowDeltaY)
		polygon.AddPoint(0, 0)
		polygon.AddPoint(-arrowDeltaX, 0)
	case klimt.ArrowPartBottomPart:
		polygon.AddPoint(-arrowDeltaX, 0)
		polygon.AddPoint(0, 0)
		polygon.AddPoint(-arrowDeltaX, arrowDeltaY)
	default:
		polygon.AddPoint(-arrowDeltaX, -arrowDeltaY)
		polygon.AddPoint(0, 0)
		polygon.AddPoint(-arrowDeltaX, +arrowDeltaY)
		if c.niceArrow {
			polygon.AddPoint(-arrowDeltaX+4, 0)
		}
	}
	return polygon
}

func (c *ComponentRoseArrow) polygonReverse(part klimt.ArrowPart) *klimt.UPolygon {
	polygon := klimt.NewUPolygon()
	switch part {
	case klimt.ArrowPartTopPart:
		polygon.AddPoint(arrowDeltaX, -arrowDeltaY)
		polygon.AddPoint(0, 0)
		polygon.AddPoint(arrowDeltaX, 0)
	case klimt.ArrowPartBottomPart:
		polygon.AddPoint(arrowDeltaX, 0)
		polygon.AddPoint(0, 0)
		polygon.AddPoint(arrowDeltaX, arrowDeltaY)
	default:
		polygon.AddPoint(arrowDeltaX, -arrowDeltaY)
		polygon.AddPoint(0, 0)
		polygon.AddPoint(arrowDeltaX, arrowDeltaY)
		if c.niceArrow {
			polygon.AddPoint(arrowDeltaX-4, 0)
		}
	}
	return polygon
}

// StartPoint returns the point where the message leaves its participant.
func (c *ComponentRoseArrow) StartPoint(stringBounder font.StringBounder, dimensionToUse geom.XDimension2D) geom.XPoint2D {
	y := c.YPoint(stringBounder)
	if c.direction2() == klimt.ArrowDirectionLeftToRightNormal {
		return geom.NewXPoint2D(c.PaddingX(), y+float64(c.inclination2))
	}
	return geom.NewXPoint2D(dimensionToUse.Width()+c.PaddingX(), y+float64(c.inclination2))
}

// EndPoint returns the point where the message reaches its participant.
func (c *ComponentRoseArrow) EndPoint(stringBounder font.StringBounder, dimensionToUse geom.XDimension2D) geom.XPoint2D {
	y := c.YPoint(stringBounder)
	if c.direction2() == klimt.ArrowDirectionLeftToRightNormal {
		return geom.NewXPoint2D(dimensionToUse.Width()+c.PaddingX(), y)
	}
	return geom.NewXPoint2D(c.PaddingX(), y)
}

// YPoint returns the vertical position of the line.
func (c *ComponentRoseArrow) YPoint(stringBounder font.StringBounder) float64 {
	if c.isBelowForResponse() {
		return c.PaddingY()
	}
	return c.TextHeight(stringBounder) + c.PaddingY()
}

func (c *ComponentRoseArrow) direction2() klimt.ArrowDirection {
	return c.ArrowConfiguration().ArrowDirection()
}

// PreferredHeight returns the height of the label, the heads, the padding
// and the inclinations.
func (c *ComponentRoseArrow) PreferredHeight(stringBounder font.StringBounder) float64 {
	return c.TextHeight(stringBounder) + arrowDeltaY + 2*c.PaddingY() + float64(c.inclination1) +
		float64(c.inclination2)
}

// PreferredWidth returns the width of the label and a head.
func (c *ComponentRoseArrow) PreferredWidth(stringBounder font.StringBounder) float64 {
	return c.TextWidth(stringBounder) + arrowDeltaX
}
