// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/mindmap/MindMapDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package mindmap

import (
	"io"

	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// ExportDiagram draws page index of the diagram in an output format, as
// upstream's UgDiagram.exportDiagram: the mindmaps, with the annotations
// around them (title, header, footer, caption, legend, main frame). A panic
// while drawing is returned as an error.
func (d *MindMapDiagram) ExportDiagram(w io.Writer, index int,
	fileFormatOption export.FileFormatOption) (export.ImageData, error) {
	return uml.ExportTitled(w, d, d.Title(), func() klimt.TextBlock {
		return d.TextBlock(index, fileFormatOption)
	}, fileFormatOption)
}

// TextBlock is upstream's getTextBlock: the mindmaps one below the other.
func (d *MindMapDiagram) TextBlock(int, export.FileFormatOption) klimt.TextBlock {
	return &mindmapsTextBlock{d: d}
}

// mindmapsTextBlock is the anonymous text block of upstream's getTextBlock.
type mindmapsTextBlock struct {
	klimt.TextBlockDefaults
	d *MindMapDiagram
}

func (t *mindmapsTextBlock) DrawU(ug klimt.UGraphic) {
	for _, mindmap := range t.d.mindmaps {
		mindmap.DrawU(ug)
		dim := mindmap.CalculateDimension(ug.StringBounder())
		ug = ug.Apply(klimt.NewUTranslate(0, dim.Height()))
	}
}

// CalculateDimension is 10 wider than the widest mindmap, as upstream's.
func (t *mindmapsTextBlock) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	width := 0.0
	height := 0.0
	for _, mindmap := range t.d.mindmaps {
		dim := mindmap.CalculateDimension(stringBounder)
		height += dim.Height()
		width = max(width, dim.Width())
	}
	return geom.NewXDimension2D(width+10, height)
}
