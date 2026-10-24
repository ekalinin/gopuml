// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/objectdiagram/command/CommandCreateMap.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/objectdiagram/command/CommandCreateJson.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/objectdiagram/command/CommandCreateJsonSingleLine.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// The map and JSON commands are shared by class and description diagrams,
// as upstream types them on AbstractEntityDiagram.

// objectDisplay is the display of a map or JSON element: the display
// given, or the name in simple-line creole.
func objectDisplay(c *CucaDiagram, displayString string, hasDisplay bool, name string) *klimt.Display {
	display := klimt.DisplayWithNewlinesString(c.Pragma(), displayString, hasDisplay)
	if klimt.DisplayIsNull(display) {
		display = klimt.DisplayWithNewlines(c.Pragma(), name).WithCreoleMode(klimt.CreoleModeSimpleLine)
	}
	return display
}

// setStereotypeSpotted sets the spotted stereotype STEREO of the match, if
// any.
func setStereotypeSpotted(c *CucaDiagram, entity *Entity, arg *regex.Result) command.Result {
	if stereotype, ok := arg.Get("STEREO", 0); ok {
		st, err := SpottedStereotype(c, stereotype)
		if err != nil {
			return ColorResult(err)
		}
		entity.SetStereotype(st)
	}
	return command.Ok()
}

// setBackColor sets the COLOR of the match, if any, as the background
// color.
func setBackColor(entity *Entity, arg *regex.Result) command.Result {
	if s, ok := arg.Get("COLOR", 0); ok {
		c, err := GetColor(s)
		if err != nil {
			return ColorResult(err)
		}
		entity.SetSpecificColorTOBEREMOVED(color.TypeBack, c)
	}
	return command.Ok()
}

// mapColors are the back color and the "##[style]color" line color of a
// map declaration.
func mapColors(arg *regex.Result) (*color.Colors, error) {
	colors, err := color.SimpleColor(color.TypeBack).Color(arg, color.Instance())
	if err != nil {
		return nil, err
	}
	if s, ok := arg.Get("LINECOLOR", 1); ok {
		lineColor, err := GetColor(s)
		if err != nil {
			return nil, err
		}
		if lineColor != nil {
			colors = colors.Add(color.TypeLine, lineColor)
		}
	}
	if s, ok := arg.Get("LINECOLOR", 0); ok {
		colors = colors.AddLegacyStroke(s)
	}
	return colors, nil
}

// CommandCreateMap is `map code {` ... `}`, the body lines being
// "key => value" entries or "key *--> entity" links.
func CommandCreateMap[D Diagram]() command.Command[D] {
	return Multilines(&uml.Multilines2[D]{
		Start: regex.NewConcat(regex.Start(),
			regex.NewNamedLeaf(0, "TYPE", "map"),
			regex.SpaceOneOrMore(),
			uml.Leaf(2, "NAME", "(?:[%g]([^%g]+)[%g][%s]+as[%s]+)?([%pLN_.]+)"),
			klimt.StereotypePatternOptional("STEREO"),
			klimt.URLBuilderOptional(),
			regex.SpaceZeroOrMore(),
			color.SimpleColor(color.TypeBack).Regex(),
			regex.SpaceZeroOrMore(),
			regex.Optional(regex.NewConcat(uml.Lit("##"),
				uml.Leaf(2, "LINECOLOR", `(?:\[(dotted|dashed|bold)\])?(\w+)?`))),
			regex.SpaceZeroOrMore(),
			uml.Lit(`\{`),
			regex.End()),
		Strategy: uml.RemoveStartingQuote,
		TrimEnd:  uml.TrimBoth,
		End:      uml.Compile(`^[%s]*\}[%s]*$`),
	}, func(m *uml.Multilines2[D], d D, lines command.BlocLines) command.Result {
		c := d.Cuca()
		lines = lines.Trim().RemoveEmptyLines()
		line0 := m.Start.Matcher(lines.At(0).Trimmed().S)
		name := line0.Str("NAME", 1)
		quark := c.QuarkInContext(true, c.CleanID(name))
		if quark.Data() != nil {
			return command.Error("Map already exists: " + name)
		}
		displayString, hasDisplay := line0.Get("NAME", 0)
		display := objectDisplay(c, displayString, hasDisplay, name)
		entity1 := c.ReallyCreateLeaf(uml.Location(lines), quark, display, LeafTypeMap, USymbolNull)
		if r := setStereotypeSpotted(c, entity1, line0); !r.IsOK() {
			return r
		}
		colors, err := mapColors(line0)
		if err != nil {
			return ColorResult(err)
		}
		entity1.SetColors(colors)
		lines = uml.SubExtract(lines, 1, 1)
		for _, sl := range lines.Lines() {
			line := sl.S
			if !entity1.Bodier().AddFieldOrMethod(line) {
				return command.Error("Map definition should contains key => value")
			}
			linkStr, ok := BodierMapLinkedEntry(line)
			if !ok {
				continue
			}
			x := strings.Index(line, linkStr)
			key := text.JavaTrim(line[:x])
			dest := text.JavaTrim(line[x+len(linkStr):])
			ident2 := c.QuarkInContext(true, dest)
			entity2 := ident2.Data()
			if entity2 == nil {
				return command.Error("No such entity " + ident2.Name())
			}
			linkType := decoration.NewLinkType(decoration.LinkDecorArrow, decoration.LinkDecorNone)
			length := uml.JavaLength(linkStr) - 2
			link := NewLink(uml.Location(lines), c, c.SkinParam().CurrentStyleBuilder(), entity1,
				entity2, linkType, LinkArgNoDisplay(length))
			link.SetPortMembers(Some(key), OptString{})
			c.AddLink(link)
		}
		return command.Ok()
	})
}

// jsonValue is upstream's CommandCreateJson.getJsonValue: the body lines
// joined in braces, or joined alone; nil when neither is JSON.
func jsonValue(lines command.BlocLines) json.Value {
	body := uml.SubExtract(lines, 1, 1)
	var sb strings.Builder
	for _, sl := range body.Lines() {
		sb.WriteString(sl.S)
	}
	if v, err := json.Parse("{" + sb.String() + "}"); err == nil {
		return v
	}
	if v, err := json.Parse(sb.String()); err == nil {
		return v
	}
	return nil
}

// jsonMultilines is CommandCreateJson: a multi-line command whose last
// line must also close valid JSON data (upstream's finalVerification).
type jsonMultilines[D Diagram] struct {
	*uml.Multilines2[D]
}

// IsValid implements command.Command.
func (c jsonMultilines[D]) IsValid(lines command.BlocLines) command.Control {
	result := c.Multilines2.IsValid(lines)
	if result != command.OK {
		return result
	}
	if jsonValue(uml.RemoveStartingQuote.CleanList(lines)) == nil {
		return command.OKPartial
	}
	return command.OK
}

// CommandCreateJSON is `json code {` ... `}`, the body being JSON data.
func CommandCreateJSON[D Diagram]() command.Command[D] {
	return jsonMultilines[D]{Multilines(&uml.Multilines2[D]{
		Start: regex.NewConcat(regex.Start(),
			regex.NewNamedLeaf(0, "TYPE", "json"),
			regex.SpaceOneOrMore(),
			NameAndCode(),
			klimt.StereotypePatternOptional("STEREO"),
			klimt.URLBuilderOptional(),
			regex.SpaceZeroOrMore(),
			color.Exp1(),
			regex.SpaceZeroOrMore(),
			uml.Lit(`\{`),
			regex.End()),
		Strategy: uml.RemoveStartingQuote,
		TrimEnd:  uml.TrimBoth,
		End:      uml.Compile(`^[%s]*\}[%s]*$`),
	}, func(m *uml.Multilines2[D], d D, lines command.BlocLines) command.Result {
		c := d.Cuca()
		lines = lines.Trim().RemoveEmptyLines()
		line0 := m.Start.Matcher(lines.At(0).Trimmed().S)
		code, _ := line0.Lazy("CODE", 0)
		quark := c.QuarkInContext(true, c.CleanID(code))
		if quark.Data() != nil {
			return command.Error("JSON already exists: " + code)
		}
		displayString, hasDisplay := line0.Lazy("DISPLAY", 0)
		display := objectDisplay(c, displayString, hasDisplay, quark.Name())
		entity1 := c.ReallyCreateLeaf(uml.Location(lines), quark, display, LeafTypeJSON, USymbolNull)
		if r := setStereotypeSpotted(c, entity1, line0); !r.IsOK() {
			return r
		}
		if r := setBackColor(entity1, line0); !r.IsOK() {
			return r
		}
		value := jsonValue(lines)
		if value == nil {
			return command.Error("Bad data")
		}
		entity1.Bodier().(*BodierJSon).SetJSON(value)
		return command.Ok()
	})}
}

// CommandCreateJSONSingleLine is `json code <value>`, the value being a
// JSON literal, number, string, array or object.
func CommandCreateJSONSingleLine[D Diagram]() command.Command[D] {
	return uml.Single(regex.NewConcat(regex.Start(),
		regex.NewNamedLeaf(0, "TYPE", "json"),
		regex.SpaceOneOrMore(),
		uml.Leaf(2, "NAME", "(?:[%g]([^%g]+)[%g][%s]+as[%s]+)?([%pLN_.]+)"),
		klimt.StereotypePatternOptional("STEREO"),
		klimt.URLBuilderOptional(),
		regex.SpaceZeroOrMore(),
		color.Exp1(),
		regex.SpaceZeroOrMore(),
		regex.Or("",
			uml.Leaf(1, "DATA_BOOLEAN", "(true|false)"),
			uml.Leaf(1, "DATA_NUMBER", `(-?\d+)`),
			uml.Leaf(1, "DATA_NULL", "(null)"),
			uml.Leaf(1, "DATA_STRING", `(".*")`),
			uml.Leaf(1, "DATA_ARRAY", `(\[.*\])`),
			uml.Leaf(2, "DATA_OBJECT", `(\{[%s]*[%g](\\[%g]|[^%g])+[%g][%s]*:.*\})`)),
		regex.End()),
		func(d D, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			c := d.Cuca()
			name := arg.Str("NAME", 1)
			quark := c.QuarkInContext(false, c.CleanID(name))
			if quark.Data() != nil {
				return command.Error("JSON already exists: " + name)
			}
			displayString, hasDisplay := arg.Get("NAME", 0)
			display := objectDisplay(c, displayString, hasDisplay, name)
			entity1 := c.ReallyCreateLeaf(location, quark, display, LeafTypeJSON, USymbolNull)
			if r := setStereotypeSpotted(c, entity1, arg); !r.IsOK() {
				return r
			}
			if r := setBackColor(entity1, arg); !r.IsOK() {
				return r
			}
			data, _ := arg.Lazy("DATA_", 0)
			value, err := json.Parse(data)
			if err != nil {
				return command.Error("Bad data")
			}
			entity1.Bodier().(*BodierJSon).SetJSON(value)
			return command.Ok()
		})
}
