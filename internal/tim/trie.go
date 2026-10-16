// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/Trie.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/TrieImpl.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

// Trie finds the known names (variables, functions) at a position of a line.
type Trie interface {
	// getLonguestMatchStartingIn returns the length, in chars, of the name
	// that starts at s[pos:], or 0. The walk follows s as long as the trie
	// does and succeeds only if it stops at the end of a name.
	getLonguestMatchStartingIn(s []uint16, pos int) int
}

// TrieImpl is a trie of UTF-16 chars; the char 0 marks the end of a name.
type TrieImpl struct {
	brothers map[uint16]*TrieImpl
}

func newTrieImpl() *TrieImpl { return &TrieImpl{brothers: map[uint16]*TrieImpl{}} }

func (t *TrieImpl) add(s string) {
	u := toUTF16(s)
	for _, c := range u {
		if c == 0 {
			throwRuntime("IllegalArgumentException", "")
		}
	}
	current := t
	for _, c := range append(u, 0) {
		current = current.getOrCreate(c)
	}
}

// remove deletes the end marker of s, leaving the path in place.
func (t *TrieImpl) remove(s string) bool {
	u := append(toUTF16(s), 0)
	current := t
	for i, first := range u {
		child := current.brothers[first]
		if child == nil {
			return false
		}
		if i == len(u)-2 {
			_, ok := child.brothers[0]
			delete(child.brothers, 0)
			return ok
		}
		current = child
	}
	throwRuntime("IllegalStateException", "")
	return false
}

func (t *TrieImpl) getOrCreate(added uint16) *TrieImpl {
	child := t.brothers[added]
	if child == nil {
		child = newTrieImpl()
		t.brothers[added] = child
	}
	return child
}

func (t *TrieImpl) getLonguestMatchStartingIn(s []uint16, pos int) int {
	current := t
	n := 0
	for current != nil {
		_, terminal := current.brothers[0]
		if len(s) == pos {
			if terminal {
				return n
			}
			return 0
		}
		child := current.brothers[s[pos]]
		if child == nil || len(child.brothers) == 0 {
			if terminal {
				return n
			}
			return 0
		}
		n++
		current = child
		pos++
	}
	return 0
}
