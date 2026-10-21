// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram/command/CommandLinkActivity.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram/command/CommandLinkLongActivity.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/descdiagram"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

type cmd = command.Command[*ActivityDiagram]

// linkSource is the optional source of a transition: the start, an
// activity code, a bar or a quoted activity, captured as FIRST.
func linkSource() regex.Part {
	return regex.Optional(regex.Or("FIRST",
		uml.Leaf(2, "STAR", `(\(\*(top)?\))`),
		uml.Leaf(1, "CODE", "([%pLN][%pLN_.]*)"),
		uml.Leaf(1, "BAR", "(?:==+)[%s]*([%pLN_.]+)[%s]*(?:==+)"),
		uml.Leaf(2, "QUOTED", "[%g]([^%g]+)[%g](?:[%s]+as[%s]+([%pLN_.]+))?")))
}

// linkArrow is the arrow of a transition and its optional label.
func linkArrow() []regex.Part {
	return []regex.Part{
		uml.Leaf(1, "ARROW_BODY1", "([-.]+)"),
		uml.Leaf(1, "ARROW_STYLE1", `(?:\[(`+descdiagram.LineStyle+`)\])?`),
		uml.Leaf(1, "ARROW_DIRECTION", `(\*|left|right|up|down|le?|ri?|up?|do?)?`),
		uml.Leaf(1, "ARROW_STYLE2", `(?:\[(`+descdiagram.LineStyle+`)\])?`),
		uml.Leaf(1, "ARROW_BODY2", "([-.]*)"),
		uml.Lit(`\>`),
		regex.SpaceZeroOrMore(),
		regex.Optional(uml.Leaf(1, "BRACKET", `\[([^\]*]+[^\]]*)\]`)),
		regex.SpaceZeroOrMore(),
	}
}

// commandLinkActivity is upstream's CommandLinkActivity: a transition from
// an activity, or from the last one, to an activity, a bar, an end, or an
// inner activity ("{").
func commandLinkActivity() cmd {
	parts := []regex.Part{regex.Start(),
		linkSource(),
		klimt.StereotypePatternOptional("STEREOTYPE"),
		color.Exp2(),
		regex.SpaceZeroOrMore(),
		klimt.URLBuilderOptional(),
	}
	parts = append(parts, linkArrow()...)
	parts = append(parts,
		regex.Or("FIRST2",
			uml.Leaf(2, "STAR2", `(\(\*(top|\d+)?\))`),
			uml.Leaf(1, "OPENBRACKET2", `(\{)`),
			uml.Leaf(1, "CODE2", "([%pLN][%pLN_.]*)"),
			uml.Leaf(1, "BAR2", "(?:==+)[%s]*([%pLN_.]+)[%s]*(?:==+)"),
			uml.Leaf(2, "QUOTED2", "[%g]([^%g]+)[%g](?:[%s]+as[%s]+([%pLN][%pLN_.]*))?"),
			uml.Leaf(1, "QUOTED_INVISIBLE2", `(\w.*?)`)),
		klimt.StereotypePatternOptional("STEREOTYPE2"),
		regex.Optional(regex.NewConcat(
			uml.Lit("in"),
			regex.SpaceOneOrMore(),
			uml.Leaf(1, "PARTITION2", `([%g][^%g]+[%g]|\S+)`))),
		regex.SpaceZeroOrMore(),
		color.Exp3(),
		regex.End())
	return uml.Single(regex.NewConcat(parts...), executeLinkActivity)
}

func executeLinkActivity(d *ActivityDiagram, location *text.LineLocation, arg *regex.Result,
	_ command.ParserPass) command.Result {
	entity1 := linkEntity(location, d, arg, true)
	if entity1 == nil {
		return command.Error("No such activity")
	}
	if stereotype, ok := arg.Get("STEREOTYPE", 0); ok {
		entity1.SetStereotype(klimt.StereotypeBuild(stereotype))
	}
	if s, ok := arg.Get("BACKCOLOR", 0); ok {
		c, err := d.Skin().IHtmlColorSet().GetColor(s)
		if err != nil {
			return abel.ColorResult(err)
		}
		entity1.SetSpecificColorTOBEREMOVED(color.TypeBack, c)
	}
	entity2 := linkEntity(location, d, arg, false)
	if entity2 == nil {
		return command.Error("No such activity")
	}
	d.setLastEntity(entity2)
	if s, ok := arg.Get("BACKCOLOR2", 0); ok {
		c, err := d.Skin().IHtmlColorSet().GetColor(s)
		if err != nil {
			return abel.ColorResult(err)
		}
		entity2.SetSpecificColorTOBEREMOVED(color.TypeBack, c)
	}
	if stereotype, ok := arg.Get("STEREOTYPE2", 0); ok {
		entity2.SetStereotype(klimt.StereotypeBuild(stereotype))
	}
	linkLabel := klimt.DisplayWithNewlinesString(d.Pragma(), arg.Str("BRACKET", 0), hasGroup(arg, "BRACKET"))
	arrowBody1, arrowBody2 := arg.Str("ARROW_BODY1", 0), arg.Str("ARROW_BODY2", 0)
	arrowDirection := arg.Str("ARROW_DIRECTION", 0)
	arrow := uml.ManageArrowForCuca(arrowBody1 + arrowDirection + arrowBody2 + ">")
	length := uml.JavaLength(arrow) - 1
	if strings.Contains(arrowDirection, "*") {
		length = 2
	}
	typ := decoration.NewLinkType(decoration.LinkDecorArrow, decoration.LinkDecorNone)
	if strings.Contains(arrowBody1+arrowBody2, ".") {
		typ = typ.GoDotted()
	}
	linkArg := abel.LinkArgBuildManage(linkLabel, length, d.SkinParam().ClassAttributeIconSize() > 0)
	link := abel.NewLink(location, d.Cuca(), d.SkinParam().CurrentStyleBuilder(), entity1, entity2, typ, linkArg)
	if strings.Contains(arrowDirection, "*") {
		link.SetConstraint(false)
	}
	direction := uml.ArrowDirection(arrowBody1 + arrowDirection + arrowBody2 + ">")
	if direction == klimt.DirectionLeft || direction == klimt.DirectionUp {
		link = link.Inv()
	}
	if url, ok := arg.Get(klimt.URLBuilderKey, 0); ok {
		link.SetURL(abel.URLBuilderOf(d.Cuca()).URL(url))
	}
	link.ApplyStyle(arg.Lazy("ARROW_STYLE", 0))
	d.AddLink(link)
	return command.Ok()
}

func hasGroup(arg *regex.Result, name string) bool {
	_, ok := arg.Get(name, 0)
	return ok
}

// inPartition opens the partition of partitionCode, if any, as a package
// with display.
func inPartition(location *text.LineLocation, d *ActivityDiagram, partition string, hasPartition bool,
	display func(quark *abel.Quark) *klimt.Display) {
	if !hasPartition {
		return
	}
	quark := d.QuarkInContext(true, d.CleanID(partition))
	d.GotoGroup(location, quark, display(quark), abel.GroupTypePackage, abel.USymbolNull)
}

// quarkDisplay is upstream's Display.getWithNewlines(quark): the name of the
// quark, read without the pragmas of the diagram.
func quarkDisplay(quark *abel.Quark) *klimt.Display {
	return klimt.DisplayWithNewlines(klimt.PragmaCreateEmpty(), quark.Name())
}

// linkEntity is upstream's CommandLinkActivity.getEntity: the source
// (start) or the target of a transition, created when needed.
func linkEntity(location *text.LineLocation, d *ActivityDiagram, arg *regex.Result, start bool) *abel.Entity {
	suf := "2"
	if start {
		suf = ""
	}
	if _, ok := arg.Get("OPENBRACKET"+suf, 0); ok {
		return d.createInnerActivity(location)
	}
	if _, ok := arg.Get("STAR"+suf, 0); ok {
		if start {
			return d.getStart(location)
		}
		suppID, hasSuppID := arg.Get("STAR"+suf, 1)
		return d.getEnd(location, suppID, hasSuppID)
	}
	partition, hasPartition := arg.Get("PARTITION"+suf, 0)
	if hasPartition {
		partition = text.RemoveQuotes(partition)
	}
	if idShort, ok := arg.Get("CODE"+suf, 0); ok {
		inPartition(location, d, partition, hasPartition, quarkDisplay)
		ident := d.QuarkInContext(true, d.CleanID(idShort))
		result := ident.Data()
		if result == nil {
			result = d.ReallyCreateLeaf(location, ident, klimt.DisplayWithNewlines(d.Pragma(), idShort),
				abel.LeafTypeActivity, abel.USymbolNull)
		}
		if hasPartition {
			d.EndGroup()
		}
		return result
	}
	if bar, ok := arg.Get("BAR"+suf, 0); ok {
		quark := d.QuarkInContext(true, d.CleanID(bar))
		result := quark.Data()
		if result == nil {
			result = d.ReallyCreateLeaf(location, quark, klimt.DisplayWithNewlines(d.Pragma(), bar),
				abel.LeafTypeSynchroBar, abel.USymbolNull)
		}
		return result
	}
	if quoted, ok := arg.Get("QUOTED"+suf, 0); ok {
		quotedString := quoted
		if code, hasCode := arg.Get("QUOTED"+suf, 1); hasCode {
			quotedString = code
		}
		inPartition(location, d, partition, hasPartition, func(*abel.Quark) *klimt.Display {
			return klimt.DisplayWithNewlines(d.Pragma(), partition)
		})
		quark := d.QuarkInContext(true, d.CleanID(quotedString))
		result := quark.Data()
		if result == nil {
			result = d.ReallyCreateLeaf(location, quark, klimt.DisplayWithNewlines(d.Pragma(), quoted),
				abel.LeafTypeActivity, abel.USymbolNull)
		}
		if hasPartition {
			d.EndGroup()
		}
		return result
	}
	if invisible, ok := arg.Get("QUOTED_INVISIBLE"+suf, 0); ok {
		inPartition(location, d, partition, hasPartition, quarkDisplay)
		identInvisible := d.QuarkInContext(true, d.CleanID(invisible))
		result := identInvisible.Data()
		if result == nil {
			result = d.ReallyCreateLeaf(location, identInvisible,
				klimt.DisplayWithNewlines(d.Pragma(), identInvisible.Name()), abel.LeafTypeActivity, abel.USymbolNull)
		}
		if hasPartition {
			d.EndGroup()
		}
		return result
	}
	if _, ok := arg.Get("FIRST"+suf, 0); !ok {
		return d.lastEntity()
	}
	return nil
}

// ubrexIfSource is upstream's CommandLinkActivity.ubrexGetEntityForIfOnly:
// the activity an "if" starts from.
func ubrexIfSource(location *text.LineLocation, d *ActivityDiagram, arg *regex.Result) *abel.Entity {
	if _, ok := arg.Get("STAR", 0); ok {
		return d.getStart(location)
	}
	if idShort, ok := arg.Get("CODE", 0); ok {
		ident := d.QuarkInContext(true, d.CleanID(idShort))
		result := ident.Data()
		if result == nil {
			result = d.ReallyCreateLeaf(location, ident, klimt.DisplayWithNewlines(d.Pragma(), idShort),
				abel.LeafTypeActivity, abel.USymbolNull)
		}
		return result
	}
	if bar, ok := arg.Get("BAR", 0); ok {
		quark := d.QuarkInContext(true, d.CleanID(bar))
		result := quark.Data()
		if result == nil {
			result = d.ReallyCreateLeaf(location, quark, klimt.DisplayWithNewlines(d.Pragma(), bar),
				abel.LeafTypeSynchroBar, abel.USymbolNull)
		}
		return result
	}
	if quoted1, ok := arg.Get("QUOTED1", 0); ok {
		quotedString := quoted1
		if quoted2, hasCode := arg.Get("QUOTED2", 0); hasCode {
			quotedString = quoted2
		}
		quark := d.QuarkInContext(true, d.CleanID(quotedString))
		result := quark.Data()
		if result == nil {
			result = d.ReallyCreateLeaf(location, quark, klimt.DisplayWithNewlines(d.Pragma(), quoted1),
				abel.LeafTypeActivity, abel.USymbolNull)
		}
		return result
	}
	return d.lastEntity()
}

// longActivityEnd is the last line of CommandLinkLongActivity: the end of
// the text, then an optional code, stereotype, partition and color.
var longActivityEnd = uml.Compile(`^[%s]*([^%g]*)[%g](?:[%s]+as[%s]+([%pLN][%pLN_.]*))?[%s]*(\<\<.*\>\>)?[%s]*(?:in[%s]+([%g][^%g]+[%g]|\S+))?[%s]*(#\w+)?$`)

// commandLinkLongActivity is upstream's CommandLinkLongActivity: a
// transition to an activity whose quoted text spans several lines.
func commandLinkLongActivity() cmd {
	parts := []regex.Part{regex.Start(),
		linkSource(),
		klimt.StereotypePatternOptional("STEREOTYPE"),
		uml.Leaf(1, "BACKCOLOR", `(#\w+)?`),
		regex.SpaceZeroOrMore(),
		klimt.URLBuilderOptional(),
	}
	parts = append(parts, linkArrow()...)
	parts = append(parts,
		uml.Lit("[%g]"),
		uml.Leaf(1, "DESC", "([^%g]*?)"),
		regex.SpaceZeroOrMore(),
		regex.End())
	c := &uml.Multilines2[*ActivityDiagram]{
		Start:    regex.NewConcat(parts...),
		Strategy: uml.RemoveStartingQuote,
		TrimEnd:  uml.TrimBoth,
		End:      longActivityEnd,
	}
	c.Exec = func(d *ActivityDiagram, lines command.BlocLines, _ command.ParserPass) command.Result {
		return executeLinkLongActivity(c, d, lines)
	}
	return c
}

// split is upstream's StringUtils.getSplit: the groups of the first match
// of re in line (ok false for those that do not participate), nil when it
// does not match.
func split(re interface {
	FindStringSubmatchIndex(s string) []int
}, line string) []regexGroup {
	loc := re.FindStringSubmatchIndex(line)
	if loc == nil {
		return nil
	}
	var result []regexGroup
	for i := 2; i < len(loc); i += 2 {
		if loc[i] < 0 {
			result = append(result, regexGroup{})
		} else {
			result = append(result, regexGroup{line[loc[i]:loc[i+1]], true})
		}
	}
	return result
}

type regexGroup struct {
	s  string
	ok bool
}

func executeLinkLongActivity(c *uml.Multilines2[*ActivityDiagram], d *ActivityDiagram,
	lines command.BlocLines) command.Result {
	lines = lines.Trim()
	location := lines.At(0).Location
	line0 := c.StartArg(lines)
	entity1 := linkEntity(location, d, line0, true)
	if entity1 == nil {
		return command.Error("No such entity")
	}
	if stereotype, ok := line0.Get("STEREOTYPE", 0); ok {
		entity1.SetStereotype(klimt.StereotypeBuild(stereotype))
	}
	if s, ok := line0.Get("BACKCOLOR", 0); ok {
		c, err := d.Skin().IHtmlColorSet().GetColor(s)
		if err != nil {
			return abel.ColorResult(err)
		}
		entity1.SetSpecificColorTOBEREMOVED(color.TypeBack, c)
	}
	newline := string(klimt.JawsBlockE1Newline)
	var sb strings.Builder
	urlBuilder := abel.URLBuilderOf(d.Cuca())
	desc0, hasDesc0 := line0.Get("DESC", 0)
	var urlActivity *klimt.URL
	if !uml.IsEmpty(desc0, hasDesc0) {
		urlActivity = urlBuilder.URL(desc0)
		if urlActivity == nil {
			sb.WriteString(desc0)
			sb.WriteString(newline)
		}
	}
	middle := lines.Lines()[1 : lines.Size()-1]
	for i, cs := range middle {
		if i == 0 && urlActivity == nil {
			urlActivity = urlBuilder.URL(cs.S)
			if urlActivity != nil {
				continue
			}
		}
		sb.WriteString(cs.S)
		if i+1 < lines.Size()-2 {
			sb.WriteString(newline)
		}
	}
	lineLast := split(longActivityEnd, lines.Last().S)
	if !uml.IsEmpty(lineLast[0].s, lineLast[0].ok) {
		if sb.Len() > 0 && !strings.HasSuffix(sb.String(), newline) {
			sb.WriteString(newline)
		}
		sb.WriteString(lineLast[0].s)
	}
	displayString := sb.String()
	idShort := displayString
	if lineLast[1].ok {
		idShort = lineLast[1].s
	}
	partition, hasPartition := lineLast[3].s, lineLast[3].ok
	if hasPartition {
		partition = text.RemoveQuotes(partition)
		quark := d.QuarkInContext(true, d.CleanID(partition))
		d.GotoGroup(location, quark, klimt.DisplayWithNewlines(d.Pragma(), partition), abel.GroupTypePackage,
			abel.USymbolNull)
	}
	ident := d.QuarkInContext(true, d.CleanID(idShort))
	entity2 := ident.Data()
	if entity2 == nil {
		entity2 = d.ReallyCreateLeaf(location, ident, klimt.DisplayWithNewlines(d.Pragma(), displayString),
			abel.LeafTypeActivity, abel.USymbolNull)
	}
	d.setLastEntity(entity2)
	if hasPartition {
		d.EndGroup()
	}
	if urlActivity != nil {
		entity2.AddURL(urlActivity)
	}
	if lineLast[2].ok {
		entity2.SetStereotype(klimt.StereotypeBuild(lineLast[2].s))
	}
	if lineLast[4].ok {
		c, err := d.Skin().IHtmlColorSet().GetColor(lineLast[4].s)
		if err != nil {
			return abel.ColorResult(err)
		}
		entity2.SetSpecificColorTOBEREMOVED(color.TypeBack, c)
	}
	arrowBody1, arrowBody2 := line0.Str("ARROW_BODY1", 0), line0.Str("ARROW_BODY2", 0)
	arrowDirection := line0.Str("ARROW_DIRECTION", 0)
	arrow := uml.ManageArrowForCuca(arrowBody1 + arrowDirection + arrowBody2 + ">")
	length := uml.JavaLength(arrow) - 1
	linkLabel := klimt.DisplayWithNewlinesString(d.Pragma(), line0.Str("BRACKET", 0), hasGroup(line0, "BRACKET"))
	typ := decoration.NewLinkType(decoration.LinkDecorArrow, decoration.LinkDecorNone)
	if strings.Contains(arrow, ".") {
		typ = typ.GoDotted()
	}
	linkArg := abel.LinkArgBuildManage(linkLabel, length, d.SkinParam().ClassAttributeIconSize() > 0)
	link := abel.NewLink(location, d.Cuca(), d.SkinParam().CurrentStyleBuilder(), entity1, entity2, typ, linkArg)
	direction := uml.ArrowDirection(arrowBody1 + arrowDirection + arrowBody2 + ">")
	if direction == klimt.DirectionLeft || direction == klimt.DirectionUp {
		link = link.Inv()
	}
	if url, ok := line0.Get(klimt.URLBuilderKey, 0); ok {
		link.SetURL(urlBuilder.URL(url))
	}
	link.ApplyStyle(line0.Lazy("ARROW_STYLE", 0))
	d.AddLink(link)
	return command.Ok()
}
