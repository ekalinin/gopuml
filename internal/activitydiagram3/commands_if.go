// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandIf4.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandIf2.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandIfLegacy1.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandElseIf3.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandElseIf2.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandElse3.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandElseLegacy1.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandEndif3.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/command/regex"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/text"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// optionalColor is HColorSet.getColor of the group COLOR, nil when it did
// not match.
func optionalColor(d *ActivityDiagram3, arg *regex.Result) (color.HColor, error) {
	s, ok := arg.Get("COLOR", 0)
	if !ok {
		return nil, nil
	}
	return getColor(d, s)
}

// isOrEquals is the unnamed "(is|equals?)" group of the if and loop
// commands.
func isOrEquals() *regex.Leaf { return uml.Leaf(1, "", "(is|equals?)") }

// commandIf4 is "if (test) is (value) then".
func commandIf4() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		c, err := optionalColor(d, arg)
		if err != nil {
			return colorResult(err)
		}
		test := testDisplay(d, arg, "TEST")
		d.StartIf(test, withNewlines(d, arg, "WHEN"), c, nil, nil)
		return command.Ok()
	},
		regex.Start(),
		color.Exp4(),
		uml.Lit("if"),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\(`),
		uml.Leaf(1, "TEST", "(.*?)"),
		uml.Lit(`\)`),
		regex.SpaceZeroOrMore(),
		isOrEquals(),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\(`),
		uml.Leaf(1, "WHEN", "(.+?)"),
		uml.Lit(`\)`),
		regex.SpaceZeroOrMore(),
		uml.Lit("then"),
		regex.SpaceZeroOrMore(),
		uml.Lit(";?"),
		regex.SpaceZeroOrMore(),
		klimt.StereogroupOptional(),
		regex.End())
}

// commandIf2 is "if (test) then (label)".
func commandIf2() *command.SingleLine[*ActivityDiagram3] {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		test := testDisplay(d, arg, "TEST")
		stereogroup := klimt.StereogroupBuildFrom(arg)
		colors, err := innerColors(d, stereogroup)
		if err != nil {
			return colorResult(err)
		}
		url := optionalURL(d, arg)
		stereotype := stereogroup.BuildStereotype()
		c := colors.Color(color.TypeBack)
		d.StartIf(test, withNewlines(d, arg, "WHEN"), c, url, stereotype)
		return command.Ok()
	},
		regex.Start(),
		klimt.URLBuilderOptional(),
		regex.SpaceZeroOrMore(),
		color.Exp4(),
		uml.Lit("if"),
		klimt.StereotypePatternOptionalIGNORED(),
		uml.Lit(`\(`),
		uml.Leaf(1, "TEST", "(.*?)"),
		uml.Lit(`\)`),
		regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewConcat(
			uml.Lit("then"),
			regex.SpaceZeroOrMore(),
			regex.Optional(uml.Leaf(1, "WHEN", `\((.+?)\)`)))),
		uml.Lit(";?"),
		regex.SpaceZeroOrMore(),
		klimt.StereogroupOptional(),
		regex.End())
}

// commandIfLegacy1 is "if (test) then when label".
func commandIfLegacy1() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		d.StartIf(withNewlines(d, arg, "TEST"), withNewlines(d, arg, "WHEN"), nil, nil, nil)
		return command.Ok()
	},
		regex.Start(),
		uml.Lit("if"),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\(`),
		uml.Leaf(1, "TEST", "(.+?)"),
		uml.Lit(`\)`),
		regex.SpaceZeroOrMore(),
		uml.Lit("then"),
		regex.SpaceOneOrMore(),
		uml.Lit("when"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "WHEN", "(.*)"),
		uml.Lit(";?"),
		regex.End())
}

// incomingArrow is the optional "(label)" or "(-[#red]-> label)" before
// "else" and "elseif", captured as INCOMING and INCOMING_COLOR.
func incomingArrow() regex.Part {
	return regex.Optional(regex.NewConcat(
		uml.Lit(`\(`),
		regex.Optional(regex.Or("",
			uml.Lit("->"),
			uml.Leaf(1, "INCOMING_COLOR", styleColorsMultiples))),
		uml.Leaf(1, "INCOMING", "(.*?)"),
		uml.Lit(`\)`)))
}

// whenArrowColor is the optional "->" or "-[#red]->" of a "then (...)",
// captured as WHEN_COLOR.
func whenArrowColor() regex.Part {
	return regex.Optional(regex.Or("",
		uml.Lit("->"),
		uml.Leaf(1, "WHEN_COLOR", styleColorsMultiples)))
}

// executeElseIf is the end of the executeArg of the elseif commands.
func executeElseIf(d *ActivityDiagram3, arg *regex.Result, c color.HColor) command.Result {
	incoming, err := backRendering(d, arg, "INCOMING")
	if err != nil {
		return colorResult(err)
	}
	when, err := backRendering(d, arg, "WHEN")
	if err != nil {
		return colorResult(err)
	}
	return d.ElseIf(incoming, testDisplay(d, arg, "TEST"), when, c)
}

// commandElseIf3 is "elseif (test) is (value) then".
func commandElseIf3() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		c, err := optionalColor(d, arg)
		if err != nil {
			return colorResult(err)
		}
		return executeElseIf(d, arg, c)
	},
		regex.Start(),
		color.Exp4(),
		regex.SpaceZeroOrMore(),
		incomingArrow(),
		regex.SpaceZeroOrMore(),
		uml.Lit("else"),
		regex.SpaceZeroOrMore(),
		uml.Lit("if"),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\(`),
		uml.Leaf(1, "TEST", "(.*?)"),
		uml.Lit(`\)`),
		regex.SpaceZeroOrMore(),
		isOrEquals(),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\(`),
		uml.Leaf(1, "WHEN", "(.+?)"),
		uml.Lit(`\)`),
		regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewConcat(
			uml.Lit("then"),
			regex.SpaceZeroOrMore(),
			regex.Optional(regex.NewConcat(
				uml.Lit(`\(`),
				whenArrowColor(),
				uml.Lit(`\)`))))),
		uml.Lit(";?"),
		regex.End())
}

// commandElseIf2 is "elseif (test) then (label)".
func commandElseIf2() *command.SingleLine[*ActivityDiagram3] {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		stereogroup := klimt.StereogroupBuildFrom(arg)
		colors, err := innerColors(d, stereogroup)
		if err != nil {
			return colorResult(err)
		}
		return executeElseIf(d, arg, colors.Color(color.TypeBack))
	},
		regex.Start(),
		color.Exp4(),
		regex.SpaceZeroOrMore(),
		incomingArrow(),
		regex.SpaceZeroOrMore(),
		uml.Lit("else"),
		regex.SpaceZeroOrMore(),
		uml.Lit("if"),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\(`),
		uml.Leaf(1, "TEST", "(.*?)"),
		uml.Lit(`\)`),
		regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewConcat(
			uml.Lit("then"),
			regex.SpaceZeroOrMore(),
			regex.Optional(regex.NewConcat(
				uml.Lit(`\(`),
				whenArrowColor(),
				uml.Leaf(1, "WHEN", "(.*?)"),
				uml.Lit(`\)`))))),
		uml.Lit(";?"),
		regex.SpaceZeroOrMore(),
		klimt.StereogroupOptional(),
		regex.End())
}

// commandElse3 is "else (label)".
func commandElse3() *command.SingleLine[*ActivityDiagram3] {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		when, err := backRendering(d, arg, "WHEN")
		if err != nil {
			return colorResult(err)
		}
		return d.Else2(when)
	},
		regex.Start(),
		uml.Lit("else"),
		regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewConcat(
			uml.Lit(`\(`),
			whenArrowColor(),
			uml.Leaf(1, "WHEN", "(.*?)"),
			uml.Lit(`\)`))),
		regex.SpaceZeroOrMore(),
		uml.Lit(";?"),
		regex.End())
}

// commandElseLegacy1 is "else when label".
func commandElseLegacy1() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		when := withNewlines(d, arg, "WHEN")
		return d.Else2(LinkRenderingNone().WithDisplay(when))
	},
		regex.Start(),
		uml.Lit("else"),
		regex.SpaceOneOrMore(),
		uml.Lit("when"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "WHEN", "(.*)"),
		uml.Lit(";?"),
		regex.End())
}

// commandEndif3 is "endif".
func commandEndif3() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		colors, err := innerColors(d, klimt.StereogroupBuildFrom(arg))
		if err != nil {
			return colorResult(err)
		}
		return d.Endif(colors)
	},
		regex.Start(),
		uml.Lit("end"),
		regex.SpaceZeroOrMore(),
		uml.Lit("if"),
		uml.Lit(";?"),
		regex.SpaceZeroOrMore(),
		klimt.StereogroupOptional(),
		regex.End())
}
