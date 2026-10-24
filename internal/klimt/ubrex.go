// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/UnicodeBracketedExpression.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/UMatcher.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/CaptureLookup.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/TextNavigator.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/Challenge.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/ChallengeResult.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/Capture.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/CaptureEntry.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/SafeList.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/CompositeList.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/CompositeNamed.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"
	"slices"
	"strings"
	"unicode/utf16"
)

// The ubrex*.go files are upstream's UBrex engine (com.plantuml.ubrex), the
// "Unicode bracketed expressions" that some creole commands use instead of
// regular expressions. It matches like a PEG: quantifiers are greedy or stop
// at a given end, and nothing backtracks. Texts are UTF-16 code units, as Java's
// CharSequence. The builder package (com.plantuml.ubrex.builder) is not
// ported: no ported code uses it.

// ubrexNoMatch is upstream's Challenge.NO_MATCH.
const ubrexNoMatch = math.MinInt32

// ubrexText is upstream's TextNavigator: a view of a UTF-16 text, possibly
// reversed, whose start moves with jump.
type ubrexText struct {
	content  []uint16
	p1, p2   int
	reversed bool
}

func newUbrexText(content []uint16) *ubrexText {
	return &ubrexText{content: content, p2: len(content)}
}

func newUbrexTextRange(content []uint16, p1, p2 int, reversed bool) *ubrexText {
	if p2 < p1 {
		throw(illegalArgument, "")
	}
	return &ubrexText{content: content, p1: p1, p2: p2, reversed: reversed}
}

func (t *ubrexText) reverse(pos int) *ubrexText {
	if t.reversed {
		throw(illegalState, "")
	}
	return newUbrexTextRange(t.content, 0, t.p1+pos, true)
}

func (t *ubrexText) length() int { return t.p2 - t.p1 }

func (t *ubrexText) charAt(index int) uint16 {
	if index < 0 || index >= t.length() {
		throw(indexOutOfBounds, "Index "+itoa(index)+" out of bounds for length "+itoa(t.length()))
	}
	if t.reversed {
		return t.content[t.p2-index-1]
	}
	return t.content[t.p1+index]
}

func (t *ubrexText) indexOf(ch uint16) int {
	for i := 0; i < t.length(); i++ {
		if t.charAt(i) == ch {
			return i
		}
	}
	return -1
}

func (t *ubrexText) subSequence(beginIndex, endIndex int) *ubrexText {
	if beginIndex < 0 || endIndex > t.length() || beginIndex > endIndex {
		throw(indexOutOfBounds, "subSequence("+itoa(beginIndex)+", "+itoa(endIndex)+") out of bounds for length "+
			itoa(t.length()))
	}
	if t.reversed {
		return newUbrexTextRange(t.content, t.p2-endIndex, t.p2-beginIndex, true)
	}
	return newUbrexTextRange(t.content, t.p1+beginIndex, t.p1+endIndex, false)
}

// String returns the text in reading order.
func (t *ubrexText) String() string {
	u := make([]uint16, t.length())
	for i := range u {
		u[i] = t.charAt(i)
	}
	return string(utf16.Decode(u))
}

func (t *ubrexText) jump(step int) {
	if t.reversed {
		t.p2 -= step
	} else {
		t.p1 += step
	}
}

// ubrexEntry is upstream's CaptureEntry.
type ubrexEntry struct{ key, value string }

// ubrexCapture is upstream's Capture: the named values of a match, in match
// order. It is immutable: the methods return new captures (nil is
// Capture.EMPTY).
type ubrexCapture []ubrexEntry

func (c ubrexCapture) withEntry(key, value string) ubrexCapture {
	return slices.Concat(c, ubrexCapture{{key, value}})
}

func (c ubrexCapture) merge(other ubrexCapture) ubrexCapture {
	if len(other) == 0 {
		return c
	}
	if len(c) == 0 {
		return other
	}
	return slices.Concat(c, other)
}

func (c ubrexCapture) withPrefixedKeys(prefix string) ubrexCapture {
	if len(c) == 0 {
		return c
	}
	result := make(ubrexCapture, len(c))
	for i, e := range c {
		result[i] = ubrexEntry{prefix + "/" + e.key, e.value}
	}
	return result
}

func (c ubrexCapture) findValuesByKey(key string) []string {
	result := []string{}
	for _, e := range c {
		if e.key == key {
			result = append(result, e.value)
		}
	}
	return result
}

// findFirstValueByKey returns the first value of key; ok is false when there
// is none (null upstream).
func (c ubrexCapture) findFirstValueByKey(key string) (string, bool) {
	for _, e := range c {
		if e.key == key {
			return e.value, true
		}
	}
	return "", false
}

// findFirstValuesByKeyPrefix returns the values of the first key starting
// with keyPrefix, nil when there is none (null upstream).
func (c ubrexCapture) findFirstValuesByKeyPrefix(keyPrefix string) []string {
	for _, e := range c {
		if strings.HasPrefix(e.key, keyPrefix) {
			return c.findValuesByKey(e.key)
		}
	}
	return nil
}

func (c ubrexCapture) extractByPrefix(keyPrefix string) ubrexCapture {
	result := ubrexCapture{}
	for _, e := range c {
		if strings.HasPrefix(e.key, keyPrefix+"/") {
			result = append(result, ubrexEntry{e.key[len(keyPrefix)+1:], e.value})
		}
	}
	return result
}

// ubrexResult is upstream's ChallengeResult: the length of the match
// (ubrexNoMatch for none) and its captures.
type ubrexResult struct {
	length  int
	capture ubrexCapture
}

var (
	ubrexResultNoMatch = ubrexResult{length: ubrexNoMatch}
	ubrexResultZero    = ubrexResult{length: 0}
	ubrexResultOne     = ubrexResult{length: 1}
)

// ubrexChallenge is upstream's Challenge: something that matches at a
// position of a text.
type ubrexChallenge interface {
	run(s *ubrexText, position int) ubrexResult
}

// ubrexMatcher is upstream's UMatcher: the result of matching a whole
// expression at a position.
type ubrexMatcher struct {
	result   ubrexResult
	accepted string
	position int
	size     int
}

func (m *ubrexMatcher) startMatch() bool { return m.result.length >= 0 }

func (m *ubrexMatcher) exactMatch() bool {
	return m.startMatch() && m.position+m.result.length == m.size
}

// acceptedMatch returns the matched text, "" when there is no match.
func (m *ubrexMatcher) acceptedMatch() string { return m.accepted }

func (m *ubrexMatcher) findValuesByKey(key string) []string {
	return m.result.capture.findValuesByKey(key)
}

func (m *ubrexMatcher) findFirstValueByKey(key string) (string, bool) {
	tmp := m.findValuesByKey(key)
	if len(tmp) == 0 {
		return "", false
	}
	return tmp[0], true
}

// ubrexExpression is upstream's UnicodeBracketedExpression.
type ubrexExpression struct {
	challenge ubrexChallenge
}

// ubrexBuild parses an expression; a malformed expression panics, as
// upstream throws. The expressions are constants of the program.
func ubrexBuild(expression string) *ubrexExpression {
	return &ubrexExpression{challenge: ubrexParseAndBuild(newUbrexText(utf16.Encode([]rune(expression))))}
}

// match matches the expression at position pos (in UTF-16 units) of s.
func (e *ubrexExpression) match(s []uint16, position int) *ubrexMatcher {
	text := newUbrexText(s)
	result := e.challenge.run(text, position)
	accepted := ""
	if result.length >= 0 {
		accepted = text.subSequence(position, position+result.length).String()
	}
	return &ubrexMatcher{result: result, accepted: accepted, position: position, size: text.length()}
}

// matchString matches the expression at the start of s.
func (e *ubrexExpression) matchString(s string) *ubrexMatcher {
	return e.match(utf16.Encode([]rune(s)), 0)
}

// ubrexList is upstream's CompositeList: challenges matched one after the
// other.
type ubrexList struct {
	challenges []ubrexChallenge
}

// ubrexParseAndBuild is upstream's CompositeList.parseAndBuild.
func ubrexParseAndBuild(input *ubrexText) *ubrexList {
	if input.length() == 0 {
		throw(unsupportedOperation, "")
	}
	result := &ubrexList{}
	result.parseAndConsumeNow(newUbrexTextRange(input.content, input.p1, input.p2, input.reversed))
	return result
}

// ubrexParseAndBuildFromText is upstream's
// CompositeList.parseAndBuildFromTextNavigator: it consumes input.
func ubrexParseAndBuildFromText(input *ubrexText) *ubrexList {
	result := &ubrexList{}
	result.parseAndConsumeNow(input)
	return result
}

func (l *ubrexList) parseAndConsumeNow(input *ubrexText) {
	for input.length() > 0 {
		if input.charAt(0) == ' ' {
			input.jump(1)
		} else {
			l.challenges = append(l.challenges, ubrexParse(input)...)
		}
	}
}

func (l *ubrexList) run(s *ubrexText, position int) ubrexResult {
	var capture ubrexCapture
	current := position
	for _, challenge := range l.challenges {
		r := challenge.run(s, current)
		if r.length < 0 {
			return ubrexResultNoMatch
		}
		current += r.length
		capture = capture.merge(r.capture)
	}
	return ubrexResult{current - position, capture}
}

// ubrexNamed is upstream's CompositeNamed: a capture.
type ubrexNamed struct {
	name       string
	challenges []ubrexChallenge
}

func (n *ubrexNamed) run(s *ubrexText, position int) ubrexResult {
	var capture ubrexCapture
	current := position
	for _, challenge := range n.challenges {
		r := challenge.run(s, current)
		if r.length < 0 {
			return ubrexResultNoMatch
		}
		current += r.length
		capture = capture.merge(r.capture)
	}
	value := s.subSequence(position, current).String()
	capture = capture.withPrefixedKeys(n.name)
	capture = capture.withEntry(n.name, value)
	return ubrexResult{current - position, capture}
}
