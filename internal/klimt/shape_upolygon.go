// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/UPolygon.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// UPolygon is a closed polygon. Its bounds follow the added points.
type UPolygon struct {
	AbstractShadowable
	all             []geom.XPoint2D
	name            string
	minmax          geom.MinMax
	compressionMode *CompressionMode
}

var _ Shadowable = (*UPolygon)(nil)

// NewUPolygon returns an empty polygon.
func NewUPolygon() *UPolygon { return NewUPolygonNamed("") }

// NewUPolygonPoints returns the polygon of points.
func NewUPolygonPoints(points []geom.XPoint2D) *UPolygon {
	p := NewUPolygon()
	p.all = append(p.all, points...)
	for _, pt := range p.all {
		p.manageMinMax(pt.X, pt.Y)
	}
	return p
}

// NewUPolygonNamed returns an empty polygon whose String is name; an empty
// name stands for upstream's null.
func NewUPolygonNamed(name string) *UPolygon {
	return &UPolygon{name: name, minmax: geom.MinMaxEmpty(false)}
}

// Point returns the point at idx.
func (p *UPolygon) Point(idx int) geom.XPoint2D { return p.all[idx] }

// CheckMiddleContactForSpecificTriangle returns the point before the first
// side whose middle is closer than 1 to center; ok is false when there is
// none (null upstream).
func (p *UPolygon) CheckMiddleContactForSpecificTriangle(center geom.XPoint2D) (pt geom.XPoint2D, ok bool) {
	for i := 0; i < len(p.all)-1; i++ {
		pt1 := p.all[i]
		pt2 := p.all[i+1]
		middle := geom.NewXPoint2D((pt1.X+pt2.X)/2, (pt1.Y+pt2.Y)/2)
		delta := middle.Distance(center)
		if delta < 1 {
			return p.all[(i+len(p.all)-1)%len(p.all)], true
		}
	}
	return geom.XPoint2D{}, false
}

// AddPoint appends a point.
func (p *UPolygon) AddPoint(x, y float64) {
	p.all = append(p.all, geom.NewXPoint2D(x, y))
	p.manageMinMax(x, y)
}

// AddPointPt appends a point.
func (p *UPolygon) AddPointPt(point geom.XPoint2D) { p.AddPoint(point.X, point.Y) }

func (p *UPolygon) manageMinMax(x, y float64) { p.minmax = p.minmax.AddPointXY(x, y) }

// Points returns the points; as upstream, the slice is the polygon's own.
func (p *UPolygon) Points() []geom.XPoint2D { return p.all }

// Translate returns a new polygon moved by (dx, dy), without name and
// shadow.
func (p *UPolygon) Translate(dx, dy float64) *UPolygon {
	result := NewUPolygon()
	for _, pt := range p.all {
		result.AddPoint(pt.X+dx, pt.Y+dy)
	}
	return result
}

// Rotate rotates the points in place by theta radians. As upstream, the
// bounds are not updated.
func (p *UPolygon) Rotate(theta float64) {
	if theta == 0 {
		return
	}
	p.Affine(XAffineTransformRotateInstance(theta))
}

// Affine transforms the points in place. As upstream, the bounds are not
// updated.
func (p *UPolygon) Affine(rotate *XAffineTransform) {
	for i := range p.all {
		p.all[i] = rotate.Transform(p.all[i])
	}
}

// String returns the name, or the points when there is none (upstream adds
// Java's identity hash before them).
func (p *UPolygon) String() string {
	if p.name != "" {
		return p.name
	}
	parts := make([]string, len(p.all))
	for i, pt := range p.all {
		parts[i] = pt.String()
	}
	return "UPolygon [" + strings.Join(parts, ", ") + "]"
}

// Height returns the height of the bounds.
func (p *UPolygon) Height() float64 { return p.minmax.Height() }

// Width returns the width of the bounds.
func (p *UPolygon) Width() float64 { return p.minmax.Width() }

// MinX returns the smallest x of the bounds.
func (p *UPolygon) MinX() float64 { return p.minmax.MinX() }

// MinY returns the smallest y of the bounds.
func (p *UPolygon) MinY() float64 { return p.minmax.MinY() }

// MaxX returns the largest x of the bounds.
func (p *UPolygon) MaxX() float64 { return p.minmax.MaxX() }

// MaxY returns the largest y of the bounds.
func (p *UPolygon) MaxY() float64 { return p.minmax.MaxY() }

// MinMax returns the bounds.
func (p *UPolygon) MinMax() geom.MinMax { return p.minmax }

// PointArray returns the coordinates moved by (x, y), as x0, y0, x1, ...
func (p *UPolygon) PointArray(x, y float64) []float64 {
	points := make([]float64, 0, len(p.all)*2)
	for _, pt := range p.all {
		points = append(points, pt.X+x, pt.Y+y)
	}
	return points
}

// CompressionMode returns the compression mode, nil for none.
func (p *UPolygon) CompressionMode() *CompressionMode { return p.compressionMode }

// SetCompressionMode sets the compression mode.
func (p *UPolygon) SetCompressionMode(mode *CompressionMode) { p.compressionMode = mode }
