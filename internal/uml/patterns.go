// SPDX-License-Identifier: MIT
// Provenance: original

package uml

import (
	"regexp"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/command/regex"
)

// javaDot is the character class of Java's regex ".": any character but
// the line terminators, which RE2's "." also matches (except '\n').
const javaDot = `[^\n\r\x{85}\x{2028}\x{2029}]`

// JavaDots rewrites the "." of an upstream pattern outside character
// classes so that it does not match line terminators, as in Java.
func JavaDots(pattern string) string {
	if !strings.Contains(pattern, ".") {
		return pattern
	}
	var sb strings.Builder
	inClass := false
	for i := 0; i < len(pattern); i++ {
		c := pattern[i]
		switch {
		case c == '\\' && i+1 < len(pattern):
			sb.WriteByte(c)
			i++
			sb.WriteByte(pattern[i])
			continue
		case c == '[' && !inClass:
			inClass = true
		case c == ']' && inClass:
			inClass = false
		case c == '.' && !inClass:
			sb.WriteString(javaDot)
			continue
		}
		sb.WriteByte(c)
	}
	return sb.String()
}

// Leaf is regex.NewNamedLeaf with Java's ".".
func Leaf(count int, name, pattern string) *regex.Leaf {
	return regex.NewNamedLeaf(count, name, JavaDots(pattern))
}

// Lit is regex.NewLeaf with Java's ".".
func Lit(pattern string) *regex.Leaf { return regex.NewLeaf(JavaDots(pattern)) }

// Compile is regex.MustCompile (upstream's Pattern2.cmpile) with Java's ".".
func Compile(pattern string) *regexp.Regexp { return regex.MustCompile(JavaDots(pattern)) }
