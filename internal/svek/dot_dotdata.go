// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/dot/DotData.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
)

// dotData is upstream's dot.DotData: what is laid out (the links in their
// order, the leaves, the top group) and the diagram that owns them.
type dotData struct {
	diagram       *abel.CucaDiagram
	links         []*abel.Link
	leafs         []*abel.Entity
	topParent     *abel.Entity
	portionShower PortionShower
}

func newDotData(diagram *abel.CucaDiagram, topParent *abel.Entity, links []*abel.Link, leafs []*abel.Entity,
	portionShower PortionShower) *dotData {
	if topParent == nil {
		throwNPE()
	}
	return &dotData{diagram: diagram, topParent: topParent, links: links, leafs: leafs, portionShower: portionShower}
}

func (d *dotData) rootGroup() *abel.Entity { return d.diagram.RootGroup() }

func (d *dotData) isDegeneratedWithFewEntities(nb int) bool {
	return len(d.diagram.Groups()) == 0 && len(d.links) == 0 && len(d.leafs) == nb
}

func (d *dotData) isHideEmptyDescriptionForState() bool {
	return d.diagram.IsHideEmptyDescriptionForState()
}

func (d *dotData) diagramType() core.DiagramType { return d.diagram.DiagramType() }

func (d *dotData) skinParam() klimt.ISkinParam { return d.diagram.SkinParam() }

func (d *dotData) isEmpty(g *abel.Entity) bool { return d.diagram.IsEmpty(g) }

// removeIrrelevantSametail gives the extends links the same tail, and
// keeps it only where a parent has at least groupInheritance children.
func (d *dotData) removeIrrelevantSametail(bibliotekon *Bibliotekon) {
	var keys []string
	sametails := map[string]int{}
	for _, link := range d.links {
		if link.Type().Decor2().IsExtendsLike() {
			link.SetSametail(abel.Some(link.Entity1().UID()))
		}
		sametail := link.Sametail()
		if !sametail.OK {
			continue
		}
		if _, ok := sametails[sametail.S]; !ok {
			keys = append(keys, sametail.S)
		}
		sametails[sametail.S]++
	}
	toremove := map[string]bool{}
	limit := d.skinParam().GroupInheritance()
	for _, key := range klimt.JavaHashMapOrder(keys) {
		if sametails[key] < limit {
			toremove[key] = true
		} else {
			var some []*abel.Link
			for _, link := range d.links {
				if st := link.Sametail(); st.OK && st.S == key {
					some = append(some, link)
				}
			}
			leaf := d.leaf(key)
			if leaf == nil {
				throwNPE()
			}
			// Upstream sets it on the entity (Entity.setNeighborhood).
			bibliotekon.neighborhoods[leaf] = newNeighborhood(leaf, some, d.linksOfThisLeaf(leaf))
		}
	}
	for _, link := range d.links {
		if sametail := link.Sametail(); sametail.OK && toremove[sametail.S] {
			link.SetSametail(abel.OptString{})
		}
	}
}

func (d *dotData) linksOfThisLeaf(leaf *abel.Entity) []*abel.Link {
	var result []*abel.Link
	for _, link := range d.links {
		if link.Contains(leaf) {
			result = append(result, link)
		}
	}
	return result
}

func (d *dotData) leaf(key string) *abel.Entity {
	for _, entity := range d.leafs {
		if entity.UID() == key {
			return entity
		}
	}
	return nil
}
