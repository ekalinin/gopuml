// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/FileFormatOption.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/FileFormat.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package export

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// optString is a string that may be missing (null upstream).
type optString struct {
	value string
	ok    bool
}

// FileFormatOption is an output format with its parameters. The With
// methods return a changed copy, as upstream.
type FileFormatOption struct {
	fileFormat          font.FileFormat
	withMetadata        bool
	useRedForError      bool
	svgLinkTarget       optString
	hoverColor          optString
	tikzFontDistortion  *klimt.TikzFontDistortion
	scale               float64
	preserveAspectRatio optString
	watermark           optString
	colorMapper         *color.ColorMapper
	decimal             int
	debugSvek           bool
}

// NewFileFormatOption returns the option of a format, with metadata.
func NewFileFormatOption(fileFormat font.FileFormat) FileFormatOption {
	return NewFileFormatOptionMetadata(fileFormat, true)
}

// NewFileFormatOptionMetadata returns the option of a format, with or
// without metadata: identity color mapper, scale 1, default decimals (-1).
func NewFileFormatOptionMetadata(fileFormat font.FileFormat, withMetadata bool) FileFormatOption {
	return FileFormatOption{fileFormat: fileFormat, withMetadata: withMetadata,
		tikzFontDistortion: klimt.TikzFontDistortionDefault(), scale: 1.0, colorMapper: color.IdentityMapper,
		decimal: -1}
}

// DefaultStringBounder returns the text measurer of the format, upstream's
// FileFormat.getDefaultStringBounder. charSizeHack nil means none. Formats
// whose measurer is not ported yet (SVG_DETERMINISTIC and LATEX
// width tables, TikZ, braille, DEBUG, ATXT, UTXT, PDF) give an
// ErrNotImplemented error.
func (o FileFormatOption) DefaultStringBounder(charSizeHack font.SvgCharSizeHack,
	pragma *klimt.Pragma) (font.StringBounder, error) {
	switch o.fileFormat {
	case font.FormatSVG:
		return font.NewStringBounderSvg(charSizeHack), nil
	case font.FormatLATEX, font.FormatLATEXNoPreamble, font.FormatLATEXDeterministic, font.FormatSVGDeterministic:
		return nil, notImplemented(o.fileFormat, "StringBounderFromWidthTable")
	case font.FormatBraillePNG:
		return nil, notImplemented(o.fileFormat, "StringBounderBraille")
	case font.FormatDEBUG:
		return nil, notImplemented(o.fileFormat, "StringBounderDebug")
	case font.FormatATXT, font.FormatUTXT:
		return nil, notImplemented(o.fileFormat, "TextStringBounder")
	case font.FormatPDF:
		return nil, notImplemented(o.fileFormat, "StringBounderOpenPdf")
	}
	return font.StringBounderAwt{}, nil
}

// SvgLinkTarget returns the SVG link target; ok is false when it is not set.
func (o FileFormatOption) SvgLinkTarget() (string, bool) {
	return o.svgLinkTarget.value, o.svgLinkTarget.ok
}

// IsWithMetadata reports whether the diagram source is embedded.
func (o FileFormatOption) IsWithMetadata() bool { return o.withMetadata }

// PreserveAspectRatio returns the SVG preserveAspectRatio; ok is false when
// it is not set.
func (o FileFormatOption) PreserveAspectRatio() (string, bool) {
	return o.preserveAspectRatio.value, o.preserveAspectRatio.ok
}

// WithUseRedForError returns the option with red error images.
func (o FileFormatOption) WithUseRedForError() FileFormatOption {
	o.useRedForError = true
	return o
}

// WithTikzFontDistortion returns the option with a TikZ font distortion,
// which must not be nil.
func (o FileFormatOption) WithTikzFontDistortion(tikzFontDistortion *klimt.TikzFontDistortion) FileFormatOption {
	if tikzFontDistortion == nil {
		panic(nullPointerError())
	}
	o.tikzFontDistortion = tikzFontDistortion
	return o
}

// WithSvgLinkTarget returns the option with an SVG link target.
func (o FileFormatOption) WithSvgLinkTarget(svgLinkTarget string) FileFormatOption {
	o.svgLinkTarget = optString{svgLinkTarget, true}
	return o
}

// WithPreserveAspectRatio returns the option with an SVG preserveAspectRatio.
func (o FileFormatOption) WithPreserveAspectRatio(preserveAspectRatio string) FileFormatOption {
	o.preserveAspectRatio = optString{preserveAspectRatio, true}
	return o
}

// WithHoverColor returns the option with the stroke color of hovered SVG
// paths.
func (o FileFormatOption) WithHoverColor(hoverColor string) FileFormatOption {
	o.hoverColor = optString{hoverColor, true}
	return o
}

// WithScale returns the option with a scale coefficient.
func (o FileFormatOption) WithScale(scale float64) FileFormatOption {
	o.scale = scale
	return o
}

// WithWatermark returns the option with a PNG watermark (upstream's
// withWartermark).
func (o FileFormatOption) WithWatermark(watermark string) FileFormatOption {
	o.watermark = optString{watermark, true}
	return o
}

// WithColorMapper returns the option with a color mapper, the hook of dark
// mode (color.DarkModeMapper) and monochrome output.
func (o FileFormatOption) WithColorMapper(colorMapper *color.ColorMapper) FileFormatOption {
	o.colorMapper = colorMapper
	return o
}

// WithDecimal returns the option with the number of decimals of SVG
// coordinates; a negative value keeps the default.
func (o FileFormatOption) WithDecimal(decimal int) FileFormatOption {
	o.decimal = decimal
	return o
}

// String returns the format name.
func (o FileFormatOption) String() string { return o.fileFormat.String() }

// FileFormat returns the format.
func (o FileFormatOption) FileFormat() font.FileFormat { return o.fileFormat }

// IsUseRedForError reports whether error images are red.
func (o FileFormatOption) IsUseRedForError() bool { return o.useRedForError }

// SetDebugSvek sets the svek debug flag; as upstream, it changes the option
// in place.
func (o *FileFormatOption) SetDebugSvek(debugSvek bool) { o.debugSvek = debugSvek }

// IsDebugSvek reports the svek debug flag.
func (o FileFormatOption) IsDebugSvek() bool { return o.debugSvek }

// HoverColor returns the hovered path color; ok is false when it is not set.
func (o FileFormatOption) HoverColor() (string, bool) { return o.hoverColor.value, o.hoverColor.ok }

// HideMetadata removes the metadata; as upstream, it changes the option in
// place.
func (o *FileFormatOption) HideMetadata() { o.withMetadata = false }

// TikzFontDistortion returns the TikZ font distortion.
func (o FileFormatOption) TikzFontDistortion() *klimt.TikzFontDistortion { return o.tikzFontDistortion }

// Watermark returns the watermark; ok is false when it is not set.
func (o FileFormatOption) Watermark() (string, bool) { return o.watermark.value, o.watermark.ok }

// ColorMapper returns the color mapper.
func (o FileFormatOption) ColorMapper() *color.ColorMapper { return o.colorMapper }

// Decimal returns the number of decimals, -1 for the default.
func (o FileFormatOption) Decimal() int { return o.decimal }

// ScaleCoef returns the scale coefficient.
func (o FileFormatOption) ScaleCoef() float64 { return o.scale }
