// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/image/CircleStart.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/image/CircleEnd.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// CircleStart is the filled circle of an initial node (upstream's
// svek.image.CircleStart): 20 x 20, the colors of the style unless colors
// override them.
type CircleStart struct {
	TextBlockDefaults
	skinParam ISkinParam
	style     *Style
	colors    *color.Colors
}

// NewCircleStart returns the start circle.
func NewCircleStart(skinParam ISkinParam, style *Style, colors *color.Colors) *CircleStart {
	return &CircleStart{skinParam: skinParam, style: style, colors: colors}
}

// CalculateDimension implements TextBlock.
func (c *CircleStart) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(20, 20)
}

// DrawU implements TextBlock.
func (c *CircleStart) DrawU(ug UGraphic) {
	circle := UEllipseBuild(20, 20)
	backColor := ColorsColorOf(c.colors, c.style, PNameBackGroundColor, c.skinParam.IHtmlColorSet())
	lineColor := ColorsColorOf(c.colors, c.style, PNameLineColor, c.skinParam.IHtmlColorSet())
	circle.SetDeltaShadow(c.style.Shadowing())
	ug.Apply(lineColor).Apply(Bg(backColor)).Apply(c.style.Stroke()).Draw(circle)
}

// CircleEnd is the bull's eye of a final node (upstream's
// svek.image.CircleEnd): 22 x 22.
type CircleEnd struct {
	TextBlockDefaults
	skinParam ISkinParam
	style     *Style
	colors    *color.Colors
}

// NewCircleEnd returns the end circle.
func NewCircleEnd(skinParam ISkinParam, style *Style, colors *color.Colors) *CircleEnd {
	return &CircleEnd{skinParam: skinParam, style: style, colors: colors}
}

// CalculateDimension implements TextBlock.
func (c *CircleEnd) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(22, 22)
}

// DrawU implements TextBlock.
func (c *CircleEnd) DrawU(ug UGraphic) {
	const size = 22
	circle := UEllipseBuild(size, size)
	backColor := ColorsColorOf(c.colors, c.style, PNameBackGroundColor, c.skinParam.IHtmlColorSet())
	lineColor := ColorsColorOf(c.colors, c.style, PNameLineColor, c.skinParam.IHtmlColorSet())
	circle.SetDeltaShadow(c.style.Shadowing())
	ug.Apply(Bg(color.None())).Apply(lineColor).Apply(c.style.Stroke()).Draw(circle)

	const delta = 5
	circleSmall := UEllipseBuild(size-delta*2, size-delta*2)
	// Because of anti-aliasing, upstream uses the middle color when the
	// line and background colors are the same
	if lineColor.Equals(backColor) {
		middle := color.Middle(lineColor, backColor)
		ug = ug.Apply(Bg(backColor))
		ug = ug.Apply(middle)
	} else {
		ug = ug.Apply(Bg(backColor))
		ug = ug.Apply(lineColor)
	}
	ug.Apply(NewUTranslate(delta, delta)).Draw(circleSmall)
}
