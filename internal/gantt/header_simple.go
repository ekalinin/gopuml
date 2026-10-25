// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/header/TimeHeaderSimple.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/header/TimeHeaderVariant.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"strconv"

	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// timeHeaderSimple is upstream's TimeHeaderSimple: the numbers of the days
// or weeks of a project without a start.
type timeHeaderSimple struct {
	timeHeader
	delta int
}

func (h *timeHeaderSimple) locale() Locale { return "" }

func (h *timeHeaderSimple) isZeroOnDay(TimePoint) bool { return false }

// TimeHeaderHeight implements TimeHeader.
func (h *timeHeaderSimple) TimeHeaderHeight(font.StringBounder) float64 {
	return h.timelineStyle.FontSizeDay() + 6
}

// TimeFooterHeight implements TimeHeader.
func (h *timeHeaderSimple) TimeFooterHeight(font.StringBounder) float64 {
	return h.timelineStyle.FontSizeDay() + 6
}

// FullHeaderHeight implements TimeHeader.
func (h *timeHeaderSimple) FullHeaderHeight(sb font.StringBounder) float64 {
	return h.TimeHeaderHeight(sb)
}

// incrementPoint is increment(TimePoint): delta steps of the print scale.
func (h *timeHeaderSimple) incrementPoint(day TimePoint) TimePoint {
	if h.delta == 0 {
		h.initDelta(day)
	}
	for i := 0; i < h.delta; i++ {
		day = day.IncrementScale(h.scaleConfig.printScale)
	}
	return day
}

// incrementDay is increment(LocalDate).
func (h *timeHeaderSimple) incrementDay(day jtime.LocalDate) jtime.LocalDate {
	if h.delta == 0 {
		h.initDelta(TimePointOfStartOfDay(day))
	}
	for i := 0; i < h.delta; i++ {
		if h.scaleConfig.printScale == PrintScaleWeekly {
			day = day.PlusDays(7)
		} else {
			day = day.PlusDays(1)
		}
	}
	return day
}

// initDelta is initDelta: daily, the number of days of a step at least 16
// pixels wide; one step otherwise.
func (h *timeHeaderSimple) initDelta(day TimePoint) {
	if h.scaleConfig.printScale != PrintScaleDaily {
		h.delta = 1
		return
	}
	x1 := h.timeScale.Position(day)
	for {
		h.delta++
		day = day.Increment()
		if h.timeScale.Position(day) >= x1+16 {
			return
		}
	}
}

func (h *timeHeaderSimple) drawSmallVlinesDay(ug klimt.UGraphic, totalHeightWithoutFooter float64) {
	ug = ug.Apply(h.lineColor()).Apply(klimt.UTranslateDy(6))
	vbar := klimt.ULineV(totalHeightWithoutFooter + 2)
	for day := h.minDay(); day.CompareTo(h.maxDay().PlusDays(1)) <= 0; day = h.incrementDay(day) {
		x1 := h.timeScale.Position(TimePointOfStartOfDay(day))
		ug.Apply(klimt.UTranslateDx(x1)).Draw(vbar)
	}
}

func (h *timeHeaderSimple) drawSimpleDayCounter(ug klimt.UGraphic) {
	fc := fontConfiguration(h.timelineStyle.Font(klimt.SNameDay), false, h.openFontColor())
	for day := h.minDay(); day.CompareTo(h.maxDay().PlusDays(1)) <= 0; day = h.incrementDay(day) {
		wink := TimePointOfStartOfDay(day)
		var value int
		if h.scaleConfig.printScale == PrintScaleWeekly {
			value = wink.AbsoluteDayNum()/7 + 1
		} else {
			value = wink.AbsoluteDayNum() + 1
		}
		num := klimt.DisplayWithNewlines(klimt.PragmaCreateEmpty(), strconv.Itoa(value)).Create(fc,
			geom.HorizontalAlignmentLeft, klimt.SpriteContainerEmpty{})
		x1 := h.timeScale.Position(wink)
		var x2 float64
		if h.scaleConfig.printScale == PrintScaleWeekly {
			x2 = h.timeScale.Position(wink.AddDays(6)) + h.timeScale.Width(wink.AddDays(6))
		} else {
			x2 = h.timeScale.Position(h.incrementPoint(wink))
		}
		width := num.CalculateDimension(ug.StringBounder()).Width()
		delta := (x2 - x1) - width
		if wink.CompareTo(TimePointOfStartOfDay(h.maxDay().PlusDays(1))) < 0 {
			num.DrawU(ug.Apply(klimt.UTranslateDx(x1 + delta/2)))
		}
	}
}

// DrawTimeHeader implements TimeHeader.
func (h *timeHeaderSimple) DrawTimeHeader(ug klimt.UGraphic, totalHeightWithoutFooter float64) {
	h.drawColorsBackground(ug, totalHeightWithoutFooter)
	h.drawSmallVlinesDay(ug, totalHeightWithoutFooter)
	h.printVerticalSeparators(ug, totalHeightWithoutFooter)
	h.drawSimpleDayCounter(ug)
}

// DrawTimeFooter implements TimeHeader.
func (h *timeHeaderSimple) DrawTimeFooter(ug klimt.UGraphic) {
	h.drawSimpleDayCounter(ug.Apply(klimt.UTranslateDy(3)))
}

// timeHeaderVariant is upstream's TimeHeaderVariant: the daily calendar of
// a project with working hours, the days drawn with their hours.
type timeHeaderVariant struct {
	timeHeaderCalendar
	workingHours *WorkingHours
	fc1, fc2     *klimt.FontConfiguration
}

func (h *timeHeaderVariant) h1() float64 { return h.timelineStyle.FontSizeMonth() + 2 }

func (h *timeHeaderVariant) h2() float64 { return h.h1() + h.timelineStyle.FontSizeDay() + 2 }

func (h *timeHeaderVariant) h3() float64 { return h.h2() + h.timelineStyle.FontSizeDay() + 3 }

// TimeHeaderHeight implements TimeHeader.
func (h *timeHeaderVariant) TimeHeaderHeight(font.StringBounder) float64 { return h.h3() }

// TimeFooterHeight implements TimeHeader: none.
func (h *timeHeaderVariant) TimeFooterHeight(font.StringBounder) float64 { return 0 }

// FullHeaderHeight implements TimeHeader.
func (h *timeHeaderVariant) FullHeaderHeight(sb font.StringBounder) float64 {
	return h.TimeHeaderHeight(sb)
}

// DrawTimeFooter implements TimeHeader: nothing.
func (h *timeHeaderVariant) DrawTimeFooter(klimt.UGraphic) {}

func (h *timeHeaderVariant) drawSmallVlinesDay(ug klimt.UGraphic, totalHeightWithoutFooter float64) {
	ug = ug.Apply(h.lineColor())
	vbar := klimt.ULineV(totalHeightWithoutFooter)
	for day := h.minDay(); day.CompareTo(h.maxDay().PlusDays(1)) <= 0; day = day.PlusDays(1) {
		x1 := h.timeScale.Position(TimePointOfStartOfDay(day))
		ug.Apply(klimt.UTranslateDx(x1)).Draw(vbar)
	}
}

func (h *timeHeaderVariant) drawTextsDayOfWeek(ug klimt.UGraphic) {
	for day := h.minDay(); day.CompareTo(h.maxDay()) <= 0; day = day.PlusDays(1) {
		wink := TimePointOfStartOfDay(day)
		if h.isHidden(wink) {
			continue
		}
		x1 := h.timeScale.Position(wink)
		x2 := h.timeScale.Position(wink.Increment())
		h.printCentered(ug, false, x1, x2, wink, h.fc(wink), FormatDayOfWeekShort, FormatDayOfWeekLong)
	}
}

func (h *timeHeaderVariant) fc(wink TimePoint) *klimt.FontConfiguration {
	if h.isZeroOnDay(wink) {
		if h.fc1 == nil {
			h.fc1 = h.fontConfigurationOf(klimt.SNameDay, false, h.closedFontColor())
		}
		return h.fc1
	}
	if h.fc2 == nil {
		h.fc2 = h.fontConfigurationOf(klimt.SNameDay, false, h.openFontColor())
	}
	return h.fc2
}

func (h *timeHeaderVariant) isHidden(wink TimePoint) bool {
	return h.scaleConfig.hideClosed && h.dayCalendar.OpenClose().IsClosed(wink.Day())
}

// DrawTimeHeader implements TimeHeader.
func (h *timeHeaderVariant) DrawTimeHeader(ug klimt.UGraphic, totalHeightWithoutFooter float64) {
	h.drawColorsBackground(ug, totalHeightWithoutFooter)
	h.drawNonWorkingPeriod(ug, totalHeightWithoutFooter)
	h.drawMonths(ug)
	h.drawTextsDayOfWeek(ug.Apply(klimt.UTranslateDy(h.h1())))
	h.drawTextDayOfMonth(ug.Apply(klimt.UTranslateDy(h.h2())))
	h.drawSmallVlinesDay(ug.Apply(klimt.UTranslateDy(h.h1())), totalHeightWithoutFooter)
}

// drawNonWorkingPeriod draws the closed hours of each day, below the
// header and as high as the diagram, as upstream.
func (h *timeHeaderVariant) drawNonWorkingPeriod(ug klimt.UGraphic, height float64) {
	backColor := h.closedBackgroundColor()
	ug = ug.Apply(backColor).Apply(klimt.Bg(backColor)).Apply(klimt.UTranslateDy(h.h3()))
	for day := h.minDay(); day.CompareTo(h.maxDay()) <= 0; day = day.PlusDays(1) {
		x1 := h.timeScale.Position(TimePointOfStartOfDay(day))
		h.workingHours.DrawOneDay(ug.Apply(klimt.UTranslateDx(x1)), height)
	}
}

func (h *timeHeaderVariant) drawTextDayOfMonth(ug klimt.UGraphic) {
	for day := h.minDay(); day.CompareTo(h.maxDay()) <= 0; day = day.PlusDays(1) {
		wink := TimePointOfStartOfDay(day)
		if h.isHidden(wink) {
			continue
		}
		x1 := h.timeScale.Position(wink)
		x2 := h.timeScale.Position(wink.Increment())
		h.printCentered(ug, false, x1, x2, wink, h.fc(wink), FormatDayOfMonth)
	}
}

func (h *timeHeaderVariant) drawMonths(ug klimt.UGraphic) {
	fc := h.fontConfigurationOf(klimt.SNameMonth, true, h.openFontColor())
	var last *TimePoint
	lastChangeMonth := -1.0
	for day := h.minDay(); day.CompareTo(h.maxDay()) <= 0; day = day.PlusDays(1) {
		wink := TimePointOfStartOfDay(day)
		if h.isHidden(wink) {
			continue
		}
		x1 := h.timeScale.Position(wink)
		if last == nil || wink.MonthYear() != last.MonthYear() {
			if last != nil {
				h.printCentered(ug, false, lastChangeMonth, x1, *last, fc, FormatMonthShort, FormatMonthLong,
					FormatMonthYearLong)
			}
			lastChangeMonth = x1
			w := wink
			last = &w
		}
	}
	x1 := h.timeScale.Position(TimePointOfStartOfDay(h.maxDay().PlusDays(1)))
	if x1 > lastChangeMonth {
		h.printCentered(ug, false, lastChangeMonth, x1, derefTimePoint(last), fc, FormatMonthShort,
			FormatMonthLong, FormatMonthYearLong)
	}
}
