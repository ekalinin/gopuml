// SPDX-License-Identifier: MIT
// Provenance: original

package statediagram

import (
	"testing"

	"github.com/ekalinin/gopuml/internal/abel"
)

// TestPassesTransitionBeforeDeclaration checks that a transition written
// before the composite state holding one of its states links that state,
// because transitions run in the second pass.
func TestPassesTransitionBeforeDeclaration(t *testing.T) {
	d := mustParse(t, commands, "[*] --> B", "state Active {", "state B", "}")
	b := entity(t, d, "B")
	if b.ParentContainer().Name() != "Active" {
		t.Fatalf("B in %q", b.ParentContainer().Name())
	}
	if ls := d.Links(); len(ls) != 1 || ls[0].Entity2() != b {
		t.Errorf("links %v", ls)
	}
	count := 0
	for _, q := range d.Quarks() {
		if q.Name() == "B" && q.Data() != nil {
			count++
		}
	}
	if count != 1 {
		t.Errorf("%d states named B", count)
	}
}

// TestPassesNoteBeforeDeclaration checks that a note on a state written
// before the state is attached to the declared state, because notes on
// states run in the third pass.
func TestPassesNoteBeforeDeclaration(t *testing.T) {
	d := mustParse(t, commands, "note right of Idle : waiting", "state Idle : busy")
	idle := entity(t, d, "Idle")
	if idle.LeafType() != abel.LeafTypeState || len(idle.Bodier().RawBody()) != 1 {
		t.Fatalf("Idle: %v", idle.LeafType())
	}
	var note *abel.Entity
	for _, leaf := range d.Leafs() {
		if leaf.LeafType() == abel.LeafTypeNote {
			note = leaf
		}
	}
	if note == nil || displayOf(note) != "waiting" {
		t.Fatal("no note")
	}
	if ls := d.Links(); len(ls) != 1 || (ls[0].Entity1() != idle && ls[0].Entity2() != idle) {
		t.Errorf("links %v", ls)
	}
}

// TestPassesNoteOnTransition checks that a note on a transition is
// attached to it, in the second pass after the transition.
func TestPassesNoteOnTransition(t *testing.T) {
	d := mustParse(t, commands, "state A", "A --> B", "note on link : moving")
	ls := d.Links()
	if len(ls) != 1 || ls[0].Note() == nil {
		t.Fatalf("links %v", ls)
	}
}
