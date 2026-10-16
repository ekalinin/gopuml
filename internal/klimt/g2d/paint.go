// SPDX-License-Identifier: MIT
// Provenance: original

package g2d

import "math"

// paint is a java.awt.Paint: the non-premultiplied ARGB color of each
// device pixel.
type paint interface {
	colorAt(x, y int) uint32
}

// solidPaint is a java.awt.Color.
type solidPaint uint32

func (c solidPaint) colorAt(int, int) uint32 { return uint32(c) }

// inverse returns the inverse of a transform; a singular transform gives
// the zero transform.
func (a affine) inverse() affine {
	det := float64(a.m00*a.m11) - float64(a.m01*a.m10)
	if det == 0 || math.IsNaN(det) {
		return affine{}
	}
	return affine{
		m00: a.m11 / det, m10: -a.m10 / det,
		m01: -a.m01 / det, m11: a.m00 / det,
		m02: (float64(a.m01*a.m12) - float64(a.m11*a.m02)) / det,
		m12: (float64(a.m10*a.m02) - float64(a.m00*a.m12)) / det,
	}
}

// gradientPaint is an acyclic java.awt.GradientPaint between two points in
// user space (stored as float32, as Point2D.Float). As Java's
// GradientPaintContext, the color of a device pixel is read at its top-left
// corner from a table of 257 colors interpolated in 256 steps.
type gradientPaint struct {
	x1, y1, dx, dy float64 // device position of the first point, and steps per device pixel
	interp         [257]uint32
	flat           bool
}

func newGradientPaint(x1, y1 float32, c1 uint32, x2, y2 float32, c2 uint32, tx affine) *gradientPaint {
	g := &gradientPaint{}
	inv := tx.inverse()
	xvx, xvy := inv.delta(1, 0)
	yvx, yvy := inv.delta(0, 1)
	px1, py1, px2, py2 := float64(x1), float64(y1), float64(x2), float64(y2)
	udx, udy := px2-px1, py2-py1
	ulenSq := float64(udx*udx) + float64(udy*udy)
	if ulenSq <= math.SmallestNonzeroFloat64 {
		g.flat = true
	} else {
		g.dx = (float64(xvx*udx) + float64(xvy*udy)) / ulenSq
		g.dy = (float64(yvx*udx) + float64(yvy*udy)) / ulenSq
		if g.dx < 0 {
			px1, py1 = px2, py2
			c1, c2 = c2, c1
			g.dx, g.dy = -g.dx, -g.dy
		}
	}
	g.x1, g.y1 = tx.apply(px1, py1)
	a1, r1, gr1, b1 := int32(c1>>24), int32(c1>>16&0xFF), int32(c1>>8&0xFF), int32(c1&0xFF)
	a2, r2, gr2, b2 := int32(c2>>24), int32(c2>>16&0xFF), int32(c2>>8&0xFF), int32(c2&0xFF)
	lerp := func(v1, v2 int32, rel float32) uint32 {
		return uint32(int32(float32(v1) + float32(float32(v2-v1)*rel)))
	}
	for i := range g.interp {
		rel := float32(i) / 256
		g.interp[i] = lerp(a1, a2, rel)<<24 | lerp(r1, r2, rel)<<16 | lerp(gr1, gr2, rel)<<8 | lerp(b1, b2, rel)
	}
	return g
}

func (g *gradientPaint) colorAt(x, y int) uint32 {
	if g.flat {
		return g.interp[0]
	}
	rel := float64((float64(x)-g.x1)*g.dx) + float64((float64(y)-g.y1)*g.dy)
	switch {
	case rel <= 0:
		return g.interp[0]
	case rel >= 1:
		return g.interp[256]
	}
	return g.interp[int(rel*256)]
}

// The cycle methods of java.awt.MultipleGradientPaint.
const (
	cycleNone = iota
	cycleReflect
	cycleRepeat
)

// linearGradientPaint is a java.awt.LinearGradientPaint as Java's
// MultipleGradientPaintContext renders it: stops completed to 0 and 1 with
// the end colors, colors interpolated in sRGB into lookup tables (one table
// sized after the smallest interval, or one table of 256 colors per interval
// when that one would be too large), and the gradient position of each
// device pixel computed in float32 from the inverse transform.
type linearGradientPaint struct {
	dgdX, dgdY, gc float32
	cycle          int
	fractions      []float32
	intervals      []float32
	simple         []uint32   // the single table, nil for per-interval tables
	tables         [][]uint32 // the per-interval tables
	// The position of the last pixel computed, to continue a row.
	cachePos                  float32
	cacheX, cacheY, cacheTile int
	cacheOK                   bool
}

// tiledPaint is a paint whose colors depend on where the tile they are
// computed in starts, as Java computes paint rasters tile by tile.
type tiledPaint interface {
	colorAtFrom(x, y, tileX int) uint32
}

// tileWidth is the width of the tiles of Java's antialiased renderer.
const tileWidth = 128

const (
	gradientSize         = 256
	gradientSizeIndex    = gradientSize - 1
	maxGradientArraySize = 5000
)

// interpolate fills out with the colors from rgb1 toward rgb2, as
// MultipleGradientPaintContext.interpolate.
func interpolate(rgb1, rgb2 uint32, out []uint32) {
	step := float32(1) / float32(len(out))
	ch := func(v uint32, sh uint) int32 { return int32(v>>sh) & 0xFF }
	a1, r1, g1, b1 := ch(rgb1, 24), ch(rgb1, 16), ch(rgb1, 8), ch(rgb1, 0)
	da, dr, dg, db := ch(rgb2, 24)-a1, ch(rgb2, 16)-r1, ch(rgb2, 8)-g1, ch(rgb2, 0)-b1
	mix := func(c1, d int32, i int) uint32 {
		return uint32(int32(float32(c1)+float32(float32(int32(i)*d)*step)+0.5)) & 0xFF
	}
	for i := range out {
		out[i] = mix(a1, da, i)<<24 | mix(r1, dr, i)<<16 | mix(g1, dg, i)<<8 | mix(b1, db, i)
	}
}

func newLinearGradientPaint(sx, sy, ex, ey float32, fractions []float32, colors []uint32, cycle int,
	tx affine) *linearGradientPaint {
	g := &linearGradientPaint{cycle: cycle}
	// Complete the stops to 0 and 1 with the end colors.
	if fractions[0] != 0 {
		fractions = append([]float32{0}, fractions...)
		colors = append([]uint32{colors[0]}, colors...)
	}
	if fractions[len(fractions)-1] != 1 {
		fractions = append(fractions, 1)
		colors = append(colors, colors[len(colors)-1])
	}
	g.fractions = fractions
	g.intervals = make([]float32, len(fractions)-1)
	imin := float32(1)
	for i := range g.intervals {
		g.intervals[i] = fractions[i+1] - fractions[i]
		imin = min(imin, g.intervals[i])
	}
	estimated := 0
	for _, iv := range g.intervals {
		estimated += int(float32(iv/imin) * gradientSize)
	}
	if estimated > maxGradientArraySize {
		g.tables = make([][]uint32, len(g.intervals))
		for i := range g.intervals {
			g.tables[i] = make([]uint32, gradientSize)
			interpolate(colors[i], colors[i+1], g.tables[i])
		}
	} else {
		total := 1
		parts := make([][]uint32, len(g.intervals))
		for i, iv := range g.intervals {
			parts[i] = make([]uint32, int(float32(iv/imin)*255))
			interpolate(colors[i], colors[i+1], parts[i])
			total += len(parts[i])
		}
		g.simple = make([]uint32, 0, total)
		for _, p := range parts {
			g.simple = append(g.simple, p...)
		}
		g.simple = append(g.simple, colors[len(colors)-1])
	}
	inv := tx.inverse()
	a00, a10, a01, a11 := float32(inv.m00), float32(inv.m10), float32(inv.m01), float32(inv.m11)
	a02, a12 := float32(inv.m02), float32(inv.m12)
	dx, dy := ex-sx, ey-sy
	dSq := float32(dx*dx) + float32(dy*dy)
	constX, constY := dx/dSq, dy/dSq
	g.dgdX = float32(a00*constX) + float32(a10*constY)
	g.dgdY = float32(a01*constX) + float32(a11*constY)
	g.gc = float32((a02-sx)*constX) + float32((a12-sy)*constY)
	return g
}

func (g *linearGradientPaint) colorAt(x, y int) uint32 { return g.colorAtFrom(x, y, x) }

// colorAtFrom returns the color of (x, y) in a tile that starts at column
// tileX: Java computes the position at the start of each row of a tile and
// adds the step of a column for each pixel, in float32.
func (g *linearGradientPaint) colorAtFrom(x, y, tileX int) uint32 {
	if !g.cacheOK || g.cacheY != y || g.cacheTile != tileX || g.cacheX > x {
		g.cachePos = float32(float32(g.dgdX*float32(tileX))+g.gc) + float32(g.dgdY*float32(y))
		g.cacheX, g.cacheY, g.cacheTile, g.cacheOK = tileX, y, tileX, true
	}
	for g.cacheX < x {
		g.cachePos += g.dgdX
		g.cacheX++
	}
	return g.lookup(g.cachePos)
}

func (g *linearGradientPaint) lookup(pos float32) uint32 {
	switch g.cycle {
	case cycleRepeat:
		pos -= float32(int32(pos))
		if pos < 0 {
			pos++
		}
	case cycleReflect:
		if pos < 0 {
			pos = -pos
		}
		part := int32(pos)
		pos -= float32(part)
		if part&1 == 1 {
			pos = 1 - pos
		}
	default:
		pos = max(0, min(1, pos))
	}
	if g.simple != nil {
		return g.simple[int(pos*float32(len(g.simple)-1))]
	}
	for i := range g.tables {
		if pos < g.fractions[i+1] {
			delta := pos - g.fractions[i]
			return g.tables[i][int(float32(delta/g.intervals[i])*gradientSizeIndex)]
		}
	}
	return g.tables[len(g.tables)-1][gradientSizeIndex]
}

// texturePaint is a java.awt.TexturePaint of a 4x4 tile anchored at the
// user space origin.
type texturePaint struct {
	tile [16]uint32
	inv  affine
}

func (t *texturePaint) colorAt(x, y int) uint32 {
	ux, uy := t.inv.apply(float64(x)+0.5, float64(y)+0.5)
	i := int(math.Floor(ux)) & 3
	j := int(math.Floor(uy)) & 3
	return t.tile[j*4+i]
}
