// SPDX-License-Identifier: MIT
// Provenance: original

package g2d

import (
	"math"
	"slices"
)

// This file turns flattened paths into coverage masks, the way the Java2D
// pipelines upstream goes through do on the reference JVM:
//
//   - Marlin, the renderer of general shapes and strokes, samples each
//     pixel on 8 sub-scanlines (at their centers) with 256 sub-columns:
//     an edge crosses a sub-scanline at ceil(x - 0.5) sub-columns, and the
//     coverage of a pixel is the number of covered samples, mapped to
//     0..255 by (n * 255 + 512) / 2048. Without antialiasing there is one
//     sample at the pixel center.
//   - The parallelogram loops (rectangle fills and outlines, lines) and
//     FreeType (glyphs) compute the exact area covered in each pixel.
//
// The code is integer or float64 arithmetic with explicit float64
// conversions of products, so the masks do not depend on the architecture.

// edge is a line segment in device pixels.
type edge struct{ x0, y0, x1, y1 float64 }

// irect is a pixel rectangle [x0, x1) x [y0, y1).
type irect struct{ x0, y0, x1, y1 int }

func (r irect) empty() bool { return r.x0 >= r.x1 || r.y0 >= r.y1 }

func (r irect) intersect(o irect) irect {
	return irect{max(r.x0, o.x0), max(r.y0, o.y0), min(r.x1, o.x1), min(r.y1, o.y1)}
}

// mask is the coverage of a shape over a rectangle of the image: alpha
// values from 0 to 255, row by row.
type mask struct {
	r irect
	a []uint8
}

func (m *mask) at(x, y int) uint8 { return m.a[(y-m.r.y0)*(m.r.x1-m.r.x0)+x-m.r.x0] }

// edgeBounds returns the pixel rectangle that contains the edges.
func edgeBounds(edges []edge) irect {
	if len(edges) == 0 {
		return irect{}
	}
	x0, y0 := math.Inf(1), math.Inf(1)
	x1, y1 := math.Inf(-1), math.Inf(-1)
	for _, e := range edges {
		x0 = math.Min(x0, math.Min(e.x0, e.x1))
		x1 = math.Max(x1, math.Max(e.x0, e.x1))
		y0 = math.Min(y0, math.Min(e.y0, e.y1))
		y1 = math.Max(y1, math.Max(e.y0, e.y1))
	}
	if math.IsNaN(x0+x1+y0+y1) || math.IsInf(x0+x1+y0+y1, 0) {
		return irect{}
	}
	const limit = 1 << 28
	clampI := func(v float64) int { return int(math.Max(-limit, math.Min(limit, v))) }
	return irect{clampI(math.Floor(x0)), clampI(math.Floor(y0)), clampI(math.Ceil(x1)) + 1,
		clampI(math.Ceil(y1)) + 1}
}

// The sampling of Marlin with antialiasing.
const (
	marlinSubX = 256
	marlinSubY = 8
)

// rasterSampled computes the coverage of the polygons made of edges, within
// clip, by sampling subX x subY points per pixel as Marlin does.
func rasterSampled(edges []edge, clip irect, evenOdd bool, subX, subY int) *mask {
	r := edgeBounds(edges).intersect(clip)
	if r.empty() {
		return nil
	}
	w, h := r.x1-r.x0, r.y1-r.y0
	type crossing struct {
		x   int
		dir int
	}
	nsub := h * subY
	rows := make([][]crossing, nsub)
	sy0 := r.y0 * subY
	minX, maxX := r.x0*subX, r.x1*subX
	fx, fy := float64(subX), float64(subY)
	for _, e := range edges {
		ya, yb := float64(e.y0*fy), float64(e.y1*fy)
		if ya == yb {
			continue
		}
		xa, xb := float64(e.x0*fx), float64(e.x1*fx)
		dir := 1
		if ya > yb {
			ya, yb, xa, xb, dir = yb, ya, xb, xa, -1
		}
		first := int(math.Ceil(ya - 0.5))
		last := int(math.Ceil(yb - 0.5))
		first = max(first, sy0)
		last = min(last, sy0+nsub)
		slope := (xb - xa) / (yb - ya)
		for s := first; s < last; s++ {
			x := xa + float64((float64(s)+0.5-ya)*slope)
			xi := int(math.Ceil(x - 0.5))
			xi = min(max(xi, minX), maxX)
			rows[s-sy0] = append(rows[s-sy0], crossing{x: xi, dir: dir})
		}
	}
	m := &mask{r: r, a: make([]uint8, w*h)}
	d := make([]int, w+2)
	maxAlpha := subX * subY
	inside := func(wind int) bool {
		if evenOdd {
			return wind&1 != 0
		}
		return wind != 0
	}
	addSpan := func(xa, xb int) {
		if xa >= xb {
			return
		}
		xa -= minX
		xb -= minX
		pa, fa := xa/subX, xa%subX
		pb, fb := xb/subX, xb%subX
		if pa == pb {
			d[pa] += xb - xa
			d[pa+1] -= xb - xa
			return
		}
		d[pa] += subX - fa
		d[pa+1] -= subX - fa
		d[pa+1] += subX
		d[pb] -= subX
		if fb > 0 {
			d[pb] += fb
			d[pb+1] -= fb
		}
	}
	for row := range h {
		clear(d)
		any := false
		for k := range subY {
			cs := rows[row*subY+k]
			if len(cs) == 0 {
				continue
			}
			any = true
			slices.SortFunc(cs, func(a, b crossing) int { return a.x - b.x })
			wind, start := 0, 0
			for _, c := range cs {
				was := inside(wind)
				wind += c.dir
				now := inside(wind)
				switch {
				case !was && now:
					start = c.x
				case was && !now:
					addSpan(start, c.x)
				}
			}
		}
		if !any {
			continue
		}
		cov := 0
		line := m.a[row*w : (row+1)*w]
		for x := range w {
			cov += d[x]
			if cov > 0 {
				line[x] = uint8((cov*255 + maxAlpha/4) / maxAlpha)
			}
		}
	}
	return m
}

// clipEdgesX replaces the parts of the edges left of x0 and right of x1 by
// vertical edges on those lines. Coverage inside [x0, x1] is unchanged: it
// only depends on where edges cross each row left of a pixel.
func clipEdgesX(edges []edge, x0, x1 float64) []edge {
	out := make([]edge, 0, len(edges))
	var clip func(e edge, bound float64, left bool)
	add := func(e edge) {
		if e.x0 < x0 || e.x1 < x0 {
			clip(e, x0, true)
			return
		}
		if e.x0 > x1 || e.x1 > x1 {
			clip(e, x1, false)
			return
		}
		out = append(out, e)
	}
	clip = func(e edge, bound float64, left bool) {
		outside := func(x float64) bool {
			if left {
				return x < bound
			}
			return x > bound
		}
		o0, o1 := outside(e.x0), outside(e.x1)
		switch {
		case o0 && o1:
			out = append(out, edge{bound, e.y0, bound, e.y1})
		case o0 || o1:
			t := (bound - e.x0) / (e.x1 - e.x0)
			ym := e.y0 + float64(t*(e.y1-e.y0))
			if o0 {
				out = append(out, edge{bound, e.y0, bound, ym})
				add(edge{bound, ym, e.x1, e.y1})
			} else {
				add(edge{e.x0, e.y0, bound, ym})
				out = append(out, edge{bound, ym, bound, e.y1})
			}
		default:
			add(e)
		}
	}
	for _, e := range edges {
		add(e)
	}
	return out
}

// rasterExact computes the exact area of the polygons made of edges in each
// pixel within clip, and maps it to alpha with toAlpha.
func rasterExact(edges []edge, clip irect, evenOdd bool, toAlpha func(float64) uint8) *mask {
	r := edgeBounds(edges).intersect(clip)
	if r.empty() {
		return nil
	}
	w, h := r.x1-r.x0, r.y1-r.y0
	edges = clipEdgesX(edges, float64(r.x0), float64(r.x1))
	m := &mask{r: r, a: make([]uint8, w*h)}
	// Bands of rows bound the memory of the accumulation buffer.
	const band = 64
	acc := make([]float64, (w+2)*min(band, h))
	ox := float64(r.x0)
	for by := 0; by < h; by += band {
		bh := min(band, h-by)
		clear(acc)
		oy := float64(r.y0 + by)
		for _, e := range edges {
			if math.Max(e.y0, e.y1) <= oy || math.Min(e.y0, e.y1) >= oy+float64(bh) {
				continue
			}
			accumulate(acc, w+2, bh, e.x0-ox, e.y0-oy, e.x1-ox, e.y1-oy)
		}
		for y := range bh {
			var cov float64
			line := acc[y*(w+2):]
			out := m.a[(by+y)*w : (by+y+1)*w]
			for x := range w {
				cov += line[x]
				c := math.Abs(cov)
				if evenOdd {
					c = math.Mod(c, 2)
					if c > 1 {
						c = 2 - c
					}
				} else if c > 1 {
					c = 1
				}
				out[x] = toAlpha(c)
			}
		}
	}
	return m
}

// accumulate adds the signed area contributions of a line segment to acc,
// a buffer of h rows of stride values whose prefix sums along a row are the
// covered fractions of the pixels. x is within [0, stride-2].
func accumulate(acc []float64, stride, h int, x0, y0, x1, y1 float64) {
	if y0 == y1 {
		return
	}
	dir := 1.0
	if y0 > y1 {
		x0, y0, x1, y1, dir = x1, y1, x0, y0, -1
	}
	dxdy := (x1 - x0) / (y1 - y0)
	xmax := float64(stride - 2)
	x := x0
	ys := y0
	if ys < 0 {
		x -= float64(ys * dxdy)
		ys = 0
	}
	ye := math.Min(y1, float64(h))
	for y := int(math.Floor(ys)); float64(y) < ye; y++ {
		top := math.Max(float64(y), ys)
		bottom := math.Min(float64(y+1), ye)
		dy := bottom - top
		if dy <= 0 {
			continue
		}
		xnext := x + float64(dxdy*dy)
		d := float64(dy * dir)
		line := acc[y*stride : (y+1)*stride]
		// Interpolated positions may leave [0, w] by a rounding error.
		xa, xb := math.Max(0, math.Min(x, xmax)), math.Max(0, math.Min(xnext, xmax))
		if xa > xb {
			xa, xb = xb, xa
		}
		xaFloor := math.Floor(xa)
		xai := int(xaFloor)
		xbCeil := math.Ceil(xb)
		xbi := int(xbCeil)
		if xbi <= xai+1 {
			// Within one pixel: the part right of the edge, at its mean x.
			xmf := float64(0.5*(xa+xb)) - xaFloor
			line[xai] += d - float64(d*xmf)
			line[xai+1] += float64(d * xmf)
		} else {
			s := 1 / (xb - xa)
			xaf := xa - xaFloor
			a0 := float64(float64(0.5*s) * float64((1-xaf)*(1-xaf)))
			xbf := xb - xbCeil + 1
			am := float64(float64(0.5*s) * float64(xbf*xbf))
			line[xai] += float64(d * a0)
			if xbi == xai+2 {
				line[xai+1] += float64(d * (1 - a0 - am))
			} else {
				a1 := float64(s * (1.5 - xaf))
				line[xai+1] += float64(d * (a1 - a0))
				for xi := xai + 2; xi < xbi-1; xi++ {
					line[xi] += float64(d * s)
				}
				a2 := a1 + float64(float64(xbi-xai-3)*s)
				line[xbi-1] += float64(d * (1 - a2 - am))
			}
			line[xbi] += float64(d * am)
		}
		x = xnext
	}
}

// alphaRound maps a covered fraction to alpha by rounding, as the
// parallelogram loops do.
func alphaRound(c float64) uint8 { return uint8(math.Floor(float64(c*255) + 0.5)) }

// alphaFreeType maps a covered fraction to alpha as FreeType's gray
// rasterizer does: 256 levels, the full coverage clamped to 255.
func alphaFreeType(c float64) uint8 { return uint8(min(255, math.Floor(c*256))) }

// flatten turns a path into polygons (one per subpath, implicitly closed
// for filling), approximating curves with lines whose distance to the curve
// stays within tol pixels.
func flatten(p *path, tol float64) [][]float64 {
	var polys [][]float64
	var cur []float64
	var sx, sy, cx, cy float64
	flush := func() {
		if len(cur) >= 4 {
			polys = append(polys, cur)
		}
		cur = nil
	}
	at := 0
	for _, c := range p.cmds {
		pts := p.pts[at : at+2*c.points()]
		at += 2 * c.points()
		switch c {
		case cmdMove:
			flush()
			sx, sy, cx, cy = pts[0], pts[1], pts[0], pts[1]
			cur = []float64{cx, cy}
		case cmdLine:
			if cur == nil {
				cur = []float64{cx, cy}
			}
			cx, cy = pts[0], pts[1]
			cur = append(cur, cx, cy)
		case cmdQuad:
			if cur == nil {
				cur = []float64{cx, cy}
			}
			cur = flattenQuad(cur, cx, cy, pts[0], pts[1], pts[2], pts[3], tol)
			cx, cy = pts[2], pts[3]
		case cmdCubic:
			if cur == nil {
				cur = []float64{cx, cy}
			}
			cur = flattenCubic(cur, cx, cy, pts[0], pts[1], pts[2], pts[3], pts[4], pts[5], tol)
			cx, cy = pts[4], pts[5]
		case cmdClose:
			if cur != nil {
				cur = append(cur, sx, sy)
			}
			flush()
			// A segment after a close starts at the start of the subpath.
			cx, cy = sx, sy
		}
	}
	flush()
	return polys
}

// flattenQuad appends the points of a quadratic curve (without its start).
func flattenQuad(out []float64, x0, y0, x1, y1, x2, y2, tol float64) []float64 {
	ddx := x0 - 2*x1 + x2
	ddy := y0 - 2*y1 + y2
	dd := math.Sqrt(float64(ddx*ddx) + float64(ddy*ddy))
	n := int(math.Ceil(math.Sqrt(dd / (4 * tol))))
	n = max(1, min(n, 1000))
	for i := 1; i <= n; i++ {
		t := float64(i) / float64(n)
		if i == n {
			out = append(out, x2, y2)
			break
		}
		out = append(out, quadAt(x0, x1, x2, t), quadAt(y0, y1, y2, t))
	}
	return out
}

// flattenCubic appends the points of a cubic curve (without its start).
func flattenCubic(out []float64, x0, y0, x1, y1, x2, y2, x3, y3, tol float64) []float64 {
	d1x, d1y := x0-2*x1+x2, y0-2*y1+y2
	d2x, d2y := x1-2*x2+x3, y1-2*y2+y3
	dd := math.Max(math.Sqrt(float64(d1x*d1x)+float64(d1y*d1y)), math.Sqrt(float64(d2x*d2x)+float64(d2y*d2y)))
	n := int(math.Ceil(math.Sqrt(float64(0.75*dd) / tol)))
	n = max(1, min(n, 1000))
	for i := 1; i <= n; i++ {
		if i == n {
			out = append(out, x3, y3)
			break
		}
		t := float64(i) / float64(n)
		out = append(out, cubicAt(x0, x1, x2, x3, t), cubicAt(y0, y1, y2, y3, t))
	}
	return out
}

// polyEdges returns the edges of closed polygons.
func polyEdges(polys [][]float64) []edge {
	var edges []edge
	for _, poly := range polys {
		n := len(poly) / 2
		for i := range n {
			j := (i + 1) % n
			edges = append(edges, edge{poly[2*i], poly[2*i+1], poly[2*j], poly[2*j+1]})
		}
	}
	return edges
}
