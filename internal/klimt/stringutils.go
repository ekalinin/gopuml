// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/StringUtils.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/utils/SignatureUtils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"crypto/md5"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"
)

// The private-use characters that upstream's StringUtils substitutes for
// visibility modifiers.
const (
	charPrivateField         = ''
	charProtectedField       = ''
	charPackagePrivateField  = ''
	charPublicField          = ''
	charPrivateMethod        = ''
	charProtectedMethod      = ''
	charPackagePrivateMethod = ''
	charPublicMethod         = ''
	charIEMandatory          = ''
)

// isDoubleQuote is upstream's StringUtils.isDoubleQuote.
func isDoubleQuote(c rune) bool {
	return c == '"' || c == '“' || c == '”' || c == '«' || c == '»'
}

// eventuallyRemoveStartingAndEndingDoubleQuoteFormat removes the enclosing
// characters of s that format allows: double quotes (any of the quote
// characters of isDoubleQuote), parentheses, brackets or colons.
func eventuallyRemoveStartingAndEndingDoubleQuoteFormat(s, format string) string {
	u := utf16.Encode([]rune(s))
	if len(u) < 2 {
		return s
	}
	inner := func() string { return string(utf16.Decode(u[1 : len(u)-1])) }
	if strings.Contains(format, "\"") && isDoubleQuote(rune(u[0])) && isDoubleQuote(rune(u[len(u)-1])) {
		return inner()
	}
	if strings.Contains(format, "(") && strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		return inner()
	}
	if strings.Contains(format, "[") && strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		return inner()
	}
	if strings.Contains(format, ":") && strings.HasPrefix(s, ":") && strings.HasSuffix(s, ":") {
		return inner()
	}
	return s
}

// eventuallyRemoveStartingAndEndingDoubleQuote removes enclosing double
// quotes, parentheses, brackets or colons.
func eventuallyRemoveStartingAndEndingDoubleQuote(s string) string {
	return eventuallyRemoveStartingAndEndingDoubleQuoteFormat(s, "\"([:")
}

// isEmptyJava is upstream's StringUtils.isEmpty: only spaces, tabs, line
// breaks and NUL characters.
func isEmptyJava(s string) bool {
	for _, c := range s {
		if !(c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == 0) {
			return false
		}
	}
	return true
}

var unicodeUplus = regexp.MustCompile(`<U\+([0-9a-fA-F]{4,5})>`)

// manageUnicodeNotationUplus replaces "<U+hhhh>" by the character.
func manageUnicodeNotationUplus(s string) string {
	if !strings.Contains(s, "<U+") {
		return s
	}
	return unicodeUplus.ReplaceAllStringFunc(s, func(m string) string {
		v, _ := strconv.ParseInt(m[3:len(m)-1], 16, 32)
		return string(rune(v))
	})
}

// md5Hex is upstream's SignatureUtils.getMD5Hex: the MD5 of the UTF-8
// bytes, in lower-case hexadecimal.
func md5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
