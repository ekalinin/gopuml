// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/shape/TextBlock.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/shape/UDrawable.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/shape/Line.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/MagneticBorder.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/MagneticBorderNone.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// UDrawable is something that draws itself.
type UDrawable interface {
	DrawU(ug UGraphic)
}

// TextBlock is a drawable block with a size: the building block of every
// diagram element. It is also a UShape: drawing a TextBlock on a UGraphic
// draws its content.
//
// Implementations embed TextBlockDefaults for upstream's default methods.
type TextBlock interface {
	UDrawable
	CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D
	// MinMax returns the bounds of what the block draws. Upstream's default
	// throws UnsupportedOperationException.
	MinMax(stringBounder font.StringBounder) geom.MinMax
	// InnerPosition returns the position of a member inside the block, nil
	// when there is none. Upstream's default throws
	// UnsupportedOperationException.
	InnerPosition(member string, stringBounder font.StringBounder) *geom.XRectangle2D
	MagneticBorder() MagneticBorder
	// Backcolor returns the background color, nil for none.
	Backcolor() color.HColor
}

// TextBlockDefaults provides upstream's default TextBlock methods; embed it
// in TextBlock implementations.
type TextBlockDefaults struct{}

// MinMax panics, as upstream's default throws UnsupportedOperationException.
func (TextBlockDefaults) MinMax(font.StringBounder) geom.MinMax {
	throw(unsupportedOperation, "class net.sourceforge.plantuml.klimt.shape.TextBlock")
	return geom.MinMax{}
}

// InnerPosition panics, as upstream's default throws
// UnsupportedOperationException.
func (TextBlockDefaults) InnerPosition(member string, _ font.StringBounder) *geom.XRectangle2D {
	throw(unsupportedOperation, "member="+member+" class net.sourceforge.plantuml.klimt.shape.TextBlock")
	return nil
}

// MagneticBorder returns the border without force.
func (TextBlockDefaults) MagneticBorder() MagneticBorder { return MagneticBorderNone{} }

// Backcolor returns nil.
func (TextBlockDefaults) Backcolor() color.HColor { return nil }

// Line is a text block that is one line of a text.
type Line interface {
	TextBlock
	HorizontalAlignment() geom.HorizontalAlignment
}

// MagneticBorder is the force that pushes things away from a block. Upstream
// defines it in klimt.geom; it lives here because it returns a UTranslate.
type MagneticBorder interface {
	ForceAt(stringBounder font.StringBounder, position geom.XPoint2D) UTranslate
}

// MagneticBorderNone is the border without force.
type MagneticBorderNone struct{}

// ForceAt returns the null translation.
func (MagneticBorderNone) ForceAt(font.StringBounder, geom.XPoint2D) UTranslate {
	return UTranslateNone()
}

// MagneticBorderFunc adapts a function to MagneticBorder.
type MagneticBorderFunc func(stringBounder font.StringBounder, position geom.XPoint2D) UTranslate

// ForceAt calls f.
func (f MagneticBorderFunc) ForceAt(stringBounder font.StringBounder, position geom.XPoint2D) UTranslate {
	return f(stringBounder, position)
}

// dim converts a measured text size to upstream's XDimension2D.
func dim(d font.Dimension) geom.XDimension2D { return geom.NewXDimension2D(d.Width, d.Height) }
