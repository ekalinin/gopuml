// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/CommandCreateDomain.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/classdiagram/command/GenericRegexProducer.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package descdiagram

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// genericPattern is upstream's GenericRegexProducer.PATTERN: generic
// parameters with up to four levels of nested "<...>".
var genericPattern = `[^\<\>/]` + genericRegex(4)

func genericRegex(level int) string {
	const part1 = `(?:[^\<\>/]|\<`
	const part2 = `\>)*`
	if level == 0 {
		return part1 + part2
	}
	return part1 + genericRegex(level-1) + part2
}

// domainStereotypes are the stereotypes that select the Problem Frames
// symbol of a domain.
var domainStereotypes = []string{"Machine", "Causal", "Designed", "Lexical", "Biddable"}

// commandCreateDomain is upstream's CommandCreateDomain: a Problem Frames
// domain or requirement, `domain "Display" as code <<Machine>>`, a group
// when the line ends with "{".
func commandCreateDomain() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "TYPE", "(requirement|domain)"),
		regex.SpaceOneOrMore(),
		uml.Leaf(2, "DISPLAY", `[%g](.+?)(?:\<(`+genericPattern+`)\>)?[%g]`),
		regex.SpaceOneOrMore(),
		uml.Lit("as"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "CODE", "([a-zA-Z0-9]+)"),
		klimt.StereotypePatternOptional("STEREO"),
		uml.Leaf(1, "GROUP", `(\{)?`),
		regex.End()),
		func(d *DescriptionDiagram, location *text.LineLocation, arg *regex.Result,
			_ command.ParserPass) command.Result {
			typeString := arg.Str("TYPE", 0)
			displayString, _ := arg.Lazy("DISPLAY", 0)
			codeString, ok := arg.Lazy("CODE", 0)
			if !ok {
				codeString = displayString
			}
			stereotype, hasStereotype := arg.Get("STEREO", 0)
			isDomain := uml.EqualsIgnoreCase(typeString, "domain")
			typ, typ2 := abel.GroupTypeRequirement, abel.LeafTypeRequirement
			if isDomain {
				typ, typ2 = abel.GroupTypeDomain, abel.LeafTypeDomain
				for _, name := range domainStereotypes {
					if hasStereotype && uml.EqualsIgnoreCase(stereotype, "<<"+name+">>") {
						typeString = uml.LowerCase(name)
					}
				}
			}
			usymbol := abel.USymbolFromSkin(typeString, d.SkinParam())
			quark := d.QuarkInContext(true, d.CleanID(codeString))
			if quark.Data() != nil {
				return command.Error("Object already exists : " + codeString)
			}
			display := klimt.DisplayWithNewlines(d.Pragma(), displayString)
			var entity *abel.Entity
			if _, ok := arg.Get("GROUP", 0); ok {
				d.GotoGroup(location, quark, display, typ, usymbol)
				entity = d.CurrentGroup()
			} else {
				entity = d.ReallyCreateLeaf(location, quark, display, typ2, usymbol)
			}
			if hasStereotype {
				st, err := abel.SpottedStereotype(&d.CucaDiagram, stereotype)
				if err != nil {
					return abel.ColorResult(err)
				}
				entity.SetStereotype(st)
			}
			// Upstream also reads a URL and a COLOR, which its pattern does not
			// have: the back color is always set to null.
			entity.SetSpecificColorTOBEREMOVED(color.TypeBack, nil)
			return command.Ok()
		})
}
