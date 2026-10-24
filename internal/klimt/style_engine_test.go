// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"slices"
	"testing"
)

func TestSpecificityOrder(t *testing.T) {
	// From the least to the most specific, one tier at a time.
	ordered := []Specificity{
		SpecificityAtOrder(5),
		SpecificityAtOrder(6),
		SpecificityAtOrder(1).WithStereotypeCount(1),
		SpecificityAtOrder(1).WithStereotypeCount(2),
		SpecificityAtOrder(9).WithStereotypeCount(3).WithAncestorRank(-2),
		SpecificityAtOrder(1).WithAncestorRank(-1),
		SpecificityAtOrder(1).WithAncestorRank(0),
		SpecificityForcedOverride(),
	}
	for i := range ordered {
		for j := range ordered {
			if got, want := ordered[i].IsBiggerThan(ordered[j]), i > j; got != want {
				t.Errorf("%v bigger than %v: %v", ordered[i], ordered[j], got)
			}
		}
	}
	for _, c := range []struct {
		s    Specificity
		want string
	}{
		{SpecificityAtOrder(3), "3"},
		{SpecificityAtOrder(3).WithStereotypeCount(1), "3,stereo=1"},
		{SpecificityAtOrder(3).WithAncestorRank(-1), "3,stereo=0,ancestor=-1"},
		{SpecificityForcedOverride(), "override"},
	} {
		if c.s.String() != c.want {
			t.Errorf("%q, want %q", c.s.String(), c.want)
		}
	}
}

func TestDarkStringMergeBySpecificity(t *testing.T) {
	red, blue, white := "red", "blue", "white"
	light := NewDarkString(&red, nil, SpecificityAtOrder(2))
	dark := NewDarkString(nil, &white, SpecificityAtOrder(1))
	// A regular and a dark value combine, keeping the regular one's
	// specificity.
	if got := light.MergeWith(&dark).String(); got != "red/white (2)" {
		t.Error(got)
	}
	if got := dark.MergeWith(&light).String(); got != "red/white (2)" {
		t.Error(got)
	}
	// Otherwise the more specific wins, a stereotype over a later plain
	// value, and the other one on a tie.
	stereo := NewDarkString(&blue, nil, SpecificityAtOrder(1).WithStereotypeCount(1))
	if got := light.MergeWith(&stereo).String(); got != "blue/null (1,stereo=1)" {
		t.Error(got)
	}
	tie := NewDarkString(&blue, nil, SpecificityAtOrder(2))
	if got := light.MergeWith(&tie).String(); got != "blue/null (2)" {
		t.Error(got)
	}
}

func TestStyleQueryAtoms(t *testing.T) {
	q := StyleSignatureBasicOf(SNameNote, SNameRoot).AddStereotype("Zeta").AddStereotype("al_p.ha").AddStereotype("é")
	var got []string
	for _, a := range q.atoms() {
		if a.isStereotype {
			got = append(got, "."+a.stereotype)
		} else {
			got = append(got, a.sname.String())
		}
	}
	// Names in enum order (note before root), then the cleaned stereotypes
	// in Java's order.
	want := []string{"note", "root", ".alpha", ".zeta", ".é"}
	if !slices.Equal(got, want) {
		t.Errorf("atoms %v, want %v", got, want)
	}
	if q.String() != "[note, root, .alpha, .zeta, .é] " {
		t.Error(q.String())
	}
	// One query for a stereotype with two labels.
	two := StyleSignatureBasicOf(SNameRoot).WithTOBECHANGED(StereotypeBuild("<<foo>><<bar>>"))
	if s, ok := two.(StyleSignatureBasic); !ok || !slices.Equal(s.Stereotypes(), []string{"bar", "foo"}) {
		t.Errorf("%v", two)
	}
}

func TestLevelConstraint(t *testing.T) {
	for _, c := range []struct {
		declLevel            int
		declStar             bool
		queryLevel           int
		queryStar, wantMatch bool
	}{
		{levelNone, false, levelNone, false, true},
		{levelNone, false, 3, false, true},
		{2, false, 2, false, true},
		{2, false, 3, false, false},
		{2, false, levelNone, false, false},
		{2, true, 3, false, true},
		{2, true, 1, false, false},
		{levelNone, false, 3, true, false}, // a starred query needs a starred declaration
		{levelNone, true, 3, true, true},
	} {
		if got := levelConstraintMatches(c.declLevel, c.declStar, c.queryLevel, c.queryStar); got != c.wantMatch {
			t.Errorf("%+v: %v", c, got)
		}
	}
	q := StyleSignatureBasicOf(SNameNode).AddLevel(2).AddStar()
	if q.String() != "[node] depth(2)*" || q.AddLevel(3).String() != "[node] depth(3)*" {
		t.Error(q.String())
	}
	if class, _ := expectPanic(t, func() { q.AddLevel(64) }); class != illegalArgument {
		t.Error(class)
	}
}

func TestStyleTrieMatchesScan(t *testing.T) {
	for _, skin := range []string{"plantuml.skin", "rose.skin", "debug.skin"} {
		b := mustLoadSkin(t, skin)
		all := b.Styles()
		for _, spec := range readLines(t, "signatures.txt") {
			query := signatureOf(spec)
			// The declarations whose atoms are all in the query's, scanned
			// in load order.
			var scan []*Style
			for _, style := range all {
				decl := style.Signature()
				if query.snames.containsAll(decl.snames) &&
					!slices.ContainsFunc(decl.Stereotypes(), func(st string) bool {
						return !slices.Contains(query.Stereotypes(), st)
					}) &&
					levelConstraintMatches(decl.level, decl.star, query.level, query.star) {
					scan = append(scan, style)
				}
			}
			// The trie walk finds the same declarations, in its own order.
			got := slices.Clone(b.index.findMatching(query))
			slices.SortFunc(got, func(x, y *Style) int { return slices.Index(all, x) - slices.Index(all, y) })
			if !slices.Equal(got, scan) {
				t.Errorf("%s %q: trie %d styles, scan %d", skin, spec, len(got), len(scan))
			}
		}
	}
}
