// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/atmp/CucaDiagram.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/UgDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package cheneer

import (
	"context"
	"io"

	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/svek"
	"github.com/ekalinin/gopuml/internal/uml"
)

// ExportDiagram draws the diagram, as ExportDiagramContext without
// cancellation.
func (d *ChenEerDiagram) ExportDiagram(w io.Writer, index int,
	fileFormatOption export.FileFormatOption) (export.ImageData, error) {
	return d.ExportDiagramContext(context.Background(), w, index, fileFormatOption)
}

// ExportDiagramContext draws the diagram in an output format, as
// upstream's UgDiagram.exportDiagram: the graph laid out by Graphviz dot
// (internal/svek) with the annotations around it. The dot executable and
// its timeout come from the render context of the diagram; cancelling ctx
// stops dot and returns the context's error without output.
//
// A layout error (dot missing, failing or timed out) is returned together
// with the data of the image, which shows the error. The text formats and
// XMI are not ported yet.
func (d *ChenEerDiagram) ExportDiagramContext(ctx context.Context, w io.Writer, index int,
	fileFormatOption export.FileFormatOption) (export.ImageData, error) {
	return d.ExportDiagramLayout(ctx, w, index, fileFormatOption, svek.LayoutFromContext(d.RenderContext()))
}

// ExportDiagramLayout is ExportDiagramContext with a given layout engine.
func (d *ChenEerDiagram) ExportDiagramLayout(ctx context.Context, w io.Writer, index int,
	fileFormatOption export.FileFormatOption, layout svek.Layout) (export.ImageData, error) {
	format := fileFormatOption.FileFormat()
	switch {
	case format == font.FormatATXT || format == font.FormatUTXT:
		return nil, &export.NotImplementedError{Format: format, What: "CucaDiagramTxtMaker"}
	case format.IsXmi():
		return nil, &export.NotImplementedError{Format: format, What: "CucaDiagramXmiMaker"}
	}
	var layoutErr error
	data, err := uml.ExportTitled(w, d, d.Title(), func() klimt.TextBlock {
		tb, lerr := svek.TextBlock(ctx, d.Cuca(), d.DotStrings(), fileFormatOption, layout)
		if tb == nil {
			// The context was cancelled: nothing is written.
			panic(abortedError{lerr})
		}
		layoutErr = lerr
		return tb
	}, fileFormatOption)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, ctxErr
	}
	if err != nil {
		return nil, err
	}
	return data, layoutErr
}

// abortedError stops the drawing when the layout gives no image.
type abortedError struct{ err error }

func (e abortedError) Error() string { return e.err.Error() }
func (e abortedError) Unwrap() error { return e.err }
