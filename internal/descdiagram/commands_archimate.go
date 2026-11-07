// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/descdiagram/command/CommandArchimate.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/descdiagram/command/CommandArchimateMultilines.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/descdiagram/command/CommandArchimatePackage.java
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

// archimateColor is the background color of the archimate commands,
// upstream's color(). The keyword and the optional color are each followed
// by spaces, so that one space without a color does not match, as upstream.
func archimateColor() *color.ColorParser { return color.SimpleColor(color.TypeBack) }

// archimateName is the name of an archimate element or group: upstream's
// four forms of CommandCreateElementFull, with an archimate icon before "as".
func archimateName() regex.Part {
	return regex.Or("",
		uml.Leaf(1, "CODE1", abel.ElementCodeWithQuote),
		regex.NewConcat(
			uml.Leaf(1, "DISPLAY2", abel.ElementDisplay),
			klimt.StereotypePatternOptionalArchimate("STEREOTYPE2"),
			uml.Lit("as"),
			regex.SpaceOneOrMore(),
			uml.Leaf(1, "CODE2", abel.ElementCode)),
		regex.NewConcat(
			uml.Leaf(1, "CODE3", abel.ElementCode),
			klimt.StereotypePatternOptionalArchimate("STEREOTYPE3"),
			uml.Lit("as"),
			regex.SpaceZeroOrMore(),
			uml.Leaf(1, "DISPLAY3", abel.ElementDisplay)),
		regex.NewConcat(
			uml.Leaf(1, "DISPLAY4", abel.ElementDisplayWithoutQuote),
			klimt.StereotypePatternOptionalArchimate("STEREOTYPE4"),
			uml.Lit("as"),
			regex.SpaceOneOrMore(),
			uml.Leaf(1, "CODE4", abel.ElementCode)))
}

// archimateDisplay is the display of a matched name without its quotes, or
// the name of the quark when there is none.
func archimateDisplay(arg *regex.Result, quark *abel.Quark) string {
	if display, ok := arg.Lazy("DISPLAY", 0); ok {
		return uml.RemoveQuotes(display)
	}
	return quark.Name()
}

// setArchimateIconAndColors gives an entity the "<<$archimate/icon>>"
// stereotype of the icon, which draws the internal sprite of that name, and
// the background color of the command.
func setArchimateIconAndColors(d *DescriptionDiagram, entity *abel.Entity, arg *regex.Result) command.Result {
	if icon, ok := arg.Lazy("STEREOTYPE", 0); ok {
		st, err := abel.SpottedStereotype(&d.CucaDiagram,
			"<<$archimate/"+klimt.StereotypePatternRemoveChevronBrackets(icon)+">>")
		if err != nil {
			return abel.ColorResult(err)
		}
		entity.SetStereotype(st)
	}
	colors, err := archimateColor().Color(arg, d.Skin().IHtmlColorSet())
	if err != nil {
		return abel.ColorResult(err)
	}
	entity.SetColors(colors)
	return command.Ok()
}

// commandArchimate is upstream's CommandArchimate: `archimate #color
// "Display" <<icon>> as code` creates an element with the archimate symbol,
// or updates the element of that name.
func commandArchimate() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(0, "SYMBOL", "archimate"),
		regex.SpaceOneOrMore(),
		archimateColor().Regex(),
		regex.SpaceOneOrMore(),
		archimateName(),
		klimt.StereotypePatternOptionalArchimate("STEREOTYPE"),
		regex.End()),
		func(d *DescriptionDiagram, location *text.LineLocation, arg *regex.Result,
			_ command.ParserPass) command.Result {
			codeRaw, _ := arg.Lazy("CODE", 0)
			quark := d.QuarkInContext(true, d.CleanID(codeRaw))
			display := archimateDisplay(arg, quark)
			entity := quark.Data()
			if entity == nil {
				entity = d.ReallyCreateLeaf(location, quark, klimt.DisplayWithNewlines(d.Pragma(), display),
					abel.LeafTypeDescription, abel.USymbolArchimate)
			}
			entity.SetDisplay(klimt.DisplayWithNewlines(d.Pragma(), display))
			return setArchimateIconAndColors(d, entity, arg)
		})
}

// commandArchimateMultilines is upstream's CommandArchimateMultilines:
// `archimate #color code <<icon>> [` ... `]` creates an element whose text
// is the lines between the brackets, with the rectangle symbol, as upstream.
// The URL, the text after "[" and the text before "]" are not used, as
// upstream.
func commandArchimateMultilines() cmd {
	return abel.Multilines(&uml.Multilines2[*DescriptionDiagram]{
		Start: regex.NewConcat(regex.Start(),
			uml.Lit("archimate"),
			regex.SpaceOneOrMore(),
			archimateColor().Regex(),
			regex.SpaceOneOrMore(),
			uml.Leaf(1, "CODE", "([%pLN_.]+)"),
			klimt.StereotypePatternOptionalArchimate("STEREOTYPE"),
			klimt.URLBuilderOptional(),
			regex.SpaceZeroOrMore(),
			color.Exp1(),
			regex.SpaceZeroOrMore(),
			uml.Lit(`\[`),
			uml.Leaf(1, "DESC", "(.*)"),
			regex.End()),
		Strategy: uml.RemoveStartingQuote,
		TrimEnd:  uml.TrimBoth,
		End:      uml.Compile(`^(.*)\]$`),
	}, func(c *uml.Multilines2[*DescriptionDiagram], d *DescriptionDiagram, lines command.BlocLines) command.Result {
		lines = lines.Trim()
		line0 := c.StartArg(lines)
		quark := d.QuarkInContext(false, d.CleanID(line0.Str("CODE", 0)))
		if quark.Data() != nil {
			return command.Error("Already exists " + quark.Name())
		}
		entity := d.ReallyCreateLeaf(uml.Location(lines), quark,
			klimt.DisplayWithNewlines(klimt.PragmaCreateEmpty(), quark.Name()), abel.LeafTypeDescription,
			abel.USymbolRectangle)
		entity.SetDisplay(klimt.DisplayCreateFoo(uml.SubExtract(lines, 1, 1).Lines()))
		return setArchimateIconAndColors(d, entity, line0)
	})
}

// commandArchimatePackage is upstream's CommandArchimatePackage: `archimate
// #color "Display" <<icon>> as code {` opens a group with the archimate
// symbol.
func commandArchimatePackage() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(0, "SYMBOL", "archimate"),
		regex.SpaceOneOrMore(),
		archimateColor().Regex(),
		regex.SpaceOneOrMore(),
		archimateName(),
		klimt.StereotypePatternOptionalArchimate("STEREOTYPE"),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\{`),
		regex.End()),
		func(d *DescriptionDiagram, location *text.LineLocation, arg *regex.Result,
			_ command.ParserPass) command.Result {
			codeRaw, _ := arg.Lazy("CODE", 0)
			quark := d.QuarkInContext(true, d.CleanID(codeRaw))
			display := archimateDisplay(arg, quark)
			status := d.GotoGroup(location, quark, klimt.DisplayWithNewlines(d.Pragma(), display),
				abel.GroupTypePackage, abel.USymbolArchimate)
			if !status.IsOK() {
				return status
			}
			p := d.CurrentGroup()
			p.SetDisplay(klimt.DisplayWithNewlines(d.Pragma(), display))
			return setArchimateIconAndColors(d, p, arg)
		})
}
