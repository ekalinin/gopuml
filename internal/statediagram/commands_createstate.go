// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/statediagram/command/CommandCreateState.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/stereo/Stereogroup.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package statediagram

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

type cmd = command.Command[*StateDiagram]

// allPasses are the passes of the state declarations.
var allPasses = []command.ParserPass{command.One, command.Two, command.Three}

// lineColorLeaf is the "##[style]color" of states.
func lineColorLeaf() regex.Part {
	return regex.Optional(uml.Leaf(2, "LINECOLOR", `##(?:\[(dotted|dashed|bold)\])?(\w+)?`))
}

// commandCreateState is `state "Long name" as Code <<stereo>> #color :
// description`.
func commandCreateState() cmd {
	c := uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("state"),
		regex.SpaceOneOrMore(),
		regex.Or("",
			regex.NewConcat(
				uml.Leaf(1, "CODE1", "([%pLN_.]+)"),
				regex.SpaceOneOrMore(),
				uml.Lit("as"),
				regex.SpaceOneOrMore(),
				uml.Leaf(1, "DISPLAY1", "[%g]([^%g]+)[%g]")),
			regex.NewConcat(
				uml.Leaf(1, "DISPLAY2", "[%g]([^%g]+)[%g]"),
				regex.SpaceOneOrMore(),
				uml.Lit("as"),
				regex.SpaceOneOrMore(),
				uml.Leaf(1, "CODE2", "([%pLN_.]+)")),
			uml.Leaf(1, "CODE3", "([%pLN_.]+)"),
			uml.Leaf(1, "CODE4", "[%g]([^%g]+)[%g]")),
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
		regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewConcat(
			uml.Lit(":"),
			regex.SpaceZeroOrMore(),
			uml.Leaf(1, "ADDFIELD", "(.*)"))),
		regex.End()),
		func(d *StateDiagram, location *text.LineLocation, arg *regex.Result, pass command.ParserPass) command.Result {
			idShort, _ := arg.Lazy("CODE", 0)
			quark := d.QuarkInContext(true, d.CleanID(idShort))
			display, ok := arg.Lazy("DISPLAY", 0)
			if !ok {
				display = quark.Name()
			}
			stereogroup := klimt.StereogroupBuildFrom(arg)
			typ := stereogroupLeafType(stereogroup)
			if typ == abel.LeafTypeNull {
				typ = abel.LeafTypeState
			}
			if !d.checkConcurrentStateOk(quark) {
				return command.Error("The state " + quark.Name() +
					" has been created in a concurrent state : it cannot be used here.")
			}
			ent := quark.Data()
			if ent == nil {
				ent = d.ReallyCreateLeaf(location, quark, klimt.DisplayWithNewlines(d.Pragma(), display), typ,
					abel.USymbolNull)
			} else {
				d.SetLastEntity(ent)
			}
			d.ensureParentState(location, quark)
			if pass != command.One {
				return command.Ok()
			}
			ent.SetDisplay(klimt.DisplayWithNewlines(d.Pragma(), display))
			ent.SetStereotype(stereogroup.BuildStereotype())
			if urlString, ok := arg.Get(klimt.URLBuilderKey, 0); ok {
				ent.AddURL(abel.URLBuilderOf(d.Cuca()).URL(urlString))
			}
			colorSet := d.Skin().IHtmlColorSet()
			colors, err := color.SimpleColor(color.TypeBack).Color(arg, colorSet)
			if err != nil {
				return abel.ColorResult(err)
			}
			if s, ok := arg.Get("LINECOLOR", 1); ok {
				lineColor, err := colorSet.GetColor(s)
				if err != nil {
					return abel.ColorResult(err)
				}
				colors = colors.Add(color.TypeLine, lineColor)
			}
			if stroke, ok := arg.Get("LINECOLOR", 0); ok {
				colors = colors.AddLegacyStroke(stroke)
			}
			inner, err := stereogroup.InnerColors(colorSet)
			if err != nil {
				return abel.ColorResult(err)
			}
			colors = colors.MergeWith(inner)
			ent.SetColors(colors)
			if addFields, ok := arg.Get("ADDFIELD", 0); ok {
				ent.Bodier().AddFieldOrMethod(addFields)
			}
			abel.AddTags(ent, abel.Opt(arg.Lazy("TAGS", 0)))
			if ent.ParentContainer().IsRoot() && !ent.EntityPosition().IsNormal() {
				return command.Error("You cannot use this stereotype here")
			}
			return command.Ok()
		})
	c.Passes = allPasses
	return c
}

// stereogroupLeafType is upstream's Stereogroup.getLeafType: the pseudo
// state named by the first stereotype of the group, LeafTypeNull for the
// others.
func stereogroupLeafType(g *klimt.Stereogroup) abel.LeafType {
	labels := g.Labels()
	if len(labels) == 0 {
		return abel.LeafTypeNull
	}
	switch strings.ToLower(labels[0]) {
	case "choice":
		return abel.LeafTypeStateChoice
	case "fork", "join":
		return abel.LeafTypeStateForkJoin
	case "start":
		return abel.LeafTypeCircleStart
	case "end":
		return abel.LeafTypeCircleEnd
	case "history":
		return abel.LeafTypePseudoState
	case "history*":
		return abel.LeafTypeDeepHistory
	}
	return abel.LeafTypeNull
}
