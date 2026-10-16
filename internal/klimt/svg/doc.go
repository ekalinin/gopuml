// SPDX-License-Identifier: MIT
// Provenance: original

// Package svg is the SVG output driver, upstream's klimt.drawing.svg:
// UGraphicSvg dispatches shapes to the Driver*Svg drivers, which write them
// into SvgGraphics, a minimal DOM (XMLDocument, XMLNode) serialized by
// XMLWriter without any newline, as upstream.
//
// The output is byte-identical to upstream's for the same drawing
// operations, with these exceptions:
//
//   - The <?plantuml-src ...?> metadata is encoded with Go's compress/flate
//     (design D10), which compresses differently from Java's zlib: the text
//     decodes to the same source but differs from upstream's.
//   - Raster images (UImage) are PNG data URLs encoded with Go's image/png
//     (klimt.PortableImage.ToPngDataURL), not Java's ImageIO.
//
// Centered characters (the letter of class headers, activity spots and
// stereotype spots) are drawn as glyph outlines centered with upstream's
// UnusedSpace, from tables of the DejaVu fonts (see centeredchar.go): they
// cover the printable ASCII characters of the logical fonts in sizes 4 to
// 32. Other characters, fonts or sizes panic with a message that starts
// with "not ported yet:", as do:
//
//   - DriverTextAsPathSvg (texts as glyph outlines); upstream's exporter never
//     selects it.
//   - The TeaVM drivers, and SvgData, which nothing uses upstream.
//
// Settings that come from process state upstream are SvgOption fields: the
// security profile (INSECURE inlines SVG images) and whether "javascript:"
// links are kept.
package svg
