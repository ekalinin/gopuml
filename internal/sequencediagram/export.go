// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/SequenceDiagram.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/LinkAnchor.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"fmt"
	"io"

	"github.com/ekalinin/gopuml/internal/asciiart"
	"github.com/ekalinin/gopuml/internal/asciiverse"
	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/rose"
	"github.com/ekalinin/gopuml/internal/uml"
)

// ExportDiagram draws page index of the diagram in an output format, as
// upstream's UgDiagram.exportDiagram: the Teoz drawing of the page with
// the annotations around it (title of the page, header, footer, caption,
// legend, main frame), or its text drawing for ATXT and UTXT. XMI is not
// ported yet: it gives an error wrapping export.ErrNotImplemented.
func (d *SequenceDiagram) ExportDiagram(w io.Writer, index int,
	fileFormatOption export.FileFormatOption) (export.ImageData, error) {
	format := fileFormatOption.FileFormat()
	switch {
	case format == font.FormatATXT || format == font.FormatUTXT:
		return d.exportTxt(w, index, format)
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

// exportTxt writes the text drawing of the diagram, as upstream's
// SequenceDiagram.exportTxt: a Teoz layout measured with the
// TextStringBounder of the format, its AsciiBlock with the title above it,
// drawn on an InfinitePlan and written as text. As upstream, every page is
// drawn whatever index is. A panic while drawing (upstream's
// UnsupportedOperationException for the elements without text drawing,
// such as delays, dividers and notes over participants) is returned as an
// error and nothing is written; upstream writes a crash report instead.
func (d *SequenceDiagram) exportTxt(w io.Writer, index int, fileFormat font.FileFormat) (data export.ImageData,
	err error) {
	var plan *asciiverse.InfinitePlan
	func() {
		defer func() {
			if r := recover(); r != nil {
				if e, ok := r.(error); ok {
					err = fmt.Errorf("drawing crashed: %w", e)
				} else {
					err = fmt.Errorf("drawing crashed: %v", r)
				}
			}
		}()
		// FileFormatOption.getDefaultStringBounder of a text format
		maker := NewSequenceDiagramFileMakerTeoz(d, rose.NewRose(), asciiart.NewTextStringBounder(fileFormat))
		rawBlock := maker.AsciiBlock(index)
		block := uml.DiagramChromeCreateASCII(rawBlock, d)
		plan = asciiverse.NewInfinitePlan(fileFormat)
		block.AsciiDraw(plan)
	}()
	if err != nil {
		return nil, err
	}
	if err := plan.ExportTxt(w); err != nil {
		return nil, err
	}
	return export.ImageDataOK(), nil
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
