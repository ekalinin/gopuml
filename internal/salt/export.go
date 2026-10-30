// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/PSystemSalt.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package salt

import (
	"io"

	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// ExportDiagram draws the diagram in an output format, as upstream's
// UgDiagram.exportDiagram: the widgets, with the annotations around them
// (title, header, footer, caption, legend, main frame). A panic while
// drawing is returned as an error.
func (d *PSystemSalt) ExportDiagram(w io.Writer, index int,
	fileFormatOption export.FileFormatOption) (export.ImageData, error) {
	return uml.ExportTitled(w, d, d.Title(), func() klimt.TextBlock {
		return d.TextBlock(index, fileFormatOption)
	}, fileFormatOption)
}

// TextBlock is upstream's getTextBlock: a new tree of widgets, measured and
// drawn in two passes. When an element draws past its preferred dimension
// (an open droplist), a dry run of both passes on a limit finder grows the
// size to what they draw.
func (d *PSystemSalt) TextBlock(_ int, fileFormatOption export.FileFormatOption) klimt.TextBlock {
	salt := createElement(d.manageSprite(), d.dictionary)
	stringBounder, err := fileFormatOption.DefaultStringBounder(d.Skin(), d.Pragma())
	if err != nil {
		panic(err)
	}
	preferred := salt.preferredDimension(stringBounder, 0, 0)
	size := preferred
	if salt.mayDrawBeyondPreferredDimension() {
		size = withActualExtent(salt, preferred, stringBounder)
	}
	return &saltTextBlock{salt: salt, size: size, back: d.SkinParam().BackgroundColor()}
}

// withActualExtent is the preferred size grown to what both passes draw.
func withActualExtent(salt element, preferred geom.XDimension2D, stringBounder font.StringBounder) geom.XDimension2D {
	minmax := klimt.TextBlockUtilsMinMax(drawBothPasses{salt: salt, dim: preferred}, stringBounder, true)
	return preferred.AtLeast(minmax.MaxX(), minmax.MaxY())
}

// drawBothPasses draws the passes 0 and 1 of an element with a size.
type drawBothPasses struct {
	salt element
	dim  geom.XDimension2D
}

func (p drawBothPasses) DrawU(ug klimt.UGraphic) {
	dim := p.dim
	p.salt.drawU(ug, 0, &dim)
	p.salt.drawU(ug, 1, &dim)
}

// saltTextBlock is the text block of getTextBlock, in black, on the
// background of the skin parameters.
type saltTextBlock struct {
	klimt.TextBlockDefaults
	salt element
	size geom.XDimension2D
	back color.HColor
}

func (t *saltTextBlock) DrawU(ug klimt.UGraphic) {
	ug = ug.Apply(color.Black)
	drawBothPasses{salt: t.salt, dim: geom.NewXDimension2D(t.size.Width(), t.size.Height())}.DrawU(ug)
}

func (t *saltTextBlock) CalculateDimension(font.StringBounder) geom.XDimension2D { return t.size }

func (t *saltTextBlock) Backcolor() color.HColor { return t.back }
