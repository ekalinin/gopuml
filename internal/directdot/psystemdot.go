// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/directdot/PSystemDot.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/directdot/CounterOutputStream.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/DirectOsDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package directdot

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/dot"
	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/security"
	"github.com/ekalinin/gopuml/internal/svek"
)

// statusCrash is upstream's FileImageData.CRASH, the status of the error
// images of a dot diagram.
const statusCrash = 503

// suppressedSVG is upstream's SVG of a dot diagram in the security profiles
// that do not allow the SVG of the layout engine.
const suppressedSVG = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
	"<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\">\n" +
	"<a xlink:href=\"https://github.com/plantuml/plantuml/issues/2495\">\n" +
	"<text x=\"10\" y=\"30\" font-family=\"sans-serif\" font-size=\"14\" fill=\"blue\" text-decoration=\"underline\">This feature has been suppressed</text>\n" +
	"</a>\n" + "</svg>"

// PSystemDot is a dot diagram: a Graphviz graph that the layout engine of
// the render draws itself, as upstream's PSystemDot. It is not a
// core.TypedDiagram, as upstream's DirectOsDiagram is not a TitledDiagram.
type PSystemDot struct {
	data    strings.Builder
	seed    int64
	profile security.Profile
	layout  svek.Layout
}

func newPSystemDot(ctx *core.RenderContext, src *core.UmlSource) *PSystemDot {
	return &PSystemDot{seed: seed(src), profile: profileOf(ctx), layout: svek.LayoutFromContext(ctx)}
}

// Description implements core.Diagram.
func (*PSystemDot) Description() string { return "(Dot)" }

// Data returns the DOT source: the graph header and the lines that follow
// it, each ended by a newline.
func (d *PSystemDot) Data() string { return d.data.String() }

// ExportDiagram writes the image of the diagram, see ExportDiagramContext.
func (d *PSystemDot) ExportDiagram(w io.Writer, page int, opt export.FileFormatOption) (export.ImageData, error) {
	return d.ExportDiagramContext(context.Background(), w, page, opt)
}

// ExportDiagramContext writes the image of the diagram, as upstream's
// exportDiagramNow. In the security profiles that do not allow it, the SVG
// is upstream's suppressed-feature document, and the layout engine does not
// run. Otherwise the output of the layout engine for the format is written
// unchanged, without metadata: the SVG or PNG of the graph, or the messages
// of the engine when it gives them. A named dot that cannot be used gives
// upstream's "There is an issue with your Dot/Graphviz installation" image,
// and an engine that fails, times out or writes nothing (upstream's
// CounterOutputStream counts nothing) its "GraphViz has crashed" image, both
// with the CRASH status. Formats other than SVG and PNG are not ported.
func (d *PSystemDot) ExportDiagramContext(ctx context.Context, w io.Writer, _ int, opt export.FileFormatOption) (export.ImageData, error) {
	format := opt.FileFormat()
	if format != font.FormatSVG && format != font.FormatPNG {
		return nil, &export.NotImplementedError{Format: format, What: "PSystemDot"}
	}
	if format == font.FormatSVG && !d.profile.AllowDotSVG() {
		if _, err := io.WriteString(w, suppressedSVG); err != nil {
			return nil, err
		}
		return export.ImageDataOK(), nil
	}
	out, err := d.layout.RenderFormat(ctx, d.Data(), strings.ToLower(format.String()))
	if _, ok := errors.AsType[*dot.NotFoundError](err); ok {
		return d.crash(w, opt, "There is an issue with your Dot/Graphviz installation")
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, ctxErr
	}
	if err != nil || len(out) == 0 {
		return d.crash(w, opt, "GraphViz has crashed")
	}
	if _, err := w.Write(out); err != nil {
		return nil, err
	}
	return export.ImageDataOK(), nil
}

// crash writes upstream's error image of one line, with the CRASH status.
func (d *PSystemDot) crash(w io.Writer, opt export.FileFormatOption, line string) (export.ImageData, error) {
	tb := klimt.GraphicStringsCreateForError([]string{line}, false)
	return export.NewBuilder(tb, opt, false).Seed(d.seed).Status(statusCrash).Build().ExportTo(w)
}

// profileOf is the security profile of a render, LEGACY by default.
func profileOf(ctx *core.RenderContext) security.Profile {
	if ctx == nil {
		return security.Legacy
	}
	if ctx.Security != nil {
		return ctx.Security.Profile
	}
	p, _ := security.ParseProfile(ctx.SecurityProfile)
	return p
}

// seed is UmlSource.seed: a hash of the lines of the source.
func seed(src *core.UmlSource) int64 {
	h := int64(1125899906842597)
	for _, l := range src.Lines() {
		h = 31*h + int64(klimt.JavaStringHashCode(l.S))
		h = 31*h + '\n'
	}
	return h
}
