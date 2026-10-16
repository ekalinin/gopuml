// SPDX-License-Identifier: MIT
// Provenance: original

package tim

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// Java strings are sequences of UTF-16 code units. The preprocessor works
// char by char (positions, look-ahead, substrings), so the eaters operate on
// []uint16 and convert to Go strings at their boundaries. A lone surrogate,
// which a Go string cannot hold, becomes U+FFFD.

// toUTF16 returns the UTF-16 code units of s.
func toUTF16(s string) []uint16 { return utf16.Encode([]rune(s)) }

// fromUTF16 returns the Go string of UTF-16 code units.
func fromUTF16(u []uint16) string {
	ascii := true
	for _, c := range u {
		if c >= 0x80 {
			ascii = false
			break
		}
	}
	if ascii {
		b := make([]byte, len(u))
		for i, c := range u {
			b[i] = byte(c)
		}
		return string(b)
	}
	return string(utf16.Decode(u))
}

// javaLength is Java's String.length: the number of UTF-16 code units.
func javaLength(s string) int {
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

// javaCompare is the sign of Java's String.compareTo: lexicographic order of
// the UTF-16 code units, then length.
func javaCompare(a, b string) int {
	if a == b {
		return 0
	}
	return slices.Compare(toUTF16(a), toUTF16(b))
}

// javaHashCode is Java's String.hashCode.
func javaHashCode(s string) int32 {
	var h int32
	for _, c := range toUTF16(s) {
		h = 31*h + int32(c)
	}
	return h
}

// isJavaLetter is Character.isLetter(char).
func isJavaLetter(ch uint16) bool { return unicode.IsLetter(rune(ch)) }

// isJavaSpaceChar is Character.isSpaceChar(char): a space, line or paragraph
// separator.
func isJavaSpaceChar(ch uint16) bool {
	return unicode.In(rune(ch), unicode.Zs, unicode.Zl, unicode.Zp)
}

// isJavaWhitespace is Character.isWhitespace(char).
func isJavaWhitespace(ch uint16) bool { return text.IsJavaWhitespace(rune(ch)) }

// isJavaLetterOrDigit is Character.isLetterOrDigit(char).
func isJavaLetterOrDigit(ch uint16) bool { return text.IsJavaLetterOrDigit(rune(ch)) }

// isJavaIdentifierPart is Character.isJavaIdentifierPart(char).
func isJavaIdentifierPart(ch uint16) bool {
	r := rune(ch)
	if unicode.In(r, unicode.L, unicode.Sc, unicode.Pc, unicode.Nd, unicode.Nl, unicode.Mn, unicode.Mc) {
		return true
	}
	// Character.isIdentifierIgnorable
	return r <= 8 || (r >= 0x0e && r <= 0x1b) || (r >= 0x7f && r <= 0x9f) || unicode.Is(unicode.Cf, r)
}

// javaDigit is Character.digit(char, 10): the value of a decimal digit of
// any script, or -1.
func javaDigit(ch uint16) int {
	r := rune(ch)
	if r >= '0' && r <= '9' {
		return int(r - '0')
	}
	if !unicode.Is(unicode.Nd, r) {
		return -1
	}
	// Decimal digits come in contiguous runs of ten, from zero to nine.
	n := 0
	for unicode.Is(unicode.Nd, r-1) {
		r--
		n++
	}
	return n % 10
}

// javaParseInt is Integer.parseInt(s): an optional sign followed by decimal
// digits of any script, within the int range. ok is false where Java throws
// NumberFormatException.
func javaParseInt(s string) (v int32, ok bool) {
	u := toUTF16(s)
	if len(u) == 0 {
		return 0, false
	}
	i, neg := 0, false
	if u[0] == '-' || u[0] == '+' {
		neg = u[0] == '-'
		i = 1
		if len(u) == 1 {
			return 0, false
		}
	}
	var acc int64
	for ; i < len(u); i++ {
		d := javaDigit(u[i])
		if d < 0 {
			return 0, false
		}
		acc = acc*10 + int64(d)
		if acc > 1<<31 {
			return 0, false
		}
	}
	if neg {
		acc = -acc
	}
	if acc > 1<<31-1 || acc < -(1<<31) {
		return 0, false
	}
	return int32(acc), true
}

// javaSplitNewline is Java's s.split("\n"): trailing empty strings are
// removed, so "a\n" gives ["a"] and "\n" gives no element at all.
func javaSplitNewline(s string) []string {
	if s == "" {
		return []string{""}
	}
	parts := strings.Split(s, "\n")
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

// javaToUpperCase is Java's String.toUpperCase in a non-Turkish locale,
// including the mappings of one character to several (for example ß to SS).
func javaToUpperCase(s string) string {
	var b strings.Builder
	for _, r := range s {
		if m, ok := upperSpecialCasing[r]; ok {
			b.WriteString(m)
			continue
		}
		b.WriteRune(unicode.ToUpper(r))
	}
	return b.String()
}

// upperSpecialCasing holds the unconditional one-to-many upper case mappings
// of Unicode's SpecialCasing.txt, which Java applies in String.toUpperCase.
var upperSpecialCasing = map[rune]string{
	0x00DF: "SS", 0x0149: "ʼN", 0x01F0: "J̌", 0x0390: "Ϊ́",
	0x03B0: "Ϋ́", 0x0587: "ԵՒ", 0x1E96: "H̱", 0x1E97: "T̈",
	0x1E98: "W̊", 0x1E99: "Y̊", 0x1E9A: "Aʾ", 0x1F50: "Υ̓",
	0x1F52: "Υ̓̀", 0x1F54: "Υ̓́", 0x1F56: "Υ̓͂",
	0x1F80: "ἈΙ", 0x1F81: "ἉΙ", 0x1F82: "ἊΙ", 0x1F83: "ἋΙ",
	0x1F84: "ἌΙ", 0x1F85: "ἍΙ", 0x1F86: "ἎΙ", 0x1F87: "ἏΙ",
	0x1F88: "ἈΙ", 0x1F89: "ἉΙ", 0x1F8A: "ἊΙ", 0x1F8B: "ἋΙ",
	0x1F8C: "ἌΙ", 0x1F8D: "ἍΙ", 0x1F8E: "ἎΙ", 0x1F8F: "ἏΙ",
	0x1F90: "ἨΙ", 0x1F91: "ἩΙ", 0x1F92: "ἪΙ", 0x1F93: "ἫΙ",
	0x1F94: "ἬΙ", 0x1F95: "ἭΙ", 0x1F96: "ἮΙ", 0x1F97: "ἯΙ",
	0x1F98: "ἨΙ", 0x1F99: "ἩΙ", 0x1F9A: "ἪΙ", 0x1F9B: "ἫΙ",
	0x1F9C: "ἬΙ", 0x1F9D: "ἭΙ", 0x1F9E: "ἮΙ", 0x1F9F: "ἯΙ",
	0x1FA0: "ὨΙ", 0x1FA1: "ὩΙ", 0x1FA2: "ὪΙ", 0x1FA3: "ὫΙ",
	0x1FA4: "ὬΙ", 0x1FA5: "ὭΙ", 0x1FA6: "ὮΙ", 0x1FA7: "ὯΙ",
	0x1FA8: "ὨΙ", 0x1FA9: "ὩΙ", 0x1FAA: "ὪΙ", 0x1FAB: "ὫΙ",
	0x1FAC: "ὬΙ", 0x1FAD: "ὭΙ", 0x1FAE: "ὮΙ", 0x1FAF: "ὯΙ",
	0x1FB2: "ᾺΙ", 0x1FB3: "ΑΙ", 0x1FB4: "ΆΙ", 0x1FB6: "Α͂",
	0x1FB7: "Α͂Ι", 0x1FBC: "ΑΙ", 0x1FC2: "ῊΙ", 0x1FC3: "ΗΙ",
	0x1FC4: "ΉΙ", 0x1FC6: "Η͂", 0x1FC7: "Η͂Ι", 0x1FCC: "ΗΙ",
	0x1FD2: "Ϊ̀", 0x1FD3: "Ϊ́", 0x1FD6: "Ι͂",
	0x1FD7: "Ϊ͂", 0x1FE2: "Ϋ̀", 0x1FE3: "Ϋ́",
	0x1FE4: "Ρ̓", 0x1FE6: "Υ͂", 0x1FE7: "Ϋ͂",
	0x1FF2: "ῺΙ", 0x1FF3: "ΩΙ", 0x1FF4: "ΏΙ", 0x1FF6: "Ω͂",
	0x1FF7: "Ω͂Ι", 0x1FFC: "ΩΙ", 0xFB00: "FF", 0xFB01: "FI", 0xFB02: "FL",
	0xFB03: "FFI", 0xFB04: "FFL", 0xFB05: "ST", 0xFB06: "ST", 0xFB13: "ՄՆ",
	0xFB14: "ՄԵ", 0xFB15: "ՄԻ", 0xFB16: "ՎՆ", 0xFB17: "ՄԽ",
}

// javaToLowerCase is Java's String.toLowerCase in a non-Turkish,
// non-Lithuanian locale: the simple lower case mapping of each character,
// except U+0130, which becomes "i" followed by U+0307, and the capital sigma,
// which becomes a final sigma at the end of a word.
func javaToLowerCase(s string) string {
	if !strings.ContainsAny(s, "İΣ") {
		return strings.Map(unicode.ToLower, s)
	}
	rs := []rune(s)
	var b strings.Builder
	for i, r := range rs {
		switch r {
		case 0x130:
			b.WriteString("i̇")
		case 0x3a3:
			if isFinalCased(rs, i) {
				b.WriteRune(0x3c2)
			} else {
				b.WriteRune(0x3c3)
			}
		default:
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

// isFinalCased is the Final_Cased condition of Java's
// ConditionalSpecialCasing: within the word that holds rs[index], a cased
// letter comes before it and none comes after it. Java finds the word with
// a word BreakIterator; here a word is a run of letters, marks, digits,
// connector punctuation and apostrophes, which covers the usual cases.
func isFinalCased(rs []rune, index int) bool {
	before := false
	for i := index - 1; i >= 0 && isWordRune(rs[i]); i-- {
		if isCased(rs[i]) {
			before = true
			break
		}
	}
	if !before {
		return false
	}
	for i := index + 1; i < len(rs) && isWordRune(rs[i]); i++ {
		if isCased(rs[i]) {
			return false
		}
	}
	return true
}

func isWordRune(r rune) bool {
	return r == '\'' || r == 0x2019 || unicode.In(r, unicode.L, unicode.M, unicode.Nd, unicode.Pc)
}

// isCased is ConditionalSpecialCasing.isCased: a lower, upper or title case
// letter, or a character with the Other_Lowercase or Other_Uppercase
// property.
func isCased(r rune) bool {
	return unicode.In(r, unicode.Ll, unicode.Lu, unicode.Lt, unicode.Other_Lowercase, unicode.Other_Uppercase)
}

// javaEqualsIgnoreCase is Java's String.equalsIgnoreCase: the strings have
// the same length and their characters are equal after conversion to upper
// case, or to upper then lower case.
func javaEqualsIgnoreCase(a, b string) bool {
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
		if u1 != u2 && unicode.ToLower(u1) != unicode.ToLower(u2) {
			return false
		}
	}
	return true
}

// javaIndexOf is Java's String.indexOf(String): the index in UTF-16 code
// units of the first occurrence of sub in s, or -1.
func javaIndexOf(s, sub string) int {
	i := strings.Index(s, sub)
	if i < 0 {
		return -1
	}
	return javaLength(s[:i])
}

// javaDigitRadix is Character.digit(char, radix): the value of a decimal
// digit of any script, of an ASCII or full-width Latin letter, or -1 when
// the value is not below radix.
func javaDigitRadix(ch uint16, radix int) int {
	v := -1
	switch {
	case ch >= 'a' && ch <= 'z':
		v = int(ch-'a') + 10
	case ch >= 'A' && ch <= 'Z':
		v = int(ch-'A') + 10
	case ch >= 0xff41 && ch <= 0xff5a:
		v = int(ch-0xff41) + 10
	case ch >= 0xff21 && ch <= 0xff3a:
		v = int(ch-0xff21) + 10
	default:
		v = javaDigit(ch)
	}
	if v >= radix {
		return -1
	}
	return v
}

// javaParseIntRadix is Integer.parseInt(s, radix). ok is false where Java
// throws NumberFormatException.
func javaParseIntRadix(s string, radix int) (v int32, ok bool) {
	u := toUTF16(s)
	if len(u) == 0 {
		return 0, false
	}
	i, neg := 0, false
	if u[0] < '0' {
		if u[0] != '-' && u[0] != '+' {
			return 0, false
		}
		neg = u[0] == '-'
		i = 1
		if len(u) == 1 {
			return 0, false
		}
	}
	var acc int64
	for ; i < len(u); i++ {
		d := javaDigitRadix(u[i], radix)
		if d < 0 {
			return 0, false
		}
		acc = acc*int64(radix) + int64(d)
		if acc > 1<<31 {
			return 0, false
		}
	}
	if neg {
		acc = -acc
	}
	if acc > 1<<31-1 {
		return 0, false
	}
	return int32(acc), true
}
