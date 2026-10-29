// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandRobustConcise.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandClock.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandAnalog.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandBinary.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// playerCode is upstream's CommandTimeMessage.PLAYER_CODE: the code of a
// player in the lines that use it. The declarations accept more codes
// (digits first, "@"), which these lines cannot name.
const playerCode = `([\p{L}_][%pLN_.]*)`

// compactPart is the optional "compact" before a declaration.
func compactPart() regex.Part {
	return regex.Optional(regex.NewConcat(uml.Leaf(1, "COMPACT", "(compact)"), regex.SpaceOneOrMore()))
}

// stereotypeOf is the stereotype of the first of the groups that matched,
// nil for none.
func stereotypeOf(arg *regex.Result, names ...string) *klimt.Stereotype {
	for _, name := range names {
		if s, ok := arg.Get(name, 0); ok {
			return klimt.StereotypeBuild(s)
		}
	}
	return nil
}

func playerColor() *color.ColorParser { return color.SimpleColor(color.TypeBack) }

// commandRobustConcise declares a robust, concise or rectangle player, with
// an optional title, stereotype and background color.
func commandRobustConcise() cmd {
	return single([]string{"compact", "concise", "rectangle", "robust"}, regex.NewConcat(regex.Start(),
		compactPart(),
		uml.Leaf(1, "TYPE", "(robust|concise|rectangle)"),
		regex.SpaceOneOrMore(),
		regex.Optional(regex.NewConcat(
			uml.Leaf(1, "FULL", "[%g]([^%g]+)[%g]"),
			klimt.StereotypePatternOptional("STEREOTYPE"),
			uml.Lit("as"),
			regex.SpaceOneOrMore())),
		uml.Leaf(1, "CODE", `([%pLN_.@]+)`),
		klimt.StereotypePatternOptional("STEREOTYPE2"),
		regex.SpaceZeroOrMore(),
		regex.Or("", playerColor().Regex()),
		regex.End()),
		func(d *TimingDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			_, compact := arg.Get("COMPACT", 0)
			code := arg.Str("CODE", 0)
			full := arg.Str("FULL", 0)
			stereotype := stereotypeOf(arg, "STEREOTYPE", "STEREOTYPE2")
			colors, err := playerColor().Color(arg, d.SkinParam().IHtmlColorSet())
			if err != nil {
				return command.BadColor()
			}
			backColor := colors.Color(color.TypeBack)
			switch strings.ToUpper(arg.Str("TYPE", 0)) {
			case "ROBUST":
				d.createPlayerRobust(code, full, compact, stereotype, backColor)
			case "CONCISE":
				d.createPlayerConcise(code, full, compact, stereotype, backColor)
			default:
				d.createPlayerRectangle(code, full, compact, stereotype, backColor)
			}
			return command.Ok()
		})
}

// bigDecimalOrZero is the decimal of a group, 0 when it did not match.
func bigDecimalOrZero(value string, ok bool) jmath.BigDecimal {
	if !ok {
		return jmath.BigDecimalZero
	}
	return jmath.BigDecimalOf(value)
}

// commandClock declares a clock with its period and an optional pulse and
// offset. "compact" is read and ignored, as upstream does.
func commandClock() cmd {
	return single([]string{"clock", "compact"}, regex.NewConcat(regex.Start(),
		compactPart(),
		uml.Leaf(0, "TYPE", "clock"),
		regex.SpaceOneOrMore(),
		regex.Optional(regex.NewConcat(
			uml.Leaf(1, "FULL", "[%g]([^%g]+)[%g]"),
			regex.SpaceOneOrMore(),
			uml.Lit("as"),
			regex.SpaceOneOrMore())),
		uml.Leaf(1, "CODE", `([%pLN_.@]+)`),
		regex.SpaceOneOrMore(),
		uml.Lit("with"),
		regex.SpaceOneOrMore(),
		uml.Lit("period"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "PERIOD", `([0-9]+(?:\.[0-9]+)?)`),
		regex.Optional(regex.NewConcat(
			regex.SpaceOneOrMore(),
			uml.Lit("pulse"),
			regex.SpaceOneOrMore(),
			uml.Leaf(1, "PULSE", `([0-9]+(?:\.[0-9]+)?)`))),
		regex.Optional(regex.NewConcat(
			regex.SpaceOneOrMore(),
			uml.Lit("offset"),
			regex.SpaceOneOrMore(),
			uml.Leaf(1, "OFFSET", `([0-9]+(?:\.[0-9]+)?)`))),
		regex.SpaceZeroOrMore(),
		klimt.StereotypePatternOptional("STEREO"),
		regex.SpaceZeroOrMore(),
		regex.End()),
		func(d *TimingDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			code := arg.Str("CODE", 0)
			period := jmath.BigDecimalOf(arg.Str("PERIOD", 0))
			pulse := bigDecimalOrZero(arg.Get("PULSE", 0))
			offset := bigDecimalOrZero(arg.Get("OFFSET", 0))
			full := arg.Str("FULL", 0)
			d.createPlayerClock(code, full, period, pulse, offset, stereotypeOf(arg, "STEREO"))
			return command.Ok()
		})
}

// commandAnalog declares an analog player, with its required title and
// optional bounds. "compact" is read and ignored, as upstream does.
func commandAnalog() cmd {
	return single([]string{"analog", "compact"}, regex.NewConcat(regex.Start(),
		compactPart(),
		uml.Lit("analog"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "FULL", "[%g]([^%g]+)[%g]"),
		klimt.StereotypePatternOptional("STEREOTYPE"),
		regex.Optional(regex.NewConcat(
			regex.Or("", uml.Lit("between"), uml.Lit("from")),
			regex.SpaceOneOrMore(),
			uml.Leaf(1, "START", `(-?[0-9]*\.?[0-9]+)`),
			regex.SpaceOneOrMore(),
			regex.Or("", uml.Lit("and"), uml.Lit("to")),
			regex.SpaceOneOrMore(),
			uml.Leaf(1, "END", `(-?[0-9]*\.?[0-9]+)`),
			regex.SpaceOneOrMore())),
		uml.Lit("as"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "CODE", `([%pLN_.@]+)`),
		klimt.StereotypePatternOptional("STEREOTYPE2"),
		regex.End()),
		func(d *TimingDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			code := arg.Str("CODE", 0)
			full := arg.Str("FULL", 0)
			p := d.createPlayerAnalog(code, full, stereotypeOf(arg, "STEREOTYPE", "STEREOTYPE2"))
			start, hasStart := arg.Get("START", 0)
			end, hasEnd := arg.Get("END", 0)
			if hasStart && hasEnd {
				p.setBounds(start, end)
			}
			return command.Ok()
		})
}

// commandBinary declares a binary player. "compact" is read and ignored,
// as upstream does.
func commandBinary() cmd {
	return single([]string{"binary", "compact"}, regex.NewConcat(regex.Start(),
		compactPart(),
		uml.Lit("binary"),
		regex.SpaceOneOrMore(),
		regex.Optional(regex.NewConcat(
			uml.Leaf(1, "FULL", "[%g]([^%g]+)[%g]"),
			klimt.StereotypePatternOptional("STEREOTYPE"),
			uml.Lit("as"),
			regex.SpaceOneOrMore())),
		uml.Leaf(1, "CODE", `([%pLN_.@]+)`),
		klimt.StereotypePatternOptional("STEREOTYPE2"),
		regex.End()),
		func(d *TimingDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			d.createPlayerBinary(arg.Str("CODE", 0), arg.Str("FULL", 0), stereotypeOf(arg, "STEREOTYPE", "STEREOTYPE2"))
			return command.Ok()
		})
}
