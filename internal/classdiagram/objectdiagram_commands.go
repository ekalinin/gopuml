// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/objectdiagram/command/CommandCreateEntityObject.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/objectdiagram/command/CommandCreateEntityObjectMultilines.java
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
