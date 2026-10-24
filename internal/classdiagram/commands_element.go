// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/classdiagram/command/CommandCreateElementFull2.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/descdiagram/command/CommandCreateElementParenthesis.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package classdiagram

import (
	"regexp"
	"strings"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// elementForbidden is the FORBIDDEN_PATTERN of the element commands: a
// line made of a single name.
var elementForbidden = regexp.MustCompile(`^[\p{L}0-9_.]+$`)

func isElementForbidden(line string) bool { return elementForbidden.MatchString(line) }

// charEncoding is upstream's getCharEncoding: the first UTF-16 unit of a
// text longer than two units, 0 otherwise (and for null).
func charEncoding(s string, ok bool) rune {
	if !ok {
		return 0
	}
	u := utf16.Encode([]rune(s))
	if len(u) > 2 {
		return rune(u[0])
	}
	return 0
}

// Element modes of upstream's CommandCreateElementFull2.
const (
	modeNormalKeyword = iota
	modeWithMixPrefix
)

// commandCreateElementFull2 is a description element in a class diagram:
// "component foo", "(use case)", ":actor:"... It needs "allowmixing",
// unless written with the "mix_" prefix (mode modeWithMixPrefix).
func commandCreateElementFull2(mode int) cmd {
	parts := []regex.Part{regex.Start()}
	if mode == modeWithMixPrefix {
		parts = append(parts, uml.Lit("mix_"))
	}
	parts = append(parts,
		uml.Leaf(1, "SYMBOL", "(state|"+abel.ElementAllTypes+")"),
		regex.SpaceOneOrMore(),
		regex.Or("",
			uml.Leaf(1, "CODE1", abel.ElementCodeWithQuote),
			regex.NewConcat(
				uml.Leaf(1, "DISPLAY2", abel.ElementDisplay),
				klimt.StereotypePatternOptional("STEREOTYPE2"),
				uml.Lit("as"),
				regex.SpaceOneOrMore(),
				uml.Leaf(1, "CODE2", abel.ElementCode))),
		regex.SpaceZeroOrMore(),
		abel.TagsLeaf("TAGS1"),
		klimt.StereotypePatternOptional("STEREOTYPE"),
		abel.TagsLeaf("TAGS2"),
		regex.SpaceZeroOrMore(),
		klimt.URLBuilderOptional(),
		regex.SpaceZeroOrMore(),
		color.Exp1(),
		regex.End())
	c := uml.Single(regex.NewConcat(parts...),
		func(d *ClassDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			if mode == modeNormalKeyword && !d.IsAllowMixing() {
				return command.Error("Use 'allowmixing' if you want to mix classes and other UML elements.")
			}
			codeRaw, _ := arg.Lazy("CODE", 0)
			displayRaw, hasDisplayRaw := abel.RemoveQuotesOpt(arg.Lazy("DISPLAY", 0))
			codeChar := charEncoding(codeRaw, true)
			codeDisplay := charEncoding(displayRaw, hasDisplayRaw)
			var symbol string
			switch {
			case strings.HasPrefix(codeRaw, "()"):
				symbol = "interface"
				codeRaw = uml.RemoveQuotes(text.Trin(codeRaw[2:]))
			case codeChar == '(' || codeDisplay == '(':
				symbol = "usecase"
			case codeChar == ':' || codeDisplay == ':':
				symbol = "actor"
			case codeChar == '[' || codeDisplay == '[':
				symbol = "component"
			default:
				symbol = arg.Str("SYMBOL", 0)
			}
			var typ abel.LeafType
			usymbol := abel.USymbolNull
			switch {
			case uml.EqualsIgnoreCase(symbol, "port"), uml.EqualsIgnoreCase(symbol, "portin"):
				typ = abel.LeafTypePortin
			case uml.EqualsIgnoreCase(symbol, "portout"):
				typ = abel.LeafTypePortout
			case uml.EqualsIgnoreCase(symbol, "usecase"):
				typ = abel.LeafTypeUsecase
			case uml.EqualsIgnoreCase(symbol, "usecase/"):
				typ = abel.LeafTypeUsecaseBusiness
			case uml.EqualsIgnoreCase(symbol, "state"):
				typ = abel.LeafTypeState
			default:
				typ = abel.LeafTypeDescription
				usymbol = abel.USymbolFromSkin(symbol, d.SkinParam())
				if usymbol == abel.USymbolNull {
					uml.Throw(uml.IllegalStateException, "")
				}
			}
			idShort := uml.RemoveQuotes(codeRaw)
			displayText := idShort
			if hasDisplayRaw {
				displayText = displayRaw
			}
			display := klimt.DisplayWithNewlines(d.Pragma(), displayText)
			quark := d.QuarkInContext(true, idShort)
			entity := quark.Data()
			if entity == nil {
				entity = d.ReallyCreateLeaf(location, quark, display, typ, usymbol)
			}
			entity.SetDisplay(display)
			if stereotype, ok := arg.Lazy("STEREOTYPE", 0); ok {
				st, err := abel.SpottedStereotype(&d.CucaDiagram, stereotype)
				if err != nil {
					return abel.ColorResult(err)
				}
				entity.SetStereotype(st)
			}
			abel.AddTags(entity, abel.Opt(arg.Lazy("TAGS", 0)))
			if urlString, ok := arg.Get(klimt.URLBuilderKey, 0); ok {
				entity.AddURL(abel.URLBuilderOf(&d.CucaDiagram).URL(urlString))
			}
			return setBackColor(entity, arg)
		})
	c.Forbidden = isElementForbidden
	return c
}

// commandCreateElementParenthesis is "() name": an interface drawn as a
// circle.
func commandCreateElementParenthesis() cmd {
	stereo := func(name string) regex.Part {
		return regex.Optional(regex.NewConcat(regex.SpaceOneOrMore(), uml.Leaf(1, name, `(\<\<.+\>\>)`)))
	}
	c := uml.Single(regex.NewConcat(regex.Start(),
		uml.Lit(`\(\)[%s]+`),
		color.SimpleColorID(color.TypeBack, "COLOR2").Regex(),
		regex.SpaceZeroOrMore(),
		regex.Or("",
			uml.Leaf(1, "CODE1", abel.ElementCodeWithQuote),
			regex.NewConcat(
				uml.Leaf(1, "DISPLAY2", abel.ElementDisplay),
				stereo("STEREOTYPE2"),
				regex.SpaceZeroOrMore(),
				uml.Lit("as"),
				regex.SpaceOneOrMore(),
				uml.Leaf(1, "CODE2", abel.ElementCode)),
			regex.NewConcat(
				uml.Leaf(1, "CODE3", abel.ElementCode),
				stereo("STEREOTYPE3"),
				regex.SpaceOneOrMore(),
				uml.Lit("as"),
				regex.SpaceZeroOrMore(),
				uml.Leaf(1, "DISPLAY3", abel.ElementDisplay)),
			regex.NewConcat(
				uml.Leaf(1, "DISPLAY4", abel.ElementDisplayWithoutQuote),
				stereo("STEREOTYPE4"),
				regex.SpaceZeroOrMore(),
				uml.Lit("as"),
				regex.SpaceOneOrMore(),
				uml.Leaf(1, "CODE4", abel.ElementCode))),
		regex.Optional(regex.NewConcat(regex.SpaceZeroOrMore(), uml.Leaf(1, "STEREOTYPE", `(\<\<.+\>\>)`))),
		regex.SpaceZeroOrMore(),
		klimt.URLBuilderOptional(),
		regex.SpaceZeroOrMore(),
		backColor().Regex(), regex.End()),
		func(d *ClassDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			codeRaw, hasCode := arg.Lazy("CODE", 0)
			displayRaw, hasDisplay := arg.Lazy("DISPLAY", 0)
			usymbol := abel.USymbolFromSkin("interface", d.SkinParam())
			quark := abel.QuarkInContextOpt(&d.CucaDiagram, false, d.CleanID(codeRaw), hasCode)
			if quark.Data() != nil {
				return command.Error("This element (" + quark.Name() + ") is already defined")
			}
			display := displayRaw
			if !hasDisplay {
				display = quark.Name()
			}
			display = uml.RemoveQuotes(display)
			if abel.ExistsWithBadType3(quark, abel.LeafTypeDescription, usymbol) {
				return command.Error("This element (" + quark.Name() + ") is already defined")
			}
			entity := d.ReallyCreateLeaf(location, quark, klimt.DisplayWithNewlines(d.Pragma(), display),
				abel.LeafTypeDescription, usymbol)
			entity.SetDisplay(klimt.DisplayWithNewlines(d.Pragma(), display))
			if stereotype, ok := arg.Lazy("STEREOTYPE", 0); ok {
				st, err := abel.SpottedStereotype(&d.CucaDiagram, stereotype)
				if err != nil {
					return abel.ColorResult(err)
				}
				entity.SetStereotype(st)
			}
			if urlString, ok := arg.Get(klimt.URLBuilderKey, 0); ok {
				entity.AddURL(abel.URLBuilderOf(&d.CucaDiagram).URL(urlString))
			}
			colors, err := backColor().Color(arg, color.Instance())
			if err != nil {
				return abel.ColorResult(err)
			}
			entity.SetColors(colors)
			return command.Ok()
		})
	c.Forbidden = isElementForbidden
	return c
}
