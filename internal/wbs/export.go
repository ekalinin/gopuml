// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/wbs/WBSDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package wbs

import (
	"io"

	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// ExportDiagram draws page index of the diagram in an output format, as
// upstream's UgDiagram.exportDiagram: the tree and the links, with the
// annotations around them (title, header, footer, caption, legend, main
// frame). A panic while drawing is returned as an error.
func (d *WBSDiagram) ExportDiagram(w io.Writer, index int,
	fileFormatOption export.FileFormatOption) (export.ImageData, error) {
	return uml.ExportTitled(w, d, d.Title(), func() klimt.TextBlock {
		return d.TextBlock(index, fileFormatOption)
	}, fileFormatOption)
}

// TextBlock is upstream's getTextBlock, which returns the diagram itself.
func (d *WBSDiagram) TextBlock(int, export.FileFormatOption) klimt.TextBlock {
	return &wbsDiagramTextBlock{d: d}
}

// wbsDiagramTextBlock is the diagram as a text block: the tree with a
// margin of 10.
type wbsDiagramTextBlock struct {
	klimt.TextBlockDefaults
	d *WBSDiagram
}

func (t *wbsDiagramTextBlock) DrawU(ug klimt.UGraphic) {
	t.d.drawMe(ug.Apply(klimt.NewUTranslate(10, 10)))
}

func (t *wbsDiagramTextBlock) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return t.d.drawingElement().CalculateDimension(stringBounder).Delta(20)
}

// drawMe draws the tree, then, on a graphic that records the positions of
// the nodes, the links, with the translation of the graphic undone.
func (d *WBSDiagram) drawMe(ug klimt.UGraphic) {
	translate, ok := translateOf(ug)
	f := d.drawingElement()
	f.DrawU(ug)
	if !ok {
		return
	}
	ug = ug.Apply(translate.Reverse())
	for _, link := range d.links {
		link.DrawU(ug)
	}
}

// drawingElement is a new drawing of the tree, as upstream builds one on
// each call.
func (d *WBSDiagram) drawingElement() *fork { return newFork(d.SkinParam(), d.root) }
