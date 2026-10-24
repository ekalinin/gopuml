// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/PositionableUtils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

func positionableUtilsConvert(positionable Positionable) XRectangle2D {
	position := positionable.Position()
	size := positionable.Size()
	return XRectangle2D{X: position.X, Y: position.Y, Width: size.Width(), Height: size.Height()}
}

// PositionableUtilsIntersect reports whether the rectangles of big and small
// intersect.
func PositionableUtilsIntersect(big, small Positionable) bool {
	bigR := positionableUtilsConvert(big)
	smallR := positionableUtilsConvert(small)
	return bigR.Intersects(smallR)
}

// PositionableUtilsAddMargin returns a view of pos enlarged by widthMargin
// on the left and right sides and by heightMargin on the top and bottom
// sides. Moving the view moves pos.
func PositionableUtilsAddMargin(pos Positionable, widthMargin, heightMargin float64) Positionable {
	return &marginPositionable{pos: pos, widthMargin: widthMargin, heightMargin: heightMargin}
}

// marginPositionable is the anonymous class of upstream's addMargin.
type marginPositionable struct {
	pos                       Positionable
	widthMargin, heightMargin float64
}

func (m *marginPositionable) Position() XPoint2D {
	p := m.pos.Position()
	return XPoint2D{X: p.X - m.widthMargin, Y: p.Y - m.heightMargin}
}

func (m *marginPositionable) Size() XDimension2D {
	return m.pos.Size().DeltaWH(float64(2*m.widthMargin), float64(2*m.heightMargin))
}

func (m *marginPositionable) MoveDelta(deltaX, deltaY float64) {
	m.pos.MoveDelta(deltaX, deltaY)
}

func positionableUtilsCenter(p Positionable) XPoint2D {
	pt := p.Position()
	dim := p.Size()
	return XPoint2D{X: pt.X + dim.Width()/2, Y: pt.Y + dim.Height()/2}
}

func positionableUtilsMove(p Positionable, deltaX, deltaY float64) Positionable {
	pt := p.Position()
	dim := p.Size()
	return NewPositionableImpl(pt.X+deltaX, pt.Y+deltaY, dim)
}

// PositionableUtilsMoveAwayFrom returns a copy of toMove moved away from the
// center of fixe, along the line between both centers, just enough to stop
// intersecting it (found by bisection). It panics when they do not intersect,
// as upstream throws IllegalArgumentException.
func PositionableUtilsMoveAwayFrom(fixe, toMove Positionable) Positionable {
	centerFixe := positionableUtilsCenter(fixe)
	centerToMove := positionableUtilsCenter(toMove)

	deltaX := centerToMove.X - centerFixe.X
	deltaY := centerToMove.Y - centerFixe.Y

	minimum := 0.0
	if !doesIntersectWithThisCoef(fixe, toMove, deltaX, deltaY, minimum) {
		panic("geom: positionables do not intersect")
	}

	maximum := 0.1
	for doesIntersectWithThisCoef(fixe, toMove, deltaX, deltaY, maximum) {
		maximum = float64(maximum * 2)
	}

	for range 5 {
		candidate := (minimum + maximum) / 2.0
		if doesIntersectWithThisCoef(fixe, toMove, deltaX, deltaY, candidate) {
			minimum = candidate
		} else {
			maximum = candidate
		}
	}
	candidate := (minimum + maximum) / 2.0
	return positionableUtilsMove(toMove, float64(deltaX*candidate), float64(deltaY*candidate))
}

func doesIntersectWithThisCoef(fixe, toMove Positionable, deltaX, deltaY, c float64) bool {
	result := positionableUtilsMove(toMove, float64(deltaX*c), float64(deltaY*c))
	return PositionableUtilsIntersect(fixe, result)
}
