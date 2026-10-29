// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/graphic/TimeArrow.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/graphic/IntricatedPoint.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"math"

	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// intricatedPoint is the point of a lane at a time: the two ends of a
// vertical segment, or the same point twice.
type intricatedPoint struct {
	pta geom.XPoint2D
	ptb geom.XPoint2D
}

func (p *intricatedPoint) translated(translate klimt.UTranslate) *intricatedPoint {
	return &intricatedPoint{translate.Translated(p.pta), translate.Translated(p.ptb)}
}

// timeArrow is the arrow of a message, from start to end, with its label.
type timeArrow struct {
	start        geom.XPoint2D
	end          geom.XPoint2D
	label        *klimt.Display
	skinParam    klimt.ISkinParam
	styleBuilder *klimt.StyleBuilder
	message      *timeMessage
}

// newTimeArrow is upstream's TimeArrow.create: the shortest of the four
// arrows between the ends of the two points, the later one on a tie.
func newTimeArrow(pt1, pt2 *intricatedPoint, label *klimt.Display, skinParam klimt.ISkinParam,
	message *timeMessage) *timeArrow {
	arrow1 := newTimeArrowOf(pt1.pta, pt2.pta, label, skinParam, message)
	arrow2 := newTimeArrowOf(pt1.pta, pt2.ptb, label, skinParam, message)
	arrow3 := newTimeArrowOf(pt1.ptb, pt2.pta, label, skinParam, message)
	arrow4 := newTimeArrowOf(pt1.ptb, pt2.ptb, label, skinParam, message)
	return shorter(shorter(arrow1, arrow2), shorter(arrow3, arrow4))
}

func newTimeArrowOf(start, end geom.XPoint2D, label *klimt.Display, skinParam klimt.ISkinParam,
	message *timeMessage) *timeArrow {
	return &timeArrow{start: start, message: message, end: end, label: label, skinParam: skinParam,
		styleBuilder: skinParam.CurrentStyleBuilder()}
}

func (a *timeArrow) angle() float64 { return jmath.Atan2(a.end.X-a.start.X, a.end.Y-a.start.Y) }

func shorter(arrow1, arrow2 *timeArrow) *timeArrow {
	if arrow1.length() < arrow2.length() {
		return arrow1
	}
	return arrow2
}

func (a *timeArrow) length() float64 { return a.start.Distance(a.end) }

// onCircle is the point 8 pixels from pt in the direction alpha.
func onCircle(pt geom.XPoint2D, alpha float64) geom.XPoint2D {
	const radius = 8
	x := pt.X - float64(jmath.Sin(alpha)*radius)
	y := pt.Y - float64(jmath.Cos(alpha)*radius)
	return geom.NewXPoint2D(x, y)
}

func (a *timeArrow) fontConfiguration() *klimt.FontConfiguration {
	return a.style().FontConfiguration(a.skinParam.IHtmlColorSet(), nil)
}

// DrawU draws the line, the head and the label at the middle of the base
// of the head, above it when the arrow goes down.
func (a *timeArrow) DrawU(ug klimt.UGraphic) {
	angle := a.angle()
	ug = ug.Apply(a.message.SpecificColor()).Apply(a.message.uStroke())
	line := klimt.NewULine(a.end.X-a.start.X, a.end.Y-a.start.Y)
	ug.Apply(klimt.UTranslatePoint(a.start)).Draw(line)
	// Java rounds after each operation, where a Go constant would not.
	pi := math.Pi
	delta := float64(20.0*pi) / 180.0
	pt1 := onCircle(a.end, angle+delta)
	pt2 := onCircle(a.end, angle-delta)
	polygon := klimt.NewUPolygon()
	polygon.AddPoint(pt1.X, pt1.Y)
	polygon.AddPoint(pt2.X, pt2.Y)
	polygon.AddPoint(a.end.X, a.end.Y)
	ug = ug.Apply(klimt.Bg(a.message.SpecificColor()))
	ug.Draw(polygon)
	textLabel := a.textBlock(a.label)
	xText := (pt1.X + pt2.X) / 2
	yText := (pt1.Y + pt2.Y) / 2
	if a.start.Y < a.end.Y {
		dimLabel := textLabel.CalculateDimension(ug.StringBounder())
		yText -= dimLabel.Height()
	}
	textLabel.DrawU(ug.Apply(klimt.NewUTranslate(xText, yText)))
}

func (a *timeArrow) textBlock(display *klimt.Display) klimt.TextBlock {
	return display.Create(a.fontConfiguration(), geom.HorizontalAlignmentLeft, a.skinParam)
}

func (a *timeArrow) style() *klimt.Style {
	return timingStyle(klimt.SNameArrow).MergedStyle(a.styleBuilder)
}
