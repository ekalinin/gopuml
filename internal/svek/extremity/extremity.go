// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/Extremity.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityFactory.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/MiddleFactory.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/AbstractExtremityFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"math"
	"strconv"

	"github.com/ekalinin/plantuml-in-go/internal/jmath"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// pi is Math.PI as a variable, so that expressions built on it are rounded
// step by step as Java's double arithmetic does (Go evaluates constant
// expressions exactly).
var pi = math.Pi

// Factory is upstream's ExtremityFactory: it builds the decoration at one
// end of a link. A nil side is Java's null.
type Factory interface {
	// CreateTBRDrawableLegacy builds the decoration from the three points of
	// the polygon that Graphviz drew for the arrow head: p1 is its tip.
	CreateTBRDrawableLegacy(p0, p1, p2 geom.XPoint2D, side *geom.Side) klimt.UDrawable
	// CreateUDrawable builds the decoration at p0, pointing along angle.
	CreateUDrawable(p0 geom.XPoint2D, angle float64, side *geom.Side) klimt.UDrawable
}

// MiddleFactory is upstream's MiddleFactory: it builds the decoration in the
// middle of a link.
type MiddleFactory interface {
	CreateUDrawable(angle float64) klimt.UDrawable
}

// Extremity is upstream's abstract Extremity: the decoration at one end (or
// in the middle) of a link.
type Extremity interface {
	klimt.UDrawable
	// SomePoint returns a point of the decoration, nil for none.
	SomePoint() *geom.XPoint2D
	// DeltaForKal returns the move of a qualifier box placed on the
	// decoration.
	DeltaForKal() klimt.UTranslate
	// DecorationLength returns the length that the decoration takes on the
	// line.
	DecorationLength() float64
}

// extremityBase holds the default methods of upstream's Extremity.
type extremityBase struct{}

// DeltaForKal returns the null translation, upstream's default.
func (extremityBase) DeltaForKal() klimt.UTranslate { return klimt.UTranslateNone() }

// DecorationLength returns 8, upstream's default.
func (extremityBase) DecorationLength() float64 { return 8 }

// manageround snaps an angle within 0.05 degree of a multiple of 90 degrees
// to it.
func manageround(angle float64) float64 {
	deg := angle * 180.0 / pi
	switch {
	case isCloseTo(0, deg):
		return 0
	case isCloseTo(90, deg):
		return quadrant(90)
	case isCloseTo(180, deg):
		return quadrant(180)
	case isCloseTo(270, deg):
		return quadrant(270)
	case isCloseTo(360, deg):
		return 0
	}
	return angle
}

// quadrant is upstream's deg * Math.PI / 180.0, rounded step by step.
func quadrant(deg float64) float64 { return deg * pi / 180.0 }

func isCloseTo(value, variable float64) bool { return math.Abs(value-variable) < 0.05 }

// point returns a pointer to a copy of p, for SomePoint.
func point(p geom.XPoint2D) *geom.XPoint2D { return &p }

// abstractFactory holds upstream's AbstractExtremityFactory: the default
// createUDrawable and the angle helper. name is the simple class name, for
// the message of the default createUDrawable.
type abstractFactory struct{ name string }

// CreateUDrawable panics, as upstream's default throws
// UnsupportedOperationException.
func (f abstractFactory) CreateUDrawable(geom.XPoint2D, float64, *geom.Side) klimt.UDrawable {
	panic(&color.RuntimeError{Class: "java.lang.UnsupportedOperationException",
		Message: "class net.sourceforge.plantuml.svek.extremity." + f.name})
}

// atan2 returns the angle of the vector p0 to p1, in [0, 2 pi).
func atan2(p1, p0 geom.XPoint2D) float64 {
	a := jmath.Atan2(p1.Y-p0.Y, p1.X-p0.X)
	if a < 0 {
		a += pi * 2
	}
	return a
}

// unsupported panics as Java's new UnsupportedOperationException().
func unsupported() {
	panic(&color.RuntimeError{Class: "java.lang.UnsupportedOperationException"})
}

// drawLine draws the segment p1 to p2 at (x, y), upstream's private
// drawLine of several extremities.
func drawLine(ug klimt.UGraphic, x, y float64, p1, p2 geom.XPoint2D) {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y
	ug.Apply(klimt.NewUTranslate(x+p1.X, y+p1.Y)).Draw(klimt.NewULine(dx, dy))
}

// polygonPoint is UPolygon.getPoint with Java's index check.
func polygonPoint(p *klimt.UPolygon, idx int) geom.XPoint2D {
	if n := len(p.Points()); idx >= n {
		panic(&color.RuntimeError{Class: "java.lang.IndexOutOfBoundsException",
			Message: "Index " + strconv.Itoa(idx) + " out of bounds for length " + strconv.Itoa(n)})
	}
	return p.Point(idx)
}

// rotate is XAffineTransform.getRotateInstance(theta) applied to p.
func rotate(p geom.XPoint2D, theta float64) geom.XPoint2D {
	return klimt.XAffineTransformRotateInstance(theta).Transform(p)
}
