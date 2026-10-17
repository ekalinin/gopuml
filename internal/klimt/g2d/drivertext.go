// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/g2d/DriverTextG2d.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/g2d/DriverTextAsPathG2d.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/text/StyledString.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package g2d

import (
	"math"
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// The rich text markers of upstream's StringUtils.
const (
	boldStart = '\uEEF2'
	boldEnd   = '\uEEF1'
)

// styledString is a part of a rich text: plain, or bold between the
// markers.
type styledString struct {
	text string
	bold bool
}

// buildStyledStrings is upstream's StyledString.build. As there, the end
// marker is searched from the start of the text, so an end marker before
// the start marker cuts the bold part short; here such a cut (which fails
// with an index error upstream) gives an empty bold part.
func buildStyledStrings(s string) []styledString {
	var result []styledString
	for len(s) > 0 {
		i1 := strings.IndexRune(s, boldStart)
		if i1 == -1 {
			result = append(result, styledString{text: s})
			break
		}
		i2 := strings.IndexRune(s, boldEnd)
		if i1 > 0 {
			result = append(result, styledString{text: s[:i1]})
		}
		start := i1 + len(string(boldStart))
		switch {
		case i2 == -1:
			result = append(result, styledString{text: s[start:], bold: true})
			s = ""
		case i2 < start:
			result = append(result, styledString{bold: true})
			s = s[i2+len(string(boldEnd)):]
		default:
			result = append(result, styledString{text: s[start:i2], bold: true})
			s = s[i2+len(string(boldEnd)):]
		}
	}
	return result
}

// driverText draws texts with Java2D's text rendering.
type driverText struct {
	visible       geom.EnsureVisible
	stringBounder font.StringBounder
}

func (d driverText) Draw(shape *klimt.UText, x, y float64, mapper *color.ColorMapper, param klimt.UParam,
	g *graphics) {
	fc := shape.FontConfiguration()

	if fc.Color().IsTransparent() {
		return
	}

	for _, ss := range buildStyledStrings(shape.Text()) {
		cfg := fc
		if ss.bold {
			cfg = fc.Bold()
		}
		x += d.printSingleText(g, cfg, ss.text, x, y, mapper, shape.Orientation())
	}
}

func (d driverText) printSingleText(g *graphics, fc *klimt.FontConfiguration, text string, x, y float64,
	mapper *color.ColorMapper, orientation int) float64 {
	f := fc.Font()
	extended := fc.ExtendedColor()

	dim := d.stringBounder.CalculateDimension(f, text)
	height := math.Max(10, dim.Height)
	width := dim.Width

	// Upstream sets text antialiasing and fractional metrics here; they
	// stay on for what is drawn next.
	g.fracMetrics = true
	awt := f.UnderlyingFont(text)

	switch orientation {
	case 90:
		orig := g.tx
		g.tx = g.tx.translate(x, y).rotate(-math.Pi / 2)
		g.drawString(text, awt, 0, 0)
		g.tx = orig
	case 270:
		orig := g.tx
		g.tx = g.tx.translate(x, y).rotate(math.Pi / 2)
		g.drawString(text, awt, 0, 0)
		g.tx = orig
	case 0:
		if fc.ContainsStyle(font.FontStyleBackcolor) {
			area := rect2D{x, y - height + 1.5, width, height}
			if isGradient(extended) {
				g.setPaint(paintGradient(x, y, mapper, width, height, extended, g.tx))
				g.fill(area)
			} else {
				backColor := extended.ToColor(mapper)
				g.setColor(argb(backColor))
				g.fill(area)
			}
		}
		d.visible.EnsureVisible(x, y-height+1.5)
		d.visible.EnsureVisible(x+width, y+1.5)

		g.setColor(argb(fc.Color().ToColor(mapper)))
		g.drawString(text, awt, float64(float32(x)), float64(float32(y)))

		if fc.ContainsStyle(font.FontStyleUnderline) {
			if extended != nil {
				g.setColor(argb(extended.ToColor(mapper)))
			}
			// Upstream fails when the configuration has no underline
			// stroke; here there is no underline (design D16).
			if stroke := fc.UnderlineStroke(); stroke != nil && stroke.Thickness() > 0 {
				manageStroke(*stroke, g)
				ypos := int(y + 2.5)
				g.drawLine(int(x), ypos, int(x+width), ypos)
				g.setStroke(newBasicStroke(1))
			}
		}
		if fc.ContainsStyle(font.FontStyleWave) {
			drawWave(g, extended, mapper, x, y, width)
		}
		if fc.ContainsStyle(font.FontStyleStrike) {
			drawStrike(g, extended, mapper, awt, x, y, width)
		}
	}
	return width
}

// drawWave draws the wave under a text, with the current stroke.
func drawWave(g *graphics, extended color.HColor, mapper *color.ColorMapper, x, y, width float64) {
	ypos := int(y+2.5) - 1
	if extended != nil {
		g.setColor(argb(extended.ToColor(mapper)))
	}
	for i := int(x); float64(i) < x+width-5; i += 6 {
		g.drawLine(i, ypos-0, i+3, ypos+1)
		g.drawLine(i+3, ypos+1, i+6, ypos-0)
	}
}

// drawStrike draws the line through a text, at FontMetrics' integer descent
// above the baseline.
func drawStrike(g *graphics, extended color.HColor, mapper *color.ColorMapper, awt font.AwtFont, x, y,
	width float64) {
	descent := int(float32(0.95) + float32(awt.Descent()))
	ypos := int(y - float64(descent) - 0.5)
	if extended != nil {
		g.setColor(argb(extended.ToColor(mapper)))
	}
	g.setStroke(newBasicStroke(float64(float32(1.5))))
	g.drawLine(int(x), ypos, int(x+width), ypos)
	g.setStroke(newBasicStroke(1))
}

// driverTextAsPath draws texts as filled outlines, which scale with the
// dpi.
type driverTextAsPath struct {
	visible       geom.EnsureVisible
	stringBounder font.StringBounder
}

func (d driverTextAsPath) Draw(shape *klimt.UText, x, y float64, mapper *color.ColorMapper, param klimt.UParam,
	g *graphics) {
	fc := shape.FontConfiguration()

	f := fc.Font()
	text := shape.Text()
	dim := d.stringBounder.CalculateDimension(f, text)
	height := math.Max(10, dim.Height)
	width := dim.Width

	if fc.ContainsStyle(font.FontStyleBackcolor) {
		extended := fc.ExtendedColor().ToColor(mapper)
		g.setColor(argb(extended))
		g.fill(rect2D{x, y - height + 1.5, width, height})
	}
	d.visible.EnsureVisible(x, y-height+1.5)
	d.visible.EnsureVisible(x+width, y+1.5)

	awt := f.UnderlyingFont(text)
	g.setColor(argb(fc.Color().ToColor(mapper)))
	orig := g.tx
	g.tx = g.tx.translate(x, y)
	g.fill(pathShape{g.textOutline(text, awt, 0, 0)})
	g.tx = orig

	if fc.ContainsStyle(font.FontStyleUnderline) {
		extended := fc.ExtendedColor()
		if extended != nil {
			g.setColor(argb(extended.ToColor(mapper)))
		}
		ypos := int(y + 2.5)
		g.setStroke(newBasicStroke(1))
		g.drawLine(int(x), ypos, int(x+width), ypos)
		g.setStroke(newBasicStroke(1))
	}
	if fc.ContainsStyle(font.FontStyleWave) {
		drawWave(g, fc.ExtendedColor(), mapper, x, y, width)
	}
	if fc.ContainsStyle(font.FontStyleStrike) {
		drawStrike(g, fc.ExtendedColor(), mapper, awt, x, y, width)
	}
}
