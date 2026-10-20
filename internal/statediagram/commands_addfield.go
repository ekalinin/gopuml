// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/statediagram/command/CommandAddField.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package statediagram

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// commandAddField is "Code : description line", which adds a line to the
// description of a state, creating the state when needed.
func commandAddField() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		regex.Or("",
			uml.Leaf(1, "CODE3", "([%pLN_.]+)"),
			uml.Leaf(1, "CODE4", "[%g]([^%g]+)[%g]")),
		regex.SpaceZeroOrMore(),
		uml.Lit(":"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "FIELD", "(.*)"),
		regex.End()),
		func(d *StateDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			codeString, _ := arg.Lazy("CODE", 0)
			var quark *abel.Quark
			if d.CurrentGroup().Name() == codeString {
				quark = d.CurrentGroup().Quark()
			} else {
				quark = d.QuarkInContext(true, d.CleanID(codeString))
			}
			entity := quark.Data()
			if entity == nil {
				entity = d.ReallyCreateLeaf(location, quark, quarkDisplay(quark), abel.LeafTypeState,
					abel.USymbolNull)
			}
			field, _ := arg.Get("FIELD", 0)
			entity.Bodier().AddFieldOrMethod(field)
			return command.Ok()
		})
}
