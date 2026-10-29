// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/TimeProjected.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/graphic/Panels.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/graphic/PanelsNoLeft.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// panels draws the lane of a player: the left panel (the labels of its
// levels), the right panel (its states along the time axis), and the
// points of its lane at a time, upstream's TimeProjected.
type panels interface {
	drawLeftPanel(ug klimt.UGraphic, fullAvailableWidth float64)
	drawRightPanel(ug klimt.UGraphic)
	getFullHeight(stringBounder font.StringBounder) float64
	getLeftPanelWidth(stringBounder font.StringBounder) float64
	// getTimeProjection is nil for a nil tick and where the lane has no
	// point at the time.
	getTimeProjection(stringBounder font.StringBounder, tick *TimeTick) *intricatedPoint
}

// The layout constants of the panels.
const (
	panelsMarginY           = 8.0
	panelsMarginX           = 12.0
	panelsBottomMargin      = 10.0
	panelsLeftPanelMinWidth = 5.0
)

// panelsBase is the part of the panels that upstream's Panels holds.
type panelsBase struct {
	skinParam       klimt.ISkinParam
	ruler           *TimingRuler
	suggestedHeight int
	style           *klimt.Style
	constraints     []*timeConstraint
	notes           []*timingNote
	// constraintDeltaY is upstream's getConstraintDeltaY, 0 for every
	// constraint when nil.
	constraintDeltaY func(constraint *timeConstraint) float64
}

func newPanelsBase(ruler *TimingRuler, skinParam klimt.ISkinParam, suggestedHeight int, style *klimt.Style,
	notes []*timingNote, constraints []*timeConstraint) panelsBase {
	return panelsBase{ruler: ruler, skinParam: skinParam, suggestedHeight: suggestedHeight, style: style,
		notes: notes, constraints: constraints}
}

func (p *panelsBase) deltaY(constraint *timeConstraint) float64 {
	if p.constraintDeltaY == nil {
		return 0
	}
	return p.constraintDeltaY(constraint)
}

func (p *panelsBase) drawConstraints(ug klimt.UGraphic) {
	for _, constraint := range p.constraints {
		constraint.drawU(ug.Apply(klimt.UTranslateDy(p.deltaY(constraint))), p.ruler)
	}
}

// heightForConstraints is the height of the labels of the constraints
// above the lane.
func (p *panelsBase) heightForConstraints(stringBounder font.StringBounder) float64 {
	if len(p.constraints) == 0 {
		return 0
	}
	result := 0.0
	for _, constraint := range p.constraints {
		result = max(result, constraint.constraintHeight(stringBounder)-p.deltaY(constraint))
	}
	return result
}

// drawNotes draws the notes at position, at their times, or at 0 for a
// note given before any time.
func (p *panelsBase) drawNotes(ug klimt.UGraphic, position abel.Position) {
	for _, note := range p.notes {
		if note.position == position {
			x := 0.0
			if note.when != nil {
				x = p.ruler.getPosInPixel(note.when)
			}
			note.drawU(ug.Apply(klimt.UTranslateDx(x)))
		}
	}
}

func (p *panelsBase) heightForNotes(stringBounder font.StringBounder, position abel.Position) float64 {
	height := 0.0
	for _, note := range p.notes {
		if note.position == position {
			height = max(height, note.height(stringBounder))
		}
	}
	return height
}

// context is the fashion of the lane: the colors and the stroke of the
// style.
func (p *panelsBase) context() *klimt.Fashion {
	lineColor := p.style.Value(klimt.PNameLineColor).AsColor(p.skinParam.IHtmlColorSet())
	backgroundColor := p.style.Value(klimt.PNameBackGroundColor).AsColor(p.skinParam.IHtmlColorSet())
	return klimt.NewFashion(backgroundColor, lineColor).WithStroke(p.stroke())
}

func (p *panelsBase) stroke() klimt.UStroke { return p.style.Stroke() }

func (p *panelsBase) fontConfiguration() *klimt.FontConfiguration {
	return klimt.FontConfigurationFromStyle(p.skinParam, p.style, nil)
}

func (p *panelsBase) createTextBlock(value string) klimt.TextBlock {
	display := klimt.DisplayWithNewlines(p.skinParam.Pragma(), value)
	return display.Create(p.fontConfiguration(), geom.HorizontalAlignmentLeft, p.skinParam)
}

func drawHLineFromPoint(ug klimt.UGraphic, start geom.XPoint2D, length float64) {
	ug.Apply(klimt.UTranslatePoint(start)).Draw(klimt.ULineH(length))
}

// drawHorizontalBetweenTimes draws a line between two times, up to the end
// of the axis.
func (p *panelsBase) drawHorizontalBetweenTimes(ug klimt.UGraphic, startTime, endTime float64) {
	x1 := p.xOfTime(startTime)
	x2 := min(p.ruler.getWidth(), p.xOfTime(endTime))
	hline := klimt.ULineH(x2 - x1)
	ug.Apply(klimt.UTranslateDx(x1)).Draw(hline)
}

// xOfTime is the x of a time given as a double, through upstream's exact
// new BigDecimal(double).
func (p *panelsBase) xOfTime(t float64) float64 {
	return p.ruler.getPosInPixel(newTimeTick(jmath.BigDecimalOfFloat64(t), formatDecimal))
}

func drawVline(ug klimt.UGraphic, x, y float64, vline *klimt.ULine) {
	ug.Apply(klimt.NewUTranslate(x, y)).Draw(vline)
}
