// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/UStroke.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

// UStroke is a line style: dash lengths and thickness. It is a comparable
// value: == is upstream's equals (field by field, with Java's double
// comparison).
type UStroke struct {
	dashVisible float64
	dashSpace   float64
	thickness   float64
}

// NewUStroke returns a stroke; a zero dashVisible means a plain line.
func NewUStroke(dashVisible, dashSpace, thickness float64) UStroke {
	return UStroke{dashVisible: dashVisible, dashSpace: dashSpace, thickness: thickness}
}

// UStrokeWithThickness returns a plain line of the given thickness.
func UStrokeWithThickness(thickness float64) UStroke { return UStroke{thickness: thickness} }

// UStrokeSimple returns the plain line of thickness 1.
func UStrokeSimple() UStroke { return UStroke{thickness: 1.0} }

// HashCode is upstream's hashCode.
func (s UStroke) HashCode() int32 {
	return javaDoubleHashCode(s.dashVisible) + javaDoubleHashCode(s.dashSpace) + javaDoubleHashCode(s.thickness)
}

// String is upstream's toString, for example "0.0-0.0-1.0".
func (s UStroke) String() string {
	return jstr(s.dashVisible) + "-" + jstr(s.dashSpace) + "-" + jstr(s.thickness)
}

// OnlyThickness returns the plain line of the same thickness.
func (s UStroke) OnlyThickness() UStroke { return UStroke{thickness: s.thickness} }

// DashVisible returns the length of the visible dashes.
func (s UStroke) DashVisible() float64 { return s.dashVisible }

// DashSpace returns the length of the gaps.
func (s UStroke) DashSpace() float64 { return s.dashSpace }

// Thickness returns the line width.
func (s UStroke) Thickness() float64 { return s.thickness }

// DasharraySvg returns the SVG dash array, nil for a plain line.
func (s UStroke) DasharraySvg() []float64 {
	if s.dashVisible == 0 {
		return nil
	}
	return []float64{s.dashVisible, s.dashSpace}
}

// DashTikz returns the TikZ dash pattern; ok is false for a plain line (null
// upstream).
func (s UStroke) DashTikz() (pattern string, ok bool) {
	if s.dashVisible == 0 {
		return "", false
	}
	return "on " + jstr(s.dashVisible) + "pt off " + jstr(s.dashSpace) + "pt", true
}

// strokePtr returns a pointer to a copy of s, for the APIs where upstream's
// stroke may be null.
func strokePtr(s UStroke) *UStroke { return &s }
