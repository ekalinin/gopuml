// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/packetdiag/PacketDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package packetdiag

import (
	"io"

	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// ExportDiagram draws page index of the diagram in an output format, as
// upstream's UgDiagram.exportDiagram: the scale and the rows of fields, with
// the annotations around them (title, header, footer, caption, legend,
// main frame). A panic while drawing, such as the negative width of a
// reversed range, is returned as an error.
func (d *PacketDiagram) ExportDiagram(w io.Writer, index int,
	fileFormatOption export.FileFormatOption) (export.ImageData, error) {
	return uml.ExportTitled(w, d, d.Title(), func() klimt.TextBlock {
		return d.TextBlock(index, fileFormatOption)
	}, fileFormatOption)
}

// TextBlock is upstream's getTextBlock: the scale above the rows of fields.
func (d *PacketDiagram) TextBlock(int, export.FileFormatOption) klimt.TextBlock {
	return &packetTextBlock{d: d}
}

// packetTextBlock is the anonymous text block of upstream's getTextBlock.
type packetTextBlock struct {
	klimt.TextBlockDefaults
	d *PacketDiagram
}

func (t *packetTextBlock) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	d := t.d
	indicatorTotalWidth := 0.0
	indicatorMaxHeight := 0.0
	for _, indicator := range d.packetIndicators {
		iDim := indicator.CalculateDimension(stringBounder)
		indicatorTotalWidth = indicatorTotalWidth + d.bitWidth
		indicatorMaxHeight = max(iDim.Height(), indicatorMaxHeight)
	}
	gridMaxWidth := 0.0
	gridTotalHeight := 0.0
	for _, blocks := range d.packetGrid {
		rowWidth := 0.0
		rowHeight := 0.0
		for _, block := range blocks {
			tb := block.ShapeTextBlock(stringBounder, d.bitWidth, d.bitHeight)
			dim := tb.CalculateDimension(stringBounder)
			rowWidth += dim.Width()
			rowHeight = dim.Height()
		}
		gridMaxWidth = max(gridMaxWidth, rowWidth)
		gridTotalHeight += rowHeight
	}
	totalWidth := max(indicatorTotalWidth, gridMaxWidth)
	totalHeight := indicatorMaxHeight + gridTotalHeight
	return geom.NewXDimension2D(totalWidth, totalHeight)
}

func (t *packetTextBlock) DrawU(ug klimt.UGraphic) {
	d := t.d
	maxWidth := 0.0
	maxHeight := 0.0
	for _, indicator := range d.packetIndicators {
		indicator.DrawU(ug)
		ug = ug.Apply(klimt.NewUTranslate(d.bitWidth, 0))
		iDim := indicator.CalculateDimension(ug.StringBounder())
		maxWidth = maxWidth + d.bitWidth
		maxHeight = max(iDim.Height(), maxHeight)
	}
	ug = ug.Apply(klimt.NewUTranslate(-maxWidth, maxHeight))
	// Draw packets
	for _, blocks := range d.packetGrid {
		totalWidth := 0.0
		drawHeight := 0.0
		for _, block := range blocks {
			tb := block.ShapeTextBlock(ug.StringBounder(), d.bitWidth, d.bitHeight)
			dim := tb.CalculateDimension(ug.StringBounder())
			totalWidth += dim.Width()
			drawHeight = dim.Height()
			tb.DrawU(ug)
			ug = ug.Apply(klimt.NewUTranslate(dim.Width(), 0))
		}
		ug = ug.Apply(klimt.NewUTranslate(-totalWidth, drawHeight))
	}
}
