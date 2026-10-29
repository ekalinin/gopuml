// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/graphic/PanelsAnalog.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"math"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// panelsAnalog is upstream's PanelsAnalog: a line through the values, with
// the bounds or the ticks on the left. It draws no notes.
type panelsAnalog struct {
	panelsBase
	timeSeries   *timeSeries
	initialState *float64
	ticksEvery   *int
}

func newPanelsAnalog(ruler *TimingRuler, skinParam klimt.ISkinParam, suggestedHeight int, style *klimt.Style,
	timeSeries *timeSeries, constraints []*timeConstraint, initialState *float64, ticksEvery *int) *panelsAnalog {
	return &panelsAnalog{panelsBase: newPanelsBase(ruler, skinParam, suggestedHeight, style, nil, constraints),
		timeSeries: timeSeries, initialState: initialState, ticksEvery: ticksEvery}
}

func (p *panelsAnalog) getFullHeight(stringBounder font.StringBounder) float64 {
	return p.heightForConstraints(stringBounder) + float64(p.suggestedHeight)
}

func (p *panelsAnalog) getTimeProjection(stringBounder font.StringBounder, tick *TimeTick) *intricatedPoint {
	if tick == nil {
		return nil
	}
	x := p.ruler.getPosInPixel(tick)
	value := p.timeSeries.valueAt(tick)
	return &intricatedPoint{geom.NewXPoint2D(x, p.ypos(stringBounder, value)),
		geom.NewXPoint2D(x, p.ypos(stringBounder, value))}
}

func (p *panelsAnalog) ypos(stringBounder font.StringBounder, value float64) float64 {
	y := (value - p.timeSeries.getMin()) * (float64(p.suggestedHeight) - 2*panelsMarginY) /
		(p.timeSeries.getMax() - p.timeSeries.getMin())
	return p.heightForConstraints(stringBounder) + float64(p.suggestedHeight) - panelsMarginY - y
}

// tickRange is the integers between the bounds, Java's (int) casts of
// their ceiling and floor.
func (p *panelsAnalog) tickRange() (first, last int32) {
	return int32(klimt.JavaD2I(math.Ceil(p.timeSeries.getMin()))),
		int32(klimt.JavaD2I(math.Floor(p.timeSeries.getMax())))
}

// isTick reports whether i is a multiple of ticksEvery; 0 throws upstream's
// ArithmeticException.
func (p *panelsAnalog) isTick(i int32) bool {
	if *p.ticksEvery == 0 {
		uml.Throw(uml.ArithmeticException, "/ by zero")
	}
	return i%int32(*p.ticksEvery) == 0
}

func (p *panelsAnalog) drawLeftPanel(ug klimt.UGraphic, fullAvailableWidth float64) {
	if p.ticksEvery == nil {
		p.drawScaleLabel(ug, p.timeSeries.getMin(), fullAvailableWidth)
		p.drawScaleLabel(ug, p.timeSeries.getMax(), fullAvailableWidth)
		return
	}
	first, last := p.tickRange()
	for i := first; i <= last; i++ {
		if p.isTick(i) {
			p.drawScaleLabel(ug, float64(i), fullAvailableWidth)
		}
	}
}

func (p *panelsAnalog) getLeftPanelWidth(stringBounder font.StringBounder) float64 {
	return panelsLeftPanelMinWidth + p.maxWidthForTicks(stringBounder)
}

func (p *panelsAnalog) maxWidthForTicks(stringBounder font.StringBounder) float64 {
	if p.ticksEvery == nil {
		return max(p.widthLabel(stringBounder, p.timeSeries.getMin()),
			p.widthLabel(stringBounder, p.timeSeries.getMax()))
	}
	result := 0.0
	first, last := p.tickRange()
	for i := first; i <= last; i++ {
		if p.isTick(i) {
			result = max(result, p.widthLabel(stringBounder, float64(i)))
		}
	}
	return result
}

func (p *panelsAnalog) widthLabel(stringBounder font.StringBounder, value float64) float64 {
	return p.valueTextBlock(value).CalculateDimension(stringBounder).Width()
}

func (p *panelsAnalog) drawScaleLabel(ug klimt.UGraphic, value, fullAvailableWidth float64) {
	label := p.valueTextBlock(value)
	dim := label.CalculateDimension(ug.StringBounder())
	ug = ug.Apply(klimt.UTranslateDx(fullAvailableWidth - dim.Width() - 2))
	label.DrawU(ug.Apply(klimt.UTranslateDy(p.ypos(ug.StringBounder(), value) - dim.Height()/2)))
}

func (p *panelsAnalog) valueTextBlock(value float64) klimt.TextBlock {
	formattedValue := p.timeSeries.displayValue(value)
	display := klimt.DisplayWithNewlines(p.skinParam.Pragma(), formattedValue)
	return display.Create(p.fontConfiguration(), geom.HorizontalAlignmentLeft, p.skinParam)
}

func (p *panelsAnalog) drawTickHlines(ug klimt.UGraphic) {
	ug = applyForVLines(ug, p.style, p.skinParam)
	first, last := p.tickRange()
	hline := klimt.ULineH(p.ruler.getWidth())
	for i := first; i <= last; i++ {
		if p.isTick(i) {
			ug.Apply(klimt.UTranslateDy(p.ypos(ug.StringBounder(), float64(i)))).Draw(hline)
		}
	}
}

func (p *panelsAnalog) drawRightPanel(ug klimt.UGraphic) {
	if p.ticksEvery != nil {
		p.drawTickHlines(ug)
	}
	ug = p.context().Apply(ug)
	lastx := 0.0
	lastValue := 0.0
	if p.initialState != nil {
		lastValue = *p.initialState
	}
	for _, ent := range p.timeSeries.values {
		y1 := p.ypos(ug.StringBounder(), lastValue)
		y2 := p.ypos(ug.StringBounder(), ent.value)
		x := p.ruler.getPosInPixel(ent.tick)
		ug.Apply(klimt.NewUTranslate(lastx, y1)).Draw(klimt.NewULine(x-lastx, y2-y1))
		lastx = x
		lastValue = ent.value
	}
	ug.Apply(klimt.NewUTranslate(lastx, p.ypos(ug.StringBounder(), lastValue))).
		Draw(klimt.ULineH(p.ruler.getWidth() - lastx))
	p.drawConstraints(ug.Apply(klimt.UTranslateDy(p.heightForConstraints(ug.StringBounder()))))
}
