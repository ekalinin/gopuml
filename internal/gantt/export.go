// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/GanttDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"io"

	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/uml"
)

// ExportDiagram draws the diagram in an output format, as upstream's
// UgDiagram.exportDiagram: the main block with the annotations around it
// (title, header, footer, caption, legend, main frame), or the table of
// the tasks for ATXT and UTXT. A panic while drawing is returned as an
// error.
func (d *GanttDiagram) ExportDiagram(w io.Writer, index int,
	fileFormatOption export.FileFormatOption) (export.ImageData, error) {
	format := fileFormatOption.FileFormat()
	if format == font.FormatATXT || format == font.FormatUTXT {
		return d.exportTxt(w)
	}
	return uml.ExportTitled(w, d, d.Title(), func() klimt.TextBlock {
		return d.TextBlock(index, fileFormatOption)
	}, fileFormatOption)
}

// TextBlock is upstream's getTextBlock: the calendar from the start of the
// project, or of "print between", to its end, and the main block laid out
// with the string bounder of the format.
func (d *GanttDiagram) TextBlock(_ int, fileFormatOption export.FileFormatOption) klimt.TextBlock {
	stringBounder, err := fileFormatOption.DefaultStringBounder(d.Skin(), d.Pragma())
	if err != nil {
		panic(err)
	}
	if start, ok := d.timeBounds.PrintStart(); !ok {
		d.initMinMax()
	} else {
		d.timeBounds.SetMinDay(start)
		d.timeBounds.SetMaxDay(d.timeBounds.PrintEnd())
	}
	timelineStyle := &TimelineStyleData{d, d.ColorSet(), d.SkinParam()}
	return newGanttDiagramMainBlock(d, timelineStyle, d.newTimeHeader(timelineStyle), stringBounder)
}
