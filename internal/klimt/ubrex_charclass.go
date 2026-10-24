// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/ChallengeAlternative.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/ChallengeCharClass.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/ChallengeCharSet.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/CaseMode.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/CharClass.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/CharClassRaw.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/CharClassType.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/CharSet.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"slices"
	"unicode"
	"unicode/utf16"
)

// ubrexAlternative is upstream's ChallengeAlternative: the first
// alternative that matches.
type ubrexAlternative struct {
	alternatives []ubrexChallenge
}

func (a *ubrexAlternative) run(s *ubrexText, position int) ubrexResult {
	for _, pattern := range a.alternatives {
		if r := pattern.run(s, position); r.length >= 0 {
			return r
		}
	}
	return ubrexResultNoMatch
}

// ubrexCharClassRaw is upstream's CharClassRaw.
type ubrexCharClassRaw int

const (
	ubrexAny ubrexCharClassRaw = iota
	ubrexSpace
	ubrexGuillemet
	ubrexWord
	ubrexDigit
	ubrexAlphaNumeric
	ubrexLetter
)

func (c ubrexCharClassRaw) definitionLength() int {
	switch c {
	case ubrexAlphaNumeric:
		return len("an")
	case ubrexLetter:
		return len("le")
	}
	return 1
}

func (c ubrexCharClassRaw) internalMatches(ch uint16) bool {
	switch c {
	case ubrexAny:
		return true
	case ubrexSpace:
		return ch == ' '
	case ubrexGuillemet:
		return ch == '"' || ch == '“' || ch == '”'
	case ubrexWord:
		return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '_'
	case ubrexDigit:
		return ch >= '0' && ch <= '9' || ch == '_'
	case ubrexAlphaNumeric:
		return javaIsLetter(ch) || unicode.IsDigit(rune(ch))
	case ubrexLetter:
		return javaIsLetter(ch)
	}
	return false
}

// javaIsLetter is Java's Character.isLetter(char): surrogates are not
// letters.
func javaIsLetter(ch uint16) bool { return !utf16.IsSurrogate(rune(ch)) && unicode.IsLetter(rune(ch)) }

func ubrexCharClassRawFromDefinition(nav *ubrexText) ubrexCharClassRaw {
	switch ch := nav.charAt(0); ch {
	case 'S', 's':
		return ubrexSpace
	case 'G', 'g':
		return ubrexGuillemet
	case 'D', 'd':
		return ubrexDigit
	case 'W', 'w':
		return ubrexWord
	case 'A', 'a':
		// Not finished because we don't test the "n"
		return ubrexAlphaNumeric
	case 'L', 'l':
		// Not finished because we don't test the "e"
		return ubrexLetter
	case '.':
		return ubrexAny
	default:
		throw(unsupportedOperation, "wip01class "+string(rune(ch)))
	}
	return ubrexAny
}

// ubrexCharClass is upstream's ChallengeCharClass over a CharClass: an
// upper-case class letter negates the class.
type ubrexCharClass struct {
	raw      ubrexCharClassRaw
	negative bool
}

func (c *ubrexCharClass) run(s *ubrexText, position int) ubrexResult {
	if s.length() == position {
		return ubrexResultNoMatch
	}
	if c.raw.internalMatches(s.charAt(position)) != c.negative {
		return ubrexResultOne
	}
	return ubrexResultNoMatch
}

// ubrexEnsureLowercase is upstream's CaseMode.ensureLowercase.
func ubrexEnsureLowercase(ch uint16) uint16 {
	if ch >= 'A' && ch <= 'Z' {
		return ch - 'A' + 'a'
	}
	return ch
}

// javaShl is Java's shift of a long, which only uses the six low bits of
// the distance.
func javaShl(v uint64, n int) uint64 { return v << (uint(n) & 63) }

// ubrexCharSetMask is upstream's CharSet: case-insensitive ASCII
// characters as two bit masks.
type ubrexCharSetMask struct{ mask1, mask2 uint64 }

func (c *ubrexCharSetMask) addChar(ch uint16) {
	ch = ubrexEnsureLowercase(ch)
	offset := int(ch) - 32
	if ch < 32 || ch > 128 {
		throw(illegalArgument, "Bad char : "+string(rune(ch)))
	}
	if offset < 64 {
		c.mask1 |= javaShl(1, offset)
	} else {
		c.mask2 |= javaShl(1, offset-64)
	}
}

func (c *ubrexCharSetMask) addRange(start, end uint16) {
	if start > end {
		throw(illegalArgument, "Invalid range: '"+string(rune(start))+"' is greater than '"+string(rune(end))+"'.")
	}
	if start < 32 || end > 128 {
		throw(illegalArgument, "Characters must be in the range 32 to 128.")
	}
	start = ubrexEnsureLowercase(start)
	end = ubrexEnsureLowercase(end)
	startOffset := int(start) - 32
	endOffset := int(end) - 32
	all := ^uint64(0)
	switch {
	case endOffset < 64:
		// Entire range is in mask1 (offsets 0 to 63)
		length := endOffset - startOffset + 1
		rangeMask := javaShl(1, length) - 1
		if length == 64 {
			rangeMask = all
		}
		c.mask1 |= javaShl(rangeMask, startOffset)
	case startOffset >= 64:
		// Entire range is in mask2 (offsets 64 and above)
		adjustedStart := startOffset - 64
		length := endOffset - 64 - adjustedStart + 1
		c.mask2 |= javaShl(javaShl(1, length)-1, adjustedStart)
	default:
		// Range spans both mask1 and mask2
		length1 := 64 - startOffset
		maskPart1 := javaShl(1, length1) - 1
		if length1 == 64 {
			maskPart1 = all
		}
		c.mask1 |= javaShl(maskPart1, startOffset)
		c.mask2 |= javaShl(1, endOffset-64+1) - 1
	}
}

func (c *ubrexCharSetMask) contains(ch uint16) bool {
	ch = ubrexEnsureLowercase(ch)
	offset := int(ch) - 32
	if offset > 128 {
		return false
	}
	if offset < 64 {
		return c.mask1&javaShl(1, offset) != 0
	}
	return c.mask2&javaShl(1, offset-64) != 0
}

// ubrexCharSet is upstream's ChallengeCharSet: one character of a set
// ("「...」"), possibly negated with "〤".
type ubrexCharSet struct {
	classes  []ubrexCharClassRaw
	set      ubrexCharSetMask
	reversed bool
}

func ubrexCharSetBuild(pattern *ubrexText) *ubrexCharSet {
	if pattern.length() == 0 {
		throw(illegalArgument, "Empty!")
	}
	if pattern.charAt(pattern.length()-1) == '〜' {
		throw(illegalArgument, "Range operator '〜' must be followed by a character.")
	}
	result := &ubrexCharSet{}
	nav := newUbrexTextRange(pattern.content, pattern.p1, pattern.p2, pattern.reversed)
	for ; nav.length() > 0; nav.jump(1) {
		ch := nav.charAt(0)
		switch {
		case ch == '〤':
			result.reversed = true
		case ch == ' ':
		case ch == '〴':
			nav.jump(1)
			charClass := ubrexCharClassRawFromDefinition(nav)
			if !slices.Contains(result.classes, charClass) {
				result.classes = append(result.classes, charClass)
			}
			nav.jump(charClass.definitionLength() - 1)
		case nav.length() > 2 && nav.charAt(1) == '〜':
			nav.jump(2)
			result.set.addRange(ch, nav.charAt(0))
		case ch == '〃':
			result.set.addChar('"')
		case ch == '∙':
			result.set.addChar(' ')
		default:
			result.set.addChar(ch)
		}
	}
	return result
}

func (c *ubrexCharSet) run(s *ubrexText, position int) ubrexResult {
	if s.length() == position {
		return ubrexResultNoMatch
	}
	ch := s.charAt(position)
	matches := c.set.contains(ch)
	for _, class := range c.classes {
		if class.internalMatches(ch) {
			matches = true
		}
	}
	if matches != c.reversed {
		return ubrexResultOne
	}
	return ubrexResultNoMatch
}
