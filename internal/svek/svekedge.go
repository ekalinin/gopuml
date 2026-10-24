// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/SvekEdge.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cruise/XAbstractEdge.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cruise/XEdge.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"math"
	"strconv"
	"strings"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/dot"
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/klimt/rose"
	"github.com/ekalinin/gopuml/internal/uml"
)

// constraintSpot is the size of the label of a link constraint.
var constraintSpot = geom.NewXDimension2D(10, 10)

// SvekEdge is a link as a dot edge: its decorations, its label, its
// quantifiers and the colors that identify them in dot's output, and the
// path dot computed for it.
type SvekEdge struct {
	link        *abel.Link
	skinParam   klimt.ISkinParam
	bibliotekon *Bibliotekon

	ltail, lhead *Cluster

	startUID, endUID entityPort

	startTailText     klimt.TextBlock
	endHeadText       klimt.TextBlock
	startTailRoleText klimt.TextBlock
	endHeadRoleText   klimt.TextBlock
	labelText         klimt.TextBlock

	divideLabelWidthByTwo bool

	lineColor      int32
	noteLabelColor int32
	startTailColor int32
	endHeadColor   int32

	stringBounder font.StringBounder

	dotPath     *klimt.DotPath
	dotPathInit *klimt.DotPath

	startTailLabelXY geom.Positionable
	endHeadLabelXY   geom.Positionable
	labelXY          geom.Positionable

	extremity1, extremity2 klimt.UDrawable

	dx, dy float64

	opale             bool
	projectionCluster *Cluster

	pragma             *klimt.Pragma
	backgroundColor    color.HColor
	useRankSame        bool
	defaultThickness   *klimt.UStroke
	arrowLollipopColor color.HColor

	labelShield float64

	kal1, kal2 *Kal

	ids map[string]bool
}

var _ klimt.GuideLine = (*SvekEdge)(nil)

func (e *SvekEdge) diagramType() core.DiagramType { return e.skinParam.DiagramType() }

func (e *SvekEdge) linkStrategy() decoration.LinkStrategy { return e.link.LinkStrategy() }

// ArrowDirection is the direction of the label arrow (GuideLine).
func (e *SvekEdge) ArrowDirection() klimt.Direction {
	if e.link.LinkArrow() == abel.LinkArrowBackward {
		return e.arrowDirectionInternal().Inv()
	}
	return e.arrowDirectionInternal()
}

func (e *SvekEdge) arrowDirectionInternal() klimt.Direction {
	if e.isAutolink() {
		return klimt.DirectionLeft
	}
	start := e.dotPath.StartPoint()
	end := e.dotPath.EndPoint()
	ang := jmath.Atan2(end.X-start.X, end.Y-start.Y)
	pi := math.Pi
	switch {
	case ang > -pi/4 && ang < pi/4:
		return klimt.DirectionDown
	case ang > pi*3/4 || ang < -pi*3/4:
		return klimt.DirectionUp
	case end.X > start.X:
		return klimt.DirectionRight
	}
	return klimt.DirectionLeft
}

// ArrowDirectionInRadian is the angle of the label arrow (GuideLine).
func (e *SvekEdge) ArrowDirectionInRadian() float64 {
	if e.link.LinkArrow() == abel.LinkArrowBackward {
		return math.Pi + e.arrowDirectionInRadianInternal()
	}
	return e.arrowDirectionInRadianInternal()
}

func (e *SvekEdge) arrowDirectionInRadianInternal() float64 {
	if e.isAutolink() {
		return e.dotPath.StartAngle()
	}
	start := e.dotPath.StartPoint()
	end := e.dotPath.EndPoint()
	return jmath.Atan2(end.X-start.X, end.Y-start.Y)
}

func cluster2(bibliotekon *Bibliotekon, entity *abel.Entity) *Cluster {
	for _, cl := range bibliotekon.allCluster {
		if cl.group == entity {
			return cl
		}
	}
	throwIllegalArgument("")
	return nil
}

func newSvekEdge(link *abel.Link, skinParam klimt.ISkinParam, stringBounder font.StringBounder,
	fc, cardinalityFont *klimt.FontConfiguration, bibliotekon *Bibliotekon, pragma *klimt.Pragma,
	graphvizVersion dot.GraphvizVersion) *SvekEdge {
	if link == nil || skinParam == nil || bibliotekon == nil {
		throwNPE()
	}
	e := &SvekEdge{link: link, skinParam: skinParam, bibliotekon: bibliotekon}

	if graphvizVersion.UseShieldForQuantifier() &&
		(link.LinkArg().Quantifier1().OK || link.LinkArg().Role1().OK) {
		link.Entity1().EnsureMargins(klimt.MarginsUniform(16))
	}
	if graphvizVersion.UseShieldForQuantifier() &&
		(link.LinkArg().Quantifier2().OK || link.LinkArg().Role2().OK) {
		link.Entity2().EnsureMargins(klimt.MarginsUniform(16))
	}
	if k := link.LinkArg().Kal1(); k.OK {
		e.kal1 = newKal(e, k.S, skinParam, link.Entity1(), link, stringBounder)
	}
	if k := link.LinkArg().Kal2(); k.OK {
		e.kal2 = newKal(e, k.S, skinParam, link.Entity2(), link, stringBounder)
	}

	e.useRankSame = skinParam.UseRankSame()
	e.startUID = linkEntityPort1(link, bibliotekon)
	e.endUID = linkEntityPort2(link, bibliotekon)

	if e.startUID.startsWith(clusterCenterID) {
		e.ltail = cluster2(bibliotekon, link.Entity1())
	}
	if e.endUID.startsWith(clusterCenterID) {
		e.lhead = cluster2(bibliotekon, link.Entity2())
	}

	// Upstream tests link.getColors() != null, which always holds.
	skinParam = klimt.ColorsMute(link.Colors(), skinParam)
	fc = fc.Mute(link.Colors())

	e.backgroundColor = skinParam.BackgroundColor()
	e.defaultThickness = skinParam.Thickness(klimt.LineParamArrow, nil)
	e.arrowLollipopColor = skinParam.HtmlColor(klimt.ColorParamArrowLollipop, nil, false)
	if e.arrowLollipopColor == nil {
		e.arrowLollipopColor = e.backgroundColor
	}
	e.pragma = pragma
	e.stringBounder = stringBounder

	e.lineColor = bibliotekon.colorSequence.value()
	e.noteLabelColor = bibliotekon.colorSequence.value()
	e.startTailColor = bibliotekon.colorSequence.value()
	e.endHeadColor = bibliotekon.colorSequence.value()

	var labelOnly klimt.TextBlock
	if klimt.DisplayIsNull(link.Label()) {
		labelOnly = klimt.TextBlockUtilsEmptyTextBlock()
		if link.LinkArrow() != abel.LinkArrowNoneOrSeveral {
			labelOnly = addMagicArrow(labelOnly, e, fc)
		}
	} else {
		alignment := messageTextAlignment(e.diagramType(), skinParam)
		hasSeveralGuideLines := link.Label().HasSeveralGuideLines()
		arrowStyle := e.defaultStyleDefinition(link.Stereotype()).MergedStyle(skinParam.CurrentStyleBuilder())
		styleWidth := arrowStyle.WrapWidth()
		wrapWidth := skinParam.MaxMessageSize()
		if styleWidth.MaxWidth() > 0 {
			wrapWidth = styleWidth
		}
		var block klimt.TextBlock
		if hasSeveralGuideLines {
			block = addSeveralMagicArrows(link.Label(), e, fc, alignment, skinParam)
		} else {
			block = link.Label().Create0(fc, alignment, skinParam, wrapWidth, klimt.CreoleModeSimpleLine, nil, nil)
		}
		labelOnly = e.addVisibilityModifier(block, link, skinParam)
		if link.LinkArrow() != abel.LinkArrowNoneOrSeveral && !hasSeveralGuideLines {
			labelOnly = addMagicArrow(labelOnly, e, fc)
		}
	}

	if note := link.Note(); note == nil {
		e.labelText = labelOnly
	} else {
		noteOnly := newEntityImageNoteLink(note.Display(), note.Colors(), skinParam, link.StyleBuilder())
		if note.Strategy() == abel.NoteLinkStrategyHalfNotPrinted ||
			note.Strategy() == abel.NoteLinkStrategyHalfPrintedFull {
			e.divideLabelWidthByTwo = true
		}
		switch note.Position() {
		case abel.PositionLeft:
			e.labelText = klimt.TextBlockUtilsMergeLR(noteOnly, labelOnly, geom.VerticalAlignmentCenter)
		case abel.PositionRight:
			e.labelText = klimt.TextBlockUtilsMergeLR(labelOnly, noteOnly, geom.VerticalAlignmentCenter)
		case abel.PositionTop:
			e.labelText = klimt.TextBlockUtilsMergeTB(noteOnly, labelOnly, geom.HorizontalAlignmentCenter)
		default:
			e.labelText = klimt.TextBlockUtilsMergeTB(labelOnly, noteOnly, geom.HorizontalAlignmentCenter)
		}
	}

	cardinality := func(s abel.OptString) klimt.TextBlock {
		if !s.OK {
			return nil
		}
		return klimt.DisplayWithNewlines(skinParam.Pragma(), s.S).Create(cardinalityFont,
			geom.HorizontalAlignmentCenter, skinParam)
	}
	e.startTailText = cardinality(link.Quantifier1())
	e.endHeadText = cardinality(link.Quantifier2())
	if r := link.Role1(); r.OK {
		e.startTailRoleText = roleLabelsCreate(r.S, skinParam.Pragma(), cardinalityFont, skinParam)
	}
	if r := link.Role2(); r.OK {
		e.endHeadRoleText = roleLabelsCreate(r.S, skinParam.Pragma(), cardinalityFont, skinParam)
	}

	if link.Type().MiddleDecor() == decoration.LinkMiddleDecorNone {
		e.labelShield = 0
	} else {
		e.labelShield = 7
	}
	return e
}

// reserveRoomForRoles makes room for the roles of link around its
// entities: Graphviz only knows the size of the quantifier (appendLine), but
// the role is drawn on the other side of the line, so without room it lands
// on the neighbouring nodes and labels. It must run for all the links before
// the first SvekEdge is built: building an edge computes (and caches) the
// shield of its two nodes.
func reserveRoomForRoles(link *abel.Link, skinParam klimt.ISkinParam, stringBounder font.StringBounder,
	cardinalityFont *klimt.FontConfiguration) {
	if r := link.Role1(); r.OK {
		link.Entity1().EnsureMargins(marginsForRole(r.S, skinParam, stringBounder, cardinalityFont))
	}
	if r := link.Role2(); r.OK {
		link.Entity2().EnsureMargins(marginsForRole(r.S, skinParam, stringBounder, cardinalityFont))
	}
}

func marginsForRole(role string, skinParam klimt.ISkinParam, stringBounder font.StringBounder,
	cardinalityFont *klimt.FontConfiguration) klimt.Margins {
	block := roleLabelsCreate(role, skinParam.Pragma(), cardinalityFont, skinParam)
	width := math.Ceil(block.CalculateDimension(stringBounder).Width()) + 2
	return klimt.NewMargins(width, width, 0, 0)
}

func (e *SvekEdge) addVisibilityModifier(block klimt.TextBlock, link *abel.Link,
	skinParam klimt.ISkinParam) klimt.TextBlock {
	if visibilityModifier, ok := link.VisibilityModifier(); ok {
		fore := rose.NewRose().HtmlColor(skinParam, visibilityModifier.Foreground())
		visibility := visibilityModifier.UBlock(skinParam.ClassAttributeIconSize(), fore, nil, false)
		visibility = klimt.TextBlockUtilsWithMargin4(visibility, 0, 1, 2, 0)
		block = klimt.TextBlockUtilsMergeLR(visibility, block, geom.VerticalAlignmentCenter)
	}
	marginLabel := 1.0
	if e.startUID.equalsID(e.endUID) {
		marginLabel = 6
	}
	return klimt.TextBlockUtilsWithMargin(block, marginLabel, marginLabel)
}

func messageTextAlignment(diagramType core.DiagramType, skinParam klimt.ISkinParam) geom.HorizontalAlignment {
	if diagramType == core.State {
		return skinParam.HorizontalAlignment(klimt.AlignmentParamStateMessageAlignment, klimt.ArrowDirectionNull,
			false, klimt.HorizontalAlignmentNull)
	}
	return skinParam.DefaultTextAlignment(geom.HorizontalAlignmentCenter)
}

func (e *SvekEdge) hasNoteLabelText() bool {
	return e.labelText != nil && e.labelText != klimt.TextBlockUtilsEmptyTextBlock()
}

func (e *SvekEdge) appendLine(graphvizVersion dot.GraphvizVersion, sb *strings.Builder, dotMode DotMode,
	dotSplines klimt.DotSplines) {
	sb.WriteString(e.startUID.fullString())
	sb.WriteString("->")
	sb.WriteString(e.endUID.fullString())
	sb.WriteString("[")
	linkType := e.link.TypePatchCluster()
	decoration := linkType.SpecificDecorationSvek(e.linkStrategy())
	if len(decoration) > 0 && !strings.HasSuffix(decoration, ",") {
		decoration += ","
	}
	sb.WriteString(decoration)

	length := e.link.Length()
	if graphvizVersion.IgnoreHorizontalLinks() && length == 1 {
		length = 2
	}
	if e.useRankSame {
		if e.pragma.IsDefine(klimt.PragmaHorizontalLineBetweenDifferentPackageAllowed) || e.link.IsInvis() ||
			length != 1 {
			sb.WriteString("minlen=" + strconv.Itoa(length-1))
			sb.WriteString(",")
		}
	} else {
		sb.WriteString("minlen=" + strconv.Itoa(length-1))
		sb.WriteString(",")
	}
	sb.WriteString("color=\"" + hexColor(e.lineColor) + "\"")
	if e.hasNoteLabelText() || e.link.LinkConstraint() != nil {
		sb.WriteString(",")
		if graphvizVersion.UseXLabelInsteadOfLabel() || dotMode == DotModeNoLeftRightAndXLabel ||
			dotSplines == klimt.DotSplinesOrtho {
			sb.WriteString("xlabel=<")
		} else {
			sb.WriteString("label=<")
		}
		dimNote := constraintSpot
		if e.hasNoteLabelText() {
			dimNote = e.labelText.CalculateDimension(e.stringBounder)
		}
		dimNote = dimNote.Delta(2 * e.labelShield)
		appendTable(sb, e.eventuallyDivideByTwo(dimNote), e.noteLabelColor)
		sb.WriteString(">")
	}
	if e.startTailText != nil {
		sb.WriteString(",")
		sb.WriteString("taillabel=<")
		appendTable(sb, e.startTailText.CalculateDimension(e.stringBounder), e.startTailColor)
		sb.WriteString(">")
	} else if e.startTailRoleText != nil {
		sb.WriteString(",")
		sb.WriteString("taillabel=<")
		appendTable(sb, e.startTailRoleText.CalculateDimension(e.stringBounder), e.startTailColor)
		sb.WriteString(">")
	}
	if e.endHeadText != nil {
		sb.WriteString(",")
		sb.WriteString("headlabel=<")
		appendTable(sb, e.endHeadText.CalculateDimension(e.stringBounder), e.endHeadColor)
		sb.WriteString(">")
	} else if e.endHeadRoleText != nil {
		sb.WriteString(",")
		sb.WriteString("headlabel=<")
		appendTable(sb, e.endHeadRoleText.CalculateDimension(e.stringBounder), e.endHeadColor)
		sb.WriteString(">")
	}
	if e.link.IsInvis() {
		sb.WriteString(",")
		sb.WriteString("style=invis")
	}
	if !e.link.IsConstraint() || e.link.HasTwoEntryPointsSameContainer() {
		sb.WriteString(",constraint=false")
	}
	if st := e.link.Sametail(); st.OK {
		sb.WriteString(",sametail=" + st.S)
	}
	sb.WriteString("];")
	println(sb)
}

func (e *SvekEdge) eventuallyDivideByTwo(dim geom.XDimension2D) geom.XDimension2D {
	if e.divideLabelWidthByTwo {
		return geom.NewXDimension2D(dim.Width()/2, dim.Height())
	}
	return dim
}

// rankSame is the {rank=same} constraint of a horizontal link.
func (e *SvekEdge) rankSame() (string, bool) {
	if !e.pragma.IsDefine(klimt.PragmaHorizontalLineBetweenDifferentPackageAllowed) && e.link.Length() == 1 {
		return "{rank=same; " + e.startUIDPrefix() + "; " + e.endUIDPrefix() + "}", true
	}
	return "", false
}

// appendTable writes the HTML table that reserves the room of a label.
func appendTable(sb *strings.Builder, dim geom.XDimension2D, col int32) {
	appendTableWH(sb, klimt.JavaD2I(dim.Width()), klimt.JavaD2I(dim.Height()), col)
}

func appendTableWH(sb *strings.Builder, w, h int, col int32) {
	sb.WriteString("<TABLE ")
	sb.WriteString("BGCOLOR=\"" + hexColor(col) + "\" ")
	sb.WriteString("FIXEDSIZE=\"TRUE\" WIDTH=\"" + strconv.Itoa(w) + "\" HEIGHT=\"" + strconv.Itoa(h) + "\">")
	sb.WriteString("<TR")
	sb.WriteString(">")
	sb.WriteString("<TD")
	sb.WriteString(">")
	sb.WriteString("</TD>")
	sb.WriteString("</TR>")
	sb.WriteString("</TABLE>")
}

func (e *SvekEdge) startUIDPrefix() string { return e.startUID.prefix() }

func (e *SvekEdge) endUIDPrefix() string { return e.endUID.prefix() }

func (e *SvekEdge) extremitySpecial(center geom.XPoint2D, decor decoration.LinkDecor, angle float64) klimt.UDrawable {
	extremityFactory := extremityFactoryLegacy(decor, e.backgroundColor)
	return extremityFactory.CreateUDrawable(center, angle, nil)
}

func closestSide(node *SvekNode, pt geom.XPoint2D) *geom.Side {
	if node == nil {
		return nil
	}
	if side, ok := node.rectangleArea().ClosestSide(pt); ok {
		return &side
	}
	return nil
}

func (e *SvekEdge) extremitySimplier(center geom.XPoint2D, extremityFactory extremityFactory, angle float64,
	nodeContact *SvekNode, isStart bool, kal *Kal) klimt.UDrawable {
	if extremityFactory == nil {
		return nil
	}
	side := closestSide(nodeContact, center)
	var translateForKal klimt.UTranslate
	if kal == nil {
		translateForKal = klimt.NewUTranslate(0, 0)
	} else {
		translateForKal = kal.translateForDecoration()
		center = translateForKal.Translated(center)
	}
	ext := extremityFactory.CreateUDrawable(center, angle, side).(extremityDrawable)
	decorationLength := ext.DecorationLength()
	move := translateForKal.Compose(klimt.NewUTranslate(decorationLength, 0).Rotate(angle - math.Pi))
	if isStart {
		e.dotPath.MoveStartPointBy(move)
	} else {
		e.dotPath.MoveEndPointBy(move)
	}
	return ext
}

func (e *SvekEdge) extremity(center geom.XPoint2D, decor decoration.LinkDecor, pointListIterator *pointListIterator,
	angle float64, cluster *Cluster, nodeContact *SvekNode) klimt.UDrawable {
	extremityFactory := extremityFactoryLegacy(decor, e.backgroundColor)
	if cluster != nil {
		if extremityFactory != nil {
			return extremityFactory.CreateUDrawable(center, angle, nil)
		}
		if decor == decoration.LinkDecorExtends {
			return extremityFactoryExtends(e.backgroundColor).CreateUDrawable(center, angle, nil)
		}
		return nil
	}
	switch {
	case extremityFactory != nil:
		points := pointListIterator.next()
		if len(points) == 0 {
			return nil
		}
		p0, p1, p2 := points[0], points[1], points[2]
		side := closestSide(nodeContact, p1)
		return extremityFactory.CreateTBRDrawableLegacy(p0, p1, p2, side)
	case decor == decoration.LinkDecorNone:
		sh := klimt.NewUPolygonPoints(pointListIterator.cloneMe().next())
		if contact, ok := sh.CheckMiddleContactForSpecificTriangle(center); ok {
			return udrawableFunc(func(ug klimt.UGraphic) {
				line := klimt.NewULine(contact.X-center.X, contact.Y-center.Y)
				ug = ug.Apply(klimt.UTranslatePoint(center))
				ug.Draw(line)
			})
		}
	default:
		sh := klimt.NewUPolygonPoints(pointListIterator.next())
		return extremityOther(sh)
	}
	return nil
}

// udrawableFunc adapts a function to UDrawable.
type udrawableFunc func(ug klimt.UGraphic)

func (f udrawableFunc) DrawU(ug klimt.UGraphic) { f(ug) }

// solveLine reads the path of the edge, its decorations and the positions
// of its labels in dot's output.
func (e *SvekEdge) solveLine(fullSvg SvgResult) {
	if e.link.IsInvis() {
		return
	}
	idx := fullSvg.indexFromColor(e.lineColor)
	if idx == -1 {
		return
	}
	idx = fullSvg.indexOf("d=\"", idx)
	if idx == -1 {
		throwIllegalState("")
	}
	end := fullSvg.indexOf("\"", idx+3)
	path := fullSvg.substring2(idx+3, end)
	if !path.isPathConsistent() {
		return
	}
	e.dotPath = path.toDotPath()

	tmpStartPoint := e.dotPath.StartPoint()
	tmpEndPoint := e.dotPath.EndPoint()
	svekNode1 := e.svekNode1()
	svekNode2 := e.svekNode2()
	if svekNode1 != nil && svekNode2 != nil {
		tmpPos1 := svekNode1.rectangleArea().PointCenter()
		tmpPos2 := svekNode2.rectangleArea().PointCenter()
		normal := tmpStartPoint.Distance(tmpPos1) + tmpEndPoint.Distance(tmpPos2)
		inversed := tmpStartPoint.Distance(tmpPos2) + tmpEndPoint.Distance(tmpPos1)
		// Sometime, Graphviz inverses the result line.
		if inversed < normal {
			e.dotPath = e.dotPath.Reverse()
		}
	}
	// Used for Kal
	e.dotPathInit = e.dotPath.Copy()

	if e.projectionCluster != nil {
		e.projectionCluster.manageEntryExitPoint(e.stringBounder)
	}
	var head, tail *geom.RectangleArea
	if e.lhead != nil {
		head = e.lhead.rectangleArea
	}
	if e.ltail != nil {
		tail = e.ltail.rectangleArea
	}
	e.dotPath = e.dotPath.SimulateCompound(head, tail)

	lineSvg := fullSvg.substring(end)
	linkType := e.link.Type()

	if e.linkStrategy() == decoration.LinkStrategySimplest {
		e.extremity1 = e.extremitySimplier(e.dotPath.StartPoint(),
			extremityFactoryComplete(linkType.Decor2(), e.backgroundColor), e.dotPath.StartAngle()+math.Pi,
			svekNode1, true, e.kal1)
		e.extremity2 = e.extremitySimplier(e.dotPath.EndPoint(),
			extremityFactoryComplete(linkType.Decor1(), e.backgroundColor), e.dotPath.EndAngle(), svekNode2, false,
			e.kal2)
	} else {
		pointListIterator := lineSvg.pointsWithThisColor(e.lineColor)
		if e.link.Length() == 1 && e.isThereTwo(linkType) && count(pointListIterator.cloneMe()) == 2 {
			// Upstream issue 1353.
			points := pointListIterator.next()
			p1 := points[1]
			startPoint := e.dotPath.StartPoint()
			endPoint := e.dotPath.EndPoint()
			if p1.Distance(startPoint) < p1.Distance(endPoint) {
				startPoint = p1
			} else {
				endPoint = p1
			}
			e.extremity1 = e.extremitySpecial(startPoint, linkType.Decor2(), e.dotPath.StartAngle()+math.Pi)
			e.extremity2 = e.extremitySpecial(endPoint, linkType.Decor1(), e.dotPath.EndAngle())
		} else {
			e.extremity1 = e.extremity(e.dotPath.StartPoint(), linkType.Decor2(), pointListIterator,
				e.dotPath.StartAngle()+math.Pi, e.ltail, svekNode1)
			e.extremity2 = e.extremity(e.dotPath.EndPoint(), linkType.Decor1(), pointListIterator,
				e.dotPath.EndAngle(), e.lhead, svekNode2)
		}
	}

	if e.link.Entity1().LeafType() == abel.LeafTypeLollipopHalf {
		svekNode1.addImpact(e.dotPath.StartAngle() + math.Pi)
	}
	if e.link.Entity2().LeafType() == abel.LeafTypeLollipopHalf {
		svekNode2.addImpact(e.dotPath.EndAngle())
	}

	if e.hasNoteLabelText() || e.link.LinkConstraint() != nil {
		if pos, ok := xy(fullSvg, e.noteLabelColor); ok {
			if e.hasNoteLabelText() {
				e.labelXY = klimt.TextBlockUtilsAsPositionable(e.labelText, e.stringBounder, pos)
			} else {
				e.labelXY = klimt.TextBlockUtilsAsPositionableDim(constraintSpot, pos)
			}
		}
	}
	if e.startTailText != nil || e.startTailRoleText != nil {
		if pos, ok := xy(fullSvg, e.startTailColor); ok {
			forSize := e.startTailText
			if forSize == nil {
				forSize = e.startTailRoleText
			}
			e.startTailLabelXY = klimt.TextBlockUtilsAsPositionable(forSize, e.stringBounder, pos)
		}
	}
	if e.endHeadText != nil || e.endHeadRoleText != nil {
		if pos, ok := xy(fullSvg, e.endHeadColor); ok {
			forSize := e.endHeadText
			if forSize == nil {
				forSize = e.endHeadRoleText
			}
			e.endHeadLabelXY = klimt.TextBlockUtilsAsPositionable(forSize, e.stringBounder, pos)
		}
	}
	if !e.isOpalisable() {
		e.setOpale(false)
	}
}

func (e *SvekEdge) isThereTwo(linkType decoration.LinkType) bool {
	return extremityFactoryLegacy(linkType.Decor2(), e.backgroundColor) != nil &&
		extremityFactoryLegacy(linkType.Decor1(), e.backgroundColor) != nil
}

func count(it *pointListIterator) int {
	nb := 0
	for it.hasNext() {
		it.next()
		nb++
	}
	return nb
}

func (e *SvekEdge) svekNode2() *SvekNode { return e.bibliotekon.node(e.link.Entity2()) }

func (e *SvekEdge) svekNode1() *SvekNode { return e.bibliotekon.node(e.link.Entity1()) }

func (e *SvekEdge) svekCluster1() *Cluster { return e.bibliotekon.cluster(e.link.Entity1()) }

func (e *SvekEdge) svekCluster2() *Cluster { return e.bibliotekon.cluster(e.link.Entity2()) }

func (e *SvekEdge) isOpalisable() bool { return len(e.dotPath.Beziers()) <= 1 }

func xy(svgResult SvgResult, c int32) (geom.XPoint2D, bool) {
	idx := svgResult.indexFromColor(c)
	if idx == -1 {
		return geom.XPoint2D{}, false
	}
	return svekGetMinXY(svgResult.substring(idx).extractList(svgPointsEquals)), true
}

func (e *SvekEdge) defaultStyleDefinition(stereotype *klimt.Stereotype) klimt.StyleSignature {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement,
		uml.DiagramTypeStyleName(e.diagramType()), klimt.SNameArrow).WithTOBECHANGED(stereotype)
}

// DrawU draws the edge: its path, decorations, labels, quantifiers and
// qualifier boxes.
func (e *SvekEdge) DrawU(ug klimt.UGraphic) {
	if e.opale || e.link.IsInvis() || e.dotPath == nil {
		return
	}
	link := e.link
	ug.Draw(klimt.NewUComment(link.CommentForSvg()))
	group := klimt.NewUGroupAt(link.Location())
	group.Put(klimt.UGroupDataUID, link.UID())
	group.Put(klimt.UGroupClass, "link")
	group.Put(klimt.UGroupID, "link_"+link.Entity1().Name()+"_"+link.Entity2().Name())
	group.Put(klimt.UGroupDataEntity1, link.Entity1().Name())
	group.Put(klimt.UGroupDataEntity2, link.Entity2().Name())
	group.Put(klimt.UGroupDataEntity1UID, link.Entity1().UID())
	group.Put(klimt.UGroupDataEntity2UID, link.Entity2().UID())
	if linkTypeName, ok := link.Type().LinkTypeName(); ok {
		group.Put(klimt.UGroupDataLinkType, linkTypeName)
	}
	ug.StartGroup(group)
	x, y := 0.0, 0.0
	url := link.URL()
	if url != nil {
		ug.StartURL(url)
	}
	if link.IsAutoLinkOfAGroup() {
		if cl := e.bibliotekon.cluster(link.Entity1()); cl != nil {
			x += cl.rectangleArea.Width()
			x -= e.dotPath.StartPoint().X - cl.rectangleArea.MinX()
		}
	}
	x += e.dx
	y += e.dy

	styleLine := e.defaultStyleDefinition(link.Stereotype()).MergedStyle(link.StyleBuilder())
	suggestedStroke := styleLine.Stroke()
	rainbow := decoration.RainbowBuild(styleLine, e.skinParam.IHtmlColorSet())

	arrowHeadColor := rainbow.ArrowHeadColor()
	c := rainbow.Color()
	// Upstream tests link.getColors() != null, which always holds.
	if newColor := link.Colors().FirstColor(color.TypeArrow, color.TypeLine); newColor != nil {
		c = newColor
		arrowHeadColor = c
	}

	ug = ug.Apply(klimt.Bg(color.None())).Apply(c)
	linkType := link.Type()
	var stroke klimt.UStroke
	if !linkType.Style().IsNormal() {
		stroke = linkType.Stroke3(e.defaultThickness)
	} else {
		stroke = linkType.Stroke3(&suggestedStroke)
	}
	if s := klimt.ColorsSpecificLineStroke(link.Colors()); s != nil {
		stroke = *s
	}
	ug = ug.Apply(stroke)

	todraw := e.dotPath.Copy()

	if radiusStr, ok := e.pragma.Value(klimt.PragmaEdgeCornerRadius); ok {
		if radius, err := strconv.ParseFloat(strings.TrimSpace(radiusStr), 64); err == nil && radius > 0 {
			todraw.MuteToRoundOrthogonalPaths(radius)
		}
	}

	magneticForce1 := klimt.UTranslateNone()
	if n := e.svekNode1(); n != nil {
		magneticForce1 = n.magneticBorder().ForceAt(ug.StringBounder(), todraw.StartPoint().Move(e.dx, e.dy))
		todraw.MoveStartPointBy(magneticForce1)
	} else if cl := e.svekCluster1(); cl != nil {
		magneticForce1 = cl.magneticBorder().ForceAt(ug.StringBounder(), todraw.StartPoint().Move(e.dx, e.dy))
		todraw.MoveStartPointBy(magneticForce1)
	}
	magneticForce2 := klimt.UTranslateNone()
	if n := e.svekNode2(); n != nil {
		magneticForce2 = n.magneticBorder().ForceAt(ug.StringBounder(), todraw.EndPoint().Move(e.dx, e.dy))
		todraw.MoveEndPointBy(magneticForce2)
	} else if cl := e.svekCluster2(); cl != nil {
		magneticForce2 = cl.magneticBorder().ForceAt(ug.StringBounder(), todraw.EndPoint().Move(e.dx, e.dy))
		todraw.MoveEndPointBy(magneticForce2)
	}

	codeLine, _ := link.CodeLine()
	todraw.SetCommentAndCodeLine(uniq(e.ids, link.IDCommentForSvg()), codeLine)

	e.drawRainbow(ug.Apply(klimt.NewUTranslate(x, y)), c, arrowHeadColor, todraw, link.SupplementaryColors(),
		stroke, magneticForce1, magneticForce2)

	ug = ug.Apply(klimt.UStrokeSimple()).Apply(c)

	if e.hasNoteLabelText() && e.labelXY != nil &&
		(link.Note() == nil || link.Note().Strategy() != abel.NoteLinkStrategyHalfNotPrinted) {
		e.labelText.DrawU(ug.Apply(klimt.NewUTranslate(x+e.labelXY.Position().X+e.labelShield,
			y+e.labelXY.Position().Y+e.labelShield)))
	}
	if e.startTailLabelXY != nil {
		pos := e.startTailLabelXY.Position()
		labelX := x + pos.X
		labelY := y + pos.Y
		if e.startTailText != nil {
			e.startTailText.DrawU(ug.Apply(klimt.NewUTranslate(labelX, labelY)))
			if e.startTailRoleText != nil {
				e.drawRoleLabel(ug, e.startTailRoleText, e.startTailText, pos, e.dotPath.StartPoint(),
					e.dotPath.EndPoint(), e.dotPath.Sample(), x, y)
			}
		} else if e.startTailRoleText != nil {
			e.startTailRoleText.DrawU(ug.Apply(klimt.NewUTranslate(labelX, labelY)))
		}
	}
	if e.endHeadLabelXY != nil {
		pos := e.endHeadLabelXY.Position()
		labelX := x + pos.X
		labelY := y + pos.Y
		if e.endHeadText != nil {
			e.endHeadText.DrawU(ug.Apply(klimt.NewUTranslate(labelX, labelY)))
			if e.endHeadRoleText != nil {
				e.drawRoleLabel(ug, e.endHeadRoleText, e.endHeadText, pos, e.dotPath.EndPoint(),
					e.dotPath.StartPoint(), e.dotPath.Sample(), x, y)
			}
		} else if e.endHeadRoleText != nil {
			e.endHeadRoleText.DrawU(ug.Apply(klimt.NewUTranslate(labelX, labelY)))
		}
	}
	if linkType.MiddleDecor() != decoration.LinkMiddleDecorNone {
		middle := e.dotPath.Middle()
		angleRad := middle.Angle()
		angleDeg := -angleRad * 180.0 / math.Pi
		mi := middleFactoryFor(linkType.MiddleDecor(), e.arrowLollipopColor, e.backgroundColor).
			CreateUDrawable(angleDeg - 45)
		mi.DrawU(ug.Apply(klimt.NewUTranslate(x+middle.X(), y+middle.Y())))
	}
	if url != nil {
		ug.CloseURL()
	}
	if lc := link.LinkConstraint(); lc != nil {
		xConstraint := x + e.labelXY.Position().X
		yConstraint := y + e.labelXY.Position().Y
		square := constraintSquare(xConstraint, yConstraint)
		bez := todraw.Sample()
		var minPt *geom.XPoint2D
		minDist := math.MaxFloat64
		for i := range square {
			pt := square[i]
			for _, pt2 := range bez {
				distance := pt2.Distance(pt)
				if minPt == nil || distance < minDist {
					minPt = &square[i]
					minDist = distance
				}
			}
		}
		if minPt == nil {
			throwNPE()
		}
		lc.SetPosition(link, *minPt)
		lc.DrawMe(ug, e.skinParam)
	}
	if e.kal1 != nil {
		e.kal1.DrawU(ug)
	}
	if e.kal2 != nil {
		e.kal2.DrawU(ug)
	}
	ug.CloseGroup()
}

// drawRoleLabel draws a role on the other side of the line than the
// quantifier, out of the line and of the areas of the two nodes.
func (e *SvekEdge) drawRoleLabel(ug klimt.UGraphic, role, quantifier klimt.TextBlock, quantifierPos,
	thisEndpoint, otherEndpoint geom.XPoint2D, pathSamples []geom.XPoint2D, x, y float64) {
	pos := roleLabelsPosition(quantifier.CalculateDimension(e.stringBounder),
		role.CalculateDimension(e.stringBounder), quantifierPos, thisEndpoint, otherEndpoint, pathSamples,
		e.nodeAreas())
	role.DrawU(ug.Apply(klimt.NewUTranslate(x+pos.X, y+pos.Y)))
}

// nodeAreas returns the areas of the two nodes of the link, in the
// coordinates of the path.
func (e *SvekEdge) nodeAreas() []geom.RectangleArea {
	var result []geom.RectangleArea
	for _, node := range []*SvekNode{e.svekNode1(), e.svekNode2()} {
		if node != nil {
			result = append(result, node.rectangleArea().Move(-e.dx, -e.dy))
		}
	}
	return result
}

func (e *SvekEdge) computeKal() {
	if e.kal1 != nil {
		tr := klimt.UTranslatePoint(e.dotPathInit.StartPoint()).Compose(klimt.NewUTranslate(e.dx, e.dy))
		e.kal1.setTranslate(tr)
	}
	if e.kal2 != nil {
		tr := klimt.UTranslatePoint(e.dotPathInit.EndPoint()).Compose(klimt.NewUTranslate(e.dx, e.dy))
		e.kal2.setTranslate(tr)
	}
}

func constraintSquare(x, y float64) []geom.XPoint2D {
	return []geom.XPoint2D{
		geom.NewXPoint2D(x, y), geom.NewXPoint2D(x+5, y), geom.NewXPoint2D(x+10, y),
		geom.NewXPoint2D(x, y+5), geom.NewXPoint2D(x+10, y+5),
		geom.NewXPoint2D(x, y+10), geom.NewXPoint2D(x+5, y+10), geom.NewXPoint2D(x+10, y+10),
	}
}

// uniq returns comment, or comment-1, comment-2... when it is already used.
func uniq(ids map[string]bool, comment string) string {
	if !ids[comment] {
		ids[comment] = true
		return comment
	}
	for i := 1; ; i++ {
		candidate := comment + "-" + strconv.Itoa(i)
		if !ids[candidate] {
			ids[candidate] = true
			return candidate
		}
	}
}

func (e *SvekEdge) drawRainbow(ug klimt.UGraphic, c, headColor color.HColor, todraw *klimt.DotPath,
	supplementaryColors []*color.Colors, stroke klimt.UStroke, magneticForce1, magneticForce2 klimt.UTranslate) {
	ug.Draw(todraw)
	linkType := e.link.Type()

	if headColor.IsTransparent() {
		if line := extremityArrowLine(e.extremity1); line != nil {
			ugHead := ug.Apply(c).Apply(stroke.OnlyThickness())
			line(ugHead.Apply(magneticForce1))
		}
	} else if e.extremity1 != nil {
		ugHead := ug.Apply(headColor).Apply(stroke.OnlyThickness())
		if linkType.Decor2().IsFill() {
			ugHead = ugHead.Apply(klimt.Bg(c))
		} else {
			ugHead = ugHead.Apply(klimt.Bg(color.None()))
		}
		e.extremity1.DrawU(ugHead.Apply(magneticForce1))
	}

	if headColor.IsTransparent() {
		if line := extremityArrowLine(e.extremity2); line != nil {
			ugHead := ug.Apply(c).Apply(stroke.OnlyThickness())
			line(ugHead.Apply(magneticForce2))
		}
	} else if e.extremity2 != nil {
		ugHead := ug.Apply(headColor).Apply(stroke.OnlyThickness())
		if linkType.Decor1().IsFill() {
			ugHead = ugHead.Apply(klimt.Bg(c))
		} else {
			ugHead = ugHead.Apply(klimt.Bg(color.None()))
		}
		e.extremity2.DrawU(ugHead.Apply(magneticForce2))
	}

	for i, colors := range supplementaryColors {
		d := float64(2 * (i + 1))
		ug.Apply(klimt.NewUTranslate(d, d)).Apply(colors.Color(color.TypeLine)).Draw(todraw)
	}
}

func (e *SvekEdge) isInverted() bool { return e.link.IsInverted() }

func (e *SvekEdge) decorDzeta() float64 {
	linkType := e.link.Type()
	return float64(linkType.Decor1().Margin() + linkType.Decor2().Margin())
}

func (e *SvekEdge) horizontalDzeta(stringBounder font.StringBounder) float64 {
	if e.startUID.equalsID(e.endUID) {
		return e.decorDzeta()
	}
	if !e.isHorizontal() {
		return 0
	}
	strategy := &arithmeticStrategySum{}
	if e.hasNoteLabelText() {
		strategy.eat(e.labelText.CalculateDimension(stringBounder).Width())
	}
	if e.startTailText != nil {
		strategy.eat(e.startTailText.CalculateDimension(stringBounder).Width())
	}
	if e.endHeadText != nil {
		strategy.eat(e.endHeadText.CalculateDimension(stringBounder).Width())
	}
	return strategy.result() + e.decorDzeta()
}

func (e *SvekEdge) isHorizontal() bool { return e.link.Length() == 1 }

func (e *SvekEdge) verticalDzeta(stringBounder font.StringBounder) float64 {
	if e.startUID.equalsID(e.endUID) {
		return e.decorDzeta()
	}
	if e.isHorizontal() {
		return 0
	}
	strategy := &arithmeticStrategySum{}
	if e.hasNoteLabelText() {
		strategy.eat(e.labelText.CalculateDimension(stringBounder).Height())
	}
	if e.startTailText != nil {
		strategy.eat(e.startTailText.CalculateDimension(stringBounder).Height())
	}
	if e.endHeadText != nil {
		strategy.eat(e.endHeadText.CalculateDimension(stringBounder).Height())
	}
	return strategy.result() + e.decorDzeta()
}

// manageCollision moves the quantifiers out of the nodes they overlap.
func (e *SvekEdge) manageCollision(allNodes []*SvekNode) {
	for _, sh := range allNodes {
		cl := geom.PositionableUtilsAddMargin(sh, 8, 8)
		if (e.startTailText != nil || e.startTailRoleText != nil) && e.startTailLabelXY != nil &&
			geom.PositionableUtilsIntersect(cl, e.startTailLabelXY) {
			e.startTailLabelXY = geom.PositionableUtilsMoveAwayFrom(cl, e.startTailLabelXY)
		}
		if (e.endHeadText != nil || e.endHeadRoleText != nil) && e.endHeadLabelXY != nil &&
			geom.PositionableUtilsIntersect(cl, e.endHeadLabelXY) {
			e.endHeadLabelXY = geom.PositionableUtilsMoveAwayFrom(cl, e.endHeadLabelXY)
		}
	}
}

// MoveDelta implements geom.Moveable.
func (e *SvekEdge) MoveDelta(deltaX, deltaY float64) {
	e.dx += deltaX
	e.dy += deltaY
}

// DotPath returns the path of the edge, moved to its final position.
func (e *SvekEdge) DotPath() *klimt.DotPath {
	result := e.dotPath.Copy()
	result.MoveDelta(e.dx, e.dy)
	return result
}

func (e *SvekEdge) length() int { return e.link.Length() }

func (e *SvekEdge) setOpale(opale bool) {
	e.link.SetOpale(opale)
	e.opale = opale
}

// IsHidden reports whether the link is hidden.
func (e *SvekEdge) IsHidden() bool { return e.link.IsHidden() }

func (e *SvekEdge) isLinkFromOrTo(group *abel.Entity) bool {
	return e.link.Entity1() == group || e.link.Entity2() == group
}

func (e *SvekEdge) hasEntryPoint() bool { return e.link.HasEntryPoint() }

func (e *SvekEdge) sameConnections(other *SvekEdge) bool { return e.link.SameConnections(other.link) }

func (e *SvekEdge) isAutolink() bool { return e.link.Entity1() == e.link.Entity2() }

func (e *SvekEdge) myPoint(entity *abel.Entity) geom.XPoint2D {
	switch entity {
	case e.link.Entity1():
		return klimt.NewUTranslate(e.dx, e.dy).Translated(e.dotPath.StartPoint())
	case e.link.Entity2():
		return klimt.NewUTranslate(e.dx, e.dy).Translated(e.dotPath.EndPoint())
	}
	throwIllegalArgument("")
	return geom.XPoint2D{}
}

func (e *SvekEdge) startContactPoint() (geom.XPoint2D, bool) {
	if e.dotPath == nil {
		return geom.XPoint2D{}, false
	}
	return klimt.NewUTranslate(e.dx, e.dy).Translated(e.dotPath.StartPoint()), true
}

func (e *SvekEdge) endContactPoint() (geom.XPoint2D, bool) {
	if e.dotPath == nil {
		return geom.XPoint2D{}, false
	}
	return klimt.NewUTranslate(e.dx, e.dy).Translated(e.dotPath.EndPoint()), true
}

func (e *SvekEdge) moveStartPoint(dx, dy float64) {
	e.dotPath.MoveStartPoint(dx, dy)
	e.dotPathInit.MoveStartPoint(dx, dy)
}

func (e *SvekEdge) moveEndPoint(dx, dy float64) {
	e.dotPath.MoveEndPoint(dx, dy)
	e.dotPathInit.MoveEndPoint(dx, dy)
}

func (e *SvekEdge) replaceDotPath(newPath *klimt.DotPath) {
	e.dotPath = newPath
	e.dotPathInit = newPath.Copy()
}
