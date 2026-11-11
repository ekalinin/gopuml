// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/yaml/parser/Monomorph.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/yaml/parser/MonomorphType.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package yamlparser

import "strconv"

// MonomorphType is what a Monomorph has become.
type MonomorphType int

// The kinds of Monomorphs.
const (
	Undeterminate MonomorphType = iota
	Scalar
	List
	Map
)

// Monomorph is a YAML value that the lines fill: undetermined when it is
// created, then a scalar, a list or a map, whose keys keep their order.
type Monomorph struct {
	typ   MonomorphType
	value string
	list  []*Monomorph
	keys  []string
	m     map[string]*Monomorph
}

func (m *Monomorph) setValue(value string) {
	if m.typ == List || m.typ == Map {
		throw(IllegalStateException, "")
	}
	m.value = value
	m.list = nil
	m.keys, m.m = nil, nil
	m.typ = Scalar
}

func (m *Monomorph) addInList(elementToBeAdded *Monomorph) {
	if m.typ == Undeterminate {
		m.list = nil
		m.typ = List
	} else if m.typ != List {
		throw(IllegalStateException, "Monomorph is not of type LIST.")
	}
	m.list = append(m.list, elementToBeAdded)
}

func (m *Monomorph) putInMap(key string, value *Monomorph) {
	if m.typ == Undeterminate {
		m.m = map[string]*Monomorph{}
		m.typ = Map
	} else if m.typ != Map {
		throw(IllegalStateException, "Monomorph is not of type MAP.")
	}
	if _, ok := m.m[key]; !ok {
		m.keys = append(m.keys, key)
	}
	m.m[key] = value
}

// Type returns what the Monomorph has become.
func (m *Monomorph) Type() MonomorphType { return m.typ }

// Value returns the value of a scalar.
func (m *Monomorph) Value() string {
	if m.typ != Scalar {
		throw(IllegalStateException, "Not a scalar value.")
	}
	return m.value
}

// ElementAt returns an element of a list.
func (m *Monomorph) ElementAt(index int) *Monomorph {
	if m.typ != List {
		throw(IllegalStateException, "Not a list.")
	}
	if index < 0 || index >= len(m.list) {
		throw(IndexOutOfBoundsException, "Index "+strconv.Itoa(index)+" out of bounds for length "+
			strconv.Itoa(len(m.list)))
	}
	return m.list[index]
}

// MapValue returns the value of a key of a map.
func (m *Monomorph) MapValue(key string) *Monomorph {
	if m.typ != Map {
		throw(IllegalStateException, "Not a map.")
	}
	return m.m[key]
}

// Keys returns the keys of a map, in their order.
func (m *Monomorph) Keys() []string {
	if m.typ != Map {
		throw(IllegalStateException, "Not a map.")
	}
	return m.keys
}

// Size returns the number of elements of a list or a map.
func (m *Monomorph) Size() int {
	switch m.typ {
	case List:
		return len(m.list)
	case Map:
		return len(m.keys)
	}
	throw(IllegalStateException, "Not a container type.")
	return 0
}

func scalar(value string) *Monomorph {
	result := &Monomorph{}
	result.setValue(value)
	return result
}

func list(input []string) *Monomorph {
	// A list, even when input is empty.
	result := &Monomorph{typ: List}
	for _, s := range input {
		result.addInList(scalar(s))
	}
	return result
}
