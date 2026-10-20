// SPDX-License-Identifier: MIT
// Provenance: original

package statediagram

import (
	"testing"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
)

func newDiagram() *StateDiagram {
	lines := []text.StringLocated{{S: "@startuml"}, {S: "@enduml"}}
	return NewStateDiagram(core.NewRenderContext(), core.NewUmlSource(lines, lines, true), nil)
}

var here = &text.LineLocation{}

// gotoState opens the composite state code.
func gotoState(d *StateDiagram, code string) {
	quark := d.QuarkInContext(true, d.CleanID(code))
	d.GotoGroup(here, quark, quarkDisplay(quark), abel.GroupTypeState, abel.USymbolNull)
}

// TestStartAndEndPerGroup checks that the root and each composite state
// have their own start and end, created once.
func TestStartAndEndPerGroup(t *testing.T) {
	d := newDiagram()
	rootStart, rootEnd := d.getStart(here), d.getEnd(here)
	if rootStart == rootEnd || rootStart.LeafType() != abel.LeafTypeCircleStart ||
		rootEnd.LeafType() != abel.LeafTypeCircleEnd {
		t.Fatalf("root start %v, end %v", rootStart.LeafType(), rootEnd.LeafType())
	}
	if d.getStart(here) != rootStart {
		t.Error("the root start is created twice")
	}
	gotoState(d, "Active")
	start, end := d.getStart(here), d.getEnd(here)
	if start == rootStart || end == rootEnd {
		t.Error("the composite state shares the root start or end")
	}
	if start.ParentContainer().Name() != "Active" || start.Name() != "*start*Active" {
		t.Errorf("start %q in %q", start.Name(), start.ParentContainer().Name())
	}
	d.EndGroup()
	if !d.CurrentGroup().IsRoot() {
		t.Error("the composite state is still open")
	}
}

// TestHistoricalOfNamedState checks that "X[H]" creates the history state
// inside X, and "X[H*]" the deep history state, from outside of X.
func TestHistoricalOfNamedState(t *testing.T) {
	d := newDiagram()
	h := d.getHistoricalOf(here, "Active")
	if h.LeafType() != abel.LeafTypePseudoState || h.ParentContainer().Name() != "Active" {
		t.Errorf("history %v in %q", h.LeafType(), h.ParentContainer().Name())
	}
	if d.getHistoricalOf(here, "Active") != h {
		t.Error("the history state is created twice")
	}
	deep := d.getDeepHistoryOf(here, "Active")
	if deep.LeafType() != abel.LeafTypeDeepHistory || deep.ParentContainer() != h.ParentContainer() {
		t.Errorf("deep history %v in %q", deep.LeafType(), deep.ParentContainer().Name())
	}
	if active := h.ParentContainer(); active.GroupType() != abel.GroupTypeState {
		t.Errorf("Active is a %v", active.GroupType())
	}
	if !d.CurrentGroup().IsRoot() {
		t.Error("Active is still open")
	}
}

// TestConcurrentRegions checks that each separator opens a new region in
// the current state, closing the previous one, and that ending the state
// closes the last region with it.
func TestConcurrentRegions(t *testing.T) {
	d := newDiagram()
	gotoState(d, "Active")
	active := d.CurrentGroup()
	d.concurrentState(here, '-')
	first := d.CurrentGroup()
	d.concurrentState(here, '-')
	second := d.CurrentGroup()
	if first == second || first.GroupType() != abel.GroupTypeConcurrentState ||
		second.ParentContainer() != active || first.ParentContainer() != active {
		t.Fatalf("regions %q and %q", first.Name(), second.Name())
	}
	if active.ConcurrentSeparator() != '-' || second.ConcurrentSeparator() != '-' {
		t.Errorf("separators %q and %q", active.ConcurrentSeparator(), second.ConcurrentSeparator())
	}
	d.EndGroup()
	if !d.CurrentGroup().IsRoot() {
		t.Errorf("current group %q", d.CurrentGroup().Name())
	}
}

// TestTransitionOutOfRegion checks upstream's final error for a link from
// a state of a concurrent region to a state outside of it.
func TestTransitionOutOfRegion(t *testing.T) {
	d := newDiagram()
	gotoState(d, "Active")
	d.concurrentState(here, '-')
	b := d.ReallyCreateLeaf(here, d.QuarkInContext(true, "B"), klimt.DisplayCreate("B"), abel.LeafTypeState,
		abel.USymbolNull)
	if !d.checkConcurrentStateOk(b.Quark()) {
		t.Error("B cannot be used in its own region")
	}
	d.EndGroup()
	if d.checkConcurrentStateOk(b.Quark()) {
		t.Error("B can be used outside of its region")
	}
	other := d.ReallyCreateLeaf(here, d.QuarkInContext(true, "Other"), klimt.DisplayCreate("Other"),
		abel.LeafTypeState, abel.USymbolNull)
	if msg := d.CheckFinalError(); msg != "" {
		t.Fatalf("error without link: %q", msg)
	}
	d.AddLink(abel.NewLink(here, d.Cuca(), d.SkinParam().CurrentStyleBuilder(), b, other,
		decoration.NewLinkType(decoration.LinkDecorArrow, decoration.LinkDecorNone), abel.LinkArgNoDisplay(2)))
	want := "State within concurrent state cannot be linked out of this concurrent state (between B and Other)"
	if msg := d.CheckFinalError(); msg != want {
		t.Errorf("error %q", msg)
	}
}
