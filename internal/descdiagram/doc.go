// SPDX-License-Identifier: MIT
// Provenance: original

// Package descdiagram parses description diagrams (use case, component and
// deployment diagrams): the DescriptionDiagram model (upstream's
// descdiagram package), the factory and its commands in upstream's order
// (upstream's descdiagram.command package, whose Labels and StringWithArrow
// the class diagram commands also use), and the drawing, laid out by
// Graphviz dot through internal/svek.
package descdiagram
