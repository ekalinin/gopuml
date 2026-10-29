// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/PlayerAnalog.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/TimeSeries.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"slices"

	"github.com/ekalinin/gopuml/internal/jfmt"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/uml"
)

// playerAnalog is an analog player: a line through its values.
type playerAnalog struct {
	playerBase
	timeSeries  *timeSeries
	constraints []*timeConstraint
	// initialState and ticksEvery are nil when they are not given.
	initialState *float64
	ticksEvery   *int
}

func newPlayerAnalog(code string, skinParam klimt.ISkinParam, ruler *TimingRuler, compact bool,
	stereotype *klimt.Stereotype) *playerAnalog {
	return &playerAnalog{playerBase: newPlayerBase(code, skinParam, ruler, compact, stereotype, nil,
		klimt.SNameAnalog, 100), timeSeries: &timeSeries{}}
}

func (p *playerAnalog) buildPlayerPanels() panels {
	return newPanelsAnalog(p.ruler, p.skinParam, p.suggestedHeight, p.style(), p.timeSeries, p.constraints,
		p.initialState, p.ticksEvery)
}

func (p *playerAnalog) defineState(string, string) { unsupported() }

// setState reads the value as Double.parseDouble, 0 when it is not a
// number; the first value is also the initial one when none was given.
func (p *playerAnalog) setState(now *TimeTick, _ *string, _ *color.Colors, valueString ...string) {
	value, err := jfmt.ParseDouble(valueString[0])
	if err != nil {
		value = 0
	}
	if now == nil {
		p.initialState = &value
	} else {
		p.timeSeries.put(now, value)
	}
	if p.initialState == nil {
		p.initialState = &value
	}
}

func (p *playerAnalog) createConstraint(tick1, tick2 *TimeTick, message string, hasMessage bool,
	config *klimt.ArrowConfiguration) {
	p.constraints = append(p.constraints, newTimeConstraint(1, tick1, tick2, message, hasMessage, p.skinParam,
		config))
}

func (p *playerAnalog) setBounds(min, max string) { p.timeSeries.setBounds(min, max) }

func (p *playerAnalog) setTicks(ticksEvery int) { p.ticksEvery = &ticksEvery }

// timeValue is a value of an analog player at a time; the tick of the
// first value given at that time is kept with the last value, as upstream's
// TreeMap.put.
type timeValue struct {
	tick  *TimeTick
	value float64
}

// timeSeries is upstream's TimeSeries: the values of an analog player in
// time order, and its bounds.
type timeSeries struct {
	values      []timeValue
	minOverride *float64
	maxOverride *float64
	format      *jfmt.DecimalFormat
}

func (s *timeSeries) put(now *TimeTick, value float64) {
	i, found := slices.BinarySearchFunc(s.values, now, func(v timeValue, t *TimeTick) int { return v.tick.CompareTo(t) })
	if found {
		s.values[i].value = value
		return
	}
	s.values = slices.Insert(s.values, i, timeValue{tick: now, value: value})
}

// get returns the value at a time, ok false when there is none.
func (s *timeSeries) get(tick *TimeTick) (float64, bool) {
	i, found := slices.BinarySearchFunc(s.values, tick, func(v timeValue, t *TimeTick) int { return v.tick.CompareTo(t) })
	if !found {
		return 0, false
	}
	return s.values[i].value, true
}

// getMin is the lower bound, or the smallest of 0 and the values.
func (s *timeSeries) getMin() float64 {
	if s.minOverride != nil {
		return *s.minOverride
	}
	m := 0.0
	for _, v := range s.values {
		m = min(m, v.value)
	}
	return m
}

// getMax is the upper bound, or the largest of 0 and the values, 10 when
// it is 0.
func (s *timeSeries) getMax() float64 {
	if s.maxOverride != nil {
		return *s.maxOverride
	}
	m := 0.0
	for _, v := range s.values {
		m = max(m, v.value)
	}
	if m == 0 {
		return 10
	}
	return m
}

// setBounds sets the bounds, and the format of the labels from the
// fraction digits of the bounds as written.
func (s *timeSeries) setBounds(minText, maxText string) {
	minValue := parseDouble(minText)
	maxValue := parseDouble(maxText)
	s.minOverride = &minValue
	s.maxOverride = &maxValue
	format1 := deduceFormatFrom(minText)
	format2 := deduceFormatFrom(maxText)
	s.format = format1.mergeWith(format2).decimalFormat()
}

// parseDouble is Double.parseDouble, which throws NumberFormatException on
// a string that is not a number.
func parseDouble(s string) float64 {
	v, err := jfmt.ParseDouble(s)
	if err != nil {
		uml.Throw(uml.NumberFormatException, err.Error())
	}
	return v
}

// displayValue is a value in the format of the bounds, or as Java's
// Double.toString without bounds.
func (s *timeSeries) displayValue(value float64) string {
	if s.format != nil {
		return s.format.Format(value)
	}
	return jfmt.DoubleToString(value)
}

// valueAt is the value at a time: the value given at that time, or the
// value interpolated between the values around it.
func (s *timeSeries) valueAt(tick *TimeTick) float64 {
	if result, ok := s.get(tick); ok {
		return result
	}
	var last *timeValue
	for i := range s.values {
		ent := &s.values[i]
		if ent.tick.CompareTo(tick) > 0 {
			v2 := ent.value
			if last == nil {
				return v2
			}
			t2 := ent.tick.time.DoubleValue()
			v1 := last.value
			t1 := last.tick.time.DoubleValue()
			p := (tick.time.DoubleValue() - t1) / (t2 - t1)
			return v1 + float64((v2-v1)*p)
		}
		last = ent
	}
	if last == nil {
		uml.ThrowNPE(`Cannot invoke "java.util.Map$Entry.getValue()" because "last" is null`)
	}
	return last.value
}
