// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/TileArguments.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/rose"
)

// TileArguments are what the tiles share: the text measurer, the living
// spaces, the skin, the skin parameters and the origins of the two lines.
type TileArguments struct {
	stringBounder font.StringBounder
	xorigin       Real
	yorigin       Real
	livingSpaces  *LivingSpaces
	skin          *rose.Rose
	skinParam     klimt.ISkinParam
	// bordered is set once the main tile is built. As upstream, the copies
	// made by WithBackColorGeneral and WithBackColor have none; no tile
	// built from a copy reads the borders.
	bordered Bordered
}

// NewTileArguments returns the arguments.
func NewTileArguments(stringBounder font.StringBounder, livingSpaces *LivingSpaces, skin *rose.Rose,
	skinParam klimt.ISkinParam, xorigin, yorigin Real) *TileArguments {
	return &TileArguments{stringBounder: stringBounder, xorigin: xorigin, yorigin: yorigin,
		livingSpaces: livingSpaces, skin: skin, skinParam: skinParam}
}

// WithBackColorGeneral returns the arguments with background colors.
func (a *TileArguments) WithBackColorGeneral(backColorElement, backColorGeneral color.HColor) *TileArguments {
	return NewTileArguments(a.stringBounder, a.livingSpaces, a.skin,
		klimt.NewSkinParamBackcolored(a.skinParam, backColorElement, backColorGeneral, false), a.xorigin, a.yorigin)
}

// WithBackColor returns the arguments with the colors of a reference.
func (a *TileArguments) WithBackColor(reference *Reference) *TileArguments {
	return NewTileArguments(a.stringBounder, a.livingSpaces, a.skin,
		klimt.NewSkinParamBackcoloredReference(a.skinParam, reference.BackColorElement(),
			reference.BackColorGeneral()), a.xorigin, a.yorigin)
}

// StringBounder returns the text measurer.
func (a *TileArguments) StringBounder() font.StringBounder { return a.stringBounder }

// XOrigin returns the origin of the X line.
func (a *TileArguments) XOrigin() Real { return a.xorigin }

// YOrigin returns the origin of the Y line.
func (a *TileArguments) YOrigin() Real { return a.yorigin }

// LivingSpaces returns the living spaces.
func (a *TileArguments) LivingSpaces() *LivingSpaces { return a.livingSpaces }

// Skin returns the skin.
func (a *TileArguments) Skin() *rose.Rose { return a.skin }

// SkinParam returns the skin parameters.
func (a *TileArguments) SkinParam() klimt.ISkinParam { return a.skinParam }

// LivingSpace returns the living space of a participant.
func (a *TileArguments) LivingSpace(p *Participant) *LivingSpace { return a.livingSpaces.Get(p) }

// FirstLivingSpace returns the first living space.
func (a *TileArguments) FirstLivingSpace() *LivingSpace { return a.livingSpaces.Values()[0] }

// LastLivingSpace returns the last living space, nil for none.
func (a *TileArguments) LastLivingSpace() *LivingSpace {
	values := a.livingSpaces.Values()
	if len(values) == 0 {
		return nil
	}
	return values[len(values)-1]
}

// SetBordered sets the source of the borders.
func (a *TileArguments) SetBordered(bordered Bordered) { a.bordered = bordered }

// Border1 returns the left border.
func (a *TileArguments) Border1() float64 { return a.bordered.Border1() }

// Border2 returns the right border.
func (a *TileArguments) Border2() float64 { return a.bordered.Border2() }
