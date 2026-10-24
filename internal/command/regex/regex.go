// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/regex/Pattern2.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

// Package regex compiles upstream's command patterns for Go's RE2 engine.
//
// Upstream patterns are Java regular expressions compiled case-insensitively,
// with macros for character sets: %s (whitespace and no-break space), %q
// (single quotes), %g (double quotes) and %pLN (Unicode letters and digits).
// Java's \s also matches the vertical tab, which RE2's \s does not, so \s is
// rewritten to an explicit set. Java's case-insensitive matching folds ASCII
// letters only, while RE2 uses Unicode simple folding; the difference only
// affects a few letters such as the Kelvin sign.
package regex

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// invisibleQuote is upstream's Jaws.BLOCK_E1_INVISIBLE_QUOTE.
const invisibleQuote = "\uE121"

const javaSpaces = `\t\n\x0B\f\r `

// macroExpansions mirrors Pattern2's QUOTED_REPLACEMENTS.
var macroExpansions = strings.NewReplacer(
	"%pLN", `\p{L}\p{N}`,
	"%s", `\s\x{00A0}`,
	"%q", `'\x{2018}\x{2019}`,
	"%g", `"\x{201C}\x{201D}`+invisibleQuote,
)

// Transform expands upstream's macros and rewrites Java-only escapes, without
// adding the case-insensitive flag.
func Transform(pattern string) (string, error) {
	expanded := macroExpansions.Replace(pattern)
	var b strings.Builder
	inClass := false
	for i := 0; i < len(expanded); i++ {
		c := expanded[i]
		switch {
		case c == '\\' && i+1 < len(expanded):
			next := expanded[i+1]
			i++
			switch {
			case next == 's' && inClass:
				b.WriteString(javaSpaces)
			case next == 's':
				b.WriteString("[" + javaSpaces + "]")
			case next == 'S' && !inClass:
				b.WriteString("[^" + javaSpaces + "]")
			case next == 'S':
				return "", fmt.Errorf(`regex: \S inside a character class is not supported in %q`, pattern)
			default:
				b.WriteByte('\\')
				b.WriteByte(next)
			}
		case c == '[' && !inClass:
			inClass = true
			b.WriteByte(c)
		case c == ']' && inClass:
			inClass = false
			b.WriteByte(c)
		default:
			b.WriteByte(c)
		}
	}
	return b.String(), nil
}

// Compile translates an upstream pattern and compiles it case-insensitively.
func Compile(pattern string) (*regexp.Regexp, error) {
	t, err := Transform(pattern)
	if err != nil {
		return nil, err
	}
	return regexp.Compile("(?i)" + t)
}

// MustCompile is like Compile but panics on error. It is meant for patterns
// that are constants of the program.
func MustCompile(pattern string) *regexp.Regexp {
	re, err := Compile(pattern)
	if err != nil {
		panic(err)
	}
	return re
}

var cache sync.Map // pattern -> *regexp.Regexp

// Cached returns the compiled pattern, compiling it once.
func Cached(pattern string) *regexp.Regexp {
	if re, ok := cache.Load(pattern); ok {
		return re.(*regexp.Regexp)
	}
	re, _ := cache.LoadOrStore(pattern, MustCompile(pattern))
	return re.(*regexp.Regexp)
}
