// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram/command/UBrexCommandIf.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram/command/UBrexCommandElse.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram/command/UBrexCommandEndif.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/descdiagram"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// ifPattern is the UBrex pattern of upstream's UBrexCommandIf: an optional
// source (the start, a bar, a quoted activity or a code), an optional
// arrow with a label in brackets, and the "if" with its label.
func ifPattern() klimt.UBrexPart {
	simple := klimt.UBrexConcat(
		klimt.UBrexSpaceZeroOrMore(),
		klimt.UBrexOptional(klimt.UBrexConcat(
			klimt.UBrexNamed("ARROW_BODY1", klimt.UBrexLeaf("〇+「-.」")),
			klimt.UBrexLeaf("〇?〘 [ 〶$ARROW_STYLE1=〘"+descdiagram.UBrexLineStyle+"〙] 〙"),
			klimt.UBrexNamed("ARROW_DIRECTION",
				klimt.UBrexLeaf("〇?【 *┇left┇right┇up┇down┇l〇?e┇r〇?i┇u〇?p┇d〇?o】")),
			klimt.UBrexLeaf("〇?〘 [  〶$ARROW_STYLE2=〘"+descdiagram.UBrexLineStyle+"〙] 〙"),
			klimt.UBrexNamed("ARROW_BODY2", klimt.UBrexLeaf("〇*「-.」")),
			klimt.UBrexLeaf(">"))),
		klimt.UBrexSpaceZeroOrMore(),
		klimt.UBrexOptional(klimt.UBrexLeaf("[ 〶$BRACKET=〘〇+「〤]*」  〇*「〤]」〙 ]")),
		klimt.UBrexSpaceZeroOrMore(),
		klimt.UBrexOr(
			klimt.UBrexLeaf("if 〇*〴s 〴g  〶$IF1=〇*〴G 〴g 〇*〴s 〇?〘as 〇+〴s 〶$ASIF1=〇+「〴an_.」 〇+〴s 〙〘then〙"),
			klimt.UBrexLeaf("if 〇+〴s 〶$IF2=〇l+〴. 〘then〙 "),
			klimt.UBrexLeaf("if 〇*〴s 〴g  〶$IF1=〇*〴G 〴g 〇*〴s 〇?〘as 〇+〴s 〶$ASIF1=〇+「〴an_.」 〇+〴s 〙"),
			klimt.UBrexLeaf("if 〇+〴s 〶$IF2=〇+〴.")))
	return klimt.UBrexConcat(
		klimt.UBrexOr(simple,
			klimt.UBrexConcat(klimt.UBrexNamed("STAR", klimt.UBrexLeaf("(* 〇?〘top〙)")), simple),
			klimt.UBrexConcat(klimt.UBrexNamed("BAR", klimt.UBrexLeaf("=〇+= 〇*〴s 〇+「〴an_.」〇*〴s =〇+=")), simple),
			klimt.UBrexConcat(
				klimt.UBrexLeaf("〴g 〶$QUOTED1=〇*〴G 〴g 〇?〘〇+〴s as 〇+〴s 〶$QUOTED2=〇+「〴an_.」 〙 "), simple),
			klimt.UBrexConcat(klimt.UBrexNamed("CODE", klimt.UBrexLeaf("〇+「〴an_.」")), simple)),
		klimt.UBrexEnd())
}

// commandIf is upstream's UBrexCommandIf: "if "label" then" opens a branch,
// linked from the source or the last activity.
func commandIf() cmd {
	return uml.UBrex(ifPattern(),
		func(d *ActivityDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			entity1 := ubrexIfSource(location, d, arg)
			if entity1 == nil {
				return command.Error("No if possible at this point")
			}
			var ifCode, ifLabel abel.OptString
			if ifLabel2, ok := arg.Get("IF2", 0); ok {
				ifLabel = abel.Some(ifLabel2)
			} else {
				ifCode = abel.Opt(arg.Get("ASIF1", 0))
				ifLabel = abel.Opt(arg.Get("IF1", 0))
			}
			d.startIf(location, ifCode.S, ifCode.OK)
			length := 2
			if arrowBody1, ok := arg.Get("ARROW_BODY1", 0); ok {
				arrow := uml.ManageArrowForCuca(arrowBody1 + arg.Str("ARROW_DIRECTION", 0) + arg.Str("ARROW_BODY2", 0) +
					">")
				length = uml.JavaLength(arrow) - 1
			}
			branch := d.currentContext.Branch()
			bracket, hasBracket := arg.Get("BRACKET", 0)
			linkArg := abel.LinkArgBuild(klimt.DisplayWithNewlinesString(d.Pragma(), bracket, hasBracket), length).
				WithQuantifier(abel.OptString{}, ifLabel).WithDistanceAngle(d.Labeldistance(), d.Labelangle())
			link := abel.NewLink(location, d.Cuca(), d.SkinParam().CurrentStyleBuilder(), entity1, branch,
				decoration.NewLinkType(decoration.LinkDecorArrow, decoration.LinkDecorNone), linkArg)
			// The pattern has no ARROW capture, so upstream never inverts the link.
			if arrow, ok := arg.Get("ARROW", 0); ok {
				if direction := uml.ArrowDirection(arrow); direction == klimt.DirectionLeft ||
					direction == klimt.DirectionUp {
					link = link.Inv()
				}
			}
			link.ApplyStyle(arg.Lazy("ARROW_STYLE", 0))
			d.AddLink(link)
			return command.Ok()
		})
}

// commandElse is upstream's UBrexCommandElse: the next transitions start
// from the branch of the current "if".
func commandElse() cmd {
	return uml.UBrexTokens([]string{"else"}, klimt.UBrexConcat(klimt.UBrexLeaf("else"), klimt.UBrexEnd()),
		func(d *ActivityDiagram, _ *text.LineLocation, _ *regex.Result, _ command.ParserPass) command.Result {
			if d.lastEntity() == nil || d.currentContext == nil {
				return command.Error("No if for this else")
			}
			d.setLastEntity(d.currentContext.Branch())
			return command.Ok()
		})
}

// commandEndif is upstream's UBrexCommandEndif: "endif" closes the current
// "if".
func commandEndif() cmd {
	return uml.UBrexTokens([]string{"end", "endif"}, klimt.UBrexConcat(klimt.UBrexLeaf("end"), klimt.UBrexSpaceZeroOrMore(), klimt.UBrexLeaf("if"),
		klimt.UBrexEnd()),
		func(d *ActivityDiagram, _ *text.LineLocation, _ *regex.Result, _ command.ParserPass) command.Result {
			if d.lastEntity() == nil || d.currentContext == nil {
				return command.Error("No if for this endif")
			}
			d.endif()
			return command.Ok()
		})
}
