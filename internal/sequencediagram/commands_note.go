// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/note/sequence/FactorySequenceNoteCommand.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/note/sequence/FactorySequenceNoteOverSeveralCommand.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/note/sequence/FactorySequenceNoteAcrossCommand.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/note/sequence/FactorySequenceNoteOnArrowCommand.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/color/Colors.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"regexp"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

var endNote = uml.Compile("^end[%s]?(note|hnote|rnote)$")

// noteBody is the text of a multi-line note: the lines between the first
// and the last one, without their shared indentation.
func noteBody(lines command.BlocLines) (command.BlocLines, *klimt.Display) {
	lines = uml.SubExtract(lines, 1, 1).RemoveEmptyColumns()
	return lines, klimt.DisplayCreateFoo(lines.Lines())
}

// applyStereotypeForNote is Colors.applyStereotypeForNote with the note
// colors: the colors of the stereotype for the background and the border
// when colors has none, and the shadowing of the stereotype.
func applyStereotypeForNote(note *Note, colors *color.Colors, stereotype *klimt.Stereotype,
	skin *klimt.SkinParam) (*color.Colors, error) {
	var err error
	for _, param := range []klimt.ColorParam{klimt.ColorParamNoteBackground, klimt.ColorParamNoteBorder} {
		colors, err = klimt.ColorsApplyStereotype(colors, stereotype, skin, param)
		if err != nil {
			return nil, err
		}
	}
	shadowing := skin.ShadowingForNote(stereotype)
	note.shadowing = &shadowing
	return colors, nil
}

// noteDecorations sets the colors, stereotype, style and link of a note,
// as the note commands do.
func noteDecorations(d *SequenceDiagram, note *Note, arg *regex.Result, colorParser *color.ColorParser) error {
	colors, err := colorParser.Color(arg, d.Skin().IHtmlColorSet())
	if err != nil {
		return err
	}
	if stereotypeString, ok := arg.Lazy("STEREO", 0); ok {
		stereotype := klimt.StereotypeBuild(stereotypeString)
		colors, err = applyStereotypeForNote(note, colors, stereotype, d.Skin())
		if err != nil {
			return err
		}
		note.SetStereotype(stereotype)
	}
	note.SetColors(colors)
	note.SetNoteStyle(noteStyleOf(arg.Str("STYLE", 0)))
	if url, ok := arg.Get(klimt.URLBuilderKey, 0); ok {
		note.SetURL(urlBuilder(d).URL(url))
	}
	return nil
}

func noteColor() *color.ColorParser { return color.SimpleColor(color.TypeBack) }

// noteSingleLine builds the one-line flavor of a note command.
func noteSingleLine(parts []regex.Part, exec func(d *SequenceDiagram, location *text.LineLocation,
	arg *regex.Result, display *klimt.Display) command.Result) cmd {
	parts = append(parts,
		regex.SpaceZeroOrMore(),
		klimt.URLBuilderOptional(),
		regex.SpaceZeroOrMore(),
		uml.Lit(":"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "NOTE", "(.*)"), regex.End())
	return uml.Single(regex.NewConcat(parts...),
		func(d *SequenceDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			display := klimt.DisplayWithNewlines(d.Pragma(), arg.Str("NOTE", 0))
			return exec(d, location, arg, d.ManageVariable(display))
		})
}

// noteMultiLine builds the multi-line flavor of a note command.
func noteMultiLine(parts []regex.Part, end *regexp.Regexp, exec func(d *SequenceDiagram, location *text.LineLocation,
	arg *regex.Result, display *klimt.Display) command.Result) cmd {
	parts = append(parts,
		regex.SpaceZeroOrMore(),
		klimt.URLBuilderOptional(),
		regex.End())
	c := &uml.Multilines2[*SequenceDiagram]{
		Start:    regex.NewConcat(parts...),
		Strategy: uml.KeepStartingQuote,
		TrimEnd:  uml.TrimBoth,
		End:      end,
	}
	c.Exec = func(d *SequenceDiagram, lines command.BlocLines, _ command.ParserPass) command.Result {
		arg := c.StartArg(lines)
		lines, display := noteBody(lines)
		return exec(d, uml.Location(lines), arg, d.ManageVariable(display))
	}
	return c
}

// noteParts is the start of the patterns of "note left of A".
func noteParts(of string) []regex.Part {
	return []regex.Part{regex.Start(),
		uml.Leaf(1, "PARALLEL", "(&[%s]*)?"),
		uml.Leaf(1, "VMERGE", "(/)?"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "STYLE", "(note|hnote|rnote)"),
		klimt.StereotypePatternOptional("STEREO1"),
		uml.Leaf(1, "POSITION", "(right|left|over)"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "PARTICIPANT", "(?:of"+of+")?([%pLN_.@]+|[%g][^%g]+[%g])"),
		klimt.StereotypePatternOptional("STEREO2"),
		noteColor().Regex(),
	}
}

// executeNote is FactorySequenceNoteCommand.executeInternal.
func executeNote(d *SequenceDiagram, location *text.LineLocation, arg *regex.Result, display *klimt.Display) command.Result {
	p := d.OrCreateParticipant(location, uml.RemoveQuotes(arg.Str("PARTICIPANT", 0)))
	position := notePositionOf(uml.UpperCase(arg.Str("POSITION", 0)))
	if display.Size() > 0 {
		_, tryMerge := arg.Get("VMERGE", 0)
		_, parallel := arg.Get("PARALLEL", 0)
		note := newNote(p, nil, position, display, d.CurrentStyleBuilder())
		if err := noteDecorations(d, note, arg, noteColor()); err != nil {
			return colorResult(err)
		}
		if parallel {
			note.GoParallel()
		}
		d.AddNote(note, tryMerge)
	}
	return command.Ok()
}

// commandNoteSingle is "note left of A : text".
func commandNoteSingle() cmd { return noteSingleLine(noteParts("[%s]"), executeNote) }

// commandNoteMulti is "note left of A" ... "end note".
func commandNoteMulti() cmd { return noteMultiLine(noteParts("[%s]+"), endNote, executeNote) }

// noteOverSeveralParts is the start of the patterns of "note over A, B".
func noteOverSeveralParts() []regex.Part {
	return []regex.Part{regex.Start(),
		uml.Leaf(1, "PARALLEL", "(&[%s]*)?"),
		uml.Leaf(1, "VMERGE", "(/)?"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "STYLE", "(note|hnote|rnote)"),
		klimt.StereotypePatternOptional("STEREO1"),
		uml.Lit("over"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "P1", "([%pLN_.@]+|[%g][^%g]+[%g])"),
		regex.SpaceZeroOrMore(),
		uml.Lit(","),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "P2", "([%pLN_.@]+|[%g][^%g]+[%g])"),
		klimt.StereotypePatternOptional("STEREO2"),
		noteColor().Regex(),
	}
}

// executeNoteOverSeveral is FactorySequenceNoteOverSeveralCommand.executeInternal.
func executeNoteOverSeveral(d *SequenceDiagram, location *text.LineLocation, arg *regex.Result,
	display *klimt.Display) command.Result {
	p1 := d.OrCreateParticipant(location, uml.RemoveQuotes(arg.Str("P1", 0)))
	p2 := d.OrCreateParticipant(location, uml.RemoveQuotes(arg.Str("P2", 0)))
	if display.Size() > 0 {
		_, tryMerge := arg.Get("VMERGE", 0)
		_, parallel := arg.Get("PARALLEL", 0)
		note := newNote(p1, p2, NoteOverSeveral, display, d.CurrentStyleBuilder())
		if err := noteDecorations(d, note, arg, noteColor()); err != nil {
			return colorResult(err)
		}
		if parallel {
			note.GoParallel()
		}
		d.AddNote(note, tryMerge)
	}
	return command.Ok()
}

// commandNoteOverSeveralSingle is "note over A, B : text".
func commandNoteOverSeveralSingle() cmd {
	return noteSingleLine(noteOverSeveralParts(), executeNoteOverSeveral)
}

// commandNoteOverSeveralMulti is "note over A, B" ... "end note".
func commandNoteOverSeveralMulti() cmd {
	return noteMultiLine(noteOverSeveralParts(), endNote, executeNoteOverSeveral)
}

// noteAcrossParts is the start of the patterns of "note across".
func noteAcrossParts() []regex.Part {
	return []regex.Part{regex.Start(),
		uml.Leaf(1, "VMERGE", "(/)?"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "STYLE", "(note|hnote|rnote)"),
		klimt.StereotypePatternOptional("STEREO1"),
		uml.Leaf(1, "ACROSS", "(accross|across)"),
		klimt.StereotypePatternOptional("STEREO2"),
		noteColor().Regex(),
	}
}

// executeNoteAcross is FactorySequenceNoteAcrossCommand.executeInternal.
func executeNoteAcross(d *SequenceDiagram, _ *text.LineLocation, arg *regex.Result, display *klimt.Display) command.Result {
	if uml.EqualsIgnoreCase(arg.Str("ACROSS", 0), "accross") {
		return command.Error("Use 'across' instead of 'accross'")
	}
	if display.Size() > 0 {
		_, tryMerge := arg.Get("VMERGE", 0)
		note := newNote(nil, nil, NoteOverSeveral, display, d.CurrentStyleBuilder())
		if err := noteDecorations(d, note, arg, noteColor()); err != nil {
			return colorResult(err)
		}
		d.AddNote(note, tryMerge)
	}
	return command.Ok()
}

// commandNoteAcrossSingle is "note across : text".
func commandNoteAcrossSingle() cmd { return noteSingleLine(noteAcrossParts(), executeNoteAcross) }

// commandNoteAcrossMulti is "note across" ... "end note".
func commandNoteAcrossMulti() cmd {
	return noteMultiLine(noteAcrossParts(), endNote, executeNoteAcross)
}

// noteOnArrowParts is the start of the patterns of a note on the previous
// message.
func noteOnArrowParts() []regex.Part {
	return []regex.Part{regex.Start(),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "STYLE", "(note|hnote|rnote)"),
		klimt.StereotypePatternOptional("STEREO1"),
		uml.Leaf(1, "POSITION", "(right|left|bottom|top)"),
		klimt.StereotypePatternOptional("STEREO2"),
		color.Exp1(),
	}
}

// executeNoteOnArrow is FactorySequenceNoteOnArrowCommand.executeInternal.
func executeNoteOnArrow(d *SequenceDiagram, _ *text.LineLocation, arg *regex.Result, display *klimt.Display) command.Result {
	event := d.LastEventWithNote()
	if event == nil {
		return command.Ok()
	}
	position := notePositionOf(uml.UpperCase(arg.Str("POSITION", 0)))
	var url *klimt.URL
	if s, ok := arg.Get(klimt.URLBuilderKey, 0); ok {
		url = urlBuilder(d).URL(s)
	}
	style := noteStyleOf(arg.Str("STYLE", 0))
	var back color.HColor
	if s, ok := arg.Get("COLOR", 0); ok {
		c, err := getColor(s)
		if err != nil {
			return colorResult(err)
		}
		back = c
	}
	colors := color.EmptyColors().Add(color.TypeBack, back)
	note := newNote(nil, nil, position, display, d.CurrentStyleBuilder())
	note.noteStyle = style
	if stereotypeString, ok := arg.Lazy("STEREO", 0); ok {
		stereotype := klimt.StereotypeBuild(stereotypeString)
		c, err := applyStereotypeForNote(note, colors, stereotype, d.Skin())
		if err != nil {
			return colorResult(err)
		}
		colors = c
		note.SetStereotype(stereotype)
	}
	note.SetURL(url)
	note.SetColors(colors)
	d.AddWarning(event.AddNote(note))
	return command.Ok()
}

// commandNoteOnArrowSingle is "note left : text" after a message.
func commandNoteOnArrowSingle() cmd { return noteSingleLine(noteOnArrowParts(), executeNoteOnArrow) }

// commandNoteOnArrowMulti is "note left" ... "end note" after a message.
func commandNoteOnArrowMulti() cmd {
	return noteMultiLine(noteOnArrowParts(), uml.Compile("^[%s]*end[%s]?note$"), executeNoteOnArrow)
}
