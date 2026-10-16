// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

// Accessors of ArrowConfiguration fields that upstream keeps private. The
// sequence diagram model tests dump them to compare the model with
// upstream's (whose harness reads the fields by reflection).

// Thickness returns the thickness of the arrow.
func (a *ArrowConfiguration) Thickness() float64 { return a.thickness }

// Inclination returns the raw inclination of the arrow, before
// Inclination1 and Inclination2 select the end it applies to.
func (a *ArrowConfiguration) Inclination() int { return a.inclination }
