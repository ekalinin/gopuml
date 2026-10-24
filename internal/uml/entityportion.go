// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/abel/EntityPortion.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package uml

// EntityPortion is a part of an element that "hide" and "show" act on.
type EntityPortion int

// Entity portions, in upstream's order.
const (
	PortionField EntityPortion = iota
	PortionMethod
	PortionMember
	PortionCircledCharacter
	PortionStereotype
)

var entityPortionNames = [...]string{"FIELD", "METHOD", "MEMBER", "CIRCLED_CHARACTER", "STEREOTYPE"}

// String returns upstream's name, for example "FIELD".
func (p EntityPortion) String() string { return entityPortionNames[p] }

// EntityPortions is a set of portions, upstream's EnumSet<EntityPortion>.
type EntityPortions uint8

// Contains reports whether p is in the set.
func (s EntityPortions) Contains(p EntityPortion) bool { return s&(1<<p) != 0 }

// With returns the set with the portions of other added.
func (s EntityPortions) With(other EntityPortions) EntityPortions { return s | other }

// Without returns the set without the portions of other.
func (s EntityPortions) Without(other EntityPortions) EntityPortions { return s &^ other }

// List returns the portions of the set in enum order.
func (s EntityPortions) List() []EntityPortion {
	var out []EntityPortion
	for p := PortionField; p <= PortionStereotype; p++ {
		if s.Contains(p) {
			out = append(out, p)
		}
	}
	return out
}

// AsSet returns the portions p stands for: fields and methods for
// PortionMember, p alone otherwise.
func (p EntityPortion) AsSet() EntityPortions {
	if p == PortionMember {
		return 1<<PortionField | 1<<PortionMethod
	}
	return 1 << p
}
