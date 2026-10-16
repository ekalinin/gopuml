// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/Kal.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/plantuml-in-go/internal/abel"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// Kal is the qualifier box at one end of a qualified association
// ("A [key] --> B"): a small rectangle with the qualifier, stuck to the side
// of the entity where the link leaves it.
type Kal struct {
	textBlock klimt.TextBlock
	position  klimt.Direction
	dim       geom.XDimension2D
	translate klimt.UTranslate
	svekEdge  *SvekEdge
	entity    *abel.Entity
	link      *abel.Link
	skinParam klimt.ISkinParam
	style     *klimt.Style
}

var _ abel.Kal = (*Kal)(nil)

// translateForDecoration is getTranslateForDecoration: the move of the
// arrow head out of the box.
func (k *Kal) translateForDecoration() klimt.UTranslate {
	switch k.position {
	case klimt.DirectionRight:
		return klimt.UTranslateDx(k.dim.Width())
	case klimt.DirectionLeft:
		return klimt.UTranslateDx(-k.dim.Width())
	case klimt.DirectionDown:
		return klimt.UTranslateDy(k.dim.Height())
	case klimt.DirectionUp:
		return klimt.UTranslateDy(-k.dim.Height())
	}
	throwIllegalState("")
	return klimt.UTranslate{}
}

// newKal creates the box of text at the end of the link that touches
// entity, reserves its room in the margins of the entity and adds it to the
// entity.
func newKal(svekEdge *SvekEdge, text string, skinParam klimt.ISkinParam, entity *abel.Entity, link *abel.Link,
	stringBounder font.StringBounder) *Kal {
	k := &Kal{svekEdge: svekEdge, entity: entity, link: link, skinParam: skinParam}
	k.style = klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameClassDiagram,
		klimt.SNameClass, klimt.SNameQualified).
		WithTOBECHANGED(entity.Stereotype()).
		With(entity.Stereostyles()).
		MergedStyle(skinParam.CurrentStyleBuilder())
	fc := k.style.FontConfiguration(skinParam.IHtmlColorSet(), nil)
	k.textBlock = klimt.DisplayWithNewlines(skinParam.Pragma(), text).Create7(fc, geom.HorizontalAlignmentLeft,
		skinParam, klimt.CreoleModeSimpleLine)
	k.dim = k.textBlock.CalculateDimension(stringBounder).DeltaWH(4, 2)
	switch {
	case link.Length() == 1 && link.Entity1() == entity:
		k.position = klimt.DirectionRight
		entity.EnsureMargins(klimt.NewMargins(0, k.dim.Width(), 0, 0))
	case link.Length() == 1 && link.Entity2() == entity:
		k.position = klimt.DirectionLeft
		entity.EnsureMargins(klimt.NewMargins(k.dim.Width(), 0, 0, 0))
	case link.Entity1() == entity:
		k.position = klimt.DirectionDown
		entity.EnsureMargins(klimt.NewMargins(0, 0, 0, k.dim.Height()))
	case link.Entity2() == entity:
		k.position = klimt.DirectionUp
		entity.EnsureMargins(klimt.NewMargins(0, 0, k.dim.Height(), 0))
	default:
		throwIllegalState("")
	}
	entity.AddKal(k)
	return k
}

// Dimension returns the size of the box.
func (k *Kal) Dimension() geom.XDimension2D { return k.dim }

// DrawU draws the box with its text.
func (k *Kal) DrawU(ug klimt.UGraphic) {
	rect := klimt.URectangleBuildDim(k.dim)
	classBackground := k.style.Value(klimt.PNameBackGroundColor).AsColor(k.skinParam.IHtmlColorSet())
	classBorder := k.style.Value(klimt.PNameLineColor).AsColor(k.skinParam.IHtmlColorSet())
	ug = ug.Apply(k.getTranslate())
	ug.Apply(klimt.Bg(classBackground)).Apply(classBorder).Apply(klimt.UStrokeWithThickness(0.5)).Draw(rect)
	k.textBlock.DrawU(ug.Apply(klimt.NewUTranslate(2, 1)))
}

func (k *Kal) getTranslate() klimt.UTranslate { return k.textDelta().Compose(k.translate) }

func (k *Kal) x1() float64 { return k.getTranslate().Dx() - 5 }

func (k *Kal) x2() float64 { return k.x1() + k.dim.Width() + 10 }

func (k *Kal) textDelta() klimt.UTranslate {
	switch k.position {
	case klimt.DirectionRight:
		return klimt.UTranslateDy(-k.dim.Height() / 2)
	case klimt.DirectionLeft:
		return klimt.NewUTranslate(-k.dim.Width()+0.5, -k.dim.Height()/2)
	case klimt.DirectionDown:
		return klimt.UTranslateDx(-k.dim.Width() / 2)
	case klimt.DirectionUp:
		return klimt.NewUTranslate(-k.dim.Width()/2, -k.dim.Height()+0.5)
	}
	throwIllegalState("")
	return klimt.UTranslate{}
}

// Position returns the side of the entity where the box is.
func (k *Kal) Position() klimt.Direction { return k.position }

// setTranslate places the box at the end of its link.
func (k *Kal) setTranslate(translate klimt.UTranslate) { k.translate = translate }

// moveX moves the box horizontally (and the start of its link when the link
// starts at the entity).
func (k *Kal) moveX(dx float64) {
	if dx == 0 {
		return
	}
	k.translate = k.translate.Compose(klimt.UTranslateDx(dx))
	if k.link.Entity1() == k.entity {
		k.svekEdge.moveStartPoint(dx, 0)
	}
}
