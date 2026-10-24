// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/command/CommandSwimlane.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/command/CommandSwimlane2.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/command/CommandPartition3.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/command/CommandCloseGroup3.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/command/CommandCloseGroupLegacy3.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// executeSwimlane is the executeArg of CommandSwimlane and CommandSwimlane2.
func executeSwimlane(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
	var c color.HColor
	if s, ok := arg.Get("COLOR", 0); ok {
		var err error
		if c, err = getColor(d, s); err != nil {
			return colorResult(err)
		}
	}
	name := arg.Str("SWIMLANE", 0)
	label := withNewlines(d, arg, "LABEL")
	return d.Swimlane(name, c, label)
}

// commandSwimlane is "|#color|name|label".
func commandSwimlane() cmd {
	return single(executeSwimlane,
		regex.Start(),
		uml.Lit(`\|`),
		color.Exp6(),
		uml.Leaf(1, "SWIMLANE", "([^|]+)"),
		uml.Lit(`\|`),
		uml.Leaf(1, "LABEL", "([^|]+)?"),
		regex.End())
}

// commandSwimlane2 is "swimlane #color name as label".
func commandSwimlane2() cmd {
	return single(executeSwimlane,
		regex.Start(),
		uml.Lit("swimlane"),
		regex.SpaceOneOrMore(),
		color.Exp7(),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "SWIMLANE", "([^|]+?)"),
		regex.Optional(regex.NewConcat(
			regex.SpaceOneOrMore(),
			uml.Lit("as"),
			regex.SpaceOneOrMore(),
			uml.Leaf(1, "LABEL", "([^|]+)"))),
		regex.End())
}

// partitionColor is CommandPartition3.color.
func partitionColor(id string) *color.ColorParser { return color.SimpleColorID(color.TypeBack, id) }

// partitionSymbol is CommandPartition3.getUSymbol.
func partitionSymbol(typ string) GroupSymbol {
	switch {
	case uml.EqualsIgnoreCase("partition", typ):
		return SymbolPartition
	case uml.EqualsIgnoreCase("package", typ):
		return SymbolPackage
	case uml.EqualsIgnoreCase("rectangle", typ):
		return SymbolRectangle
	case uml.EqualsIgnoreCase("card", typ):
		return SymbolCard
	case uml.EqualsIgnoreCase("group", typ):
		return SymbolGroup
	}
	uml.Throw(uml.IllegalArgumentException, "")
	return SymbolPartition
}

// commandPartition3 is "partition #color name <<stereo>> {" (also package,
// rectangle, card and group).
func commandPartition3() cmd {
	return single(executePartition3,
		regex.Start(),
		uml.Leaf(1, "TYPE", "(partition|package|rectangle|card|group)"),
		regex.SpaceOneOrMore(),
		regex.Optional(regex.NewConcat(
			partitionColor("BACK1").Regex(),
			regex.SpaceOneOrMore())),
		uml.Leaf(1, "NAME", "([%g][^%g]+[%g]|.*?)"),
		regex.Optional(regex.NewConcat(
			regex.SpaceOneOrMore(),
			partitionColor("BACK2").Regex())),
		klimt.StereotypePatternOptional("STEREO"),
		uml.Leaf(1, "BRACKET", `(\{?)`),
		regex.End())
}

func executePartition3(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
	name := uml.RemoveQuotes(arg.Str("NAME", 0))
	id := "BACK1"
	if _, ok := arg.Get("BACK1", 0); !ok {
		id = "BACK2"
	}
	colors, err := partitionColor(id).Color(arg, d.Skin().IHtmlColorSet())
	if err != nil {
		return colorResult(err)
	}
	typ := arg.Str("TYPE", 0)
	symbol := partitionSymbol(typ)
	var stereotype *klimt.Stereotype
	if stereo, ok := arg.Get("STEREO", 0); ok {
		stereotype = klimt.StereotypeBuild(stereo)
	}
	if arg.Str("BRACKET", 0) == "" {
		d.AddWarning(klimt.NewWarning("You should use a bracket ({) when defining your container '" + typ + "' " + name))
	}
	stylePartition := groupStyleSignature(symbol).WithTOBECHANGED(stereotype).MergedStyle(d.CurrentStyleBuilder())
	backColor := colors.Color(color.TypeBack)
	if backColor == nil {
		backColor = stylePartition.Value(klimt.PNameBackGroundColor).AsColor(d.Skin().IHtmlColorSet())
	}
	d.StartGroup(klimt.DisplayWithNewlines(d.Pragma(), name), backColor, symbol, stylePartition)
	return command.Ok()
}

// commandCloseGroup3 is "}".
func commandCloseGroup3() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, _ *regex.Result) command.Result {
		return d.CloseGroup()
	}, regex.Start(), uml.Lit(`\}`), regex.End())
}

// commandCloseGroupLegacy3 is "end group".
func commandCloseGroupLegacy3() cmd {
	return single(func(d *ActivityDiagram3, _ *text.LineLocation, arg *regex.Result) command.Result {
		d.AddWarning(klimt.NewWarning("You should use a bracket (}) instead of '" + arg.Str("CMD", 0) + "'"))
		return d.CloseGroup()
	},
		regex.Start(),
		uml.Leaf(1, "CMD", "(end ?group|group ?end)"),
		uml.Lit(";?"),
		regex.End())
}
