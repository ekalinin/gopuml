// SPDX-License-Identifier: MIT
// Provenance: original

package activitydiagram

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

func branchCommands() []cmd {
	return []cmd{commandPartition(), commandEndPartition(), commandLinkLongActivity(), commandIf(), commandElse(),
		commandEndif(), commandLinkActivity()}
}

// ifLink returns the link into the branch of the first "if" of body.
func ifLink(t *testing.T, body ...string) *abel.Link {
	t.Helper()
	d := mustParse(t, branchCommands, body...)
	for _, l := range d.Links() {
		if l.Entity2().LeafType() == abel.LeafTypeBranch {
			return l
		}
	}
	t.Fatalf("%q: no link into a branch", body)
	return nil
}

// TestIfForms checks the "if" forms and sources upstream accepts.
func TestIfForms(t *testing.T) {
	for _, c := range []struct {
		body   []string
		source string
		label  string
		branch string
	}{
		{[]string{`(*) --> "A"`, `if "valid?" then`}, "A", "valid?", ""},
		{[]string{`(*) --> "A"`, `if "valid?"`}, "A", "valid?", ""},
		{[]string{`(*) --> "A"`, `if "valid?" as V then`}, "A", "valid?", "V"},
		// As upstream, the unquoted label keeps the space before "then".
		{[]string{`(*) --> "A"`, `if valid then`}, "A", "valid ", ""},
		{[]string{`(*) --> "A"`, `if valid`}, "A", "valid", ""},
		{[]string{`(*) if "a" then`}, "start", "a", ""},
		{[]string{`===B1=== if "a" then`}, "===B1===", "a", ""},
		{[]string{`"Act" if "a" then`}, "Act", "a", ""},
		{[]string{`Act if "a" then`}, "Act", "a", ""},
		{[]string{`(*) --> "A"`, `-->[label] if "a" then`}, "A", "a", ""},
	} {
		l := ifLink(t, c.body...)
		if l.Entity1().Name() != c.source {
			t.Errorf("%q: source %q, want %q", c.body, l.Entity1().Name(), c.source)
		}
		if got := l.Quantifier2(); !got.OK || got.S != c.label {
			t.Errorf("%q: label %v, want %q", c.body, got, c.label)
		}
		if c.branch != "" && l.Entity2().Name() != c.branch {
			t.Errorf("%q: branch %q, want %q", c.body, l.Entity2().Name(), c.branch)
		}
	}
	if l := ifLink(t, `(*) --> "A"`, `-->[label] if "a" then`); l.Label().Get(0).String() != "label" {
		t.Errorf("bracket label %v", l.Label())
	}
}

// TestElseAndNesting checks that "else" links from the branch, and that
// nested branches close in order.
func TestElseAndNesting(t *testing.T) {
	d := mustParse(t, branchCommands,
		`(*) --> "A"`,
		`if "outer" then`,
		`  -->[yes] "B"`,
		`  if "inner" then`,
		`    --> "C"`,
		`  else`,
		`    --> "D"`,
		`  endif`,
		`else`,
		`  -->[no] "E"`,
		`endif`)
	from := map[string]string{}
	for _, l := range d.Links() {
		from[l.Entity2().Name()] = l.Entity1().Name()
	}
	outer, inner := from["B"], from["C"]
	if outer == "" || inner == "" || outer == inner || from[inner] != "B" {
		t.Fatalf("links %v", from)
	}
	if from["D"] != inner || from["E"] != outer {
		t.Errorf("else links: D from %q, E from %q", from["D"], from["E"])
	}
	if d.currentContext != nil {
		t.Error("a branch is still open")
	}
}

// TestElseEndifWithoutIf checks upstream's errors for "else" and "endif"
// without "if".
func TestElseEndifWithoutIf(t *testing.T) {
	if msg := parseError(t, branchCommands, `(*) --> "A"`, `else`); !strings.HasPrefix(msg, "No if for this else") {
		t.Errorf("else: %q", msg)
	}
	if msg := parseError(t, branchCommands, `(*) --> "A"`, `endif`); !strings.HasPrefix(msg, "No if for this endif") {
		t.Errorf("endif: %q", msg)
	}
}

// TestPartitions checks partitions closed by "}" and by "end partition",
// with a color and a stereotype.
func TestPartitions(t *testing.T) {
	d := mustParse(t, branchCommands,
		`partition Payment #pink<<money>> {`,
		`(*) --> "Pay"`,
		`}`,
		`partition "Second one"`,
		`"Pay" --> "Ship"`,
		`end partition`,
		`"Ship" --> (*)`)
	pay, ship := d.Links()[0].Entity2(), d.Links()[1].Entity2()
	payment, second := pay.ParentContainer(), ship.ParentContainer()
	if payment.Name() != "Payment" || payment.Colors().Color(color.TypeBack) == nil || payment.Stereotype() == nil {
		t.Errorf("Payment: %q", payment.Name())
	}
	if second.Name() != "Second one" || second.Stereotype() != nil {
		t.Errorf("second partition: %q", second.Name())
	}
	if !d.Links()[2].Entity2().ParentContainer().IsRoot() {
		t.Error("the end is in a partition")
	}
}
