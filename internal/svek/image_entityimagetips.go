// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/image/EntityImageTips.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"math"
	"strings"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

const tipsYSpacing = 10

// EntityImageTips is the image of the notes on members of a class ("note
// left of A::field"): one note per member, pointing at the member.
type EntityImageTips struct {
	abstractEntityImage
	memo        klimt.TextBlockMemoized
	bibliotekon *Bibliotekon
}

var _ klimt.IEntityImage = (*EntityImageTips)(nil)

func newEntityImageTips(entity *abel.Entity, bibliotekon *Bibliotekon) *EntityImageTips {
	return &EntityImageTips{abstractEntityImage: newAbstractEntityImage(entity), bibliotekon: bibliotekon}
}

// styleFor is the style of one tip: each tip keeps its own stereotype
// (upstream issue 2814).
func (t *EntityImageTips) styleFor(tip *abel.Tip) *klimt.Style {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, t.styleName(), klimt.SNameNote).
		WithTOBECHANGED(tip.Stereotype()).MergedStyle(t.skinParam.CurrentStyleBuilder())
}

func (t *EntityImageTips) noteBackgroundColor(tip *abel.Tip, style *klimt.Style) color.HColor {
	if fromTip := tip.Colors().Color(color.TypeBack); fromTip != nil {
		return fromTip
	}
	return style.Value(klimt.PNameBackGroundColor).AsColor(t.skinParam.IHtmlColorSet())
}

func (t *EntityImageTips) position() abel.Position {
	if strings.HasSuffix(t.entity.Name(), abel.PositionRight.String()) {
		return abel.PositionRight
	}
	return abel.PositionLeft
}

// ShapeType returns RECTANGLE.
func (t *EntityImageTips) ShapeType() klimt.ShapeType { return klimt.ShapeTypeRectangle }

// CalculateDimension returns the size of the notes, stacked.
func (t *EntityImageTips) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return t.memo.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		width, height := 0.0, 0.0
		members, tips := t.entity.Tips()
		for _, member := range members {
			dim := t.opale(tips[member]).CalculateDimension(sb)
			height += dim.Height()
			height += tipsYSpacing
			width = math.Max(width, dim.Width())
		}
		return geom.NewXDimension2D(width, height)
	})
}

// DrawU draws the notes, each pointing at its member.
func (t *EntityImageTips) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	other := t.bibliotekon.onlyOther(t.entity)
	if other == nil {
		// Upstream prints "Error1 in EntityImageTips" on standard error.
		return
	}
	nodeMe := t.bibliotekon.node(t.entity)
	nodeOther := t.bibliotekon.node(other)
	if nodeOther == nil {
		// Upstream prints "Error2 in EntityImageTips" on standard error.
		return
	}
	positionOther := nodeOther.Position()
	direction := t.position().ReverseDirection()
	positionMe := nodeMe.Position()
	height := 0.0
	members, tips := t.entity.Tips()
	for _, member := range members {
		tip := tips[member]
		bestMatch, ok := nodeOther.leaf.BestMatch(member)
		if !ok {
			return
		}
		memberPosition := nodeOther.image.InnerPosition(bestMatch, stringBounder)
		if memberPosition == nil {
			return
		}
		opale := t.opale(tip)
		dim := opale.CalculateDimension(stringBounder)
		pp1 := geom.NewXPoint2D(0, dim.Height()/2)
		x := positionOther.X - positionMe.X
		if direction == klimt.DirectionRight && x < 0 {
			direction = direction.Inv()
		}
		if direction == klimt.DirectionLeft {
			x += memberPosition.MaxX()
		} else {
			x += memberPosition.MinX()
		}
		y := positionOther.Y - positionMe.Y - height + memberPosition.CenterY()
		pp2 := geom.NewXPoint2D(x, y)
		opale.SetOpale(direction, pp1, pp2)
		opale.DrawU(klimt.UGraphicStencilCreateDim(ug, dim))
		ug = ug.Apply(klimt.UTranslateDy(dim.Height() + tipsYSpacing))
		height += dim.Height()
		height += tipsYSpacing
	}
}

func (t *EntityImageTips) opale(tip *abel.Tip) *klimt.Opale {
	style := t.styleFor(tip)
	noteBackgroundColor := t.noteBackgroundColor(tip, style)
	borderColor := style.Value(klimt.PNameLineColor).AsColor(t.skinParam.IHtmlColorSet())
	shadowing := style.Shadowing()
	fc := style.FontConfiguration(t.skinParam.IHtmlColorSet(), nil)
	stroke := style.Stroke()
	textBlock := klimt.BodyFactoryCreate3(tip.Display(), t.skinParam, geom.HorizontalAlignmentLeft, fc,
		style.WrapWidth(), style)
	return klimt.NewOpale(shadowing, borderColor, noteBackgroundColor, textBlock, true, &stroke)
}
