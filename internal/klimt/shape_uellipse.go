// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/UEllipse.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"

	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// UEllipse is an ellipse inscribed in a width x height box, or an arc of it
// when extend is not 0 (angles in degrees, as java.awt.geom.Arc2D).
type UEllipse struct {
	AbstractShadowable
	width  float64
	height float64
	start  float64
	extend float64
}

var (
	_ UShapeSized = (*UEllipse)(nil)
	_ Shadowable  = (*UEllipse)(nil)
)

// UEllipseBuild returns the full ellipse of a box.
func UEllipseBuild(width, height float64) *UEllipse { return NewUEllipse(width, height, 0, 0) }

// NewUEllipse returns an ellipse or an arc.
func NewUEllipse(width, height, start, extend float64) *UEllipse {
	return &UEllipse{width: width, height: height, start: start, extend: extend}
}

// Width returns the width of the box.
func (e *UEllipse) Width() float64 { return e.width }

// Height returns the height of the box.
func (e *UEllipse) Height() float64 { return e.height }

// Start returns the start angle of an arc.
func (e *UEllipse) Start() float64 { return e.start }

// Extend returns the angular extent of an arc, 0 for a full ellipse.
func (e *UEllipse) Extend() float64 { return e.extend }

// Dimension returns the size of the box.
func (e *UEllipse) Dimension() geom.XDimension2D { return geom.NewXDimension2D(e.width, e.height) }

// Bigger returns the full ellipse grown by more, with the same shadow.
func (e *UEllipse) Bigger(more float64) *UEllipse {
	result := UEllipseBuild(e.width+more, e.height+more)
	result.SetDeltaShadow(e.DeltaShadow())
	return result
}

// Scale returns the full ellipse scaled by factor, with the same shadow.
func (e *UEllipse) Scale(factor float64) *UEllipse {
	result := UEllipseBuild(e.width*factor, e.height*factor)
	result.SetDeltaShadow(e.DeltaShadow())
	return result
}

// StartingX returns the x of the left side of the ellipse at height y.
func (e *UEllipse) StartingX(y float64) float64 {
	y = y / e.height * 2
	x := 1 - math.Sqrt(1-float64((y-1)*(y-1)))
	return x * e.width / 2
}

// EndingX returns the x of the right side of the ellipse at height y.
func (e *UEllipse) EndingX(y float64) float64 {
	y = y / e.height * 2
	x := 1 + math.Sqrt(1-float64((y-1)*(y-1)))
	return x * e.width / 2
}

// PointAtAngle returns the point of the ellipse at angle alpha radians.
func (e *UEllipse) PointAtAngle(alpha float64) geom.XPoint2D {
	x := e.width/2 + float64(e.width/2*jmath.Cos(alpha))
	y := e.height/2 + float64(e.height/2*jmath.Sin(alpha))
	return geom.NewXPoint2D(x, y)
}
