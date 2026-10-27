// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/mindmap/Finger.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/mindmap/FingerImpl.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package mindmap

import (
	"github.com/ekalinin/gopuml/internal/activitydiagram3"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// fingerImpl draws a node (its phalanx) and its children (its nail), packed
// next to it. Upstream's Finger interface has fingerImpl as its only
// implementation, which Branch.getX12 casts it to, so Go keeps the struct
// alone. The thickness is across the direction of the drawing, the
// elongation along it.
type fingerImpl struct {
	idea      *idea
	skinParam klimt.ISkinParam
	// direction is 1 for the right (or lower) side, -1 for the other one.
	direction   float64
	drawPhalanx bool
	nail        []*fingerImpl
	tetris      *tetris
}

// buildFinger returns the finger of idea and of its descendants.
func buildFinger(idea *idea, skinParam klimt.ISkinParam, direction bool) *fingerImpl {
	result := &fingerImpl{idea: idea, skinParam: skinParam, direction: -1, drawPhalanx: true}
	if direction {
		result.direction = 1
	}
	for _, child := range idea.children {
		result.nail = append(result.nail, buildFinger(child, skinParam, direction))
	}
	return result
}

func (f *fingerImpl) isTopToBottom() bool { return f.skinParam.Rankdir() == geom.RankdirTopToBottom }

func (f *fingerImpl) margin() klimt.ClockwiseTopRightBottomLeft { return f.style().Margin() }

// DrawU draws the phalanx at the origin, then each child at its place in
// the nail, with a link from the end of the phalanx to the child.
func (f *fingerImpl) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	phalanx := f.phalanx()
	dimPhalanx := phalanx.CalculateDimension(stringBounder)
	if f.drawPhalanx {
		var posX, posY float64
		if f.isTopToBottom() {
			posX = -f.phalanxThickness(stringBounder) / 2
			if f.direction == 1 {
				posY = 0
			} else {
				posY = -dimPhalanx.Height()
			}
		} else {
			if f.direction == 1 {
				posX = 0
			} else {
				posX = -dimPhalanx.Width()
			}
			posY = -f.phalanxThickness(stringBounder) / 2
		}
		phalanx.DrawU(ug.Apply(klimt.NewUTranslate(posX, posY)))
	}
	var p1 geom.XPoint2D
	if f.isTopToBottom() {
		p1 = geom.XPoint2D{X: 0, Y: f.direction * dimPhalanx.Height()}
	} else {
		p1 = geom.XPoint2D{X: f.direction * dimPhalanx.Width(), Y: 0}
	}
	for i, child := range f.nail {
		stp := f.getTetris(stringBounder).elements[i]
		var p2 geom.XPoint2D
		if f.isTopToBottom() {
			p2 = geom.XPoint2D{X: stp.y, Y: f.direction * (dimPhalanx.Height() + f.x12())}
		} else {
			p2 = geom.XPoint2D{X: f.direction * (dimPhalanx.Width() + f.x12()), Y: stp.y}
		}
		child.DrawU(ug.Apply(klimt.UTranslatePoint(p2)))
		linkColor := f.linkColor()
		if !linkColor.IsTransparent() {
			f.drawLine(ug.Apply(linkColor).Apply(f.uStroke()), p1, p2)
		}
	}
}

func (f *fingerImpl) linkColor() color.HColor {
	return f.idea.styleArrow().Value(klimt.PNameLineColor).AsColor(f.skinParam.IHtmlColorSet())
}

func (f *fingerImpl) uStroke() klimt.UStroke { return f.idea.styleArrow().Stroke() }

// drawLine draws the link: a short line, a curve, and a short line.
func (f *fingerImpl) drawLine(ug klimt.UGraphic, p1, p2 geom.XPoint2D) {
	path := klimt.UPathNone()
	path.MoveToPoint(p1)
	if f.isTopToBottom() {
		delta1 := f.direction * 3
		delta2 := f.direction * 10
		path.LineTo(p1.X, p1.Y+delta1)
		path.CubicTo(p1.X, p1.Y+delta2, p2.X, p2.Y-delta2, p2.X, p2.Y-delta1)
	} else {
		delta1 := f.direction * 10
		delta2 := f.direction * 25
		path.LineTo(p1.X+delta1, p1.Y)
		path.CubicTo(p1.X+delta2, p1.Y, p2.X-delta2, p2.Y, p2.X-delta1, p2.Y)
	}
	path.LineToPoint(p2)
	ug.Draw(path)
}

// getTetris packs the children, once, with the string bounder of the first
// call.
func (f *fingerImpl) getTetris(stringBounder font.StringBounder) *tetris {
	if f.tetris == nil {
		f.tetris = newTetris()
		for _, child := range f.nail {
			f.tetris.add(child.asSymetricalTee(stringBounder))
		}
		f.tetris.balance()
	}
	return f.tetris
}

func (f *fingerImpl) asSymetricalTee(stringBounder font.StringBounder) *symetricalTee {
	thickness1 := f.phalanxThickness(stringBounder)
	elongation1 := f.phalanxElongation(stringBounder)
	if len(f.nail) == 0 {
		return newSymetricalTee(thickness1, elongation1, 0, 0)
	}
	thickness2 := f.nailThickness(stringBounder)
	elongation2 := f.nailElongation(stringBounder)
	return newSymetricalTee(thickness1, elongation1+f.x1(), thickness2, f.x2()+elongation2)
}

// x1 is the space before the links, x2 the space after them.
func (f *fingerImpl) x1() float64 {
	if f.isTopToBottom() {
		return f.margin().Top()
	}
	return f.margin().Left()
}

func (f *fingerImpl) x2() float64 {
	if f.isTopToBottom() {
		return f.margin().Bottom() + 5
	}
	return f.margin().Right() + 30
}

func (f *fingerImpl) x12() float64 { return f.x1() + f.x2() }

func (f *fingerImpl) phalanxThickness(stringBounder font.StringBounder) float64 {
	if f.isTopToBottom() {
		return f.phalanx().CalculateDimension(stringBounder).Width()
	}
	return f.phalanx().CalculateDimension(stringBounder).Height()
}

func (f *fingerImpl) phalanxElongation(stringBounder font.StringBounder) float64 {
	if f.isTopToBottom() {
		return f.phalanx().CalculateDimension(stringBounder).Height()
	}
	return f.phalanx().CalculateDimension(stringBounder).Width()
}

// phalanx is the drawing of the node: its box with the margins across the
// direction of the drawing, or its text with small margins.
func (f *fingerImpl) phalanx() klimt.TextBlock {
	if !f.drawPhalanx {
		return klimt.TextBlockUtilsEmpty(0, 0)
	}
	style := f.style()
	if f.idea.shape == IdeaShapeBox {
		foo := klimt.NewSkinParamColors(f.skinParam, color.EmptyColors().Add(color.TypeBack, f.idea.backColor))
		box := activitydiagram3.FtileBoxOldCreateMindMap(style, foo, f.idea.label)
		margin := f.margin()
		if f.isTopToBottom() {
			return klimt.TextBlockUtilsWithMargin4(box, margin.Left(), margin.Right(), 0, 0)
		}
		return klimt.TextBlockUtilsWithMargin4(box, 0, 0, margin.Top(), margin.Bottom())
	}
	text := f.idea.label.Create0(style.FontConfiguration(f.skinParam.IHtmlColorSet(), nil),
		style.HorizontalAlignment(), f.skinParam, style.WrapWidth(), klimt.CreoleModeFull, nil, nil)
	if f.direction == 1 {
		return klimt.TextBlockUtilsWithMargin4(text, 3, 0, 1, 1)
	}
	return klimt.TextBlockUtilsWithMargin4(text, 0, 3, 1, 1)
}

func (f *fingerImpl) style() *klimt.Style {
	if len(f.nail) != len(f.idea.children) {
		uml.Throw(uml.IllegalStateException, "")
	}
	return f.idea.style()
}

func (f *fingerImpl) nailThickness(stringBounder font.StringBounder) float64 {
	return f.getTetris(stringBounder).height()
}

func (f *fingerImpl) nailElongation(stringBounder font.StringBounder) float64 {
	return f.getTetris(stringBounder).width()
}

func (f *fingerImpl) fullThickness(stringBounder font.StringBounder) float64 {
	return max(f.phalanxThickness(stringBounder), f.nailThickness(stringBounder))
}

func (f *fingerImpl) fullElongation(stringBounder font.StringBounder) float64 {
	return f.phalanxElongation(stringBounder) + f.nailElongation(stringBounder)
}

func (f *fingerImpl) doNotDrawFirstPhalanx() { f.drawPhalanx = false }
