// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/XRectangle2D.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

// XRectangle2D is an immutable rectangle given by its top left corner and its
// size.
type XRectangle2D struct {
	X, Y, Width, Height float64
}

// NewXRectangle2D returns the rectangle at (x, y) of size width x height.
func NewXRectangle2D(x, y, width, height float64) XRectangle2D {
	return XRectangle2D{X: x, Y: y, Width: width, Height: height}
}

// CenterX returns the abscissa of the center.
func (r XRectangle2D) CenterX() float64 { return r.X + r.Width/2 }

// CenterY returns the ordinate of the center.
func (r XRectangle2D) CenterY() float64 { return r.Y + r.Height/2 }

// MinX returns the left side abscissa.
func (r XRectangle2D) MinX() float64 { return r.X }

// MaxX returns the right side abscissa.
func (r XRectangle2D) MaxX() float64 { return r.X + r.Width }

// MinY returns the top side ordinate.
func (r XRectangle2D) MinY() float64 { return r.Y }

// MaxY returns the bottom side ordinate.
func (r XRectangle2D) MaxY() float64 { return r.Y + r.Height }

// Intersects reports whether both rectangles are non-empty and overlap.
// Rectangles that only touch do not intersect.
func (r XRectangle2D) Intersects(other XRectangle2D) bool {
	return r.intersects(other.X, other.Y, other.Width, other.Height)
}

func (r XRectangle2D) intersects(x, y, w, h float64) bool {
	return w > 0 && h > 0 && r.Width > 0 && r.Height > 0 && x < r.X+r.Width && x+w > r.X &&
		y < r.Y+r.Height && y+h > r.Y
}

// ContainsPoint is upstream's contains(XPoint2D), which is not supported: it
// always panics.
func (r XRectangle2D) ContainsPoint(point XPoint2D) bool {
	panic("geom: XRectangle2D.contains(XPoint2D) is not supported")
}

// Contains reports whether (xp, yp) is inside the rectangle, the left and top
// sides included and the right and bottom sides excluded. It panics when the
// rectangle is empty, as upstream throws IllegalStateException.
func (r XRectangle2D) Contains(xp, yp float64) bool {
	if r.Width <= 0 || r.Height <= 0 {
		panic("geom: contains on an empty rectangle")
	}
	return xp >= r.MinX() && xp < r.MaxX() && yp >= r.MinY() && yp < r.MaxY()
}

// Intersect returns the first intersection of line with the sides of the
// rectangle, trying the top, right, bottom and left sides in that order. The
// boolean is false when there is none (upstream returns null).
func (r XRectangle2D) Intersect(line XLine2D) (XPoint2D, bool) {
	a := XPoint2D{X: r.X, Y: r.Y}
	b := XPoint2D{X: r.X + r.Width, Y: r.Y}
	c := XPoint2D{X: r.X + r.Width, Y: r.Y + r.Height}
	d := XPoint2D{X: r.X, Y: r.Y + r.Height}
	for _, side := range [...]XLine2D{XLine2DLine(a, b), XLine2DLine(b, c), XLine2DLine(c, d), XLine2DLine(d, a)} {
		if result, ok := line.Intersect(side); ok {
			return result, true
		}
	}
	return XPoint2D{}, false
}
