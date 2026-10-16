// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/geom/ImgValign.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package geom

// ImgValign is the vertical alignment of an image.
type ImgValign int

// Values of ImgValign, in upstream's order.
const (
	ImgValignTop ImgValign = iota
	ImgValignBottom
	ImgValignMiddle
)

var imgValignNames = []string{"TOP", "BOTTOM", "MIDDLE"}

// String returns upstream's name, for example "TOP".
func (v ImgValign) String() string { return enumName("ImgValign", imgValignNames, v) }

// ImgValignValueOf returns the value named exactly s, as Java's valueOf.
func ImgValignValueOf(s string) (ImgValign, bool) {
	return enumValueOf[ImgValign](imgValignNames, s)
}
