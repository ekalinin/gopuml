// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/TimeTickBuilder.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandAtTime.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandAtPlayer.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandScalePixel.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandHideTimeAxis.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandModeCompact.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandTicks.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandPixelHeight.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandUseDateFormat.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"strconv"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// cmd is a command of timing diagrams.
type cmd = command.Command[*TimingDiagram]

// The explain methods of upstream's commands and TimeTickBuilder's
// describeTime, which only its @Explain tooling calls, are not ported.

// anyLine declares that a command may accept any line, as upstream's
// FirstTokens.ANYTHING, whatever its pattern.
type anyLine struct {
	*command.SingleLine[*TimingDiagram]
}

// MandatoryFirstTokens implements command.FirstTokensCommand.
func (anyLine) MandatoryFirstTokens() []string { return nil }

type exec = func(d *TimingDiagram, location *text.LineLocation, arg *regex.Result, pass command.ParserPass) command.Result

// single is a single line command with upstream's first tokens; nil
// tokens accept any line.
func single(firstTokens []string, pattern *regex.Concat, exec exec) cmd {
	c := &command.SingleLine[*TimingDiagram]{Pattern: pattern, Exec: exec, FirstTokens: firstTokens}
	if firstTokens == nil {
		return anyLine{c}
	}
	return c
}

// expressionAtWithoutArobase is a time: a named time with an optional
// offset (:start+5), a date (2024/07/15), an hour (0:15:00), a number
// relative to the current time when it starts with "+" (+50), or a
// multiple of the period of a clock (clk*3).
func expressionAtWithoutArobase(name string) regex.Part {
	return regex.Or("",
		uml.Leaf(2, name+"CODE", `:([%pLN_.]+)([-+][.\d]+)?`),
		uml.Leaf(3, name+"DATE", `(\d+)/(\d+)/(\d+)`),
		uml.Leaf(3, name+"HOUR", `(\d+):(\d+):(\d+)`),
		uml.Leaf(2, name+"DIGIT", `(\+?)(-?\d+\.?\d*)`),
		uml.Leaf(2, name+"CLOCK", `([%pLN_.@]+)\*(\d+)`))
}

func expressionAtWithArobase(name string) *regex.Concat {
	return regex.NewConcat(uml.Lit("@"), expressionAtWithoutArobase(name))
}

func optionalExpressionAtWithArobase(name string) regex.Part {
	return regex.Optional(expressionAtWithArobase(name))
}

// parseTimeTick is the tick of the time name: nil for an unknown named
// time or clock; the current time, possibly nil, when the time is not
// given.
func parseTimeTick(name string, arg *regex.Result, d *TimingDiagram) *TimeTick {
	if code, ok := arg.Get(name+"CODE", 0); ok {
		delta, hasDelta := arg.Get(name+"CODE", 1)
		result := d.codeValue(code)
		if !hasDelta {
			return result
		}
		if result == nil {
			uml.ThrowNPE(`Cannot invoke "net.sourceforge.plantuml.timingdiagram.TimeTick.getTime()" because "result" is null`)
		}
		return newTimeTick(result.time.Add(jmath.BigDecimalOf(delta)), formatDecimal)
	}
	if clockName, ok := arg.Get(name+"CLOCK", 0); ok {
		number := uml.JavaParseInt(arg.Str(name+"CLOCK", 1))
		return d.clockValue(clockName, number)
	}
	if _, ok := arg.Get(name+"HOUR", 0); ok {
		h := int32(uml.JavaParseInt(arg.Str(name+"HOUR", 0)))
		m := int32(uml.JavaParseInt(arg.Str(name+"HOUR", 1)))
		s := int32(uml.JavaParseInt(arg.Str(name+"HOUR", 2)))
		return newTimeTick(jmath.BigDecimalOfInt64(int64(3600*h+60*m+s)), formatHour)
	}
	if _, ok := arg.Get(name+"DATE", 0); ok {
		yy := uml.JavaParseInt(arg.Str(name+"DATE", 0))
		mm := uml.JavaParseInt(arg.Str(name+"DATE", 1))
		dd := uml.JavaParseInt(arg.Str(name+"DATE", 2))
		return createDate(yy, mm, dd, d.timingFormatDate())
	}
	number, ok := arg.Get(name+"DIGIT", 1)
	if !ok {
		return d.getNow()
	}
	isRelative := arg.Str(name+"DIGIT", 0) == "+"
	value := jmath.BigDecimalOf(number)
	if isRelative && d.getNow() != nil {
		value = d.getNow().time.Add(value)
	}
	return newTimeTick(value, formatDecimal)
}

// commandAtTime is "@<time>", optionally named with "as :<name>".
func commandAtTime() cmd {
	return single([]string{"@"}, regex.NewConcat(regex.Start(),
		expressionAtWithArobase("TIME"),
		regex.Optional(regex.NewConcat(
			regex.SpaceOneOrMore(),
			uml.Lit("as"),
			regex.SpaceOneOrMore(),
			uml.Lit(":"),
			uml.Leaf(1, "CODE", `([%pLN_.]+)`))),
		regex.SpaceZeroOrMore(),
		regex.End()),
		func(d *TimingDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			timeTick := parseTimeTick("TIME", arg, d)
			if timeTick == nil {
				return command.Error("What time?")
			}
			d.addTime(timeTick, arg.Str("CODE", 0))
			return command.Ok()
		})
}

// commandAtPlayer is "@<player>": the player of the "<time> is <state>"
// lines that follow.
func commandAtPlayer() cmd {
	return single([]string{"@"}, regex.NewConcat(regex.Start(),
		uml.Lit("@"),
		uml.Leaf(1, "PLAYER", playerCode),
		regex.SpaceZeroOrMore(),
		regex.End()),
		func(d *TimingDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			code := arg.Str("PLAYER", 0)
			p := d.player(code)
			if p == nil {
				return command.Error("No such participant " + code)
			}
			d.setLastPlayer(p)
			return command.Ok()
		})
}

// javaParseLong is Java's Long.parseLong for the digits of a pattern.
func javaParseLong(s string) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		uml.Throw(uml.NumberFormatException, `For input string: "`+s+`"`)
	}
	return v
}

// commandScalePixel is "scale <n> as <m> pixels": a tick every n time
// units (s, m, h, d or y), m pixels apart.
func commandScalePixel() cmd {
	return single([]string{"scale"}, regex.NewConcat(regex.Start(),
		uml.Lit("scale"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "TICK", `(\d+)`),
		regex.Optional(regex.NewConcat(
			regex.SpaceZeroOrOne(),
			uml.Leaf(1, "UNIT", `([smhdDyY])`))),
		regex.SpaceOneOrMore(),
		uml.Lit("as"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "PIXEL", `(\d+)`),
		regex.SpaceOneOrMore(),
		uml.Lit("pixels?"),
		regex.SpaceZeroOrMore(), regex.End()),
		func(d *TimingDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			tick := javaParseLong(arg.Str("TICK", 0)) * unitFactor(arg.Get("UNIT", 0))
			pixel := javaParseLong(arg.Str("PIXEL", 0))
			if tick <= 0 || pixel <= 0 {
				return command.Error("Bad value")
			}
			d.scaleInPixels(tick, pixel)
			return command.Ok()
		})
}

// unitFactor is the number of seconds of a unit; the pattern is case
// insensitive and the units are not, so S, M and H are 1, as upstream.
func unitFactor(unit string, ok bool) int64 {
	if !ok {
		return 1
	}
	switch unit {
	case "m":
		return 60
	case "h":
		return 3600
	case "D", "d":
		return 3600 * 24
	case "Y", "y":
		return 3600 * 8766 // 24 * 365.25 = 8766
	}
	return 1
}

// commandHideTimeAxis is "hide time-axis" and "manual time-axis".
func commandHideTimeAxis() cmd {
	return single([]string{"hide", "manual"}, regex.NewConcat(regex.Start(),
		uml.Leaf(1, "COMMAND", "(hide|manual)"),
		regex.SpaceOneOrMore(),
		uml.Lit("time"),
		uml.Lit(".?"),
		uml.Lit("axis"),
		regex.End()),
		func(d *TimingDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			if uml.EqualsIgnoreCase(arg.Str("COMMAND", 0), "MANUAL") {
				return d.setTimeAxisStategy(TimeAxisManual)
			}
			return d.setTimeAxisStategy(TimeAxisHidden)
		})
}

// commandModeCompact is "mode compact": the players declared after it are
// compact.
func commandModeCompact() cmd {
	return single([]string{"mode"}, regex.NewConcat(regex.Start(),
		uml.Lit("mode"),
		regex.SpaceOneOrMore(),
		uml.Lit("compact"),
		regex.End()),
		func(d *TimingDiagram, _ *text.LineLocation, _ *regex.Result, _ command.ParserPass) command.Result {
			d.goCompactMode()
			return command.Ok()
		})
}

// commandTicks is "<player> ticks num on multiple <n>", which only analog
// players read.
func commandTicks() cmd {
	return single(nil, regex.NewConcat(regex.Start(),
		uml.Leaf(1, "PLAYER", `([%pLN_.@]+)`),
		regex.SpaceOneOrMore(),
		uml.Lit("ticks"),
		regex.SpaceOneOrMore(),
		regex.Or("", uml.Lit("every"),
			regex.NewConcat(uml.Lit("num"), regex.SpaceOneOrMore(), uml.Lit("on"),
				regex.SpaceOneOrMore(), uml.Lit("multiple"))),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "NUM", "([0-9]+)"),
		regex.End()),
		func(d *TimingDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			code := arg.Str("PLAYER", 0)
			p := d.player(code)
			if p == nil {
				return command.Error("No such participant " + code)
			}
			if analog, ok := p.(*playerAnalog); ok {
				analog.setTicks(uml.JavaParseInt(arg.Str("NUM", 0)))
			}
			return command.Ok()
		})
}

// commandPixelHeight is "<player> is <n> pixels height".
func commandPixelHeight() cmd {
	return single(nil, regex.NewConcat(regex.Start(),
		uml.Leaf(1, "PLAYER", `([%pLN_.@]+)`),
		regex.SpaceOneOrMore(),
		uml.Lit("is"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "NUM", "([0-9]+)"),
		regex.SpaceOneOrMore(),
		uml.Lit("pixels?"),
		regex.SpaceOneOrMore(),
		uml.Lit("height"),
		regex.End()),
		func(d *TimingDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			code := arg.Str("PLAYER", 0)
			p := d.player(code)
			if p == nil {
				return command.Error("No such participant " + code)
			}
			p.base().setHeight(uml.JavaParseInt(arg.Str("NUM", 0)))
			return command.Ok()
		})
}

// commandUseDateFormat is "use date format "<pattern>"", the format of the
// dates that follow.
func commandUseDateFormat() cmd {
	return single([]string{"use", "usedate", "usedateformat"}, regex.NewConcat(regex.Start(),
		uml.Lit("use"),
		regex.SpaceZeroOrMore(),
		uml.Lit("date"),
		regex.SpaceZeroOrMore(),
		uml.Lit("format"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "FORMAT", "[%g]([^%g]+)[%g]"),
		regex.End()),
		func(d *TimingDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			return d.useDateFormat(arg.Str("FORMAT", 0))
		})
}
