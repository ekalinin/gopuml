// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/DotStringFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/dot"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// DotStringFactory writes the DOT graph of the nodes, edges and clusters
// of a Bibliotekon, and reads dot's SVG output back into their positions
// (solve).
type DotStringFactory struct {
	diagramType     core.DiagramType
	skinParam       klimt.ISkinParam
	bibliotekon     *Bibliotekon
	root            *Cluster
	graphvizVersion dot.GraphvizVersion
}

func newDotStringFactory(bibliotekon *Bibliotekon, root *Cluster, diagramType core.DiagramType,
	skinParam klimt.ISkinParam, graphvizVersion dot.GraphvizVersion) *DotStringFactory {
	return &DotStringFactory{bibliotekon: bibliotekon, root: root, diagramType: diagramType, skinParam: skinParam,
		graphvizVersion: graphvizVersion}
}

func (f *DotStringFactory) horizontalDzeta(stringBounder font.StringBounder) float64 {
	result := 0.0
	for _, l := range f.bibliotekon.allLines {
		if c := l.horizontalDzeta(stringBounder); c > result {
			result = c
		}
	}
	return result / 10
}

func (f *DotStringFactory) verticalDzeta(stringBounder font.StringBounder) float64 {
	result := 0.0
	for _, l := range f.bibliotekon.allLines {
		if c := l.verticalDzeta(stringBounder); c > result {
			result = c
		}
	}
	if f.root.diagram.Pragma().IsTrue(klimt.PragmaKermor) {
		return result / 100
	}
	return result / 10
}

// createDotString returns the DOT graph.
func (f *DotStringFactory) createDotString(stringBounder font.StringBounder, dotMode DotMode,
	dotStrings []string) string {
	var sb strings.Builder

	nodesep := f.horizontalDzeta(stringBounder)
	if nodesep < float64(f.minNodeSep()) {
		nodesep = float64(f.minNodeSep())
	}
	if f.skinParam.Nodesep() != 0 {
		nodesep = f.skinParam.Nodesep()
	}
	nodesepInches := PixelToInches(nodesep)
	ranksep := f.verticalDzeta(stringBounder)
	if ranksep < float64(f.minRankSep()) {
		ranksep = float64(f.minRankSep())
	}
	if f.skinParam.Ranksep() != 0 {
		ranksep = f.skinParam.Ranksep()
	}
	ranksepInches := PixelToInches(ranksep)
	sb.WriteString("digraph unix {")
	println(&sb)

	for _, s := range dotStrings {
		switch {
		case strings.HasPrefix(s, "ranksep"):
			sb.WriteString("ranksep=" + ranksepInches + ";")
		case strings.HasPrefix(s, "nodesep"):
			sb.WriteString("nodesep=" + nodesepInches + ";")
		default:
			sb.WriteString(s)
		}
		println(&sb)
	}
	sb.WriteString("remincross=true;")
	println(&sb)
	sb.WriteString("searchsize=500;")
	println(&sb)

	dotSplines := f.skinParam.DotSplines()
	switch dotSplines {
	case klimt.DotSplinesPolyline:
		sb.WriteString("splines=polyline;")
		println(&sb)
	case klimt.DotSplinesOrtho:
		sb.WriteString("splines=ortho;")
		sb.WriteString("forcelabels=true;")
		println(&sb)
	}
	if f.skinParam.Rankdir() == geom.RankdirLeftToRight {
		sb.WriteString("rankdir=LR;")
		println(&sb)
	}
	f.manageMinMaxCluster(&sb)

	if f.root.diagram.Pragma().IsTrue(klimt.PragmaKermor) {
		notPorted("Cluster.printCluster3_forKermor")
	}
	f.root.printCluster1(&sb, f.bibliotekon.allLines, stringBounder)
	for _, line := range f.bibliotekon.lines0 {
		line.appendLine(f.graphvizVersion, &sb, dotMode, dotSplines)
	}
	f.root.printCluster2(&sb, f.bibliotekon.allLines, stringBounder, dotMode, f.graphvizVersion, f.diagramType)
	for _, line := range f.bibliotekon.lines1 {
		line.appendLine(f.graphvizVersion, &sb, dotMode, dotSplines)
	}
	println(&sb)
	sb.WriteString("}")
	return sb.String()
}

func (f *DotStringFactory) manageMinMaxCluster(sb *strings.Builder) {
	var minPointCluster, maxPointCluster []string
	for _, cluster := range f.bibliotekon.allCluster {
		if minPoint, ok := cluster.minPoint(f.diagramType); ok {
			minPointCluster = append(minPointCluster, minPoint)
		}
		if maxPoint, ok := cluster.maxPoint(f.diagramType); ok {
			maxPointCluster = append(maxPointCluster, maxPoint)
		}
	}
	if len(minPointCluster) > 0 {
		sb.WriteString("{rank=min;")
		for _, s := range minPointCluster {
			sb.WriteString(s)
			sb.WriteString(" [shape=point,width=.01,label=\"\"]")
			sb.WriteString(";")
		}
		sb.WriteString("}")
		println(sb)
	}
	if len(maxPointCluster) > 0 {
		sb.WriteString("{rank=max;")
		for _, s := range maxPointCluster {
			sb.WriteString(s)
			sb.WriteString(" [shape=point,width=.01,label=\"\"]")
			sb.WriteString(";")
		}
		sb.WriteString("}")
		println(sb)
	}
}

func (f *DotStringFactory) minRankSep() int {
	if f.diagramType == core.Activity {
		return 40
	}
	if f.root.diagram.Pragma().IsTrue(klimt.PragmaKermor) {
		return 40
	}
	return 60
}

func (f *DotStringFactory) minNodeSep() int {
	if f.diagramType == core.Activity {
		return 20
	}
	return 35
}

var svgSizePattern = regexp.MustCompile(`(?m)<svg\s+width="(\d+)pt"\s+height="(\d+)pt"`)

// solve reads the positions of the nodes, clusters and edges in dot's SVG
// output.
func (f *DotStringFactory) solve(svg string) {
	if len(svg) == 0 {
		panic(errEmptySvg)
	}
	mGraph := svgSizePattern.FindStringSubmatch(svg)
	if mGraph == nil {
		throwIllegalState("")
	}
	fullHeight, err := strconv.Atoi(mGraph[2])
	if err != nil {
		throwNumberFormat(mGraph[2])
	}
	move := newYDelta(float64(fullHeight))
	svgResult := newSvgResult(svg, move)
	for _, node := range f.bibliotekon.allNodes() {
		idx := strings.Index(svg, "<title>"+node.uid+"</title>")
		switch node.typ {
		case klimt.ShapeTypeRectangle, klimt.ShapeTypeRectangleHTMLForPorts, klimt.ShapeTypeRectangleWithCircleInside,
			klimt.ShapeTypeFolder, klimt.ShapeTypeDiamond, klimt.ShapeTypeRectanglePort:
			points := svgResult.substring(idx).extractList(svgPointsEquals)
			minPt := svekGetMinXY(points)
			node.MoveDelta(minPt.X, minPt.Y)
		case klimt.ShapeTypeRoundRectangle:
			idx2 := indexFrom(svg, "d=\"", idx+1)
			idx = indexFrom(svg, "points=\"", idx+1)
			var points []geom.XPoint2D
			if idx2 != -1 && (idx == -1 || idx2 < idx) {
				// Graphviz 2.30
				points = svgResult.substring(idx2).extractList(svgDEquals)
			} else {
				points = svgResult.substring(idx).extractList(svgPointsEquals)
				for range 3 {
					idx = indexFrom(svg, "points=\"", idx+1)
					points = append(points, svgResult.substring(idx).extractList(svgPointsEquals)...)
				}
			}
			minPt := svekGetMinXY(points)
			node.MoveDelta(minPt.X, minPt.Y)
		case klimt.ShapeTypeOctagon, klimt.ShapeTypeHexagon:
			idx = indexFrom(svg, "points=\"", idx+1)
			points := svgResult.substring(idx).extractList(svgPointsEquals)
			minPt := svekGetMinXY(points)
			node.MoveDelta(minPt.X, minPt.Y)
			node.setPolygon(minPt.X, minPt.Y, points)
		case klimt.ShapeTypeCircle, klimt.ShapeTypeOval:
			cx := svekGetValue(svg, idx, "cx")
			cy := svekGetValue(svg, idx, "cy") + float64(fullHeight)
			rx := svekGetValue(svg, idx, "rx")
			ry := svekGetValue(svg, idx, "ry")
			node.MoveDelta(cx-rx, cy-ry)
		default:
			throwIllegalState(node.typ.String() + " " + node.uid)
		}
	}

	for _, cluster := range f.bibliotekon.allCluster {
		if cluster.group.IsPacked() {
			continue
		}
		idx := clusterIndex(svg, cluster.color)
		points := svgResult.substring(idx).extractList(svgPointsEquals)
		minPt := svekGetMinXY(points)
		maxPt := svekGetMaxXY(points)
		cluster.setPosition(minPt, maxPt)
		if cluster.titleAndAttributeWidth() == 0 || cluster.titleAndAttributeHeight() == 0 {
			continue
		}
		idx = clusterIndex(svg, cluster.colorTitle)
		pointsTitle := svgResult.substring(idx).extractList(svgPointsEquals)
		cluster.setTitlePosition(svekGetMinXY(pointsTitle))
		if f.root.diagram.Pragma().IsTrue(klimt.PragmaKermor) {
			notPorted("Kermor cluster notes")
		}
	}

	for _, line := range f.bibliotekon.allLines {
		line.solveLine(svgResult)
	}
	if f.skinParam.DotSplines() == klimt.DotSplinesOrtho {
		f.alignEdgesAtLabelNodes()
	}
	for _, line := range f.bibliotekon.allLines {
		line.manageCollision(f.bibliotekon.allNodes())
	}
}

// alignEdgesAtLabelNodes inserts horizontal corners in orthogonal edges to
// align them through the centers of the label nodes ("transition_...").
func (f *DotStringFactory) alignEdgesAtLabelNodes() {
	var labelNodes []*SvekNode
	for _, node := range f.bibliotekon.allNodes() {
		entity := f.bibliotekon.leaf(node)
		if entity != nil && strings.HasPrefix(entity.Name(), "transition_") {
			labelNodes = append(labelNodes, node)
		}
	}
	for _, labelNode := range labelNodes {
		f.alignEdgesThroughLabel(labelNode)
	}
}

func (f *DotStringFactory) alignEdgesThroughLabel(labelNode *SvekNode) {
	labelEntity := f.bibliotekon.leaf(labelNode)
	if labelEntity == nil {
		return
	}
	connectedEdges := f.bibliotekon.allLineConnectedTo(labelEntity)
	if len(connectedEdges) != 2 {
		return
	}
	edge0, edge1 := connectedEdges[0], connectedEdges[1]
	if edge0.dotPath == nil || edge1.dotPath == nil {
		return
	}
	labelCenterY := labelNode.minY + labelNode.Height()/2.0
	edge0EndY := edge0.DotPath().EndPoint().Y
	edge0StartY := edge0.DotPath().StartPoint().Y
	edgeIn, edgeOut := edge1, edge0
	if math.Abs(edge0EndY-labelCenterY) < math.Abs(edge0StartY-labelCenterY) {
		edgeIn, edgeOut = edge0, edge1
	}
	f.insertCornersForAlignment(edgeIn, edgeOut, labelNode)
}

func (f *DotStringFactory) insertCornersForAlignment(edgeIn, edgeOut *SvekEdge, labelNode *SvekNode) {
	pathIn := edgeIn.DotPath()
	pathOut := edgeOut.DotPath()
	labelCenterX := labelNode.minX + labelNode.Width()/2.0
	labelCenterY := labelNode.minY + labelNode.Height()/2.0
	inStart, inEnd := pathIn.StartPoint(), pathIn.EndPoint()
	outStart, outEnd := pathOut.StartPoint(), pathOut.EndPoint()
	if math.Abs(inEnd.X-outStart.X) < 5.0 {
		return
	}
	if math.Abs(inEnd.Y-outStart.Y) > 30.0 {
		return
	}
	if newPathIn := rebuildPathWithCorner(inStart, inEnd, labelCenterX, labelCenterY, true); newPathIn != nil {
		edgeIn.replaceDotPath(newPathIn)
	}
	if newPathOut := rebuildPathWithCorner(outStart, outEnd, labelCenterX, labelCenterY, false); newPathOut != nil {
		edgeOut.replaceDotPath(newPathOut)
	}
}

func rebuildPathWithCorner(start, end geom.XPoint2D, targetX, _ float64, isIncoming bool) *klimt.DotPath {
	startX, startY := start.X, start.Y
	endX, endY := end.X, end.Y
	if math.Abs(startX-targetX) < 1.0 && math.Abs(endX-targetX) < 1.0 {
		return nil
	}
	var beziers []geom.XCubicCurve2D
	if isIncoming {
		horizontalY := endY - 20
		beziers = append(beziers,
			geom.NewXCubicCurve2D(startX, startY, startX, startY, startX, horizontalY, startX, horizontalY),
			geom.NewXCubicCurve2D(startX, horizontalY, startX, horizontalY, targetX, horizontalY, targetX, horizontalY),
			geom.NewXCubicCurve2D(targetX, horizontalY, targetX, horizontalY, targetX, endY, targetX, endY))
	} else {
		horizontalY := startY + 20
		beziers = append(beziers,
			geom.NewXCubicCurve2D(targetX, startY, targetX, startY, targetX, horizontalY, targetX, horizontalY),
			geom.NewXCubicCurve2D(targetX, horizontalY, targetX, horizontalY, endX, horizontalY, endX, horizontalY),
			geom.NewXCubicCurve2D(endX, horizontalY, endX, horizontalY, endX, endY, endX, endY))
	}
	return klimt.DotPathFromBeziers(beziers)
}

// clusterIndex is getClusterIndex: the position of the element drawn with
// the color of a cluster.
func clusterIndex(svg string, colorInt int32) int {
	colorString := strings.ToLower(color.ToHexRGBColor(colorInt))
	idx := strings.Index(svg, "=\""+colorString+"\"")
	if idx == -1 {
		idx = strings.Index(svg, "stroke:"+colorString+";")
	}
	if idx == -1 {
		throwIllegalState("Cannot find color " + colorString)
	}
	return idx
}

// MoveDelta implements geom.Moveable: it moves every node, edge and
// cluster.
func (f *DotStringFactory) MoveDelta(deltaX, deltaY float64) {
	for _, sh := range f.bibliotekon.allNodes() {
		sh.MoveDelta(deltaX, deltaY)
	}
	for _, line := range f.bibliotekon.allLines {
		line.MoveDelta(deltaX, deltaY)
	}
	for _, cl := range f.bibliotekon.allCluster {
		cl.MoveDelta(deltaX, deltaY)
	}
}

// Bibliotekon returns the nodes, edges and clusters of the graph.
func (f *DotStringFactory) Bibliotekon() *Bibliotekon { return f.bibliotekon }
