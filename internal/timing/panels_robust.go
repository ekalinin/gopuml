// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/graphic/PanelsRobust.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"slices"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

const (
	histogramBottomMargin = 12.0
	histogramInitialWidth = 40.0
)

// panelsRobust is upstream's PanelsRobust: a line that moves between the
// levels of the states, with their labels on the left. It draws no notes.
type panelsRobust struct {
	panelsBase
	changes []*changeState
	// allStates are the levels from the bottom: the declared states in
	// reverse order, then the initial state, then the states of the
	// changes.
	allStates    []string
	initialState *string
}

func newPanelsRobust(ruler *TimingRuler, skinParam klimt.ISkinParam, someStates []string, suggestedHeight int,
	style *klimt.Style, constraints []*timeConstraint) *panelsRobust {
	p := &panelsRobust{panelsBase: newPanelsBase(ruler, skinParam, suggestedHeight, style, nil, constraints)}
	p.allStates = slices.Clone(someStates)
	slices.Reverse(p.allStates)
	p.constraintDeltaY = p.getConstraintDeltaY
	return p
}

func (p *panelsRobust) getTimeProjection(stringBounder font.StringBounder, tick *TimeTick) *intricatedPoint {
	if tick == nil {
		return nil
	}
	x := p.ruler.getPosInPixel(tick)
	states := p.statesAt(tick)
	if len(states) == 0 {
		return nil
	}
	heightForConstraints := p.heightForConstraints(stringBounder)
	if len(states) == 1 {
		y := p.yOfState(states[0]) + heightForConstraints
		return &intricatedPoint{geom.NewXPoint2D(x, y), geom.NewXPoint2D(x, y)}
	}
	y1 := p.yOfState(states[0]) + heightForConstraints
	y2 := p.yOfState(states[1]) + heightForConstraints
	return &intricatedPoint{geom.NewXPoint2D(x, y1), geom.NewXPoint2D(x, y2)}
}

// statesAt returns the states at a time: the two states around a change at
// that time, or the state of the change before it.
func (p *panelsRobust) statesAt(tick *TimeTick) []string {
	if len(p.changes) == 0 {
		return nil
	}
	for i, change := range p.changes {
		comparison := change.when.CompareTo(tick)
		if comparison == 0 {
			if i == 0 && p.initialState == nil {
				return []string{change.state()}
			}
			if i == 0 && p.initialState != nil {
				return []string{*p.initialState, change.state()}
			}
			return []string{p.changes[i-1].state(), change.state()}
		}
		if comparison > 0 {
			// A time before the first change is drawn at the first one.
			changeIndex := i - 1
			if i == 0 {
				changeIndex = 0
			}
			return []string{p.changes[changeIndex].state()}
		}
	}
	return []string{p.changes[len(p.changes)-1].state()}
}

func (p *panelsRobust) addChange(change *changeState) {
	p.changes = append(p.changes, change)
	if change.isCompletelyHidden() {
		return
	}
	for _, state := range change.states {
		if !slices.Contains(p.allStates, state) {
			p.allStates = append(p.allStates, state)
		}
	}
}

func (p *panelsRobust) points(n int) []geom.XPoint2D {
	change := p.changes[n]
	x := p.ruler.getPosInPixel(change.when)
	if len(change.states) == 2 {
		return []geom.XPoint2D{geom.NewXPoint2D(x, p.yOfState(change.states[0])),
			geom.NewXPoint2D(x, p.yOfState(change.states[1]))}
	}
	return []geom.XPoint2D{geom.NewXPoint2D(x, p.yOfState(change.states[0]))}
}

func (p *panelsRobust) pointX(n int) float64 { return p.ruler.getPosInPixel(p.changes[n].when) }

func (p *panelsRobust) pointMinY(n int) float64 {
	states := p.changes[n].states
	if len(states) == 2 {
		return min(p.yOfState(states[0]), p.yOfState(states[1]))
	}
	return p.yOfState(states[0])
}

func (p *panelsRobust) pointMaxY(n int) float64 {
	states := p.changes[n].states
	if len(states) == 2 {
		return max(p.yOfState(states[0]), p.yOfState(states[1]))
	}
	return p.yOfState(states[0])
}

// drawLeftPanel draws the labels of the levels, right-aligned 5 pixels
// before the lane when there is room.
func (p *panelsRobust) drawLeftPanel(ug klimt.UGraphic, fullAvailableWidth float64) {
	stringBounder := ug.StringBounder()
	ug = ug.Apply(klimt.UTranslateDy(p.heightForConstraints(stringBounder)))
	width := p.statesWidth(stringBounder)
	if p.initialState != nil {
		width += histogramInitialWidth
	}
	if fullAvailableWidth > width+5 {
		ug = ug.Apply(klimt.UTranslateDx(fullAvailableWidth - width - 5))
	} else {
		ug = ug.Apply(klimt.UTranslateDx(fullAvailableWidth - width))
	}
	for _, state := range p.allStates {
		label := p.createTextBlock(state)
		dim := label.CalculateDimension(stringBounder)
		label.DrawU(ug.Apply(klimt.UTranslateDy(p.yOfState(state) - dim.Height()/2 + 1)))
	}
}

// drawRightPanel draws nothing without changes, not even the initial
// state.
func (p *panelsRobust) drawRightPanel(ug klimt.UGraphic) {
	if len(p.changes) == 0 {
		return
	}
	ug = p.context().Apply(ug)
	ug = ug.Apply(klimt.UTranslateDy(p.heightForConstraints(ug.StringBounder())))
	p.drawHlines(ug)
	p.drawVlines(ug)
	p.drawLabels(ug)
	p.drawConstraints(ug.Apply(klimt.UTranslateDy(-constraintTopMargin)))
}

func (p *panelsRobust) statesWidth(stringBounder font.StringBounder) float64 {
	result := 0.0
	for _, state := range p.allStates {
		result = max(result, p.createTextBlock(state).CalculateDimension(stringBounder).Width())
	}
	return result
}

func (p *panelsRobust) getLeftPanelWidth(stringBounder font.StringBounder) float64 {
	if p.initialState == nil {
		return p.statesWidth(stringBounder)
	}
	return p.statesWidth(stringBounder) + histogramInitialWidth
}

func (p *panelsRobust) drawHlines(ug klimt.UGraphic) {
	if p.initialState != nil {
		for _, pt := range p.points(0) {
			drawHLineFromPoint(ug, p.initialPoint(), histogramInitialWidth+pt.X)
		}
	}
	for i, change := range p.changes {
		if change.isCompletelyHidden() {
			continue
		}
		x2 := p.ruler.getWidth()
		if i < len(p.changes)-1 {
			x2 = p.pointX(i + 1)
		}
		length := x2 - p.pointX(i)
		points := p.points(i)
		if len(points) == 2 {
			drawHBlock(ug.Apply(klimt.Bg(change.backColor(p.skinParam, p.style))), points[0], points[1], length)
		}
		if i < len(p.changes)-1 {
			for _, pt := range points {
				drawHLineFromPoint(ug, pt, length)
			}
		}
	}
	if !p.changes[len(p.changes)-1].isCompletelyHidden() {
		for _, pt := range p.points(len(p.changes) - 1) {
			drawHLineFromPoint(ug, pt, p.ruler.getWidth()-pt.X)
		}
	}
}

// drawHBlock draws two states at once: a box between their levels with
// vertical lines every 5 pixels.
func drawHBlock(ug klimt.UGraphic, pt1, pt2 geom.XPoint2D, length float64) {
	minY := min(pt1.Y, pt2.Y)
	maxY := max(pt1.Y, pt2.Y)
	pt := geom.NewXPoint2D(pt1.X, minY)
	ug = ug.Apply(klimt.UTranslatePoint(pt))
	ug.Draw(klimt.URectangleBuild(length, maxY-minY))
	for x := 0.0; x < length; x += 5 {
		ug.Apply(klimt.UTranslateDx(x)).Draw(klimt.ULineV(maxY - minY))
	}
}

func (p *panelsRobust) drawVlines(ug klimt.UGraphic) {
	if p.initialState != nil {
		before := p.initialPoint()
		current := p.points(0)[0]
		drawVline(ug, current.X, current.Y, klimt.ULineV(before.Y-current.Y))
	}
	for i := 1; i < len(p.changes); i++ {
		if p.changes[i-1].isCompletelyHidden() || p.changes[i].isCompletelyHidden() {
			continue
		}
		minY := min(p.pointMinY(i), p.pointMinY(i-1))
		maxY := max(p.pointMaxY(i), p.pointMaxY(i-1))
		drawVline(ug, p.pointX(i), minY, klimt.ULineV(maxY-minY))
	}
}

func (p *panelsRobust) drawLabels(ug klimt.UGraphic) {
	for i, change := range p.changes {
		ptLabel := p.points(i)[0]
		if change.comment == nil {
			continue
		}
		label := p.createTextBlock(*change.comment)
		dim := label.CalculateDimension(ug.StringBounder())
		label.DrawU(ug.Apply(klimt.UTranslatePoint(ptLabel).Compose(klimt.NewUTranslate(2, -dim.Height()))))
	}
}

// getConstraintDeltaY is the highest level of the states between the two
// times of a constraint.
func (p *panelsRobust) getConstraintDeltaY(constraint *timeConstraint) float64 {
	y := p.yOfStateAt(constraint.tick1)
	for _, change := range p.changes {
		if constraint.containsStrict(change.when) {
			y = min(y, p.yOfStateAt(change.when))
		}
	}
	return y
}

func (p *panelsRobust) initialPoint() geom.XPoint2D {
	return geom.NewXPoint2D(-histogramInitialWidth, p.yOfState(*p.initialState))
}

func (p *panelsRobust) heightForConstraints(stringBounder font.StringBounder) float64 {
	return max(10, p.panelsBase.heightForConstraints(stringBounder))
}

func (p *panelsRobust) getFullHeight(stringBounder font.StringBounder) float64 {
	height := p.heightForConstraints(stringBounder)
	if len(p.allStates) > 0 {
		height += float64(p.stepHeight() * float64(len(p.allStates)-1))
	}
	height += histogramBottomMargin
	return height + 6
}

// yOfState is the y of the level of a state; an unknown state is one step
// above the top level.
func (p *panelsRobust) yOfState(state string) float64 {
	nb := len(p.allStates) - 1 - slices.Index(p.allStates, state)
	return float64(p.stepHeight() * float64(nb))
}

// yOfStateAt is the y of the last state at a time; a time without states
// throws upstream's IndexOutOfBoundsException.
func (p *panelsRobust) yOfStateAt(when *TimeTick) float64 {
	states := p.statesAt(when)
	if len(states) == 0 {
		uml.Throw(uml.IndexOutOfBoundsException, "Index: -1")
	}
	return p.yOfState(states[len(states)-1])
}

// stepHeight is the distance between two levels: 20, or the suggested
// height shared between the levels with Java's int division.
func (p *panelsRobust) stepHeight() float64 {
	if p.suggestedHeight == 0 || len(p.allStates) <= 1 {
		return 20
	}
	return float64(int32(p.suggestedHeight) / int32(len(p.allStates)-1))
}

func (p *panelsRobust) setInitialState(initialState *string) {
	p.initialState = initialState
	if initialState != nil && !slices.Contains(p.allStates, *initialState) {
		p.allStates = append(p.allStates, *initialState)
	}
}
