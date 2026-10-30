// SPDX-License-Identifier: MIT
// Provenance: original

package cheneer

import (
	"testing"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// linkFrom returns the only link that starts at the entity of a name.
func linkFrom(t *testing.T, d *ChenEerDiagram, name string) *abel.Link {
	t.Helper()
	var result *abel.Link
	for _, l := range d.Links() {
		if l.Entity1().Name() == name {
			if result != nil {
				t.Fatalf("several links from %q", name)
			}
			result = l
		}
	}
	if result == nil {
		t.Fatalf("no link from %q", name)
	}
	return result
}

func TestAttribute(t *testing.T) {
	d := parse(t, `entity CUSTOMER {`, `Number <<key>>`, `"No." as Code`, `}`)
	number := leaf(t, d, "CUSTOMER/Number")
	if number.LeafType() != abel.LeafTypeChenAttribute || number.Display().String() != "[Number]" {
		t.Errorf("Number: %v %v", number.LeafType(), number.Display())
	}
	if st := number.Stereotype(); st == nil || st.String() != "<<key>>" {
		t.Errorf("stereotype %v", st)
	}
	link := linkFrom(t, d, "CUSTOMER/Number")
	if link.Entity2() != leaf(t, d, "CUSTOMER") || link.Length() != 2 || !klimt.DisplayIsNull(link.Label()) ||
		link.Type().Decor1() != decoration.LinkDecorNone || link.Type().Decor2() != decoration.LinkDecorNone {
		t.Errorf("link %v %d %v %v", link.Entity2().Name(), link.Length(), link.Label(), link.Type())
	}
	if code := leaf(t, d, "CUSTOMER/Code"); code.Display().String() != "[No.]" {
		t.Errorf("Code: %v", code.Display())
	}
}

func TestCompositeAttribute(t *testing.T) {
	d := parse(t, `entity DIRECTOR {`, `Name {`, `Fname`, `}`, `Born`, `}`)
	if l := linkFrom(t, d, "DIRECTOR/Name/Fname"); l.Entity2().Name() != "DIRECTOR/Name" {
		t.Errorf("Fname is linked to %q", l.Entity2().Name())
	}
	if l := linkFrom(t, d, "DIRECTOR/Born"); l.Entity2().Name() != "DIRECTOR" {
		t.Errorf("Born is linked to %q", l.Entity2().Name())
	}
}

func TestLineReadAsAttribute(t *testing.T) {
	d := parse(t, `entity A {`, `class Foo`, `Born : DATE`, `}`)
	if e := leaf(t, d, "A/class Foo"); e.Display().String() != "[class Foo]" {
		t.Errorf("class Foo: %v", e.Display())
	}
	if e := leaf(t, d, "A/Born : DATE"); e.Display().String() != "[Born : DATE]" {
		t.Errorf("Born : DATE: %v", e.Display())
	}
}

func TestAttributeColors(t *testing.T) {
	// The main color of an attribute is its line color, which line:orange
	// then replaces: the attribute has no background.
	d := parse(t, `entity PERSON {`, `Name #lime;line:orange`, `}`)
	for what, colors := range map[string]*color.Colors{
		"attribute": leaf(t, d, "PERSON/Name").Colors(),
		"link":      linkFrom(t, d, "PERSON/Name").Colors(),
	} {
		if line := colors.Color(color.TypeLine); line == nil || line.AsString() != "#FFA500" {
			t.Errorf("%s: line %v", what, line)
		}
		if back := colors.Color(color.TypeBack); back != nil {
			t.Errorf("%s: background %v", what, back)
		}
	}
}

func TestSameAttributeInTwoEntities(t *testing.T) {
	d := parse(t, `entity DIRECTOR {`, `Number`, `}`, `entity CUSTOMER {`, `Number`, `}`)
	leaf(t, d, "DIRECTOR/Number")
	leaf(t, d, "CUSTOMER/Number")
}

func TestDottedNames(t *testing.T) {
	d := parse(t, `entity shop.Customer {`, `first.name`, `}`)
	if e := leaf(t, d, "shop.Customer"); e.Display().String() != "[shop.Customer]" {
		t.Errorf("entity %v", e.Display())
	}
	if e := leaf(t, d, "shop.Customer/first.name"); e.Display().String() != "[first.name]" {
		t.Errorf("attribute %v", e.Display())
	}
}

func TestAttributeErrors(t *testing.T) {
	want := "Attribute must be inside an entity, relationship or another attribute" + assumed
	if got := parseError(t, `entity A {`, `}`, `Name`); got != want {
		t.Errorf("outside: %q", got)
	}
	if got := parseError(t, `entity A {`, `Name`, `Name`, `}`); got != "Attribute already exists"+assumed {
		t.Errorf("twice: %q", got)
	}
}
