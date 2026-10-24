// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/note/CommandFactoryNote.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/note/CommandFactoryNoteOnEntity.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/note/CommandFactoryNoteOnLink.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/note/CommandFactoryTipOnEntity.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/note/SingleMultiFactoryCommand.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/color/Colors.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// colorsApplyStereotypeForNote is upstream's
// Colors.applyStereotypeForNote: the note colors of the stereotype that the
// colors do not set, and the note shadowing of the stereotype.
func colorsApplyStereotypeForNote(colors *color.Colors, stereotype *klimt.Stereotype, skinParam klimt.ISkinParam,
	params ...klimt.ColorParam) (*color.Colors, error) {
	if stereotype == nil {
		uml.ThrowNPE("")
	}
	result := colors
	for _, param := range params {
		var err error
		if result, err = klimt.ColorsApplyStereotype(result, stereotype, skinParam, param); err != nil {
			return nil, err
		}
	}
	return result.WithShadowing(skinParam.ShadowingForNote(stereotype)), nil
}

// noteBody is the multi-line part shared by the note commands: the lines
// between the first and the last, without their common indentation, as a
// display, with the location of the first of them.
func noteBody(lines command.BlocLines) (*klimt.Display, *text.LineLocation) {
	lines = uml.SubExtract(lines, 1, 1).RemoveEmptyColumns()
	return klimt.DisplayCreateFoo(lines.Lines()), uml.Location(lines)
}

func noteFactoryExecute[D Diagram](d D, location *text.LineLocation, arg *regex.Result,
	display *klimt.Display) command.Result {
	c := d.Cuca()
	quark := c.QuarkInContext(false, c.CleanID(arg.Str("CODE", 0)))
	if quark.Data() != nil {
		return command.Error("Note already created: " + quark.Name())
	}
	entity := c.ReallyCreateLeaf(location, quark, display, LeafTypeNote, USymbolNull)
	if s, ok := arg.Get("COLOR", 0); ok {
		col, err := GetColor(s)
		if err != nil {
			return ColorResult(err)
		}
		entity.SetSpecificColorTOBEREMOVED(color.TypeBack, col)
	}
	if stereotypeString, ok := arg.Get("STEREO", 0); ok {
		entity.SetStereotype(klimt.StereotypeBuild(stereotypeString))
	}
	AddTags(entity, Opt(arg.Get("TAGS", 0)))
	return command.Ok()
}

// CommandNoteSingle is CommandFactoryNote.createSingleLine:
// `note "text" as code <<stereotype>> #color`.
func CommandNoteSingle[D Diagram]() command.Command[D] {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("note"),
		regex.SpaceOneOrMore(),
		uml.Lit("[%g]"),
		uml.Leaf(1, "DISPLAY", "([^%g]+)"),
		uml.Lit("[%g]"),
		regex.SpaceOneOrMore(),
		uml.Lit("as"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "CODE", "([%pLN_.]+)"),
		regex.SpaceZeroOrMore(),
		TagsLeaf("TAGS"),
		klimt.StereotypePatternOptional("STEREO"),
		color.Exp1(),
		regex.End()),
		func(d D, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			display := klimt.DisplayWithNewlines(d.Cuca().Pragma(), arg.Str("DISPLAY", 0))
			return noteFactoryExecute(d, location, arg, display)
		})
}

// CommandNoteMulti is CommandFactoryNote.createMultiLine: "note as code"
// ... "end note".
func CommandNoteMulti[D Diagram]() command.Command[D] {
	return Multilines(&uml.Multilines2[D]{
		Start: regex.NewConcat(regex.Start(),
			uml.Lit("note"),
			regex.SpaceOneOrMore(),
			uml.Lit("as"),
			regex.SpaceOneOrMore(),
			uml.Leaf(1, "CODE", "([%pLN_.]+)"),
			regex.SpaceZeroOrMore(),
			TagsLeaf("TAGS"),
			klimt.StereotypePatternOptional("STEREO"),
			color.Exp1(),
			regex.End()),
		Strategy: uml.KeepStartingQuote,
		TrimEnd:  uml.TrimBoth,
		End:      uml.Compile(`^[%s]*end[%s]?note$`),
	}, func(c *uml.Multilines2[D], d D, lines command.BlocLines) command.Result {
		line0 := c.StartArg(lines)
		display, location := noteBody(lines)
		return noteFactoryExecute(d, location, line0, display)
	})
}

// Multilines sets the execution of a multi-line command, which needs the
// command (for its starting pattern), and returns the command.
func Multilines[D command.Diagram](c *uml.Multilines2[D],
	exec func(c *uml.Multilines2[D], d D, lines command.BlocLines) command.Result) *uml.Multilines2[D] {
	c.Exec = func(d D, lines command.BlocLines, _ command.ParserPass) command.Result { return exec(c, d, lines) }
	return c
}

// noteOnEntityHead is the pattern of "note left|right|top|bottom [of
// partial]", shared by the note on entity commands.
func noteOnEntityHead(partialPattern regex.Part) []regex.Part {
	return []regex.Part{regex.Start(),
		uml.Lit("note"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "POSITION", "(right|left|top|bottom)"),
		regex.Or("",
			regex.NewConcat(regex.SpaceOneOrMore(),
				uml.Lit("of"),
				regex.SpaceOneOrMore(), partialPattern),
			uml.Lit("")),
		regex.SpaceZeroOrMore(),
		TagsLeaf("TAGS1"),
		klimt.StereotypePatternOptional("STEREO"),
		TagsLeaf("TAGS2"),
		regex.SpaceZeroOrMore(),
		color.SimpleColor(color.TypeBack).Regex(),
		regex.SpaceZeroOrMore(),
		klimt.URLBuilderOptional(),
	}
}

// CommandNoteOnEntitySingle is CommandFactoryNoteOnEntity.createSingleLine:
// "note left of code : text". partialPattern captures the entity as CODE.
func CommandNoteOnEntitySingle[D Diagram](partialPattern regex.Part) command.Command[D] {
	parts := append(noteOnEntityHead(partialPattern),
		regex.SpaceZeroOrMore(),
		uml.Lit(":"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "NOTE", "(.*)"),
		regex.End())
	return uml.Single(regex.NewConcat(parts...),
		func(d D, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			display := klimt.DisplayWithNewlines(d.Cuca().Pragma(), arg.Str("NOTE", 0))
			return noteOnEntityExecute(d, location, arg, nil, display)
		})
}

// CommandNoteOnEntityMulti is CommandFactoryNoteOnEntity.createMultiLine:
// "note left of code" ... "end note", or "note left of code {" ... "}"
// withBracket.
func CommandNoteOnEntityMulti[D Diagram](partialPattern regex.Part, withBracket bool) command.Command[D] {
	parts := noteOnEntityHead(partialPattern)
	end := uml.Compile(`^[%s]*(end[%s]?note)$`)
	if withBracket {
		parts = append(parts, regex.SpaceZeroOrMore(), uml.Lit(`\{`))
		end = uml.Compile(`^(\})$`)
	}
	parts = append(parts, regex.End())
	return Multilines(&uml.Multilines2[D]{
		Start:    regex.NewConcat(parts...),
		Strategy: uml.KeepStartingQuote,
		TrimEnd:  uml.TrimBoth,
		End:      end,
	}, func(c *uml.Multilines2[D], d D, lines command.BlocLines) command.Result {
		line0 := c.StartArg(lines)
		display, location := noteBody(lines)
		var url *klimt.URL
		if s, ok := line0.Get(klimt.URLBuilderKey, 0); ok {
			url = URLBuilderOf(d.Cuca()).URL(s)
		}
		return noteOnEntityExecute(d, location, line0, url, display)
	})
}

func noteOnEntityExecute[D Diagram](d D, location *text.LineLocation, line0 *regex.Result, url *klimt.URL,
	display *klimt.Display) command.Result {
	c := d.Cuca()
	pos := line0.Str("POSITION", 0)
	var cl1 *Entity
	if code, ok := line0.Get("CODE", 0); !ok {
		cl1 = c.LastEntity()
		if cl1 == nil {
			return command.Error("Nothing to note to")
		}
	} else {
		idShort := c.CleanID(code)
		cl1 = c.QuarkInContext(true, idShort).Data()
		if cl1 == nil {
			return command.Error("Not known: " + idShort)
		}
	}
	position := PositionValueOf(uml.UpperCase(pos)).WithRankdir(c.Skin().Rankdir())
	colors, err := color.SimpleColor(color.TypeBack).Color(line0, c.Skin().IHtmlColorSet())
	if err != nil {
		return ColorResult(err)
	}
	stereotypeString, hasStereotype := line0.Get("STEREO", 0)
	var stereotype *klimt.Stereotype
	if hasStereotype {
		stereotype = klimt.StereotypeBuild(stereotypeString)
		colors, err = colorsApplyStereotypeForNote(colors, stereotype, c.SkinParam(), klimt.ColorParamNoteBackground,
			klimt.ColorParamNoteBorder)
		if err != nil {
			return ColorResult(err)
		}
	}
	if c.Pragma().IsTrue(klimt.PragmaKermor) && cl1.IsGroup() {
		cl1.AddNote(display, position, colors)
		return command.Ok()
	}
	tmp := c.UniqueSequence("GMN")
	quark := c.QuarkInContext(true, tmp)
	note := c.ReallyCreateLeaf(location, quark, display, LeafTypeNote, USymbolNull)
	if hasStereotype {
		note.SetStereotype(stereotype)
	}
	note.SetColors(colors)
	if url != nil {
		note.AddURL(url)
	}
	AddTags(note, Opt(line0.Lazy("TAGS", 0)))
	typ := decoration.NewLinkType(decoration.LinkDecorNone, decoration.LinkDecorNone).GoDashed()
	styleBuilder := c.SkinParam().CurrentStyleBuilder()
	var link *Link
	switch position {
	case PositionRight:
		link = NewLink(location, c, styleBuilder, cl1, note, typ, LinkArgNoDisplay(1))
		link.SetHorizontalSolitary(true)
	case PositionLeft:
		link = NewLink(location, c, styleBuilder, note, cl1, typ, LinkArgNoDisplay(1))
		link.SetHorizontalSolitary(true)
	case PositionBottom:
		link = NewLink(location, c, styleBuilder, cl1, note, typ, LinkArgNoDisplay(2))
	default:
		link = NewLink(location, c, styleBuilder, note, cl1, typ, LinkArgNoDisplay(2))
	}
	c.AddLink(link)
	return command.Ok()
}

// noteOnLinkHead is the pattern of "note [position] on|of link #color".
func noteOnLinkHead() []regex.Part {
	return []regex.Part{regex.Start(),
		uml.Lit("note"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "POSITION", "(right|left|top|bottom)?"),
		regex.SpaceZeroOrMore(),
		regex.NewNamedLeaf(1, "", "(on|of)"),
		regex.SpaceOneOrMore(),
		uml.Lit("link"),
		regex.SpaceZeroOrMore(),
		color.SimpleColor(color.TypeBack).Regex(),
	}
}

// CommandNoteOnLinkSingle is CommandFactoryNoteOnLink.createSingleLine:
// "note on link : text", on the last link.
func CommandNoteOnLinkSingle[D Diagram]() command.Command[D] {
	parts := append(noteOnLinkHead(),
		regex.SpaceZeroOrMore(),
		uml.Lit(":"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "NOTE", "(.*)"), regex.End())
	return uml.Single(regex.NewConcat(parts...),
		func(d D, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			display := klimt.DisplayWithNewlines(d.Cuca().Pragma(), arg.Str("NOTE", 0))
			return noteOnLinkExecute(d, arg, display)
		})
}

// CommandNoteOnLinkMulti is CommandFactoryNoteOnLink.createMultiLine:
// "note on link" ... "end note".
func CommandNoteOnLinkMulti[D Diagram]() command.Command[D] {
	return Multilines(&uml.Multilines2[D]{
		Start:    regex.NewConcat(append(noteOnLinkHead(), regex.End())...),
		Strategy: uml.KeepStartingQuote,
		TrimEnd:  uml.TrimBoth,
		End:      uml.Compile(`^end[%s]?note$`),
	}, func(c *uml.Multilines2[D], d D, lines command.BlocLines) command.Result {
		arg := c.StartArg(lines)
		body := uml.SubExtract(lines, 1, 1).RemoveEmptyColumns()
		if body.Size() > 0 {
			return noteOnLinkExecute(d, arg, klimt.DisplayCreateFoo(body.Lines()))
		}
		return command.Error("No note defined")
	})
}

func noteOnLinkExecute[D Diagram](d D, arg *regex.Result, display *klimt.Display) command.Result {
	c := d.Cuca()
	link := c.LastLink()
	if link == nil {
		return command.Error("No link defined")
	}
	position := PositionBottom
	if s, ok := arg.Get("POSITION", 0); ok {
		position = PositionValueOf(uml.UpperCase(s))
	}
	colors, err := color.SimpleColor(color.TypeBack).Color(arg, c.Skin().IHtmlColorSet())
	if err != nil {
		return ColorResult(err)
	}
	link.AddNote(NewCucaNote(display, position, colors))
	return command.Ok()
}

// CommandTipOnEntityMulti is CommandFactoryTipOnEntity.createMultiLine:
// "note left of Class::member" ... "end note", or with "{" ... "}"
// withBracket. The tips of one side of a class share one TIPS entity.
func CommandTipOnEntityMulti[D Diagram](withBracket bool) command.Command[D] {
	parts := []regex.Part{regex.Start(),
		uml.Lit("note"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "POSITION", "(right|left)"),
		regex.SpaceOneOrMore(),
		uml.Lit("of"),
		regex.SpaceOneOrMore(),
		CodeWithMemberForClass(),
		regex.SpaceZeroOrMore(),
		TagsLeaf("TAGS1"),
		klimt.StereotypePatternOptional("STEREO"),
		TagsLeaf("TAGS2"),
		regex.SpaceZeroOrMore(),
		color.Exp1(),
		regex.SpaceZeroOrMore(),
		klimt.URLBuilderOptional(),
	}
	end := uml.Compile(`^[%s]*(end[%s]?note)$`)
	if withBracket {
		parts = append(parts, regex.SpaceZeroOrMore(), uml.Lit(`\{`))
		end = uml.Compile(`^(\})$`)
	}
	parts = append(parts, regex.End())
	return Multilines(&uml.Multilines2[D]{
		Start:    regex.NewConcat(parts...),
		Strategy: uml.KeepStartingQuote,
		TrimEnd:  uml.TrimBoth,
		End:      end,
	}, func(c *uml.Multilines2[D], d D, lines command.BlocLines) command.Result {
		line0 := c.StartArg(lines)
		display, location := noteBody(lines)
		return tipOnEntityExecute(d, location, line0, display)
	})
}

func tipOnEntityExecute[D Diagram](d D, location *text.LineLocation, line0 *regex.Result,
	display *klimt.Display) command.Result {
	c := d.Cuca()
	pos := line0.Str("POSITION", 0)
	idShort := line0.Str("CODE", 0)
	member := uml.RemoveQuotes(line0.Str("CODE", 1))
	cl1 := c.QuarkInContext(true, idShort).Data()
	if cl1 == nil {
		return command.Error("Nothing to note to")
	}
	position := PositionValueOf(uml.UpperCase(pos)).WithRankdir(c.Skin().Rankdir())
	tmp := uml.RemoveQuotes(idShort + "$$$" + position.String())
	identTip := c.QuarkInContext(true, tmp)
	tips := identTip.Data()
	if tips == nil {
		tips = c.ReallyCreateLeaf(location, identTip, klimt.DisplayWithNewlines(c.Pragma(), ""), LeafTypeTips,
			USymbolNull)
		typ := decoration.NewLinkType(decoration.LinkDecorNone, decoration.LinkDecorNone).Invisible()
		styleBuilder := c.SkinParam().CurrentStyleBuilder()
		var link *Link
		if position == PositionRight {
			link = NewLink(location, c, styleBuilder, cl1, tips, typ, LinkArgNoDisplay(1))
		} else {
			link = NewLink(location, c, styleBuilder, tips, cl1, typ, LinkArgNoDisplay(1))
		}
		c.AddLink(link)
	}
	colors, err := color.SimpleColor(color.TypeBack).Color(line0, c.Skin().IHtmlColorSet())
	if err != nil {
		return ColorResult(err)
	}
	var stereotype *klimt.Stereotype
	if stereotypeString, ok := line0.Get("STEREO", 0); ok {
		stereotype = klimt.StereotypeBuild(stereotypeString)
		colors, err = colorsApplyStereotypeForNote(colors, stereotype, c.SkinParam(), klimt.ColorParamNoteBackground,
			klimt.ColorParamNoteBorder)
		if err != nil {
			return ColorResult(err)
		}
	}
	tips.PutTip(member, display, colors, stereotype)
	return command.Ok()
}
