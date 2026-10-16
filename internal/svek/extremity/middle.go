// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/MiddleCircle.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/MiddleCircleCircled.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/MiddleCircleCircledMode.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/MiddleFactoryCircle.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/MiddleFactoryCircleCircled.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/MiddleFactorySubset.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/extremity/MiddleSubset.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/plantuml-in-go/internal/jmath"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// middleCircle is upstream's MiddleCircle: a circle in the middle of a link.
type middleCircle struct {
	extremityBase
	backColor color.HColor
}

const middleCircleRadius = 6.0

// SomePoint returns nil.
func (e *middleCircle) SomePoint() *geom.XPoint2D { return nil }

// DecorationLength returns the radius.
func (e *middleCircle) DecorationLength() float64 { return middleCircleRadius }

// DrawU draws the circle centered on the origin.
func (e *middleCircle) DrawU(ug klimt.UGraphic) {
	ug.Apply(klimt.Bg(e.backColor)).Apply(klimt.UStrokeWithThickness(1.5)).
		Apply(klimt.NewUTranslate(-middleCircleRadius, -middleCircleRadius)).
		Draw(klimt.UEllipseBuild(2*middleCircleRadius, 2*middleCircleRadius))
}

// MiddleCircleCircledMode is upstream's MiddleCircleCircledMode: which arcs
// surround the circle.
type MiddleCircleCircledMode int

// Modes, in upstream's order.
const (
	MiddleCircleCircledBoth MiddleCircleCircledMode = iota
	MiddleCircleCircledMode1
	MiddleCircleCircledMode2
)

// String returns upstream's constant name.
func (m MiddleCircleCircledMode) String() string {
	return [...]string{"BOTH", "MODE1", "MODE2"}[m]
}

// middleCircleCircled is upstream's MiddleCircleCircled: a circle with one
// or two arcs around it.
type middleCircleCircled struct {
	extremityBase
	diagramBackColor color.HColor
	angle            float64
	mode             MiddleCircleCircledMode
	backColor        color.HColor
}

const (
	middleCircleCircledRadius1 = 6.0
	middleCircleCircledRadius2 = 10.0
)

// SomePoint returns nil.
func (e *middleCircleCircled) SomePoint() *geom.XPoint2D { return nil }

// DrawU draws the arcs and the circle centered on the origin.
func (e *middleCircleCircled) DrawU(ug klimt.UGraphic) {
	const r1, r2 = middleCircleCircledRadius1, middleCircleCircledRadius2
	if e.mode == MiddleCircleCircledBoth {
		ug.Apply(e.diagramBackColor).Apply(klimt.Bg(e.diagramBackColor)).Apply(klimt.NewUTranslate(-r2, -r2)).
			Draw(klimt.UEllipseBuild(2*r2, 2*r2))
	}
	ug = ug.Apply(klimt.Bg(e.backColor))
	ug = ug.Apply(klimt.UStrokeWithThickness(1.5))

	const d = 0.0
	if e.mode == MiddleCircleCircledMode1 || e.mode == MiddleCircleCircledBoth {
		arc1 := klimt.NewUEllipse(2*r2, 2*r2, e.angle, 90)
		ug.Apply(klimt.Bg(color.None())).Apply(klimt.NewUTranslate(-r2+d, -r2+d)).Draw(arc1)
	}
	if e.mode == MiddleCircleCircledMode2 || e.mode == MiddleCircleCircledBoth {
		arc2 := klimt.NewUEllipse(2*r2, 2*r2, e.angle+180, 90)
		ug.Apply(klimt.Bg(color.None())).Apply(klimt.NewUTranslate(-r2+d, -r2+d)).Draw(arc2)
	}
	ug.Apply(klimt.NewUTranslate(-r1, -r1)).Draw(klimt.UEllipseBuild(2*r1, 2*r1))
}

// middleSubset is upstream's MiddleSubset: a half circle with two lines,
// the subset or superset sign.
type middleSubset struct {
	extremityBase
	angle   float64
	reverse bool
}

const (
	middleSubsetRadius = 6.0
	middleSubsetLength = 10.0
)

// SomePoint returns nil.
func (e *middleSubset) SomePoint() *geom.XPoint2D { return nil }

// DecorationLength returns the radius.
func (e *middleSubset) DecorationLength() float64 { return middleSubsetRadius }

// DrawU draws the half circle and the two lines.
func (e *middleSubset) DrawU(ug klimt.UGraphic) {
	const radius, length = middleSubsetRadius, middleSubsetLength
	ug = ug.Apply(klimt.UStrokeWithThickness(1.5))
	rot := 135.0
	if e.reverse {
		rot = -45
	}
	ug.Apply(klimt.NewUTranslate(-radius, -radius)).Draw(klimt.NewUEllipse(2*radius, 2*radius, e.angle+rot, 180))

	sin := jmath.Sin(jmath.ToRadians(e.angle + rot))
	cos := jmath.Cos(jmath.ToRadians(e.angle + rot))

	ug.Apply(klimt.NewUTranslate(float64(radius*cos), float64(radius*-sin))).
		Draw(klimt.NewULine(float64(length*sin), float64(length*cos)))
	ug.Apply(klimt.NewUTranslate(float64(radius*-cos), float64(radius*sin))).
		Draw(klimt.NewULine(float64(length*sin), float64(length*cos)))
}

// MiddleFactoryCircle is upstream's MiddleFactoryCircle.
type MiddleFactoryCircle struct{ backColor color.HColor }

// NewMiddleFactoryCircle returns the factory of middle circles.
func NewMiddleFactoryCircle(backColor color.HColor) *MiddleFactoryCircle {
	return &MiddleFactoryCircle{backColor}
}

// CreateUDrawable returns the circle.
func (f *MiddleFactoryCircle) CreateUDrawable(float64) klimt.UDrawable {
	return &middleCircle{backColor: f.backColor}
}

// MiddleFactoryCircleCircled is upstream's MiddleFactoryCircleCircled.
type MiddleFactoryCircleCircled struct {
	mode                        MiddleCircleCircledMode
	backColor, diagramBackColor color.HColor
}

// NewMiddleFactoryCircleCircled returns the factory of circled middle
// circles.
func NewMiddleFactoryCircleCircled(mode MiddleCircleCircledMode, backColor,
	diagramBackColor color.HColor) *MiddleFactoryCircleCircled {
	return &MiddleFactoryCircleCircled{mode, backColor, diagramBackColor}
}

// CreateUDrawable returns the circled circle with its arcs at angle.
func (f *MiddleFactoryCircleCircled) CreateUDrawable(angle float64) klimt.UDrawable {
	return &middleCircleCircled{angle: angle, mode: f.mode, backColor: f.backColor,
		diagramBackColor: f.diagramBackColor}
}

// MiddleFactorySubset is upstream's MiddleFactorySubset.
type MiddleFactorySubset struct{ reverse bool }

// NewMiddleFactorySubset returns the factory of subset (reverse false) or
// superset signs.
func NewMiddleFactorySubset(reverse bool) *MiddleFactorySubset { return &MiddleFactorySubset{reverse} }

// CreateUDrawable returns the sign at angle.
func (f *MiddleFactorySubset) CreateUDrawable(angle float64) klimt.UDrawable {
	return &middleSubset{angle: angle, reverse: f.reverse}
}
