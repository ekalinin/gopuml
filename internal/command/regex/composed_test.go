// SPDX-License-Identifier: MIT
// Provenance: original

package regex

import (
	"strings"
	"testing"
)

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

// fakeCaptures are UBrex captures for TestResultFromCaptures.
type fakeCaptures map[string][]string

func (f fakeCaptures) FindValuesByKey(key string) []string { return f[key] }

func (f fakeCaptures) FindFirstValuesByKeyPrefix(keyPrefix string) []string {
	for _, key := range []string{"ARROW_STYLE1", "ARROW_STYLE2", "IF1", "IF2"} {
		if strings.HasPrefix(key, keyPrefix) && len(f[key]) > 0 {
			return f[key]
		}
	}
	return nil
}

// TestResultFromCaptures checks that a result backed by UBrex captures
// reads them as upstream's RegexResult(UMatcher): get returns a value of
// the key, getLazzy a value of the first key with the prefix, and an empty
// list is null.
func TestResultFromCaptures(t *testing.T) {
	r := NewResultFromCaptures(fakeCaptures{"IF2": {"valid?"}, "ARROW_STYLE2": {"dotted"}, "BAR": {"B1", "B2"}})
	for _, c := range []struct {
		name string
		num  int
		want string
		ok   bool
	}{{"IF2", 0, "valid?", true}, {"BAR", 1, "B2", true}, {"IF1", 0, "", false}, {"NONE", 0, "", false}} {
		if got, ok := r.Get(c.name, c.num); got != c.want || ok != c.ok {
			t.Errorf("Get(%q, %d) = %q, %v", c.name, c.num, got, ok)
		}
	}
	if got, ok := r.Lazy("ARROW_STYLE", 0); got != "dotted" || !ok {
		t.Errorf("Lazy = %q, %v", got, ok)
	}
	if got, ok := r.Lazy("STEREO", 0); ok {
		t.Errorf("Lazy of a missing prefix = %q", got)
	}
	if got := r.Str("IF2", 0); got != "valid?" {
		t.Errorf("Str = %q", got)
	}
}

// The groups are in the order the pattern fills them, the group of an
// alternative after the groups inside it, and Lazy takes the first matching
// group in that order, as upstream's RegexResult does.
func TestResultOrder(t *testing.T) {
	c := NewConcat(Start(),
		NewNamedLeaf(1, "TAGS1", "(a)?"),
		Or("CODE",
			NewNamedLeaf(1, "CODE1", "(b)"),
			NewNamedLeaf(1, "CODE2", "(c)"),
		),
		NewNamedLeaf(1, "TAGS2", "(d)?"),
		End())
	r := c.Matcher("abd")
	if r == nil {
		t.Fatal("no match")
	}
	if got := strings.Join(r.order, ","); got != "TAGS1,CODE1,CODE2,CODE,TAGS2" {
		t.Errorf("order %s", got)
	}
	if got, _ := r.Lazy("TAGS", 0); got != "a" {
		t.Errorf("Lazy TAGS = %q, want the first matching group", got)
	}
	r = c.Matcher("cd")
	if got, _ := r.Lazy("TAGS", 0); got != "d" {
		t.Errorf("Lazy TAGS = %q, want d", got)
	}
	if got, _ := r.Lazy("CODE", 0); got != "c" {
		t.Errorf("Lazy CODE = %q, want c, the first matching CODE group", got)
	}
}
