// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/GanttConstraint.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// GanttConstraint is upstream's GanttConstraint: a link between two
// instants, drawn as an arrow, with the styles of the time it was created.
type GanttConstraint struct {
	decoration.WithLinkType
	source, dest *TaskInstant
	styleBuilder *klimt.StyleBuilder
}

// NewGanttConstraint is new GanttConstraint; forcedColor is the color of
// "with ... link", nil for none.
func NewGanttConstraint(styleBuilder *klimt.StyleBuilder, source, dest *TaskInstant,
	forcedColor color.HColor) *GanttConstraint {
	c := &GanttConstraint{source: source, dest: dest, styleBuilder: styleBuilder}
	c.WithLinkType = decoration.InitWithLinkType(
		decoration.NewLinkType(decoration.LinkDecorNone, decoration.LinkDecorNone), func() {})
	c.SetSpecificColor(forcedColor, 0)
	return c
}

// IsOn is isOn: whether the constraint starts or ends on task.
func (c *GanttConstraint) IsOn(task Task) bool {
	return c.source.Moment() == Moment(task) || c.dest.Moment() == Moment(task)
}

// IsThereRightArrow is isThereRightArrow: whether an arrow arrives at the
// right of task.
func (c *GanttConstraint) IsThereRightArrow(task Task) bool {
	if c.dest.Moment() == Moment(task) && c.dest.Attribute() == TaskAttributeEnd {
		return true
	}
	return c.source.Moment() == Moment(task) && c.dest.Attribute() == TaskAttributeEnd &&
		c.source.Attribute() == TaskAttributeEnd
}

// IsHidden is isHidden: whether an end of the constraint is outside
// [min, max].
func (c *GanttConstraint) IsHidden(min, max TimePoint) bool {
	return isOutside(c.source.InstantPrecise(), min, max) || isOutside(c.dest.InstantPrecise(), min, max)
}

func isOutside(now, min, max TimePoint) bool { return now.CompareTo(min) < 0 || now.CompareTo(max) > 0 }
