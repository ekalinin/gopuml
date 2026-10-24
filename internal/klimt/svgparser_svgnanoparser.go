// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svg/parser/SvgNanoParser.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// Emojji from https://twemoji.twitter.com/
// Shorcut from https://api.github.com/emojis

// SvgNanoParser is upstream's default SVG sprite parser: it finds the path,
// g, circle, ellipse and text elements with regular expressions and draws
// them; other elements are ignored. It is safe for concurrent use, as the
// standard library keeps its sprites for the process.
type SvgNanoParser struct {
	svg string

	dataOnce sync.Once
	data     []string

	grayMu  sync.Mutex
	minGray int
	maxGray int
}

var (
	_ ISvgSpriteParser = (*SvgNanoParser)(nil)
	_ GrayLevelRange   = (*SvgNanoParser)(nil)
)

// The patterns of upstream; "." is Java's (javaDots).
var (
	svgNanoPTextOrDraw = regexp.MustCompile(javaDots(
		`(\<text .*?\</text\>)|(\<(svg|path|g|circle|ellipse)[^<>]*\>)|(\</[^<>]*\>)`))

	svgNanoPText     = regexp.MustCompile(javaDots(`\<text[^<>]*\>(.*?)\</text\>`))
	svgNanoPFontSize = regexp.MustCompile(`^(\d+)p[tx]$`)

	svgNanoPMatrix = regexp.MustCompile(
		`matrix\(([-.0-9]+)[ ,]+([-.0-9]+)[ ,]+([-.0-9]+)[ ,]+([-.0-9]+)[ ,]+([-.0-9]+)[ ,]+([-.0-9]+)\)`)

	svgNanoPRotate = regexp.MustCompile(`rotate\(([-.0-9]+)[ ,]+([-.0-9]+)[ ,]+([-.0-9]+)\)`)

	svgNanoPTranslate1 = regexp.MustCompile(`translate\(([-.0-9]+)[ ,]+([-.0-9]+)\)`)
	svgNanoPTranslate2 = regexp.MustCompile(`translate\(([-.0-9]+)\)`)

	svgNanoPScale1 = regexp.MustCompile(`scale\(([-.0-9]+)\)`)
	svgNanoPScale2 = regexp.MustCompile(`scale\(([-.0-9]+)[ ,]+([-.0-9]+)\)`)

	svgNanoDataCx          = svgNanoEqualsSomething("cx")
	svgNanoDataCy          = svgNanoEqualsSomething("cy")
	svgNanoDataFill        = svgNanoEqualsSomething("fill")
	svgNanoDataFontFamily  = svgNanoEqualsSomething("font-family")
	svgNanoDataFontSize    = svgNanoEqualsSomething("font-size")
	svgNanoDataR           = svgNanoEqualsSomething("r")
	svgNanoDataRx          = svgNanoEqualsSomething("rx")
	svgNanoDataRy          = svgNanoEqualsSomething("ry")
	svgNanoDataStroke      = svgNanoEqualsSomething("stroke")
	svgNanoDataStrokeWidth = svgNanoEqualsSomething("stroke-width")
	svgNanoDataStyle       = svgNanoEqualsSomething("style")
	svgNanoDataTransform   = svgNanoEqualsSomething("transform")
	svgNanoDataX           = svgNanoEqualsSomething("x")
	svgNanoDataY           = svgNanoEqualsSomething("y")

	svgNanoStyleFill       = svgNanoColonSomething("fill")
	svgNanoStyleFontSize   = svgNanoColonSomething("font-size")
	svgNanoStyleFontFamily = svgNanoColonSomething("font-family")

	svgNanoLowerLetters = regexp.MustCompile(`[a-z]`)
)

func svgNanoEqualsSomething(name string) *regexp.Regexp {
	return regexp.MustCompile(regexp.QuoteMeta(name) + `="([^"]+)"`)
}

func svgNanoColonSomething(name string) *regexp.Regexp {
	return regexp.MustCompile(regexp.QuoteMeta(name) + `:([^;"]+)`)
}

// svgNanoExtract returns the first group of the first match of p in s; ok is
// false without a match (null upstream).
func svgNanoExtract(p *regexp.Regexp, s string) (string, bool) {
	m := p.FindStringSubmatch(s)
	if m == nil {
		return "", false
	}
	return m[1], true
}

// svgNanoParseExtracted is upstream's Double.parseDouble(extract(p, s)): a
// missing attribute panics with a NullPointerException.
func svgNanoParseExtracted(p *regexp.Regexp, s string) float64 {
	v, ok := svgNanoExtract(p, s)
	if !ok {
		throw(nullPointer, "")
	}
	return svgParseDouble(v)
}

// NewSvgNanoParser returns the parser of an SVG text.
func NewSvgNanoParser(svg string) *SvgNanoParser {
	return &SvgNanoParser{svg: svg, minGray: 999, maxGray: -1}
}

// DrawU draws the sprite.
func (p *SvgNanoParser) DrawU(ug UGraphic, scale float64, fontColor, forcedColor color.HColor) {
	colorResolver := NewColorResolver(fontColor, forcedColor, p)
	ugs := NewUGraphicWithScale(ug, colorResolver, scale)

	var stack []*UGraphicWithScale
	var stackG []string // first element is the innermost group
	for _, s := range p.getData() {
		switch {
		case strings.HasPrefix(s, "<path "):
			p.drawPath(ugs, s, stackG)
		case strings.HasPrefix(s, "</g>"):
			if len(stack) == 0 {
				throw(indexOutOfBounds, "Index 0 out of bounds for length 0")
			}
			ugs = stack[0]
			stack = stack[1:]
			stackG = stackG[1:]
		case strings.HasPrefix(s, "<g>"):
			stack = append([]*UGraphicWithScale{ugs}, stack...)
			stackG = append([]string{s}, stackG...)
		case strings.HasPrefix(s, "<g "):
			stack = append([]*UGraphicWithScale{ugs}, stack...)
			stackG = append([]string{s}, stackG...)
			ugs = p.applyFillAndStroke(ugs, s, stackG)
			ugs = p.applyTransform(ugs, s)
		case strings.HasPrefix(s, "<circle "):
			p.drawCircle(ugs, s, stackG)
		case strings.HasPrefix(s, "<ellipse "):
			p.drawEllipse(ugs, s, stackG)
		case strings.HasPrefix(s, "<text "):
			p.drawText(ugs, s, stackG)
		}
	}
}

// getData returns the elements of the SVG text that the parser draws, found
// once.
func (p *SvgNanoParser) getData() []string {
	p.dataOnce.Do(func() {
		for _, m := range svgNanoPTextOrDraw.FindAllString(p.svg, -1) {
			if strings.HasPrefix(m, "<path") || strings.HasPrefix(m, "<g ") || strings.HasPrefix(m, "<g>") ||
				strings.HasPrefix(m, "</g>") || strings.HasPrefix(m, "<circle ") ||
				strings.HasPrefix(m, "<ellipse ") || strings.HasPrefix(m, "<text ") {
				p.data = append(p.data, m)
			}
			// <svg and </svg are ignored, and so is the rest
		}
	})
	return p.data
}

func (p *SvgNanoParser) applyFillAndStroke(ugs *UGraphicWithScale, s string, stackG []string) *UGraphicWithScale {
	fillString, hasFill := p.getFillString(s, stackG)
	strokeString, hasStroke := svgNanoExtract(svgNanoDataStroke, s)

	if strokeWidth, ok := svgNanoExtract(svgNanoDataStrokeWidth, s); ok {
		scale := ugs.InitialScale()
		ugs = ugs.Apply(UStrokeWithThickness(scale * svgParseDouble(strokeWidth)))
	}

	if hasStroke {
		stroke := ugs.TrueColor(strokeString)
		ugs = ugs.Apply(stroke)
		if !hasFill {
			return ugs.Apply(Bg(ugs.DefaultColor()))
		}
	}

	if hasFill && fillString == "none" {
		ugs = ugs.Apply(Bg(color.None()))
	} else {
		var fill color.HColor
		if hasFill {
			fill = ugs.TrueColor(fillString)
		} else {
			fill = ugs.DefaultColor()
		}
		if !hasStroke {
			ugs = ugs.Apply(fill)
		}
		ugs = ugs.Apply(Bg(fill))
	}

	return ugs
}

func (p *SvgNanoParser) drawCircle(ugs *UGraphicWithScale, s string, stackG []string) {
	ugs = p.applyFillAndStroke(ugs, s, stackG)
	ugs = p.applyTransform(ugs, s)

	scalex := ugs.AffineTransform().ScaleX()
	scaley := ugs.AffineTransform().ScaleY()

	deltax := ugs.AffineTransform().TranslateX()
	deltay := ugs.AffineTransform().TranslateY()

	cx := float64(svgNanoParseExtracted(svgNanoDataCx, s) * scalex)
	cy := float64(svgNanoParseExtracted(svgNanoDataCy, s) * scaley)
	rx := float64(svgNanoParseExtracted(svgNanoDataR, s) * scalex)
	ry := float64(svgNanoParseExtracted(svgNanoDataR, s) * scaley)

	translate := NewUTranslate(deltax+cx-rx, deltay+cy-ry)
	ugs.Apply(translate).Draw(UEllipseBuild(rx*2, ry*2))
}

func (p *SvgNanoParser) drawEllipse(ugs *UGraphicWithScale, s string, stackG []string) {
	ugs = p.applyFillAndStroke(ugs, s, stackG)
	ugs = p.applyTransform(ugs, s)

	cx := svgNanoParseExtracted(svgNanoDataCx, s)
	cy := svgNanoParseExtracted(svgNanoDataCy, s)
	rx := svgNanoParseExtracted(svgNanoDataRx, s)
	ry := svgNanoParseExtracted(svgNanoDataRy, s)

	path := UPathNone()
	path.MoveTo(0, ry)
	path.ArcTo(rx, ry, 0, 0, 1, rx, 0)
	path.ArcTo(rx, ry, 0, 0, 1, 2*rx, ry)
	path.ArcTo(rx, ry, 0, 0, 1, rx, 2*ry)
	path.ArcTo(rx, ry, 0, 0, 1, 0, ry)
	path.ClosePath()

	path = path.Translate(cx-rx, cy-ry)
	path = path.Affine(ugs.AffineTransform(), ugs.Angle(), ugs.InitialScale())

	ugs.Draw(path)
}

func (p *SvgNanoParser) drawText(ugs *UGraphicWithScale, s string, stackG []string) {
	ugs = p.applyFillAndStroke(ugs, s, stackG)
	ugs = p.applyTransform(ugs, s)

	scalex := ugs.AffineTransform().ScaleX()
	scaley := ugs.AffineTransform().ScaleY()

	deltax := ugs.AffineTransform().TranslateX()
	deltay := ugs.AffineTransform().TranslateY()

	x := float64(svgNanoParseExtracted(svgNanoDataX, s)*scalex) + deltax
	y := float64(svgNanoParseExtracted(svgNanoDataY, s)*scaley) + deltay
	fontColor, hasFontColor := p.getFillString(s, stackG)
	fontSize := javaD2I(float64(p.getTextFontSize(s)) * ugs.InitialScale())

	if m := svgNanoPText.FindStringSubmatch(s); m != nil {
		text := m[1]
		if !hasFontColor {
			// upstream's HColorSet.getColorOrWhite(null)
			throw(nullPointer, "")
		}
		c := color.Instance().GetColorOrWhite(fontColor)
		fontFamily, ok := p.getTextFontFamily(s, stackG)
		if !ok {
			fontFamily = "SansSerif"
		}
		f := font.Build(fontFamily, font.NormalFace(), fontSize)
		fc := NewFontConfiguration(f, c, c, nil)
		utext := UTextBuild(text, fc)
		ug := ugs.Ug()
		ug = ug.Apply(NewUTranslate(x, y))
		ug.Draw(utext)
	}
}

func (p *SvgNanoParser) getTextFontFamily(s string, stackG []string) (string, bool) {
	family, ok := svgNanoExtract(svgNanoDataFontFamily, s)
	if !ok {
		if style, hasStyle := svgNanoExtract(svgNanoDataStyle, s); hasStyle {
			family, ok = svgNanoExtract(svgNanoStyleFontFamily, style)
		}
	}
	if !ok {
		for _, g := range stackG {
			if family, ok = p.getTextFontFamily(g, nil); ok {
				return family, true
			}
		}
	}
	return family, ok
}

func (p *SvgNanoParser) getFillString(s string, stackG []string) (string, bool) {
	c, ok := svgNanoExtract(svgNanoDataFill, s)
	if !ok {
		if style, hasStyle := svgNanoExtract(svgNanoDataStyle, s); hasStyle {
			c, ok = svgNanoExtract(svgNanoStyleFill, style)
		}
	}

	if !ok {
		for _, g := range stackG {
			if c, ok = p.getFillString(g, nil); ok {
				return c, true
			}
		}
	}

	return c, ok
}

func (p *SvgNanoParser) getTextFontSize(s string) int {
	fontSize, ok := svgNanoExtract(svgNanoDataFontSize, s)
	if !ok {
		if style, hasStyle := svgNanoExtract(svgNanoDataStyle, s); hasStyle {
			fontSize, ok = svgNanoExtract(svgNanoStyleFontSize, style)
		}
	}
	if !ok {
		// Not perfect, by let's take a default value
		return 14
	}

	if svgNanoPFontSize.MatchString(fontSize) {
		return javaParseInt(svgNanoLowerLetters.ReplaceAllString(fontSize, ""))
	}

	return javaParseInt(fontSize)
}

func (p *SvgNanoParser) drawPath(ugs *UGraphicWithScale, s string, stackG []string) {
	s = strings.ReplaceAll(s, "id=\"", "ID=\"")
	ugs = p.applyFillAndStroke(ugs, s, stackG)
	ugs = p.applyTransform(ugs, s)

	// s starts with "<path ", so byte and UTF-16 indexes agree up to x1+3
	x1 := strings.Index(s, "d=\"")
	x2 := -1
	if i := strings.IndexByte(s[x1+3:], '"'); i >= 0 {
		x2 = x1 + 3 + i
	}
	if x2 < 0 {
		throw("java.lang.StringIndexOutOfBoundsException", "begin "+itoa(x1+3)+", end -1, length "+
			itoa(javaLen(s)))
	}
	tmp := s[x1+3 : x2]

	svgPath := NewSvgPath(tmp, UTranslateNone())
	svgPath.DrawMe(ugs.Ug(), ugs.AffineTransform())
}

func (p *SvgNanoParser) applyTransform(ugs *UGraphicWithScale, s string) *UGraphicWithScale {
	transform, ok := svgNanoExtract(svgNanoDataTransform, s)
	if !ok {
		return ugs
	}

	if strings.Contains(transform, "rotate(") {
		return p.applyRotate(ugs, transform)
	}

	if strings.Contains(transform, "matrix(") {
		return p.applyMatrix(ugs, transform)
	}

	scale := p.getScale(transform)
	translate := p.getTranslate(transform)
	ugs = ugs.ApplyTranslate(translate.Dx(), translate.Dy())

	return ugs.ApplyScale(scale[0], scale[1])
}

func (p *SvgNanoParser) applyMatrix(ugs *UGraphicWithScale, transform string) *UGraphicWithScale {
	if m3 := svgNanoPMatrix.FindStringSubmatch(transform); m3 != nil {
		v1 := svgParseDouble(m3[1])
		v2 := svgParseDouble(m3[2])
		v3 := svgParseDouble(m3[3])
		v4 := svgParseDouble(m3[4])
		v5 := svgParseDouble(m3[5])
		v6 := svgParseDouble(m3[6])
		ugs = ugs.ApplyMatrix(v1, v2, v3, v4, v5, v6)
	} else {
		fmt.Fprintln(os.Stderr, "WARNING: "+transform)
	}
	return ugs
}

func (p *SvgNanoParser) applyRotate(ugs *UGraphicWithScale, transform string) *UGraphicWithScale {
	if m3 := svgNanoPRotate.FindStringSubmatch(transform); m3 != nil {
		angle := svgParseDouble(m3[1])
		x := svgParseDouble(m3[2])
		y := svgParseDouble(m3[3])
		ugs = ugs.ApplyRotate(angle, x, y)
	} else {
		fmt.Fprintln(os.Stderr, "WARNING: "+transform)
	}
	return ugs
}

func (p *SvgNanoParser) getTranslate(transform string) UTranslate {
	x, y := 0.0, 0.0

	if m3 := svgNanoPTranslate1.FindStringSubmatch(transform); m3 != nil {
		x = svgParseDouble(m3[1])
		y = svgParseDouble(m3[2])
	} else if m4 := svgNanoPTranslate2.FindStringSubmatch(transform); m4 != nil {
		x = svgParseDouble(m4[1])
		y = svgParseDouble(m4[1])
	}
	return NewUTranslate(x, y)
}

func (p *SvgNanoParser) getScale(transform string) [2]float64 {
	scale := [2]float64{1, 1}
	if m1 := svgNanoPScale1.FindStringSubmatch(transform); m1 != nil {
		scale[0] = svgParseDouble(m1[1])
		scale[1] = scale[0]
	} else if m2 := svgNanoPScale2.FindStringSubmatch(transform); m2 != nil {
		scale[0] = svgParseDouble(m2[1])
		scale[1] = svgParseDouble(m2[2])
	}
	return scale
}

// AsTextBlock returns the sprite as a block of the size of the SVG (its
// viewBox, else its width and height) times scale, over backColor when it is
// not nil. A size that cannot be found panics, as upstream throws
// IllegalStateException.
func (p *SvgNanoParser) AsTextBlock(fontColor, forcedColor color.HColor, scale float64,
	backColor color.HColor) TextBlock {
	return newSvgSpriteTextBlock(p.svg, p, fontColor, forcedColor, scale, backColor)
}

// svgSpriteTextBlock is the text block of the SVG sprite parsers.
type svgSpriteTextBlock struct {
	TextBlockDefaults
	parser                 ISvgSpriteParser
	fontColor, forcedColor color.HColor
	backColor              color.HColor
	scale                  float64
	width, height          float64
}

func newSvgSpriteTextBlock(svg string, parser ISvgSpriteParser, fontColor, forcedColor color.HColor,
	scale float64, backColor color.HColor) *svgSpriteTextBlock {
	data := NewUImageSvg(svg, scale)
	width := data.Width()
	height := data.Height()
	return &svgSpriteTextBlock{parser: parser, fontColor: fontColor, forcedColor: forcedColor,
		backColor: backColor, scale: scale, width: width, height: height}
}

// DrawU draws the background, then the sprite.
func (t *svgSpriteTextBlock) DrawU(ug UGraphic) {
	if t.backColor != nil {
		ug.Apply(Bg(t.backColor)).Apply(t.backColor).Draw(URectangleBuildDim(
			t.CalculateDimension(ug.StringBounder())))
	}
	t.parser.DrawU(ug, t.scale, t.fontColor, t.forcedColor)
}

// CalculateDimension returns the size of the SVG times the scale.
func (t *svgSpriteTextBlock) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(t.width, t.height)
}

func (p *SvgNanoParser) computeMinMaxGray() {
	for _, s := range p.getData() {
		if strings.Contains(s, "<path ") || strings.Contains(s, "<g ") || strings.Contains(s, "<circle ") ||
			strings.Contains(s, "<ellipse ") {
			fillString, hasFill := p.getFillString(s, nil)
			strokeString, hasStroke := svgNanoExtract(svgNanoDataStroke, s)

			p.updateMinMax(strokeString, hasStroke)
			p.updateMinMax(fillString, hasFill)
		}
	}
}

func (p *SvgNanoParser) updateMinMax(colorString string, ok bool) {
	if ok {
		c := color.Instance().GetColorOrWhite(colorString)
		gray := color.GrayScaleColor(c.ToColor(color.MonochromeMapper)).Green()
		p.minGray = min(p.minGray, gray)
		p.maxGray = max(p.maxGray, gray)
	}
}

// MinGrayLevel returns the lowest gray level of the fill and stroke colors,
// 999 when there is none.
func (p *SvgNanoParser) MinGrayLevel() int {
	p.grayMu.Lock()
	defer p.grayMu.Unlock()
	if p.maxGray == -1 {
		p.computeMinMaxGray()
	}
	return p.minGray
}

// MaxGrayLevel returns the highest gray level of the fill and stroke colors,
// -1 when there is none.
func (p *SvgNanoParser) MaxGrayLevel() int {
	p.grayMu.Lock()
	defer p.grayMu.Unlock()
	if p.maxGray == -1 {
		p.computeMinMaxGray()
	}
	return p.maxGray
}
