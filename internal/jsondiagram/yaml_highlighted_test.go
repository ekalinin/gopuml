// SPDX-License-Identifier: MIT
// Provenance: original

package jsondiagram

import (
	"slices"
	"testing"

	"github.com/ekalinin/gopuml/internal/klimt/color"
)

func TestBuildHighlighted(t *testing.T) {
	h := BuildHighlighted(`#highlight "a" / "b"`)
	if !slices.Equal(h.paths, []string{"a", "b"}) || h.Stereotype() != nil {
		t.Errorf(`"a" / "b": paths %q, stereotype %v`, h.paths, h.Stereotype())
	}
	h = BuildHighlighted(`#highlight "a" <<h1>>`)
	if !slices.Equal(h.paths, []string{"a"}) || h.Stereotype() == nil || h.Stereotype().String() != "<<h1>>" {
		t.Errorf(`"a" <<h1>>: paths %q, stereotype %v`, h.paths, h.Stereotype())
	}
	if !MatchesHighlightedDefinition(`#highlight "a"`) || MatchesHighlightedDefinition(`#foo bar`) {
		t.Error("MatchesHighlightedDefinition does not recognize the #highlight lines")
	}
}

func TestBuildHighlightedThrows(t *testing.T) {
	for _, line := range []string{`#foo bar`, `#highlight <<h1>>`} {
		func() {
			defer func() {
				if e, ok := recover().(*color.RuntimeError); !ok || e.Class != "java.lang.IllegalStateException" {
					t.Errorf("%q: recovered %v, want an IllegalStateException", line, e)
				}
			}()
			BuildHighlighted(line)
		}()
	}
}

func TestHighlightedUpOneLevel(t *testing.T) {
	for _, tc := range []struct {
		line, key string
		want      []string // nil: no path
	}{
		{`#highlight "a" / "b"`, "a", []string{"b"}},
		{`#highlight "*" / "c"`, "k", []string{"c"}},
		{`#highlight "**" / "d"`, "m", []string{"**", "d"}},
		{`#highlight "x" / "b"`, "a", nil},
		{`#highlight "a"`, "a", nil},
	} {
		got := BuildHighlighted(tc.line).upOneLevel(tc.key)
		switch {
		case tc.want == nil && got != nil:
			t.Errorf("%s up %q: %q, want nil", tc.line, tc.key, got.paths)
		case tc.want != nil && (got == nil || !slices.Equal(got.paths, tc.want)):
			t.Errorf("%s up %q: %v, want %q", tc.line, tc.key, got, tc.want)
		}
	}
}

func TestHighlightedIsKeyHighlight(t *testing.T) {
	for _, tc := range []struct {
		line string
		want bool
	}{
		{`#highlight "**" / "d"`, true},
		{`#highlight "d"`, true},
		{`#highlight "a" / "d"`, false},
		{`#highlight "e"`, false},
	} {
		if got := BuildHighlighted(tc.line).isKeyHighlight("d"); got != tc.want {
			t.Errorf("%s isKeyHighlight(d) = %v, want %v", tc.line, got, tc.want)
		}
	}
}
