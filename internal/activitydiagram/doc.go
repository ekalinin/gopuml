// SPDX-License-Identifier: MIT
// Provenance: original

// Package activitydiagram parses activity diagrams in the legacy syntax
// ("(*) --> "First step""): the ActivityDiagram model (upstream's
// activitydiagram package), the factory and its commands in upstream's
// order (upstream's activitydiagram.command package and its note command),
// and the drawing, laid out by Graphviz dot through internal/svek. The
// activity diagrams of the new syntax are in internal/activitydiagram3.
package activitydiagram
