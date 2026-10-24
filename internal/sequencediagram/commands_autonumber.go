// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandAutonumber.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandAutonumberIncrement.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandAutonumberResume.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandAutonumberStop.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandAutoNewpage.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandNewpage.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/command/CommandIgnoreNewpage.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// commandAutoNewpage is "autonewpage n", which upstream ignores.
func commandAutoNewpage() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("autonewpage"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "VALUE", `(\d+)`), regex.End()),
		func(d *SequenceDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			d.SetAutonewpage(uml.JavaParseInt(arg.Str("VALUE", 0)))
			return command.Ok()
		})
}

// commandAutonumber is `autonumber [start] [step] ["format"]`.
func commandAutonumber() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("autonumber"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "START", `(\d(?:(?:[^%pLN%s]+|\d+)*\d)?)?`),
		regex.Optional(regex.NewConcat(regex.SpaceOneOrMore(), uml.Leaf(1, "STEP", `(\d+)`))),
		regex.Optional(regex.NewConcat(regex.SpaceOneOrMore(), uml.Leaf(1, "FORMAT", "[%g]([^%g]+)[%g]"))),
		regex.SpaceZeroOrMore(), regex.End()),
		func(d *SequenceDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			start := NewDottedNumber("1")
			if s, ok := arg.Get("START", 0); ok {
				start = NewDottedNumber(s)
			}
			inc := 1
			if s, ok := arg.Get("STEP", 0); ok {
				inc = uml.JavaParseInt(s)
			}
			df := "<b>0</b>"
			if s, ok := arg.Get("FORMAT", 0); ok {
				df = s
			}
			format, err := newDecimalFormat(df)
			if err != nil {
				return command.Error("Error in pattern : " + df)
			}
			d.AutonumberGo(start, inc, format)
			return command.Ok()
		})
}

// commandAutonumberStop is "autonumber stop".
func commandAutonumberStop() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("autonumber"),
		regex.SpaceOneOrMore(),
		uml.Lit("stop"),
		regex.SpaceZeroOrMore(), regex.End()),
		func(d *SequenceDiagram, _ *text.LineLocation, _ *regex.Result, _ command.ParserPass) command.Result {
			d.AutonumberStop()
			return command.Ok()
		})
}

// commandAutonumberResume is `autonumber resume [inc] ["format"]`.
func commandAutonumberResume() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("autonumber"),
		regex.SpaceOneOrMore(),
		uml.Lit("resume"),
		regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewConcat(regex.SpaceOneOrMore(), uml.Leaf(1, "INC", `(\d+)`))),
		regex.Optional(regex.NewConcat(regex.SpaceOneOrMore(), uml.Leaf(1, "DF", "[%g]([^%g]+)[%g]"))),
		regex.End()),
		func(d *SequenceDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			var format *decimalFormat
			if df, ok := arg.Get("DF", 0); ok {
				f, err := newDecimalFormat(df)
				if err != nil {
					return command.Error("Error in pattern : " + df)
				}
				format = f
			}
			if inc, ok := arg.Get("INC", 0); ok {
				d.AutoNumber().ResumeIncrement(uml.JavaParseInt(inc), format)
			} else {
				d.AutoNumber().Resume(format)
			}
			return command.Ok()
		})
}

// commandAutonumberIncrement is "autonumber inc [A]".
func commandAutonumberIncrement() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("autonumber"),
		regex.SpaceOneOrMore(),
		uml.Lit("inc"),
		regex.Optional(regex.NewConcat(regex.SpaceOneOrMore(), uml.Leaf(1, "POS", "([A-Za-z])"))),
		regex.SpaceZeroOrMore(), regex.End()),
		func(d *SequenceDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			if position, ok := arg.Get("POS", 0); ok {
				d.AutoNumber().IncrementIntermediateAt(int(uml.LowerCase(position)[0]) - 'a')
			} else {
				d.AutoNumber().IncrementIntermediate()
			}
			return command.Ok()
		})
}

// commandIgnoreNewpage is "ignore newpage".
func commandIgnoreNewpage() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("ignore"),
		regex.SpaceZeroOrMore(),
		uml.Lit("newpage"), regex.End()),
		func(d *SequenceDiagram, _ *text.LineLocation, _ *regex.Result, _ command.ParserPass) command.Result {
			d.IgnoreNewpage()
			return command.Ok()
		})
}

// commandNewpage is "newpage [title]".
func commandNewpage() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("@?"),
		uml.Lit("newpage"),
		regex.Optional(regex.NewConcat(
			uml.Lit("(?:[%s]*:[%s]*|[%s]+)"),
			uml.Leaf(1, "LABEL", "(.*[%pLN_.].*)"))), regex.End()),
		func(d *SequenceDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			label, ok := arg.Get("LABEL", 0)
			strings := klimt.DisplayNULL()
			if ok {
				strings = klimt.DisplayWithNewlines(d.Pragma(), label)
			}
			d.Newpage(uml.DisplayPositionedSingle(location, strings, geom.HorizontalAlignmentCenter,
				geom.VerticalAlignmentTop))
			return command.Ok()
		})
}
