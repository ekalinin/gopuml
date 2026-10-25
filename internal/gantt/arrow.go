// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/GanttArrow.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/GArrows.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/GanttConstraint.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"math"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// UDrawable is getUDrawable: the arrow of the constraint, with the styles
// of the time it was created, its color and its link style.
func (c *GanttConstraint) UDrawable(colorSet *color.HColorSet, timeScale TimeScale,
	toTaskDraw *TaskDrawRegistryData) klimt.UDrawable {
	style := ganttStyleQuery(klimt.SNameArrow).MergedStyle(c.styleBuilder).
		EventuallyOverrideColor(klimt.PNameLineColor, c.SpecificColor())
	defaultStroke := style.Stroke()
	stroke := c.LinkType.Stroke3(&defaultStroke)
	style = style.EventuallyOverrideStroke(&stroke)
	return newGanttArrow(colorSet, style, timeScale, c.source, c.dest, toTaskDraw)
}

// ganttArrow is upstream's GanttArrow: an arrow from the start or end of a
// task to the start or end of another one.
type ganttArrow struct {
	timeScale  TimeScale
	atStart    GSide
	source     *TaskInstant
	atEnd      GSide
	dest       *TaskInstant
	colorSet   *color.HColorSet
	style      *klimt.Style
	toTaskDraw *TaskDrawRegistryData
}

func newGanttArrow(colorSet *color.HColorSet, style *klimt.Style, timeScale TimeScale, source, dest *TaskInstant,
	toTaskDraw *TaskDrawRegistryData) *ganttArrow {
	a := &ganttArrow{timeScale: timeScale, source: source, dest: dest, colorSet: colorSet, style: style,
		toTaskDraw: toTaskDraw}
	switch {
	case source.Attribute() == TaskAttributeEnd && dest.Attribute() == TaskAttributeStart:
		a.atStart = GSideBottomRight
		if source.SameRowAs(dest) {
			a.atStart = GSideRight
		}
		a.atEnd = GSideLeft
	case source.Attribute() == TaskAttributeEnd && dest.Attribute() == TaskAttributeEnd:
		a.atStart = GSideRight
		a.atEnd = GSideRight
	case source.Attribute() == TaskAttributeStart && dest.Attribute() == TaskAttributeStart:
		a.atStart = GSideLeft
		a.atEnd = GSideLeft
	case source.Attribute() == TaskAttributeStart && dest.Attribute() == TaskAttributeEnd:
		a.atStart = GSideBottomLeft
		if source.SameRowAs(dest) {
			a.atStart = GSideLeft
		}
		a.atEnd = GSideRight
	default:
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return a
}

func (a *ganttArrow) drawOf(instant *TaskInstant) TaskDraw {
	if task, ok := instant.Moment().(Task); ok {
		return a.toTaskDraw.TaskDraw(task)
	}
	return nil
}

// DrawU implements klimt.UDrawable.
func (a *ganttArrow) DrawU(ug klimt.UGraphic) {
	ug = a.style.ApplyStrokeAndLineColor(ug, a.colorSet)
	start := a.drawOf(a.source)
	end := a.drawOf(a.dest)
	if start == nil || end == nil {
		return
	}
	sb := ug.StringBounder()
	x1 := start.X(sb, a.atStart, GArrowOutgoing)
	y1 := start.SideY(sb, a.atStart)
	x2 := end.X(sb, a.atEnd, GArrowIncoming)
	y2 := end.SideY(sb, a.atEnd)
	if a.atStart.IsBottom() && y2 < y1 {
		y1 = start.SideY(sb, a.atStart.ReverseBottomTop())
	}
	const minimalWidth = 6
	arrows := &gArrows{atEnd: a.atEnd, backColor: a.style.Value(klimt.PNameLineColor).AsColor(a.colorSet)}
	switch {
	case a.atStart.IsBottom() && a.atEnd == GSideLeft:
		if x2 > x1+minimalWidth {
			arrows.addPoint(x1, y1)
			arrows.addPoint(x1, y2)
			arrows.addPoint(x2, y2)
		} else {
			x1 = start.X(sb, GSideRight, GArrowOutgoing)
			y1 = start.SideY(sb, GSideRight)
			y1b := end.RealY(sb).CurrentValue()
			arrows.addPoint(x1, y1)
			arrows.addPoint(x1+6, y1)
			arrows.addPoint(x1+6, y1b)
			arrows.addPoint(x2-8, y1b)
			arrows.addPoint(x2-8, y2)
			arrows.addPoint(x2, y2)
		}
	case a.atStart == GSideRight && a.atEnd == GSideRight:
		xmax := math.Max(x1, x2) + 8
		arrows.addPoint(x1, y1)
		arrows.addPoint(xmax, y1)
		arrows.addPoint(xmax, y2)
		arrows.addPoint(x2, y2)
	case a.atStart == GSideLeft && a.atEnd == GSideLeft:
		xmin := math.Min(x1, x2) - 8
		arrows.addPoint(x1, y1)
		arrows.addPoint(xmin, y1)
		arrows.addPoint(xmin, y2)
		arrows.addPoint(x2, y2)
	default:
		arrows.addPoint(x1, y1)
		arrows.addPoint(x1, y2)
		arrows.addPoint(x2, y2)
	}
	arrows.drawU(ug)
}

// gArrows is upstream's GArrows: the path of an arrow and its head.
type gArrows struct {
	atEnd     GSide
	backColor color.HColor
	points    []geom.XPoint2D
}

const arrowDelta2 = 4

func (g *gArrows) addPoint(x, y float64) { g.points = append(g.points, geom.NewXPoint2D(x, y)) }

func (g *gArrows) drawU(ug klimt.UGraphic) {
	path := klimt.UPathNone()
	path.MoveTo(g.points[0].X, g.points[0].Y)
	for i := 1; i < len(g.points)-1; i++ {
		path.LineTo(g.points[i].X, g.points[i].Y)
	}
	last := g.lastPointPatched()
	path.LineTo(last.X, last.Y)
	path.ClosePath()
	ug.Draw(path)
	ug = ug.Apply(klimt.UStrokeSimple()).Apply(klimt.Bg(g.backColor))
	ug.Draw(g.head())
}

// lastPointPatched is the last point, 3 pixels into the head.
func (g *gArrows) lastPointPatched() geom.XPoint2D {
	last := g.points[len(g.points)-1]
	switch g.atEnd {
	case GSideLeft:
		return geom.NewXPoint2D(last.X-3, last.Y)
	case GSideRight:
		return geom.NewXPoint2D(last.X+3, last.Y)
	}
	return last
}

func (g *gArrows) head() *klimt.UPolygon {
	last := g.points[len(g.points)-1]
	x, y := last.X, last.Y
	switch g.atEnd {
	case GSideLeft:
		polygon := klimt.NewUPolygonNamed("asToRight")
		polygon.AddPoint(x-4, y-arrowDelta2)
		polygon.AddPoint(x, y)
		polygon.AddPoint(x-4, y+arrowDelta2)
		polygon.AddPoint(x-4, y-arrowDelta2)
		return polygon
	case GSideRight:
		polygon := klimt.NewUPolygonNamed("asToLeft")
		polygon.AddPoint(x+4, y-arrowDelta2)
		polygon.AddPoint(x, y)
		polygon.AddPoint(x+4, y+arrowDelta2)
		polygon.AddPoint(x+4, y-arrowDelta2)
		return polygon
	}
	return klimt.NewUPolygon()
}
