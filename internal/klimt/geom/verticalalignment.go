// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/VerticalAlignment.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

// VerticalAlignment is a vertical alignment.
type VerticalAlignment int

// Vertical alignments, in upstream's order.
const (
	VerticalAlignmentTop VerticalAlignment = iota
	VerticalAlignmentCenter
	VerticalAlignmentBottom
)

var verticalAlignmentNames = []string{"TOP", "CENTER", "BOTTOM"}

// String returns upstream's name, for example "TOP".
func (a VerticalAlignment) String() string {
	return enumName("VerticalAlignment", verticalAlignmentNames, a)
}

// VerticalAlignmentValueOf returns the alignment named exactly s, as Java's
// valueOf.
func VerticalAlignmentValueOf(s string) (VerticalAlignment, bool) {
	return enumValueOf[VerticalAlignment](verticalAlignmentNames, s)
}

// VerticalAlignmentFromString returns VerticalAlignmentTop when s is "TOP"
// ignoring case, as Java's equalsIgnoreCase, and VerticalAlignmentBottom
// otherwise: like upstream, it never returns VerticalAlignmentCenter.
func VerticalAlignmentFromString(s string) VerticalAlignment {
	if javaEqualsIgnoreCase(VerticalAlignmentTop.String(), s) {
		return VerticalAlignmentTop
	}
	if javaEqualsIgnoreCase(VerticalAlignmentBottom.String(), s) {
		return VerticalAlignmentBottom
	}
	return VerticalAlignmentBottom
}
