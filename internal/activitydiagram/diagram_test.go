// SPDX-License-Identifier: MIT
// Provenance: original

package activitydiagram

import (
	"testing"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
)

func newDiagram() *ActivityDiagram {
	lines := []text.StringLocated{{S: "@startuml"}, {S: "@enduml"}}
	return NewActivityDiagram(core.NewRenderContext(), core.NewUmlSource(lines, lines, true), nil)
}

var here = &text.LineLocation{}

// TestStartAndEnds checks that the start is created once, and that the
// ends with and without a suffix are different entities.
func TestStartAndEnds(t *testing.T) {
	d := newDiagram()
	start := d.getStart(here)
	if start.LeafType() != abel.LeafTypeCircleStart || d.getStart(here) != start {
		t.Fatalf("start %v", start.LeafType())
	}
	end, endTop, end1 := d.getEnd(here, "", false), d.getEnd(here, "top", true), d.getEnd(here, "1", true)
	if end == endTop || end == end1 || endTop == end1 || end.LeafType() != abel.LeafTypeCircleEnd {
		t.Errorf("ends %q %q %q", end.Name(), endTop.Name(), end1.Name())
	}
	if d.getEnd(here, "top", true) != endTop {
		t.Error("the end of (*top) is created twice")
	}
	if d.lastEntity() != end1 {
		t.Errorf("last entity %q", d.lastEntity().Name())
	}
}

// TestBranchesNest checks that branches open nested contexts, closed in
// order, and that the branch is the last entity.
func TestBranchesNest(t *testing.T) {
	d := newDiagram()
	d.startIf(here, "", false)
	outer := d.currentContext.Branch()
	d.startIf(here, "inner", true)
	inner := d.currentContext.Branch()
	if outer == inner || inner.Name() != "inner" || d.currentContext.Parent().Branch() != outer {
		t.Fatalf("branches %q and %q", outer.Name(), inner.Name())
	}
	if d.lastEntity() != inner || d.lastEntityBrancheConsulted != inner {
		t.Error("the inner branch is not the last entity")
	}
	d.endif()
	if d.currentContext.Branch() != outer {
		t.Error("endif does not return to the outer branch")
	}
	d.endif()
	if d.currentContext != nil {
		t.Error("the outer context is still open")
	}
}

// TestNoteKeepsLastEntity checks that a note does not become the last
// activity.
func TestNoteKeepsLastEntity(t *testing.T) {
	d := newDiagram()
	start := d.getStart(here)
	d.createNote(here, d.QuarkInContext(true, "GN1"), klimt.DisplayCreate("text"))
	if d.lastEntity() != start {
		t.Errorf("last entity %q", d.lastEntity().Name())
	}
	if got := d.Description(); got != "(2 activities)" {
		t.Errorf("description %q", got)
	}
}
