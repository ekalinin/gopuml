// SPDX-License-Identifier: MIT
// Provenance: original

// Package export writes a drawn diagram in an output format: upstream's
// core.TextBlockExporter with FileFormatOption, the Scale kinds and the
// ImageData results.
//
// A diagram package builds its fully decorated klimt.TextBlock, then:
//
//	data, err := export.NewBuilder(block, export.NewFileFormatOption(font.FormatSVG), false).
//		Styled(diagram).Build().ExportTo(w)
//
// The error image uses the same builder without a diagram (Backcolor,
// Margin, Metadata, Seed...). Dark mode is the color mapper of the format
// option (FileFormatOption.WithColorMapper(color.DarkModeMapper)).
//
// SVG and PNG output are ported (and NULL, which writes nothing). TXT and
// the other formats give an error wrapping ErrNotImplemented and naming the
// format; their graphics plug into TextBlockExporter.createUGraphic, and
// their text measurers into FileFormatOption.DefaultStringBounder, without
// changing the callers. The handwritten mode is not ported either.
package export
