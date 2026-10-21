// SPDX-License-Identifier: MIT
// Provenance: original

package activitydiagram

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// parseWith parses body (the lines between @startuml and @enduml) with a
// factory of the given commands only.
func parseWith(commands func() []cmd, body ...string) core.Diagram {
	lines := []text.StringLocated{{S: "@startuml"}}
	for _, s := range body {
		lines = append(lines, text.StringLocated{S: s})
	}
	lines = append(lines, text.StringLocated{S: "@enduml"})
	return uml.NewFactory(core.Activity, commands, NewActivityDiagram).
		CreateSystem(core.NewRenderContext(), core.NewUmlSource(lines, lines, true))
}

func mustParse(t *testing.T, commands func() []cmd, body ...string) *ActivityDiagram {
	t.Helper()
	result := parseWith(commands, body...)
	d, ok := result.(*ActivityDiagram)
	if !ok {
		if e, isError := result.(*core.ErrorDiagram); isError {
			t.Fatalf("%q: error at line %d: %s", body, e.Position(), e.Error.Message())
		}
		t.Fatalf("result %T", result)
	}
	return d
}

func parseError(t *testing.T, commands func() []cmd, body ...string) string {
	t.Helper()
	e, ok := parseWith(commands, body...).(*core.ErrorDiagram)
	if !ok {
		t.Fatalf("no error for %q", body)
	}
	return e.Error.Message()
}

func displayOf(e *abel.Entity) string {
	var lines []string
	for i := range e.Display().Size() {
		lines = append(lines, e.Display().Get(i).String())
	}
	return strings.Join(lines, "\n")
}

func ends(l *abel.Link) string { return l.Entity1().Name() + "->" + l.Entity2().Name() }

func linkCommands() []cmd { return []cmd{commandLinkLongActivity(), commandLinkActivity()} }

// TestLinkStartAndEnds checks the start and the ends (*), (*top) and (*1).
func TestLinkStartAndEnds(t *testing.T) {
	d := mustParse(t, linkCommands, `(*) --> "A"`, `"A" --> (*)`, `"A" --> (*top)`, `"A" --> (*1)`, `(*top) --> "B"`)
	ls := d.Links()
	if ls[0].Entity1().LeafType() != abel.LeafTypeCircleStart || ls[4].Entity1() != ls[0].Entity1() {
		t.Error("the start is not shared")
	}
	if ls[1].Entity2() == ls[2].Entity2() || ls[2].Entity2() == ls[3].Entity2() ||
		ls[3].Entity2().LeafType() != abel.LeafTypeCircleEnd {
		t.Errorf("ends %s %s %s", ends(ls[1]), ends(ls[2]), ends(ls[3]))
	}
}

// TestLinkActivities checks quoted activities with and without "as", bare
// names, bars and the omitted source.
func TestLinkActivities(t *testing.T) {
	d := mustParse(t, linkCommands,
		`(*) --> "Long name" as L`,
		`L --> Bare`,
		`--> ===B1===`,
		`===B1=== --> "Quoted"`,
		`"Quoted" --> (*)`)
	ls := d.Links()
	want := []string{"start->L", "L->Bare", "Bare->B1", "B1->Quoted", "Quoted->end"}
	for i, w := range want {
		if ends(ls[i]) != w {
			t.Errorf("link %d: %s, want %s", i, ends(ls[i]), w)
		}
	}
	if displayOf(ls[0].Entity2()) != "Long name" || ls[0].Entity2().LeafType() != abel.LeafTypeActivity {
		t.Errorf("L: %q %v", displayOf(ls[0].Entity2()), ls[0].Entity2().LeafType())
	}
	if ls[2].Entity2().LeafType() != abel.LeafTypeSynchroBar {
		t.Errorf("B1: %v", ls[2].Entity2().LeafType())
	}
}

// TestLinkArrows checks labels, dotted bodies, direction hints, bracket
// styles, "-*->" and URLs.
func TestLinkArrows(t *testing.T) {
	d := mustParse(t, linkCommands,
		`"A" -->[yes] "B"`,
		`"B" ..> "C"`,
		`"C" -left-> "D"`,
		`"D" -[#red,dashed]-> "E"`,
		`"E" -*-> "F"`,
		`"F" [[http://example.com]]--> "G"`,
		`"G" ---> "H"`)
	ls := d.Links()
	if ls[0].Label().Get(0).String() != "yes" || ls[0].Length() != 2 {
		t.Errorf("label link: %v length %d", ls[0].Label(), ls[0].Length())
	}
	if ls[1].Type().Style() != decoration.NewLinkType(decoration.LinkDecorArrow, decoration.LinkDecorNone).GoDotted().Style() {
		t.Errorf("dotted link: %v", ls[1].Type().Style())
	}
	if ends(ls[2]) != "D->C" || !ls[2].IsInverted() || ls[2].Length() != 1 {
		t.Errorf("left link: %s inverted %v length %d", ends(ls[2]), ls[2].IsInverted(), ls[2].Length())
	}
	if ls[3].Colors().Color(color.TypeLine) == nil {
		t.Error("styled link has no color")
	}
	if ls[4].IsConstraint() || ls[4].Length() != 2 {
		t.Errorf("-*-> link: constraint %v length %d", ls[4].IsConstraint(), ls[4].Length())
	}
	if ls[5].URL() == nil || ls[5].URL().URL() != "http://example.com" {
		t.Errorf("url link: %v", ls[5].URL())
	}
	if ls[6].Length() != 3 {
		t.Errorf("long link length %d", ls[6].Length())
	}
}

// TestLinkStereotypesColorsAndPartition checks stereotypes and colors of
// both ends, and "in Partition".
func TestLinkStereotypesColorsAndPartition(t *testing.T) {
	d := mustParse(t, linkCommands, `"A" <<s1>> #pink --> "B" <<s2>> in "My partition" #blue`)
	l := d.Links()[0]
	a, b := l.Entity1(), l.Entity2()
	if a.Stereotype() == nil || b.Stereotype() == nil {
		t.Error("missing stereotype")
	}
	if a.Colors().Color(color.TypeBack) == nil || b.Colors().Color(color.TypeBack) == nil {
		t.Error("missing background color")
	}
	if p := b.ParentContainer(); p.Name() != "My partition" || p.GroupType() != abel.GroupTypePackage {
		t.Errorf("B in %q", p.Name())
	}
	if !a.ParentContainer().IsRoot() || !d.CurrentGroup().IsRoot() {
		t.Error("the partition is still open")
	}
}

// TestLinkInnerActivity checks that "--> {" opens an inner activity.
func TestLinkInnerActivity(t *testing.T) {
	d := mustParse(t, linkCommands, `(*) --> "A"`, `--> {`)
	if g := d.CurrentGroup(); g.GroupType() != abel.GroupTypeInnerActivity || d.lastEntity() != g {
		t.Errorf("current group %v", g.GroupType())
	}
}

// TestLinkLongActivity checks an activity on several lines, with a code, a
// stereotype, a partition and a color on its last line.
func TestLinkLongActivity(t *testing.T) {
	d := mustParse(t, linkCommands, `(*) -->[go] "First line`, `second line`, `third line" as T <<s>> in P #red`, `T --> (*)`)
	ls := d.Links()
	target := ls[0].Entity2()
	if target.Name() != "T" || displayOf(target) != "First line\nsecond line\nthird line" {
		t.Errorf("target %q %q", target.Name(), displayOf(target))
	}
	if target.Stereotype() == nil || target.Colors().Color(color.TypeBack) == nil || target.ParentContainer().Name() != "P" {
		t.Error("missing stereotype, color or partition")
	}
	if ls[0].Label().Get(0).String() != "go" || ends(ls[1]) != "T->end" {
		t.Errorf("links %s %s", ends(ls[0]), ends(ls[1]))
	}
}

// TestLinkWithoutSource checks the error of a transition without source
// at the start of the diagram.
func TestLinkWithoutSource(t *testing.T) {
	if msg := parseError(t, linkCommands, `--> "A"`); !strings.HasPrefix(msg, "No such activity") {
		t.Errorf("error %q", msg)
	}
}
