// SPDX-License-Identifier: MIT
// Provenance: original

package dejavu

import "encoding/binary"

// Point is a point of an outline in 26.6 fixed point, y growing downward
// as in Java's user space.
type Point struct{ X, Y int64 }

// Op is the kind of an outline segment.
type Op uint8

// The segment kinds of an outline.
const (
	MoveTo Op = iota
	LineTo
	QuadTo
	Close
)

// Segment is an outline segment: MoveTo and LineTo use P[0]; QuadTo has its
// control point in P[0] and its end in P[1]; Close has no point.
type Segment struct {
	Op Op
	P  [2]Point
}

// Outline returns the outline of a glyph at an integer pixel size, as
// Java's FreeType scaler gives it with hinting off (fractional metrics):
// the TrueType points scaled with FreeType's FT_MulFix and FT_DivFix,
// translated so that the first phantom point is at the origin, and
// decomposed as FT_Outline_Decompose does (quadratic segments, implied
// on-curve points at the middle of two off-curve points, rounded toward
// zero). A missing or empty glyph gives no segment. The result is not
// cached: callers that draw a glyph often keep it.
func (f *Face) Outline(gid, ppem int) []Segment {
	scale := divFix(int64(ppem)<<6, int64(f.unitsPerEm))
	g := f.loadGlyph(gid, scale, 0)
	if len(g.points) > 0 {
		if data := f.glyphData(gid); len(data) >= 10 {
			xMin := int64(int16(binary.BigEndian.Uint16(data[2:])))
			shift := mulFix(xMin-int64(f.leftSideBearing(gid)), scale)
			if shift != 0 {
				for i := range g.points {
					g.points[i].x -= shift
				}
			}
		}
	}
	return decompose(g)
}

// HintedAdvance returns the advance of a glyph in whole pixels as FreeType
// gives it for a hinted load (Java without fractional metrics): the scaled
// advance width rounded to the pixel.
func (f *Face) HintedAdvance(gid, ppem int) int {
	scale := divFix(int64(ppem)<<6, int64(f.unitsPerEm))
	return int((mulFix(int64(f.AdvanceWidth(gid)), scale) + 32) >> 6)
}

// mulFix is FreeType's FT_MulFix: a * b / 65536 rounded half away from
// zero.
func mulFix(a, b int64) int64 {
	ab := a * b
	if ab < 0 {
		return -((-ab + 0x8000) >> 16)
	}
	return (ab + 0x8000) >> 16
}

// divFix is FreeType's FT_DivFix: a * 65536 / b rounded half away from
// zero.
func divFix(a, b int64) int64 {
	neg := (a < 0) != (b < 0)
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	q := (a<<16 + b/2) / b
	if neg {
		return -q
	}
	return q
}

// glyphPoint is a scaled point of a TrueType contour.
type glyphPoint struct {
	x, y int64
	on   bool
}

// glyphPoints are the scaled contours of a glyph, y growing upward.
type glyphPoints struct {
	points []glyphPoint
	ends   []int
}

// The composite glyph flags.
const (
	argsAreWords      = 0x0001
	argsAreXYValues   = 0x0002
	haveScale         = 0x0008
	moreComponents    = 0x0020
	haveXYScale       = 0x0040
	haveTwoByTwo      = 0x0080
	scaledCompoOffset = 0x0800
)

// loadGlyph loads the scaled points of a glyph, recursing into the
// components of a composite glyph as FreeType's TrueType loader does for
// unhinted loads.
func (f *Face) loadGlyph(gid int, scale int64, depth int) glyphPoints {
	data := f.glyphData(gid)
	if len(data) < 10 || depth > 8 {
		return glyphPoints{}
	}
	nContours := int(int16(binary.BigEndian.Uint16(data)))
	if nContours >= 0 {
		g := parseSimple(data, nContours)
		for i := range g.points {
			g.points[i].x = mulFix(g.points[i].x, scale)
			g.points[i].y = mulFix(g.points[i].y, scale)
		}
		return g
	}
	var base glyphPoints
	p := 10
	for {
		if p+4 > len(data) {
			break
		}
		flags := binary.BigEndian.Uint16(data[p:])
		component := int(binary.BigEndian.Uint16(data[p+2:]))
		p += 4
		var arg1, arg2 int64
		if flags&argsAreWords != 0 {
			if p+4 > len(data) {
				break
			}
			if flags&argsAreXYValues != 0 {
				arg1 = int64(int16(binary.BigEndian.Uint16(data[p:])))
				arg2 = int64(int16(binary.BigEndian.Uint16(data[p+2:])))
			} else {
				arg1 = int64(binary.BigEndian.Uint16(data[p:]))
				arg2 = int64(binary.BigEndian.Uint16(data[p+2:]))
			}
			p += 4
		} else {
			if p+2 > len(data) {
				break
			}
			if flags&argsAreXYValues != 0 {
				arg1 = int64(int8(data[p]))
				arg2 = int64(int8(data[p+1]))
			} else {
				arg1 = int64(data[p])
				arg2 = int64(data[p+1])
			}
			p += 2
		}
		// The 2x2 transform in 16.16 (FreeType reads F2Dot14 values << 2).
		xx, xy, yx, yy := int64(0x10000), int64(0), int64(0), int64(0x10000)
		f2dot14 := func(o int) int64 { return int64(int16(binary.BigEndian.Uint16(data[o:]))) << 2 }
		transformed := false
		switch {
		case flags&haveScale != 0 && p+2 <= len(data):
			xx = f2dot14(p)
			yy = xx
			p += 2
			transformed = true
		case flags&haveXYScale != 0 && p+4 <= len(data):
			xx = f2dot14(p)
			yy = f2dot14(p + 2)
			p += 4
			transformed = true
		case flags&haveTwoByTwo != 0 && p+8 <= len(data):
			xx = f2dot14(p)
			yx = f2dot14(p + 2)
			xy = f2dot14(p + 4)
			yy = f2dot14(p + 6)
			p += 8
			transformed = true
		}
		comp := f.loadGlyph(component, scale, depth+1)
		if transformed {
			for i := range comp.points {
				x, y := comp.points[i].x, comp.points[i].y
				comp.points[i].x = mulFix(x, xx) + mulFix(y, xy)
				comp.points[i].y = mulFix(x, yx) + mulFix(y, yy)
			}
		}
		var dx, dy int64
		if flags&argsAreXYValues != 0 {
			x, y := arg1, arg2
			if transformed && flags&scaledCompoOffset != 0 {
				x, y = mulFix(arg1, xx)+mulFix(arg2, xy), mulFix(arg1, yx)+mulFix(arg2, yy)
			}
			dx, dy = mulFix(x, scale), mulFix(y, scale)
		} else {
			// Point matching: the component point arg2 lands on the point
			// arg1 of the glyph built so far.
			p1, p2 := int(arg1), int(arg2)
			if p1 < len(base.points) && p2 < len(comp.points) {
				dx = base.points[p1].x - comp.points[p2].x
				dy = base.points[p1].y - comp.points[p2].y
			}
		}
		offset := len(base.points)
		for _, pt := range comp.points {
			base.points = append(base.points, glyphPoint{x: pt.x + dx, y: pt.y + dy, on: pt.on})
		}
		for _, e := range comp.ends {
			base.ends = append(base.ends, e+offset)
		}
		if flags&moreComponents == 0 {
			break
		}
	}
	return base
}

// parseSimple decodes the contours of a simple glyph record, in font
// units.
func parseSimple(g []byte, nContours int) glyphPoints {
	if nContours == 0 || len(g) < 12+2*nContours {
		return glyphPoints{}
	}
	u16 := func(o int) int { return int(binary.BigEndian.Uint16(g[o:])) }
	ends := make([]int, nContours)
	for i := range ends {
		ends[i] = u16(10 + 2*i)
	}
	nPoints := ends[nContours-1] + 1
	p := 12 + 2*nContours + u16(10+2*nContours)
	flags := make([]byte, 0, nPoints)
	for len(flags) < nPoints {
		if p >= len(g) {
			return glyphPoints{}
		}
		fl := g[p]
		p++
		flags = append(flags, fl)
		if fl&8 != 0 {
			if p >= len(g) {
				return glyphPoints{}
			}
			repeat := int(g[p])
			p++
			for range repeat {
				flags = append(flags, fl)
			}
		}
	}
	flags = flags[:nPoints]
	points := make([]glyphPoint, nPoints)
	ok := true
	coord := func(short, same byte, set func(i int, v int64)) {
		var v int64
		for i, fl := range flags {
			switch {
			case fl&short != 0:
				if p >= len(g) {
					ok = false
					return
				}
				d := int64(g[p])
				p++
				if fl&same == 0 {
					d = -d
				}
				v += d
			case fl&same == 0:
				if p+2 > len(g) {
					ok = false
					return
				}
				v += int64(int16(binary.BigEndian.Uint16(g[p:])))
				p += 2
			}
			set(i, v)
		}
	}
	coord(0x02, 0x10, func(i int, v int64) {
		points[i].x = v
		points[i].on = flags[i]&1 != 0
	})
	coord(0x04, 0x20, func(i int, v int64) { points[i].y = v })
	if !ok {
		return glyphPoints{}
	}
	return glyphPoints{points: points, ends: ends}
}

// decompose turns the contours into segments as FT_Outline_Decompose does,
// with y negated (Java's user space grows downward). Each contour starts
// with a move and is closed.
func decompose(g glyphPoints) []Segment {
	if len(g.points) == 0 {
		return nil
	}
	var out []Segment
	pt := func(p glyphPoint) Point { return Point{X: p.x, Y: -p.y} }
	mid := func(a, b glyphPoint) glyphPoint { return glyphPoint{x: (a.x + b.x) / 2, y: (a.y + b.y) / 2, on: true} }
	first := 0
	for _, last := range g.ends {
		if last < first || last >= len(g.points) {
			break
		}
		start := g.points[first]
		limit := last
		i := first
		if !start.on {
			if g.points[last].on {
				start = g.points[last]
				limit--
			} else {
				start = mid(start, g.points[last])
			}
			i--
		}
		out = append(out, Segment{Op: MoveTo, P: [2]Point{pt(start)}})
		closed := false
		for i < limit && !closed {
			i++
			if g.points[i].on {
				out = append(out, Segment{Op: LineTo, P: [2]Point{pt(g.points[i])}})
				continue
			}
			control := g.points[i]
			for {
				if i >= limit {
					out = append(out, Segment{Op: QuadTo, P: [2]Point{pt(control), pt(start)}})
					closed = true
					break
				}
				i++
				if g.points[i].on {
					out = append(out, Segment{Op: QuadTo, P: [2]Point{pt(control), pt(g.points[i])}})
					break
				}
				m := mid(control, g.points[i])
				out = append(out, Segment{Op: QuadTo, P: [2]Point{pt(control), pt(m)}})
				control = g.points[i]
			}
		}
		if !closed {
			out = append(out, Segment{Op: LineTo, P: [2]Point{pt(start)}})
		}
		out = append(out, Segment{Op: Close})
		first = last + 1
	}
	return out
}
