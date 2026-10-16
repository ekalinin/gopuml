// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/classdiagram/command/CommandLinkClass.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/classdiagram/command/CommandLinkLollipop.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/command/Labels.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/command/StringWithArrow.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/command/CommandLinkElement.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/StringUtils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package classdiagram

import (
	"regexp"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/abel"
	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/command/regex"
	"github.com/ekalinin/plantuml-in-go/internal/decoration"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/text"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// The class name patterns of upstream's CommandLinkClass.
var (
	singleName          = separator + `?[%pLN_]+(?:` + separator + `[%pLN_]+)*`
	singleNameGuillemet = `[%g]` + separator + `?[%pLN_]+(?:` + separator + `[%pLN_]+)*[%g]`
	singleName2         = `(?:` + singleName + `|` + singleNameGuillemet + `)`
	couple              = `\([%s]*(` + singleName2 + `)[%s]*,[%s]*(` + singleName2 + `)[%s]*\)`
	// classIdentifier is upstream's CommandLinkClass.getClassIdentifier.
	classIdentifier = `(` + separator + `?[%pLN_$]+(?:` + separator + `[%pLN_$]+)*|[%g][^%g]+[%g])`
)

// lineStyle is upstream's CommandLinkElement.LINE_STYLE: colors and
// options in "-[...]->".
const (
	lineStyleKey1 = `dotted|dashed|plain|bold|hidden|norank|single|node|thickness=\d+`
	lineStyleKey2 = `,dotted|,dashed|,plain|,bold|,hidden|,norank|,single|,node|,thickness=\d+`
	lineStyle     = `(?:#\w+|` + lineStyleKey1 + `)(?:,#\w+|` + lineStyleKey2 + `)*`
)

func qualifierPart(name string) regex.Part {
	return regex.Optional(regex.NewConcat(
		regex.SpaceOneOrMore(),
		uml.Lit(`[\[]`),
		uml.Leaf(1, name, `([^\[\]]+)`),
		uml.Lit(`[\]]`),
		regex.SpaceOneOrMore()))
}

// commandLinkClass is a relation: "A "1" *-- "many" B : label >", with the
// heads, the body ("--", "..", "-up->"...), options in brackets, the
// association classes "(A, B) .. C" and the weight header "@2".
//
// Upstream's INSIDE leaf, (0|\(0\)|\(0|0\))(?=[-=.~]), has a lookahead
// that RE2 lacks: the middle circle must be followed by the second body.
// The pattern is rewritten as two alternatives, the circle with a
// non-empty second body (ARROW_BODY2B), then the optional style and the
// second body; their priority is the backtracking order of upstream. The
// rewrite does not accept a circle followed by '~' (the start of a second
// name whose first character is the separator '~'), which upstream does.
func commandLinkClass() cmd {
	style := `(?:\[(` + lineStyle + `)\])?`
	return uml.Single(regex.NewConcat(regex.Start(),
		regex.Optional(regex.NewConcat(
			uml.Leaf(1, "HEADER", `@([\d.]+)`),
			regex.SpaceOneOrMore())),
		regex.Or("",
			uml.Leaf(1, "ENT1", classIdentifier),
			uml.Leaf(2, "COUPLE1", couple)),
		regex.SpaceZeroOrMore(),
		qualifierPart("QUALIFIER1"),
		regex.Optional(uml.Leaf(1, "FIRST_LABEL", "[%g]([^%g]+)[%g]")),
		regex.Optional(uml.Leaf(1, "FIRST_ROLE", "/([^%s]+|[%g][^%g]+[%g])")),
		regex.SpaceZeroOrMore(),
		regex.NewConcat(
			uml.Leaf(1, "ARROW_HEAD1", decoration.RegexDecors1()),
			uml.Leaf(1, "ARROW_BODY1", "([-=.]+)"),
			uml.Leaf(1, "ARROW_STYLE1", style),
			uml.Leaf(1, "ARROW_DIRECTION", "(left|right|up|down|le?|ri?|up?|do?)?"),
			regex.Or("",
				regex.NewConcat(
					uml.Leaf(1, "INSIDE", `(0|\(0\)|\(0|0\))`),
					uml.Leaf(1, "ARROW_BODY2B", "([-=.]+)")),
				regex.NewConcat(
					uml.Leaf(1, "ARROW_STYLE2", style),
					uml.Leaf(1, "ARROW_BODY2", "([-=.]*)"))),
			uml.Leaf(1, "ARROW_HEAD2", decoration.RegexDecors2())),
		regex.SpaceZeroOrMore(),
		regex.Optional(uml.Leaf(1, "SECOND_LABEL", "[%g]([^%g]+)[%g]")),
		regex.Optional(uml.Leaf(1, "SECOND_ROLE", "/([^%s]+|[%g][^%g]+[%g])")),
		qualifierPart("QUALIFIER2"),
		regex.SpaceZeroOrMore(),
		regex.Or("",
			uml.Leaf(1, "ENT2", classIdentifier),
			uml.Leaf(2, "COUPLE2", couple)),
		regex.SpaceZeroOrMore(),
		color.SimpleColor(color.TypeLine).Regex(),
		regex.SpaceZeroOrMore(),
		klimt.URLBuilderOptional(),
		regex.SpaceZeroOrMore(),
		klimt.StereotypePatternOptional("STEREOTYPE"),
		regex.Optional(regex.NewConcat(
			uml.Lit(":"),
			regex.SpaceZeroOrMore(),
			uml.Leaf(1, "LABEL_LINK", "(.+)"))),
		regex.End()),
		executeLinkClass)
}

func executeLinkClass(d *ClassDiagram, location *text.LineLocation, arg *regex.Result,
	_ command.ParserPass) command.Result {
	ent1, hasEnt1 := arg.Get("ENT1", 0)
	ent2, hasEnt2 := arg.Get("ENT2", 0)
	switch {
	case !hasEnt1 && !hasEnt2:
		return executeLinkSpecial3(location, d, arg)
	case !hasEnt1:
		return executeLinkSpecial(location, d, arg, 1)
	case !hasEnt2:
		return executeLinkSpecial(location, d, arg, 2)
	}
	ent1String := d.CleanID(ent1)
	ent2String := d.CleanID(ent2)
	var port1, port2 abel.OptString
	linkType := linkClassType(arg)
	if strings.Contains(ent1String, "::") && d.FirstWithName(ent1String) == nil {
		port1 = d.PortID(ent1String)
		ent1String = d.RemovePortID(ent1String)
	}
	if strings.Contains(ent2String, "::") && d.FirstWithName(ent2String) == nil {
		port2 = d.PortID(ent2String)
		ent2String = d.RemovePortID(ent2String)
	}
	quark1, failure, _ := d.QuarkInContextSafe(true, ent1String)
	if failure != "" {
		return command.Error(failure)
	}
	quark2, failure, _ := d.QuarkInContextSafe(true, ent2String)
	if failure != "" {
		return command.Error(failure)
	}
	cl1 := quark1.Data()
	if cl1 == nil {
		cl1 = d.ReallyCreateLeaf(location, quark1, klimt.DisplayWithNewlines(d.Pragma(), quark1.Name()),
			abel.LeafTypeClass, abel.USymbolNull)
	}
	cl2 := quark2.Data()
	if cl2 == nil {
		cl2 = d.ReallyCreateLeaf(location, quark2, klimt.DisplayWithNewlines(d.Pragma(), quark2.Name()),
			abel.LeafTypeClass, abel.USymbolNull)
	}
	dir := linkDirection(arg)
	queue := 1
	if dir != klimt.DirectionLeft && dir != klimt.DirectionRight {
		queue = linkQueueLength(arg)
	}
	labels := newLabels(arg)
	linkArg := abel.LinkArgBuildManage(labels.display(d.Pragma()), queue, d.SkinParam().ClassAttributeIconSize() > 0).
		WithQuantifier(labels.firstLabel, labels.secondLabel).
		WithRole(labels.firstRole, labels.secondRole).
		WithDistanceAngle(d.Labeldistance(), d.Labelangle()).
		WithKal(abel.Opt(arg.Get("QUALIFIER1", 0)), abel.Opt(arg.Get("QUALIFIER2", 0)))
	link := abel.NewLink(location, &d.CucaDiagram, d.SkinParam().CurrentStyleBuilder(), cl1, cl2, linkType, linkArg)
	if urlString, ok := arg.Get(klimt.URLBuilderKey, 0); ok {
		link.SetURL(abel.URLBuilderOf(&d.CucaDiagram).URL(urlString))
	}
	link.SetPortMembers(port1, port2)
	if dir == klimt.DirectionLeft || dir == klimt.DirectionUp {
		link = link.Inv()
	}
	link.SetLinkArrow(labels.linkArrow)
	colors, err := color.SimpleColor(color.TypeLine).Color(arg, color.Instance())
	if err != nil {
		return abel.ColorResult(err)
	}
	link.SetColors(colors)
	link.ApplyStyle(arg.Lazy("ARROW_STYLE", 0))
	if stereotype, ok := arg.Get("STEREOTYPE", 0); ok {
		link.SetStereotype(klimt.StereotypeBuild(stereotype))
	}
	link.SetCodeLine(location)
	addWeightedLink(d, link, arg)
	return command.Ok()
}

// addWeightedLink adds a link with the weight of the "@weight" header.
func addWeightedLink(d *ClassDiagram, link *abel.Link, arg *regex.Result) {
	d.AddLink(link)
	if weight, ok := arg.Get("HEADER", 0); ok {
		link.SetWeight(uml.JavaParseDouble(weight))
	}
}

// executeLinkSpecial is the association class "(A, B) .. C" (couple 1) or
// "C .. (A, B)" (couple 2).
func executeLinkSpecial(location *text.LineLocation, d *ClassDiagram, arg *regex.Result,
	couple int) command.Result {
	coupleName, entName := "COUPLE1", "ENT2"
	if couple == 2 {
		coupleName, entName = "COUPLE2", "ENT1"
	}
	nameA := uml.RemoveQuotes(arg.Str(coupleName, 0))
	nameB := uml.RemoveQuotes(arg.Str(coupleName, 1))
	quarkA := d.QuarkInContext(true, nameA)
	quarkB := d.QuarkInContext(true, nameB)
	if quarkA.Data() == nil {
		return command.Error("No class " + nameA)
	}
	if quarkB.Data() == nil {
		return command.Error("No class " + nameB)
	}
	id := text.RemoveQuotesFormat(arg.Str(entName, 0), `"`)
	ent := d.QuarkInContext(true, id)
	cl := ent.Data()
	if cl == nil {
		cl = d.ReallyCreateLeaf(location, ent, klimt.DisplayWithNewlines(d.Pragma(), ent.Name()), abel.LeafTypeClass,
			abel.USymbolNull)
	}
	linkType := linkClassType(arg)
	label := klimt.DisplayWithNewlinesString(d.Pragma(), abel.Opt(arg.Get("LABEL_LINK", 0)).S,
		abel.Opt(arg.Get("LABEL_LINK", 0)).OK)
	if !d.AssociationClass(location, couple, quarkA.Data(), quarkB.Data(), cl, linkType, label) {
		return command.Error("Cannot have more than 2 assocications")
	}
	return command.Ok()
}

// executeLinkSpecial3 is the link between two association classes,
// "(A, B) .. (C, D)".
func executeLinkSpecial3(location *text.LineLocation, d *ClassDiagram, arg *regex.Result) command.Result {
	name1A := uml.RemoveQuotes(arg.Str("COUPLE1", 0))
	name1B := uml.RemoveQuotes(arg.Str("COUPLE1", 1))
	name2A := uml.RemoveQuotes(arg.Str("COUPLE2", 0))
	name2B := uml.RemoveQuotes(arg.Str("COUPLE2", 1))
	quark1A := d.QuarkInContext(true, name1A)
	quark1B := d.QuarkInContext(true, name1B)
	quark2A := d.QuarkInContext(true, name2A)
	quark2B := d.QuarkInContext(true, name2B)
	for _, q := range []struct {
		quark *abel.Quark
		name  string
	}{{quark1A, name1A}, {quark1B, name1B}, {quark2A, name2A}, {quark2B, name2B}} {
		if q.quark.Data() == nil {
			return command.Error("No class " + q.name)
		}
	}
	linkType := linkClassType(arg)
	label := klimt.DisplayWithNewlinesString(d.Pragma(), abel.Opt(arg.Get("LABEL_LINK", 0)).S,
		abel.Opt(arg.Get("LABEL_LINK", 0)).OK)
	return d.AssociationClass4(location, quark1A.Data(), quark1B.Data(), quark2A.Data(), quark2B.Data(), linkType,
		label)
}

// arrowBody2 is the second body of the arrow, in either alternative of the
// rewritten pattern (never null upstream).
func arrowBody2(arg *regex.Result) string {
	if s, ok := arg.Get("ARROW_BODY2B", 0); ok {
		return s
	}
	return arg.Str("ARROW_BODY2", 0)
}

// linkClassType is CommandLinkClass.getLinkType: the decorations of the
// heads (the first head is the decoration of the second end), dashed for
// a dotted body, and the middle circle.
func linkClassType(arg *regex.Result) decoration.LinkType {
	decors1 := decoration.LookupDecors1(arrowHead(arg, "ARROW_HEAD1"), true)
	decors2 := decoration.LookupDecors2(arrowHead(arg, "ARROW_HEAD2"), true)
	result := decoration.NewLinkType(decors2, decors1)
	if strings.Contains(arg.Str("ARROW_BODY1", 0), ".") || strings.Contains(arrowBody2(arg), ".") {
		result = result.GoDashed()
	}
	middle, _ := arg.Get("INSIDE", 0)
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

var (
	notQueueChars     = regexp.MustCompile(`[^-.=]`)
	notDirectionChars = regexp.MustCompile(`[^-.=a-zA-Z0-9_]`)
)

func linkQueueLength(arg *regex.Result) int {
	return len(notQueueChars.ReplaceAllString(fullArrow(arg), ""))
}

func linkDirection(arg *regex.Result) klimt.Direction {
	s := notDirectionChars.ReplaceAllString(fullArrow(arg), "")
	s = strings.TrimPrefix(s, "o")
	s = strings.TrimSuffix(s, "o")
	return queueDirection(s)
}

// queueDirection is upstream's StringUtils.getQueueDirection: the
// direction named in an arrow body, DOWN by default (RIGHT for a single
// character).
func queueDirection(s string) klimt.Direction {
	if strings.ContainsAny(s, "<>") {
		uml.Throw(uml.IllegalArgumentException, s)
	}
	s = strings.ToLower(s)
	switch {
	case strings.Contains(s, "left"):
		return klimt.DirectionLeft
	case strings.Contains(s, "right"):
		return klimt.DirectionRight
	case strings.Contains(s, "up"):
		return klimt.DirectionUp
	case strings.Contains(s, "down"):
		return klimt.DirectionDown
	case strings.Contains(s, "l"):
		return klimt.DirectionLeft
	case strings.Contains(s, "r"):
		return klimt.DirectionRight
	case strings.Contains(s, "u"):
		return klimt.DirectionUp
	case strings.Contains(s, "d"):
		return klimt.DirectionDown
	case uml.JavaLength(s) == 1:
		return klimt.DirectionRight
	}
	return klimt.DirectionDown
}

func arrowHead(arg *regex.Result, key string) string {
	return strings.ReplaceAll(arg.Str(key, 0), "_", "")
}

func fullArrow(arg *regex.Result) string {
	return arrowHead(arg, "ARROW_HEAD1") + arg.Str("ARROW_BODY1", 0) + arg.Str("ARROW_DIRECTION", 0) +
		arrowBody2(arg) + arrowHead(arg, "ARROW_HEAD2")
}

// labels is upstream's Labels: the cardinalities, roles and label of a
// link, the cardinalities possibly written in the label
// (`: "1" label "many"`).
type labels struct {
	firstLabel, secondLabel abel.OptString
	firstRole, secondRole   abel.OptString
	label                   abel.OptString
	linkArrow               abel.LinkArrow
}

var (
	labelsBoth       = uml.Compile(`^[%g]([^%g]+)[%g]([^%g]+)[%g]([^%g]+)[%g]$`)
	labelsFirstOnly  = uml.Compile(`^[%g]([^%g]+)[%g]([^%g]+)$`)
	labelsSecondOnly = uml.Compile(`^([^%g]+)[%g]([^%g]+)[%g]$`)
)

func cleanRole(role string, ok bool) abel.OptString {
	if !ok {
		return abel.OptString{}
	}
	return abel.Some(text.RemoveQuotesFormat(role, `"`))
}

func newLabels(arg *regex.Result) *labels {
	l := &labels{
		firstLabel:  abel.Opt(arg.Get("FIRST_LABEL", 0)),
		secondLabel: abel.Opt(arg.Get("SECOND_LABEL", 0)),
		firstRole:   cleanRole(arg.Get("FIRST_ROLE", 0)),
		secondRole:  cleanRole(arg.Get("SECOND_ROLE", 0)),
	}
	labelLink, ok := arg.Get("LABEL_LINK", 0)
	if ok {
		labelLink = l.init(labelLink)
	}
	l.label, l.linkArrow = stringWithArrow(labelLink, ok)
	return l
}

func (l *labels) init(labelLink string) string {
	if !l.firstLabel.OK && !l.secondLabel.OK {
		if m := labelsBoth.FindStringSubmatch(labelLink); m != nil {
			l.firstLabel = abel.Some(m[1])
			l.secondLabel = abel.Some(m[3])
			return text.Trin(uml.RemoveQuotes(text.Trin(m[2])))
		}
		if m := labelsFirstOnly.FindStringSubmatch(labelLink); m != nil {
			l.firstLabel = abel.Some(m[1])
			l.secondLabel = abel.OptString{}
			return text.Trin(uml.RemoveQuotes(text.Trin(m[2])))
		}
		if m := labelsSecondOnly.FindStringSubmatch(labelLink); m != nil {
			l.firstLabel = abel.OptString{}
			l.secondLabel = abel.Some(m[2])
			return text.Trin(uml.RemoveQuotes(text.Trin(m[1])))
		}
	}
	return text.RemoveQuotesFormat(labelLink, `"`)
}

func (l *labels) display(pragma *klimt.Pragma) *klimt.Display {
	return klimt.DisplayWithNewlinesString(pragma, l.label.S, l.label.OK)
}

// stringWithArrow is upstream's StringWithArrow: a label with a leading or
// trailing "<" or ">" gives the label arrow.
func stringWithArrow(completeLabel string, ok bool) (abel.OptString, abel.LinkArrow) {
	if !ok {
		return abel.OptString{}, abel.LinkArrowNoneOrSeveral
	}
	completeLabel = text.RemoveQuotesFormat(completeLabel, `"`)
	if klimt.DisplayHasSeveralGuideLines(completeLabel) {
		return abel.Some(completeLabel), abel.LinkArrowNoneOrSeveral
	}
	switch {
	case completeLabel == "<":
		return abel.OptString{}, abel.LinkArrowBackward
	case completeLabel == ">":
		return abel.OptString{}, abel.LinkArrowDirectNormal
	case strings.HasPrefix(completeLabel, "< "):
		return abel.Some(text.Trin(completeLabel[2:])), abel.LinkArrowBackward
	case strings.HasPrefix(completeLabel, "> "):
		return abel.Some(text.Trin(completeLabel[2:])), abel.LinkArrowDirectNormal
	case strings.HasSuffix(completeLabel, " >"):
		return abel.Some(text.Trin(completeLabel[:len(completeLabel)-2])), abel.LinkArrowDirectNormal
	case strings.HasSuffix(completeLabel, " <"):
		return abel.Some(text.Trin(completeLabel[:len(completeLabel)-2])), abel.LinkArrowBackward
	}
	return abel.Some(completeLabel), abel.LinkArrowNoneOrSeveral
}

// lollipopKeywords is upstream's CommandLinkLollipop.optionalKeywords for
// class diagrams.
const lollipopKeywords = "(interface|enum|annotation|abstract[%s]+class|abstract|class|entity|protocol|struct|" +
	"exception|metaclass|stereotype|dataclass|record)"

// commandLinkLollipop is a lollipop interface: "Foo ()- Bar" or
// "Bar -() Foo", a full ("()") or half ("((" or "))") circle linked to an
// existing class.
func commandLinkLollipop() cmd {
	ent := "(?:" + lollipopKeywords + "[%s]+)?" + classIdentifier + `[%s]*(\<\<.*\>\>)?`
	return uml.Single(regex.NewConcat(regex.Start(),
		regex.Optional(regex.NewConcat(
			uml.Leaf(1, "HEADER", `@([\d.]+)`),
			regex.SpaceOneOrMore())),
		uml.Leaf(3, "ENT1", ent),
		regex.SpaceZeroOrMore(),
		regex.Optional(uml.Leaf(1, "FIRST_LABEL", "[%g]([^%g]+)[%g]")),
		regex.SpaceZeroOrMore(),
		regex.Or("",
			uml.Leaf(2, "LOL_THEN_ENT", `([()]\))([-=.]+)`),
			uml.Leaf(2, "ENT_THEN_LOL", `([-=.]+)(\([()])`)),
		regex.SpaceZeroOrMore(),
		regex.Optional(uml.Leaf(1, "SECOND_LABEL", "[%g]([^%g]+)[%g]")),
		regex.SpaceZeroOrMore(),
		uml.Leaf(3, "ENT2", ent),
		regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewConcat(
			uml.Lit(":"),
			regex.SpaceZeroOrMore(),
			uml.Leaf(1, "LABEL_LINK", "(.+)"))),
		regex.End()),
		executeLinkLollipop)
}

var (
	lollipopP1 = uml.Compile(`^"([^"]+)"([^"]+)"([^"]+)"$`)
	lollipopP2 = uml.Compile(`^"([^"]+)"([^"]+)$`)
	lollipopP3 = uml.Compile(`^([^"]+)"([^"]+)"$`)
)

func lollipopType(desc string) abel.LeafType {
	if desc[0] == desc[1] {
		return abel.LeafTypeLollipopHalf
	}
	return abel.LeafTypeLollipopFull
}

func executeLinkLollipop(d *ClassDiagram, location *text.LineLocation, arg *regex.Result,
	_ command.ParserPass) command.Result {
	ent1 := arg.Str("ENT1", 1)
	ent2 := arg.Str("ENT2", 1)
	var cl1, cl2, normalEntity *abel.Entity
	suffix := d.UniqueSequence("lol")
	if _, ok := arg.Get("LOL_THEN_ENT", 1); !ok {
		quark := d.QuarkInContext(true, d.CleanID(ent1))
		cl1 = quark.Data()
		if cl1 == nil {
			return command.Error("No class " + quark.Name())
		}
		idNewLong := d.QuarkInContext(true, d.CleanID(ent1)+suffix)
		typ := lollipopType(arg.Str("ENT_THEN_LOL", 1))
		cl2 = d.ReallyCreateLeaf(location, idNewLong, klimt.DisplayWithNewlines(d.Pragma(), ent2), typ,
			abel.USymbolNull)
		normalEntity = cl1
	} else {
		quark := d.QuarkInContext(true, d.CleanID(ent2))
		cl2 = quark.Data()
		if cl2 == nil {
			return command.Error("No class " + quark.Name())
		}
		idNewLong := d.QuarkInContext(true, d.CleanID(ent2)+suffix)
		typ := lollipopType(arg.Str("LOL_THEN_ENT", 0))
		cl1 = d.ReallyCreateLeaf(location, idNewLong, klimt.DisplayWithNewlines(d.Pragma(), ent1), typ,
			abel.USymbolNull)
		normalEntity = cl2
	}
	linkType := decoration.NewLinkType(decoration.LinkDecorNone, decoration.LinkDecorNone)
	length := uml.JavaLength(lollipopQueue(arg))
	if length == 1 && d.NbOfHozizontalLollipop(normalEntity) > 1 {
		length++
	}
	firstLabel := abel.Opt(arg.Get("FIRST_LABEL", 0))
	secondLabel := abel.Opt(arg.Get("SECOND_LABEL", 0))
	labelLink := abel.OptString{}
	if s, ok := arg.Get("LABEL_LINK", 0); ok {
		labelLink = abel.Some(s)
		if !firstLabel.OK && !secondLabel.OK {
			if m := lollipopP1.FindStringSubmatch(s); m != nil {
				firstLabel = abel.Some(m[1])
				labelLink = abel.Some(text.Trin(uml.RemoveQuotes(text.Trin(m[2]))))
				secondLabel = abel.Some(m[3])
			} else if m := lollipopP2.FindStringSubmatch(s); m != nil {
				firstLabel = abel.Some(m[1])
				labelLink = abel.Some(text.Trin(uml.RemoveQuotes(text.Trin(m[2]))))
				secondLabel = abel.OptString{}
			} else if m := lollipopP3.FindStringSubmatch(s); m != nil {
				firstLabel = abel.OptString{}
				labelLink = abel.Some(text.Trin(uml.RemoveQuotes(text.Trin(m[1]))))
				secondLabel = abel.Some(m[2])
			}
		}
		labelLink = abel.Some(uml.RemoveQuotes(labelLink.S))
	}
	linkArg := abel.LinkArgBuildManage(klimt.DisplayWithNewlinesString(d.Pragma(), labelLink.S, labelLink.OK), length,
		d.SkinParam().ClassAttributeIconSize() > 0)
	link := abel.NewLink(location, &d.CucaDiagram, d.SkinParam().CurrentStyleBuilder(), cl1, cl2, linkType,
		linkArg.WithQuantifier(firstLabel, secondLabel).WithDistanceAngle(d.Labeldistance(), d.Labelangle()))
	d.ResetPragmaLabel()
	addWeightedLink(d, link, arg)
	return command.Ok()
}

func lollipopQueue(arg *regex.Result) string {
	if s, ok := arg.Get("LOL_THEN_ENT", 1); ok {
		return text.Trin(s)
	}
	if s, ok := arg.Get("ENT_THEN_LOL", 0); ok {
		return text.Trin(s)
	}
	uml.Throw(uml.IllegalArgumentException, "")
	return ""
}
