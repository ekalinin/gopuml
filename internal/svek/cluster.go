// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/Cluster.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/ClusterDecoration.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"math"
	"strconv"
	"strings"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/dot"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// clusterCenterID prefixes the special point of a group that links start
// from or end at.
const clusterCenterID = "za"

// Cluster is a group (package, namespace...) as a dot subgraph: its nodes,
// its child clusters, the colors that identify it and its title in dot's
// output, and its position once laid out.
type Cluster struct {
	parentCluster   *Cluster
	group           *abel.Entity
	nodes           []*SvekNode
	children        []*Cluster
	color           int32
	colorTitle      int32
	colorNoteTop    int32
	colorNoteBottom int32
	skinParam       klimt.ISkinParam
	diagram         *abel.CucaDiagram

	clusterHeader *ClusterHeader

	xyTitle      *geom.XPoint2D
	xyNoteTop    *geom.XPoint2D
	xyNoteBottom *geom.XPoint2D

	rectangleArea *geom.RectangleArea

	togetherCounter int
}

// MoveDelta implements geom.Moveable.
func (c *Cluster) MoveDelta(deltaX, deltaY float64) {
	move := func(p *geom.XPoint2D) *geom.XPoint2D {
		if p == nil {
			return nil
		}
		m := p.Move(deltaX, deltaY)
		return &m
	}
	c.xyNoteTop = move(c.xyNoteTop)
	c.xyNoteBottom = move(c.xyNoteBottom)
	c.xyTitle = move(c.xyTitle)
	if c.rectangleArea != nil {
		r := c.rectangleArea.Move(deltaX, deltaY)
		c.rectangleArea = &r
	}
}

// entityPositionsExceptNormal returns the positions other than NORMAL of
// the nodes, as an EnumSet.
func (c *Cluster) entityPositionsExceptNormal() []abel.EntityPosition {
	return entityPositionsExceptNormal(c.nodes)
}

func entityPositionsExceptNormal(nodes []*SvekNode) []abel.EntityPosition {
	var seen [16]bool
	for _, sh := range nodes {
		if sh.entityPosition != abel.EntityPositionNormal {
			seen[sh.entityPosition] = true
		}
	}
	var result []abel.EntityPosition
	for p, ok := range seen {
		if ok {
			result = append(result, abel.EntityPosition(p))
		}
	}
	return result
}

func newRootCluster(diagram *abel.CucaDiagram, colors *colorSequence, root *abel.Entity) *Cluster {
	return newCluster(diagram, nil, colors, root)
}

func newCluster(diagram *abel.CucaDiagram, parentCluster *Cluster, colors *colorSequence, group *abel.Entity) *Cluster {
	if group == nil {
		throwIllegalState("")
	}
	c := &Cluster{parentCluster: parentCluster, group: group, diagram: diagram}
	c.color = colors.value()
	c.colorTitle = colors.value()
	c.colorNoteTop = colors.value()
	c.colorNoteBottom = colors.value()
	c.skinParam = klimt.ColorsMute(group.Colors(), diagram.SkinParam())
	return c
}

func (c *Cluster) addNode(node *SvekNode) {
	if node == nil {
		throwNPE()
	}
	c.nodes = append(c.nodes, node)
	node.cluster = c
}

func (c *Cluster) nodesOrderedTop(lines []*SvekEdge) []*SvekNode {
	var firsts []*SvekNode
	shs := map[string]*SvekNode{}
	for _, node := range c.nodes {
		shs[node.uid] = node
	}
	for _, l := range lines {
		if l.isInverted() {
			if sh := shs[l.startUIDPrefix()]; sh != nil && isNormalPosition(sh) {
				firsts = insertAt(firsts, 0, sh)
			}
		}
	}
	return firsts
}

func isNormalPosition(sh *SvekNode) bool { return sh.entityPosition == abel.EntityPositionNormal }

func (c *Cluster) nodesOrderedWithoutTop(lines []*SvekEdge) []*SvekNode {
	var all []*SvekNode
	shs := map[string]*SvekNode{}
	for _, sh := range c.nodes {
		if !isNormalPosition(sh) {
			continue
		}
		all = append(all, sh)
		shs[sh.uid] = sh
	}
	for _, l := range lines {
		if l.isInverted() {
			if sh := shs[l.startUIDPrefix()]; sh != nil {
				all = removeFirst(all, sh)
			}
		}
	}
	return all
}

// removeFirst is List.remove(Object).
func removeFirst[T comparable](s []T, v T) []T {
	for i, x := range s {
		if x == v {
			return append(s[:i:i], s[i+1:]...)
		}
	}
	return s
}

func (c *Cluster) createChild(header *ClusterHeader, colors *colorSequence, g *abel.Entity) *Cluster {
	child := newCluster(c.diagram, c, colors, g)
	child.clusterHeader = header
	c.children = append(c.children, child)
	return child
}

// Group returns the group of the cluster.
func (c *Cluster) Group() *abel.Entity { return c.group }

func (c *Cluster) titleAndAttributeWidth() int {
	minimumWidth := c.style().Value(klimt.PNameMinimumWidth).AsDouble()
	return max(c.clusterHeader.titleAndAttributeWidth, klimt.JavaD2I(math.Ceil(minimumWidth)))
}

func (c *Cluster) titleAndAttributeHeight() int { return c.clusterHeader.titleAndAttributeHeight }

func (c *Cluster) setTitlePosition(pos geom.XPoint2D) { c.xyTitle = &pos }

func (c *Cluster) setNoteTopPosition(pos geom.XPoint2D) { c.xyNoteTop = &pos }

func (c *Cluster) setNoteBottomPosition(pos geom.XPoint2D) { c.xyNoteBottom = &pos }

// clusterDefaultStyleDefinition is Cluster.getDefaultStyleDefinition.
func clusterDefaultStyleDefinition(diagramStyleName klimt.SName, symbol uSymbol,
	groupType abel.GroupType) klimt.StyleSignatureBasic {
	switch {
	case diagramStyleName == klimt.SNameStateDiagram:
		return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameStateDiagram,
			klimt.SNameState, klimt.SNameGroup)
	case symbol != nil:
		return klimt.StyleSignatureBasicOf5(klimt.SNameRoot, klimt.SNameElement, diagramStyleName, klimt.SNameGroup,
			symbol.SNames())
	case groupType == abel.GroupTypePackage:
		return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, diagramStyleName, klimt.SNamePackage,
			klimt.SNameGroup)
	}
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, diagramStyleName, klimt.SNameGroup)
}

// DrawU draws the cluster: its decoration (folder, frame, node...) with
// its title.
func (c *Cluster) DrawU(ug klimt.UGraphic) {
	if c.group.IsHidden() {
		return
	}
	if c.diagram.Pragma().IsTrue(klimt.PragmaKermor) {
		notPorted("ClusterDotStringKermor")
	}
	fullName := c.group.Name()
	if !strings.HasPrefix(fullName, "##") {
		ug.Draw(klimt.NewUComment("cluster " + fullName))
	}
	diagramType := c.diagram.DiagramType()
	style := c.style()
	shadowing := style.Shadowing()
	var borderColor color.HColor
	if lc := c.group.Colors().Color(color.TypeLine); lc != nil {
		borderColor = lc
	} else {
		borderColor = style.Value(klimt.PNameLineColor).AsColor(c.skinParam.IHtmlColorSet())
	}
	rounded := style.Value(klimt.PNameRoundCorner).AsDouble()
	if c.skinParam.StrictUmlStyle() {
		rounded = 0
	}
	diagonalCorner := style.Value(klimt.PNameDiagonalCorner).AsDouble()

	uGroup := klimt.NewUGroupAt(c.group.Location())
	uGroup.Put(klimt.UGroupClass, "cluster")
	uGroup.Put(klimt.UGroupID, "cluster_"+fullName)
	uGroup.Put(klimt.UGroupDataEntity, c.group.Name())
	uGroup.Put(klimt.UGroupDataUID, c.group.UID())
	uGroup.Put(klimt.UGroupDataQualifiedName, c.group.Quark().QualifiedName())
	ug.StartGroup(uGroup)

	url := c.group.URL99()
	if url != nil {
		ug.StartURL(url)
	}
	defer func() {
		if url != nil {
			ug.CloseURL()
		}
		ug.CloseGroup()
	}()

	if len(c.entityPositionsExceptNormal()) > 0 {
		c.manageEntryExitPoint(ug.StringBounder())
	}
	if c.skinParam.UseSwimlanes(diagramType) {
		c.drawSwinLinesState(ug, borderColor)
		return
	}
	isState := diagramType == core.State
	if isState && c.group.USymbol() == abel.USymbolNull {
		c.drawUState(ug, rounded, shadowing)
		return
	}
	packageStyle, ok := c.group.PackageStyle()
	if !ok {
		packageStyle = c.skinParam.PackageStyle()
	}
	stroke := clusterStrokeInternal(c.group, style)

	backColor := c.backColor(style)
	backColor = clusterBackColor(backColor, c.group.Stereotype(), uml.DiagramTypeStyleName(diagramType),
		usymbolOf(c.group.USymbol()), c.skinParam.CurrentStyleBuilder(), c.skinParam.IHtmlColorSet(),
		c.group.GroupType())

	decoration := newClusterDecoration(packageStyle, usymbolOf(c.group.USymbol()), c.clusterHeader.title,
		c.clusterHeader.stereo, *c.rectangleArea, stroke)
	horizontalAlignment := c.clusterHeader.titleHorizontalAlignment()
	decoration.drawU(ug, backColor, borderColor, shadowing, rounded, horizontalAlignment,
		c.skinParam.StereotypeAlignment(), diagonalCorner)
}

// drawUState draws a composite state laid out with its parent: a rounded
// box with the title, the description lines and the oo symbol.
func (c *Cluster) drawUState(ug klimt.UGraphic, rounded, shadowing float64) {
	styleBuilder := c.skinParam.CurrentStyleBuilder()
	colorSet := c.skinParam.IHtmlColorSet()
	stereotype := c.group.Stereotype()
	borderColor := c.group.Colors().Color(color.TypeLine)
	if borderColor == nil {
		borderColor = styleStateOf(stereotype, styleBuilder).Value(klimt.PNameLineColor).AsColor(colorSet)
	}
	northBackcolor := c.group.Colors().Color(color.TypeBack)
	centerBackColor, southBackcolor := northBackcolor, northBackcolor
	if northBackcolor == nil {
		backColor := func(part klimt.SName) color.HColor {
			return stateStyle().AddSName(part).WithTOBECHANGED(stereotype).MergedStyle(styleBuilder).
				Value(klimt.PNameBackGroundColor).AsColor(colorSet)
		}
		northBackcolor = backColor(klimt.SNameName)
		centerBackColor = backColor(klimt.SNameDescription)
		southBackcolor = backColor(klimt.SNameBody)
	}
	attribute := stateDescription(c.group, c.skinParam)
	attributeHeight := attribute.CalculateDimension(ug.StringBounder()).Height()
	area := *c.rectangleArea
	if area.Dimension().Width() == 0 {
		// Upstream prints "Cluster::drawUState issue".
		return
	}
	var stroke klimt.UStroke
	if s := klimt.ColorsSpecificLineStroke(c.group.Colors()); s != nil {
		stroke = *s
	} else {
		stroke = styleStateOf(stereotype, styleBuilder).Stroke()
	}
	descriptionHeight := attributeHeight
	if attributeHeight > 0 {
		descriptionHeight += klimt.IEntityImageMargin
	}
	nameHeight := c.clusterHeader.title.CalculateDimension(ug.StringBounder()).Height() + klimt.IEntityImageMargin
	newRoundedContainer(borderColor, area.Dimension(), nameHeight, descriptionHeight, northBackcolor,
		centerBackColor, southBackcolor, stroke, rounded, shadowing).
		DrawU(ug.Apply(klimt.NewUTranslate(area.MinX(), area.MinY())))
	drawAligned(c.clusterHeader.titleHorizontalAlignment(), ug.Apply(klimt.NewUTranslate(area.MinX(), c.xyTitle.Y)),
		c.clusterHeader.title, klimt.IEntityImageMargin, 0, area.Width())
	if attributeHeight > 0 {
		attribute.DrawU(ug.Apply(klimt.NewUTranslate(area.MinX()+klimt.IEntityImageMargin,
			area.MinY()+nameHeight+klimt.IEntityImageMargin/2.0)))
	}
	if stereotype != nil && stereotype.IsWithOOSymbol() {
		drawStateSymbol(ug.Apply(borderColor), area.MaxX(), area.MaxY())
	}
}

func (c *Cluster) style() *klimt.Style {
	diagramType := c.diagram.DiagramType()
	sym := c.group.USymbol()
	if sym == abel.USymbolNull {
		sym = abel.USymbolPackage
	}
	return clusterDefaultStyleDefinition(uml.DiagramTypeStyleName(diagramType), usymbolOf(sym), c.group.GroupType()).
		WithTOBECHANGED(c.group.Stereotype()).MergedStyle(c.skinParam.CurrentStyleBuilder())
}

// clusterStrokeInternal is Cluster.getStrokeInternal.
func clusterStrokeInternal(group *abel.Entity, style *klimt.Style) klimt.UStroke {
	if s := klimt.ColorsSpecificLineStroke(group.Colors()); s != nil {
		return *s
	}
	return style.Stroke()
}

// manageEntryExitPoint is Cluster.manageEntryExitPoint: the cluster box
// becomes the frontier of its inner nodes and clusters and of its entry
// and exit points, and the title is centered on it.
func (c *Cluster) manageEntryExitPoint(stringBounder font.StringBounder) {
	var insides []geom.RectangleArea
	var points []geom.XPoint2D
	for _, sh := range c.nodes {
		if isNormalPosition(sh) {
			insides = append(insides, sh.rectangleArea())
		} else {
			points = append(points, sh.rectangleArea().PointCenter())
		}
	}
	for _, in := range c.children {
		// Upstream prints "Frontier null for" a child without a box.
		if in.rectangleArea != nil {
			insides = append(insides, *in.rectangleArea)
		}
	}
	calculator := newFrontierCalculator(*c.rectangleArea, insides, points, c.skinParam.Rankdir())
	if c.titleAndAttributeWidth() > 0 && c.titleAndAttributeHeight() > 0 {
		calculator.ensureMinWidth(float64(c.titleAndAttributeWidth() + 10))
	}
	r := calculator.suggestedPosition()
	c.rectangleArea = &r
	widthTitle := c.clusterHeader.title.CalculateDimension(stringBounder).Width()
	c.xyTitle = &geom.XPoint2D{X: r.MinX() + (r.Width()-widthTitle)/2, Y: r.MinY() + klimt.IEntityImageMargin}
}

func (c *Cluster) drawSwinLinesState(ug klimt.UGraphic, borderColor color.HColor) {
	c.clusterHeader.title.DrawU(ug.Apply(klimt.UTranslateDx(c.xyTitle.X)))
	line := klimt.ULineV(c.rectangleArea.Height())
	ug = ug.Apply(borderColor)
	ug.Apply(klimt.UTranslateDx(c.rectangleArea.MinX())).Draw(line)
	ug.Apply(klimt.UTranslateDx(c.rectangleArea.MaxX())).Draw(line)
}

func (c *Cluster) setPosition(minPt, maxPt geom.XPoint2D) {
	r := geom.RectangleAreaBuild(minPt, maxPt)
	c.rectangleArea = &r
}

func (c *Cluster) printCluster1(sb *strings.Builder, lines []*SvekEdge, stringBounder font.StringBounder) bool {
	tmp := c.nodesOrderedTop(lines)
	if len(tmp) == 0 {
		return false
	}
	for _, node := range tmp {
		node.appendShape(sb, stringBounder)
	}
	return true
}

func (c *Cluster) printTogether(together *abel.Together, otherTogethers []*abel.Together, sb *strings.Builder,
	nodesOrderedWithoutTop []*SvekNode, stringBounder font.StringBounder, lines []*SvekEdge, dotMode DotMode,
	graphvizVersion dot.GraphvizVersion, typ core.DiagramType) {
	sb.WriteString("subgraph " + c.clusterID() + "t" + strconv.Itoa(c.togetherCounter) + " {\n")
	for _, node := range nodesOrderedWithoutTop {
		if node.together() == together {
			node.appendShape(sb, stringBounder)
		}
	}
	for _, child := range c.children {
		if child.group.Together() == together {
			child.printInternal(sb, lines, stringBounder, dotMode, graphvizVersion, typ)
		}
	}
	for _, otherTogether := range otherTogethers {
		if otherTogether.Parent() == together {
			c.printTogether(otherTogether, otherTogethers, sb, nodesOrderedWithoutTop, stringBounder, lines, dotMode,
				graphvizVersion, typ)
		}
	}
	sb.WriteString("}\n")
	c.togetherCounter++
}

func (c *Cluster) printCluster2(sb *strings.Builder, lines []*SvekEdge, stringBounder font.StringBounder,
	dotMode DotMode, graphvizVersion dot.GraphvizVersion, typ core.DiagramType) *SvekNode {
	var added *SvekNode
	var togethers []*abel.Together // a LinkedHashSet
	nodesOrderedWithoutTop := c.nodesOrderedWithoutTop(lines)
	for _, node := range nodesOrderedWithoutTop {
		if together := node.together(); together == nil {
			node.appendShape(sb, stringBounder)
		} else {
			togethers = addTogetherWithParents(togethers, together)
		}
		added = node
	}
	for _, child := range c.children {
		if t := child.group.Together(); t != nil {
			togethers = addTogetherWithParents(togethers, t)
		}
	}
	for _, together := range togethers {
		if together.Parent() == nil {
			c.printTogether(together, togethers, sb, nodesOrderedWithoutTop, stringBounder, lines, dotMode,
				graphvizVersion, typ)
		}
	}
	if c.skinParam.UseRankSame() && dotMode != DotModeNoLeftRightAndXLabel &&
		!graphvizVersion.IgnoreHorizontalLinks() {
		c.appendRankSame(sb, lines)
	}
	for _, child := range c.children {
		if child.group.Together() == nil {
			child.printInternal(sb, lines, stringBounder, dotMode, graphvizVersion, typ)
		}
	}
	return added
}

func addTogetherWithParents(collection []*abel.Together, together *abel.Together) []*abel.Together {
	for t := together; t != nil; t = t.Parent() {
		found := false
		for _, x := range collection {
			if x == t {
				found = true
				break
			}
		}
		if !found {
			collection = append(collection, t)
		}
	}
	return collection
}

func (c *Cluster) printInternal(sb *strings.Builder, lines []*SvekEdge, stringBounder font.StringBounder,
	dotMode DotMode, graphvizVersion dot.GraphvizVersion, typ core.DiagramType) {
	clusterDotString{cluster: c, skinParam: c.skinParam}.printInternal(sb, lines, stringBounder, dotMode,
		graphvizVersion, typ)
}

func (c *Cluster) appendRankSame(sb *strings.Builder, lines []*SvekEdge) {
	for _, same := range c.rankSame(lines) {
		sb.WriteString(same)
		println(sb)
	}
}

// rankSame returns the {rank=same} constraints of the horizontal links
// inside the cluster, in the order of upstream's HashSet.
func (c *Cluster) rankSame(lines []*SvekEdge) []string {
	var rankSame []string
	seen := map[string]bool{}
	for _, l := range lines {
		if l.hasEntryPoint() {
			continue
		}
		startUID := l.startUIDPrefix()
		endUID := l.endUIDPrefix()
		if c.isInCluster(startUID) && c.isInCluster(endUID) {
			if same, ok := l.rankSame(); ok && !seen[same] {
				seen[same] = true
				rankSame = append(rankSame, same)
			}
		}
	}
	return klimt.JavaHashMapOrder(rankSame)
}

func (c *Cluster) isInCluster(uid string) bool {
	for _, node := range c.nodes {
		if node.uid == uid {
			return true
		}
	}
	return false
}

func (c *Cluster) clusterID() string { return "cluster" + strconv.Itoa(int(c.color)) }

// clusterSpecialPointID is Cluster.getSpecialPointId: the point node of a
// group that links start from or end at.
func clusterSpecialPointID(group *abel.Entity) string { return clusterCenterID + group.UID() }

func (c *Cluster) minPoint(typ core.DiagramType) (string, bool) {
	if c.skinParam.UseSwimlanes(typ) {
		return "minPoint" + strconv.Itoa(int(c.color)), true
	}
	return "", false
}

func (c *Cluster) maxPoint(typ core.DiagramType) (string, bool) {
	if c.skinParam.UseSwimlanes(typ) {
		return "maxPoint" + strconv.Itoa(int(c.color)), true
	}
	return "", false
}

func (c *Cluster) isLabel() bool {
	return c.titleAndAttributeHeight() > 0 || c.titleAndAttributeWidth() > 0
}

func (c *Cluster) backColor(style *klimt.Style) color.HColor {
	if c.group.IsRoot() {
		return nil
	}
	if result := c.group.Colors().Color(color.TypeBack); result != nil {
		return result
	}
	return style.Value(klimt.PNameBackGroundColor).AsColor(c.skinParam.IHtmlColorSet())
}

// clusterBackColor is Cluster.getBackColor(HColor, ...): the background of
// a group, transparent when it has none.
func clusterBackColor(backColor color.HColor, stereotype *klimt.Stereotype, styleName klimt.SName, symbol uSymbol,
	styleBuilder *klimt.StyleBuilder, colorSet *color.HColorSet, groupType abel.GroupType) color.HColor {
	style := clusterDefaultStyleDefinition(styleName, symbol, groupType).MergedStyle(styleBuilder)
	if backColor == nil {
		backColor = style.Value(klimt.PNameBackGroundColor).AsColor(colorSet)
	}
	if backColor == nil || backColor.Equals(color.Transparent()) {
		backColor = color.Transparent()
	}
	return backColor
}

func (c *Cluster) titleDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return c.clusterHeader.title.CalculateDimension(stringBounder)
}

// magneticBorder is the magnetic border of the decoration of the group, at
// the position of the cluster.
func (c *Cluster) magneticBorder() klimt.MagneticBorder {
	if c.group.USymbol() == abel.USymbolNull && c.group.GroupType() != abel.GroupTypePackage {
		return klimt.MagneticBorderNone{}
	}
	sym := usymbolOf(c.group.USymbol())
	packageStyle, ok := c.group.PackageStyle()
	if !ok {
		packageStyle = c.skinParam.PackageStyle()
	}
	style := clusterDefaultStyleDefinition(uml.DiagramTypeStyleName(core.Class), sym, c.group.GroupType()).
		WithTOBECHANGED(c.group.Stereotype()).MergedStyle(c.skinParam.CurrentStyleBuilder())
	stroke := clusterStrokeInternal(c.group, style)
	decoration := newClusterDecoration(packageStyle, sym, c.clusterHeader.title, c.clusterHeader.stereo,
		*c.rectangleArea, stroke)
	textBlock := decoration.textBlock(color.Black, color.Black, 0, 0,
		c.skinParam.HorizontalAlignment(klimt.AlignmentParamPackageTitleAlignment, klimt.ArrowDirectionNull, false,
			klimt.HorizontalAlignmentNull),
		c.skinParam.StereotypeAlignment(), 0)
	orig := textBlock.MagneticBorder()
	rect := *c.rectangleArea
	return klimt.MagneticBorderFunc(func(stringBounder font.StringBounder, position geom.XPoint2D) klimt.UTranslate {
		return orig.ForceAt(stringBounder, position.Move(-rect.MinX(), -rect.MinY()))
	})
}

// clusterDecoration is upstream's ClusterDecoration: the symbol of a group
// drawn big, with its title and stereotype.
type clusterDecoration struct {
	defaultStroke klimt.UStroke
	symbol        uSymbol
	title, stereo klimt.TextBlock
	rectangleArea geom.RectangleArea
}

func newClusterDecoration(style klimt.PackageStyle, symbol uSymbol, title, stereo klimt.TextBlock,
	rectangleArea geom.RectangleArea, stroke klimt.UStroke) clusterDecoration {
	if symbol == nil {
		symbol = packageStyleToUSymbol(style)
	}
	return clusterDecoration{symbol: symbol, stereo: stereo, title: title, rectangleArea: rectangleArea,
		defaultStroke: stroke}
}

func (d clusterDecoration) drawU(ug klimt.UGraphic, backColor, borderColor color.HColor, shadowing, roundCorner float64,
	titleAlignment, stereoAlignment geom.HorizontalAlignment, diagonalCorner float64) {
	asBig := d.textBlock(backColor, borderColor, shadowing, roundCorner, titleAlignment, stereoAlignment,
		diagonalCorner)
	asBig.DrawU(ug.Apply(klimt.UTranslatePoint(geom.NewXPoint2D(d.rectangleArea.MinX(), d.rectangleArea.MinY()))))
}

func (d clusterDecoration) textBlock(backColor, borderColor color.HColor, shadowing, roundCorner float64,
	titleAlignment, stereoAlignment geom.HorizontalAlignment, diagonalCorner float64) klimt.TextBlock {
	biColor := klimt.NewFashion(backColor, borderColor)
	if d.symbol == nil {
		throwUnsupported("")
	}
	symbolContext := biColor.WithShadow(shadowing).WithStroke(d.defaultStroke).WithCorner(roundCorner, diagonalCorner)
	return d.symbol.AsBig(d.title, titleAlignment, d.stereo, d.rectangleArea.Width(), d.rectangleArea.Height(),
		symbolContext, stereoAlignment)
}
