// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/mindmap/CommandMindMapDirection.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/mindmap/CommandMindMapOrgmode.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/mindmap/CommandMindMapOrgmodeMultiline.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/mindmap/CommandMindMapPlus.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/mindmap/CommandMindMapRoot.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package mindmap

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// cmd is a command of mindmap diagrams.
type cmd = command.Command[*MindMapDiagram]

// The explain methods of upstream's commands, which only its @Explain
// tooling calls, are not ported.

// backColorOf reads the optional BACKCOLOR group with the color set of the
// skin parameters; ok is false for an unknown color.
func backColorOf(d *MindMapDiagram, arg *regex.Result) (c color.HColor, ok bool) {
	s, found := arg.Get("BACKCOLOR", 0)
	if !found {
		return nil, true
	}
	c, err := d.SkinParam().IHtmlColorSet().GetColor(s)
	if err != nil {
		return nil, false
	}
	return c, true
}

// commandMindMapDirection is "left side", "right side" and the other lines
// that name a side: the side of the next "*" and "#" nodes.
func commandMindMapDirection() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit("[^*#]*"),
		uml.Lit(`\b`),
		uml.Leaf(1, "DIRECTION", "(left|right|top|bottom)"),
		uml.Lit(`\b`),
		uml.Lit("[^*#]*"),
		uml.Leaf(1, "", "(side|direction)"),
		uml.Lit("[^*#]*"),
		regex.End()),
		func(d *MindMapDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			d.setDefaultDirection(klimt.DirectionLazzyValueOf(arg.Str("DIRECTION", 0)))
			return command.Ok()
		})
}

// commandMindMapOrgmode is a node written with "*" or "#" markers, which
// may be indented; the line is not trimmed, so that getSmartLevel sees the
// indentation.
func commandMindMapOrgmode() cmd {
	return &command.SingleLine[*MindMapDiagram]{
		Pattern: regex.NewConcat(regex.Start(),
			uml.Leaf(1, "TYPE", "([ \t]*[*#]+)"),
			regex.Optional(uml.Leaf(1, "BACKCOLOR", `\[(#\w+)\]`)),
			uml.Leaf(1, "SHAPE", "(_)?"),
			regex.SpaceZeroOrMore(),
			uml.Leaf(1, "LABEL", "(.*)"),
			regex.End()),
		NoTrim:      true,
		FirstTokens: []string{"#", "*"},
		Exec: func(d *MindMapDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			backColor, ok := backColorOf(d, arg)
			if !ok {
				return command.BadColor()
			}
			return d.addIdea(backColor, d.smartLevel(arg.Str("TYPE", 0)),
				klimt.DisplayWithNewlines(d.Pragma(), arg.Str("LABEL", 0)), IdeaShapeFromDesc(arg.Str("SHAPE", 0)))
		},
	}
}

// multilineEnd is the last line of a multiline node: the text before ";",
// then an optional stereotype.
var multilineEnd = uml.Compile(`^(.*);\s*(\<\<(.+)\>\>)?$`)

// commandMindMapOrgmodeMultiline is a node from "**:first line" to a line
// that ends with ";", optionally followed by a stereotype.
func commandMindMapOrgmodeMultiline() cmd {
	c := &uml.Multilines2[*MindMapDiagram]{
		Start: regex.NewConcat(regex.Start(),
			uml.Leaf(1, "TYPE", "([*#]+)"),
			regex.Optional(uml.Leaf(1, "BACKCOLOR", `\[(#\w+)\]`)),
			uml.Leaf(1, "SHAPE", "(_)?"),
			uml.Lit(":"),
			uml.Leaf(1, "DATA", "(.*)"),
			regex.End()),
		Strategy: uml.RemoveStartingQuote,
		TrimEnd:  uml.TrimBoth,
		End:      multilineEnd,
	}
	c.Exec = func(d *MindMapDiagram, lines command.BlocLines, _ command.ParserPass) command.Result {
		line0 := c.StartArg(lines)
		// The last line is read untrimmed: a space after the stereotype
		// leaves it unmatched, as upstream's getSplit returns null.
		lineLast := multilineEnd.FindStringSubmatch(lines.Last().S)
		if lineLast == nil {
			uml.ThrowNPE(`Cannot invoke "java.util.List.get(int)" because "lineLast" is null`)
		}
		// Upstream removes the last character, whatever it is.
		lines = lines.RemoveStartingAndEnding(line0.Str("DATA", 0), 1)
		stereotype := lineLast[2]
		if stereotype != "" {
			lines = lines.OverrideLastLine(lineLast[1])
		}
		typ := line0.Str("TYPE", 0)
		backColor, ok := backColorOf(d, line0)
		if !ok {
			return command.BadColor()
		}
		label := klimt.DisplayCreateFoo(lines.Lines())
		shape := IdeaShapeFromDesc(line0.Str("SHAPE", 0))
		if stereotype == "" {
			return d.addIdea(backColor, len(typ)-1, label, shape)
		}
		return d.addIdeaStereotype(klimt.StereotypeBuild(stereotype), backColor, len(typ)-1, label, shape)
	}
	return c
}

// commandMindMapPlus is a node written with "+" markers (right side) or
// with markers that contain "-" (left side).
func commandMindMapPlus() cmd {
	return &command.SingleLine[*MindMapDiagram]{
		Pattern: regex.NewConcat(regex.Start(),
			uml.Leaf(1, "TYPE", "([+-]+)"),
			regex.Optional(uml.Leaf(1, "BACKCOLOR", `\[(#\w+)\]`)),
			uml.Leaf(1, "SHAPE", "(_)?"),
			regex.SpaceZeroOrMore(),
			uml.Leaf(1, "LABEL", "(.*)"),
			regex.End()),
		NoTrim:      true,
		FirstTokens: []string{"+", "-"},
		Exec: func(d *MindMapDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			typ := arg.Str("TYPE", 0)
			backColor, ok := backColorOf(d, arg)
			if !ok {
				return command.BadColor()
			}
			direction := !strings.Contains(typ, "-")
			return d.addIdeaDirection(backColor, len(typ)-1,
				klimt.DisplayWithNewlines(d.Pragma(), arg.Str("LABEL", 0)), IdeaShapeFromDesc(arg.Str("SHAPE", 0)),
				direction)
		},
	}
}

// commandMindMapRoot is "0 label": the root.
func commandMindMapRoot() cmd {
	c := uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "TYPE", "(0)"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "LABEL", "(.*)"),
		regex.End()),
		func(d *MindMapDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			return d.addIdeaDirection(nil, 0, klimt.DisplayWithNewlines(d.Pragma(), arg.Str("LABEL", 0)),
				IdeaShapeBox, true)
		})
	c.FirstTokens = []string{"0"}
	return c
}
