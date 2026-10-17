// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandNote3.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/command/CommandNoteLong3.java
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

// noteColor is the color parser of the note commands.
func noteColor() *color.ColorParser { return color.SimpleColor(color.TypeBack) }

// noteTags is the "$tag" leaf of the note commands.
func noteTags() *regex.Leaf { return uml.Leaf(4, "TAGS", klimt.StereotagPattern()+"?") }

// executeNote is the end of the note commands: the note with the position,
// type, colors and stereotype of arg.
func executeNote(d *ActivityDiagram3, arg *regex.Result, note *klimt.Display) command.Result {
	position := notePositionDefaultLeft(arg.Get("POSITION", 0))
	typ := noteTypeDefault(arg.Get("TYPE", 0))
	colors, err := noteColor().Color(arg, d.Skin().IHtmlColorSet())
	if err != nil {
		return colorResult(err)
	}
	var stereotype *klimt.Stereotype
	if s, ok := arg.Get("STEREO", 0); ok {
		stereotype = klimt.StereotypeBuild(s)
	}
	return d.AddNote(note, position, typ, colors, stereotype)
}

// commandNote3 is "note left #color: text" or "floating note: text".
func commandNote3() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		return executeNote(d, arg, klimt.DisplayWithNewlines(d.Pragma(), arg.Str("NOTE", 0)))
	},
		regex.Start(),
		uml.Leaf(1, "TYPE", "(note|floating note)"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "POSITION", "(left|right)?"),
		regex.SpaceZeroOrMore(),
		noteColor().Regex(),
		regex.SpaceZeroOrMore(),
		noteTags(),
		klimt.StereotypePatternOptional("STEREO"),
		uml.Lit(":"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "NOTE", "(.*)"),
		regex.End())
}

var endNote = uml.Compile("^end[%s]?note$")

// commandNoteLong3 is "note left" ... "end note".
func commandNoteLong3() cmd {
	c := &uml.Multilines2[*ActivityDiagram3]{
		Start: regex.NewConcat(regex.Start(),
			uml.Leaf(1, "TYPE", "(note|floating note)"),
			regex.SpaceZeroOrMore(),
			uml.Leaf(1, "POSITION", "(left|right)?"),
			regex.SpaceZeroOrMore(),
			noteTags(),
			klimt.StereotypePatternOptional("STEREO"),
			noteColor().Regex(),
			regex.End()),
		Strategy: uml.RemoveStartingQuote,
		TrimEnd:  uml.TrimBoth,
		End:      endNote,
	}
	c.Exec = func(d *ActivityDiagram3, lines command.BlocLines, _ command.ParserPass) command.Result {
		line0 := c.StartArg(lines)
		lines = uml.SubExtract(lines, 1, 1).RemoveEmptyColumns()
		return executeNote(d, line0, klimt.DisplayCreateFoo(lines.Lines()))
	}
	return c
}
