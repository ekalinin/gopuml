// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/GanttI18n.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/utils/I18nTimeData.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"strconv"
	"strings"

	"github.com/ekalinin/gopuml/internal/jtime"
)

// i18nTask is GanttI18n.task: the header of the column of the tasks.
func i18nTask(locale Locale) string {
	switch locale {
	case "fr":
		return "Tâche"
	case "es":
		return "Tarea"
	case "de":
		return "Aufgabe"
	case "ja":
		return "タスク"
	case "ko":
		return "작업"
	case "ru":
		return "Задача"
	case "zh":
		return "任务"
	}
	return "Task"
}

// i18nStart is GanttI18n.start: the header of the column of the starts.
func i18nStart(locale Locale) string {
	switch locale {
	case "fr":
		return "Début"
	case "es":
		return "Inicio"
	case "de":
		return "Start"
	case "ja":
		return "開始"
	case "ko":
		return "시작"
	case "ru":
		return "Начало"
	case "zh":
		return "开始"
	}
	return "Start"
}

// i18nEnd is GanttI18n.end: the header of the column of the ends.
func i18nEnd(locale Locale) string {
	switch locale {
	case "fr", "es":
		return "Fin"
	case "de":
		return "Ende"
	case "ja":
		return "終了"
	case "ko":
		return "끝"
	case "ru":
		return "Окончание"
	case "zh":
		return "结束"
	}
	return "End"
}

// i18nDuration is GanttI18n.duration: the header of the column of the
// durations.
func i18nDuration(locale Locale) string {
	switch locale {
	case "fr":
		return "Durée"
	case "es":
		return "Duración"
	case "de":
		return "Dauer"
	case "ja":
		return "期間"
	case "ko":
		return "기간"
	case "ru":
		return "Длительность"
	case "zh":
		return "持续时间"
	}
	return "Duration"
}

// i18nDurationHumanReadable is GanttI18n.durationHumanReadable: the days,
// hours, minutes and seconds of a duration, the units that are zero left
// out.
func i18nDurationHumanReadable(locale Locale, duration jtime.Duration) string {
	seconds := duration.Seconds()
	units := i18nUnitsOf(locale)
	if seconds == 0 {
		return units.none
	}
	days := seconds / (24 * 3600)
	seconds %= 24 * 3600
	hours := seconds / 3600
	seconds %= 3600
	minutes := seconds / 60
	seconds %= 60
	var b strings.Builder
	units.append(&b, days, 0)
	units.append(&b, hours, 1)
	units.append(&b, minutes, 2)
	units.append(&b, seconds, 3)
	return b.String()
}

// i18nUnits are the Units of GanttI18n: the names of the units of a
// language and how to write them.
type i18nUnits struct {
	none, separator string
	// suffixNoSpace writes the unit right after the number.
	suffixNoSpace bool
	// labels are the forms of the day, hour, minute and second: singular and
	// plural, one invariable form, or the three Russian forms (singular,
	// genitive plural, paucal).
	labels [4][]string
}

func (u *i18nUnits) append(b *strings.Builder, value int64, unit int) {
	if value <= 0 {
		return
	}
	if b.Len() > 0 {
		b.WriteString(u.separator)
	}
	b.WriteString(strconv.FormatInt(value, 10))
	if !u.suffixNoSpace {
		b.WriteByte(' ')
	}
	b.WriteString(u.label(unit, value))
}

func (u *i18nUnits) label(unit int, value int64) string {
	forms := u.labels[unit]
	switch len(forms) {
	case 1:
		return forms[0]
	case 3:
		mod100, mod10 := value%100, value%10
		if mod10 == 1 && mod100 != 11 {
			return forms[0]
		}
		if mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14) {
			return forms[2]
		}
		return forms[1]
	}
	if value > 1 {
		return forms[1]
	}
	return forms[0]
}

func i18nUnitsOf(locale Locale) *i18nUnits {
	switch locale {
	case "fr":
		return &i18nUnits{"aucune", ", ", false, [4][]string{{"jour", "jours"}, {"heure", "heures"},
			{"minute", "minutes"}, {"seconde", "secondes"}}}
	case "es":
		return &i18nUnits{"ninguna", ", ", false, [4][]string{{"día", "días"}, {"hora", "horas"},
			{"minuto", "minutos"}, {"segundo", "segundos"}}}
	case "de":
		return &i18nUnits{"keine", ", ", false, [4][]string{{"Tag", "Tage"}, {"Stunde", "Stunden"},
			{"Minute", "Minuten"}, {"Sekunde", "Sekunden"}}}
	case "ja":
		return &i18nUnits{"なし", "", true, [4][]string{{"日"}, {"時間"}, {"分"}, {"秒"}}}
	case "ko":
		return &i18nUnits{"없음", ", ", true, [4][]string{{"일"}, {"시간"}, {"분"}, {"초"}}}
	case "ru":
		return &i18nUnits{"нет", ", ", false, [4][]string{{"день", "дней", "дня"}, {"час", "часов", "часа"},
			{"минута", "минут", "минуты"}, {"секунда", "секунд", "секунды"}}}
	case "zh":
		return &i18nUnits{"无", "", true, [4][]string{{"天"}, {"小时"}, {"分钟"}, {"秒"}}}
	}
	return &i18nUnits{"none", ", ", false, [4][]string{{"day", "days"}, {"hour", "hours"},
		{"minute", "minutes"}, {"second", "seconds"}}}
}

// i18nDayNumber is GanttI18n.dayNumber: "Day N" for the dates of a project
// without a start.
func i18nDayNumber(locale Locale, dayNum int) string {
	n := strconv.Itoa(dayNum)
	switch locale {
	case "fr":
		return "Jour " + n
	case "es":
		return "Día " + n
	case "de":
		return "Tag " + n
	case "ja":
		return n + "日目"
	case "ko":
		return n + "일째"
	case "ru":
		return "День " + n
	case "zh":
		return "第" + n + "天"
	}
	return "Day " + n
}

// dayOfWeekLong is I18nTimeData.dayOfWeekLong.
func dayOfWeekLong(d jtime.DayOfWeek, locale Locale) string {
	if names, ok := dayOfWeekLongNames[string(locale)]; ok && names[d.Ordinal()] != "" {
		return names[d.Ordinal()]
	}
	return dayOfWeekLongEnglish[d.Ordinal()]
}

// dayOfWeekShort is I18nTimeData.dayOfWeekShort.
func dayOfWeekShort(d jtime.DayOfWeek, locale Locale) string {
	if names, ok := dayOfWeekShortNames[string(locale)]; ok && names[d.Ordinal()] != "" {
		return names[d.Ordinal()]
	}
	return dayOfWeekShortEnglish[d.Ordinal()]
}

// monthShort is I18nTimeData.monthShort.
func monthShort(m jtime.Month, locale Locale) string {
	if names, ok := monthShortNames[string(locale)]; ok && names[m.Ordinal()] != "" {
		return names[m.Ordinal()]
	}
	return monthShortEnglish[m.Ordinal()]
}

// monthLong is I18nTimeData.monthLong.
func monthLong(m jtime.Month, locale Locale) string {
	if names, ok := monthLongNames[string(locale)]; ok && names[m.Ordinal()] != "" {
		return names[m.Ordinal()]
	}
	return monthLongEnglish[m.Ordinal()]
}

// dayOfWeekLongNames are the names of upstream's I18nTimeData.dayOfWeekLong by language, "" where
// the language has none, then English.
var dayOfWeekLongNames = map[string][7]string{
	"de": {"Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag", "Sonntag"},
	"es": {"lunes", "martes", "miércoles", "jueves", "viernes", "sábado", "domingo"},
	"fr": {"lundi", "mardi", "mercredi", "jeudi", "vendredi", "samedi", "dimanche"},
	"ja": {"月曜日", "火曜日", "水曜日", "木曜日", "金曜日", "土曜日", "日曜日"},
	"ko": {"월요일", "화요일", "수요일", "목요일", "금요일", "토요일", "일요일"},
	"ru": {"понедельник", "вторник", "среда", "четверг", "пятница", "суббота", "воскресенье"},
	"zh": {"星期一", "星期二", "星期三", "星期四", "星期五", "星期六", "星期日"},
}

var dayOfWeekLongEnglish = [7]string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}

// dayOfWeekShortNames are the names of upstream's I18nTimeData.dayOfWeekShort by language, "" where
// the language has none, then English.
var dayOfWeekShortNames = map[string][7]string{
	"de": {"Mo", "Di", "Mi", "Do", "Fr", "Sa", "So"},
	"es": {"lu", "ma", "mi", "ju", "vi", "sá", "do"},
	"fr": {"lu", "ma", "me", "je", "ve", "sa", "di"},
	"ja": {"月", "火", "水", "木", "金", "土", "日"},
	"ko": {"월", "화", "수", "목", "금", "토", "일"},
	"ru": {"пн", "вт", "ср", "чт", "пт", "сб", "вс"},
	"zh": {"周一", "周二", "周三", "周四", "周五", "周六", "周日"},
}

var dayOfWeekShortEnglish = [7]string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}

// monthShortNames are the names of upstream's I18nTimeData.monthShort by language, "" where
// the language has none, then English.
var monthShortNames = map[string][12]string{
	"de": {"Jan", "Feb", "Mär", "Apr", "Mai", "Jun", "Jul", "Aug", "Sep", "Okt", "Nov", "Dez"},
	"es": {"ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sept", "oct", "nov", "dic"},
	"fr": {"janv.", "févr.", "mars", "avr.", "mai", "juin", "juil.", "août", "sept.", "oct.", "nov.", "déc."},
	"ja": {"1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"},
	"ko": {"1월", "2월", "3월", "4월", "5월", "6월", "7월", "8월", "9월", "10월", "11월", "12월"},
	"ru": {"янв.", "февр.", "март", "апр.", "май", "июнь", "июль", "авг.", "сент.", "окт.", "нояб.", "дек."},
	"zh": {"1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"},
}

var monthShortEnglish = [12]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

// monthLongNames are the names of upstream's I18nTimeData.monthLong by language, "" where
// the language has none, then English.
var monthLongNames = map[string][12]string{
	"de": {"Januar", "Februar", "März", "April", "Mai", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"},
	"es": {"enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"},
	"fr": {"janvier", "février", "mars", "avril", "mai", "juin", "juillet", "août", "septembre", "octobre", "novembre", "décembre"},
	"ja": {"1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"},
	"ko": {"1월", "2월", "3월", "4월", "5월", "6월", "7월", "8월", "9월", "10월", "11월", "12월"},
	"ru": {"январь", "февраль", "март", "апрель", "май", "июнь", "июль", "август", "сентябрь", "октябрь", "ноябрь", "декабрь"},
	"zh": {"一月", "二月", "三月", "四月", "五月", "六月", "七月", "八月", "九月", "十月", "十一月", "十二月"},
}

var monthLongEnglish = [12]string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
