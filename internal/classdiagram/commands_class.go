// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/classdiagram/command/CommandCreateClass.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/classdiagram/command/CommandCreateClassMultilines.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/classdiagram/command/GenericRegexProducer.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/NameAndCodeParser.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package classdiagram

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

type cmd = command.Command[*ClassDiagram]

// genericPattern is upstream's GenericRegexProducer.PATTERN: generic
// parameters with up to four levels of nested "<...>".
var genericPattern = `[^\<\>/]` + genericRegex(4)

func genericRegex(level int) string {
	const part1 = `(?:[^\<\>/]|\<`
	const part2 = `\>)*`
	if level == 0 {
		return part1 + part2
	}
	return part1 + genericRegex(level-1) + part2
}

// The separators of class names, upstream's CommandLinkClass constants: a
// single character that is not part of an identifier nor a syntactic
// character, or "::" or "\\".
const (
	separatorCharSingle = `[^%pLN%s_$#\:{}<>%g]`
	separatorCharDouble = `[\\]{2}|::`
	separator           = `(?:` + separatorCharSingle + `|` + separatorCharDouble + `)`
)

// classCodes is upstream's CommandCreateClassMultilines.CODES: class names
// separated by commas.
var classCodes = func() string {
	code := separator + `?[%pLN_$]+(?:` + separator + `[%pLN_$]+)*`
	return code + `(?:\s*,\s*` + code + `)*`
}()

const (
	nameCodeCode = `[^%s{}%g<>]+`
)

// nameAndCodeForClassWithGeneric is
// NameAndCodeParser.nameAndCodeForClassWithGeneric: the quoted displays
// may end with generic parameters ("Foo<T>" as code).
func nameAndCodeForClassWithGeneric() regex.Part {
	displayWithGeneric := `[%g](.+?)(?:\<(` + genericPattern + `)\>)?[%g]`
	return regex.Or("",
		regex.NewConcat(
			uml.Leaf(2, "DISPLAY1", displayWithGeneric),
			regex.SpaceOneOrMore(),
			uml.Lit("as"),
			regex.SpaceOneOrMore(),
			uml.Leaf(1, "CODE1", "("+nameCodeCode+")")),
		regex.NewConcat(
			uml.Leaf(1, "CODE2", "("+nameCodeCode+")"),
			regex.SpaceOneOrMore(),
			uml.Lit("as"),
			regex.SpaceOneOrMore(),
			uml.Leaf(2, "DISPLAY2", displayWithGeneric)),
		uml.Leaf(1, "CODE3", "("+nameCodeCode+")"),
		uml.Leaf(1, "CODE4", "[%g]([^%g]+)[%g]"))
}

func backColor() *color.ColorParser { return color.SimpleColor(color.TypeBack) }

// classHead is the start of the class declaration patterns, up to the
// line color; types are the accepted keywords.
func classHead(types string) []regex.Part {
	return []regex.Part{regex.Start(),
		uml.Leaf(1, "VISIBILITY", "("+klimt.VisibilityRegexForVisibilityCharacter()+")?"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "TYPE", "("+types+")"),
		regex.SpaceOneOrMore(),
		nameAndCodeForClassWithGeneric(),
		regex.Optional(regex.NewConcat(regex.SpaceZeroOrMore(),
			uml.Leaf(1, "GENERIC", `\<(`+genericPattern+`)\>`))),
		regex.SpaceZeroOrMore(),
		abel.TagsLeaf("TAGS1"),
		klimt.StereotypePatternOptional("STEREO"),
		abel.TagsLeaf("TAGS2"),
		regex.SpaceZeroOrMore(),
		klimt.URLBuilderOptional(),
		regex.SpaceZeroOrMore(),
		backColor().Regex(),
		regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewConcat(uml.Lit("##"),
			uml.Leaf(2, "LINECOLOR", `(?:\[(dotted|dashed|bold)\])?(\w+)?`))),
	}
}

// commandCreateClass is `class Name<T> <<stereotype>> #color ##line
// extends A implements B {}`, and the other class-like keywords.
func commandCreateClass() cmd {
	parts := append(classHead(
		"interface|enum|annotation|abstract[%s]+class|static[%s]+class|abstract|class|entity|circle|diamond|"+
			"protocol|struct|exception|metaclass|stereotype|dataclass|record|map"),
		regex.Optional(regex.NewConcat(regex.SpaceOneOrMore(),
			uml.Leaf(3, "EXTENDS", "(extends)[%s]+("+classCodes+"|[%g]([^%g]+)[%g])"))),
		regex.Optional(regex.NewConcat(regex.SpaceOneOrMore(),
			uml.Leaf(3, "IMPLEMENTS", "(implements)[%s]+("+classCodes+"|[%g]([^%g]+)[%g])"))),
		regex.Optional(regex.NewConcat(regex.SpaceZeroOrMore(), uml.Lit(`\{`),
			regex.SpaceZeroOrMore(), uml.Lit(`\}`))),
		regex.End())
	return uml.Single(regex.NewConcat(parts...),
		func(d *ClassDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			typeString := uml.UpperCase(arg.Str("TYPE", 0))
			typ := abel.LeafTypeOf(typeString)
			code, _ := arg.Lazy("CODE", 0)
			idShort := d.CleanID(code)
			displayString, hasDisplay := arg.Lazy("DISPLAY", 0)
			quark, failure, score := d.QuarkInContextSafe(false, idShort)
			if failure != "" {
				return command.ErrorScore(failure, score)
			}
			entity := quark.Data()
			if entity == nil {
				display := klimt.DisplayWithNewlinesString(d.Pragma(), displayString, hasDisplay)
				if klimt.DisplayIsNull(display) {
					display = klimt.DisplayWithNewlines(d.Pragma(), quark.Name()).WithCreoleMode(klimt.CreoleModeSimpleLine)
				}
				entity = d.ReallyCreateLeaf(location, quark, display, typ, abel.USymbolNull)
			} else if !entity.MuteToTypeSymbol(typ, abel.USymbolNull) {
				return command.Error("Bad name")
			}
			if r := d.CheckIfPackageHierarchyIsOk(entity); !r.IsOK() {
				return r
			}
			if r := setClassProperties(d, entity, arg); !r.IsOK() {
				return r
			}
			manageExtends(location, "EXTENDS", d, arg, entity)
			manageExtends(location, "IMPLEMENTS", d, arg, entity)
			abel.AddTags(entity, abel.Opt(arg.Lazy("TAGS", 0)))
			if strings.Contains(typeString, "STATIC") {
				entity.SetStatic(true)
			}
			return command.Ok()
		})
}

// setClassProperties sets what the class declarations share, in upstream's
// order: the last entity, visibility, stereotype, generic, link and
// colors.
func setClassProperties(d *ClassDiagram, entity *abel.Entity, arg *regex.Result) command.Result {
	d.SetLastEntity(entity)
	if visibilityString, ok := arg.Get("VISIBILITY", 0); ok {
		entity.SetVisibilityModifier(klimt.VisibilityOf(visibilityString+"FOO", false))
	} else {
		entity.SetVisibilityModifier(0, false)
	}
	if stereo, ok := arg.Get("STEREO", 0); ok {
		st, err := abel.SpottedStereotype(&d.CucaDiagram, stereo)
		if err != nil {
			return abel.ColorResult(err)
		}
		entity.SetStereotype(st)
		entity.SetStereostyle(stereo)
	}
	if generic, ok := classGeneric(arg); ok {
		entity.SetGeneric(generic)
	}
	if urlString, ok := arg.Get(klimt.URLBuilderKey, 0); ok {
		entity.AddURL(abel.URLBuilderOf(&d.CucaDiagram).URL(urlString))
	}
	colors, err := classColors(arg)
	if err != nil {
		return abel.ColorResult(err)
	}
	entity.SetColors(colors)
	return command.Ok()
}

// classGeneric is the generic parameters: in the quoted display, or after
// the name.
func classGeneric(arg *regex.Result) (string, bool) {
	if genericOption, ok := arg.Lazy("DISPLAY", 1); ok {
		return genericOption, true
	}
	return arg.Get("GENERIC", 0)
}

// classColors returns the background color, with the line color and line
// style given after "##".
func classColors(arg *regex.Result) (*color.Colors, error) {
	colors, err := backColor().Color(arg, color.Instance())
	if err != nil {
		return nil, err
	}
	if s, ok := arg.Get("LINECOLOR", 1); ok {
		lineColor, err := abel.GetColor(s)
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

// commandCreateClassMultilines is `class Name {` ... `}`, the body lines
// being fields and methods.
func commandCreateClassMultilines() cmd {
	extendsGeneric := regex.Optional(regex.NewConcat(regex.SpaceZeroOrMore(),
		uml.Leaf(1, "", `\<(`+genericPattern+`)\>`)))
	parts := append(classHead(
		"interface|enum|annotation|abstract[%s]+class|static[%s]+class|abstract|class|entity|protocol|struct|"+
			"exception|metaclass|stereotype|dataclass|record"),
		regex.Optional(regex.NewConcat(regex.SpaceOneOrMore(),
			uml.Leaf(3, "EXTENDS", "(extends)[%s]+("+classCodes+"|[%g]([^%g]+)[%g])"),
			extendsGeneric)),
		regex.Optional(regex.NewConcat(regex.SpaceOneOrMore(),
			uml.Leaf(3, "IMPLEMENTS", "(implements)[%s]+("+classCodes+"|[%g]([^%g]+)[%g])"),
			extendsGeneric)),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\{`),
		regex.SpaceZeroOrMore(),
		regex.End())
	return abel.Multilines(&uml.Multilines2[*ClassDiagram]{
		Start:        regex.NewConcat(parts...),
		Strategy:     uml.RemoveStartingQuote,
		TrimEnd:      uml.TrimBoth,
		End:          uml.Compile(`^[%s]*\}[%s]*$`),
		FinalBracket: true,
	}, func(c *uml.Multilines2[*ClassDiagram], d *ClassDiagram, lines command.BlocLines) command.Result {
		lines = uml.TrimSmart(lines, 1)
		line0 := c.Start.Matcher(lines.At(0).Trimmed().S)
		typeString := uml.UpperCase(line0.Str("TYPE", 0))
		typ := abel.LeafTypeOf(typeString)
		code, _ := line0.Lazy("CODE", 0)
		idShort := d.CleanID(code)
		displayString, hasDisplay := line0.Lazy("DISPLAY", 0)
		quark, failure, _ := d.QuarkInContextSafe(false, idShort)
		if failure != "" {
			return command.Error(failure)
		}
		entity := quark.Data()
		display := klimt.DisplayWithNewlinesString(d.Pragma(), displayString, hasDisplay)
		if entity == nil {
			if klimt.DisplayIsNull(display) {
				display = klimt.DisplayWithNewlines(d.Pragma(), quark.Name()).WithCreoleMode(klimt.CreoleModeSimpleLine)
			}
			entity = d.ReallyCreateLeaf(uml.Location(lines), quark, display, typ, abel.USymbolNull)
		} else {
			if !entity.MuteToTypeSymbol(typ, abel.USymbolNull) {
				return command.Error("Cannot create " + idShort + " because it already exists")
			}
			if !klimt.DisplayIsNull(display) {
				entity.SetDisplay(display)
			}
		}
		if r := d.CheckIfPackageHierarchyIsOk(entity); !r.IsOK() {
			return r
		}
		d.SetLastEntity(entity)
		if visibilityString, ok := line0.Get("VISIBILITY", 0); ok {
			entity.SetVisibilityModifier(klimt.VisibilityOf(visibilityString+"FOO", false))
		} else {
			entity.SetVisibilityModifier(0, false)
		}
		if stereo, ok := line0.Get("STEREO", 0); ok {
			st, err := abel.SpottedStereotype(&d.CucaDiagram, stereo)
			if err != nil {
				return abel.ColorResult(err)
			}
			entity.SetStereotype(st)
			entity.SetStereostyle(stereo)
		}
		if urlString, ok := line0.Get(klimt.URLBuilderKey, 0); ok {
			entity.AddURL(abel.URLBuilderOf(&d.CucaDiagram).URL(urlString))
		}
		colors, err := classColors(line0)
		if err != nil {
			return abel.ColorResult(err)
		}
		entity.SetColors(colors)
		if generic, ok := classGeneric(line0); ok {
			entity.SetGeneric(generic)
		}
		if strings.Contains(typeString, "STATIC") {
			entity.SetStatic(true)
		}
		if lines.Size() > 1 {
			lines = uml.SubExtract(lines, 1, 1)
			for _, s := range lines.Lines() {
				if s.S != "" && klimt.VisibilityIsVisibilityCharacter(s.S) {
					d.SetVisibilityModifierPresent(true)
				}
				entity.Bodier().AddFieldOrMethod(s.S)
			}
		}
		manageExtends(uml.Location(lines), "EXTENDS", d, line0, entity)
		manageExtends(uml.Location(lines), "IMPLEMENTS", d, line0, entity)
		abel.AddTags(entity, abel.Opt(line0.Lazy("TAGS", 0)))
		return command.Ok()
	})
}

// manageExtends is upstream's CommandCreateClassMultilines.manageExtends:
// "extends A, B" or "implements C" add the parents, as classes or
// interfaces when they do not exist, and the extension links.
func manageExtends(location *text.LineLocation, keyword string, d *ClassDiagram, arg *regex.Result,
	entity *abel.Entity) {
	kw, ok := arg.Get(keyword, 0)
	if !ok {
		return
	}
	extends := uml.EqualsIgnoreCase(kw, "extends")
	type2 := abel.LeafTypeClass
	if !extends {
		type2 = abel.LeafTypeInterface
	}
	if extends && entity.LeafType() == abel.LeafTypeInterface {
		type2 = abel.LeafTypeInterface
	}
	codes := uml.RemoveQuotes(arg.Str(keyword, 1))
	for _, s := range splitTrailing(codes, ",") {
		idShort := text.Trin(s)
		quark := d.QuarkInContext(false, d.CleanID(idShort))
		cl2 := quark.Data()
		if cl2 == nil {
			cl2 = d.ReallyCreateLeaf(location, quark, klimt.DisplayWithNewlines(d.Pragma(), quark.Name()), type2,
				abel.USymbolNull)
		}
		typeLink := decoration.NewLinkType(decoration.LinkDecorNone, decoration.LinkDecorExtends)
		if type2 == abel.LeafTypeInterface && entity.LeafType() != abel.LeafTypeInterface {
			typeLink = typeLink.GoDashed()
		}
		linkArg := abel.LinkArgNoDisplay(2)
		link := abel.NewLink(location, &d.CucaDiagram, d.SkinParam().CurrentStyleBuilder(), cl2, entity, typeLink,
			linkArg.WithQuantifier(abel.OptString{}, abel.OptString{}).WithDistanceAngle(d.Labeldistance(),
				d.Labelangle()))
		d.AddLink(link)
	}
}

// splitTrailing is Java's String.split for a literal separator: the
// trailing empty pieces are removed.
func splitTrailing(s, sep string) []string {
	parts := strings.Split(s, sep)
	for len(parts) > 1 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	if len(parts) == 1 && parts[0] == "" && s != "" {
		return nil
	}
	return parts
}
