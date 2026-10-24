// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/asciiverse/AElseSeparator.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package asciiverse

// AElseSeparator is the dashed divider across a group frame at an "else":
// its width is the frame's, given by the group.
type AElseSeparator struct {
	width int
	label []uint16
}

// NewAElseSeparator returns the divider of a width with a label; a null
// label (ok false) is empty.
func NewAElseSeparator(width int, label string, ok bool) *AElseSeparator {
	if !ok {
		label = ""
	}
	return &AElseSeparator{width: width, label: javaChars(label)}
}

// AsciiDimension returns the width by one row.
func (s *AElseSeparator) AsciiDimension() ADimension2D { return NewADimension2D(s.width, 1) }

// AsciiDraw draws a dotted line inside the frame borders (columns 1 to
// width-2), then the label near the left corner, between spaces, when it
// fits.
func (s *AElseSeparator) AsciiDraw(plan *InfinitePlan) {
	if s.width > 2 {
		plan.DrawHLineDotted(1, s.width-2, 0, true)
	}
	if len(s.label) > 0 && s.width > len(s.label)+4 {
		plan.Move(2, 0).drawChars(spaced(s.label))
	}
}

// spaced returns " " + chars + " ".
func spaced(chars []uint16) []uint16 {
	result := make([]uint16, 0, len(chars)+2)
	result = append(result, ' ')
	result = append(result, chars...)
	return append(result, ' ')
}
