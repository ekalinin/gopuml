// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/Arrows.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/ArrowsRegular.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/ArrowsTriangle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

// Arrows builds the arrow heads of activity diagrams. Upstream's two
// implementations live in activitydiagram3.ftile; they are here because
// ISkinParam.arrows() returns them.
type Arrows interface {
	AsToUp() *UPolygon
	AsToDown() *UPolygon
	AsToRight() *UPolygon
	AsToLeft() *UPolygon
}

// ArrowsAsTo returns the head pointing to direction.
func ArrowsAsTo(a Arrows, direction Direction) *UPolygon {
	switch direction {
	case DirectionUp:
		return a.AsToUp()
	case DirectionDown:
		return a.AsToDown()
	case DirectionLeft:
		return a.AsToLeft()
	case DirectionRight:
		return a.AsToRight()
	}
	throw(illegalArgument, "")
	return nil
}

const (
	arrowsDelta1 = 10.0
	arrowsDelta2 = 4.0
)

// ArrowsRegular are the default heads, with a notch.
type ArrowsRegular struct{}

// AsToUp returns the head pointing up.
func (ArrowsRegular) AsToUp() *UPolygon {
	polygon := NewUPolygonNamed("asToUp")
	polygon.AddPoint(-arrowsDelta2, arrowsDelta1)
	polygon.AddPoint(0, 0)
	polygon.AddPoint(arrowsDelta2, arrowsDelta1)
	polygon.AddPoint(0, arrowsDelta1-4)
	return polygon
}

// AsToDown returns the head pointing down.
func (ArrowsRegular) AsToDown() *UPolygon {
	polygon := NewUPolygonNamed("asToDown")
	polygon.AddPoint(-arrowsDelta2, -arrowsDelta1)
	polygon.AddPoint(0, 0)
	polygon.AddPoint(arrowsDelta2, -arrowsDelta1)
	polygon.AddPoint(0, -arrowsDelta1+4)
	return polygon
}

// AsToRight returns the head pointing right.
func (ArrowsRegular) AsToRight() *UPolygon {
	polygon := NewUPolygonNamed("asToRight")
	polygon.AddPoint(-arrowsDelta1, -arrowsDelta2)
	polygon.AddPoint(0, 0)
	polygon.AddPoint(-arrowsDelta1, arrowsDelta2)
	polygon.AddPoint(-arrowsDelta1+4, 0)
	return polygon
}

// AsToLeft returns the head pointing left.
func (ArrowsRegular) AsToLeft() *UPolygon {
	polygon := NewUPolygonNamed("asToLeft")
	polygon.AddPoint(arrowsDelta1, -arrowsDelta2)
	polygon.AddPoint(0, 0)
	polygon.AddPoint(arrowsDelta1, arrowsDelta2)
	polygon.AddPoint(arrowsDelta1-4, 0)
	return polygon
}

// ArrowsTriangle are the plain triangle heads of "skinparam style
// strictuml".
type ArrowsTriangle struct{}

// AsToUp returns the head pointing up.
func (ArrowsTriangle) AsToUp() *UPolygon {
	polygon := NewUPolygonNamed("asToUp")
	polygon.AddPoint(-arrowsDelta2, arrowsDelta1)
	polygon.AddPoint(0, 0)
	polygon.AddPoint(arrowsDelta2, arrowsDelta1)
	return polygon
}

// AsToDown returns the head pointing down.
func (ArrowsTriangle) AsToDown() *UPolygon {
	polygon := NewUPolygonNamed("asToDown")
	polygon.AddPoint(-arrowsDelta2, -arrowsDelta1)
	polygon.AddPoint(arrowsDelta2, -arrowsDelta1)
	polygon.AddPoint(0, 0)
	return polygon
}

// AsToRight returns the head pointing right.
func (ArrowsTriangle) AsToRight() *UPolygon {
	polygon := NewUPolygonNamed("asToRight")
	polygon.AddPoint(-arrowsDelta1, -arrowsDelta2)
	polygon.AddPoint(0, 0)
	polygon.AddPoint(-arrowsDelta1, arrowsDelta2)
	return polygon
}

// AsToLeft returns the head pointing left.
func (ArrowsTriangle) AsToLeft() *UPolygon {
	polygon := NewUPolygonNamed("asToLeft")
	polygon.AddPoint(arrowsDelta1, -arrowsDelta2)
	polygon.AddPoint(0, 0)
	polygon.AddPoint(arrowsDelta1, arrowsDelta2)
	return polygon
}
