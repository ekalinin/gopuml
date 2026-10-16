// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/UgDiagram.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/core/AbstractDiagram.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/core/UmlSource.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/FileImageData.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package errdiagram

import (
	"errors"
	"io"
	"slices"
	"strings"
	"unicode/utf16"

	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/export"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/text"
	"github.com/ekalinin/plantuml-in-go/internal/version"
)

// The status codes of upstream's FileImageData, reported in the image data.
const (
	// StatusError is the status of error and unsupported-diagram images.
	StatusError = 400
	// StatusCrash is the status of a diagram whose drawing crashed.
	StatusCrash = 503
)

// ErrUnhandled is returned by Export for a diagram that is not drawn by
// this package.
var ErrUnhandled = errors.New("errdiagram: not an error, unsupported or welcome diagram")

// Source is the block of a diagram as Export reads it, for the metadata and
// the line count of upstream's UmlSource.
type Source struct {
	// Data are the preprocessed lines of the block, from the start line to
	// the end line (engine.Block.Data). nil means the Source of an
	// ErrorDiagram.
	Data []text.StringLocated
	// Raw are the lines of the block before preprocessing
	// (engine.Block.Raw). nil means the same as Data. It is ignored for the
	// error of the preprocessor, whose source has no raw lines upstream
	// (PSystemErrorPreprocessor builds it with UmlSource.create).
	Raw []text.StringLocated
}

// Handles reports whether Export draws d: a *core.ErrorDiagram, a
// *core.UnsupportedDiagram or a *Welcome.
func Handles(d core.Diagram) bool {
	switch d.(type) {
	case *core.ErrorDiagram, *core.UnsupportedDiagram, *Welcome:
		return true
	}
	return false
}

// Export draws the image of d, an error, unsupported or welcome diagram, to
// w, as upstream's UgDiagram.exportDiagram: no margin, no scale, the
// metadata when opt has metadata, and status StatusError for errors and
// unsupported diagrams, 0 for the welcome screen. Other diagrams give
// ErrUnhandled. Output formats that are not ported give an error wrapping
// export.ErrNotImplemented.
func Export(w io.Writer, d core.Diagram, src Source, opt export.FileFormatOption) (export.ImageData, error) {
	var (
		tb             klimt.TextBlock
		source         *core.UmlSource
		status         int
		warningOrError string
	)
	switch d := d.(type) {
	case *core.ErrorDiagram:
		data := src.Data
		if data == nil {
			data = d.Source
		}
		raw := src.Raw
		if raw == nil {
			raw = data
		}
		if isPreprocessorError(d) {
			raw = nil
		}
		source = newUmlSource(data, raw)
		tb = errorTextBlock(d, source)
		status = StatusError
		warningOrError = errorWarningOrError(source)
	case *core.UnsupportedDiagram:
		raw := src.Raw
		if raw == nil {
			raw = src.Data
		}
		source = newUmlSource(src.Data, raw)
		tb = unsupportedTextBlock(d)
		status = StatusError
	case *Welcome:
		source = d.source
		tb = welcomeTextBlock(geom.GraphicPositionBackgroundCornerBottomRight)
	default:
		return nil, ErrUnhandled
	}
	b := export.NewBuilder(tb, opt, false).Seed(seed(source)).WarningOrError(warningOrError).
		Margin(klimt.ClockwiseSame(0)).Status(status)
	if opt.IsWithMetadata() {
		b.Metadata(metadata(source))
	}
	return b.Build().ExportTo(w)
}

// newUmlSource is the UmlSource of a block, whose lines ending with a
// backslash are joined for @startuml blocks.
func newUmlSource(data, raw []text.StringLocated) *core.UmlSource {
	isSequence := len(data) > 0 && slices.Contains(core.FindStartTypes(data[0].S), core.Sequence)
	return core.NewUmlSource(data, raw, isSequence)
}

// joinLines is UmlSource.getPlainString and getRawString: each line
// followed by a newline.
func joinLines(lines []text.StringLocated) string {
	var sb strings.Builder
	for _, l := range lines {
		sb.WriteString(l.S)
		sb.WriteByte('\n')
	}
	return sb.String()
}

// metadata is AbstractDiagram.getMetadata: the raw source, the
// preprocessed source when it differs, and the version.
func metadata(source *core.UmlSource) string {
	rawString := joinLines(source.Raw())
	plainString := joinLines(source.Lines())
	if rawString == plainString {
		return rawString + "\n" + version.Upstream
	}
	return rawString + "\n" + plainString + "\n" + version.Upstream
}

// seed is UmlSource.seed: a hash of the lines, which seeds the SVG ids.
func seed(source *core.UmlSource) int64 {
	h := int64(1125899906842597)
	for _, l := range source.Lines() {
		h = 31*h + int64(javaStringHashCode(l.S))
		h = 31*h + '\n'
	}
	return h
}

// javaStringHashCode is Java's String.hashCode, over UTF-16 units.
func javaStringHashCode(s string) int32 {
	var h int32
	for _, u := range utf16.Encode([]rune(s)) {
		h = 31*h + int32(u)
	}
	return h
}
