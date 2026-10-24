// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/rose/ComponentRoseLine.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/rose/ComponentRoseActiveLine.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/rose/ComponentRoseDelayLine.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/rose/ComponentRoseDelayText.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/rose/ComponentRoseDestroy.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/rose/ComponentRoseNewpage.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/rose/ComponentRoseDivider.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package rose

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// ComponentRoseLine is a lifeline: a vertical line with a transparent
// hover target that carries the participant name as tooltip.
type ComponentRoseLine struct {
	klimt.AbstractComponent
	color            color.HColor
	continueLine     bool
	stroke           klimt.UStroke
	stringsToDisplay *klimt.Display
}

// NewComponentRoseLine returns a lifeline; stringsToDisplay may be nil (the
// null display).
func NewComponentRoseLine(style *klimt.Style, continueLine bool, stringsToDisplay *klimt.Display,
	skinParam klimt.ISkinParam) *ComponentRoseLine {
	c := &ComponentRoseLine{}
	c.InitAbstractComponent(c, style, skinParam)
	c.color = c.ColorLine()
	c.stroke = c.Stroke()
	c.continueLine = continueLine
	// Ideally, stringsToDisplay should never be null. However, as a
	// safeguard, upstream defaults to Display.NULL when it is null.
	if stringsToDisplay == nil {
		stringsToDisplay = klimt.DisplayNULL()
	}
	c.stringsToDisplay = stringsToDisplay
	return c
}

// DrawInternalU draws the hover target and the line in a titled group.
func (c *ComponentRoseLine) DrawInternalU(ug klimt.UGraphic, area *klimt.Area) {
	dimensionToUse := area.DimensionToUse()
	ug = ug.Apply(c.color)
	ug = ug.Apply(c.stroke)

	ug.StartGroup(klimt.UGroupSingleton(klimt.UGroupTitle, c.stringsToDisplay.ToTooltipText()))
	c.drawTitleHoverTargetRect(ug, dimensionToUse.Width(), dimensionToUse.Height())

	x := javaD2I(dimensionToUse.Width() / 2)
	ug.Apply(klimt.UTranslateDx(float64(x))).Draw(klimt.ULineV(dimensionToUse.Height()))
	ug.CloseGroup()
}

// PreferredHeight returns 20.
func (c *ComponentRoseLine) PreferredHeight(font.StringBounder) float64 { return 20 }

// PreferredWidth returns 1.
func (c *ComponentRoseLine) PreferredWidth(font.StringBounder) float64 { return 1 }

func (c *ComponentRoseLine) drawTitleHoverTargetRect(ug klimt.UGraphic, width, height float64) {
	if height > 0 {
		const hoverTargetWidth = 8
		ug = ug.Apply(klimt.UStrokeWithThickness(0))
		ug = ug.Apply(color.Transparent())
		ug = ug.Apply(klimt.Bg(color.TransparentWith(color.WithFillOpacity)))
		ug = ug.Apply(klimt.UTranslateDx((width - hoverTargetWidth) / 2))
		ug.Draw(klimt.URectangleBuild(hoverTargetWidth, height))
	}
}

// ComponentRoseActiveLine is an activation box on a lifeline, closed at
// its top and bottom or open where it continues.
type ComponentRoseActiveLine struct {
	klimt.AbstractComponent
	symbolContext    *klimt.Fashion
	closeUp          bool
	closeDown        bool
	stringsToDisplay *klimt.Display
}

// NewComponentRoseActiveLine returns an activation box; stringsToDisplay
// may be nil (the null display).
func NewComponentRoseActiveLine(style *klimt.Style, closeUp, closeDown bool, stringsToDisplay *klimt.Display,
	skinParam klimt.ISkinParam) *ComponentRoseActiveLine {
	c := &ComponentRoseActiveLine{}
	c.InitAbstractComponent(c, style, skinParam)
	c.symbolContext = c.SymbolContext()
	c.closeUp = closeUp
	c.closeDown = closeDown
	// Ideally, stringsToDisplay should never be null. However, as a
	// safeguard, upstream defaults to Display.NULL when it is null.
	if stringsToDisplay == nil {
		stringsToDisplay = klimt.DisplayNULL()
	}
	c.stringsToDisplay = stringsToDisplay
	return c
}

// DrawInternalU draws the box in a titled group, with its sides only when
// it is open.
func (c *ComponentRoseActiveLine) DrawInternalU(ug klimt.UGraphic, area *klimt.Area) {
	dimensionToUse := area.DimensionToUse()
	stringBounder := ug.StringBounder()
	x := javaD2I(dimensionToUse.Width()-c.PreferredWidth(stringBounder)) / 2

	if dimensionToUse.Height() == 0 {
		return
	}

	ug.StartGroup(klimt.UGroupSingleton(klimt.UGroupTitle, c.stringsToDisplay.ToTooltipText()))

	rect := klimt.URectangleBuild(c.PreferredWidth(stringBounder), dimensionToUse.Height())
	if c.symbolContext.IsShadowing() {
		rect.SetDeltaShadow(1)
	}

	ug = ug.Apply(c.symbolContext.ForeColor())
	if c.closeUp && c.closeDown {
		ug.Apply(klimt.Bg(c.symbolContext.BackColor())).Apply(klimt.UTranslateDx(float64(x))).Draw(rect)
	} else {
		ug.Apply(klimt.Bg(c.symbolContext.BackColor())).Apply(c.symbolContext.BackColor()).
			Apply(klimt.UTranslateDx(float64(x))).Draw(rect)

		vline := klimt.ULineV(dimensionToUse.Height())
		ug.Apply(klimt.UTranslateDx(float64(x))).Draw(vline)
		ug.Apply(klimt.UTranslateDx(float64(x) + c.PreferredWidth(stringBounder))).Draw(vline)

		hline := klimt.ULineH(c.PreferredWidth(stringBounder))
		if c.closeUp {
			ug.Apply(klimt.UTranslateDx(float64(x))).Draw(hline)
		}
		if c.closeDown {
			ug.Apply(klimt.NewUTranslate(float64(x), dimensionToUse.Height())).Draw(hline)
		}
	}

	ug.CloseGroup()
}

// PreferredHeight returns 0.
func (c *ComponentRoseActiveLine) PreferredHeight(font.StringBounder) float64 { return 0 }

// PreferredWidth returns 10.
func (c *ComponentRoseActiveLine) PreferredWidth(font.StringBounder) float64 { return 10 }

// ComponentRoseDelayLine is the lifeline during a delay.
type ComponentRoseDelayLine struct {
	klimt.AbstractComponent
	color color.HColor
}

// NewComponentRoseDelayLine returns the lifeline of a delay.
func NewComponentRoseDelayLine(style *klimt.Style, skinParam klimt.ISkinParam) *ComponentRoseDelayLine {
	c := &ComponentRoseDelayLine{}
	c.InitAbstractComponent(c, style, skinParam)
	c.color = c.ColorLine()
	return c
}

// DrawInternalU draws the line without anti-aliasing.
func (c *ComponentRoseDelayLine) DrawInternalU(ug klimt.UGraphic, area *klimt.Area) {
	dimensionToUse := area.DimensionToUse()
	ug = ug.Apply(c.Stroke()).Apply(c.color)
	x := javaD2I(dimensionToUse.Width() / 2)
	ug.Apply(klimt.AntiAliasingOff).Apply(klimt.UTranslateDx(float64(x))).Draw(klimt.ULineV(dimensionToUse.Height()))
}

// PreferredHeight returns 20.
func (c *ComponentRoseDelayLine) PreferredHeight(font.StringBounder) float64 { return 20 }

// PreferredWidth returns 1.
func (c *ComponentRoseDelayLine) PreferredWidth(font.StringBounder) float64 { return 1 }

// ComponentRoseDelayText is the text of a delay ("...text...").
type ComponentRoseDelayText struct {
	klimt.AbstractTextualComponent
}

// NewComponentRoseDelayText returns the text of a delay.
func NewComponentRoseDelayText(style *klimt.Style, stringsToDisplay *klimt.Display,
	skinParam klimt.ISkinParam) *ComponentRoseDelayText {
	c := &ComponentRoseDelayText{}
	c.InitAbstractTextualComponent(c, style, style, klimt.LineBreakStrategyNone,
		klimt.ClockwiseTopRightBottomLeftOf(4, 0, 4, 0), skinParam, stringsToDisplay, false)
	return c
}

// DrawInternalU draws the text centered in the area.
func (c *ComponentRoseDelayText) DrawInternalU(ug klimt.UGraphic, area *klimt.Area) {
	dimensionToUse := area.DimensionToUse()
	textBlock := c.TextBlock()
	stringBounder := ug.StringBounder()
	textWidth := c.TextWidth(stringBounder)
	textHeight := c.TextHeight(stringBounder)

	xpos := (dimensionToUse.Width() - textWidth) / 2
	ypos := (dimensionToUse.Height() - textHeight) / 2

	textBlock.DrawU(ug.Apply(klimt.NewUTranslate(xpos, ypos+c.OldPaddingY())))
}

// PreferredHeight returns the height of the text and 20.
func (c *ComponentRoseDelayText) PreferredHeight(stringBounder font.StringBounder) float64 {
	return c.TextHeight(stringBounder) + 20
}

// PreferredWidth returns the width of the text.
func (c *ComponentRoseDelayText) PreferredWidth(stringBounder font.StringBounder) float64 {
	return c.PureTextWidth(stringBounder)
}

// destroyCrossSize is the half size of the destroy cross.
const destroyCrossSize = 9

// ComponentRoseDestroy is the cross that ends a destroyed lifeline.
type ComponentRoseDestroy struct {
	klimt.AbstractComponent
	foregroundColor color.HColor
}

// NewComponentRoseDestroy returns the destroy cross.
func NewComponentRoseDestroy(style *klimt.Style, skinParam klimt.ISkinParam) *ComponentRoseDestroy {
	c := &ComponentRoseDestroy{}
	c.InitAbstractComponent(c, style, skinParam)
	c.foregroundColor = c.ColorLine()
	return c
}

// DrawInternalU draws the cross.
func (c *ComponentRoseDestroy) DrawInternalU(ug klimt.UGraphic, _ *klimt.Area) {
	ug = ug.Apply(c.Stroke()).Apply(c.foregroundColor)

	ug.Draw(klimt.NewULine(2*destroyCrossSize, 2*destroyCrossSize))
	ug.Apply(klimt.UTranslateDy(2 * destroyCrossSize)).Draw(klimt.NewULine(2*destroyCrossSize, -2*destroyCrossSize))
}

// PreferredHeight returns 18.
func (c *ComponentRoseDestroy) PreferredHeight(font.StringBounder) float64 {
	return destroyCrossSize * 2
}

// PreferredWidth returns 18.
func (c *ComponentRoseDestroy) PreferredWidth(font.StringBounder) float64 {
	return destroyCrossSize * 2
}

// ComponentRoseNewpage is the line of a page break.
type ComponentRoseNewpage struct {
	klimt.AbstractComponent
	foregroundColor color.HColor
}

// NewComponentRoseNewpage returns the line of a page break.
func NewComponentRoseNewpage(style *klimt.Style, skinParam klimt.ISkinParam) *ComponentRoseNewpage {
	c := &ComponentRoseNewpage{}
	c.InitAbstractComponent(c, style, skinParam)
	c.foregroundColor = c.ColorLine()
	return c
}

// DrawInternalU draws the line across the area.
func (c *ComponentRoseNewpage) DrawInternalU(ug klimt.UGraphic, area *klimt.Area) {
	dimensionToUse := area.DimensionToUse()
	ug = ug.Apply(c.Stroke()).Apply(c.foregroundColor)
	ug.Draw(klimt.ULineH(dimensionToUse.Width()))
}

// PreferredHeight returns 1.
func (c *ComponentRoseNewpage) PreferredHeight(font.StringBounder) float64 { return 1 }

// PreferredWidth returns 0.
func (c *ComponentRoseNewpage) PreferredWidth(font.StringBounder) float64 { return 0 }

// ComponentRoseDivider is a divider ("== text =="): a double line across
// the diagram with the text in a box.
type ComponentRoseDivider struct {
	klimt.AbstractTextualComponent
	borderColor color.HColor
	background  color.HColor
	empty       bool
	shadow      float64
	stroke      klimt.UStroke
	roundCorner float64
}

// NewComponentRoseDivider returns a divider; an empty first line gives the
// double line alone.
func NewComponentRoseDivider(style *klimt.Style, stringsToDisplay *klimt.Display,
	skinParam klimt.ISkinParam) *ComponentRoseDivider {
	c := &ComponentRoseDivider{}
	c.InitAbstractTextualComponent(c, style, style, klimt.LineBreakStrategyNone,
		klimt.ClockwiseTopRightBottomLeftOf(4, 4, 4, 4), skinParam, stringsToDisplay, false)

	c.background = c.ColorBackGround()
	c.borderColor = c.ColorLine()
	c.stroke = c.Stroke()
	c.roundCorner = c.RoundCorner()
	c.shadow = c.Shadowing()

	c.empty = stringsToDisplay.Get(0).String() == ""
	return c
}

// DrawInternalU draws the double line and, with a text, the box and the
// text.
func (c *ComponentRoseDivider) DrawInternalU(ug klimt.UGraphic, area *klimt.Area) {
	dimensionToUse := area.DimensionToUse()

	ug = ug.Apply(klimt.Bg(c.background))
	if c.empty {
		c.drawSep(ug.Apply(klimt.UTranslateDy(dimensionToUse.Height()/2)), dimensionToUse.Width())
	} else {
		textBlock := c.TextBlock()
		stringBounder := ug.StringBounder()
		textWidth := c.TextWidth(stringBounder)
		textHeight := c.TextHeight(stringBounder)
		const deltaX = 6
		xpos := (dimensionToUse.Width() - textWidth - deltaX) / 2
		ypos := (dimensionToUse.Height() - textHeight) / 2

		c.drawSep(ug.Apply(klimt.UTranslateDy(dimensionToUse.Height()/2)), dimensionToUse.Width())

		ug = ug.Apply(c.borderColor)
		ug = ug.Apply(c.stroke)
		rect := klimt.URectangleBuild(textWidth+deltaX, textHeight).Rounded(c.roundCorner)
		rect.SetDeltaShadow(c.shadow)

		ug.Apply(klimt.NewUTranslate(xpos, ypos)).Draw(rect)
		textBlock.DrawU(ug.Apply(klimt.NewUTranslate(xpos+deltaX, ypos+c.OldPaddingY())))
	}
}

func (c *ComponentRoseDivider) drawSep(ug klimt.UGraphic, width float64) {
	ug = ug.Apply(c.background)
	c.drawRectLong(ug.Apply(klimt.UTranslateDy(-1)), width)
	c.drawDoubleLine(ug, width)
}

func (c *ComponentRoseDivider) drawRectLong(ug klimt.UGraphic, width float64) {
	rectLong := klimt.URectangleBuild(width, 3).Rounded(c.roundCorner)
	rectLong.SetDeltaShadow(c.shadow)

	ug = ug.Apply(klimt.UStrokeSimple())
	ug.Draw(rectLong)
}

func (c *ComponentRoseDivider) drawDoubleLine(ug klimt.UGraphic, width float64) {
	ug = ug.Apply(klimt.UStrokeWithThickness(c.stroke.Thickness() / 2)).Apply(c.borderColor)
	line := klimt.ULineH(width)
	ug.Apply(klimt.UTranslateDy(-1)).Draw(line)
	ug.Apply(klimt.UTranslateDy(2)).Draw(line)
}

// PreferredHeight returns the height of the text and 20.
func (c *ComponentRoseDivider) PreferredHeight(stringBounder font.StringBounder) float64 {
	return c.TextHeight(stringBounder) + 20
}

// PreferredWidth returns the width of the text and 30.
func (c *ComponentRoseDivider) PreferredWidth(stringBounder font.StringBounder) float64 {
	return c.TextWidth(stringBounder) + 30
}
