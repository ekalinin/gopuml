// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandActivate.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandActivate2.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandDeactivateShort.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandAutoactivate.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandReturn.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

func optionalColor(arg *regex.Result, name string) (color.HColor, error) {
	s, ok := arg.Get(name, 0)
	if !ok {
		return nil, nil
	}
	return getColor(s)
}

// commandActivate is "activate|deactivate|destroy|create name [#back] [#line]".
func commandActivate() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "TYPE", "(activate|deactivate|destroy|create)"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "WHO", "([%pLN_.@]+|[%g][^%g]+[%g])"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "BACK", `(#\w+)?`),
		regex.Optional(regex.NewConcat(regex.SpaceOneOrMore(), uml.Leaf(1, "LINE", `(#\w+)`))),
		regex.End()),
		func(d *SequenceDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			typ := lifeEventTypeOf(uml.UpperCase(arg.Str("TYPE", 0)))
			p := d.OrCreateParticipant(location, uml.RemoveQuotes(arg.Str("WHO", 0)))
			back, err := optionalColor(arg, "BACK")
			if err != nil {
				return colorResult(err)
			}
			line, err := optionalColor(arg, "LINE")
			if err != nil {
				return colorResult(err)
			}
			if msg := d.Activate(p, typ, back, line); msg != "" {
				return command.Error(msg)
			}
			return command.Ok()
		})
}

// commandActivate2 is "name++" and "name--".
func commandActivate2() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "NAME", "([%pLN_.@]+)"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "TYPE", `(\+\+|--)`),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "COLOR", `(#\w+)?`),
		regex.End()),
		func(d *SequenceDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			typ := LifeEventDeactivate
			if arg.Str("TYPE", 0) == "++" {
				typ = LifeEventActivate
			}
			p := d.OrCreateParticipant(location, arg.Str("NAME", 0))
			c, err := optionalColor(arg, "COLOR")
			if err != nil {
				return colorResult(err)
			}
			if msg := d.Activate(p, typ, c, nil); msg != "" {
				return command.Error(msg)
			}
			return command.Ok()
		})
}

// commandDeactivateShort is "deactivate" alone: it ends the last
// activation.
func commandDeactivateShort() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(0, "TYPE", "deactivate"),
		regex.SpaceZeroOrMore(),
		regex.End()),
		func(d *SequenceDiagram, _ *text.LineLocation, _ *regex.Result, _ command.ParserPass) command.Result {
			message := d.ActivatingMessage()
			if message == nil {
				return command.Error("Nothing to deactivate.")
			}
			if msg := d.Activate(message.Participant2(), LifeEventDeactivate, nil, nil); msg != "" {
				return command.Error(msg)
			}
			return command.Ok()
		})
}

// commandAutoactivate is "autoactivate on|off".
func commandAutoactivate() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("autoactivate"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "ON", "(off|on)"), regex.End()),
		func(d *SequenceDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			d.SetAutoactivate(uml.EqualsIgnoreCase("on", arg.Str("ON", 0)))
			return command.Ok()
		})
}

// commandReturn is "return label": a dotted message back from the last
// activation. Its first tokens are declared, as upstream declares them:
// the pattern accepts any line ("(.*)" follows "return"), but only the
// lines that start with "&" or the word "return" are tried.
func commandReturn() cmd {
	c := uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "PARALLEL", "(&[%s]*)?"),
		regex.SpaceZeroOrMore(),
		uml.Lit("return"),
		regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewConcat(uml.Leaf(1, "COLOR", `(#\w+)`), regex.SpaceOneOrMore())),
		uml.Leaf(1, "MESSAGE", "(.*)"), regex.End()),
		func(d *SequenceDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			message1 := d.ActivatingMessage()
			doDeactivation := true
			if message1 == nil {
				last, ok := d.LastEventWithDeactivate().(*Message)
				if !ok {
					return command.Error("Nowhere to return to.")
				}
				message1 = last
				doDeactivation = false
			}
			arrow := message1.Base().ArrowConfiguration().WithBody(klimt.ArrowBodyDotted)
			if s, ok := arg.Get("COLOR", 0); ok {
				c, err := getColor(s)
				if err != nil {
					return colorResult(err)
				}
				arrow = arrow.WithColor(c)
			}
			display := klimt.DisplayWithNewlines(d.Pragma(), arg.Str("MESSAGE", 0))
			var message2 AbstractMessage
			if exo1, ok := message1.(*MessageExo); ok {
				number, hasNumber := d.NextMessageNumber()
				message2 = NewMessageExo(d.CurrentStyleBuilder(), exo1.Participant(), exo1.Type().Reverse(), display,
					arrow, number, hasNumber, false, location)
			} else {
				number, hasNumber := d.NextMessageNumber()
				m := NewMessage(d.CurrentStyleBuilder(), message1.Participant2(), message1.Participant1(), display,
					arrow, number, hasNumber, location)
				if _, ok := arg.Get("PARALLEL", 0); ok {
					m.GoParallel()
				}
				message2 = m
			}
			if status := d.AddMessage(message2); !status.IsOK() {
				return status
			}
			if doDeactivation {
				if msg := d.Activate(message1.Participant2(), LifeEventDeactivate, nil, nil); msg != "" {
					return command.Error(msg)
				}
			}
			return command.Ok()
		})
	c.FirstTokens = []string{"&", "return"}
	return c
}
