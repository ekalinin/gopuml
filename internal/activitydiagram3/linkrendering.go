// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/LinkRendering.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// LinkRendering is the label and the colors of an arrow. It is immutable;
// a nil *LinkRendering is upstream's null.
type LinkRendering struct {
	rainbow *Rainbow
	display *klimt.Display
}

// LinkRenderingCreate returns an arrow of the given colors without label.
func LinkRenderingCreate(rainbow *Rainbow) *LinkRendering {
	return newLinkRendering(rainbow, klimt.DisplayNULL())
}

// LinkRenderingNone returns an arrow without colors and label.
func LinkRenderingNone() *LinkRendering { return LinkRenderingCreate(RainbowNone()) }

func newLinkRendering(rainbow *Rainbow, display *klimt.Display) *LinkRendering {
	if rainbow == nil {
		uml.ThrowNPE("")
	}
	return &LinkRendering{rainbow: rainbow, display: display}
}

// WithRainbow returns the arrow with other colors.
func (l *LinkRendering) WithRainbow(rainbow *Rainbow) *LinkRendering {
	return newLinkRendering(rainbow, l.display)
}

// WithDisplay returns the arrow with another label.
func (l *LinkRendering) WithDisplay(display *klimt.Display) *LinkRendering {
	return newLinkRendering(l.rainbow, display)
}

// Display returns the label.
func (l *LinkRendering) Display() *klimt.Display { return l.display }

// Rainbow returns the colors.
func (l *LinkRendering) Rainbow() *Rainbow { return l.rainbow }

// RainbowOr is upstream's getRainbow(Rainbow): the colors, or defaultColor
// when there are none.
func (l *LinkRendering) RainbowOr(defaultColor *Rainbow) *Rainbow {
	if l.rainbow.Size() == 0 {
		return defaultColor
	}
	return l.rainbow
}

// IsNone reports whether the arrow has neither label nor colors.
func (l *LinkRendering) IsNone() bool {
	return klimt.DisplayIsNull(l.display) && l.rainbow.Size() == 0
}
