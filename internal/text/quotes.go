// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/StringUtils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package text

import (
	"strings"
	"unicode/utf16"
)

// isDoubleQuote is StringUtils.isDoubleQuote.
func isDoubleQuote(c uint16) bool {
	return c == '"' || c == '“' || c == '”' || c == '«' || c == '»'
}

// RemoveQuotesFormat is StringUtils.eventuallyRemoveStartingAndEndingDoubleQuote
// with a format: it removes the enclosing characters of s that format
// allows, double quotes (any quote character of isDoubleQuote),
// parentheses, brackets or colons.
func RemoveQuotesFormat(s, format string) string {
	u := utf16.Encode([]rune(s))
	if len(u) < 2 {
		return s
	}
	inner := func() string { return string(utf16.Decode(u[1 : len(u)-1])) }
	switch {
	case strings.Contains(format, `"`) && isDoubleQuote(u[0]) && isDoubleQuote(u[len(u)-1]),
		strings.Contains(format, "(") && strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")"),
		strings.Contains(format, "[") && strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]"),
		strings.Contains(format, ":") && strings.HasPrefix(s, ":") && strings.HasSuffix(s, ":"):
		return inner()
	}
	return s
}

// RemoveQuotes is StringUtils.eventuallyRemoveStartingAndEndingDoubleQuote:
// it removes enclosing double quotes, parentheses, brackets or colons.
func RemoveQuotes(s string) string { return RemoveQuotesFormat(s, `"([:`) }
