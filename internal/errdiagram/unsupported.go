// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/error/PSystemUnsupported.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/SourceStringReader.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package errdiagram

import (
	"io"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/version"
)

// unsupportedTextBlock is PSystemUnsupported.getTextBlock: the text of the
// diagram in black on white, with the logo at the top right.
func unsupportedTextBlock(d *core.UnsupportedDiagram) klimt.TextBlock {
	return klimt.GraphicStringsCreateBlackOnWhiteImage(d.Lines(version.FullDescription(), version.License),
		plantumlImage(), geom.GraphicPositionBackgroundCornerTopRight)
}

// noValidStartMessage is the text of the image of a source without any
// @start/@end block.
const noValidStartMessage = "No valid @start/@end found, please check the version"

// ExportNoValidStart writes upstream's image for a source without any
// @start/@end block (SourceStringReader.noValidStartFound): one bold line,
// green on black, or black on light red with
// FileFormatOption.WithUseRedForError. The status is 0, as upstream's.
func ExportNoValidStart(w io.Writer, opt export.FileFormatOption) (export.ImageData, error) {
	tb := klimt.GraphicStringsCreateForError([]string{noValidStartMessage}, opt.IsUseRedForError())
	return export.NewBuilder(tb, opt, false).Build().ExportTo(w)
}
