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
// Not ported yet: the text output (upstream's asciiverse methods of the
// tiles, LivingSpace and PlayingSpaceWithParticipants: AsciiBlock,
// asciiDraw, asciiDimension, asciiAddConstraints, getAsciiMinX and
// getAsciiMaxX), the XMI output, and the drawing of links between anchored
// messages (LinkAnchor.drawAnchor, which needs activitydiagram3's Snake).
package sequencediagram
