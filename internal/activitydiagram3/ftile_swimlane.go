// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/Swimlane.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/Swimable.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Swimlane is a lane of an activity diagram, with its layout (translation,
// actual width and bounds of its drawing).
type Swimlane struct {
	name    string
	order   int
	display *klimt.Display
	colors  *color.Colors

	translate   klimt.UTranslate
	actualWidth float64
	minMax      *geom.MinMax
}

// NewSwimlane returns the lane name at position order.
func NewSwimlane(name string, order int, pragma *klimt.Pragma) *Swimlane {
	return &Swimlane{name: name, order: order, display: klimt.DisplayWithNewlines(pragma, name),
		colors: color.EmptyColors()}
}

// String returns the name.
func (s *Swimlane) String() string { return s.name }

// Name returns the name.
func (s *Swimlane) Name() string { return s.name }

// Order returns the position of the lane.
func (s *Swimlane) Order() int { return s.order }

// Display returns the title.
func (s *Swimlane) Display() *klimt.Display { return s.display }

// SetDisplay sets the title.
func (s *Swimlane) SetDisplay(label *klimt.Display) { s.display = label }

// Colors returns the colors of the lane.
func (s *Swimlane) Colors() *color.Colors { return s.colors }

// SetSpecificColorTOBEREMOVED adds a color, unless it is nil.
func (s *Swimlane) SetSpecificColorTOBEREMOVED(typ color.ColorType, c color.HColor) {
	if c != nil {
		s.colors = s.colors.Add(typ, c)
	}
}

// SetColors sets the colors.
func (s *Swimlane) SetColors(colors *color.Colors) { s.colors = colors }

// Translate returns the position of the lane.
func (s *Swimlane) Translate() klimt.UTranslate { return s.translate }

// SetTranslate sets the position of the lane.
func (s *Swimlane) SetTranslate(translate klimt.UTranslate) { s.translate = translate }

// SetWidth sets the actual width of the lane.
func (s *Swimlane) SetWidth(actualWidth float64) { s.actualWidth = actualWidth }

// ActualWidth returns the actual width of the lane.
func (s *Swimlane) ActualWidth() float64 { return s.actualWidth }

// SetMinMax sets the bounds of the drawing of the lane.
func (s *Swimlane) SetMinMax(minMax geom.MinMax) { s.minMax = &minMax }

// MinMax returns the bounds of the drawing of the lane; it panics before
// they are set, as upstream dereferences null.
func (s *Swimlane) MinMax() geom.MinMax {
	if s.minMax == nil {
		uml.ThrowNPE("")
	}
	return *s.minMax
}

// CompareTo orders lanes by position.
func (s *Swimlane) CompareTo(other *Swimlane) int {
	switch {
	case s.order < other.order:
		return -1
	case s.order > other.order:
		return 1
	}
	return 0
}

// IsSmallerThanAllOthers reports whether no lane of others is before s,
// others being more than s alone.
func (s *Swimlane) IsSmallerThanAllOthers(others SwimlaneSet) bool {
	if len(others) == 1 && others.Contains(s) {
		return false
	}
	for _, sw := range others {
		if sw.CompareTo(s) < 0 {
			return false
		}
	}
	return true
}

// SwimlaneSet is a set of lanes (upstream's HashSet<Swimlane>), in the order
// of insertion. It may hold nil, as upstream's set may hold null.
type SwimlaneSet []*Swimlane

// Contains reports whether s is in the set.
func (set SwimlaneSet) Contains(s *Swimlane) bool {
	for _, x := range set {
		if x == s {
			return true
		}
	}
	return false
}

// add returns the set with s added.
func (set SwimlaneSet) add(s *Swimlane) SwimlaneSet {
	if set.Contains(s) {
		return set
	}
	return append(set, s)
}

// addAll returns the set with the lanes of other added.
func (set SwimlaneSet) addAll(other SwimlaneSet) SwimlaneSet {
	for _, s := range other {
		set = set.add(s)
	}
	return set
}

// Swimable is an element that lies in lanes.
type Swimable interface {
	Swimlanes() SwimlaneSet
	SwimlaneIn() *Swimlane
	SwimlaneOut() *Swimlane
}
