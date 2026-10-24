// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/compress/CompressionMode.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/compress/UShapeIgnorableForCompression.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

// CompressionMode is the axis of upstream's layout compression
// (klimt.compress, whose other classes are not ported in this step).
type CompressionMode int

// Compression axes, in upstream's order.
const (
	CompressionOnX CompressionMode = iota
	CompressionOnY
)

// String returns upstream's constant name.
func (m CompressionMode) String() string {
	return enumName("CompressionMode", []string{"ON_X", "ON_Y"}, int(m))
}

// UShapeIgnorableForCompression is a shape that layout compression may skip.
type UShapeIgnorableForCompression interface {
	IsIgnoreForCompressionOn(mode CompressionMode) bool
	DrawWhenCompressed(ug UGraphic, mode CompressionMode)
}
