// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityDiamond.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/ExtremityFactoryDiamond.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// diamond is upstream's ExtremityDiamond: the diamond of a composition
// (filled) or an aggregation.
type diamond struct {
	extremityBase
	polygon     *klimt.UPolygon
	fill        bool
	contact     geom.XPoint2D
	deltaForKal klimt.UTranslate
}

func newDiamond(p1 geom.XPoint2D, angle float64, fill bool) *diamond {
	e := &diamond{polygon: klimt.NewUPolygon(), fill: fill, contact: geom.NewXPoint2D(p1.X, p1.Y)}
	angle = manageround(angle)
	e.polygon.AddPoint(0, 0)
	const xWing = 6
	const yAperture = 4
	e.polygon.AddPoint(-xWing, -yAperture)
	e.polygon.AddPoint(-xWing*2, 0)
	e.polygon.AddPoint(-xWing, yAperture)
	e.polygon.AddPoint(0, 0)
	e.polygon.Rotate(angle + pi/2)
	e.deltaForKal = klimt.UTranslatePoint(polygonPoint(e.polygon, 2)).Reverse()
	e.polygon = e.polygon.Translate(p1.X, p1.Y)
	return e
}

// SomePoint returns the contact point.
func (e *diamond) SomePoint() *geom.XPoint2D { return point(e.contact) }

// DeltaForKal returns the move to the far corner of the diamond.
func (e *diamond) DeltaForKal() klimt.UTranslate { return e.deltaForKal }

// DecorationLength returns 12.
func (e *diamond) DecorationLength() float64 { return 12 }

// DrawU draws the diamond, filled with the current color for a composition.
func (e *diamond) DrawU(ug klimt.UGraphic) {
	if e.fill {
		ug = ug.Apply(klimt.ChangeBack(ug))
	} else {
		ug = ug.Apply(klimt.Bg(color.None()))
	}
	ug.Draw(e.polygon)
}

// FactoryDiamond is upstream's ExtremityFactoryDiamond.
type FactoryDiamond struct {
	abstractFactory
	fill bool
}

// NewFactoryDiamond returns the factory of diamonds.
func NewFactoryDiamond(fill bool) *FactoryDiamond {
	return &FactoryDiamond{abstractFactory{"ExtremityFactoryDiamond"}, fill}
}

// CreateUDrawable returns the diamond at p0.
func (f *FactoryDiamond) CreateUDrawable(p0 geom.XPoint2D, angle float64, _ *geom.Side) klimt.UDrawable {
	return newDiamond(p0, angle-pi/2, f.fill)
}

// CreateTBRDrawableLegacy returns the diamond at p1.
func (f *FactoryDiamond) CreateTBRDrawableLegacy(p0, p1, p2 geom.XPoint2D, _ *geom.Side) klimt.UDrawable {
	ortho := atan2(p0, p2)
	return newDiamond(p1, ortho, f.fill)
}
