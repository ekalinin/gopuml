// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandArrow3.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandArrowLong3.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandLink3.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/command/regex"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/text"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// arrowStart is the "->" or "-[#color,style]->" of the arrow commands.
func arrowStart() regex.Part {
	return regex.Or("", uml.Lit("->"), uml.Leaf(1, "COLOR", styleColorsMultiples))
}

// setArrowColor sets the colors captured as COLOR on the next arrow.
func setArrowColor(d *ActivityDiagram3, arg *regex.Result) error {
	if _, ok := arg.Get("COLOR", 0); !ok {
		return nil
	}
	rainbow, err := rainbowOf(d, arg, "COLOR")
	if err != nil {
		return err
	}
	d.SetColorNextArrow(rainbow)
	return nil
}

// commandArrow3 is "-> label;" or "-[#red]-> label;".
func commandArrow3() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		if err := setArrowColor(d, arg); err != nil {
			return colorResult(err)
		}
		if label, ok := arg.Get("LABEL", 0); ok && label != "" {
			d.SetLabelNextArrow(klimt.DisplayWithNewlines(d.Pragma(), label))
		}
		return command.Ok()
	},
		regex.Start(),
		arrowStart(),
		regex.SpaceZeroOrMore(),
		regex.Or("", uml.Leaf(1, "LABEL", "(.*);"), regex.NewLeaf("")),
		regex.End())
}

var arrowLongEnd = uml.Compile("^(.*);$")

// commandArrowLong3 is an arrow label on several lines, ending with ";".
func commandArrowLong3() cmd {
	c := &uml.Multilines2[*ActivityDiagram3]{
		Start: regex.NewConcat(regex.Start(),
			arrowStart(),
			regex.SpaceZeroOrMore(),
			uml.Leaf(1, "LABEL", "(.*)"),
			regex.End()),
		Strategy: uml.RemoveStartingQuote,
		TrimEnd:  uml.TrimBoth,
		End:      arrowLongEnd,
	}
	c.Exec = func(d *ActivityDiagram3, lines command.BlocLines, _ command.ParserPass) command.Result {
		lines = lines.RemoveEmptyColumns()
		line0 := c.Start.Matcher(lines.At(0).Trimmed().S)
		if err := setArrowColor(d, line0); err != nil {
			return colorResult(err)
		}
		lines = lines.RemoveStartingAndEnding(line0.Str("LABEL", 0), 1)
		d.SetLabelNextArrow(klimt.DisplayCreateFoo(lines.Lines()))
		return command.Ok()
	}
	return c
}

// commandLink3 is "link #color".
func commandLink3() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		c, err := getColor(d, arg.Str("COLOR", 0))
		if err != nil {
			return colorResult(err)
		}
		if c != nil {
			d.SetColorNextArrow(RainbowFromColor(c, nil))
		}
		return command.Ok()
	},
		regex.Start(),
		uml.Lit("link"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "COLOR", `(#\w+)`),
		uml.Lit(";?"),
		regex.End())
}
