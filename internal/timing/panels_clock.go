// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/graphic/PanelsClock.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"math"

	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// panelsClock is upstream's PanelsClock: a square wave from time 0, low
// until the offset, then high for the pulse (half the period without one)
// and low for the rest of each period, up to the end of the axis.
type panelsClock struct {
	panelsBase
	period jmath.BigDecimal
	pulse  jmath.BigDecimal
	offset jmath.BigDecimal
}

func newPanelsClock(ruler *TimingRuler, skinParam klimt.ISkinParam, suggestedHeight int, style *klimt.Style,
	period, pulse, offset jmath.BigDecimal) *panelsClock {
	return &panelsClock{panelsBase: newPanelsBase(ruler, skinParam, suggestedHeight, style, nil, nil),
		period: period, pulse: pulse, offset: offset}
}

func (p *panelsClock) drawLeftPanel(klimt.UGraphic, float64) {}

func (p *panelsClock) getLeftPanelWidth(font.StringBounder) float64 { return panelsLeftPanelMinWidth }

func (p *panelsClock) getFullHeight(font.StringBounder) float64 { return float64(p.suggestedHeight) }

func (p *panelsClock) lineHeight() float64 { return float64(p.suggestedHeight) - 2*panelsMarginY }

func (p *panelsClock) yhigh() float64 { return panelsMarginY }

func (p *panelsClock) ylow() float64 { return panelsMarginY + p.lineHeight() }

// isHighAt reports whether the wave is high at a time, as it is drawn.
func (p *panelsClock) isHighAt(t float64) bool {
	periodValue := p.period.DoubleValue()
	if periodValue <= 0 {
		return false
	}
	offsetValue := p.offset.DoubleValue()
	if t < offsetValue {
		return false
	}
	vpulse := p.pulse.DoubleValue()
	if vpulse == 0 {
		vpulse = periodValue / 2.0
	}
	phase := math.Mod(t-offsetValue, periodValue)
	// Java's % can be negative for negative dividends.
	normalized := phase
	if phase < 0 {
		normalized = phase + periodValue
	}
	return normalized < vpulse
}

func (p *panelsClock) getTimeProjection(_ font.StringBounder, tick *TimeTick) *intricatedPoint {
	if tick == nil {
		return nil
	}
	x := p.ruler.getPosInPixel(tick)
	y := p.ylow()
	if p.isHighAt(tick.time.DoubleValue()) {
		y = p.yhigh()
	}
	point := geom.NewXPoint2D(x, y)
	return &intricatedPoint{point, point}
}

func (p *panelsClock) drawRightPanel(ug klimt.UGraphic) {
	ug = p.context().Apply(ug)
	vline := klimt.ULineV(p.lineHeight())
	ug = ug.Apply(klimt.UTranslateDy(panelsMarginY))
	value := 0.0
	if p.offset.DoubleValue() != 0 {
		p.drawHorizontalBetweenTimes(ug.Apply(klimt.UTranslateDy(vline.DY())), value, p.offset.DoubleValue())
		value += p.offset.DoubleValue()
	}
	if p.xOfTime(value) > p.ruler.getWidth() {
		return
	}
	drawVline(ug, p.xOfTime(value), 0, vline)
	vpulse := p.pulse.DoubleValue()
	if vpulse == 0 {
		vpulse = p.period.DoubleValue() / 2.0
	}
	remain := p.period.DoubleValue() - vpulse
	for range 1000 {
		p.drawHorizontalBetweenTimes(ug, value, value+vpulse)
		value += vpulse
		if p.xOfTime(value) > p.ruler.getWidth() {
			return
		}
		drawVline(ug, p.xOfTime(value), 0, vline)
		p.drawHorizontalBetweenTimes(ug.Apply(klimt.UTranslateDy(vline.DY())), value, value+remain)
		value += remain
		if p.xOfTime(value) > p.ruler.getWidth() {
			return
		}
		drawVline(ug, p.xOfTime(value), 0, vline)
	}
}
