// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/UMotif.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import "github.com/ekalinin/plantuml-in-go/internal/klimt/geom"

// UMotif is a repeated pattern of points, drawn as a path along a rectangle
// border.
type UMotif struct {
	points []geom.XPoint2D
}

// NewUMotifInts returns the motif of x, y pairs. An odd number of values
// panics, as upstream reads past the end of its array.
func NewUMotifInts(data ...int) *UMotif {
	m := &UMotif{}
	for i := 0; i < len(data); i += 2 {
		if i+1 >= len(data) {
			throw("java.lang.ArrayIndexOutOfBoundsException", "Index "+itoa(i+1)+" out of bounds for length "+
				itoa(len(data)))
		}
		m.points = append(m.points, geom.NewXPoint2D(float64(data[i]), float64(data[i+1])))
	}
	return m
}

// NewUMotif returns the motif encoded by s: each character is a move of the
// previous point (see UMotifConvertPoint).
func NewUMotif(s string) *UMotif {
	m := &UMotif{}
	last := geom.NewXPoint2D(0, 0)
	for _, c := range s {
		read := UMotifConvertPoint(c)
		last = last.MovePoint(read)
		m.points = append(m.points, last)
	}
	return m
}

// Length returns the distance between the first and last points.
func (m *UMotif) Length() float64 { return m.points[0].Distance(m.points[len(m.points)-1]) }

// Points returns the points.
func (m *UMotif) Points() []geom.XPoint2D { return m.points }

// Rectangle returns the motif repeated along the top and right sides of a
// rectangle.
func (m *UMotif) Rectangle(width, height float64) *DotPath {
	l := m.Length()
	nb1 := javaD2I(width / l)
	h1 := m.drawHorizontalPath(0, 0, nb1)
	nb2 := javaD2I(height / l)
	v1 := m.drawVerticalPath(h1.EndPoint().X, h1.EndPoint().Y, nb2)
	return h1.AddAfterPath(v1)
}

// UMotifConvertPoint decodes a character of a motif: 'A' to 'Z' then 'a' to
// 'w' are the 49 moves of a 7x7 grid centered on 0. Other characters panic,
// as upstream throws IllegalArgumentException.
func UMotifConvertPoint(c rune) geom.XPoint2D {
	v := UMotifConvertFromChar(c)
	x := v % 7
	y := v / 7
	return geom.NewXPoint2D(float64(x-3), float64(y-3))
}

// UMotifConvertFromChar returns the index of a motif character.
func UMotifConvertFromChar(c rune) int {
	if c >= 'A' && c <= 'Z' {
		return int(c - 'A')
	}
	if c >= 'a' && c <= 'w' {
		return int(c-'a') + 26
	}
	throw(illegalArgument, "")
	return 0
}

// DrawHorizontal draws the motif nb times horizontally from (x, y).
func (m *UMotif) DrawHorizontal(ug UGraphic, x, y float64, nb int) {
	ug.Draw(m.drawHorizontalPath(x, y, nb))
}

// DrawVertical draws the motif nb times vertically from (x, y).
func (m *UMotif) DrawVertical(ug UGraphic, x, y float64, nb int) {
	ug.Draw(m.drawVerticalPath(x, y, nb))
}

func (m *UMotif) drawHorizontalPath(x, y float64, nb int) *DotPath {
	path := NewDotPath()
	for i := 0; i < nb; i++ {
		path = m.addHorizontal(x, y, path)
		x = path.EndPoint().X
		y = path.EndPoint().Y
	}
	return path
}

func (m *UMotif) drawVerticalPath(x, y float64, nb int) *DotPath {
	path := NewDotPath()
	for i := 0; i < nb; i++ {
		path = m.addVertical(x, y, path)
		x = path.EndPoint().X
		y = path.EndPoint().Y
	}
	return path
}

func (m *UMotif) addHorizontal(x, y float64, path *DotPath) *DotPath {
	lastx, lasty := 0.0, 0.0
	for _, p := range m.points {
		x1 := lastx + x
		y1 := lasty + y
		x2 := p.X + x
		y2 := p.Y + y
		path = path.AddAfter(geom.NewXCubicCurve2D(x1, y1, x1, y1, x2, y2, x2, y2))
		lastx = p.X
		lasty = p.Y
	}
	return path
}

func (m *UMotif) addVertical(x, y float64, path *DotPath) *DotPath {
	lastx, lasty := 0.0, 0.0
	for _, p := range m.points {
		x1 := lastx + x
		y1 := lasty + y
		x2 := p.Y + x
		y2 := p.X + y
		path = path.AddAfter(geom.NewXCubicCurve2D(x1, y1, x1, y1, x2, y2, x2, y2))
		lastx = p.Y
		lasty = p.X
	}
	return path
}
