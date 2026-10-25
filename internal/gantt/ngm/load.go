// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/ngm/Load.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/ngm/NGMAllocation.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/ngm/NGMTask.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/ngm/NGMTaskFixedTotalEffort.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/ngm/math/LoadIntegrator.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package ngm

import "github.com/ekalinin/gopuml/internal/jtime"

// Load is upstream's Load: an amount of work, as a duration of full-time
// work.
type Load struct {
	duration jtime.Duration
}

// LoadOf is Load.of, which fails with Java's IllegalArgumentException for a
// negative duration.
func LoadOf(duration jtime.Duration) Load {
	if duration.IsNegative() {
		throw(IllegalArgumentException, "Effort cannot be negative: "+duration.String())
	}
	return Load{duration}
}

// LoadZero is Load.zero.
func LoadZero() Load { return Load{jtime.ZeroDuration} }

// Seconds is toSeconds: the whole seconds of the work, through its
// milliseconds.
func (l Load) Seconds() Fraction { return NewFraction(l.duration.ToMillis()/1000, 1) }

// String is toString: the duration in ISO-8601 format.
func (l Load) String() string { return l.duration.String() }

// Allocation is upstream's NGMAllocation: the share of full-time work that
// can be done at each instant.
type Allocation struct {
	loadFunction PiecewiseConstant
}

// AllocationOf is NGMAllocation.of.
func AllocationOf(loadFunction PiecewiseConstant) Allocation { return Allocation{loadFunction} }

// AllocationFullTime is NGMAllocation.fullTime: full-time work every day.
func AllocationFullTime() Allocation { return Allocation{WeekdayOf(One)} }

// LoadFunction is getLoadFunction.
func (a Allocation) LoadFunction() PiecewiseConstant { return a.loadFunction }

// TaskFixedTotalEffort is upstream's NGMTaskFixedTotalEffort, the only kind
// of NGMTask that gantt diagrams create: a task with a fixed amount of work,
// whose end follows its start, or whose start follows its end.
type TaskFixedTotalEffort struct {
	allocation  Allocation
	totalEffort Load
	start, end  jtime.LocalDateTime
}

// TaskWithFixedTotalEffort is NGMTask.withFixedTotalEffort: a task that
// starts at start, with its end computed from the work.
func TaskWithFixedTotalEffort(allocation Allocation, start jtime.LocalDateTime, totalEffort Load) *TaskFixedTotalEffort {
	t := &TaskFixedTotalEffort{allocation: allocation, totalEffort: totalEffort, start: start}
	t.end = NewLoadIntegrator(allocation.loadFunction, totalEffort).ComputeEnd(start)
	return t
}

// Allocation is getAllocation.
func (t *TaskFixedTotalEffort) Allocation() Allocation { return t.allocation }

// Start is getStart.
func (t *TaskFixedTotalEffort) Start() jtime.LocalDateTime { return t.start }

// SetStart is setStart: the end follows.
func (t *TaskFixedTotalEffort) SetStart(start jtime.LocalDateTime) {
	t.start = start
	t.end = NewLoadIntegrator(t.allocation.loadFunction, t.totalEffort).ComputeEnd(start)
}

// End is getEnd.
func (t *TaskFixedTotalEffort) End() jtime.LocalDateTime { return t.end }

// SetEnd is setEnd: the start follows.
func (t *TaskFixedTotalEffort) SetEnd(end jtime.LocalDateTime) {
	t.end = end
	t.start = NewLoadIntegrator(t.allocation.loadFunction, t.totalEffort).ComputeStart(end)
}

// SetEffort is setEffort: the end follows.
func (t *TaskFixedTotalEffort) SetEffort(effort Load) {
	t.totalEffort = effort
	t.end = NewLoadIntegrator(t.allocation.loadFunction, t.totalEffort).ComputeEnd(t.start)
}

// Duration is getDuration: the time from the start to the end.
func (t *TaskFixedTotalEffort) Duration() jtime.Duration { return jtime.Between(t.start, t.end) }

// TotalEffort is getTotalEffort.
func (t *TaskFixedTotalEffort) TotalEffort() Load { return t.totalEffort }

// LoadIntegrator is upstream's LoadIntegrator: it walks a load function
// until a load is done.
type LoadIntegrator struct {
	loadFunction PiecewiseConstant
	totalLoad    Load
}

// NewLoadIntegrator is new LoadIntegrator.
func NewLoadIntegrator(loadFunction PiecewiseConstant, totalLoad Load) *LoadIntegrator {
	return &LoadIntegrator{loadFunction, totalLoad}
}

// ComputeEnd is computeEnd: the instant when the load is done, starting at
// start.
func (li *LoadIntegrator) ComputeEnd(start jtime.LocalDateTime) jtime.LocalDateTime {
	return li.integrate(start, Forward)
}

// ComputeStart is computeStart: the instant to start at for the load to be
// done at end.
func (li *LoadIntegrator) ComputeStart(end jtime.LocalDateTime) jtime.LocalDateTime {
	return li.integrate(end, Backward)
}

func (li *LoadIntegrator) integrate(startTime jtime.LocalDateTime, direction TimeDirection) jtime.LocalDateTime {
	remainingLoad := li.totalLoad.Seconds()
	currentTime := startTime
	it := IterateSegmentsFrom(li.loadFunction, startTime, direction)
	for it.HasNext() {
		segment := it.Next()
		if remainingLoad == Zero {
			break
		}
		loadRate := segment.Value()
		if loadRate == Zero {
			currentTime = segment.EndExclusive()
			continue
		}
		effectiveStart := segment.ClampedStart(currentTime)
		segmentSeconds := jtime.Between(effectiveStart, segment.EndExclusive()).Seconds()
		if segmentSeconds < 0 {
			segmentSeconds = -segmentSeconds
		}
		segmentLoad := loadRate.MultiplyByLong(segmentSeconds)
		if segmentLoad.CompareTo(remainingLoad) >= 0 {
			secondsNeeded := remainingLoad.Divide(loadRate)
			return direction.AdjustBySeconds(effectiveStart, secondsNeeded.WholePart())
		}
		remainingLoad = remainingLoad.Subtract(segmentLoad)
		currentTime = segment.EndExclusive()
	}
	return currentTime
}
