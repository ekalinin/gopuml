// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/geom/HorizontalAlignment.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

import "strings"

// HorizontalAlignment is a horizontal alignment.
//
// Upstream's draw(UGraphic, TextBlock, ...) and
// asPlacementStrategy(StringBounder) are not ported yet.
type HorizontalAlignment int

// Horizontal alignments, in upstream's order.
const (
	HorizontalAlignmentLeft HorizontalAlignment = iota
	HorizontalAlignmentCenter
	HorizontalAlignmentRight
)

var horizontalAlignmentNames = []string{"LEFT", "CENTER", "RIGHT"}

// String returns upstream's name, for example "LEFT".
func (a HorizontalAlignment) String() string {
	return enumName("HorizontalAlignment", horizontalAlignmentNames, a)
}

// HorizontalAlignmentValueOf returns the alignment named exactly s, as Java's
// valueOf.
func HorizontalAlignmentValueOf(s string) (HorizontalAlignment, bool) {
	return enumValueOf[HorizontalAlignment](horizontalAlignmentNames, s)
}

// HorizontalAlignmentFromString returns the alignment whose name equals s
// ignoring case, as Java's equalsIgnoreCase. The boolean is false when there
// is none (upstream returns null).
func HorizontalAlignmentFromString(s string) (HorizontalAlignment, bool) {
	for _, a := range [...]HorizontalAlignment{HorizontalAlignmentLeft, HorizontalAlignmentCenter,
		HorizontalAlignmentRight} {
		if javaEqualsIgnoreCase(a.String(), s) {
			return a, true
		}
	}
	return 0, false
}

// HorizontalAlignmentFromStringDefault returns the alignment named s ignoring
// case, or defaultValue when there is none. It is upstream's
// fromString(String, HorizontalAlignment); a null s is the caller's concern.
func HorizontalAlignmentFromStringDefault(s string, defaultValue HorizontalAlignment) HorizontalAlignment {
	// Upstream upper-cases s with full case mappings first. The mappings
	// strings.ToUpper lacks (such as "ß" to "SS") cannot produce a name.
	s = strings.ToUpper(s)
	result, ok := HorizontalAlignmentFromString(s)
	if !ok {
		return defaultValue
	}
	return result
}

// GraphVizValue returns the first letter of the name in lower case: "l", "c"
// or "r".
func (a HorizontalAlignment) GraphVizValue() string {
	return strings.ToLower(a.String()[:1])
}

// Position returns the offset of an item of the given width inside
// fullWidth.
func (a HorizontalAlignment) Position(width, fullWidth float64) float64 {
	if a == HorizontalAlignmentRight {
		return fullWidth - width
	} else if a == HorizontalAlignmentCenter {
		return (fullWidth - width) / 2
	}
	return 0
}
