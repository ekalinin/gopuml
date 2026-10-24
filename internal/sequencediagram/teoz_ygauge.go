// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/YGauge.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"github.com/ekalinin/gopuml/internal/jfmt"
	"github.com/ekalinin/gopuml/internal/uml"
)

// YGauge is the vertical extent of a tile on the Y line: its top (min),
// its bottom (max), and for tiles with an arrow, the contact line shared
// by parallel (&) messages and the origin of their group.
//
// The max of a gauge is the chaining point of the next tile: it is always
// built with AddAtLeast on a point created by AddAtLeast, so that chains
// of tiles do not stack delegations (upstream's CHAIN ANCHORING note).
type YGauge struct {
	min, max        Real
	contact, origin Real
}

// NewYGauge returns a gauge without contact line.
func NewYGauge(min, max Real) *YGauge { return &YGauge{min: min, max: max} }

func newYGaugeContact(min, max, contact, origin Real) *YGauge {
	return &YGauge{min: min, max: max, contact: contact, origin: origin}
}

// String is upstream's toString.
func (g *YGauge) String() string {
	return jfmt.DoubleToString(g.min.CurrentValue()) + " " + jfmt.DoubleToString(g.max.CurrentValue())
}

// YGaugeCreate returns the gauge of a tile of the height starting at min.
func YGaugeCreate(min Real, height float64) *YGauge {
	if height < 0 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return NewYGauge(min, min.AddAtLeast(height))
}

// YGaugeCreateWithContact returns the gauge of a tile with a contact line
// (a message arrow) chained after currentY: its min floats so that a
// parallel member with a taller label can push it down.
func YGaugeCreateWithContact(currentY *YGauge, contactRelative, height float64) *YGauge {
	if height < 0 || contactRelative < 0 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	origin := currentY.Max()
	contact := origin.AddAtLeast(contactRelative)
	min := contact.AddFixed(-contactRelative)
	max := contact.AddAtLeast(height - contactRelative)
	return newYGaugeContact(min, max, contact, origin)
}

// YGaugeCreateParallel returns the gauge of a parallel (&) tile: its
// contact line is the one of the group, pushed down when this member's
// label is taller, and its max dominates the previous member's.
func YGaugeCreateParallel(currentY *YGauge, contactRelative, height float64) *YGauge {
	contact := currentY.Contact()
	origin := currentY.Origin()
	if contact == nil || origin == nil {
		// The previous tile has no contact line: fall back to top alignment
		return YGaugeCreate(currentY.Min(), height)
	}
	contact.EnsureBiggerThan(origin.AddFixed(contactRelative))
	min := contact.AddFixed(-contactRelative)
	max := contact.AddAtLeast(height - contactRelative)
	max.EnsureBiggerThan(currentY.Max())
	return newYGaugeContact(min, max, contact, origin)
}

// YGaugeCreatePropagating returns the gauge of a tile without contact line
// placed between parallel members (life events, spaces): the contact line
// and origin go through it.
func YGaugeCreatePropagating(currentY *YGauge, height float64) *YGauge {
	if height < 0 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	min := currentY.Max()
	return newYGaugeContact(min, min.AddAtLeast(height), currentY.Contact(), currentY.Origin())
}

// Min returns the top of the tile.
func (g *YGauge) Min() Real { return g.min }

// Max returns the bottom of the tile.
func (g *YGauge) Max() Real { return g.max }

// Contact returns the contact line, nil for none.
func (g *YGauge) Contact() Real { return g.contact }

// Origin returns the origin of the parallel group, nil for none.
func (g *YGauge) Origin() Real { return g.origin }
