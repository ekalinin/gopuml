// SPDX-License-Identifier: MIT
// Provenance: original

package statediagram

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/uml"
)

func linkCommands() []cmd {
	return uml.AddCommonCommands1([]cmd{commandCreateState(), commandLinkState(), commandLinkStateReverse()})
}

// links parses body and returns the links of the diagram.
func links(t *testing.T, body ...string) (*StateDiagram, []*abel.Link) {
	t.Helper()
	d := mustParse(t, linkCommands, body...)
	return d, d.Links()
}

func ends(l *abel.Link) string { return l.Entity1().Name() + "->" + l.Entity2().Name() }

// TestLinkStartAndEnd checks that "[*]" is the start on the left of an
// arrow and the end on the right.
func TestLinkStartAndEnd(t *testing.T) {
	_, ls := links(t, "[*] --> Idle", "Idle --> [*]")
	if len(ls) != 2 || ls[0].Entity1().LeafType() != abel.LeafTypeCircleStart ||
		ls[1].Entity2().LeafType() != abel.LeafTypeCircleEnd || ls[0].Length() != 2 {
		t.Fatalf("links %v", ls)
	}
}

// TestLinkHistoryAndBars checks the history states of the current group
// and of a named state, and the synchronization bars.
func TestLinkHistoryAndBars(t *testing.T) {
	_, ls := links(t,
		"state Active",
		"A --> [H]",
		"B --> [H*]",
		"C --> Active[H]",
		"D --> Active[H*]",
		"E --> ==sync==",
		"==sync== --> F")
	want := []struct {
		typ    abel.LeafType
		parent string
	}{
		{abel.LeafTypePseudoState, ""}, {abel.LeafTypeDeepHistory, ""},
		{abel.LeafTypePseudoState, "Active"}, {abel.LeafTypeDeepHistory, "Active"},
		{abel.LeafTypeSynchroBar, ""},
	}
	for i, w := range want {
		e := ls[i].Entity2()
		parent := ""
		if p := e.ParentContainer(); p != nil && !p.IsRoot() {
			parent = p.Name()
		}
		if e.LeafType() != w.typ || parent != w.parent {
			t.Errorf("link %d: %v in %q", i, e.LeafType(), parent)
		}
	}
	if ls[4].Entity2() != ls[5].Entity1() {
		t.Error("the bar is created twice")
	}
}

// TestLinkDirectionStyleAndLabel checks direction hints, bracket styles
// and colors, labels and stereotypes.
func TestLinkDirectionStyleAndLabel(t *testing.T) {
	_, ls := links(t,
		"state A",
		"A -left-> B : back",
		"A -[#blue,dashed]-> C",
		"A ---> D <<important>> : far")
	if ends(ls[0]) != "B->A" || !ls[0].IsInverted() || ls[0].Length() != 1 {
		t.Errorf("left link %s inverted %v length %d", ends(ls[0]), ls[0].IsInverted(), ls[0].Length())
	}
	if ls[0].Label().Get(0).String() != "back" {
		t.Errorf("label %v", ls[0].Label())
	}
	if ls[1].Type().Style() != decoration.LinkStyleDashed() || ls[1].Colors().Color(color.TypeLine) == nil {
		t.Errorf("styled link %v %v", ls[1].Type().Style(), ls[1].Colors())
	}
	if ls[2].Length() != 3 || ls[2].Stereotype() == nil {
		t.Errorf("long link length %d stereotype %v", ls[2].Length(), ls[2].Stereotype())
	}
}

// TestLinkDecorations checks the "x" start and the "o" end.
func TestLinkDecorations(t *testing.T) {
	_, ls := links(t, "state A", "A x--> B", "A -->o C")
	if ls[0].Type().Decor2() != decoration.LinkDecorCircleCross || ls[0].Type().Decor1() != decoration.LinkDecorArrow {
		t.Errorf("cross start %v %v", ls[0].Type().Decor1(), ls[0].Type().Decor2())
	}
	if ls[1].Type().Decor1() != decoration.LinkDecorArrowAndCircle {
		t.Errorf("circle end %v", ls[1].Type().Decor1())
	}
}

// TestLinkReversed checks that "B <-- A" goes from A to B, to the left
// when no direction is written.
func TestLinkReversed(t *testing.T) {
	_, ls := links(t, "state A", "state B", "B <-- A : back", "C <-up- A")
	if ends(ls[0]) != "B->A" || !ls[0].IsInverted() || ls[0].Length() != 1 ||
		ls[0].Label().Get(0).String() != "back" {
		t.Errorf("reversed link %s inverted %v length %d", ends(ls[0]), ls[0].IsInverted(), ls[0].Length())
	}
	if ends(ls[1]) != "C->A" || !ls[1].IsInverted() || ls[1].Length() != 2 {
		t.Errorf("reversed up link %s inverted %v length %d", ends(ls[1]), ls[1].IsInverted(), ls[1].Length())
	}
}

// TestLinkUnusableState checks the error for a state of a concurrent
// region used outside of it.
func TestLinkUnusableState(t *testing.T) {
	commands := func() []cmd {
		return []cmd{commandCreateState(), commandLinkState(), commandCreatePackageState(), commandEndState(),
			commandConcurrentState()}
	}
	msg := parseError(t, commands, "state Active {", "state A", "--", "state B", "}", "state Other", "B --> Other")
	if !strings.HasPrefix(msg, "The state B cannot be used here.") {
		t.Errorf("error %q", msg)
	}
}

// TestLinkLabelNode checks that "-[node]->" and the skin parameter draw
// the label as a node, with deterministic names.
func TestLinkLabelNode(t *testing.T) {
	for _, body := range [][]string{
		{"state A", "A -[node]-> B : go"},
		{"skinparam stateDiagramEdgeLabelStyle node", "state A", "A --> B : go"},
	} {
		d, ls := links(t, body...)
		if len(ls) != 2 {
			t.Fatalf("%q: %d links", body, len(ls))
		}
		node := ls[0].Entity2()
		if node.LeafType() != abel.LeafTypeStateTransitionLabel || ls[1].Entity1() != node ||
			node.Name() != "transition_A_B__go__1" {
			t.Errorf("%q: node %q %v", body, node.Name(), node.LeafType())
		}
		if label, _ := node.Stereotype().Label(klimt.GuillemetDoubleComparator()); label != "<<transition>>" {
			t.Errorf("%q: stereotype %q", body, label)
		}
		if ls[0].Type().Decor1() != decoration.LinkDecorNone || ls[1].Type().Decor1() != decoration.LinkDecorArrow ||
			ls[0].Weight() != 10 || ls[1].Weight() != 10 {
			t.Errorf("%q: links %v %v", body, ls[0].Type(), ls[1].Type())
		}
		if d.transitionNodes != 1 {
			t.Errorf("%q: %d label nodes", body, d.transitionNodes)
		}
	}
}
