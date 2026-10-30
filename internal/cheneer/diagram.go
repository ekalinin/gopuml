// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cheneer/ChenEerDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package cheneer

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
)

// ChenEerDiagram is a parsed Chen EER diagram: entities, relationships and
// attributes with the links between them.
type ChenEerDiagram struct {
	abel.AbstractEntityDiagram
	// owners is upstream's ownerStack: the entities, relationships and
	// composite attributes whose block is open.
	owners []*abel.Entity
}

var (
	_ core.TypedDiagram = (*ChenEerDiagram)(nil)
	_ abel.Diagram      = (*ChenEerDiagram)(nil)
)

// NewChenEerDiagram returns an empty Chen EER diagram for src; md5map holds
// the base64 images of the source. Upstream's ChenEerDiagram overrides no
// method of CucaDiagram, so the diagram has no hooks.
func NewChenEerDiagram(ctx *core.RenderContext, src *core.UmlSource, md5map map[string]string) *ChenEerDiagram {
	d := &ChenEerDiagram{}
	d.InitCucaDiagram(ctx, src, core.ChenEER, md5map, nil)
	return d
}

// PushOwner pushes the owner of the following attributes.
func (d *ChenEerDiagram) PushOwner(group *abel.Entity) { d.owners = append(d.owners, group) }

// PopOwner pops an owner, and reports false when there is none.
func (d *ChenEerDiagram) PopOwner() bool {
	if len(d.owners) == 0 {
		return false
	}
	d.owners = d.owners[:len(d.owners)-1]
	return true
}

// PeekOwner returns the owner of the current attribute, nil when there is
// none. The attributes are linked to their owner by where they appear in the
// source, without nesting them in it as packages are, so the current group
// of CucaDiagram is not used.
func (d *ChenEerDiagram) PeekOwner() *abel.Entity {
	if len(d.owners) == 0 {
		return nil
	}
	return d.owners[len(d.owners)-1]
}
