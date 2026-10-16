// SPDX-License-Identifier: MIT
// Provenance: original

package svg

import (
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
)

// Upstream's unchecked exceptions are panics with a *color.RuntimeError
// value, the convention of the klimt packages; checked exceptions are
// errors.
const (
	illegalState         = "java.lang.IllegalStateException"
	illegalArgument      = "java.lang.IllegalArgumentException"
	unsupportedOperation = "java.lang.UnsupportedOperationException"
	nullPointer          = "java.lang.NullPointerException"
)

func throw(class, message string) { panic(&color.RuntimeError{Class: class, Message: message}) }

// notPorted panics for a code path whose upstream class is not ported yet,
// naming it, so that the gap is never silent.
func notPorted(upstreamClass, what string) {
	panic(&color.RuntimeError{Class: unsupportedOperation,
		Message: "not ported yet: " + upstreamClass + " (" + what + "); see internal/klimt/svg/doc.go"})
}

// javaLen is Java's String.length: the number of UTF-16 code units.
func javaLen(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x10000 {
			n += 2
		} else {
			n++
		}
	}
	return n
}

// isJavaBlank reports whether s matches Java's "^\s*$": only the ASCII
// whitespace of Java regular expressions, [ \t\n\x0B\f\r].
func isJavaBlank(s string) bool {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case ' ', '\t', '\n', '\x0B', '\f', '\r':
		default:
			return false
		}
	}
	return true
}

// javaEqualsIgnoreCase is Java's String.equalsIgnoreCase.
func javaEqualsIgnoreCase(a, b string) bool {
	if a == b {
		return true
	}
	ra, rb := []rune(a), []rune(b)
	if len(ra) != len(rb) {
		return false
	}
	for i, c1 := range ra {
		c2 := rb[i]
		if c1 == c2 {
			continue
		}
		u1, u2 := unicode.ToUpper(c1), unicode.ToUpper(c2)
		if u1 == u2 || unicode.ToLower(u1) == unicode.ToLower(u2) {
			continue
		}
		return false
	}
	return true
}

// javaToLowerCase is Java's String.toLowerCase for the default (English)
// locale; Go's lower-case mapping matches it except for a few special
// cases (final sigma) that do not matter where it is used.
func javaToLowerCase(s string) string { return strings.ToLower(s) }

// javaUTF16String turns UTF-16 code units into a string as Java's
// getBytes(UTF_8) encodes them: surrogate pairs become their code point and
// lone surrogates become '?'.
func javaUTF16String(units []uint16) string {
	var b strings.Builder
	for i := 0; i < len(units); i++ {
		u := rune(units[i])
		switch {
		case utf16.IsSurrogate(u) && u < 0xDC00 && i+1 < len(units) && units[i+1] >= 0xDC00 && units[i+1] < 0xE000:
			b.WriteRune(utf16.DecodeRune(u, rune(units[i+1])))
			i++
		case utf16.IsSurrogate(u):
			b.WriteByte('?')
		default:
			b.WriteRune(u)
		}
	}
	return b.String()
}

// toUTF16 returns the UTF-16 code units of s.
func toUTF16(s string) []uint16 {
	if !utf8.ValidString(s) {
		s = strings.ToValidUTF8(s, "�")
	}
	return utf16.Encode([]rune(s))
}
