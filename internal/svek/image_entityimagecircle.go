// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/image/EntityImageCircleStart.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/image/EntityImageCircleEnd.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// circleGroup is the group of a start or end circle: the group of the
// entity with its own class.
func circleGroup(e *abel.Entity, class string) *klimt.UGroup {
	group := entityGroup(e)
	group.Put(klimt.UGroupClass, class)
	return group
}

// circleStyle is the style of a start or end circle of the diagram.
func circleStyle(skinParam klimt.ISkinParam, end klimt.SName) *klimt.Style {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement,
		uml.DiagramTypeStyleName(skinParam.DiagramType()), klimt.SNameCircle, end).
		MergedStyle(skinParam.CurrentStyleBuilder())
}

// EntityImageCircleStart is upstream's EntityImageCircleStart: the filled
// circle of a start.
type EntityImageCircleStart struct {
	abstractEntityImage
	circle *klimt.CircleStart
}

var _ klimt.IEntityImage = (*EntityImageCircleStart)(nil)

func newEntityImageCircleStart(entity *abel.Entity) *EntityImageCircleStart {
	e := &EntityImageCircleStart{abstractEntityImage: newAbstractEntityImage(entity)}
	// A junction pseudostate is the smaller circle of the initial node.
	junction := entity.LeafType() == abel.LeafTypeJunction
	e.circle = klimt.NewCircleStartJunction(e.skinParam, circleStyle(e.skinParam, klimt.SNameStart), entity.Colors(),
		junction)
	return e
}

// CalculateDimension returns the size of the circle.
func (e *EntityImageCircleStart) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return e.circle.CalculateDimension(stringBounder)
}

// DrawU draws the circle.
func (e *EntityImageCircleStart) DrawU(ug klimt.UGraphic) {
	ug.StartGroup(circleGroup(e.entity, "start_entity"))
	e.circle.DrawU(ug)
	ug.CloseGroup()
}

// ShapeType returns CIRCLE.
func (e *EntityImageCircleStart) ShapeType() klimt.ShapeType { return klimt.ShapeTypeCircle }

// circleEndSize is the size of upstream's EntityImageCircleEnd.
const circleEndSize = 22

// EntityImageCircleEnd is upstream's EntityImageCircleEnd: the circled dot
// of an end.
type EntityImageCircleEnd struct {
	abstractEntityImage
	circle *klimt.CircleEnd
}

var _ klimt.IEntityImage = (*EntityImageCircleEnd)(nil)

func newEntityImageCircleEnd(entity *abel.Entity) *EntityImageCircleEnd {
	e := &EntityImageCircleEnd{abstractEntityImage: newAbstractEntityImage(entity)}
	e.circle = klimt.NewCircleEnd(e.skinParam, circleStyle(e.skinParam, klimt.SNameEnd), entity.Colors())
	return e
}

// CalculateDimension returns 22 x 22.
func (e *EntityImageCircleEnd) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(circleEndSize, circleEndSize)
}

// DrawU draws the circle.
func (e *EntityImageCircleEnd) DrawU(ug klimt.UGraphic) {
	ug.StartGroup(circleGroup(e.entity, "end_entity"))
	e.circle.DrawU(ug)
	ug.CloseGroup()
}

// ShapeType returns CIRCLE.
func (e *EntityImageCircleEnd) ShapeType() klimt.ShapeType { return klimt.ShapeTypeCircle }
