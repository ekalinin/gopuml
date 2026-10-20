// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/GroupMakerState.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cucadiagram/PortionShower.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/dot"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/uml"
)

// portionShowerAll is upstream's PortionShower.ALL: every portion shown,
// no stereotype label.
type portionShowerAll struct{}

func (portionShowerAll) ShowPortion(uml.EntityPortion, *abel.Entity) bool { return true }
func (portionShowerAll) VisibleStereotypeLabels(*abel.Entity) []string    { return nil }

// groupMakerState is upstream's GroupMakerState: the image of a composite
// state laid out on its own, with its own Graphviz run.
type groupMakerState struct {
	diagram       *abel.CucaDiagram
	group         *abel.Entity
	stringBounder font.StringBounder
	dotMode       DotMode
	le            *layoutEngine
	version       dot.GraphvizVersion
}

func newGroupMakerState(diagram *abel.CucaDiagram, group *abel.Entity, stringBounder font.StringBounder,
	dotMode DotMode, le *layoutEngine, version dot.GraphvizVersion) *groupMakerState {
	if !group.IsGroup() {
		throwIllegalArgument("")
	}
	return &groupMakerState{diagram: diagram, group: group, stringBounder: stringBounder, dotMode: dotMode, le: le,
		version: version}
}

// innerGroupHierarchy is upstream's GroupMakerState.InnerGroupHierarchy:
// the groups of the composite state, without its concurrent regions, are
// the top groups of its layout.
type innerGroupHierarchy struct {
	diagram *abel.CucaDiagram
	group   *abel.Entity
}

func (h innerGroupHierarchy) ChildrenGroups(parent *abel.Entity) []*abel.Entity {
	if parent.IsRoot() {
		return withoutConcurrentStates(h.diagram.ChildrenGroups(h.group))
	}
	return withoutConcurrentStates(h.diagram.ChildrenGroups(parent))
}

func (h innerGroupHierarchy) IsEmpty(g *abel.Entity) bool { return h.diagram.IsEmpty(g) }

func withoutConcurrentStates(groups []*abel.Entity) []*abel.Entity {
	var result []*abel.Entity
	for _, g := range groups {
		if g.GroupType() != abel.GroupTypeConcurrentState {
			result = append(result, g)
		}
	}
	return result
}

// pureInnerLinks are the links between entities of the group.
func (m *groupMakerState) pureInnerLinks() []*abel.Link {
	var result []*abel.Link
	for _, link := range m.diagram.Links() {
		if abel.EntityUtilsIsPureInnerLink12(m.group, link) {
			result = append(result, link)
		}
	}
	return result
}

// image returns the image of the group. The error is a layout error,
// returned with the image that shows it, or the error of the context,
// without image.
func (m *groupMakerState) image() (klimt.IEntityImage, error) {
	if m.group.CountChildren() == 0 && len(m.group.Groups()) == 0 {
		return newEntityImageState(m.group), nil
	}
	if m.group.GroupType() == abel.GroupTypeConcurrentState {
		return m.createGeneralImageBuilder(m.group.Leafs()).buildImage(m.stringBounder, []string{}, m.le)
	}
	if m.group.GroupType() != abel.GroupTypeState {
		throwUnsupported(m.group.GroupType().String())
	}
	image, err := m.createGeneralImageBuilder(withoutConcurrentLeafs(m.group.Leafs())).
		buildImage(m.stringBounder, []string{}, m.le)
	if image == nil || err != nil {
		return image, err
	}
	if m.containsSomeConcurrentStates() {
		inners := []klimt.IEntityImage{image}
		for _, inner := range m.group.Leafs() {
			if inner.LeafType() == abel.LeafTypeStateConcurrent {
				inners = append(inners, inner.SvekImage())
			}
		}
		image = newConcurrentStates(inners, m.group.ConcurrentSeparator(), m.diagram.SkinParam(),
			m.group.Stereotype())
	}
	return newInnerStateAutonom(image, m.group), nil
}

// containsSomeConcurrentStates reports whether the group has concurrent
// regions, laid out already.
func (m *groupMakerState) containsSomeConcurrentStates() bool {
	for _, entity := range m.group.Leafs() {
		if entity.LeafType() == abel.LeafTypeStateConcurrent {
			return true
		}
	}
	return false
}

// createGeneralImageBuilder is the builder of the layout of leafs, with
// the links inside the group.
func (m *groupMakerState) createGeneralImageBuilder(leafs []*abel.Entity) *graphvizImageBuilder {
	links := m.pureInnerLinks()
	bibliotekon := newBibliotekon(links)
	data := newDotData(m.diagram, m.group, links, leafs, innerGroupHierarchy{m.diagram, m.group},
		portionShowerAll{})
	root := newRootCluster(m.diagram, bibliotekon.colorSequence, data.rootGroup())
	cm := &clusterManager{bibliotekon: bibliotekon, current: root}
	dotStringFactory := newDotStringFactory(bibliotekon, root, m.diagram.DiagramType(), m.diagram.SkinParam(),
		m.version)
	return &graphvizImageBuilder{dotData: data, pragma: m.diagram.Pragma(), styleName: klimt.SNameStateDiagram,
		dotMode: m.dotMode, dotStringFactory: dotStringFactory, clusterManager: cm}
}

// withoutConcurrentLeafs are the leafs that are not concurrent regions
// laid out already.
func withoutConcurrentLeafs(leafs []*abel.Entity) []*abel.Entity {
	var result []*abel.Entity
	for _, leaf := range leafs {
		if leaf.LeafType() != abel.LeafTypeStateConcurrent {
			result = append(result, leaf)
		}
	}
	return result
}
