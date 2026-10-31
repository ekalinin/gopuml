// SPDX-License-Identifier: MIT
// Provenance: original

package cheneer

import (
	"slices"
	"testing"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// parseResult parses body (the lines between @startchen and @endchen) with
// the commands of Chen diagrams.
func parseResult(body ...string) core.Diagram {
	lines := []text.StringLocated{{S: "@startchen"}}
	for _, s := range body {
		lines = append(lines, text.StringLocated{S: s})
	}
	lines = append(lines, text.StringLocated{S: "@endchen"})
	return uml.NewFactory(core.ChenEER, commands, NewChenEerDiagram).
		CreateSystem(core.NewRenderContext(), core.NewUmlSource(lines, lines, true))
}

// parse parses body and fails the test on an error diagram.
func parse(t *testing.T, body ...string) *ChenEerDiagram {
	t.Helper()
	result := parseResult(body...)
	d, ok := result.(*ChenEerDiagram)
	if !ok {
		if e, ok := result.(*core.ErrorDiagram); ok {
			t.Fatalf("error at line %d: %s", e.Position(), e.Error.Message())
		}
		t.Fatalf("result %T", result)
	}
	return d
}

// assumed ends the message of every parse error of a Chen diagram.
const assumed = " (Assumed diagram type: chen_eer)"

// parseError parses body and returns the message of its error diagram.
func parseError(t *testing.T, body ...string) string {
	t.Helper()
	e, ok := parseResult(body...).(*core.ErrorDiagram)
	if !ok {
		t.Fatalf("no error for %q", body)
	}
	return e.Error.Message()
}

// leaf returns the entity of a full name.
func leaf(t *testing.T, d *ChenEerDiagram, name string) *abel.Entity {
	t.Helper()
	e := d.QuarkInContext(false, name).Data()
	if e == nil {
		t.Fatalf("no entity %q", name)
	}
	return e
}

func TestFirstTokens(t *testing.T) {
	// Upstream declares "entity" and "relationship" for the entity command,
	// "}" for the end of a block, and any line for the others. The tokens
	// read from the patterns must not reject a line upstream accepts.
	for _, c := range []struct {
		cmd  cmd
		want []string
	}{
		{commandCreateAttribute(), nil},
		{commandAssociate(), nil},
		{commandSimpleSubclass(), nil},
		{commandMultiSubclass(), nil},
	} {
		if got := command.FirstTokensOf(c.cmd); got != nil {
			t.Errorf("%T: first tokens %q, want any line", c.cmd, got)
		}
	}
	if got := command.FirstTokensOf(commandCreateEntity()); got != nil &&
		!(slices.Contains(got, "entity") && slices.Contains(got, "relationship")) {
		t.Errorf("entity: first tokens %q", got)
	}
	if got := command.FirstTokensOf(commandEndGroup()); got != nil && !slices.Contains(got, "}") {
		t.Errorf("end group: first tokens %q", got)
	}
}

func TestEntitiesAndRelationships(t *testing.T) {
	d := parse(t, `entity A {`, `}`, `relationship "x" as R {`, `}`)
	if a := leaf(t, d, "A"); a.LeafType() != abel.LeafTypeChenEntity || a.Display().String() != "[A]" {
		t.Errorf("A: %v %v", a.LeafType(), a.Display())
	}
	if r := leaf(t, d, "R"); r.LeafType() != abel.LeafTypeChenRelationship || r.Display().String() != "[x]" {
		t.Errorf("R: %v %v", r.LeafType(), r.Display())
	}
	if got := d.Description(); got != "(2 entities)" {
		t.Errorf("description %q", got)
	}
	if d.PeekOwner() != nil {
		t.Error("the blocks are closed")
	}
}

func TestEntityStereotypeAndColors(t *testing.T) {
	d := parse(t, `entity CHILD <<weak>> {`, `}`, `entity P #pink;line:red {`, `}`)
	if st := leaf(t, d, "CHILD").Stereotype(); st == nil || st.String() != "<<weak>>" {
		t.Errorf("stereotype %v", st)
	}
	colors := leaf(t, d, "P").Colors()
	if back := colors.Color(color.TypeBack); back == nil || back.AsString() != "#FFC0CB" {
		t.Errorf("background %v", back)
	}
	if line := colors.Color(color.TypeLine); line == nil || line.AsString() != "#FF0000" {
		t.Errorf("line %v", line)
	}
}

func TestEntityDeclaredAgain(t *testing.T) {
	d := parse(t, `entity A {`, `}`, `entity A {`)
	if n := len(d.Leafs()); n != 1 {
		t.Errorf("%d leaves", n)
	}
	if owner := d.PeekOwner(); owner == nil || owner.Name() != "A" {
		t.Errorf("owner %v", owner)
	}
}

func TestEntityErrors(t *testing.T) {
	if got := parseError(t, `entity A {`, `}`, `relationship A {`, `}`); got != "Bad name"+assumed {
		t.Errorf("relationship A: %q", got)
	}
	if got := parseError(t, `entity A {`, `}`, `}`); got != "Unbalanced brackets"+assumed {
		t.Errorf("extra }: %q", got)
	}
}

func TestUnclosedBlock(t *testing.T) {
	d := parse(t, `entity A {`)
	if owner := d.PeekOwner(); owner == nil || owner.Name() != "A" {
		t.Errorf("owner %v", owner)
	}
}

func TestDirectionAndTitle(t *testing.T) {
	d := parse(t, `left to right direction`, `title Movies`, `entity A {`, `}`)
	if got := d.Skin().Rankdir(); got != geom.RankdirLeftToRight {
		t.Errorf("rankdir %v", got)
	}
	if got := d.Title().Display().String(); got != "[Movies]" {
		t.Errorf("title %s", got)
	}
}
