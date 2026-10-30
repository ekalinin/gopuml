// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/shape/TextBlockTitle.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/shape/TextBlockArrow.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/shape/TextBlockArrow2.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/shape/CircledCharacter.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/GuideLine.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"

	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// TextBlockTitle is a centered title with a margin of 2 on each side.
type TextBlockTitle struct {
	TextBlockDefaults
	outMargin float64
	textBlock TextBlock
}

// NewTextBlockTitle returns the title block of a display. A display of one
// empty line panics, as upstream throws IllegalArgumentException.
func NewTextBlockTitle(f *FontConfiguration, stringsToDisplay *Display, spriteContainer ISkinSimple) *TextBlockTitle {
	if stringsToDisplay.Size() == 1 && stringsToDisplay.Get(0).String() == "" {
		throw(illegalArgument, "")
	}
	lineBreak := LineBreakStrategyNone
	textBlock := stringsToDisplay.Create0(f, geom.HorizontalAlignmentCenter, spriteContainer, lineBreak,
		CreoleModeFull, nil, nil)
	return &TextBlockTitle{outMargin: 2, textBlock: textBlock}
}

// DrawU draws the title after the margin.
func (t *TextBlockTitle) DrawU(ug UGraphic) { t.textBlock.DrawU(ug.Apply(UTranslateDx(t.outMargin))) }

// CalculateDimension returns the size of the title plus the margins.
func (t *TextBlockTitle) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	textDim := t.textBlock.CalculateDimension(stringBounder)
	width := textDim.Width() + float64(t.outMargin*2)
	height := textDim.Height()
	return geom.NewXDimension2D(width, height)
}

// MinMax panics, as upstream throws UnsupportedOperationException.
func (t *TextBlockTitle) MinMax(font.StringBounder) geom.MinMax {
	throw(unsupportedOperation, "")
	return geom.MinMax{}
}

// InnerPosition returns nil.
func (t *TextBlockTitle) InnerPosition(string, font.StringBounder) *geom.XRectangle2D { return nil }

// MagneticBorder returns the border of the title, shifted by the margin.
func (t *TextBlockTitle) MagneticBorder() MagneticBorder {
	return MagneticBorderFunc(func(stringBounder font.StringBounder, position geom.XPoint2D) UTranslate {
		return t.textBlock.MagneticBorder().ForceAt(stringBounder, position.Move(-t.outMargin, 0))
	})
}

// Backcolor returns the background color of the title.
func (t *TextBlockTitle) Backcolor() color.HColor { return t.textBlock.Backcolor() }

// TextBlockArrow is a small filled triangle the size of a font, pointing in
// a direction (creole's <&...> arrows in links).
type TextBlockArrow struct {
	TextBlockDefaults
	size  float64
	arrow Direction
	color color.HColor
}

// NewTextBlockArrow returns the arrow for a font configuration.
func NewTextBlockArrow(arrow Direction, fontConfiguration *FontConfiguration) *TextBlockArrow {
	return &TextBlockArrow{arrow: arrow, size: fontConfiguration.Font().Size2D(),
		color: fontConfiguration.Color()}
}

// DrawU draws the triangle.
func (t *TextBlockArrow) DrawU(ug UGraphic) {
	ug = ug.Apply(Bg(t.color))
	ug = ug.Apply(t.color)
	triSize := javaD2I(float64(t.size*.8) - 3)
	if triSize%2 == 1 {
		triSize--
	}
	triangle := t.triangle(triSize)
	ug.Apply(NewUTranslate(2, (t.size-float64(triSize))-2)).Draw(triangle)
}

func (t *TextBlockArrow) triangle(triSize int) *UPolygon {
	triangle := NewUPolygon()
	s := float64(triSize)
	half := float64(triSize / 2)
	switch t.arrow {
	case DirectionRight:
		triangle.AddPoint(0, 0)
		triangle.AddPoint(s, half)
		triangle.AddPoint(0, s)
		triangle.AddPoint(0, 0)
	case DirectionLeft:
		triangle.AddPoint(s, 0)
		triangle.AddPoint(0, half)
		triangle.AddPoint(s, s)
		triangle.AddPoint(s, 0)
	case DirectionUp:
		triangle.AddPoint(0, s)
		triangle.AddPoint(half, 0)
		triangle.AddPoint(s, s)
		triangle.AddPoint(0, s)
	case DirectionDown:
		triangle.AddPoint(0, 0)
		triangle.AddPoint(half, s)
		triangle.AddPoint(s, 0)
		triangle.AddPoint(0, 0)
	default:
		throw(illegalState, "")
	}
	return triangle
}

// CalculateDimension returns a square of the font size.
func (t *TextBlockArrow) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(t.size, t.size)
}

// GuideLine is the direction of a link label arrow. Upstream defines it in
// svek.
type GuideLine interface {
	ArrowDirection() Direction
	ArrowDirectionInRadian() float64
}

// TextBlockArrow2 is a small triangle pointing along a guide line.
type TextBlockArrow2 struct {
	TextBlockDefaults
	size  float64
	angle GuideLine
	color color.HColor
}

// NewTextBlockArrow2 returns the arrow for a guide line and a font
// configuration.
func NewTextBlockArrow2(angle GuideLine, fontConfiguration *FontConfiguration) *TextBlockArrow2 {
	if angle == nil {
		throw(nullPointer, "")
	}
	return &TextBlockArrow2{angle: angle, size: fontConfiguration.Font().Size2D(), color: fontConfiguration.Color()}
}

// DrawU draws the triangle.
func (t *TextBlockArrow2) DrawU(ug UGraphic) {
	triSize := javaD2I(t.size * .80)
	ug = ug.Apply(t.color)
	ug = ug.Apply(Bg(t.color))
	ug = ug.Apply(NewUTranslate(float64(triSize/2), t.size/2))
	triangle := NewUPolygon()
	pi := math.Pi // a float64 variable: Java rounds Math.PI * 4 before dividing
	beta := pi * 4 / 5
	half := float64(triSize / 2)
	triangle.AddPointPt(arrow2Point(half, t.angle.ArrowDirectionInRadian()))
	triangle.AddPointPt(arrow2Point(half, t.angle.ArrowDirectionInRadian()+beta))
	triangle.AddPointPt(arrow2Point(half, t.angle.ArrowDirectionInRadian()-beta))
	triangle.AddPointPt(arrow2Point(half, t.angle.ArrowDirectionInRadian()))
	ug.Draw(triangle)
}

func arrow2Point(l, alpha float64) geom.XPoint2D {
	dx := l * jmath.Sin(alpha)
	dy := l * jmath.Cos(alpha)
	return geom.NewXPoint2D(dx, dy)
}

// CalculateDimension returns a square of the font size.
func (t *TextBlockArrow2) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(t.size, t.size)
}

// CircledCharacter is a character in a filled circle (the spot of a class).
type CircledCharacter struct {
	TextBlockDefaults
	c             string
	font          font.UFont
	spotBackColor color.HColor
	spotBorder    color.HColor
	fontColor     color.HColor
	radius        float64
}

// NewCircledCharacter returns the spot; spotBorder may be nil. The font
// color is resolved against the background color.
func NewCircledCharacter(c rune, radius float64, f font.UFont, spotBackColor, spotBorder,
	fontColor color.HColor) *CircledCharacter {
	return &CircledCharacter{c: string(c), radius: radius, font: f, spotBackColor: spotBackColor,
		spotBorder: spotBorder, fontColor: fontColor.AppropriateColor(spotBackColor)}
}

// DrawU draws the circle and the character.
func (c *CircledCharacter) DrawU(ug UGraphic) {
	if c.spotBorder != nil {
		ug = ug.Apply(c.spotBorder)
	}
	ug = ug.Apply(Bg(c.spotBackColor))
	ug.Draw(UEllipseBuild(c.radius*2, c.radius*2))
	ug = ug.Apply(c.fontColor)
	ug = ug.Apply(NewUTranslate(c.radius, c.radius))
	ug.Draw(NewUCenteredCharacter([]rune(c.c)[0], c.font))
}

// PreferredWidth returns the diameter.
func (c *CircledCharacter) PreferredWidth(font.StringBounder) float64 { return 2 * c.radius }

// PreferredHeight returns the diameter.
func (c *CircledCharacter) PreferredHeight(font.StringBounder) float64 { return 2 * c.radius }

// CalculateDimension returns the square of the diameter.
func (c *CircledCharacter) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(c.PreferredWidth(stringBounder), c.PreferredHeight(stringBounder))
}
