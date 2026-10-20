// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/note/CommandFactoryNoteOnEntity.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/note/CommandFactoryNoteOnLink.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package command

import (
	"math"
	"slices"
)

// InPasses makes cmd eligible for the given passes only, as the
// selectedPass of upstream's CommandFactoryNoteOnEntity and
// CommandFactoryNoteOnLink, which state diagrams set to their second and
// third passes.
func InPasses[D Diagram](cmd Command[D], passes ...ParserPass) Command[D] {
	return &inPasses[D]{Command: cmd, passes: passes}
}

type inPasses[D Diagram] struct {
	Command[D]
	passes []ParserPass
}

// EligibleFor implements Command.
func (c *inPasses[D]) EligibleFor(pass ParserPass) bool { return slices.Contains(c.passes, pass) }

// MaxLines implements MaxLinesCommand with the limit of the wrapped
// command, if it has one.
func (c *inPasses[D]) MaxLines() int {
	if m, ok := c.Command.(MaxLinesCommand); ok {
		return m.MaxLines()
	}
	return math.MaxInt
}
