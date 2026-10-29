// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/TimeConstraint.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// timeConstraint is a double arrow between two times in the lane of a
// player, with its label above it.
type timeConstraint struct {
	tick1        *TimeTick
	tick2        *TimeTick
	label        *klimt.Display
	skinParam    klimt.ISkinParam
	styleBuilder *klimt.StyleBuilder
	config       *klimt.ArrowConfiguration
	marginx      float64
}

// newTimeConstraint builds a constraint; label is the label, hasLabel
// false for none. Nil ticks throw NullPointerException, as upstream's
// requireNonNull.
func newTimeConstraint(marginx float64, tick1, tick2 *TimeTick, label string, hasLabel bool,
	skinParam klimt.ISkinParam, config *klimt.ArrowConfiguration) *timeConstraint {
	if tick1 == nil || tick2 == nil {
		uml.Throw(uml.NullPointerException, "")
	}
	return &timeConstraint{marginx: marginx, tick1: tick1, tick2: tick2,
		label: klimt.DisplayWithNewlinesString(skinParam.Pragma(), label, hasLabel), skinParam: skinParam,
		styleBuilder: skinParam.CurrentStyleBuilder(), config: config}
}

// containsStrict reports whether other is strictly between the two times.
func (c *timeConstraint) containsStrict(other *TimeTick) bool {
	return c.tick1.CompareTo(other) < 0 && c.tick2.CompareTo(other) > 0
}

func (c *timeConstraint) textBlock(display *klimt.Display) klimt.TextBlock {
	return display.Create(c.fontConfiguration(), geom.HorizontalAlignmentLeft, c.skinParam)
}

func (c *timeConstraint) fontConfiguration() *klimt.FontConfiguration {
	return c.style().FontConfiguration(c.skinParam.IHtmlColorSet(), nil)
}

// drawU draws the line with its heads outward when the times are more than
// 20 pixels apart, inward otherwise, and the label centered above it.
func (c *timeConstraint) drawU(ug klimt.UGraphic, ruler *TimingRuler) {
	arrowColor := c.arrowColor()
	ug = ug.Apply(arrowColor).Apply(klimt.Bg(arrowColor))
	x1 := ruler.getPosInPixel(c.tick1) + c.marginx
	x2 := ruler.getPosInPixel(c.tick2) - c.marginx
	if x2-x1 > 20 {
		c.drawLine(ug, x1+3, x2-3)
		ug.Apply(klimt.UStrokeSimple()).Draw(polygonOf(klimt.DirectionLeft, geom.NewXPoint2D(x1, 0)))
		ug.Apply(klimt.UStrokeSimple()).Draw(polygonOf(klimt.DirectionRight, geom.NewXPoint2D(x2, 0)))
	} else {
		c.drawLine(ug, x1-1, x2+1)
		ug.Apply(klimt.UStrokeSimple()).Draw(polygonOf(klimt.DirectionRight, geom.NewXPoint2D(x1, 0)))
		ug.Apply(klimt.UStrokeSimple()).Draw(polygonOf(klimt.DirectionLeft, geom.NewXPoint2D(x2, 0)))
	}
	text := c.textBlock(c.label)
	dimText := text.CalculateDimension(ug.StringBounder())
	x := x1 + (x2-x1-dimText.Width())/2
	text.DrawU(ug.Apply(klimt.NewUTranslate(x, -c.constraintHeight(ug.StringBounder()))))
}

func (c *timeConstraint) drawLine(ug klimt.UGraphic, x1, x2 float64) {
	ug.Apply(klimt.UTranslateDx(x1)).Apply(c.uStroke()).Draw(klimt.ULineH(x2 - x1))
}

func (c *timeConstraint) arrowColor() color.HColor {
	if configColor := c.config.Color(); configColor != nil {
		return configColor
	}
	return c.style().Value(klimt.PNameLineColor).AsColor(c.skinParam.IHtmlColorSet())
}

func (c *timeConstraint) style() *klimt.Style {
	return timingStyle(klimt.SNameConstraintArrow).MergedStyle(c.styleBuilder)
}

func (c *timeConstraint) uStroke() klimt.UStroke { return c.style().Stroke() }

// constraintHeight is the height of the label and of the top margin.
func (c *timeConstraint) constraintHeight(stringBounder font.StringBounder) float64 {
	text := c.textBlock(c.label)
	dimText := text.CalculateDimension(stringBounder)
	return dimText.Height() + constraintTopMargin
}

const constraintTopMargin = 5

// polygonOf is a head of 8 by 4 pixels at end, pointing to dir.
func polygonOf(dir klimt.Direction, end geom.XPoint2D) *klimt.UPolygon {
	const dx = 8.0
	const dy = 4.0
	var pt1, pt2 geom.XPoint2D
	if dir == klimt.DirectionRight {
		pt1 = end.Move(-dx, dy)
		pt2 = end.Move(-dx, -dy)
	} else {
		pt1 = end.Move(dx, dy)
		pt2 = end.Move(dx, -dy)
	}
	polygon := klimt.NewUPolygon()
	polygon.AddPointPt(pt1)
	polygon.AddPointPt(pt2)
	polygon.AddPointPt(end)
	return polygon
}
