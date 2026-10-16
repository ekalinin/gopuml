// SPDX-License-Identifier: MIT
// Provenance: original

// Package symbol is upstream's decoration.symbol: the shapes drawn around a
// text (database, queue, node, cloud, folder...), small as a participant
// head or an entity (AsSmall), or big as a container (AsBig).
//
// Every symbol of upstream's USymbols registry is ported, as a package
// variable named after the upstream constant (Folder, Package, Node,
// Cloud...); FromKey maps the registry keys that internal/abel uses to
// name symbols, and PackageStyleToUSymbol, ComponentStyleToUSymbol and
// ActorStyleToUSymbol are upstream's toUSymbol methods of the styles. The
// actor-like symbols (actors, boundary, control, entity, interface), the
// person and the use cases have no container form: their AsBig panics, as
// upstream throws UnsupportedOperationException.
//
// The svek helpers that the symbols draw live in internal/klimt (design D2):
// Control, Boundary and EntityDomain, CircleInterface2 for the interface,
// and the ellipse helpers of the use case (Footprint, ContainingEllipse,
// RotatedEllipse, TextBlockInContainingEllipse). The cloud draws its
// bubbles with an emulation of java.util.Random seeded by its size, so its
// outline is the same as upstream's.
package symbol
