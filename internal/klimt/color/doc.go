// SPDX-License-Identifier: MIT
// Provenance: original

// Package color is PlantUML's color model: XColor values, the HColor kinds
// diagrams use (simple colors, automatic colors, color schemes, gradients),
// the color parser with the HTML color names, the per-element Colors of a
// command, the color mappers for dark mode, monochrome and reversed output,
// and the HSL and HSLuv arithmetic of the preprocessor color functions.
//
// Upstream's unchecked exceptions are panics with a *RuntimeError value, as
// elsewhere in this module; checked exceptions are errors.
package color
