// SPDX-License-Identifier: MIT
// Provenance: original

package descdiagram

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/decoration"
)

func linkCommands() []cmd { return []cmd{commandLinkElement(), commandCreateElementFull()} }

// parseLink parses one link line and returns its link.
func parseLink(t *testing.T, line string) (*DescriptionDiagram, *abel.Link) {
	t.Helper()
	result := parseWith(t, linkCommands, line)
	d, ok := result.(*DescriptionDiagram)
	if !ok {
		if e, ok := result.(*core.ErrorDiagram); ok {
			t.Fatalf("%s: error %s", line, e.Error.Message())
		}
		t.Fatalf("%s: result %T", line, result)
	}
	links := d.Links()
	if len(links) != 1 {
		t.Fatalf("%s: %d links", line, len(links))
	}
	return d, links[0]
}

func TestLinkElementHeadsAndStyles(t *testing.T) {
	// The decorations of the type are (second head, first head); a link
	// to the left or up is inverted, which swaps the ends and the heads.
	for _, c := range []struct {
		line, typ  string
		from, to   string
		length     int
		middle     decoration.LinkMiddleDecor
		label      string
		quantifier [2]string
	}{
		{"A --> B", "ARROW-NORMAL(null)-NONE", "A", "B", 2, decoration.LinkMiddleDecorNone, "", [2]string{}},
		{"A <-- B", "NONE-NORMAL(null)-ARROW", "A", "B", 2, decoration.LinkMiddleDecorNone, "", [2]string{}},
		{"A ---> B", "ARROW-NORMAL(null)-NONE", "A", "B", 3, decoration.LinkMiddleDecorNone, "", [2]string{}},
		{"A ..> B", "ARROW-DASHED(null)-NONE", "A", "B", 2, decoration.LinkMiddleDecorNone, "", [2]string{}},
		{"A ~~> B", "ARROW-DOTTED(null)-NONE", "A", "B", 2, decoration.LinkMiddleDecorNone, "", [2]string{}},
		{"A ==> B", "ARROW-BOLD(null)-NONE", "A", "B", 2, decoration.LinkMiddleDecorNone, "", [2]string{}},
		{"A --|> B", "EXTENDS-NORMAL(null)-NONE", "A", "B", 2, decoration.LinkMiddleDecorNone, "",
			[2]string{}},
		{"A -down-> B", "ARROW-NORMAL(null)-NONE", "A", "B", 2, decoration.LinkMiddleDecorNone, "",
			[2]string{}},
		{"A -up-> B", "NONE-NORMAL(null)-ARROW", "B", "A", 2, decoration.LinkMiddleDecorNone, "", [2]string{}},
		{"A .left.> B : uses", "NONE-DASHED(null)-ARROW", "B", "A", 1, decoration.LinkMiddleDecorNone, "uses",
			[2]string{}},
		{"A -r-> B", "ARROW-NORMAL(null)-NONE", "A", "B", 1, decoration.LinkMiddleDecorNone, "", [2]string{}},
		{"A - B", "NONE-NORMAL(null)-NONE", "A", "B", 1, decoration.LinkMiddleDecorNone, "", [2]string{}},
		{"A -0- B", "NONE-NORMAL(null)-NONE", "A", "B", 2, decoration.LinkMiddleDecorCircle, "", [2]string{}},
		{`A "1" --> "many" B`, "ARROW-NORMAL(null)-NONE", "A", "B", 2, decoration.LinkMiddleDecorNone, "",
			[2]string{"1", "many"}},
	} {
		_, link := parseLink(t, c.line)
		if got := link.Type().String(); got != c.typ {
			t.Errorf("%s: type %s, want %s", c.line, got, c.typ)
		}
		if link.Entity1().Name() != c.from || link.Entity2().Name() != c.to {
			t.Errorf("%s: from %s to %s", c.line, link.Entity1().Name(), link.Entity2().Name())
		}
		if link.Length() != c.length {
			t.Errorf("%s: length %d, want %d", c.line, link.Length(), c.length)
		}
		if link.Type().MiddleDecor() != c.middle {
			t.Errorf("%s: middle %v", c.line, link.Type().MiddleDecor())
		}
		if label := link.Label(); c.label != "" && (label == nil || label.Get(0).String() != c.label) {
			t.Errorf("%s: label %v", c.line, label)
		}
		if c.quantifier != ([2]string{}) && (link.Quantifier1().S != c.quantifier[0] ||
			link.Quantifier2().S != c.quantifier[1]) {
			t.Errorf("%s: quantifiers %v %v", c.line, link.Quantifier1(), link.Quantifier2())
		}
	}
}

func TestLinkElementBracketStyles(t *testing.T) {
	for _, c := range []struct{ line, style string }{
		{"A -[#red,dashed]-> B", "DASHED"},
		{"A -[bold]-> B", "BOLD"},
		{"A -up[#blue,dotted]-> B", "DOTTED"},
		{"A -[dashed]up-> B", "DASHED"},
	} {
		_, link := parseLink(t, c.line)
		if got := link.Type().Style().String(); !strings.HasPrefix(got, c.style) {
			t.Errorf("%s: style %s, want %s", c.line, got, c.style)
		}
	}
	_, link := parseLink(t, "A -[#red]-> B")
	if link.SpecificColor() == nil {
		t.Error("-[#red]->: no line color")
	}
	_, link = parseLink(t, "A --> B #green")
	if link.SpecificColor() == nil {
		t.Error("#green: no line color")
	}
	_, link = parseLink(t, "A --> B : label >")
	if link.LinkArrow() != abel.LinkArrowDirectNormal || link.Label().Get(0).String() != "label" {
		t.Errorf("label arrow %v, label %v", link.LinkArrow(), link.Label())
	}
}

func TestLinkElementShortFormEnds(t *testing.T) {
	d, link := parseLink(t, ":Admin: --> (Reset password)")
	if e := link.Entity1(); e.Name() != "Admin" || e.USymbol() != abel.USymbolActorStickman {
		t.Errorf("first end %s, symbol %v", e.Name(), e.USymbol())
	}
	if e := link.Entity2(); e.Name() != "Reset password" || e.LeafType() != abel.LeafTypeUsecase {
		t.Errorf("second end %s, type %v", e.Name(), e.LeafType())
	}
	_, link = parseLink(t, "[API] --> () REST")
	if e := link.Entity1(); e.Name() != "API" || e.USymbol() != abel.USymbolComponent2 {
		t.Errorf("first end %s, symbol %v", e.Name(), e.USymbol())
	}
	if e := link.Entity2(); e.Name() != "REST" || e.USymbol() != abel.USymbolInterface {
		t.Errorf("second end %s, symbol %v", e.Name(), e.USymbol())
	}
	_ = d
	// A name only used in links stays still unknown until makeDiagramReady.
	_, link = parseLink(t, "Web --> Db")
	if e := link.Entity1(); e.LeafType() != abel.LeafTypeDescription || e.USymbol() != abel.USymbolInterface {
		t.Errorf("Web: type %v, symbol %v", e.LeafType(), e.USymbol())
	}
}

func TestLinkElementStateMarkersRejected(t *testing.T) {
	for _, line := range []string{"[*] --> A", "A --> (*)"} {
		if _, ok := parseWith(t, linkCommands, line).(*core.ErrorDiagram); !ok {
			t.Errorf("%s was accepted", line)
		}
	}
	// The rewrite of upstream's negative lookaheads keeps longer contents.
	_, link := parseLink(t, "[*x] --> (*y)")
	if link.Entity1().Name() != "*x" || link.Entity2().Name() != "*y" {
		t.Errorf("ends %s %s", link.Entity1().Name(), link.Entity2().Name())
	}
}
