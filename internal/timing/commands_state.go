// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandDefineStateShort.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandDefineStateLong.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandChangeState.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandChangeStateByPlayerCode.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/command/CommandChangeStateByTime.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// commandDefineStateShort is "<player> has A,B,C": the states of a player,
// without spaces after the commas.
func commandDefineStateShort() cmd {
	return single(nil, regex.NewConcat(regex.Start(),
		uml.Leaf(1, "PLAYER", `([%pLN_.@]+)`),
		regex.SpaceOneOrMore(),
		uml.Lit("has"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "STATE", `([-%pLN_.@]+)`),
		uml.Leaf(3, "STATES", `((,([-%pLN_.@]+))*)`),
		regex.End()),
		func(d *TimingDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			playerCode := arg.Str("PLAYER", 0)
			p := d.player(playerCode)
			if p == nil {
				return command.Error("Unknown " + playerCode)
			}
			stateCode := arg.Str("STATE", 0)
			p.defineState(stateCode, stateCode)
			// StringTokenizer skips the empty tokens.
			for _, token := range strings.FieldsFunc(arg.Str("STATES", 0), func(r rune) bool { return r == ',' }) {
				p.defineState(token, token)
			}
			return command.Ok()
		})
}

// commandDefineStateLong is "<player> has "Long label" as L".
func commandDefineStateLong() cmd {
	return single(nil, regex.NewConcat(regex.Start(),
		uml.Leaf(1, "PLAYER", `([%pLN_.@]+)`),
		regex.SpaceOneOrMore(),
		uml.Lit("has"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "LABEL", "[%g]([^%g]+)[%g]"),
		regex.SpaceOneOrMore(),
		uml.Lit("as"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "STATE", `([%pLN_.@]+)`), regex.End()),
		func(d *TimingDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			playerCode := arg.Str("PLAYER", 0)
			p := d.player(playerCode)
			if p == nil {
				return command.Error("Unknown " + playerCode)
			}
			p.defineState(arg.Str("STATE", 0), arg.Str("LABEL", 0))
			return command.Ok()
		})
}

// stateCode is upstream's CommandChangeState.STATE_CODE.
const stateCode = `([-%pLN_][-%pLN_.]*)`

func stateColor() *color.ColorParser { return color.SimpleColor(color.TypeBack) }

// stateOrHidden is a state: a quoted label, a code, {hidden}, {...}, {-},
// {?} or two states {A,B}.
func stateOrHidden() regex.Part {
	return regex.Or("",
		uml.Leaf(1, "STATE1", "[%g]([^%g]*)[%g]"),
		uml.Leaf(1, "STATE2", stateCode),
		uml.Leaf(1, "STATE3", `(\{hidden\})`),
		uml.Leaf(1, "STATE4", `(\{\.\.\.\})`),
		uml.Leaf(1, "STATE5", `(\{-\})`),
		uml.Leaf(1, "STATE6", `(\{\?\})`),
		uml.Leaf(2, "STATE7", `(?:\{`+stateCode+","+stateCode+`\})`))
}

// stateTail is the state, its color and its comment.
func stateTail() []regex.Part {
	return []regex.Part{
		stateOrHidden(),
		regex.SpaceZeroOrMore(),
		stateColor().Regex(),
		regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewConcat(
			regex.SpaceZeroOrMore(),
			uml.Lit(":"),
			regex.SpaceZeroOrMore(),
			uml.Leaf(1, "COMMENT", "(.*?)"))),
		regex.SpaceZeroOrMore(), regex.End(),
	}
}

// addState is CommandChangeState.addState: the state, its comment and its
// colors at now.
func addState(d *TimingDiagram, arg *regex.Result, p player, now *TimeTick) command.Result {
	var comment *string
	if s, ok := arg.Get("COMMENT", 0); ok {
		comment = &s
	}
	colors, err := stateColor().Color(arg, d.SkinParam().IHtmlColorSet())
	if err != nil {
		return command.BadColor()
	}
	p.setState(now, comment, colors, statesOf(arg)...)
	return command.Ok()
}

func statesOf(arg *regex.Result) []string {
	if state1, ok := arg.Get("STATE7", 0); ok {
		return []string{state1, arg.Str("STATE7", 1)}
	}
	state, _ := arg.Lazy("STATE", 0)
	return []string{state}
}

// commandChangeStateByPlayerCode is "<player> is <state>" at the current
// time, or the initial state before any time. The spaces around "is" are
// optional, as upstream.
func commandChangeStateByPlayerCode() cmd {
	return single(nil, regex.NewConcat(append([]regex.Part{regex.Start(),
		uml.Leaf(1, "CODE", playerCode),
		regex.SpaceZeroOrMore(),
		uml.Lit("is"),
		regex.SpaceZeroOrMore()},
		stateTail()...)...),
		func(d *TimingDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			code := arg.Str("CODE", 0)
			p := d.player(code)
			if p == nil {
				return command.Error(`Unknown "` + code + `"`)
			}
			return addState(d, arg, p, d.getNow())
		})
}

// commandChangeStateByTime is "<time> is <state>" in the block of the last
// player; the time becomes the current time.
func commandChangeStateByTime() cmd {
	return single(nil, regex.NewConcat(append([]regex.Part{regex.Start(),
		regex.SpaceZeroOrMore(),
		expressionAtWithoutArobase("TIME"),
		regex.SpaceZeroOrMore(),
		uml.Lit("is"),
		regex.SpaceZeroOrMore()},
		stateTail()...)...),
		func(d *TimingDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			p := d.getLastPlayer()
			if p == nil {
				return command.Error("Missing @ line before this")
			}
			tick := parseTimeTick("TIME", arg, d)
			d.addTime(tick, "")
			return addState(d, arg, p, tick)
		})
}
