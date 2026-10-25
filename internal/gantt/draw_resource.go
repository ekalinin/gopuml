// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/ResourceDraw.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/ResourceDrawNumbers.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"strconv"

	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// ResourceDraw is upstream's ResourceDraw: the drawing of a resource below
// the tasks.
type ResourceDraw interface {
	klimt.UDrawable
	Height(stringBounder font.StringBounder) float64
	Y() float64
}

// ResourceDrawNumbers is upstream's ResourceDrawNumbers: the name of a
// resource and its average load per period of the calendar.
type ResourceDrawNumbers struct {
	res                         *Resource
	timeScale                   TimeScale
	y                           float64
	min                         TimePoint
	maxTimePointPrintedEndOfDay TimePoint
	modelData                   *GanttModelData
	timelineStyle               *TimelineStyleData
}

// DrawU implements klimt.UDrawable: the name in a serif font of 13, a
// line, then the loads in a serif font of 9, red above 100.
func (r *ResourceDrawNumbers) DrawU(ug klimt.UGraphic) {
	title := textOf(r.timelineStyle.Pragma(), r.res.Name(), resourceFontConfiguration(13, color.Black))
	title.DrawU(ug)
	line := klimt.ULineH(r.timeScale.Position(r.maxTimePointPrintedEndOfDay) - r.timeScale.Position(r.min))
	ug.Apply(color.Black).Apply(klimt.UTranslateDy(title.CalculateDimension(ug.StringBounder()).Height())).Draw(line)
	startingPosition := -1.0
	totalLoad, activeDays := 0, 0
	isRed := false
	for i := r.min; i.CompareTo(r.maxTimePointPrintedEndOfDay) <= 0; i = i.Increment() {
		isBreaking := r.timeScale.IsBreaking(i)
		load := r.modelData.LoadForResource(r.res, i)
		if load > 100 {
			isRed = true
		}
		totalLoad += load
		if load > 0 {
			activeDays++
		}
		if !isBreaking {
			if startingPosition == -1 {
				startingPosition = r.timeScale.Position(i)
			}
			continue
		}
		if totalLoad > 0 {
			displayedLoad := int(jmath.Round(float64(totalLoad) / float64(activeDays)))
			c := color.Black
			if isRed {
				c = color.Red
			}
			value := textOf(r.timelineStyle.Pragma(), strconv.Itoa(displayedLoad), resourceFontConfiguration(9, c))
			if startingPosition == -1 {
				startingPosition = r.timeScale.Position(i)
			}
			endingPosition := r.timeScale.Position(i) + r.timeScale.Width(i)
			start := (startingPosition+endingPosition)/2 - value.CalculateDimension(ug.StringBounder()).Width()/2
			value.DrawU(ug.Apply(klimt.NewUTranslate(start, 16)))
		}
		startingPosition = -1
		totalLoad = 0
		activeDays = 0
		isRed = false
	}
}

func resourceFontConfiguration(size int, c color.HColor) *klimt.FontConfiguration {
	return klimt.NewFontConfiguration(font.Serif(size), c, c, nil)
}

// Height implements ResourceDraw: 32.
func (r *ResourceDrawNumbers) Height(font.StringBounder) float64 { return 16 * 2 }

// Y implements ResourceDraw.
func (r *ResourceDrawNumbers) Y() float64 { return r.y }
