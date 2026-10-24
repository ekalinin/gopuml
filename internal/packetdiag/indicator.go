// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/packetdiag/PacketIndicator.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package packetdiag

import (
	"strconv"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// Tick heights of the scale, in pixels.
const (
	VLineFull  = 32.0
	VLineShort = VLineFull / 2
)

// PacketIndicator is a tick of the bit scale: a full or short vertical line,
// with the number of the bit above it when it is numbered.
type PacketIndicator struct {
	klimt.TextBlockDefaults
	klimt.TextBlockMemoized

	full      bool
	numbered  bool
	bitNumber string
	style     *klimt.Style
	skinParam klimt.ISkinParam
}

// NewPacketIndicator returns the tick of bit bitNumber.
func NewPacketIndicator(full, numbered bool, bitNumber int32, style *klimt.Style, skinParam klimt.ISkinParam) *PacketIndicator {
	return &PacketIndicator{full: full, numbered: numbered, bitNumber: strconv.Itoa(int(bitNumber)), style: style,
		skinParam: skinParam}
}

// DrawU draws the number, centered on the tick, and the tick. The number is
// always measured, so that the ticks start at the same height.
func (p *PacketIndicator) DrawU(ug klimt.UGraphic) {
	numberTb := p.numberTb()
	numberDim := numberTb.CalculateDimension(ug.StringBounder())
	lineOffsetY, lineLength := VLineShort, VLineShort
	if p.full {
		lineOffsetY, lineLength = 0, VLineFull
	}
	if p.numbered {
		numberYOffset := VLineShort
		if p.full {
			numberYOffset = 0
		}
		numberTb.DrawU(ug.Apply(klimt.NewUTranslate(-numberDim.Width()/2, numberYOffset)))
	}
	vLineTb := p.vLineTb(lineOffsetY, lineLength)
	vLineTb.DrawU(ug.Apply(klimt.UTranslateDy(numberDim.Height())))
}

// CalculateDimension is the number above the tick.
func (p *PacketIndicator) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return p.Dimension(stringBounder, p.calculateDimensionSlow)
}

func (p *PacketIndicator) calculateDimensionSlow(stringBounder font.StringBounder) geom.XDimension2D {
	numberDim := p.numberTb().CalculateDimension(stringBounder)
	vLineHeight := VLineShort
	if p.full {
		vLineHeight = VLineFull
	}
	return numberDim.MergeTB(geom.NewXDimension2D(0, vLineHeight))
}

func (p *PacketIndicator) numberTb() klimt.TextBlock {
	fg := p.style.FontConfiguration(p.skinParam.IHtmlColorSet(), nil)
	return createNumberTextBlock(fg, p.skinParam, p.bitNumber)
}

func (p *PacketIndicator) vLineTb(y, height float64) klimt.TextBlock {
	return createVLineTextBlock(p.style, p.skinParam, y, height)
}

// createNumberTextBlock is the text of a bit number, without creole.
func createNumberTextBlock(fg *klimt.FontConfiguration, skinParam klimt.ISkinParam, s ...string) klimt.TextBlock {
	return klimt.DisplayCreate(s...).Create8(fg, geom.HorizontalAlignmentLeft, skinParam, klimt.CreoleModeNoCreole,
		klimt.LineBreakStrategyNone)
}

// createVLineTextBlock is the vertical line of a tick, from y to y + height,
// with the line color and stroke of style.
func createVLineTextBlock(style *klimt.Style, skinParam klimt.ISkinParam, y, height float64) klimt.TextBlock {
	return &vLine{
		lineColor: style.SymbolContext(skinParam.IHtmlColorSet(), nil).ForeColor(),
		stroke:    style.Stroke(),
		y:         y,
		height:    height,
	}
}

// vLine is the anonymous text block of upstream's createVLineTextBlock.
type vLine struct {
	klimt.TextBlockDefaults
	lineColor color.HColor
	stroke    klimt.UStroke
	y, height float64
}

func (v *vLine) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(v.stroke.Thickness(), v.height)
}

func (v *vLine) DrawU(ug klimt.UGraphic) {
	ug = ug.Apply(v.lineColor).Apply(v.stroke)
	newVerticalLine(v.y, v.y+v.height, nil).DrawU(ug)
}
