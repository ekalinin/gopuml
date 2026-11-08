// SPDX-License-Identifier: MIT
// Provenance: original

package descdiagram

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

func archimateCommands() []cmd {
	return []cmd{abel.CommandEndPackage[*DescriptionDiagram](), commandArchimate(), commandArchimateMultilines(),
		commandArchimatePackage()}
}

func parseArchimate(t *testing.T, body ...string) *DescriptionDiagram {
	t.Helper()
	result := parseWith(t, archimateCommands, body...)
	d, ok := result.(*DescriptionDiagram)
	if !ok {
		if e, ok := result.(*core.ErrorDiagram); ok {
			t.Fatalf("error at line %d: %s", e.Position(), e.Error.Message())
		}
		t.Fatalf("result %T", result)
	}
	return d
}

// checkArchimate checks the symbol, the display (its lines joined with
// "\n"), the background color and the icon of an entity; "" for no color or
// no icon.
func checkArchimate(t *testing.T, e *abel.Entity, symbol abel.USymbol, display, back, icon string) {
	t.Helper()
	if e.USymbol() != symbol {
		t.Errorf("%s: symbol %v, want %v", e.Quark().Name(), e.USymbol(), symbol)
	}
	var lines []string
	for _, l := range e.Display().AsList() {
		lines = append(lines, l.String())
	}
	if got := strings.Join(lines, "\n"); got != display {
		t.Errorf("%s: display %q, want %q", e.Quark().Name(), got, display)
	}
	got := ""
	if c := e.Colors().Color(color.TypeBack); c != nil {
		got = c.AsString()
	}
	if got != back {
		t.Errorf("%s: back color %q, want %q", e.Quark().Name(), got, back)
	}
	label := ""
	if st := e.Stereotype(); st != nil {
		label, _ = st.Label(klimt.GuillemetDoubleComparator())
	}
	if label != icon {
		t.Errorf("%s: icon %q, want %q", e.Quark().Name(), label, icon)
	}
}

func TestArchimate(t *testing.T) {
	d := parseArchimate(t, `archimate #Business "Customer" <<business-actor>> as c`)
	e := entity(t, d, "c")
	if e.LeafType() != abel.LeafTypeDescription {
		t.Errorf("leaf type %v", e.LeafType())
	}
	checkArchimate(t, e, abel.USymbolArchimate, "Customer", "#FFFFCC", "<<business-actor>>")
}

func TestArchimateNameForms(t *testing.T) {
	for _, tc := range []struct{ line, code, display string }{
		{`archimate #Business Customer`, "Customer", "Customer"},
		{`archimate #Business "Display" as d`, "d", "Display"},
		{`archimate #Business d as "Display"`, "d", "Display"},
		{`archimate #Business Display as d`, "d", "Display"},
	} {
		d := parseArchimate(t, tc.line)
		checkArchimate(t, entity(t, d, tc.code), abel.USymbolArchimate, tc.display, "#FFFFCC", "")
	}
}

func TestArchimateUpdated(t *testing.T) {
	d := parseArchimate(t, `archimate #Business "Old" as u`,
		`archimate #Application "Updated" <<application-component>> as u`)
	if n := len(d.Leafs()); n != 1 {
		t.Errorf("%d leafs, want 1", n)
	}
	checkArchimate(t, entity(t, d, "u"), abel.USymbolArchimate, "Updated", "#C2F0FF", "<<application-component>>")
}

func TestArchimateMultilines(t *testing.T) {
	d := parseArchimate(t, `archimate #Business m <<business-actor>> [`, "first line", "second line", "]")
	checkArchimate(t, entity(t, d, "m"), abel.USymbolRectangle, "first line\nsecond line", "#FFFFCC",
		"<<business-actor>>")

	result := parseWith(t, archimateCommands, `archimate #Business c [`, "first", "]", `archimate #Business c [`,
		"second", "]")
	e2, ok := result.(*core.ErrorDiagram)
	if !ok {
		t.Fatalf("result %T, want an error", result)
	}
	// The error is on the first line of the block, as upstream; the
	// conformance case archimate-multiline-twice checks the error image
	if e2.Error.Error != "Already exists c" || e2.Error.Line == nil || e2.Error.Line.S != "archimate #Business c [" {
		t.Errorf("error %q at %v", e2.Error.Error, e2.Error.Line)
	}
}

func TestArchimatePackage(t *testing.T) {
	d := parseArchimate(t, `archimate #Technology "Domain" <<technology-node>> as d {`,
		`archimate #Application "App" as a`, "}")
	g := entity(t, d, "d")
	if !g.IsGroup() || g.GroupType() != abel.GroupTypePackage {
		t.Errorf("group %v, type %v", g.IsGroup(), g.GroupType())
	}
	checkArchimate(t, g, abel.USymbolArchimate, "Domain", "#C9FFC9", "<<technology-node>>")
	a := d.QuarkInContext(false, "d.a").Data()
	if a == nil {
		a = entity(t, d, "a")
	}
	if a.Quark().Parent().Data() != g {
		t.Errorf("parent of a is %v", a.Quark().Parent().Name())
	}
}
