// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/statediagram/command/CommandEndState.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/statediagram/command/CommandConcurrentState.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package statediagram

import (
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// commandEndState is "}" or "end state", which closes a composite state.
func commandEndState() cmd {
	c := uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "", `(end[%s]?state|\})`),
		regex.End()),
		func(d *StateDiagram, _ *text.LineLocation, _ *regex.Result, _ command.ParserPass) command.Result {
			if current := d.CurrentGroup(); current == nil || current.IsRoot() {
				return command.Error("No inner state defined")
			}
			d.EndGroup()
			return command.Ok()
		})
	c.Passes = allPasses
	return c
}

// commandConcurrentState is "--" or "||", which starts a new concurrent
// region of the current state.
func commandConcurrentState() cmd {
	c := uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "TYPE", `(--+|\|\|+)`),
		regex.End()),
		func(d *StateDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			separator, _ := arg.Get("TYPE", 0)
			if d.concurrentState(location, rune(separator[0])) {
				return command.Ok()
			}
			return command.Error("Error 42")
		})
	c.Passes = allPasses
	return c
}
