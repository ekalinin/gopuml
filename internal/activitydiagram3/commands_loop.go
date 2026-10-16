// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandRepeat3.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandRepeatWhile3.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandRepeatWhile3Multilines.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandBackward3.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandBackwardLong3.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandWhile3.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandWhileEnd3.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandBreak.java
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

// commandRepeat3 is "repeat :label; <<stereo>>".
func commandRepeat3() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		label := withNewlines(d, arg, "LABEL")
		stereogroup := klimt.StereogroupBuildFrom(arg)
		boxStyle := boxStyleOf(stereogroup)
		if c, ok := arg.Get("COLOR", 0); ok {
			d.AddWarning(deprecatedColorWarning(c))
		}
		colors, err := innerColors(d, stereogroup)
		if err != nil {
			return colorResult(err)
		}
		d.StartRepeat(label, boxStyle, colors, stereogroup)
		return command.Ok()
	},
		regex.Start(),
		color.Exp4(),
		uml.Lit("repeat"),
		regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewConcat(
			uml.Lit(":"),
			uml.Leaf(1, "LABEL", "(.*?)"),
			uml.Lit(";"),
			regex.SpaceZeroOrMore())),
		klimt.StereogroupOptional(),
		regex.End())
}

// lazyDisplay is Display.getWithNewlines(pragma, arg.getLazzy(key, 0)).
func lazyDisplay(d *ActivityDiagram3, arg *regex.Result, key string) *klimt.Display {
	s, ok := arg.Lazy(key, 0)
	return klimt.DisplayWithNewlinesString(d.Pragma(), s, ok)
}

// commandRepeatWhile3 is "repeat while (test) is (yes) not (out) -> label;".
func commandRepeatWhile3() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		test := lazyDisplay(d, arg, "TEST")
		yes := lazyDisplay(d, arg, "WHEN")
		out := lazyDisplay(d, arg, "OUT")
		rainbow, err := rainbowOf(d, arg, "XCOLOR")
		if err != nil {
			return colorResult(err)
		}
		stereogroup := klimt.StereogroupBuildFrom(arg)
		linkLabel := withNewlines(d, arg, "LABEL")
		return d.RepeatWhile(test, yes, out, linkLabel, rainbow, stereogroup)
	},
		regex.Start(),
		uml.Lit("repeat"),
		regex.SpaceZeroOrMore(),
		uml.Lit("while"),
		regex.SpaceZeroOrMore(),
		regex.Or("",
			regex.NewConcat(uml.Leaf(1, "TEST3", `\((.*?)\)`),
				regex.SpaceZeroOrMore(),
				isOrEquals(),
				regex.SpaceZeroOrMore(),
				uml.Leaf(1, "WHEN3", `\((.+?)\)`),
				regex.SpaceZeroOrMore(),
				uml.Leaf(1, "", "(not)"),
				regex.SpaceZeroOrMore(),
				uml.Leaf(1, "OUT3", `\((.+?)\)`)),
			regex.NewConcat(uml.Leaf(1, "TEST4", `\((.*?)\)`),
				regex.SpaceZeroOrMore(),
				uml.Leaf(1, "", "(not)"),
				regex.SpaceZeroOrMore(),
				uml.Leaf(1, "OUT4", `\((.+?)\)`)),
			regex.NewConcat(uml.Leaf(1, "TEST2", `\((.*?)\)`),
				regex.SpaceZeroOrMore(),
				isOrEquals(),
				regex.SpaceZeroOrMore(),
				uml.Leaf(1, "WHEN2", `\((.+?)\)`)),
			regex.Optional(uml.Leaf(1, "TEST1", `\((.*)\)`))),
		regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewConcat(
			regex.Or("",
				uml.Lit("->"),
				uml.Leaf(1, "XCOLOR", styleColorsMultiples)),
			regex.SpaceZeroOrMore(),
			regex.Or("",
				uml.Leaf(1, "LABEL", "(.*)"),
				regex.NewLeaf("")))),
		uml.Lit(";?"),
		regex.SpaceZeroOrMore(),
		klimt.StereogroupOptional(),
		regex.End())
}

var isOrEqualsSeparator = uml.Compile(`\)[%s]*(is|equals?)[%s]*\(`)

// commandRepeatWhile3Multilines is "repeat while (test" ... "test)".
func commandRepeatWhile3Multilines() cmd {
	c := &multilines3{
		start: regex.NewConcat(regex.Start(),
			uml.Lit("repeat"),
			regex.SpaceZeroOrMore(),
			uml.Lit("while"),
			regex.SpaceZeroOrMore(),
			uml.Lit(`\(`),
			uml.Leaf(1, "TEST1", "(.*)"),
			regex.End()),
		strategy: uml.RemoveStartingQuote,
		trimEnd:  uml.TrimBoth,
		end: regex.NewConcat(
			uml.Leaf(1, "TEST1", "(.*)"),
			uml.Lit(`\)`),
			uml.Lit(";?"),
			regex.End()),
	}
	c.exec = func(d *ActivityDiagram3, lines command.BlocLines) command.Result {
		lines = lines.Trim()
		line0 := c.start.Matcher(text.Trin(lines.At(0).S))
		lineLast := c.end.Matcher(lines.Last().S)
		testDisplay := klimt.DisplayWithNewlines(d.Pragma(), line0.Str("TEST1", 0))
		for _, s := range uml.SubExtract(lines, 1, 1).Lines() {
			testDisplay = testDisplay.Add(klimt.Text(s.S))
		}
		if trailTest, ok := lineLast.Get("TEST1", 0); !uml.IsEmpty(trailTest, ok) {
			testDisplay = testDisplay.Add(klimt.Text(trailTest))
		}
		yes := klimt.DisplayNULL()
		if split := testDisplay.SplitMultiline(isOrEqualsSeparator); len(split) == 2 {
			testDisplay = split[0]
			yes = split[1]
		}
		return d.RepeatWhile(testDisplay, yes, klimt.DisplayNULL(), klimt.DisplayNULL(), RainbowNone(),
			klimt.StereogroupNone())
	}
	return c
}

// arrowLabelPart is the optional "(label)" or "(-[#red]-> label)" around a
// "backward" action, captured as name and name+"_COLOR".
func arrowLabelPart(name string) regex.Part {
	return regex.Optional(regex.NewConcat(
		uml.Lit(`\(`),
		regex.Optional(regex.Or("",
			uml.Lit("->"),
			uml.Leaf(1, name+"_COLOR", styleColorsMultiples))),
		uml.Leaf(1, name, "(.*?)"),
		uml.Lit(`\)`)))
}

// commandBackward3 is "(in) backward :label; <<stereo>> (out)".
func commandBackward3() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		stereogroup := klimt.StereogroupBuildFrom(arg)
		boxStyle := boxStyleOf(stereogroup)
		stereotype := stereogroup.BuildStereotype()
		label := withNewlines(d, arg, "LABEL")
		in, err := backRendering(d, arg, "INCOMING")
		if err != nil {
			return colorResult(err)
		}
		out, err := backRendering(d, arg, "OUTCOMING")
		if err != nil {
			return colorResult(err)
		}
		return d.Backward(label, boxStyle, in, out, stereotype)
	},
		regex.Start(),
		arrowLabelPart("INCOMING"),
		regex.SpaceZeroOrMore(),
		uml.Lit("backward"),
		regex.SpaceZeroOrMore(),
		uml.Lit(":"),
		uml.Leaf(1, "LABEL", "(.*?)"),
		uml.Lit(";"),
		regex.SpaceZeroOrMore(),
		klimt.StereogroupOptional(),
		regex.SpaceZeroOrMore(),
		arrowLabelPart("OUTCOMING"),
		regex.SpaceZeroOrMore(),
		regex.End())
}

// commandBackwardLong3 is a "backward" action on several lines.
func commandBackwardLong3() cmd {
	c := &multilines3{
		start: regex.NewConcat(regex.Start(),
			regex.SpaceZeroOrMore(),
			uml.Lit("backward"),
			regex.SpaceZeroOrMore(),
			uml.Lit(":"),
			regex.SpaceZeroOrMore(),
			uml.Leaf(1, "DATA", "(.*)"),
			regex.End()),
		strategy: uml.RemoveStartingQuote,
		trimEnd:  uml.TrimBoth,
		end: regex.NewConcat(
			uml.Leaf(1, "TEXT", "(.*)"),
			uml.Lit(";"),
			regex.SpaceZeroOrMore(),
			klimt.StereogroupOptional(),
			regex.End()),
	}
	c.exec = func(d *ActivityDiagram3, lines command.BlocLines) command.Result {
		lines = lines.RemoveEmptyColumns()
		line0 := c.start.Matcher(lines.At(0).Trimmed().S)
		lineLast := c.end.Matcher(lines.Last().S)
		stereogroup := klimt.StereogroupBuildFrom(lineLast)
		stereotype := stereogroup.BuildStereotype()
		style := boxStyleOf(stereogroup)
		lines = lines.RemoveStartingAndEnding(line0.Str("DATA", 0), 0)
		lines = lines.OverrideLastLine(lineLast.Str("TEXT", 0))
		return d.Backward(klimt.DisplayCreateFoo(lines.Lines()), style, LinkRenderingNone(), LinkRenderingNone(),
			stereotype)
	}
	return c
}

// commandWhile3 is "while (test) is (yes)".
func commandWhile3() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		colors, err := innerColors(d, klimt.StereogroupBuildFrom(arg))
		if err != nil {
			return colorResult(err)
		}
		c := colors.Color(color.TypeBack)
		d.DoWhile(withNewlines(d, arg, "TEST"), withNewlines(d, arg, "YES"), c)
		return command.Ok()
	},
		regex.Start(),
		color.Exp4(),
		uml.Lit("while"),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\(`),
		uml.Leaf(1, "TEST", "(.*?)"),
		uml.Lit(`\)`),
		regex.Optional(regex.NewConcat(
			regex.SpaceZeroOrMore(),
			isOrEquals(),
			regex.SpaceZeroOrMore(),
			uml.Leaf(1, "YES", `\((.+?)\)`))),
		uml.Lit(";?"),
		regex.SpaceZeroOrMore(),
		klimt.StereogroupOptional(),
		regex.End())
}

// commandWhileEnd3 is "endwhile (out)".
func commandWhileEnd3() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		return d.Endwhile(withNewlines(d, arg, "OUT"))
	},
		regex.Start(),
		regex.Or("",
			regex.NewConcat(
				uml.Lit("end"),
				regex.SpaceZeroOrMore(),
				uml.Lit("while")),
			regex.NewConcat(
				uml.Lit("while"),
				regex.SpaceZeroOrMore(),
				uml.Lit("end"))),
		regex.SpaceZeroOrMore(),
		regex.Optional(uml.Leaf(1, "OUT", `\((.+?)\)`)),
		uml.Lit(";?"),
		regex.End())
}

// commandBreak is "break".
func commandBreak() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, _ *regex.Result) command.Result {
		d.BreakInstruction()
		return command.Ok()
	}, regex.Start(), uml.Lit("break"), uml.Lit(";?"), regex.End())
}
