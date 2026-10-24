// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/drawing/svg/DriverTextSvg.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/drawing/svg/DriverTextAsPathSvg.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/drawing/svg/DriverCenteredCharacterSvg.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/drawing/svg/DriverCenteredCharacterSvgDeterministic.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svg

import (
	"strconv"
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/text"
)

// extraLines are the underlines and strike-throughs of a custom color,
// which CSS text-decoration cannot draw: they become lines.
type extraLines struct {
	colors  []color.HColor
	deltaYs []float64
}

func (e *extraLines) add(c color.HColor, deltaY float64) {
	e.colors = append(e.colors, c)
	e.deltaYs = append(e.deltaYs, deltaY)
}

func (e *extraLines) drawAll(x, y, width float64, f font.UFont, mapper *color.ColorMapper, svg *SvgGraphics) {
	for i, c := range e.colors {
		svg.SetStrokeColor(c.ToSvg(mapper))
		svg.SetStrokeWidth(f.Size2D()/28.0, nil)
		svg.SvgLine(x, y+e.deltaYs[i], x+width, y+e.deltaYs[i], 0)
	}
}

// DriverTextSvg draws texts as <text> elements; nothing is drawn when the
// start point is outside the clip or the color is transparent.
type DriverTextSvg struct {
	stringBounder font.StringBounder
	clipContainer klimt.ClipContainer
}

// NewDriverTextSvg returns the driver.
func NewDriverTextSvg(stringBounder font.StringBounder, clipContainer klimt.ClipContainer) DriverTextSvg {
	return DriverTextSvg{stringBounder: stringBounder, clipContainer: clipContainer}
}

// Draw implements klimt.UDriver.
//
// Upstream reads the thickness of the hyperlink underline stroke of an
// underlined text, which is null with "skinparam hyperlinkUnderline false"
// and throws NullPointerException; here a missing stroke draws no underline
// (design D16).
func (d DriverTextSvg) Draw(shape *klimt.UText, x, y float64, mapper *color.ColorMapper, param klimt.UParam,
	svg *SvgGraphics) {
	if clip := d.clipContainer.Clip(); clip != nil && !clip.IsInside(x, y) {
		return
	}

	fc := shape.FontConfiguration()
	if fc.Color().IsTransparent() {
		return
	}

	f := fc.Font()
	face := fc.FontFace()

	// The full CSS weight ("300", "500", "700"...), not only bold or
	// normal.
	fontWeight := ""
	if fc.ContainsStyle(font.FontStyleBold) {
		if face.CssWeight() >= 700 {
			fontWeight = face.ToCssWeightString()
		} else {
			fontWeight = "700"
		}
	} else if face.CssWeight() != 400 {
		fontWeight = face.ToCssWeightString()
	}

	fontStyle := ""
	if fc.ContainsStyle(font.FontStyleItalic) || face.IsItalic() {
		fontStyle = "italic"
	}

	txt := shape.Text()
	if isJavaBlank(txt) {
		txt = strings.ReplaceAll(txt, " ", "\u00a0")
	}

	if strings.HasPrefix(txt, " ") {
		space := d.stringBounder.CalculateDimension(f, " ").Width
		for strings.HasPrefix(txt, " ") {
			x += space
			txt = txt[1:]
		}
	}
	txt = text.JavaTrim(txt)
	dim := d.stringBounder.CalculateDimension(f, txt)
	width := dim.Width
	height := dim.Height

	var extra extraLines
	var decorations strings.Builder

	if fc.ContainsStyle(font.FontStyleUnderline) {
		if us := fc.UnderlineStroke(); us != nil && us.Thickness() > 0 {
			if fc.ExtendedColor() == nil {
				decorations.WriteString("underline ")
			} else {
				extra.add(fc.ExtendedColor(), f.Size2D()/14.0)
			}
		}
	}

	if fc.ContainsStyle(font.FontStyleStrike) {
		if fc.ExtendedColor() == nil {
			decorations.WriteString("line-through ")
		} else {
			extra.add(fc.ExtendedColor(), -f.Size2D()/4.0)
		}
	}

	if fc.ContainsStyle(font.FontStyleWave) {
		// Some SVG renderers draw a straight line (Chrome).
		decorations.WriteString("wavy underline ")
	}

	textDecoration := text.JavaTrim(decorations.String())

	backColor := ""
	if fc.ContainsStyle(font.FontStyleBackcolor) {
		back := fc.ExtendedColor()
		if gr, ok := back.(*color.HColorGradient); ok {
			id := svg.CreateSvgGradient(gr.Color1().ToRGB(mapper), gr.Color2().ToRGB(mapper), gr.Policy())
			svg.SetFillColor("url(#" + id + ")")
			svg.SetStrokeColor("")
			const deltaPatch = 2
			svg.SvgRectangle(x, y-height+deltaPatch, width, height, 0, 0, 0)
		} else {
			backColor = back.ToRGB(mapper)
		}
	}

	svg.SetFillColor(fc.Color().ToSvg(mapper))
	svg.Text(txt, x, y, f.Size(), width, TextOptions{FontFamily: f.Family(txt, font.ContextSVG),
		FontWeight: fontWeight, FontStyle: fontStyle, TextDecoration: textDecoration, Attributes: fc.Attributes(),
		TextBackColor: backColor, Orientation: shape.Orientation()})

	extra.drawAll(x, y, width, f, mapper, svg)
}

// DriverTextAsPathSvg draws texts as glyph outlines. It needs Java2D's
// TextLayout outlines, which are not ported: drawing panics. Upstream's
// TextBlockExporter never selects it.
type DriverTextAsPathSvg struct{ clipContainer klimt.ClipContainer }

// NewDriverTextAsPathSvg returns the driver.
func NewDriverTextAsPathSvg(clipContainer klimt.ClipContainer) DriverTextAsPathSvg {
	return DriverTextAsPathSvg{clipContainer: clipContainer}
}

// Draw implements klimt.UDriver; it panics unless the text is clipped out.
func (d DriverTextAsPathSvg) Draw(shape *klimt.UText, x, y float64, mapper *color.ColorMapper,
	param klimt.UParam, svg *SvgGraphics) {
	if clip := d.clipContainer.Clip(); clip != nil && !clip.IsInside(x, y) {
		return
	}
	notPorted("java.awt.font.TextLayout", "text drawn as glyph outlines")
}

// DriverCenteredCharacterSvg draws a character centered on a point (the
// letter of a class or spot circle). With SVG_DETERMINISTIC it writes a
// monospace <text>; with SVG it writes the glyph outline as a filled path,
// centered with upstream's UnusedSpace (see centeredchar.go). Characters,
// fonts or sizes outside of the glyph tables panic with a "not ported yet"
// message.
type DriverCenteredCharacterSvg struct{ fileFormat font.FileFormat }

// NewDriverCenteredCharacterSvg returns the driver of a format.
func NewDriverCenteredCharacterSvg(fileFormat font.FileFormat) DriverCenteredCharacterSvg {
	return DriverCenteredCharacterSvg{fileFormat: fileFormat}
}

// Draw implements klimt.UDriver.
func (d DriverCenteredCharacterSvg) Draw(characterCircled *klimt.UCenteredCharacter, x, y float64,
	mapper *color.ColorMapper, param klimt.UParam, svg *SvgGraphics) {
	if d.fileFormat == font.FormatSVGDeterministic {
		DriverCenteredCharacterSvgDeterministic{}.Draw(characterCircled, x, y, mapper, param, svg)
		return
	}
	c := characterCircled.Char()
	f := characterCircled.Font()
	centerX, centerY, ok := unusedSpaceCenter(f, c)
	if !ok {
		notPorted("net.sourceforge.plantuml.klimt.font.UnusedSpace",
			"centered character "+strconv.QuoteRune(c)+" in "+f.UnderlyingFont(string(c)).FontName()+" "+
				strconv.Itoa(f.Size()))
	}
	xpos := x - centerX - 0.5
	ypos := y - centerY - 0.5

	outline, ok := glyphOutline(f, c)
	if !ok {
		notPorted("java.awt.font.TextLayout", "outline of "+strconv.QuoteRune(c))
	}
	svg.SetFillColor(param.Color().ToSvg(mapper))

	svg.DrawPathIterator(xpos, ypos, &outlineIterator{segments: outline})
}

// DriverCenteredCharacterSvgDeterministic draws a centered character as a
// monospace <text> of size 14, offset by (-5, 5).
type DriverCenteredCharacterSvgDeterministic struct{}

// Draw implements klimt.UDriver.
func (DriverCenteredCharacterSvgDeterministic) Draw(characterCircled *klimt.UCenteredCharacter, x, y float64,
	mapper *color.ColorMapper, param klimt.UParam, svg *SvgGraphics) {
	svg.SetFillColor(param.Color().ToSvg(mapper))
	svg.Text(string(characterCircled.Char()), x-5, y+5, 14, 0, TextOptions{FontFamily: "monospace"})
}
