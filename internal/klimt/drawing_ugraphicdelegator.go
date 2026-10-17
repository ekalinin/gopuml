// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/UGraphicDelegator.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/UGraphicFilter.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/UGraphicDispatchDrawable.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/AbstractUGraphicHorizontalLine.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/UGraphicStencil.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"io"
	"reflect"
	"slices"

	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// UGraphicDelegator forwards everything to another graphic. Types embedding
// it implement Apply (upstream leaves it abstract) and may replace Draw.
type UGraphicDelegator struct {
	ug UGraphic
}

// NewUGraphicDelegator returns a delegator to ug.
func NewUGraphicDelegator(ug UGraphic) UGraphicDelegator { return UGraphicDelegator{ug: ug} }

// MatchesProperty forwards to the delegate.
func (d *UGraphicDelegator) MatchesProperty(propertyName string) bool {
	return d.ug.MatchesProperty(propertyName)
}

// StringBounder forwards to the delegate.
func (d *UGraphicDelegator) StringBounder() font.StringBounder { return d.ug.StringBounder() }

// Param forwards to the delegate.
func (d *UGraphicDelegator) Param() UParam { return d.ug.Param() }

// Draw forwards to the delegate.
func (d *UGraphicDelegator) Draw(shape UShape) { d.ug.Draw(shape) }

// ColorMapper forwards to the delegate.
func (d *UGraphicDelegator) ColorMapper() *color.ColorMapper { return d.ug.ColorMapper() }

// StartURL forwards to the delegate.
func (d *UGraphicDelegator) StartURL(url *URL) { d.ug.StartURL(url) }

// CloseURL forwards to the delegate.
func (d *UGraphicDelegator) CloseURL() { d.ug.CloseURL() }

// StartGroup forwards to the delegate.
func (d *UGraphicDelegator) StartGroup(group *UGroup) { d.ug.StartGroup(group) }

// CloseGroup forwards to the delegate.
func (d *UGraphicDelegator) CloseGroup() { d.ug.CloseGroup() }

// Ug returns the delegate.
func (d *UGraphicDelegator) Ug() UGraphic { return d.ug }

// FlushUg forwards to the delegate.
func (d *UGraphicDelegator) FlushUg() { d.ug.FlushUg() }

// DefaultBackground forwards to the delegate.
func (d *UGraphicDelegator) DefaultBackground() color.HColor { return d.ug.DefaultBackground() }

// WriteToStream forwards to the delegate.
func (d *UGraphicDelegator) WriteToStream(w io.Writer, metadata string, dpi int) error {
	return d.ug.WriteToStream(w, metadata, dpi)
}

// UGraphicFilter draws only the shapes of some exact types.
type UGraphicFilter struct {
	UGraphicDelegator
	toprint []reflect.Type
}

var _ UGraphic = (*UGraphicFilter)(nil)

// NewUGraphicFilter returns the filter; toprint are sample values of the
// shape types to draw, for example (*URectangle)(nil).
func NewUGraphicFilter(ug UGraphic, toprint ...UShape) *UGraphicFilter {
	f := &UGraphicFilter{UGraphicDelegator: NewUGraphicDelegator(ug)}
	for _, t := range toprint {
		f.toprint = append(f.toprint, reflect.TypeOf(t))
	}
	return f
}

// Apply panics, as upstream throws UnsupportedOperationException.
func (f *UGraphicFilter) Apply(UChange) UGraphic {
	throw(unsupportedOperation, "")
	return nil
}

// Draw draws shape when its type is one of the filter's.
func (f *UGraphicFilter) Draw(shape UShape) {
	if slices.Contains(f.toprint, reflect.TypeOf(shape)) {
		f.Ug().Draw(shape)
	}
}

// UGraphicDispatchDrawable draws drawable shapes (text blocks) by calling
// their DrawU on itself, and hands the other shapes, and images, to the
// delegate.
type UGraphicDispatchDrawable struct {
	UGraphicDelegator
}

var _ UGraphic = (*UGraphicDispatchDrawable)(nil)

// NewUGraphicDispatchDrawable returns the dispatcher over ug.
func NewUGraphicDispatchDrawable(ug UGraphic) *UGraphicDispatchDrawable {
	return &UGraphicDispatchDrawable{UGraphicDelegator: NewUGraphicDelegator(ug)}
}

// Draw dispatches the shape.
func (d *UGraphicDispatchDrawable) Draw(shape UShape) {
	_, isImage := shape.(*UImage)
	if drawable, ok := shape.(UDrawable); ok && !isImage {
		// UImage implements TextBlock/UDrawable but must be drawn by the
		// low-level driver
		drawable.DrawU(d)
		return
	}
	d.Ug().Draw(shape)
}

// Apply returns the dispatcher over the changed delegate.
func (d *UGraphicDispatchDrawable) Apply(change UChange) UGraphic {
	return NewUGraphicDispatchDrawable(d.Ug().Apply(change))
}

// HorizontalLineGraphic is a graphic built on AbstractUGraphicHorizontalLine,
// upstream's abstract methods copy and drawHline.
type HorizontalLineGraphic interface {
	UGraphic
	// HorizontalLineBase returns the embedded base.
	HorizontalLineBase() *AbstractUGraphicHorizontalLine
	// CopyHorizontalLine returns a graphic of the same kind over ug.
	CopyHorizontalLine(ug UGraphic) HorizontalLineGraphic
	// DrawHline draws a separator at the vertical translation.
	DrawHline(ug UGraphic, line *UHorizontalLine, translate UTranslate)
}

// AbstractUGraphicHorizontalLine keeps its own translation, so that
// UHorizontalLine separators are drawn across their container whatever the
// horizontal translation. Concrete graphics embed it and call SetSelf.
type AbstractUGraphicHorizontalLine struct {
	UGraphicDelegator
	translate UTranslate
	self      HorizontalLineGraphic
}

// NewAbstractUGraphicHorizontalLine returns the base over ug.
func NewAbstractUGraphicHorizontalLine(ug UGraphic) AbstractUGraphicHorizontalLine {
	return AbstractUGraphicHorizontalLine{UGraphicDelegator: NewUGraphicDelegator(ug)}
}

// SetSelf records the concrete graphic.
func (a *AbstractUGraphicHorizontalLine) SetSelf(self HorizontalLineGraphic) { a.self = self }

// HorizontalLineBase returns a itself.
func (a *AbstractUGraphicHorizontalLine) HorizontalLineBase() *AbstractUGraphicHorizontalLine {
	return a
}

// Apply keeps translations for itself and applies other changes to the
// delegate, a clip after the translation.
func (a *AbstractUGraphicHorizontalLine) Apply(change UChange) UGraphic {
	var result HorizontalLineGraphic
	switch ch := change.(type) {
	case UTranslate:
		result = a.self.CopyHorizontalLine(a.Ug())
		result.HorizontalLineBase().translate = a.translate.Compose(ch)
	case UClip:
		clip := ch.TranslateBy(a.translate)
		result = a.self.CopyHorizontalLine(a.Ug().Apply(clip))
		result.HorizontalLineBase().translate = a.translate
	default:
		result = a.self.CopyHorizontalLine(a.Ug().Apply(change))
		result.HorizontalLineBase().translate = a.translate
	}
	return result
}

// Draw draws separators with DrawHline and other shapes translated.
func (a *AbstractUGraphicHorizontalLine) Draw(shape UShape) {
	if line, ok := shape.(*UHorizontalLine); ok {
		a.self.DrawHline(a.Ug(), line, UTranslateDy(a.translate.Dy()))
		return
	}
	a.Ug().Apply(a.translate).Draw(shape)
}

// UGraphicStencil draws separators across a stencil.
type UGraphicStencil struct {
	AbstractUGraphicHorizontalLine
	stencil       Stencil
	defaultStroke UStroke
}

var _ HorizontalLineGraphic = (*UGraphicStencil)(nil)

// UGraphicStencilCreate returns ug drawing separators across stencil.
func UGraphicStencilCreate(ug UGraphic, stencil Stencil, defaultStroke UStroke) UGraphic {
	return newUGraphicStencil(ug, stencil, defaultStroke)
}

// UGraphicStencilCreateDim returns ug drawing separators across a
// rectangle of a dimension.
func UGraphicStencilCreateDim(ug UGraphic, d geom.XDimension2D) UGraphic {
	return newUGraphicStencil(ug, rectangleStencil(d), UStrokeSimple())
}

func rectangleStencil(d geom.XDimension2D) Stencil {
	return StencilFunc{
		Starting: func(font.StringBounder, float64) float64 { return 0 },
		Ending:   func(font.StringBounder, float64) float64 { return d.Width() },
	}
}

func newUGraphicStencil(ug UGraphic, stencil Stencil, defaultStroke UStroke) *UGraphicStencil {
	s := &UGraphicStencil{AbstractUGraphicHorizontalLine: NewAbstractUGraphicHorizontalLine(ug), stencil: stencil,
		defaultStroke: defaultStroke}
	s.SetSelf(s)
	return s
}

// CopyHorizontalLine returns a stencil graphic over ug.
func (s *UGraphicStencil) CopyHorizontalLine(ug UGraphic) HorizontalLineGraphic {
	return newUGraphicStencil(ug, s.stencil, s.defaultStroke)
}

// DrawHline draws the separator across the stencil.
func (s *UGraphicStencil) DrawHline(ug UGraphic, line *UHorizontalLine, translate UTranslate) {
	line.DrawLineInternal(ug, s.stencil, translate.Dy(), s.defaultStroke)
}
