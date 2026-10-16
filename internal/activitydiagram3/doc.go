// SPDX-License-Identifier: MIT
// Provenance: original

// Package activitydiagram3 parses, lays out and draws activity diagrams in
// the new syntax, upstream's activitydiagram3 and activitydiagram3.ftile
// packages (one Go package: they form a cycle upstream; the ftile files
// carry the "ftile_" prefix).
//
// Ported so far: the commands of ActivityDiagramFactory3 in upstream's
// order (Factory), the diagram (ActivityDiagram3) and the instruction tree
// it builds (Instruction and its implementations, Branch, LinkRendering,
// PositionedNote), the lanes (Swimlane, Swimlanes without their drawing),
// the identity of the box shapes (BoxStyle) and the arrow colors of
// upstream's decoration package (Rainbow, HtmlColorAndStyle; their line
// styles are internal/decoration's LinkStyle). The parse result matches
// upstream's instruction tree, which testdata/upstream/Harness.java dumps
// for the differential tests.
//
// Not ported yet (the ftile layout and the drawing): the createFtile and
// createGtile methods of the instructions, the ftile factories and their
// decorations, Swimlanes as a text block, BoxStyle shapes, Recentred and
// the compression of the diagram. Export returns ErrNotDrawable.
package activitydiagram3
