// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/UShape.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/UChange.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/UBackground.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/UShapeSized.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/Shadowable.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/AbstractShadowable.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/CopyForegroundColorToBackgroundColor.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/UPattern.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/UAntiAliasing.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/AffineTransformType.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/ClipContainer.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import "github.com/ekalinin/gopuml/internal/klimt/color"

// UShape is anything a UGraphic can draw. Upstream's UShape is a marker
// interface; drawing code dispatches on the concrete type (URectangle,
// UText, a TextBlock, ...), so any value is accepted here.
type UShape any

// UChange is anything that UGraphic.Apply accepts: a UTranslate, a UClip, a
// UStroke, a UPattern, a UHidden, a UBackground, a color.HColor (the
// foreground color), CopyForegroundColorToBackgroundColor or an
// UAntiAliasing. Upstream's UChange is a marker interface implemented by
// HColor, which lives in the color package, so any value is accepted here.
type UChange any

// UBackground is the change that sets the background (fill) color. It is
// built by Bg.
type UBackground interface {
	BackColor() color.HColor
}

// UShapeSized is a shape with a size.
type UShapeSized interface {
	Width() float64
	Height() float64
}

// Shadowable is a shape that can draw a shadow.
type Shadowable interface {
	SetDeltaShadow(deltaShadow float64)
	DeltaShadow() float64
}

// AbstractShadowable holds the shadow size of a shape; shapes embed it.
type AbstractShadowable struct {
	deltaShadow float64
}

// DeltaShadow returns the shadow size, 0 for no shadow.
func (s *AbstractShadowable) DeltaShadow() float64 { return s.deltaShadow }

// SetDeltaShadow sets the shadow size.
func (s *AbstractShadowable) SetDeltaShadow(deltaShadow float64) { s.deltaShadow = deltaShadow }

// CopyForegroundColorToBackgroundColor is the change that makes the
// background color the current foreground color.
type CopyForegroundColorToBackgroundColor struct{}

// UPattern is a fill pattern.
type UPattern int

// Fill patterns, in upstream's order.
const (
	UPatternFull UPattern = iota
	UPatternHorizontalStripe
	UPatternVerticalStripe
	UPatternSmallCircle
)

var upatternNames = []string{"FULL", "HORIZONTAL_STRIPE", "VERTICAL_STRIPE", "SMALL_CIRCLE"}

// String returns upstream's constant name.
func (p UPattern) String() string { return enumName("UPattern", upatternNames, int(p)) }

// UAntiAliasing switches anti-aliasing for raster output.
type UAntiAliasing int

// Anti-aliasing modes, in upstream's order. Upstream's apply(Graphics2D)
// belongs to the raster driver.
const (
	AntiAliasingOn UAntiAliasing = iota
	AntiAliasingOff
)

// String returns upstream's constant name.
func (a UAntiAliasing) String() string {
	return enumName("UAntiAliasing", []string{"ANTI_ALIASING_ON", "ANTI_ALIASING_OFF"}, int(a))
}

// AffineTransformType is the interpolation used to scale images.
type AffineTransformType int

// Interpolations, in upstream's order.
const (
	AffineTransformNearestNeighbor AffineTransformType = iota
	AffineTransformBilinear
)

// ToLegacyInt returns the value of the matching java.awt.image
// AffineTransformOp constant.
func (t AffineTransformType) ToLegacyInt() int {
	if t == AffineTransformBilinear {
		return 2 // AffineTransformOp.TYPE_BILINEAR
	}
	return 1 // AffineTransformOp.TYPE_NEAREST_NEIGHBOR
}

// ClipContainer is something with a clip.
type ClipContainer interface {
	Clip() *UClip
}

// UParam is the drawing state a driver receives: colors, stroke, visibility
// and pattern.
type UParam interface {
	Color() color.HColor
	Backcolor() color.HColor
	Stroke() UStroke
	IsHidden() bool
	Pattern() UPattern
}

// UParamNull is the drawing state of graphics that draw nothing: black,
// simple stroke, full pattern.
type UParamNull struct{}

// Color returns black.
func (UParamNull) Color() color.HColor { return color.Black }

// Backcolor returns black.
func (UParamNull) Backcolor() color.HColor { return color.Black }

// Stroke returns the simple stroke.
func (UParamNull) Stroke() UStroke { return UStrokeSimple() }

// IsHidden returns false.
func (UParamNull) IsHidden() bool { return false }

// Pattern returns the full pattern.
func (UParamNull) Pattern() UPattern { return UPatternFull }

// enumName returns names[v], or "Type(v)" when v is out of range.
func enumName(typ string, names []string, v int) string {
	if v < 0 || v >= len(names) {
		return typ + "(" + itoa(v) + ")"
	}
	return names[v]
}
