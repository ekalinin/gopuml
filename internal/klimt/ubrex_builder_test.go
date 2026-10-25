// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"slices"
	"testing"
)

// TestUBrexBuilderMatchesString checks that an expression built from parts
// matches and captures as the same expression written as a string.
func TestUBrexBuilderMatchesString(t *testing.T) {
	built := UBrexConcat(
		UBrexLeaf("if"),
		UBrexSpaceOneOrMore(),
		UBrexOr(
			UBrexNamed("CODE", UBrexLeaf("〇+「〴an_.」")),
			UBrexLeaf("〴g 〶$QUOTED=〇*〴G 〴g")),
		UBrexOptional(UBrexConcat(UBrexSpaceOneOrMore(), UBrexLeaf("then"))),
		UBrexEnd())
	written := ubrexBuild("if 〇+〴s 【 〶$CODE=〇+「〴an_.」 ┇ 〴g 〶$QUOTED=〇*〴G 〴g 】 〇?〘 〇+〴s then 〙 〒$")
	for _, line := range []string{"if ok", `if "is valid?" then`, "if ok then", "if", "iff ok", `if "x" else`} {
		got, want := built.Match(line), written.matchString(line)
		if got.ExactMatch() != want.exactMatch() {
			t.Errorf("%q: exact %v, want %v", line, got.ExactMatch(), want.exactMatch())
		}
		for _, key := range []string{"CODE", "QUOTED"} {
			if g, w := got.FindValuesByKey(key), want.findValuesByKey(key); !slices.Equal(g, w) {
				t.Errorf("%q: %s = %q, want %q", line, key, g, w)
			}
		}
	}
	if m := built.Match("if ok"); !m.ExactMatch() || !slices.Equal(m.FindValuesByKey("CODE"), []string{"ok"}) {
		t.Errorf("if ok: exact %v, CODE %q", m.ExactMatch(), m.FindValuesByKey("CODE"))
	}
	if m := built.Match("iff ok"); m.ExactMatch() {
		t.Error("iff ok matches")
	}
	if m := built.Match(`if "is valid?" then`); !slices.Equal(m.FindFirstValuesByKeyPrefix("QUO"), []string{"is valid?"}) {
		t.Errorf("prefix lookup %q", m.FindFirstValuesByKeyPrefix("QUO"))
	}
}

// TestUBrexBuilderRepetitions checks the builders of gantt's words:
// repetitions that do not give back what they matched, Upto, which needs
// one character and stops at the first position where its end matches, and
// the case folding of ASCII letters only.
func TestUBrexBuilderRepetitions(t *testing.T) {
	accepted := func(p UBrexPart, s string) (string, bool) {
		m := p.Match(s)
		return m.m.acceptedMatch(), m.m.startMatch()
	}
	for _, tc := range []struct {
		name  string
		part  UBrexPart
		input string
		want  string
		ok    bool
	}{
		{"one or more", UBrexOneOrMore(UBrexLeaf("a")), "aaab", "aaa", true},
		{"one or more needs one", UBrexOneOrMore(UBrexLeaf("a")), "b", "", false},
		{"one or more keeps", UBrexConcat(UBrexOneOrMore(UBrexLeaf("a")), UBrexLeaf("a")), "aaa", "", false},
		{"zero or more", UBrexZeroOrMore(UBrexLeaf("ab")), "ababa", "abab", true},
		{"zero or more empty", UBrexZeroOrMore(UBrexLeaf("a")), "b", "", true},
		{"zero or more keeps", UBrexConcat(UBrexZeroOrMore(UBrexLeaf("a")), UBrexLeaf("a")), "aa", "", false},
		{"upto", UBrexUpto(UBrexLeaf("〴."), UBrexLeaf("jan")), "5th of jan 2020", "5th of jan", true},
		{"upto first", UBrexUpto(UBrexLeaf("〴."), UBrexLeaf("jan")), "1 jan jan", "1 jan", true},
		{"upto needs one", UBrexUpto(UBrexLeaf("〴."), UBrexLeaf("jan")), "jan", "", false},
		{"upto no end", UBrexUpto(UBrexLeaf("〴."), UBrexLeaf("jan")), "1 feb", "", false},
		{"ascii case", UBrexLeaf("dec"), "DECember", "DEC", true},
		{"non ascii case", UBrexLeaf("é"), "É", "", false},
	} {
		got, ok := accepted(tc.part, tc.input)
		if got != tc.want || ok != tc.ok {
			t.Errorf("%s: %q: accepted %q (%v), want %q (%v)", tc.name, tc.input, got, ok, tc.want, tc.ok)
		}
	}
	m := UBrexUpto(UBrexLeaf("〴D"), UBrexNamed("M", UBrexLeaf("〇+〴le"))).Match(" dec 18")
	if !slices.Equal(m.FindValuesByKey("M"), []string{"dec"}) {
		t.Errorf("upto capture %q", m.FindValuesByKey("M"))
	}
}

// TestUbrexURLBuilder checks the links of upstream's UbrexUrlBuilder, as the
// regex builder reads them. A link without a tooltip or a label uses the
// link, as upstream's Url.
func TestUbrexURLBuilder(t *testing.T) {
	for _, tc := range []struct {
		s, url, tooltip, label string
	}{
		{"[[https://example.com]]", "https://example.com", "https://example.com", "https://example.com"},
		{"[[https://example.com label]]", "https://example.com", "https://example.com", "label"},
		{"[[https://example.com{tip}]]", "https://example.com", "tip", "https://example.com"},
		{"[[https://example.com{tip} label]]", "https://example.com", "tip", "label"},
		{`[["https://example.com/a b"{tip} label]]`, "https://example.com/a b", "tip", "label"},
		{"[[{only tip}]]", "", "only tip", ""},
		{"[[page]]", "top/page", "top/page", "top/page"},
	} {
		u := NewUbrexURLBuilder("top/", URLModeStrict).URL(tc.s)
		if u == nil {
			t.Errorf("%q: no link", tc.s)
			continue
		}
		if u.URL() != tc.url || u.Tooltip() != tc.tooltip || u.Label() != tc.label {
			t.Errorf("%q: %q %q %q, want %q %q %q", tc.s, u.URL(), u.Tooltip(), u.Label(), tc.url, tc.tooltip, tc.label)
		}
	}
	if NewUbrexURLBuilder("", URLModeStrict).URL("x [[a]]") != nil {
		t.Error("strict mode finds a link after text")
	}
	if u := NewUbrexURLBuilder("", URLModeAnywhere).URL("x [[a]] y"); u == nil || u.URL() != "a" {
		t.Errorf("anywhere: %v", u)
	}
}

// TestUBrexNavigator checks the matching of the gantt sentences: at the
// start of a navigator that moves forward.
func TestUBrexNavigator(t *testing.T) {
	n := NewUBrexNavigator("  [Ünïcode task] lasts 5 days")
	spaces := UBrexLeaf("〇+〴s")
	n.Jump(spaces.MatchNavigator(n).AcceptedLength())
	m := UBrexConcat(UBrexLeaf("["), UBrexNamed("SUBJECT", UBrexLeaf("〇+「〤[]」")), UBrexLeaf("]")).MatchNavigator(n)
	if v, ok := m.FindFirstValueByKey("SUBJECT"); !ok || v != "Ünïcode task" || m.AcceptedLength() != 14 {
		t.Fatalf("subject %q %v, length %d", v, ok, m.AcceptedLength())
	}
	c := n.Copy()
	c.Jump(m.AcceptedLength())
	if got := UBrexLeaf("lasts").MatchNavigator(n).AcceptedMatch(); got != "" {
		t.Errorf("the copy moved the navigator: %q", got)
	}
	c.Jump(spaces.MatchNavigator(c).AcceptedLength())
	if got := UBrexLeaf("lasts").MatchNavigator(c); !got.StartMatch() || got.ExactMatch() || got.AcceptedMatch() != "lasts" {
		t.Errorf("verb %q", got.AcceptedMatch())
	}
	e := UBrexNamed("A", UBrexConcat(UBrexNamed("X", UBrexLeaf("a")), UBrexNamed("Y", UBrexLeaf("b")))).Match("ab").ExtractByPrefix("A")
	if x, _ := e.FindFirstValueByKey("X"); x != "a" || len(e.FindValuesByKey("Y")) != 1 {
		t.Errorf("extract %q", x)
	}
}
