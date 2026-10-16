// SPDX-License-Identifier: MIT
// Provenance: original

package regex

import "testing"

// participantPattern resembles upstream's sequence participant command.
func participantPattern() *Concat {
	return NewConcat(
		Start(),
		NewNamedLeaf(1, "TYPE", "(participant|actor|database)"),
		SpaceOneOrMore(),
		Or("CODE",
			NewNamedLeaf(1, "QUOTED", `[%g]([^%g]+)[%g]`),
			NewNamedLeaf(1, "PLAIN", `([%pLN_.]+)`),
		),
		Optional(NewConcat(
			SpaceOneOrMore(),
			NewLeaf("as"),
			SpaceOneOrMore(),
			NewNamedLeaf(1, "ALIAS", `([%pLN_.]+)`),
		)),
		SpaceZeroOrMore(),
		Optional(Or("",
			NewNamedLeaf(1, "COLOR1", `(#\w+)`),
			NewNamedLeaf(1, "COLOR2", `\[(#\w+)\]`),
		)),
		End(),
	)
}

func TestConcatMatcher(t *testing.T) {
	p := participantPattern()
	r := p.Matcher(`ACTOR "John Doe" as jd #red`)
	if r == nil {
		t.Fatal("expected a match")
	}
	check := func(name string, want string, wantOK bool) {
		t.Helper()
		got, ok := r.Get(name, 0)
		if got != want || ok != wantOK {
			t.Errorf("%s = %q, %v; want %q, %v", name, got, ok, want, wantOK)
		}
	}
	check("TYPE", "ACTOR", true) // case-insensitive
	check("QUOTED", "John Doe", true)
	check("PLAIN", "", false) // the other branch did not match: null upstream
	check("CODE", `"John Doe"`, true)
	check("ALIAS", "jd", true)
	check("COLOR1", "#red", true)
	if s, ok := r.Lazy("COLOR", 0); s != "#red" || !ok {
		t.Errorf("Lazy(COLOR) = %q, %v", s, ok)
	}
	if r.Str("MISSING", 0) != "" {
		t.Error("unknown names return empty")
	}

	r = p.Matcher("database db [#blue]")
	if r == nil || r.Str("PLAIN", 0) != "db" || r.Str("COLOR2", 0) != "#blue" {
		t.Fatalf("second form: %+v", r)
	}
	if _, ok := r.Get("ALIAS", 0); ok {
		t.Error("ALIAS must be unmatched")
	}
	if p.Matcher("class A") != nil || p.Match("class A") {
		t.Error("class A must not match")
	}
}

func TestCounts(t *testing.T) {
	p := participantPattern()
	// TYPE 1 + Or(CODE) 1+2 + Optional(ALIAS) 1 + unnamed Or 1+2 = 8
	if p.Count() != 8 {
		t.Errorf("Count = %d, want 8", p.Count())
	}
	if want := `^(participant|actor|database)[%s]+([%g]([^%g]+)[%g]|([%pLN_.]+))(?:[%s]+as[%s]+([%pLN_.]+))?[%s]*(?:(?:(#\w+)|\[(#\w+)\]))?$`; p.Pattern() != want {
		t.Errorf("Pattern =\n%s\nwant\n%s", p.Pattern(), want)
	}
}

func TestUnnamedOneOrMoreShiftsGroups(t *testing.T) {
	// Upstream does not consume the group of an unnamed OneOrMore, so the
	// next named leaf receives it.
	p := NewConcat(Start(), OneOrMore("", NewLeaf("ab")), NewNamedLeaf(1, "X", "(c)"), End())
	r := p.Matcher("ababc")
	if got := r.Str("X", 0); got != "ab" {
		t.Errorf("X = %q, want the shifted group %q", got, "ab")
	}
}

func TestNamedLeafCountChecked(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a wrong group count must panic")
		}
	}()
	NewNamedLeaf(2, "X", "(a)")
}
