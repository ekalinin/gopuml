// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/AbstractEntityImageBorder.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/EntityImagePort.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/EntityImageStateBorder.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// abstractEntityImageBorder is upstream's AbstractEntityImageBorder: the
// image of an entity drawn on the border of its cluster.
type abstractEntityImageBorder struct {
	abstractEntityImage
	entityPosition abel.EntityPosition
	parent         *Cluster
	bibliotekon    *Bibliotekon
	rankdir        geom.Rankdir
}

func newAbstractEntityImageBorder(leaf *abel.Entity, parent *Cluster,
	bibliotekon *Bibliotekon) abstractEntityImageBorder {
	b := abstractEntityImageBorder{abstractEntityImage: newAbstractEntityImage(leaf), parent: parent,
		bibliotekon: bibliotekon, entityPosition: leaf.EntityPosition()}
	b.rankdir = b.skinParam.Rankdir()
	if b.entityPosition == abel.EntityPositionNormal {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return b
}

func (b *abstractEntityImageBorder) style(signature klimt.StyleSignatureBasic) *klimt.Style {
	return signature.WithTOBECHANGED(b.entity.Stereotype()).MergedStyle(b.skinParam.CurrentStyleBuilder())
}

func (b *abstractEntityImageBorder) desc(style *klimt.Style) klimt.TextBlock {
	fc := klimt.FontConfigurationFromStyle(b.skinParam, style, nil)
	return b.entity.Display().Create(fc, geom.HorizontalAlignmentCenter, b.skinParam)
}

// EntityImagePort is the square of a port ("portin", "portout") on the
// border of its component, with its name above or below.
type EntityImagePort struct {
	abstractEntityImageBorder
}

var _ klimt.IEntityImage = (*EntityImagePort)(nil)

func newEntityImagePort(leaf *abel.Entity, parent *Cluster, bibliotekon *Bibliotekon) *EntityImagePort {
	return &EntityImagePort{abstractEntityImageBorder: newAbstractEntityImageBorder(leaf, parent, bibliotekon)}
}

func (p *EntityImagePort) styleSignature() klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, p.styleName(), klimt.SNamePort)
}

// upPosition reports whether the port is in the upper half of its
// cluster: its name is then drawn above it.
func (p *EntityImagePort) upPosition() bool {
	if p.parent == nil {
		return false
	}
	clusterCenter := p.parent.rectangleArea.PointCenter()
	node := p.bibliotekon.node(p.entity)
	return node.MinY() < clusterCenter.Y
}

// CalculateDimension returns the size of the square.
func (p *EntityImagePort) CalculateDimension(font.StringBounder) geom.XDimension2D {
	sp := abel.EntityPositionRadius * 2
	return geom.NewXDimension2D(sp, sp)
}

// MaxWidthFromLabelForEntryExit is the width of the name, which the
// layout reserves around the port.
func (p *EntityImagePort) MaxWidthFromLabelForEntryExit(stringBounder font.StringBounder) float64 {
	return p.desc(p.style(p.styleSignature())).CalculateDimension(stringBounder).Width()
}

// DrawU draws the name and the square.
func (p *EntityImagePort) DrawU(ug klimt.UGraphic) {
	style := p.style(p.styleSignature())
	desc := p.desc(style)
	y := 0.0
	dimDesc := desc.CalculateDimension(ug.StringBounder())
	x := 0 - (dimDesc.Width()-2*abel.EntityPositionRadius)/2
	if p.upPosition() {
		y -= 2*abel.EntityPositionRadius + dimDesc.Height()
	} else {
		y += 2 * abel.EntityPositionRadius
	}
	ug.StartGroup(entityGroup(p.entity))
	desc.DrawU(ug.Apply(klimt.NewUTranslate(x, y)))
	set := p.skinParam.IHtmlColorSet()
	backcolor := p.entity.Colors().Color(color.TypeBack)
	borderColor := p.entity.Colors().Color(color.TypeLine)
	if borderColor == nil {
		borderColor = style.Value(klimt.PNameLineColor).AsColor(set)
	}
	if backcolor == nil {
		backcolor = style.Value(klimt.PNameBackGroundColor).AsColor(set)
	}
	ug = ug.Apply(borderColor)
	ug = ug.Apply(klimt.UStrokeWithThickness(1.5)).Apply(klimt.Bg(backcolor))
	ug.Draw(klimt.URectangleBuild(abel.EntityPositionRadius*2, abel.EntityPositionRadius*2))
	ug.CloseGroup()
}

// ShapeType returns RECTANGLE_PORT.
func (p *EntityImagePort) ShapeType() klimt.ShapeType { return klimt.ShapeTypeRectanglePort }

// CalculateDimension returns the size of the symbol of the position, as
// upstream's AbstractEntityImageBorder.
func (b *abstractEntityImageBorder) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return b.entityPosition.Dimension(b.rankdir)
}

// ShapeType returns the shape of the position, as upstream's
// AbstractEntityImageBorder.
func (b *abstractEntityImageBorder) ShapeType() klimt.ShapeType { return b.entityPosition.ShapeType() }

// EntityImageStateBorder is upstream's EntityImageStateBorder: an entry or
// exit point, a pin or an expansion node on the border of its composite
// state, with its name above or below.
type EntityImageStateBorder struct {
	abstractEntityImageBorder
}

var _ klimt.IEntityImage = (*EntityImageStateBorder)(nil)

func newEntityImageStateBorder(leaf *abel.Entity, stateParent *Cluster, bibliotekon *Bibliotekon) *EntityImageStateBorder {
	return &EntityImageStateBorder{abstractEntityImageBorder: newAbstractEntityImageBorder(leaf, stateParent,
		bibliotekon)}
}

func (e *EntityImageStateBorder) styleSignature() klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameStateDiagram, klimt.SNameState)
}

// upPosition reports whether the entity is in the upper half of its
// cluster: its name is then drawn above it.
func (e *EntityImageStateBorder) upPosition() bool {
	if e.parent == nil {
		return false
	}
	clusterCenter := e.parent.rectangleArea.PointCenter()
	node := e.bibliotekon.node(e.entity)
	return node.MinY() < clusterCenter.Y
}

// MaxWidthFromLabelForEntryExit is the width of the name, which the
// layout reserves around the entity.
func (e *EntityImageStateBorder) MaxWidthFromLabelForEntryExit(stringBounder font.StringBounder) float64 {
	return e.desc(e.style(e.styleSignature())).CalculateDimension(stringBounder).Width()
}

// DrawU draws the name and the symbol.
func (e *EntityImageStateBorder) DrawU(ug klimt.UGraphic) {
	style := e.style(e.styleSignature())
	desc := e.desc(style)
	y := 0.0
	dimDesc := desc.CalculateDimension(ug.StringBounder())
	x := 0 - (dimDesc.Width()-2*abel.EntityPositionRadius)/2
	if e.upPosition() {
		y -= 2*abel.EntityPositionRadius + dimDesc.Height()
	} else {
		y += 2 * abel.EntityPositionRadius
	}
	desc.DrawU(ug.Apply(klimt.NewUTranslate(x, y)))
	set := e.skinParam.IHtmlColorSet()
	borderColor := style.Value(klimt.PNameLineColor).AsColor(set)
	backcolor := e.entity.Colors().Color(color.TypeBack)
	if backcolor == nil {
		backcolor = style.Value(klimt.PNameBackGroundColor).AsColor(set)
	}
	if backcolor.IsTransparent() {
		backcolor = e.skinParam.BackgroundColor()
	}
	ug = ug.Apply(klimt.UStrokeWithThickness(1.5)).Apply(borderColor)
	ug = ug.Apply(klimt.Bg(backcolor))
	e.entityPosition.DrawSymbol(ug, e.rankdir)
}
