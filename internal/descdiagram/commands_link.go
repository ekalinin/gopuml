// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/command/CommandLinkElement.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package descdiagram

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

// LineStyle is upstream's CommandLinkElement.LINE_STYLE: colors and
// options in "-[...]->".
const (
	lineStyleKey1 = `dotted|dashed|plain|bold|hidden|norank|single|node|thickness=\d+`
	lineStyleKey2 = `,dotted|,dashed|,plain|,bold|,hidden|,norank|,single|,node|,thickness=\d+`
	LineStyle     = `(?:#\w+|` + lineStyleKey1 + `)(?:,#\w+|` + lineStyleKey2 + `)*`
	// lineStyleMultiples are several line styles separated by ";".
	lineStyleMultiples = LineStyle + `(?:(?:;` + LineStyle + `)*)`
)

// UBrexLineStyle is upstream's CommandLinkElement.UBREX_LINE_STYLE: LineStyle
// written with UBrex, for the commands written with UBrex.
const (
	ubrexKey1      = "dotted┇dashed┇plain┇bold┇hidden┇norank┇single┇node┇thickness=〇+〴d"
	ubrexKey2      = ",dotted┇,dashed┇,plain┇,bold┇,hidden┇,norank┇,single┇,node┇,thickness=〇+〴d"
	UBrexLineStyle = "【 #〇+〴w ┇" + ubrexKey1 + "】 〇*【 ,#〇+〴w ┇" + ubrexKey2 + "】"
)

// linkElementEnd is upstream's CommandLinkElement.getGroup: a name, a
// quoted name, "() name", ":actor:", "[component]" or "(use case)". The
// negative lookaheads of upstream, which keep "[*]" and "(*)" for state
// diagrams, are rewritten for RE2 on the content of the brackets.
func linkElementEnd(name string) *regex.Leaf {
	return uml.Leaf(1, name, `(`+
		`[%pLN_.]+`+
		`|[%g][^%g]+[%g]`+
		`|\(\)[%s]*[%pLN_.]+`+
		`|\(\)[%s]*[%g][^%g]+[%g]`+
		`|:[^:]+:/?`+
		`|\[(?:[^\[\]*][^\[\]]*|\*[^\[\]]+)\]`+
		`|\((?:[^)*][^)]*|\*[^)]+)\)/?`+
		`)`)
}

const (
	linkDirection = `(left|right|up|down|le?|ri?|up?|do?)`
	linkInside    = `(0|\(0\)|\(0|0\))`
	linkStyle2    = `\[(` + LineStyle + `)\]`
)

// commandLinkElement is upstream's CommandLinkElement: a link between two
// description elements, created when needed with the shape of their
// notation.
//
// Upstream's DIRECTION leaf has the lookahead (?=[-=.~0()\[]): a direction
// is followed by the middle circle, a style or the second body. Its INSIDE
// leaf has the lookahead (?=[-=.~]): the middle circle is followed by a
// non-empty second body, without a style. RE2 lacks lookaheads, so the
// part between the first body and the second head is rewritten as
// alternatives in upstream's backtracking order: with a direction (and
// the middle circle, a style or a body), then without one. The rewrite
// does not accept a direction followed directly by a second head that
// starts with "0", "(" or ")", which upstream does.
func commandLinkElement() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		linkElementEnd("ENT1"),
		regex.SpaceZeroOrMore(),
		regex.Optional(uml.Leaf(1, "FIRST_LABEL", "[%g]([^%g]+)[%g]")),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "HEAD1", decoration.RegexDecors1()),
		uml.Leaf(1, "BODY1", "([-=.~]+)"),
		uml.Leaf(1, "ARROW_STYLE1", `(?:\[(`+lineStyleMultiples+`)\])?`),
		regex.Or("",
			regex.NewConcat(
				uml.Leaf(1, "DIRECTION", linkDirection),
				uml.Leaf(1, "INSIDE", linkInside),
				uml.Leaf(1, "BODY2", "([-=.~]+)")),
			regex.NewConcat(
				uml.Leaf(1, "DIRECTION_S", linkDirection),
				uml.Leaf(1, "STYLE2_S", linkStyle2),
				uml.Leaf(1, "BODY2_S", "([-=.~]*)")),
			regex.NewConcat(
				uml.Leaf(1, "DIRECTION_B", linkDirection),
				uml.Leaf(1, "BODY2_B", "([-=.~]+)")),
			regex.NewConcat(
				uml.Leaf(1, "INSIDE_N", linkInside),
				uml.Leaf(1, "BODY2_I", "([-=.~]+)")),
			regex.NewConcat(
				uml.Leaf(1, "STYLE2_N", `(?:`+linkStyle2+`)?`),
				uml.Leaf(1, "BODY2_N", "([-=.~]*)"))),
		uml.Leaf(1, "HEAD2", decoration.RegexDecors2()),
		regex.SpaceZeroOrMore(),
		regex.Optional(uml.Leaf(1, "SECOND_LABEL", "[%g]([^%g]+)[%g]")),
		regex.SpaceZeroOrMore(),
		linkElementEnd("ENT2"),
		regex.SpaceZeroOrMore(),
		color.SimpleColor(color.TypeLine).Regex(),
		klimt.StereotypePatternOptional("STEREOTYPE"),
		uml.Leaf(1, "LABEL_LINK", "(?::[%s]*(.+))?"),
		regex.End()),
		executeLinkElement)
}

// firstOf returns the first matched group of names.
func firstOf(arg *regex.Result, names ...string) (string, bool) {
	for _, name := range names {
		if s, ok := arg.Get(name, 0); ok {
			return s, true
		}
	}
	return "", false
}

func linkHead(arg *regex.Result, key string) string {
	return strings.ReplaceAll(uml.LowerCase(text.Trin(arg.Str(key, 0))), "_", "")
}

// linkQueue is upstream's getQueue: the first and second bodies.
func linkQueue(arg *regex.Result) string {
	body2, _ := firstOf(arg, "BODY2", "BODY2_S", "BODY2_B", "BODY2_I", "BODY2_N")
	return arg.Str("BODY1", 0) + body2
}

// linkElementType is upstream's getLinkType: the decorations of the heads
// (the first head is the decoration of the second end), the line style of
// the bodies, and the middle circle.
func linkElementType(arg *regex.Result) decoration.LinkType {
	d1 := decoration.LookupDecors1(linkHead(arg, "HEAD1"), true)
	d2 := decoration.LookupDecors2(linkHead(arg, "HEAD2"), true)
	result := decoration.NewLinkType(d2, d1)
	queue := linkQueue(arg)
	switch {
	case strings.Contains(queue, "."):
		result = result.GoDashed()
	case strings.Contains(queue, "~"):
		result = result.GoDotted()
	case strings.Contains(queue, "="):
		result = result.GoBold()
	}
	middle, _ := firstOf(arg, "INSIDE", "INSIDE_N")
	switch middle {
	case "0":
		result = result.WithMiddleCircle()
	case "0)":
		result = result.WithMiddleCircleCircled1()
	case "(0":
		result = result.WithMiddleCircleCircled2()
	case "(0)":
		result = result.WithMiddleCircleCircled()
	}
	return result
}

// linkElementDirection is upstream's getDirection: the direction written
// in the arrow, or the one of its bodies.
func linkElementDirection(arg *regex.Result) klimt.Direction {
	if dir, ok := firstOf(arg, "DIRECTION", "DIRECTION_S", "DIRECTION_B"); ok {
		return uml.QueueDirection(dir)
	}
	return uml.QueueDirection(linkQueue(arg))
}

// linkElementStyle is upstream's getLazzy("ARROW_STYLE", 0): the first
// style, then the second, in upstream's HashMap order.
func linkElementStyle(arg *regex.Result) (string, bool) {
	return firstOf(arg, "ARROW_STYLE1", "STYLE2_S", "STYLE2_N")
}

func executeLinkElement(d *DescriptionDiagram, location *text.LineLocation, arg *regex.Result,
	_ command.ParserPass) command.Result {
	ent1 := arg.Str("ENT1", 0)
	ent2 := arg.Str("ENT2", 0)
	ent1clean := d.CleanID(ent1)
	ent2clean := d.CleanID(ent2)
	linkType := linkElementType(arg)
	dir := linkElementDirection(arg)
	queue := linkQueue(arg)
	if dir == klimt.DirectionLeft || dir == klimt.DirectionRight {
		queue = "-"
	}
	labels := NewLabels(arg)
	var cl1, cl2 *abel.Entity
	if d.IsGroupNamed(ent1clean) && d.IsGroupNamed(ent2clean) {
		cl1 = d.Group(ent1clean)
		cl2 = d.Group(ent2clean)
	} else {
		cl1 = linkDummy(location, d, ent1)
		cl2 = linkDummy(location, d, ent2)
	}
	linkArg := abel.LinkArgBuildManage(labels.Display(d.Pragma()), uml.JavaLength(queue),
		d.SkinParam().ClassAttributeIconSize() > 0).
		WithQuantifier(labels.FirstLabel, labels.SecondLabel).
		WithDistanceAngle(d.Labeldistance(), d.Labelangle())
	link := abel.NewLink(location, &d.CucaDiagram, d.SkinParam().CurrentStyleBuilder(), cl1, cl2, linkType,
		linkArg)
	if dir == klimt.DirectionLeft || dir == klimt.DirectionUp {
		link = link.Inv()
	}
	link.SetLinkArrow(labels.LinkArrow)
	colors, err := color.SimpleColor(color.TypeLine).Color(arg, d.Skin().IHtmlColorSet())
	if err != nil {
		return abel.ColorResult(err)
	}
	link.SetColors(colors)
	link.ApplyStyle(linkElementStyle(arg))
	if stereotype, ok := arg.Get("STEREOTYPE", 0); ok {
		link.SetStereotype(klimt.StereotypeBuild(stereotype))
	}
	d.AddLink(link)
	return command.Ok()
}

// linkDummy is upstream's getDummy: the entity of a link end, created with
// the shape of its notation when it does not exist yet.
func linkDummy(location *text.LineLocation, d *DescriptionDiagram, ident string) *abel.Entity {
	if strings.HasPrefix(ident, "()") {
		ident = d.CleanID(ident)
		quark := d.QuarkInContext(true, ident)
		if quark.Data() != nil {
			return quark.Data()
		}
		return d.ReallyCreateLeaf(location, quark, klimt.DisplayWithNewlines(d.Pragma(), quark.Name()),
			abel.LeafTypeDescription, abel.USymbolInterface)
	}
	codeChar := charEncoding(ident, true)
	endWithSlash := strings.HasSuffix(ident, "/")
	ident = d.CleanID(ident)
	quark := d.QuarkInContext(true, ident)
	if d.IsGroupQuark(quark) {
		return quark.Data()
	}
	if quark.Data() != nil {
		return quark.Data()
	}
	display := klimt.DisplayWithNewlines(d.Pragma(), quark.Name())
	switch codeChar {
	case '(':
		if endWithSlash {
			return d.ReallyCreateLeaf(location, quark, display, abel.LeafTypeUsecaseBusiness,
				abel.USymbolUsecaseBusiness)
		}
		return d.ReallyCreateLeaf(location, quark, display, abel.LeafTypeUsecase, abel.USymbolUsecase)
	case ':':
		if endWithSlash {
			return d.ReallyCreateLeaf(location, quark, display, abel.LeafTypeDescription,
				abel.USymbolActorStickmanBusiness)
		}
		return d.ReallyCreateLeaf(location, quark, display, abel.LeafTypeDescription,
			abel.ActorStyleUSymbol(d.SkinParam().ActorStyle()))
	case '[':
		return d.ReallyCreateLeaf(location, quark, display, abel.LeafTypeDescription,
			abel.ComponentStyleUSymbol(d.SkinParam().ComponentStyle()))
	}
	return d.ReallyCreateLeaf(location, quark, display, abel.LeafTypeStillUnknown, abel.USymbolNull)
}
