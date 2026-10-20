// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/statediagram/StateDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package statediagram

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
)

// concurrentPrefix names the groups of the concurrent regions.
const concurrentPrefix = "CONC"

// StateDiagram is a parsed state diagram.
type StateDiagram struct {
	abel.AbstractEntityDiagram

	hideEmptyDescription bool
	// transitionNodes counts the label nodes of transitions.
	transitionNodes int
}

var (
	_ core.TypedDiagram = (*StateDiagram)(nil)
	_ abel.Diagram      = (*StateDiagram)(nil)
)

// NewStateDiagram returns an empty state diagram for src; md5map holds the
// base64 images of the source.
func NewStateDiagram(ctx *core.RenderContext, src *core.UmlSource, md5map map[string]string) *StateDiagram {
	d := &StateDiagram{}
	d.InitCucaDiagram(ctx, src, core.State, md5map, stateHooks{d})
	d.SetNamespaceSeparator(".", true)
	return d
}

// stateHooks are the CucaDiagram methods StateDiagram overrides.
type stateHooks struct{ d *StateDiagram }

func (stateHooks) UpdateLasts(*abel.Entity)               {}
func (h stateHooks) IsHideEmptyDescriptionForState() bool { return h.d.hideEmptyDescription }
func (stateHooks) CleanID(id string) string               { return abel.DefaultCleanID(id) }

// RequiredPasses implements command.Diagram: states are declared in every
// pass, transitions in the second and notes on states in the third.
func (d *StateDiagram) RequiredPasses() []command.ParserPass {
	return []command.ParserPass{command.One, command.Two, command.Three}
}

// SetHideEmptyDescription implements uml.HideEmptyDescriptionSetter.
func (d *StateDiagram) SetHideEmptyDescription(hide bool) { d.hideEmptyDescription = hide }

// checkConcurrentStateOk reports whether the state of a quark may be used
// in the current group: a state of a concurrent region cannot be used
// outside of it, nor a state outside of the current concurrent region.
func (d *StateDiagram) checkConcurrentStateOk(code *abel.Quark) bool {
	existing := code.Data()
	if existing == nil {
		return true
	}
	current := d.CurrentGroup()
	if current.GroupType() == abel.GroupTypeConcurrentState && current != existing.ParentContainer() {
		return false
	}
	if parent := existing.ParentContainer(); parent != nil &&
		parent.GroupType() == abel.GroupTypeConcurrentState && current != parent {
		return false
	}
	return true
}

// pseudoState returns the pseudo state named prefix in the current group,
// creating it with type typ.
func (d *StateDiagram) pseudoState(location *text.LineLocation, prefix string, typ abel.LeafType) *abel.Entity {
	idShort := prefix
	if g := d.CurrentGroup(); !g.IsRoot() {
		idShort += g.Name()
	}
	quark := d.QuarkInContext(true, d.CleanID(idShort))
	if quark.Data() == nil {
		d.ReallyCreateLeaf(location, quark, klimt.DisplayWithNewlines(d.Pragma(), ""), typ, abel.USymbolNull)
	}
	return quark.Data()
}

// getStart returns the start of the current group.
func (d *StateDiagram) getStart(location *text.LineLocation) *abel.Entity {
	return d.pseudoState(location, "*start*", abel.LeafTypeCircleStart)
}

// getEnd returns the end of the current group.
func (d *StateDiagram) getEnd(location *text.LineLocation) *abel.Entity {
	return d.pseudoState(location, "*end*", abel.LeafTypeCircleEnd)
}

// getHistorical returns the history state of the current group.
func (d *StateDiagram) getHistorical(location *text.LineLocation) *abel.Entity {
	return d.pseudoState(location, "*historical*", abel.LeafTypePseudoState)
}

// getDeepHistory returns the deep history state of the current group.
func (d *StateDiagram) getDeepHistory(location *text.LineLocation) *abel.Entity {
	return d.pseudoState(location, "*deephistory*", abel.LeafTypeDeepHistory)
}

// getHistoricalOf returns the history state of the state idShort,
// creating the state as a composite state when needed.
func (d *StateDiagram) getHistoricalOf(location *text.LineLocation, idShort string) *abel.Entity {
	quark := d.QuarkInContext(true, d.CleanID(idShort))
	d.GotoGroup(location, quark, quarkDisplay(quark), abel.GroupTypeState, abel.USymbolNull)
	ident := d.QuarkInContext(true, "*historical*"+d.CurrentGroup().Name())
	result := ident.Data()
	if result == nil {
		result = d.ReallyCreateLeaf(location, ident, klimt.DisplayWithNewlines(d.Pragma(), ""),
			abel.LeafTypePseudoState, abel.USymbolNull)
	}
	d.EndGroup()
	return result
}

// getDeepHistoryOf returns the deep history state of the state idShort,
// creating the state as a composite state when needed.
func (d *StateDiagram) getDeepHistoryOf(location *text.LineLocation, idShort string) *abel.Entity {
	quark := d.QuarkInContext(true, d.CleanID(idShort))
	d.GotoGroup(location, quark, quarkDisplay(quark), abel.GroupTypeState, abel.USymbolNull)
	ident := d.QuarkInContext(true, d.CleanID("*deephistory*"+d.CurrentGroup().Name()))
	result := ident.Data()
	if result == nil {
		result = d.ReallyCreateLeaf(location, ident, klimt.DisplayWithNewlines(d.Pragma(), ""),
			abel.LeafTypeDeepHistory, abel.USymbolNull)
	}
	d.EndGroup()
	return result
}

// quarkDisplay is upstream's Display.getWithNewlines(quark): the name of
// the quark, read without the pragmas of the diagram.
func quarkDisplay(quark *abel.Quark) *klimt.Display {
	return klimt.DisplayWithNewlines(klimt.PragmaCreateEmpty(), quark.Name())
}

// concurrentState starts a new concurrent region in the current state,
// separated by direction ('-' or '|'), closing the previous region.
func (d *StateDiagram) concurrentState(location *text.LineLocation, direction rune) bool {
	cur := d.CurrentGroup()
	cur.SetConcurrentSeparator(direction)
	if cur.GroupType() == abel.GroupTypeConcurrentState {
		d.CucaDiagram.EndGroup()
	}
	ident := d.QuarkInContext(true, d.CleanID(d.UniqueSequence2(concurrentPrefix)))
	d.GotoGroup(location, ident, klimt.DisplayCreate(""), abel.GroupTypeConcurrentState, abel.USymbolNull)
	d.CurrentGroup().SetConcurrentSeparator(direction)
	return true
}

// EndGroup closes the current group, and the state of the current
// concurrent region with it.
func (d *StateDiagram) EndGroup() bool {
	if d.CurrentGroup().GroupType() == abel.GroupTypeConcurrentState {
		d.CucaDiagram.EndGroup()
	}
	return d.CucaDiagram.EndGroup()
}

// CheckFinalError implements command.Diagram: a state of a concurrent
// region cannot be linked out of it.
func (d *StateDiagram) CheckFinalError() string {
	for _, link := range d.Links() {
		parent1 := concurrentParent(link.Entity1())
		parent2 := concurrentParent(link.Entity2())
		if parent1 != parent2 {
			return "State within concurrent state cannot be linked out of this concurrent state (between " +
				link.Entity1().Name() + " and " + link.Entity2().Name() + ")"
		}
	}
	return ""
}

// concurrentParent returns the innermost concurrent region holding ent,
// nil for none.
func concurrentParent(ent *abel.Entity) *abel.Entity {
	for parent := ent.ParentContainer(); parent != nil; parent = parent.ParentContainer() {
		if parent.GroupType() == abel.GroupTypeConcurrentState {
			return parent
		}
	}
	return nil
}

// ensureParentState creates the missing states of the parents of a
// dotted name ("A.x" creates the composite state A).
func (d *StateDiagram) ensureParentState(location *text.LineLocation, current *abel.Quark) {
	for current != nil {
		parent := current.Parent()
		if parent == nil || parent.Data() != nil {
			return
		}
		parentGroup := d.CreateGroup(location, parent, abel.GroupTypeState)
		parentGroup.SetDisplay(quarkDisplay(parent))
		current = parent
	}
}
