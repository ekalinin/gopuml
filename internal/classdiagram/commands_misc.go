// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/classdiagram/command/CommandAddMethod.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/classdiagram/command/CommandStereotype.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/classdiagram/command/CommandAllowMixing.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/classdiagram/command/CommandLayoutNewLine.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/classdiagram/command/CommandDiamondAssociation.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandNamespace.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandNamespace2.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/command/CommandNamespaceEmpty.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package classdiagram

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// commandAddMethod is "Name : member": the member is added to the class,
// created when it does not exist.
func commandAddMethod() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "NAME", "([%pLN_.]+|[%g][^%g]+[%g])"),
		regex.SpaceOneOrMore(),
		uml.Lit(":"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "DATA", "(.*)"),
		regex.End()),
		func(d *ClassDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			quark := d.QuarkInContext(true, d.CleanID(arg.Str("NAME", 0)))
			entity := quark.Data()
			if entity == nil {
				entity = d.ReallyCreateLeaf(location, quark,
					klimt.DisplayWithNewlines(klimt.PragmaCreateEmpty(), quark.Name()), abel.LeafTypeClass,
					abel.USymbolNull)
			}
			field := arg.Str("DATA", 0)
			if field != "" && klimt.VisibilityIsVisibilityCharacter(field) {
				d.SetVisibilityModifierPresent(true)
			}
			entity.Bodier().AddFieldOrMethod(field)
			return command.Ok()
		})
}

// commandStereotype is "Name <<stereotype>>".
func commandStereotype() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "NAME", "([%pLN_.]+|[%g][^%g]+[%g])"),
		regex.SpaceZeroOrMore(),
		klimt.StereotypePatternMandatory("STEREO"),
		regex.End()),
		func(d *ClassDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			quark := d.QuarkInContext(true, d.CleanID(arg.Str("NAME", 0)))
			entity := quark.Data()
			if entity == nil {
				return command.Error("No such class " + quark.Name())
			}
			st, err := abel.SpottedStereotype(&d.CucaDiagram, arg.Str("STEREO", 0))
			if err != nil {
				return abel.ColorResult(err)
			}
			entity.SetStereotype(st)
			return command.Ok()
		})
}

// commandAllowMixing is "allowmixing" (or "allow_mixing"): description
// elements may be used.
func commandAllowMixing() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("allow"),
		uml.Lit("_?"),
		uml.Lit("mixing"),
		regex.End()),
		func(d *ClassDiagram, _ *text.LineLocation, _ *regex.Result, _ command.ParserPass) command.Result {
			d.SetAllowMixing(true)
			return command.Ok()
		})
}

// commandLayoutNewLine is "layout_new_line".
func commandLayoutNewLine() cmd {
	return uml.Single(regex.NewConcat(regex.Start(), uml.Lit("layout_new_line"), regex.End()),
		func(d *ClassDiagram, _ *text.LineLocation, _ *regex.Result, _ command.ParserPass) command.Result {
			d.LayoutNewLine()
			return command.Ok()
		})
}

// commandDiamondAssociation is "<> code": the diamond of an n-ary
// association.
func commandDiamondAssociation() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit(`\<\>`),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "CODE", "([%pLN_.]+)"),
		regex.End()),
		func(d *ClassDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			quark := d.QuarkInContext(true, d.CleanID(arg.Str("CODE", 0)))
			if quark.Data() != nil {
				return command.Error("Already existing : " + quark.Name())
			}
			d.ReallyCreateLeaf(location, quark, klimt.DisplayWithNewlines(d.Pragma(), ""), abel.LeafTypeAssociation,
				abel.USymbolNull)
			return command.Ok()
		})
}

// namespaceRegex is upstream's CommandNamespace.NAMESPACE_REGEX.
const namespaceRegex = `([%pLN_][-%pLN_.:\\/]*)`

// namespaceTail sets the link and the background color of the opened
// namespace.
func namespaceTail(d *ClassDiagram, p *abel.Entity, arg *regex.Result) command.Result {
	if urlString, ok := arg.Get(klimt.URLBuilderKey, 0); ok {
		p.AddURL(abel.URLBuilderOf(&d.CucaDiagram).URL(urlString))
	}
	if s, ok := arg.Get("COLOR", 0); ok {
		c, err := abel.GetColor(s)
		if err != nil {
			return abel.ColorResult(err)
		}
		p.SetSpecificColorTOBEREMOVED(color.TypeBack, c)
	}
	return command.Ok()
}

// commandNamespace is "namespace name <<stereotype>> #color {".
func commandNamespace() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("namespace"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "NAME", namespaceRegex),
		klimt.StereotypePatternOptional("STEREOTYPE"),
		klimt.URLBuilderOptional(),
		regex.SpaceZeroOrMore(),
		color.Exp1(),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\{`), regex.End()),
		func(d *ClassDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			quark := d.QuarkInContext(false, d.CleanID(arg.Str("NAME", 0)))
			stereotype, hasStereotype := arg.Get("STEREOTYPE", 0)
			skin := d.Skin()
			usymbol := abel.USymbolFromStyles(stereotype, hasStereotype, skin.ActorStyle(), skin.ComponentStyle(),
				skin.PackageStyle())
			status := d.GotoGroup(location, quark, klimt.DisplayWithNewlines(d.Pragma(), quark.Name()),
				abel.GroupTypePackage, usymbol)
			if !status.IsOK() {
				return status
			}
			p := d.CurrentGroup()
			if hasStereotype && usymbol == abel.USymbolNull {
				p.SetStereotype(klimt.StereotypeBuild(stereotype))
			}
			return namespaceTail(d, p, arg)
		})
}

// commandNamespace2 is `namespace "display" as name <<stereotype>> #color
// {`.
func commandNamespace2() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("namespace"),
		regex.SpaceOneOrMore(),
		uml.Lit("[%g]"),
		uml.Leaf(1, "DISPLAY", "([^%g]+)"),
		uml.Lit("[%g]"),
		regex.SpaceOneOrMore(),
		uml.Lit("as"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "NAME", namespaceRegex),
		klimt.StereotypePatternOptional("STEREOTYPE"),
		klimt.URLBuilderOptional(),
		regex.SpaceZeroOrMore(),
		color.Exp1(),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\{`), regex.End()),
		func(d *ClassDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			quark := d.QuarkInContext(false, d.CleanID(arg.Str("NAME", 0)))
			display := klimt.DisplayWithNewlinesString(d.Pragma(), abel.Opt(arg.Lazy("DISPLAY", 0)).S,
				abel.Opt(arg.Lazy("DISPLAY", 0)).OK)
			status := d.GotoGroup(location, quark, display, abel.GroupTypePackage, abel.USymbolNull)
			if !status.IsOK() {
				return status
			}
			p := d.CurrentGroup()
			if stereotype, ok := arg.Get("STEREOTYPE", 0); ok {
				p.SetStereotype(klimt.StereotypeBuild(stereotype))
			}
			return namespaceTail(d, p, arg)
		})
}

// commandNamespaceEmpty is "namespace name <<stereotype>> #color {}".
func commandNamespaceEmpty() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("namespace"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "NAME", namespaceRegex),
		klimt.StereotypePatternOptional("STEREOTYPE"),
		klimt.URLBuilderOptional(),
		regex.SpaceZeroOrMore(),
		color.Exp1(),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\{`),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\}`),
		regex.End()),
		func(d *ClassDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			quark := d.QuarkInContext(false, d.CleanID(arg.Str("NAME", 0)))
			if quark.Data() != nil {
				return command.Error("Already exists " + quark.Name())
			}
			display := klimt.DisplayWithNewlines(d.Pragma(), quark.QualifiedName())
			status := d.GotoGroup(location, quark, display, abel.GroupTypePackage, abel.USymbolNull)
			if !status.IsOK() {
				return status
			}
			p := d.CurrentGroup()
			if stereotype, ok := arg.Get("STEREOTYPE", 0); ok {
				p.SetStereotype(klimt.StereotypeBuild(stereotype))
			}
			if r := namespaceTail(d, p, arg); !r.IsOK() {
				return r
			}
			d.EndGroup()
			return command.Ok()
		})
}
