// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/FileFormat.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package font

import (
	"strconv"
	"strings"
)

// FileFormat is an output format, with upstream's values. It lives here
// because every StringBounder reports the format it measures for.
type FileFormat int

// The output formats, in upstream's order.
const (
	FormatEPS FileFormat = iota
	FormatEPSText
	FormatATXT
	FormatUTXT
	FormatXMIStandard
	FormatXMIStar
	FormatXMIArgo
	FormatXMICustom
	FormatXMIScript
	FormatSCXML
	FormatGRAPHML
	FormatPDF
	FormatHTML
	FormatHTML5
	FormatVDX
	FormatLATEX
	FormatLATEXNoPreamble
	FormatLATEXDeterministic
	FormatBASE64
	FormatBraillePNG
	FormatOBFUSCATE
	FormatDEBUG
	FormatNULL
	FormatPREPROC
	FormatPNG
	FormatPNGEmpty
	FormatRAW
	FormatSVGDeterministic
	FormatSVG
)

// fileFormatInfo holds upstream's enum name, the format name used on the
// command line and the MIME type.
type fileFormatInfo struct {
	enumName, name, mimeType string
}

var fileFormats = [...]fileFormatInfo{
	FormatEPS:                {"EPS", "eps", "application/postscript"},
	FormatEPSText:            {"EPS_TEXT", "eps-text", "application/postscript"},
	FormatATXT:               {"ATXT", "txt", "text/plain"},
	FormatUTXT:               {"UTXT", "utxt", "text/plain;charset=UTF-8"},
	FormatXMIStandard:        {"XMI_STANDARD", "xmi", "application/vnd.xmi+xml"},
	FormatXMIStar:            {"XMI_STAR", "xmi-star", "application/vnd.xmi+xml"},
	FormatXMIArgo:            {"XMI_ARGO", "xmi-argo", "application/vnd.xmi+xml"},
	FormatXMICustom:          {"XMI_CUSTOM", "xmi-custom", "application/vnd.xmi+xml"},
	FormatXMIScript:          {"XMI_SCRIPT", "xmi-script", "application/vnd.xmi+xml"},
	FormatSCXML:              {"SCXML", "scxml", "application/scxml+xml"},
	FormatGRAPHML:            {"GRAPHML", "graphml", "application/graphml+xml"},
	FormatPDF:                {"PDF", "pdf", "application/pdf"},
	FormatHTML:               {"HTML", "html", "text/html"},
	FormatHTML5:              {"HTML5", "html5", "text/html"},
	FormatVDX:                {"VDX", "vdx", "application/vnd.visio.xml"},
	FormatLATEX:              {"LATEX", "eps", "application/x-latex"},
	FormatLATEXNoPreamble:    {"LATEX_NO_PREAMBLE", "eps-no-preamble", "application/x-latex"},
	FormatLATEXDeterministic: {"LATEX_DETERMINISTIC", "eps", "application/x-latex"},
	FormatBASE64:             {"BASE64", "base64", "text/plain; charset=x-user-defined"},
	FormatBraillePNG:         {"BRAILLE_PNG", "braille-png", "image/png"},
	FormatOBFUSCATE:          {"OBFUSCATE", "obfuscate", "text/plain"},
	FormatDEBUG:              {"DEBUG", "debug", "text/plain"},
	FormatNULL:               {"NULL", "null", "text/plain"},
	FormatPREPROC:            {"PREPROC", "preproc", "text/plain"},
	FormatPNG:                {"PNG", "png", "image/png"},
	FormatPNGEmpty:           {"PNG_EMPTY", "png-empty", "image/png"},
	FormatRAW:                {"RAW", "raw", "image/raw"},
	FormatSVGDeterministic:   {"SVG_DETERMINISTIC", "svg", "image/svg+xml"},
	FormatSVG:                {"SVG", "svg", "image/svg+xml"},
}

func (f FileFormat) info() (fileFormatInfo, bool) {
	if f < 0 || int(f) >= len(fileFormats) {
		return fileFormatInfo{}, false
	}
	return fileFormats[f], true
}

// String returns upstream's enum name, for example "SVG" or "EPS_TEXT".
func (f FileFormat) String() string {
	if info, ok := f.info(); ok {
		return info.enumName
	}
	return "FileFormat(" + strconv.Itoa(int(f)) + ")"
}

// MimeType returns the MIME type of the format.
func (f FileFormat) MimeType() string {
	info, _ := f.info()
	return info.mimeType
}

// FileFormatFromCli returns the format of a command-line format name such
// as "svg". Several formats share a name; as in upstream, which keeps the
// last one put in its map, the last of them in enum order wins.
func FileFormatFromCli(formatName string) (FileFormat, bool) {
	for f := len(fileFormats) - 1; f >= 0; f-- {
		if fileFormats[f].name == formatName {
			return FileFormat(f), true
		}
	}
	return 0, false
}

// FileSuffix returns the file name suffix of the format, starting with a
// dot.
func (f FileFormat) FileSuffix() string {
	switch {
	case f == FormatXMICustom:
		return ".xmi_custom"
	case f.IsXmi():
		return ".xmi"
	case f == FormatLATEX || f == FormatLATEXNoPreamble || f == FormatLATEXDeterministic:
		return ".tex"
	case f == FormatBraillePNG:
		return ".braille.png"
	case f == FormatSVGDeterministic:
		return ".svg"
	case f == FormatEPSText:
		return FormatEPS.FileSuffix()
	}
	return "." + strings.ToLower(f.String())
}

// IsEps reports whether the format is Encapsulated PostScript.
func (f FileFormat) IsEps() bool { return f == FormatEPS || f == FormatEPSText }

// IsXmi reports whether the format is one of the XMI formats.
func (f FileFormat) IsXmi() bool { return strings.HasPrefix(f.String(), "XMI") }

// DoesSupportMetadata reports whether the format can embed the diagram
// source: PNG and SVG.
func (f FileFormat) DoesSupportMetadata() bool { return f == FormatPNG || f == FormatSVG }

// JavaDimension returns the size of text drawn with font, as upstream's
// FileFormat.getJavaDimension measures it with Java2D: the bounds of
// FontMetrics.getStringBounds for the font Java uses for the text, and 0x0
// for an empty text.
func JavaDimension(font UFont, text string) Dimension {
	if text == "" {
		return Dimension{}
	}
	return font.UnderlyingFont(text).StringBounds(text)
}
