// SPDX-License-Identifier: MIT
// Provenance: original

package abel

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// testDiagram is a CucaDiagram with the namespace separator of class
// diagrams.
type testDiagram struct {
	AbstractEntityDiagram
}

func newTestDiagram(separator string, hasSeparator bool) *testDiagram {
	lines := []text.StringLocated{{S: "@startuml"}, {S: "@enduml"}}
	d := &testDiagram{}
	d.InitCucaDiagram(core.NewRenderContext(), core.NewUmlSource(lines, lines, true), core.Class, nil, nil)
	d.SetNamespaceSeparator(separator, hasSeparator)
	return d
}

func (d *testDiagram) leaf(t *testing.T, name string, typ LeafType) *Entity {
	t.Helper()
	q := d.QuarkInContext(true, name)
	if q.Data() != nil {
		return q.Data()
	}
	return d.ReallyCreateLeaf(nil, q, klimt.DisplayWithNewlines(d.Pragma(), q.Name()), typ, USymbolNull)
}

func (d *testDiagram) link(a, b *Entity) *Link {
	l := NewLink(nil, &d.CucaDiagram, nil, a, b, decoration.NewLinkType(decoration.LinkDecorNone,
		decoration.LinkDecorArrow), LinkArgNoDisplay(2))
	d.AddLink(l)
	return l
}

func TestQuarkNamespace(t *testing.T) {
	d := newTestDiagram(".", true)
	c := d.leaf(t, "a.b.C", LeafTypeClass)
	if got := c.Quark().QualifiedName(); got != "a.b.C" {
		t.Errorf("qualified name = %q", got)
	}
	if c.Name() != "C" || c.Quark().Parent().Name() != "b" {
		t.Errorf("name %q, parent %q", c.Name(), c.Quark().Parent().Name())
	}
	// The class creates the packages of its namespace.
	b := c.ParentContainer()
	if b == nil || !b.IsGroup() || b.GroupType() != GroupTypePackage || b.Display().Get(0).String() != "b" {
		t.Fatalf("phantom group = %v", b)
	}
	// A unique simple name is found from anywhere.
	if q := d.QuarkInContext(true, "C"); q != c.Quark() {
		t.Errorf("reuse of a unique name: %v", q)
	}
	// Bad names.
	if _, err, score := d.QuarkInContextSafe(false, "x..y"); err != "Bad name since . is a separator" || score != 3 {
		t.Errorf("bad name: %q %d", err, score)
	}
	d.leaf(t, "L", LeafTypeClass)
	if _, err, _ := d.QuarkInContextSafe(false, "L.x"); err != "Not a package: L" {
		t.Errorf("not a package: %q", err)
	}
	// Without separator, dots are part of the names.
	n := newTestDiagram("", false)
	e := n.leaf(t, "a.b.C", LeafTypeClass)
	if e.Name() != "a.b.C" || e.ParentContainer() != n.RootGroup() {
		t.Errorf("no separator: %q in %v", e.Name(), e.ParentContainer())
	}
	if !strings.HasPrefix(e.UID(), "ent") || n.RootGroup().UID() != "entroot" {
		t.Errorf("uids %q %q", e.UID(), n.RootGroup().UID())
	}
}

func TestHideShowRules(t *testing.T) {
	d := newTestDiagram(".", true)
	foo := d.leaf(t, "FooBar", LeafTypeClass)
	baz := d.leaf(t, "FooBaz", LeafTypeClass)
	other := d.leaf(t, "Other", LeafTypeClass)
	other.SetStereotype(klimt.StereotypeBuild("<<st>>"))
	tagged := d.leaf(t, "Tagged", LeafTypeClass)
	tagged.AddStereotag(klimt.NewStereotag("tag"))
	d.HideOrShow2("Foo*", false)
	d.HideOrShow2("*Baz", true)
	d.HideOrShow2("<<st>>", false)
	d.HideOrShow2("$tag", false)
	for _, c := range []struct {
		e      *Entity
		hidden bool
	}{{foo, true}, {baz, false}, {other, true}, {tagged, true}} {
		if got := c.e.IsHidden(); got != c.hidden {
			t.Errorf("%s hidden = %v, want %v", c.e.Name(), got, c.hidden)
		}
	}
	// A note linked to a single entity follows it.
	note := d.leaf(t, "N", LeafTypeNote)
	d.link(foo, note)
	if !note.IsHidden() {
		t.Error("note of a hidden entity is not hidden")
	}
}

func TestRemoveUnlinked(t *testing.T) {
	d := newTestDiagram(".", true)
	a := d.leaf(t, "A", LeafTypeClass)
	b := d.leaf(t, "B", LeafTypeClass)
	c := d.leaf(t, "C", LeafTypeClass)
	d.link(a, b)
	d.RemoveOrRestore("@unlinked", false)
	if a.IsRemoved() || b.IsRemoved() || !c.IsRemoved() {
		t.Errorf("removed: A %v B %v C %v", a.IsRemoved(), b.IsRemoved(), c.IsRemoved())
	}
	d.RemoveOrRestore("C", true)
	if c.IsRemoved() {
		t.Error("restore C")
	}
	d.RemoveOrRestore("<<x>>", false)
	if !d.IsStereotypeRemoved(klimt.StereotypeBuild("<<x>>")) || d.IsStereotypeRemoved(klimt.StereotypeBuild("<<y>>")) {
		t.Error("stereotype removal")
	}
}

func TestRulesInPackage(t *testing.T) {
	d := newTestDiagram(".", true)
	q := d.QuarkInContext(false, "p")
	d.GotoGroup(nil, q, klimt.DisplayWithNewlines(d.Pragma(), "p"), GroupTypePackage, USymbolNull)
	a := d.leaf(t, "A", LeafTypeClass)
	d.HideOrShow2("A", false)
	d.EndGroup()
	if got := d.Hides2()[0].What(); got != "p.A" {
		t.Errorf("rule in package = %q", got)
	}
	if !a.IsHidden() {
		t.Error("p.A not hidden")
	}
}

func TestPortions(t *testing.T) {
	d := newTestDiagram(".", true)
	a := d.leaf(t, "A", LeafTypeClass)
	i := d.leaf(t, "I", LeafTypeInterface)
	a.Bodier().AddFieldOrMethod("field")
	d.HideOrShow(GenderAll(), uml.PortionMember, false)
	d.HideOrShow(GenderByEntityType(LeafTypeInterface), uml.PortionMethod, true)
	d.HideOrShow(GenderAnd(GenderAll(), GenderEmptyFields()), uml.PortionCircledCharacter, false)
	for _, c := range []struct {
		e       *Entity
		portion uml.EntityPortion
		show    bool
	}{
		{a, uml.PortionField, false}, {a, uml.PortionMethod, false}, {i, uml.PortionMethod, true},
		{a, uml.PortionCircledCharacter, true}, {i, uml.PortionCircledCharacter, false},
	} {
		if got := d.ShowPortion(c.portion, c.e); got != c.show {
			t.Errorf("%s %s = %v, want %v", c.e.Name(), c.portion, got, c.show)
		}
	}
}

func TestMembers(t *testing.T) {
	for _, c := range []struct {
		line     string
		method   bool
		display  string
		vis      string
		static   bool
		abstract bool
	}{
		{"-id : int", false, "id : int", "PRIVATE_FIELD", false, false},
		{"{static} +create() : Order", true, "create() : Order", "PUBLIC_METHOD", true, false},
		{"{abstract} #run()", true, "run()", "PROTECTED_METHOD", false, true},
		{"{classifier} x", false, "x", "", true, false},
		{"<<create>> make()", true, "«create» make()", "", false, false},
		{"--", false, "--", "", false, false},
	} {
		var m *Member
		if c.method {
			m = MemberMethod(c.line)
		} else {
			m = MemberField(c.line)
		}
		vis := ""
		if v, ok := m.VisibilityModifier(); ok {
			vis = v.String()
		}
		if m.Display(false) != c.display || vis != c.vis || m.IsStatic() != c.static || m.IsAbstract() != c.abstract {
			t.Errorf("%q: display %q vis %q static %v abstract %v", c.line, m.Display(false), vis, m.IsStatic(),
				m.IsAbstract())
		}
	}
	// The link of a member is in brackets: "[[[target]]]".
	m := MemberField("field [[[http://x]]]")
	if !m.HasURL() || m.URL().URL() != "http://x" || m.Display(false) != "field" {
		t.Errorf("member url: %v %q", m.URL(), m.Display(false))
	}
}

func TestBodierSplitsFieldsAndMethods(t *testing.T) {
	var hidden VisibilitySet
	b := newBodierLikeClassOrObject(LeafTypeClass, &hidden)
	for _, s := range []string{"a", "m1()", "", "m2()", "-pp", "{method} x", "{field} y()", ""} {
		b.AddFieldOrMethod(s)
	}
	// An empty line between methods is a method; other empty lines are
	// fields, but not the first and last ones.
	fields := strings.Join(memberDisplays(b.FieldsToDisplay()), ",")
	methods := strings.Join(memberDisplays(b.MethodsToDisplay()), ",")
	if fields != "a, ,pp,y()" || methods != "m1(), ,m2(),x" {
		t.Errorf("fields %q methods %q", fields, methods)
	}
	// The hidden visibilities are a live view of the diagram's set.
	hidden = hidden.With(klimt.VisibilityPrivateField)
	b.AddFieldOrMethod("z")
	if fields := strings.Join(memberDisplays(b.FieldsToDisplay()), ","); fields != "a, ,y(), ,z" {
		t.Errorf("fields after hide = %q", fields)
	}
}

func memberDisplays(d *klimt.Display) []string {
	var out []string
	for _, cs := range d.AsList() {
		out = append(out, cs.(*Member).Display(false))
	}
	return out
}

func TestMapBody(t *testing.T) {
	b := newBodierMap()
	if !b.AddFieldOrMethod("k1 => v1") || !b.AddFieldOrMethod("k2 *--> T") || b.AddFieldOrMethod("bad") {
		t.Fatal("map lines")
	}
	keys, values := b.Entries()
	if strings.Join(keys, ",") != "k1,k2" || values["k1"] != "v1" || values["k2"] != "\x00" {
		t.Errorf("entries %v %q", keys, values)
	}
}

func TestSingleStrategy(t *testing.T) {
	d := newTestDiagram(".", true)
	for _, n := range []string{"A", "B", "C", "D", "E"} {
		d.leaf(t, n, LeafTypeClass)
	}
	d.ApplySingleStrategy()
	var got []string
	for _, l := range d.Links() {
		got = append(got, l.Entity1().Name()+l.Entity2().Name()+":"+itoa(l.Length()))
	}
	// Five standalones make a 3-wide square: A-B-C, A|D, D-E.
	if strings.Join(got, " ") != "AB:1 BC:1 AD:2 DE:1" {
		t.Errorf("links %v", got)
	}
	for _, l := range d.Links() {
		if !l.Type().IsInvisible() {
			t.Error("visible square link")
		}
	}
}

func TestLinkInverse(t *testing.T) {
	d := newTestDiagram(".", true)
	a := d.leaf(t, "A", LeafTypeClass)
	b := d.leaf(t, "B", LeafTypeClass)
	l := NewLink(nil, &d.CucaDiagram, nil, a, b, decoration.NewLinkType(decoration.LinkDecorExtends,
		decoration.LinkDecorNone), LinkArgNoDisplay(2).WithQuantifier(Some("1"), Some("*")))
	l.SetLinkArrow(LinkArrowDirectNormal)
	inv := l.Inv()
	if inv.Entity1() != b || inv.Type().String() != "NONE-NORMAL(null)-EXTENDS" || inv.Quantifier1().S != "*" ||
		inv.LinkArrow() != LinkArrowBackward || !inv.IsInverted() {
		t.Errorf("inverse: %v %v %v %v", inv.Entity1().Name(), inv.Type(), inv.Quantifier1(), inv.LinkArrow())
	}
	if l.UID() == inv.UID() {
		t.Error("inverse shares the uid")
	}
}

func TestNewpagedDiagramPages(t *testing.T) {
	d1 := newTestDiagram(".", true)
	d2 := newTestDiagram(".", true)
	n := NewNewpagedDiagram(nil, d1, d2)
	if n.Description() != "(0 entities) (0 entities)" || n.Pages() != 1 || n.LastDiagram() != d2 {
		t.Errorf("newpaged: %q %d", n.Description(), n.Pages())
	}
}
