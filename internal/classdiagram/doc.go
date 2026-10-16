// SPDX-License-Identifier: MIT
// Provenance: original

// Package classdiagram parses class diagrams, and object diagrams, which
// upstream also parses with its class factory: the ClassDiagram model
// (upstream's classdiagram and objectdiagram packages, which refer to each
// other and form one Go package per design D2, objectdiagram_*.go), the
// factory and its commands in upstream's order. The graph layout and the
// drawing of class diagrams are not ported yet: Export returns
// abel.ErrNotDrawable.
package classdiagram
