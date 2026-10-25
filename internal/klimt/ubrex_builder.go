// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/builder/UBrexPart.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/builder/UBrexConcat.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/builder/UBrexLeaf.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/builder/UBrexNamed.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/builder/UBrexOneOrMore.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/builder/UBrexOptional.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/builder/UBrexOr.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/builder/UBrexUpto.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/builder/UBrexZeroOrMore.java
// Provenance: plantuml-mit 1.2026.9 com/plantuml/ubrex/UMatcher.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import "unicode/utf16"

// UBrexPart is upstream's UBrexPart (com.plantuml.ubrex.builder): a part of
// a UBrex expression built in code, as the commands written with UBrex
// build their patterns. Only the parts these commands use are ported.
type UBrexPart struct {
	challenge ubrexChallenge
}

// UBrexLeaf is upstream's UBrexLeaf: a part written in UBrex syntax.
func UBrexLeaf(definition string) UBrexPart {
	return UBrexPart{ubrexParseAndBuild(newUbrexText(utf16.Encode([]rune(definition))))}
}

// UBrexEnd is UBrexLeaf.end(): the end of the text.
func UBrexEnd() UBrexPart { return UBrexLeaf("〒$") }

// UBrexSpaceOneOrMore is UBrexLeaf.spaceOneOrMore().
func UBrexSpaceOneOrMore() UBrexPart { return UBrexLeaf("〇+〴s") }

// UBrexSpaceZeroOrMore is UBrexLeaf.spaceZeroOrMore().
func UBrexSpaceZeroOrMore() UBrexPart { return UBrexLeaf("〇*〴s") }

// UBrexConcat is upstream's UBrexConcat.build: the parts one after the
// other.
func UBrexConcat(parts ...UBrexPart) UBrexPart {
	result := &ubrexList{}
	for _, p := range parts {
		result.challenges = append(result.challenges, p.challenge)
	}
	return UBrexPart{result}
}

// UBrexNamed is upstream's UBrexNamed: a part captured under name.
func UBrexNamed(name string, part UBrexPart) UBrexPart {
	return UBrexPart{&ubrexNamed{name: name, challenges: []ubrexChallenge{part.challenge}}}
}

// UBrexOr is upstream's UBrexOr: the first of the parts that matches.
func UBrexOr(parts ...UBrexPart) UBrexPart {
	result := &ubrexAlternative{}
	for _, p := range parts {
		result.alternatives = append(result.alternatives, p.challenge)
	}
	return UBrexPart{result}
}

// UBrexOptional is upstream's UBrexOptional: the part, or nothing.
func UBrexOptional(part UBrexPart) UBrexPart { return UBrexPart{&ubrexOptional{part.challenge}} }

// UBrexOneOrMore is upstream's UBrexOneOrMore: the part once or more,
// without giving back what it matched.
func UBrexOneOrMore(part UBrexPart) UBrexPart { return UBrexPart{&ubrexOneOrMore{part.challenge}} }

// UBrexZeroOrMore is upstream's UBrexZeroOrMore: the part any number of
// times, without giving back what it matched.
func UBrexZeroOrMore(part UBrexPart) UBrexPart { return UBrexPart{&ubrexZeroOrMore{part.challenge}} }

// UBrexUpto is upstream's UBrexUpto: what, once or more, up to the first
// position where stop matches, then stop.
func UBrexUpto(what, stop UBrexPart) UBrexPart {
	return UBrexPart{&ubrexList{challenges: []ubrexChallenge{
		&ubrexOneOrMoreUpTo{origin: what.challenge, end: stop.challenge}, stop.challenge}}}
}

// Match matches the part at the start of s.
func (p UBrexPart) Match(s string) *UBrexMatcher {
	return &UBrexMatcher{(&ubrexExpression{challenge: p.challenge}).matchString(s)}
}

// UBrexMatcher is upstream's UMatcher: the result of matching a part, with
// its captures.
type UBrexMatcher struct {
	m *ubrexMatcher
}

// ExactMatch reports whether the part matches the whole text.
func (m *UBrexMatcher) ExactMatch() bool { return m.m.exactMatch() }

// FindValuesByKey returns the values captured under key.
func (m *UBrexMatcher) FindValuesByKey(key string) []string { return m.m.findValuesByKey(key) }

// FindFirstValuesByKeyPrefix returns the values of the first captured key
// that starts with keyPrefix.
func (m *UBrexMatcher) FindFirstValuesByKeyPrefix(keyPrefix string) []string {
	return m.m.result.capture.findFirstValuesByKeyPrefix(keyPrefix)
}

// UBrexNavigator is upstream's TextNavigator as the gantt sentences use it:
// a text whose start moves forward as the words are read.
type UBrexNavigator struct {
	text *ubrexText
}

// NewUBrexNavigator is TextNavigator.build.
func NewUBrexNavigator(s string) *UBrexNavigator {
	return &UBrexNavigator{newUbrexText(utf16.Encode([]rune(s)))}
}

// Copy is copy.
func (n *UBrexNavigator) Copy() *UBrexNavigator {
	t := *n.text
	return &UBrexNavigator{&t}
}

// Jump is jump: the start moves by step UTF-16 units.
func (n *UBrexNavigator) Jump(step int) { n.text.jump(step) }

// MatchNavigator matches the part at the start of the navigator, as
// UBrexPart.match(TextNavigator).
func (p UBrexPart) MatchNavigator(n *UBrexNavigator) *UBrexMatcher {
	result := p.challenge.run(n.text, 0)
	accepted := ""
	if result.length >= 0 {
		accepted = n.text.subSequence(0, result.length).String()
	}
	return &UBrexMatcher{&ubrexMatcher{result: result, accepted: accepted, size: n.text.length()}}
}

// StartMatch reports whether the part matches the start of the text.
func (m *UBrexMatcher) StartMatch() bool { return m.m.startMatch() }

// AcceptedMatch is getAcceptedMatch: the text matched, "" for none.
func (m *UBrexMatcher) AcceptedMatch() string { return m.m.acceptedMatch() }

// AcceptedLength is the length of AcceptedMatch in UTF-16 units, as Java's
// String.length.
func (m *UBrexMatcher) AcceptedLength() int {
	if m.m.result.length < 0 {
		return 0
	}
	return m.m.result.length
}

// FindFirstValueByKey returns the first value captured under key; ok is
// false when there is none (null upstream).
func (m *UBrexMatcher) FindFirstValueByKey(key string) (string, bool) {
	return m.m.findFirstValueByKey(key)
}

// ExtractByPrefix is extractByPrefix: the values captured under
// keyPrefix/..., under the rest of their keys.
func (m *UBrexMatcher) ExtractByPrefix(keyPrefix string) *UBrexCapture {
	return &UBrexCapture{m.m.result.capture.extractByPrefix(keyPrefix)}
}

// UBrexCapture is upstream's Capture: named values.
type UBrexCapture struct {
	c ubrexCapture
}

// FindValuesByKey returns the values captured under key.
func (c *UBrexCapture) FindValuesByKey(key string) []string { return c.c.findValuesByKey(key) }

// FindFirstValueByKey returns the first value captured under key; ok is
// false when there is none (null upstream).
func (c *UBrexCapture) FindFirstValueByKey(key string) (string, bool) {
	return c.c.findFirstValueByKey(key)
}
