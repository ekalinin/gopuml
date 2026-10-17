// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/EntityDomain.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// entityDomainSuppY is the gap between the circle and the bar of the
// entity head.
const entityDomainSuppY = 2

// EntityDomain is the head of an entity participant: a circle above a bar.
// Upstream defines it in svek; it lives here because the skin components
// (skin.rose) and the symbols (decoration.symbol) draw it.
type EntityDomain struct {
	TextBlockDefaults
	symbolContext *Fashion
}

// NewEntityDomain returns the entity head drawn with symbolContext.
func NewEntityDomain(symbolContext *Fashion) *EntityDomain {
	return &EntityDomain{symbolContext: symbolContext}
}

// DrawU draws the circle and the bar.
func (e *EntityDomain) DrawU(ug UGraphic) {
	x := 0.0
	y := 0.0
	x += svekHeadMargin
	y += svekHeadMargin
	ug = e.symbolContext.Apply(ug)
	circle := UEllipseBuild(svekHeadRadius*2, svekHeadRadius*2)
	circle.SetDeltaShadow(e.symbolContext.DeltaShadow())
	ug.Apply(NewUTranslate(x, y)).Draw(circle)
	ug.Apply(NewUTranslate(x, y+2*svekHeadRadius+entityDomainSuppY)).Draw(ULineH(2 * svekHeadRadius))
}

// CalculateDimension returns the size of the head with its margin.
func (e *EntityDomain) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(svekHeadRadius*2+2*svekHeadMargin, svekHeadRadius*2+2*svekHeadMargin)
}
