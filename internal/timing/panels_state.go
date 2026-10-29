// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/graphic/PanelsState.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/graphic/PanelsConcise.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/graphic/PanelsRectangle.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/graphic/HexaShape.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/graphic/PentaAShape.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/graphic/PentaBShape.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/graphic/PentaAShapeRectangle.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/graphic/PentaBShapeRectangle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

const defaultRibbonHeight = 24

// stateShapes are the shapes of the segments of a ribbon: between two
// changes (hexa), before the first change (pentaA) and after the last one
// (pentaB).
type stateShapes interface {
	drawHexa(ug klimt.UGraphic, length float64, change *changeState)
	drawPentaA(ug klimt.UGraphic, length float64, change *changeState)
	drawPentaB(ug klimt.UGraphic, length float64, change *changeState)
}

// panelsState is upstream's PanelsState: the ribbon of a concise or a
// rectangle player, with the label of each state, the comments above it,
// the constraints and the notes.
type panelsState struct {
	panelsBase
	shapes        stateShapes
	changes       []*changeState
	initialState  *string
	initialColors *color.Colors
}

func newPanelsState(ruler *TimingRuler, skinParam klimt.ISkinParam, suggestedHeight int, style *klimt.Style,
	notes []*timingNote, constraints []*timeConstraint) panelsState {
	if suggestedHeight == 0 {
		suggestedHeight = defaultRibbonHeight
	}
	return panelsState{panelsBase: newPanelsBase(ruler, skinParam, suggestedHeight, style, notes, constraints)}
}

func (p *panelsState) getTimeProjection(stringBounder font.StringBounder, tick *TimeTick) *intricatedPoint {
	if tick == nil {
		return nil
	}
	x := p.ruler.getPosInPixel(tick)
	y := p.heightForConstraints(stringBounder) + p.heightForNotes(stringBounder, abel.PositionTop) +
		p.heightForTopComment(stringBounder) + p.ribbonHeight()/2
	for _, change := range p.changes {
		if change.when.CompareTo(tick) == 0 {
			return &intricatedPoint{geom.NewXPoint2D(x, y), geom.NewXPoint2D(x, y)}
		}
	}
	return &intricatedPoint{geom.NewXPoint2D(x, y-p.ribbonHeight()/2), geom.NewXPoint2D(x, y+p.ribbonHeight()/2)}
}

func (p *panelsState) addChange(change *changeState) { p.changes = append(p.changes, change) }

func (p *panelsState) posInPixel(change *changeState) float64 {
	return p.ruler.getPosInPixel(change.when)
}

func (p *panelsState) drawLeftPanel(klimt.UGraphic, float64) {}

func (p *panelsState) getLeftPanelWidth(stringBounder font.StringBounder) float64 {
	return p.initialWidth(stringBounder)
}

func (p *panelsState) drawRightPanel(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	ugRibbon := ug.Apply(klimt.UTranslateDy(p.heightForConstraints(stringBounder) +
		p.heightForTopComment(stringBounder) + p.heightForNotes(stringBounder, abel.PositionTop)))
	p.drawBeforeZeroState(ugRibbon)
	p.drawBeforeZeroStateLabel(ugRibbon.Apply(klimt.UTranslateDy(p.ribbonHeight() / 2)))
	p.drawStates(ugRibbon)
	p.drawStatesLabels(ugRibbon.Apply(klimt.UTranslateDy(p.ribbonHeight() / 2)))
	p.drawConstraints(ug.Apply(klimt.UTranslateDy(p.heightForConstraints(stringBounder) / 2)))
	p.drawNotes(ug, abel.PositionTop)
	// Upstream leaves out the height of the top comments here.
	p.drawNotes(ug.Apply(klimt.UTranslateDy(p.heightForConstraints(stringBounder)+p.ribbonHeight()+
		p.heightForNotes(stringBounder, abel.PositionTop))), abel.PositionBottom)
}

func (p *panelsState) initialWidth(stringBounder font.StringBounder) float64 {
	if p.initialState == nil {
		return 0
	}
	return p.createTextBlock(*p.initialState).CalculateDimension(stringBounder).Width() + 2*panelsMarginX
}

func (p *panelsState) drawFlat(ug klimt.UGraphic, length float64, change *changeState) {
	line := klimt.ULineH(length)
	change.context(p.skinParam, p.style).Apply(ug).Apply(klimt.UTranslateDy(p.ribbonHeight() / 2)).Draw(line)
}

func (p *panelsState) ribbonHeight() float64 { return float64(p.suggestedHeight) }

func (p *panelsState) setInitialState(initialState *string, initialColors *color.Colors) {
	p.initialState = initialState
	p.initialColors = initialColors
}

func (p *panelsState) drawBeforeZeroState(ug klimt.UGraphic) {
	if p.initialState == nil {
		return
	}
	stringBounder := ug.StringBounder()
	if len(p.changes) == 0 {
		p.drawSingle(ug.Apply(klimt.UTranslateDx(-p.initialWidth(stringBounder))),
			p.initialWidth(stringBounder)+p.ruler.getWidth())
		return
	}
	a := p.posInPixel(p.changes[0])
	if isFlat(*p.initialState) {
		p.drawFlat(ug.Apply(klimt.UTranslateDx(-p.initialWidth(stringBounder))), p.initialWidth(stringBounder)+a,
			p.changes[0])
	} else {
		p.shapes.drawPentaA(ug.Apply(klimt.UTranslateDx(-p.initialWidth(stringBounder))),
			p.initialWidth(stringBounder)+a, p.changes[0])
	}
}

func (p *panelsState) drawBeforeZeroStateLabel(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	if p.initialState != nil && !isFlat(*p.initialState) {
		initial := p.createTextBlock(*p.initialState)
		dimInitial := initial.CalculateDimension(stringBounder)
		initial.DrawU(ug.Apply(klimt.NewUTranslate(-panelsMarginX-dimInitial.Width(), -dimInitial.Height()/2)))
	}
}

// drawSingle draws the initial state of a ribbon without changes.
func (p *panelsState) drawSingle(ug klimt.UGraphic, length float64) {
	back := p.style.Value(klimt.PNameBackGroundColor).AsColor(p.skinParam.IHtmlColorSet())
	line := p.style.Value(klimt.PNameLineColor).AsColor(p.skinParam.IHtmlColorSet())
	ctx := klimt.NewFashion(back, back).WithStroke(p.style.Stroke())
	ug = ctx.Apply(ug)
	border := klimt.ULineH(length)
	height := p.ribbonHeight()
	if isFlat(*p.initialState) {
		ug = ug.Apply(line)
		ug.Apply(klimt.UTranslateDy(height / 2)).Draw(border)
		return
	}
	rect := klimt.URectangleBuild(length, height)
	ug.Draw(rect)
	ug = ug.Apply(line)
	ug.Draw(border)
	ug.Apply(klimt.UTranslateDy(height)).Draw(border)
}

func (p *panelsState) drawStates(ug klimt.UGraphic) {
	for i := 0; i < len(p.changes)-1; i++ {
		a := p.posInPixel(p.changes[i])
		b := p.posInPixel(p.changes[i+1])
		switch {
		case p.changes[i].isFlat():
			p.drawFlat(ug.Apply(klimt.UTranslateDx(a)), b-a, p.changes[i])
		case !p.changes[i].isCompletelyHidden():
			p.shapes.drawHexa(ug.Apply(klimt.UTranslateDx(a)), b-a, p.changes[i])
		}
	}
	if len(p.changes) >= 1 {
		last := p.changes[len(p.changes)-1]
		a := p.posInPixel(last)
		switch {
		case last.isFlat():
			p.drawFlat(ug.Apply(klimt.UTranslateDx(a)), p.ruler.getWidth()-a, last)
		case !last.isCompletelyHidden():
			p.shapes.drawPentaB(ug.Apply(klimt.UTranslateDx(a)), p.ruler.getWidth()-a, last)
		}
	}
}

func (p *panelsState) drawStatesLabels(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	for i, change := range p.changes {
		x := p.ruler.getPosInPixel(change.when)
		if !change.isBlank() && !change.isCompletelyHidden() && !change.isFlat() {
			state := p.createTextBlock(change.state())
			dim := state.CalculateDimension(stringBounder)
			var xtext float64
			if i == len(p.changes)-1 {
				xtext = x + panelsMarginX
			} else {
				x2 := p.ruler.getPosInPixel(p.changes[i+1].when)
				xtext = (x+x2)/2 - dim.Width()/2
			}
			state.DrawU(ug.Apply(klimt.NewUTranslate(xtext, -dim.Height()/2)))
		}
		commentTopBlock := p.commentTopBlock(change)
		dimComment := commentTopBlock.CalculateDimension(stringBounder)
		commentTopBlock.DrawU(ug.Apply(klimt.NewUTranslate(x+panelsMarginX, -p.ribbonHeight()/2-dimComment.Height())))
	}
}

func (p *panelsState) commentTopBlock(change *changeState) klimt.TextBlock {
	if change.comment == nil {
		return klimt.TextBlockUtilsEmpty(0, 0)
	}
	return p.createTextBlock(*change.comment)
}

func (p *panelsState) heightForTopComment(stringBounder font.StringBounder) float64 {
	result := 0.0
	for _, change := range p.changes {
		result = max(result, p.commentTopBlock(change).CalculateDimension(stringBounder).Height())
	}
	return result
}

// contextWithInitialColors is the fashion of a change with the colors of
// the initial state.
func (p *panelsState) contextWithInitialColors(change *changeState) *klimt.Fashion {
	ctx := change.context(p.skinParam, p.style)
	back := p.initialColors.Color(color.TypeBack)
	line := p.initialColors.Color(color.TypeLine)
	if back != nil {
		ctx = ctx.WithBackColor(back)
	}
	if line != nil {
		ctx = ctx.WithForeColor(line)
	}
	return ctx
}

func (p *panelsState) heightForConstraints(stringBounder font.StringBounder) float64 {
	return max(5, p.panelsBase.heightForConstraints(stringBounder))
}

func (p *panelsState) getFullHeight(stringBounder font.StringBounder) float64 {
	return p.heightForConstraints(stringBounder) + p.heightForTopComment(stringBounder) +
		p.heightForNotes(stringBounder, abel.PositionTop) + p.ribbonHeight() +
		p.heightForNotes(stringBounder, abel.PositionBottom) + panelsBottomMargin
}

// panelsConcise draws the segments of a ribbon as hexagons.
type panelsConcise struct{ panelsState }

func newPanelsConcise(ruler *TimingRuler, skinParam klimt.ISkinParam, notes []*timingNote, suggestedHeight int,
	style *klimt.Style, constraints []*timeConstraint) *panelsConcise {
	p := &panelsConcise{newPanelsState(ruler, skinParam, suggestedHeight, style, notes, constraints)}
	p.shapes = p
	return p
}

func (p *panelsConcise) drawHexa(ug klimt.UGraphic, length float64, change *changeState) {
	drawHexaShape(ug, length, p.ribbonHeight(), change.context(p.skinParam, p.style))
}

func (p *panelsConcise) drawPentaA(ug klimt.UGraphic, length float64, change *changeState) {
	drawPentaAShape(ug, length, p.ribbonHeight(), p.contextWithInitialColors(change))
}

func (p *panelsConcise) drawPentaB(ug klimt.UGraphic, length float64, change *changeState) {
	drawPentaBShape(ug, length, p.ribbonHeight(), change.context(p.skinParam, p.style))
}

// panelsRectangle draws the segments of a ribbon as rectangles.
type panelsRectangle struct{ panelsState }

func newPanelsRectangle(ruler *TimingRuler, skinParam klimt.ISkinParam, notes []*timingNote, suggestedHeight int,
	style *klimt.Style, constraints []*timeConstraint) *panelsRectangle {
	p := &panelsRectangle{newPanelsState(ruler, skinParam, suggestedHeight, style, notes, constraints)}
	p.shapes = p
	return p
}

func (p *panelsRectangle) drawHexa(ug klimt.UGraphic, length float64, change *changeState) {
	shape := klimt.URectangleBuild(length, p.ribbonHeight())
	change.context(p.skinParam, p.style).Apply(ug).Draw(shape)
}

func (p *panelsRectangle) drawPentaA(ug klimt.UGraphic, length float64, change *changeState) {
	drawPentaAShapeRectangle(ug, length, p.ribbonHeight(), p.contextWithInitialColors(change))
}

func (p *panelsRectangle) drawPentaB(ug klimt.UGraphic, length float64, change *changeState) {
	drawPentaBShapeRectangle(ug, length, p.ribbonHeight(), change.context(p.skinParam, p.style))
}

const shapeDelta = 12.0

// drawHexaShape is upstream's HexaShape: a segment with pointed ends.
func drawHexaShape(ug klimt.UGraphic, width, height float64, ctx *klimt.Fashion) {
	ug = ctx.Apply(ug)
	polygon := klimt.NewUPolygon()
	polygon.AddPoint(shapeDelta, 0)
	polygon.AddPoint(width-shapeDelta, 0)
	polygon.AddPoint(width, height/2)
	polygon.AddPoint(width-shapeDelta, height)
	polygon.AddPoint(shapeDelta, height)
	polygon.AddPoint(0, height/2)
	ug.Draw(polygon)
}

// drawPentaAShape is upstream's PentaAShape: the segment before the first
// change, filled, then outlined without its left side.
func drawPentaAShape(ug klimt.UGraphic, width, height float64, ctx *klimt.Fashion) {
	polygon := klimt.NewUPolygon()
	polygon.AddPoint(0, 0)
	polygon.AddPoint(width-shapeDelta, 0)
	polygon.AddPoint(width, height/2)
	polygon.AddPoint(width-shapeDelta, height)
	polygon.AddPoint(0, height)
	ctx.WithForeColor(ctx.BackColor()).Apply(ug).Draw(polygon)
	path := klimt.UPathNone()
	path.MoveTo(0, 0)
	path.LineTo(width-shapeDelta, 0)
	path.LineTo(width, height/2)
	path.LineTo(width-shapeDelta, height)
	path.LineTo(0, height)
	ctx.Apply(ug).Draw(path)
}

// drawPentaBShape is upstream's PentaBShape: the segment after the last
// change, filled, then outlined without its right side.
func drawPentaBShape(ug klimt.UGraphic, width, height float64, ctx *klimt.Fashion) {
	polygon := klimt.NewUPolygon()
	polygon.AddPoint(shapeDelta, 0)
	polygon.AddPoint(width, 0)
	polygon.AddPoint(width, height)
	polygon.AddPoint(shapeDelta, height)
	polygon.AddPoint(0, height/2)
	ctx.WithForeColor(ctx.BackColor()).Apply(ug).Draw(polygon)
	path := klimt.UPathNone()
	path.MoveTo(width, 0)
	path.LineTo(shapeDelta, 0)
	path.LineTo(0, height/2)
	path.LineTo(shapeDelta, height)
	path.LineTo(width, height)
	ctx.Apply(ug).Draw(path)
}

// drawPentaAShapeRectangle is upstream's PentaAShapeRectangle.
func drawPentaAShapeRectangle(ug klimt.UGraphic, width, height float64, ctx *klimt.Fashion) {
	polygon := klimt.URectangleBuild(width, height)
	ctx.WithForeColor(ctx.BackColor()).Apply(ug).Draw(polygon)
	path := klimt.UPathNone()
	path.MoveTo(0, 0)
	path.LineTo(width, 0)
	path.LineTo(width, height)
	path.LineTo(0, height)
	ctx.Apply(ug).Draw(path)
}

// drawPentaBShapeRectangle is upstream's PentaBShapeRectangle.
func drawPentaBShapeRectangle(ug klimt.UGraphic, width, height float64, ctx *klimt.Fashion) {
	polygon := klimt.URectangleBuild(width, height)
	ctx.WithForeColor(ctx.BackColor()).Apply(ug).Draw(polygon)
	path := klimt.UPathNone()
	path.MoveTo(width, 0)
	path.LineTo(0, 0)
	path.LineTo(0, height)
	path.LineTo(width, height)
	ctx.Apply(ug).Draw(path)
}
