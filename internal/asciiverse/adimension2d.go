// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/asciiverse/ADimension2D.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/asciiverse/ATranslate.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package asciiverse

import "strconv"

// ADimension2D is a size on the character grid: a number of columns and a
// number of rows, never negative.
type ADimension2D struct {
	width, height int
}

// NewADimension2D returns the size; a negative width or height panics with
// IllegalArgumentException, as upstream.
func NewADimension2D(width, height int) ADimension2D {
	if width < 0 {
		throw(illegalArgument, "")
	}
	if height < 0 {
		throw(illegalArgument, "")
	}
	return ADimension2D{width: width, height: height}
}

// String returns "[width,height]".
func (d ADimension2D) String() string {
	return "[" + strconv.Itoa(d.width) + "," + strconv.Itoa(d.height) + "]"
}

// Height returns the number of rows.
func (d ADimension2D) Height() int { return d.height }

// Width returns the number of columns.
func (d ADimension2D) Width() int { return d.width }

// Delta returns the size grown by delta in both directions.
func (d ADimension2D) Delta(delta int) ADimension2D { return d.DeltaWH(delta, delta) }

// WithWidth returns the size with another width.
func (d ADimension2D) WithWidth(newWidth int) ADimension2D {
	return NewADimension2D(newWidth, d.height)
}

// DeltaWH returns the size grown by deltaWidth columns and deltaHeight rows
// (upstream's delta(int, int)).
func (d ADimension2D) DeltaWH(deltaWidth, deltaHeight int) ADimension2D {
	if deltaHeight == 0 && deltaWidth == 0 {
		return d
	}
	return NewADimension2D(d.width+deltaWidth, d.height+deltaHeight)
}

// MergeTB returns the size of this block above bottom.
func (d ADimension2D) MergeTB(bottom ADimension2D) ADimension2D {
	return NewADimension2D(max(d.width, bottom.width), d.height+bottom.height)
}

// MergeLR returns the size of this block left of right.
func (d ADimension2D) MergeLR(right ADimension2D) ADimension2D {
	return NewADimension2D(d.width+right.width, max(d.height, right.height))
}

// AtLeast returns the size with at least minWidth columns and minHeight
// rows.
func (d ADimension2D) AtLeast(minWidth, minHeight int) ADimension2D {
	h := d.height
	w := d.width
	if w > minWidth && h > minHeight {
		return d
	}
	if h < minHeight {
		h = minHeight
	}
	if w < minWidth {
		w = minWidth
	}
	return NewADimension2D(w, h)
}

// ADimension2DMax returns the size that contains both sizes.
func ADimension2DMax(dim1, dim2 ADimension2D) ADimension2D {
	return dim1.AtLeast(dim2.width, dim2.height)
}

// ATranslate is a move on the character grid. Nothing in upstream uses it
// yet.
type ATranslate struct {
	dx, dy int
}

// String returns "translate dx=.. dy=..".
func (t ATranslate) String() string {
	return "translate dx=" + strconv.Itoa(t.dx) + " dy=" + strconv.Itoa(t.dy)
}

// ATranslateNone returns the null move.
func ATranslateNone() ATranslate { return ATranslate{} }

// NewATranslate returns the move.
func NewATranslate(dx, dy int) ATranslate { return ATranslate{dx: dx, dy: dy} }

// ATranslateDx returns a horizontal move.
func ATranslateDx(dx int) ATranslate { return ATranslate{dx: dx} }

// ATranslateDy returns a vertical move.
func ATranslateDy(dy int) ATranslate { return ATranslate{dy: dy} }

// Dx returns the horizontal move.
func (t ATranslate) Dx() float64 { return float64(t.dx) }

// Dy returns the vertical move.
func (t ATranslate) Dy() float64 { return float64(t.dy) }

// IsAlmostSame reports whether the moves share dx or dy (upstream's test,
// an "or").
func (t ATranslate) IsAlmostSame(other ATranslate) bool { return t.dx == other.dx || t.dy == other.dy }

// Compose returns this move followed by other.
func (t ATranslate) Compose(other ATranslate) ATranslate {
	return ATranslate{dx: t.dx + other.dx, dy: t.dy + other.dy}
}

// Reverse returns the opposite move.
func (t ATranslate) Reverse() ATranslate { return ATranslate{dx: -t.dx, dy: -t.dy} }
