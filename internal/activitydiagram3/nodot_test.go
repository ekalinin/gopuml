// SPDX-License-Identifier: MIT
// Provenance: original

package activitydiagram3

import (
	"strings"
	"testing"
)

// TestActivityRendersWithoutDot checks that activity diagrams draw without
// a dot executable: the ftile layout never runs Graphviz.
func TestActivityRendersWithoutDot(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	t.Setenv("GRAPHVIZ_DOT", "/nonexistent/dot")
	svg, err := renderSVG("@startuml\nstart\n:a;\nif (c) then (yes)\n:b;\nelse (no)\n:c;\nendif\n" +
		"while (more?)\n:d;\nendwhile\nfork\n:e;\nfork again\n:f;\nend fork\nstop\n@enduml\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(svg, "<svg") || !strings.Contains(svg, ">a</text>") {
		t.Errorf("no drawing: %.300s", svg)
	}
}
