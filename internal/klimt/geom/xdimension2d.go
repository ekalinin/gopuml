// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/geom/XDimension2D.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

import "math"

// XDimension2D is an immutable, non-negative width and height. The zero value
// is the empty dimension.
//
// Upstream's applyTranslate(UTranslate) and fromDimension(java.awt.Dimension)
// are not ported yet.
type XDimension2D struct {
	width, height float64
}

// NewXDimension2D returns the dimension width x height. It panics when either
// value is negative or NaN, as upstream throws IllegalArgumentException.
func NewXDimension2D(width, height float64) XDimension2D {
	if width < 0 {
		panic("geom: negative width")
	}
	if height < 0 {
		panic("geom: negative height")
	}
	if math.IsNaN(width) || math.IsNaN(height) {
		panic("geom: NaN dimension")
	}
	return XDimension2D{width: width, height: height}
}

// String returns upstream's format, for example "[10.0,20.5]".
func (d XDimension2D) String() string {
	return "[" + javaDoubleString(d.width) + "," + javaDoubleString(d.height) + "]"
}

// Height returns the height.
func (d XDimension2D) Height() float64 { return d.height }

// Width returns the width.
func (d XDimension2D) Width() float64 { return d.width }

// Delta returns the dimension grown by delta in both directions.
func (d XDimension2D) Delta(delta float64) XDimension2D {
	return d.DeltaWH(delta, delta)
}

// WithWidth returns the dimension with another width.
func (d XDimension2D) WithWidth(newWidth float64) XDimension2D {
	return NewXDimension2D(newWidth, d.height)
}

// DeltaWH returns the dimension grown by deltaWidth and deltaHeight. It is
// upstream's delta(double, double).
func (d XDimension2D) DeltaWH(deltaWidth, deltaHeight float64) XDimension2D {
	if deltaHeight == 0 && deltaWidth == 0 {
		return d
	}
	return NewXDimension2D(d.width+deltaWidth, d.height+deltaHeight)
}

// MergeTB stacks bottom below d: the widest width and the sum of heights.
func (d XDimension2D) MergeTB(bottom XDimension2D) XDimension2D {
	width := max(d.width, bottom.width)
	height := d.height + bottom.height
	return NewXDimension2D(width, height)
}

// MergeTB3 stacks d, b and c from top to bottom. It is upstream's
// mergeTB(XDimension2D, XDimension2D).
func (d XDimension2D) MergeTB3(b, c XDimension2D) XDimension2D {
	width := max(max(d.width, b.width), c.width)
	height := d.height + b.height + c.height
	return NewXDimension2D(width, height)
}

// MergeLR puts right at the right of d: the sum of widths and the tallest
// height.
func (d XDimension2D) MergeLR(right XDimension2D) XDimension2D {
	height := max(d.height, right.height)
	width := d.width + right.width
	return NewXDimension2D(width, height)
}

// AtLeast returns the dimension enlarged to at least minWidth x minHeight.
func (d XDimension2D) AtLeast(minWidth, minHeight float64) XDimension2D {
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
	return NewXDimension2D(w, h)
}

// XDimension2DMergeLayoutT12B3 returns the dimension of top1, top2 and bottom
// stacked from top to bottom.
func XDimension2DMergeLayoutT12B3(top1, top2, bottom XDimension2D) XDimension2D {
	width := max(max(top1.width, top2.width), bottom.width)
	height := top1.height + top2.height + bottom.height
	return NewXDimension2D(width, height)
}

// XDimension2DMax returns the smallest dimension containing dim1 and dim2.
func XDimension2DMax(dim1, dim2 XDimension2D) XDimension2D {
	return dim1.AtLeast(dim2.width, dim2.height)
}
