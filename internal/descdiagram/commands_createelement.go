// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/descdiagram/command/CommandCreateElementFull.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package descdiagram

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

type cmd = command.Command[*DescriptionDiagram]

// elementForbidden is upstream's FORBIDDEN_PATTERN: a line made of a
// single name is not an element declaration.
var elementForbidden = regexp.MustCompile(`^[\p{L}0-9_.]+$`)

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

// elementSymbol is the symbol deduction of executeArg: the shape of the
// code or of the display ("()...", "(...)", ":...:", "[...]") overrides the
// keyword, and a trailing "/" selects the business variant of use cases
// and actors. It returns the symbol (false for none) and the code and
// display without the "/".
func elementSymbol(arg *regex.Result, codeRaw, displayRaw string, hasDisplay bool) (string, bool, string,
	string) {
	codeChar := charEncoding(codeRaw, true)
	codeDisplay := charEncoding(displayRaw, hasDisplay)
	keyword, hasKeyword := arg.Get("SYMBOL", 0)
	business := hasKeyword && strings.HasSuffix(keyword, "/")
	switch {
	case strings.HasPrefix(codeRaw, "()"):
		return "interface", true, uml.RemoveQuotes(text.Trin(codeRaw[2:])), displayRaw
	case codeChar == '(' || codeDisplay == '(':
		switch {
		case business:
			return "usecase/", true, codeRaw, displayRaw
		case hasDisplay && strings.HasSuffix(displayRaw, ")/"):
			return "usecase/", true, codeRaw, displayRaw[:len(displayRaw)-1]
		case strings.HasSuffix(codeRaw, ")/"):
			return "usecase/", true, codeRaw[:len(codeRaw)-1], displayRaw
		}
		return "usecase", true, codeRaw, displayRaw
	case codeChar == ':' || codeDisplay == ':':
		switch {
		case business:
			return "actor/", true, codeRaw, displayRaw
		case hasDisplay && strings.HasSuffix(displayRaw, ":/"):
			return "actor/", true, codeRaw, displayRaw[:len(displayRaw)-1]
		case strings.HasSuffix(codeRaw, ":/"):
			return "actor/", true, codeRaw[:len(codeRaw)-1], displayRaw
		}
		return "actor", true, codeRaw, displayRaw
	case codeChar == '[' || codeDisplay == '[':
		return "component", true, codeRaw, displayRaw
	}
	return keyword, hasKeyword, codeRaw, displayRaw
}

// commandCreateElementFull is upstream's CommandCreateElementFull: an
// element declared with a keyword ("node Server"), a short form
// ("(Login)", ":Admin:", "[API]", "() REST"), or both, with an optional
// display, alias, stereotype, tags, URL and color.
func commandCreateElementFull() cmd {
	c := uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "SYMBOL", "(?:("+abel.ElementAllTypes+`|\(\))[%s]+)?`),
		color.SimpleColorID(color.TypeBack, "COLOR2").Regex(),
		regex.SpaceZeroOrMore(),
		regex.Or("",
			uml.Leaf(1, "CODE1", abel.ElementCodeWithQuote),
			regex.NewConcat(
				uml.Leaf(1, "DISPLAY2", abel.ElementDisplay),
				klimt.StereotypePatternOptional("STEREOTYPE2"),
				uml.Lit("as"),
				regex.SpaceOneOrMore(),
				uml.Leaf(1, "CODE2", abel.ElementCode)),
			regex.NewConcat(
				uml.Leaf(1, "CODE3", abel.ElementCode),
				klimt.StereotypePatternOptional("STEREOTYPE3"),
				uml.Lit("as"),
				regex.SpaceZeroOrMore(),
				uml.Leaf(1, "DISPLAY3", abel.ElementDisplay)),
			regex.NewConcat(
				uml.Leaf(1, "DISPLAY4", abel.ElementDisplayWithoutQuote),
				klimt.StereotypePatternOptional("STEREOTYPE4"),
				uml.Lit("as"),
				regex.SpaceOneOrMore(),
				uml.Leaf(1, "CODE4", abel.ElementCode))),
		regex.SpaceZeroOrMore(),
		abel.TagsLeaf("TAGS1"),
		klimt.StereotypePatternOptional("STEREOTYPE"),
		abel.TagsLeaf("TAGS2"),
		regex.SpaceZeroOrMore(),
		klimt.URLBuilderOptional(),
		regex.SpaceZeroOrMore(),
		color.SimpleColor(color.TypeBack).Regex(),
		regex.End()),
		func(d *DescriptionDiagram, location *text.LineLocation, arg *regex.Result,
			_ command.ParserPass) command.Result {
			codeRaw, _ := arg.Lazy("CODE", 0)
			displayRaw, hasDisplay := arg.Lazy("DISPLAY", 0)
			symbol, hasSymbol, codeRaw, displayRaw := elementSymbol(arg, codeRaw, displayRaw, hasDisplay)
			var typ abel.LeafType
			usymbol := abel.USymbolNull
			switch {
			case !hasSymbol:
				typ = abel.LeafTypeDescription
				usymbol = abel.ActorStyleUSymbol(d.SkinParam().ActorStyle())
			case uml.EqualsIgnoreCase(symbol, "portin"), uml.EqualsIgnoreCase(symbol, "port"):
				typ = abel.LeafTypePortin
			case uml.EqualsIgnoreCase(symbol, "portout"):
				typ = abel.LeafTypePortout
			case uml.EqualsIgnoreCase(symbol, "usecase"):
				typ = abel.LeafTypeUsecase
			case uml.EqualsIgnoreCase(symbol, "usecase/"):
				typ = abel.LeafTypeUsecaseBusiness
			case uml.EqualsIgnoreCase(symbol, "circle"):
				typ = abel.LeafTypeCircle
			default:
				typ = abel.LeafTypeDescription
				usymbol = abel.USymbolFromSkin(symbol, d.SkinParam())
				if usymbol == abel.USymbolNull {
					uml.Throw(uml.IllegalStateException, "")
				}
			}
			quark := d.QuarkInContext(false, d.CleanID(codeRaw))
			if d.IsGroupQuark(quark) {
				return command.Error("This element (" + quark.Name() + ") is already defined")
			}
			display := displayRaw
			if !hasDisplay {
				display = quark.Name()
			}
			display = uml.RemoveQuotes(display)
			if abel.ExistsWithBadType3(quark, typ, usymbol) {
				return command.Error("This element (" + quark.Name() + ") is already defined")
			}
			if (typ == abel.LeafTypePortin || typ == abel.LeafTypePortout) && d.CurrentGroup().IsRoot() {
				return command.Error("Port can only be used inside an element and not at root level")
			}
			entity := quark.Data()
			if entity == nil {
				entity = d.ReallyCreateLeaf(location, quark, klimt.DisplayWithNewlines(d.Pragma(), display), typ,
					usymbol)
			}
			entity.SetDisplay(klimt.DisplayWithNewlines(d.Pragma(), display))
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
			colors, err := color.SimpleColor(color.TypeBack).Color(arg, d.Skin().IHtmlColorSet())
			if err != nil {
				return abel.ColorResult(err)
			}
			if s, ok := arg.Get("LINECOLOR", 1); ok {
				lineColor, err := abel.GetColor(s)
				if err != nil {
					return abel.ColorResult(err)
				}
				if lineColor != nil {
					colors = colors.Add(color.TypeLine, lineColor)
				}
			}
			entity.SetColors(colors)
			return command.Ok()
		})
	c.Forbidden = elementForbidden.MatchString
	return c
}
