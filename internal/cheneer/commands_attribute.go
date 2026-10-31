// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cheneer/command/CommandCreateAttribute.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package cheneer

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// commandCreateAttribute is upstream's CommandCreateAttribute: an attribute
// of the current owner, `"Name" as CODE <<stereo>> #color`, which opens
// the block of a composite attribute when it ends with "{". The name may
// hold spaces and ":", so the command reads every line that the commands
// before it leave, apart from the links and "}", as upstream does.
func commandCreateAttribute() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewConcat(
			uml.Leaf(1, "DISPLAY", "[%g]([^%g]+)[%g]"),
			regex.SpaceOneOrMore(),
			uml.Lit("as"),
			regex.SpaceOneOrMore(),
		)),
		uml.Leaf(1, "CODE", "([%pLN%s_.:]+)"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "STEREO", "(<<.*>>)?"),
		regex.SpaceZeroOrMore(),
		color.SimpleColor(color.TypeLine).Regex(),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "COMPOSITE", `(\{)?`),
		regex.End()),
		func(d *ChenEerDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			owner := d.PeekOwner()
			if owner == nil {
				return command.Error("Attribute must be inside an entity, relationship or another attribute")
			}
			idShort := d.CleanID(text.JavaTrim(arg.Str("CODE", 0)))
			quark := d.QuarkInContext(true, owner.Name()+"/"+idShort)
			displayText, ok := arg.Get("DISPLAY", 0)
			if !ok {
				displayText = idShort
			}
			_, composite := arg.Get("COMPOSITE", 0)
			colors, err := color.SimpleColor(color.TypeLine).Color(arg, d.Skin().IHtmlColorSet())
			if err != nil {
				return abel.ColorResult(err)
			}
			if quark.Data() != nil {
				return command.Error("Attribute already exists")
			}
			entity := d.ReallyCreateLeaf(location, quark, klimt.DisplayWithNewlines(d.Pragma(), displayText),
				abel.LeafTypeChenAttribute, abel.USymbolNull)
			if stereo, ok := arg.Get("STEREO", 0); ok {
				entity.SetStereotype(klimt.StereotypeBuild(stereo))
				entity.SetStereostyle(stereo)
			}
			entity.SetColors(colors)
			link := abel.NewLink(location, &d.CucaDiagram, d.SkinParam().CurrentStyleBuilder(), entity, owner,
				decoration.NewLinkType(decoration.LinkDecorNone, decoration.LinkDecorNone),
				abel.LinkArgBuild(klimt.DisplayNULL(), 2))
			link.SetColors(colors)
			d.AddLink(link)
			if composite {
				d.PushOwner(entity)
			}
			return command.Ok()
		})
}
