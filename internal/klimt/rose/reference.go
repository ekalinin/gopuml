// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/rose/ComponentRoseReference.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/rose/ComponentRoseEnglober.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package rose

import (
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// The sizes of a reference frame.
const (
	referenceCornersize   = 10
	referenceHeightFooter = 5
	referenceXMargin      = 2
)

// ComponentRoseReference is a reference ("ref over"): a frame with a tab
// holding the header and the text below.
type ComponentRoseReference struct {
	klimt.AbstractTextualComponent
	textHeader          klimt.TextBlock
	position            geom.HorizontalAlignment
	symbolContextHeader *klimt.Fashion
	symbolContextBody   *klimt.Fashion
	roundCorner         float64
}

// NewComponentRoseReference returns a reference: the first line of
// stringsToDisplay is the header, the others the text.
func NewComponentRoseReference(style, styleHeader *klimt.Style, stringsToDisplay *klimt.Display,
	skinParam klimt.ISkinParam) *ComponentRoseReference {
	c := &ComponentRoseReference{}
	c.InitAbstractTextualComponent(c, style, style, klimt.LineBreakStrategyNone,
		klimt.ClockwiseTopRightBottomLeftOf(4, 4, 4, 4), skinParam, stringsToDisplay.SubList(1, stringsToDisplay.Size()),
		false)

	c.symbolContextHeader = styleHeader.SymbolContext(c.IHtmlColorSet(), nil)
	c.symbolContextBody = c.SymbolContext()
	c.roundCorner = c.RoundCorner()
	fcHeader := styleHeader.FontConfiguration(c.IHtmlColorSet(), nil)
	c.position = c.HorizontalAlignment()

	c.textHeader = stringsToDisplay.SubList(0, 1).Create(fcHeader, geom.HorizontalAlignmentLeft, skinParam)
	return c
}

// DrawInternalU draws the frame, the tab, the header and the text.
func (c *ComponentRoseReference) DrawInternalU(ug klimt.UGraphic, area *klimt.Area) {
	dimensionToUse := area.DimensionToUse()
	stringBounder := ug.StringBounder()
	textHeaderWidth := float64(javaD2I(c.headerWidth(stringBounder)))
	textHeaderHeight := float64(javaD2I(c.headerHeight(stringBounder)))

	rect := klimt.URectangleBuild(dimensionToUse.Width()-referenceXMargin*2-c.symbolContextBody.DeltaShadow(),
		dimensionToUse.Height()-referenceHeightFooter)
	if c.roundCorner != 0 {
		rect = rect.Rounded(c.roundCorner)
	}

	rect.SetDeltaShadow(c.symbolContextBody.DeltaShadow())
	ug = c.symbolContextBody.Apply(ug)
	ug.Apply(klimt.UTranslateDx(referenceXMargin)).Draw(rect)

	corner := klimt.UPathNone()
	if c.roundCorner == 0 {
		corner.MoveTo(0, 0)
		corner.LineTo(textHeaderWidth, 0)

		corner.LineTo(textHeaderWidth, textHeaderHeight-referenceCornersize)
		corner.LineTo(textHeaderWidth-referenceCornersize, textHeaderHeight)

		corner.LineTo(0, textHeaderHeight)
		corner.LineTo(0, 0)
	} else {
		corner.MoveTo(c.roundCorner/2, 0)
		corner.LineTo(textHeaderWidth, 0)

		corner.LineTo(textHeaderWidth, textHeaderHeight-referenceCornersize)
		corner.LineTo(textHeaderWidth-referenceCornersize, textHeaderHeight)

		corner.LineTo(0, textHeaderHeight)
		corner.LineTo(0, c.roundCorner/2)

		corner.ArcTo(c.roundCorner/2, c.roundCorner/2, 0, 0, 1, c.roundCorner/2, 0)
	}

	ug = c.symbolContextHeader.Apply(ug)
	ug.Apply(klimt.UTranslateDx(referenceXMargin)).Draw(corner)

	ug = ug.Apply(klimt.UStrokeSimple())

	c.textHeader.DrawU(ug.Apply(klimt.NewUTranslate(15, 2)))
	var textPos float64
	switch c.position {
	case geom.HorizontalAlignmentCenter:
		textWidth := c.TextBlock().CalculateDimension(stringBounder).Width()
		textPos = (dimensionToUse.Width() - textWidth) / 2
	case geom.HorizontalAlignmentRight:
		textWidth := c.TextBlock().CalculateDimension(stringBounder).Width()
		textPos = dimensionToUse.Width() - textWidth - c.OldPaddingX2() - referenceXMargin
	default:
		textPos = c.OldPaddingX1() + referenceXMargin
	}
	c.TextBlock().DrawU(ug.Apply(klimt.NewUTranslate(textPos, c.OldPaddingY()+textHeaderHeight)))
}

func (c *ComponentRoseReference) headerHeight(stringBounder font.StringBounder) float64 {
	headerDim := c.textHeader.CalculateDimension(stringBounder)
	return headerDim.Height() + 2*1
}

func (c *ComponentRoseReference) headerWidth(stringBounder font.StringBounder) float64 {
	headerDim := c.textHeader.CalculateDimension(stringBounder)
	return headerDim.Width() + 30 + 15
}

// PreferredHeight returns the height of the text, the header and the
// footer.
func (c *ComponentRoseReference) PreferredHeight(stringBounder font.StringBounder) float64 {
	return c.TextHeight(stringBounder) + c.headerHeight(stringBounder) + referenceHeightFooter
}

// PreferredWidth returns the width of the text or of the header, the
// margins and the shadow.
func (c *ComponentRoseReference) PreferredWidth(stringBounder font.StringBounder) float64 {
	return jmath.Max(c.TextWidth(stringBounder), c.headerWidth(stringBounder)) + referenceXMargin*2 +
		c.symbolContextBody.DeltaShadow()
}

// ComponentRoseEnglober is a box around participants ("box ... end box"),
// drawn in the background with its title.
type ComponentRoseEnglober struct {
	klimt.AbstractTextualComponent
	symbolContext *klimt.Fashion
	roundCorner   float64
}

// NewComponentRoseEnglober returns the box of participants with its title.
func NewComponentRoseEnglober(style *klimt.Style, strings *klimt.Display,
	skinParam klimt.ISkinParam) *ComponentRoseEnglober {
	c := &ComponentRoseEnglober{}
	c.InitAbstractTextualComponent(c, style, style, klimt.LineBreakStrategyNone,
		klimt.ClockwiseTopRightBottomLeftOf(1, 3, 1, 3), skinParam, strings, false)
	c.roundCorner = c.RoundCorner()
	c.symbolContext = c.SymbolContext()
	return c
}

// DrawBackgroundInternalU draws the box and the centered title.
func (c *ComponentRoseEnglober) DrawBackgroundInternalU(ug klimt.UGraphic, area *klimt.Area) {
	dimensionToUse := area.DimensionToUse()
	ug = c.symbolContext.Apply(ug)
	ug.Draw(klimt.URectangleBuild(dimensionToUse.Width(), dimensionToUse.Height()).Rounded(c.roundCorner))
	xpos := (dimensionToUse.Width() - c.PureTextWidth(ug.StringBounder())) / 2
	c.TextBlock().DrawU(ug.Apply(klimt.UTranslateDx(xpos)))
}

// DrawInternalU draws nothing: the box is drawn in the background.
func (c *ComponentRoseEnglober) DrawInternalU(klimt.UGraphic, *klimt.Area) {}

// PreferredHeight returns the height of the title and 3.
func (c *ComponentRoseEnglober) PreferredHeight(stringBounder font.StringBounder) float64 {
	return c.TextHeight(stringBounder) + 3
}

// PreferredWidth returns the width of the title.
func (c *ComponentRoseEnglober) PreferredWidth(stringBounder font.StringBounder) float64 {
	return c.TextWidth(stringBounder)
}
