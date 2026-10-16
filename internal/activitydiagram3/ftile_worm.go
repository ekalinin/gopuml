// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/Worm.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/WormMutation.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"math"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// directionNull is a null direction (upstream's Direction.fromVector of a
// null vector).
const directionNull klimt.Direction = -1

// directionOf is Direction.fromVector with directionNull for a null vector.
func directionOf(p1, p2 geom.XPoint2D) klimt.Direction {
	d, ok := klimt.DirectionFromVector(p1, p2)
	if !ok {
		return directionNull
	}
	return d
}

// directionInv is Direction.getInv; a null direction panics, as upstream
// dereferences null.
func directionInv(d klimt.Direction) klimt.Direction {
	if d == directionNull {
		uml.ThrowNPE("")
	}
	return d.Inv()
}

// Worm is the polyline of a snake. A moved worm shares the points of the
// original until either changes them.
type Worm struct {
	points       []geom.XPoint2D
	sharedPoints bool
	style        *klimt.Style
	arrows       klimt.Arrows
	tr           *klimt.UTranslate

	ignoreForCompression bool
}

// NewWorm returns an empty worm.
func NewWorm(style *klimt.Style, arrows klimt.Arrows) *Worm {
	return newWormSize(style, arrows, 10)
}

func newWormSize(style *klimt.Style, arrows klimt.Arrows, initSize int) *Worm {
	return &Worm{style: style, arrows: arrows, points: make([]geom.XPoint2D, 0, max(initSize, 10))}
}

// Move returns the worm moved by (dx, dy), sharing the points.
func (w *Worm) Move(dx, dy float64) *Worm {
	result := w.CloneEmpty()
	if w.tr == nil {
		t := klimt.NewUTranslate(dx, dy)
		result.tr = &t
	} else {
		t := w.tr.Compose(klimt.NewUTranslate(dx, dy))
		result.tr = &t
	}
	result.points = w.points
	result.sharedPoints = true
	w.sharedPoints = true
	return result
}

func (w *Worm) ensureOwnPoints() {
	if w.sharedPoints {
		w.points = append(make([]geom.XPoint2D, 0, max(len(w.points), 10)), w.points...)
		w.sharedPoints = false
	}
}

// IsPureHorizontal reports whether the worm is a single horizontal segment.
func (w *Worm) IsPureHorizontal() bool {
	return w.Size() == 2 && w.Point(0).Y == w.Point(1).Y
}

// CloneEmpty returns an empty worm of the same style.
func (w *Worm) CloneEmpty() *Worm {
	result := newWormSize(w.style, w.arrows, w.Size())
	result.ignoreForCompression = w.ignoreForCompression
	return result
}

// SetIgnoreForCompression marks the worm as ignored by the compression on
// x; it panics once points are added, as upstream.
func (w *Worm) SetIgnoreForCompression() {
	if w.Size() > 0 {
		uml.Throw(uml.IllegalStateException, "")
	}
	w.ignoreForCompression = true
}

func (w *Worm) drawInternalOneColor(startDecoration *klimt.UPolygon, ug klimt.UGraphic,
	colorAndStyle *HtmlColorAndStyle, strokeValue float64, emphasizeDirection klimt.Direction, hasEmphasize bool,
	endDecoration *klimt.UPolygon) {
	arrowColor := colorAndStyle.ArrowColor()
	if arrowColor == nil {
		uml.ThrowNPE("")
	}
	linkStyle := colorAndStyle.Style()
	if linkStyle.IsInvisible() {
		return
	}
	ug = ug.Apply(arrowColor)
	ug = ug.Apply(klimt.Bg(arrowColor))
	if linkStyle.IsNormal() {
		ug = ug.Apply(w.style.Stroke())
	} else {
		ug = ug.Apply(linkStyle.GoThickness(strokeValue).Stroke3())
	}
	drawn := false
	for i := 0; i < w.Size()-1; i++ {
		p1 := w.Point(i)
		p2 := w.Point(i + 1)
		line := geom.XLine2DLine(p1, p2)
		if !drawn && hasEmphasize && directionOf(p1, p2) == emphasizeDirection {
			w.drawLine(ug, line, emphasizeDirection, true)
			drawn = true
		} else {
			w.drawLine(ug, line, 0, false)
		}
	}
	arrowHeadColor := colorAndStyle.ArrowHeadColor()
	if arrowHeadColor == nil {
		uml.Throw(uml.IllegalStateException, "")
	}
	ug = ug.Apply(arrowHeadColor)
	ug = ug.Apply(klimt.Bg(arrowHeadColor))
	onX := klimt.CompressionOnX
	if startDecoration != nil {
		ug = ug.Apply(klimt.UStrokeWithThickness(1.5))
		start := w.Point(0)
		if w.ignoreForCompression {
			startDecoration.SetCompressionMode(&onX)
		}
		ug.Apply(klimt.UTranslatePoint(start)).Apply(klimt.UStrokeSimple()).Draw(startDecoration)
	}
	if endDecoration != nil {
		ug = ug.Apply(klimt.UStrokeWithThickness(1.5))
		end := w.Point(w.Size() - 1)
		if w.ignoreForCompression {
			endDecoration.SetCompressionMode(&onX)
		}
		ug.Apply(klimt.UTranslatePoint(end)).Apply(klimt.UStrokeSimple()).Draw(endDecoration)
	}
}

func (w *Worm) drawLine(ug klimt.UGraphic, line geom.XLine2D, direction klimt.Direction, hasDirection bool) {
	x1, y1, x2, y2 := line.X1, line.Y1, line.X2, line.Y2
	ug = ug.Apply(klimt.NewUTranslate(x1, y1))
	if hasDirection {
		ug.Apply(klimt.NewUTranslate((x2-x1)/2, (y2-y1)/2)).Draw(klimt.ArrowsAsTo(w.arrows, direction))
	}
	ug.Draw(klimt.NewULine(x2-x1, y2-y1))
}

// MoveFirstPoint returns the worm with its first point moved horizontally
// or vertically, the second point following it on the same axis.
func (w *Worm) MoveFirstPoint(move klimt.UTranslate) *Worm {
	dx := move.Dx()
	dy := move.Dy()
	if dx != 0 && dy != 0 {
		uml.Throw(uml.IllegalArgumentException, "move="+move.String())
	}
	result := newWormSize(w.style, w.arrows, w.Size())
	x0 := w.Point(0).X
	y0 := w.Point(0).Y
	x1 := w.Point(1).X
	y1 := w.Point(1).Y
	if dx != 0 && x0 == x1 {
		x1 += dx
	}
	if dy != 0 && y0 == y1 {
		y1 += dy
	}
	x0 += dx
	y0 += dy
	result.AddPoint(x0, y0)
	result.AddPoint(x1, y1)
	for i := 2; i < w.Size(); i++ {
		result.AddPointPt(w.Point(i))
	}
	return result
}

// MoveLastPoint returns the worm with its last point moved horizontally or
// vertically, the point before following it on the same axis.
func (w *Worm) MoveLastPoint(move klimt.UTranslate) *Worm {
	dx := move.Dx()
	dy := move.Dy()
	if dx != 0 && dy != 0 {
		uml.Throw(uml.IllegalArgumentException, "move="+move.String())
	}
	result := newWormSize(w.style, w.arrows, w.Size())
	x8 := w.Point(w.Size() - 2).X
	y8 := w.Point(w.Size() - 2).Y
	x9 := w.Point(w.Size() - 1).X
	y9 := w.Point(w.Size() - 1).Y
	if dx != 0 && x8 == x9 {
		x8 += dx
	}
	if dy != 0 && y8 == y9 {
		y8 += dy
	}
	x9 += dx
	y9 += dy
	for i := 0; i < w.Size()-2; i++ {
		result.AddPointPt(w.Point(i))
	}
	result.AddPoint(x8, y8)
	result.AddPoint(x9, y9)
	return result
}

// AddPoint appends a point, unless it repeats the last one.
func (w *Worm) AddPoint(x, y float64) {
	if w.tr != nil {
		x -= w.tr.Dx()
		y -= w.tr.Dy()
	}
	if math.IsNaN(x) || math.IsNaN(y) {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	if w.Size() > 0 {
		last := w.Last()
		if last.X == x && last.Y == y {
			return
		}
	}
	w.ensureOwnPoints()
	w.points = append(w.points, geom.NewXPoint2D(x, y))
}

// AddPointPt appends a point, unless it repeats the last one.
func (w *Worm) AddPointPt(pt geom.XPoint2D) { w.AddPoint(pt.X, pt.Y) }

// Direction returns the direction of the first segment.
func (w *Worm) Direction() SnakeDirection {
	if w.Size() < 2 {
		uml.Throw(uml.IllegalStateException, "")
	}
	return SnakeDirectionOf(w.Point(0), w.Point(1))
}

// DirectionsCode returns the first letters of the directions of the
// segments, for example "DRD".
func (w *Worm) DirectionsCode() string {
	var result strings.Builder
	for i := 0; i < w.Size()-1; i++ {
		direction := directionOf(w.Point(i), w.Point(i+1))
		if direction == directionNull {
			uml.ThrowNPE("")
		}
		result.WriteString(direction.ShortCode())
	}
	return result.String()
}

func (w *Worm) patternAt(i int) [4]klimt.Direction {
	return [4]klimt.Direction{w.directionAtPoint(i), w.directionAtPoint(i + 1), w.directionAtPoint(i + 2),
		w.directionAtPoint(i + 3)}
}

func (w *Worm) isForwardAndBackwardAt(i int) bool {
	return w.directionAtPoint(i) == directionInv(w.directionAtPoint(i+1))
}

func (w *Worm) directionAtPoint(i int) klimt.Direction { return directionOf(w.Point(i), w.Point(i+1)) }

// DoesHorizontalCross reports whether a horizontal segment crosses area.
func (w *Worm) DoesHorizontalCross(area geom.MinMax) bool {
	for i := 0; i < w.Size()-1; i++ {
		pt1 := w.Point(i)
		pt2 := w.Point(i + 1)
		if pt1.Y == pt2.Y && area.DoesHorizontalCross(pt1, pt2) {
			return true
		}
	}
	return false
}

// Size returns the number of points.
func (w *Worm) Size() int { return len(w.points) }

// Point returns point i.
func (w *Worm) Point(i int) geom.XPoint2D {
	if i < 0 || i >= len(w.points) {
		uml.ThrowIndexOutOfBounds(i, len(w.points))
	}
	return w.resolve(w.points[i])
}

func (w *Worm) resolve(point geom.XPoint2D) geom.XPoint2D {
	if w.tr == nil {
		return point
	}
	return w.tr.Translated(point)
}

// First returns the first point.
func (w *Worm) First() geom.XPoint2D { return w.Point(0) }

// Last returns the last point.
func (w *Worm) Last() geom.XPoint2D { return w.Point(w.Size() - 1) }

// MinX returns the smallest abscissa.
func (w *Worm) MinX() float64 {
	result := w.Point(0).X
	for _, pt := range w.points {
		result = math.Min(result, w.resolve(pt).X)
	}
	return result
}

// MaxX returns the largest abscissa.
func (w *Worm) MaxX() float64 {
	result := w.Point(0).X
	for _, pt := range w.points {
		result = math.Max(result, w.resolve(pt).X)
	}
	return result
}

// MaxY returns the largest ordinate.
func (w *Worm) MaxY() float64 {
	result := w.Point(0).Y
	for _, pt := range w.points {
		result = math.Max(result, w.resolve(pt).Y)
	}
	return result
}

// Merge returns the worm joining w and other, simplified.
func (w *Worm) Merge(other *Worm, merge MergeStrategy) *Worm {
	if !snakeSame(w.Last(), other.First()) {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	result := newWormSize(w.style, w.arrows, w.Size()+other.Size())
	for _, pt := range w.points {
		result.AddPointPt(w.resolve(pt))
	}
	for _, pt := range other.points {
		result.AddPointPt(other.resolve(pt))
	}
	result.mergeMe(merge)
	return result
}

func (w *Worm) mergeMe(merge MergeStrategy) {
	if w.tr != nil {
		uml.Throw(uml.UnsupportedOperationException, "")
	}
	w.ensureOwnPoints()
	for {
		change := w.removeNullVector() || w.removeRedondantDirection() || w.removePattern1() ||
			w.removePattern2() || w.removePattern3() || w.removePattern4() || w.removePattern5() ||
			w.removePattern6() || w.removePattern7()
		if !change && merge == MergeFull {
			change = w.removePattern8()
		}
		if !change {
			return
		}
	}
}

func (w *Worm) removeAt(i int) { w.points = append(w.points[:i], w.points[i+1:]...) }

// replace3 replaces points i+1, i+2 and i+3 by newPoint.
func (w *Worm) replace3(i int, newPoint geom.XPoint2D) {
	w.removeAt(i + 3)
	w.removeAt(i + 2)
	w.points[i+1] = newPoint
}

func (w *Worm) removeNullVector() bool {
	for i := 0; i < w.Size()-1; i++ {
		if w.directionAtPoint(i) == directionNull {
			w.removeAt(i)
			return true
		}
	}
	return false
}

func (w *Worm) removeRedondantDirection() bool {
	for i := 0; i < w.Size()-2; i++ {
		if w.directionAtPoint(i) == w.directionAtPoint(i+1) {
			w.removeAt(i + 1)
			return true
		}
	}
	return false
}

var (
	patternDLDR = [4]klimt.Direction{klimt.DirectionDown, klimt.DirectionLeft, klimt.DirectionDown, klimt.DirectionRight}
	patternDRDL = [4]klimt.Direction{klimt.DirectionDown, klimt.DirectionRight, klimt.DirectionDown, klimt.DirectionLeft}
	patternRDLD = [4]klimt.Direction{klimt.DirectionRight, klimt.DirectionDown, klimt.DirectionLeft, klimt.DirectionDown}
	patternRDRU = [4]klimt.Direction{klimt.DirectionRight, klimt.DirectionDown, klimt.DirectionRight, klimt.DirectionUp}
	patternLDLU = [4]klimt.Direction{klimt.DirectionLeft, klimt.DirectionDown, klimt.DirectionLeft, klimt.DirectionUp}
	patternDRDR = [4]klimt.Direction{klimt.DirectionDown, klimt.DirectionRight, klimt.DirectionDown, klimt.DirectionRight}
	patternDLDL = [4]klimt.Direction{klimt.DirectionDown, klimt.DirectionLeft, klimt.DirectionDown, klimt.DirectionLeft}
	patternLDLD = [4]klimt.Direction{klimt.DirectionLeft, klimt.DirectionDown, klimt.DirectionLeft, klimt.DirectionDown}
	patternRDRD = [4]klimt.Direction{klimt.DirectionRight, klimt.DirectionDown, klimt.DirectionRight, klimt.DirectionDown}
)

func (w *Worm) removePattern1() bool {
	for i := 0; i < w.Size()-5; i++ {
		patternAt := w.patternAt(i)
		if patternAt == patternDLDR || patternAt == patternDRDL {
			w.replace3(i, geom.NewXPoint2D(w.Point(i+1).X, w.Point(i+3).Y))
			return true
		}
	}
	return false
}

func (w *Worm) removePattern7() bool {
	if w.Size() > 4 {
		const i = 0
		if w.patternAt(i) == patternRDLD && w.Point(i+3).X > w.Point(i).X {
			newPoint := geom.NewXPoint2D(w.Point(i+3).X, w.Point(i).Y)
			w.removeAt(i + 2)
			w.points[i+1] = newPoint
			return true
		}
	}
	return false
}

func (w *Worm) removePattern2() bool {
	for i := 0; i < w.Size()-5; i++ {
		patternAt := w.patternAt(i)
		if patternAt == patternRDRU || patternAt == patternLDLU {
			w.replace3(i, geom.NewXPoint2D(w.Point(i+3).X, w.Point(i+1).Y))
			return true
		}
	}
	return false
}

func (w *Worm) removePattern3() bool {
	for i := 0; i < w.Size()-4; i++ {
		patternAt := w.patternAt(i)
		if patternAt == patternDRDR || patternAt == patternDLDL {
			w.replace3(i, geom.NewXPoint2D(w.Point(i+1).X, w.Point(i+3).Y))
			return true
		}
	}
	return false
}

func (w *Worm) removePattern4() bool {
	i := w.Size() - 5
	if i >= 0 && w.patternAt(i) == patternDLDR {
		p1 := w.Point(i + 1)
		p4 := w.Point(i + 4)
		if p4.X > p1.X {
			w.replace3(i, geom.NewXPoint2D(w.Point(i+1).X, w.Point(i+3).Y))
			return true
		}
	}
	return false
}

func (w *Worm) removePattern5() bool {
	i := w.Size() - 5
	if i >= 0 && w.patternAt(i) == patternDRDL {
		p1 := w.Point(i + 1)
		p4 := w.Point(i + 4)
		if p4.X+4 < p1.X {
			w.replace3(i, geom.NewXPoint2D(w.Point(i+1).X, w.Point(i+3).Y))
			return true
		}
	}
	return false
}

func (w *Worm) removePattern6() bool {
	for i := 0; i < w.Size()-2; i++ {
		if w.isForwardAndBackwardAt(i) {
			w.removeAt(i + 1)
			return true
		}
	}
	return false
}

func (w *Worm) removePattern8() bool {
	for i := 0; i < w.Size()-4; i++ {
		patternAt := w.patternAt(i)
		if patternAt == patternLDLD || patternAt == patternRDRD {
			w.replace3(i, geom.NewXPoint2D(w.Point(i+3).X, w.Point(i+1).Y))
			return true
		}
	}
	return false
}

// Style returns the arrow style of the worm.
func (w *Worm) Style() *klimt.Style { return w.style }

// WormMutation is the list of translations that move each point of a worm
// to draw the parallel lines of a rainbow arrow.
type WormMutation struct {
	translations []klimt.UTranslate
}

// WormMutationCreate returns the mutation of worm for a gap of delta.
func WormMutationCreate(worm *Worm, delta float64) *WormMutation {
	signature := worm.DirectionsCode()
	definition, ok := wormMutationDefinition(signature)
	if !ok {
		return wormMutationFromLongSignature(signature, delta)
	}
	return newWormMutation(definition, delta)
}

func wormMutationFromLongSignature(signature string, delta float64) *WormMutation {
	result := &WormMutation{}
	for i := 0; i < len(signature)-1; i++ {
		definition, ok := wormMutationDefinition(signature[i : i+2])
		if !ok {
			uml.ThrowNPE("")
		}
		tmp := newWormMutation(definition, delta)
		if i == 0 {
			result.translations = append(result.translations, tmp.translations[0])
		} else {
			last := result.Last()
			if !last.IsAlmostSame(tmp.translations[0]) {
				tmp = tmp.reverse()
			}
		}
		result.translations = append(result.translations, tmp.translations[1])
		if i == len(signature)-2 {
			result.translations = append(result.translations, tmp.translations[2])
		}
	}
	return result
}

func (m *WormMutation) reverse() *WormMutation {
	result := &WormMutation{}
	for _, tr := range m.translations {
		result.translations = append(result.translations, tr.Reverse())
	}
	return result
}

// Last returns the last translation.
func (m *WormMutation) Last() klimt.UTranslate { return m.translations[len(m.translations)-1] }

// First returns the first translation.
func (m *WormMutation) First() klimt.UTranslate { return m.translations[0] }

// Size returns the number of translations.
func (m *WormMutation) Size() int { return len(m.translations) }

func wormMutationDefinition(signature string) (string, bool) {
	switch signature {
	case "D", "U":
		return "33", true
	case "L", "R":
		return "55", true
	case "RD":
		return "123", true
	case "RU":
		return "543", true
	case "LD":
		return "187", true
	case "DL":
		return "345", true
	case "DR":
		return "765", true
	case "UL":
		return "321", true
	case "UR":
		return "781", true
	}
	return "", false
}

func newWormMutation(definition string, delta float64) *WormMutation {
	result := &WormMutation{}
	for i := range len(definition) {
		result.translations = append(result.translations, wormMutationTranslation(int(definition[i]-'0'), delta))
	}
	return result
}

func wormMutationTranslation(typ int, delta float64) klimt.UTranslate {
	switch typ {
	case 1:
		return klimt.UTranslateDy(-delta)
	case 2:
		return klimt.NewUTranslate(delta, -delta)
	case 3:
		return klimt.UTranslateDx(delta)
	case 4:
		return klimt.NewUTranslate(delta, delta)
	case 5:
		return klimt.UTranslateDy(delta)
	case 6:
		return klimt.NewUTranslate(-delta, delta)
	case 7:
		return klimt.UTranslateDx(-delta)
	case 8:
		return klimt.NewUTranslate(-delta, -delta)
	}
	uml.Throw(uml.IllegalArgumentException, "")
	return klimt.UTranslate{}
}

// TextTranslate returns the move of the labels of a rainbow of size
// colors.
func (m *WormMutation) TextTranslate(size int) klimt.UTranslate {
	// Java's Double.MIN_VALUE, the smallest positive double
	minV, maxV := math.MaxFloat64, 4.9e-324
	for _, tr := range m.translations {
		v := tr.Dx()
		if v > maxV {
			maxV = v
		}
		if v < minV {
			minV = v
		}
	}
	extreme := minV
	if math.Abs(maxV) > math.Abs(minV) {
		extreme = maxV
	}
	return klimt.UTranslateDx(extreme * float64(size-1))
}

// IsDxNegative reports whether the first translation goes left.
func (m *WormMutation) IsDxNegative() bool { return m.translations[0].Dx() < 0 }

// Mute returns original with its points translated.
func (m *WormMutation) Mute(original *Worm) *Worm {
	result := original.CloneEmpty()
	for i := 0; i < original.Size(); i++ {
		result.AddPointPt(m.translations[i].Translated(original.Point(i)))
	}
	return result
}
