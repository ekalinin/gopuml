// SPDX-License-Identifier: MIT
// Provenance: original

package g2d

import (
	"math"

	"github.com/ekalinin/gopuml/internal/jmath"
)

// The cap and join styles of java.awt.BasicStroke, with Java's values.
const (
	capButt   = 0
	capRound  = 1
	capSquare = 2

	joinMiter = 0
	joinRound = 1
	joinBevel = 2
)

// basicStroke is a java.awt.BasicStroke.
type basicStroke struct {
	width      float64
	cap        int
	join       int
	miterLimit float64
	dash       []float64
	phase      float64
}

// newBasicStroke is new BasicStroke(width): square caps, miter joins with a
// limit of 10, no dashes.
func newBasicStroke(width float64) basicStroke {
	return basicStroke{width: width, cap: capSquare, join: joinMiter, miterLimit: 10}
}

// polyline is a flattened subpath: points (x, y pairs) without consecutive
// duplicates; closed when the subpath ends with a close segment. A
// subpath whose segments all have zero length keeps its single point and is
// marked dot.
type polyline struct {
	pts    []float64
	closed bool
	dot    bool
}

// strokeOutline returns the area of a path stroked with s, in the same
// space as the path, as polygons to fill with the non-zero rule, as
// Java's stroker gives it: the path is flattened (curves within tol), then
// dashed (the pattern restarts with each subpath; on a closed subpath the
// last dash continues into the first), then each piece is covered by a
// quadrilateral per segment, a join between segments (miter within the
// miter limit, else bevel; or round) and caps at the ends of open pieces
// (butt, square or round). A subpath of zero length gets the caps of a
// segment going right (nothing with butt caps). The polygons overlap where
// segments meet; all run in the same direction, so the non-zero rule fills
// their union.
func strokeOutline(p *path, s basicStroke, tol float64) [][]float64 {
	hw := s.width / 2
	if !(hw > 0) {
		return nil
	}
	st := stroker{hw: hw, s: s, tol: tol}
	for _, pl := range flattenSubpaths(p, tol) {
		switch {
		case pl.dot:
			st.dot(pl.pts[0], pl.pts[1])
		case dashed(s.dash):
			for _, piece := range dash(pl, s.dash, s.phase) {
				st.polyline(piece, false)
			}
		default:
			st.polyline(pl.pts, pl.closed)
		}
	}
	return st.out
}

// dashed reports whether a dash pattern draws gaps: Java rejects patterns
// with negative lengths or a zero sum, which never reach here.
func dashed(d []float64) bool {
	if len(d) == 0 {
		return false
	}
	sum := 0.0
	for _, v := range d {
		if v < 0 {
			return false
		}
		sum += v
	}
	return sum > 0
}

// flattenSubpaths splits a path into flattened subpaths. A segment after a
// close starts a new subpath at the start of the closed one.
func flattenSubpaths(p *path, tol float64) []polyline {
	var out []polyline
	var cur []float64
	var sx, sy, cx, cy float64
	open := false
	hadSegment := false
	add := func(x, y float64) {
		n := len(cur)
		if n >= 2 && cur[n-2] == x && cur[n-1] == y {
			return
		}
		cur = append(cur, x, y)
	}
	flush := func(closed bool) {
		if !open {
			return
		}
		switch {
		case len(cur) >= 4:
			// The closing segment of a closed polyline is implied.
			if closed && cur[0] == cur[len(cur)-2] && cur[1] == cur[len(cur)-1] {
				cur = cur[:len(cur)-2]
			}
			out = append(out, polyline{pts: cur, closed: closed})
		case hadSegment:
			out = append(out, polyline{pts: cur, dot: true})
		}
		cur = nil
		open = false
		hadSegment = false
	}
	start := func(x, y float64) {
		cur = []float64{x, y}
		open = true
	}
	at := 0
	for _, c := range p.cmds {
		pts := p.pts[at : at+2*c.points()]
		at += 2 * c.points()
		switch c {
		case cmdMove:
			flush(false)
			sx, sy, cx, cy = pts[0], pts[1], pts[0], pts[1]
			start(cx, cy)
			continue
		case cmdClose:
			if open {
				add(sx, sy)
			}
			flush(true)
			cx, cy = sx, sy
			continue
		}
		if !open {
			start(cx, cy)
		}
		hadSegment = true
		switch c {
		case cmdLine:
			add(pts[0], pts[1])
			cx, cy = pts[0], pts[1]
		case cmdQuad:
			var q []float64
			q = flattenQuad(q, cx, cy, pts[0], pts[1], pts[2], pts[3], tol)
			for i := 0; i < len(q); i += 2 {
				add(q[i], q[i+1])
			}
			cx, cy = pts[2], pts[3]
		case cmdCubic:
			var q []float64
			q = flattenCubic(q, cx, cy, pts[0], pts[1], pts[2], pts[3], pts[4], pts[5], tol)
			for i := 0; i < len(q); i += 2 {
				add(q[i], q[i+1])
			}
			cx, cy = pts[4], pts[5]
		}
	}
	flush(false)
	return out
}

// dash cuts a flattened subpath into the polylines of the "on" intervals
// of a dash pattern starting phase into it. On a closed subpath the
// closing segment is dashed too, and a dash running at the end continues
// into the first dash.
func dash(pl polyline, d []float64, phase float64) [][]float64 {
	pts := pl.pts
	if pl.closed {
		pts = append(append([]float64(nil), pts...), pts[0], pts[1])
	}
	// The dash index and the length left in it, after the phase.
	idx := 0
	on := true
	left := d[0]
	for phase > 0 {
		if phase < left {
			left -= phase
			break
		}
		phase -= left
		idx = (idx + 1) % len(d)
		on = !on
		left = d[idx]
	}
	startsOn := on
	var pieces [][]float64
	var cur []float64
	if on {
		cur = []float64{pts[0], pts[1]}
	}
	firstDone := false // the first dash of a closed subpath is complete
	var first []float64
	emit := func() {
		if len(cur) >= 4 {
			if pl.closed && startsOn && !firstDone {
				first = cur
			} else {
				pieces = append(pieces, cur)
			}
		}
		firstDone = true
		cur = nil
	}
	for i := 0; i+3 < len(pts); i += 2 {
		x0, y0, x1, y1 := pts[i], pts[i+1], pts[i+2], pts[i+3]
		dx, dy := x1-x0, y1-y0
		length := math.Sqrt(float64(dx*dx) + float64(dy*dy))
		t := 0.0
		for length-t > left {
			t += left
			f := t / length
			x, y := x0+float64(dx*f), y0+float64(dy*f)
			if on {
				cur = append(cur, x, y)
				emit()
			} else {
				cur = []float64{x, y}
				if !firstDone {
					firstDone = true
				}
			}
			idx = (idx + 1) % len(d)
			on = !on
			left = d[idx]
		}
		left -= length - t
		if on {
			cur = append(cur, x1, y1)
		}
	}
	if on && len(cur) >= 4 {
		if pl.closed && startsOn && first != nil {
			// The last dash runs through the start into the first one.
			cur = append(cur, first[2:]...)
			pieces = append(pieces, cur)
			first = nil
		} else {
			pieces = append(pieces, cur)
		}
	}
	if first != nil {
		pieces = append(pieces, first)
	}
	return pieces
}

// stroker collects the polygons of a stroke.
type stroker struct {
	hw  float64
	s   basicStroke
	tol float64
	out [][]float64
}

// add adds a polygon, oriented with a positive signed area.
func (st *stroker) add(poly ...float64) {
	area := 0.0
	n := len(poly) / 2
	for i := range n {
		j := (i + 1) % n
		area += float64(poly[2*i]*poly[2*j+1]) - float64(poly[2*j]*poly[2*i+1])
	}
	if area == 0 {
		return
	}
	if area < 0 {
		for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
			poly[2*i], poly[2*j] = poly[2*j], poly[2*i]
			poly[2*i+1], poly[2*j+1] = poly[2*j+1], poly[2*i+1]
		}
	}
	st.out = append(st.out, poly)
}

// unit returns the unit direction from (x0, y0) to (x1, y1).
func unit(x0, y0, x1, y1 float64) (float64, float64) {
	dx, dy := x1-x0, y1-y0
	l := math.Sqrt(float64(dx*dx) + float64(dy*dy))
	return dx / l, dy / l
}

// polyline strokes a polyline: a quadrilateral per segment, the joins, and
// caps at the ends of an open one. Repeated points are skipped; a piece of
// zero length draws nothing.
func (st *stroker) polyline(raw []float64, closed bool) {
	pts := make([]float64, 0, len(raw))
	for i := 0; i+1 < len(raw); i += 2 {
		if n := len(pts); n >= 2 && pts[n-2] == raw[i] && pts[n-1] == raw[i+1] {
			continue
		}
		pts = append(pts, raw[i], raw[i+1])
	}
	if closed && len(pts) >= 4 && pts[0] == pts[len(pts)-2] && pts[1] == pts[len(pts)-1] {
		pts = pts[:len(pts)-2]
	}
	n := len(pts) / 2
	if n < 2 {
		return
	}
	hw := st.hw
	for i := 0; i+1 < n; i++ {
		st.segment(pts[2*i], pts[2*i+1], pts[2*i+2], pts[2*i+3])
	}
	if closed {
		st.segment(pts[2*n-2], pts[2*n-1], pts[0], pts[1])
		for i := range n {
			p := (i + n - 1) % n
			q := (i + 1) % n
			st.join(pts[2*p], pts[2*p+1], pts[2*i], pts[2*i+1], pts[2*q], pts[2*q+1])
		}
		return
	}
	for i := 1; i+1 < n; i++ {
		st.join(pts[2*i-2], pts[2*i-1], pts[2*i], pts[2*i+1], pts[2*i+2], pts[2*i+3])
	}
	dx, dy := unit(pts[2], pts[3], pts[0], pts[1])
	st.cap(pts[0], pts[1], dx, dy, hw)
	dx, dy = unit(pts[2*n-4], pts[2*n-3], pts[2*n-2], pts[2*n-1])
	st.cap(pts[2*n-2], pts[2*n-1], dx, dy, hw)
}

// segment covers a segment with a quadrilateral of the stroke width.
func (st *stroker) segment(x0, y0, x1, y1 float64) {
	dx, dy := unit(x0, y0, x1, y1)
	nx, ny := float64(-dy*st.hw), float64(dx*st.hw)
	st.add(x0+nx, y0+ny, x1+nx, y1+ny, x1-nx, y1-ny, x0-nx, y0-ny)
}

// join covers the outer side of the turn at (x1, y1) between the segment
// from (x0, y0) and the segment to (x2, y2).
func (st *stroker) join(x0, y0, x1, y1, x2, y2 float64) {
	d0x, d0y := unit(x0, y0, x1, y1)
	d1x, d1y := unit(x1, y1, x2, y2)
	cross := float64(d0x*d1y) - float64(d0y*d1x)
	dot := float64(d0x*d1x) + float64(d0y*d1y)
	if cross == 0 && dot > 0 {
		return
	}
	hw := st.hw
	// The outer side is on the right of a left turn and on the left of a
	// right turn.
	s := hw
	if cross > 0 {
		s = -hw
	}
	o0x, o0y := float64(-d0y*s), float64(d0x*s)
	o1x, o1y := float64(-d1y*s), float64(d1x*s)
	switch st.s.join {
	case joinRound:
		st.arcFan(x1, y1, o0x, o0y, o1x, o1y)
	case joinMiter:
		// The tip is hw / cos(a / 2) away, a the angle between the
		// normals: within the limit when 1 + cos(a) >= 2 / limit^2.
		limit := st.s.miterLimit
		if 1+dot >= 2/float64(limit*limit) && 1+dot > 0 {
			tx, ty := x1+(o0x+o1x)/(1+dot), y1+(o0y+o1y)/(1+dot)
			st.add(x1, y1, x1+o0x, y1+o0y, tx, ty, x1+o1x, y1+o1y)
			return
		}
		st.add(x1, y1, x1+o0x, y1+o0y, x1+o1x, y1+o1y)
	default:
		st.add(x1, y1, x1+o0x, y1+o0y, x1+o1x, y1+o1y)
	}
}

// cap covers the end (x, y) of an open polyline going in direction
// (dx, dy).
func (st *stroker) cap(x, y, dx, dy, hw float64) {
	nx, ny := float64(-dy*hw), float64(dx*hw)
	switch st.s.cap {
	case capSquare:
		ex, ey := float64(dx*hw), float64(dy*hw)
		st.add(x+nx, y+ny, x+nx+ex, y+ny+ey, x-nx+ex, y-ny+ey, x-nx, y-ny)
	case capRound:
		// Two quarter turns through the direction of the end.
		ex, ey := float64(dx*hw), float64(dy*hw)
		st.arcFan(x, y, nx, ny, ex, ey)
		st.arcFan(x, y, ex, ey, -nx, -ny)
	}
}

// dot covers a subpath of zero length as Java does: the caps of a segment
// going right.
func (st *stroker) dot(x, y float64) {
	hw := st.hw
	switch st.s.cap {
	case capSquare:
		st.add(x-hw, y-hw, x+hw, y-hw, x+hw, y+hw, x-hw, y+hw)
	case capRound:
		st.arcFan(x, y, hw, 0, 0, hw)
		st.arcFan(x, y, 0, hw, -hw, 0)
		st.arcFan(x, y, -hw, 0, 0, -hw)
		st.arcFan(x, y, 0, -hw, hw, 0)
	}
}

// arcFan covers the circular sector of center (x, y) from the vector
// (ax, ay) to (bx, by), going the short way (less than a half turn), with
// the flattening tolerance.
func (st *stroker) arcFan(x, y, ax, ay, bx, by float64) {
	hw := st.hw
	a0 := jmath.Atan2(ay, ax)
	sweep := jmath.Atan2(float64(ax*by)-float64(ay*bx), float64(ax*bx)+float64(ay*by))
	if sweep == 0 {
		return
	}
	// Steps whose chords stay within the tolerance.
	step := 2 * jmath.Acos(math.Max(-1, 1-st.tol/hw))
	n := max(1, int(math.Ceil(math.Abs(sweep)/step)))
	poly := []float64{x, y}
	for i := 0; i <= n; i++ {
		a := a0 + float64(sweep*float64(i))/float64(n)
		poly = append(poly, x+float64(hw*jmath.Cos(a)), y+float64(hw*jmath.Sin(a)))
	}
	st.add(poly...)
}
