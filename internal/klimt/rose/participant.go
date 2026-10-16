// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/rose/ComponentRoseParticipant.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package rose

import (
	"github.com/ekalinin/plantuml-in-go/internal/jmath"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
)

// ComponentRoseParticipant is the head or tail of a participant or of
// collections: the name in a box (two stacked boxes for collections).
type ComponentRoseParticipant struct {
	klimt.AbstractTextualComponent
	back            color.HColor
	foregroundColor color.HColor
	deltaShadow     float64
	roundCorner     float64
	diagonalCorner  float64
	stroke          klimt.UStroke
	minWidth        float64
	collections     bool
	padding         klimt.ClockwiseTopRightBottomLeft
	margin          klimt.ClockwiseTopRightBottomLeft
}

// NewComponentRoseParticipant returns the box of a participant; the text
// padding is the style's, padding is the position of the text in the box
// and margin the space around the box.
func NewComponentRoseParticipant(style, stereo *klimt.Style, stringsToDisplay *klimt.Display,
	skinParam klimt.ISkinParam, minWidth float64, collections bool, padding,
	margin klimt.ClockwiseTopRightBottomLeft) *ComponentRoseParticipant {
	c := &ComponentRoseParticipant{}
	c.InitAbstractTextualComponent(c, style, stereo, klimt.LineBreakStrategyNone, style.Padding(), skinParam,
		stringsToDisplay, false)

	c.roundCorner = c.RoundCorner()
	c.diagonalCorner = c.DiagonalCorner()
	biColor := c.SymbolContext()
	c.stroke = c.Stroke()

	c.padding = padding
	c.margin = margin
	c.minWidth = minWidth
	c.collections = collections
	c.back = biColor.BackColor()
	c.deltaShadow = biColor.DeltaShadow()
	c.foregroundColor = biColor.ForeColor()
	return c
}

// DrawInternalU draws the box (two for collections) and the name.
func (c *ComponentRoseParticipant) DrawInternalU(ug klimt.UGraphic, _ *klimt.Area) {
	stringBounder := ug.StringBounder()

	// Margin: empty space outside the box
	ug = ug.Apply(klimt.NewUTranslate(c.margin.Left(), c.margin.Top()))

	if c.foregroundColor != nil {
		ug = ug.Apply(c.foregroundColor)
	}
	if c.back != nil {
		ug = ug.Apply(klimt.Bg(c.back))
	}
	ug = ug.Apply(c.stroke)

	// Padding: space inside the box, between border and text
	boxWidth := c.TextWidth(stringBounder)
	boxHeight := c.TextHeight(stringBounder)

	rect := klimt.URectangleBuild(boxWidth, boxHeight).Rounded(c.roundCorner).DiagonalCorner(c.diagonalCorner)
	rect.SetDeltaShadow(c.deltaShadow)
	if c.collections {
		ug.Apply(klimt.UTranslateDx(c.deltaCollection())).Draw(rect)
		ug = ug.Apply(klimt.UTranslateDy(c.deltaCollection()))
	}
	ug.Draw(rect)
	ug = ug.Apply(klimt.UStrokeSimple())

	// Text: centered inside the box, offset by padding
	textBlock := c.TextBlock()
	textBlock.DrawU(ug.Apply(c.padding.Translate()))
}

func (c *ComponentRoseParticipant) deltaCollection() float64 {
	if c.collections {
		return 4
	}
	return 0
}

// PreferredHeight returns the height of the box, its margin, its shadow
// and the second box of collections.
func (c *ComponentRoseParticipant) PreferredHeight(stringBounder font.StringBounder) float64 {
	return c.TextHeight(stringBounder) + c.margin.Top() + c.margin.Bottom() + c.deltaShadow + 1 + c.deltaCollection()
}

// PreferredWidth returns the width of the box, its margin, its shadow and
// the second box of collections.
func (c *ComponentRoseParticipant) PreferredWidth(stringBounder font.StringBounder) float64 {
	return c.TextWidth(stringBounder) + c.margin.Left() + c.margin.Right() + c.deltaShadow + c.deltaCollection()
}

// PureTextWidth returns the width of the name, at least the minimum width
// of the style. Upstream overrides getPureTextWidth, so TextWidth uses it.
func (c *ComponentRoseParticipant) PureTextWidth(stringBounder font.StringBounder) float64 {
	return jmath.Max(c.AbstractTextualComponent.PureTextWidth(stringBounder), c.minWidth)
}

// TextWidth returns the width of the name, at least the minimum width,
// with the horizontal padding.
func (c *ComponentRoseParticipant) TextWidth(stringBounder font.StringBounder) float64 {
	return c.PureTextWidth(stringBounder) + c.OldPaddingX1() + c.OldPaddingX2()
}
