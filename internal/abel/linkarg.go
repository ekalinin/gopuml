// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/abel/LinkArg.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"github.com/ekalinin/gopuml/internal/klimt"
)

// OptString is a Java String that may be null: S is meaningful when OK.
type OptString struct {
	S  string
	OK bool
}

// Some returns the present string s.
func Some(s string) OptString { return OptString{S: s, OK: true} }

// Opt returns s when ok, null otherwise.
func Opt(s string, ok bool) OptString {
	if !ok {
		return OptString{}
	}
	return OptString{S: s, OK: true}
}

// String returns the string, "null" when absent.
func (o OptString) String() string {
	if !o.OK {
		return "null"
	}
	return o.S
}

// LinkArg holds the label and the end texts of a link: its label,
// cardinalities (quantifiers), roles, qualifiers (kal) and label distance
// and angle, with its length and the visibility of its label.
type LinkArg struct {
	label              *klimt.Display
	quantifier1        OptString
	quantifier2        OptString
	role1              OptString
	role2              OptString
	labeldistance      OptString
	labelangle         OptString
	kal1               OptString
	kal2               OptString
	visibilityModifier klimt.VisibilityModifier
	hasVisibility      bool
	length             int
}

// LinkArgBuild is upstream's LinkArg.build(label, length): the visibility
// character of the label is managed.
func LinkArgBuild(label *klimt.Display, length int) *LinkArg {
	return LinkArgBuildManage(label, length, true)
}

// LinkArgNoDisplay is upstream's LinkArg.noDisplay: a link without label.
func LinkArgNoDisplay(length int) *LinkArg {
	return LinkArgBuildManage(klimt.DisplayNULL(), length, true)
}

// LinkArgBuildManage is upstream's LinkArg.build(label, length,
// manageVisibilityModifier): the label with guillemets, and its visibility
// character turned into a visibility when asked.
func LinkArgBuildManage(label *klimt.Display, length int, manageVisibilityModifier bool) *LinkArg {
	result := &LinkArg{length: length}
	if klimt.DisplayIsNull(label) {
		result.label = klimt.DisplayNULL()
	} else {
		result.label = label.ManageGuillemet(manageVisibilityModifier)
		if manageVisibilityModifier && klimt.VisibilityIsVisibilityCharacter(label.Get(0).String()) {
			result.visibilityModifier, result.hasVisibility = klimt.VisibilityOf(label.Get(0).String(), false)
		}
	}
	return result
}

func (a *LinkArg) copy() *LinkArg {
	c := *a
	return &c
}

// WithQuantifier returns the arguments with the cardinalities of both ends.
func (a *LinkArg) WithQuantifier(quantifier1, quantifier2 OptString) *LinkArg {
	c := a.copy()
	c.quantifier1, c.quantifier2 = quantifier1, quantifier2
	return c
}

// WithRole returns the arguments with the roles of both ends.
func (a *LinkArg) WithRole(role1, role2 OptString) *LinkArg {
	c := a.copy()
	c.role1, c.role2 = role1, role2
	return c
}

// WithKal returns the arguments with the qualifiers of both ends.
func (a *LinkArg) WithKal(kal1, kal2 OptString) *LinkArg {
	c := a.copy()
	c.kal1, c.kal2 = kal1, kal2
	return c
}

// WithDistanceAngle returns the arguments with the label distance and
// angle.
func (a *LinkArg) WithDistanceAngle(labeldistance, labelangle OptString) *LinkArg {
	c := a.copy()
	c.labeldistance, c.labelangle = labeldistance, labelangle
	return c
}

// Inv returns the arguments seen from the other end.
func (a *LinkArg) Inv() *LinkArg {
	c := a.copy()
	c.quantifier1, c.quantifier2 = a.quantifier2, a.quantifier1
	c.kal1, c.kal2 = a.kal2, a.kal1
	c.role1, c.role2 = a.role2, a.role1
	return c
}

// Label returns the label.
func (a *LinkArg) Label() *klimt.Display { return a.label }

// Length returns the length (the number of dashes).
func (a *LinkArg) Length() int { return a.length }

// Quantifier1 returns the cardinality of the first end.
func (a *LinkArg) Quantifier1() OptString { return a.quantifier1 }

// Quantifier2 returns the cardinality of the second end.
func (a *LinkArg) Quantifier2() OptString { return a.quantifier2 }

// Labeldistance returns the label distance.
func (a *LinkArg) Labeldistance() OptString { return a.labeldistance }

// Labelangle returns the label angle.
func (a *LinkArg) Labelangle() OptString { return a.labelangle }

// VisibilityModifier returns the visibility of the label; ok is false for
// none.
func (a *LinkArg) VisibilityModifier() (klimt.VisibilityModifier, bool) {
	return a.visibilityModifier, a.hasVisibility
}

// SetVisibilityModifier sets the visibility of the label; ok false clears
// it.
func (a *LinkArg) SetVisibilityModifier(v klimt.VisibilityModifier, ok bool) {
	a.visibilityModifier, a.hasVisibility = v, ok
}

// SetLength sets the length.
func (a *LinkArg) SetLength(length int) { a.length = length }

// Kal1 returns the qualifier of the first end.
func (a *LinkArg) Kal1() OptString { return a.kal1 }

// Kal2 returns the qualifier of the second end.
func (a *LinkArg) Kal2() OptString { return a.kal2 }

// Role1 returns the role of the first end.
func (a *LinkArg) Role1() OptString { return a.role1 }

// Role2 returns the role of the second end.
func (a *LinkArg) Role2() OptString { return a.role2 }
