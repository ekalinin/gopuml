// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/packetdiag/PacketBlock.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package packetdiag

import (
	"math"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// PacketBlock is a field, or the part of a field, drawn in one row: width
// bits wide and height bits high, with its text inside. A field split across
// rows has open edges where it continues, which upstream records but does
// not draw differently.
type PacketBlock struct {
	width, height int32
	label         string
	skinParam     klimt.ISkinParam
	leftOpen      bool
	rightOpen     bool
}

// openerStroke is the stroke of the open edges, unused, as upstream.
var openerStroke = klimt.NewUStroke(5, 5, 1)

// NewPacketBlock returns a block of width bits by height bits with label.
func NewPacketBlock(width, height int32, label string, skinParam klimt.ISkinParam, leftOpen, rightOpen bool) *PacketBlock {
	return &PacketBlock{width: width, height: height, label: label, skinParam: skinParam, leftOpen: leftOpen, rightOpen: rightOpen}
}

// OpenLeft sets whether the left edge continues a previous row.
func (b *PacketBlock) OpenLeft(leftOpen bool) { b.leftOpen = leftOpen }

// OpenRight sets whether the right edge continues on the next row.
func (b *PacketBlock) OpenRight(rightOpen bool) { b.rightOpen = rightOpen }

// Width returns the width in bits.
func (b *PacketBlock) Width() int32 { return b.width }

// Height returns the height in bits.
func (b *PacketBlock) Height() int32 { return b.height }

// SetHeight sets the height in bits.
func (b *PacketBlock) SetHeight(height int32) { b.height = height }

// DrawWidth returns the width in pixels for a bit of scale pixels.
func (b *PacketBlock) DrawWidth(scale float64) float64 { return float64(b.width) * scale }

// DrawHeight returns the height in pixels for a bit of scale pixels.
func (b *PacketBlock) DrawHeight(scale float64) float64 { return float64(b.height) * scale }

// Label returns the text of the block.
func (b *PacketBlock) Label() string { return b.label }

func (b *PacketBlock) style() *klimt.Style {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNamePacketdiagDiagram,
		klimt.SNameRectangle).MergedStyle(b.skinParam.CurrentStyleBuilder())
}

func (b *PacketBlock) fashion() *klimt.Fashion {
	return b.style().SymbolContext(b.skinParam.IHtmlColorSet(), nil)
}

func (b *PacketBlock) labelTextBlock(label string) klimt.TextBlock {
	return klimt.DisplayWithNewlines(b.skinParam.Pragma(), label).Create8(
		b.style().FontConfiguration(b.skinParam.IHtmlColorSet(), nil), geom.HorizontalAlignmentCenter, b.skinParam,
		klimt.CreoleModeSimpleLine, klimt.LineBreakStrategyNone)
}

// javaSubstring is Java's s.substring(0, n): the first n UTF-16 code units
// of s. A cut inside a surrogate pair gives U+FFFD, where Java keeps the
// lone surrogate.
func javaSubstring(s string, n int) string {
	return string(utf16.Decode(utf16.Encode([]rune(s))[:n]))
}

// labelTextBlockAbbr is the text of the block, shortened character by
// character and ended with "..." when it does not fit in the block.
func (b *PacketBlock) labelTextBlockAbbr(stringBounder font.StringBounder, bitWidth float64) klimt.TextBlock {
	f := b.style().UFont()
	const pad = "..."
	padDim := stringBounder.CalculateDimension(f, pad)
	labelDim := stringBounder.CalculateDimension(f, b.label)
	reqWidth := b.DrawWidth(bitWidth)
	if labelDim.Width < reqWidth {
		return b.labelTextBlock(b.label)
	}
	// Shrink label char by char, this is technically wrong when there's
	// multiple Unicode code points
	abbr := b.label
	for i := len(utf16.Encode([]rune(b.label))); i > 0; i-- {
		if labelDim.Width+padDim.Width < reqWidth {
			break
		}
		abbr = javaSubstring(b.label, i) + pad
		labelDim = stringBounder.CalculateDimension(f, abbr)
	}
	return b.labelTextBlock(abbr)
}

// ShapeTextBlock returns the drawing of the block for a bit of bitWidth by
// bitHeight pixels: its rectangle, at least as high as its text with
// vertical margins, and the text centered in it.
func (b *PacketBlock) ShapeTextBlock(stringBounder font.StringBounder, bitWidth, bitHeight float64) klimt.TextBlock {
	const vMargin = 10.0
	reqWidth := b.DrawWidth(bitWidth)
	fashion := b.fashion()
	label := b.labelTextBlockAbbr(stringBounder, bitWidth)
	labelHeight := label.CalculateDimension(stringBounder).Height()
	totalSpare := math.Max(0, b.DrawHeight(bitHeight)-labelHeight-vMargin-vMargin)
	topSpare := totalSpare / 2.0
	bottomSpare := totalSpare - topSpare
	return &blockShape{
		reqWidth:     reqWidth,
		height:       topSpare + labelHeight + bottomSpare + vMargin + vMargin,
		vMargin:      vMargin,
		fashion:      fashion,
		label:        label,
		topSpacer:    klimt.TextBlockUtilsEmpty(reqWidth, topSpare),
		bottomSpacer: klimt.TextBlockUtilsEmpty(reqWidth, bottomSpare),
	}
}

// blockShape is the anonymous text block of upstream's getShapeTextBlock.
type blockShape struct {
	klimt.TextBlockDefaults
	reqWidth, height        float64
	vMargin                 float64
	fashion                 *klimt.Fashion
	label                   klimt.TextBlock
	topSpacer, bottomSpacer klimt.TextBlock
}

func (s *blockShape) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(s.reqWidth, s.height)
}

func (s *blockShape) DrawU(ug klimt.UGraphic) {
	dim := s.CalculateDimension(ug.StringBounder())
	if dim.Width() == 0 || dim.Height() == 0 {
		return
	}
	ug = klimt.UGraphicStencilCreateDim(ug, dim)
	ug = s.fashion.Apply(ug)
	rect := klimt.URectangleBuild(dim.Width(), dim.Height())
	shape := rect.Rounded(s.fashion.RoundCorner())
	shape.SetDeltaShadow(s.fashion.DeltaShadow())
	ug.Draw(shape)
	tb := klimt.TextBlockUtilsMergeTB(s.topSpacer, s.label, geom.HorizontalAlignmentCenter)
	full := klimt.TextBlockUtilsMergeTB(tb, s.bottomSpacer, geom.HorizontalAlignmentCenter)
	full.DrawU(ug.Apply(klimt.NewUTranslate(0, s.vMargin)))
}
