// SPDX-License-Identifier: MIT
// Provenance: original

package statediagram

import (
	"testing"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

func compositeCommands() []cmd {
	return []cmd{commandCreateState(), commandLinkState(), commandCreatePackageState(), commandCreatePackage2(),
		commandEndState(), commandConcurrentState()}
}

// TestCompositeStatesNested checks nested composite states closed by "}"
// and by "end state", with their display names and colors.
func TestCompositeStatesNested(t *testing.T) {
	d := mustParse(t, compositeCommands,
		`state "Outer state" as Outer #pink ##[dashed]red {`,
		`  state Inner begin`,
		`    state Leaf`,
		`  end state`,
		`}`,
		`state After`)
	outer, inner, leaf := entity(t, d, "Outer"), entity(t, d, "Inner"), entity(t, d, "Leaf")
	if outer.GroupType() != abel.GroupTypeState || displayOf(outer) != "Outer state" {
		t.Errorf("Outer: %v %q", outer.GroupType(), displayOf(outer))
	}
	if outer.Colors().Color(color.TypeBack) == nil || outer.Colors().LineStyle() != color.LineStyleDashed {
		t.Errorf("Outer colors %v", outer.Colors())
	}
	if inner.ParentContainer() != outer || leaf.ParentContainer() != inner {
		t.Errorf("Inner in %q, Leaf in %q", inner.ParentContainer().Name(), leaf.ParentContainer().Name())
	}
	if after := entity(t, d, "After"); !after.ParentContainer().IsRoot() {
		t.Errorf("After in %q", after.ParentContainer().Name())
	}
}

// TestFrame checks that "frame" opens a package drawn as a frame.
func TestFrame(t *testing.T) {
	d := mustParse(t, compositeCommands, `frame "Big frame" as F <<boxed>> {`, `state A`, `}`)
	f := entity(t, d, "F")
	if f.GroupType() != abel.GroupTypePackage || f.USymbol() != abel.USymbolFrame || displayOf(f) != "Big frame" ||
		f.Stereotype() == nil {
		t.Errorf("F: %v %v %q", f.GroupType(), f.USymbol(), displayOf(f))
	}
	if entity(t, d, "A").ParentContainer() != f {
		t.Error("A is not in F")
	}
}

// TestRegions checks the "--" and "||" regions: the states of the first
// region stay in the composite state, each separator opens a region, and
// both carry the separator.
func TestRegions(t *testing.T) {
	for _, c := range []struct {
		separator string
		want      rune
	}{{"--", '-'}, {"||", '|'}} {
		d := mustParse(t, compositeCommands,
			`state Active {`, `state A`, c.separator, `state B`, c.separator, `state C`, `}`, `state Other`)
		active := entity(t, d, "Active")
		a, b, cc := entity(t, d, "A"), entity(t, d, "B"), entity(t, d, "C")
		if a.ParentContainer() != active || b.ParentContainer().GroupType() != abel.GroupTypeConcurrentState ||
			cc.ParentContainer() == b.ParentContainer() || cc.ParentContainer().ParentContainer() != active {
			t.Errorf("%s: A in %q, B in %q, C in %q", c.separator, a.ParentContainer().Name(),
				b.ParentContainer().Name(), cc.ParentContainer().Name())
		}
		if active.ConcurrentSeparator() != c.want || b.ParentContainer().ConcurrentSeparator() != c.want {
			t.Errorf("%s: separators %q and %q", c.separator, active.ConcurrentSeparator(),
				b.ParentContainer().ConcurrentSeparator())
		}
		if !entity(t, d, "Other").ParentContainer().IsRoot() {
			t.Errorf("%s: Other is not at the top level", c.separator)
		}
	}
}

// TestEndStateWithoutState checks upstream's error for "}" at the top
// level.
func TestEndStateWithoutState(t *testing.T) {
	if msg := parseError(t, compositeCommands, `state A`, `}`); msg != "No inner state defined (Assumed diagram type: state)" {
		t.Errorf("error %q", msg)
	}
}
