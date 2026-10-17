// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/objectdiagram/command/CommandCreateEntityObject.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/objectdiagram/command/CommandCreateEntityObjectMultilines.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/objectdiagram/command/CommandCreateMap.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/objectdiagram/command/CommandCreateJson.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/objectdiagram/command/CommandCreateJsonSingleLine.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package classdiagram

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// objectDisplay is the display of an object, map or JSON element: the
// display given, or the name in simple-line creole.
func objectDisplay(d *ClassDiagram, displayString string, hasDisplay bool, name string) *klimt.Display {
	display := klimt.DisplayWithNewlinesString(d.Pragma(), displayString, hasDisplay)
	if klimt.DisplayIsNull(display) {
		display = klimt.DisplayWithNewlines(d.Pragma(), name).WithCreoleMode(klimt.CreoleModeSimpleLine)
	}
	return display
}

// setStereotypeSpotted sets the spotted stereotype STEREO of the match, if
// any.
func setStereotypeSpotted(d *ClassDiagram, entity *abel.Entity, arg *regex.Result) command.Result {
	if stereotype, ok := arg.Get("STEREO", 0); ok {
		st, err := abel.SpottedStereotype(&d.CucaDiagram, stereotype)
		if err != nil {
			return abel.ColorResult(err)
		}
		entity.SetStereotype(st)
	}
	return command.Ok()
}

// setBackColor sets the COLOR of the match, if any, as the background
// color.
func setBackColor(entity *abel.Entity, arg *regex.Result) command.Result {
	if s, ok := arg.Get("COLOR", 0); ok {
		c, err := abel.GetColor(s)
		if err != nil {
			return abel.ColorResult(err)
		}
		entity.SetSpecificColorTOBEREMOVED(color.TypeBack, c)
	}
	return command.Ok()
}

// commandCreateEntityObject is `object "display" as code <<stereotype>>
// #color`.
func commandCreateEntityObject() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		regex.NewNamedLeaf(0, "TYPE", "object"),
		regex.SpaceOneOrMore(),
		abel.NameAndCode(),
		klimt.StereotypePatternOptional("STEREO"),
		klimt.URLBuilderOptional(),
		regex.SpaceZeroOrMore(),
		color.Exp1(), regex.End()),
		func(d *ClassDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			code, _ := arg.Lazy("CODE", 0)
			idShort := d.CleanID(code)
			quark := d.QuarkInContext(true, d.CleanID(idShort))
			displayString, hasDisplay := arg.Lazy("DISPLAY", 0)
			if quark.Data() != nil {
				return command.Error("Object already exists: " + quark.Name())
			}
			display := objectDisplay(d, displayString, hasDisplay, quark.Name())
			entity := d.ReallyCreateLeaf(location, quark, display, abel.LeafTypeObject, abel.USymbolNull)
			if r := setStereotypeSpotted(d, entity, arg); !r.IsOK() {
				return r
			}
			if urlString, ok := arg.Get(klimt.URLBuilderKey, 0); ok {
				entity.AddURL(abel.URLBuilderOf(&d.CucaDiagram).URL(urlString))
			}
			return setBackColor(entity, arg)
		})
}

// commandCreateEntityObjectMultilines is `object code {` ... `}`, the body
// lines being fields.
func commandCreateEntityObjectMultilines() cmd {
	return abel.Multilines(&uml.Multilines2[*ClassDiagram]{
		Start: regex.NewConcat(regex.Start(),
			regex.NewNamedLeaf(0, "TYPE", "object"),
			regex.SpaceOneOrMore(),
			abel.NameAndCode(),
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
	}, func(c *uml.Multilines2[*ClassDiagram], d *ClassDiagram, lines command.BlocLines) command.Result {
		lines = lines.Trim().RemoveEmptyLines()
		line0 := c.Start.Matcher(lines.At(0).Trimmed().S)
		code, _ := line0.Lazy("CODE", 0)
		quark := d.QuarkInContext(true, d.CleanID(code))
		displayString, hasDisplay := line0.Lazy("DISPLAY", 0)
		display := objectDisplay(d, displayString, hasDisplay, quark.Name())
		entity := quark.Data()
		if entity == nil {
			entity = d.ReallyCreateLeaf(uml.Location(lines), quark, display, abel.LeafTypeObject, abel.USymbolNull)
		}
		if r := setStereotypeSpotted(d, entity, line0); !r.IsOK() {
			return r
		}
		if r := setBackColor(entity, line0); !r.IsOK() {
			return r
		}
		lines = uml.SubExtract(lines, 1, 1)
		for _, s := range lines.Lines() {
			if klimt.VisibilityIsVisibilityCharacter(s.S) {
				d.SetVisibilityModifierPresent(true)
			}
			entity.Bodier().AddFieldOrMethod(s.S)
		}
		return command.Ok()
	})
}

// commandCreateMap is `map code {` ... `}`, the body lines being
// "key => value" entries or "key *--> entity" links.
func commandCreateMap() cmd {
	return abel.Multilines(&uml.Multilines2[*ClassDiagram]{
		Start: regex.NewConcat(regex.Start(),
			regex.NewNamedLeaf(0, "TYPE", "map"),
			regex.SpaceOneOrMore(),
			uml.Leaf(2, "NAME", "(?:[%g]([^%g]+)[%g][%s]+as[%s]+)?([%pLN_.]+)"),
			klimt.StereotypePatternOptional("STEREO"),
			klimt.URLBuilderOptional(),
			regex.SpaceZeroOrMore(),
			backColor().Regex(),
			regex.SpaceZeroOrMore(),
			regex.Optional(regex.NewConcat(uml.Lit("##"),
				uml.Leaf(2, "LINECOLOR", `(?:\[(dotted|dashed|bold)\])?(\w+)?`))),
			regex.SpaceZeroOrMore(),
			uml.Lit(`\{`),
			regex.End()),
		Strategy: uml.RemoveStartingQuote,
		TrimEnd:  uml.TrimBoth,
		End:      uml.Compile(`^[%s]*\}[%s]*$`),
	}, func(c *uml.Multilines2[*ClassDiagram], d *ClassDiagram, lines command.BlocLines) command.Result {
		lines = lines.Trim().RemoveEmptyLines()
		line0 := c.Start.Matcher(lines.At(0).Trimmed().S)
		name := line0.Str("NAME", 1)
		quark := d.QuarkInContext(true, d.CleanID(name))
		if quark.Data() != nil {
			return command.Error("Map already exists: " + name)
		}
		displayString, hasDisplay := line0.Get("NAME", 0)
		display := objectDisplay(d, displayString, hasDisplay, name)
		entity1 := d.ReallyCreateLeaf(uml.Location(lines), quark, display, abel.LeafTypeMap, abel.USymbolNull)
		if r := setStereotypeSpotted(d, entity1, line0); !r.IsOK() {
			return r
		}
		colors, err := classColors(line0)
		if err != nil {
			return abel.ColorResult(err)
		}
		entity1.SetColors(colors)
		lines = uml.SubExtract(lines, 1, 1)
		for _, sl := range lines.Lines() {
			line := sl.S
			if !entity1.Bodier().AddFieldOrMethod(line) {
				return command.Error("Map definition should contains key => value")
			}
			linkStr, ok := abel.BodierMapLinkedEntry(line)
			if !ok {
				continue
			}
			x := strings.Index(line, linkStr)
			key := text.Trin(line[:x])
			dest := text.Trin(line[x+len(linkStr):])
			ident2 := d.QuarkInContext(true, dest)
			entity2 := ident2.Data()
			if entity2 == nil {
				return command.Error("No such entity " + ident2.Name())
			}
			linkType := decoration.NewLinkType(decoration.LinkDecorArrow, decoration.LinkDecorNone)
			length := uml.JavaLength(linkStr) - 2
			link := abel.NewLink(uml.Location(lines), &d.CucaDiagram, d.SkinParam().CurrentStyleBuilder(), entity1,
				entity2, linkType, abel.LinkArgNoDisplay(length))
			link.SetPortMembers(abel.Some(key), abel.OptString{})
			d.AddLink(link)
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
type jsonMultilines struct {
	*uml.Multilines2[*ClassDiagram]
}

// IsValid implements command.Command.
func (c jsonMultilines) IsValid(lines command.BlocLines) command.Control {
	result := c.Multilines2.IsValid(lines)
	if result != command.OK {
		return result
	}
	if jsonValue(uml.RemoveStartingQuote.CleanList(lines)) == nil {
		return command.OKPartial
	}
	return command.OK
}

// commandCreateJSON is `json code {` ... `}`, the body being JSON data.
func commandCreateJSON() cmd {
	return jsonMultilines{abel.Multilines(&uml.Multilines2[*ClassDiagram]{
		Start: regex.NewConcat(regex.Start(),
			regex.NewNamedLeaf(0, "TYPE", "json"),
			regex.SpaceOneOrMore(),
			abel.NameAndCode(),
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
	}, func(c *uml.Multilines2[*ClassDiagram], d *ClassDiagram, lines command.BlocLines) command.Result {
		lines = lines.Trim().RemoveEmptyLines()
		line0 := c.Start.Matcher(lines.At(0).Trimmed().S)
		code, _ := line0.Lazy("CODE", 0)
		quark := d.QuarkInContext(true, d.CleanID(code))
		if quark.Data() != nil {
			return command.Error("JSON already exists: " + code)
		}
		displayString, hasDisplay := line0.Lazy("DISPLAY", 0)
		display := objectDisplay(d, displayString, hasDisplay, quark.Name())
		entity1 := d.ReallyCreateLeaf(uml.Location(lines), quark, display, abel.LeafTypeJSON, abel.USymbolNull)
		if r := setStereotypeSpotted(d, entity1, line0); !r.IsOK() {
			return r
		}
		if r := setBackColor(entity1, line0); !r.IsOK() {
			return r
		}
		value := jsonValue(lines)
		if value == nil {
			return command.Error("Bad data")
		}
		entity1.Bodier().(*abel.BodierJSon).SetJSON(value)
		return command.Ok()
	})}
}

// commandCreateJSONSingleLine is `json code <value>`, the value being a
// JSON literal, number, string, array or object.
func commandCreateJSONSingleLine() cmd {
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
		func(d *ClassDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			name := arg.Str("NAME", 1)
			quark := d.QuarkInContext(false, d.CleanID(name))
			if quark.Data() != nil {
				return command.Error("JSON already exists: " + name)
			}
			displayString, hasDisplay := arg.Get("NAME", 0)
			display := objectDisplay(d, displayString, hasDisplay, name)
			entity1 := d.ReallyCreateLeaf(location, quark, display, abel.LeafTypeJSON, abel.USymbolNull)
			if r := setStereotypeSpotted(d, entity1, arg); !r.IsOK() {
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
			entity1.Bodier().(*abel.BodierJSon).SetJSON(value)
			return command.Ok()
		})
}
