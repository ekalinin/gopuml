// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/PlayerBinary.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"slices"
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

const (
	lowString  = "0"
	highString = "1"
)

// binaryValue is a value of a binary player at a time; the tick of the
// first value given at that time is kept with the last value, as upstream's
// TreeMap.put.
type binaryValue struct {
	tick  *TimeTick
	value *changeState
}

// playerBinary is a binary player: a line at its high or its low level.
type playerBinary struct {
	playerBase
	constraints  []*timeConstraint
	values       []binaryValue
	initialState *changeState
}

func newPlayerBinary(code string, skinParam klimt.ISkinParam, ruler *TimingRuler, compact bool,
	stereotype *klimt.Stereotype) *playerBinary {
	return &playerBinary{playerBase: newPlayerBase(code, skinParam, ruler, compact, stereotype, nil,
		klimt.SNameBinary, 30)}
}

func (p *playerBinary) buildPlayerPanels() panels {
	return newPanelsBinary(p.ruler, p.skinParam, p.suggestedHeight, p.style(), p.values, p.constraints,
		p.initialState, p.notes)
}

func (p *playerBinary) defineState(string, string) { unsupported() }

// setState reads 1 and high, in any case, as high, and every other value as
// low; the last value given at a time wins.
func (p *playerBinary) setState(now *TimeTick, comment *string, colors *color.Colors, states ...string) {
	cs := newChangeState(now, comment, colors, convertBinary(states)...)
	if now == nil {
		p.initialState = cs
		return
	}
	i, found := slices.BinarySearchFunc(p.values, now, func(v binaryValue, t *TimeTick) int {
		return v.tick.CompareTo(t)
	})
	if found {
		p.values[i].value = cs
		return
	}
	p.values = slices.Insert(p.values, i, binaryValue{tick: now, value: cs})
}

func convertBinary(states []string) []string {
	if len(states) == 1 {
		return []string{convertBinaryValue(states[0])}
	}
	return []string{convertBinaryValue(states[0]), convertBinaryValue(states[1])}
}

func convertBinaryValue(value string) string {
	if value == "1" || strings.EqualFold(value, "high") {
		return highString
	}
	return lowString
}

func (p *playerBinary) createConstraint(tick1, tick2 *TimeTick, message string, hasMessage bool,
	config *klimt.ArrowConfiguration) {
	p.constraints = append(p.constraints, newTimeConstraint(2.5, tick1, tick2, message, hasMessage, p.skinParam,
		config))
}
