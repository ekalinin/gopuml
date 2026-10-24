// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/UClip.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"

	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// UClip is a clipping rectangle, an immutable value.
type UClip struct {
	x, y, width, height float64
}

// NewUClip returns the clip of the rectangle.
func NewUClip(x, y, width, height float64) UClip {
	return UClip{x: x, y: y, width: width, height: height}
}

// Enlarge returns the clip grown by delta on every side.
func (c UClip) Enlarge(delta float64) UClip {
	return UClip{x: c.x - delta, y: c.y - delta, width: c.width + float64(2*delta), height: c.height + float64(2*delta)}
}

// String is upstream's toString.
func (c UClip) String() string {
	return "CLIP x=" + jstr(c.x) + " y=" + jstr(c.y) + " w=" + jstr(c.width) + " h=" + jstr(c.height)
}

// Translate returns the clip moved by (dx, dy).
func (c UClip) Translate(dx, dy float64) UClip {
	return UClip{x: c.x + dx, y: c.y + dy, width: c.width, height: c.height}
}

// TranslateBy returns the clip moved by t.
func (c UClip) TranslateBy(t UTranslate) UClip { return c.Translate(t.Dx(), t.Dy()) }

// X returns the left side.
func (c UClip) X() float64 { return c.x }

// Y returns the top side.
func (c UClip) Y() float64 { return c.y }

// Width returns the width.
func (c UClip) Width() float64 { return c.width }

// Height returns the height.
func (c UClip) Height() float64 { return c.height }

// IsInsidePoint reports whether pt is inside the clip, borders included.
func (c UClip) IsInsidePoint(pt geom.XPoint2D) bool { return c.IsInside(pt.X, pt.Y) }

// IsInside reports whether (xp, yp) is inside the clip, borders included.
func (c UClip) IsInside(xp, yp float64) bool {
	if xp < c.x {
		return false
	}
	if xp > c.x+c.width {
		return false
	}
	if yp < c.y {
		return false
	}
	if yp > c.y+c.height {
		return false
	}
	return true
}

// ClippedRectangle returns the intersection of r and the clip.
func (c UClip) ClippedRectangle(r geom.XRectangle2D) geom.XRectangle2D {
	x1 := math.Max(r.X, c.x)
	y1 := math.Max(r.Y, c.y)
	x2 := math.Min(r.X+r.Width, c.x+c.width)
	y2 := math.Min(r.Y+r.Height, c.y+c.height)
	return geom.NewXRectangle2D(x1, y1, x2-x1, y2-y1)
}

// ClippedLine returns the part of a horizontal or vertical line inside the
// clip; ok is false when there is none (null upstream).
func (c UClip) ClippedLine(line geom.XLine2D) (clipped geom.XLine2D, ok bool) {
	in1 := c.IsInside(line.X1, line.Y1)
	in2 := c.IsInside(line.X2, line.Y2)
	if in1 && in2 {
		return line, true
	}
	if !in1 && !in2 {
		if line.X1 == line.X2 {
			newy1 := c.clippedY(line.Y1)
			newy2 := c.clippedY(line.Y2)
			if newy1 != newy2 {
				return geom.NewXLine2D(line.X1, newy1, line.X2, newy2), true
			}
		}
		return geom.XLine2D{}, false
	}
	if line.X1 != line.X2 && line.Y1 != line.Y2 {
		return geom.XLine2D{}, false
	}
	if line.Y1 == line.Y2 {
		return geom.NewXLine2D(c.clippedX(line.X1), line.Y1, c.clippedX(line.X2), line.Y2), true
	}
	if line.X1 == line.X2 {
		return geom.NewXLine2D(line.X1, c.clippedY(line.Y1), line.X2, c.clippedY(line.Y2)), true
	}
	throw(illegalState, "")
	return geom.XLine2D{}, false
}

// IsInsidePath reports whether the bounds of shape, drawn at (x, y), are
// inside the clip.
func (c UClip) IsInsidePath(x, y float64, shape *UPath) bool {
	return c.IsInside(x+shape.MinX(), y+shape.MinY()) && c.IsInside(x+shape.MaxX(), y+shape.MaxY())
}

func (c UClip) clippedX(xp float64) float64 {
	if xp < c.x {
		return c.x
	}
	if xp > c.x+c.width {
		return c.x + c.width
	}
	return xp
}

func (c UClip) clippedY(yp float64) float64 {
	if yp < c.y {
		return c.y
	}
	if yp > c.y+c.height {
		return c.y + c.height
	}
	return yp
}
