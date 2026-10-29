// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/PlayerClock.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// playerClock is a clock: a square wave of a period, with a pulse and an
// offset. It has no states, no constraints and no notes.
type playerClock struct {
	playerBase
	period jmath.BigDecimal
	pulse  jmath.BigDecimal
	offset jmath.BigDecimal
}

func newPlayerClock(title string, skinParam klimt.ISkinParam, ruler *TimingRuler, period, pulse,
	offset jmath.BigDecimal, compact bool, stereotype *klimt.Stereotype) *playerClock {
	return &playerClock{playerBase: newPlayerBase(title, skinParam, ruler, compact, stereotype, nil,
		klimt.SNameClock, 30), period: period, pulse: pulse, offset: offset}
}

func (p *playerClock) buildPlayerPanels() panels {
	return newPanelsClock(p.ruler, p.skinParam, p.suggestedHeight, p.style(), p.period, p.pulse, p.offset)
}

func (p *playerClock) defineState(string, string) { unsupported() }

func (p *playerClock) setState(*TimeTick, *string, *color.Colors, ...string) { unsupported() }

func (p *playerClock) createConstraint(*TimeTick, *TimeTick, string, bool, *klimt.ArrowConfiguration) {
	unsupported()
}
