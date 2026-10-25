// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/GanttStyle.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/data/TimelineStyleData.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/header/TimeHeaderFactory.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/header/TimeHeader.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/draw/header/TimeHeaderCalendar.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"math"

	"github.com/ekalinin/gopuml/internal/gantt/ngm"
	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// GanttStyle is upstream's GanttStyle: the styles of a gantt diagram, of
// the time of the drawing.
type GanttStyle interface {
	Style(params ...klimt.SName) *klimt.Style
}

// Style is getStyle: the merged style of ganttDiagram and params, with the
// styles of the end of the diagram.
func (d *GanttDiagram) Style(params ...klimt.SName) *klimt.Style {
	names := append([]klimt.SName{klimt.SNameRoot, klimt.SNameElement, klimt.SNameGanttDiagram}, params...)
	return klimt.StyleSignatureBasicOf(names...).MergedStyle(d.CurrentStyleBuilder())
}

// TimelineStyleData is upstream's TimelineStyleData: the styles of the
// calendar.
type TimelineStyleData struct {
	ganttStyle GanttStyle
	colorSet   *color.HColorSet
	skinParam  klimt.ISkinParam
}

// Pragma is getPragma.
func (t *TimelineStyleData) Pragma() *klimt.Pragma { return t.skinParam.Pragma() }

// FontSizeDay is getFontSizeDay.
func (t *TimelineStyleData) FontSizeDay() float64 {
	return t.styleDay().Value(klimt.PNameFontSize).AsDouble()
}

// FontSizeMonth is getFontSizeMonth.
func (t *TimelineStyleData) FontSizeMonth() float64 {
	return t.ganttStyle.Style(klimt.SNameTimeline, klimt.SNameMonth).Value(klimt.PNameFontSize).AsDouble()
}

// FontSizeYear is getFontSizeYear.
func (t *TimelineStyleData) FontSizeYear() float64 {
	return t.ganttStyle.Style(klimt.SNameTimeline, klimt.SNameYear).Value(klimt.PNameFontSize).AsDouble()
}

// Font is getFont.
func (t *TimelineStyleData) Font(param klimt.SName) font.UFont {
	return t.ganttStyle.Style(klimt.SNameTimeline, param).UFont()
}

// ClosedBackgroundColor is getClosedBackgroundColor.
func (t *TimelineStyleData) ClosedBackgroundColor() color.HColor {
	return t.ganttStyle.Style(klimt.SNameClosed).Value(klimt.PNameBackGroundColor).AsColor(t.colorSet)
}

// ClosedFontColor is getClosedFontColor.
func (t *TimelineStyleData) ClosedFontColor() color.HColor {
	return t.ganttStyle.Style(klimt.SNameClosed).Value(klimt.PNameFontColor).AsColor(t.colorSet)
}

// OpenFontColor is getOpenFontColor.
func (t *TimelineStyleData) OpenFontColor() color.HColor {
	return t.ganttStyle.Style(klimt.SNameTimeline).Value(klimt.PNameFontColor).AsColor(t.colorSet)
}

// LineColor is getLineColor.
func (t *TimelineStyleData) LineColor() color.HColor {
	return t.ganttStyle.Style(klimt.SNameTimeline).Value(klimt.PNameLineColor).AsColor(t.colorSet)
}

// ColorSet is getColorSet.
func (t *TimelineStyleData) ColorSet() *color.HColorSet { return t.colorSet }

// ApplyVerticalSeparatorStyle is applyVerticalSeparatorStyle.
func (t *TimelineStyleData) ApplyVerticalSeparatorStyle(ug klimt.UGraphic) klimt.UGraphic {
	style := t.ganttStyle.Style(klimt.SNameVerticalSeparator)
	return ug.Apply(style.Value(klimt.PNameLineColor).AsColor(t.colorSet)).Apply(style.Stroke())
}

// CellWidth is getCellWidth: the font size of the days times 1.6.
func (t *TimelineStyleData) CellWidth() float64 {
	w := t.styleDay().Value(klimt.PNameFontSize).AsDouble()
	return w * 1.6
}

func (t *TimelineStyleData) styleDay() *klimt.Style {
	return t.ganttStyle.Style(klimt.SNameTimeline, klimt.SNameDay)
}

// TimeHeader is upstream's TimeHeader: the calendar above and below the
// tasks.
type TimeHeader interface {
	TimeHeaderHeight(stringBounder font.StringBounder) float64
	TimeFooterHeight(stringBounder font.StringBounder) float64
	FullHeaderHeight(stringBounder font.StringBounder) float64
	DrawTimeHeader(ug klimt.UGraphic, totalHeightWithoutFooter float64)
	DrawTimeFooter(ug klimt.UGraphic)
	TimeScale() TimeScale
}

// headerSelf are the methods of a header that the shared code of
// TimeHeader calls.
type headerSelf interface {
	TimeHeaderHeight(stringBounder font.StringBounder) float64
	FullHeaderHeight(stringBounder font.StringBounder) float64
	isZeroOnDay(instant TimePoint) bool
	locale() Locale
}

// timeHeader is the shared part of upstream's TimeHeader.
type timeHeader struct {
	self          headerSelf
	timeScale     TimeScale
	dayCalendar   *DayCalendarData
	timeBounds    *TimeBoundsData
	scaleConfig   *TimeScaleConfigData
	timelineStyle *TimelineStyleData
	weekConfig    *WeekConfigData
}

// TimeScale is getTimeScale.
func (h *timeHeader) TimeScale() TimeScale { return h.timeScale }

func (h *timeHeader) colorOfDay(wink TimePoint) color.HColor { return h.dayCalendar.DayColor(wink) }

func (h *timeHeader) colorOfDayOfWeek(dayOfWeek jtime.DayOfWeek) color.HColor {
	return h.dayCalendar.DayOfWeekColor(dayOfWeek)
}

func (h *timeHeader) isBold(wink jtime.LocalDate) bool { return h.dayCalendar.HasSeparatorBefore(wink) }

func (h *timeHeader) minDay() jtime.LocalDate { return h.timeBounds.MinDay() }

func (h *timeHeader) maxDay() jtime.LocalDate { return h.timeBounds.MaxDay() }

func (h *timeHeader) closedBackgroundColor() color.HColor {
	return h.timelineStyle.ClosedBackgroundColor()
}

func (h *timeHeader) closedFontColor() color.HColor { return h.timelineStyle.ClosedFontColor() }

func (h *timeHeader) openFontColor() color.HColor { return h.timelineStyle.OpenFontColor() }

func (h *timeHeader) lineColor() color.HColor { return h.timelineStyle.LineColor() }

// drawHline is drawHline: a line across the calendar at y.
func (h *timeHeader) drawHline(ug klimt.UGraphic, y float64) {
	xmin := h.timeScale.Position(TimePointOfStartOfDay(h.timeBounds.MinDay()))
	xmax := h.timeScale.Position(TimePointOfStartOfDay(h.timeBounds.MaxDay()).AddDays(1))
	ug.Apply(h.lineColor()).Apply(klimt.UTranslateDy(y)).Draw(klimt.ULineH(xmax - xmin))
}

// drawVline is drawVline.
func (h *timeHeader) drawVline(ug klimt.UGraphic, x, y1, y2 float64) {
	ug.Apply(klimt.NewUTranslate(x, y1)).Draw(klimt.ULineV(y2 - y1))
}

func fontConfiguration(f font.UFont, bold bool, c color.HColor) *klimt.FontConfiguration {
	if bold {
		f = f.WithFontFace(font.BoldFace())
	}
	return klimt.NewFontConfiguration(f, c, c, nil)
}

// drawColorsBackground is drawColorsBackground: the colored and closed
// days, merged when the same color follows.
func (h *timeHeader) drawColorsBackground(ug klimt.UGraphic, totalHeightWithoutFooter float64) {
	height := totalHeightWithoutFooter - h.self.FullHeaderHeight(ug.StringBounder())
	var p *pending
	for day := h.minDay(); day.CompareTo(h.maxDay()) <= 0; day = day.PlusDays(1) {
		wink := TimePointOfStartOfDay(day)
		x1 := h.timeScale.Position(wink)
		x2 := h.timeScale.Position(wink) + h.timeScale.Width(wink)
		back := h.colorOfDay(wink)
		// Day of week should be stronger than period of time (back color).
		if backDoW := h.colorOfDayOfWeek(wink.DayOfWeek()); backDoW != nil {
			back = backDoW
		}
		if back == nil && h.self.isZeroOnDay(wink) {
			back = h.closedBackgroundColor()
		}
		if back == nil {
			if p != nil {
				h.drawPending(ug, p, height)
			}
			p = nil
		} else {
			if p != nil && !p.color.Equals(back) {
				h.drawPending(ug, p, height)
				p = nil
			}
			if p == nil {
				p = &pending{x1, x2, back}
			} else {
				p.x2 = x2
			}
		}
	}
	if p != nil {
		h.drawPending(ug, p, height)
	}
}

// pending is upstream's TimeHeader.Pending: a run of days of a color.
type pending struct {
	x1, x2 float64
	color  color.HColor
}

func (h *timeHeader) drawPending(ug klimt.UGraphic, p *pending, height float64) {
	h.drawRectangle(ug.Apply(klimt.Bg(p.color)), height, p.x1, p.x2)
}

// textBlock is getTextBlockSLOW.
func (h *timeHeader) textBlock(text string, fc *klimt.FontConfiguration) klimt.TextBlock {
	return klimt.DisplayWithNewlines(klimt.PragmaCreateEmpty(), text).Create(fc, geom.HorizontalAlignmentLeft,
		klimt.SpriteContainerEmpty{})
}

// fontConfigurationOf is getFontConfigurationSLOW.
func (h *timeHeader) fontConfigurationOf(param klimt.SName, bold bool, c color.HColor) *klimt.FontConfiguration {
	return fontConfiguration(h.timelineStyle.Font(param), bold, c)
}

// printCentered is printCentered: the last of the formats that fits
// between start and end, centered, or the first one, from start.
func (h *timeHeader) printCentered(ug klimt.UGraphic, hideIfTooBig bool, start, end float64, when TimePoint,
	fc *klimt.FontConfiguration, formats ...TimePointFormat) {
	available := end - start
	for i := len(formats) - 1; i >= 0; i-- {
		text := h.textBlock(formats[i].Format(when, h.self.locale()), fc)
		width := text.CalculateDimension(ug.StringBounder()).Width()
		if (i == 0 && !hideIfTooBig) || width <= available {
			diff := math.Max(0, available-width)
			text.DrawU(ug.Apply(klimt.UTranslateDx(start + diff/2)))
			return
		}
	}
}

// drawRectangle is drawRectangle: a rectangle below the header.
func (h *timeHeader) drawRectangle(ug klimt.UGraphic, height, x1, x2 float64) {
	if height == 0 {
		return
	}
	ug = ug.Apply(color.None())
	ug = ug.Apply(klimt.NewUTranslate(x1, h.self.FullHeaderHeight(ug.StringBounder())))
	if x2 > x1 {
		ug.Draw(klimt.URectangleBuild(x2-x1, height))
	}
}

// printVerticalSeparators is printVerticalSeparators: the separators of
// "separator just before".
func (h *timeHeader) printVerticalSeparators(ug klimt.UGraphic, totalHeightWithoutFooter float64) {
	ug = h.timelineStyle.ApplyVerticalSeparatorStyle(ug)
	for day := h.minDay(); day.CompareTo(h.maxDay()) <= 0; day = day.PlusDays(1) {
		if h.isBold(day) {
			h.drawVline(ug, h.timeScale.Position(TimePointOfStartOfDay(day)),
				h.self.FullHeaderHeight(ug.StringBounder()), totalHeightWithoutFooter)
		}
	}
}

// timeHeaderCalendar is upstream's TimeHeaderCalendar: the headers of a
// project with a start.
type timeHeaderCalendar struct {
	timeHeader
}

func (h *timeHeaderCalendar) locale() Locale { return h.weekConfig.locale }

func (h *timeHeaderCalendar) isZeroOnDay(instant TimePoint) bool {
	return ngm.IsZeroOnDay(h.dayCalendar.OpenClose().AsPiecewiseConstant(), instant.Day())
}

// printNamedDays is printNamedDays: the name of each run of named days,
// below the calendar.
func (h *timeHeaderCalendar) printNamedDays(ug klimt.UGraphic) {
	if len(h.dayCalendar.NameDays()) == 0 {
		return
	}
	var last *string
	fc := h.fontConfigurationOf(klimt.SNameMonth, false, h.openFontColor())
	for day := h.minDay(); day.CompareTo(h.maxDay()) <= 0; day = day.PlusDays(1) {
		wink := TimePointOfStartOfDay(day)
		name, ok := h.dayCalendar.DayName(wink)
		if ok && (last == nil || name != *last) {
			x1 := h.timeScale.Position(wink)
			label := h.textBlock(name, fc)
			position := h.self.TimeHeaderHeight(ug.StringBounder())
			label.DrawU(ug.Apply(klimt.NewUTranslate(x1, position)))
		}
		if ok {
			last = &name
		} else {
			last = nil
		}
	}
}

// headerNameDayHeight is the height of the names of the days below the
// calendar of the daily, weekly and monthly headers.
func (h *timeHeaderCalendar) headerNameDayHeight() float64 {
	if len(h.dayCalendar.NameDays()) > 0 {
		return h.timelineStyle.FontSizeDay() + 6
	}
	return 0
}

// The widths of a day with working hours: 120 pixels, the closed hours 5
// times narrower.
const (
	headerCloseFactor = 0.2
	headerDayWidth    = 120
)

// newTimeHeader is TimeHeaderFactory.createTimeHeader.
func (d *GanttDiagram) newTimeHeader(timelineStyle *TimelineStyleData) TimeHeader {
	base := timeHeader{dayCalendar: d.dayCalendar, timeBounds: d.timeBounds, scaleConfig: d.scaleConfig,
		timelineStyle: timelineStyle, weekConfig: d.weekConfig}
	start := TimePointOfStartOfDay(d.timeBounds.MinDay())
	cellWidth := timelineStyle.CellWidth()
	effectiveScale := d.scaleConfig.EffectiveScale()
	if d.timeBounds.MinDay() == TimePointEpoch() {
		base.timeScale = NewTimeScaleWink(cellWidth, effectiveScale, d.scaleConfig.printScale)
		h := &timeHeaderSimple{timeHeader: base}
		h.self = h
		return h
	}
	if d.dayCalendar.HasWorkingTimeRanges() {
		workingHours := &WorkingHours{d.dayCalendar.WorkingTimeRanges(), headerDayWidth, headerCloseFactor}
		base.timeScale = &VariantTimeScale{d.timeBounds.MinDay(), d.timeBounds.MaxDay(), workingHours,
			d.dayCalendar.OpenClose()}
		h := &timeHeaderVariant{timeHeaderCalendar: timeHeaderCalendar{base}, workingHours: workingHours}
		h.self = h
		return h
	}
	compressed := func() TimeScale {
		return NewTimeScaleCompressed(cellWidth, start, effectiveScale, d.timeBounds.printStart)
	}
	switch d.scaleConfig.printScale {
	case PrintScaleDaily:
		if d.scaleConfig.hideClosed {
			base.timeScale = NewTimeScaleDailyHideClosed(cellWidth, start, effectiveScale, d.dayCalendar.OpenClose())
		} else {
			base.timeScale = NewTimeScaleDaily(cellWidth, start, effectiveScale, d.timeBounds.printStart)
		}
		h := &timeHeaderDaily{timeHeaderCalendar: timeHeaderCalendar{base}}
		h.self = h
		return h
	case PrintScaleWeekly:
		base.timeScale = compressed()
		h := &timeHeaderWeekly{timeHeaderCalendar{base}}
		h.self = h
		return h
	case PrintScaleMonthly:
		base.timeScale = compressed()
		h := &timeHeaderMonthly{timeHeaderCalendar{base}}
		h.self = h
		return h
	case PrintScaleQuarterly:
		base.timeScale = compressed()
		h := &timeHeaderQuarterly{timeHeaderCalendar{base}}
		h.self = h
		return h
	}
	base.timeScale = compressed()
	h := &timeHeaderYearly{timeHeaderCalendar{base}}
	h.self = h
	return h
}
