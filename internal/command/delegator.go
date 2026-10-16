// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/NewpagedDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package command

// Delegator is a diagram that runs the commands on another diagram, as
// upstream's NewpagedDiagram.executeCommand runs them on its last page.
// The factory calls Delegate with the execution of the command; unlike the
// commands run on the diagram itself, a panic of the command is not turned
// into an error result (NewpagedDiagram does not wrap the command in a
// ProtectedCommand), so it ends the parsing of the factory.
type Delegator interface {
	Diagram
	// Delegate runs exec on the target diagram and returns the result to
	// continue with.
	Delegate(exec func(target Diagram) Result) Result
}
