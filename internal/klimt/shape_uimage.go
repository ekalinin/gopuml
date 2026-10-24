// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/shape/UImage.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/shape/UImageSvg.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/shape/UImageTikz.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"
	"regexp"
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// UImage is a raster image. It is a TextBlock that draws itself as a shape,
// so the drivers receive it.
type UImage struct {
	TextBlockDefaults
	image       MutableImage
	formula     string
	rawFileName string
}

var _ TextBlock = (*UImage)(nil)

// NewUImage returns the image.
func NewUImage(image MutableImage) *UImage { return &UImage{image: image} }

// WithRawFileName returns the image with the name of its file.
func (u *UImage) WithRawFileName(rawFileName string) *UImage {
	return &UImage{image: u.image, rawFileName: rawFileName, formula: u.formula}
}

// WithFormula returns the image with the formula it was rendered from.
func (u *UImage) WithFormula(formula string) *UImage {
	return &UImage{image: u.image, rawFileName: u.rawFileName, formula: formula}
}

// RawFileName returns the name of the file, "" for none.
func (u *UImage) RawFileName() string { return u.rawFileName }

// Formula returns the formula, "" for none.
func (u *UImage) Formula() string { return u.formula }

// Scale returns the image scaled.
func (u *UImage) Scale(scale float64) *UImage {
	return &UImage{image: u.image.WithScale(scale), rawFileName: u.rawFileName, formula: u.formula}
}

// Image returns the pixels at a scale.
func (u *UImage) Image(withScale float64) PortableImage { return u.image.WithScale(withScale).Image() }

// Width returns the width minus one, as upstream.
func (u *UImage) Width() int { return u.image.Image().Width() - 1 }

// Height returns the height minus one, as upstream.
func (u *UImage) Height() int { return u.image.Image().Height() - 1 }

// MuteColor returns the image recolored.
func (u *UImage) MuteColor(newColor color.XColor) *UImage {
	return &UImage{image: u.image.MuteColor(newColor), rawFileName: u.rawFileName, formula: u.formula}
}

// MuteTransparentColor returns the image with its transparent color
// replaced.
func (u *UImage) MuteTransparentColor(newColor color.XColor) *UImage {
	return &UImage{image: u.image.MuteTransparentColor(newColor), rawFileName: u.rawFileName, formula: u.formula}
}

// Monochrome returns the image in gray levels.
func (u *UImage) Monochrome() *UImage {
	return &UImage{image: u.image.Monochrome(), rawFileName: u.rawFileName, formula: u.formula}
}

// ImageScale returns the scale of the image.
func (u *UImage) ImageScale() float64 { return u.image.Scale() }

// DrawU draws the image as a shape.
func (u *UImage) DrawU(ug UGraphic) { ug.Draw(u) }

// CalculateDimension returns the width and height (each minus one).
func (u *UImage) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(float64(u.Width()), float64(u.Height()))
}

// MinMax returns the bounds from (0, 0) to the size.
func (u *UImage) MinMax(font.StringBounder) geom.MinMax {
	return geom.MinMaxFromMax(float64(u.Width()), float64(u.Height()))
}

// UImageSvg is an embedded SVG image.
type UImageSvg struct {
	svg          string
	scale        float64
	cachedWidth  int
	cachedHeight int
}

// NewUImageSvg returns the SVG image at a scale.
func NewUImageSvg(svg string, scale float64) *UImageSvg {
	return &UImageSvg{svg: svg, scale: scale, cachedWidth: -1, cachedHeight: -1}
}

// MD5Hex returns the MD5 of the SVG text.
func (u *UImageSvg) MD5Hex() string { return md5Hex(u.svg) }

// ContainsXlink reports whether the SVG declares the xlink namespace.
func (u *UImageSvg) ContainsXlink() bool {
	return strings.Contains(u.svg, "xmlns:xlink=\"http://www.w3.org/1999/xlink\"")
}

var (
	svgBackgroundPattern = regexp.MustCompile(`background:([^;]+)`)
	svgStylePattern      = regexp.MustCompile(`(?i)<svg[^>]+style="([^">]+)"`)
	svgViewBoxPattern    = regexp.MustCompile(
		`viewBox[= "']+([0-9.]+)[\t\n\v\f\r ,]+([0-9.]+)[\t\n\v\f\r ,]+([0-9.]+)[\t\n\v\f\r ,]+([0-9.]+)`)
	svgWidthPattern  = regexp.MustCompile(`(?i)<svg[^>]+width\W+(\d+)`)
	svgHeightPattern = regexp.MustCompile(`(?i)<svg[^>]+height\W+(\d+)`)
)

// Svg returns the SVG text: raw, or reduced to a <svg> element without
// attributes, the background of its style turned into a rectangle.
func (u *UImageSvg) Svg(raw bool) string {
	result := u.svg
	if raw {
		return result
	}
	if strings.HasPrefix(result, "<?xml") {
		idx := strings.Index(result, "<svg")
		if idx < 0 {
			throw("java.lang.StringIndexOutOfBoundsException", "begin -1, end "+itoa(javaLen(result))+
				", length "+itoa(javaLen(result)))
		}
		result = result[idx:]
	}
	if strings.HasPrefix(result, "<svg") {
		idx := strings.Index(result, ">")
		result = "<svg>" + result[idx+1:]
	}
	if style, ok := u.extractSvgStyle(); ok {
		if background, ok := extractBackground(style); ok {
			width := u.Data("width")
			height := u.Data("height")
			result = javaReplaceFirstLiteral(result, "<g>", "<g><rect fill=\""+background+"\" style=\""+style+
				"\" width=\""+itoa(width)+"\" height=\""+itoa(height)+"\"/> ")
		}
	}
	if !strings.HasPrefix(result, "<svg>") {
		throw(illegalArgument, "")
	}
	return result
}

func extractBackground(style string) (string, bool) {
	m := svgBackgroundPattern.FindStringSubmatch(style)
	if m == nil {
		return "", false
	}
	return m[1], true
}

func (u *UImageSvg) extractSvgStyle() (string, bool) {
	m := svgStylePattern.FindStringSubmatch(u.svg)
	if m == nil {
		return "", false
	}
	return m[1], true
}

// Data returns the "width" or "height" of the SVG: from its viewBox,
// rounded up, or else from its attribute. It panics when there is none, as
// upstream throws IllegalStateException.
func (u *UImageSvg) Data(name string) int {
	if m2 := svgViewBoxPattern.FindStringSubmatch(u.svg); m2 != nil {
		switch name {
		case "width":
			return javaD2I(math.Ceil(javaParseDouble(m2[3])))
		case "height":
			return javaD2I(math.Ceil(javaParseDouble(m2[4])))
		}
	}
	var p *regexp.Regexp
	switch name {
	case "width":
		p = svgWidthPattern
	case "height":
		p = svgHeightPattern
	default:
		p = regexp.MustCompile(`(?i)<svg[^>]+` + regexp.QuoteMeta(name) + `\W+(\d+)`)
	}
	if m := p.FindStringSubmatch(u.svg); m != nil {
		return javaParseInt(m[1])
	}
	throw(illegalState, "Cannot find "+name)
	return 0
}

// Height returns the height times the scale.
func (u *UImageSvg) Height() float64 {
	if u.cachedHeight == -1 {
		u.cachedHeight = u.Data("height")
	}
	return float64(u.cachedHeight) * u.scale
}

// Width returns the width times the scale.
func (u *UImageSvg) Width() float64 {
	if u.cachedWidth == -1 {
		u.cachedWidth = u.Data("width")
	}
	return float64(u.cachedWidth) * u.scale
}

// Scale returns the scale.
func (u *UImageSvg) Scale() float64 { return u.scale }

// UImageTikz is an embedded TikZ picture.
type UImageTikz struct {
	tikzCode    string
	width       float64
	height      float64
	hasURL      bool
	scaleFactor float64
}

// NewUImageTikz returns the picture of a TikZ code of a size.
func NewUImageTikz(tikzCode string, width, height float64) *UImageTikz {
	u := &UImageTikz{tikzCode: tikzCode, width: width, height: height, scaleFactor: 1}
	u.scan(tikzCode)
	return u
}

func (u *UImageTikz) scan(tikzCode string) {
	beforeTikzPicture := true
	for _, line := range strings.Split(tikzCode, "\n") {
		trimmed := strings.TrimFunc(line, func(r rune) bool { return r <= ' ' })
		if beforeTikzPicture && strings.HasPrefix(trimmed, "\\scalebox{") {
			u.scaleFactor = parseScaleFactor(trimmed)
		}
		if strings.HasPrefix(trimmed, "\\begin{tikzpicture}") {
			beforeTikzPicture = false
		}
		if containsTikzURL(trimmed) {
			u.hasURL = true
		}
	}
}

// TikzCode returns the code.
func (u *UImageTikz) TikzCode() string { return u.tikzCode }

// HasURL reports whether the code contains a link.
func (u *UImageTikz) HasURL() bool { return u.hasURL }

// Width returns the width times the \scalebox factor.
func (u *UImageTikz) Width() float64 { return u.width * u.scaleFactor }

// Height returns the height times the \scalebox factor.
func (u *UImageTikz) Height() float64 { return u.height * u.scaleFactor }

func parseScaleFactor(texLine string) float64 {
	start := strings.IndexByte(texLine, '{') + 1
	end := -1
	if start > 0 {
		if i := strings.IndexByte(texLine[start:], '}'); i >= 0 {
			end = start + i
		}
	}
	if start > 0 && end > start {
		if v, ok := javaTryParseDouble(texLine[start:end]); ok {
			return v
		}
	}
	return 1
}

func containsTikzURL(texLine string) bool {
	return strings.Contains(texLine, "\\href{") || strings.Contains(texLine, "\\hyperref[") ||
		strings.Contains(texLine, "href node") || strings.Contains(texLine, "hyperref node")
}

// javaReplaceFirstLiteral is Java's s.replaceFirst(regex, replacement) for
// a regex without metacharacters: the replacement keeps Java's escapes, a
// backslash quoting the next character and "$0" standing for the match;
// other group references panic as Java throws.
func javaReplaceFirstLiteral(s, literal, replacement string) string {
	idx := strings.Index(s, literal)
	if idx < 0 {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(replacement); i++ {
		c := replacement[i]
		switch {
		case c == '\\':
			i++
			if i >= len(replacement) {
				throw(illegalArgument, "character to be escaped is missing")
			}
			b.WriteByte(replacement[i])
		case c == '$':
			i++
			if i >= len(replacement) {
				throw(illegalArgument, "Illegal group reference: group index is missing")
			}
			if replacement[i] != '0' {
				if replacement[i] >= '1' && replacement[i] <= '9' {
					throw(indexOutOfBounds, "No group "+string(replacement[i]))
				}
				throw(illegalArgument, "Illegal group reference")
			}
			b.WriteString(literal)
		default:
			b.WriteByte(c)
		}
	}
	return s[:idx] + b.String() + s[idx+len(literal):]
}
