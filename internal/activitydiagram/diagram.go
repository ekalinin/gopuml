// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram/ActivityDiagram.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram/ConditionalContext.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram

import (
	"strconv"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// ConditionalContext is upstream's ConditionalContext: the branch of an
// open "if" and the context it is nested in.
type ConditionalContext struct {
	branch    *abel.Entity
	direction klimt.Direction
	parent    *ConditionalContext
}

func newConditionalContext(parent *ConditionalContext, branch *abel.Entity, direction klimt.Direction) *ConditionalContext {
	if branch == nil {
		uml.ThrowNPE("")
	}
	if branch.LeafType() != abel.LeafTypeBranch {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return &ConditionalContext{branch: branch, direction: direction, parent: parent}
}

// Direction returns the direction of the branch.
func (c *ConditionalContext) Direction() klimt.Direction { return c.direction }

// Parent returns the enclosing context, nil for none.
func (c *ConditionalContext) Parent() *ConditionalContext { return c.parent }

// Branch returns the branch entity.
func (c *ConditionalContext) Branch() *abel.Entity { return c.branch }

// ActivityDiagram is a parsed activity diagram in the legacy syntax.
type ActivityDiagram struct {
	abel.CucaDiagram

	lastEntityConsulted        *abel.Entity
	lastEntityBrancheConsulted *abel.Entity
	currentContext             *ConditionalContext
}

var (
	_ core.TypedDiagram = (*ActivityDiagram)(nil)
	_ abel.Diagram      = (*ActivityDiagram)(nil)
)

// NewActivityDiagram returns an empty activity diagram for src; md5map holds
// the base64 images of the source.
func NewActivityDiagram(ctx *core.RenderContext, src *core.UmlSource, md5map map[string]string) *ActivityDiagram {
	d := &ActivityDiagram{}
	d.InitCucaDiagram(ctx, src, core.Activity, md5map, activityHooks{d})
	d.SetNamespaceSeparator("", false)
	return d
}

// activityHooks are the CucaDiagram methods ActivityDiagram overrides.
type activityHooks struct{ d *ActivityDiagram }

// UpdateLasts records the last activity and the last branch created; notes
// are not activities.
func (h activityHooks) UpdateLasts(result *abel.Entity) {
	if result == nil || result.LeafType() == abel.LeafTypeNote {
		return
	}
	h.d.lastEntityConsulted = result
	if result.LeafType() == abel.LeafTypeBranch {
		h.d.lastEntityBrancheConsulted = result
	}
}

func (activityHooks) IsHideEmptyDescriptionForState() bool { return false }
func (activityHooks) CleanID(id string) string             { return abel.DefaultCleanID(id) }

// autoBranch is the code of a branch without code.
func (d *ActivityDiagram) autoBranch() string { return d.UniqueSequence("#") }

// startIf creates the branch of an "if", with optionalCode or an automatic
// code, and opens its context.
func (d *ActivityDiagram) startIf(location *text.LineLocation, optionalCode string, hasCode bool) {
	idShort := optionalCode
	if !hasCode {
		idShort = d.autoBranch()
	}
	quark := d.QuarkInContext(true, d.CleanID(idShort))
	br := d.ReallyCreateLeaf(location, quark, klimt.DisplayCreate(""), abel.LeafTypeBranch, abel.USymbolNull)
	d.currentContext = newConditionalContext(d.currentContext, br, klimt.DirectionDown)
}

// endif closes the context of the current "if".
func (d *ActivityDiagram) endif() { d.currentContext = d.currentContext.Parent() }

// getStart returns the start, created once.
func (d *ActivityDiagram) getStart(location *text.LineLocation) *abel.Entity {
	quark := d.QuarkInContext(true, "start")
	if quark.Data() == nil {
		d.ReallyCreateLeaf(location, quark, klimt.DisplayWithNewlines(d.Pragma(), "start"), abel.LeafTypeCircleStart,
			abel.USymbolNull)
	}
	return quark.Data()
}

// getEnd returns the end, or the end named by suppID ("(*top)", "(*1)").
func (d *ActivityDiagram) getEnd(location *text.LineLocation, suppID string, hasSuppID bool) *abel.Entity {
	tmp := "end"
	if hasSuppID {
		tmp = "end$" + suppID
	}
	quark := d.QuarkInContext(true, tmp)
	if quark.Data() == nil {
		d.ReallyCreateLeaf(location, quark, klimt.DisplayWithNewlines(d.Pragma(), "end"), abel.LeafTypeCircleEnd,
			abel.USymbolNull)
	}
	return quark.Data()
}

// createNote creates a note.
func (d *ActivityDiagram) createNote(location *text.LineLocation, quark *abel.Quark, display *klimt.Display) *abel.Entity {
	if quark == nil {
		uml.ThrowNPE("")
	}
	return d.ReallyCreateLeaf(location, quark, display, abel.LeafTypeNote, abel.USymbolNull)
}

// DotStrings returns the DOT settings of legacy activity diagrams.
func (d *ActivityDiagram) DotStrings() []string {
	return []string{"nodesep=.20;", "ranksep=0.4;", "edge [fontsize=11,labelfontsize=11];", "node [fontsize=11];"}
}

// Description implements core.Diagram: "(N activities)".
func (d *ActivityDiagram) Description() string {
	return "(" + strconv.Itoa(len(d.Leafs())) + " activities)"
}

// lastEntity is upstream's getLastEntityConsulted.
func (d *ActivityDiagram) lastEntity() *abel.Entity { return d.lastEntityConsulted }

// setLastEntity is upstream's setLastEntityConsulted.
func (d *ActivityDiagram) setLastEntity(last *abel.Entity) { d.lastEntityConsulted = last }

// createInnerActivity opens an inner activity ("--> {") and returns it.
func (d *ActivityDiagram) createInnerActivity(location *text.LineLocation) *abel.Entity {
	quark := d.QuarkInContext(true, d.UniqueSequence("##"))
	d.GotoGroup(location, quark, klimt.DisplayWithNewlines(d.Pragma(), quark.Name()), abel.GroupTypeInnerActivity,
		abel.USymbolNull)
	g := d.CurrentGroup()
	d.lastEntityConsulted = nil
	d.lastEntityBrancheConsulted = nil
	return g
}

// concurrentActivity starts a concurrent part of the current inner
// activity. Upstream's CommandInnerConcurrent, its only caller, is not
// registered.
func (d *ActivityDiagram) concurrentActivity(location *text.LineLocation) {
	if d.CurrentGroup().GroupType() == abel.GroupTypeConcurrentActivity {
		d.EndGroup()
	}
	idShort := d.UniqueSequence("##")
	if d.CurrentGroup().GroupType() != abel.GroupTypeInnerActivity {
		uml.Throw(uml.IllegalStateException, "type="+d.CurrentGroup().GroupType().String())
	}
	quark := d.QuarkInContext(true, idShort)
	d.GotoGroup(location, quark, klimt.DisplayWithNewlines(d.Pragma(), "code"), abel.GroupTypeConcurrentActivity,
		abel.USymbolNull)
	d.lastEntityConsulted = nil
	d.lastEntityBrancheConsulted = nil
}
