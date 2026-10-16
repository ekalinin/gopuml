// SPDX-License-Identifier: MIT
// Provenance: original

// Package text holds located source lines and string helpers with the exact
// semantics of the Java methods upstream relies on.
//
// Java strings are sequences of UTF-16 code units, so Java character tests
// see a supplementary character as two surrogate units that are neither
// letters nor whitespace. The helpers below reproduce that for runes above
// U+FFFF.
package text

import "unicode"

// IsJavaWhitespace reports whether r is whitespace according to Java's
// Character.isWhitespace: Unicode space, line or paragraph separators except
// the non-breaking ones (U+00A0, U+2007, U+202F), plus the ASCII controls
// U+0009..U+000D and U+001C..U+001F.
func IsJavaWhitespace(r rune) bool {
	switch {
	case r >= '\t' && r <= '\r', r >= 0x1c && r <= 0x1f:
		return true
	case r == 0xa0 || r == 0x2007 || r == 0x202f || r > 0xffff:
		return false
	}
	return unicode.In(r, unicode.Zs, unicode.Zl, unicode.Zp)
}

// IsJavaLetterOrDigit reports whether r is a letter or digit according to
// Java's Character.isLetterOrDigit applied to a single UTF-16 unit.
func IsJavaLetterOrDigit(r rune) bool {
	if r > 0xffff {
		return false
	}
	return unicode.IsLetter(r) || unicode.Is(unicode.Nd, r)
}
