// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/packetdiag/command/CommandPacketDiagStart.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/packetdiag/command/CommandPacketDiagEnd.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/packetdiag/command/CommandColWidth.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/packetdiag/command/CommandNodeHeight.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/packetdiag/command/CommandScaleDirection.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/packetdiag/command/CommandScaleInterval.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/packetdiag/command/CommandSameHeight.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package packetdiag

import (
	"strconv"
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

type cmd = command.Command[*PacketDiagram]

// commandPacketDiagStart is the optional "packetdiag {" that opens the
// fields. It accepts an empty line too.
func commandPacketDiagStart() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "TYPE", `(packetdiag)?`),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\{?`), regex.End()),
		func(*PacketDiagram, *text.LineLocation, *regex.Result, command.ParserPass) command.Result {
			return command.Ok()
		})
}

// commandPacketDiagEnd is the optional "}" that closes the fields.
func commandPacketDiagEnd() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\}?`),
		regex.SpaceZeroOrMore(), regex.End()),
		func(*PacketDiagram, *text.LineLocation, *regex.Result, command.ParserPass) command.Result {
			return command.Ok()
		})
}

// directive returns the pattern "name = value" with an optional ";", value
// being captured by group.
func directive(name, group, value string) *regex.Concat {
	return regex.NewConcat(regex.Start(),
		uml.Lit(name),
		regex.SpaceZeroOrMore(),
		uml.Lit(`=`),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, group, `(`+value+`);?`),
		regex.SpaceZeroOrMore(), regex.End())
}

// parseInt is Java's Integer.parseInt for the digits the patterns capture;
// ok is false where Java throws NumberFormatException.
func parseInt(s string) (int32, bool) {
	v, err := strconv.ParseInt(s, 10, 32)
	return int32(v), err == nil
}

// commandColWidth is "colwidth = N", the number of bits of a row.
func commandColWidth() cmd {
	return uml.Single(directive("colwidth", "WIDTH", `\d{1,3}`),
		func(d *PacketDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			if width, ok := arg.Get("WIDTH", 0); ok {
				w, ok := parseInt(width)
				if !ok {
					return command.Error("Width must be an integer")
				}
				d.SetColWidth(w)
			}
			return command.Ok()
		})
}

// commandNodeHeight is "node_height = N", the height of a bit in pixels.
func commandNodeHeight() cmd {
	return uml.Single(directive("node_height", "HEIGHT", `\d{1,3}`),
		func(d *PacketDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			// The original python implementation takes int, but it really should be a float number
			h, ok := parseInt(arg.Str("HEIGHT", 0))
			if !ok {
				return command.Error("Height must be an integer")
			}
			d.UpdateNodeHeight(h)
			return command.Ok()
		})
}

// commandScaleDirection is "scale_direction = ltr" or "rtl".
func commandScaleDirection() cmd {
	return uml.Single(directive("scale_direction", "DIR", `ltr|rtl`),
		func(d *PacketDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			dir := "LTR"
			if v, ok := arg.Get("DIR", 0); ok {
				dir = strings.ToUpper(v)
			}
			d.SetScaleDirection(dir)
			return command.Ok()
		})
}

// commandScaleInterval is "scale_interval = N", the interval of the
// numbered bits.
func commandScaleInterval() cmd {
	return uml.Single(directive("scale_interval", "INTERVAL", `\d+`),
		func(d *PacketDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			interval, ok := parseInt(arg.Str("INTERVAL", 0))
			if !ok {
				return command.Error("Scale interval invalid")
			}
			d.UpdateScaleInterval(interval)
			return command.Ok()
		})
}

// commandSameHeight is "same_height = true" or "false".
func commandSameHeight() cmd {
	return uml.Single(directive("same_height", "VALUE", `true|false`),
		func(d *PacketDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			d.SetSameHeight(strings.EqualFold(arg.Str("VALUE", 0), "true"))
			return command.Ok()
		})
}
