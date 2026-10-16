// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/command/CommandParticipant.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/command/CommandParticipantA.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/command/CommandParticipantA2.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/command/CommandParticipantA3.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/command/CommandParticipantA4.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/command/CommandParticipantMultilines.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/command/CommandBoxStart.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/command/CommandBoxEnd.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/command/CommandUrl.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"errors"

	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/command/regex"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/text"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

type cmd = command.Command[*SequenceDiagram]

// colorResult turns a color error into upstream's result: an unknown
// color is CommandExecutionResult.badColor().
func colorResult(err error) command.Result {
	var noSuch *color.NoSuchColorError
	if errors.As(err, &noSuch) {
		return command.BadColor()
	}
	panic(err)
}

// getColor is HColorSet.getColor; ok is false for an unknown color.
func getColor(s string) (color.HColor, error) { return color.Instance().GetColor(s) }

// orderRegex is CommandParticipant.getOrderRegex.
func orderRegex() regex.Part {
	return regex.Optional(regex.NewConcat(uml.Lit("order"), regex.SpaceOneOrMore(),
		uml.Leaf(1, "ORDER", `(-?\d{1,7})`)))
}

// regexType is CommandParticipant.getRegexType.
func regexType() regex.Part {
	return regex.Or("",
		uml.Leaf(1, "TYPE", "(participant|actor|create|boundary|control|entity|queue|database|collections)"),
		uml.Leaf(1, "CREATE", "create[%s](participant|actor|boundary|control|entity|queue|database|collections)"))
}

// urlBuilder is UrlBuilder(skinParam.getValue("topurl"), UrlMode.STRICT).
func urlBuilder(d *SequenceDiagram) *klimt.URLBuilder {
	topurl, ok := d.Skin().Value("topurl")
	return klimt.NewURLBuilder(topurl, ok, klimt.URLModeStrict)
}

// participantStereotype is the spotted stereotype of a participant.
func participantStereotype(d *SequenceDiagram, label string) (*klimt.Stereotype, error) {
	skin := d.Skin()
	font := skin.Font(nil, false, klimt.FontParamCircledCharacter)
	return klimt.StereotypeBuildSpotted(label, float64(skin.CircledCharacterRadius()), font, skin.IHtmlColorSet())
}

// executeParticipant is CommandParticipant.executeArg.
func executeParticipant(d *SequenceDiagram, location *text.LineLocation, arg *regex.Result) command.Result {
	code := arg.Str("CODE", 0)
	if d.ParticipantsContainsKey(code) {
		d.PutParticipantInLast(code)
		return command.Ok()
	}
	strings := klimt.DisplayNULL()
	if full, ok := arg.Get("FULL", 0); ok {
		strings = klimt.DisplayWithNewlines(d.Pragma(), full)
	}
	typeString1, _ := arg.Get("TYPE", 0)
	typeCreate1, hasCreate := arg.Get("CREATE", 0)
	var typ ParticipantType
	var create bool
	switch {
	case hasCreate:
		typ, create = participantTypeOf(uml.UpperCase(typeCreate1)), true
	case uml.EqualsIgnoreCase(typeString1, "CREATE"):
		typ, create = TypeParticipant, true
	default:
		typ = participantTypeOf(uml.UpperCase(typeString1))
	}
	order := 0
	if s, ok := arg.Get("ORDER", 0); ok {
		order = uml.JavaParseInt(s)
	}
	participant := d.CreateNewParticipant(location, typ, code, strings, order)
	if stereotype, ok := arg.Get("STEREO", 0); ok {
		st, err := participantStereotype(d, stereotype)
		if err != nil {
			return colorResult(err)
		}
		participant.SetStereotype(st, d.Skin().StereotypePositionTop())
	}
	var back color.HColor
	if s, ok := arg.Get("COLOR", 0); ok {
		c, err := getColor(s)
		if err != nil {
			return colorResult(err)
		}
		back = c
	}
	participant.SetSpecificColorTOBEREMOVED(color.TypeBack, back)
	if urlString, ok := arg.Get(klimt.URLBuilderKey, 0); ok {
		participant.SetURL(urlBuilder(d).URL(urlString))
	}
	if create {
		if err := d.Activate(participant, LifeEventCreate, nil, nil); err != "" {
			return command.Error(err)
		}
	}
	return command.Ok()
}

func participantCommand(parts ...regex.Part) cmd {
	return uml.Single(regex.NewConcat(parts...),
		func(d *SequenceDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			return executeParticipant(d, location, arg)
		})
}

func participantTail() []regex.Part {
	return []regex.Part{
		klimt.StereotypePatternOptional("STEREO"),
		orderRegex(),
		regex.SpaceZeroOrMore(),
		klimt.URLBuilderOptional(),
		regex.SpaceZeroOrMore(),
		color.Exp1(), regex.End(),
	}
}

// commandParticipantA is `participant "Long name" as code`.
func commandParticipantA() cmd {
	return participantCommand(append([]regex.Part{regex.Start(),
		regexType(),
		regex.SpaceOneOrMore(),
		regex.Optional(regex.NewConcat(
			uml.Leaf(1, "FULL", "[%g]([^%g]+)[%g]"),
			regex.SpaceOneOrMore(),
			uml.Lit("as"),
			regex.SpaceOneOrMore())),
		uml.Leaf(1, "CODE", "([%pLN_.@]+)")}, participantTail()...)...)
}

// commandParticipantA2 is `participant code as "Long name"`.
func commandParticipantA2() cmd {
	return participantCommand(append([]regex.Part{regex.Start(),
		regexType(),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "CODE", "([%pLN_.@]+)"),
		regex.SpaceOneOrMore(),
		uml.Lit("as"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "FULL", "[%g]([^%g]+)[%g]")}, participantTail()...)...)
}

// commandParticipantA3 is `participant Name as code`.
func commandParticipantA3() cmd {
	return participantCommand(append([]regex.Part{regex.Start(),
		regexType(),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "FULL", "([%pLN_.@]+)"),
		regex.SpaceOneOrMore(),
		uml.Lit("as"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "CODE", "([%pLN_.@]+)")}, participantTail()...)...)
}

// commandParticipantA4 is `participant "code"`.
func commandParticipantA4() cmd {
	return participantCommand(append([]regex.Part{regex.Start(),
		regexType(),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "CODE", "[%g]([^%g]+)[%g]")}, participantTail()...)...)
}

// commandParticipantMultilines is "participant code [" ... "]".
func commandParticipantMultilines() cmd {
	c := &uml.Multilines2[*SequenceDiagram]{
		Start: regex.NewConcat(regex.Start(),
			uml.Leaf(1, "TYPE", "(participant)"),
			regex.SpaceOneOrMore(),
			uml.Leaf(1, "CODE", "([%pLN_.@]+)"),
			klimt.StereotypePatternOptional("STEREO"),
			orderRegex(),
			regex.SpaceZeroOrMore(),
			klimt.URLBuilderOptional(),
			regex.SpaceZeroOrMore(),
			color.Exp1(),
			regex.SpaceZeroOrMore(),
			uml.Lit(`\[`),
			regex.End()),
		Strategy: uml.RemoveStartingQuote,
		TrimEnd:  uml.TrimBoth,
		End:      uml.Compile(`^([^\[\]]*)\]$`),
	}
	c.Exec = func(d *SequenceDiagram, lines command.BlocLines, _ command.ParserPass) command.Result {
		arg := c.StartArg(lines)
		code := arg.Str("CODE", 0)
		if d.ParticipantsContainsKey(code) {
			d.PutParticipantInLast(code)
			return command.Ok()
		}
		lines = uml.SubExtract(lines, 1, 1).RemoveEmptyColumns()
		strings := klimt.DisplayCreateFoo(lines.Lines())
		order := 0
		if s, ok := arg.Get("ORDER", 0); ok {
			order = uml.JavaParseInt(s)
		}
		participant := d.CreateNewParticipant(uml.Location(lines), TypeParticipant, code, strings, order)
		if stereotype, ok := arg.Get("STEREO", 0); ok {
			st, err := participantStereotype(d, stereotype)
			if err != nil {
				return colorResult(err)
			}
			participant.SetStereotype(st, d.Skin().StereotypePositionTop())
		}
		if urlString, ok := arg.Get(klimt.URLBuilderKey, 0); ok {
			participant.SetURL(urlBuilder(d).URL(urlString))
		}
		return command.Ok()
	}
	return c
}

// commandBoxStart is `box ["title"] [<<stereotype>>] [#color]`.
func commandBoxStart() cmd {
	colorParser := color.SimpleColor(color.TypeBack)
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("box"),
		regex.Optional(regex.Or("",
			regex.NewConcat(regex.SpaceOneOrMore(), uml.Leaf(1, "NAME1", "[%g]([^%g]+)[%g]")),
			regex.NewConcat(regex.SpaceOneOrMore(), uml.Leaf(1, "NAME2", "([^#]+)")))),
		klimt.StereotypePatternOptional("STEREO"),
		colorParser.Regex(),
		regex.End()),
		func(d *SequenceDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			argTitle, hasTitle := arg.Lazy("NAME", 0)
			var stereotype *klimt.Stereotype
			if stereo, ok := arg.Get("STEREO", 0); ok {
				stereotype = klimt.StereotypeBuild(stereo)
			}
			colors, err := colorParser.Color(arg, d.Skin().IHtmlColorSet())
			if err != nil {
				return colorResult(err)
			}
			title := ""
			if hasTitle {
				title = argTitle
			}
			d.BoxStart(klimt.DisplayWithNewlines(d.Pragma(), title), colors.Color(color.TypeBack), stereotype)
			return command.Ok()
		})
}

// commandBoxEnd is "end box".
func commandBoxEnd() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("end"),
		regex.SpaceZeroOrMore(),
		uml.Lit("box"), regex.End()),
		func(d *SequenceDiagram, _ *text.LineLocation, _ *regex.Result, _ command.ParserPass) command.Result {
			if !d.IsBoxPending() {
				return command.Error("Missing starting box")
			}
			d.EndBox()
			return command.Ok()
		})
}

// commandURL is "url of code is [[link]]".
func commandURL() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("url"),
		regex.SpaceZeroOrMore(),
		regex.Optional(uml.Lit("of|for")),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "CODE", "([%pLN_.@]+|[%g][^%g]+[%g])"),
		regex.SpaceOneOrMore(),
		regex.Optional(uml.Lit("is")),
		regex.SpaceZeroOrMore(),
		klimt.URLBuilderMandatory(),
		regex.End()),
		func(d *SequenceDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			p := d.OrCreateParticipant(location, arg.Str("CODE", 0))
			p.SetURL(urlBuilder(d).URL(arg.Str(klimt.URLBuilderKey, 0)))
			return command.Ok()
		})
}
