// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/header/TimeHeaderMonthly.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/header/TimeHeaderQuarterly.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/header/TimeHeaderYearly.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// periods draws the vertical lines and the labels of runs of days, for
// the years, months and quarters of the compressed calendars. inclusive
// walks up to the last day, as the yearly header does; the other headers
// stop the day before. same says whether two days are in the same run.
func (h *timeHeaderCalendar) periods(ug klimt.UGraphic, lineHeight float64, inclusive bool,
	same func(a, b TimePoint) bool, print func(when TimePoint, start, end float64)) {
	var last *TimePoint
	lastChange := -1.0
	for day := h.minDay(); ; day = day.PlusDays(1) {
		if c := day.CompareTo(h.maxDay()); c > 0 || (c == 0 && !inclusive) {
			break
		}
		wink := TimePointOfStartOfDay(day)
		x1 := h.timeScale.Position(wink)
		if last == nil || !same(wink, *last) {
			h.drawVline(ug.Apply(h.lineColor()), x1, 0, lineHeight)
			if last != nil {
				print(*last, lastChange, x1)
			}
			lastChange = x1
			w := wink
			last = &w
		}
	}
	x1 := h.timeScale.Position(TimePointOfStartOfDay(h.maxDay().PlusDays(1)))
	if x1 > lastChange {
		print(derefTimePoint(last), lastChange, x1)
	}
	h.drawVline(ug.Apply(h.lineColor()), x1, 0, lineHeight)
}

func sameYear(a, b TimePoint) bool { return a.MonthYear().Year == b.MonthYear().Year }

func sameMonth(a, b TimePoint) bool { return a.MonthYear() == b.MonthYear() }

func sameQuarter(a, b TimePoint) bool { return a.Quarter() == b.Quarter() }

// timeHeaderMonthly is upstream's TimeHeaderMonthly: the years and the
// months.
type timeHeaderMonthly struct {
	timeHeaderCalendar
}

func (h *timeHeaderMonthly) h1() float64 { return h.timelineStyle.FontSizeYear() + 2 }

func (h *timeHeaderMonthly) h2() float64 { return h.h1() + h.timelineStyle.FontSizeMonth() + 2 }

// TimeHeaderHeight implements TimeHeader.
func (h *timeHeaderMonthly) TimeHeaderHeight(font.StringBounder) float64 { return h.h2() }

// TimeFooterHeight implements TimeHeader.
func (h *timeHeaderMonthly) TimeFooterHeight(font.StringBounder) float64 {
	return h.timelineStyle.FontSizeYear() + h.timelineStyle.FontSizeMonth() + 5
}

// FullHeaderHeight implements TimeHeader.
func (h *timeHeaderMonthly) FullHeaderHeight(sb font.StringBounder) float64 {
	return h.TimeHeaderHeight(sb) + h.headerNameDayHeight()
}

// DrawTimeHeader implements TimeHeader.
func (h *timeHeaderMonthly) DrawTimeHeader(ug klimt.UGraphic, totalHeightWithoutFooter float64) {
	h.drawColorsBackground(ug, totalHeightWithoutFooter)
	h.drawYears(ug)
	h1, h2 := h.h1(), h.h2()
	h.drawMonths(ug.Apply(klimt.UTranslateDy(h1)))
	h.printVerticalSeparators(ug, totalHeightWithoutFooter)
	h.printNamedDays(ug)
	h.drawHline(ug, 0)
	h.drawHline(ug, h1)
	h.drawHline(ug, h2)
}

// DrawTimeFooter implements TimeHeader.
func (h *timeHeaderMonthly) DrawTimeFooter(ug klimt.UGraphic) {
	h1 := h.timelineStyle.FontSizeYear()
	h2 := h.timelineStyle.FontSizeMonth()
	h.drawMonths(ug)
	h.drawYears(ug.Apply(klimt.UTranslateDy(h2 + 2)))
	h.drawHline(ug, 0)
	h.drawHline(ug, h2+2)
	h.drawHline(ug, h1+2+h2+2)
}

func (h *timeHeaderMonthly) drawYears(ug klimt.UGraphic) {
	fc := h.fontConfigurationOf(klimt.SNameMonth, true, h.openFontColor())
	h.periods(ug, h.timelineStyle.FontSizeYear()+2, false, sameYear, func(when TimePoint, start, end float64) {
		h.printCentered(ug, false, start, end, when, fc, FormatYear)
	})
}

func (h *timeHeaderMonthly) drawMonths(ug klimt.UGraphic) {
	fc := h.fontConfigurationOf(klimt.SNameDay, false, h.openFontColor())
	h.periods(ug, h.timelineStyle.FontSizeMonth()+2, false, sameMonth, func(when TimePoint, start, end float64) {
		h.printCentered(ug, false, start, end, when, fc, FormatMonthShort, FormatMonthLong)
	})
}

// timeHeaderQuarterly is upstream's TimeHeaderQuarterly: the years and the
// quarters.
type timeHeaderQuarterly struct {
	timeHeaderCalendar
}

// TimeHeaderHeight implements TimeHeader.
func (h *timeHeaderQuarterly) TimeHeaderHeight(font.StringBounder) float64 {
	return h.timelineStyle.FontSizeYear() + h.timelineStyle.FontSizeMonth() + 5
}

// TimeFooterHeight implements TimeHeader.
func (h *timeHeaderQuarterly) TimeFooterHeight(font.StringBounder) float64 {
	return h.timelineStyle.FontSizeYear() + h.timelineStyle.FontSizeMonth() + 5
}

// FullHeaderHeight implements TimeHeader.
func (h *timeHeaderQuarterly) FullHeaderHeight(sb font.StringBounder) float64 {
	return h.TimeHeaderHeight(sb)
}

// DrawTimeHeader implements TimeHeader.
func (h *timeHeaderQuarterly) DrawTimeHeader(ug klimt.UGraphic, totalHeightWithoutFooter float64) {
	h.drawColorsBackground(ug, totalHeightWithoutFooter)
	h.drawYears(ug)
	h1 := h.timelineStyle.FontSizeYear()
	h2 := h.timelineStyle.FontSizeMonth()
	h.drawQuarters(ug.Apply(klimt.UTranslateDy(h1 + 2)))
	h.printVerticalSeparators(ug, totalHeightWithoutFooter)
	h.drawHline(ug, 0)
	h.drawHline(ug, h1+2)
	h.drawHline(ug, h1+2+h2+2)
}

// DrawTimeFooter implements TimeHeader.
func (h *timeHeaderQuarterly) DrawTimeFooter(ug klimt.UGraphic) {
	h1 := h.timelineStyle.FontSizeYear()
	h2 := h.timelineStyle.FontSizeMonth()
	h.drawQuarters(ug)
	h.drawYears(ug.Apply(klimt.UTranslateDy(h2 + 2)))
	h.drawHline(ug, 0)
	h.drawHline(ug, h2+2)
	h.drawHline(ug, h1+2+h2+2)
}

func (h *timeHeaderQuarterly) drawYears(ug klimt.UGraphic) {
	fc := h.fontConfigurationOf(klimt.SNameMonth, true, h.openFontColor())
	h.periods(ug, h.timelineStyle.FontSizeYear()+2, false, sameYear, func(when TimePoint, start, end float64) {
		h.printCentered(ug, false, start, end, when, fc, FormatYear)
	})
}

func (h *timeHeaderQuarterly) drawQuarters(ug klimt.UGraphic) {
	fc := h.fontConfigurationOf(klimt.SNameDay, false, h.openFontColor())
	h.periods(ug, h.timelineStyle.FontSizeMonth()+2, false, sameQuarter, func(when TimePoint, start, end float64) {
		h.printCentered(ug, false, start, end, when, fc, FormatQuarter)
	})
}

// timeHeaderYearly is upstream's TimeHeaderYearly: the years, whose labels
// are left out when they do not fit.
type timeHeaderYearly struct {
	timeHeaderCalendar
}

// TimeHeaderHeight implements TimeHeader.
func (h *timeHeaderYearly) TimeHeaderHeight(font.StringBounder) float64 {
	return h.timelineStyle.FontSizeYear() + 3
}

// TimeFooterHeight implements TimeHeader.
func (h *timeHeaderYearly) TimeFooterHeight(font.StringBounder) float64 {
	return h.timelineStyle.FontSizeYear() + 3
}

// FullHeaderHeight implements TimeHeader.
func (h *timeHeaderYearly) FullHeaderHeight(sb font.StringBounder) float64 {
	return h.TimeHeaderHeight(sb)
}

// DrawTimeHeader implements TimeHeader.
func (h *timeHeaderYearly) DrawTimeHeader(ug klimt.UGraphic, totalHeightWithoutFooter float64) {
	h.drawColorsBackground(ug, totalHeightWithoutFooter)
	h.drawYears(ug)
	h.printVerticalSeparators(ug, totalHeightWithoutFooter)
	h.drawHline(ug, 0)
	h.drawHline(ug, h.FullHeaderHeight(ug.StringBounder()))
}

// DrawTimeFooter implements TimeHeader.
func (h *timeHeaderYearly) DrawTimeFooter(ug klimt.UGraphic) {
	h.drawYears(ug)
	h.drawHline(ug, 0)
	h.drawHline(ug, h.TimeFooterHeight(ug.StringBounder()))
}

func (h *timeHeaderYearly) drawYears(ug klimt.UGraphic) {
	fc := h.fontConfigurationOf(klimt.SNameYear, true, h.openFontColor())
	h.periods(ug, h.timelineStyle.FontSizeYear()+2, true, sameYear, func(when TimePoint, start, end float64) {
		h.printCentered(ug, true, start, end, when, fc, FormatYear)
	})
}
