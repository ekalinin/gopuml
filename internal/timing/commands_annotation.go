// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandTimeMessage.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandNote.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandNoteLong.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandConstraint.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandHighlight.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandArrow.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/descdiagram/command/CommandLinkElement.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// lineStyle is CommandLinkElement.LINE_STYLE.
const lineStyle = `(?:#\w+|dotted|dashed|plain|bold|hidden|norank|single|node|thickness=\d+)` +
	`(?:,#\w+|,dotted|,dashed|,plain|,bold|,hidden|,norank|,single|,node|,thickness=\d+)*`

// colorOrStylePattern is CommandArrow.getColorOrStylePattern: "[#red,dashed]".
const colorOrStylePattern = `(?:\[(` + lineStyle + `)\])?`

// applyStyle is CommandArrow.applyStyle: "dashed" and "dotted" dot the
// line, "bold" changes nothing, "hidden" hides it, and the other styles
// are colors, which fail with "No such color" when they are unknown.
func applyStyle(arrowStyle string, ok bool, config *klimt.ArrowConfiguration) (*klimt.ArrowConfiguration, error) {
	if !ok {
		return config, nil
	}
	for _, s := range strings.Split(arrowStyle, ",") {
		if s == "" {
			continue
		}
		switch {
		case uml.EqualsIgnoreCase(s, "dashed"), uml.EqualsIgnoreCase(s, "dotted"):
			config = config.WithBody(klimt.ArrowBodyDotted)
		case uml.EqualsIgnoreCase(s, "bold"):
		case uml.EqualsIgnoreCase(s, "hidden"):
			config = config.WithBody(klimt.ArrowBodyHidden)
		default:
			c, err := color.Instance().GetColor(s)
			if err != nil {
				return config, err
			}
			config = config.WithColor(c)
		}
	}
	return config, nil
}

// commandTimeMessage is "<player>[@<time>] -> <player>[@<time>] : <label>",
// with an optional style between the dashes and ">".
func commandTimeMessage() cmd {
	return single(nil, regex.NewConcat(regex.Start(),
		uml.Leaf(1, "PART1", playerCode),
		optionalExpressionAtWithArobase("TIME1"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "ARROW_BODY", "(-+)"),
		uml.Leaf(1, "ARROW_STYLE", `(?:\[(`+lineStyle+`)\])?`),
		uml.Leaf(0, "ARROW_HEAD", `\>`),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "PART2", playerCode),
		optionalExpressionAtWithArobase("TIME2"),
		regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewConcat(
			uml.Lit(":"),
			regex.SpaceZeroOrMore(),
			uml.Leaf(1, "MESSAGE", "(.*)"))),
		regex.SpaceZeroOrMore(), regex.End()),
		func(d *TimingDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			player1 := d.player(arg.Str("PART1", 0))
			if player1 == nil {
				return command.Error("No such element: " + arg.Str("PART1", 0))
			}
			player2 := d.player(arg.Str("PART2", 0))
			if player2 == nil {
				return command.Error("No such element: " + arg.Str("PART2", 0))
			}
			tick1 := parseTimeTick("TIME1", arg, d)
			tick2 := parseTimeTick("TIME2", arg, d)
			message, hasMessage := arg.Get("MESSAGE", 0)
			result := d.createTimeMessage(player1, tick1, player2, tick2, message, hasMessage)
			result.ApplyStyle(arg.Lazy("ARROW_STYLE", 0))
			return command.Ok()
		})
}

// positionOf is Position.fromString.
func positionOf(s string) abel.Position { return abel.PositionValueOf(strings.ToUpper(s)) }

// commandNote is "note top of <player> : <text>" at the current time. Its
// tags are read and not used, as upstream does.
func commandNote() cmd {
	return single([]string{"note"}, regex.NewConcat(regex.Start(),
		regex.SpaceZeroOrMore(),
		uml.Lit("note"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "POSITION", "(top|bottom)"),
		regex.SpaceOneOrMore(),
		uml.Lit("of"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "CODE", playerCode),
		regex.SpaceZeroOrMore(),
		uml.Leaf(4, "TAGS", klimt.StereotagPattern()+"?"),
		klimt.StereotypePatternOptional("STEREO"),
		regex.SpaceZeroOrMore(),
		uml.Lit(":"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "NOTE", "(.+)"),
		regex.SpaceZeroOrMore(),
		regex.End()),
		func(d *TimingDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			code := arg.Str("CODE", 0)
			p := d.player(code)
			if p == nil {
				return command.Error(`Unknown "` + code + `"`)
			}
			note := klimt.DisplayWithNewlines(d.Pragma(), arg.Str("NOTE", 0))
			p.base().addNote(d.getNow(), note, positionOf(arg.Str("POSITION", 0)), stereotypeOf(arg, "STEREO"))
			return command.Ok()
		})
}

// endNote is the last line of a long note.
var endNote = uml.Compile(`^end[%s]?note$`)

// commandNoteLong is "note top of <player>" ... "end note".
func commandNoteLong() cmd {
	c := &uml.Multilines2[*TimingDiagram]{
		Start: regex.NewConcat(regex.Start(),
			regex.SpaceZeroOrMore(),
			uml.Lit("note"),
			regex.SpaceOneOrMore(),
			uml.Leaf(1, "POSITION", "(top|bottom)"),
			regex.SpaceOneOrMore(),
			uml.Lit("of"),
			regex.SpaceOneOrMore(),
			uml.Leaf(1, "CODE", playerCode),
			regex.SpaceZeroOrMore(),
			klimt.StereotypePatternOptional("STEREO"),
			regex.SpaceZeroOrMore(),
			regex.End()),
		Strategy: uml.RemoveStartingQuote,
		TrimEnd:  uml.TrimBoth,
		End:      endNote,
	}
	c.Exec = func(d *TimingDiagram, lines command.BlocLines, _ command.ParserPass) command.Result {
		line0 := c.StartArg(lines)
		lines = uml.SubExtract(lines, 1, 1).RemoveEmptyColumns()
		code := line0.Str("CODE", 0)
		p := d.player(code)
		if p == nil {
			return command.Error(`Unknown "` + code + `"`)
		}
		note := klimt.DisplayCreateFoo(lines.Lines())
		p.base().addNote(d.getNow(), note, positionOf(line0.Str("POSITION", 0)), stereotypeOf(line0, "STEREO"))
		return command.Ok()
	}
	return c
}

// commandConstraint is "[<player>]@<time> <-> @<time> : <label>"; the second
// time is relative to the first one.
func commandConstraint() cmd {
	return single(nil, regex.NewConcat(regex.Start(),
		uml.Leaf(2, "PART1", "("+playerCode+")?"),
		expressionAtWithArobase("TIME1"),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\<`),
		uml.Leaf(1, "", "(-+)"),
		uml.Leaf(1, "ARROW_STYLE1", colorOrStylePattern),
		uml.Leaf(1, "", "(-*)"),
		uml.Lit(`\>`),
		regex.SpaceZeroOrMore(),
		expressionAtWithArobase("TIME2"),
		regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewConcat(
			uml.Lit(":"),
			regex.SpaceZeroOrMore(),
			uml.Leaf(1, "MESSAGE", "(.*)"))),
		regex.SpaceZeroOrMore(), regex.End()),
		func(d *TimingDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			var player1 player
			if part1, ok := arg.Get("PART1", 0); !ok {
				player1 = d.getLastPlayer()
				if player1 == nil {
					return command.Error("You have to provide a participant")
				}
			} else {
				player1 = d.player(part1)
				if player1 == nil {
					return command.Error("No such participant " + part1)
				}
			}
			tick1 := parseTimeTick("TIME1", arg, d)
			if tick1 == nil {
				return command.Error("Unknown time label")
			}
			restore := d.getNow()
			d.updateNow(tick1)
			tick2 := parseTimeTick("TIME2", arg, d)
			d.updateNow(restore)
			if tick2 == nil {
				return command.Error("Unknown time label")
			}
			s, ok := arg.Lazy("ARROW_STYLE", 0)
			config, err := applyStyle(s, ok, klimt.ArrowConfigurationWithDirectionBoth())
			if err != nil {
				return command.BadColor()
			}
			message, hasMessage := arg.Get("MESSAGE", 0)
			player1.createConstraint(tick1, tick2, message, hasMessage, config)
			return command.Ok()
		})
}

func highlightColor() *color.ColorParser { return color.SimpleColor(color.TypeBack) }

// commandHighlight is "highlight <time> to <time> <color> : <caption>".
func commandHighlight() cmd {
	return single([]string{"highlight"}, regex.NewConcat(regex.Start(),
		regex.SpaceZeroOrMore(),
		uml.Lit("highlight"),
		regex.SpaceOneOrMore(),
		expressionAtWithoutArobase("FROM"),
		regex.SpaceOneOrMore(),
		uml.Lit("to"),
		regex.SpaceOneOrMore(),
		expressionAtWithoutArobase("TO"),
		regex.SpaceZeroOrMore(),
		highlightColor().Regex(),
		regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewConcat(
			uml.Lit(":"),
			regex.SpaceZeroOrMore(),
			uml.Leaf(1, "CAPTION", "(.*)"))),
		regex.End()),
		func(d *TimingDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			tickFrom := parseTimeTick("FROM", arg, d)
			tickTo := parseTimeTick("TO", arg, d)
			caption, hasCaption := arg.Get("CAPTION", 0)
			display := klimt.DisplayWithNewlinesString(d.Pragma(), caption, hasCaption)
			colors, err := highlightColor().Color(arg, d.SkinParam().IHtmlColorSet())
			if err != nil {
				return command.BadColor()
			}
			return d.highlight(tickFrom, tickTo, display, colors)
		})
}
