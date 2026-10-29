// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/graphic/PanelsBinary.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"slices"
	"strings"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// panelsBinary is upstream's PanelsBinary: a line at the high or the low
// level of the lane, with hatching for two values at once.
type panelsBinary struct {
	panelsBase
	values       []binaryValue
	initialState *changeState
}

func newPanelsBinary(ruler *TimingRuler, skinParam klimt.ISkinParam, suggestedHeight int, style *klimt.Style,
	values []binaryValue, constraints []*timeConstraint, initialState *changeState,
	notes []*timingNote) *panelsBinary {
	return &panelsBinary{panelsBase: newPanelsBase(ruler, skinParam, suggestedHeight, style, notes, constraints),
		values: values, initialState: initialState}
}

func (p *panelsBinary) drawLeftPanel(klimt.UGraphic, float64) {}

func (p *panelsBinary) getLeftPanelWidth(font.StringBounder) float64 { return panelsLeftPanelMinWidth }

func (p *panelsBinary) getFullHeight(stringBounder font.StringBounder) float64 {
	return p.heightForConstraints(stringBounder) + p.heightForNotes(stringBounder, abel.PositionTop) +
		float64(p.suggestedHeight) + p.heightForNotes(stringBounder, abel.PositionBottom)
}

// getTimeProjection is always at the high level, as upstream does.
func (p *panelsBinary) getTimeProjection(stringBounder font.StringBounder, tick *TimeTick) *intricatedPoint {
	if tick == nil {
		return nil
	}
	x := p.ruler.getPosInPixel(tick)
	return &intricatedPoint{geom.NewXPoint2D(x, p.ypos(stringBounder, highString)),
		geom.NewXPoint2D(x, p.ypos(stringBounder, highString))}
}

func (p *panelsBinary) ypos(stringBounder font.StringBounder, state string) float64 {
	if strings.EqualFold(state, lowString) {
		return p.ylow(stringBounder)
	}
	return p.yhigh(stringBounder)
}

func (p *panelsBinary) yhigh(stringBounder font.StringBounder) float64 {
	return panelsMarginY + p.heightForConstraints(stringBounder) + p.heightForNotes(stringBounder, abel.PositionTop)
}

func (p *panelsBinary) ylow(stringBounder font.StringBounder) float64 {
	return p.heightForConstraints(stringBounder) + p.heightForNotes(stringBounder, abel.PositionTop) +
		float64(p.suggestedHeight) - panelsMarginY
}

func (p *panelsBinary) drawRightPanel(ug klimt.UGraphic) {
	ug = p.context().Apply(ug)
	lastx := 0.0
	lastValues := []string{lowString}
	if p.initialState != nil {
		lastValues = p.initialState.states
	}
	stringBounder := ug.StringBounder()
	yhigh := p.yhigh(stringBounder)
	ylow := p.ylow(stringBounder)
	vline := klimt.ULineV(ylow - yhigh)
	for _, ent := range p.values {
		value := ent.value
		x := p.ruler.getPosInPixel(ent.tick)
		if len(lastValues) == 1 {
			ug.Apply(klimt.NewUTranslate(lastx, p.ypos(stringBounder, lastValues[0]))).Draw(klimt.ULineH(x - lastx))
		} else {
			for tmpx := lastx; tmpx < x; tmpx += 5 {
				ug.Apply(klimt.NewUTranslate(tmpx, yhigh)).Draw(vline)
			}
		}
		if !slices.Equal(lastValues, value.states) {
			ug.Apply(klimt.NewUTranslate(x, yhigh)).Draw(vline)
		}
		if value.comment != nil {
			label := p.createTextBlock(*value.comment)
			label.DrawU(ug.Apply(klimt.NewUTranslate(x+2, yhigh)))
		}
		lastx = x
		lastValues = value.states
	}
	ug.Apply(klimt.NewUTranslate(lastx, p.ypos(stringBounder, lastValues[0]))).
		Draw(klimt.ULineH(p.ruler.getWidth() - lastx))
	p.drawConstraints(ug.Apply(klimt.UTranslateDy(p.heightForConstraints(ug.StringBounder()))))
	p.drawNotes(ug.Apply(klimt.UTranslateDy(panelsMarginY)), abel.PositionTop)
	p.drawNotes(ug.Apply(klimt.UTranslateDy(p.heightForConstraints(stringBounder)+
		p.heightForNotes(stringBounder, abel.PositionTop)+float64(p.suggestedHeight)-panelsMarginY/2)),
		abel.PositionBottom)
}
