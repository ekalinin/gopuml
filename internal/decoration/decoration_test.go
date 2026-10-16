// SPDX-License-Identifier: MIT
// Provenance: original

package decoration

import (
	"regexp"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/command/regex"
)

func TestLookupDecors(t *testing.T) {
	for _, c := range []struct {
		head1, head2 string
		decor1       LinkDecor
		decor2       LinkDecor
	}{
		{"<|", "|>", LinkDecorExtends, LinkDecorExtends},
		{"^", "^", LinkDecorExtends, LinkDecorExtends},
		{"*", "*", LinkDecorComposition, LinkDecorComposition},
		{"o", "o", LinkDecorAggregation, LinkDecorAggregation},
		{"}o", "o{", LinkDecorCircleCrowfoot, LinkDecorCircleCrowfoot},
		{"<", ">", LinkDecorArrow, LinkDecorArrow},
		{"0)", "(0", LinkDecorCircleConnect, LinkDecorCircleConnect},
		{"", `\\`, LinkDecorNone, LinkDecorHalfArrowUp},
		{" <| ", "?", LinkDecorExtends, LinkDecorNone},
	} {
		if got := LookupDecors1(c.head1, true); got != c.decor1 {
			t.Errorf("decors1 %q = %v, want %v", c.head1, got, c.decor1)
		}
		if got := LookupDecors2(c.head2, true); got != c.decor2 {
			t.Errorf("decors2 %q = %v, want %v", c.head2, got, c.decor2)
		}
	}
	if LookupDecors1("<|", false) != LinkDecorNone {
		t.Error("null head")
	}
}

func TestRegexDecors(t *testing.T) {
	re := regex.MustCompile("^" + RegexDecors1() + "$")
	for _, s := range []string{"", "<|", "<||", "<|:", "}o", "|o", "o", "<_", "0)", "#"} {
		if !re.MatchString(s) {
			t.Errorf("decors1 does not match %q", s)
		}
	}
	// The longer keys come first: "<||" is one head, not "<|" and "|".
	head := regexp.MustCompile("^" + RegexDecors1())
	if m := head.FindString("<||--"); m != "<||" {
		t.Errorf("longest head = %q", m)
	}
	// A leading "o" needs a word boundary.
	re2 := regex.MustCompile("^a" + RegexDecors2() + "$")
	if re2.MatchString("ao") {
		t.Error(`"o" matched without a word boundary`)
	}
}

func TestLinkType(t *testing.T) {
	lt := NewLinkType(LinkDecorNone, LinkDecorExtends).GoDashed()
	if lt.String() != "NONE-DASHED(null)-EXTENDS" {
		t.Errorf("String = %q", lt)
	}
	inv := lt.WithMiddleCircleCircled1().Inversed()
	if inv.Decor1() != LinkDecorExtends || inv.Decor2() != LinkDecorNone ||
		inv.MiddleDecor() != LinkMiddleDecorCircleCircled2 {
		t.Errorf("inversed = %v %v", inv, inv.MiddleDecor())
	}
	if !lt.Invisible().IsInvisible() || lt.IsInvisible() {
		t.Error("invisible")
	}
	if got := lt.GoThickness(2).String(); got != "NONE-DASHED(2.0)-EXTENDS" {
		t.Errorf("thickness = %q", got)
	}
	if name, _ := lt.LinkTypeName(); name != "extension" {
		t.Errorf("type name = %q", name)
	}
	if got := NewLinkType(LinkDecorArrow, LinkDecorNone).SpecificDecorationSvek(LinkStrategyLegacyToBeRemoved); got !=
		"arrowsize=0.5" {
		t.Errorf("svek decoration = %q", got)
	}
}

func TestApplyStyle(t *testing.T) {
	norank := false
	w := InitWithLinkType(NewLinkType(LinkDecorNone, LinkDecorNone), func() { norank = true })
	w.ApplyStyle("#red,dashed;#blue,thickness=3,hidden,single,node,norank", true)
	if w.LinkType.String() != "NONE-DASHED(3.0)-NONE" || !w.Hidden || !w.IsSingle() || !w.UseNodeStyle() || !norank {
		t.Errorf("style: %v hidden %v single %v node %v norank %v", w.LinkType, w.Hidden, w.IsSingle(),
			w.UseNodeStyle(), norank)
	}
	if w.SpecificColor() == nil || len(w.SupplementaryColors()) != 1 {
		t.Errorf("colors: %v %v", w.SpecificColor(), w.SupplementaryColors())
	}
}
