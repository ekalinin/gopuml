// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/AbstractCommonUGraphic.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/AbstractUGraphic.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"reflect"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// CommonUGraphic is a concrete graphic built on AbstractCommonUGraphic: it
// gives access to its common state and copies itself for Apply.
type CommonUGraphic interface {
	UGraphic
	// Common returns the embedded common state.
	Common() *AbstractCommonUGraphic
	// CopyUGraphic returns a copy of the graphic, upstream's abstract
	// copyUGraphic; Apply then changes the copy's common state.
	CopyUGraphic() CommonUGraphic
}

// AbstractCommonUGraphic is the state shared by the output graphics: stroke,
// pattern, visibility, colors, translation, clip and color mapper. Concrete
// graphics embed it, call SetSelf with themselves, and implement
// CommonUGraphic.
type AbstractCommonUGraphic struct {
	self              CommonUGraphic
	stroke            UStroke
	pattern           UPattern
	hidden            bool
	backColor         color.HColor
	color             color.HColor
	enlargeClip       bool
	stringBounder     font.StringBounder
	translate         UTranslate
	colorMapper       *color.ColorMapper
	clip              *UClip
	defaultBackground color.HColor
}

// NewAbstractCommonUGraphic returns the initial state: simple stroke, full
// pattern, transparent colors.
func NewAbstractCommonUGraphic(stringBounder font.StringBounder) AbstractCommonUGraphic {
	return AbstractCommonUGraphic{stroke: UStrokeSimple(), pattern: UPatternFull, backColor: color.None(),
		color: color.None(), stringBounder: stringBounder}
}

// SetSelf records the concrete graphic, so that Apply copies it.
func (a *AbstractCommonUGraphic) SetSelf(self CommonUGraphic) { a.self = self }

// Common returns a itself.
func (a *AbstractCommonUGraphic) Common() *AbstractCommonUGraphic { return a }

// BasicCopy sets the default background (which must not be nil) and the
// color mapper.
func (a *AbstractCommonUGraphic) BasicCopy(defaultBackground color.HColor, colorMapper *color.ColorMapper) {
	if defaultBackground == nil {
		throw(nullPointer, "")
	}
	a.colorMapper = colorMapper
	a.defaultBackground = defaultBackground
}

// BasicCopyFrom copies the drawing state of other (not the string bounder
// nor the self reference).
func (a *AbstractCommonUGraphic) BasicCopyFrom(other *AbstractCommonUGraphic) {
	if other.defaultBackground == nil {
		throw(nullPointer, "")
	}
	a.defaultBackground = other.defaultBackground
	a.enlargeClip = other.enlargeClip
	a.colorMapper = other.colorMapper
	a.translate = other.translate
	a.clip = other.clip
	a.stroke = other.stroke
	a.pattern = other.pattern
	a.hidden = other.hidden
	a.color = other.color
	a.backColor = other.backColor
}

// DefaultBackground returns the background of the whole drawing.
func (a *AbstractCommonUGraphic) DefaultBackground() color.HColor { return a.defaultBackground }

// DpiFactor returns 1.
func (a *AbstractCommonUGraphic) DpiFactor() float64 { return 1 }

// Apply returns a copy of the concrete graphic with change applied.
func (a *AbstractCommonUGraphic) Apply(change UChange) UGraphic {
	if change == nil {
		throw(nullPointer, "")
	}
	cp := a.self.CopyUGraphic()
	c := cp.Common()
	switch ch := change.(type) {
	case UTranslate:
		c.translate = ch.Compose(c.translate)
	case UClip:
		clip := ch.Translate(a.TranslateX(), a.TranslateY())
		c.clip = &clip
	case UStroke:
		c.stroke = ch
	case UPattern:
		c.pattern = ch
	case UHidden:
		c.hidden = ch == UHiddenHidden
	case UBackground:
		c.backColor = ch.BackColor()
	case color.HColor:
		c.color = ch
	case CopyForegroundColorToBackgroundColor:
		c.backColor = a.color
	}
	return cp
}

// Clip returns the clip, enlarged by 1 when EnlargeClip was called; nil for
// none.
func (a *AbstractCommonUGraphic) Clip() *UClip {
	if a.enlargeClip && a.clip != nil {
		e := a.clip.Enlarge(1)
		return &e
	}
	return a.clip
}

// EnlargeClip makes Clip return the clip enlarged by 1.
func (a *AbstractCommonUGraphic) EnlargeClip() { a.enlargeClip = true }

// commonParam is the UParam of a graphic, read live from its state as
// upstream's anonymous class does.
type commonParam struct{ a *AbstractCommonUGraphic }

func (p commonParam) IsHidden() bool          { return p.a.hidden }
func (p commonParam) Stroke() UStroke         { return p.a.stroke }
func (p commonParam) Color() color.HColor     { return p.a.color }
func (p commonParam) Backcolor() color.HColor { return p.a.backColor }
func (p commonParam) Pattern() UPattern       { return p.a.pattern }

// Param returns the drawing state.
func (a *AbstractCommonUGraphic) Param() UParam { return commonParam{a: a} }

// StringBounder returns the bounder of the output format.
func (a *AbstractCommonUGraphic) StringBounder() font.StringBounder { return a.stringBounder }

// TranslateX returns the current horizontal translation.
func (a *AbstractCommonUGraphic) TranslateX() float64 { return a.translate.Dx() }

// TranslateY returns the current vertical translation.
func (a *AbstractCommonUGraphic) TranslateY() float64 { return a.translate.Dy() }

// ColorMapper returns the color mapper.
func (a *AbstractCommonUGraphic) ColorMapper() *color.ColorMapper { return a.colorMapper }

// FlushUg does nothing.
func (a *AbstractCommonUGraphic) FlushUg() {}

// StartURL does nothing.
func (a *AbstractCommonUGraphic) StartURL(*URL) {}

// CloseURL does nothing.
func (a *AbstractCommonUGraphic) CloseURL() {}

// StartGroup does nothing.
func (a *AbstractCommonUGraphic) StartGroup(*UGroup) {}

// CloseGroup does nothing.
func (a *AbstractCommonUGraphic) CloseGroup() {}

// MatchesProperty returns false.
func (a *AbstractCommonUGraphic) MatchesProperty(string) bool { return false }

// Translate returns the current translation.
func (a *AbstractCommonUGraphic) Translate() UTranslate { return a.translate }

// The optional hooks of a graphic built on AbstractUGraphic, upstream's
// overridable methods; the defaults apply when the concrete graphic does not
// implement them.
type (
	// ManageHiddenAutomatically reports whether Draw skips hidden shapes
	// (default true).
	hookManageHidden interface{ ManageHiddenAutomatically() bool }
	// BeforeDraw runs before each driver call.
	hookBeforeDraw interface{ BeforeDraw() }
	// AfterDraw runs after each driver call.
	hookAfterDraw interface{ AfterDraw() }
	// DrawComment draws a comment (default: nothing).
	hookDrawComment interface{ DrawComment(c *UComment) }
)

// driverFunc is a registered driver with its shape type erased.
type driverFunc[O any] func(shape UShape, x, y float64, mapper *color.ColorMapper, param UParam, o O)

// AbstractUGraphic is the base of the output graphics: it dispatches each
// shape to the driver registered for its exact type and tracks the bounds of
// what is drawn. O is the output object (an SVG writer, an image, ...).
//
// A concrete graphic registers its drivers when it is built and in
// CopyUGraphic, and may implement ManageHiddenAutomatically() bool,
// BeforeDraw(), AfterDraw() and DrawComment(*UComment) to replace upstream's
// overridable defaults.
type AbstractUGraphic[O any] struct {
	AbstractCommonUGraphic
	graphic O
	minmax  *geom.MinMaxMutable
	drivers map[reflect.Type]driverFunc[O]
}

// NewAbstractUGraphic returns the base state of an output graphic.
func NewAbstractUGraphic[O any](stringBounder font.StringBounder) AbstractUGraphic[O] {
	return AbstractUGraphic[O]{AbstractCommonUGraphic: NewAbstractCommonUGraphic(stringBounder),
		drivers: map[reflect.Type]driverFunc[O]{}}
}

// Copy initializes a new graphic: default background, color mapper, output
// object and empty bounds.
func (a *AbstractUGraphic[O]) Copy(defaultBackground color.HColor, colorMapper *color.ColorMapper, graphic O) {
	a.BasicCopy(defaultBackground, colorMapper)
	a.graphic = graphic
	a.minmax = geom.MinMaxMutableEmpty(true)
}

// CopyFrom initializes a copy of other: it shares other's output object and
// bounds. As upstream, the drivers are not copied: each graphic registers its
// own, so that drivers built on the graphic (to read its clip, for example)
// see the copy.
func (a *AbstractUGraphic[O]) CopyFrom(other *AbstractUGraphic[O]) {
	a.BasicCopyFrom(&other.AbstractCommonUGraphic)
	a.graphic = other.graphic
	a.minmax = other.minmax
}

// GraphicObject returns the output object.
func (a *AbstractUGraphic[O]) GraphicObject() O { return a.graphic }

// RegisterDriver registers the driver of shapes of type S (the exact type,
// for example *URectangle).
func RegisterDriver[S UShape, O any](a *AbstractUGraphic[O], driver UDriver[S, O]) {
	a.drivers[reflect.TypeFor[S]()] = func(shape UShape, x, y float64, mapper *color.ColorMapper, param UParam, o O) {
		driver.Draw(shape.(S), x, y, mapper, param, o)
	}
}

// IgnoreShape registers a driver that draws nothing for shapes of type S.
func IgnoreShape[S UShape, O any](a *AbstractUGraphic[O]) {
	a.drivers[reflect.TypeFor[S]()] = func(UShape, float64, float64, *color.ColorMapper, UParam, O) {}
}

// Draw draws a shape with the driver of its type, after updating the
// bounds. A shape without driver panics, as upstream throws
// UnsupportedOperationException.
func (a *AbstractUGraphic[O]) Draw(shape UShape) {
	switch s := shape.(type) {
	case *SpecialText:
		s.Title().DrawU(a.self)
		return
	case *UEmpty:
		return
	case *UComment:
		if h, ok := a.self.(hookDrawComment); ok {
			h.DrawComment(s)
		}
		return
	}
	a.updateMinMax(shape)
	driver := a.drivers[reflect.TypeOf(shape)]
	if driver == nil {
		throw(unsupportedOperation, reflect.TypeOf(shape).String()+" "+reflect.TypeOf(a.self).String())
	}
	manageHidden := true
	if h, ok := a.self.(hookManageHidden); ok {
		manageHidden = h.ManageHiddenAutomatically()
	}
	if a.Param().IsHidden() && manageHidden {
		return
	}
	if h, ok := a.self.(hookBeforeDraw); ok {
		h.BeforeDraw()
	}
	driver(shape, a.TranslateX(), a.TranslateY(), a.ColorMapper(), a.Param(), a.graphic)
	if h, ok := a.self.(hookAfterDraw); ok {
		h.AfterDraw()
	}
}

// atomTextMarker is implemented by creole's AtomText (ported with creole):
// the SPECIALTXT graphics do not count its bounds.
type atomTextMarker interface{ isAtomText() }

func (a *AbstractUGraphic[O]) updateMinMax(shape UShape) {
	if _, ok := shape.(atomTextMarker); ok && a.self.MatchesProperty("SPECIALTXT") {
		return
	}
	x := a.Translate().Dx()
	y := a.Translate().Dy()
	limitsOf(shape, x, y, a.self, a.StringBounder(), a.addPoint)
}

func (a *AbstractUGraphic[O]) addPoint(x, y float64) {
	if clip := a.Clip(); clip == nil || clip.IsInside(x, y) {
		a.minmax.AddPointXY(x, y)
	}
}

// MaxX returns the right of the bounds.
func (a *AbstractUGraphic[O]) MaxX() float64 { return a.minmax.MaxX() }

// MaxY returns the bottom of the bounds.
func (a *AbstractUGraphic[O]) MaxY() float64 { return a.minmax.MaxY() }

// ResetMax resets the bounds.
func (a *AbstractUGraphic[O]) ResetMax() { a.minmax.Reset() }

// hackXForPolygon is upstream's HACK_X_FOR_POLYGON.
const hackXForPolygon = 10

// limitsOf reports the points that bound shape drawn at (x, y), as
// upstream's AbstractUGraphic.updateMinMax and LimitFinder.draw do: a text
// block is drawn on ug instead. Shapes without bounds are ignored; an
// unknown shape panics, as upstream throws UnsupportedOperationException.
func limitsOf(shape UShape, x, y float64, ug UGraphic, stringBounder font.StringBounder,
	addPoint func(x, y float64)) {
	switch s := shape.(type) {
	case *UText:
		d := dim(stringBounder.CalculateDimension(s.FontConfiguration().Font(), s.Text()))
		y -= d.Height() - 1.5
		addPoint(x, y)
		addPoint(x, y+d.Height())
		addPoint(x+d.Width(), y)
		addPoint(x+d.Width(), y+d.Height())
	case *ULine:
		addPoint(x, y)
		addPoint(x+s.DX(), y+s.DY())
	case *UEllipse:
		addPoint(x, y)
		addPoint(x+s.Width()-1+float64(s.DeltaShadow()*2), y+s.Height()-1+float64(s.DeltaShadow()*2))
	case *UPolygon:
		if len(s.Points()) == 0 {
			return
		}
		addPoint(x+s.MinX()-hackXForPolygon, y+s.MinY())
		addPoint(x+s.MaxX()+hackXForPolygon, y+s.MaxY())
	case *UPath:
		addPoint(x+s.MinX(), y+s.MinY())
		addPoint(x+s.MaxX(), y+s.MaxY())
	case *URectangle:
		addPoint(x-1, y-1)
		addPoint(x+s.Width()-1+float64(s.DeltaShadow()*2), y+s.Height()-1+float64(s.DeltaShadow()*2))
	case *DotPath:
		shapeMinMax := s.MinMax()
		addPoint(x+shapeMinMax.MinX(), y+shapeMinMax.MinY())
		addPoint(x+shapeMinMax.MaxX(), y+shapeMinMax.MaxY())
	case *UImage:
		addPoint(x, y)
		addPoint(x+float64(s.Width())-1, y+float64(s.Height())-1)
	case *UImageSvg:
		addPoint(x, y)
		addPoint(x+s.Width()-1, y+s.Height()-1)
	case *UImageTikz:
		addPoint(x, y)
		addPoint(x+s.Width()-1, y+s.Height()-1)
	case *UComment:
	case *UEmpty:
		addPoint(x, y)
		addPoint(x+s.Width(), y+s.Height())
	case TextBlock:
		s.DrawU(ug)
	case *UCenteredCharacter:
		// To be done
	case *CenteredText:
		// Ignored
	case *SpecialText:
		// Ignored
	case CopyForegroundColorToBackgroundColor:
		// Ignored
	case *UPixel, UPixel:
		addPoint(x, y)
	default:
		if shape == nil {
			throw(nullPointer, "")
		}
		throw(unsupportedOperation, reflect.TypeOf(shape).String())
	}
}
