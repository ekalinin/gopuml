// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/Player.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/AbstractStatePlayer.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/PlayerRobust.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/PlayerConcise.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/PlayerRectangle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"slices"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// player is upstream's abstract Player: a lane of the diagram.
type player interface {
	base() *playerBase
	defineState(stateCode, label string)
	// setState changes the state at now, or sets the initial state when now
	// is nil; comment is nil for none.
	setState(now *TimeTick, comment *string, colors *color.Colors, states ...string)
	createConstraint(tick1, tick2 *TimeTick, message string, hasMessage bool, config *klimt.ArrowConfiguration)
	buildPlayerPanels() panels
}

// playerBase is the part of the players that upstream's Player holds.
type playerBase struct {
	skinParam              klimt.ISkinParam
	ruler                  *TimingRuler
	compact                bool
	title                  *klimt.Display
	suggestedHeight        int
	stereotype             *klimt.Stereotype
	generalBackgroundColor color.HColor
	cached                 panels
	notes                  []*timingNote
	playerFrame            *playerFrame
	sname                  klimt.SName
}

// newPlayerBase builds the title of the player with the styles of the time
// of its declaration, as upstream's constructor does.
func newPlayerBase(title string, skinParam klimt.ISkinParam, ruler *TimingRuler, compact bool,
	stereotype *klimt.Stereotype, generalBackgroundColor color.HColor, sname klimt.SName,
	suggestedHeight int) playerBase {
	b := playerBase{generalBackgroundColor: generalBackgroundColor, sname: sname, stereotype: stereotype,
		skinParam: skinParam, compact: compact, ruler: ruler, suggestedHeight: suggestedHeight,
		title: klimt.DisplayWithNewlines(skinParam.Pragma(), title)}
	b.playerFrame = newPlayerFrame(b.titleBlock(), skinParam, compact)
	return b
}

func (b *playerBase) base() *playerBase { return b }

// addNote adds a note at now, with the note style of its stereotype.
func (b *playerBase) addNote(now *TimeTick, note *klimt.Display, position abel.Position,
	stereotype *klimt.Stereotype) {
	style := timingStyle(klimt.SNameNote).AddStereotypeOf(stereotype).MergedStyle(b.skinParam.CurrentStyleBuilder())
	b.notes = append(b.notes, newTimingNote(now, note, position, b.skinParam, style))
}

// style is the style of the lane: timingDiagram, then the type of the
// player, with its stereotype.
func (b *playerBase) style() *klimt.Style {
	return timingStyle(b.sname).AddStereotypeOf(b.stereotype).MergedStyle(b.skinParam.CurrentStyleBuilder())
}

// titleBlock is the title, or the empty text block when it is white.
func (b *playerBase) titleBlock() klimt.TextBlock {
	if b.title.IsWhite() {
		return klimt.TextBlockUtilsEmptyTextBlock()
	}
	fontConfiguration := klimt.FontConfigurationFromStyle(b.skinParam,
		timingStyle().MergedStyle(b.skinParam.CurrentStyleBuilder()), nil)
	return b.title.Create(fontConfiguration, geom.HorizontalAlignmentLeft, b.skinParam)
}

func (b *playerBase) drawFrameTitle(ug klimt.UGraphic) { b.playerFrame.drawFrameTitle(ug) }

func (b *playerBase) frameHeight(stringBounder font.StringBounder) float64 {
	return b.playerFrame.height(stringBounder)
}

func (b *playerBase) setHeight(height int) { b.suggestedHeight = height }

// panelsOf returns the panels of a player, built on their first use.
func panelsOf(p player) panels {
	b := p.base()
	if b.cached == nil {
		b.cached = p.buildPlayerPanels()
	}
	return b.cached
}

// abstractStatePlayer is the part of robust, concise and rectangle players
// that upstream's AbstractStatePlayer holds: the changes of state in time
// order, the first change of a time kept, the constraints and the labels
// of the states.
type abstractStatePlayer struct {
	playerBase
	changes     []*changeState
	constraints []*timeConstraint
	// statesLabel maps the code of a state to its label; statesOrder keeps
	// the codes in the order of their first declaration.
	statesLabel      map[string]string
	statesOrder      []string
	initialState     *string
	initialColors    *color.Colors
	constraintOffset float64
}

func newAbstractStatePlayer(full string, skinParam klimt.ISkinParam, ruler *TimingRuler, compact bool,
	stereotype *klimt.Stereotype, generalBackgroundColor color.HColor, sname klimt.SName,
	constraintOffset float64) abstractStatePlayer {
	return abstractStatePlayer{playerBase: newPlayerBase(full, skinParam, ruler, compact, stereotype,
		generalBackgroundColor, sname, 0), statesLabel: map[string]string{}, constraintOffset: constraintOffset}
}

func (p *abstractStatePlayer) createConstraint(tick1, tick2 *TimeTick, message string, hasMessage bool,
	config *klimt.ArrowConfiguration) {
	p.constraints = append(p.constraints, newTimeConstraint(p.constraintOffset, tick1, tick2, message, hasMessage,
		p.skinParam, config))
}

func (p *abstractStatePlayer) setState(now *TimeTick, comment *string, colors *color.Colors, states ...string) {
	for i := range states {
		states[i] = p.decodeState(states[i])
	}
	if now == nil {
		p.initialState = &states[0]
		p.initialColors = colors
		return
	}
	change := newChangeState(now, comment, colors, states...)
	i, found := slices.BinarySearchFunc(p.changes, change, (*changeState).compareTo)
	if !found {
		p.changes = slices.Insert(p.changes, i, change)
	}
}

func (p *abstractStatePlayer) decodeState(code string) string {
	if label, ok := p.statesLabel[code]; ok {
		return label
	}
	return code
}

func (p *abstractStatePlayer) defineState(stateCode, label string) {
	if _, ok := p.statesLabel[stateCode]; !ok {
		p.statesOrder = append(p.statesOrder, stateCode)
	}
	p.statesLabel[stateCode] = label
}

// statesLabels returns the labels of the declared states in the order of
// their declarations.
func (p *abstractStatePlayer) statesLabels() []string {
	result := make([]string, len(p.statesOrder))
	for i, code := range p.statesOrder {
		result[i] = p.statesLabel[code]
	}
	return result
}

// playerRobust is a robust player: a line that moves between the levels of
// its states.
type playerRobust struct{ abstractStatePlayer }

func newPlayerRobust(full string, skinParam klimt.ISkinParam, ruler *TimingRuler, compact bool,
	stereotype *klimt.Stereotype, generalBackgroundColor color.HColor) *playerRobust {
	return &playerRobust{newAbstractStatePlayer(full, skinParam, ruler, compact, stereotype, generalBackgroundColor,
		klimt.SNameRobust, 2.5)}
}

func (p *playerRobust) buildPlayerPanels() panels {
	result := newPanelsRobust(p.ruler, p.skinParam, p.statesLabels(), p.suggestedHeight, p.style(), p.constraints)
	result.setInitialState(p.initialState)
	for _, change := range p.changes {
		result.addChange(change)
	}
	return result
}

// playerConcise is a concise player: a ribbon with the label of each state.
type playerConcise struct{ abstractStatePlayer }

func newPlayerConcise(full string, skinParam klimt.ISkinParam, ruler *TimingRuler, compact bool,
	stereotype *klimt.Stereotype, generalBackgroundColor color.HColor) *playerConcise {
	return &playerConcise{newAbstractStatePlayer(full, skinParam, ruler, compact, stereotype, generalBackgroundColor,
		klimt.SNameConcise, 1.0)}
}

func (p *playerConcise) buildPlayerPanels() panels {
	result := newPanelsConcise(p.ruler, p.skinParam, p.notes, p.suggestedHeight, p.style(), p.constraints)
	result.setInitialState(p.initialState, p.initialColors)
	for _, change := range p.changes {
		result.addChange(change)
	}
	return result
}

// playerRectangle is a rectangle player: a ribbon of rectangles.
type playerRectangle struct{ abstractStatePlayer }

func newPlayerRectangle(full string, skinParam klimt.ISkinParam, ruler *TimingRuler, compact bool,
	stereotype *klimt.Stereotype, generalBackgroundColor color.HColor) *playerRectangle {
	return &playerRectangle{newAbstractStatePlayer(full, skinParam, ruler, compact, stereotype,
		generalBackgroundColor, klimt.SNameRectangle, 1)}
}

func (p *playerRectangle) buildPlayerPanels() panels {
	result := newPanelsRectangle(p.ruler, p.skinParam, p.notes, p.suggestedHeight, p.style(), p.constraints)
	result.setInitialState(p.initialState, p.initialColors)
	for _, change := range p.changes {
		result.addChange(change)
	}
	return result
}

// unsupported throws the UnsupportedOperationException of the methods that
// a type of player does not implement.
func unsupported() { uml.Throw(uml.UnsupportedOperationException, "") }
