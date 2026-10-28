// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/wbs/WBSLink.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package wbs

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/svek/extremity"
	"github.com/ekalinin/gopuml/internal/uml"
)

// wbsLink is a link from one node to another, drawn over the tree.
type wbsLink struct {
	element1 *wElement
	element2 *wElement
	color    color.HColor
}

// newWBSLink returns the link; a nil color throws IllegalArgumentException,
// as upstream.
func newWBSLink(element1, element2 *wElement, c color.HColor) *wbsLink {
	if c == nil {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return &wbsLink{element1: element1, element2: element2, color: c}
}

// DrawU implements klimt.UDrawable: an arrow between the centers of the two
// nodes, from the border of the first to the border of the second, when
// both were drawn.
func (l *wbsLink) DrawU(ug klimt.UGraphic) {
	position1 := l.element1.position
	position2 := l.element2.position
	dim1 := l.element1.dimension
	dim2 := l.element2.dimension
	if position1 == nil || position2 == nil {
		return
	}
	rect1 := geom.NewXRectangle2D(position1.Dx(), position1.Dy(), dim1.Width(), dim1.Height())
	rect2 := geom.NewXRectangle2D(position2.Dx(), position2.Dy(), dim2.Width(), dim2.Height())
	line := geom.NewXLine2D(rect1.CenterX(), rect1.CenterY(), rect2.CenterX(), rect2.CenterY())
	c1, ok1 := rect1.Intersect(line)
	c2, ok2 := rect2.Intersect(line)
	// Upstream builds the line from the two crossings, which are null when
	// the line does not cross a border.
	if !ok1 {
		uml.ThrowNPE(`Cannot invoke "net.sourceforge.plantuml.klimt.geom.XPoint2D.getX()" because "p1" is null`)
	}
	if !ok2 {
		uml.ThrowNPE(`Cannot invoke "net.sourceforge.plantuml.klimt.geom.XPoint2D.getX()" because "p2" is null`)
	}
	line = geom.XLine2DLine(c1, c2)
	ug = ug.Apply(l.color)
	// Upstream's XLine2D.drawU.
	ug.Apply(klimt.NewUTranslate(line.X1, line.Y1)).Draw(klimt.NewULine(line.X2-line.X1, line.Y2-line.Y1))
	angle := line.Angle()
	arrow := extremity.NewArrow(c2, angle)
	arrow.DrawU(ug)
}
