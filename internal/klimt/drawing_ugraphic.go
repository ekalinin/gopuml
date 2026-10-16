// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/UGraphic.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/UDriver.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"io"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
)

// UGraphic is a drawing surface. Apply returns a new surface with a change
// (translation, colors, stroke, clip, ...) applied; the receiver is left
// unchanged, so callers pass derived surfaces down the drawing code.
type UGraphic interface {
	StringBounder() font.StringBounder
	Param() UParam
	Draw(shape UShape)
	Apply(change UChange) UGraphic
	ColorMapper() *color.ColorMapper
	StartURL(url *URL)
	CloseURL()
	StartGroup(group *UGroup)
	CloseGroup()
	FlushUg()
	MatchesProperty(propertyName string) bool
	DefaultBackground() color.HColor
	// WriteToStream writes the drawing; upstream throws IOException.
	WriteToStream(w io.Writer, metadata string, dpi int) error
}

// UDriver draws one kind of shape at (x, y) on an output object of type O.
type UDriver[S UShape, O any] interface {
	Draw(shape S, x, y float64, mapper *color.ColorMapper, param UParam, object O)
}

// UDriverFunc adapts a function to UDriver.
type UDriverFunc[S UShape, O any] func(shape S, x, y float64, mapper *color.ColorMapper, param UParam, object O)

// Draw calls f.
func (f UDriverFunc[S, O]) Draw(shape S, x, y float64, mapper *color.ColorMapper, param UParam, object O) {
	f(shape, x, y, mapper, param, object)
}
