// SPDX-License-Identifier: MIT
// Provenance: original

package text

import "testing"

func TestIsJavaWhitespace(t *testing.T) {
	for r, want := range map[rune]bool{
		' ': true, '\t': true, '\n': true, '\v': true, '\f': true, '\r': true, 0x1c: true, 0x1f: true,
		0x2028: true, 0x3000: true, // line separator, ideographic space
		0xa0: false, 0x2007: false, 0x202f: false, // non-breaking spaces
		0x85: false, 'a': false, 0x1d400: false,
	} {
		if got := IsJavaWhitespace(r); got != want {
			t.Errorf("IsJavaWhitespace(%U) = %v, want %v", r, got, want)
		}
	}
}

func TestIsJavaLetterOrDigit(t *testing.T) {
	for r, want := range map[rune]bool{
		'a': true, 'Z': true, '7': true, 'é': true, 'Ж': true, '٣': true,
		'_': false, '~': false, ' ': false,
		0x1d400: false, // supplementary letter: Java sees two surrogates
	} {
		if got := IsJavaLetterOrDigit(r); got != want {
			t.Errorf("IsJavaLetterOrDigit(%U) = %v, want %v", r, got, want)
		}
	}
}

func TestTrinAndBackslash(t *testing.T) {
	if got := Trin(" \t\x01 a b \r\n"); got != "a b" {
		t.Errorf("Trin = %q", got)
	}
	if got := Trin(" a "); got != "a" {
		t.Errorf("Trin must remove no-break spaces, got %q", got)
	}
	for s, want := range map[string]bool{`a\`: true, `a\\`: false, `a`: false, `\`: true} {
		if got := EndsWithBackslash(s); got != want {
			t.Errorf("EndsWithBackslash(%q) = %v", s, got)
		}
	}
}

func TestRemoveInnerComment(t *testing.T) {
	for in, want := range map[string]string{
		"/' c '/ A -> B":     " A -> B",
		"A -> B /' c '/":     "A -> B ",
		"A /'''x-1'''/ -> B": "A  -> B",
		"A -> B":             "A -> B",
		"\t/' c '/A":         "A",
	} {
		if got := NewStringLocated(in, nil).RemoveInnerComment().S; got != want {
			t.Errorf("RemoveInnerComment(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLineLocation(t *testing.T) {
	l := NewLineLocation("", nil)
	if l.Desc != "?" || l.Position != -1 {
		t.Fatalf("new location = %+v", l)
	}
	if n := l.OneLineRead().OneLineRead(); n.Position != 1 || l.Position != -1 {
		t.Fatalf("OneLineRead must not mutate: %+v %+v", n, l)
	}
}

// Trin is upstream's StringUtils.trin, which also trims U+00A0; JavaTrim is
// Java's String.trim, which keeps it.
func TestTrinAndJavaTrim(t *testing.T) {
	if got := Trin("  x\t "); got != "x" {
		t.Errorf("Trin %q", got)
	}
	if got := JavaTrim("  x\t "); got != "  x\t " {
		t.Errorf("JavaTrim %q", got)
	}
	if got := JavaTrim("  x "); got != " x" {
		t.Errorf("JavaTrim %q", got)
	}
	if got := NewStringLocated(" Alice -> Bob ", nil).Trimmed().S; got != "Alice -> Bob" {
		t.Errorf("Trimmed %q", got)
	}
}

// FirstToken is upstream's StringLocated.getFirstToken.
func TestFirstToken(t *testing.T) {
	for s, want := range map[string]string{
		"Title foo": "title", "endif": "endif", "note2": "note", "}": "}", "-> B": "-", "@0": "@",
		"12:00": "1", "éa": "é", "": "", " \t ": "", "  Alice -> Bob": "alice", " Note": "note",
	} {
		if got := NewStringLocated(s, nil).FirstToken(); got != want {
			t.Errorf("FirstToken(%q) = %q, want %q", s, got, want)
		}
	}
}
