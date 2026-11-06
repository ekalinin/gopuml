// SPDX-License-Identifier: MIT
// Provenance: original

package core

import (
	"slices"
	"testing"
)

func TestRemoveInitialNoise(t *testing.T) {
	src := NewUmlSource(srcLines("@startdot\nskinparam monochrome true\n\n!pragma layout smetana\ndigraph G {\n}\n@enddot\n"), nil, false)
	got := src.RemoveInitialNoise()
	var lines []string
	var positions []int
	for _, l := range got.Lines() {
		lines = append(lines, l.S)
		positions = append(positions, l.Location.Position)
	}
	if want := []string{"@startdot", "digraph G {", "}", "@enddot"}; !slices.Equal(lines, want) {
		t.Errorf("lines = %q, want %q", lines, want)
	}
	if want := []int{0, 4, 5, 6}; !slices.Equal(positions, want) {
		t.Errorf("positions = %v, want %v", positions, want)
	}

	for _, s := range []string{"@startdot\nskinparamfoo\n@enddot\n", "@startdot\ndigraph G {\n\n}\n@enddot\n"} {
		src := NewUmlSource(srcLines(s), nil, false)
		if got := src.RemoveInitialNoise(); got != src {
			t.Errorf("%q: the source changed to %v", s, got.Lines())
		}
	}
}

func TestIsStartDef(t *testing.T) {
	for s, want := range map[string]bool{"@startdef\n@enddef\n": true, "@startdot\n@enddot\n": false, "@startuml\n@enduml\n": false} {
		if got := NewUmlSource(srcLines(s), nil, false).IsStartDef(); got != want {
			t.Errorf("IsStartDef of %q = %v, want %v", s, got, want)
		}
	}
}
