// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/Bibliotekon.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/ClusterManager.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// Bibliotekon holds the nodes, edges and clusters of one layout, and the
// color sequence that names them in the DOT graph.
type Bibliotekon struct {
	allCluster []*Cluster

	// nodeMap is upstream's LinkedHashMap<Entity, SvekNode>.
	nodeKeys []*abel.Entity
	nodeMap  map[*abel.Entity]*SvekNode

	lines0, lines1, allLines []*SvekEdge

	links         []*abel.Link
	colorSequence *colorSequence

	// neighborhoods are upstream's Entity.getNeighborhood, set by
	// DotData.removeIrrelevantSametail.
	neighborhoods map[*abel.Entity]*neighborhood
}

func newBibliotekon(links []*abel.Link) *Bibliotekon {
	return &Bibliotekon{links: links, colorSequence: newColorSequence(), nodeMap: map[*abel.Entity]*SvekNode{},
		neighborhoods: map[*abel.Entity]*neighborhood{}}
}

func (b *Bibliotekon) createNode(ent *abel.Entity, image klimt.IEntityImage, stringBounder font.StringBounder) *SvekNode {
	node := newSvekNode(ent, image, b.colorSequence, stringBounder)
	if _, ok := b.nodeMap[ent]; !ok {
		b.nodeKeys = append(b.nodeKeys, ent)
	}
	b.nodeMap[ent] = node
	return node
}

// cluster returns the cluster of a group, nil when there is none.
func (b *Bibliotekon) cluster(ent *abel.Entity) *Cluster {
	for _, cl := range b.allCluster {
		if cl.group == ent {
			return cl
		}
	}
	return nil
}

func (b *Bibliotekon) addLine(line *SvekEdge) {
	b.allLines = append(b.allLines, line)
	if line.length() == 1 {
		if line.hasNoteLabelText() {
			for i, other := range b.lines0 {
				if !other.hasNoteLabelText() && line.sameConnections(other) {
					b.lines0 = insertAt(b.lines0, i, line)
					return
				}
			}
		}
		b.lines0 = append(b.lines0, line)
	} else {
		b.lines1 = append(b.lines1, line)
	}
}

func insertAt[T any](s []T, i int, v T) []T {
	s = append(s, v)
	copy(s[i+1:], s[i:])
	s[i] = v
	return s
}

func (b *Bibliotekon) addCluster(current *Cluster) { b.allCluster = append(b.allCluster, current) }

func (b *Bibliotekon) node(ent *abel.Entity) *SvekNode { return b.nodeMap[ent] }

// nodeUID is getNodeUid: the dot node of an entity, with the ":h" port of
// shielded nodes, or the special point of a group.
func (b *Bibliotekon) nodeUID(ent *abel.Entity) string {
	if result := b.node(ent); result != nil {
		uid := result.uid
		if result.isShielded() {
			uid += ":h"
		}
		return uid
	}
	if ent.IsGroup() {
		return clusterSpecialPointID(ent)
	}
	throwIllegalState("")
	return ""
}

func (b *Bibliotekon) allNodes() []*SvekNode {
	result := make([]*SvekNode, len(b.nodeKeys))
	for i, k := range b.nodeKeys {
		result[i] = b.nodeMap[k]
	}
	return result
}

func (b *Bibliotekon) allLineConnectedTo(leaf *abel.Entity) []*SvekEdge {
	var result []*SvekEdge
	for _, line := range b.allLines {
		if line.isLinkFromOrTo(leaf) {
			result = append(result, line)
		}
	}
	return result
}

func (b *Bibliotekon) line(link *abel.Link) *SvekEdge {
	for _, line := range b.allLines {
		if line.link == link {
			return line
		}
	}
	throwIllegalArgument("")
	return nil
}

// onlyOther is getOnlyOther: the other end of the first link of entity.
func (b *Bibliotekon) onlyOther(entity *abel.Entity) *abel.Entity {
	for _, link := range b.links {
		if link.Contains(entity) {
			if other := link.Other(entity); other != nil {
				return other
			}
		}
	}
	return nil
}

func (b *Bibliotekon) leaf(node *SvekNode) *abel.Entity {
	for _, k := range b.nodeKeys {
		if b.nodeMap[k] == node {
			return k
		}
	}
	throwIllegalArgument("")
	return nil
}

// clusterManager is upstream's ClusterManager: the cluster being filled
// while the groups are visited.
type clusterManager struct {
	bibliotekon *Bibliotekon
	current     *Cluster
}

func (m *clusterManager) addNode(node *SvekNode) { m.current.addNode(node) }

func (m *clusterManager) openCluster(g *abel.Entity, header *ClusterHeader) {
	m.current = m.current.createChild(header, m.bibliotekon.colorSequence, g)
	m.bibliotekon.addCluster(m.current)
}

func (m *clusterManager) closeCluster() {
	if m.current.parentCluster == nil {
		throwIllegalState("")
	}
	m.current = m.current.parentCluster
}
