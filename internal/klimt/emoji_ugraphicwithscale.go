// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/emoji/UGraphicWithScale.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"

	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// UGraphicWithScale is a graphic with the affine transform, the rotation
// angle and the scale of the SVG element being drawn, and the color
// resolver of the sprite. It is immutable: each method returns a new one.
type UGraphicWithScale struct {
	ug            UGraphic
	at            *XAffineTransform
	angle         float64
	scale         float64
	colorResolver *ColorResolver
}

// NewUGraphicWithScale returns the graphic of a sprite drawn at scale, its
// colors (foreground and background) set to the default color.
func NewUGraphicWithScale(ug UGraphic, colorResolver *ColorResolver, scale float64) *UGraphicWithScale {
	return &UGraphicWithScale{ug: ugraphicWithScaleUpdateColor(ug, colorResolver),
		colorResolver: colorResolver, at: XAffineTransformScaleInstance(scale, scale), angle: 0, scale: scale}
}

func ugraphicWithScaleUpdateColor(ug UGraphic, colorResolver *ColorResolver) UGraphic {
	c := colorResolver.DefaultColor()
	return ug.Apply(c).Apply(Bg(c))
}

func (u *UGraphicWithScale) with(ug UGraphic, at *XAffineTransform, angle, scale float64) *UGraphicWithScale {
	return &UGraphicWithScale{ug: ug, colorResolver: u.colorResolver, at: at, angle: angle, scale: scale}
}

// Ug returns the graphic.
func (u *UGraphicWithScale) Ug() UGraphic { return u.ug }

// Apply returns the graphic with a change applied.
func (u *UGraphicWithScale) Apply(change UChange) *UGraphicWithScale {
	return u.with(u.ug.Apply(change), u.at, u.angle, u.scale)
}

// TrueColor returns the color of an SVG color code (ColorResolver.TrueColor).
func (u *UGraphicWithScale) TrueColor(code string) color.HColor {
	return u.colorResolver.TrueColor(code)
}

// DefaultColor returns the default color of the sprite.
func (u *UGraphicWithScale) DefaultColor() color.HColor { return u.colorResolver.DefaultColor() }

// ApplyScale returns the graphic with the transform scaled. Different
// factors panic, as upstream throws IllegalArgumentException. As upstream,
// the scale becomes changex, not the product with the current scale.
func (u *UGraphicWithScale) ApplyScale(changex, changey float64) *UGraphicWithScale {
	if changex != changey {
		throw(illegalArgument, "")
	}
	cp := u.at.Copy()
	cp.Scale(changex, changey)
	return u.with(u.ug, cp, u.angle, 1*changex)
}

// Draw draws a shape on the graphic.
func (u *UGraphicWithScale) Draw(shape UShape) { u.ug.Draw(shape) }

// ApplyRotate returns the graphic with the transform rotated by deltaAngle
// degrees around (x, y).
func (u *UGraphicWithScale) ApplyRotate(deltaAngle, x, y float64) *UGraphicWithScale {
	cp := u.at.Copy()
	cp.Rotate(deltaAngle*math.Pi/180, x, y)
	return u.with(u.ug, cp, u.angle+deltaAngle, u.scale)
}

// ApplyTranslate returns the graphic with the transform translated.
func (u *UGraphicWithScale) ApplyTranslate(x, y float64) *UGraphicWithScale {
	cp := u.at.Copy()
	cp.Translate(x, y)
	return u.with(u.ug, cp, u.angle, u.scale)
}

// AffineTransform returns the transform; it must not be modified.
func (u *UGraphicWithScale) AffineTransform() *XAffineTransform { return u.at }

// ApplyMatrix returns the graphic with the SVG matrix(v1 v2 v3 v4 v5 v6)
// concatenated to the transform.
func (u *UGraphicWithScale) ApplyMatrix(v1, v2, v3, v4, v5, v6 float64) *UGraphicWithScale {
	cp := u.at.Copy()
	cp.Concatenate(NewXAffineTransformFlat([]float64{v1, v2, v3, v4, v5, v6}))
	return u.with(u.ug, cp, u.angle, u.scale)
}

// Angle returns the rotation in degrees.
func (u *UGraphicWithScale) Angle() float64 { return u.angle }

// InitialScale returns the scale.
func (u *UGraphicWithScale) InitialScale() float64 { return u.scale }
