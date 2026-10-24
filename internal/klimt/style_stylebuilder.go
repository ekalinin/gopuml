// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/StyleBuilder.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/StyleIndex.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/StyleAtomTrie.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/utils/StringTrie.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"strings"
	"sync"
)

// orderedMap is an insertion-ordered map, upstream's LinkedHashMap: a put
// on an existing key keeps its position.
type orderedMap[K comparable, V any] struct {
	keys   []K
	values map[K]V
}

func newOrderedMap[K comparable, V any]() *orderedMap[K, V] {
	return &orderedMap[K, V]{values: map[K]V{}}
}

func (m *orderedMap[K, V]) get(k K) (V, bool) {
	v, ok := m.values[k]
	return v, ok
}

func (m *orderedMap[K, V]) put(k K, v V) {
	if _, ok := m.values[k]; !ok {
		m.keys = append(m.keys, k)
	}
	m.values[k] = v
}

func (m *orderedMap[K, V]) putAll(other *orderedMap[K, V]) {
	for _, k := range other.keys {
		m.put(k, other.values[k])
	}
}

func (m *orderedMap[K, V]) size() int { return len(m.keys) }

// StyleStorage holds the declared styles: those with stereotypes keyed by
// signature, the others by key, each in declaration order.
type StyleStorage struct {
	legacy *orderedMap[StyleSignatureBasic, *Style]
	plain  *orderedMap[StyleKey, *Style]
}

// NewStyleStorage returns an empty storage.
func NewStyleStorage() *StyleStorage {
	return &StyleStorage{legacy: newOrderedMap[StyleSignatureBasic, *Style](),
		plain: newOrderedMap[StyleKey, *Style]()}
}

// PrintMe writes the stereotype styles to the standard error.
func (s *StyleStorage) PrintMe() {
	for _, k := range s.legacy.keys {
		s.legacy.values[k].PrintMe()
	}
}

// PutAll copies the styles of other.
func (s *StyleStorage) PutAll(other *StyleStorage) {
	s.legacy.putAll(other.legacy)
	s.plain.putAll(other.plain)
}

// Get returns the style declared for a signature, nil when there is none.
func (s *StyleStorage) Get(signature StyleSignatureBasic) *Style {
	if signature.stereotypes == "" {
		v, _ := s.plain.get(signature.Key())
		return v
	}
	v, _ := s.legacy.get(signature)
	return v
}

// Put declares a style for its signature.
func (s *StyleStorage) Put(modifiedStyle *Style) {
	signature := modifiedStyle.Signature()
	if signature.stereotypes == "" {
		s.plain.put(signature.Key(), modifiedStyle)
	} else {
		s.legacy.put(signature, modifiedStyle)
	}
}

// Styles returns the stereotype styles then the others, each in declaration
// order.
func (s *StyleStorage) Styles() []*Style {
	result := make([]*Style, 0, s.legacy.size()+s.plain.size())
	for _, k := range s.legacy.keys {
		result = append(result, s.legacy.values[k])
	}
	for _, k := range s.plain.keys {
		result = append(result, s.plain.values[k])
	}
	return result
}

// ComputeMergedStyle merges, in storage order, the styles whose signature
// applies to signature; nil when none does.
func (s *StyleStorage) ComputeMergedStyle(signature StyleSignatureBasic) *Style {
	var mergedStyle *Style
	for _, style := range s.Styles() {
		key := style.Signature()
		if !key.MatchAll(signature) {
			continue
		}
		if mergedStyle == nil {
			mergedStyle = style
		} else {
			mergedStyle = mergedStyle.MergeWith(style, MergeOverwriteExistingValue)
		}
	}
	return mergedStyle
}

// StyleBuilder holds the styles of a diagram and resolves the style of a
// signature. It also numbers the declared values (AutomaticCounter), so that
// later declarations win.
//
// The merged styles are cached, as upstream does; the cache is guarded so
// that a builder can be read concurrently. Upstream's log of the used
// signatures is not kept.
type StyleBuilder struct {
	storage *StyleStorage
	counter int

	mu    sync.Mutex
	cache map[StyleSignatureBasic]*Style
}

var _ AutomaticCounter = (*StyleBuilder)(nil)

// NewStyleBuilder returns an empty builder.
func NewStyleBuilder() *StyleBuilder {
	return &StyleBuilder{storage: NewStyleStorage(), cache: map[StyleSignatureBasic]*Style{}}
}

// PrintMe writes the stereotype styles to the standard error.
func (b *StyleBuilder) PrintMe() { b.storage.PrintMe() }

// CloneMe returns a builder with the same styles and counter, and an empty
// cache.
func (b *StyleBuilder) CloneMe() *StyleBuilder {
	result := NewStyleBuilder()
	result.storage.PutAll(b.storage)
	result.counter = b.counter
	return result
}

// CreateStyleStereotype returns the style declared for a stereotype name
// (lower-cased), or an empty style. A name with "*" panics, as upstream
// throws IllegalArgumentException.
func (b *StyleBuilder) CreateStyleStereotype(name string) *Style {
	if strings.Contains(name, StyleSignatureStar) {
		throw(illegalArgument, "")
	}
	name = javaLower(name)
	signature := StyleSignatureBasicCreateStereotype(name)
	if result := b.storage.Get(signature); result != nil {
		return result
	}
	return NewStyle(signature, nil)
}

// MuteStyle returns a new builder with the modified styles merged over the
// declared ones (later values win).
func (b *StyleBuilder) MuteStyle(modifiedStyles []*Style) *StyleBuilder {
	result := NewStyleBuilder()
	result.counter = b.counter
	result.storage.PutAll(b.storage)
	for _, modifiedStyle := range modifiedStyles {
		signature := modifiedStyle.Signature()
		orig := result.storage.Get(signature)
		if orig == nil {
			result.storage.Put(modifiedStyle)
		} else {
			result.storage.Put(orig.MergeWith(modifiedStyle, MergeOverwriteExistingValue))
		}
	}
	return result
}

// LoadInternal declares a style, merging it over a style of the same
// signature. A starred signature panics, as upstream throws
// IllegalArgumentException.
func (b *StyleBuilder) LoadInternal(signature StyleSignatureBasic, newStyle *Style) {
	if signature.IsStarred() {
		throw(illegalArgument, "")
	}
	orig := b.storage.Get(signature)
	if orig == nil {
		b.storage.Put(newStyle)
	} else {
		b.storage.Put(orig.MergeWith(newStyle, MergeOverwriteExistingValue))
	}
}

// NextInt returns the next value priority.
func (b *StyleBuilder) NextInt() int {
	b.counter++
	return b.counter
}

// MergedStyle returns the style of a signature, merging the declared styles
// that apply to it (cached); nil when none does.
func (b *StyleBuilder) MergedStyle(signature StyleSignatureBasic) *Style {
	b.mu.Lock()
	defer b.mu.Unlock()
	if s, ok := b.cache[signature]; ok {
		return s
	}
	s := b.storage.ComputeMergedStyle(signature)
	// Upstream's ConcurrentHashMap.computeIfAbsent does not store null.
	if s != nil {
		b.cache[signature] = s
	}
	return s
}

// MergedStyleSpecial is MergedStyle without cache, with the priorities of
// the starred styles increased by deltaPriority.
func (b *StyleBuilder) MergedStyleSpecial(signature StyleSignatureBasic, deltaPriority int) *Style {
	var mergedStyle *Style
	for _, style := range b.storage.Styles() {
		key := style.Signature()
		if !key.MatchAll(signature) {
			continue
		}
		tmp := style
		if key.IsStarred() {
			tmp = tmp.DeltaPriority(deltaPriority)
		}
		if mergedStyle == nil {
			mergedStyle = tmp
		} else {
			mergedStyle = mergedStyle.MergeWith(tmp, MergeOverwriteExistingValue)
		}
	}
	return mergedStyle
}

// Storage returns the declared styles.
func (b *StyleBuilder) Storage() *StyleStorage { return b.storage }

// StringTrie maps strings to values, ignoring case.
type StringTrie[V any] struct {
	root *trieNode[V]
}

type trieNode[V any] struct {
	value    V
	set      bool
	children map[rune]*trieNode[V]
}

// NewStringTrie returns an empty trie.
func NewStringTrie[V any]() *StringTrie[V] {
	return &StringTrie[V]{root: &trieNode[V]{children: map[rune]*trieNode[V]{}}}
}

// Put maps key to value.
func (t *StringTrie[V]) Put(key string, value V) {
	node := t.root
	for _, c := range key {
		c = javaCharLower(c)
		next := node.children[c]
		if next == nil {
			next = &trieNode[V]{children: map[rune]*trieNode[V]{}}
			node.children[c] = next
		}
		node = next
	}
	node.value = value
	node.set = true
}

// Get returns the value of key; ok is false when there is none.
func (t *StringTrie[V]) Get(key string) (value V, ok bool) {
	node := t.root
	for _, c := range key {
		node = node.children[javaCharLower(c)]
		if node == nil {
			return value, false
		}
	}
	return node.value, node.set
}
