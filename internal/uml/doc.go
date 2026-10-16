// SPDX-License-Identifier: MIT
// Provenance: original

// Package uml holds what every UML diagram type shares: the diagram base
// (upstream's AbstractDiagram, UgDiagram and TitledDiagram), the annotations
// (title, caption, legend, header, footer, main frame, scale), the
// multi-line command shapes and the common commands of
// net.sourceforge.plantuml.command.CommonCommands. It does not depend on any
// diagram type: sequence, class and activity diagrams embed Base and add the
// common commands to their own command lists.
package uml
