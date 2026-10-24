// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/note/CommandFactoryNoteActivity.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// noteActivityHead is "note left|right|top|bottom #color".
func noteActivityHead() []regex.Part {
	return []regex.Part{regex.Start(),
		uml.Lit("note"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "POSITION", "(right|left|top|bottom)"),
		regex.SpaceZeroOrMore(),
		color.Exp1(),
		regex.SpaceZeroOrMore(),
	}
}

// commandNoteActivityMulti is upstream's
// CommandFactoryNoteActivity.createMultiLine: "note right" ... "end note",
// on the last activity. A URL on the first line goes to the note.
func commandNoteActivityMulti() cmd {
	return abel.Multilines(&uml.Multilines2[*ActivityDiagram]{
		Start:    regex.NewConcat(append(noteActivityHead(), regex.End())...),
		Strategy: uml.KeepStartingQuote,
		TrimEnd:  uml.TrimBoth,
		End:      uml.Compile(`^[%s]*end[%s]?note$`),
	}, func(c *uml.Multilines2[*ActivityDiagram], d *ActivityDiagram, lines command.BlocLines) command.Result {
		arg := c.StartArg(lines)
		lines = uml.SubExtract(lines, 1, 1).RemoveEmptyColumns()
		strings := klimt.DisplayCreateFoo(lines.Lines())
		var url *klimt.URL
		if strings.Size() > 0 {
			url = abel.URLBuilderOf(d.Cuca()).URL(strings.Get(0).String())
		}
		if url != nil {
			strings = strings.SubList(1, strings.Size())
		}
		quark := d.QuarkInContext(true, d.UniqueSequence("GMN"))
		location := uml.Location(lines)
		note := d.ReallyCreateLeaf(location, quark, strings, abel.LeafTypeNote, abel.USymbolNull)
		if url != nil {
			note.AddURL(url)
		}
		return noteActivityExecute(location, d, arg, note)
	})
}

// commandNoteActivitySingle is upstream's
// CommandFactoryNoteActivity.createSingleLine: "note right: text" on the
// last activity.
func commandNoteActivitySingle() cmd {
	parts := append(noteActivityHead(),
		uml.Lit(":"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "NOTE", "(.*)"),
		regex.End())
	return uml.Single(regex.NewConcat(parts...),
		func(d *ActivityDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			tmp := d.UniqueSequence("GN")
			quark := d.QuarkInContext(true, d.CleanID(tmp))
			note := d.createNote(location, quark, klimt.DisplayWithNewlines(d.Pragma(), arg.Str("NOTE", 0)))
			return noteActivityExecute(location, d, arg, note)
		})
}

// noteActivityExecute colors the note and links it with a dashed line to
// the last activity, or to the start.
func noteActivityExecute(location *text.LineLocation, d *ActivityDiagram, arg *regex.Result,
	note *abel.Entity) command.Result {
	var back color.HColor
	if s, ok := arg.Get("COLOR", 0); ok {
		c, err := d.Skin().IHtmlColorSet().GetColor(s)
		if err != nil {
			return abel.ColorResult(err)
		}
		back = c
	}
	note.SetSpecificColorTOBEREMOVED(color.TypeBack, back)
	activity := d.lastEntity()
	if activity == nil {
		activity = d.getStart(location)
	}
	position := abel.PositionValueOf(uml.UpperCase(arg.Str("POSITION", 0))).WithRankdir(d.Skin().Rankdir())
	typ := decoration.NewLinkType(decoration.LinkDecorNone, decoration.LinkDecorNone).GoDashed()
	styleBuilder := d.SkinParam().CurrentStyleBuilder()
	var link *abel.Link
	switch position {
	case abel.PositionRight:
		link = abel.NewLink(location, d.Cuca(), styleBuilder, activity, note, typ, abel.LinkArgNoDisplay(1))
	case abel.PositionLeft:
		link = abel.NewLink(location, d.Cuca(), styleBuilder, note, activity, typ, abel.LinkArgNoDisplay(1))
	case abel.PositionBottom:
		link = abel.NewLink(location, d.Cuca(), styleBuilder, activity, note, typ, abel.LinkArgNoDisplay(2))
	case abel.PositionTop:
		link = abel.NewLink(location, d.Cuca(), styleBuilder, note, activity, typ, abel.LinkArgNoDisplay(2))
	default:
		uml.Throw(uml.IllegalArgumentException, "")
	}
	d.AddLink(link)
	return command.Ok()
}
