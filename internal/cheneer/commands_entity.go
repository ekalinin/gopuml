// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cheneer/command/CommandCreateEntity.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cheneer/command/CommandEndGroup.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package cheneer

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// commandCreateEntity is upstream's CommandCreateEntity: an entity or a
// relationship that opens the block of its attributes,
// `entity "Name" as CODE <<stereo>> #color {`.
func commandCreateEntity() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "TYPE", "(entity|relationship)"),
		regex.SpaceOneOrMore(),
		regex.Optional(regex.NewConcat(
			uml.Leaf(1, "DISPLAY", "[%g]([^%g]+)[%g]"),
			regex.SpaceOneOrMore(),
			uml.Lit("as"),
			regex.SpaceOneOrMore(),
		)),
		uml.Leaf(1, "CODE", "([%pLN_.]+)"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "STEREO", "(<<.+>>)?"),
		regex.SpaceZeroOrMore(),
		color.SimpleColor(color.TypeBack).Regex(),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\{`),
		regex.End()),
		func(d *ChenEerDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			typ := abel.LeafTypeChenEntity
			if arg.Str("TYPE", 0) == "relationship" {
				typ = abel.LeafTypeChenRelationship
			}
			quark := d.QuarkInContext(true, d.CleanID(arg.Str("CODE", 0)))
			displayText, ok := arg.Get("DISPLAY", 0)
			if !ok {
				displayText = quark.Name()
			}
			entity := quark.Data()
			if entity == nil {
				entity = d.ReallyCreateLeaf(location, quark, klimt.DisplayWithNewlines(d.Pragma(), displayText), typ,
					abel.USymbolNull)
			} else if !entity.MuteToTypeSymbol(typ, abel.USymbolNull) {
				return command.Error("Bad name")
			}
			if stereo, ok := arg.Get("STEREO", 0); ok {
				entity.SetStereotype(klimt.StereotypeBuild(stereo))
				entity.SetStereostyle(stereo)
			}
			colors, err := color.SimpleColor(color.TypeBack).Color(arg, d.Skin().IHtmlColorSet())
			if err != nil {
				return abel.ColorResult(err)
			}
			entity.SetColors(colors)
			d.PushOwner(entity)
			return command.Ok()
		})
}

// commandEndGroup is upstream's CommandEndGroup: "}" closes the block of
// the current owner.
func commandEndGroup() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit(`\}`),
		regex.End()),
		func(d *ChenEerDiagram, _ *text.LineLocation, _ *regex.Result, _ command.ParserPass) command.Result {
			if !d.PopOwner() {
				return command.Error("Unbalanced brackets")
			}
			return command.Ok()
		})
}
