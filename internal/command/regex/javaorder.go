// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/regex/RegexComposed.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/regex/RegexOr.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/regex/RegexRepeatedOneOrMore.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/regex/RegexResult.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package regex

import (
	"math"
	"unicode/utf16"
)

// Upstream's RegexResult keeps the groups in a HashMap that every composed
// part builds with putAll of the maps of its sub-parts (a RegexOr or a
// named RegexRepeatedOneOrMore then puts its own name), and getLazzy
// returns the first matching name in the iteration order of that map. The
// order only depends on the structure of the pattern, so it is computed
// once per pattern by replaying java.util.HashMap (Java 19 and later) on
// the names.

// javaNames returns the names of a part in the iteration order of the
// HashMap that upstream's createPartialMatch returns for it.
func javaNames(p Part) []string {
	switch p := p.(type) {
	case *Leaf:
		if p.name == "" {
			return nil
		}
		return []string{p.name}
	case *named:
		inner := &javaHashMap{}
		for _, child := range p.parts {
			inner.putAll(javaNames(child))
		}
		m := &javaHashMap{}
		m.putAll(inner.keys())
		if p.name != "" {
			m.put(p.name)
		}
		return m.keys()
	case *optional:
		return javaComposedNames(p.parts)
	case *zeroOrMore:
		return javaComposedNames(p.parts)
	case *Concat:
		return javaComposedNames(p.parts)
	}
	return nil
}

func javaComposedNames(parts []Part) []string {
	m := &javaHashMap{}
	for _, child := range parts {
		m.putAll(javaNames(child))
	}
	return m.keys()
}

// javaHashMap replays the bucket layout of a java.util.HashMap with the
// default load factor: keys are only added.
type javaHashMap struct {
	table     [][]javaKey
	size      int
	threshold int
}

type javaKey struct {
	key  string
	hash int32
}

const (
	javaLoadFactor        = 0.75
	javaTreeifyThreshold  = 8
	javaMinTreeifyCapcity = 64
)

// javaStringHash is String.hashCode spread as HashMap.hash does.
func javaStringHash(s string) int32 {
	var h int32
	for _, u := range utf16.Encode([]rune(s)) {
		h = 31*h + int32(u)
	}
	return h ^ int32(uint32(h)>>16)
}

// javaTableSizeFor is HashMap.tableSizeFor: the power of two at least c.
func javaTableSizeFor(c int) int {
	n := 1
	for n < c {
		n <<= 1
	}
	return n
}

// putAll is HashMap.putMapEntries: presize an empty table, grow a table
// that cannot hold the new entries, then put each key.
func (m *javaHashMap) putAll(keys []string) {
	s := len(keys)
	if s == 0 {
		return
	}
	if m.table == nil {
		t := int(math.Ceil(float64(s) / javaLoadFactor))
		if t > m.threshold {
			m.threshold = javaTableSizeFor(t)
		}
	} else {
		for s > m.threshold {
			m.resize()
		}
	}
	for _, k := range keys {
		m.put(k)
	}
}

// put is HashMap.putVal for a new key: at the end of its bucket, with the
// resize of a crowded bucket in a small table (treeifyBin) and the resize
// above the threshold.
func (m *javaHashMap) put(key string) {
	if m.table == nil {
		m.resize()
	}
	h := javaStringHash(key)
	i := int(h) & (len(m.table) - 1)
	for _, e := range m.table[i] {
		if e.key == key {
			return
		}
	}
	binCount := len(m.table[i]) - 1
	m.table[i] = append(m.table[i], javaKey{key: key, hash: h})
	if binCount >= javaTreeifyThreshold-1 && len(m.table) < javaMinTreeifyCapcity {
		m.resize()
	}
	m.size++
	if m.size > m.threshold {
		m.resize()
	}
}

// resize is HashMap.resize: the table doubles (or gets its first size) and
// the keys keep their relative order in each bucket.
func (m *javaHashMap) resize() {
	oldCap := len(m.table)
	oldThr := m.threshold
	newCap, newThr := 0, 0
	switch {
	case oldCap > 0:
		newCap = oldCap << 1
		if oldCap >= 16 {
			newThr = oldThr << 1
		}
	case oldThr > 0:
		newCap = oldThr
	default:
		newCap, newThr = 16, 12
	}
	if newThr == 0 {
		newThr = int(float32(newCap) * javaLoadFactor)
	}
	table := make([][]javaKey, newCap)
	for _, bucket := range m.table {
		for _, e := range bucket {
			j := int(e.hash) & (newCap - 1)
			table[j] = append(table[j], e)
		}
	}
	m.table = table
	m.threshold = newThr
}

// keys returns the keys in iteration order.
func (m *javaHashMap) keys() []string {
	var result []string
	for _, bucket := range m.table {
		for _, e := range bucket {
			result = append(result, e.key)
		}
	}
	return result
}
