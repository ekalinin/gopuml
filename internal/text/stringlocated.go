// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/text/StringLocated.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/utils/LineLocationImpl.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/StringUtils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package text

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// LineLocation is the position of a line in its source.
type LineLocation struct {
	// Desc describes the source, for example a file name or "string".
	Desc string
	// Position is the 0-based index of the line in the source.
	Position int
	// Parent is the location of the directive that included this source.
	Parent *LineLocation
}

// NewLineLocation returns the location before the first line of a source.
func NewLineLocation(desc string, parent *LineLocation) *LineLocation {
	if desc == "" {
		desc = "?"
	}
	return &LineLocation{Desc: desc, Position: -1, Parent: parent}
}

// OneLineRead returns the location of the next line.
func (l *LineLocation) OneLineRead() *LineLocation {
	return &LineLocation{Desc: l.Desc, Position: l.Position + 1, Parent: l.Parent}
}

// StringLocated is a source line with its location.
type StringLocated struct {
	S                 string
	Location          *LineLocation
	PreprocessorError string
}

// NewStringLocated returns a located line.
func NewStringLocated(s string, location *LineLocation) StringLocated {
	return StringLocated{S: s, Location: location}
}

// Substring returns the line from byte offset start to end.
func (sl StringLocated) Substring(start, end int) StringLocated {
	return StringLocated{S: sl.S[start:end], Location: sl.Location, PreprocessorError: sl.PreprocessorError}
}

// MergeEndBackslash joins a line ending with a backslash with the next one.
func (sl StringLocated) MergeEndBackslash(next StringLocated) StringLocated {
	if !EndsWithBackslash(sl.S) {
		panic("text: MergeEndBackslash on a line without a trailing backslash")
	}
	return StringLocated{S: sl.S[:len(sl.S)-1] + next.S, Location: sl.Location, PreprocessorError: sl.PreprocessorError}
}

// Trimmed returns the line without leading and trailing characters <= ' '.
func (sl StringLocated) Trimmed() StringLocated {
	return StringLocated{S: Trin(sl.S), Location: sl.Location, PreprocessorError: sl.PreprocessorError}
}

var specialInnerComment = regexp.MustCompile(`/'''[-\w]*'''/`)

func removeSpecialInnerComment(s string) string {
	if strings.Contains(s, "/'''") && strings.Contains(s, "'''/") {
		return specialInnerComment.ReplaceAllString(s, "")
	}
	return s
}

// RemoveInnerComment removes a block comment at the start or the end of the
// line, and special inline markers anywhere. Both forms are shown below:
//
//	/' comment '/ code
//	code /'''name'''/ code
func (sl StringLocated) RemoveInnerComment() StringLocated {
	trim := strings.TrimFunc(strings.ReplaceAll(sl.S, "\t", " "), isJavaTrimSpace)
	if strings.HasPrefix(trim, "/'") {
		if idx := strings.Index(sl.S, "'/"); idx != -1 {
			return StringLocated{S: removeSpecialInnerComment(sl.S[idx+2:]), Location: sl.Location, PreprocessorError: sl.PreprocessorError}
		}
	}
	if strings.HasSuffix(trim, "'/") {
		if idx := strings.LastIndex(sl.S, "/'"); idx != -1 {
			return StringLocated{S: removeSpecialInnerComment(sl.S[:idx]), Location: sl.Location, PreprocessorError: sl.PreprocessorError}
		}
	}
	if strings.Contains(trim, "/'''") && strings.Contains(trim, "'''/") {
		return StringLocated{S: removeSpecialInnerComment(sl.S), Location: sl.Location, PreprocessorError: sl.PreprocessorError}
	}
	return sl
}

// FirstToken is the token the command index files the line under,
// upstream's getFirstToken: once the leading blanks are skipped (those of
// StringUtils.isTrimmable), the run of ASCII letters the line starts with,
// lower-cased, or its first character alone, or "" for a blank line.
func (sl StringLocated) FirstToken() string {
	s := strings.TrimLeftFunc(sl.S, isTrimmable)
	if s == "" {
		return ""
	}
	end := 0
	for end < len(s) && (s[end] >= 'a' && s[end] <= 'z' || s[end] >= 'A' && s[end] <= 'Z') {
		end++
	}
	if end > 0 {
		return strings.ToLower(s[:end])
	}
	_, size := utf8.DecodeRuneInString(s)
	return s[:size]
}

// isJavaTrimSpace matches the characters removed by Java's String.trim.
func isJavaTrimSpace(r rune) bool { return r <= ' ' }

// isTrimmable is upstream's StringUtils.isTrimmable: a character up to the
// space, or U+00A0.
func isTrimmable(r rune) bool { return r <= ' ' || r == '\u00A0' }

// Trin removes the leading and trailing characters up to the space, and
// U+00A0, like upstream's StringUtils.trin.
func Trin(s string) string { return strings.TrimFunc(s, isTrimmable) }

// JavaTrim removes the leading and trailing characters up to the space,
// like Java's String.trim (U+00A0 is kept).
func JavaTrim(s string) string { return strings.TrimFunc(s, isJavaTrimSpace) }

// EndsWithBackslash reports whether s ends with a single backslash.
func EndsWithBackslash(s string) bool {
	return strings.HasSuffix(s, `\`) && !strings.HasSuffix(s, `\\`)
}
