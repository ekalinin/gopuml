// SPDX-License-Identifier: MIT
// Provenance: original

package g2d

import (
	"math"
)

// graphics emulates the java.awt.Graphics2D of a BufferedImage, for the
// operations upstream's g2d drivers use. It follows the pipelines the
// reference JVM (OpenJDK 21, Marlin renderer) chooses with the default
// stroke control (normalization on):
//
//   - fill of a Rectangle2D: the exact transformed rectangle, with exact
//     area coverage for a color and Marlin sampling for other paints;
//   - draw of a Rectangle2D (miter joins, no dashes) and of a Line2D (no
//     round caps, no dashes): parallelograms whose corners are snapped to
//     floor(v + 0.001) + 0.499, square caps extending lines by half the
//     width;
//   - other fills: the shape as is, sampled by Marlin;
//   - other strokes: the path snapped to pixel centers (floor(v) + 0.5,
//     control points moved with their end points), dashed, stroked in
//     device space with the scaled width, sampled by Marlin;
//   - without antialiasing: one sample per pixel center, paths snapped to
//     floor(v + 0.25) + 0.25, strokes at least one pixel wide.
//
// It is not safe for concurrent use; each render has its own.
type graphics struct {
	im       *Image
	tx       affine
	bounds   irect // the image
	clipRect irect
	clipMask []bool // nil, or a pixel mask over the image (shape clips)
	paint    paint
	color    uint32
	stroke   basicStroke
	aa       bool
	// fracMetrics is the fractional metrics hint, off by default.
	fracMetrics bool
	glyphs      glyphCache
}

// flatTol is the flattening tolerance of curves, in device pixels.
const flatTol = 0.02

// minPenAA is the narrowest stroke Marlin draws with antialiasing, in
// device pixels.
const minPenAA = 1.0 / 8

func newGraphics(im *Image) *graphics {
	b := irect{0, 0, im.width, im.height}
	return &graphics{im: im, tx: identity, bounds: b, clipRect: b, paint: solidPaint(0xFFFFFFFF),
		color: 0xFFFFFFFF, stroke: newBasicStroke(1)}
}

func (g *graphics) setColor(argb uint32) {
	g.color = argb
	g.paint = solidPaint(argb)
}

func (g *graphics) setPaint(p paint) { g.paint = p }

func (g *graphics) setStroke(s basicStroke) { g.stroke = s }

// setClip sets the clip to a rectangle in user space, nil for none. As
// Java's clip regions, the transformed rectangle is rounded to whole
// pixels with ceil(v - 0.5).
func (g *graphics) setClip(r *rect2D) {
	g.clipMask = nil
	if r == nil {
		g.clipRect = g.bounds
		return
	}
	x0, y0 := g.tx.apply(r.x, r.y)
	x1, y1 := g.tx.apply(r.x+r.w, r.y+r.h)
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	round := func(v float64) int {
		v = math.Ceil(v - 0.5)
		return int(math.Max(-1<<30, math.Min(1<<30, v)))
	}
	g.clipRect = irect{round(x0), round(y0), round(x1), round(y1)}.intersect(g.bounds)
}

// setClipShapeMinus sets the clip to a device rectangle minus a device
// shape, as java.awt.geom.Area does; Java clips to the pixels whose centers
// are inside.
func (g *graphics) setClipShapeMinus(full rect2D, minus *path) {
	w, h := g.im.width, g.im.height
	m := make([]bool, w*h)
	inFull := rasterSampled(polyEdges(flatten(full.path(), flatTol)), g.bounds, false, 1, 1)
	inMinus := rasterSampled(polyEdges(flatten(minus, flatTol)), g.bounds, minus.evenOdd, 1, 1)
	r := g.bounds
	if inFull != nil {
		r = inFull.r
	}
	g.clipRect = r
	for y := r.y0; y < r.y1; y++ {
		for x := r.x0; x < r.x1; x++ {
			in := inFull != nil && inFull.at(x, y) != 0
			if in && inMinus != nil && x >= inMinus.r.x0 && x < inMinus.r.x1 && y >= inMinus.r.y0 &&
				y < inMinus.r.y1 && inMinus.at(x, y) != 0 {
				in = false
			}
			m[y*w+x] = in
		}
	}
	g.clipMask = m
}

// clipState saves the clip.
type clipState struct {
	r irect
	m []bool
}

func (g *graphics) saveClip() clipState     { return clipState{g.clipRect, g.clipMask} }
func (g *graphics) restoreClip(s clipState) { g.clipRect, g.clipMask = s.r, s.m }

// composite blends a coverage mask with the current paint.
func (g *graphics) composite(m *mask) {
	if m == nil {
		return
	}
	r := m.r.intersect(g.clipRect)
	w := m.r.x1 - m.r.x0
	solid, isSolid := g.paint.(solidPaint)
	tiled, isTiled := g.paint.(tiledPaint)
	for y := r.y0; y < r.y1; y++ {
		row := (y - m.r.y0) * w
		for x := r.x0; x < r.x1; x++ {
			a := m.a[row+x-m.r.x0]
			if a == 0 {
				continue
			}
			i := y*g.im.width + x
			if g.clipMask != nil && !g.clipMask[i] {
				continue
			}
			var src uint32
			switch {
			case isSolid:
				src = uint32(solid)
			case isTiled:
				src = tiled.colorAtFrom(x, y, m.r.x0+(x-m.r.x0)/tileWidth*tileWidth)
			default:
				src = g.paint.colorAt(x, y)
			}
			g.im.blend(i, src, a)
		}
	}
}

// fillPolygonsExact fills device polygons with exact coverage (the
// parallelogram loops of a color paint), or Marlin sampling for other
// paints.
func (g *graphics) fillParallelograms(polys [][]float64, evenOdd bool) {
	edges := polyEdges(polys)
	if _, ok := g.paint.(solidPaint); ok {
		g.composite(rasterExact(edges, g.clipRect, evenOdd, alphaRound))
		return
	}
	g.composite(rasterSampled(edges, g.clipRect, evenOdd, marlinSubX, marlinSubY))
}

// reversed returns a polygon with its points in the other order.
func reversed(poly []float64) []float64 {
	n := len(poly) / 2
	out := make([]float64, len(poly))
	for i := range n {
		out[2*i], out[2*i+1] = poly[2*(n-1-i)], poly[2*(n-1-i)+1]
	}
	return out
}

// pgram returns the polygon of a device parallelogram.
func pgram(px, py, dx1, dy1, dx2, dy2 float64) []float64 {
	return []float64{px, py, px + dx1, py + dy1, px + dx1 + dx2, py + dy1 + dy2, px + dx2, py + dy2}
}

// fill fills a shape with the current paint.
func (g *graphics) fill(s shape) {
	if !g.aa {
		g.fillNonAA(s)
		return
	}
	if r, ok := s.(rect2D); ok {
		if r.w < 0 || r.h < 0 {
			return
		}
		px, py := g.tx.apply(r.x, r.y)
		dx1, dy1 := g.tx.delta(r.w, 0)
		dx2, dy2 := g.tx.delta(0, r.h)
		g.fillParallelograms([][]float64{pgram(px, py, dx1, dy1, dx2, dy2)}, false)
		return
	}
	p := s.path().transformed(g.tx)
	g.composite(rasterSampled(polyEdges(flatten(p, flatTol)), g.clipRect, p.evenOdd, marlinSubX, marlinSubY))
}

// fillRect is Graphics.fillRect.
func (g *graphics) fillRect(x, y, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	if g.tx.isIdentity() {
		r := irect{x, y, x + w, y + h}.intersect(g.clipRect)
		solid, isSolid := g.paint.(solidPaint)
		for py := r.y0; py < r.y1; py++ {
			for px := r.x0; px < r.x1; px++ {
				i := py*g.im.width + px
				if g.clipMask != nil && !g.clipMask[i] {
					continue
				}
				src := uint32(solid)
				if !isSolid {
					src = g.paint.colorAt(px, py)
				}
				g.im.blend(i, src, 0xFF)
			}
		}
		return
	}
	g.fill(rect2D{float64(x), float64(y), float64(w), float64(h)})
}

// normP2P snaps a device coordinate as the antialiased parallelogram
// pipeline does.
func normP2P(v float64) float64 { return math.Floor(v+0.001) + 0.499 }

// draw strokes a shape with the current stroke and paint.
func (g *graphics) draw(s shape) {
	bs := g.stroke
	if !g.aa {
		g.drawNonAA(s)
		return
	}
	switch v := s.(type) {
	case rect2D:
		if bs.join == joinMiter && len(bs.dash) == 0 {
			if v.w >= 0 && v.h >= 0 {
				g.drawRectangle(v, bs.width)
			}
			return
		}
	case line2D:
		if bs.cap != capRound && len(bs.dash) == 0 {
			g.drawGeneralLine(v.x1, v.y1, v.x2, v.y2)
			return
		}
	}
	g.strokeShape(s.path(), bs, true)
}

// drawRectangle draws the outline of a rectangle as two parallelograms.
func (g *graphics) drawRectangle(r rect2D, lw float64) {
	t := g.tx
	px, py := t.apply(r.x, r.y)
	lw1 := math.Hypot(t.m00, t.m10) * lw
	lw2 := math.Hypot(t.m01, t.m11) * lw
	dx1, dy1 := float64(t.m00*r.w), float64(t.m10*r.w)
	dx2, dy2 := float64(t.m01*r.h), float64(t.m11*r.h)
	nx, ny := normP2P(px), normP2P(py)
	dx1, dy1 = normP2P(px+dx1)-nx, normP2P(py+dy1)-ny
	dx2, dy2 = normP2P(px+dx2)-nx, normP2P(py+dy2)-ny
	px, py = nx, ny
	lw1 = math.Max(lw1, minPenAA)
	lw2 = math.Max(lw2, minPenAA)
	len1 := math.Hypot(dx1, dy1)
	len2 := math.Hypot(dx2, dy2)
	if lw1 >= len1 || lw2 >= len2 {
		g.fillOuterParallelogram(px, py, dx1, dy1, dx2, dy2, len1, len2, lw1, lw2)
		return
	}
	f1, f2 := lw1/len1, lw2/len2
	ldx1, ldy1 := float64(dx1*f1), float64(dy1*f1)
	ldx2, ldy2 := float64(dx2*f2), float64(dy2*f2)
	ox, oy := px-float64((ldx1+ldx2)/2), py-float64((ldy1+ldy2)/2)
	odx1, ody1, odx2, ody2 := dx1+ldx1, dy1+ldy1, dx2+ldx2, dy2+ldy2
	polys := [][]float64{pgram(ox, oy, odx1, ody1, odx2, ody2)}
	if f1 <= 1 || f2 <= 1 {
		// The hole, in the other direction: the non-zero rule then gives
		// the exact area of the ring, where even-odd area sums would not.
		ix, iy := ox+ldx1+ldx2, oy+ldy1+ldy2
		polys = append(polys, reversed(pgram(ix, iy, odx1-float64(2*ldx1), ody1-float64(2*ldy1), odx2-float64(2*ldx2),
			ody2-float64(2*ldy2))))
	}
	g.fillParallelograms(polys, false)
}

func (g *graphics) fillOuterParallelogram(px, py, dx1, dy1, dx2, dy2, len1, len2, lw1, lw2 float64) {
	var udx1, udy1, udx2, udy2 float64
	if len1 != 0 {
		udx1, udy1 = dx1/len1, dy1/len1
	}
	if len2 != 0 {
		udx2, udy2 = dx2/len2, dy2/len2
	}
	switch {
	case len1 == 0:
		if len2 == 0 {
			udx2, udy2 = 0, 1
		}
		udx1, udy1 = udy2, -udx2
	case len2 == 0:
		udx2, udy2 = udy1, -udx1
	}
	udx1 = float64(udx1 * lw1)
	udy1 = float64(udy1 * lw1)
	udx2 = float64(udx2 * lw2)
	udy2 = float64(udy2 * lw2)
	px -= float64((udx1 + udx2) / 2)
	py -= float64((udy1 + udy2) / 2)
	g.fillParallelograms([][]float64{pgram(px, py, dx1+udx1, dy1+udy1, dx2+udx2, dy2+udy2)}, false)
}

// drawGeneralLine draws a line as a parallelogram: the end points snapped,
// extended by half the width with square caps.
func (g *graphics) drawGeneralLine(ux1, uy1, ux2, uy2 float64) {
	bs := g.stroke
	lw := bs.width
	dx, dy := ux2-ux1, uy2-uy1
	x1, y1 := g.tx.apply(ux1, uy1)
	x2, y2 := g.tx.apply(ux2, uy2)
	x1, y1, x2, y2 = normP2P(x1), normP2P(y1), normP2P(x2), normP2P(y2)
	if !g.tx.isTranslation() {
		l := math.Hypot(dx, dy)
		if l == 0 {
			dx, l = 1, 1
		}
		ux, uy := g.tx.delta(dy/l, -dx/l)
		lw *= math.Hypot(ux, uy)
	}
	lw = math.Max(lw, minPenAA)
	dx, dy = x2-x1, y2-y1
	l := math.Hypot(dx, dy)
	var udx, udy float64
	if l == 0 {
		if bs.cap == capButt {
			return
		}
		udx, udy = lw, 0
	} else {
		udx, udy = float64(lw*dx/l), float64(lw*dy/l)
	}
	px, py := x1+float64(udy/2), y1-float64(udx/2)
	if bs.cap == capSquare {
		px -= float64(udx / 2)
		py -= float64(udy / 2)
		dx += udx
		dy += udy
	}
	g.fillParallelograms([][]float64{pgram(px, py, -udy, udx, dx, dy)}, false)
}

// drawLine is Graphics.drawLine.
func (g *graphics) drawLine(x1, y1, x2, y2 int) {
	g.draw(line2D{float64(x1), float64(y1), float64(x2), float64(y2)})
}

// strokeShape strokes a user space path with Marlin: transformed, snapped,
// dashed and stroked in device space (the transforms upstream uses scale
// uniformly), then sampled.
func (g *graphics) strokeShape(up *path, bs basicStroke, aa bool) {
	scale, uniform := g.tx.uniformScale()
	p := up.transformed(g.tx)
	if !uniform {
		// Not used by upstream: approximate with the mean scale.
		scale = math.Sqrt(math.Abs(float64(g.tx.m00*g.tx.m11) - float64(g.tx.m01*g.tx.m10)))
	}
	ds := bs
	ds.width = bs.width * scale
	if len(bs.dash) > 0 {
		ds.dash = make([]float64, len(bs.dash))
		for i, d := range bs.dash {
			ds.dash[i] = d * scale
		}
		ds.phase = bs.phase * scale
	}
	if aa {
		ds.width = math.Max(ds.width, minPenAA)
		p = normalizeCenter(p)
	} else {
		ds.width = math.Max(ds.width, 1)
		p = normalizeQuarter(p)
	}
	polys := strokeOutline(p, ds, flatTol)
	edges := polyEdges(polys)
	if aa {
		g.composite(rasterSampled(edges, g.clipRect, false, marlinSubX, marlinSubY))
	} else {
		g.composite(rasterSampled(edges, g.clipRect, false, 1, 1))
	}
}

// normalizeCenter snaps a device path as Marlin does with antialiasing:
// end points to the pixel centers, floor(v) + 0.5, control points moved by
// the moves of the end points they belong to.
func normalizeCenter(p *path) *path {
	return normalizePath(p, func(v float64) float64 { return math.Floor(v) + 0.5 })
}

// normalizeQuarter snaps a device path as Java does without antialiasing:
// floor(v + 0.25) + 0.25.
func normalizeQuarter(p *path) *path {
	return normalizePath(p, func(v float64) float64 { return math.Floor(v+0.25) + 0.25 })
}

func normalizePath(p *path, norm func(float64) float64) *path {
	q := &path{cmds: p.cmds, pts: make([]float64, len(p.pts)), evenOdd: p.evenOdd}
	copy(q.pts, p.pts)
	var curX, curY, movX, movY float64
	at := 0
	for _, c := range p.cmds {
		n := 2 * c.points()
		pts := q.pts[at : at+n]
		at += n
		if c == cmdClose {
			curX, curY = movX, movY
			continue
		}
		last := n - 2
		ax := norm(pts[last])
		ay := norm(pts[last+1])
		adjX, adjY := ax-pts[last], ay-pts[last+1]
		pts[last], pts[last+1] = ax, ay
		switch c {
		case cmdMove:
			movX, movY = adjX, adjY
		case cmdQuad:
			pts[0] += float64((curX + adjX) / 2)
			pts[1] += float64((curY + adjY) / 2)
		case cmdCubic:
			pts[0] += curX
			pts[1] += curY
			pts[2] += adjX
			pts[3] += adjY
		}
		curX, curY = adjX, adjY
	}
	return q
}

// drawImage draws an ARGB image with its top-left corner at (x, y) in user
// space. Under a translation (the transforms upstream uses around images,
// up to rounding) the pixels are copied; under another transform each
// device pixel takes the image pixel under its center (nearest neighbor,
// Java's default interpolation).
func (g *graphics) drawImage(src *Image, x, y int) {
	if math.Abs(g.tx.m00-1) > 1e-9 || math.Abs(g.tx.m11-1) > 1e-9 || g.tx.m01 != 0 || g.tx.m10 != 0 {
		g.drawImageTransformed(src, x, y)
		return
	}
	dx, dy := g.tx.apply(float64(x), float64(y))
	ox, oy := int(math.Floor(dx+0.5)), int(math.Floor(dy+0.5))
	r := irect{ox, oy, ox + src.width, oy + src.height}.intersect(g.clipRect)
	for py := r.y0; py < r.y1; py++ {
		for px := r.x0; px < r.x1; px++ {
			i := py*g.im.width + px
			if g.clipMask != nil && !g.clipMask[i] {
				continue
			}
			s := src.pix[(py-oy)*src.width+px-ox]
			if !src.hasAlpha {
				s |= 0xFF000000
			}
			g.im.blend(i, s, 0xFF)
		}
	}
}

// fillMask blends the current paint through a mask.
func (g *graphics) fillMask(m *mask) { g.composite(m) }

// normQuarter snaps a device coordinate as the parallelogram pipeline does
// without antialiasing.
func normQuarter(v float64) float64 { return math.Floor(v+0.25) + 0.25 }

// thin reports whether the stroke is one device pixel wide or less and not
// dashed: Java then draws lines pixel by pixel and fills paths with its
// path loops.
func (g *graphics) thin() bool {
	scale := math.Sqrt(math.Abs(float64(g.tx.m00*g.tx.m11) - float64(g.tx.m01*g.tx.m10)))
	return len(g.stroke.dash) == 0 && g.stroke.width*scale <= 1
}

// fillNonAA fills without antialiasing: a rectangle with its corners
// snapped to floor(v + 0.25) + 0.25 and the pixels whose centers are
// inside; with a thin stroke, other shapes as is and the pixels whose
// top-left corners are inside (Java's path fill loop); otherwise the shape
// snapped and the pixel centers.
func (g *graphics) fillNonAA(s shape) {
	if r, ok := s.(rect2D); ok {
		if r.w < 0 || r.h < 0 {
			return
		}
		px, py := g.tx.apply(r.x, r.y)
		dx1, dy1 := g.tx.delta(r.w, 0)
		dx2, dy2 := g.tx.delta(0, r.h)
		nx, ny := normQuarter(px), normQuarter(py)
		dx1, dy1 = normQuarter(px+dx1)-nx, normQuarter(py+dy1)-ny
		dx2, dy2 = normQuarter(px+dx2)-nx, normQuarter(py+dy2)-ny
		edges := polyEdges([][]float64{pgram(nx, ny, dx1, dy1, dx2, dy2)})
		g.composite(rasterSampled(edges, g.clipRect, false, 1, 1))
		return
	}
	p := s.path().transformed(g.tx)
	if g.thin() {
		p = p.transformed(identity.translate(0.5, 0.5))
	} else {
		p = normalizeQuarter(p)
	}
	g.composite(rasterSampled(polyEdges(flatten(p, flatTol)), g.clipRect, p.evenOdd, 1, 1))
}

// drawNonAA strokes without antialiasing.
func (g *graphics) drawNonAA(s shape) {
	bs := g.stroke
	thin := g.thin()
	switch v := s.(type) {
	case rect2D:
		if bs.join == joinMiter && len(bs.dash) == 0 {
			if v.w >= 0 && v.h >= 0 {
				g.drawRectangleNonAA(v, thin)
			}
			return
		}
	case line2D:
		if bs.cap != capRound && len(bs.dash) == 0 {
			if thin {
				x1, y1 := g.tx.apply(v.x1, v.y1)
				x2, y2 := g.tx.apply(v.x2, v.y2)
				g.drawPixelLine(int(math.Floor(x1)), int(math.Floor(y1)), int(math.Floor(x2)), int(math.Floor(y2)))
				return
			}
			g.strokeShape(v.path(), bs, false)
			return
		}
	}
	if thin {
		// Java's path loop draws the segments pixel by pixel.
		for _, poly := range flattenOpen(s.path().transformed(g.tx), flatTol) {
			for i := 0; i+3 < len(poly); i += 2 {
				g.drawPixelLine(int(math.Floor(poly[i])), int(math.Floor(poly[i+1])), int(math.Floor(poly[i+2])),
					int(math.Floor(poly[i+3])))
			}
		}
		return
	}
	g.strokeShape(s.path(), bs, false)
}

// drawRectangleNonAA draws a rectangle outline as the parallelogram
// pipeline does without antialiasing: corners snapped to
// floor(v + 0.25) + 0.25, at least one pixel wide.
func (g *graphics) drawRectangleNonAA(r rect2D, thin bool) {
	t := g.tx
	lw := g.stroke.width
	if thin {
		lw = 1
	}
	px, py := t.apply(r.x, r.y)
	lw1 := math.Hypot(t.m00, t.m10) * lw
	lw2 := math.Hypot(t.m01, t.m11) * lw
	dx1, dy1 := float64(t.m00*r.w), float64(t.m10*r.w)
	dx2, dy2 := float64(t.m01*r.h), float64(t.m11*r.h)
	nx, ny := normQuarter(px), normQuarter(py)
	dx1, dy1 = normQuarter(px+dx1)-nx, normQuarter(py+dy1)-ny
	dx2, dy2 = normQuarter(px+dx2)-nx, normQuarter(py+dy2)-ny
	lw1 = math.Max(lw1, 1)
	lw2 = math.Max(lw2, 1)
	len1 := math.Hypot(dx1, dy1)
	len2 := math.Hypot(dx2, dy2)
	var polys [][]float64
	if lw1 >= len1 || lw2 >= len2 {
		var udx1, udy1, udx2, udy2 float64
		if len1 != 0 {
			udx1, udy1 = dx1/len1, dy1/len1
		}
		if len2 != 0 {
			udx2, udy2 = dx2/len2, dy2/len2
		}
		udx1, udy1, udx2, udy2 = float64(udx1*lw1), float64(udy1*lw1), float64(udx2*lw2), float64(udy2*lw2)
		polys = [][]float64{pgram(nx-float64((udx1+udx2)/2), ny-float64((udy1+udy2)/2), dx1+udx1, dy1+udy1, dx2+udx2,
			dy2+udy2)}
	} else {
		f1, f2 := lw1/len1, lw2/len2
		ldx1, ldy1 := float64(dx1*f1), float64(dy1*f1)
		ldx2, ldy2 := float64(dx2*f2), float64(dy2*f2)
		ox, oy := nx-float64((ldx1+ldx2)/2), ny-float64((ldy1+ldy2)/2)
		odx1, ody1, odx2, ody2 := dx1+ldx1, dy1+ldy1, dx2+ldx2, dy2+ldy2
		polys = [][]float64{pgram(ox, oy, odx1, ody1, odx2, ody2),
			reversed(pgram(ox+ldx1+ldx2, oy+ldy1+ldy2, odx1-float64(2*ldx1), ody1-float64(2*ldy1),
				odx2-float64(2*ldx2), ody2-float64(2*ldy2)))}
	}
	g.composite(rasterSampled(polyEdges(polys), g.clipRect, false, 1, 1))
}

// drawPixelLine draws a one pixel line between two pixels, both included,
// stepping along the longer axis and rounding the other half up.
func (g *graphics) drawPixelLine(x1, y1, x2, y2 int) {
	plot := func(x, y int) {
		r := g.clipRect
		if x < r.x0 || y < r.y0 || x >= r.x1 || y >= r.y1 {
			return
		}
		i := y*g.im.width + x
		if g.clipMask != nil && !g.clipMask[i] {
			return
		}
		g.im.blend(i, g.paint.colorAt(x, y), 0xFF)
	}
	dx, dy := x2-x1, y2-y1
	adx, ady := dx, dy
	if adx < 0 {
		adx = -adx
	}
	if ady < 0 {
		ady = -ady
	}
	if adx == 0 && ady == 0 {
		plot(x1, y1)
		return
	}
	if adx >= ady {
		step := 1
		if dx < 0 {
			step = -1
		}
		for i := 0; i <= adx; i++ {
			// y = y1 + dy * i / adx, rounded half up.
			num := 2*dy*i + adx
			y := y1 + floorDiv(num, 2*adx)
			plot(x1+step*i, y)
		}
		return
	}
	step := 1
	if dy < 0 {
		step = -1
	}
	for i := 0; i <= ady; i++ {
		num := 2*dx*i + ady
		x := x1 + floorDiv(num, 2*ady)
		plot(x, y1+step*i)
	}
}

func floorDiv(a, b int) int {
	q := a / b
	if (a%b != 0) && ((a < 0) != (b < 0)) {
		q--
	}
	return q
}

// flattenOpen turns a path into polylines, one per subpath, closed only by
// a close segment.
func flattenOpen(p *path, tol float64) [][]float64 {
	var polys [][]float64
	var cur []float64
	var sx, sy, cx, cy float64
	flush := func() {
		if len(cur) >= 2 {
			polys = append(polys, cur)
		}
		cur = nil
	}
	at := 0
	for _, c := range p.cmds {
		pts := p.pts[at : at+2*c.points()]
		at += 2 * c.points()
		if c != cmdMove && c != cmdClose && cur == nil {
			cur = []float64{cx, cy}
		}
		switch c {
		case cmdMove:
			flush()
			sx, sy, cx, cy = pts[0], pts[1], pts[0], pts[1]
			cur = []float64{cx, cy}
		case cmdLine:
			cx, cy = pts[0], pts[1]
			cur = append(cur, cx, cy)
		case cmdQuad:
			cur = flattenQuad(cur, cx, cy, pts[0], pts[1], pts[2], pts[3], tol)
			cx, cy = pts[2], pts[3]
		case cmdCubic:
			cur = flattenCubic(cur, cx, cy, pts[0], pts[1], pts[2], pts[3], pts[4], pts[5], tol)
			cx, cy = pts[4], pts[5]
		case cmdClose:
			if cur != nil {
				cur = append(cur, sx, sy)
			}
			flush()
			cx, cy = sx, sy
		}
	}
	flush()
	return polys
}

// drawImageTransformed draws an image under a general transform with
// nearest neighbor sampling.
func (g *graphics) drawImageTransformed(src *Image, x, y int) {
	t := g.tx.translate(float64(x), float64(y))
	inv := t.inverse()
	var poly []float64
	for _, c := range [][2]float64{{0, 0}, {float64(src.width), 0}, {float64(src.width), float64(src.height)},
		{0, float64(src.height)}} {
		px, py := t.apply(c[0], c[1])
		poly = append(poly, px, py)
	}
	r := edgeBounds(polyEdges([][]float64{poly})).intersect(g.clipRect)
	for py := r.y0; py < r.y1; py++ {
		for px := r.x0; px < r.x1; px++ {
			sx, sy := inv.apply(float64(px)+0.5, float64(py)+0.5)
			if sx < 0 || sy < 0 || sx >= float64(src.width) || sy >= float64(src.height) {
				continue
			}
			i := py*g.im.width + px
			if g.clipMask != nil && !g.clipMask[i] {
				continue
			}
			s := src.pix[int(sy)*src.width+int(sx)]
			if !src.hasAlpha {
				s |= 0xFF000000
			}
			g.im.blend(i, s, 0xFF)
		}
	}
}
