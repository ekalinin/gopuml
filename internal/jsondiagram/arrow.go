// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/jsondiagram/Arrow.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package jsondiagram

import (
	"math"

	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// arrow is the head of an arrow, from p1, the end of its curve, to p2, its
// tip.
type arrow struct {
	p1, p2 geom.XPoint2D
}

func (a arrow) drawArrow(ug klimt.UGraphic) {
	ug = ug.Apply(klimt.UStrokeSimple())

	p1p2 := klimt.ULineCreate(a.p1, a.p2)
	dist := a.p1.Distance(a.p2)

	alpha := jmath.Atan2(p1p2.DX(), p1p2.DY())

	const factor = .4
	const factor2 = .3

	p3 := arrowPoint(a.p1, alpha+math.Pi/2, dist*factor)
	p4 := arrowPoint(a.p1, alpha-math.Pi/2, dist*factor)
	p11 := arrowPoint(a.p1, alpha, dist*factor2)

	path := klimt.UPathNone()
	path.MoveToPoint(p4)
	path.LineToPoint(p11)
	path.LineToPoint(p3)
	path.LineToPoint(a.p2)
	path.LineToPoint(p4)
	path.ClosePath()

	ug.Draw(path)
}

func arrowPoint(center geom.XPoint2D, alpha, length float64) geom.XPoint2D {
	return geom.NewXPoint2D(center.X+length*jmath.Sin(alpha), center.Y+length*jmath.Cos(alpha))
}
