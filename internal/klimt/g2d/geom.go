// SPDX-License-Identifier: MIT
// Provenance: original

package g2d

import (
	"math"

	"github.com/ekalinin/plantuml-in-go/internal/jmath"
)

// This file holds the geometry the drivers hand to the emulated Graphics2D:
// affine transforms and paths, and the paths of the java.awt.geom shapes
// upstream draws (rectangles, rounded rectangles, ellipses, arcs, lines and
// general paths), with the same segments, directions and control points as
// Java's path iterators, so that strokes start, join and dash at the same
// places.
//
// Products that feed an addition are written float64(a*b): the explicit
// conversion keeps the compiler from fusing them into one multiply-add,
// whose result depends on the architecture (design D17).

// affine is a java.awt.geom.AffineTransform:
//
//	[m00 m01 m02]
//	[m10 m11 m12]
type affine struct{ m00, m10, m01, m11, m02, m12 float64 }

var identity = affine{m00: 1, m11: 1}

// apply transforms a point.
func (a affine) apply(x, y float64) (float64, float64) {
	return float64(x*a.m00) + float64(y*a.m01) + a.m02, float64(x*a.m10) + float64(y*a.m11) + a.m12
}

// delta transforms a vector.
func (a affine) delta(x, y float64) (float64, float64) {
	return float64(x*a.m00) + float64(y*a.m01), float64(x*a.m10) + float64(y*a.m11)
}

// scale concatenates a scaling, as AffineTransform.scale.
func (a affine) scale(sx, sy float64) affine {
	a.m00 *= sx
	a.m10 *= sx
	a.m01 *= sy
	a.m11 *= sy
	return a
}

// translate concatenates a translation, as AffineTransform.translate.
func (a affine) translate(tx, ty float64) affine {
	a.m02 = float64(tx*a.m00) + float64(ty*a.m01) + a.m02
	a.m12 = float64(tx*a.m10) + float64(ty*a.m11) + a.m12
	return a
}

// rotate concatenates a rotation by theta radians. As AffineTransform.rotate,
// the quarter turns (whose sine is exactly 1 or -1) are exact.
func (a affine) rotate(theta float64) affine {
	sin := jmath.Sin(theta)
	switch sin {
	case 1:
		return affine{m00: a.m01, m10: a.m11, m01: -a.m00, m11: -a.m10, m02: a.m02, m12: a.m12}
	case -1:
		return affine{m00: -a.m01, m10: -a.m11, m01: a.m00, m11: a.m10, m02: a.m02, m12: a.m12}
	}
	cos := jmath.Cos(theta)
	if cos == 1 {
		return a
	}
	return affine{
		m00: float64(a.m00*cos) + float64(a.m01*sin), m01: float64(-a.m00*sin) + float64(a.m01*cos),
		m10: float64(a.m10*cos) + float64(a.m11*sin), m11: float64(-a.m10*sin) + float64(a.m11*cos),
		m02: a.m02, m12: a.m12,
	}
}

// isIdentity reports whether a is the identity.
func (a affine) isIdentity() bool { return a == identity }

// isTranslation reports whether a only translates.
func (a affine) isTranslation() bool { return a.m00 == 1 && a.m11 == 1 && a.m01 == 0 && a.m10 == 0 }

// uniformScale returns the scale of a transform that is a constant multiple
// of an orthogonal transform (Marlin then strokes in device space with a
// scaled width), and false for other transforms.
func (a affine) uniformScale() (float64, bool) {
	// Marlin's nearZero: |v| < 2 * ulp(v), which holds for zero and
	// subnormal values.
	near := func(v float64) bool {
		v = math.Abs(v)
		return v < 2*(math.Nextafter(v, math.Inf(1))-v)
	}
	if near(float64(a.m00*a.m01)+float64(a.m10*a.m11)) &&
		near(float64(a.m00*a.m00)+float64(a.m10*a.m10)-(float64(a.m01*a.m01)+float64(a.m11*a.m11))) {
		return math.Sqrt(float64(a.m00*a.m00) + float64(a.m10*a.m10)), true
	}
	return 0, false
}

// pathCmd is the kind of a path segment, with Java's PathIterator values.
type pathCmd uint8

const (
	cmdMove pathCmd = iota
	cmdLine
	cmdQuad
	cmdCubic
	cmdClose
)

// points returns the number of points of a segment.
func (c pathCmd) points() int {
	switch c {
	case cmdMove, cmdLine:
		return 1
	case cmdQuad:
		return 2
	case cmdCubic:
		return 3
	}
	return 0
}

// path is a sequence of segments, as a Java path iterator returns them.
type path struct {
	cmds    []pathCmd
	pts     []float64 // x, y of each point
	evenOdd bool      // WIND_EVEN_ODD, else WIND_NON_ZERO
}

func (p *path) moveTo(x, y float64) {
	p.cmds = append(p.cmds, cmdMove)
	p.pts = append(p.pts, x, y)
}

func (p *path) lineTo(x, y float64) {
	p.cmds = append(p.cmds, cmdLine)
	p.pts = append(p.pts, x, y)
}

func (p *path) quadTo(x1, y1, x2, y2 float64) {
	p.cmds = append(p.cmds, cmdQuad)
	p.pts = append(p.pts, x1, y1, x2, y2)
}

func (p *path) cubicTo(x1, y1, x2, y2, x3, y3 float64) {
	p.cmds = append(p.cmds, cmdCubic)
	p.pts = append(p.pts, x1, y1, x2, y2, x3, y3)
}

func (p *path) closePath() { p.cmds = append(p.cmds, cmdClose) }

// isEmpty reports whether the path has no segment.
func (p *path) isEmpty() bool { return len(p.cmds) == 0 }

// transformed returns the path with every point transformed.
func (p *path) transformed(a affine) *path {
	q := &path{cmds: p.cmds, pts: make([]float64, len(p.pts)), evenOdd: p.evenOdd}
	for i := 0; i < len(p.pts); i += 2 {
		q.pts[i], q.pts[i+1] = a.apply(p.pts[i], p.pts[i+1])
	}
	return q
}

// bounds returns the bounds of the end and control points.
func (p *path) bounds() (x0, y0, x1, y1 float64) {
	if len(p.pts) == 0 {
		return 0, 0, 0, 0
	}
	x0, y0, x1, y1 = p.pts[0], p.pts[1], p.pts[0], p.pts[1]
	for i := 2; i < len(p.pts); i += 2 {
		x0 = math.Min(x0, p.pts[i])
		x1 = math.Max(x1, p.pts[i])
		y0 = math.Min(y0, p.pts[i+1])
		y1 = math.Max(y1, p.pts[i+1])
	}
	return x0, y0, x1, y1
}

// shape is a java.awt.Shape: a path in user space and the bounds
// getBounds2D returns.
type shape interface {
	path() *path
	bounds2D() (x, y, w, h float64)
}

// rect2D is a Rectangle2D.Double.
type rect2D struct{ x, y, w, h float64 }

func (r rect2D) path() *path {
	p := &path{}
	if r.w < 0 || r.h < 0 {
		return p
	}
	p.moveTo(r.x, r.y)
	p.lineTo(r.x+r.w, r.y)
	p.lineTo(r.x+r.w, r.y+r.h)
	p.lineTo(r.x, r.y+r.h)
	p.lineTo(r.x, r.y)
	p.closePath()
	return p
}

func (r rect2D) bounds2D() (x, y, w, h float64) { return r.x, r.y, r.w, r.h }

// line2D is a Line2D.Double.
type line2D struct{ x1, y1, x2, y2 float64 }

func (l line2D) path() *path {
	p := &path{}
	p.moveTo(l.x1, l.y1)
	p.lineTo(l.x2, l.y2)
	return p
}

func (l line2D) bounds2D() (x, y, w, h float64) {
	x, y = math.Min(l.x1, l.x2), math.Min(l.y1, l.y2)
	return x, y, math.Max(l.x1, l.x2) - x, math.Max(l.y1, l.y2) - y
}

// kappa is the length of the control vectors of the cubic curve that
// approximates a quarter of a unit circle, 4/3 * (sqrt(2) - 1).
const kappa = 0.5522847498307933

// ellipse2D is an Ellipse2D.Double. Its path starts at the rightmost point
// and runs through the bottom, left and top points, a cubic curve for each
// quarter.
type ellipse2D struct{ x, y, w, h float64 }

func (e ellipse2D) path() *path {
	p := &path{}
	if e.w < 0 || e.h < 0 {
		return p
	}
	pcv := 0.5 + float64(kappa*0.5)
	ncv := 0.5 - float64(kappa*0.5)
	px := func(v float64) float64 { return e.x + float64(v*e.w) }
	py := func(v float64) float64 { return e.y + float64(v*e.h) }
	p.moveTo(px(1), py(0.5))
	p.cubicTo(px(1), py(pcv), px(pcv), py(1), px(0.5), py(1))
	p.cubicTo(px(ncv), py(1), px(0), py(pcv), px(0), py(0.5))
	p.cubicTo(px(0), py(ncv), px(ncv), py(0), px(0.5), py(0))
	p.cubicTo(px(pcv), py(0), px(1), py(ncv), px(1), py(0.5))
	p.closePath()
	return p
}

func (e ellipse2D) bounds2D() (x, y, w, h float64) { return e.x, e.y, e.w, e.h }

// roundRect2D is a RoundRectangle2D.Double: arcW and arcH are the width and
// height of the corner arcs (twice the radii). Its path starts on the left
// side below the top-left corner and runs counterclockwise on screen (down
// the left side first), a cubic curve for each corner.
type roundRect2D struct{ x, y, w, h, arcW, arcH float64 }

func (r roundRect2D) path() *path {
	p := &path{}
	if r.w < 0 || r.h < 0 {
		return p
	}
	aw := math.Min(r.w, math.Abs(r.arcW))
	ah := math.Min(r.h, math.Abs(r.arcH))
	// The control points sit at acv times the arc size from the corner.
	acv := (1 - kappa) / 2
	pt := func(v0, v1, v2, v3 float64) (float64, float64) {
		return r.x + float64(v0*r.w) + float64(v1*aw), r.y + float64(v2*r.h) + float64(v3*ah)
	}
	p.moveTo(pt(0, 0, 0, 0.5))
	p.lineTo(pt(0, 0, 1, -0.5))
	x1, y1 := pt(0, 0, 1, -acv)
	x2, y2 := pt(0, acv, 1, 0)
	x3, y3 := pt(0, 0.5, 1, 0)
	p.cubicTo(x1, y1, x2, y2, x3, y3)
	p.lineTo(pt(1, -0.5, 1, 0))
	x1, y1 = pt(1, -acv, 1, 0)
	x2, y2 = pt(1, 0, 1, -acv)
	x3, y3 = pt(1, 0, 1, -0.5)
	p.cubicTo(x1, y1, x2, y2, x3, y3)
	p.lineTo(pt(1, 0, 0, 0.5))
	x1, y1 = pt(1, 0, 0, acv)
	x2, y2 = pt(1, -acv, 0, 0)
	x3, y3 = pt(1, -0.5, 0, 0)
	p.cubicTo(x1, y1, x2, y2, x3, y3)
	p.lineTo(pt(0, 0.5, 0, 0))
	x1, y1 = pt(0, acv, 0, 0)
	x2, y2 = pt(0, 0, 0, acv)
	x3, y3 = pt(0, 0, 0, 0.5)
	p.cubicTo(x1, y1, x2, y2, x3, y3)
	p.closePath()
	return p
}

func (r roundRect2D) bounds2D() (x, y, w, h float64) { return r.x, r.y, r.w, r.h }

// arc2D is an open Arc2D.Double: the arc of the ellipse framed by x, y, w,
// h from start degrees over extent degrees (counterclockwise on screen for
// positive angles, as Java's arcs).
type arc2D struct{ x, y, w, h, start, extent float64 }

func (a arc2D) path() *path { return arcPath(a.x, a.y, a.w, a.h, a.start, a.extent, identity) }

// bounds2D returns the frame of the ellipse: upstream only asks the bounds
// of shapes it shadows, and arcs have no shadow.
func (a arc2D) bounds2D() (x, y, w, h float64) { return a.x, a.y, a.w, a.h }

// arcPath returns the open arc of the ellipse framed by x, y, w, h, from
// start over extent degrees, transformed by at, as Java's arc iterator
// gives it: a move to the start point, then one cubic curve per piece of at
// most 90 degrees, with control vectors of 4/3 * tan(piece / 4) of the
// radius.
func arcPath(x, y, w, h, start, extent float64, at affine) *path {
	p := &path{}
	if w < 0 || h < 0 {
		return p
	}
	rw, rh := float64(w/2), float64(h/2)
	cx, cy := x+rw, y+rh
	angle := -jmath.ToRadians(start)
	ext := -extent
	var segs int
	var increment, cv float64
	if ext >= 360 || ext <= -360 {
		segs = 4
		increment = math.Pi / 2
		cv = kappa
		if ext < 0 {
			increment = -increment
			cv = -cv
		}
	} else {
		segs = int(math.Ceil(math.Abs(ext) / 90))
		increment = jmath.ToRadians(ext / float64(segs))
		half := increment / 2
		cv = 4.0 / 3.0 * jmath.Sin(half) / (1 + jmath.Cos(half))
		if cv == 0 {
			segs = 0
		}
	}
	p.moveTo(at.apply(cx+float64(jmath.Cos(angle)*rw), cy+float64(jmath.Sin(angle)*rh)))
	for i := range segs {
		a0 := angle + float64(increment*float64(i))
		relx, rely := jmath.Cos(a0), jmath.Sin(a0)
		x1, y1 := at.apply(cx+float64((relx-float64(cv*rely))*rw), cy+float64((rely+float64(cv*relx))*rh))
		a1 := a0 + increment
		relx, rely = jmath.Cos(a1), jmath.Sin(a1)
		x2, y2 := at.apply(cx+float64((relx+float64(cv*rely))*rw), cy+float64((rely-float64(cv*relx))*rh))
		x3, y3 := at.apply(cx+float64(relx*rw), cy+float64(rely*rh))
		p.cubicTo(x1, y1, x2, y2, x3, y3)
	}
	return p
}

// current returns the end of the last segment; ok is false for an empty
// path or a path that ends with a close.
func (p *path) current() (x, y float64, ok bool) {
	if len(p.cmds) == 0 || p.cmds[len(p.cmds)-1] == cmdClose {
		return 0, 0, false
	}
	n := len(p.pts)
	return p.pts[n-2], p.pts[n-1], true
}

// generalPath is a java.awt.geom.GeneralPath (a Path2D.Float): its points
// are stored as float32. Appending with connect turns a leading move into a
// line, dropped when it goes nowhere, as Path2D.append does.
type generalPath struct{ p path }

func f32(v float64) float64 { return float64(float32(v)) }

func (g *generalPath) moveTo(x, y float64) {
	// Path2D replaces a move that follows a move.
	if n := len(g.p.cmds); n > 0 && g.p.cmds[n-1] == cmdMove {
		g.p.pts[len(g.p.pts)-2], g.p.pts[len(g.p.pts)-1] = f32(x), f32(y)
		return
	}
	g.p.moveTo(f32(x), f32(y))
}

func (g *generalPath) lineTo(x, y float64) { g.p.lineTo(f32(x), f32(y)) }

func (g *generalPath) quadTo(x1, y1, x2, y2 float64) { g.p.quadTo(f32(x1), f32(y1), f32(x2), f32(y2)) }

func (g *generalPath) cubicTo(x1, y1, x2, y2, x3, y3 float64) {
	g.p.cubicTo(f32(x1), f32(y1), f32(x2), f32(y2), f32(x3), f32(y3))
}

func (g *generalPath) closePath() {
	if n := len(g.p.cmds); n == 0 || g.p.cmds[n-1] != cmdClose {
		g.p.closePath()
	}
}

// appendPath appends the segments of q, as Path2D.append(shape, connect).
func (g *generalPath) appendPath(q *path, connect bool) {
	at := 0
	for i, c := range q.cmds {
		pts := q.pts[at : at+2*c.points()]
		at += 2 * c.points()
		if i == 0 && connect && c == cmdMove && len(g.p.cmds) > 0 {
			// After a close, the move always becomes a line.
			if x, y, ok := g.p.current(); ok && f32(pts[0]) == x && f32(pts[1]) == y {
				continue
			}
			g.lineTo(pts[0], pts[1])
			continue
		}
		switch c {
		case cmdMove:
			g.moveTo(pts[0], pts[1])
		case cmdLine:
			g.lineTo(pts[0], pts[1])
		case cmdQuad:
			g.quadTo(pts[0], pts[1], pts[2], pts[3])
		case cmdCubic:
			g.cubicTo(pts[0], pts[1], pts[2], pts[3], pts[4], pts[5])
		case cmdClose:
			g.closePath()
		}
	}
}

func (g *generalPath) path() *path { return &g.p }

// bounds2D returns the tight bounds of the path (Path2D.getBounds2D since
// Java 19: curves contribute their extrema, not their control points).
func (g *generalPath) bounds2D() (x, y, w, h float64) {
	x0, y0, x1, y1, ok := tightBounds(&g.p)
	if !ok {
		return 0, 0, 0, 0
	}
	return x0, y0, x1 - x0, y1 - y0
}

// tightBounds returns the bounds of the end points and of the extrema of the
// curves of a path.
func tightBounds(p *path) (x0, y0, x1, y1 float64, ok bool) {
	add := func(x, y float64) {
		if !ok {
			x0, y0, x1, y1, ok = x, y, x, y, true
			return
		}
		x0, y0 = math.Min(x0, x), math.Min(y0, y)
		x1, y1 = math.Max(x1, x), math.Max(y1, y)
	}
	at := 0
	var cx, cy float64
	for _, c := range p.cmds {
		pts := p.pts[at : at+2*c.points()]
		at += 2 * c.points()
		switch c {
		case cmdMove, cmdLine:
			add(pts[0], pts[1])
			cx, cy = pts[0], pts[1]
		case cmdQuad:
			add(pts[2], pts[3])
			for _, t := range quadExtrema(cx, pts[0], pts[2]) {
				add(quadAt(cx, pts[0], pts[2], t), quadAt(cy, pts[1], pts[3], t))
			}
			for _, t := range quadExtrema(cy, pts[1], pts[3]) {
				add(quadAt(cx, pts[0], pts[2], t), quadAt(cy, pts[1], pts[3], t))
			}
			cx, cy = pts[2], pts[3]
		case cmdCubic:
			add(pts[4], pts[5])
			ts := append(cubicExtrema(cx, pts[0], pts[2], pts[4]), cubicExtrema(cy, pts[1], pts[3], pts[5])...)
			for _, t := range ts {
				add(cubicAt(cx, pts[0], pts[2], pts[4], t), cubicAt(cy, pts[1], pts[3], pts[5], t))
			}
			cx, cy = pts[4], pts[5]
		}
	}
	return x0, y0, x1, y1, ok
}

func quadAt(p0, p1, p2, t float64) float64 {
	u := 1 - t
	return float64(u*u*p0) + float64(2*u*t*p1) + float64(t*t*p2)
}

func cubicAt(p0, p1, p2, p3, t float64) float64 {
	u := 1 - t
	return float64(u*u*u*p0) + float64(3*u*u*t*p1) + float64(3*u*t*t*p2) + float64(t*t*t*p3)
}

// quadExtrema returns the parameters in (0, 1) where a quadratic curve
// coordinate has a zero derivative.
func quadExtrema(p0, p1, p2 float64) []float64 {
	d := p0 - float64(2*p1) + p2
	if d == 0 {
		return nil
	}
	t := (p0 - p1) / d
	if t > 0 && t < 1 {
		return []float64{t}
	}
	return nil
}

// cubicExtrema returns the parameters in (0, 1) where a cubic curve
// coordinate has a zero derivative.
func cubicExtrema(p0, p1, p2, p3 float64) []float64 {
	a := -p0 + float64(3*p1) - float64(3*p2) + p3
	b := 2 * (p0 - float64(2*p1) + p2)
	c := p1 - p0
	var ts []float64
	keep := func(t float64) {
		if t > 0 && t < 1 {
			ts = append(ts, t)
		}
	}
	if math.Abs(a) < 1e-12 {
		if b != 0 {
			keep(-c / b)
		}
		return ts
	}
	disc := float64(b*b) - float64(4*a*c)
	if disc < 0 {
		return nil
	}
	sq := math.Sqrt(disc)
	keep((-b + sq) / (2 * a))
	keep((-b - sq) / (2 * a))
	return ts
}

// pathShape is a shape over a path built by a driver, with the tight bounds
// of Path2D.
type pathShape struct{ p *path }

func (s pathShape) path() *path { return s.p }

func (s pathShape) bounds2D() (x, y, w, h float64) {
	x0, y0, x1, y1, ok := tightBounds(s.p)
	if !ok {
		return 0, 0, 0, 0
	}
	return x0, y0, x1 - x0, y1 - y0
}
