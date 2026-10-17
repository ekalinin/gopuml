// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ActivityDiagram3.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/Recentred.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"io"

	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// ExportDiagram draws page index of the diagram in an output format, as
// upstream's UgDiagram.exportDiagram: the lanes, compressed on both axes
// and recentred, with the annotations around them (title, header, footer,
// caption, legend, main frame). The formats the exporter does not support
// yet give an error wrapping export.ErrNotImplemented.
func (d *ActivityDiagram3) ExportDiagram(w io.Writer, index int,
	fileFormatOption export.FileFormatOption) (export.ImageData, error) {
	return uml.ExportTitled(w, d, d.Title(), func() klimt.TextBlock {
		return d.TextBlock(index, fileFormatOption)
	}, fileFormatOption)
}

// TextBlock is upstream's getTextBlock: the drawing of the diagram without
// annotations.
func (d *ActivityDiagram3) TextBlock(int, export.FileFormatOption) klimt.TextBlock {
	var result klimt.TextBlock = d.swimlanes
	// BUG42
	result = klimt.CompressionXorYBuilderBuild(klimt.CompressionOnX, result)
	result = klimt.CompressionXorYBuilderBuild(klimt.CompressionOnY, result)
	return newRecentred(result)
}

// recentred moves a text block so that its drawing starts at (5, 5), its
// bounds enlarged by 15.
type recentred struct {
	klimt.TextBlockDefaults
	textBlock klimt.TextBlock
	minMax    *geom.MinMax
}

func newRecentred(textBlock klimt.TextBlock) *recentred { return &recentred{textBlock: textBlock} }

func (r *recentred) DrawU(ug klimt.UGraphic) {
	minMax := r.MinMax(ug.StringBounder())
	r.textBlock.DrawU(ug.Apply(klimt.NewUTranslate(-minMax.MinX()+5, -minMax.MinY()+5)))
}

func (r *recentred) MinMax(stringBounder font.StringBounder) geom.MinMax {
	if r.minMax == nil {
		m := r.textBlock.MinMax(stringBounder).Enlarge(15, 15)
		r.minMax = &m
	}
	return *r.minMax
}

func (r *recentred) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return r.MinMax(stringBounder).Dimension()
}

func (r *recentred) Backcolor() color.HColor { return r.textBlock.Backcolor() }
