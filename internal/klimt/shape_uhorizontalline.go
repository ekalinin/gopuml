// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/UHorizontalLine.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/Stencil.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import "github.com/ekalinin/plantuml-in-go/internal/klimt/font"

// Stencil gives the horizontal extent of a shape at each height, so that
// separators span the shape they are drawn in. Upstream defines it in
// klimt.creole.
type Stencil interface {
	StartingX(stringBounder font.StringBounder, y float64) float64
	EndingX(stringBounder font.StringBounder, y float64) float64
}

// StencilFunc is a stencil of two functions.
type StencilFunc struct {
	Starting func(stringBounder font.StringBounder, y float64) float64
	Ending   func(stringBounder font.StringBounder, y float64) float64
}

// StartingX calls Starting.
func (s StencilFunc) StartingX(stringBounder font.StringBounder, y float64) float64 {
	return s.Starting(stringBounder, y)
}

// EndingX calls Ending.
func (s StencilFunc) EndingX(stringBounder font.StringBounder, y float64) float64 {
	return s.Ending(stringBounder, y)
}

// UHorizontalLine is a separator spanning the stencil it is drawn in (see
// UGraphicStencil), possibly with a title in its middle.
type UHorizontalLine struct {
	skipAtStart      float64
	skipAtEnd        float64
	title            TextBlock
	blankTitle       bool
	style            rune
	defaultThickness float64
}

// UHorizontalLineInfinite returns a separator of a style: '=' (double), '.'
// (dotted), '-' or '_' (simple), any other non-zero rune (default
// thickness), or 0 for the stencil's default stroke.
func UHorizontalLineInfinite(defaultThickness, skipAtStart, skipAtEnd float64, style rune) *UHorizontalLine {
	return &UHorizontalLine{defaultThickness: defaultThickness, skipAtStart: skipAtStart, skipAtEnd: skipAtEnd,
		style: style}
}

// UHorizontalLineInfiniteTitle returns a separator with a title.
func UHorizontalLineInfiniteTitle(defaultThickness, skipAtStart, skipAtEnd float64, title TextBlock,
	style rune) *UHorizontalLine {
	return &UHorizontalLine{defaultThickness: defaultThickness, skipAtStart: skipAtStart, skipAtEnd: skipAtEnd,
		title: title, style: style}
}

// IsDouble reports whether the separator is a double line.
func (h *UHorizontalLine) IsDouble() bool { return h.style == '=' }

// DrawLineInternal draws the separator at height y across the stencil.
func (h *UHorizontalLine) DrawLineInternal(ug UGraphic, stencil Stencil, y float64, defaultStroke UStroke) {
	stencil = h.addSkip(stencil)
	strokeToUse := defaultStroke
	if h.style != 0 {
		strokeToUse = h.Stroke()
	}
	ugStroke := ug.Apply(strokeToUse)
	if h.title == nil {
		h.drawHLine(stencil, y, ugStroke)
		return
	}
	dimTitle := h.title.CalculateDimension(ug.StringBounder())
	h.drawHLine(firstHalf(stencil, dimTitle.Width()), y, ugStroke)
	startingX := stencil.StartingX(ug.StringBounder(), y)
	endingX := stencil.EndingX(ug.StringBounder(), y)
	h.DrawTitleInternal(ug, startingX, endingX, y, false)
	h.drawHLine(secondHalf(stencil, dimTitle.Width()), y, ugStroke)
}

func (h *UHorizontalLine) addSkip(stencil Stencil) Stencil {
	return StencilFunc{
		Starting: func(sb font.StringBounder, y float64) float64 { return stencil.StartingX(sb, y) + h.skipAtStart },
		Ending:   func(sb font.StringBounder, y float64) float64 { return stencil.EndingX(sb, y) - h.skipAtEnd },
	}
}

func firstHalf(stencil Stencil, widthTitle float64) Stencil {
	return StencilFunc{
		Starting: stencil.StartingX,
		Ending: func(sb font.StringBounder, y float64) float64 {
			start := stencil.StartingX(sb, y)
			end := stencil.EndingX(sb, y)
			l := (end - start - widthTitle) / 2
			return start + l
		},
	}
}

func secondHalf(stencil Stencil, widthTitle float64) Stencil {
	return StencilFunc{
		Starting: func(sb font.StringBounder, y float64) float64 {
			start := stencil.StartingX(sb, y)
			end := stencil.EndingX(sb, y)
			l := (end - start - widthTitle) / 2
			return end - l
		},
		Ending: stencil.EndingX,
	}
}

func (h *UHorizontalLine) drawHLine(stencil Stencil, y float64, ug UGraphic) {
	drawSimpleHline(ug, stencil, y)
	if h.style == '=' {
		drawSimpleHline(ug, stencil, y+2)
	}
}

func drawSimpleHline(ug UGraphic, stencil Stencil, y float64) {
	startingX := stencil.StartingX(ug.StringBounder(), y)
	endingX := stencil.EndingX(ug.StringBounder(), y)
	ug.Apply(NewUTranslate(startingX, y)).Draw(ULineH(endingX - startingX))
}

// DrawTitleInternal draws the title centered between startingX and endingX
// at height y, over a cleared rectangle when clearArea is set.
func (h *UHorizontalLine) DrawTitleInternal(ug UGraphic, startingX, endingX, y float64, clearArea bool) {
	if h.title == nil || h.blankTitle {
		return
	}
	widthToUse := endingX - startingX
	dimTitle := h.title.CalculateDimension(ug.StringBounder())
	space := (widthToUse - dimTitle.Width()) / 2
	x1 := startingX + space
	y1 := y - dimTitle.Height()/2 - 0.5
	ug = ug.Apply(NewUTranslate(x1, y1))
	if clearArea {
		ug.Apply(h.Stroke()).Draw(URectangleBuildDim(dimTitle))
	}
	h.title.DrawU(ug)
}

// DrawMe draws the separator as a shape: graphics that know their stencil
// (UGraphicStencil) draw it across.
func (h *UHorizontalLine) DrawMe(ug UGraphic) { ug.Draw(h) }

// Stroke returns the stroke of the style. The default style (0) panics, as
// upstream throws IllegalStateException.
func (h *UHorizontalLine) Stroke() UStroke {
	switch h.style {
	case 0:
		throw(illegalState, "")
	case '=':
		return UStrokeSimple()
	case '.':
		return NewUStroke(1, 2, 1)
	case '-':
		return UStrokeSimple()
	}
	return UStrokeWithThickness(h.defaultThickness)
}
