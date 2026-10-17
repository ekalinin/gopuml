// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/utils/Direction.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import "github.com/ekalinin/gopuml/internal/klimt/geom"

// Direction is one of the four directions. Upstream defines it in
// net.sourceforge.plantuml.utils; it lives here because Arrows, GuideLine
// and TextBlockArrow need it. Upstream's getWBSDirection belongs to the WBS
// diagram.
type Direction int

// Directions, in upstream's order.
const (
	DirectionRight Direction = iota
	DirectionLeft
	DirectionDown
	DirectionUp
)

var directionNames = []string{"RIGHT", "LEFT", "DOWN", "UP"}

// String returns upstream's constant name.
func (d Direction) String() string { return enumName("Direction", directionNames, int(d)) }

// Inv returns the opposite direction.
func (d Direction) Inv() Direction {
	switch d {
	case DirectionRight:
		return DirectionLeft
	case DirectionLeft:
		return DirectionRight
	case DirectionDown:
		return DirectionUp
	case DirectionUp:
		return DirectionDown
	}
	throw(illegalState, "")
	return d
}

// ShortCode returns the first letter of the name.
func (d Direction) ShortCode() string { return d.String()[:1] }

// DirectionFromChar returns the direction of '<', '>' or '^', and Down for
// any other character.
func DirectionFromChar(c rune) Direction {
	switch c {
	case '<':
		return DirectionLeft
	case '>':
		return DirectionRight
	case '^':
		return DirectionUp
	}
	return DirectionDown
}

// Clockwise returns the direction turned by a quarter clockwise.
func (d Direction) Clockwise() Direction {
	switch d {
	case DirectionRight:
		return DirectionDown
	case DirectionLeft:
		return DirectionUp
	case DirectionDown:
		return DirectionLeft
	case DirectionUp:
		return DirectionRight
	}
	throw(illegalState, "")
	return d
}

// DirectionLeftOrRight returns Left when p1 is left of p2 and Right when it
// is right of it; it panics when they have the same x, as upstream throws.
func DirectionLeftOrRight(p1, p2 geom.XPoint2D) Direction {
	if p1.X < p2.X {
		return DirectionLeft
	}
	if p1.X > p2.X {
		return DirectionRight
	}
	throw(illegalArgument, "")
	return DirectionLeft
}

// DirectionFromVector returns the direction of a horizontal or vertical
// vector; ok is false for a null vector (null upstream). A diagonal vector
// panics, as upstream throws.
func DirectionFromVector(p1, p2 geom.XPoint2D) (d Direction, ok bool) {
	x1, y1, x2, y2 := p1.X, p1.Y, p2.X, p2.Y
	if x1 == x2 && y1 == y2 {
		return 0, false
	}
	if x1 == x2 {
		if y2 > y1 {
			return DirectionDown, true
		}
		return DirectionUp, true
	}
	if y1 == y2 {
		if x2 > x1 {
			return DirectionRight, true
		}
		return DirectionLeft, true
	}
	throw(illegalArgument, "Not a H or V line!")
	return 0, false
}

// DirectionLazzyValueOf returns the direction named s ignoring case, with
// "TOP" for Up and "BOTTOM" for Down. An unknown name panics, as Java's
// valueOf throws.
func DirectionLazzyValueOf(s string) Direction {
	s = javaUpper(s)
	switch s {
	case "TOP":
		return DirectionUp
	case "BOTTOM":
		return DirectionDown
	}
	for i, name := range directionNames {
		if name == s {
			return Direction(i)
		}
	}
	throw(illegalArgument, "No enum constant net.sourceforge.plantuml.utils.Direction."+s)
	return 0
}
