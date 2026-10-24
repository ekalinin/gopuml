// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/command/CommandFork3.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/command/CommandForkAgain3.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/command/CommandForkEnd3.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/command/CommandSplit3.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/command/CommandSplitAgain3.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/command/CommandSplitEnd3.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// commandFork3 is "fork <<stereo>>".
func commandFork3() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		colors, err := innerColors(d, klimt.StereogroupBuildFrom(arg))
		if err != nil {
			return colorResult(err)
		}
		d.Fork(colors)
		return command.Ok()
	},
		regex.Start(),
		uml.Lit("fork"),
		uml.Lit(";?"),
		regex.SpaceZeroOrMore(),
		klimt.StereogroupOptional(),
		regex.End())
}

// commandForkAgain3 is "fork again".
func commandForkAgain3() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, _ *regex.Result) command.Result {
		return d.ForkAgain()
	},
		regex.Start(),
		uml.Lit("fork"),
		regex.SpaceZeroOrMore(),
		uml.Lit("again"),
		uml.Lit(";?"),
		regex.End())
}

// commandForkEnd3 is "end fork {label}" or "end merge".
func commandForkEnd3() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		forkStyle := ForkStyleFork
		if strings.Contains(arg.Str("STYLE", 0), "merge") {
			forkStyle = ForkStyleMerge
		}
		label, ok := arg.Get("LABEL", 0)
		return d.EndFork(forkStyle, label, ok)
	},
		regex.Start(),
		regex.Or("STYLE",
			regex.NewConcat(
				uml.Lit("end"),
				regex.SpaceZeroOrMore(),
				uml.Lit("fork")),
			regex.NewConcat(
				uml.Lit("fork"),
				regex.SpaceZeroOrMore(),
				uml.Lit("end")),
			regex.NewConcat(
				uml.Lit("end"),
				regex.SpaceZeroOrMore(),
				uml.Lit("merge"))),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "LABEL", `(\{.+\})?`),
		uml.Lit(";?"),
		regex.End())
}

// commandSplit3 is "split".
func commandSplit3() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, _ *regex.Result) command.Result {
		d.Split()
		return command.Ok()
	}, regex.Start(), uml.Lit("split"), uml.Lit(";?"), regex.End())
}

// commandSplitAgain3 is "split again".
func commandSplitAgain3() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, _ *regex.Result) command.Result {
		return d.SplitAgain()
	},
		regex.Start(),
		uml.Lit("split"),
		regex.SpaceZeroOrMore(),
		uml.Lit("again"),
		uml.Lit(";?"),
		regex.End())
}

// commandSplitEnd3 is "end split".
func commandSplitEnd3() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, _ *regex.Result) command.Result {
		return d.EndSplit()
	},
		regex.Start(),
		regex.Or("",
			regex.NewConcat(
				uml.Lit("end"),
				regex.SpaceZeroOrMore(),
				uml.Lit("split")),
			regex.NewConcat(
				uml.Lit("split"),
				regex.SpaceZeroOrMore(),
				uml.Lit("end"))),
		uml.Lit(";?"),
		regex.End())
}
