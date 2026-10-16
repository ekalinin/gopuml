// SPDX-License-Identifier: MIT
// Provenance: original

// Package activitydiagram3 parses, lays out and draws activity diagrams in
// the new syntax: upstream's activitydiagram3 package with its ftile,
// ftile.vertical, ftile.vcompact and ftile.vcompact.cond subpackages (one Go
// package, as they form a cycle upstream; the ftile files carry the
// "ftile_", "ftile_vertical_", "ftile_vcompact_" and "ftile_vcompact_cond_"
// prefixes) and svek's UGraphicForSnake.
//
// The commands of ActivityDiagramFactory3 (Factory) build the diagram
// (ActivityDiagram3) and its instruction tree (Instruction and its
// implementations, Branch, LinkRendering, PositionedNote). The
// differential tests compare that tree with the one upstream builds, which
// testdata/upstream/Harness.java dumps.
//
// ExportDiagram draws the diagram as upstream's getTextBlock: Swimlanes
// builds the tiles (Ftile) of the instructions with the factory chain
// (VCompactFactory, then the FtileFactoryDelegator decorators for URLs,
// assembly, if, switch, while, repeat, fork/split/merge, notes and
// groups), computes the lane widths, and draws the tiles, their arrows
// (Snake, merged by UGraphicForSnake), the lane dividers and titles; the
// result is compressed on both axes (internal/klimt's
// CompressionXorYBuilder) and recentred, then exported with the
// annotations through uml.ExportTitled. Upstream's other layout (the Gtile
// classes behind Gtile.USE_GTILE, which is false) is not ported.
package activitydiagram3
