// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/compress/CompressionXorYBuilder.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/compress/UGraphicCompressOnXorY.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/compress/SlotFinder.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/compress/SlotSet.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/compress/Slot.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/compress/CompressionTransform.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/compress/PiecewiseAffineTransform.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"reflect"
	"slices"

	"github.com/ekalinin/gopuml/internal/jfmt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// The layout compression of activity diagrams: the empty bands of a
// drawing (the gaps between the shapes along one axis, minus a margin) are
// removed.

// PiecewiseAffineTransform maps coordinates along one axis.
type PiecewiseAffineTransform interface {
	Transform(v float64) float64
}

// CompressionXorYBuilder is a text block drawn with its empty bands along
// one axis removed.
type CompressionXorYBuilder struct {
	TextBlockDefaults
	mode         CompressionMode
	textBlock    TextBlock
	cachedAffine PiecewiseAffineTransform
	cachedMinMax *geom.MinMax
}

// CompressionXorYBuilderBuild returns textBlock compressed along mode.
func CompressionXorYBuilderBuild(mode CompressionMode, textBlock TextBlock) TextBlock {
	return &CompressionXorYBuilder{mode: mode, textBlock: textBlock}
}

func (c *CompressionXorYBuilder) affineTransform(stringBounder font.StringBounder) PiecewiseAffineTransform {
	if c.cachedAffine == nil {
		slotFinder := SlotFinderCreate(c.mode, stringBounder)
		c.textBlock.DrawU(slotFinder)
		slotSet := slotFinder.SlotSet().Reverse().Smaller(5.0)
		c.cachedAffine = NewCompressionTransform(slotSet)
	}
	return c.cachedAffine
}

// DrawU implements TextBlock.
func (c *CompressionXorYBuilder) DrawU(ug UGraphic) {
	affine := c.affineTransform(ug.StringBounder())
	c.textBlock.DrawU(UGraphicCompressOnXorYCreate(c.mode, ug, affine))
}

// CalculateDimension implements TextBlock.
func (c *CompressionXorYBuilder) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	affine := c.affineTransform(stringBounder)
	dim := c.textBlock.CalculateDimension(stringBounder)
	if c.mode == CompressionOnX {
		return geom.NewXDimension2D(affine.Transform(dim.Width()), dim.Height())
	}
	return geom.NewXDimension2D(dim.Width(), affine.Transform(dim.Height()))
}

// MinMax implements TextBlock: the bounds of the drawing, cached.
func (c *CompressionXorYBuilder) MinMax(stringBounder font.StringBounder) geom.MinMax {
	if c.cachedMinMax == nil {
		m := TextBlockUtilsMinMax(c, stringBounder, false)
		c.cachedMinMax = &m
	}
	return *c.cachedMinMax
}

// UGraphicCompressOnXorY draws shapes with their coordinates along one axis
// transformed.
type UGraphicCompressOnXorY struct {
	UGraphicDelegator
	mode                 CompressionMode
	compressionTransform PiecewiseAffineTransform
	translate            UTranslate
}

var _ UGraphic = (*UGraphicCompressOnXorY)(nil)

// UGraphicCompressOnXorYCreate returns the transforming graphic over ug.
func UGraphicCompressOnXorYCreate(mode CompressionMode, ug UGraphic,
	compressionTransform PiecewiseAffineTransform) *UGraphicCompressOnXorY {
	return &UGraphicCompressOnXorY{UGraphicDelegator: NewUGraphicDelegator(ug), mode: mode,
		compressionTransform: compressionTransform, translate: UTranslateNone()}
}

// String is upstream's toString.
func (u *UGraphicCompressOnXorY) String() string { return "UGraphicCompressOnXorY " + u.mode.String() }

// Apply implements UGraphic: translations are kept here, strokes and colors
// go to the delegate, other changes panic, as upstream.
func (u *UGraphicCompressOnXorY) Apply(change UChange) UGraphic {
	switch c := change.(type) {
	case UTranslate:
		return &UGraphicCompressOnXorY{UGraphicDelegator: NewUGraphicDelegator(u.Ug()), mode: u.mode,
			compressionTransform: u.compressionTransform, translate: u.translate.Compose(c)}
	case UStroke, UBackground, color.HColor:
		return &UGraphicCompressOnXorY{UGraphicDelegator: NewUGraphicDelegator(u.Ug().Apply(change)), mode: u.mode,
			compressionTransform: u.compressionTransform, translate: u.translate}
	}
	throw(unsupportedOperation, "")
	return nil
}

// Draw implements UGraphic.
func (u *UGraphicCompressOnXorY) Draw(shape UShape) {
	x := u.translate.Dx()
	y := u.translate.Dy()
	if rect, ok := shape.(*URectangle); ok {
		if u.mode == CompressionOnX {
			x2 := u.ct(x + rect.Width())
			shape = rect.WithWidth(x2 - u.ct(x))
		} else {
			y2 := u.ct(y + rect.Height())
			shape = rect.WithHeight(y2 - u.ct(y))
		}
	}
	if centeredText, ok := shape.(*CenteredText); ok {
		text := centeredText.Text()
		totalWidth := centeredText.TotalWidth()
		realSpaceWidth := totalWidth
		if u.mode == CompressionOnX {
			realSpaceWidth = u.ct(x+totalWidth) - u.ct(x)
		}
		textWidth := text.CalculateDimension(u.StringBounder()).Width()
		pos := (realSpaceWidth - textWidth) / 2
		text.DrawU(u.Ug().Apply(u.translateOf(x, y)).Apply(UTranslateDx(pos)))
		return
	}
	if line, ok := shape.(*ULine); ok {
		u.drawLine(x, y, line)
		return
	}
	u.Ug().Apply(u.translateOf(x, y)).Draw(shape)
}

func (u *UGraphicCompressOnXorY) translateOf(x, y float64) UTranslate {
	if u.mode == CompressionOnX {
		return NewUTranslate(u.ct(x), y)
	}
	return NewUTranslate(x, u.ct(y))
}

func (u *UGraphicCompressOnXorY) drawLine(x, y float64, shape *ULine) {
	if u.mode == CompressionOnX {
		u.drawLine4(u.ct(x), y, u.ct(x+shape.DX()), y+shape.DY())
	} else {
		u.drawLine4(x, u.ct(y), x+shape.DX(), u.ct(y+shape.DY()))
	}
}

func (u *UGraphicCompressOnXorY) ct(v float64) float64 { return u.compressionTransform.Transform(v) }

func (u *UGraphicCompressOnXorY) drawLine4(x1, y1, x2, y2 float64) {
	if y1 > y2 {
		u.drawLine4(x2, y2, x1, y1)
		return
	}
	u.Ug().Apply(NewUTranslate(x1, y1)).Draw(NewULine(x2-x1, y2-y1))
}

// SlotFinder records the bands along one axis that the shapes drawn on it
// occupy.
type SlotFinder struct {
	UGraphicNo
	slot *SlotSet
	mode CompressionMode
}

var _ UGraphic = (*SlotFinder)(nil)

// SlotFinderCreate returns an empty finder.
func SlotFinderCreate(mode CompressionMode, stringBounder font.StringBounder) *SlotFinder {
	return &SlotFinder{UGraphicNo: NewUGraphicNo(stringBounder, UTranslateNone()), slot: &SlotSet{}, mode: mode}
}

// Apply implements UGraphic: translations move the finder; colors and
// strokes are ignored; other changes panic, as upstream.
func (s *SlotFinder) Apply(change UChange) UGraphic {
	switch change.(type) {
	case UBackground, color.HColor, UStroke, UTranslate:
	default:
		throw(unsupportedOperation, reflect.TypeOf(change).String())
	}
	tmp := s.Translate()
	if t, ok := change.(UTranslate); ok {
		tmp = s.Translate().Compose(t)
	}
	return &SlotFinder{UGraphicNo: NewUGraphicNo(s.StringBounder(), tmp), slot: s.slot, mode: s.mode}
}

// Draw implements UGraphic: the band of the shape is recorded; text blocks
// and other shapes are ignored.
func (s *SlotFinder) Draw(sh UShape) {
	x := s.Translate().Dx()
	y := s.Translate().Dy()
	if shape, ok := sh.(UShapeIgnorableForCompression); ok && shape.IsIgnoreForCompressionOn(s.mode) {
		shape.DrawWhenCompressed(s, s.mode)
		return
	}
	switch shape := sh.(type) {
	case *URectangle:
		s.add(x, x+shape.Width(), y, y+shape.Height())
	case *UPath:
		s.add(x+shape.MinX(), x+shape.MaxX(), y+shape.MinY(), y+shape.MaxY())
	case *UPolygon:
		if mode := shape.CompressionMode(); mode != nil && *mode == s.mode {
			return
		}
		s.add(x+shape.MinX(), x+shape.MaxX(), y+shape.MinY(), y+shape.MaxY())
	case *UEllipse:
		s.add(x, x+shape.Width(), y, y+shape.Height())
	case *UText:
		finder := TextLimitFinderCreate(s.StringBounder(), false)
		finder.Apply(NewUTranslate(x, y)).Draw(shape)
		s.add(finder.MinX(), finder.MaxX(), finder.MinY(), finder.MaxY())
	case *UEmpty:
		s.add(x, x+shape.Width(), y, y+shape.Height())
	}
}

// add records [x1, x2] or [y1, y2], following the mode.
func (s *SlotFinder) add(x1, x2, y1, y2 float64) {
	if s.mode == CompressionOnX {
		s.slot.AddSlot(x1, x2)
	} else {
		s.slot.AddSlot(y1, y2)
	}
}

// ColorMapper implements UGraphic: the identity.
func (s *SlotFinder) ColorMapper() *color.ColorMapper { return color.IdentityMapper }

// SlotSet returns the recorded bands.
func (s *SlotFinder) SlotSet() *SlotSet { return s.slot }

// SlotSet is a set of disjoint bands.
type SlotSet struct {
	all []Slot
}

// Filter returns the parts of the bands inside [start, end].
func (s *SlotSet) Filter(start, end float64) *SlotSet {
	result := &SlotSet{}
	for _, slot := range s.all {
		if intersec, ok := slot.IntersectRange(start, end); ok {
			result.all = append(result.all, intersec)
		}
	}
	return result
}

// AddAll adds the bands of other.
func (s *SlotSet) AddAll(other *SlotSet) { s.all = append(s.all, other.all...) }

// AddSlot adds [start, end], merged with the bands it touches.
func (s *SlotSet) AddSlot(start, end float64) {
	var collisions []Slot
	newSlot := NewSlot(start, end)
	kept := s.all[:0:0]
	for _, slot := range s.all {
		if slot.Intersect(newSlot) {
			collisions = append(collisions, slot)
		} else {
			kept = append(kept, slot)
		}
	}
	s.all = kept
	for _, slot := range collisions {
		newSlot = newSlot.Merge(slot)
	}
	s.all = append(s.all, newSlot)
}

// Smaller returns the bands shrunk by margin on each side, dropping those
// not wider than twice the margin.
func (s *SlotSet) Smaller(margin float64) *SlotSet {
	result := &SlotSet{}
	for _, sl := range s.all {
		if sl.Size() <= 2*margin {
			continue
		}
		result.AddSlot(sl.Start()+margin, sl.End()-margin)
	}
	return result
}

// String is upstream's toString.
func (s *SlotSet) String() string {
	result := "["
	for i, slot := range s.all {
		if i > 0 {
			result += ", "
		}
		result += slot.String()
	}
	return result + "]"
}

// Slots returns the bands.
func (s *SlotSet) Slots() []Slot { return s.all }

// Reverse returns the gaps between the bands; the bands get sorted, as
// upstream sorts its list in place.
func (s *SlotSet) Reverse() *SlotSet {
	result := &SlotSet{}
	slices.SortStableFunc(s.all, func(a, b Slot) int { return a.CompareTo(b) })
	var last *Slot
	for i := range s.all {
		slot := s.all[i]
		if last != nil {
			result.AddSlot(last.End(), slot.Start())
		}
		last = &s.all[i]
	}
	return result
}

// Slot is a band [start, end] with start < end.
type Slot struct {
	start, end float64
}

// NewSlot returns [start, end]; it panics unless start < end, as upstream.
func NewSlot(start, end float64) Slot {
	if start >= end {
		throw(illegalArgument, "start="+jfmt.DoubleToString(start)+" end="+jfmt.DoubleToString(end))
	}
	return Slot{start: start, end: end}
}

// String is upstream's toString.
func (s Slot) String() string {
	return "(" + jfmt.DoubleToString(s.start) + "," + jfmt.DoubleToString(s.end) + ")"
}

// Start returns the start.
func (s Slot) Start() float64 { return s.start }

// End returns the end.
func (s Slot) End() float64 { return s.end }

// Size returns the width of the band.
func (s Slot) Size() float64 { return s.end - s.start }

// Contains reports whether v is inside the band, bounds included.
func (s Slot) Contains(v float64) bool { return v >= s.start && v <= s.end }

// Intersect reports whether the bands touch.
func (s Slot) Intersect(other Slot) bool {
	return s.Contains(other.start) || s.Contains(other.end) || other.Contains(s.start) || other.Contains(s.end)
}

// Merge returns the band covering both.
func (s Slot) Merge(other Slot) Slot {
	return NewSlot(min(s.start, other.start), max(s.end, other.end))
}

// IntersectRange returns the part of the band inside [otherStart,
// otherEnd]; ok is false when they do not overlap.
func (s Slot) IntersectRange(otherStart, otherEnd float64) (Slot, bool) {
	if otherStart >= s.end || otherEnd <= s.start {
		return Slot{}, false
	}
	return NewSlot(max(s.start, otherStart), min(s.end, otherEnd)), true
}

// CompareTo orders bands by start.
func (s Slot) CompareTo(other Slot) int {
	switch {
	case s.start < other.start:
		return -1
	case s.start > other.start:
		return 1
	}
	return 0
}

// CompressionTransform removes bands: a coordinate moves back by the width
// of the bands before it.
type CompressionTransform struct {
	all []Slot
}

// NewCompressionTransform returns the transform removing the bands of
// slotSet.
func NewCompressionTransform(slotSet *SlotSet) *CompressionTransform {
	return &CompressionTransform{all: slotSet.Slots()}
}

// Transform implements PiecewiseAffineTransform.
func (c *CompressionTransform) Transform(v float64) float64 { return v - c.compressDelta(v) }

func (c *CompressionTransform) compressDelta(v float64) float64 {
	result := 0.0
	for _, s := range c.all {
		if s.Start() > v {
			continue
		}
		if v > s.End() {
			result += s.Size()
		} else {
			result += v - s.Start()
		}
	}
	return result
}
