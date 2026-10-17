// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/CommandPackage.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/CommandEndPackage.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/CommandPackageEmpty.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/command/CommandPackageWithUSymbol.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/command/CommandTogether.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// skinUSymbol is USymbols.fromString(s, actorStyle, componentStyle,
// packageStyle) with the styles of the diagram.
func skinUSymbol(c *CucaDiagram, s string, ok bool) USymbol {
	skin := c.Skin()
	return USymbolFromStyles(s, ok, skin.ActorStyle(), skin.ComponentStyle(), skin.PackageStyle())
}

// CommandPackage is `package name [as code] <<stereotype>> #color {`; a
// stereotype naming a symbol ("<<Node>>") selects the shape of the package.
func CommandPackage[D Diagram]() command.Command[D] {
	c := uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "VISIBILITY", "("+klimt.VisibilityRegexForVisibilityCharacter()+")?"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "TYPE", "(package)"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "NAME", "([%g][^%g]+[%g]|[^#%s{}]*)"),
		regex.Optional(regex.NewConcat(
			regex.SpaceOneOrMore(),
			uml.Lit("as"),
			regex.SpaceOneOrMore(),
			uml.Leaf(1, "AS", "([%pLN_.]+)"))),
		regex.SpaceZeroOrMore(),
		TagsLeaf("TAGS1"),
		klimt.StereotypePatternOptional("STEREOTYPE"),
		TagsLeaf("TAGS2"),
		regex.SpaceZeroOrMore(),
		klimt.URLBuilderOptional(),
		regex.SpaceZeroOrMore(),
		color.SimpleColor(color.TypeBack).Regex(),
		regex.SpaceZeroOrMore(), uml.Lit(`\{`), regex.End()),
		func(d D, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			c := d.Cuca()
			name := uml.RemoveQuotes(arg.Str("NAME", 0))
			as, hasAs := arg.Get("AS", 0)
			if !hasAs && name == "" {
				c.UniqueSequence("##")
				return command.Error("Error in name")
			}
			var quark *Quark
			var display string
			if !hasAs {
				quark = c.QuarkInContext(false, c.CleanID(name))
				display = quark.Name()
			} else {
				quark = c.QuarkInContext(false, c.CleanID(as))
				display = name
			}
			stereotype, hasStereotype := arg.Get("STEREOTYPE", 0)
			usymbol := skinUSymbol(c, stereotype, hasStereotype)
			status := c.GotoGroup(location, quark, klimt.DisplayWithNewlines(c.Pragma(), display), GroupTypePackage,
				usymbol)
			if !status.IsOK() {
				return status
			}
			p := c.CurrentGroup()
			if visibilityString, ok := arg.Get("VISIBILITY", 0); ok {
				p.SetVisibilityModifier(klimt.VisibilityOf(visibilityString+"FOO", false))
			}
			if hasStereotype && usymbol == USymbolNull {
				p.SetStereotype(klimt.StereotypeBuild(stereotype))
			}
			AddTags(p, Opt(arg.Lazy("TAGS", 0)))
			if urlString, ok := arg.Get(klimt.URLBuilderKey, 0); ok {
				p.AddURL(URLBuilderOf(c).URL(urlString))
			}
			colors, err := color.SimpleColor(color.TypeBack).Color(arg, c.Skin().IHtmlColorSet())
			if err != nil {
				return ColorResult(err)
			}
			p.SetColors(colors)
			return command.Ok()
		})
	c.FinalBracket = true
	return c
}

// CommandEndPackage is "}": it closes the current package, namespace or
// together block.
func CommandEndPackage[D Diagram]() command.Command[D] {
	return uml.Single(regex.NewConcat(regex.Start(), uml.Lit(`\}`), regex.End()),
		func(d D, _ *text.LineLocation, _ *regex.Result, _ command.ParserPass) command.Result {
			if !d.Cuca().EndGroup() {
				return command.Error("No package or namespace defined")
			}
			return command.Ok()
		})
}

// CommandPackageEmpty is `package name [as code] #color {}`.
func CommandPackageEmpty[D Diagram]() command.Command[D] {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("package"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "DISPLAY", "([%g][^%g]+[%g]|[^#%s{}]*)"),
		regex.Optional(regex.NewConcat(
			regex.SpaceOneOrMore(),
			uml.Lit("as"),
			regex.SpaceOneOrMore(),
			uml.Leaf(1, "CODE", "([%pLN_.]+)"))),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "COLOR", `(#[0-9a-fA-F]{6}|#?\w+)?`),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\{`),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\}`),
		regex.End()),
		func(d D, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			c := d.Cuca()
			var idShort, display string
			hasDisplay := true
			if code, ok := arg.Get("CODE", 0); !ok {
				if uml.RemoveQuotes(arg.Str("DISPLAY", 0)) == "" {
					idShort = c.UniqueSequence("##")
					hasDisplay = false
				} else {
					idShort = uml.RemoveQuotes(arg.Str("DISPLAY", 0))
					display = idShort
				}
			} else {
				display = uml.RemoveQuotes(arg.Str("DISPLAY", 0))
				idShort = code
			}
			quark := c.QuarkInContext(false, c.CleanID(idShort))
			status := c.GotoGroup(location, quark, klimt.DisplayWithNewlinesString(c.Pragma(), display, hasDisplay),
				GroupTypePackage, USymbolNull)
			if !status.IsOK() {
				return status
			}
			p := c.CurrentGroup()
			if s, ok := arg.Get("COLOR", 0); ok {
				col, err := GetColor(s)
				if err != nil {
					return ColorResult(err)
				}
				p.SetSpecificColorTOBEREMOVED(color.TypeBack, col)
			}
			c.EndGroup()
			return command.Ok()
		})
}

// CommandPackageWithUSymbol is `node "name" as code <<stereotype>> #color {`
// and the other container keywords (rectangle, frame, cloud, database...).
func CommandPackageWithUSymbol[D Diagram]() command.Command[D] {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "SYMBOL",
			"(package|rectangle|hexagon|node|artifact|folder|file|frame|cloud|action|process|database|storage|component|card|queue|stack)"),
		regex.SpaceOneOrMore(),
		regex.Or("",
			regex.NewConcat(
				uml.Leaf(1, "DISPLAY1", "([%g].+?[%g])"),
				regex.Optional(regex.NewConcat(
					regex.SpaceOneOrMore(),
					uml.Leaf(1, "STEREOTYPE1", `(\<\<.+\>\>)`))),
				regex.SpaceZeroOrMore(),
				uml.Lit("as"),
				regex.SpaceOneOrMore(),
				uml.Leaf(1, "CODE1", "([^#%s{}]+)")),
			regex.NewConcat(
				uml.Leaf(1, "CODE2", "([^#%s{}%g]+)"),
				regex.Optional(regex.NewConcat(
					regex.SpaceOneOrMore(),
					uml.Leaf(1, "STEREOTYPE2", `(\<\<.+\>\>)`))),
				regex.SpaceZeroOrMore(),
				uml.Lit("as"),
				regex.SpaceOneOrMore(),
				uml.Leaf(1, "DISPLAY2", "([%g].+?[%g])")),
			regex.NewConcat(
				uml.Leaf(1, "DISPLAY3", "([^#%s{}%g]+)"),
				regex.Optional(regex.NewConcat(
					regex.SpaceOneOrMore(),
					uml.Leaf(1, "STEREOTYPE3", `(\<\<.+\>\>)`))),
				regex.SpaceZeroOrMore(),
				uml.Lit("as"),
				regex.SpaceOneOrMore(),
				uml.Leaf(1, "CODE3", "([^#%s{}%g]+)")),
			uml.Leaf(1, "CODE8", "([%g][^%g]+[%g])"),
			uml.Leaf(1, "CODE9", "([^#%s{}%g]*)")),
		regex.SpaceZeroOrMore(),
		TagsLeaf("TAGS1"),
		klimt.StereotypePatternOptional("STEREOTYPE"),
		TagsLeaf("TAGS2"),
		regex.SpaceZeroOrMore(),
		klimt.URLBuilderOptional(),
		regex.SpaceZeroOrMore(),
		color.SimpleColor(color.TypeBack).Regex(),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\{`), regex.End()),
		func(d D, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			c := d.Cuca()
			codeRaw, _ := arg.Lazy("CODE", 0)
			codeArg := uml.RemoveQuotes(codeRaw)
			id := codeArg
			if codeArg == "" {
				id = c.UniqueSequence("##")
			}
			ident := c.QuarkInContext(false, c.CleanID(id))
			displayArg, hasDisplayArg := RemoveQuotesOpt(arg.Lazy("DISPLAY", 0))
			display, hasDisplay := displayArg, true
			switch {
			case codeArg == "":
				hasDisplay = false
			case !hasDisplayArg:
				display = ident.Name()
			}
			symbol, hasSymbol := arg.Get("SYMBOL", 0)
			usymbol := skinUSymbol(c, symbol, hasSymbol)
			status := c.GotoGroup(location, ident, klimt.DisplayWithNewlinesString(c.Pragma(), display, hasDisplay),
				GroupTypePackage, usymbol)
			if !status.IsOK() {
				return status
			}
			p := c.CurrentGroup()
			if stereotype, ok := arg.Lazy("STEREOTYPE", 0); ok {
				p.SetStereotype(klimt.StereotypeBuildAuto(stereotype, false))
			}
			if urlString, ok := arg.Get(klimt.URLBuilderKey, 0); ok {
				p.AddURL(URLBuilderOf(c).URL(urlString))
			}
			AddTags(p, Opt(arg.Lazy("TAGS", 0)))
			colors, err := color.SimpleColor(color.TypeBack).Color(arg, c.Skin().IHtmlColorSet())
			if err != nil {
				return ColorResult(err)
			}
			p.SetColors(colors)
			return command.Ok()
		})
}

// CommandTogether is "together {": the elements up to "}" are kept close.
func CommandTogether[D Diagram]() command.Command[D] {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("together"),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\{`), regex.End()),
		func(d D, _ *text.LineLocation, _ *regex.Result, _ command.ParserPass) command.Result {
			return d.Cuca().GotoTogether()
		})
}
