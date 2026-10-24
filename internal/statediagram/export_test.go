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

// labelNodes returns the label nodes of the transitions, by label.
func labelNodes(d *statediagram.StateDiagram) map[string]*abel.Entity {
	result := map[string]*abel.Entity{}
	for _, leaf := range d.Leafs() {
		if leaf.LeafType() == abel.LeafTypeStateTransitionLabel {
			result[leaf.Display().Get(0).String()] = leaf
		}
	}
	return result
}

// TestOrphanedTransitionLabel checks that the label node of a transition is
// hidden or removed with a state at either end, and only then.
func TestOrphanedTransitionLabel(t *testing.T) {
	const states = "state Idle $idle\nstate Charging $charging\nstate Faulted $fault\n" +
		"Idle --> Charging : start\nCharging --> Faulted : fault\n"
	cases := []struct {
		rule            string
		removed, hidden []string
		visible         []string
	}{
		{"remove $fault", []string{"fault"}, []string{"fault"}, []string{"start"}},
		{"hide $fault", []string{"fault"}, []string{"fault"}, []string{"start"}},
		{"remove $idle", []string{"start"}, []string{"start"}, []string{"fault"}},
		{"remove $charging", []string{"start", "fault"}, []string{"start", "fault"}, nil},
		{"remove $fault\nrestore $fault", nil, nil, []string{"start", "fault"}},
	}
	for _, c := range cases {
		src := "@startuml\nskinparam stateDiagramEdgeLabelStyle node\n" + states + c.rule + "\n@enduml\n"
		nodes := labelNodes(parseBlock(t, core.NewRenderContext(), src).Diagram.(*statediagram.StateDiagram))
		for _, label := range c.removed {
			if !nodes[label].IsRemoved() {
				t.Errorf("%q: label %q is not removed", c.rule, label)
			}
		}
		for _, label := range c.hidden {
			if !nodes[label].IsHidden() {
				t.Errorf("%q: label %q is not hidden", c.rule, label)
			}
		}
		for _, label := range c.visible {
			if nodes[label].IsRemoved() || nodes[label].IsHidden() {
				t.Errorf("%q: label %q is removed or hidden", c.rule, label)
			}
		}
	}
}

// TestRemovedStateWithLabelNode checks the state-diagram scenario "Removed
// state with a label node": the label node of a transition to a removed
// state is not drawn.
func TestRemovedStateWithLabelNode(t *testing.T) {
	src := "@startuml\nskinparam stateDiagramEdgeLabelStyle node\nstate Idle\nstate Charging\n" +
		"state Faulted $fault\nIdle --> Charging : start\nCharging --> Faulted : fault\nremove $fault\n@enduml\n"
	nodes := labelNodes(parseBlock(t, core.NewRenderContext(), src).Diagram.(*statediagram.StateDiagram))
	if !nodes["fault"].IsRemoved() || nodes["start"].IsRemoved() {
		t.Fatalf("fault removed %v, start removed %v", nodes["fault"].IsRemoved(), nodes["start"].IsRemoved())
	}
	if _, err := exec.LookPath("dot"); err != nil {
		t.Skip("dot is not installed: the render is not checked")
	}
	var buf bytes.Buffer
	block := parseBlock(t, core.NewRenderContext(), src)
	if _, err := block.Render(context.Background(), &buf, 0, engine.Options{Format: font.FormatSVG}); err != nil {
		t.Fatal(err)
	}
	svg := buf.String()
	for _, absent := range []string{">Faulted</text>", ">fault</text>"} {
		if strings.Contains(svg, absent) {
			t.Errorf("the SVG has %q", absent)
		}
	}
	for _, want := range []string{">Idle</text>", ">Charging</text>", ">start</text>",
		"<!--link Idle to transition_Idle_Charging__start__1-->",
		"<!--link transition_Idle_Charging__start__1 to Charging-->"} {
		if !strings.Contains(svg, want) {
			t.Errorf("the SVG has no %q", want)
		}
	}
}

// TestStyledHistoryStates checks the state-diagram scenario "Styled history
// states": a ".history" style reaches the four forms of history states,
// which the conformance comparison does not tell apart by color.
func TestStyledHistoryStates(t *testing.T) {
	if _, err := exec.LookPath("dot"); err != nil {
		t.Skip("dot is not installed")
	}
	src := "@startuml\n<style>\nstateDiagram {\n  .history {\n    BackGroundColor palegreen\n  }\n}\n</style>\n" +
		"state A\nstate B\nstate hh <<history>>\nstate hd <<history*>>\n[H] --> A\n[H*] --> B\nhh --> A\n" +
		"hd --> B\n@enduml\n"
	var buf bytes.Buffer
	block := parseBlock(t, core.NewRenderContext(), src)
	if _, err := block.Render(context.Background(), &buf, 0, engine.Options{Format: font.FormatSVG}); err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(buf.String(), `<ellipse cx=`); n != 4 {
		t.Fatalf("%d circles, want 4", n)
	}
	if n := strings.Count(buf.String(), `fill="#98FB98"`); n != 4 {
		t.Errorf("%d palegreen circles, want 4", n)
	}
}
