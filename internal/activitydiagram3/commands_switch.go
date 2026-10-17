// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandSwitch.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandCase.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandEndSwitch.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// commandSwitch is "switch (test)".
func commandSwitch() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		colors, err := innerColors(d, klimt.StereogroupBuildFrom(arg))
		if err != nil {
			return colorResult(err)
		}
		d.StartSwitch(testDisplay(d, arg, "TEST"), colors)
		return command.Ok()
	},
		regex.Start(),
		color.Exp4(),
		uml.Lit("switch"),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\(`),
		uml.Leaf(1, "TEST", "(.*?)"),
		uml.Lit(`\)`),
		regex.SpaceZeroOrMore(),
		klimt.StereogroupOptional(),
		regex.End())
}

// commandCase is "case (label)".
func commandCase() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		return d.SwitchCase(testDisplay(d, arg, "TEST"))
	},
		regex.Start(),
		uml.Lit("case"),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\(`),
		uml.Leaf(1, "TEST", "(.*?)"),
		uml.Lit(`\)`),
		regex.SpaceZeroOrMore(),
		regex.End())
}

// commandEndSwitch is "endswitch".
func commandEndSwitch() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		colors, err := innerColors(d, klimt.StereogroupBuildFrom(arg))
		if err != nil {
			return colorResult(err)
		}
		return d.EndSwitch(colors)
	},
		regex.Start(),
		uml.Lit("endswitch"),
		regex.SpaceZeroOrMore(),
		klimt.StereogroupOptional(),
		regex.End())
}
