// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/PositionableImpl.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

// PositionableImpl is a Positionable with a fixed size.
type PositionableImpl struct {
	pos XPoint2D
	dim XDimension2D
}

// NewPositionableImpl returns an item of size dim at (x, y).
func NewPositionableImpl(x, y float64, dim XDimension2D) *PositionableImpl {
	return &PositionableImpl{pos: XPoint2D{X: x, Y: y}, dim: dim}
}

// PositionableImplCreate returns an item of size dim at pt.
func PositionableImplCreate(pt XPoint2D, dim XDimension2D) *PositionableImpl {
	return NewPositionableImpl(pt.X, pt.Y, dim)
}

// Position returns the top left corner.
func (p *PositionableImpl) Position() XPoint2D { return p.pos }

// Size returns the size.
func (p *PositionableImpl) Size() XDimension2D { return p.dim }

// MoveDelta moves the item by (deltaX, deltaY).
func (p *PositionableImpl) MoveDelta(deltaX, deltaY float64) {
	p.pos = p.pos.Move(deltaX, deltaY)
}
