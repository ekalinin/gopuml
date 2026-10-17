// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/SvgResult.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/PointListIterator.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/PointListIteratorImpl.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// The attributes of dot's SVG output that hold coordinates.
const (
	svgDEquals      = "d=\""
	svgPointsEquals = "points=\""
)

// SvgResult is upstream's SvgResult: a part of the SVG document written by
// dot, and the conversion of its coordinates to the drawing's (dot's y axis
// points up from the bottom of the graph).
type SvgResult struct {
	svg      string
	function point2DFunction
}

func newSvgResult(svg string, function point2DFunction) SvgResult {
	return SvgResult{svg: svg, function: function}
}

// pointsWithThisColor is getPointsWithThisColor.
func (s SvgResult) pointsWithThisColor(lineColor int32) *pointListIterator {
	return newPointListIterator(s, lineColor)
}

// extractList returns the points of the first attribute searched (d=" or
// points=").
func (s SvgResult) extractList(searched string) []geom.XPoint2D {
	p2 := s.indexOf(searched, 0)
	if p2 == -1 {
		return nil
	}
	p3 := s.indexOf("\"", p2+len(searched))
	if p3 == -1 {
		return nil
	}
	return s.substring2(p2+len(searched), p3).points(" MC")
}

// indexFromColor is getIndexFromColor: the position of the first element
// drawn with the color, -1 when there is none.
func (s SvgResult) indexFromColor(c int32) int {
	hex := strings.ToLower(color.ToHexRGBColor(c))
	if idx := strings.Index(s.svg, "stroke=\""+hex+"\""); idx != -1 {
		return idx
	}
	if idx := strings.Index(s.svg, ";stroke:"+hex+";"); idx != -1 {
		return idx
	}
	if idx := strings.Index(s.svg, "fill=\""+hex+"\""); idx != -1 {
		return idx
	}
	return -1
}

// points returns the points of the text, separated by the characters of
// separator; a number that cannot be read gives no point at all.
func (s SvgResult) points(separator string) (result []geom.XPoint2D) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(interface{ Error() string }); ok && isNumberFormat(e) {
				result = nil
				return
			}
			panic(r)
		}
	}()
	for _, tok := range klimtTokenize(s.svg, separator) {
		result = append(result, s.firstPoint(tok))
	}
	return result
}

func isNumberFormat(e interface{ Error() string }) bool {
	re, ok := e.(*color.RuntimeError)
	return ok && re.Class == uml.NumberFormatException
}

// nextPoint is getNextPoint.
func (s SvgResult) nextPoint() geom.XPoint2D { return s.firstPoint(s.svg) }

func (s SvgResult) firstPoint(tmp string) geom.XPoint2D {
	tokens := klimtTokenize(tmp, ",")
	if len(tokens) < 2 {
		uml.Throw(uml.NoSuchElementException, "")
	}
	startX := javaParseDouble(tokens[0])
	startY := javaParseDouble(tokens[1])
	return s.function.apply(geom.NewXPoint2D(startX, startY))
}

func (s SvgResult) indexOf(substr string, pos int) int { return indexFrom(s.svg, substr, pos) }

func (s SvgResult) substring(pos int) SvgResult {
	return SvgResult{svg: s.svg[pos:], function: s.function}
}

func (s SvgResult) substring2(start, end int) SvgResult {
	return SvgResult{svg: s.svg[start:end], function: s.function}
}

// toDotPath reads the path ("M x,y C x1,y1 x2,y2 x,y ...") as Bézier
// curves.
func (s SvgResult) toDotPath() *klimt.DotPath {
	if !s.isPathConsistent() {
		throwIllegalArgument("")
	}
	posC := s.indexOf("C", 0)
	if posC == -1 {
		throwIllegalArgument("")
	}
	start := s.substring2(1, posC).nextPoint()
	pts := s.substring(posC + 1).points(" ")
	var beziers []geom.XCubicCurve2D
	x, y := start.X, start.Y
	for i := 0; i < len(pts); i += 3 {
		if i+2 >= len(pts) {
			uml.Throw(uml.NoSuchElementException, "")
		}
		p1, p2, p := pts[i], pts[i+1], pts[i+2]
		beziers = append(beziers, geom.NewXCubicCurve2D(x, y, p1.X, p1.Y, p2.X, p2.Y, p.X, p.Y))
		x, y = p.X, p.Y
	}
	return klimt.DotPathFromBeziers(beziers)
}

// isPathConsistent reports whether the text starts with "M".
func (s SvgResult) isPathConsistent() bool { return strings.HasPrefix(s.svg, "M") }

// pointListIterator is upstream's PointListIteratorImpl: the successive
// points="..." lists after the first element of a color.
type pointListIterator struct {
	svg SvgResult
	pos int
}

func newPointListIterator(svg SvgResult, lineColor int32) *pointListIterator {
	result := &pointListIterator{svg: svg}
	if svg.indexFromColor(lineColor) == -1 {
		result.pos = -1
	}
	return result
}

func (it *pointListIterator) cloneMe() *pointListIterator {
	return &pointListIterator{svg: it.svg, pos: it.pos}
}

func (it *pointListIterator) hasNext() bool { return it.pos != -2 }

func (it *pointListIterator) next() []geom.XPoint2D {
	if it.pos < 0 {
		it.pos = -2
		return nil
	}
	if it.pos > len(it.svg.svg) {
		// StringIndexOutOfBoundsException, logged by upstream.
		return nil
	}
	result := it.svg.substring(it.pos).extractList(svgPointsEquals)
	if len(result) == 0 {
		it.pos = -2
	} else {
		it.pos = it.svg.indexOf(svgPointsEquals, it.pos) + len(svgPointsEquals) + 1
	}
	return result
}

// klimtTokenize is java.util.StringTokenizer without the delimiters.
func klimtTokenize(s, delims string) []string {
	return strings.FieldsFunc(s, func(r rune) bool { return strings.ContainsRune(delims, r) })
}
