// SPDX-License-Identifier: MIT
// Provenance: original

// Package symbol is upstream's decoration.symbol: the shapes drawn around a
// text (database, queue, node, cloud...), small as a participant head or an
// entity, or big as a container.
//
// Only the symbols that the sequence diagram skin (skin.rose) needs are
// ported: the database and the queue. The other symbols also need svek
// classes that are not ported yet (CircleInterface2 for the interface,
// RotatedEllipse and the TextBlockInEllipse footprint for the use case), so
// they are left with the registry (USymbols.fromString) for the class
// diagram step; see internal/klimt/HANDOFF.md.
package symbol
