// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/rose/ComponentRoseNote.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/rose/ComponentRoseNoteBox.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/rose/ComponentRoseNoteHexagonal.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package rose

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// ComponentRoseNote is a note: a text in a sheet with a folded corner.
type ComponentRoseNote struct {
	klimt.AbstractTextualComponent
	paddingX      float64
	paddingY      float64
	symbolContext *klimt.Fashion
	roundCorner   float64
	position      geom.HorizontalAlignment
}

var _ klimt.Stencil = (*ComponentRoseNote)(nil)

// NewComponentRoseNote returns a note; textAlignment chooses the padding
// (a centered text is padded on both sides) and position places the text
// in a wider area. Both may be HorizontalAlignmentNull; colors may be nil.
func NewComponentRoseNote(style *klimt.Style, strings *klimt.Display, paddingX, paddingY float64,
	skinParam klimt.ISkinParam, textAlignment, position geom.HorizontalAlignment,
	colors *color.Colors) *ComponentRoseNote {
	c := &ComponentRoseNote{}
	padding := klimt.ClockwiseTopRightBottomLeftOf(5, 15, 5, 6)
	if textAlignment == geom.HorizontalAlignmentCenter {
		padding = klimt.ClockwiseTopRightBottomLeftOf(5, 15, 5, 15)
	}
	c.InitAbstractTextualComponent(c, style, style, style.WrapWidth(), padding, skinParam, strings, true)
	c.paddingX = paddingX
	c.paddingY = paddingY
	c.position = position
	c.symbolContext = c.SymbolContextColors(colors)
	c.roundCorner = c.RoundCorner()
	return c
}

// PreferredWidth returns the width of the text, the padding and the shadow.
func (c *ComponentRoseNote) PreferredWidth(stringBounder font.StringBounder) float64 {
	result := c.TextWidth(stringBounder) + 2*c.PaddingX() + c.symbolContext.DeltaShadow()
	return result
}

// PreferredHeight returns the height of the text, the padding and the
// shadow.
func (c *ComponentRoseNote) PreferredHeight(stringBounder font.StringBounder) float64 {
	return c.TextHeight(stringBounder) + 2*c.PaddingY() + c.symbolContext.DeltaShadow()
}

// PaddingX returns the horizontal padding given to the constructor.
func (c *ComponentRoseNote) PaddingX() float64 { return c.paddingX }

// PaddingY returns the vertical padding given to the constructor.
func (c *ComponentRoseNote) PaddingY() float64 { return c.paddingY }

// DrawInternalU draws the sheet and the text. An area narrower than the
// note panics, as upstream throws IllegalArgumentException.
func (c *ComponentRoseNote) DrawInternalU(ug klimt.UGraphic, area *klimt.Area) {
	stringBounder := ug.StringBounder()
	textHeight := javaD2I(c.TextHeight(stringBounder))

	x2 := javaD2I(c.TextWidth(stringBounder))
	diffX := area.DimensionToUse().Width() - c.PreferredWidth(stringBounder)
	if diffX < 0 {
		throw(illegalArgument, "")
	}
	if area.DimensionToUse().Width() > c.PreferredWidth(stringBounder) {
		x2 = javaD2I(area.DimensionToUse().Width() - 2*c.PaddingX())
	}

	polygon := klimt.OpalePolygonNormal(float64(x2), float64(textHeight), c.roundCorner)
	polygon.SetDeltaShadow(c.symbolContext.DeltaShadow())

	ug = c.symbolContext.Apply(ug)
	ug.Draw(polygon)

	ug.Draw(klimt.OpaleCorner(float64(x2), c.roundCorner))
	ug2 := klimt.UGraphicStencilCreate(ug, c, klimt.UStrokeSimple())

	switch c.position {
	case geom.HorizontalAlignmentLeft:
		ug2 = ug2.Apply(klimt.NewUTranslate(c.OldPaddingX1(), c.OldPaddingY()))
	case geom.HorizontalAlignmentRight:
		ug2 = ug2.Apply(klimt.NewUTranslate(area.DimensionToUse().Width()-c.TextWidth(stringBounder),
			c.OldPaddingY()))
	default:
		ug2 = ug2.Apply(klimt.NewUTranslate(c.OldPaddingX1()+diffX/2, c.OldPaddingY()))
	}

	c.TextBlock().DrawU(ug2)
}

// StartingX returns 0: separators start at the left border.
func (c *ComponentRoseNote) StartingX(font.StringBounder, float64) float64 { return 0 }

// EndingX returns the width of the text with its padding.
func (c *ComponentRoseNote) EndingX(stringBounder font.StringBounder, _ float64) float64 {
	return c.TextWidth(stringBounder)
}

// ComponentRoseNoteBox is a note drawn as a rectangle ("rnote").
type ComponentRoseNoteBox struct {
	klimt.AbstractTextualComponent
	symbolContext *klimt.Fashion
	roundCorner   float64
}

// NewComponentRoseNoteBox returns a rectangular note; colors may be nil.
func NewComponentRoseNoteBox(style *klimt.Style, strings *klimt.Display, skinParam klimt.ISkinParam,
	colors *color.Colors) *ComponentRoseNoteBox {
	c := &ComponentRoseNoteBox{}
	c.InitAbstractTextualComponent(c, style, style, style.WrapWidth(), klimt.ClockwiseTopRightBottomLeftOf(4, 4, 4, 4),
		skinParam, strings, false)
	c.symbolContext = c.SymbolContextColors(colors)
	c.roundCorner = c.RoundCorner()
	return c
}

// PreferredWidth returns the width of the text and the padding.
func (c *ComponentRoseNoteBox) PreferredWidth(stringBounder font.StringBounder) float64 {
	result := c.TextWidth(stringBounder) + 2*c.PaddingX()
	return result
}

// PreferredHeight returns the height of the text and the padding.
func (c *ComponentRoseNoteBox) PreferredHeight(stringBounder font.StringBounder) float64 {
	return c.TextHeight(stringBounder) + 2*c.PaddingY()
}

// PaddingX returns 5.
func (c *ComponentRoseNoteBox) PaddingX() float64 { return 5 }

// PaddingY returns 5.
func (c *ComponentRoseNoteBox) PaddingY() float64 { return 5 }

// DrawInternalU draws the rectangle and the text. An area narrower than
// the note panics, as upstream throws IllegalArgumentException.
func (c *ComponentRoseNoteBox) DrawInternalU(ug klimt.UGraphic, area *klimt.Area) {
	stringBounder := ug.StringBounder()
	textHeight := javaD2I(c.TextHeight(stringBounder))

	x2 := javaD2I(c.TextWidth(stringBounder))
	diffX := area.DimensionToUse().Width() - c.PreferredWidth(stringBounder)
	if diffX < 0 {
		throw(illegalArgument, "")
	}
	if area.DimensionToUse().Width() > c.PreferredWidth(stringBounder) {
		x2 = javaD2I(area.DimensionToUse().Width() - 2*c.PaddingX())
	}

	ug = c.symbolContext.Apply(ug)
	rect := klimt.URectangleBuild(float64(x2), float64(textHeight)).Rounded(c.roundCorner)
	rect.SetDeltaShadow(c.symbolContext.DeltaShadow())
	ug.Draw(rect)
	ug = ug.Apply(klimt.UStrokeSimple())

	c.TextBlock().DrawU(ug.Apply(klimt.NewUTranslate(c.OldPaddingX1()+diffX/2, c.OldPaddingY())))
}

// noteHexagonalCornersize is the width of the pointed ends of the
// hexagonal note.
const noteHexagonalCornersize = 10

// ComponentRoseNoteHexagonal is a note drawn as a hexagon ("hnote").
type ComponentRoseNoteHexagonal struct {
	klimt.AbstractTextualComponent
	symbolContext *klimt.Fashion
}

// NewComponentRoseNoteHexagonal returns a hexagonal note; colors may be nil.
func NewComponentRoseNoteHexagonal(style *klimt.Style, strings *klimt.Display, skinParam klimt.ISkinParam,
	colors *color.Colors) *ComponentRoseNoteHexagonal {
	c := &ComponentRoseNoteHexagonal{}
	c.InitAbstractTextualComponent(c, style, style, style.WrapWidth(),
		klimt.ClockwiseTopRightBottomLeftOf(4, 12, 4, 12), skinParam, strings, false)
	c.symbolContext = c.SymbolContextColors(colors)
	return c
}

// PreferredWidth returns the width of the text and the padding.
func (c *ComponentRoseNoteHexagonal) PreferredWidth(stringBounder font.StringBounder) float64 {
	result := c.TextWidth(stringBounder) + 2*c.PaddingX()
	return result
}

// PreferredHeight returns the height of the text and the padding.
func (c *ComponentRoseNoteHexagonal) PreferredHeight(stringBounder font.StringBounder) float64 {
	return c.TextHeight(stringBounder) + 2*c.PaddingY()
}

// PaddingX returns 5.
func (c *ComponentRoseNoteHexagonal) PaddingX() float64 { return 5 }

// PaddingY returns 5.
func (c *ComponentRoseNoteHexagonal) PaddingY() float64 { return 5 }

// DrawInternalU draws the hexagon and the text. An area narrower than the
// note panics, as upstream throws IllegalArgumentException.
func (c *ComponentRoseNoteHexagonal) DrawInternalU(ug klimt.UGraphic, area *klimt.Area) {
	stringBounder := ug.StringBounder()
	textHeight := javaD2I(c.TextHeight(stringBounder))

	x2 := javaD2I(c.TextWidth(stringBounder))
	diffX := area.DimensionToUse().Width() - c.PreferredWidth(stringBounder)
	if diffX < 0 {
		throw(illegalArgument, "")
	}
	if area.DimensionToUse().Width() > c.PreferredWidth(stringBounder) {
		x2 = javaD2I(area.DimensionToUse().Width() - 2*c.PaddingX())
	}

	polygon := klimt.NewUPolygon()
	polygon.AddPoint(noteHexagonalCornersize, 0)
	polygon.AddPoint(float64(x2-noteHexagonalCornersize), 0)
	polygon.AddPoint(float64(x2), float64(textHeight/2))
	polygon.AddPoint(float64(x2-noteHexagonalCornersize), float64(textHeight))
	polygon.AddPoint(noteHexagonalCornersize, float64(textHeight))
	polygon.AddPoint(0, float64(textHeight/2))
	polygon.AddPoint(noteHexagonalCornersize, 0)

	ug = c.symbolContext.Apply(ug)
	polygon.SetDeltaShadow(c.symbolContext.DeltaShadow())
	ug.Draw(polygon)
	ug = ug.Apply(klimt.UStrokeSimple())

	c.TextBlock().DrawU(ug.Apply(klimt.NewUTranslate(c.OldPaddingX1()+diffX/2, c.OldPaddingY())))
}
