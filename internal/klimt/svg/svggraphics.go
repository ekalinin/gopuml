// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/drawing/svg/SvgGraphics.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svg

import (
	"embed"
	"encoding/base64"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/ekalinin/gopuml/internal/code"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/security"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/version"
)

const (
	defaultFontFamily   = "sans-serif"
	defaultLengthAdjust = "spacing"

	xlinkTitle1 = "title"
	xlinkTitle2 = "xlink:title"
	xlinkHref1  = "href"
	xlinkHref2  = "xlink:href"
)

// MetaHeader is upstream's META_HEADER, the start of the legacy metadata
// comment.
const MetaHeader = "<!--SRC=["

// assets holds upstream's /svg resources of the interactive mode.
//
//go:embed assets/*.min.css assets/*.min.js
var assets embed.FS

// SvgGraphics builds an SVG document: the drivers set the fill, stroke and
// stroke width, then add shapes; CreateXML writes the document once the
// canvas size is known.
type SvgGraphics struct {
	document *XMLDocument
	root     *XMLNode
	defs     *XMLNode
	gRoot    *XMLNode

	fill            string
	stroke          string
	strokeWidth     string
	strokeDasharray string // "" for none

	backcolorString string
	hasBackcolor    bool

	maxX, maxY int

	filterUID  string
	shadowID   string
	gradientID string

	option *SvgOption

	pendingBackground *XMLNode
	robotoAdded       bool

	gradients       map[string]string
	gradientCount   int
	filterBackColor map[string]string
	images          []inlinedImage

	withShadow bool
	hidden     bool

	currentPath *strings.Builder

	pendingElements []*XMLNode
	activeLinks     []linkData
}

// inlinedImage is an SVG image spliced verbatim in place of the empty
// element named key.
type inlinedImage struct{ key, svg string }

func (s *SvgGraphics) ensureVisible(x, y float64) {
	if x > float64(s.maxX) {
		s.maxX = javaD2I(x + 1)
	}
	if y > float64(s.maxY) {
		s.maxY = javaD2I(y + 1)
	}
}

// NewSvgGraphics returns the document of an option. seed makes the ids of
// filters and gradients.
func NewSvgGraphics(seed int64, option *SvgOption) *SvgGraphics {
	s := &SvgGraphics{document: &XMLDocument{}, fill: "black", stroke: "black", maxX: 10, maxY: 10,
		option: option, gradients: map[string]string{}, filterBackColor: map[string]string{}}
	minDim := option.MinDim()
	s.ensureVisible(minDim.Width(), minDim.Height())

	s.root = s.rootNode()
	for _, a := range option.rootAttributes {
		s.root.SetAttribute(a.name, a.value)
	}

	s.defs = s.simpleElement("defs")
	s.gRoot = s.simpleElement("g")
	s.gRoot.SetAttribute("font-family", defaultFontFamily)
	switch option.LengthAdjust() {
	case klimt.LengthAdjustSpacing:
		s.gRoot.SetAttribute("lengthAdjust", defaultLengthAdjust)
	case klimt.LengthAdjustSpacingAndGlyphs:
		s.gRoot.SetAttribute("lengthAdjust", "spacingAndGlyphs")
	}
	s.strokeWidth = s.Format(1)
	s.filterUID = "b" + seedString(seed)
	s.shadowID = "f" + seedString(seed)
	s.gradientID = "g" + seedString(seed)
	if option.Hover() != "" {
		s.defs.AppendChild(s.pathHover(option.Hover()))
	}

	if option.IsInteractive() {
		if styles := s.stylesForInteractiveMode(); styles != nil {
			s.defs.AppendChild(styles)
		}
		if script := s.scriptForInteractiveMode(); script != nil {
			s.defs.AppendChild(script)
		}
	}

	switch backcolor := option.Backcolor().(type) {
	case *color.HColorGradient:
		id := s.CreateSvgGradient(backcolor.Color1().ToRGB(option.ColorMapper()),
			backcolor.Color2().ToRGB(option.ColorMapper()), backcolor.Policy())
		s.paintBackcolor("url(#" + id + ")")
	case nil:
	default:
		s.backcolorString = backcolor.ToSvg(option.ColorMapper())
		s.hasBackcolor = true
		c := backcolor.ToSvg(option.ColorMapper())
		if c != "#00000000" && c != "#000000" && c != "#FFFFFF" {
			s.paintBackcolor(c)
		}
	}
	return s
}

func (s *SvgGraphics) addRoboto() {
	if s.robotoAdded {
		return
	}
	style := s.document.CreateElement("style")
	style.SetAttribute("type", "text/css")
	style.SetTextContent("@import url('https://fonts.googleapis.com/css?family=Roboto:400,100,100italic,300," +
		"300italic,400italic,500,500italic,700,700italic,900,900italic');")
	s.defs.AppendChild(style)
	s.robotoAdded = true
}

func (s *SvgGraphics) paintBackcolor(back string) {
	s.SetFillColor(back)
	s.SetStrokeColor("")
	s.pendingBackground = s.createRectangleInternal(0, 0, 0, 0)
	s.G().AppendChild(s.pendingBackground)
}

func (s *SvgGraphics) stylesForInteractiveMode() *XMLNode {
	data, ok := assetText(s.option.InteractiveBaseFilename() + ".min.css")
	if !ok {
		return nil
	}
	style := s.document.CreateElement("style")
	style.SetAttribute("type", "text/css")
	style.AppendCData(data)
	return style
}

func (s *SvgGraphics) scriptForInteractiveMode() *XMLNode {
	data, ok := assetText(s.option.InteractiveBaseFilename() + ".min.js")
	if !ok {
		return nil
	}
	script := s.document.CreateElement("script")
	script.SetTextContent(data)
	return script
}

// assetText reads an interactive asset as upstream's FileUtils.readText:
// each line followed by "\n". A missing asset gives ok false (upstream logs
// an error and continues).
func assetText(name string) (string, bool) {
	data, err := assets.ReadFile("assets/" + name)
	if err != nil {
		return "", false
	}
	s := text.DecodeUTF8(data)
	var b strings.Builder
	for len(s) > 0 {
		line := s
		i := strings.IndexAny(s, "\r\n")
		if i >= 0 {
			line = s[:i]
			if s[i] == '\r' && i+1 < len(s) && s[i+1] == '\n' {
				s = s[i+2:]
			} else {
				s = s[i+1:]
			}
		} else {
			s = ""
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String(), true
}

func (s *SvgGraphics) pathHover(hover string) *XMLNode {
	style := s.document.CreateElement("style")
	style.SetAttribute("type", "text/css")
	style.AppendCData("path:hover { stroke: " + hover + " !important;}")
	return style
}

// seedString is upstream's Long.toString(Math.abs(seed), 36);
// Math.abs(Long.MIN_VALUE) stays negative, as Go's negation.
func seedString(seed int64) string {
	if seed < 0 {
		seed = -seed
	}
	return strconv.FormatInt(seed, 36)
}

func (s *SvgGraphics) simpleElement(typ string) *XMLNode {
	elt := s.document.CreateElement(typ)
	s.root.AppendChild(elt)
	return elt
}

func (s *SvgGraphics) rootNode() *XMLNode {
	svg := s.document.CreateElement("svg")
	s.document.SetRoot(svg)

	// The version as first child of <svg>, for Confluence.
	svg.AppendProcessingInstruction("plantuml", version.Upstream)

	svg.SetAttribute("xmlns", "http://www.w3.org/2000/svg")
	svg.SetAttribute("xmlns:xlink", "http://www.w3.org/1999/xlink")
	svg.SetAttribute("version", "1.1")

	if title, ok := s.option.Title(); ok {
		elt := s.document.CreateElement("title")
		elt.SetTextContent(title)
		svg.AppendChild(elt)
	}
	if desc, ok := s.option.Desc(); ok {
		elt := s.document.CreateElement("desc")
		elt.SetTextContent(desc)
		svg.AppendChild(elt)
	}
	return svg
}

// Document returns the document being built.
func (s *SvgGraphics) Document() *XMLDocument { return s.document }

// SvgEllipse adds an ellipse centered on (x, y).
func (s *SvgGraphics) SvgEllipse(x, y, xRadius, yRadius, deltaShadow float64) {
	s.manageShadow(deltaShadow)
	if !s.hidden {
		elt := s.document.CreateElement("ellipse")
		elt.SetAttribute("cx", s.Format(x))
		elt.SetAttribute("cy", s.Format(y))
		elt.SetAttribute("rx", s.Format(xRadius))
		elt.SetAttribute("ry", s.Format(yRadius))
		s.fillMe(elt)
		s.styleMe(elt, "")
		s.addFilterShadowID(elt, deltaShadow)
		s.G().AppendChild(elt)
	}
	s.ensureVisible(x+xRadius+float64(deltaShadow*2), y+yRadius+float64(deltaShadow*2))
}

// SvgArcEllipse adds an elliptical arc from (x1, y1) to (x2, y2).
func (s *SvgGraphics) SvgArcEllipse(rx, ry, x1, y1, x2, y2 float64) {
	if !s.hidden {
		path := "M" + s.Format(x1) + "," + s.Format(y1) + " A" + s.Format(rx) + "," + s.Format(ry) + " 0 0 0 " +
			s.Format(x2) + " " + s.Format(y2)
		elt := s.document.CreateElement("path")
		elt.SetAttribute("d", path)
		s.fillMe(elt)
		s.styleMe(elt, "")
		s.G().AppendChild(elt)
	}
	s.ensureVisible(x1, y1)
	s.ensureVisible(x2, y2)
}

// CreateSvgGradient returns the id of the two-color gradient of a policy
// ('|', '\\', '-' or another for the diagonal), adding it to the
// definitions on first use.
func (s *SvgGraphics) CreateSvgGradient(color1, color2 string, policy rune) string {
	key := "2\x00" + color1 + "\x00" + color2 + "\x00" + string(policy)
	if id, ok := s.gradients[key]; ok {
		return id
	}
	elt := s.document.CreateElement("linearGradient")
	switch policy {
	case '|':
		elt.SetAttribute("x1", "0%")
		elt.SetAttribute("y1", "50%")
		elt.SetAttribute("x2", "100%")
		elt.SetAttribute("y2", "50%")
	case '\\':
		elt.SetAttribute("x1", "0%")
		elt.SetAttribute("y1", "100%")
		elt.SetAttribute("x2", "100%")
		elt.SetAttribute("y2", "0%")
	case '-':
		elt.SetAttribute("x1", "50%")
		elt.SetAttribute("y1", "0%")
		elt.SetAttribute("x2", "50%")
		elt.SetAttribute("y2", "100%")
	default:
		elt.SetAttribute("x1", "0%")
		elt.SetAttribute("y1", "0%")
		elt.SetAttribute("x2", "100%")
		elt.SetAttribute("y2", "100%")
	}
	id := s.newGradientID(key)
	elt.SetAttribute("id", id)

	stop1 := s.document.CreateElement("stop")
	stop1.SetAttribute("stop-color", shortenColor(color1))
	stop1.SetAttribute("offset", "0%")
	stop2 := s.document.CreateElement("stop")
	stop2.SetAttribute("stop-color", shortenColor(color2))
	stop2.SetAttribute("offset", "100%")

	elt.AppendChild(stop1)
	elt.AppendChild(stop2)
	s.defs.AppendChild(elt)
	return id
}

// newGradientID records a gradient key: ids number the gradients of both
// kinds in creation order, as upstream's map size.
func (s *SvgGraphics) newGradientID(key string) string {
	id := s.gradientID + strconv.Itoa(s.gradientCount)
	s.gradientCount++
	s.gradients[key] = id
	return id
}

// CreateSvgLinearGradient returns the id of a linear gradient with stops,
// adding it to the definitions on first use.
func (s *SvgGraphics) CreateSvgLinearGradient(gr *color.HColorLinearGradient, mapper *color.ColorMapper) string {
	key := linearGradientKey(gr, mapper)
	if id, ok := s.gradients[key]; ok {
		return id
	}
	elt := s.document.CreateElement("linearGradient")
	if gr.IsUserSpaceOnUse() {
		elt.SetAttribute("gradientUnits", "userSpaceOnUse")
		elt.SetAttribute("x1", s.Format(gr.X1()))
		elt.SetAttribute("y1", s.Format(gr.Y1()))
		elt.SetAttribute("x2", s.Format(gr.X2()))
		elt.SetAttribute("y2", s.Format(gr.Y2()))
	} else {
		elt.SetAttribute("x1", s.FormatPercent(gr.X1()))
		elt.SetAttribute("y1", s.FormatPercent(gr.Y1()))
		elt.SetAttribute("x2", s.FormatPercent(gr.X2()))
		elt.SetAttribute("y2", s.FormatPercent(gr.Y2()))
	}
	if gr.SpreadMethod() != color.SpreadPad {
		elt.SetAttribute("spreadMethod", strings.ToLower(gr.SpreadMethod().String()))
	}
	id := s.newGradientID(key)
	elt.SetAttribute("id", id)

	for _, stop := range gr.Stops() {
		stopElt := s.document.CreateElement("stop")
		stopElt.SetAttribute("offset", s.FormatPercent(stop.Offset()))
		c := stop.Color().ToColor(mapper)
		stopElt.SetAttribute("stop-color", shortenColor(color.ToHexRGBColor(c.RGB())))
		opacity := float64(c.Alpha()) / 255.0 * stop.Opacity()
		if opacity < 0.9999 {
			stopElt.SetAttribute("stop-opacity", s.FormatOpacity(opacity))
		}
		elt.AppendChild(stopElt)
	}
	s.defs.AppendChild(elt)
	return id
}

// linearGradientKey is upstream's buildLinearGradientKey. Doubles compare
// as Double.equals does, by their bits.
func linearGradientKey(gr *color.HColorLinearGradient, mapper *color.ColorMapper) string {
	var b strings.Builder
	bits := func(v float64) {
		b.WriteString(strconv.FormatUint(math.Float64bits(v), 16))
		b.WriteByte(0)
	}
	b.WriteString("linear\x00")
	b.WriteString(strconv.FormatBool(gr.IsUserSpaceOnUse()))
	b.WriteByte(0)
	b.WriteString(gr.SpreadMethod().String())
	b.WriteByte(0)
	bits(gr.X1())
	bits(gr.Y1())
	bits(gr.X2())
	bits(gr.Y2())
	for _, stop := range gr.Stops() {
		c := stop.Color().ToColor(mapper)
		bits(stop.Offset())
		b.WriteString(color.ToHexRGBColor(c.RGB()))
		b.WriteByte(0)
		bits(stop.Opacity())
	}
	return b.String()
}

// SetFillColor sets the fill of the next shapes; "" (upstream's null) and
// "#00000000" become "none".
func (s *SvgGraphics) SetFillColor(fill string) { s.SetFillColorWith(fill, color.WithFillNone) }

// SetFillColorWith sets the fill; with WithFillOpacity a transparent color
// is kept as is (written as a zero fill-opacity).
func (s *SvgGraphics) SetFillColorWith(fill string, behavior color.TransparentFillBehavior) {
	switch behavior {
	case color.WithFillNone:
		s.fill = fixColor(fill)
	case color.WithFillOpacity:
		s.fill = fill
	}
}

// SetStrokeColor sets the stroke of the next shapes; "" (upstream's null)
// and "#00000000" become "none".
func (s *SvgGraphics) SetStrokeColor(stroke string) { s.stroke = fixColor(stroke) }

// setStrokeColorEmpty is upstream's setStrokeColor("") (an empty string, not
// null): the stroke becomes "", not "none".
func (s *SvgGraphics) setStrokeColorEmpty() { s.stroke = "" }

func fixColor(c string) string {
	if c == "" || c == "#00000000" {
		return "none"
	}
	return c
}

// shortenColor turns #RRGGBB into #RGB when each pair has two identical
// digits; other values are returned unchanged.
func shortenColor(c string) string {
	if len(c) != 7 || c[0] != '#' {
		return c
	}
	if c[1] == c[2] && c[3] == c[4] && c[5] == c[6] {
		return "#" + c[1:2] + c[3:4] + c[5:6]
	}
	return c
}

// SetStrokeWidth sets the stroke width and dashes (nil for a plain line) of
// the next shapes.
func (s *SvgGraphics) SetStrokeWidth(strokeWidth float64, strokeDasharray []float64) {
	s.strokeWidth = s.Format(strokeWidth)
	if strokeDasharray == nil {
		s.strokeDasharray = ""
	} else {
		s.strokeDasharray = s.Format(strokeDasharray[0]) + "," + s.Format(strokeDasharray[1])
	}
}

// G returns the element shapes are added to: the innermost pending group
// or link, else the root group.
func (s *SvgGraphics) G() *XMLNode {
	if len(s.pendingElements) == 0 {
		return s.gRoot
	}
	return s.pendingElements[0]
}

// SvgRectangle adds a rectangle; nothing is drawn when the width or the
// height is not positive.
func (s *SvgGraphics) SvgRectangle(x, y, width, height, rx, ry, deltaShadow float64) {
	if height <= 0 || width <= 0 {
		return
	}
	s.manageShadow(deltaShadow)
	if !s.hidden {
		elt := s.createRectangleInternal(x, y, width, height)
		s.addFilterShadowID(elt, deltaShadow)
		if rx > 0 && ry > 0 {
			elt.SetAttribute("rx", s.Format(rx))
			elt.SetAttribute("ry", s.Format(ry))
		}
		s.G().AppendChild(elt)
	}
	s.ensureVisible(x+width+float64(2*deltaShadow), y+height+float64(2*deltaShadow))
}

func (s *SvgGraphics) createRectangleInternal(x, y, width, height float64) *XMLNode {
	elt := s.document.CreateElement("rect")
	elt.SetAttribute("x", s.Format(x))
	elt.SetAttribute("y", s.Format(y))
	elt.SetAttribute("width", s.Format(width))
	elt.SetAttribute("height", s.Format(height))
	s.fillMe(elt)
	s.styleMe(elt, "")
	return elt
}

// SvgLine adds a line.
func (s *SvgGraphics) SvgLine(x1, y1, x2, y2, deltaShadow float64) {
	s.manageShadow(deltaShadow)
	if !s.hidden {
		elt := s.document.CreateElement("line")
		elt.SetAttribute("x1", s.Format(x1))
		elt.SetAttribute("y1", s.Format(y1))
		elt.SetAttribute("x2", s.Format(x2))
		elt.SetAttribute("y2", s.Format(y2))
		s.styleMe(elt, "")
		s.addFilterShadowID(elt, deltaShadow)
		s.G().AppendChild(elt)
	}
	s.ensureVisible(x1+float64(2*deltaShadow), y1+float64(2*deltaShadow))
	s.ensureVisible(x2+float64(2*deltaShadow), y2+float64(2*deltaShadow))
}

func (s *SvgGraphics) styleMe(elt *XMLNode, suppStyle string) {
	if s.strokeWidth == "0" {
		return
	}
	var style strings.Builder
	style.WriteString("stroke:" + shortenColor(s.stroke) + ";")
	// Without stroke, the width and the dashes are useless.
	if s.stroke != "none" {
		style.WriteString("stroke-width:" + s.strokeWidth + ";")
		if s.strokeDasharray != "" {
			style.WriteString("stroke-dasharray:" + s.strokeDasharray + ";")
		}
	}
	style.WriteString(suppStyle)
	elt.SetAttribute("style", style.String())
}

// SvgPolygon adds a polygon of points x1, y1, x2, y2...
func (s *SvgGraphics) SvgPolygon(deltaShadow float64, points ...float64) {
	s.manageShadow(deltaShadow)
	if !s.hidden {
		elt := s.document.CreateElement("polygon")
		var sb strings.Builder
		for _, coord := range points {
			if sb.Len() > 0 {
				sb.WriteByte(',')
			}
			sb.WriteString(s.Format(coord))
		}
		elt.SetAttribute("points", sb.String())
		s.fillMe(elt)
		s.styleMe(elt, "stroke-linejoin:miter;stroke-miterlimit:10;")
		s.addFilterShadowID(elt, deltaShadow)
		s.G().AppendChild(elt)
	}
	for i := 0; i+1 < len(points); i += 2 {
		s.ensureVisible(points[i]+float64(2*deltaShadow), points[i+1]+float64(2*deltaShadow))
	}
	if len(points)%2 != 0 {
		// Upstream reads points[i + 1] and throws.
		throw("java.lang.ArrayIndexOutOfBoundsException", "Index "+strconv.Itoa(len(points))+
			" out of bounds for length "+strconv.Itoa(len(points)))
	}
}

// TextOptions are the optional parts of a text: font family ("" for none),
// weight, style and decoration ("" for none), extra attributes, the color
// of a background filter ("" for none) and the orientation (0, 90 or 270).
type TextOptions struct {
	FontFamily     string
	FontWeight     string
	FontStyle      string
	TextDecoration string
	// Attributes are written in their sorted order; nil means none.
	Attributes    *klimt.SvgAttributes
	TextBackColor string
	Orientation   int
}

// Text adds a text whose baseline starts at (x, y).
func (s *SvgGraphics) Text(txt string, x, y float64, fontSize int, textLength float64, opt TextOptions) {
	if !s.hidden {
		elt := s.document.CreateElement("text")
		elt.SetAttribute("x", s.Format(x))
		elt.SetAttribute("y", s.Format(y))
		s.fillMe(elt)

		switch opt.Orientation {
		case 90:
			elt.SetAttribute("transform", "rotate(-90 "+s.Format(x)+" "+s.Format(y)+")")
		case 270:
			elt.SetAttribute("transform", "rotate(90 "+s.Format(x)+" "+s.Format(y)+")")
		}
		elt.SetAttribute("font-size", s.Format(float64(fontSize)))

		// lengthAdjust is set on the root group and inherited; textLength
		// is not inheritable. A single glyph has no spacing to adjust.
		if javaLen(txt) > 1 && (s.option.LengthAdjust() == klimt.LengthAdjustSpacing ||
			s.option.LengthAdjust() == klimt.LengthAdjustSpacingAndGlyphs) {
			elt.SetAttribute("textLength", s.Format(textLength))
		}

		if opt.FontWeight != "" {
			elt.SetAttribute("font-weight", opt.FontWeight)
		}
		if opt.FontStyle != "" {
			elt.SetAttribute("font-style", opt.FontStyle)
		}
		if opt.TextDecoration != "" {
			elt.SetAttribute("text-decoration", opt.TextDecoration)
		}

		if fontFamily := opt.FontFamily; fontFamily != "" {
			if javaEqualsIgnoreCase("roboto", fontFamily) {
				s.addRoboto()
			}
			if javaEqualsIgnoreCase("monospaced", fontFamily) {
				fontFamily = "monospace"
			}
			if !javaEqualsIgnoreCase(fontFamily, defaultFontFamily) {
				elt.SetAttribute("font-family", fontFamily)
			}
			if javaEqualsIgnoreCase(fontFamily, "monospace") || javaEqualsIgnoreCase(fontFamily, "courier") {
				txt = strings.ReplaceAll(txt, " ", "\u00a0")
			}
		}
		if opt.TextBackColor != "" {
			backFilterID := s.filterBackColorID(opt.TextBackColor)
			elt.SetAttribute("filter", "url(#"+backFilterID+")")
		}
		if opt.Attributes != nil {
			attributes := opt.Attributes.Attributes()
			for _, k := range opt.Attributes.Keys() {
				elt.SetAttribute(k, attributes[k])
			}
		}

		elt.SetTextContent(txt)
		s.G().AppendChild(elt)
	}
	s.ensureVisible(x, y)
	s.ensureVisible(x+textLength, y)
}

func (s *SvgGraphics) idFilterBackColor(c string) string {
	result, ok := s.filterBackColor[c]
	if !ok {
		result = s.filterUID + strconv.Itoa(len(s.filterBackColor))
		s.filterBackColor[c] = result
	}
	return result
}

func (s *SvgGraphics) filterBackColorID(c string) string {
	if id, ok := s.filterBackColor[c]; ok {
		return id
	}
	id := s.idFilterBackColor(c)
	filter := s.document.CreateElement("filter")
	filter.SetAttribute("id", id)
	filter.SetAttribute("x", "0")
	filter.SetAttribute("y", "0")
	filter.SetAttribute("width", "1")
	filter.SetAttribute("height", "1")
	s.addFilter(filter, "feFlood", "flood-color", c, "result", "flood")
	s.addFilter(filter, "feComposite", "in", "SourceGraphic", "in2", "flood", "operator", "over")
	s.defs.AppendChild(filter)
	return id
}

// CreateXML writes the document: the root size attributes are set from
// the drawn area, then inlined SVG images replace their placeholders.
func (s *SvgGraphics) CreateXML(w io.Writer) error {
	s.finalizeRootAttributes()
	out := s.document.ToXML(0)
	for _, img := range s.images {
		out = strings.ReplaceAll(out, "<"+img.key+"/>", img.svg)
	}
	_, err := io.WriteString(w, out)
	return err
}

func (s *SvgGraphics) finalizeRootAttributes() {
	maxXscaled := javaD2I(float64(s.maxX) * s.option.Scale())
	maxYscaled := javaD2I(float64(s.maxY) * s.option.Scale())
	style := "width:" + strconv.Itoa(maxXscaled) + "px;height:" + strconv.Itoa(maxYscaled) + "px;"

	if s.hasBackcolor && s.backcolorString != "#00000000" {
		style += "background:" + s.backcolorString + ";"
	}

	if s.option.SvgDimensionStyle() {
		s.root.SetAttribute("style", style)
		s.root.SetAttribute("width", s.Format(float64(s.maxX))+"px")
		s.root.SetAttribute("height", s.Format(float64(s.maxY))+"px")
	}
	s.root.SetAttribute("viewBox", "0 0 "+strconv.Itoa(maxXscaled)+" "+strconv.Itoa(maxYscaled))
	s.root.SetAttribute("zoomAndPan", "magnify")
	s.root.SetAttribute("preserveAspectRatio", s.option.PreserveAspectRatio())
	s.root.SetAttribute("contentStyleType", "text/css")

	if s.pendingBackground != nil {
		s.pendingBackground.SetAttribute("width", s.Format(float64(s.maxX)))
		s.pendingBackground.SetAttribute("height", s.Format(float64(s.maxY)))
	}
}

// SvgPath adds a path translated by (x, y).
func (s *SvgGraphics) SvgPath(x, y float64, path *klimt.UPath, deltaShadow float64) {
	s.manageShadow(deltaShadow)
	s.ensureVisible(x, y)
	var sb strings.Builder
	shadow2 := float64(2 * deltaShadow)
	for _, seg := range path.Segments() {
		coord := seg.Coord()
		switch seg.SegmentType() {
		case geom.SegMoveTo:
			sb.WriteString("M" + s.Format(coord[0]+x) + "," + s.Format(coord[1]+y) + " ")
			s.ensureVisible(coord[0]+x+shadow2, coord[1]+y+shadow2)
		case geom.SegLineTo:
			sb.WriteString("L" + s.Format(coord[0]+x) + "," + s.Format(coord[1]+y) + " ")
			s.ensureVisible(coord[0]+x+shadow2, coord[1]+y+shadow2)
		case geom.SegQuadTo:
			sb.WriteString("Q" + s.Format(coord[0]+x) + "," + s.Format(coord[1]+y) + " " + s.Format(coord[2]+x) +
				"," + s.Format(coord[3]+y) + " ")
			s.ensureVisible(coord[0]+x+shadow2, coord[1]+y+shadow2)
			s.ensureVisible(coord[2]+x+shadow2, coord[3]+y+shadow2)
		case geom.SegCubicTo:
			sb.WriteString("C" + s.Format(coord[0]+x) + "," + s.Format(coord[1]+y) + " " + s.Format(coord[2]+x) +
				"," + s.Format(coord[3]+y) + " " + s.Format(coord[4]+x) + "," + s.Format(coord[5]+y) + " ")
			s.ensureVisible(coord[0]+x+shadow2, coord[1]+y+shadow2)
			s.ensureVisible(coord[2]+x+shadow2, coord[3]+y+shadow2)
			s.ensureVisible(coord[4]+x+shadow2, coord[5]+y+shadow2)
		case geom.SegArcTo:
			sb.WriteString("A" + s.Format(coord[0]) + "," + s.Format(coord[1]) + " " + s.Format(coord[2]) + " " +
				formatBoolean(coord[3]) + " " + formatBoolean(coord[4]) + " " + s.Format(coord[5]+x) + "," +
				s.Format(coord[6]+y) + " ")
			s.ensureVisible(coord[5]+coord[0]+x+shadow2, coord[6]+coord[1]+y+shadow2)
		case geom.SegClose:
			// Nothing
		}
	}
	if !s.hidden {
		elt := s.document.CreateElement("path")
		elt.SetAttribute("d", text.JavaTrim(sb.String()))
		s.styleMe(elt, "")
		s.fillMe(elt)
		if id := path.Comment(); id != "" {
			elt.SetAttribute("id", id)
		}
		if codeLine := path.CodeLine(); codeLine != "" {
			elt.SetAttribute("codeLine", codeLine)
		}
		s.addFilterShadowID(elt, deltaShadow)
		s.G().AppendChild(elt)
	}
}

var hexColorWithAlpha = regexp.MustCompile(`^#[0-9A-Fa-f]{8}$`)

func (s *SvgGraphics) fillMe(elt *XMLNode) {
	if hexColorWithAlpha.MatchString(s.fill) {
		elt.SetAttribute("fill", shortenColor(s.fill[:7]))
		alpha, _ := strconv.ParseInt(s.fill[7:], 16, 32)
		elt.SetAttribute("fill-opacity", s.FormatOpacity(float64(alpha)/255.0))
	} else {
		elt.SetAttribute("fill", shortenColor(s.fill))
	}
}

func (s *SvgGraphics) addFilterShadowID(elt *XMLNode, deltaShadow float64) {
	if deltaShadow > 0 {
		elt.SetAttribute("filter", "url(#"+s.shadowID+")")
	}
}

// NewPath starts a filled path, written by Fill.
func (s *SvgGraphics) NewPath() { s.currentPath = &strings.Builder{} }

// MoveTo adds a move to the current path.
func (s *SvgGraphics) MoveTo(x, y float64) {
	s.currentPath.WriteString("M" + s.Format(x) + "," + s.Format(y) + " ")
	s.ensureVisible(x, y)
}

// LineTo adds a line to the current path.
func (s *SvgGraphics) LineTo(x, y float64) {
	s.currentPath.WriteString("L" + s.Format(x) + "," + s.Format(y) + " ")
	s.ensureVisible(x, y)
}

// ClosePath closes the current subpath.
func (s *SvgGraphics) ClosePath() { s.currentPath.WriteString("Z ") }

// CurveTo adds a cubic curve to the current path.
func (s *SvgGraphics) CurveTo(x1, y1, x2, y2, x3, y3 float64) {
	s.currentPath.WriteString("C" + s.Format(x1) + "," + s.Format(y1) + " " + s.Format(x2) + "," + s.Format(y2) +
		" " + s.Format(x3) + "," + s.Format(y3) + " ")
	s.ensureVisible(x1, y1)
	s.ensureVisible(x2, y2)
	s.ensureVisible(x3, y3)
}

// QuadTo adds a quadratic curve to the current path.
func (s *SvgGraphics) QuadTo(x1, y1, x2, y2 float64) {
	s.currentPath.WriteString("Q" + s.Format(x1) + "," + s.Format(y1) + " " + s.Format(x2) + "," + s.Format(y2) + " ")
	s.ensureVisible(x1, y1)
	s.ensureVisible(x2, y2)
}

// Fill writes the current path, filled and without stroke; the winding
// rule is not written.
func (s *SvgGraphics) Fill(windingRule int) {
	if !s.hidden {
		elt := s.document.CreateElement("path")
		elt.SetAttribute("d", s.currentPath.String())
		s.fillMe(elt)
		s.G().AppendChild(elt)
	}
	s.currentPath = nil
}

// PathIterator is java.awt.geom.PathIterator: the segments of a glyph
// outline.
type PathIterator interface {
	IsDone() bool
	// CurrentSegment fills coords and returns the segment type
	// (geom.SegMoveTo, SegLineTo, SegQuadTo, SegCubicTo or SegClose).
	CurrentSegment(coords *[6]float64) geom.USegmentType
	Next()
	WindingRule() int
}

// DrawPathIterator writes a filled path from an outline translated by
// (x, y).
func (s *SvgGraphics) DrawPathIterator(x, y float64, path PathIterator) {
	s.NewPath()
	var coord [6]float64
	for !path.IsDone() {
		switch code := path.CurrentSegment(&coord); code {
		case geom.SegMoveTo:
			s.MoveTo(coord[0]+x, coord[1]+y)
		case geom.SegLineTo:
			s.LineTo(coord[0]+x, coord[1]+y)
		case geom.SegClose:
			s.ClosePath()
		case geom.SegCubicTo:
			s.CurveTo(coord[0]+x, coord[1]+y, coord[2]+x, coord[3]+y, coord[4]+x, coord[5]+y)
		case geom.SegQuadTo:
			s.QuadTo(coord[0]+x, coord[1]+y, coord[2]+x, coord[3]+y)
		default:
			throw(unsupportedOperation, "code="+strconv.Itoa(int(code)))
		}
		path.Next()
	}
	s.Fill(path.WindingRule())
}

// SvgImagePng adds a raster image as a PNG data URL.
//
// Upstream encodes the PNG with Java's ImageIO; here PortableImage's
// ToPngDataURL encodes it with Go's image/png, so the base64 data is a
// valid PNG of the same pixels but not byte-identical.
func (s *SvgGraphics) SvgImagePng(image klimt.PortableImage, x, y float64) {
	if !s.hidden {
		elt := s.document.CreateElement("image")
		elt.SetAttribute("width", s.Format(float64(image.Width())))
		elt.SetAttribute("height", s.Format(float64(image.Height())))
		elt.SetAttribute("x", s.Format(x))
		elt.SetAttribute("y", s.Format(y))
		elt.SetAttribute("xlink:href", image.ToPngDataURL())
		s.G().AppendChild(elt)
	}
	s.ensureVisible(x, y)
	s.ensureVisible(x+float64(image.Width()), y+float64(image.Height()))
}

func (s *SvgGraphics) svgImageUnsecure(image *klimt.UImageSvg, x, y float64) {
	if !s.hidden {
		svg := s.manageScale(image)
		pos := `<svg x="` + s.Format(x) + `" y="` + s.Format(y) + `">`
		svg = pos + svg[5:]
		key := "imagesvginlined" + image.MD5Hex() + strconv.Itoa(len(s.images))
		elt := s.document.CreateElement(key)
		s.G().AppendChild(elt)
		s.images = append(s.images, inlinedImage{key: key, svg: svg})
	}
	s.ensureVisible(x, y)
	s.ensureVisible(x+float64(image.Data("width")), y+float64(image.Data("height")))
}

// SvgImageSvg adds an SVG image: embedded as a base64 data URL, or inlined
// with the INSECURE security profile.
func (s *SvgGraphics) SvgImageSvg(image *klimt.UImageSvg, x, y float64) {
	if s.option.SecurityProfile() == security.Insecure {
		s.svgImageUnsecure(image, x, y)
		return
	}
	if !s.hidden {
		elt := s.document.CreateElement("image")
		elt.SetAttribute("width", s.Format(image.Width()))
		elt.SetAttribute("height", s.Format(image.Height()))
		elt.SetAttribute("x", s.Format(x))
		elt.SetAttribute("y", s.Format(y))

		svg := s.manageScale(image)
		height := strconv.Itoa(javaD2I(image.Height() * s.option.Scale()))
		width := strconv.Itoa(javaD2I(image.Width() * s.option.Scale()))
		var svgHeader string
		if image.ContainsXlink() {
			svgHeader = `<svg height="` + height + `" width="` + width +
				`" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns="http://www.w3.org/2000/svg" >`
		} else {
			svgHeader = `<svg height="` + height + `" width="` + width + `" xmlns="http://www.w3.org/2000/svg" >`
		}
		svg = svgHeader + svg[5:]
		elt.SetAttribute("xlink:href", "data:image/svg+xml;base64,"+base64.StdEncoding.EncodeToString([]byte(svg)))
		s.G().AppendChild(elt)
	}
	s.ensureVisible(x, y)
	s.ensureVisible(x+float64(image.Data("width")), y+float64(image.Data("height")))
}

var (
	svgOpenTag  = regexp.MustCompile(`<svg>`)
	svgCloseTag = regexp.MustCompile(`</svg>`)
	gOpenTag    = regexp.MustCompile(`<g\b`)
)

func (s *SvgGraphics) manageScale(svgImage *klimt.UImageSvg) string {
	svgScale := svgImage.Scale()
	svg := svgImage.Svg(false)
	if svgScale*s.option.Scale() == 1 {
		return svg
	}
	svg2 := strings.NewReplacer("\n", " ", "\r", " ").Replace(svg)
	if !strings.Contains(svg2, "<g ") && !strings.Contains(svg2, "<g>") {
		svg = replaceFirst(svgOpenTag, svg, "<svg><g>")
		svg = replaceFirst(svgCloseTag, svg, "</g></svg>")
	}
	factor := s.Format(svgScale)
	return replaceFirst(gOpenTag, svg, `<g transform="scale(`+factor+`,`+factor+`)" `)
}

// replaceFirst is Java's String.replaceFirst with a literal replacement.
func replaceFirst(re *regexp.Regexp, s, replacement string) string {
	loc := re.FindStringIndex(s)
	if loc == nil {
		return s
	}
	return s[:loc[0]] + replacement + s[loc[1]:]
}

// manageShadow adds the shadow filter to the definitions on first use.
func (s *SvgGraphics) manageShadow(deltaShadow float64) {
	if deltaShadow != 0 {
		if !s.withShadow {
			filter := s.document.CreateElement("filter")
			filter.SetAttribute("id", s.shadowID)
			filter.SetAttribute("x", "-1")
			filter.SetAttribute("y", "-1")
			filter.SetAttribute("width", "300%")
			filter.SetAttribute("height", "300%")
			s.addFilter(filter, "feGaussianBlur", "result", "blurOut", "stdDeviation", s.Format(2))
			s.addFilter(filter, "feColorMatrix", "type", "matrix", "in", "blurOut", "result", "blurOut2", "values",
				"0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 .4 0")
			s.addFilter(filter, "feOffset", "result", "blurOut3", "in", "blurOut2", "dx", s.Format(4), "dy",
				s.Format(4))
			s.addFilter(filter, "feBlend", "in", "SourceGraphic", "in2", "blurOut3", "mode", "normal")
			s.defs.AppendChild(filter)
		}
		s.withShadow = true
	}
}

func (s *SvgGraphics) addFilter(filter *XMLNode, name string, data ...string) {
	elt := s.document.CreateElement(name)
	for i := 0; i+1 < len(data); i += 2 {
		elt.SetAttribute(data[i], data[i+1])
	}
	filter.AppendChild(elt)
}

// SetHidden makes the next shapes count in the canvas size without being
// written.
func (s *SvgGraphics) SetHidden(hidden bool) { s.hidden = hidden }

// MetadataHex is upstream's getMetadataHex: the metadata text in the URL
// text format (upstream's protected transcoder encodes like the plain
// one). The deflate step is Go's compress/flate, not Java's zlib, so the
// text can differ from upstream's while decoding to the same source.
func MetadataHex(comment string) string { return code.Encode(comment) }

// AddCommentMetadata adds the <?plantuml-src ...?> instruction with the
// encoded metadata.
func (s *SvgGraphics) AddCommentMetadata(metadata string) {
	s.G().AppendProcessingInstruction("plantuml-src", MetadataHex(metadata))
}

// AddComment adds a comment.
func (s *SvgGraphics) AddComment(comment string) { s.G().AppendComment(comment) }

// linkData is an open link.
type linkData struct {
	url    string
	title  string
	target string
}

func (s *SvgGraphics) newLinkData(url, title, target string) linkData {
	// "javascript:" links are a security issue.
	if !s.option.AllowJavascriptInLink() && isJavascriptLink(url) {
		url = ""
	}
	return linkData{url: url, title: title, target: target}
}

// isJavascriptLink is upstream's SecurityUtils.isJavascriptLink.
func isJavascriptLink(url string) bool {
	lower := javaToLowerCase(url)
	var letters strings.Builder
	for i := 0; i < len(lower); i++ {
		if c := lower[i]; c >= 'a' && c <= 'z' {
			letters.WriteByte(c)
		}
	}
	return strings.HasPrefix(letters.String(), "javascript")
}

var unicodeEscape = regexp.MustCompile(`<U\+([0-9A-Fa-f]+)>`)

// xlinkTitle is the title with its <U+XXXX> escapes decoded (as one UTF-16
// unit each, as upstream's char cast) and literal "\n" turned into
// newlines.
//
// Upstream builds the result with Matcher.appendReplacement, which reads
// '$' and '\' in the replacement: an escape of either throws there; here it
// gives the character (design D16).
func (l linkData) xlinkTitle() string {
	var units []uint16
	last := 0
	for _, m := range unicodeEscape.FindAllStringSubmatchIndex(l.title, -1) {
		units = append(units, toUTF16(l.title[last:m[0]])...)
		units = append(units, uint16(parseHexWrapping(l.title[m[2]:m[3]])))
		last = m[1]
	}
	units = append(units, toUTF16(l.title[last:])...)
	return strings.ReplaceAll(javaUTF16String(units), `\n`, "\n")
}

// parseHexWrapping parses hexadecimal digits, keeping the low bits. Upstream
// parses an int and casts it to char; values beyond the int range throw
// there and wrap here (design D16).
func parseHexWrapping(s string) uint64 {
	var v uint64
	for i := 0; i < len(s); i++ {
		c := s[i]
		var d uint64
		switch {
		case c >= '0' && c <= '9':
			d = uint64(c - '0')
		case c >= 'a' && c <= 'f':
			d = uint64(c-'a') + 10
		default:
			d = uint64(c-'A') + 10
		}
		v = v<<4 | d
	}
	return v
}

func (l linkData) updateAttributesOf(element *XMLNode) {
	element.SetAttribute("target", l.target)
	element.SetAttribute(xlinkHref1, l.url)
	element.SetAttribute(xlinkHref2, l.url)
	element.SetAttribute("xlink:type", "simple")
	element.SetAttribute("xlink:actuate", "onRequest")
	element.SetAttribute("xlink:show", "new")
	title := l.xlinkTitle()
	element.SetAttribute(xlinkTitle1, title)
	element.SetAttribute(xlinkTitle2, title)
}

// SVG does not support nested links: only the innermost open link is
// active, and opening one closes the element of the previous one.

// closeTopPendingElement moves the innermost pending element (link or
// group) to its parent, dropping it when it is empty.
func (s *SvgGraphics) closeTopPendingElement() {
	element := s.pendingElements[0]
	s.pendingElements = s.pendingElements[1:]
	if element.HasChildren() {
		s.G().AppendChild(element)
	}
}

func (s *SvgGraphics) closeTopActiveLinkIfNeeded() {
	if len(s.activeLinks) > 0 {
		if s.pendingElements[0].TagName() != "a" {
			throw(illegalState, "Expected top pending element to be a link.")
		}
		s.closeTopPendingElement()
	}
	for _, elt := range s.pendingElements {
		if elt.TagName() == "a" {
			throw(illegalState, "")
		}
	}
}

func (s *SvgGraphics) addTopOpenedLinkIfNeeded() {
	if len(s.activeLinks) > 0 {
		link := s.activeLinks[0]
		a := s.document.CreateElement("a")
		s.pendingElements = append([]*XMLNode{a}, s.pendingElements...)
		link.updateAttributesOf(a)
	}
}

// OpenLink opens a link; the title is the tooltip, upstream's non-null
// tooltip of a Url.
func (s *SvgGraphics) OpenLink(url, title, target string) {
	s.closeTopActiveLinkIfNeeded()
	s.activeLinks = append([]linkData{s.newLinkData(url, title, target)}, s.activeLinks...)
	s.addTopOpenedLinkIfNeeded()
}

// CloseLink closes the innermost link and reopens the enclosing one. It
// panics when no link is open, as upstream throws IllegalStateException.
func (s *SvgGraphics) CloseLink() {
	if len(s.pendingElements) == 0 || len(s.activeLinks) == 0 || s.pendingElements[0].TagName() != "a" {
		throw(illegalState, "Attempting to close a link in an invalid state.")
	}
	s.closeTopActiveLinkIfNeeded()
	s.activeLinks = s.activeLinks[1:]
	s.addTopOpenedLinkIfNeeded()
}

// CloseGroup closes the innermost group. It panics when none is open.
func (s *SvgGraphics) CloseGroup() {
	if len(s.pendingElements) == 0 {
		throw(illegalState, "")
	}
	s.closeTopActiveLinkIfNeeded()
	s.closeTopPendingElement()
	s.addTopOpenedLinkIfNeeded()
}

// StartGroup opens a <g> with the attributes of a group, which must not be
// empty.
func (s *SvgGraphics) StartGroup(typeIdents []klimt.UGroupEntry) {
	if len(typeIdents) == 0 {
		throw(illegalArgument, "")
	}
	s.closeTopActiveLinkIfNeeded()
	g := s.document.CreateElement("g")
	s.pendingElements = append([]*XMLNode{g}, s.pendingElements...)
	for _, e := range typeIdents {
		s.document.ApplyGroupAttribute(g, e.Key, e.Value)
	}
	s.addTopOpenedLinkIfNeeded()
}

// javaD2I is Java's (int) cast of a double: truncation toward zero, NaN to
// zero and saturation at the int range.
func javaD2I(x float64) int {
	switch {
	case math.IsNaN(x):
		return 0
	case x >= math.MaxInt32:
		return math.MaxInt32
	case x <= math.MinInt32:
		return math.MinInt32
	}
	return int(x)
}
