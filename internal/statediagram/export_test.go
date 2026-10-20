// SPDX-License-Identifier: MIT
// Provenance: original

package statediagram_test

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/dot"
	"github.com/ekalinin/gopuml/internal/engine"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/statediagram"
)

const states = "@startuml\n[*] --> Idle\nIdle --> Running : start\nRunning --> [*]\n@enduml\n"

func parseBlock(t *testing.T, rc *core.RenderContext, src string) *engine.Block {
	t.Helper()
	blocks, err := engine.Parse(context.Background(), rc, engine.Source{Text: src})
	if err != nil || len(blocks) != 1 {
		t.Fatalf("%d blocks, error %v", len(blocks), err)
	}
	if _, ok := blocks[0].Diagram.(*statediagram.StateDiagram); !ok {
		t.Fatalf("diagram %T", blocks[0].Diagram)
	}
	return blocks[0]
}

// TestStateTxtUnsupported checks the state-diagram scenario "TXT for a
// state diagram".
func TestStateTxtUnsupported(t *testing.T) {
	block := parseBlock(t, core.NewRenderContext(), states)
	for _, format := range []font.FileFormat{font.FormatATXT, font.FormatUTXT} {
		var buf bytes.Buffer
		_, err := block.Render(context.Background(), &buf, 0, engine.Options{Format: format})
		var unsupported *engine.UnsupportedFormatError
		if !errors.As(err, &unsupported) || !strings.HasSuffix(err.Error(), " output is not supported for state diagrams") {
			t.Errorf("%v: error %v", format, err)
		}
		if buf.Len() != 0 {
			t.Errorf("%v: %d bytes written", format, buf.Len())
		}
	}
}

// TestStateWithoutDot checks the state-diagram scenario "State diagram
// without dot", with composite states, which are not laid out on their own
// without dot.
func TestStateWithoutDot(t *testing.T) {
	rc := core.NewRenderContext()
	rc.DotPath = "/nonexistent/graphviz/dot"
	block := parseBlock(t, rc, "@startuml\nstate Active {\n[*] --> Running\n}\n[*] --> Active\n@enduml\n")
	var buf bytes.Buffer
	_, err := block.Render(context.Background(), &buf, 0, engine.Options{Format: font.FormatSVG})
	if !errors.Is(err, dot.ErrNotFound) {
		t.Fatalf("error %v, want dot.ErrNotFound", err)
	}
	for _, want := range []string{"Dot Executable: /nonexistent/graphviz/dot", "Cannot find Graphviz. You should try"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("the image does not say %q", want)
		}
	}
}

// TestTransitionLabelNode checks the state-diagram scenario "Transition
// label as a node": the label is a node between the states, named the same
// way on every parse, which upstream's names with the current time are
// not, and two renders give the same SVG.
func TestTransitionLabelNode(t *testing.T) {
	src := "@startuml\nskinparam stateDiagramEdgeLabelStyle node\nstate A\nA --> B : go\n@enduml\n"
	for range 2 {
		d := parseBlock(t, core.NewRenderContext(), src).Diagram.(*statediagram.StateDiagram)
		var node *abel.Entity
		for _, leaf := range d.Leafs() {
			if leaf.LeafType() == abel.LeafTypeStateTransitionLabel {
				node = leaf
			}
		}
		if node == nil || node.Name() != "transition_A_B__go__1" {
			t.Fatalf("label node %v", node)
		}
		if links := d.Links(); len(links) != 2 || links[0].Entity2() != node || links[1].Entity1() != node {
			t.Fatalf("links %v", links)
		}
	}
	if _, err := exec.LookPath("dot"); err != nil {
		t.Skip("dot is not installed: the renders are not compared")
	}
	render := func() string {
		var buf bytes.Buffer
		block := parseBlock(t, core.NewRenderContext(), src)
		if _, err := block.Render(context.Background(), &buf, 0, engine.Options{Format: font.FormatSVG}); err != nil {
			t.Fatal(err)
		}
		return buf.String()
	}
	first, second := render(), render()
	if first != second {
		t.Error("two renders differ")
	}
	for _, want := range []string{"<!--link A to transition_A_B__go__1-->", "<!--link transition_A_B__go__1 to B-->",
		">go</text>"} {
		if !strings.Contains(first, want) {
			t.Errorf("the SVG has no %q", want)
		}
	}
}
