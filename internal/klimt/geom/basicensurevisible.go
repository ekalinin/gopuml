// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/BasicEnsureVisible.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

import (
	"math"
	"strconv"
)

// BasicEnsureVisible is an EnsureVisible that tracks the bounding box of the
// points. Create it with NewBasicEnsureVisible.
type BasicEnsureVisible struct {
	minX, maxX, minY, maxY float64
}

// NewBasicEnsureVisible returns a tracker that has seen no point yet.
func NewBasicEnsureVisible() *BasicEnsureVisible {
	return &BasicEnsureVisible{
		minX: math.MaxFloat64,
		maxX: -math.MaxFloat64,
		minY: math.MaxFloat64,
		maxY: -math.MaxFloat64,
	}
}

// EnsureVisible takes (x, y) into account.
func (b *BasicEnsureVisible) EnsureVisible(x, y float64) {
	if x > b.maxX {
		b.maxX = x
	}
	if x < b.minX {
		b.minX = x
	}
	if y > b.maxY {
		b.maxY = y
	}
	if y < b.minY {
		b.minY = y
	}
}

// HasData reports whether a point was taken into account.
func (b *BasicEnsureVisible) HasData() bool {
	return b.minX != math.MaxFloat64
}

// Coords returns the scaled bounding box as "x1,y1,x2,y2", each coordinate
// truncated to an integer as Java's (int) cast does, or "0,0,0,0" when there
// is no point.
func (b *BasicEnsureVisible) Coords(scale float64) string {
	if b.minX == math.MaxFloat64 {
		return "0,0,0,0"
	}
	x1 := javaD2I(b.minX * scale)
	y1 := javaD2I(b.minY * scale)
	x2 := javaD2I(b.maxX * scale)
	y2 := javaD2I(b.maxY * scale)
	return strconv.Itoa(int(x1)) + "," + strconv.Itoa(int(y1)) + "," + strconv.Itoa(int(x2)) + "," +
		strconv.Itoa(int(y2))
}

// Surface returns the area of the bounding box, or 0 when there is no point.
func (b *BasicEnsureVisible) Surface() float64 {
	if b.minX == math.MaxFloat64 {
		return 0
	}
	return float64((b.maxX - b.minX) * (b.maxY - b.minY))
}
