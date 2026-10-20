// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/statediagram/command/CommandCreatePackageState.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/statediagram/command/CommandCreatePackage2.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package statediagram

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// packageName is the name of a composite state or frame: `Code as
// "display"`, `"display" as Code` or `Code`.
func packageName() regex.Part {
	return regex.Or("",
		regex.NewConcat(
			uml.Leaf(1, "CODE1", "([%pLN_.]+)"),
			regex.SpaceOneOrMore(),
			uml.Lit("as"),
			regex.SpaceOneOrMore(),
			uml.Leaf(1, "DISPLAY1", "[%g]([^%g]+)[%g]")),
		regex.NewConcat(
			regex.Optional(regex.NewConcat(
				uml.Leaf(1, "DISPLAY2", "[%g]([^%g]+)[%g]"),
				regex.SpaceOneOrMore(),
				uml.Lit("as"),
				regex.SpaceOneOrMore())),
			uml.Leaf(1, "CODE2", "([%pLN_.]+)")))
}

// openBlock is the "{" or " begin" that ends the line of a block.
func openBlock() *regex.Leaf { return uml.Lit(`(?:[%s]*\{|[%s]+begin)`) }

// notNull is upstream's getNotNull: v1 when it matched, else v2.
func notNull(arg *regex.Result, v1, v2 string) (string, bool) {
	if s, ok := arg.Get(v1, 0); ok {
		return s, true
	}
	return arg.Get(v2, 0)
}

// lineColors are the colors of a state or frame: the background color,
// and the line color and style of "##[style]color".
func lineColors(d *StateDiagram, arg *regex.Result) (*color.Colors, error) {
	colorSet := d.Skin().IHtmlColorSet()
	colors, err := color.SimpleColor(color.TypeBack).Color(arg, colorSet)
	if err != nil {
		return nil, err
	}
	if s, ok := arg.Get("LINECOLOR", 1); ok {
		lineColor, err := colorSet.GetColor(s)
		if err != nil {
			return nil, err
		}
		colors = colors.Add(color.TypeLine, lineColor)
	}
	if stroke, ok := arg.Get("LINECOLOR", 0); ok {
		colors = colors.AddLegacyStroke(stroke)
	}
	return colors, nil
}

// commandCreatePackageState is `state "Long name" as Code <<stereo>>
// #color {`, which opens a composite state.
func commandCreatePackageState() cmd {
	c := uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("state"),
		regex.SpaceOneOrMore(),
		packageName(),
		regex.SpaceZeroOrMore(),
		abel.TagsLeaf("TAGS1"),
		klimt.StereogroupOptional(),
		abel.TagsLeaf("TAGS2"),
		regex.SpaceZeroOrMore(),
		klimt.URLBuilderOptional(),
		regex.SpaceZeroOrMore(),
		color.SimpleColor(color.TypeBack).Regex(),
		regex.SpaceZeroOrMore(),
		lineColorLeaf(),
		openBlock(),
		regex.End()),
		func(d *StateDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			idShort, _ := notNull(arg, "CODE1", "CODE2")
			quark := d.QuarkInContext(true, idShort)
			display, hasDisplay := notNull(arg, "DISPLAY1", "DISPLAY2")
			groupDisplay := display
			if !hasDisplay {
				groupDisplay = quark.Name()
			}
			d.GotoGroup(location, quark, klimt.DisplayWithNewlines(d.Pragma(), groupDisplay), abel.GroupTypeState,
				abel.USymbolNull)
			p := d.CurrentGroup()
			if hasDisplay {
				p.SetDisplay(klimt.DisplayWithNewlines(d.Pragma(), display))
			}
			stereogroup := klimt.StereogroupBuildFrom(arg)
			p.SetStereotype(stereogroup.BuildStereotype())
			if urlString, ok := arg.Get(klimt.URLBuilderKey, 0); ok {
				p.AddURL(abel.URLBuilderOf(d.Cuca()).URL(urlString))
			}
			colors, err := lineColors(d, arg)
			if err != nil {
				return abel.ColorResult(err)
			}
			inner, err := stereogroup.InnerColors(d.Skin().IHtmlColorSet())
			if err != nil {
				return abel.ColorResult(err)
			}
			p.SetColors(colors.MergeWith(inner))
			abel.AddTags(p, abel.Opt(arg.Lazy("TAGS", 0)))
			return command.Ok()
		})
	c.Passes = allPasses
	return c
}

// commandCreatePackage2 is `frame "Long name" as Code <<stereo>> #color {`,
// which opens a group drawn as a frame.
func commandCreatePackage2() cmd {
	c := uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("frame"),
		regex.SpaceOneOrMore(),
		packageName(),
		klimt.StereotypePatternOptional("STEREOTYPE"),
		klimt.URLBuilderOptional(),
		regex.SpaceZeroOrMore(),
		color.SimpleColor(color.TypeBack).Regex(),
		regex.SpaceZeroOrMore(),
		lineColorLeaf(),
		openBlock(),
		regex.End()),
		func(d *StateDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			idShort, _ := notNull(arg, "CODE1", "CODE2")
			quark := d.QuarkInContext(true, d.CleanID(idShort))
			display, ok := notNull(arg, "DISPLAY1", "DISPLAY2")
			if !ok {
				display = quark.Name()
			}
			d.GotoGroup(location, quark, klimt.DisplayWithNewlines(d.Pragma(), display), abel.GroupTypePackage,
				abel.USymbolFrame)
			p := d.CurrentGroup()
			if stereotype, ok := arg.Get("STEREOTYPE", 0); ok {
				p.SetStereotype(klimt.StereotypeBuild(stereotype))
			}
			if urlString, ok := arg.Get(klimt.URLBuilderKey, 0); ok {
				p.AddURL(abel.URLBuilderOf(d.Cuca()).URL(urlString))
			}
			colors, err := lineColors(d, arg)
			if err != nil {
				return abel.ColorResult(err)
			}
			p.SetColors(colors)
			return command.Ok()
		})
	c.Passes = allPasses
	return c
}
