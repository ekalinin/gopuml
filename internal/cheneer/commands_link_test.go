// SPDX-License-Identifier: MIT
// Provenance: original

package cheneer

import (
	"strconv"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// entities declares the entities of names, each with an empty block.
func entities(names ...string) []string {
	var result []string
	for _, n := range names {
		result = append(result, "entity "+n+" {", "}")
	}
	return result
}

// linkBetween returns the only link from the entity of name1 to the one
// of name2.
func linkBetween(t *testing.T, d *ChenEerDiagram, name1, name2 string) *abel.Link {
	t.Helper()
	var result *abel.Link
	for _, l := range d.Links() {
		if l.Entity1().Name() == name1 && l.Entity2().Name() == name2 {
			if result != nil {
				t.Fatalf("several links from %q to %q", name1, name2)
			}
			result = l
		}
	}
	if result == nil {
		t.Fatalf("no link from %q to %q", name1, name2)
	}
	return result
}

func isBold(l *abel.Link) bool { return l.Type().Style().String() == "BOLD(null)" }

// uidNumber is the number of a uid such as "ent0005" or "lnk6", which
// share one sequence.
func uidNumber(t *testing.T, uid string) int {
	t.Helper()
	n, err := strconv.Atoi(strings.TrimLeft(uid, "abcdefghijklmnopqrstuvwxyz"))
	if err != nil {
		t.Fatalf("uid %q: %v", uid, err)
	}
	return n
}

func TestAssociations(t *testing.T) {
	d := parse(t, append(entities("A", "B", "C", "R"),
		`R -1- A`, `R -- B`, `R -(N, M)- C`, `A -N= B`, `A =N- C`, `R -1- C #lime`)...)
	if l := linkBetween(t, d, "R", "A"); l.Label().String() != "[1]" || l.Length() != 3 || isBold(l) {
		t.Errorf("R -1- A: %v %d %v", l.Label(), l.Length(), l.Type().Style())
	}
	if l := linkBetween(t, d, "R", "B"); !klimt.DisplayIsNull(l.Label()) {
		t.Errorf("R -- B: %v", l.Label())
	}
	if l := linkBetween(t, d, "A", "B"); isBold(l) {
		t.Error("A -N= B is bold")
	}
	if l := linkBetween(t, d, "A", "C"); !isBold(l) {
		t.Error("A =N- C is not bold")
	}
	var lime *abel.Link
	for _, l := range d.Links() {
		if l.Entity1().Name() == "R" && l.Entity2().Name() == "C" {
			if l.Label().String() == "[(N, M)]" {
				continue
			}
			lime = l
		}
	}
	if lime == nil {
		t.Fatal("no link R -1- C")
	}
	if line := lime.Colors().Color(color.TypeLine); line == nil || line.AsString() != "#00FF00" {
		t.Errorf("R -1- C #lime: line %v", line)
	}
	if got := parseError(t, append(entities("A"), `A -1- B`)...); got != "No such entity: B"+assumed {
		t.Errorf("unknown entity: %q", got)
	}
}

func TestSimpleSubclasses(t *testing.T) {
	d := parse(t, append(entities("A", "B", "C"), `A ->- B`, `C =<- A`)...)
	if l := linkBetween(t, d, "A", "B"); l.Type().MiddleDecor() != decoration.LinkMiddleDecorSuperset ||
		l.Length() != 3 || isBold(l) {
		t.Errorf("A ->- B: %v %d", l.Type(), l.Length())
	}
	if l := linkBetween(t, d, "C", "A"); l.Type().MiddleDecor() != decoration.LinkMiddleDecorSubset || !isBold(l) {
		t.Errorf("C =<- A: %v", l.Type())
	}
}

func TestUnion(t *testing.T) {
	d := parse(t, append(entities("PERSON", "CUSTOMER", "DIRECTOR"),
		`PERSON ->- U { CUSTOMER, DIRECTOR } #lime;line:green`)...)
	const center = "PERSON/U CUSTOMER, DIRECTOR /center"
	circle := leaf(t, d, center)
	if circle.LeafType() != abel.LeafTypeChenCircle || circle.Display().String() != "[U]" {
		t.Errorf("circle %v %v", circle.LeafType(), circle.Display())
	}
	// The circle is created before its links, as upstream does.
	for _, l := range d.Links() {
		if uidNumber(t, l.UID()) < uidNumber(t, circle.UID()) {
			t.Errorf("link %s comes before the circle %s", l.UID(), circle.UID())
		}
	}
	if back := circle.Colors().Color(color.TypeBack); back == nil || back.AsString() != "#00FF00" {
		t.Errorf("circle background %v", back)
	}
	super := linkBetween(t, d, "PERSON", center)
	if super.Type().MiddleDecor() != decoration.LinkMiddleDecorSuperset || super.Length() != 2 {
		t.Errorf("superclass link %v %d", super.Type(), super.Length())
	}
	for _, name := range []string{"CUSTOMER", "DIRECTOR"} {
		l := linkBetween(t, d, center, name)
		if l.Type().MiddleDecor() != decoration.LinkMiddleDecorNone || l.Length() != 3 {
			t.Errorf("link to %s: %v %d", name, l.Type(), l.Length())
		}
		if line := l.Colors().Color(color.TypeLine); line == nil || line.AsString() != "#008000" {
			t.Errorf("link to %s: line %v", name, line)
		}
	}
	if line := super.Colors().Color(color.TypeLine); line == nil || line.AsString() != "#008000" {
		t.Errorf("superclass link: line %v", line)
	}
}

func TestDisjointSubclassGroup(t *testing.T) {
	d := parse(t, append(entities("CHILD", "TODDLER", "TEEN"), `CHILD =>= d { TODDLER, TEEN }`)...)
	const center = "CHILD/d TODDLER, TEEN /center"
	if l := linkBetween(t, d, "CHILD", center); !isBold(l) || l.Type().MiddleDecor() != decoration.LinkMiddleDecorNone {
		t.Errorf("superclass link %v", l.Type())
	}
	for _, name := range []string{"TODDLER", "TEEN"} {
		if l := linkBetween(t, d, center, name); isBold(l) ||
			l.Type().MiddleDecor() != decoration.LinkMiddleDecorSuperset {
			t.Errorf("link to %s: %v", name, l.Type())
		}
	}
	if got := parseError(t, append(entities("A", "B"), `A ->- d { B, C }`)...); got != "No such entity: C"+assumed {
		t.Errorf("unknown subclass: %q", got)
	}
}
