// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/packetdiag/command/CommandNumRange.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package packetdiag

import (
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// commandNumRange is a field: "N: text" for the bit N, "A-B: text" for the
// bits A to B, or "* text" after the last field, with optional attributes
// in brackets.
func commandNumRange() cmd {
	//	^(\*|\d{1,7}(?:-\d{1,7})?)(?::)?\s+(.*?)(?:\s+\[(.*?)\])?$
	return uml.Single(regex.NewConcat(regex.Start(),
		regex.SpaceZeroOrMore(),
		regex.Or("",
			uml.Leaf(1, "RANGE_UNIT", `(\d{1,7})(?::)?`),
			uml.Leaf(2, "RANGE", `(\d{1,7})(?:-(\d{1,7}))(?::)?`),
			uml.Lit(`\*`)),
		uml.Leaf(1, "DESC", `\s+(.*?)`),
		uml.Leaf(1, "ATTRIBUTE", `(?:\s*\[(.*?)\])?`),
		regex.SpaceZeroOrMore(), regex.End()),
		executeNumRange)
}

// intAttribute is the value of the attribute name, or defaultValue when it
// is missing or not an integer.
func intAttribute(name string, attr map[string]string, defaultValue int32) int32 {
	v, ok := attr[name]
	if !ok {
		return defaultValue
	}
	if i, ok := parseInt(v); ok {
		return i
	}
	return defaultValue
}

func lengthAttribute(attr map[string]string) int32   { return intAttribute("len", attr, 1) }
func rotationAttribute(attr map[string]string) int32 { return intAttribute("rotate", attr, 0) }
func heightAttribute(attr map[string]string) int32   { return intAttribute("height", attr, 1) }

func executeNumRange(d *PacketDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
	r1, hasR1 := arg.Get("RANGE_UNIT", 0)
	r2, hasR2 := arg.Get("RANGE", 0)
	r3, hasR3 := arg.Get("RANGE", 1)
	desc, _ := arg.Get("DESC", 0)
	attr := toSet(arg.Get("ATTRIBUTE", 0))

	var start, end int32
	switch {
	case hasR1:
		s, ok := parseInt(r1)
		if !ok {
			return command.Error("Invalid bit start " + r1)
		}
		start = s
		end = start + lengthAttribute(attr) - 1
	case hasR2 && hasR3:
		s, ok1 := parseInt(r2)
		e, ok2 := parseInt(r3)
		if !ok1 || !ok2 {
			return command.Error("Invalid bit range " + r2 + " and " + r3)
		}
		start, end = s, e
	default:
		// auto length and start position, requires diagram context
		if last, ok := d.LastPacketEnd(); ok {
			start = last + 1
		}
		end = start + lengthAttribute(attr) - 1
	}
	// local height and rotation attribute for current packet block
	height := heightAttribute(attr)
	rotation := rotationAttribute(attr)
	d.AddPacketItemRange(start, end, height, desc, rotation)
	return command.Ok()
}
