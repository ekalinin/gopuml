// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 com/plantuml/ubrex/ChallengeEndOfText.java
// Provenance: plantuml-mit 1.2026.8 com/plantuml/ubrex/ChallengeLazzyOneOrMore.java
// Provenance: plantuml-mit 1.2026.8 com/plantuml/ubrex/LookAround.java
// Provenance: plantuml-mit 1.2026.8 com/plantuml/ubrex/ChallengeLookAhead.java
// Provenance: plantuml-mit 1.2026.8 com/plantuml/ubrex/ChallengeLookBehind.java
// Provenance: plantuml-mit 1.2026.8 com/plantuml/ubrex/ChallengeOneOrMore.java
// Provenance: plantuml-mit 1.2026.8 com/plantuml/ubrex/ChallengeOneOrMoreUpToOldVersion.java
// Provenance: plantuml-mit 1.2026.8 com/plantuml/ubrex/ChallengeOptional.java
// Provenance: plantuml-mit 1.2026.8 com/plantuml/ubrex/Repetition.java
// Provenance: plantuml-mit 1.2026.8 com/plantuml/ubrex/ChallengeRepetition.java
// Provenance: plantuml-mit 1.2026.8 com/plantuml/ubrex/ChallengeSingleChar.java
// Provenance: plantuml-mit 1.2026.8 com/plantuml/ubrex/ChallengeUpTo.java
// Provenance: plantuml-mit 1.2026.8 com/plantuml/ubrex/ChallengeZeroOrMore.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"
	"slices"
	"strings"
	"unicode/utf16"
)

// ubrexEndOfText is upstream's ChallengeEndOfText.
type ubrexEndOfText struct{}

func (ubrexEndOfText) run(s *ubrexText, position int) ubrexResult {
	if position >= s.length() {
		return ubrexResultZero
	}
	return ubrexResultNoMatch
}

// ubrexLazyOneOrMore is upstream's ChallengeLazzyOneOrMore: it repeats
// origin until stop matches, without consuming stop.
type ubrexLazyOneOrMore struct {
	origin, stop ubrexChallenge
}

func (c *ubrexLazyOneOrMore) run(s *ubrexText, position int) ubrexResult {
	var capture ubrexCapture
	currentPos := position
	for {
		match1 := c.origin.run(s, currentPos)
		if match1.length < 0 {
			return ubrexResultNoMatch
		}
		if match1.length == 0 {
			throw(illegalState, "infinite loop")
		}
		capture = capture.merge(match1.capture)
		currentPos += match1.length
		if c.stop.run(s, currentPos).length >= 0 {
			return ubrexResult{currentPos - position, capture}
		}
	}
}

// ubrexLookAround is upstream's LookAround.
type ubrexLookAround int

const (
	ubrexLookAheadPositive ubrexLookAround = iota
	ubrexLookAheadNegative
	ubrexLookBehindPositive
	ubrexLookBehindNegative
	ubrexEndOfTextLook
)

func ubrexLookAroundFrom(input *ubrexText) (ubrexLookAround, bool) {
	switch input.charAt(0) {
	case '$':
		return ubrexEndOfTextLook, true
	case '=':
		return ubrexLookAheadPositive, true
	case '!':
		return ubrexLookAheadNegative, true
	case '<':
		switch input.charAt(1) {
		case '=':
			return ubrexLookBehindPositive, true
		case '!':
			return ubrexLookBehindNegative, true
		}
	}
	return 0, false
}

func (l ubrexLookAround) definitionSize() int {
	if l == ubrexLookBehindPositive || l == ubrexLookBehindNegative {
		return 2
	}
	return 1
}

// ubrexLook is upstream's ChallengeLookAhead and ChallengeLookBehind: a
// zero-length test of origin after (ahead) or before (behind) the position.
type ubrexLook struct {
	origin ubrexChallenge
	look   ubrexLookAround
}

func (c *ubrexLook) run(s *ubrexText, position int) ubrexResult {
	var tmp int
	if c.look == ubrexLookBehindPositive || c.look == ubrexLookBehindNegative {
		tmp = c.origin.run(s.reverse(position), 0).length
	} else {
		tmp = c.origin.run(s, position).length
	}
	if (tmp >= 0) == (c.look == ubrexLookAheadPositive || c.look == ubrexLookBehindPositive) {
		return ubrexResultZero
	}
	return ubrexResultNoMatch
}

// ubrexOneOrMore is upstream's ChallengeOneOrMore (greedy).
type ubrexOneOrMore struct{ origin ubrexChallenge }

func (c *ubrexOneOrMore) run(s *ubrexText, position int) ubrexResult {
	var capture ubrexCapture
	currentPos := position
	for {
		r := c.origin.run(s, currentPos)
		if r.length < 0 {
			if currentPos > position {
				return ubrexResult{currentPos - position, capture}
			}
			return ubrexResultNoMatch
		}
		if r.length == 0 {
			throw(illegalState, "infinite loop")
		}
		capture = capture.merge(r.capture)
		currentPos += r.length
	}
}

// ubrexOneOrMoreUpTo is upstream's ChallengeOneOrMoreUpToOldVersion: it
// repeats origin until end matches, without consuming end.
type ubrexOneOrMoreUpTo struct{ origin, end ubrexChallenge }

func (c *ubrexOneOrMoreUpTo) run(s *ubrexText, position int) ubrexResult {
	var capture ubrexCapture
	currentPos := position
	for {
		r := c.origin.run(s, currentPos)
		if r.length < 0 {
			return ubrexResultNoMatch
		}
		if r.length == 0 {
			throw(illegalState, "infinite loop")
		}
		capture = capture.merge(r.capture)
		currentPos += r.length
		if c.end.run(s, currentPos).length >= 0 {
			return ubrexResult{currentPos - position, capture}
		}
	}
}

// ubrexOptional is upstream's ChallengeOptional.
type ubrexOptional struct{ origin ubrexChallenge }

func (c *ubrexOptional) run(s *ubrexText, position int) ubrexResult {
	r := c.origin.run(s, position)
	if r.length < 0 {
		return ubrexResultZero
	}
	return r
}

// ubrexRepetition is upstream's Repetition: the allowed counts.
type ubrexRepetition struct {
	values       []int
	minInclusive int
}

func ubrexRepetitionParse(input *ubrexText) *ubrexRepetition {
	result := &ubrexRepetition{minInclusive: math.MaxInt32}
	var token []uint16
	for {
		ch := input.charAt(0)
		switch ch {
		case '}':
			result.addToken(string(utf16.Decode(token)))
			input.jump(1)
			return result
		case ';':
			result.addToken(string(utf16.Decode(token)))
			token = token[:0]
		default:
			token = append(token, ch)
		}
		input.jump(1)
	}
}

func (r *ubrexRepetition) addToken(token string) {
	switch {
	case token == "":
		throw(unsupportedOperation, "empty repetition token")
	case strings.HasSuffix(token, "+"):
		r.minInclusive = javaParseInt(token[:len(token)-1])
	case strings.Contains(token, "-"):
		dash := strings.IndexByte(token, '-')
		lo, hi := javaParseInt(token[:dash]), javaParseInt(token[dash+1:])
		for i := lo; i <= hi; i++ {
			r.values = append(r.values, i)
		}
	default:
		r.values = append(r.values, javaParseInt(token))
	}
}

func (r *ubrexRepetition) match(nb int) bool {
	return nb >= r.minInclusive || slices.Contains(r.values, nb)
}

// ubrexRepeat is upstream's ChallengeRepetition: greedy repetition whose
// count must be allowed.
type ubrexRepeat struct {
	origin     ubrexChallenge
	repetition *ubrexRepetition
}

func (c *ubrexRepeat) run(s *ubrexText, position int) ubrexResult {
	var capture ubrexCapture
	currentPos := position
	count := 0
	for {
		r := c.origin.run(s, currentPos)
		if r.length < 0 {
			if currentPos > position && c.repetition.match(count) {
				return ubrexResult{currentPos - position, capture}
			}
			return ubrexResultNoMatch
		}
		if r.length == 0 {
			throw(illegalState, "infinite loop")
		}
		capture = capture.merge(r.capture)
		currentPos += r.length
		count++
	}
}

// ubrexSingleChar is upstream's ChallengeSingleChar: one character, ASCII
// letters compared case-insensitively.
type ubrexSingleChar struct{ ch uint16 }

func newUbrexSingleChar(ch uint16) *ubrexSingleChar {
	switch ch {
	case '┇', '〴', '「', '」', '〤', '〜', '〇', '〄', '〶', '〘', '〙', '【', '】':
		throw(illegalArgument, "")
	}
	return &ubrexSingleChar{ch: ubrexEnsureLowercase(ch)}
}

func (c *ubrexSingleChar) run(s *ubrexText, position int) ubrexResult {
	if s.length() == position {
		return ubrexResultNoMatch
	}
	if ubrexEnsureLowercase(s.charAt(position)) == c.ch {
		return ubrexResultOne
	}
	return ubrexResultNoMatch
}

// ubrexUpTo is upstream's ChallengeUpTo: the text up to the first position
// where origin matches, possibly empty.
type ubrexUpTo struct{ origin ubrexChallenge }

func (c *ubrexUpTo) run(s *ubrexText, position int) ubrexResult {
	currentPos := position
	for {
		if c.origin.run(s, currentPos).length >= 0 {
			return ubrexResult{length: currentPos - position}
		}
		currentPos++
		if currentPos > s.length() {
			return ubrexResultNoMatch
		}
	}
}

// ubrexZeroOrMore is upstream's ChallengeZeroOrMore (greedy).
type ubrexZeroOrMore struct{ origin ubrexChallenge }

func (c *ubrexZeroOrMore) run(s *ubrexText, position int) ubrexResult {
	var capture ubrexCapture
	currentPos := position
	for {
		r := c.origin.run(s, currentPos)
		if r.length < 0 {
			return ubrexResult{currentPos - position, capture}
		}
		if r.length == 0 {
			throw(illegalState, "infinite loop")
		}
		capture = capture.merge(r.capture)
		currentPos += r.length
	}
}
