// SPDX-License-Identifier: MIT
// Provenance: original

package g2d

import (
	"math"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/g2d/dejavu"
)

// This file draws text as Java2D does on the reference JVM with text
// antialiasing and fractional metrics (the hints DriverTextG2d sets):
//
//   - drawString: each glyph is rasterized by FreeType from its unhinted
//     outline at the font size, with exact coverage, and blitted at its
//     pen position rounded to whole pixels, floor(x + 0.5); the pen
//     advances by the fractional advances the font metrics give.
//   - text as path (DriverTextAsPathG2d): the outlines at their fractional
//     positions are filled as a shape, so they scale with the transform.
//
// Glyphs come from the embedded DejaVu fonts; a character the face lacks is
// taken from the other embedded families, else drawn as the missing glyph,
// as Java's logical fonts do with their fallback fonts.

// glyphKey identifies a cached outline.
type glyphKey struct {
	face *dejavu.Face
	gid  int
	ppem int
}

// glyphCache keeps the outlines a render uses.
type glyphCache map[glyphKey][]dejavu.Segment

func (c glyphCache) outline(face *dejavu.Face, gid, ppem int) []dejavu.Segment {
	k := glyphKey{face, gid, ppem}
	if s, ok := c[k]; ok {
		return s
	}
	s := face.Outline(gid, ppem)
	c[k] = s
	return s
}

// glyphFor returns the face and glyph Java draws a code point with.
// Control characters are invisible.
func glyphFor(f font.AwtFont, cp rune) (*dejavu.Face, int, bool) {
	if cp < 0x20 || cp == 0x7F {
		return nil, 0, false
	}
	primary := dejavu.ForAwtFont(f)
	if gid := primary.GlyphIndex(cp); gid != 0 {
		return primary, gid, true
	}
	if f.CanDisplay(cp) {
		for _, fb := range primary.Fallbacks() {
			if gid := fb.GlyphIndex(cp); gid != 0 {
				return fb, gid, true
			}
		}
	}
	return primary, 0, true
}

// drawString draws text with its baseline origin at (x, y) in user space,
// in the current color, as Graphics2D.drawString with antialiasing. With
// fractional metrics off (the default of a Graphics2D, until a text driver
// sets them on) the pen advances by whole pixels. Java then also hints the
// glyphs, which is not reproduced: the outlines stay unhinted.
func (g *graphics) drawString(text string, f font.AwtFont, x, y float64) {
	if g.glyphs == nil {
		g.glyphs = glyphCache{}
	}
	// The transforms around text are translations, quarter turns and the
	// dpi scaling; the glyph size follows the scale.
	scale := math.Hypot(g.tx.m00, g.tx.m10)
	ppem := int(math.Floor(float64(float64(f.Size())*scale) + 0.5))
	if ppem <= 0 {
		return
	}
	// Unit vectors of the rotation, exact for quarter turns.
	r00, r10, r01, r11 := g.tx.m00/scale, g.tx.m10/scale, g.tx.m01/scale, g.tx.m11/scale
	ox, oy := g.tx.apply(x, y)
	penX, penY := float32(ox), float32(oy)
	for _, cp := range text {
		face, gid, visible := glyphFor(f, cp)
		adv := f.Advance(cp)
		if !g.fracMetrics && face != nil {
			adv = float64(hintedAdvance(f, face, gid, cp))
		}
		dax, day := g.tx.delta(adv, 0)
		if visible {
			gx := math.Floor(float64(penX) + 0.5)
			gy := math.Floor(float64(penY) + 0.5)
			g.fillGlyph(g.glyphs.outline(face, gid, ppem), gx, gy, r00, r10, r01, r11)
		}
		penX += float32(dax)
		penY += float32(day)
	}
}

// fillGlyph rasterizes an outline in 26.6, rotated by the unit matrix r and
// placed at the integer device position (ox, oy), with FreeType's coverage
// and conic flattening, and blends it with the current paint.
func (g *graphics) fillGlyph(segs []dejavu.Segment, ox, oy, r00, r10, r01, r11 float64) {
	if len(segs) == 0 {
		return
	}
	pt := func(p dejavu.Point) (float64, float64) {
		x, y := float64(p.X)/64, float64(p.Y)/64
		return ox + float64(x*r00) + float64(y*r01), oy + float64(x*r10) + float64(y*r11)
	}
	var edges []edge
	var cx, cy, sx, sy float64
	for _, s := range segs {
		switch s.Op {
		case dejavu.MoveTo:
			cx, cy = pt(s.P[0])
			sx, sy = cx, cy
		case dejavu.LineTo:
			x, y := pt(s.P[0])
			edges = append(edges, edge{cx, cy, x, y})
			cx, cy = x, y
		case dejavu.QuadTo:
			qx, qy := pt(s.P[0])
			x, y := pt(s.P[1])
			edges = appendConic(edges, cx, cy, qx, qy, x, y)
			cx, cy = x, y
		case dejavu.Close:
			if cx != sx || cy != sy {
				edges = append(edges, edge{cx, cy, sx, sy})
			}
			cx, cy = sx, sy
		}
	}
	g.composite(rasterExact(edges, g.clipRect, false, alphaFreeType))
}

// appendConic appends the lines of a quadratic curve as FreeType's gray
// rasterizer draws it: 2^n lines of equal parameter steps, n the number of
// times the second difference must be quartered to fall within a quarter
// pixel.
func appendConic(edges []edge, x0, y0, x1, y1, x2, y2 float64) []edge {
	dx := math.Abs(x0 + x2 - 2*x1)
	dy := math.Abs(y0 + y2 - 2*y1)
	d := int64(math.Max(dx, dy) * 256)
	n := 1
	for d > 64 {
		d >>= 2
		n <<= 1
	}
	px, py := x0, y0
	for i := 1; i <= n; i++ {
		var x, y float64
		if i == n {
			x, y = x2, y2
		} else {
			t := float64(i) / float64(n)
			x, y = quadAt(x0, x1, x2, t), quadAt(y0, y1, y2, t)
		}
		edges = append(edges, edge{px, py, x, y})
		px, py = x, y
	}
	return edges
}

// textOutline returns the outline TextLayout.getOutline gives for text:
// the glyph outlines at their fractional pen positions from (x, y), in user
// space.
func (g *graphics) textOutline(text string, f font.AwtFont, x, y float64) *path {
	if g.glyphs == nil {
		g.glyphs = glyphCache{}
	}
	p := &path{}
	pen := float32(0)
	for _, cp := range text {
		face, gid, visible := glyphFor(f, cp)
		if visible {
			px := x + float64(pen)
			for _, s := range g.glyphs.outline(face, gid, f.Size()) {
				switch s.Op {
				case dejavu.MoveTo:
					p.moveTo(px+float64(float64(s.P[0].X)/64), y+float64(float64(s.P[0].Y)/64))
				case dejavu.LineTo:
					p.lineTo(px+float64(float64(s.P[0].X)/64), y+float64(float64(s.P[0].Y)/64))
				case dejavu.QuadTo:
					p.quadTo(px+float64(float64(s.P[0].X)/64), y+float64(float64(s.P[0].Y)/64), px+float64(float64(s.P[1].X)/64),
						y+float64(float64(s.P[1].Y)/64))
				case dejavu.Close:
					p.closePath()
				}
			}
		}
		pen += float32(f.Advance(cp))
	}
	return p
}

// glyphCenter returns the center of the bounds of a glyph outline,
// relative to its origin.
func glyphCenter(f font.AwtFont, cp rune) (float64, float64) {
	face, gid, visible := glyphFor(f, cp)
	if !visible {
		return 0, 0
	}
	segs := face.Outline(gid, f.Size())
	first := true
	var x0, y0, x1, y1 int64
	for _, s := range segs {
		n := 0
		switch s.Op {
		case dejavu.MoveTo, dejavu.LineTo:
			n = 1
		case dejavu.QuadTo:
			n = 2
		}
		for _, p := range s.P[:n] {
			if first {
				x0, y0, x1, y1, first = p.X, p.Y, p.X, p.Y, false
				continue
			}
			x0, y0 = min(x0, p.X), min(y0, p.Y)
			x1, y1 = max(x1, p.X), max(y1, p.Y)
		}
	}
	if first {
		return 0, 0
	}
	return float64(x0+x1) / 128, float64(y0+y1) / 128
}

// stringWidth returns the width of a text as FontMetrics.getStringBounds
// gives it, with or without fractional metrics.
func (g *graphics) stringWidth(text string, f font.AwtFont) float64 {
	if g.fracMetrics {
		return f.StringBounds(text).Width
	}
	w := 0
	for _, cp := range text {
		face, gid, _ := glyphFor(f, cp)
		if face == nil {
			face, gid = dejavu.ForAwtFont(f), 0
		}
		w += hintedAdvance(f, face, gid, cp)
	}
	return float64(w)
}

// dialog12Advances are the advances in whole pixels of the printable ASCII
// characters in the default font of a Graphics2D (Dialog, plain, 12)
// without fractional metrics, as FontMetrics.charWidth gives them on the
// reference JVM: hinted by FreeType, which the scaled advances do not
// reproduce.
var dialog12Advances = [...]uint8{4, 5, 5, 10, 8, 11, 10, 3, 5, 5, 6, 10, 4, 4, 4, 4, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 4, 4,
	10, 10, 10, 6, 13, 8, 8, 8, 9, 8, 7, 9, 9, 3, 3, 7, 6, 10, 9, 9, 8, 9, 8, 8, 7, 9, 8, 11, 7, 7, 9, 5, 4, 5, 10, 6, 6, 8, 8,
	7, 8, 8, 4, 8, 8, 3, 3, 7, 3, 11, 8, 8, 8, 8, 5, 7, 5, 8, 6, 9, 6, 6, 5, 8, 4, 8, 10}

// hintedAdvance returns the advance of a glyph without fractional metrics:
// the table of the default font, else the scaled advance rounded to the
// pixel (hinting may make it one pixel wider).
func hintedAdvance(f font.AwtFont, face *dejavu.Face, gid int, cp rune) int {
	if f.Size() == 12 && face.Family() == dejavu.Sans && face.Style() == font.AwtPlain && cp >= 0x20 && cp <= 0x7E {
		return int(dialog12Advances[cp-0x20])
	}
	return face.HintedAdvance(gid, f.Size())
}
