// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/utils/StartUtils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package core

import (
	"strings"
	"unicode/utf8"

	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/text"
)

var (
	patternFilename = regex.MustCompile(`^[@\\]start[^%s{}%g]+[%s{][%s%g]*([^%g]*?)[%s}%g]*$`)
	pausePattern    = regex.MustCompile(`((?:\W|\<[^<>]*\>)*)[@\\]unpause`)
	appendPattern   = regex.MustCompile(`^\W*[@\\](append|a)\b`)
)

// skipJavaWhitespace returns the byte offset of the first rune of s at or
// after from that is not Java whitespace.
func skipJavaWhitespace(s string, from int) int {
	for from < len(s) {
		r, size := utf8.DecodeRuneInString(s[from:])
		if !text.IsJavaWhitespace(r) {
			break
		}
		from += size
	}
	return from
}

// IsStartDirective reports whether s is an @start... or \start... line. The
// keyword is case-sensitive and must be followed by at least one character.
func IsStartDirective(s string) bool {
	i := skipJavaWhitespace(s, 0)
	if i >= len(s) || (s[i] != '@' && s[i] != '\\') {
		return false
	}
	return i+6 < len(s) && s[i+1:i+6] == "start"
}

// startsWithDirectiveKeyword reports whether s, from byte offset from and
// after whitespace, starts with @keyword or \keyword.
func startsWithDirectiveKeyword(s string, from int, keyword string) bool {
	i := skipJavaWhitespace(s, from)
	if i >= len(s) || (s[i] != '@' && s[i] != '\\') {
		return false
	}
	return strings.HasPrefix(s[i+1:], keyword)
}

// BeforeStartUml returns the text before an @start directive on the line,
// when that text only holds non-word characters and <...> tags. It returns
// ok false otherwise. This lets diagrams live inside comments of other files.
func BeforeStartUml(s string) (prefix string, ok bool) {
	inside := false
	for i, r := range s {
		if startsWithDirectiveKeyword(s, i, "start") {
			return s[:i], true
		}
		if inside {
			if r == '>' {
				inside = false
			}
			continue
		}
		if r == '<' {
			inside = true
		} else if r == '~' || r == '_' || text.IsJavaLetterOrDigit(r) {
			return "", false
		}
	}
	return "", false
}

// IsEndDirective reports whether s is an @end... or \end... line.
func IsEndDirective(s string) bool { return startsWithDirectiveKeyword(s, 0, "end") }

// IsPauseDirective reports whether s is an @pause line.
func IsPauseDirective(s string) bool { return startsWithDirectiveKeyword(s, 0, "pause") }

// IsUnpauseDirective reports whether s is an @unpause line.
func IsUnpauseDirective(s string) bool { return startsWithDirectiveKeyword(s, 0, "unpause") }

// IsExit reports whether s is the !exit directive.
func IsExit(s string) bool {
	return strings.TrimFunc(s, text.IsJavaWhitespace) == "!exit"
}

// PossibleAppend returns the text after an @append (or @a) marker, used to
// add lines to a block while it is paused.
func PossibleAppend(sl text.StringLocated) (text.StringLocated, bool) {
	m := appendPattern.FindStringIndex(sl.S)
	if m == nil {
		return text.StringLocated{}, false
	}
	return text.StringLocated{S: text.Trin(sl.S[m[1]:]), Location: sl.Location, PreprocessorError: sl.PreprocessorError}, true
}
