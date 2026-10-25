// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/header/TimeHeaderDaily.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/header/TimeHeaderWeekly.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"strconv"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// timeHeaderDaily is upstream's TimeHeaderDaily: the months, the days of
// the week and the days of the month.
type timeHeaderDaily struct {
	timeHeaderCalendar
	fc1, fc2 *klimt.FontConfiguration
}

func (h *timeHeaderDaily) h1() float64 { return h.timelineStyle.FontSizeMonth() + 2 }

func (h *timeHeaderDaily) h2() float64 { return h.h1() + h.timelineStyle.FontSizeDay() + 2 }

func (h *timeHeaderDaily) h3() float64 { return h.h2() + h.timelineStyle.FontSizeDay() + 3 }

// TimeHeaderHeight implements TimeHeader.
func (h *timeHeaderDaily) TimeHeaderHeight(font.StringBounder) float64 { return h.h3() }

// TimeFooterHeight implements TimeHeader.
func (h *timeHeaderDaily) TimeFooterHeight(font.StringBounder) float64 {
	h1 := h.timelineStyle.FontSizeDay()
	h2 := h.timelineStyle.FontSizeDay()
	h3 := h.timelineStyle.FontSizeMonth()
	return h1 + h2 + h3 + 8
}

// FullHeaderHeight implements TimeHeader.
func (h *timeHeaderDaily) FullHeaderHeight(sb font.StringBounder) float64 {
	return h.TimeHeaderHeight(sb) + h.headerNameDayHeight()
}

// DrawTimeHeader implements TimeHeader.
func (h *timeHeaderDaily) DrawTimeHeader(ug klimt.UGraphic, totalHeightWithoutFooter float64) {
	h.drawColorsBackground(ug, totalHeightWithoutFooter)
	h.drawMonths(ug)
	h.drawTextsDayOfWeek(ug.Apply(klimt.UTranslateDy(h.h1())))
	h.drawTextDayOfMonth(ug.Apply(klimt.UTranslateDy(h.h2())))
	h.printDailyVerticalSeparators(ug, totalHeightWithoutFooter)
	h.printNamedDays(ug)
	h.drawHline(ug, h.FullHeaderHeight(ug.StringBounder()))
	h.drawHline(ug, totalHeightWithoutFooter)
}

// printDailyVerticalSeparators is the printVerticalSeparators of the daily
// header: a line before each day, in the style of the separators after
// "separator just before" or a hidden day.
func (h *timeHeaderDaily) printDailyVerticalSeparators(ug klimt.UGraphic, totalHeightWithoutFooter float64) {
	ugVerticalSeparator := h.timelineStyle.ApplyVerticalSeparatorStyle(ug)
	ugLineColor := ug.Apply(h.lineColor())
	for day := h.minDay(); day.CompareTo(h.maxDay()) <= 0; day = day.PlusDays(1) {
		wink := TimePointOfStartOfDay(day)
		if h.isBold(day) || h.timeScale.Width(wink.Decrement()) == 0 {
			h.drawVline(ugVerticalSeparator, h.timeScale.Position(wink), h.FullHeaderHeight(ug.StringBounder()),
				totalHeightWithoutFooter)
		} else {
			h.drawVline(ugLineColor, h.timeScale.Position(wink), h.FullHeaderHeight(ug.StringBounder()),
				totalHeightWithoutFooter)
		}
	}
	end := h.timeScale.Position(TimePointOfStartOfDay(h.maxDay().PlusDays(1)))
	h.drawVline(ugLineColor, end, h.FullHeaderHeight(ug.StringBounder()), totalHeightWithoutFooter)
}

// DrawTimeFooter implements TimeHeader.
func (h *timeHeaderDaily) DrawTimeFooter(ug klimt.UGraphic) {
	height := h.timelineStyle.FontSizeDay() + 2
	h.drawTextsDayOfWeek(ug)
	h.drawTextDayOfMonth(ug.Apply(klimt.UTranslateDy(height + 2)))
	h.drawMonths(ug.Apply(klimt.UTranslateDy(2*height + 3)))
}

func (h *timeHeaderDaily) drawTextsDayOfWeek(ug klimt.UGraphic) {
	for day := h.minDay(); day.CompareTo(h.maxDay()) <= 0; day = day.PlusDays(1) {
		wink := TimePointOfStartOfDay(day)
		if h.isHidden(wink) {
			continue
		}
		x1 := h.timeScale.Position(wink)
		x2 := h.timeScale.Position(wink.Increment())
		h.printCentered(ug, false, x1, x2, wink, h.fc(wink), FormatDayOfWeekShort)
	}
}

func (h *timeHeaderDaily) drawTextDayOfMonth(ug klimt.UGraphic) {
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

func (h *timeHeaderDaily) isHidden(wink TimePoint) bool {
	return h.scaleConfig.hideClosed && h.dayCalendar.OpenClose().IsClosed(wink.Day())
}

func (h *timeHeaderDaily) fc(wink TimePoint) *klimt.FontConfiguration {
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

func (h *timeHeaderDaily) drawMonths(ug klimt.UGraphic) {
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
				h.printMonth(ug, *last, lastChangeMonth, x1, fc)
			}
			lastChangeMonth = x1
			w := wink
			last = &w
		}
	}
	x1 := h.timeScale.Position(TimePointOfStartOfDay(h.maxDay().PlusDays(1)))
	if x1 > lastChangeMonth {
		h.printMonth(ug, derefTimePoint(last), lastChangeMonth, x1, fc)
	}
}

func (h *timeHeaderDaily) printMonth(ug klimt.UGraphic, monthYear TimePoint, start, end float64,
	fc *klimt.FontConfiguration) {
	h.printCentered(ug, false, start, end, monthYear, fc, FormatMonthShort, FormatMonthLong, FormatMonthYearLong)
}

// derefTimePoint is a time point that upstream may leave null, with Java's
// NullPointerException.
func derefTimePoint(t *TimePoint) TimePoint {
	if t == nil {
		panicNPE()
	}
	return *t
}

// timeHeaderWeekly is upstream's TimeHeaderWeekly: the months and the
// weeks.
type timeHeaderWeekly struct {
	timeHeaderCalendar
}

func (h *timeHeaderWeekly) h1() float64 { return h.timelineStyle.FontSizeMonth() + 4 }

func (h *timeHeaderWeekly) h2() float64 { return h.h1() + h.timelineStyle.FontSizeDay() + 1 }

// TimeHeaderHeight implements TimeHeader.
func (h *timeHeaderWeekly) TimeHeaderHeight(font.StringBounder) float64 { return h.h2() }

// TimeFooterHeight implements TimeHeader.
func (h *timeHeaderWeekly) TimeFooterHeight(font.StringBounder) float64 {
	return h.timelineStyle.FontSizeMonth() + 4
}

// FullHeaderHeight implements TimeHeader.
func (h *timeHeaderWeekly) FullHeaderHeight(sb font.StringBounder) float64 {
	return h.TimeHeaderHeight(sb) + h.headerNameDayHeight()
}

// DrawTimeHeader implements TimeHeader.
func (h *timeHeaderWeekly) DrawTimeHeader(ug klimt.UGraphic, totalHeightWithoutFooter float64) {
	h.drawColorsBackground(ug, totalHeightWithoutFooter)
	h.printDaysOfMonth(ug)
	h.printWeeklyVerticalSeparators(ug, totalHeightWithoutFooter)
	h.printMonths(ug)
	h.printNamedDays(ug)
	h.drawHline(ug, 0)
	h.drawHline(ug, h.h1())
	h.drawHline(ug, h.FullHeaderHeight(ug.StringBounder()))
}

// DrawTimeFooter implements TimeHeader.
func (h *timeHeaderWeekly) DrawTimeFooter(ug klimt.UGraphic) {
	h.drawHline(ug, 0)
	h.printMonths(ug)
	h.drawHline(ug, h.TimeFooterHeight(ug.StringBounder()))
}

func (h *timeHeaderWeekly) printMonths(ug klimt.UGraphic) {
	fc := h.fontConfigurationOf(klimt.SNameMonth, true, h.openFontColor())
	var last *TimePoint
	lastChangeMonth := -1.0
	for day := h.minDay(); day.CompareTo(h.maxDay()) <= 0; day = day.PlusDays(1) {
		wink := TimePointOfStartOfDay(day)
		x1 := h.timeScale.Position(wink)
		if last == nil || wink.MonthYear() != last.MonthYear() {
			h.drawVline(ug.Apply(h.lineColor()), x1, 0, h.h1())
			if last != nil {
				h.printMonth(ug, *last, lastChangeMonth, x1, fc)
			}
			lastChangeMonth = x1
			w := wink
			last = &w
		}
	}
	end := h.timeScale.Position(TimePointOfStartOfDay(h.maxDay().PlusDays(1)))
	h.drawVline(ug.Apply(h.lineColor()), end, 0, h.h1())
	if last != nil && end > lastChangeMonth {
		h.printMonth(ug, *last, lastChangeMonth, end, fc)
	}
}

// printWeeklyVerticalSeparators is the printVerticalSeparators of the
// weekly header: a line before each week, then the separators of
// "separator just before".
func (h *timeHeaderWeekly) printWeeklyVerticalSeparators(ug klimt.UGraphic, totalHeightWithoutFooter float64) {
	for day := h.minDay(); day.CompareTo(h.maxDay()) <= 0; day = day.PlusDays(1) {
		wink := TimePointOfStartOfDay(day)
		if wink.DayOfWeek() == h.weekConfig.weekNumberStrategy.FirstDayOfWeek {
			h.drawVline(ug.Apply(h.lineColor()), h.timeScale.Position(wink), h.h1(), totalHeightWithoutFooter)
		}
	}
	end := h.timeScale.Position(TimePointOfStartOfDay(h.maxDay().PlusDays(1)))
	h.drawVline(ug.Apply(h.lineColor()), end, h.h1(), totalHeightWithoutFooter)
	h.printVerticalSeparators(ug, totalHeightWithoutFooter)
}

// printDaysOfMonth writes the label of each week: its number of the year,
// the day of the month of its first day, or a counter. As upstream, the
// last day of the calendar is left out.
func (h *timeHeaderWeekly) printDaysOfMonth(ug klimt.UGraphic) {
	fc := h.fontConfigurationOf(klimt.SNameDay, false, h.openFontColor())
	counter := h.weekConfig.weekStartingNumber
	for day := h.minDay(); day.CompareTo(h.maxDay().PlusDays(-1)) < 0; day = day.PlusDays(1) {
		wink := TimePointOfStartOfDay(day)
		if wink.DayOfWeek() != h.weekConfig.weekNumberStrategy.FirstDayOfWeek {
			continue
		}
		var num string
		switch h.weekConfig.weeklyHeaderStrategy {
		case FromN:
			num = strconv.Itoa(counter)
			counter++
		case DayOfMonth:
			num = strconv.Itoa(wink.DayOfMonth())
		default:
			num = strconv.Itoa(wink.WeekOfYear(h.weekConfig.weekNumberStrategy))
		}
		h.textBlock(num, fc).DrawU(ug.Apply(klimt.UTranslateDy(h.h1())).
			Apply(klimt.UTranslateDx(h.timeScale.Position(wink) + 5)))
	}
}

func (h *timeHeaderWeekly) printMonth(ug klimt.UGraphic, monthYear TimePoint, start, end float64,
	fc *klimt.FontConfiguration) {
	h.printCentered(ug, false, start, end, monthYear, fc, FormatMonthShort, FormatMonthYearShort)
}
