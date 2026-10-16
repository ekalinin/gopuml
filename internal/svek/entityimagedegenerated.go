// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/EntityImageDegenerated.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// entityImageDegeneratedDelta is the margin around the single entity.
const entityImageDegeneratedDelta = 7

// EntityImageDegenerated is the drawing of a diagram with a single entity
// and no link, which is drawn without Graphviz.
type EntityImageDegenerated struct {
	orig      klimt.IEntityImage
	backcolor color.HColor
}

var _ klimt.IEntityImage = (*EntityImageDegenerated)(nil)

// IsHidden reports whether the entity is hidden.
func (d *EntityImageDegenerated) IsHidden() bool { return d.orig.IsHidden() }

// Backcolor returns the background color of the document.
func (d *EntityImageDegenerated) Backcolor() color.HColor { return d.backcolor }

// CalculateDimension returns the size of the entity with its margins.
func (d *EntityImageDegenerated) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return d.orig.CalculateDimension(stringBounder).DeltaWH(entityImageDegeneratedDelta*2,
		entityImageDegeneratedDelta*2)
}

// MinMax returns the bounds of the entity.
func (d *EntityImageDegenerated) MinMax(stringBounder font.StringBounder) geom.MinMax {
	return d.orig.MinMax(stringBounder)
}

// InnerPosition returns the position of a member in the entity.
func (d *EntityImageDegenerated) InnerPosition(member string, stringBounder font.StringBounder) *geom.XRectangle2D {
	return d.orig.InnerPosition(member, stringBounder)
}

// DrawU draws the entity with its margins.
func (d *EntityImageDegenerated) DrawU(ug klimt.UGraphic) {
	d.orig.DrawU(ug.Apply(klimt.NewUTranslate(entityImageDegeneratedDelta, entityImageDegeneratedDelta)))
	dim := d.CalculateDimension(ug.StringBounder())
	ug.Apply(klimt.NewUTranslate(dim.Width()-entityImageDegeneratedDelta, dim.Height()-entityImageDegeneratedDelta)).
		Draw(klimt.NewUEmpty(entityImageDegeneratedDelta, entityImageDegeneratedDelta))
}

// MagneticBorder returns the border of the entity, moved by the margin.
func (d *EntityImageDegenerated) MagneticBorder() klimt.MagneticBorder {
	return klimt.MagneticBorderFunc(func(stringBounder font.StringBounder, position geom.XPoint2D) klimt.UTranslate {
		return d.orig.MagneticBorder().ForceAt(stringBounder,
			position.Move(entityImageDegeneratedDelta, entityImageDegeneratedDelta))
	})
}

// ShapeType returns the shape of the entity.
func (d *EntityImageDegenerated) ShapeType() klimt.ShapeType { return d.orig.ShapeType() }

// Shield returns the shield of the entity.
func (d *EntityImageDegenerated) Shield(stringBounder font.StringBounder) klimt.Margins {
	return d.orig.Shield(stringBounder)
}

// OverscanX returns the overscan of the entity.
func (d *EntityImageDegenerated) OverscanX(stringBounder font.StringBounder) float64 {
	return d.orig.OverscanX(stringBounder)
}

// IsCrash returns false.
func (d *EntityImageDegenerated) IsCrash() bool { return false }
