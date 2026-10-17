// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/UgDiagram.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/TitledDiagram.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/core/AbstractDiagram.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/core/UmlSource.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package uml

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/version"
)

// TitledDiagram is a diagram drawn through ExportTitled: it embeds Base and
// may override the default margins (upstream's TitledDiagram subclasses).
type TitledDiagram interface {
	Titled() *Base
	DefaultMargins() klimt.ClockwiseTopRightBottomLeft
}

// ExportTitled writes the drawing raw of page index of d, as upstream's
// UgDiagram.exportDiagram and getExporter for the graphic formats: raw is
// wrapped with the annotations (title is the title of the page), the
// background color, then exported with the margins, scale, metadata and
// color mapper of the diagram. A panic while drawing is returned as an
// error, as upstream catches it to write a crash report.
func ExportTitled(w io.Writer, d TitledDiagram, title DisplayPositioned, raw func() klimt.TextBlock,
	fileFormatOption export.FileFormatOption) (data export.ImageData, err error) {
	defer func() {
		if r := recover(); r != nil {
			data = nil
			if e, ok := r.(error); ok {
				err = fmt.Errorf("drawing crashed: %w", e)
			} else {
				err = fmt.Errorf("drawing crashed: %v", r)
			}
		}
	}()
	b := d.Titled()
	result := raw()
	// computeStatus: 0 for the diagrams drawn here
	const status = 0

	result = DiagramChromeCreate(result, b, b.SkinParam(), b.chromeWarnings(), title)

	if backColor := b.CalculateBackColor(); backColor != nil {
		result = klimt.TextBlockUtilsAddBackcolor(result, backColor)
	}

	mutedMapper := b.MuteColorMapper(fileFormatOption.ColorMapper())
	effectiveFormat := fileFormatOption.WithColorMapper(mutedMapper)

	builder := export.NewBuilder(result, effectiveFormat, b.IsHandwritten())
	builder.Styled(styledDiagram{d: d})
	builder.Status(status)
	return builder.Build().ExportTo(w)
}

// chromeWarnings is TitledDiagram.getWarnings without duplicates. The
// warnings of the preprocessor (!option) do not reach the diagram yet.
func (b *Base) chromeWarnings() []*klimt.Warning {
	var result []*klimt.Warning
	for _, w := range b.Warnings() {
		dup := false
		for _, r := range result {
			if r.Equals(w) {
				dup = true
				break
			}
		}
		if !dup {
			result = append(result, w)
		}
	}
	return result
}

// CalculateBackColor returns the background color of the document style of
// the diagram type, transparent when it has none.
func (b *Base) CalculateBackColor() color.HColor {
	style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameDocument,
		DiagramTypeStyleName(b.DiagramType())).MergedStyle(b.SkinParam().CurrentStyleBuilder())
	backgroundColor := style.Value(klimt.PNameBackGroundColor).AsColor(b.SkinParam().IHtmlColorSet())
	if backgroundColor == nil {
		backgroundColor = color.Transparent()
	}
	return backgroundColor
}

// MuteColorMapper returns the color mapper of the "monochrome" and
// "reversecolor" skin parameters and of the dark mode, else init.
func (b *Base) MuteColorMapper(init *color.ColorMapper) *color.ColorMapper {
	if klimt.SkinParamIsDark(b.SkinParam()) {
		return color.DarkModeMapper
	}
	monochrome, _ := b.Skin().Value("monochrome")
	if monochrome == "true" {
		return color.MonochromeMapper
	}
	if monochrome == "reverse" {
		return color.MonochromeReverseMapper
	}
	reversecolor, ok := b.Skin().Value("reversecolor")
	if !ok {
		return init
	}
	if EqualsIgnoreCase("dark", reversecolor) {
		return color.LigthnessInverseMapper
	}
	order, ok := color.ColorOrderFromString(reversecolor)
	if !ok {
		return init
	}
	return color.ReverseMapper(order)
}

// Metadata returns the raw source, the preprocessed source when it
// differs, and the version, as upstream's AbstractDiagram.getMetadata.
func (b *Base) Metadata() string {
	if b.source == nil {
		return version.Upstream
	}
	rawString := joinSourceLines(b.source.Raw())
	plainString := joinSourceLines(b.source.Lines())
	if rawString == plainString {
		return rawString + "\n" + version.Upstream
	}
	return rawString + "\n" + plainString + "\n" + version.Upstream
}

func joinSourceLines(lines []text.StringLocated) string {
	var sb strings.Builder
	for _, l := range lines {
		sb.WriteString(l.S)
		sb.WriteByte('\n')
	}
	return sb.String()
}

// Seed returns a hash of the source lines, which seeds the SVG ids
// (upstream's UmlSource.seed).
func (b *Base) Seed() int64 {
	if b.source == nil {
		return 42
	}
	h := int64(1125899906842597)
	for _, l := range b.source.Lines() {
		h = 31*h + int64(javaStringHash(l.S))
		h = 31*h + '\n'
	}
	return h
}

func javaStringHash(s string) int32 {
	var h int32
	for _, u := range utf16.Encode([]rune(s)) {
		h = 31*h + int32(u)
	}
	return h
}

// styledDiagram adapts a TitledDiagram to export.StyledDiagram.
type styledDiagram struct{ d TitledDiagram }

func (s styledDiagram) SkinParam() klimt.ISkinParam { return s.d.Titled().SkinParam() }

func (s styledDiagram) CalculateBackColor() color.HColor { return s.d.Titled().CalculateBackColor() }

func (s styledDiagram) DefaultMargins() klimt.ClockwiseTopRightBottomLeft {
	return s.d.DefaultMargins()
}

func (s styledDiagram) Metadata() string { return s.d.Titled().Metadata() }

func (s styledDiagram) Seed() int64 { return s.d.Titled().Seed() }

// WarningOrError is upstream's TitledDiagram.getWarningOrError, which is
// always null: the width it checks is never recorded.
func (s styledDiagram) WarningOrError() string { return "" }

func (s styledDiagram) Scale() export.Scale {
	scale := s.d.Titled().Scale()
	if scale == nil {
		return nil
	}
	return scaleAdapter{scale}
}

func (s styledDiagram) Pragma() *klimt.Pragma { return s.d.Titled().Pragma() }

func (s styledDiagram) DiagramType() core.DiagramType { return s.d.Titled().DiagramType() }

// scaleAdapter is a Scale as the exporter reads it.
type scaleAdapter struct{ s *Scale }

func (a scaleAdapter) Scale(width, height float64) float64 { return a.s.Factor(width, height) }
