// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svg/parser/SvgSaxParser.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// SvgSaxParser is the SVG sprite parser of "!pragma svgparser sax": an XML
// parser (Go's encoding/xml in place of Java's SAX parser, see
// svgparser_saxreader.go) of the SVG 1.1 core subset.
//
// Feature set: rect, circle, ellipse, line, polyline, polygon and path; text
// with font family, size, weight, style, decoration and text-anchor;
// translate, rotate, scale and matrix transforms; linear gradients (the
// first stop color for radial ones); groups (their transform only); defs,
// symbol and use; images of PNG and JPEG data URLs.
//
// Two passes: the first collects the definitions (defs, symbols, gradients)
// by id, the second draws. A document that is not well-formed fails in the
// first pass, before anything is drawn. Upstream's java.util.logging
// messages are not written.
type SvgSaxParser struct {
	svg string
}

var (
	_ ISvgSpriteParser = (*SvgSaxParser)(nil)
	_ GrayLevelRange   = (*SvgSaxParser)(nil)
)

// NewSvgSaxParser returns the parser of an SVG text.
func NewSvgSaxParser(svg string) *SvgSaxParser { return &SvgSaxParser{svg: svg} }

// svgSaxMaxUseDepth bounds the nesting of <use> references. Upstream
// recurses until the JVM throws StackOverflowError on a reference cycle; a
// Go stack overflow would end the process, so the same error is thrown at
// this depth.
const svgSaxMaxUseDepth = 256

// DrawU draws the sprite. Any exception (a document that is not
// well-formed, a number that does not parse) panics with Java's
// RuntimeException "Failed to parse SVG", after what the second pass drew.
func (p *SvgSaxParser) DrawU(ug UGraphic, scale float64, fontColor, forcedColor color.HColor) {
	colorResolver := NewColorResolver(fontColor, forcedColor, p)
	ugs := NewUGraphicWithScale(ug, colorResolver, scale)

	if err := svgSaxCatch(func() error { return p.parseSvg(p.svg, ugs, colorResolver) }); err != nil {
		throw("java.lang.RuntimeException", "Failed to parse SVG")
	}
}

// svgSaxCatch runs f as Java's try { ... } catch (Exception e): it returns
// the error of f, or the exception that f throws. Java errors (such as
// StackOverflowError) and Go runtime panics go through.
func svgSaxCatch(f func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case *color.RuntimeError:
				if strings.HasSuffix(e.Class, "Error") {
					panic(r)
				}
				err = e
			case *color.NoSuchColorRuntimeError:
				err = e
			default:
				panic(r)
			}
		}
	}()
	return f()
}

// AsTextBlock returns the sprite as a block of the size of the SVG (its
// viewBox, else its width and height) times scale, over backColor when it is
// not nil. A size that cannot be found panics, as upstream throws
// IllegalStateException.
func (p *SvgSaxParser) AsTextBlock(fontColor, forcedColor color.HColor, scale float64,
	backColor color.HColor) TextBlock {
	return newSvgSpriteTextBlock(p.svg, p, fontColor, forcedColor, scale, backColor)
}

// MinGrayLevel returns 0.
func (p *SvgSaxParser) MinGrayLevel() int { return 0 }

// MaxGrayLevel returns 255.
func (p *SvgSaxParser) MaxGrayLevel() int { return 255 }

// parseSvg parses the SVG in two passes.
func (p *SvgSaxParser) parseSvg(svgContent string, ugs *UGraphicWithScale, colorResolver *ColorResolver) error {
	// Pass 1: Collect definitions
	defsCollector := newSvgSaxDefsCollector()
	if err := saxParse(svgContent, defsCollector); err != nil {
		return err
	}

	// Pass 2: Render elements
	renderer := newSvgSaxRenderHandler(ugs, colorResolver, defsCollector.definitions, 0)
	return saxParse(svgContent, renderer)
}

// svgSaxBufferedElement is a definition: its tag, its XML text (re-parsed
// by <use>), its attributes and its gradient stops.
type svgSaxBufferedElement struct {
	tagName    string
	xmlContent string
	attributes map[string]string
	stops      []svgSaxGradientStop
}

// svgSaxGradientStop is a gradient stop as written; the opacity may be
// missing.
type svgSaxGradientStop struct {
	offset     string
	color      string
	opacity    string
	hasOpacity bool
}

// svgSaxDefsCollector is the first pass: it collects the definitions
// (defs, symbols, gradients) by id. As upstream, an element with an id
// inside a definition replaces it as the one being collected.
type svgSaxDefsCollector struct {
	definitions    map[string]*svgSaxBufferedElement
	depth          int
	inDefs         bool
	currentContent *strings.Builder
	currentID      string // "" for none
	currentTag     string // "" for none
	currentAttrs   map[string]string
	currentElement *svgSaxBufferedElement
}

func newSvgSaxDefsCollector() *svgSaxDefsCollector {
	return &svgSaxDefsCollector{definitions: map[string]*svgSaxBufferedElement{},
		currentContent: &strings.Builder{}}
}

func (c *svgSaxDefsCollector) startElement(qName string, attrs saxAttributes) {
	c.depth++

	if qName == "defs" || qName == "symbol" || qName == "linearGradient" || qName == "radialGradient" {
		c.inDefs = true
	}

	if c.inDefs || qName == "symbol" {
		if id, ok := attrs.value("id"); ok && id != "" {
			c.currentID = id
			c.currentTag = qName
			c.currentContent = &strings.Builder{}
			c.currentAttrs = map[string]string{}
			// Store all attributes for gradient processing
			for _, a := range attrs {
				c.currentAttrs[a.qName] = a.value
			}
			c.currentElement = &svgSaxBufferedElement{tagName: qName, attributes: c.currentAttrs}
		}

		// Handle gradient stops
		if qName == "stop" && c.currentElement != nil {
			offset, hasOffset := attrs.value("offset")
			stopColor, hasStopColor := attrs.value("stop-color")
			stopOpacity, hasStopOpacity := attrs.value("stop-opacity")
			if style, ok := attrs.value("style"); ok && style != "" {
				if !hasStopColor {
					stopColor, hasStopColor = svgSaxStyleValue(style, "stop-color")
				}
				if !hasStopOpacity {
					stopOpacity, hasStopOpacity = svgSaxStyleValue(style, "stop-opacity")
				}
			}
			if hasOffset && hasStopColor {
				c.currentElement.stops = append(c.currentElement.stops, svgSaxGradientStop{offset: offset,
					color: stopColor, opacity: stopOpacity, hasOpacity: hasStopOpacity})
			}
		}

		// Store element as buffered XML
		c.currentContent.WriteString("<" + qName)
		for _, a := range attrs {
			c.currentContent.WriteString(" " + a.qName + "=\"" + svgSaxEscapeXML(a.value) + "\"")
		}
		c.currentContent.WriteString(">")
	}
}

func (c *svgSaxDefsCollector) endElement(qName string) {
	if c.inDefs {
		c.currentContent.WriteString("</" + qName + ">")

		if c.currentID != "" && c.currentTag != "" && c.currentTag == qName {
			if c.currentElement != nil {
				buffered := &svgSaxBufferedElement{tagName: c.currentTag, xmlContent: c.currentContent.String(),
					attributes: c.currentAttrs}
				buffered.stops = append(buffered.stops, c.currentElement.stops...)
				c.definitions[c.currentID] = buffered
			}
			c.currentID = ""
			c.currentTag = ""
			c.currentAttrs = nil
			c.currentElement = nil
		}
	}

	c.depth--
	if (qName == "defs" || qName == "symbol") && c.depth == 1 {
		c.inDefs = false
	}
}

func (c *svgSaxDefsCollector) characters(text string) {
	if c.inDefs {
		c.currentContent.WriteString(svgSaxEscapeXML(text))
	}
}

func svgSaxEscapeXML(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	return strings.ReplaceAll(text, "\"", "&quot;")
}

// svgSaxStyleValue returns the value of a property of a style attribute
// ("fill:red;stroke:blue"), the name compared ignoring case; ok is false
// when the property is not there.
func svgSaxStyleValue(style, key string) (string, bool) {
	for _, part := range strings.Split(style, ";") {
		idx := strings.IndexByte(part, ':')
		if idx <= 0 {
			continue
		}
		name := javaTrim(part[:idx])
		if javaEqualsIgnoreCase(name, key) {
			return javaTrim(part[idx+1:]), true
		}
	}
	return "", false
}

// svgSaxGroupState is a group being drawn; only its transform is used.
type svgSaxGroupState struct {
	id        string
	className string
	transform string
}

// svgSaxRenderHandler is the second pass: it draws the elements outside
// defs, re-parsing the definitions that <use> references with the current
// transform and styling.
type svgSaxRenderHandler struct {
	initialUgs    *UGraphicWithScale
	ugs           *UGraphicWithScale
	colorResolver *ColorResolver
	definitions   map[string]*svgSaxBufferedElement
	groupStack    []svgSaxGroupState
	ugsStack      []*UGraphicWithScale
	defsDepth     int
	useDepth      int

	inText      bool
	textContent strings.Builder
	textAttrs   saxAttributes
}

func newSvgSaxRenderHandler(ugs *UGraphicWithScale, colorResolver *ColorResolver,
	definitions map[string]*svgSaxBufferedElement, useDepth int) *svgSaxRenderHandler {
	return &svgSaxRenderHandler{initialUgs: ugs, ugs: ugs, colorResolver: colorResolver, definitions: definitions,
		useDepth: useDepth}
}

func (r *svgSaxRenderHandler) startElement(qName string, attrs saxAttributes) {
	if qName == "defs" {
		r.defsDepth++
		return
	}
	if r.defsDepth > 0 {
		return
	}
	switch qName {
	case "g":
		r.handleGroupStart(attrs)
	case "rect":
		r.handleRect(attrs)
	case "circle":
		r.handleCircle(attrs)
	case "ellipse":
		r.handleEllipse(attrs)
	case "line":
		r.handleLine(attrs)
	case "polyline":
		r.handlePolyShape(attrs, false)
	case "polygon":
		r.handlePolyShape(attrs, true)
	case "path":
		r.handlePath(attrs)
	case "text":
		r.inText = true
		r.textContent.Reset()
		r.textAttrs = append(saxAttributes{}, attrs...)
	case "image":
		r.handleImage(attrs)
	case "use":
		r.handleUse(attrs)
	}
	// svg (the root element) and the other elements (symbol, title, desc,
	// metadata, clipPath, mask, filter, pattern, marker, style, script and
	// unknown elements) are skipped: CSS style blocks are not supported (use
	// inline attributes); supporting CSS would require a full CSS parser for
	// selectors, specificity, cascade, and inheritance rules
}

func (r *svgSaxRenderHandler) endElement(qName string) {
	if qName == "defs" && r.defsDepth > 0 {
		r.defsDepth--
		return
	}
	if r.defsDepth > 0 {
		return
	}
	if qName == "g" {
		r.handleGroupEnd()
	} else if qName == "text" && r.inText {
		r.handleText(r.textAttrs, r.textContent.String())
		r.inText = false
		r.textContent.Reset()
		r.textAttrs = nil
	}
}

func (r *svgSaxRenderHandler) characters(text string) {
	if r.inText {
		r.textContent.WriteString(text)
	}
}

// ---- Group Handling ----

func (r *svgSaxRenderHandler) handleGroupStart(attrs saxAttributes) {
	id, _ := attrs.value("id")
	className, _ := attrs.value("class")
	transform, _ := attrs.value("transform")
	gs := svgSaxGroupState{id: id, className: className, transform: transform}
	r.groupStack = append([]svgSaxGroupState{gs}, r.groupStack...)
	r.ugsStack = append([]*UGraphicWithScale{r.ugs}, r.ugsStack...)

	if gs.transform != "" {
		r.ugs = r.applyTransform(r.ugs, gs.transform)
	}
}

func (r *svgSaxRenderHandler) handleGroupEnd() {
	if len(r.groupStack) > 0 {
		r.groupStack = r.groupStack[1:]
	}
	if len(r.ugsStack) > 0 {
		r.ugs = r.ugsStack[0]
		r.ugsStack = r.ugsStack[1:]
	}
}

// ---- Shape Handlers ----

func (r *svgSaxRenderHandler) handleRect(attrs saxAttributes) {
	elementUgs := r.applyStyleAndTransform(attrs, r.ugs)

	x := svgSaxDoubleAttr(attrs, "x", 0)
	y := svgSaxDoubleAttr(attrs, "y", 0)
	width := svgSaxDoubleAttr(attrs, "width", 0)
	height := svgSaxDoubleAttr(attrs, "height", 0)

	path := UPathNone()
	path.MoveTo(x, y)
	path.LineTo(x+width, y)
	path.LineTo(x+width, y+height)
	path.LineTo(x, y+height)
	path.LineTo(x, y)

	path = path.Affine(elementUgs.AffineTransform(), elementUgs.Angle(), elementUgs.InitialScale())
	elementUgs.Draw(path)
}

func (r *svgSaxRenderHandler) handleCircle(attrs saxAttributes) {
	elementUgs := r.applyStyleAndTransform(attrs, r.ugs)

	cx := svgSaxDoubleAttr(attrs, "cx", 0)
	cy := svgSaxDoubleAttr(attrs, "cy", 0)
	radius := svgSaxDoubleAttr(attrs, "r", 0)

	path := svgSaxEllipsePath(cx, cy, radius, radius)
	path = path.Affine(elementUgs.AffineTransform(), elementUgs.Angle(), elementUgs.InitialScale())
	elementUgs.Draw(path)
}

func (r *svgSaxRenderHandler) handleEllipse(attrs saxAttributes) {
	elementUgs := r.applyStyleAndTransform(attrs, r.ugs)

	cx := svgSaxDoubleAttr(attrs, "cx", 0)
	cy := svgSaxDoubleAttr(attrs, "cy", 0)
	rx := svgSaxDoubleAttr(attrs, "rx", 0)
	ry := svgSaxDoubleAttr(attrs, "ry", 0)

	path := svgSaxEllipsePath(cx, cy, rx, ry)
	path = path.Affine(elementUgs.AffineTransform(), elementUgs.Angle(), elementUgs.InitialScale())
	elementUgs.Draw(path)
}

func (r *svgSaxRenderHandler) handleLine(attrs saxAttributes) {
	elementUgs := r.applyStyleAndTransform(attrs, r.ugs)

	x1 := svgSaxDoubleAttr(attrs, "x1", 0)
	y1 := svgSaxDoubleAttr(attrs, "y1", 0)
	x2 := svgSaxDoubleAttr(attrs, "x2", 0)
	y2 := svgSaxDoubleAttr(attrs, "y2", 0)

	path := UPathNone()
	path.MoveTo(x1, y1)
	path.LineTo(x2, y2)

	path = path.Affine(elementUgs.AffineTransform(), elementUgs.Angle(), elementUgs.InitialScale())
	elementUgs.Draw(path)
}

var (
	svgSaxPTransformOp  = regexp.MustCompile(`(translate|rotate|scale|matrix)[\t\n\x0B\f\r ]*\(([^)]*)\)`)
	svgSaxPPointsNumber = regexp.MustCompile(`[+-]?(?:[0-9]+\.[0-9]*|\.[0-9]+|[0-9]+)(?:[eE][+-]?[0-9]+)?`)
	svgSaxPFontSizeUnit = regexp.MustCompile(`([pP][xXtT]|[eE][mM]|[rR][eE][mM])$`)
)

func (r *svgSaxRenderHandler) handlePolyShape(attrs saxAttributes, closed bool) {
	elementUgs := r.applyStyleAndTransform(attrs, r.ugs)

	pointsStr, _ := attrs.value("points")
	if pointsStr == "" {
		return
	}

	// Per SVG 1.1 §9.7.1, whitespace alone is a valid separator between coordinates.
	// All of these are equivalent: "x,y x,y" / "x y x y" / "x,y,x,y" / mixed.
	// Extract all numeric tokens in order and pair them as (x, y).
	var nums []float64
	for _, tok := range svgSaxPPointsNumber.FindAllString(pointsStr, -1) {
		if v, ok := javaTryParseDouble(tok); ok {
			nums = append(nums, v)
		}
	}

	comment := "polyline"
	if closed {
		comment = "polygon"
	}
	path := NewUPath(comment, "")

	first := true
	var firstX, firstY float64
	for i := 0; i+1 < len(nums); i += 2 {
		x := nums[i]
		y := nums[i+1]
		if first {
			path.MoveTo(x, y)
			firstX = x
			firstY = y
			first = false
		} else {
			path.LineTo(x, y)
		}
	}

	if closed {
		// NOTE: Keep polygons open to avoid SEG_CLOSE dependency in klimt.
		if !first {
			path.LineTo(firstX, firstY)
		}
	}

	path = path.Affine(elementUgs.AffineTransform(), elementUgs.Angle(), elementUgs.InitialScale())
	elementUgs.Draw(path)
}

func svgSaxEllipsePath(cx, cy, rx, ry float64) *UPath {
	path := UPathNone()
	path.MoveTo(0, ry)
	path.ArcTo(rx, ry, 0, 0, 1, rx, 0)
	path.ArcTo(rx, ry, 0, 0, 1, 2*rx, ry)
	path.ArcTo(rx, ry, 0, 0, 1, rx, 2*ry)
	path.ArcTo(rx, ry, 0, 0, 1, 0, ry)
	path.LineTo(0, ry)

	return path.Translate(cx-rx, cy-ry)
}

func (r *svgSaxRenderHandler) handlePath(attrs saxAttributes) {
	elementUgs := r.applyStyleAndTransform(attrs, r.ugs)

	if pathData, _ := attrs.value("d"); pathData != "" {
		svgPath := NewSvgPath(pathData, UTranslateNone())
		svgPath.DrawMe(elementUgs.Ug(), elementUgs.AffineTransform())
	}
}

func (r *svgSaxRenderHandler) handleText(attrs saxAttributes, content string) {
	elementUgs := r.applyStyleAndTransform(attrs, r.ugs)
	fontFamily := svgSaxAttrOrStyle(attrs, "font-family", "font-family")
	fontSize := svgSaxAttrOrStyle(attrs, "font-size", "font-size")
	fontWeight := svgSaxAttrOrStyle(attrs, "font-weight", "font-weight")
	fontStyle := svgSaxAttrOrStyle(attrs, "font-style", "font-style")
	textDecoration := svgSaxAttrOrStyle(attrs, "text-decoration", "text-decoration")
	fillString := svgSaxAttrOrStyle(attrs, "fill", "fill")
	textAnchor, _ := attrs.value("text-anchor")
	xLocation, _ := attrs.value("x")
	yLocation, _ := attrs.value("y")

	textContent := javaTrim(content)
	fontSizeValue := svgSaxParseFontSize(fontSize, 12)
	face := svgSaxParseFontFace(fontWeight, fontStyle)

	// Use default font family if not specified
	if fontFamily == "" {
		fontFamily = "SansSerif"
	}

	f := font.Build(fontFamily, face, fontSizeValue)

	textColor := elementUgs.DefaultColor()
	if fillString != "" && fillString != "none" {
		textColor = elementUgs.TrueColor(fillString)
	}

	fontConfig := NewFontConfigurationTab(f, textColor, textColor, strokePtr(UStrokeSimple()), 8)
	fontConfig = svgSaxApplyTextDecoration(fontConfig, textDecoration)

	utext := UTextBuild(textContent, fontConfig)

	x := 0.0
	if xLocation != "" {
		x = svgParseDouble(xLocation)
	}
	y := 0.0
	if yLocation != "" {
		y = svgParseDouble(yLocation)
	}
	anchorShift := 0.0
	if textAnchor != "" {
		textWidth := utext.CalculateDimension(elementUgs.Ug().StringBounder()).Width()
		if javaEqualsIgnoreCase("middle", textAnchor) {
			anchorShift = -textWidth / 2.0
		} else if javaEqualsIgnoreCase("end", textAnchor) {
			anchorShift = -textWidth
		}
	}

	textTranslate := NewUTranslate(x+anchorShift, y)
	elementUgs.Apply(textTranslate).Draw(utext)
}

func (r *svgSaxRenderHandler) handleImage(attrs saxAttributes) {
	href := svgSaxHref(attrs)
	if href == "" {
		return
	}

	image, unknownFormat := svgSaxDecodeDataImage(href)
	if image == nil && !unknownFormat {
		return
	}

	elementUgs := r.applyStyleAndTransform(attrs, r.ugs)
	x := svgSaxDoubleAttr(attrs, "x", 0)
	y := svgSaxDoubleAttr(attrs, "y", 0)

	_, hasWidthAttr := attrs.value("width")
	_, hasHeightAttr := attrs.value("height")
	if unknownFormat {
		// Upstream reads the width of the null image that ImageIO returns
		// for an unknown format and throws NullPointerException, which fails
		// the whole sprite; the image is skipped instead, as data that does
		// not decode is (design D16).
		return
	}
	width := svgSaxDoubleAttr(attrs, "width", float64(image.Width()))
	height := svgSaxDoubleAttr(attrs, "height", float64(image.Height()))

	if width <= 0 || height <= 0 {
		return
	}

	scale := svgSaxResolveImageScale(hasWidthAttr, hasHeightAttr, width, height, image.Width(), image.Height())
	if scale <= 0 {
		return
	}

	if elementUgs.Ug().MatchesProperty("SVG") {
		svgImage := svgSaxBuildSvgImage(href, width, height)
		imageSvg := NewUImageSvg(svgImage, 1.0)
		elementUgs.Apply(NewUTranslate(x, y)).Draw(imageSvg)
		return
	}

	uimage := NewUImage(NewPixelImage(image, AffineTransformBilinear))
	if scale != 1.0 {
		scaled := uimage.Image(scale)
		uimage = NewUImage(NewPixelImage(scaled, AffineTransformBilinear))
	}

	elementUgs.Apply(NewUTranslate(x, y)).Draw(uimage)
}

func svgSaxHref(attrs saxAttributes) string {
	href, _ := attrs.value("href")
	if href == "" {
		href, _ = attrs.value("xlink:href")
	}
	return href
}

func svgSaxResolveImageScale(hasWidthAttr, hasHeightAttr bool, width, height float64, imageWidth,
	imageHeight int) float64 {
	if imageWidth <= 0 || imageHeight <= 0 {
		return 1.0
	}

	scaleX := width / float64(imageWidth)
	scaleY := height / float64(imageHeight)

	if !hasWidthAttr && !hasHeightAttr {
		return 1.0
	}

	if math.Abs(scaleX-scaleY) < 0.0001 {
		return scaleX
	}

	if hasWidthAttr {
		return scaleX
	}

	return scaleY
}

// svgSaxDecodeDataImage returns the image of a PNG or JPEG data URL, nil
// for any other reference or data that does not decode. unknownFormat is
// true for base64 data of no known image format, which upstream does not
// detect (it fails when it reads the size).
func svgSaxDecodeDataImage(href string) (image PortableImage, unknownFormat bool) {
	lowerHref := javaLower(href)
	if strings.HasPrefix(lowerHref, "data:image/svg+xml;base64,") {
		// Skipping embedded SVG image data URI
		return nil, false
	}

	prefix := svgSaxDataImagePrefix(lowerHref)
	if prefix == "" {
		// Skipping non-data image href
		return nil, false
	}

	data := href[len(prefix):]
	bytes, err := base64Decode(data)
	if err != nil {
		return nil, false
	}

	image, err = readRasterImage(bytes)
	if err != nil {
		return nil, false
	}
	return image, image == nil
}

func svgSaxDataImagePrefix(lowerHref string) string {
	for _, prefix := range []string{"data:image/png;base64,", "data:image/jpeg;base64,", "data:image/jpg;base64,"} {
		if strings.HasPrefix(lowerHref, prefix) {
			return prefix
		}
	}
	return ""
}

func svgSaxBuildSvgImage(href string, width, height float64) string {
	safeHref := strings.ReplaceAll(strings.ReplaceAll(href, "&", "&amp;"), "\"", "&quot;")
	return "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" " +
		"width=\"" + svgSaxFormatNumber(width) + "\" height=\"" + svgSaxFormatNumber(height) + "\">" +
		"<image x=\"0\" y=\"0\" width=\"" + svgSaxFormatNumber(width) + "\" height=\"" + svgSaxFormatNumber(height) +
		"\" xlink:href=\"" + safeHref + "\"/>" + "</svg>"
}

// svgSaxFormatNumber writes an integral value without decimals, others as
// Java's Double.toString.
func svgSaxFormatNumber(value float64) string {
	if value == jmath.Rint(value) {
		return strconv.FormatInt(jmath.Round(value), 10)
	}
	return jstr(value)
}

func (r *svgSaxRenderHandler) handleUse(attrs saxAttributes) {
	href := svgSaxHref(attrs)
	if href == "" || !strings.HasPrefix(href, "#") {
		return
	}

	refID := href[1:]
	referenced := r.definitions[refID]
	if referenced == nil {
		// Referenced element not found
		return
	}

	// Apply positioning from <use> element
	x := svgSaxDoubleAttr(attrs, "x", 0)
	y := svgSaxDoubleAttr(attrs, "y", 0)
	width := svgSaxDoubleAttr(attrs, "width", -1)
	height := svgSaxDoubleAttr(attrs, "height", -1)

	elementUgs := r.applyStyleAndTransform(attrs, r.ugs)

	if x != 0 || y != 0 {
		scalex := elementUgs.AffineTransform().ScaleX()
		scaley := elementUgs.AffineTransform().ScaleY()
		translate := NewUTranslate(x*scalex, y*scaley)
		elementUgs = elementUgs.Apply(translate)
	}

	r.renderUseReference(referenced, elementUgs, width, height)
}

func (r *svgSaxRenderHandler) renderUseReference(referenced *svgSaxBufferedElement,
	elementUgs *UGraphicWithScale, width, height float64) {
	if referenced.xmlContent == "" {
		// Referenced element has no buffered content
		return
	}

	if referenced.tagName == "symbol" {
		elementUgs = r.applySymbolViewport(referenced, elementUgs, width, height)
	}

	wrapped := "<svg xmlns=\"http://www.w3.org/2000/svg\">" + referenced.xmlContent + "</svg>"
	r.parseSvgFragment(wrapped, elementUgs)
}

func (r *svgSaxRenderHandler) applySymbolViewport(referenced *svgSaxBufferedElement,
	elementUgs *UGraphicWithScale, width, height float64) *UGraphicWithScale {
	viewBox := referenced.attributes["viewBox"]
	if viewBox == "" {
		return elementUgs
	}

	values := svgSaxParseNumberList(viewBox)
	if len(values) < 4 {
		return elementUgs
	}

	minX := values[0]
	minY := values[1]
	vbWidth := values[2]
	vbHeight := values[3]

	if vbWidth == 0 || vbHeight == 0 {
		return elementUgs
	}

	scaleX := 1.0
	if width > 0 {
		scaleX = width / vbWidth
	}
	scaleY := 1.0
	if height > 0 {
		scaleY = height / vbHeight
	}

	elementUgs = elementUgs.ApplyTranslate(-minX, -minY)
	elementUgs = elementUgs.ApplyScale(scaleX, scaleY)
	return elementUgs
}

// parseSvgFragment draws a referenced definition; an exception stops it
// (what it drew stays), as upstream logs it.
func (r *svgSaxRenderHandler) parseSvgFragment(svgFragment string, elementUgs *UGraphicWithScale) {
	if r.useDepth >= svgSaxMaxUseDepth {
		throw("java.lang.StackOverflowError", "")
	}
	_ = svgSaxCatch(func() error {
		renderer := newSvgSaxRenderHandler(elementUgs, r.colorResolver, r.definitions, r.useDepth+1)
		return saxParse(svgFragment, renderer)
	})
}

// ---- Style and Transform Helpers ----

// extractGradientColor returns the color of a gradient definition: a
// linear gradient of its stops (its first stop color with a single stop),
// the first stop color of a radial gradient, nil when there is none.
func (r *svgSaxRenderHandler) extractGradientColor(gradientID string) color.HColor {
	gradientElement := r.definitions[gradientID]
	if gradientElement == nil {
		// Gradient not found
		return nil
	}

	tagName := gradientElement.tagName

	if tagName == "linearGradient" {
		stops := r.buildGradientStops(gradientElement.stops)
		if len(stops) >= 2 {
			userSpaceOnUse := gradientElement.attributes["gradientUnits"] == "userSpaceOnUse"
			spreadMethod := svgSaxParseSpreadMethod(gradientElement.attributes["spreadMethod"])
			x1 := svgSaxParsePercentOrNumber(gradientElement.attributes["x1"], 0.0)
			y1 := svgSaxParsePercentOrNumber(gradientElement.attributes["y1"], 0.0)
			x2 := svgSaxParsePercentOrNumber(gradientElement.attributes["x2"], 1.0)
			y2 := svgSaxParsePercentOrNumber(gradientElement.attributes["y2"], 0.0)
			return color.NewHColorLinearGradient(x1, y1, x2, y2, userSpaceOnUse, spreadMethod, stops)
		}
		if len(gradientElement.stops) > 0 {
			if fallback := r.stopColor(&gradientElement.stops[0]); fallback != nil {
				return fallback
			}
		}
	}

	if tagName == "radialGradient" {
		if len(gradientElement.stops) > 0 {
			if fallback := r.stopColor(&gradientElement.stops[0]); fallback != nil {
				return fallback
			}
		}
	}

	return nil
}

func (r *svgSaxRenderHandler) buildGradientStops(stops []svgSaxGradientStop) []color.Stop {
	var result []color.Stop
	for i := range stops {
		stop := &stops[i]
		c := r.stopColor(stop)
		if c == nil {
			continue
		}

		offset := svgSaxParseOffset(stop.offset)
		opacity := svgSaxParseOpacity(stop.opacity)
		result = append(result, color.NewStop(offset, c, opacity))
	}
	return result
}

func (r *svgSaxRenderHandler) stopColor(stop *svgSaxGradientStop) color.HColor {
	if stop == nil || stop.color == "" || javaEqualsIgnoreCase("none", stop.color) {
		return nil
	}
	return r.ugs.TrueColor(stop.color)
}

func svgSaxParseOffset(value string) float64 {
	if value == "" {
		return 0.0
	}
	if number, ok := strings.CutSuffix(value, "%"); ok {
		v, ok := javaTryParseDouble(number)
		if !ok {
			return 0.0
		}
		return svgSaxClamp01(v / 100.0)
	}
	v, ok := javaTryParseDouble(value)
	if !ok {
		return 0.0
	}
	return svgSaxClamp01(v)
}

func svgSaxParseOpacity(value string) float64 {
	if value == "" {
		return 1.0
	}
	v, ok := javaTryParseDouble(value)
	if !ok {
		return 1.0
	}
	return svgSaxClamp01(v)
}

func svgSaxClamp01(value float64) float64 {
	if value < 0.0 {
		return 0.0
	}
	if value > 1.0 {
		return 1.0
	}
	return value
}

func svgSaxParseSpreadMethod(value string) color.SpreadMethod {
	switch javaLower(value) {
	case "reflect":
		return color.SpreadReflect
	case "repeat":
		return color.SpreadRepeat
	}
	return color.SpreadPad
}

// svgSaxParsePercentOrNumber parses a percentage ("50%") as a fraction, or
// a number; defaultValue when it is missing or does not parse.
func svgSaxParsePercentOrNumber(value string, defaultValue float64) float64 {
	if value == "" {
		return defaultValue
	}
	if number, ok := strings.CutSuffix(value, "%"); ok {
		v, ok := javaTryParseDouble(number)
		if !ok {
			return defaultValue
		}
		return v / 100.0
	}
	v, ok := javaTryParseDouble(value)
	if !ok {
		return defaultValue
	}
	return v
}

func (r *svgSaxRenderHandler) applyStyleAndTransform(attrs saxAttributes, ugs *UGraphicWithScale) *UGraphicWithScale {
	if transform, _ := attrs.value("transform"); transform != "" {
		ugs = r.applyTransform(ugs, transform)
	}

	fill := svgSaxAttrOrStyle(attrs, "fill", "fill")
	stroke := svgSaxAttrOrStyle(attrs, "stroke", "stroke")
	strokeWidth := svgSaxAttrOrStyle(attrs, "stroke-width", "stroke-width")

	if strokeWidth != "" {
		if v, ok := javaTryParseDouble(strokeWidth); ok {
			scale := ugs.InitialScale()
			ugs = ugs.Apply(UStrokeWithThickness(scale * v))
		}
	}

	if stroke != "" {
		sc := ugs.TrueColor(stroke)
		ugs = ugs.Apply(sc)
		if fill == "" {
			ugs = ugs.Apply(Bg(ugs.DefaultColor()))
		}
	}

	if fill != "" {
		switch {
		case fill == "none":
			ugs = ugs.Apply(Bg(color.None()))
		case strings.HasPrefix(fill, "url(#"):
			// Extract gradient reference: url(#Gradient) -> Gradient
			gradientID := svgSaxSubstring(fill, 5, javaLen(fill)-1)
			if gradientColor := r.extractGradientColor(gradientID); gradientColor != nil {
				// If no stroke specified, also set as foreground
				if stroke == "" {
					ugs = ugs.Apply(gradientColor)
				}
				// Always set as background/fill
				ugs = ugs.Apply(Bg(gradientColor))
			}
		default:
			fc := ugs.TrueColor(fill)
			if stroke == "" {
				ugs = ugs.Apply(fc)
			}
			ugs = ugs.Apply(Bg(fc))
		}
	}

	return ugs
}

// svgSaxSubstring is Java's s.substring(begin, end) in UTF-16 units; it
// panics as Java throws StringIndexOutOfBoundsException.
func svgSaxSubstring(s string, begin, end int) string {
	u := utf16Of(s)
	if begin < 0 || end > len(u) || begin > end {
		throw("java.lang.StringIndexOutOfBoundsException", "begin "+itoa(begin)+", end "+itoa(end)+", length "+
			itoa(len(u)))
	}
	return string(utf16.Decode(u[begin:end]))
}

// svgSaxAttrOrStyle returns the property of the style attribute when it is
// set and not empty, else the attribute ("" when missing).
func svgSaxAttrOrStyle(attrs saxAttributes, attrName, styleKey string) string {
	if styleValue, ok := svgSaxAttrStyleValue(attrs, styleKey); ok && styleValue != "" {
		return styleValue
	}
	v, _ := attrs.value(attrName)
	return v
}

func svgSaxAttrStyleValue(attrs saxAttributes, key string) (string, bool) {
	style, _ := attrs.value("style")
	if style == "" {
		return "", false
	}
	return svgSaxStyleValue(style, key)
}

// ---- Transform Parsing (Shared DOM/SAX helpers) ----

func (r *svgSaxRenderHandler) applyTransform(ugs *UGraphicWithScale, transform string) *UGraphicWithScale {
	if transform == "" {
		return ugs
	}

	for _, m := range svgSaxPTransformOp.FindAllStringSubmatch(transform, -1) {
		op := m[1]
		values := svgSaxParseNumberList(m[2])

		switch op {
		case "translate":
			tx := 0.0
			if len(values) > 0 {
				tx = values[0]
			}
			ty := 0.0
			if len(values) > 1 {
				ty = values[1]
			}
			ugs = ugs.ApplyTranslate(tx, ty)
		case "scale":
			sx := 1.0
			if len(values) > 0 {
				sx = values[0]
			}
			sy := sx
			if len(values) > 1 {
				sy = values[1]
			}
			if sx == sy {
				ugs = ugs.ApplyScale(sx, sy)
			} else {
				ugs = ugs.ApplyMatrix(sx, 0, 0, sy, 0, 0)
			}
		case "rotate":
			angle := 0.0
			if len(values) > 0 {
				angle = values[0]
			}
			cx, cy := 0.0, 0.0
			if len(values) > 2 {
				cx = values[1]
				cy = values[2]
			}
			ugs = ugs.ApplyRotate(angle, cx, cy)
		case "matrix":
			if len(values) >= 6 {
				ugs = ugs.ApplyMatrix(values[0], values[1], values[2], values[3], values[4], values[5])
			}
		}
	}

	return ugs
}

// svgSaxParseNumberList parses the numbers of a list separated by commas and
// whitespace; parts that are not numbers are skipped.
func svgSaxParseNumberList(raw string) []float64 {
	raw = javaTrim(raw)
	if raw == "" {
		return nil
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || isJavaWhitespaceRegex(r) })
	values := make([]float64, 0, len(parts))
	for _, part := range parts {
		if v, ok := javaTryParseDouble(part); ok {
			values = append(values, v)
		}
	}
	return values
}

// ---- Font/Text Helpers (Shared DOM/SAX helpers) ----

func svgSaxParseFontSize(fontSizeStr string, defaultSize int) int {
	if fontSizeStr == "" {
		return defaultSize
	}
	cleaned := svgSaxPFontSizeUnit.ReplaceAllString(javaTrim(fontSizeStr), "")
	v, err := strconv.ParseInt(cleaned, 10, 32)
	if err != nil {
		return defaultSize
	}
	return int(v)
}

// svgSaxParseFontFace returns the face of the CSS font-weight (100 to 900,
// keywords) and font-style (italic or oblique) values.
func svgSaxParseFontFace(fontWeight, fontStyle string) font.UFontFace {
	face := font.NormalFace()

	if fontWeight != "" {
		if wf, ok := font.UFontFaceFromCssWeight(javaTrim(fontWeight)); ok {
			face = face.WithWeight(wf.CssWeight())
		}
	}

	if fontStyle != "" {
		s := javaTrim(fontStyle)
		if javaEqualsIgnoreCase("italic", s) || javaEqualsIgnoreCase("oblique", s) {
			face = face.WithStyle(font.StyleItalic)
		}
	}

	return face
}

func svgSaxApplyTextDecoration(fontConfig *FontConfiguration, textDecoration string) *FontConfiguration {
	if textDecoration == "" || javaEqualsIgnoreCase("none", textDecoration) {
		return fontConfig
	}

	if strings.Contains(textDecoration, "underline") {
		fontConfig = fontConfig.Add(font.FontStyleUnderline)
	}

	if strings.Contains(textDecoration, "line-through") {
		fontConfig = fontConfig.Add(font.FontStyleStrike)
	}

	return fontConfig
}

// ---- Utility Methods ----

// svgSaxDoubleAttr returns the number of an attribute, defaultValue when it
// is missing, empty or not a number.
func svgSaxDoubleAttr(attrs saxAttributes, name string, defaultValue float64) float64 {
	value, _ := attrs.value(name)
	if value == "" {
		return defaultValue
	}
	v, ok := javaTryParseDouble(value)
	if !ok {
		return defaultValue
	}
	return v
}
