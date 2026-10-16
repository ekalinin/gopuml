// SPDX-License-Identifier: MIT
// Provenance: original

// Package klimt is the drawing model: shapes, text blocks and the UGraphic
// abstraction that output drivers implement, displays and stereotypes,
// links, the style system (CSS-like style sheets, the embedded skins,
// skinparam conversion) and the skin parameters.
//
// Upstream's packages klimt, klimt.shape, klimt.drawing, klimt.creole,
// klimt.sprite, style, style.parser, skin, stereo and url reference each
// other, so they form this one package (design D2); file names carry the
// origin package as prefix (shape_, drawing_, creole_, style_, skin_,
// stereo_, url_, ...). The leaf packages color, geom and font stay separate.
//
// Upstream's unchecked exceptions are panics with a *color.RuntimeError
// value; checked exceptions are errors. Code paths whose upstream classes are
// not ported yet panic with a message naming them (see HANDOFF.md).
package klimt
