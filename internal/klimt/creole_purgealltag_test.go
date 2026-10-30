// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"slices"
	"testing"
)

// TestSplitterPurgeAllTag checks purgeAllTag against the results of
// upstream's Splitter.purgeAllTag on the same inputs.
func TestSplitterPurgeAllTag(t *testing.T) {
	tests := []struct{ in, want string }{
		{"<b>ab</b>", "ab"},
		{"<color:red>x</color>", "x"},
		{"<size:14>big", "big"},
		{"<&person> user", "<&person> user"},
		{"[[http://plantuml.com PlantUML]] after", " after"},
		// The strike activation pattern is written in ubrex syntax, which
		// matches nothing as a regular expression.
		{"<i>it</i><u>u</u><s>s</s>", "itu<s>s"},
		{"plain text", "plain text"},
		{"<font color=red>f</font>", "f"},
		{"<B>x</B>", "x"},
		{"<sup>2</sup>", "2"},
		{"<img:a.png>", ""},
		// String.replaceAll does not expand upstream's macros.
		{"%s<b>", "%s"},
	}
	for _, tt := range tests {
		if got := SplitterPurgeAllTag(tt.in); got != tt.want {
			t.Errorf("SplitterPurgeAllTag(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestJavaTokenizeExported(t *testing.T) {
	if got, want := JavaTokenizeWithDelims("a | b}c", "|}"), []string{"a ", "|", " b", "}", "c"}; !slices.Equal(got, want) {
		t.Errorf("JavaTokenizeWithDelims = %q, want %q", got, want)
	}
	if got, want := JavaTokenize("^a^^b^", "^"), []string{"a", "b"}; !slices.Equal(got, want) {
		t.Errorf("JavaTokenize = %q, want %q", got, want)
	}
}
