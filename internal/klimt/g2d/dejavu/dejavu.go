// SPDX-License-Identifier: MIT
// Provenance: original

// Package dejavu embeds the DejaVu fonts behind Java's logical fonts on the
// reference image (fonts-dejavu 2.37) and loads their glyph outlines as
// Java gets them from FreeType: unhinted, scaled with FreeType's fixed
// point arithmetic, in 26.6 fixed point.
//
// The raster driver draws text with these outlines; text is measured with
// the metrics tables of internal/klimt/font, so the outlines are only used
// for their shapes. Host fonts are never read.
//
// The fonts are the unmodified TrueType files, compressed with gzip
// (fonts/*.ttf.gz); their license is fonts/LICENSE-DejaVu.txt (Bitstream
// Vera and Arev fonts licenses, DejaVu changes in the public domain).
package dejavu

import (
	"bytes"
	"compress/gzip"
	"embed"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
)

//go:embed fonts/*.ttf.gz
var fontFiles embed.FS

// Family is one of the three embedded DejaVu families.
type Family int

// The embedded families.
const (
	Sans Family = iota
	SansMono
	Serif
)

// faceNames are the PostScript names of the faces, by family then Java
// style (plain, bold, italic, bold italic).
var faceNames = [3][4]string{
	{"DejaVuSans", "DejaVuSans-Bold", "DejaVuSans-Oblique", "DejaVuSans-BoldOblique"},
	{"DejaVuSansMono", "DejaVuSansMono-Bold", "DejaVuSansMono-Oblique", "DejaVuSansMono-BoldOblique"},
	{"DejaVuSerif", "DejaVuSerif-Bold", "DejaVuSerif-Italic", "DejaVuSerif-BoldItalic"},
}

// Face is a parsed TrueType font. It is read-only once loaded and safe for
// concurrent use.
type Face struct {
	family Family
	style  font.AwtStyle
	name   string

	unitsPerEm  int
	headFlags   uint16
	numGlyphs   int
	longLoca    bool
	numHMetrics int
	loca        []byte
	glyf        []byte
	hmtx        []byte
	cmap        cmapTable
}

var (
	faceOnce [3][4]sync.Once
	faces    [3][4]*Face
)

// Get returns the face of a family and style; only the bold and italic bits
// of style are used. The face is decompressed and parsed on first use.
func Get(family Family, style font.AwtStyle) *Face {
	s := int(style & font.AwtBoldItalic)
	faceOnce[family][s].Do(func() {
		name := faceNames[family][s]
		f, err := load(name)
		if err != nil {
			panic(fmt.Sprintf("dejavu: %s: %v", name, err))
		}
		f.family, f.style = family, font.AwtStyle(s)
		faces[family][s] = f
	})
	return faces[family][s]
}

// FamilyOf returns the DejaVu family Java draws an AWT font with on the
// reference image: DejaVu Sans for SansSerif and Dialog (which unknown font
// names resolve to), DejaVu Sans Mono for Monospaced and DialogInput,
// DejaVu Serif for Serif.
func FamilyOf(f font.AwtFont) Family {
	switch f.Family() {
	case "Monospaced", "DialogInput", "DejaVu Sans Mono":
		return SansMono
	case "Serif", "DejaVu Serif":
		return Serif
	}
	return Sans
}

// ForAwtFont returns the face Java draws an AWT font with.
func ForAwtFont(f font.AwtFont) *Face { return Get(FamilyOf(f), f.Style()) }

// Name returns the PostScript name of the face, for example
// "DejaVuSans-Bold".
func (f *Face) Name() string { return f.name }

// Family returns the family of the face.
func (f *Face) Family() Family { return f.family }

// Style returns the Java style of the face.
func (f *Face) Style() font.AwtStyle { return f.style }

// UnitsPerEm returns the design units per em.
func (f *Face) UnitsPerEm() int { return f.unitsPerEm }

// NumGlyphs returns the number of glyphs.
func (f *Face) NumGlyphs() int { return f.numGlyphs }

// Fallbacks returns the faces searched, after f, for a character f lacks:
// the other families in the same style, then DejaVu Sans plain. Java's
// logical fonts fall back to the fonts fontconfig lists; on the reference
// image those are DejaVu fonts too, so the embedded families cover most of
// what they display.
func (f *Face) Fallbacks() []*Face {
	var order []Family
	switch f.family {
	case Sans:
		order = []Family{Serif, SansMono}
	case SansMono:
		order = []Family{Sans, Serif}
	default:
		order = []Family{Sans, SansMono}
	}
	result := make([]*Face, 0, 3)
	for _, fam := range order {
		result = append(result, Get(fam, f.style))
	}
	if f.style != font.AwtPlain {
		result = append(result, Get(Sans, font.AwtPlain))
	}
	return result
}

// errFont reports a malformed font file.
var errFont = errors.New("malformed TrueType data")

func load(name string) (*Face, error) {
	gz, err := fontFiles.ReadFile("fonts/" + name + ".ttf.gz")
	if err != nil {
		return nil, err
	}
	zr, err := gzip.NewReader(bytes.NewReader(gz))
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(zr)
	if err != nil {
		return nil, err
	}
	return parse(name, data)
}

// parse reads the tables the outlines need: head, maxp, hhea, hmtx, loca,
// glyf and cmap.
func parse(name string, data []byte) (*Face, error) {
	if len(data) < 12 {
		return nil, errFont
	}
	n := int(binary.BigEndian.Uint16(data[4:]))
	tables := map[string][]byte{}
	for i := range n {
		at := 12 + 16*i
		if at+16 > len(data) {
			return nil, errFont
		}
		rec := data[at:]
		off := int(binary.BigEndian.Uint32(rec[8:]))
		length := int(binary.BigEndian.Uint32(rec[12:]))
		if off < 0 || length < 0 || off+length > len(data) {
			return nil, errFont
		}
		tables[string(rec[:4])] = data[off : off+length]
	}
	for _, t := range []string{"head", "maxp", "hhea", "hmtx", "loca", "glyf", "cmap"} {
		if tables[t] == nil {
			return nil, fmt.Errorf("%w: no %s table", errFont, t)
		}
	}
	head, maxp, hhea := tables["head"], tables["maxp"], tables["hhea"]
	if len(head) < 54 || len(maxp) < 6 || len(hhea) < 36 {
		return nil, errFont
	}
	f := &Face{
		name:        name,
		unitsPerEm:  int(binary.BigEndian.Uint16(head[18:])),
		headFlags:   binary.BigEndian.Uint16(head[16:]),
		longLoca:    binary.BigEndian.Uint16(head[50:]) != 0,
		numGlyphs:   int(binary.BigEndian.Uint16(maxp[4:])),
		numHMetrics: int(binary.BigEndian.Uint16(hhea[34:])),
		loca:        tables["loca"],
		glyf:        tables["glyf"],
		hmtx:        tables["hmtx"],
	}
	if f.unitsPerEm == 0 || f.numHMetrics == 0 || len(f.hmtx) < 4*f.numHMetrics {
		return nil, errFont
	}
	cmap, err := parseCmap(tables["cmap"])
	if err != nil {
		return nil, err
	}
	f.cmap = cmap
	return f, nil
}

// GlyphIndex returns the glyph of a character, 0 (the missing glyph) when
// the face has none.
func (f *Face) GlyphIndex(r rune) int {
	g := f.cmap.lookup(r)
	if g >= f.numGlyphs {
		return 0
	}
	return g
}

// HasGlyph reports whether the face maps r to a glyph.
func (f *Face) HasGlyph(r rune) bool { return f.GlyphIndex(r) != 0 }

// glyphData returns the glyf record of a glyph, nil for an empty glyph.
func (f *Face) glyphData(gid int) []byte {
	if gid < 0 || gid >= f.numGlyphs {
		return nil
	}
	var start, end int
	if f.longLoca {
		if 4*gid+8 > len(f.loca) {
			return nil
		}
		start = int(binary.BigEndian.Uint32(f.loca[4*gid:]))
		end = int(binary.BigEndian.Uint32(f.loca[4*gid+4:]))
	} else {
		if 2*gid+4 > len(f.loca) {
			return nil
		}
		start = 2 * int(binary.BigEndian.Uint16(f.loca[2*gid:]))
		end = 2 * int(binary.BigEndian.Uint16(f.loca[2*gid+2:]))
	}
	if start >= end || end > len(f.glyf) {
		return nil
	}
	return f.glyf[start:end]
}

// leftSideBearing returns the left side bearing of a glyph in font units.
func (f *Face) leftSideBearing(gid int) int {
	if gid < f.numHMetrics {
		return int(int16(binary.BigEndian.Uint16(f.hmtx[4*gid+2:])))
	}
	at := 4*f.numHMetrics + 2*(gid-f.numHMetrics)
	if at+2 > len(f.hmtx) {
		return 0
	}
	return int(int16(binary.BigEndian.Uint16(f.hmtx[at:])))
}

// AdvanceWidth returns the advance width of a glyph in font units.
func (f *Face) AdvanceWidth(gid int) int {
	if gid >= f.numHMetrics {
		gid = f.numHMetrics - 1
	}
	return int(binary.BigEndian.Uint16(f.hmtx[4*gid:]))
}

// cmapTable maps characters to glyphs: a format 12 subtable when the font
// has one, else a format 4 subtable.
type cmapTable struct {
	format int
	data   []byte
}

func parseCmap(t []byte) (cmapTable, error) {
	if len(t) < 4 {
		return cmapTable{}, errFont
	}
	n := int(binary.BigEndian.Uint16(t[2:]))
	var best cmapTable
	bestRank := 0
	for i := range n {
		at := 4 + 8*i
		if at+8 > len(t) {
			return cmapTable{}, errFont
		}
		platform := binary.BigEndian.Uint16(t[at:])
		encoding := binary.BigEndian.Uint16(t[at+2:])
		off := int(binary.BigEndian.Uint32(t[at+4:]))
		if off+4 > len(t) {
			continue
		}
		format := int(binary.BigEndian.Uint16(t[off:]))
		rank := 0
		switch {
		case format == 12 && (platform == 3 && encoding == 10 || platform == 0):
			rank = 3
		case format == 4 && (platform == 3 && encoding == 1 || platform == 0):
			rank = 2
		}
		if rank > bestRank {
			bestRank = rank
			best = cmapTable{format: format, data: t[off:]}
		}
	}
	if bestRank == 0 {
		return cmapTable{}, fmt.Errorf("%w: no Unicode cmap", errFont)
	}
	return best, nil
}

func (c cmapTable) lookup(r rune) int {
	d := c.data
	if c.format == 12 {
		if len(d) < 16 {
			return 0
		}
		n := int(binary.BigEndian.Uint32(d[12:]))
		lo, hi := 0, n
		for lo < hi {
			mid := (lo + hi) / 2
			g := d[16+12*mid:]
			if len(g) < 12 {
				return 0
			}
			start := rune(binary.BigEndian.Uint32(g))
			end := rune(binary.BigEndian.Uint32(g[4:]))
			switch {
			case r < start:
				hi = mid
			case r > end:
				lo = mid + 1
			default:
				return int(binary.BigEndian.Uint32(g[8:])) + int(r-start)
			}
		}
		return 0
	}
	// Format 4: segments of BMP code points.
	if r > 0xFFFF || len(d) < 14 {
		return 0
	}
	segX2 := int(binary.BigEndian.Uint16(d[6:]))
	ends := d[14:]
	starts := d[16+segX2:]
	deltas := d[16+2*segX2:]
	offsets := d[16+3*segX2:]
	if len(offsets) < segX2 {
		return 0
	}
	c16 := uint16(r)
	lo, hi := 0, segX2/2
	for lo < hi {
		mid := (lo + hi) / 2
		end := binary.BigEndian.Uint16(ends[2*mid:])
		if end < c16 {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo >= segX2/2 {
		return 0
	}
	start := binary.BigEndian.Uint16(starts[2*lo:])
	if c16 < start {
		return 0
	}
	delta := binary.BigEndian.Uint16(deltas[2*lo:])
	ro := int(binary.BigEndian.Uint16(offsets[2*lo:]))
	if ro == 0 {
		return int(c16 + delta)
	}
	at := 2*lo + ro + 2*int(c16-start)
	if at+2 > len(offsets) {
		return 0
	}
	g := binary.BigEndian.Uint16(offsets[at:])
	if g == 0 {
		return 0
	}
	return int(g + delta)
}
