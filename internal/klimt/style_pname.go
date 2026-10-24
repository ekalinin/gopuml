// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/PName.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/StyleScheme.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/MergeStrategy.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/LengthAdjust.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

// PName is the name of a style property.
type PName int

// Style properties, in upstream's order.
const (
	PNameShadowing PName = iota
	PNameFontName
	PNameFontColor
	PNameFontSize
	PNameFontStyle
	// PNameFontWeight is the CSS font-weight: a keyword (normal, bold,
	// lighter, bolder) or a number from 100 to 900.
	PNameFontWeight
	PNameBackGroundColor
	PNameRoundCorner
	PNameLineThickness
	PNameDiagonalCorner
	PNameHyperLinkColor
	PNameHyperlinkUnderlineStyle
	PNameHyperlinkUnderlineThickness
	PNameHeadColor
	PNameLineColor
	PNameLineStyle
	PNamePadding
	PNameMargin
	PNameMaximumWidth
	PNameMinimumWidth
	PNameExportedName
	PNameImage
	PNameHorizontalAlignment
	PNameShowStereotype
	PNameImagePosition
	// Chart-specific properties
	PNameMarkerShape
	PNameMarkerSize
	PNameMarkerColor
	PNameBarWidth
	PNameWidth
	pnameCount
)

var pnameNames = [...]string{"Shadowing", "FontName", "FontColor", "FontSize", "FontStyle", "FontWeight",
	"BackGroundColor", "RoundCorner", "LineThickness", "DiagonalCorner", "HyperLinkColor",
	"HyperlinkUnderlineStyle", "HyperlinkUnderlineThickness", "HeadColor", "LineColor", "LineStyle", "Padding",
	"Margin", "MaximumWidth", "MinimumWidth", "ExportedName", "Image", "HorizontalAlignment", "ShowStereotype",
	"ImagePosition", "MarkerShape", "MarkerSize", "MarkerColor", "BarWidth", "Width"}

// String returns upstream's constant name, for example "BackGroundColor".
func (p PName) String() string { return enumName("PName", pnameNames[:], int(p)) }

// PNameFromName returns the property named name, ignoring case; ok is false
// when there is none (null upstream). The scheme is unused, as upstream.
func PNameFromName(name string, scheme StyleScheme) (PName, bool) {
	for i, n := range pnameNames {
		if javaEqualsIgnoreCase(n, name) {
			return PName(i), true
		}
	}
	return 0, false
}

// StyleScheme tells whether style values are for the regular or the dark
// mode ("@media dark").
type StyleScheme int

// Style schemes, in upstream's order.
const (
	StyleSchemeRegular StyleScheme = iota
	StyleSchemeDark
)

// MergeStrategy tells how merged styles treat the values of stereotype
// styles.
type MergeStrategy int

// Merge strategies, in upstream's order.
const (
	MergeKeepExistingValueOfStereotype MergeStrategy = iota
	MergeOverwriteExistingValue
)

// LengthAdjust is the SVG lengthAdjust of texts.
type LengthAdjust int

// Length adjustments, in upstream's order.
const (
	LengthAdjustNone LengthAdjust = iota
	LengthAdjustSpacing
	LengthAdjustSpacingAndGlyphs
)

// String returns upstream's constant name.
func (l LengthAdjust) String() string {
	return enumName("LengthAdjust", []string{"NONE", "SPACING", "SPACING_AND_GLYPHS"}, int(l))
}

// LengthAdjustDefault returns Spacing.
func LengthAdjustDefault() LengthAdjust { return LengthAdjustSpacing }
