// SPDX-License-Identifier: MIT
// Provenance: original

package regex

import (
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// unescapeGolden decodes the \uXXXX escapes of first_tokens.golden.
func unescapeGolden(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+5 < len(s) && s[i+1] == 'u' {
			if v, err := strconv.ParseUint(s[i+2:i+6], 16, 32); err == nil {
				sb.WriteRune(rune(v))
				i += 5
				continue
			}
		}
		sb.WriteByte(s[i])
	}
	return sb.String()
}

// FirstTokens reads, from every starting pattern of the commands of the
// supported factories, the tokens upstream's FirstTokens.from reads (the
// last column of testdata/first_tokens.golden).
func TestFirstTokensOfUpstreamPatterns(t *testing.T) {
	data, err := os.ReadFile("../testdata/first_tokens.golden")
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for line := range strings.SplitSeq(strings.TrimRight(string(data), "\n"), "\n") {
		f := strings.Split(line, "\t")
		if f[3] == "-" {
			continue
		}
		pattern := unescapeGolden(f[3])
		want := "ANYTHING"
		if got := FirstTokens(pattern); got != nil {
			var parts []string
			for _, tok := range got {
				if tok == "" {
					tok = "<empty>"
				}
				parts = append(parts, tok)
			}
			slices.Sort(parts)
			if strings.Join(parts, " ") != unescapeGolden(f[6]) {
				t.Errorf("%s %s: tokens %q, upstream %s", f[0], f[2], parts, f[6])
			}
		} else if f[6] != want {
			t.Errorf("%s %s: any line, upstream %s", f[0], f[2], f[6])
		}
		n++
	}
	if n < 400 {
		t.Errorf("only %d patterns", n)
	}
}

func TestFirstTokensReader(t *testing.T) {
	for pattern, want := range map[string][]string{
		"^title$":                        {"title"},
		"^end[%s]*split$":                {"end", "endsplit"},
		"^(?:(left|right))?[%s]*footer$": {"footer", "left", "leftfooter", "right", "rightfooter"},
		"^[%s]*\\}$":                     {"}"},
		"^\\d+$":                         {"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"},
		"^(?=a)ab$":                      {"ab"},
		"^[hx]-axis$":                    {"h", "x"},
		"^a?$":                           {"", "a"},
	} {
		if got := FirstTokens(pattern); !slices.Equal(got, want) {
			t.Errorf("%q: %q, want %q", pattern, got, want)
		}
	}
	for _, pattern := range []string{"title$", "^.*$", "^[^a]$", "^(?<x>a)$", "^%pLN$", "^(ab)+$"} {
		if got := FirstTokens(pattern); got != nil {
			t.Errorf("%q: %q, want any line", pattern, got)
		}
	}
}
