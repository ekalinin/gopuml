// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/shape/URectangle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import "github.com/ekalinin/gopuml/internal/klimt/geom"

// URectangle is a rectangle, possibly with round corners. Only its shadow
// can be changed after creation.
type URectangle struct {
	AbstractShadowable
	width                   float64
	height                  float64
	rx                      float64
	ry                      float64
	ignoreForCompressionOnX bool
	ignoreForCompressionOnY bool
}

var (
	_ UShapeSized                   = (*URectangle)(nil)
	_ UShapeIgnorableForCompression = (*URectangle)(nil)
	_ Shadowable                    = (*URectangle)(nil)
)

// newURectangle is upstream's private constructor. Upstream takes a comment
// and a code line but never stores them; they are dropped here too.
func newURectangle(width, height, rx, ry float64, ignoreX, ignoreY bool) *URectangle {
	if height == 0 {
		throw(illegalArgument, "height="+jstr(height))
	}
	if width == 0 {
		throw(illegalArgument, "width="+jstr(width))
	}
	return &URectangle{width: width, height: height, rx: rx, ry: ry, ignoreForCompressionOnX: ignoreX,
		ignoreForCompressionOnY: ignoreY}
}

// URectangleBuild returns the rectangle of a size. A zero width or height
// panics, as upstream throws IllegalArgumentException.
func URectangleBuild(width, height float64) *URectangle {
	return newURectangle(width, height, 0, 0, false, false)
}

// URectangleBuildDim returns the rectangle of a dimension.
func URectangleBuildDim(d geom.XDimension2D) *URectangle {
	return URectangleBuild(d.Width(), d.Height())
}

// WithHeight returns the rectangle with another height and the same shadow.
func (r *URectangle) WithHeight(newHeight float64) *URectangle {
	result := newURectangle(r.width, newHeight, r.rx, r.ry, r.ignoreForCompressionOnX, r.ignoreForCompressionOnY)
	result.SetDeltaShadow(r.DeltaShadow())
	return result
}

// WithWidth returns the rectangle with another width and the same shadow.
func (r *URectangle) WithWidth(newWidth float64) *URectangle {
	result := newURectangle(newWidth, r.height, r.rx, r.ry, r.ignoreForCompressionOnX, r.ignoreForCompressionOnY)
	result.SetDeltaShadow(r.DeltaShadow())
	return result
}

// WithCommentAndCodeLine returns a copy without shadow; upstream drops the
// comment and code line.
func (r *URectangle) WithCommentAndCodeLine(comment, codeLine string) *URectangle {
	return newURectangle(r.width, r.height, r.rx, r.ry, r.ignoreForCompressionOnX, r.ignoreForCompressionOnY)
}

// Rounded returns the rectangle with round corners of the given size, and
// the rectangle itself when there are no corners at all.
func (r *URectangle) Rounded(round float64) *URectangle {
	if round == 0 && r.rx == 0 && r.ry == 0 {
		return r
	}
	return newURectangle(r.width, r.height, round, round, r.ignoreForCompressionOnX, r.ignoreForCompressionOnY)
}

// DiagonalCorner returns the rectangle with cut corners as a path, and the
// rectangle itself for a zero corner.
func (r *URectangle) DiagonalCorner(diagonalCorner float64) Shadowable {
	if r.ignoreForCompressionOnX || r.ignoreForCompressionOnY {
		throw(illegalState, "")
	}
	if diagonalCorner == 0 {
		return r
	}
	width, height := r.width, r.height
	result := UPathNone()
	result.MoveTo(diagonalCorner, 0)
	result.LineTo(width-diagonalCorner, 0)
	result.LineTo(width, diagonalCorner)
	result.LineTo(width, height-diagonalCorner)
	result.LineTo(width-diagonalCorner, height)
	result.LineTo(diagonalCorner, height)
	result.LineTo(0, height-diagonalCorner)
	result.LineTo(0, diagonalCorner)
	result.LineTo(diagonalCorner, 0)
	return result
}

// IgnoreForCompressionOnX returns a copy ignored when compressing on x.
func (r *URectangle) IgnoreForCompressionOnX() *URectangle {
	return newURectangle(r.width, r.height, r.rx, r.ry, true, r.ignoreForCompressionOnY)
}

// IgnoreForCompressionOnY returns a copy ignored when compressing on y.
func (r *URectangle) IgnoreForCompressionOnY() *URectangle {
	return newURectangle(r.width, r.height, r.rx, r.ry, r.ignoreForCompressionOnX, true)
}

// HalfRounded returns the rectangle with its two top corners rounded, as a
// path, and the rectangle itself for a zero corner.
func (r *URectangle) HalfRounded(roundCorner float64) Shadowable {
	if roundCorner == 0 {
		return r
	}
	width, height := r.width, r.height
	path := UPathNone()
	path.MoveTo(roundCorner/2, 0)
	path.LineTo(width-roundCorner/2, 0)
	path.ArcTo(roundCorner/2, roundCorner/2, 0, 0, 1, width, roundCorner/2)
	path.LineTo(width, height)
	path.LineTo(0, height)
	path.LineTo(0, roundCorner/2)
	path.ArcTo(roundCorner/2, roundCorner/2, 0, 0, 1, roundCorner/2, 0)
	return path
}

// String is upstream's toString.
func (r *URectangle) String() string { return "width=" + jstr(r.width) + " height=" + jstr(r.height) }

// Width returns the width.
func (r *URectangle) Width() float64 { return r.width }

// Height returns the height.
func (r *URectangle) Height() float64 { return r.height }

// Rx returns the horizontal corner radius.
func (r *URectangle) Rx() float64 { return r.rx }

// Ry returns the vertical corner radius.
func (r *URectangle) Ry() float64 { return r.ry }

// Clip returns the rectangle itself, as upstream.
func (r *URectangle) Clip(clip UClip) *URectangle { return r }

// MinMax returns the bounds from (0, 0) to the size.
func (r *URectangle) MinMax() geom.MinMax { return geom.MinMaxFromMax(r.width, r.height) }

// DrawWhenCompressed draws the two borders kept by compression.
func (r *URectangle) DrawWhenCompressed(ug UGraphic, mode CompressionMode) {
	if mode == CompressionOnX {
		ug.Draw(NewUEmpty(2, r.Height()))
		ug.Apply(UTranslateDx(r.Width() - 2)).Draw(NewUEmpty(2, r.Height()))
	}
	if mode == CompressionOnY {
		ug.Draw(NewUEmpty(r.Width(), 2))
		ug.Apply(UTranslateDy(r.Height() - 2)).Draw(NewUEmpty(r.Width(), 2))
	}
}

// IsIgnoreForCompressionOn reports whether compression in mode skips the
// rectangle.
func (r *URectangle) IsIgnoreForCompressionOn(mode CompressionMode) bool {
	switch mode {
	case CompressionOnX:
		return r.ignoreForCompressionOnX
	case CompressionOnY:
		return r.ignoreForCompressionOnY
	}
	throw(illegalArgument, "")
	return false
}
