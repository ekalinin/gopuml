// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/UGraphicNo.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/LimitFinder.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/TextLimitFinder.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/UGraphicNull.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"io"
	"reflect"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// UGraphicNo is the base of the graphics that measure instead of drawing.
// Types embedding it implement Draw and Apply.
type UGraphicNo struct {
	stringBounder font.StringBounder
	translate     UTranslate
}

// NewUGraphicNo returns the base with a bounder and a translation.
func NewUGraphicNo(stringBounder font.StringBounder, translate UTranslate) UGraphicNo {
	return UGraphicNo{stringBounder: stringBounder, translate: translate}
}

// StartURL does nothing.
func (u *UGraphicNo) StartURL(*URL) {}

// StartGroup does nothing.
func (u *UGraphicNo) StartGroup(*UGroup) {}

// CloseURL does nothing.
func (u *UGraphicNo) CloseURL() {}

// CloseGroup does nothing.
func (u *UGraphicNo) CloseGroup() {}

// ColorMapper panics, as upstream throws UnsupportedOperationException.
func (u *UGraphicNo) ColorMapper() *color.ColorMapper {
	throw(unsupportedOperation, "")
	return nil
}

// DefaultBackground returns black.
func (u *UGraphicNo) DefaultBackground() color.HColor { return color.Black }

// Param returns the null drawing state.
func (u *UGraphicNo) Param() UParam { return UParamNull{} }

// StringBounder returns the bounder.
func (u *UGraphicNo) StringBounder() font.StringBounder { return u.stringBounder }

// FlushUg does nothing.
func (u *UGraphicNo) FlushUg() {}

// MatchesProperty returns false.
func (u *UGraphicNo) MatchesProperty(string) bool { return false }

// WriteToStream panics, as upstream throws UnsupportedOperationException.
func (u *UGraphicNo) WriteToStream(io.Writer, string, int) error {
	throw(unsupportedOperation, "")
	return nil
}

// Translate returns the translation.
func (u *UGraphicNo) Translate() UTranslate { return u.translate }

// LimitFinder is a graphic that draws nothing and records the bounds of
// what would be drawn. Graphics derived by Apply share the bounds.
type LimitFinder struct {
	UGraphicNo
	minmax *geom.MinMaxMutable
	clip   *UClip
}

var _ UGraphic = (*LimitFinder)(nil)

// LimitFinderCreate returns a finder with empty bounds; initToZero makes the
// bounds start at the origin.
func LimitFinderCreate(stringBounder font.StringBounder, initToZero bool) *LimitFinder {
	return &LimitFinder{UGraphicNo: NewUGraphicNo(stringBounder, UTranslateNone()),
		minmax: geom.MinMaxMutableEmpty(initToZero)}
}

// Apply returns a finder sharing the bounds, with a translation or a clip
// applied. Changes other than colors, strokes, visibility, anti-aliasing,
// translations and clips panic, as upstream throws
// UnsupportedOperationException.
func (l *LimitFinder) Apply(change UChange) UGraphic {
	if change == nil {
		throw(nullPointer, "")
	}
	tmp := l.Translate()
	if t, ok := change.(UTranslate); ok {
		tmp = l.Translate().Compose(t)
	}
	result := &LimitFinder{UGraphicNo: NewUGraphicNo(l.StringBounder(), tmp), minmax: l.minmax}
	switch change.(type) {
	case UAntiAliasing, UBackground, UClip, color.HColor, UHidden, UStroke, UTranslate,
		CopyForegroundColorToBackgroundColor:
	default:
		throw(unsupportedOperation, reflect.TypeOf(change).String())
	}
	if c, ok := change.(UClip); ok {
		clip := c.TranslateBy(result.Translate())
		result.clip = &clip
	} else {
		result.clip = l.clip
	}
	return result
}

// MatchesProperty forwards to the bounder.
func (l *LimitFinder) MatchesProperty(propertyName string) bool {
	return l.StringBounder().MatchesProperty(propertyName)
}

// Draw records the bounds of shape; text blocks are drawn on the finder.
func (l *LimitFinder) Draw(shape UShape) {
	x := l.Translate().Dx()
	y := l.Translate().Dy()
	limitsOf(shape, x, y, l, l.StringBounder(), l.addPoint)
}

func (l *LimitFinder) addPoint(x, y float64) {
	if l.clip == nil || l.clip.IsInside(x, y) {
		l.minmax.AddPointXY(x, y)
	}
}

// ColorMapper returns the identity mapper.
func (l *LimitFinder) ColorMapper() *color.ColorMapper { return color.IdentityMapper }

// MaxX returns the right of the bounds.
func (l *LimitFinder) MaxX() float64 { return l.minmax.MaxX() }

// MaxY returns the bottom of the bounds.
func (l *LimitFinder) MaxY() float64 { return l.minmax.MaxY() }

// MinX returns the left of the bounds.
func (l *LimitFinder) MinX() float64 { return l.minmax.MinX() }

// MinY returns the top of the bounds.
func (l *LimitFinder) MinY() float64 { return l.minmax.MinY() }

// MinMax returns the bounds; empty bounds (from the origin) when nothing was
// drawn.
func (l *LimitFinder) MinMax() geom.MinMax {
	if l.minmax.IsInfinity() {
		return geom.MinMaxEmpty(true)
	}
	return geom.MinMaxFromMutable(l.minmax)
}

// TextLimitFinder records the bounds of the texts only.
type TextLimitFinder struct {
	UGraphicNo
	minmax *geom.MinMaxMutable
}

var _ UGraphic = (*TextLimitFinder)(nil)

// TextLimitFinderCreate returns a finder with empty bounds.
func TextLimitFinderCreate(stringBounder font.StringBounder, initToZero bool) *TextLimitFinder {
	return &TextLimitFinder{UGraphicNo: NewUGraphicNo(stringBounder, UTranslateNone()),
		minmax: geom.MinMaxMutableEmpty(initToZero)}
}

// Apply returns a finder sharing the bounds. Changes other than colors,
// strokes and translations panic, as upstream throws
// UnsupportedOperationException.
func (t *TextLimitFinder) Apply(change UChange) UGraphic {
	if change == nil {
		throw(nullPointer, "")
	}
	switch change.(type) {
	case UBackground, color.HColor, UStroke, UTranslate:
	default:
		throw(unsupportedOperation, reflect.TypeOf(change).String())
	}
	tmp := t.Translate()
	if tr, ok := change.(UTranslate); ok {
		tmp = t.Translate().Compose(tr)
	}
	return &TextLimitFinder{UGraphicNo: NewUGraphicNo(t.StringBounder(), tmp), minmax: t.minmax}
}

// Draw records the bounds of a text.
func (t *TextLimitFinder) Draw(shape UShape) {
	text, ok := shape.(*UText)
	if !ok {
		return
	}
	x := t.Translate().Dx()
	y := t.Translate().Dy()
	d := dim(t.StringBounder().CalculateDimension(text.FontConfiguration().Font(), text.Text()))
	y -= d.Height() - 1.5
	t.minmax.AddPointXY(x, y)
	t.minmax.AddPointXY(x, y+d.Height())
	t.minmax.AddPointXY(x+d.Width(), y)
	t.minmax.AddPointXY(x+d.Width(), y+d.Height())
}

// MaxX returns the right of the bounds.
func (t *TextLimitFinder) MaxX() float64 { return t.minmax.MaxX() }

// MaxY returns the bottom of the bounds.
func (t *TextLimitFinder) MaxY() float64 { return t.minmax.MaxY() }

// MinX returns the left of the bounds.
func (t *TextLimitFinder) MinX() float64 { return t.minmax.MinX() }

// MinY returns the top of the bounds.
func (t *TextLimitFinder) MinY() float64 { return t.minmax.MinY() }

// UGraphicNull is an output graphic without drivers: drawing a visible
// shape panics, as upstream throws UnsupportedOperationException. It
// measures text for PNG.
type UGraphicNull struct {
	AbstractUGraphic[string]
}

var _ CommonUGraphic = (*UGraphicNull)(nil)

// NewUGraphicNull returns the null graphic: black background, identity
// mapper.
func NewUGraphicNull() *UGraphicNull {
	u := &UGraphicNull{AbstractUGraphic: NewAbstractUGraphic[string](font.StringBounderAwt{})}
	u.SetSelf(u)
	u.Copy(color.Black, color.IdentityMapper, "foo")
	return u
}

// CopyUGraphic returns a copy.
func (u *UGraphicNull) CopyUGraphic() CommonUGraphic {
	result := &UGraphicNull{AbstractUGraphic: NewAbstractUGraphic[string](font.StringBounderAwt{})}
	result.SetSelf(result)
	result.CopyFrom(&u.AbstractUGraphic)
	return result
}

// WriteToStream writes nothing.
func (u *UGraphicNull) WriteToStream(io.Writer, string, int) error { return nil }

// EnsureVisible does nothing.
func (u *UGraphicNull) EnsureVisible(x, y float64) {}
