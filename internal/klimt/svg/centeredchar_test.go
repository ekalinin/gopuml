// SPDX-License-Identifier: MIT
// Provenance: original

package svg

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"flag"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

var updateGlyphs = flag.Bool("update-glyphs", false,
	"regenerate glyphs/*.bin from .cache/fonts/dejavu and testdata/golden/glyphs.txt.gz")

// glyphLine is a line of golden/glyphs.txt.gz: a font, the hash of its
// outlines and its UnusedSpace centers (times 4) for ' ' to '~'.
type glyphLine struct {
	family, style string
	size          int
	hash          string
	centers       [][2]int
}

func readGlyphLines(t *testing.T) []glyphLine {
	t.Helper()
	f, err := os.Open("testdata/golden/glyphs.txt.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	var lines []glyphLine
	sc := bufio.NewScanner(zr)
	sc.Buffer(nil, 1<<20)
	for sc.Scan() {
		p := strings.Split(sc.Text(), "\t")
		size, _ := strconv.Atoi(p[2])
		l := glyphLine{family: p[0], style: p[1], size: size, hash: p[3]}
		for _, c := range strings.Fields(p[4]) {
			x, y, _ := strings.Cut(c, ",")
			cx, _ := strconv.Atoi(x)
			cy, _ := strconv.Atoi(y)
			l.centers = append(l.centers, [2]int{cx, cy})
		}
		if len(l.centers) != 0x7F-0x20 {
			t.Fatalf("bad line %q", sc.Text())
		}
		lines = append(lines, l)
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return lines
}

var styleFaces = map[string]font.UFontFace{"plain": font.NormalFace(), "bold": font.BoldFace(),
	"italic": font.ItalicFace(), "bolditalic": font.BoldItalicFace()}

// TestGlyphTablesMatchUpstream compares the outlines (by hash) and the
// UnusedSpace centers of every font, style, size and character of
// golden/glyphs.txt.gz with what upstream's TextLayout and UnusedSpace give
// on the reference JVM.
func TestGlyphTablesMatchUpstream(t *testing.T) {
	if *updateGlyphs {
		t.Skip("regenerating")
	}
	lines := readGlyphLines(t)
	if len(lines) < 400 {
		t.Fatalf("only %d lines", len(lines))
	}
	failures := 0
	for _, l := range lines {
		f := font.Build(l.family, styleFaces[l.style], l.size)
		var sb strings.Builder
		for c := rune(0x20); c <= 0x7E; c++ {
			outline, ok := glyphOutline(f, c)
			if !ok {
				t.Fatalf("%s %s %d %q: no outline", l.family, l.style, l.size, c)
			}
			sb.WriteString(outlineString(outline))
			sb.WriteByte('|')
			cx, cy, ok := unusedSpaceCenter(f, c)
			if !ok || int(cx*4) != l.centers[c-0x20][0] || int(cy*4) != l.centers[c-0x20][1] {
				failures++
				t.Errorf("%s %s %d %q: center %v,%v (%v), upstream %v", l.family, l.style, l.size, c, cx, cy, ok,
					l.centers[c-0x20])
			}
		}
		h := fnv.New64a()
		h.Write([]byte(sb.String()))
		if got := strconv.FormatUint(h.Sum64(), 16); got != l.hash {
			failures++
			t.Errorf("%s %s %d: outline hash %s, upstream %s", l.family, l.style, l.size, got, l.hash)
		}
		if failures > 20 {
			t.Fatal("too many failures")
		}
	}
}

// TestCenteredCharacterSvg checks the class header spot against the
// reference SVG of class-diagram/composition-cardinalities: an ellipse of
// radius 11 centered on (25.298, 23) and its "C" in Monospaced bold 17. The
// reference rounds the center; its path places it in [25.298375, 25.2985).
func TestCenteredCharacterSvg(t *testing.T) {
	g := Build(SvgOptionBasic().WithBackcolor(color.White), false, 42, font.NewStringBounderSvg(nil),
		font.FormatSVG)
	ug := g.Apply(klimt.NewUTranslate(25.29844, 23)).Apply(color.Black)
	ug.Draw(klimt.NewUCenteredCharacter('C', font.Build("Monospaced", font.BoldFace(), 17)))
	var buf bytes.Buffer
	if err := g.WriteToStream(&buf, "", 96); err != nil {
		t.Fatal(err)
	}
	const want = `<path d="M28.267,28.641 Q27.689,28.938 27.048,29.078 Q26.408,29.234 25.705,29.234 ` +
		`Q23.205,29.234 21.877,27.594 Q20.564,25.938 20.564,22.813 Q20.564,19.688 21.877,18.031 ` +
		`Q23.205,16.375 25.705,16.375 Q26.408,16.375 27.048,16.531 Q27.705,16.688 28.267,16.984 ` +
		`L28.267,19.703 Q27.642,19.125 27.048,18.859 Q26.455,18.578 25.83,18.578 Q24.486,18.578 23.798,19.656 ` +
		`Q23.111,20.719 23.111,22.813 Q23.111,24.906 23.798,25.984 Q24.486,27.047 25.83,27.047 ` +
		`Q26.455,27.047 27.048,26.781 Q27.642,26.5 28.267,25.922 L28.267,28.641 Z " fill="#000"/>`
	if !strings.Contains(buf.String(), want) {
		t.Errorf("got %s", buf.String())
	}
	// Outside of the tables.
	e := mustPanic(t, func() { ug.Draw(klimt.NewUCenteredCharacter('é', font.Build("Monospaced", font.BoldFace(), 17))) })
	if !strings.Contains(e.Message, "not ported yet: net.sourceforge.plantuml.klimt.font.UnusedSpace") {
		t.Errorf("panic %v", e)
	}
	mustPanic(t, func() { ug.Draw(klimt.NewUCenteredCharacter('C', font.Build("SansSerif", font.NormalFace(), 40))) })
}

// TestGenerateGlyphData regenerates the embedded tables:
//
//	go test ./internal/klimt/svg -run TestGenerateGlyphData -update-glyphs
//
// glyphs/dejavu.bin holds the glyf records of ' ' to '~' of the 12 DejaVu
// faces of the reference image (.cache/fonts/dejavu, from
// tools/reference/fetch-tools.sh); glyphs/unusedspace.bin the UnusedSpace
// centers of golden/glyphs.txt.gz for sizes 4 to 32.
func TestGenerateGlyphData(t *testing.T) {
	if !*updateGlyphs {
		t.Skip("use -update-glyphs to regenerate")
	}
	var out bytes.Buffer
	out.WriteString("PUGLYPH1")
	out.WriteByte(byte(len(glyphFaceNames)))
	for _, name := range glyphFaceNames {
		ttf, err := readTTF(filepath.Join("..", "..", "..", ".cache", "fonts", "dejavu", name+".ttf"))
		if err != nil {
			t.Fatal(err)
		}
		out.WriteByte(byte(len(name)))
		out.WriteString(name)
		binary.Write(&out, binary.BigEndian, uint16(ttf.unitsPerEm))
		binary.Write(&out, binary.BigEndian, uint16(0x20))
		binary.Write(&out, binary.BigEndian, uint16(0x7F-0x20))
		for c := rune(0x20); c <= 0x7E; c++ {
			g := ttf.glyphIndex(c)
			if g == 0 {
				t.Fatalf("%s: no glyph for %q", name, c)
			}
			rec := ttf.glyf(g)
			shift := 0
			if len(rec) > 0 {
				if int16(binary.BigEndian.Uint16(rec)) < 0 {
					t.Fatalf("%s %q: composite glyph", name, c)
				}
				shift = int(int16(binary.BigEndian.Uint16(rec[2:]))) - ttf.lsb(g)
			}
			binary.Write(&out, binary.BigEndian, int16(shift))
			binary.Write(&out, binary.BigEndian, uint16(len(rec)))
			out.Write(rec)
		}
	}
	if err := os.WriteFile("glyphs/dejavu.bin", out.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	const minSize, maxSize = 4, 32
	familyFace := map[string]int{"SansSerif": 0, "Monospaced": 1, "Serif": 2}
	styleIndex := map[string]int{"plain": 0, "bold": 1, "italic": 2, "bolditalic": 3}
	centers := make([]byte, len(glyphFaceNames)*(maxSize-minSize+1)*(0x7F-0x20)*2)
	seen := map[int]bool{}
	byKey := map[string][][2]int{}
	for _, l := range readGlyphLines(t) {
		byKey[fmt.Sprint(l.family, l.style, l.size)] = l.centers
		fam, ok := familyFace[l.family]
		if !ok || l.size < minSize || l.size > maxSize {
			continue
		}
		face := fam*4 + styleIndex[l.style]
		row := face*(maxSize-minSize+1) + l.size - minSize
		seen[row] = true
		for i, c := range l.centers {
			if c[0] < -128 || c[0] > 127 || c[1] < -128 || c[1] > 127 {
				t.Fatalf("center %v out of range", c)
			}
			centers[2*(row*(0x7F-0x20)+i)] = byte(int8(c[0]))
			centers[2*(row*(0x7F-0x20)+i)+1] = byte(int8(c[1]))
		}
	}
	if len(seen) != len(glyphFaceNames)*(maxSize-minSize+1) {
		t.Fatalf("%d rows", len(seen))
	}
	// The other names of the faces render the same glyphs.
	aliases := map[string]string{"Dialog": "SansSerif", "DialogInput": "Monospaced", "DejaVu Sans": "SansSerif",
		"DejaVu Serif": "Serif", "DejaVu Sans Mono": "Monospaced"}
	for alias, family := range aliases {
		for style := range styleIndex {
			for _, size := range []int{9, 12, 14, 17, 23} {
				a, b := byKey[fmt.Sprint(alias, style, size)], byKey[fmt.Sprint(family, style, size)]
				if fmt.Sprint(a) != fmt.Sprint(b) {
					t.Errorf("%s %s %d differs from %s", alias, style, size, family)
				}
			}
		}
	}
	var table bytes.Buffer
	table.WriteString("PUUNUSE1")
	table.WriteByte(byte(len(glyphFaceNames)))
	table.WriteByte(minSize)
	table.WriteByte(maxSize)
	binary.Write(&table, binary.BigEndian, uint16(0x20))
	binary.Write(&table, binary.BigEndian, uint16(0x7F-0x20))
	table.Write(centers)
	if err := os.WriteFile("glyphs/unusedspace.bin", table.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

// ttfFont is the part of a TrueType font the generator reads.
type ttfFont struct {
	tables     map[string][]byte
	unitsPerEm int
}

func readTTF(path string) (*ttfFont, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f := &ttfFont{tables: map[string][]byte{}}
	n := int(binary.BigEndian.Uint16(data[4:]))
	for i := range n {
		rec := data[12+16*i:]
		off := binary.BigEndian.Uint32(rec[8:])
		length := binary.BigEndian.Uint32(rec[12:])
		f.tables[string(rec[:4])] = data[off : off+length]
	}
	f.unitsPerEm = int(binary.BigEndian.Uint16(f.tables["head"][18:]))
	return f, nil
}

func be16(b []byte, o int) int { return int(binary.BigEndian.Uint16(b[o:])) }

// glyphIndex looks a BMP character up in the format 4 Unicode cmap.
func (f *ttfFont) glyphIndex(c rune) int {
	cmap := f.tables["cmap"]
	for i := range be16(cmap, 2) {
		if be16(cmap, 4+8*i) != 3 || be16(cmap, 6+8*i) != 1 {
			continue
		}
		t := cmap[binary.BigEndian.Uint32(cmap[8+8*i:]):]
		segX2 := be16(t, 6)
		for s := 0; s < segX2/2; s++ {
			end, start := be16(t, 14+2*s), be16(t, 16+segX2+2*s)
			if int(c) < start || int(c) > end {
				continue
			}
			delta, rangeOffset := be16(t, 16+2*segX2+2*s), be16(t, 16+3*segX2+2*s)
			if rangeOffset == 0 {
				return (int(c) + delta) & 0xffff
			}
			g := be16(t, 16+3*segX2+2*s+rangeOffset+2*(int(c)-start))
			if g == 0 {
				return 0
			}
			return (g + delta) & 0xffff
		}
	}
	return 0
}

func (f *ttfFont) glyf(g int) []byte {
	loca := f.tables["loca"]
	var a, b int
	if int16(binary.BigEndian.Uint16(f.tables["head"][50:])) == 0 {
		a, b = 2*be16(loca, 2*g), 2*be16(loca, 2*g+2)
	} else {
		a, b = int(binary.BigEndian.Uint32(loca[4*g:])), int(binary.BigEndian.Uint32(loca[4*g+4:]))
	}
	return f.tables["glyf"][a:b]
}

func (f *ttfFont) lsb(g int) int {
	hmtx := f.tables["hmtx"]
	n := be16(f.tables["hhea"], 34)
	if g < n {
		return int(int16(binary.BigEndian.Uint16(hmtx[4*g+2:])))
	}
	return int(int16(binary.BigEndian.Uint16(hmtx[4*n+2*(g-n):])))
}
