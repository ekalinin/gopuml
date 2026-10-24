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

// styleTrieNode is a node of StyleAtomTrie: the children by atom and the
// declarations whose atoms end here, in insertion order.
type styleTrieNode struct {
	children  map[styleAtom]*styleTrieNode
	rulesHere []*Style
}

// styleAtomTrie is upstream's StyleAtomTrie over styles: each declaration is
// stored under the path of its sorted atoms, so that a query only visits
// the declarations whose atoms are all among its own.
type styleAtomTrie struct {
	root styleTrieNode
}

func (t *styleAtomTrie) insert(style *Style) {
	current := &t.root
	for _, atom := range style.Signature().atoms() {
		child := current.children[atom]
		if child == nil {
			if current.children == nil {
				current.children = map[styleAtom]*styleTrieNode{}
			}
			child = &styleTrieNode{}
			current.children[atom] = child
		}
		current = child
	}
	current.rulesHere = append(current.rulesHere, style)
}

// findMatching returns the stored styles whose atoms are all in the query's
// and whose level constraint accepts the query's, in the order of the walk:
// the declarations of a node, then its children in the order of the query's
// atoms.
func (t *styleAtomTrie) findMatching(query StyleSignatureBasic) []*Style {
	var result []*Style
	collectStyles(&t.root, query.atoms(), 0, query, &result)
	return result
}

func collectStyles(node *styleTrieNode, queryAtoms []styleAtom, fromIndex int, query StyleSignatureBasic,
	result *[]*Style) {
	for _, stored := range node.rulesHere {
		decl := stored.Signature()
		if levelConstraintMatches(decl.level, decl.star, query.level, query.star) {
			*result = append(*result, stored)
		}
	}
	for i := fromIndex; i < len(queryAtoms); i++ {
		if child := node.children[queryAtoms[i]]; child != nil {
			collectStyles(child, queryAtoms, i+1, query, result)
		}
	}
}

// StyleIndex is upstream's StyleIndex: the declared styles, in load order,
// with a trie built on the first query, and the merged styles of the
// queries without stereotype cached for the life of the index. It is
// immutable once built, and safe for concurrent queries.
type StyleIndex struct {
	allStyles []*Style

	trieOnce sync.Once
	trie     *styleAtomTrie

	mu          sync.Mutex
	sharedCache map[StyleSignatureBasic]*Style
}

func newStyleIndex(allStyles []*Style) *StyleIndex {
	return &StyleIndex{allStyles: allStyles, sharedCache: map[StyleSignatureBasic]*Style{}}
}

var emptyStyleIndex = newStyleIndex(nil)

// withLoaded returns a new index with newStyle merged into the style of the
// same query, or appended.
func (idx *StyleIndex) withLoaded(newStyle *Style) *StyleIndex {
	updated := append(make([]*Style, 0, len(idx.allStyles)+1), idx.allStyles...)
	return newStyleIndex(mergeOrAppendStyle(updated, newStyle))
}

// withMuted returns a new index with each of newStyles merged in turn, as
// withLoaded does.
func (idx *StyleIndex) withMuted(newStyles []*Style) *StyleIndex {
	if len(newStyles) == 0 {
		return idx
	}
	updated := append(make([]*Style, 0, len(idx.allStyles)+len(newStyles)), idx.allStyles...)
	for _, modifiedStyle := range newStyles {
		updated = mergeOrAppendStyle(updated, modifiedStyle)
	}
	return newStyleIndex(updated)
}

// mergeOrAppendStyle merges newStyle into the style of list with the same
// query, in place, or appends it.
func mergeOrAppendStyle(list []*Style, newStyle *Style) []*Style {
	signature := newStyle.Signature()
	for i, style := range list {
		if style.Signature() == signature {
			list[i] = style.MergeWith(newStyle, MergeOverwriteExistingValue)
			return list
		}
	}
	return append(list, newStyle)
}

func (idx *StyleIndex) trieOf() *styleAtomTrie {
	idx.trieOnce.Do(func() {
		idx.trie = &styleAtomTrie{}
		for _, style := range idx.allStyles {
			idx.trie.insert(style)
		}
	})
	return idx.trie
}

// findMatching returns the styles that apply to query, in the order of the
// trie walk.
func (idx *StyleIndex) findMatching(query StyleSignatureBasic) []*Style {
	return idx.trieOf().findMatching(query)
}

// mergedStyle returns the merged style of query, cached in the index for a
// query without stereotype and in perDiagram (guarded by perDiagramMu)
// otherwise; nil when no style applies, which is not cached.
func (idx *StyleIndex) mergedStyle(query StyleSignatureBasic, perDiagram map[StyleSignatureBasic]*Style,
	perDiagramMu *sync.Mutex) *Style {
	cache, mu := idx.sharedCache, &idx.mu
	if query.HasStereotype() {
		cache, mu = perDiagram, perDiagramMu
	}
	mu.Lock()
	cached, ok := cache[query]
	mu.Unlock()
	if ok {
		return cached
	}
	computed := idx.computeMergedStyle(query)
	if computed != nil {
		mu.Lock()
		cache[query] = computed
		mu.Unlock()
	}
	return computed
}

func (idx *StyleIndex) computeMergedStyle(query StyleSignatureBasic) *Style {
	var mergedStyle *Style
	for _, style := range idx.findMatching(query) {
		if mergedStyle == nil {
			mergedStyle = style
		} else {
			mergedStyle = mergedStyle.MergeWith(style, MergeOverwriteExistingValue)
		}
	}
	return mergedStyle
}

// StyleBuilder holds the styles of a diagram and resolves the style of a
// query. It also numbers the declared values (AutomaticCounter), so that
// later declarations win. The merged styles of the queries with a
// stereotype are cached in the builder, the others in its index.
type StyleBuilder struct {
	index   *StyleIndex
	counter int

	mu    sync.Mutex
	cache map[StyleSignatureBasic]*Style
}

var _ AutomaticCounter = (*StyleBuilder)(nil)

// NewStyleBuilder returns an empty builder.
func NewStyleBuilder() *StyleBuilder {
	return &StyleBuilder{index: emptyStyleIndex, cache: map[StyleSignatureBasic]*Style{}}
}

// PrintMe writes the styles to the standard error.
func (b *StyleBuilder) PrintMe() {
	for _, style := range b.index.allStyles {
		style.PrintMe()
	}
}

// CloneMe returns a builder with the same index and counter, and an empty
// cache.
func (b *StyleBuilder) CloneMe() *StyleBuilder {
	result := NewStyleBuilder()
	result.index = b.index
	result.counter = b.counter
	return result
}

// CreateStyleStereotype returns the style declared for a stereotype name
// (lower-cased) with no style name, or an empty style. A name with "*"
// panics, as upstream throws IllegalArgumentException.
func (b *StyleBuilder) CreateStyleStereotype(name string) *Style {
	if strings.Contains(name, StyleSignatureStar) {
		throw(illegalArgument, "")
	}
	name = javaLower(name)
	signature := StyleSignatureBasicEmpty().AddStereotype(name)
	var result *Style
	for _, style := range b.index.allStyles {
		if style.Signature() == signature {
			if result == nil {
				result = style
			} else {
				result = result.MergeWith(style, MergeOverwriteExistingValue)
			}
		}
	}
	if result == nil {
		return NewStyle(signature, nil)
	}
	return result
}

// MuteStyle returns a new builder with the modified styles merged over the
// declared ones (later values win).
func (b *StyleBuilder) MuteStyle(modifiedStyles []*Style) *StyleBuilder {
	result := NewStyleBuilder()
	result.counter = b.counter
	result.index = b.index.withMuted(modifiedStyles)
	return result
}

// LoadInternal declares a style, merging it over a style of the same
// query. A starred query panics, as upstream throws
// IllegalArgumentException.
func (b *StyleBuilder) LoadInternal(signature StyleSignatureBasic, newStyle *Style) {
	if signature.IsStarred() {
		throw(illegalArgument, "")
	}
	b.index = b.index.withLoaded(newStyle)
	b.mu.Lock()
	clear(b.cache)
	b.mu.Unlock()
}

// NextInt returns the next value order.
func (b *StyleBuilder) NextInt() int {
	b.counter++
	return b.counter
}

// MergedStyle returns the style of a query, merging the declared styles
// that apply to it (cached); nil when none does.
func (b *StyleBuilder) MergedStyle(signature StyleSignatureBasic) *Style {
	return b.index.mergedStyle(signature, b.cache, &b.mu)
}

// MergedStyleSpecial is MergedStyle without cache, with the starred styles
// applied through the ancestor cascade at ancestorRank.
func (b *StyleBuilder) MergedStyleSpecial(signature StyleSignatureBasic, ancestorRank int) *Style {
	var mergedStyle *Style
	for _, style := range b.index.findMatching(signature) {
		tmp := style
		if style.Signature().IsStarred() {
			tmp = tmp.WithAncestorRank(ancestorRank)
		}
		if mergedStyle == nil {
			mergedStyle = tmp
		} else {
			mergedStyle = mergedStyle.MergeWith(tmp, MergeOverwriteExistingValue)
		}
	}
	return mergedStyle
}

// Styles returns the declared styles, in load order.
func (b *StyleBuilder) Styles() []*Style { return b.index.allStyles }

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
