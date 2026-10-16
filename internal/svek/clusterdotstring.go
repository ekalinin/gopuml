// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/ClusterDotString.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/abel"
	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/dot"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
)

const (
	clusterRankSource = "source"
	clusterRankSink   = "sink"
	clusterIDEE       = "ee"
)

// clusterDotString writes a cluster as a dot subgraph, with the nested
// subgraphs that keep links to the group and labels in place.
type clusterDotString struct {
	cluster   *Cluster
	skinParam klimt.ISkinParam
}

func (c clusterDotString) isPacked() bool { return c.cluster.group.IsPacked() }

func (c clusterDotString) printInternal(sb *strings.Builder, lines []*SvekEdge, stringBounder font.StringBounder,
	dotMode DotMode, graphvizVersion dot.GraphvizVersion, typ core.DiagramType) {
	cluster := c.cluster
	if cluster.diagram.Pragma().IsTrue(klimt.PragmaKermor) {
		notPorted("ClusterDotStringKermor")
	}
	if c.isPacked() {
		cluster.printCluster1(sb, lines, stringBounder)
		cluster.printCluster2(sb, lines, stringBounder, dotMode, graphvizVersion, typ)
		return
	}
	thereALinkFromOrToGroup2 := c.isThereALinkFromOrToGroup(lines)
	thereALinkFromOrToGroup1 := thereALinkFromOrToGroup2
	useProtectionWhenThereALinkFromOrToGroup := graphvizVersion.UseProtectionWhenThereALinkFromOrToGroup()
	if !useProtectionWhenThereALinkFromOrToGroup {
		thereALinkFromOrToGroup1 = false
	}
	if thereALinkFromOrToGroup1 {
		c.subgraphClusterNoLabel(sb, "a")
	}
	positionsExceptNormal := entityPositionsExceptNormal(cluster.nodes)
	if len(positionsExceptNormal) > 0 {
		for _, line := range lines {
			if line.isLinkFromOrTo(cluster.group) {
				line.projectionCluster = cluster
			}
		}
	}
	protection0 := c.protection0(typ)
	protection1 := c.protection1(typ)
	if len(positionsExceptNormal) > 0 || !useProtectionWhenThereALinkFromOrToGroup {
		protection0 = false
		protection1 = false
	}
	if protection0 {
		c.subgraphClusterNoLabel(sb, "p0")
	}
	sb.WriteString("subgraph " + cluster.clusterID() + " {")
	sb.WriteString("style=solid;")
	sb.WriteString("color=\"" + hexColor(cluster.color) + "\";")

	var label string
	if cluster.isLabel() {
		var sblabel strings.Builder
		sblabel.WriteString("<")
		appendTableWH(&sblabel, cluster.titleAndAttributeWidth(), cluster.titleAndAttributeHeight()-5,
			cluster.colorTitle)
		sblabel.WriteString(">")
		label = sblabel.String()
		align := c.skinParam.HorizontalAlignment(klimt.AlignmentParamPackageTitleAlignment, klimt.ArrowDirectionNull,
			false, klimt.HorizontalAlignmentNull)
		sb.WriteString("labeljust=\"" + strings.ToLower(align.String()[:1]) + "\";")
	} else {
		label = "\"\""
	}

	if len(positionsExceptNormal) > 0 {
		c.printRanks(clusterRankSource, c.withPosition(abel.EntityPosition.IsInput), sb, stringBounder)
		c.printRanks(clusterRankSink, c.withPosition(abel.EntityPosition.IsOutput), sb, stringBounder)
		if c.hasPort() {
			c.subgraphClusterNoLabel(sb, clusterIDEE)
		} else {
			c.subgraphClusterWithLabel(sb, clusterIDEE, label)
		}
	} else {
		sb.WriteString("label=" + label + ";")
		println(sb)
	}

	if thereALinkFromOrToGroup2 {
		sb.WriteString(clusterSpecialPointID(cluster.group) + " [shape=point,width=.01,label=\"\"];")
	}
	if thereALinkFromOrToGroup1 {
		c.subgraphClusterNoLabel(sb, "i")
	}
	if protection1 {
		c.subgraphClusterNoLabel(sb, "p1")
	}
	if c.skinParam.UseSwimlanes(typ) {
		minPoint, _ := cluster.minPoint(typ)
		maxPoint, _ := cluster.maxPoint(typ)
		sb.WriteString("{rank = source; ")
		sb.WriteString(c.sourceInPoint(typ))
		sb.WriteString(" [shape=point,width=.01,label=\"\"];")
		sb.WriteString(minPoint + "->" + c.sourceInPoint(typ) + "  [weight=999];")
		sb.WriteString("}")
		println(sb)
		sb.WriteString("{rank = sink; ")
		sb.WriteString(c.sinkInPoint(typ))
		sb.WriteString(" [shape=point,width=.01,label=\"\"];")
		sb.WriteString("}")
		sb.WriteString(c.sinkInPoint(typ) + "->" + maxPoint + "  [weight=999];")
		println(sb)
	}
	println(sb)

	cluster.printCluster1(sb, lines, stringBounder)

	added := cluster.printCluster2(sb, lines, stringBounder, dotMode, graphvizVersion, typ)
	if len(positionsExceptNormal) > 0 {
		if c.hasPort() {
			sb.WriteString(c.empty() + " [shape=rect,width=.01,height=.01,label=")
			sb.WriteString(label)
			sb.WriteString("];")
		} else if added == nil {
			sb.WriteString(c.empty() + " [shape=point,width=.01,label=\"\"];")
		}
	}
	println(sb)

	sb.WriteString("}")
	if protection1 {
		sb.WriteString("}")
	}
	if thereALinkFromOrToGroup1 {
		sb.WriteString("}")
		sb.WriteString("}")
	}
	if len(positionsExceptNormal) > 0 {
		sb.WriteString("}")
	}
	if protection0 {
		sb.WriteString("}")
	}
	println(sb)
}

func (c clusterDotString) sourceInPoint(typ core.DiagramType) string {
	if c.skinParam.UseSwimlanes(typ) {
		return "sourceIn" + itoa32(c.cluster.color)
	}
	return "null"
}

func (c clusterDotString) sinkInPoint(typ core.DiagramType) string {
	if c.skinParam.UseSwimlanes(typ) {
		return "sinkIn" + itoa32(c.cluster.color)
	}
	return "null"
}

// empty is the node used as the empty node of the cluster: the special
// point of the group (a new node cannot be put inside the nested
// subgraphs).
func (c clusterDotString) empty() string { return clusterSpecialPointID(c.cluster.group) }

func (c clusterDotString) hasPort() bool {
	for _, pos := range entityPositionsExceptNormal(c.cluster.nodes) {
		if pos.IsPort() {
			return true
		}
	}
	return false
}

func (c clusterDotString) subgraphClusterNoLabel(sb *strings.Builder, id string) {
	c.subgraphClusterWithLabel(sb, id, "\"\"")
}

func (c clusterDotString) subgraphClusterWithLabel(sb *strings.Builder, id, label string) {
	sb.WriteString("subgraph " + c.cluster.clusterID() + id + " {")
	sb.WriteString("label=" + label + ";")
}

func (c clusterDotString) printRanks(rank string, entries []*SvekNode, sb *strings.Builder,
	stringBounder font.StringBounder) {
	if len(entries) == 0 {
		return
	}
	sb.WriteString("{rank=" + rank + ";")
	for _, sh1 := range entries {
		sb.WriteString(sh1.uid + ";")
	}
	sb.WriteString("}")
	println(sb)
	for _, sh2 := range entries {
		sh2.appendShape(sb, stringBounder)
	}
	println(sb)
	if c.hasPort() {
		arrow := false
		node := "null"
		for _, sh := range entries {
			if arrow {
				sb.WriteString("->")
			}
			arrow = true
			node = sh.uid
			sb.WriteString(node)
		}
		if arrow {
			sb.WriteString(" [arrowhead=none]")
		}
		sb.WriteByte(';')
		println(sb)
		sb.WriteString(node + "->" + c.empty() + ";")
		println(sb)
	}
}

// withPosition returns the nodes whose position is in the set (upstream's
// EntityPosition.getInputs and getOutputs).
func (c clusterDotString) withPosition(in func(abel.EntityPosition) bool) []*SvekNode {
	var result []*SvekNode
	for _, sh := range c.cluster.nodes {
		if in(sh.entityPosition) {
			result = append(result, sh)
		}
	}
	return result
}

func (c clusterDotString) protection0(typ core.DiagramType) bool {
	return !c.skinParam.UseSwimlanes(typ)
}

func (c clusterDotString) protection1(typ core.DiagramType) bool {
	if c.cluster.group.USymbol() == abel.USymbolNode {
		return true
	}
	return !c.skinParam.UseSwimlanes(typ)
}

func (c clusterDotString) isThereALinkFromOrToGroup(lines []*SvekEdge) bool {
	for _, line := range lines {
		if line.isLinkFromOrTo(c.cluster.group) {
			return true
		}
	}
	return false
}
