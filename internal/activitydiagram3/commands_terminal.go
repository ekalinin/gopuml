// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/command/CommandStart3.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/command/CommandStop3.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/command/CommandEnd3.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/command/CommandKill3.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/command/CommandCircleSpot3.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/command/CommandLabel.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/command/CommandGoto.java
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

// terminalCommand is "keyword <<stereo>>", running exec with the inner
// colors of the stereotypes.
func terminalCommand(keyword string, exec func(d *ActivityDiagram3, colors *color.Colors)) cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		colors, err := innerColors(d, klimt.StereogroupBuildFrom(arg))
		if err != nil {
			return colorResult(err)
		}
		exec(d, colors)
		return command.Ok()
	},
		regex.Start(),
		uml.Lit(keyword),
		uml.Lit(";?"),
		regex.SpaceZeroOrMore(),
		klimt.StereogroupOptional(),
		regex.End())
}

// commandStart3 is "start".
func commandStart3() cmd { return terminalCommand("start", (*ActivityDiagram3).Start) }

// commandStop3 is "stop".
func commandStop3() cmd { return terminalCommand("stop", (*ActivityDiagram3).Stop) }

// commandEnd3 is "end".
func commandEnd3() cmd { return terminalCommand("end", (*ActivityDiagram3).End) }

// commandKill3 is "kill" or "detach". As upstream, the pattern is
// "^kill|detach;?$": any line starting with "kill" or ending with "detach"
// matches.
func commandKill3() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, _ *regex.Result) command.Result {
		return d.Kill()
	}, regex.Start(), uml.Lit("kill|detach"), uml.Lit(";?"), regex.End())
}

// commandCircleSpot3 is the connector "(A)".
func commandCircleSpot3() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		c, err := optionalColor(d, arg)
		if err != nil {
			return colorResult(err)
		}
		d.AddSpot(arg.Str("SPOT", 0), c)
		return command.Ok()
	},
		regex.Start(),
		color.Exp4(),
		uml.Leaf(1, "SPOT", `\((\S)\)`),
		uml.Lit(";?"),
		regex.End())
}

// commandLabel is "label name".
func commandLabel() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		return d.AddLabel(arg.Str("NAME", 0))
	},
		regex.Start(),
		uml.Lit("label"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "NAME", "([%pLN_.]+)"),
		uml.Lit(";?"),
		regex.End())
}

// commandGoto is "goto name".
func commandGoto() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		return d.AddGoto(arg.Str("NAME", 0))
	},
		regex.Start(),
		uml.Lit("goto"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "NAME", "([%pLN_.]+)"),
		uml.Lit(";?"),
		regex.End())
}
