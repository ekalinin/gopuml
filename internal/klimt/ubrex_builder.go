// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 com/plantuml/ubrex/builder/UBrexPart.java
// Provenance: plantuml-mit 1.2026.8 com/plantuml/ubrex/builder/UBrexConcat.java
// Provenance: plantuml-mit 1.2026.8 com/plantuml/ubrex/builder/UBrexLeaf.java
// Provenance: plantuml-mit 1.2026.8 com/plantuml/ubrex/builder/UBrexNamed.java
// Provenance: plantuml-mit 1.2026.8 com/plantuml/ubrex/builder/UBrexOptional.java
// Provenance: plantuml-mit 1.2026.8 com/plantuml/ubrex/builder/UBrexOr.java
// Provenance: plantuml-mit 1.2026.8 com/plantuml/ubrex/UMatcher.java
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
