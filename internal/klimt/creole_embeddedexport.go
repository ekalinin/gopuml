// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/EmbeddedDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

// EmbeddedDiagramCreateAndSkip is upstream's EmbeddedDiagram.createAndSkip
// for the packages outside klimt (the member areas of class bodies): the
// lines from start up to the matching "}}" make the diagram of type typ. It
// returns the diagram and the index of the closing line. Without an
// EmbeddedDiagramMaker in skinParam it panics with "not ported yet".
func EmbeddedDiagramCreateAndSkip(typ string, lines []CharSequence, start int, skinParam ISkinSimple) (TextBlock,
	int) {
	return embeddedDiagramCreateAndSkip(typ, lines, start, skinParam)
}
