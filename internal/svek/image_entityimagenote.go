// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/EntityImageNote.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"math"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

const (
	noteMarginX1 = 6
	noteMarginX2 = 15
	noteMarginY  = 5
)

// EntityImageNote is a note entity: the folded-corner shape, or, when it
// is attached by a single link (opale), the shape with a pointer to the
// other entity instead of the link.
type EntityImageNote struct {
	abstractEntityImage
	memo klimt.TextBlockMemoized

	noteBackgroundColor color.HColor
	borderColor         color.HColor
	shadowing           float64
	style               *klimt.Style
	textBlock           klimt.TextBlock

	opaleLine *SvekEdge
	node      *SvekNode
	other     *SvekNode
}

var (
	_ klimt.IEntityImage = (*EntityImageNote)(nil)
	_ klimt.Stencil      = (*EntityImageNote)(nil)
)

func newEntityImageNote(entity *abel.Entity) *EntityImageNote {
	n := &EntityImageNote{abstractEntityImage: newAbstractEntityImage(entity)}
	strings := entity.Display()
	n.style = n.styleSignature().MergedStyle(n.skinParam.CurrentStyleBuilder())
	if back := entity.Colors().Color(color.TypeBack); back == nil {
		n.noteBackgroundColor = n.style.Value(klimt.PNameBackGroundColor).AsColor(n.skinParam.IHtmlColorSet())
	} else {
		n.noteBackgroundColor = back
	}
	n.borderColor = n.style.Value(klimt.PNameLineColor).AsColor(n.skinParam.IHtmlColorSet())
	n.shadowing = n.style.Shadowing()
	fontConfiguration := n.style.FontConfiguration(n.skinParam.IHtmlColorSet(), nil)
	horizontalAlignment := n.style.HorizontalAlignment()
	if strings.Size() == 1 && javaLength(strings.Get(0).String()) == 0 {
		n.textBlock = klimt.NewTextBlockEmpty(0, 0)
	} else {
		n.textBlock = klimt.BodyFactoryCreate3(strings, n.skinParam, horizontalAlignment, fontConfiguration,
			n.style.WrapWidth(), n.style)
	}
	return n
}

func javaLength(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x10000 {
			n += 2
		} else {
			n++
		}
	}
	return n
}

func (n *EntityImageNote) styleSignature() klimt.StyleSignature {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, n.styleName(), klimt.SNameNote).
		WithTOBECHANGED(n.stereo())
}

func (n *EntityImageNote) textHeight(stringBounder font.StringBounder) float64 {
	return n.textBlock.CalculateDimension(stringBounder).Height() + 2*noteMarginY
}

func (n *EntityImageNote) textWidth(stringBounder font.StringBounder) float64 {
	return n.textBlock.CalculateDimension(stringBounder).Width() + noteMarginX1 + noteMarginX2
}

// CalculateDimension returns the size of the note.
func (n *EntityImageNote) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return n.memo.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		return geom.NewXDimension2D(n.textWidth(sb), n.textHeight(sb))
	})
}

// DrawU draws the note.
func (n *EntityImageNote) DrawU(ug klimt.UGraphic) {
	url := n.entity.URL99()
	ug.StartGroup(entityGroup(n.entity))
	if url != nil {
		ug.StartURL(url)
	}
	ug2 := klimt.UGraphicStencilCreate(ug, n, klimt.UStrokeSimple())
	if n.opaleLine == nil || !n.opaleLine.opale {
		n.drawNormal(ug2)
	} else {
		stringBounder := ug.StringBounder()
		path := n.opaleLine.DotPath()

		force1 := n.MagneticBorder().ForceAt(stringBounder, path.StartPoint())
		force2 := n.other.magneticBorder().ForceAt(stringBounder, path.EndPoint())

		path.MoveDelta(-n.node.minX, -n.node.minY)

		textWidth := n.textWidth(stringBounder)
		textHeight := n.textHeight(stringBounder)
		center := geom.NewXPoint2D(textWidth/2, textHeight/2)
		if path.StartPoint().Distance(center) > path.EndPoint().Distance(center) {
			path = path.Reverse()
		}
		strategy, ok := n.opaleStrategy(textWidth, textHeight, path.StartPoint())
		pp1 := force1.Translated(path.StartPoint())
		pp2 := force2.Translated(path.EndPoint())

		stroke := n.style.Stroke()
		opale := klimt.NewOpale(n.shadowing, n.borderColor, n.noteBackgroundColor, n.textBlock, true, &stroke)
		opale.SetRoundCorner(n.roundCorner())
		if ok {
			opale.SetOpale(strategy, pp1, pp2)
		}
		stroked := n.applyStroke(ug2)
		opale.DrawU(klimt.ColorsApplyStroke(stroked, n.entity.Colors()))
	}
	if url != nil {
		ug.CloseURL()
	}
	ug.CloseGroup()
}

func (n *EntityImageNote) roundCorner() float64 {
	return n.skinParam.RoundCorner(klimt.CornerParamDefault, nil)
}

func (n *EntityImageNote) drawNormal(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	polygon := klimt.OpalePolygonNormal(n.textWidth(stringBounder), n.textHeight(stringBounder), n.roundCorner())
	polygon.SetDeltaShadow(n.shadowing)
	ug = ug.Apply(klimt.Bg(n.noteBackgroundColor)).Apply(n.borderColor)
	stroked := n.applyStroke(ug)
	stroked.Draw(polygon)
	ug.Draw(klimt.OpaleCorner(n.textWidth(stringBounder), n.roundCorner()))
	n.textBlock.DrawU(ug.Apply(klimt.NewUTranslate(noteMarginX1, noteMarginY)))
}

func (n *EntityImageNote) applyStroke(ug klimt.UGraphic) klimt.UGraphic {
	return ug.Apply(n.style.Stroke())
}

func (n *EntityImageNote) opaleStrategy(width, height float64, pt geom.XPoint2D) (klimt.Direction, bool) {
	d1 := orthoDistance(geom.NewXLine2D(width, 0, width, height), pt)
	d2 := orthoDistance(geom.NewXLine2D(0, height, width, height), pt)
	d3 := orthoDistance(geom.NewXLine2D(0, 0, 0, height), pt)
	d4 := orthoDistance(geom.NewXLine2D(0, 0, width, 0), pt)
	switch {
	case d3 <= d1 && d3 <= d2 && d3 <= d4:
		return klimt.DirectionLeft, true
	case d1 <= d2 && d1 <= d3 && d1 <= d4:
		return klimt.DirectionRight, true
	case d4 <= d1 && d4 <= d2 && d4 <= d3:
		return klimt.DirectionUp, true
	case d2 <= d1 && d2 <= d3 && d2 <= d4:
		return klimt.DirectionDown, true
	}
	return 0, false
}

func orthoDistance(seg geom.XLine2D, pt geom.XPoint2D) float64 {
	switch {
	case seg.Y1 == seg.Y2:
		return math.Abs(seg.Y1 - pt.Y)
	case seg.X1 == seg.X2:
		return math.Abs(seg.X1 - pt.X)
	}
	throwIllegalArgument("")
	return 0
}

// ShapeType returns RECTANGLE.
func (n *EntityImageNote) ShapeType() klimt.ShapeType { return klimt.ShapeTypeRectangle }

func (n *EntityImageNote) setOpaleLine(line *SvekEdge, node, other *SvekNode) {
	if other == nil {
		throwNPE()
	}
	n.opaleLine = line
	n.node = node
	n.other = other
}

// StartingX implements klimt.Stencil.
func (n *EntityImageNote) StartingX(font.StringBounder, float64) float64 { return 0 }

// EndingX implements klimt.Stencil.
func (n *EntityImageNote) EndingX(stringBounder font.StringBounder, _ float64) float64 {
	return n.CalculateDimension(stringBounder).Width()
}
