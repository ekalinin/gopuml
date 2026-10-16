// SPDX-License-Identifier: MIT
// Provenance: original

package dejavu

import (
	"bufio"
	"compress/gzip"
	"hash/fnv"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
)

// outlineString writes an outline as SvgGolden.java's outline does: the
// Java segment types (0 move, 1 line, 2 quad, 4 close) and the coordinates
// in 1/64 pixel.
func outlineString(segments []Segment) string {
	var b []byte
	for i, s := range segments {
		if i > 0 {
			b = append(b, ';')
		}
		var pts []Point
		switch s.Op {
		case MoveTo:
			b = append(b, '0')
			pts = s.P[:1]
		case LineTo:
			b = append(b, '1')
			pts = s.P[:1]
		case QuadTo:
			b = append(b, '2')
			pts = s.P[:2]
		case Close:
			b = append(b, '4')
		}
		for _, p := range pts {
			b = append(b, ' ')
			b = strconv.AppendInt(b, p.X, 10)
			b = append(b, ' ')
			b = strconv.AppendInt(b, p.Y, 10)
		}
	}
	return string(b)
}

var styleFaces = map[string]font.UFontFace{"plain": font.NormalFace(), "bold": font.BoldFace(),
	"italic": font.ItalicFace(), "bolditalic": font.BoldItalicFace()}

// TestOutlinesMatchUpstream compares the outlines (by hash) and the
// UnusedSpace centers of the printable ASCII characters with what upstream's
// TextLayout and UnusedSpace give on the reference JVM, for each logical
// font, style and size of the SVG driver's golden file.
func TestOutlinesMatchUpstream(t *testing.T) {
	f, err := os.Open("../../svg/testdata/golden/glyphs.txt.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	sc := bufio.NewScanner(zr)
	sc.Buffer(nil, 1<<20)
	lines, failures := 0, 0
	for sc.Scan() {
		p := strings.Split(sc.Text(), "\t")
		size, _ := strconv.Atoi(p[2])
		uf := font.Build(p[0], styleFaces[p[1]], size)
		centers := strings.Fields(p[4])
		var sb strings.Builder
		for c := rune(0x20); c <= 0x7E; c++ {
			awt := uf.UnderlyingFont(string(c))
			face := ForAwtFont(awt)
			sb.WriteString(outlineString(face.Outline(face.GlyphIndex(c), awt.Size())))
			sb.WriteByte('|')
			cx, cy, ok := UnusedSpaceCenter(uf, c)
			want := centers[c-0x20]
			if got := strconv.Itoa(int(cx*4)) + "," + strconv.Itoa(int(cy*4)); !ok || got != want {
				failures++
				t.Errorf("%s %s %d %q: center %s (%v), upstream %s", p[0], p[1], size, c, got, ok, want)
			}
		}
		h := fnv.New64a()
		h.Write([]byte(sb.String()))
		if got := strconv.FormatUint(h.Sum64(), 16); got != p[3] {
			failures++
			t.Errorf("%s %s %d: outline hash %s, upstream %s", p[0], p[1], size, got, p[3])
		}
		if failures > 20 {
			t.Fatal("too many failures")
		}
		lines++
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if lines < 400 {
		t.Fatalf("only %d lines", lines)
	}
}

func TestFaces(t *testing.T) {
	for fam := Sans; fam <= Serif; fam++ {
		for s := font.AwtPlain; s <= font.AwtBoldItalic; s++ {
			f := Get(fam, s)
			if f.Name() != faceNames[fam][s] || f.Family() != fam || f.Style() != s || f.UnitsPerEm() != 2048 ||
				f.NumGlyphs() < 2000 {
				t.Errorf("%s: %d units, %d glyphs", f.Name(), f.UnitsPerEm(), f.NumGlyphs())
			}
			if !f.HasGlyph('A') || f.HasGlyph(0xE000) || f.GlyphIndex(-1) != 0 {
				t.Errorf("%s: cmap", f.Name())
			}
			if len(f.Fallbacks()) < 2 {
				t.Errorf("%s: fallbacks", f.Name())
			}
		}
	}
	sans := Get(Sans, font.AwtPlain)
	// A composite glyph (e acute) and a character outside the BMP.
	if out := sans.Outline(sans.GlyphIndex('é'), 12); len(out) < 6 {
		t.Errorf("e acute: %d segments", len(out))
	}
	if !sans.HasGlyph(0x10300) {
		t.Error("U+10300 is not mapped")
	}
	if sans.Outline(sans.GlyphIndex(' '), 12) != nil || sans.Outline(-1, 12) != nil {
		t.Error("empty glyphs have segments")
	}
	if sans.AdvanceWidth(sans.GlyphIndex('H')) != 1540 || sans.AdvanceWidth(1<<20) == 0 {
		t.Error("advance")
	}
	if got := FamilyOf(font.NewAwtFont("Arial", font.AwtPlain, 12)); got != Sans {
		t.Errorf("Arial: %v", got)
	}
	if got := FamilyOf(font.NewAwtFont("Monospaced", font.AwtPlain, 12)); got != SansMono {
		t.Errorf("Monospaced: %v", got)
	}
	if got := FamilyOf(font.NewAwtFont("Serif", font.AwtBold, 12)); got != Serif {
		t.Errorf("Serif: %v", got)
	}
	if _, _, ok := UnusedSpaceCenter(font.Build("SansSerif", font.NormalFace(), 40), 'A'); ok {
		t.Error("size 40 is outside of the table")
	}
}

func TestConcurrentLoad(t *testing.T) {
	var wg sync.WaitGroup
	for i := range 24 {
		wg.Go(func() {
			f := Get(Family(i%3), font.AwtStyle(i%4))
			_ = f.Outline(f.GlyphIndex('g'), 10+i)
		})
	}
	wg.Wait()
}

func TestParseErrors(t *testing.T) {
	if _, err := parse("x", []byte("short")); err == nil {
		t.Error("short data")
	}
	if _, err := parse("x", make([]byte, 64)); err == nil {
		t.Error("no tables")
	}
	if _, err := parseCmap([]byte{0, 0, 0, 1, 0, 9, 0, 9, 0, 0, 0, 12, 0, 6, 0, 0}); err == nil {
		t.Error("no Unicode cmap")
	}
	if g := parseSimple([]byte{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 1); len(g.points) != 0 {
		t.Error("truncated glyph")
	}
}
