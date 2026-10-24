// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/statediagram/command/CommandLinkStateCommon.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/statediagram/command/CommandLinkState.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/statediagram/command/CommandLinkStateReverse.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package statediagram

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/descdiagram"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// statePattern is upstream's getStatePattern: a state, a history state,
// the start or end "[*]", or a synchronization bar "==name==".
func statePattern(name string) *regex.Leaf {
	return uml.Leaf(1, name, `([%pLN_.:]+|[%pLN_.:]+\[H\*?\]|\[\*\]|\[H\*?\]|(?:==+)(?:[%pLN_.:]+)(?:==+))`)
}

// linkLabel is the optional ": label" of a transition.
func linkLabel() regex.Part {
	return regex.Optional(regex.NewConcat(
		uml.Lit(":"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "LABEL", "(.+)")))
}

const arrowDirection = "(left|right|up|down|le?|ri?|up?|do?)?"

// arrowStyle is "[style]" in an arrow.
func arrowStyle(name string) *regex.Leaf {
	return uml.Leaf(1, name, `(?:\[(`+descdiagram.LineStyle+`)\])?`)
}

// commandLinkState is "A -> B : label", upstream's CommandLinkState.
func commandLinkState() cmd {
	return linkStateCommand(regex.NewConcat(regex.Start(),
		statePattern("ENT1"),
		regex.SpaceZeroOrMore(),
		regex.NewConcat(
			uml.Leaf(1, "ARROW_CROSS_START", "(x)?"),
			uml.Leaf(1, "ARROW_BODY1", "(-+)"),
			arrowStyle("ARROW_STYLE1"),
			uml.Leaf(1, "ARROW_DIRECTION", arrowDirection),
			arrowStyle("ARROW_STYLE2"),
			uml.Leaf(1, "ARROW_BODY2", "(-*)"),
			uml.Lit(`\>`),
			uml.Leaf(1, "ARROW_CIRCLE_END", "(o[%s]+)?")),
		regex.SpaceZeroOrMore(),
		statePattern("ENT2"),
		regex.SpaceZeroOrMore(),
		klimt.StereotypePatternOptional("STEREOTYPE"),
		linkLabel(),
		regex.End()), false)
}

// commandLinkStateReverse is "B <- A : label", upstream's
// CommandLinkStateReverse.
func commandLinkStateReverse() cmd {
	return linkStateCommand(regex.NewConcat(regex.Start(),
		statePattern("ENT2"),
		regex.SpaceZeroOrMore(),
		regex.NewConcat(
			uml.Leaf(1, "ARROW_CIRCLE_END", "(o[%s]+)?"),
			uml.Lit(`\<`),
			uml.Leaf(1, "ARROW_BODY2", "(-*)"),
			arrowStyle("ARROW_STYLE2"),
			uml.Leaf(1, "ARROW_DIRECTION", arrowDirection),
			arrowStyle("ARROW_STYLE1"),
			uml.Leaf(1, "ARROW_BODY1", "(-+)"),
			uml.Leaf(1, "ARROW_CROSS_START", "(x)?")),
		regex.SpaceZeroOrMore(),
		statePattern("ENT1"),
		regex.SpaceZeroOrMore(),
		klimt.StereotypePatternOptional("STEREOTYPE"),
		linkLabel(),
		regex.End()), true)
}

// linkStateCommand is upstream's CommandLinkStateCommon, which runs in the
// second pass, once the states are declared. A reversed transition goes
// to the left when its arrow names no direction.
func linkStateCommand(pattern *regex.Concat, reversed bool) cmd {
	c := uml.Single(pattern,
		func(d *StateDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			return executeLinkState(d, location, arg, reversed)
		})
	c.Passes = []command.ParserPass{command.Two}
	return c
}

func executeLinkState(d *StateDiagram, location *text.LineLocation, arg *regex.Result, reversed bool) command.Result {
	ent1, _ := arg.Get("ENT1", 0)
	ent2, _ := arg.Get("ENT2", 0)
	cl1 := entityStart(location, d, ent1)
	if cl1 == nil {
		return command.Error("The state " + ent1 + " cannot be used here.")
	}
	cl2 := entityEnd(location, d, ent2)
	if cl2 == nil {
		return command.Error("The state " + ent2 + " cannot be used here.")
	}
	body1, _ := arg.Get("ARROW_BODY1", 0)
	body2, _ := arg.Get("ARROW_BODY2", 0)
	queue := body1 + body2
	dir, hasDir := linkDirection(arg, reversed)
	if hasDir && (dir == klimt.DirectionLeft || dir == klimt.DirectionRight) {
		queue = "-"
	}
	length := uml.JavaLength(queue)
	_, crossStart := arg.Get("ARROW_CROSS_START", 0)
	_, circleEnd := arg.Get("ARROW_CIRCLE_END", 0)
	decor1, decor2 := decoration.LinkDecorArrow, decoration.LinkDecorNone
	if circleEnd {
		decor1 = decoration.LinkDecorArrowAndCircle
	}
	if crossStart {
		decor2 = decoration.LinkDecorCircleCross
	}
	linkType := decoration.NewLinkType(decor1, decor2)
	labelString, hasLabel := arg.Get("LABEL", 0)
	label := klimt.DisplayWithNewlinesString(d.Pragma(), labelString, hasLabel)
	inv := hasDir && (dir == klimt.DirectionLeft || dir == klimt.DirectionUp)
	if useNodeStyle(d, arg) && !klimt.DisplayIsNull(label) && strings.TrimSpace(label.String()) != "" {
		createTransitionWithIntermediateNode(d, location, cl1, cl2, label, linkType, inv, arg)
		return command.Ok()
	}
	linkArg := abel.LinkArgBuildManage(label, length, d.SkinParam().ClassAttributeIconSize() > 0)
	link := abel.NewLink(location, d.Cuca(), d.SkinParam().CurrentStyleBuilder(), cl1, cl2, linkType, linkArg)
	if inv {
		link = link.Inv()
	}
	link.ApplyStyle(arg.Lazy("ARROW_STYLE", 0))
	if stereotype, ok := arg.Get("STEREOTYPE", 0); ok {
		link.SetStereotype(klimt.StereotypeBuild(stereotype))
	}
	d.AddLink(link)
	return command.Ok()
}

// linkDirection is upstream's getDirection: the direction written in the
// arrow, else LEFT for a reversed transition and none otherwise.
func linkDirection(arg *regex.Result, reversed bool) (klimt.Direction, bool) {
	if dir, ok := arg.Get("ARROW_DIRECTION", 0); ok {
		return uml.QueueDirection(dir), true
	}
	if reversed {
		return klimt.DirectionLeft, true
	}
	return klimt.DirectionDown, false
}

// entityStart is upstream's getEntityStart: "[*]" is the start of the
// current group.
func entityStart(location *text.LineLocation, d *StateDiagram, code string) *abel.Entity {
	if strings.HasPrefix(code, "[*]") {
		return d.getStart(location)
	}
	return linkEntity(location, d, code)
}

// entityEnd is upstream's getEntityEnd: "[*]" is the end of the current
// group.
func entityEnd(location *text.LineLocation, d *StateDiagram, code string) *abel.Entity {
	if strings.HasPrefix(code, "[*]") {
		return d.getEnd(location)
	}
	return linkEntity(location, d, code)
}

// linkEntity is upstream's getEntity: a history state, a synchronization
// bar, the current group, or a state, created when its parent exists. It
// returns nil for a state that cannot be used in the current group.
func linkEntity(location *text.LineLocation, d *StateDiagram, code string) *abel.Entity {
	switch {
	case strings.EqualFold(code, "[H]"):
		return d.getHistorical(location)
	case strings.HasSuffix(code, "[H]"):
		return d.getHistoricalOf(location, code[:len(code)-3])
	case strings.EqualFold(code, "[H*]"):
		return d.getDeepHistory(location)
	case strings.HasSuffix(code, "[H*]"):
		return d.getDeepHistoryOf(location, code[:len(code)-4])
	case strings.HasPrefix(code, "=") && strings.HasSuffix(code, "="):
		quark := d.QuarkInContext(true, d.CleanID(strings.Trim(code, "=")))
		if quark.Data() != nil {
			return quark.Data()
		}
		return d.ReallyCreateLeaf(location, quark, quarkDisplay(quark), abel.LeafTypeSynchroBar, abel.USymbolNull)
	case d.CurrentGroup().Name() == code:
		return d.CurrentGroup()
	}
	quark := d.QuarkInContext(true, d.CleanID(code))
	if !d.checkConcurrentStateOk(quark) {
		return nil
	}
	if quark.Data() != nil {
		return quark.Data()
	}
	if quark.Parent().Data() == nil {
		return nil
	}
	return d.ReallyCreateLeaf(location, quark, klimt.DisplayWithNewlines(d.Pragma(), quark.Name()),
		abel.LeafTypeState, abel.USymbolNull)
}

// useNodeStyle is upstream's shouldUseNodeStyle: "node" in the arrow
// style, or "skinparam stateDiagramEdgeLabelStyle node", draws the label
// as a node between the two states.
func useNodeStyle(d *StateDiagram, arg *regex.Result) bool {
	if style, ok := arg.Lazy("ARROW_STYLE", 0); ok && strings.Contains(strings.ToLower(style), "node") {
		return true
	}
	setting, ok := d.Skin().Value("statediagramedgelabelstyle")
	return ok && strings.EqualFold(setting, "node")
}

var (
	nodeStyleWord  = regexp.MustCompile(`(?i),?node,?`)
	nodeStyleEnds  = regexp.MustCompile(`^,|,$`)
	nonIdentifiers = regexp.MustCompile(`[^a-zA-Z0-9_]`)
)

// createTransitionWithIntermediateNode is upstream's method of the same
// name: the label becomes a node, linked from source and to target.
func createTransitionWithIntermediateNode(d *StateDiagram, location *text.LineLocation, source, target *abel.Entity,
	label *klimt.Display, linkType decoration.LinkType, inv bool, arg *regex.Result) {
	quark := d.QuarkInContext(true, d.CleanID(d.transitionNodeID(source, target, label)))
	node := d.ReallyCreateLeaf(location, quark, label, abel.LeafTypeStateTransitionLabel, abel.USymbolNull)
	node.SetStereotype(klimt.StereotypeBuild("<<transition>>"))
	iconSize := d.SkinParam().ClassAttributeIconSize() > 0
	styleBuilder := d.SkinParam().CurrentStyleBuilder()
	first := abel.NewLink(location, d.Cuca(), styleBuilder, source, node,
		decoration.NewLinkType(decoration.LinkDecorNone, decoration.LinkDecorNone),
		abel.LinkArgBuildManage(klimt.DisplayNULL(), 2, iconSize))
	first.SetWeight(10)
	second := abel.NewLink(location, d.Cuca(), styleBuilder, node, target, linkType,
		abel.LinkArgBuildManage(klimt.DisplayNULL(), 2, iconSize))
	second.SetWeight(10)
	if inv {
		first = first.Inv()
		second = second.Inv()
	}
	if style, ok := arg.Lazy("ARROW_STYLE", 0); ok {
		filtered := strings.ReplaceAll(nodeStyleWord.ReplaceAllString(style, ""), ",,", ",")
		filtered = nodeStyleEnds.ReplaceAllString(filtered, "")
		if filtered != "" {
			first.ApplyStyle(filtered, true)
			second.ApplyStyle(filtered, true)
		}
	}
	d.AddLink(first)
	d.AddLink(second)
}

// transitionNodeID is upstream's generateTransitionNodeId, with a sequence
// number of the diagram instead of the current time, so that the output
// is the same on every render.
func (d *StateDiagram) transitionNodeID(source, target *abel.Entity, label *klimt.Display) string {
	d.transitionNodes++
	return "transition_" + source.Name() + "_" + target.Name() + "_" +
		nonIdentifiers.ReplaceAllString(label.String(), "_") + "_" + strconv.Itoa(d.transitionNodes)
}
