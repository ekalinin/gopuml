// SPDX-License-Identifier: MIT
// Provenance: original

package g2d

import (
	"math"
	"testing"
)

// coverage returns the covered area of stroke polygons (non-zero rule),
// measured with Marlin's sampling over a 100 x 100 image.
func coverage(polys [][]float64) float64 {
	m := rasterSampled(polyEdges(polys), irect{0, 0, 100, 100}, false, marlinSubX, marlinSubY)
	if m == nil {
		return 0
	}
	sum := 0.0
	for _, a := range m.a {
		sum += float64(a) / 255
	}
	return sum
}

func strokePath(build func(p *path), s basicStroke) [][]float64 {
	p := &path{}
	build(p)
	return strokeOutline(p, s, flatTol)
}

func near(got, want, tol float64) bool { return math.Abs(got-want) <= tol }

func TestStrokeLineCaps(t *testing.T) {
	line := func(p *path) { p.moveTo(10, 10.5); p.lineTo(30, 10.5) }
	for _, c := range []struct {
		cap  int
		want float64
	}{{capButt, 20}, {capSquare, 21}, {capRound, 20 + math.Pi/4}} {
		s := newBasicStroke(1)
		s.cap = c.cap
		if got := coverage(strokePath(line, s)); !near(got, c.want, 0.05) {
			t.Errorf("cap %d: area %.3f, want %.3f", c.cap, got, c.want)
		}
	}
}

func TestStrokeJoins(t *testing.T) {
	// A right angle: the miter fills the corner square, the bevel half of
	// it, the round join a quarter disc.
	corner := func(p *path) { p.moveTo(10, 20); p.lineTo(40, 20); p.lineTo(40, 50) }
	base := 30.0*2 + 30*2 // two 30 long segments 2 wide, butt caps
	for _, c := range []struct {
		join int
		want float64
	}{{joinMiter, base}, {joinBevel, base - 0.5}, {joinRound, base - 1 + math.Pi/4}} {
		s := basicStroke{width: 2, cap: capButt, join: c.join, miterLimit: 10}
		if got := coverage(strokePath(corner, s)); !near(got, c.want, 0.1) {
			t.Errorf("join %d: area %.3f, want %.3f", c.join, got, c.want)
		}
	}
	// A turn sharper than the miter limit allows gets a bevel: the area is
	// close to that of the bevel join.
	sharp := func(p *path) { p.moveTo(10, 50); p.lineTo(60, 52); p.lineTo(10, 54) }
	miter := coverage(strokePath(sharp, basicStroke{width: 2, cap: capButt, join: joinMiter, miterLimit: 10}))
	bevel := coverage(strokePath(sharp, basicStroke{width: 2, cap: capButt, join: joinBevel, miterLimit: 10}))
	if !near(miter, bevel, 0.01) {
		t.Errorf("sharp miter %.3f, bevel %.3f", miter, bevel)
	}
}

func TestStrokeClosed(t *testing.T) {
	// A closed square has joins everywhere and no caps: a ring.
	square := func(p *path) {
		p.moveTo(20, 20)
		p.lineTo(60, 20)
		p.lineTo(60, 60)
		p.lineTo(20, 60)
		p.closePath()
	}
	if got := coverage(strokePath(square, newBasicStroke(2))); !near(got, 42*42-38*38, 0.1) {
		t.Errorf("closed square: area %.3f", got)
	}
	// A dot: square caps of a segment going right.
	dot := func(p *path) { p.moveTo(50, 50); p.lineTo(50, 50) }
	if got := coverage(strokePath(dot, newBasicStroke(4))); !near(got, 16, 0.05) {
		t.Errorf("dot: area %.3f", got)
	}
	s := newBasicStroke(4)
	s.cap = capButt
	if polys := strokePath(dot, s); len(polys) != 0 {
		t.Errorf("dot with butt caps: %d polygons", len(polys))
	}
	if polys := strokePath(func(p *path) { p.moveTo(5, 5) }, newBasicStroke(1)); len(polys) != 0 {
		t.Errorf("lone move: %d polygons", len(polys))
	}
}

func TestDash(t *testing.T) {
	open := polyline{pts: []float64{0, 0, 20, 0}}
	pieces := dash(open, []float64{5, 5}, 0)
	if len(pieces) != 2 || pieces[0][2] != 5 || pieces[1][0] != 10 || pieces[1][2] != 15 {
		t.Errorf("open dashes %v", pieces)
	}
	// A dash that runs past a vertex follows it.
	bent := polyline{pts: []float64{0, 0, 3, 0, 3, 10}}
	if pieces := dash(bent, []float64{5, 2}, 0); len(pieces[0]) != 6 || pieces[0][5] != 2 {
		t.Errorf("bent dashes %v", pieces)
	}
	// On a closed subpath (perimeter 40) the last dash runs through the
	// start into the first one: 40 = 4 * (7 + 3), the pattern ends on a
	// gap, so 4 dashes; with (7, 2) it ends inside a dash which joins the
	// first.
	closed := polyline{pts: []float64{0, 0, 10, 0, 10, 10, 0, 10}, closed: true}
	if pieces := dash(closed, []float64{7, 3}, 0); len(pieces) != 4 {
		t.Errorf("closed dashes %d", len(pieces))
	}
	pieces = dash(closed, []float64{7, 2}, 0)
	last := pieces[len(pieces)-1]
	if len(pieces) != 4 || last[len(last)-2] != 7 || last[len(last)-1] != 0 {
		t.Errorf("closed joined dashes %d %v", len(pieces), last)
	}
	// The phase starts into the pattern.
	if pieces := dash(open, []float64{5, 5}, 7); pieces[0][0] != 3 {
		t.Errorf("phase %v", pieces)
	}
	if dashed([]float64{0, 0}) || dashed([]float64{1, -1}) || !dashed([]float64{1, 0}) {
		t.Error("dashed")
	}
}

// TestStrokeDegenerateCurve strokes curves whose control points sit on
// their end points (the database cylinder): the stroke covers the curve
// only, about its length times the width.
func TestStrokeDegenerateCurve(t *testing.T) {
	curve := func(p *path) {
		p.moveTo(10, 30)
		p.cubicTo(10, 20, 30, 20, 30, 20)
		p.cubicTo(30, 20, 50, 20, 50, 30)
	}
	got := coverage(strokePath(curve, basicStroke{width: 1, cap: capButt, join: joinMiter, miterLimit: 10}))
	length := 0.0
	p := &path{}
	curve(p)
	for _, pl := range flattenSubpaths(p, 0.001) {
		for i := 0; i+3 < len(pl.pts); i += 2 {
			length += math.Hypot(pl.pts[i+2]-pl.pts[i], pl.pts[i+3]-pl.pts[i+1])
		}
	}
	// The area between a curve and its chord would add tens of pixels.
	if !near(got, length, 0.5) {
		t.Errorf("area %.3f, length %.3f", got, length)
	}
}
