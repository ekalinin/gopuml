// SPDX-License-Identifier: MIT
// Provenance: original

// Package font holds upstream's font model (UFont, UFontFace, FontStack,
// FontStyle) and the StringBounder text measurement.
//
// Upstream measures text with Java2D on the fonts installed on the host. This
// package measures it with tables dumped from the reference JVM (Java's
// logical fonts backed by the DejaVu fonts), so that widths, heights and
// descents equal upstream's on every platform without reading host fonts.
//
// Known limits of the tables: texts that need complex shaping (for example
// combining marks positioned by the font, or scripts with contextual forms)
// are measured as the sum of the code point advances, and fonts named after
// a physical DejaVu family use the coverage of the matching logical font.
package font
