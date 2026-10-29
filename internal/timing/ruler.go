// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/timingdiagram/TimingRuler.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package timing

import (
	"slices"

	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// TimingRuler holds the times of the diagram and places them on the x
// axis: a tick every tickIntervalInPixels for each tick unit.
type TimingRuler struct {
	// times are sorted by time; a time is kept once, with the tick that
	// gave it first, as upstream's TreeSet.
	times     []*TimeTick
	skinParam klimt.ISkinParam

	tickIntervalInPixels int64
	forcedTickUnitary    int64
	// format is the format of the last tick added that is not DECIMAL.
	format *TimingFormat

	highestCommonFactorInternal int64
}

func newTimingRuler(skinParam klimt.ISkinParam) *TimingRuler {
	return &TimingRuler{skinParam: skinParam, tickIntervalInPixels: 50, format: formatDecimal,
		highestCommonFactorInternal: -1}
}

// applyForVLines applies the dashed stroke of the vertical lines and the
// line color of style.
func applyForVLines(ug klimt.UGraphic, style *klimt.Style, skinParam klimt.ISkinParam) klimt.UGraphic {
	stroke := klimt.NewUStroke(3, 5, 0.5)
	c := style.Value(klimt.PNameLineColor).AsColor(skinParam.IHtmlColorSet())
	return ug.Apply(stroke).Apply(c)
}

// insert adds a tick unless its time is already there, as TreeSet.add. A
// nil tick throws upstream's NullPointerException of TreeMap.put.
func (r *TimingRuler) insert(tick *TimeTick) {
	if tick == nil {
		if len(r.times) == 0 {
			uml.ThrowNPE(`Cannot invoke "java.lang.Comparable.compareTo(Object)" because "k1" is null`)
		}
		uml.Throw(uml.NullPointerException, "")
	}
	i, found := slices.BinarySearchFunc(r.times, tick, (*TimeTick).CompareTo)
	if !found {
		r.times = slices.Insert(r.times, i, tick)
	}
}

// ensureNotEmpty adds a tick at 0 to an empty ruler and to one whose times
// go from below zero to above zero.
func (r *TimingRuler) ensureNotEmpty() {
	if len(r.times) == 0 {
		r.insert(newTimeTick(jmath.BigDecimalZero, formatDecimal))
	}
	if r.getMax().time.Signum() > 0 && r.getMin().time.Signum() < 0 {
		r.insert(newTimeTick(jmath.BigDecimalZero, formatDecimal))
	}
}

// scaleInPixels forces a tick every tick time units, pixel pixels apart.
func (r *TimingRuler) scaleInPixels(tick, pixel int64) {
	if pixel <= 0 || tick <= 0 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	r.tickIntervalInPixels = pixel
	r.forcedTickUnitary = tick
}

func (r *TimingRuler) getOptimalTickUnit() int64 {
	if r.forcedTickUnitary == 0 {
		hcfTickUnit := r.calculateHighestCommonFactor()
		const maxDiagramWidth = 4000.0
		if hcfTickUnit == 1 && r.calculateDiagramWidth(hcfTickUnit) > maxDiagramWidth {
			// A unit of 1 would make the diagram too wide: the unit is
			// approximated from the width and the time range.
			totalTimeRange := r.getMax().time.DoubleValue() - r.getMin().time.DoubleValue()
			return jmath.Round(1 + (float64(r.tickIntervalInPixels) * totalTimeRange / maxDiagramWidth))
		}
		return hcfTickUnit
	}
	return r.forcedTickUnitary
}

// getWidth is the width of the axis: 100 without times.
func (r *TimingRuler) getWidth() float64 {
	if len(r.times) == 0 {
		return 100
	}
	return r.calculateDiagramWidth(r.getOptimalTickUnit())
}

func (r *TimingRuler) calculateDiagramWidth(tickUnitary int64) float64 {
	delta := r.getMax().time.DoubleValue() - r.getMin().time.DoubleValue()
	return float64((delta/float64(tickUnitary) + 1) * float64(r.tickIntervalInPixels))
}

// calculateHighestCommonFactor is the highest common factor of the times,
// truncated to integers and without zero; -1 when there is none.
func (r *TimingRuler) calculateHighestCommonFactor() int64 {
	if r.highestCommonFactorInternal == -1 {
		for _, tick := range r.getAbsolutesTicks() {
			if r.highestCommonFactorInternal == -1 {
				r.highestCommonFactorInternal = tick
			} else {
				r.highestCommonFactorInternal = computeHighestCommonFactor(r.highestCommonFactorInternal, tick)
			}
		}
	}
	return r.highestCommonFactorInternal
}

// getAbsolutesTicks returns the distinct non-zero absolute values of the
// truncated times, from the largest.
func (r *TimingRuler) getAbsolutesTicks() []int64 {
	var result []int64
	for _, t := range r.times {
		value := t.time.LongValue()
		if value < 0 {
			value = -value
		}
		if value > 0 && !slices.Contains(result, value) {
			result = append(result, value)
		}
	}
	slices.SortFunc(result, func(a, b int64) int { return -compareInt64(a, b) })
	return result
}

func compareInt64(a, b int64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

func (r *TimingRuler) getNbTick() int {
	if len(r.times) == 0 {
		return 1
	}
	delta := r.getMax().time.LongValue() - r.getMin().time.LongValue()
	return min(1000, int(int32(1+delta/r.getOptimalTickUnit())))
}

// getPosInPixel is the x of a tick.
func (r *TimingRuler) getPosInPixel(when *TimeTick) float64 {
	return r.getPosInPixelInternal(when.time.DoubleValue())
}

func (r *TimingRuler) getPosInPixelInternal(t float64) float64 {
	t -= r.getMin().time.DoubleValue()
	return float64(t / float64(r.getOptimalTickUnit()) * float64(r.tickIntervalInPixels))
}

// addTime adds a time to the axis; a tick that is not DECIMAL sets the
// format of the labels, even when its time is already there.
func (r *TimingRuler) addTime(t *TimeTick) {
	r.highestCommonFactorInternal = -1
	r.insert(t)
	if t.format != formatDecimal {
		r.format = t.format
	}
}

func (r *TimingRuler) styleTimegrid() *klimt.Style {
	return timingStyle(klimt.SNameTimegrid).MergedStyle(r.skinParam.CurrentStyleBuilder())
}

func (r *TimingRuler) styleTimeline() *klimt.Style {
	return timingStyle(klimt.SNameTimeline).MergedStyle(r.skinParam.CurrentStyleBuilder())
}

func (r *TimingRuler) timeTextBlockLong(t int64) klimt.TextBlock {
	return r.timeTextBlock(r.format.formatLong(t))
}

func (r *TimingRuler) timeTextBlock(s string) klimt.TextBlock {
	display := klimt.DisplayWithNewlines(r.skinParam.Pragma(), s)
	fontConfiguration := klimt.FontConfigurationFromStyle(r.skinParam, r.styleTimeline(), nil)
	return display.Create(fontConfiguration, geom.HorizontalAlignmentLeft, r.skinParam)
}

// drawTimeAxis draws the axis, its ticks and their labels; codes are the
// named times, in the iteration order of upstream's HashMap.
func (r *TimingRuler) drawTimeAxis(ug klimt.UGraphic, timeAxisStategy TimeAxisStategy, codes []namedTime) {
	if timeAxisStategy == TimeAxisHidden {
		return
	}
	styleTimeline := r.styleTimeline()
	c := styleTimeline.Value(klimt.PNameLineColor).AsColor(r.skinParam.IHtmlColorSet())
	stroke := styleTimeline.Stroke()
	ug = ug.Apply(stroke).Apply(c)
	if timeAxisStategy == TimeAxisAutomatic {
		r.drawTimeAxisAutomatic(ug)
	} else {
		r.drawTimeAxisManual(ug, codes)
	}
}

func (r *TimingRuler) drawTimeAxisManual(ug klimt.UGraphic, codes []namedTime) {
	const tickHeight = 5
	line := klimt.ULineV(tickHeight)
	firstTickPosition := r.getPosInPixelInternal(r.getFirstPositiveOrZeroValue().DoubleValue())
	nb := 0
	for firstTickPosition+float64(int64(nb)*r.tickIntervalInPixels) <= r.getWidth() {
		nb++
	}
	ug.Apply(klimt.UTranslateDx(firstTickPosition)).Draw(klimt.ULineH(float64(int64(nb-1) * r.tickIntervalInPixels)))
	for _, tick := range r.times {
		ug.Apply(klimt.UTranslateDx(r.getPosInPixel(tick))).Draw(line)
		label := r.getLabel(tick, codes)
		if label == "" {
			continue
		}
		text := r.timeTextBlock(label)
		dim := text.CalculateDimension(ug.StringBounder())
		text.DrawU(ug.Apply(klimt.NewUTranslate(r.getPosInPixel(tick)-dim.Width()/2, tickHeight+1)))
	}
}

func (r *TimingRuler) drawTimeAxisAutomatic(ug klimt.UGraphic) {
	const tickHeight = 5
	line := klimt.ULineV(tickHeight)
	firstTickPosition := r.getPosInPixelInternal(r.getFirstPositiveOrZeroValue().DoubleValue())
	nb := 0
	for firstTickPosition+float64(int64(nb)*r.tickIntervalInPixels) <= r.getWidth() {
		ug.Apply(klimt.UTranslateDx(firstTickPosition + float64(int64(nb)*r.tickIntervalInPixels))).Draw(line)
		nb++
	}
	ug.Apply(klimt.UTranslateDx(firstTickPosition)).Draw(klimt.ULineH(float64(int64(nb-1) * r.tickIntervalInPixels)))
	for _, round := range r.roundValues() {
		text := r.timeTextBlockLong(round)
		dim := text.CalculateDimension(ug.StringBounder())
		text.DrawU(ug.Apply(klimt.NewUTranslate(r.getPosInPixelInternal(float64(round))-dim.Width()/2, tickHeight+1)))
	}
}

// getLabel is the name of a tick when a named time is that same tick, or
// its formatted time.
func (r *TimingRuler) getLabel(tick *TimeTick, codes []namedTime) string {
	for _, code := range codes {
		if code.tick == tick {
			return code.name
		}
	}
	return r.format.formatTime(tick.time)
}

// getFirstPositiveOrZeroValue is the first time at or above zero; it throws
// IllegalStateException when every time is below zero.
func (r *TimingRuler) getFirstPositiveOrZeroValue() jmath.BigDecimal {
	for _, t := range r.times {
		if t.time.Signum() >= 0 {
			return t.time
		}
	}
	uml.Throw(uml.IllegalStateException, "")
	return jmath.BigDecimal{}
}

// roundValues are the labeled times of the automatic axis: the truncated
// times, or every forced tick, with 0 when they go from below zero to
// above zero.
func (r *TimingRuler) roundValues() []int64 {
	var result []int64
	add := func(v int64) {
		if !slices.Contains(result, v) {
			result = append(result, v)
		}
	}
	if r.forcedTickUnitary == 0 {
		for _, tick := range r.times {
			add(tick.time.LongValue())
		}
	} else {
		nb := r.getNbTick()
		for i := 0; i <= nb; i++ {
			add(r.tickToTime(i))
		}
	}
	slices.Sort(result)
	if result[0] < 0 && result[len(result)-1] > 0 {
		add(0)
		slices.Sort(result)
	}
	return result
}

func (r *TimingRuler) tickToTime(i int) int64 {
	return r.forcedTickUnitary*int64(i) + r.getMin().time.LongValue()
}

// drawVlines draws the dashed vertical lines of the ticks.
func (r *TimingRuler) drawVlines(ug klimt.UGraphic, height float64) {
	ug = applyForVLines(ug, r.styleTimegrid(), r.skinParam)
	line := klimt.ULineV(height)
	nb := r.getNbTick()
	for i := 0; i <= nb; i++ {
		ug.Apply(klimt.UTranslateDx(float64(r.tickIntervalInPixels * int64(i)))).Draw(line)
	}
}

// getHeight is the height of the labels of the axis.
func (r *TimingRuler) getHeight(stringBounder font.StringBounder) float64 {
	return r.timeTextBlockLong(0).CalculateDimension(stringBounder).Height()
}

// getMax and getMin are the last and the first tick; an empty ruler throws
// NoSuchElementException, as TreeSet.last and first.
func (r *TimingRuler) getMax() *TimeTick {
	if len(r.times) == 0 {
		uml.Throw(uml.NoSuchElementException, "")
	}
	return r.times[len(r.times)-1]
}

func (r *TimingRuler) getMin() *TimeTick {
	if len(r.times) == 0 {
		uml.Throw(uml.NoSuchElementException, "")
	}
	return r.times[0]
}

func computeHighestCommonFactor(a, b int64) int64 {
	rem := a
	for rem != 0 {
		rem = a % b
		a = b
		b = rem
	}
	if a < 0 {
		return -a
	}
	return a
}
