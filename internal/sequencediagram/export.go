// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/SequenceDiagram.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/LinkAnchor.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"io"

	"github.com/ekalinin/plantuml-in-go/internal/export"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/rose"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// ExportDiagram draws page index of the diagram in an output format, as
// upstream's UgDiagram.exportDiagram: the Teoz drawing of the page with
// the annotations around it (title of the page, header, footer, caption,
// legend, main frame). The text formats (ATXT, UTXT) and XMI are not
// ported yet: they give an error wrapping export.ErrNotImplemented.
func (d *SequenceDiagram) ExportDiagram(w io.Writer, index int,
	fileFormatOption export.FileFormatOption) (export.ImageData, error) {
	format := fileFormatOption.FileFormat()
	switch {
	case format == font.FormatATXT || format == font.FormatUTXT:
		return nil, &export.NotImplementedError{Format: format, What: "SequenceDiagram.exportTxt"}
	case format.IsXmi():
		return nil, &export.NotImplementedError{Format: format, What: "SequenceDiagramXmiMaker"}
	}
	title := d.Title()
	if index > 0 {
		title = d.TitleOfPage(index)
	}
	return uml.ExportTitled(w, d, title, func() klimt.TextBlock {
		return d.TextBlock(index, fileFormatOption)
	}, fileFormatOption)
}

// TextBlock returns the drawing of page num, without annotations, as
// upstream's getTextBlock: a new Teoz layout is made for each call.
func (d *SequenceDiagram) TextBlock(num int, fileFormatOption export.FileFormatOption) klimt.TextBlock {
	return d.sequenceDiagramPngMaker(fileFormatOption).TextBlock(num)
}

func (d *SequenceDiagram) sequenceDiagramPngMaker(fileFormatOption export.FileFormatOption) FileMaker {
	stringBounder, err := fileFormatOption.DefaultStringBounder(d.Skin(), d.Pragma())
	if err != nil {
		panic(err)
	}
	return NewSequenceDiagramFileMakerTeoz(d, rose.NewRose(), stringBounder)
}

// drawAnchor draws the link between two anchored messages. It needs
// upstream's activitydiagram3.ftile.Snake, which is not ported yet.
func (l *LinkAnchor) drawAnchor(klimt.UGraphic, Tile, Tile, klimt.ISkinParam) {
	panic(&color.RuntimeError{Class: "java.lang.UnsupportedOperationException",
		Message: "not ported yet: LinkAnchor.drawAnchor (activitydiagram3.ftile.Snake)"})
}
