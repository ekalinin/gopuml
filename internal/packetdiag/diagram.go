// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/packetdiag/PacketDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package packetdiag

import (
	"slices"
	"strconv"
	"strings"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/uml"
)

// DefaultColWidth is the number of bits of a row when colwidth is not set.
const DefaultColWidth = 16

// ScaleDirection is the direction in which the bits grow.
type ScaleDirection int

// Scale directions.
const (
	LTR ScaleDirection = iota // left to right
	RTL                       // right to left
)

// PacketDiagram is a parsed packet diagram. Its fields are laid out in rows
// of colWidth bits once all lines are read (CheckFinalError), as upstream's
// build does. The bit numbers and widths are Java ints (int32), so that
// their sums overflow as upstream's do.
type PacketDiagram struct {
	uml.Base

	// colWidth is the number of bits of a row.
	colWidth int32
	// bitWidth and bitHeight are the size of one bit in pixels.
	bitWidth, bitHeight float64
	// scaleInterval is the interval of the numbered bits of the scale.
	scaleInterval int32
	// useDefaultScaleInterval is true until scale_interval is set.
	useDefaultScaleInterval bool
	// sameHeight makes the fields of a row as high as the highest one.
	sameHeight     bool
	scaleDirection ScaleDirection

	// packetItems are the fields in declaration order.
	packetItems []*packetItem
	// packetGrid are the fields per row, as they are drawn.
	packetGrid [][]*PacketBlock
	// packetIndicators are the ticks of the scale, as they are drawn.
	packetIndicators []*PacketIndicator

	style *klimt.Style
}

// NewPacketDiagram returns an empty packet diagram for src; md5map holds the
// base64 images of the source.
func NewPacketDiagram(ctx *core.RenderContext, src *core.UmlSource, md5map map[string]string) *PacketDiagram {
	return &PacketDiagram{
		Base:                    uml.NewBase(ctx, src, core.Packet, md5map),
		colWidth:                DefaultColWidth,
		bitWidth:                24,
		bitHeight:               32,
		scaleInterval:           DefaultColWidth / 2,
		useDefaultScaleInterval: true,
	}
}

var _ core.TypedDiagram = (*PacketDiagram)(nil)

// Description implements core.Diagram: "Packet Diagram".
func (d *PacketDiagram) Description() string { return "Packet Diagram" }

// CheckFinalError lays out the fields, as upstream's checkFinalError calls
// build.
func (d *PacketDiagram) CheckFinalError() string {
	d.build()
	return d.Base.CheckFinalError()
}

// ColWidth returns the number of bits of a row.
func (d *PacketDiagram) ColWidth() int32 { return d.colWidth }

// SetColWidth sets the number of bits of a row; values <= 0 are ignored.
func (d *PacketDiagram) SetColWidth(colWidth int32) {
	if colWidth > 0 {
		d.colWidth = colWidth
	}
}

// fullIndicatorInterval is the interval of the full ticks of the scale.
func (d *PacketDiagram) fullIndicatorInterval() int32 {
	if d.colWidth >= 4 {
		return d.colWidth / 4
	}
	return d.colWidth
}

// UpdateScaleInterval sets the interval of the numbered bits, limited to
// colWidth; values <= 0 are ignored.
func (d *PacketDiagram) UpdateScaleInterval(value int32) {
	if value > 0 {
		d.scaleInterval = min(value, d.colWidth)
		d.useDefaultScaleInterval = false
	}
}

// UpdateNodeHeight sets the height of a bit in pixels, 0 for negative
// values.
func (d *PacketDiagram) UpdateNodeHeight(nodeHeight int32) {
	d.bitHeight = float64(max(nodeHeight, 0))
}

// SetScaleDirection sets the direction from "LTR" or "RTL", in any case;
// other values give LTR.
func (d *PacketDiagram) SetScaleDirection(dir string) {
	if strings.ToUpper(dir) == "RTL" {
		d.scaleDirection = RTL
	} else {
		d.scaleDirection = LTR
	}
}

// SetSameHeight makes the fields of a row as high as the highest one.
func (d *PacketDiagram) SetSameHeight(sameHeight bool) { d.sameHeight = sameHeight }

// AddPacketItemRange adds the field of the bits start to end (inclusive),
// height bits high, with the text desc.
func (d *PacketDiagram) AddPacketItemRange(start, end, height int32, desc string, textRotation int32) {
	item := packetItemOfRange(start, end, height, desc)
	item.textRotation = textRotation
	d.packetItems = append(d.packetItems, item)
}

// Style returns the style of the diagram, computed once.
func (d *PacketDiagram) Style() *klimt.Style {
	if d.style == nil {
		d.style = klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNamePacketdiagDiagram).
			MergedStyle(d.CurrentStyleBuilder())
	}
	return d.style
}

// LastPacketEnd returns the last bit of the last field; ok is false when
// there is no field.
func (d *PacketDiagram) LastPacketEnd() (end int32, ok bool) {
	if len(d.packetItems) == 0 {
		return 0, false
	}
	return d.packetItems[len(d.packetItems)-1].bitEnd, true
}

func (d *PacketDiagram) build() {
	if d.useDefaultScaleInterval {
		d.UpdateScaleInterval(d.colWidth / 2)
	}
	d.adjustLayout()
	d.adjustColWidth()
	d.adjustIndicators()
	d.adjustBitWidth()
}

// adjustColWidth reduces colWidth to the width of the first row.
func (d *PacketDiagram) adjustColWidth() {
	if len(d.packetGrid) > 0 {
		var maxWidth int32
		for _, b := range d.packetGrid[0] {
			maxWidth += b.Width()
		}
		d.colWidth = min(maxWidth, d.colWidth)
	}
}

// javaMod is Java's int %, which throws ArithmeticException for a zero
// divisor.
func javaMod(a, b int32) int32 {
	if b == 0 {
		uml.Throw(uml.ArithmeticException, "/ by zero")
	}
	return a % b
}

// adjustIndicators creates the ticks of the scale, from 0 to colWidth (from
// colWidth to 0 right to left), as upstream's IntStream.
func (d *PacketDiagram) adjustIndicators() {
	fullLengthInterval := d.fullIndicatorInterval()
	limit := int64(d.colWidth + 1)
	if limit < 0 {
		// Stream.limit
		uml.Throw(uml.IllegalArgumentException, strconv.FormatInt(limit, 10))
	}
	first, step := int32(0), int32(1)
	if d.scaleDirection == RTL {
		first, step = d.colWidth, -1
	}
	d.packetIndicators = nil
	i := first
	for range limit {
		full := javaMod(i, fullLengthInterval) == 0
		numbered := javaMod(i, d.scaleInterval) == 0
		d.packetIndicators = append(d.packetIndicators, NewPacketIndicator(full, numbered, i, d.Style(), d.SkinParam()))
		i += step
	}
}

// adjustLayout puts the fields into rows, splitting the fields that go out
// of their row.
func (d *PacketDiagram) adjustLayout() {
	var grid [][]*PacketBlock
	var currRow []*PacketBlock
	remainRowWidth := d.colWidth
	for _, packet := range d.packetItems {
		if p := d.fitPacketInRow(packet, remainRowWidth, &currRow); p != nil {
			remainRowWidth = d.colWidth
			grid = append(grid, currRow)
			currRow = nil
			rowCnt := p.width / remainRowWidth
			remain := p.width % remainRowWidth
			for range rowCnt {
				grid = append(grid, []*PacketBlock{p.toPacketBlockSplit(remainRowWidth, true, true, d.SkinParam())})
			}
			if remain > 0 {
				currRow = append(currRow, p.toPacketBlockSplit(remainRowWidth, true, false, d.SkinParam()))
				remainRowWidth -= remain
			} else {
				grid[len(grid)-1][0].OpenRight(false)
			}
		} else {
			remainRowWidth -= packet.width
			if remainRowWidth == 0 {
				remainRowWidth = d.colWidth
				grid = append(grid, currRow)
				currRow = nil
			}
		}
	}
	if len(currRow) > 0 {
		grid = append(grid, currRow)
	}
	for _, blocks := range grid {
		if d.sameHeight {
			maxH := int32(1)
			for i, b := range blocks {
				if i == 0 || b.Height() > maxH {
					maxH = b.Height()
				}
			}
			for _, b := range blocks {
				b.SetHeight(maxH)
			}
		}
		if d.scaleDirection == RTL {
			slices.Reverse(blocks)
		}
	}
	d.packetGrid = grid
}

// adjustBitWidth sets the width of a bit from the font size.
func (d *PacketDiagram) adjustBitWidth() {
	const bitScale = 3.0
	d.bitWidth = d.Style().Value(klimt.PNameFontSize).AsDouble() * bitScale
}

// fitPacketInRow adds packet, or its part that fits, to row. It returns the
// rest of the field when it does not fit, nil otherwise.
func (d *PacketDiagram) fitPacketInRow(packet *packetItem, remainRowWidth int32, row *[]*PacketBlock) *packetItem {
	if overflow := packet.width - remainRowWidth; overflow > 0 {
		margin := packet.width - remainRowWidth
		bitEnd := packet.bitEnd - margin
		*row = append(*row, packet.toPacketBlockSplit(remainRowWidth, false, true, d.SkinParam()))
		return &packetItem{width: margin, height: packet.height, bitStart: bitEnd, bitEnd: packet.bitEnd, desc: packet.desc}
	}
	*row = append(*row, packet.toPacketBlock(d.SkinParam()))
	return nil
}

// packetItem is a declared field.
type packetItem struct {
	width, height    int32
	bitStart, bitEnd int32
	desc             string
	// textRotation is parsed, but not drawn, as upstream.
	textRotation int32
}

func packetItemOfRange(start, end, height int32, desc string) *packetItem {
	return &packetItem{width: end - start + 1, height: height, bitStart: start, bitEnd: end, desc: desc}
}

func (p *packetItem) toPacketBlock(skinParam klimt.ISkinParam) *PacketBlock {
	return NewPacketBlock(p.width, p.height, p.desc, skinParam, false, false)
}

func (p *packetItem) toPacketBlockSplit(newWidth int32, openLeft, openRight bool, skinParam klimt.ISkinParam) *PacketBlock {
	return NewPacketBlock(min(newWidth, p.width), p.height, p.desc, skinParam, openLeft, openRight)
}
