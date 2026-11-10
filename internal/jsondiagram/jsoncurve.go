// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/jsondiagram/JsonCurve.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package jsondiagram

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// jsonCurve is the arrow of a link, from the points of its spline in the
// layout: a dot, a straight start of veryFirstLine points, the spline, and
// the head to the end point of the edge.
type jsonCurve struct {
	points        []geom.XPoint2D
	xMirror       Mirror
	veryFirstLine float64
	// ep is the end point of the edge, nil when it has none.
	ep *geom.XPoint2D
}

// newJSONCurve reads the pos attribute of an edge of the layout: an
// optional start point "s,x,y", an optional end point "e,x,y", then the
// points of its spline. An edge of more than one spline is an error, as
// upstream's JsonCurve throws an IllegalStateException. An end point at the
// origin is no end point, as upstream.
func newJSONCurve(pos string, xMirror Mirror, veryFirstLine float64) (*jsonCurve, error) {
	if strings.Contains(pos, ";") {
		return nil, errors.New("edge of more than one spline: " + pos)
	}
	c := &jsonCurve{xMirror: xMirror, veryFirstLine: veryFirstLine}
	for field := range strings.FieldsSeq(pos) {
		start := strings.HasPrefix(field, "s,")
		end := strings.HasPrefix(field, "e,")
		if start || end {
			field = field[2:]
		}
		pt, err := parsePoint(field)
		if err != nil {
			return nil, err
		}
		switch {
		case start:
		case end:
			if pt.X != 0 || pt.Y != 0 {
				c.ep = &pt
			}
		default:
			c.points = append(c.points, pt)
		}
	}
	if len(c.points) < 4 || (len(c.points)-1)%3 != 0 {
		return nil, fmt.Errorf("spline of %d points: %s", len(c.points), pos)
	}
	return c, nil
}

// parsePoint reads "x,y".
func parsePoint(s string) (geom.XPoint2D, error) {
	xs, ys, ok := strings.Cut(s, ",")
	x, errX := strconv.ParseFloat(xs, 64)
	y, errY := strconv.ParseFloat(ys, 64)
	if !ok || errX != nil || errY != nil {
		return geom.XPoint2D{}, errors.New("point " + strconv.Quote(s))
	}
	return geom.NewXPoint2D(x, y), nil
}

func (c *jsonCurve) drawCurve(hcolor color.HColor, ug klimt.UGraphic) {
	path := klimt.UPathNone()

	path.MoveToPoint(c.veryFirst())
	path.LineToPoint(c.xMirror.invAndXYSwitch(c.points[0]))

	for i := 1; i < len(c.points); i += 3 {
		pt2 := c.xMirror.invAndXYSwitch(c.points[i])
		pt3 := c.xMirror.invAndXYSwitch(c.points[i+1])
		pt4 := c.xMirror.invAndXYSwitch(c.points[i+2])
		path.CubicToPoints(pt2, pt3, pt4)
	}
	ug.Draw(path)

	if c.ep != nil {
		last := c.xMirror.invAndXYSwitch(c.points[len(c.points)-1])
		trueEp := c.xMirror.invAndXYSwitch(*c.ep)
		arrow{p1: last, p2: trueEp}.drawArrow(ug.Apply(klimt.Bg(hcolor)))
	}
}

func (c *jsonCurve) drawSpot(ug klimt.UGraphic) {
	const size = 3
	ug = ug.Apply(klimt.UTranslatePoint(c.veryFirst()).Compose(klimt.NewUTranslate(-size, -size)))
	ug.Apply(klimt.UStrokeSimple()).Draw(klimt.UEllipseBuild(2*size, 2*size))
}

func (c *jsonCurve) veryFirst() geom.XPoint2D {
	return supp(c.xMirror.invAndXYSwitch(c.points[0]), c.xMirror.invAndXYSwitch(c.points[1]), c.veryFirstLine)
}

// supp returns the point at length from center, away from direction.
func supp(center, direction geom.XPoint2D, length float64) geom.XPoint2D {
	full := center.Distance(direction)
	dx := (center.X - direction.X) / full
	dy := (center.Y - direction.Y) / full
	return geom.NewXPoint2D(center.X+dx*length, center.Y+dy*length)
}
