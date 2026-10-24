// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/color/HColor.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/color/HColors.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import "github.com/ekalinin/gopuml/internal/klimt/color"

// back is upstream's HColor.Back, the background change of a color.
type back struct {
	color color.HColor
}

// BackColor returns the background color.
func (b back) BackColor() color.HColor { return b.color }

// Bg returns the change that makes c the background color: upstream's
// HColor.bg(). It lives here because the color package cannot depend on
// UChange. A nil color panics, as calling bg() on null throws
// NullPointerException.
func Bg(c color.HColor) UBackground {
	if c == nil {
		throw(nullPointer, "")
	}
	return back{color: c}
}

// ChangeBack returns the change that makes the current foreground color of
// ug the background color, transparent when there is none: upstream's
// HColors.changeBack.
func ChangeBack(ug UGraphic) UChange {
	c := ug.Param().Color()
	if c == nil {
		return Bg(color.None())
	}
	return Bg(c)
}
