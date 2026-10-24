// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/font/UnusedSpace.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svg

import (
	_ "embed"
	"encoding/binary"
	"strconv"
	"sync"

	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// The centered characters of class headers and spots are drawn as glyph
// outlines (upstream's DriverCenteredCharacterSvg), centered with
// upstream's UnusedSpace.
//
// The outline is the one Java's TextLayout gives on the reference JVM: the
// TrueType glyph scaled by FreeType without hinting, in 26.6 fixed point.
// It is computed here from the glyph records of the DejaVu fonts behind
// Java's logical fonts (glyphs/dejavu.bin), with FreeType's arithmetic.
//
// UnusedSpace rasterizes the glyph with Java2D (hinted, not antialiased)
// and searches the point that minimizes the distance to the farthest
// pixel. Hinted rasterization is not reproducible here, so its results come
// from a table dumped from the reference JVM (glyphs/unusedspace.bin, from
// testdata/golden/glyphs.txt.gz). Both tables cover the printable ASCII
// characters (and space) of the 12 faces of DejaVu Sans, DejaVu Sans Mono
// and DejaVu Serif, the tables also for sizes 4 to 32; other characters,
// fonts or sizes panic with a "not ported yet" message.
//
// The glyph data is derived from the DejaVu fonts (Bitstream Vera license,
// see glyphs/LICENSE-DejaVu.txt).

//go:embed glyphs/dejavu.bin
var dejavuGlyphs []byte

//go:embed glyphs/unusedspace.bin
var unusedSpaceTable []byte

// The faces of the glyph tables, in their order: DejaVu Sans, Sans Mono and
// Serif, each plain, bold, italic, bold italic (Java's style values).
var glyphFaceNames = [...]string{
	"DejaVuSans", "DejaVuSans-Bold", "DejaVuSans-Oblique", "DejaVuSans-BoldOblique",
	"DejaVuSansMono", "DejaVuSansMono-Bold", "DejaVuSansMono-Oblique", "DejaVuSansMono-BoldOblique",
	"DejaVuSerif", "DejaVuSerif-Bold", "DejaVuSerif-Italic", "DejaVuSerif-BoldItalic",
}

// glyphFace returns the index in glyphFaceNames of the physical font Java
// draws an AWT font with on the reference JVM; ok is false for families
// that are not DejaVu.
func glyphFace(f font.AwtFont) (int, bool) {
	var family int
	switch f.Family() {
	case "SansSerif", "Dialog", "DejaVu Sans":
		family = 0
	case "Monospaced", "DialogInput", "DejaVu Sans Mono":
		family = 1
	case "Serif", "DejaVu Serif":
		family = 2
	default:
		return 0, false
	}
	return family*4 + int(f.Style()&font.AwtBoldItalic), true
}

// glyphRecord is the TrueType glyph of a character: its glyf table record
// and the shift of FreeType's first phantom point (xMin - lsb).
type glyphRecord struct {
	shift int
	glyf  []byte
}

type glyphTables struct {
	unitsPerEm [len(glyphFaceNames)]int
	firstChar  rune
	glyphs     [len(glyphFaceNames)][]glyphRecord

	minSize, maxSize int
	centerFirst      rune
	centerCount      int
	centers          []byte
}

var (
	glyphTablesOnce sync.Once
	glyphData       *glyphTables
)

// loadGlyphTables parses the embedded tables once; they are read-only
// afterwards.
func loadGlyphTables() *glyphTables {
	glyphTablesOnce.Do(func() {
		t := &glyphTables{}
		b := dejavuGlyphs
		if string(b[:8]) != "PUGLYPH1" || int(b[8]) != len(glyphFaceNames) {
			panic("svg: bad glyphs/dejavu.bin")
		}
		p := 9
		for face := range glyphFaceNames {
			n := int(b[p])
			if string(b[p+1:p+1+n]) != glyphFaceNames[face] {
				panic("svg: bad face in glyphs/dejavu.bin")
			}
			p += 1 + n
			t.unitsPerEm[face] = int(binary.BigEndian.Uint16(b[p:]))
			t.firstChar = rune(binary.BigEndian.Uint16(b[p+2:]))
			count := int(binary.BigEndian.Uint16(b[p+4:]))
			p += 6
			t.glyphs[face] = make([]glyphRecord, count)
			for i := range count {
				shift := int(int16(binary.BigEndian.Uint16(b[p:])))
				size := int(binary.BigEndian.Uint16(b[p+2:]))
				t.glyphs[face][i] = glyphRecord{shift: shift, glyf: b[p+4 : p+4+size]}
				p += 4 + size
			}
		}
		c := unusedSpaceTable
		if string(c[:8]) != "PUUNUSE1" || int(c[8]) != len(glyphFaceNames) {
			panic("svg: bad glyphs/unusedspace.bin")
		}
		t.minSize, t.maxSize = int(c[9]), int(c[10])
		t.centerFirst = rune(binary.BigEndian.Uint16(c[11:]))
		t.centerCount = int(binary.BigEndian.Uint16(c[13:]))
		t.centers = c[15:]
		glyphData = t
	})
	return glyphData
}

// unusedSpaceCenter returns the center of upstream's UnusedSpace of a
// character drawn with f (its getCenterX and getCenterY); ok is false
// outside of the table.
func unusedSpaceCenter(f font.UFont, c rune) (centerX, centerY float64, ok bool) {
	awt := f.UnderlyingFont(string(c))
	face, ok := glyphFace(awt)
	t := loadGlyphTables()
	size := awt.Size()
	idx := int(c - t.centerFirst)
	if !ok || size < t.minSize || size > t.maxSize || idx < 0 || idx >= t.centerCount {
		return 0, 0, false
	}
	at := 2 * ((face*(t.maxSize-t.minSize+1)+size-t.minSize)*t.centerCount + idx)
	return float64(int8(t.centers[at])) / 4, float64(int8(t.centers[at+1])) / 4, true
}

// outlineSegment is a segment of a glyph outline in 26.6 fixed point, y
// growing downward as Java's paths: a move, a line, a quadratic curve or a
// close.
type outlineSegment struct {
	typ    geom.USegmentType
	coords []int64
}

// glyphOutline returns the outline of a character drawn with f, as Java's
// TextLayout.getOutline gives it on the reference JVM; ok is false outside
// of the glyph table.
func glyphOutline(f font.UFont, c rune) ([]outlineSegment, bool) {
	awt := f.UnderlyingFont(string(c))
	face, ok := glyphFace(awt)
	t := loadGlyphTables()
	idx := int(c - t.firstChar)
	if !ok || idx < 0 || idx >= len(t.glyphs[face]) {
		return nil, false
	}
	return scaledOutline(t.glyphs[face][idx], t.unitsPerEm[face], awt.Size()), true
}

// glyphPoint is a point of a TrueType contour.
type glyphPoint struct {
	x, y int64
	on   bool
}

// parseSimpleGlyph decodes a simple glyph record: contour end indices and
// points in font units.
func parseSimpleGlyph(g []byte) (ends []int, points []glyphPoint) {
	if len(g) == 0 {
		return nil, nil
	}
	u16 := func(o int) int { return int(binary.BigEndian.Uint16(g[o:])) }
	nContours := int(int16(binary.BigEndian.Uint16(g)))
	if nContours < 0 {
		panic("svg: composite glyphs are not supported")
	}
	ends = make([]int, nContours)
	for i := range ends {
		ends[i] = u16(10 + 2*i)
	}
	if nContours == 0 {
		return ends, nil
	}
	nPoints := ends[nContours-1] + 1
	p := 12 + 2*nContours + u16(10+2*nContours)
	flags := make([]byte, 0, nPoints)
	for len(flags) < nPoints {
		fl := g[p]
		p++
		flags = append(flags, fl)
		if fl&8 != 0 {
			repeat := int(g[p])
			p++
			for range repeat {
				flags = append(flags, fl)
			}
		}
	}
	points = make([]glyphPoint, nPoints)
	coord := func(short, same byte, set func(i int, v int64)) {
		var v int64
		for i, f := range flags {
			switch {
			case f&short != 0:
				d := int64(g[p])
				p++
				if f&same == 0 {
					d = -d
				}
				v += d
			case f&same == 0:
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
	return ends, points
}

// mulFix is FreeType's FT_MulFix: a * b / 65536 rounded half away from
// zero.
func mulFix(a, b int64) int64 {
	ab := a * b
	var neg int64
	if ab < 0 {
		neg = 1
	}
	return (ab + 0x8000 - neg) >> 16
}

// scaledOutline is the unhinted outline FreeType loads for a glyph at an
// integer pixel size (Java's FreetypeFontScaler sets the size in 26.6 at 72
// dpi), decomposed as FT_Outline_Decompose and Java's GeneralPath do: a
// move per contour, lines and quadratic curves (two consecutive off-curve
// points give an on-curve point at their middle, rounded toward zero), the
// contour closed by a line or curve back to its start, a close before each
// move but the first and at the end.
func scaledOutline(g glyphRecord, unitsPerEm, size int) []outlineSegment {
	ends, points := parseSimpleGlyph(g.glyf)
	if len(points) == 0 {
		return nil
	}
	ppem := int64(size)
	// FT_DivFix(ppem << 6, units_per_EM); DejaVu rounds the ppem (head
	// flags bit 3), which an integer size already is.
	scale := (ppem<<6<<16 + int64(unitsPerEm)/2) / int64(unitsPerEm)
	shift := mulFix(int64(g.shift), scale)
	for i := range points {
		points[i].x = mulFix(points[i].x, scale) - shift
		points[i].y = mulFix(points[i].y, scale)
	}
	var out []outlineSegment
	emit := func(typ geom.USegmentType, p ...glyphPoint) {
		coords := make([]int64, 0, 2*len(p))
		for _, q := range p {
			coords = append(coords, q.x, -q.y)
		}
		out = append(out, outlineSegment{typ: typ, coords: coords})
	}
	mid := func(a, b glyphPoint) glyphPoint { return glyphPoint{x: (a.x + b.x) / 2, y: (a.y + b.y) / 2, on: true} }
	first := 0
	for _, last := range ends {
		start := points[first]
		limit := last
		i := first
		if !start.on {
			if points[last].on {
				start = points[last]
				limit--
			} else {
				start = mid(start, points[last])
			}
			i--
		}
		if len(out) > 0 {
			emit(geom.SegClose)
		}
		emit(geom.SegMoveTo, start)
		closed := false
		for i < limit && !closed {
			i++
			if points[i].on {
				emit(geom.SegLineTo, points[i])
				continue
			}
			control := points[i]
			for {
				if i >= limit {
					emit(geom.SegQuadTo, control, start)
					closed = true
					break
				}
				i++
				if points[i].on {
					emit(geom.SegQuadTo, control, points[i])
					break
				}
				m := mid(control, points[i])
				emit(geom.SegQuadTo, control, m)
				control = points[i]
			}
		}
		if !closed {
			emit(geom.SegLineTo, start)
		}
		first = last + 1
	}
	emit(geom.SegClose)
	return out
}

// outlineString is the outline as SvgGolden.java's outline writes it, for
// the tests.
func outlineString(segments []outlineSegment) string {
	var b []byte
	for i, s := range segments {
		if i > 0 {
			b = append(b, ';')
		}
		b = strconv.AppendInt(b, int64(s.typ), 10)
		for _, c := range s.coords {
			b = append(b, ' ')
			b = strconv.AppendInt(b, c, 10)
		}
	}
	return string(b)
}

// outlineIterator iterates an outline in pixels, as Java's PathIterator
// over the GeneralPath of a glyph (float coordinates).
type outlineIterator struct {
	segments []outlineSegment
	pos      int
}

func (it *outlineIterator) IsDone() bool { return it.pos >= len(it.segments) }

func (it *outlineIterator) CurrentSegment(coords *[6]float64) geom.USegmentType {
	s := it.segments[it.pos]
	for i, c := range s.coords {
		coords[i] = float64(float32(c) / 64)
	}
	return s.typ
}

func (it *outlineIterator) Next() { it.pos++ }

// WindingRule returns Java's WIND_NON_ZERO, which the SVG output ignores.
func (it *outlineIterator) WindingRule() int { return 1 }
