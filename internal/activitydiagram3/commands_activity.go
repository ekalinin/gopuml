// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/command/CommandActivity3.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/command/CommandActivityLong3.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/command/CommandActivityList.java
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

// deprecatedColorWarning is the warning of the colors written before the
// ":" of an action on several lines.
func deprecatedColorWarning(c string) *klimt.Warning {
	return klimt.NewWarning("This syntax is deprecated, you must add <<" + c +
		">> at the end of the line, after the ';'")
}

// commandActivity3 is ":label; <<stereo>> [[url]]". A color or a
// stereotype before the ":" is ignored with a warning.
func commandActivity3() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		url := optionalURL(d, arg)
		stereogroup := klimt.StereogroupBuildFrom(arg)
		label := arg.Str("LABEL", 0)
		if ignored, ok := arg.Get("IGNORED", 0); ok {
			d.AddWarning(klimt.NewWarning("The stereotype " + ignored + " before the ':' is ignored: write ':" +
				label + "; " + ignored + "' instead"))
		}
		if c, ok := arg.Get("COLOR", 0); ok {
			d.AddWarning(klimt.NewWarning("This syntax is deprecated and the color is ignored: write ':" + label +
				"; <<" + c + ">>' instead of '" + c + ":" + label + ";'"))
		}
		style := boxStyleOf(stereogroup)
		display := klimt.DisplayWithNewlines(d.Pragma(), label)
		return d.AddActivity(display, style, url, stereogroup)
	},
		regex.Start(),
		color.SimpleColor(color.TypeBack).Regex(),
		klimt.StereotypePatternOptionalIGNORED(),
		uml.Lit(":"),
		uml.Leaf(1, "LABEL", "(.*?)"),
		uml.Lit(";"),
		regex.SpaceZeroOrMore(),
		klimt.StereogroupOptional(),
		regex.SpaceZeroOrMore(),
		klimt.URLBuilderOptional(),
		regex.End())
}

// commandActivityLong3 is an action on several lines, ":first line" ...
// "last line; <<stereo>> [[url]]".
func commandActivityLong3() cmd {
	c := &multilines3{
		start: regex.NewConcat(regex.Start(),
			color.SimpleColor(color.TypeBack).Regex(),
			uml.Lit(":"),
			uml.Leaf(1, "DATA", "(.*)"),
			regex.End()),
		strategy: uml.RemoveStartingQuote,
		trimEnd:  uml.TrimNone,
		end: regex.NewConcat(
			uml.Leaf(1, "TEXT", "(.*)"),
			uml.Lit(";"),
			regex.SpaceZeroOrMore(),
			klimt.StereogroupOptional(),
			regex.SpaceZeroOrMore(),
			klimt.URLBuilderOptional(),
			regex.End()),
	}
	c.exec = func(d *ActivityDiagram3, lines command.BlocLines) command.Result {
		lines = lines.RemoveEmptyColumns()
		line0 := c.start.Matcher(lines.At(0).Trimmed().S)
		lineLast := c.end.Matcher(lines.Last().S)
		if col, ok := line0.Get("COLOR", 0); ok {
			d.AddWarning(deprecatedColorWarning(col))
		}
		url := optionalURL(d, lineLast)
		stereogroup := klimt.StereogroupBuildFrom(lineLast)
		style := boxStyleOf(stereogroup)
		lines = lines.RemoveStartingAndEnding(line0.Str("DATA", 0), 0)
		lines = lines.OverrideLastLine(lineLast.Str("TEXT", 0))
		result, err := d.addActivity(klimt.DisplayCreateFoo(lines.Lines()), style, url, stereogroup)
		if err != nil {
			// CommandMultilines3 does not turn NoSuchColorException into a
			// result: ProtectedCommand reports the exception.
			throwNoSuchColor(err)
		}
		return result
	}
	return c
}

// commandActivityList is "- label" or "* label".
func commandActivityList() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		stereogroup := klimt.StereogroupBuildFrom(arg)
		url := optionalURL(d, arg)
		return d.AddActivity(klimt.DisplayWithNewlines(d.Pragma(), arg.Str("LABEL", 0)), BoxStylePlain, url,
			stereogroup)
	},
		regex.Start(),
		uml.Lit("[-*]"),
		regex.SpaceZeroOrOne(),
		uml.Leaf(1, "LABEL", "(.*?)"),
		regex.SpaceZeroOrMore(),
		klimt.StereogroupOptional(),
		regex.SpaceZeroOrMore(),
		klimt.URLBuilderOptional(),
		regex.End())
}
