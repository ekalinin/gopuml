// SPDX-License-Identifier: MIT
// Provenance: original

// Package sequencediagram parses, lays out (Teoz engine) and draws sequence
// diagrams.
//
// Upstream's sequencediagram, sequencediagram.teoz and the classes of
// sequencediagram.graphic that Teoz uses import each other, so they form
// this one package (design D2); the files of teoz and graphic carry the
// prefixes "teoz_" and "graphic_". The constraint solver of upstream's real
// package, which only Teoz uses, is in real.go.
//
// Drawing goes through SequenceDiagram.ExportDiagram: a
// SequenceDiagramFileMakerTeoz lays out the diagram (living spaces, tiles,
// X and Y lines solved once), and uml.ExportTitled adds the annotations and
// writes the page with internal/export. "!pragma teoz" with any value keeps
// the Teoz layout, the only one of upstream 1.2026.8.
//
// The text formats (ATXT, UTXT) go through SequenceDiagram.exportTxt, as
// upstream: a Teoz layout measured with asciiart.TextStringBounder, then
// the text drawing of teoz_ascii*.go (a separate X line in characters,
// tiles stacked by row) on an asciiverse.InfinitePlan, with the title
// above it (uml.DiagramChromeCreateASCII). As upstream, every page is
// drawn whatever the page index, and the tiles without a text drawing
// (delays, dividers, "|||", references, notes over participants, notes
// above or below messages) panic with UnsupportedOperationException, which
// exportTxt returns as an error (upstream writes a crash report).
//
// Not ported yet: the XMI output, and the drawing of links between anchored
// messages (LinkAnchor.drawAnchor, which needs activitydiagram3's Snake).
package sequencediagram
