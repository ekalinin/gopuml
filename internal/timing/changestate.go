// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/ChangeState.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/TickInPlayer.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/TimingStyle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/uml"
)

// changeState is a change of state at a time: one state, or two for
// {A,B}, with an optional comment and colors.
type changeState struct {
	when    *TimeTick
	states  []string
	comment *string
	colors  *color.Colors
}

func newChangeState(when *TimeTick, comment *string, colors *color.Colors, states ...string) *changeState {
	if len(states) == 0 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return &changeState{when: when, states: states, comment: comment, colors: colors}
}

func (c *changeState) compareTo(other *changeState) int { return c.when.CompareTo(other.when) }

func (c *changeState) state() string { return c.states[0] }

// backColor is the background color of the change, or of the style.
func (c *changeState) backColor(skinParam klimt.ISkinParam, style *klimt.Style) color.HColor {
	if c.colors == nil || c.colors.Color(color.TypeBack) == nil {
		return style.Value(klimt.PNameBackGroundColor).AsColor(skinParam.IHtmlColorSet())
	}
	return c.colors.Color(color.TypeBack)
}

func (c *changeState) lineColor(skinParam klimt.ISkinParam, style *klimt.Style) color.HColor {
	if c.colors == nil || c.colors.Color(color.TypeLine) == nil {
		return style.Value(klimt.PNameLineColor).AsColor(skinParam.IHtmlColorSet())
	}
	return c.colors.Color(color.TypeLine)
}

// context is the fashion of the change: its colors and the stroke of the
// style.
func (c *changeState) context(skinParam klimt.ISkinParam, style *klimt.Style) *klimt.Fashion {
	return klimt.NewFashion(c.backColor(skinParam, style), c.lineColor(skinParam, style)).WithStroke(style.Stroke())
}

// The special states are compared case-sensitively, as upstream does.

func (c *changeState) isBlank() bool { return c.states[0] == "{...}" }

func (c *changeState) isCompletelyHidden() bool { return c.states[0] == "{hidden}" }

func (c *changeState) isFlat() bool { return isFlat(c.states[0]) }

func isFlat(state string) bool { return state == "{-}" }

// tickInPlayer is a time in the lane of a player; tick is nil for none.
type tickInPlayer struct {
	player player
	tick   *TimeTick
}

// TimingStyle is the type of a state player.
type TimingStyle int

const (
	TimingStyleRobust TimingStyle = iota
	TimingStyleConcise
	TimingStyleRectangle
)
