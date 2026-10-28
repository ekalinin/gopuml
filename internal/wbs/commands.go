// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/wbs/CommandWBSItemMultilineNew.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/wbs/CommandWBSItemMultilineOld.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/wbs/CommandWBSItemNew.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/wbs/CommandWBSItemOld.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/wbs/CommandWBSLink.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/utils/Direction.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package wbs

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/mindmap"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// cmd is a command of WBS diagrams.
type cmd = command.Command[*WBSDiagram]

// The explain methods of upstream's commands, which only its @Explain
// tooling calls, are not ported.

// The group names of upstream's Constant.WBS_TYPE and WBS_DIRECTION.
const (
	wbsType      = "TYPE"
	wbsDirection = "DIRECTION"
)

// itemFirstTokens are the first tokens of the node commands.
var itemFirstTokens = []string{"*", "+", "-"}

// wbsTypePart is the markers of a node, which may be indented.
func wbsTypePart() *regex.Leaf { return uml.Leaf(1, wbsType, "([ \t]*[*+-]+)") }

// shapeAndDirection is "_" (no box) then "<" or ">" (the side). The
// parentheses of upstream's source close its second alternative after the
// side, so the alternatives are "_" then the side, the side alone and "_"
// alone, and ">_" is not read.
func shapeAndDirection() regex.Part {
	return regex.Or("",
		regex.NewConcat(
			uml.Leaf(1, "SHAPE_1", "(_)?"),
			uml.Leaf(1, wbsDirection+"_1", "([<>])?")),
		regex.NewConcat(
			uml.Leaf(1, wbsDirection+"_2", "([<>])?")),
		uml.Leaf(1, "SHAPE_2", "(_)?"))
}

// colorAndCode is the background color and the alias, in either order.
func colorAndCode() regex.Part {
	return regex.Or("",
		regex.NewConcat(
			regex.Optional(uml.Leaf(1, "BACKCOLOR_1", `\[(#\w+)\]`)),
			regex.Optional(uml.Leaf(1, "CODE_1", `\(([%pLN_]+)\)`))),
		regex.NewConcat(
			regex.Optional(uml.Leaf(1, "CODE_2", `\(([%pLN_]+)\)`)),
			regex.Optional(uml.Leaf(1, "BACKCOLOR_2", `\[(#\w+)\]`))))
}

// quotedLabelAndCode is a quoted label followed by "as" and the alias.
func quotedLabelAndCode() []regex.Part {
	return []regex.Part{
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "LABEL", "[%g](.*)[%g]"),
		regex.SpaceOneOrMore(),
		uml.Lit("as"),
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "CODE", "([%pLN_]+)"),
		regex.End(),
	}
}

// optionalLabel is a space and the label, which a pseudo node does not
// have.
func optionalLabel() regex.Part {
	return regex.Optional(regex.NewConcat(
		regex.SpaceOneOrMore(),
		uml.Leaf(1, "LABEL", "([^%s].*)")))
}

// lazy is upstream's getLazzy: the first matched group whose name starts
// with prefix, or "".
func lazy(arg *regex.Result, prefix string) string {
	s, _ := arg.Lazy(prefix, 0)
	return s
}

// backColorOf reads the background color with the color set of the skin
// parameters; ok is false for an unknown color.
func backColorOf(d *WBSDiagram, arg *regex.Result) (c color.HColor, ok bool) {
	s, found := arg.Lazy("BACKCOLOR", 0)
	if !found {
		return nil, true
	}
	c, err := d.SkinParam().IHtmlColorSet().GetColor(s)
	if err != nil {
		return nil, false
	}
	return c, true
}

// getWBSDirection is upstream's Direction.getWBSDirection: the left side for
// markers that contain "-", the right side otherwise, unless "<" (left) or
// ">" (right) follows the markers.
func getWBSDirection(arg *regex.Result) klimt.Direction {
	direction := klimt.DirectionRight
	if strings.Contains(arg.Str(wbsType, 0), "-") {
		direction = klimt.DirectionLeft
	}
	switch lazy(arg, wbsDirection) {
	case "<":
		direction = klimt.DirectionLeft
	case ">":
		direction = klimt.DirectionRight
	}
	return direction
}

// executeItem adds the node of a single line command; old adds upstream's
// warning on the order of the line, withCode with the alias.
func executeItem(d *WBSDiagram, arg *regex.Result, old, withCode bool) command.Result {
	typ := arg.Str(wbsType, 0)
	label, hasLabel := arg.Get("LABEL", 0)
	code := lazy(arg, "CODE")
	backColor, ok := backColorOf(d, arg)
	if !ok {
		return command.BadColor()
	}
	dir := getWBSDirection(arg)
	shape := mindmap.IdeaShapeFromDesc(lazy(arg, "SHAPE"))
	if !hasLabel && shape == mindmap.IdeaShapeNone {
		shape = mindmap.IdeaShapePseudo
	} else if !hasLabel {
		return command.Error("Missing label for WBS node.")
	}
	if old {
		d.AddWarning(orderWarning(arg, withCode))
	}
	return d.addIdeaLabel(code, backColor, d.smartLevel(typ), label, dir, shape)
}

// commandWBSItemNew is a node with "_" and the side before the color and
// the alias: mode 0 with an optional label, mode 1 with a quoted label and
// "as" alias. The line is not trimmed, so that getSmartLevel sees the
// indentation.
func commandWBSItemNew(mode int) cmd {
	var pattern *regex.Concat
	if mode == 0 {
		pattern = regex.NewConcat(regex.Start(),
			wbsTypePart(),
			shapeAndDirection(),
			colorAndCode(),
			optionalLabel(),
			regex.End())
	} else {
		pattern = regex.NewConcat(append([]regex.Part{regex.Start(),
			wbsTypePart(),
			shapeAndDirection(),
			regex.Optional(uml.Leaf(1, "BACKCOLOR", `\[(#\w+)\]`))},
			quotedLabelAndCode()...)...)
	}
	return &command.SingleLine[*WBSDiagram]{
		Pattern:     pattern,
		NoTrim:      true,
		FirstTokens: itemFirstTokens,
		Exec: func(d *WBSDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			return executeItem(d, arg, false, false)
		},
	}
}

// commandWBSItemOld is a node with the color and the alias before "_" and
// the side, the legacy order, which adds a warning.
func commandWBSItemOld(mode int) cmd {
	var pattern *regex.Concat
	if mode == 0 {
		pattern = regex.NewConcat(regex.Start(),
			wbsTypePart(),
			colorAndCode(),
			shapeAndDirection(),
			optionalLabel(),
			regex.End())
	} else {
		pattern = regex.NewConcat(append([]regex.Part{regex.Start(),
			wbsTypePart(),
			regex.Optional(uml.Leaf(1, "BACKCOLOR", `\[(#\w+)\]`)),
			shapeAndDirection()},
			quotedLabelAndCode()...)...)
	}
	return &command.SingleLine[*WBSDiagram]{
		Pattern:     pattern,
		NoTrim:      true,
		FirstTokens: itemFirstTokens,
		Exec: func(d *WBSDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			return executeItem(d, arg, true, mode == 0)
		},
	}
}

// orderWarning is the warning on the legacy order, where the color and the
// alias come before the shape and the side. It shows the start of the line
// rewritten in the expected order, e.g. '**_>[#red](id)' for
// '**[#red](id)_>'.
func orderWarning(arg *regex.Result, withCode bool) *klimt.Warning {
	var sb strings.Builder
	sb.WriteString(text.JavaTrim(arg.Str(wbsType, 0)))
	appendIfPresent(&sb, "", lazy(arg, "SHAPE"), "")
	appendIfPresent(&sb, "", lazy(arg, wbsDirection), "")
	appendIfPresent(&sb, "[", lazy(arg, "BACKCOLOR"), "]")
	if withCode {
		appendIfPresent(&sb, "(", lazy(arg, "CODE"), ")")
	}
	return klimt.NewWarning("Shape (_) and direction (< >) must come before color and id: start the line with '" +
		sb.String() + "'")
}

func appendIfPresent(sb *strings.Builder, before, value, after string) {
	if value != "" {
		sb.WriteString(before + value + after)
	}
}

// multilineEnd is the last line of a multiline node: the text before ";",
// then an optional stereotype.
var multilineEnd = uml.Compile(`^(.*);\s*(\<\<(.+)\>\>)?$`)

// commandWBSItemMultiline is a node from "**:first line" to a line that
// ends with ";", optionally followed by a stereotype: with "_" and the side
// before the color and the alias, or, for old, after them, which adds a
// warning.
func commandWBSItemMultiline(old bool) cmd {
	start := []regex.Part{regex.Start(), wbsTypePart(), shapeAndDirection(), colorAndCode()}
	if old {
		start = []regex.Part{regex.Start(), wbsTypePart(), colorAndCode(), shapeAndDirection()}
	}
	c := &uml.Multilines2[*WBSDiagram]{
		Start: regex.NewConcat(append(start,
			uml.Lit(":"),
			uml.Leaf(1, "DATA", "(.*)"),
			regex.End())...),
		Strategy: uml.RemoveStartingQuote,
		TrimEnd:  uml.TrimBoth,
		End:      multilineEnd,
	}
	c.Exec = func(d *WBSDiagram, lines command.BlocLines, _ command.ParserPass) command.Result {
		line0 := c.StartArg(lines)
		lineLast := multilineEnd.FindStringSubmatch(lines.Last().S)
		if lineLast == nil {
			uml.ThrowNPE(`Cannot invoke "java.util.List.get(int)" because "lineLast" is null`)
		}
		// Upstream removes the last character, whatever it is.
		lines = lines.RemoveStartingAndEnding(line0.Str("DATA", 0), 1)
		var stereotype *klimt.Stereotype
		if lineLast[2] != "" {
			lines = lines.OverrideLastLine(lineLast[1])
			stereotype = klimt.StereotypeBuild(lineLast[2])
		}
		typ := line0.Str(wbsType, 0)
		backColor, ok := backColorOf(d, line0)
		if !ok {
			return command.BadColor()
		}
		code := lazy(line0, "CODE")
		dir := getWBSDirection(line0)
		if old {
			d.AddWarning(orderWarning(line0, true))
		}
		return d.addIdea(code, backColor, d.smartLevel(typ), klimt.DisplayCreateFoo(lines.Lines()), stereotype, dir,
			mindmap.IdeaShapeFromDesc(lazy(line0, "SHAPE")))
	}
	return c
}

// linkColor is the color of a link.
func linkColor() *color.ColorParser { return color.SimpleColor(color.TypeLine) }

// commandWBSLink is "a -> b": a link from the node with the alias a to the
// node with the alias b. Its label is read and not used, as upstream does.
func commandWBSLink() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "CODE1", "([%pLN_]+)"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(0, "LINK", `[.-]+\>`),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "CODE2", "([%pLN_]+)"),
		regex.SpaceZeroOrMore(),
		linkColor().Regex(),
		klimt.StereotypePatternOptional("STEREOTYPE"),
		uml.Leaf(1, "LABEL_LINK", "(?::[%s]*(.+))?"),
		regex.End()),
		func(d *WBSDiagram, _ *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			colors, err := linkColor().Color(arg, d.SkinParam().IHtmlColorSet())
			if err != nil {
				return command.BadColor()
			}
			var stereotype *klimt.Stereotype
			if s, ok := arg.Get("STEREOTYPE", 0); ok {
				stereotype = klimt.StereotypeBuild(s)
			}
			return d.link(arg.Str("CODE1", 0), arg.Str("CODE2", 0), colors, stereotype)
		})
}
