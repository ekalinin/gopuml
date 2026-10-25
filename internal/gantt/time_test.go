// SPDX-License-Identifier: MIT
// Provenance: original

package gantt

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// The expected values of these tests were printed by upstream's classes of
// the reference jar.

func panicMessage(t *testing.T, f func()) (msg string) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(*color.RuntimeError)
			if !ok {
				t.Fatalf("panic %v is not a Java exception", r)
			}
			msg = e.Error()
		}
	}()
	f()
	return ""
}

func TestTimePointFormats(t *testing.T) {
	jun17 := TimePointOfStartOfDayYMD(2020, 6, 17)
	day := jtime.DurationOfDays(1).PlusSeconds(1)
	long := jtime.DurationOfDays(22).PlusHours(3).PlusMinutes(5)
	for _, tc := range []struct {
		language string
		want     string
	}{
		{"en", "Jun 17 | June 2020 | Jun 2020 | Wednesday | We | Q2 | Day 3 | 1 day, 1 second | 22 days, 3 hours, 5 minutes | none | []"},
		{"de", "17. Jun | Juni 2020 | Jun 2020 | Mittwoch | Mi | Q2 | Tag 3 | 1 Tag, 1 Sekunde | 22 Tage, 3 Stunden, 5 Minuten | keine | []"},
		{"fr", "17 juin | juin 2020 | juin 2020 | mercredi | me | Q2 | Jour 3 | 1 jour, 1 seconde | 22 jours, 3 heures, 5 minutes | aucune | []"},
		{"es", "17 de jun | junio 2020 | jun 2020 | miércoles | mi | Q2 | Día 3 | 1 día, 1 segundo | 22 días, 3 horas, 5 minutos | ninguna | []"},
		{"ja", "6月/17日 | 6月 2020 | 6月 2020 | 水曜日 | 水 | Q2 | 3日目 | 1日1秒 | 22日3時間5分 | なし | []"},
		{"ko", "6월 17일 | 6월 2020 | 6월 2020 | 수요일 | 수 | Q2 | 3일째 | 1일, 1초 | 22일, 3시간, 5분 | 없음 | []"},
		{"ru", "17 июнь | июнь 2020 | июнь 2020 | среда | ср | Q2 | День 3 | 1 день, 1 секунда | 22 дня, 3 часа, 5 минут | нет | []"},
		{"zh", "6月17日 | 六月 2020 | 6月 2020 | 星期三 | 周三 | Q2 | 第3天 | 1天1秒 | 22天3小时5分钟 | 无 | []"},
		{"fr_FR", "Jun 17 | June 2020 | Jun 2020 | Wednesday | We | Q2 | Day 3 | 1 day, 1 second | 22 days, 3 hours, 5 minutes | none | []"},
		{"xx", "Jun 17 | June 2020 | Jun 2020 | Wednesday | We | Q2 | Day 3 | 1 day, 1 second | 22 days, 3 hours, 5 minutes | none | []"},
		{"FR", "17 juin | juin 2020 | juin 2020 | mercredi | me | Q2 | Jour 3 | 1 jour, 1 seconde | 22 jours, 3 heures, 5 minutes | aucune | []"},
	} {
		loc := NewLocale(tc.language)
		got := strings.Join([]string{jun17.StringShort(loc), FormatMonthYearLong.Format(jun17, loc),
			FormatMonthYearShort.Format(jun17, loc), FormatDayOfWeekLong.Format(jun17, loc),
			FormatDayOfWeekShort.Format(jun17, loc), FormatQuarter.Format(jun17, loc), i18nDayNumber(loc, 3),
			i18nDurationHumanReadable(loc, day), i18nDurationHumanReadable(loc, long),
			i18nDurationHumanReadable(loc, jtime.ZeroDuration),
			"[" + i18nDurationHumanReadable(loc, jtime.DurationOfHours(-5)) + "]"}, " | ")
		if got != tc.want {
			t.Errorf("%s:\n got %s\nwant %s", tc.language, got, tc.want)
		}
	}
	for _, tc := range []struct {
		days int64
		want string
	}{
		{1, "1 день"}, {2, "2 дня"}, {5, "5 дней"}, {11, "11 дней"}, {12, "12 дней"}, {21, "21 день"},
		{22, "22 дня"}, {25, "25 дней"}, {111, "111 дней"}, {112, "112 дней"}, {101, "101 день"},
	} {
		if got := i18nDurationHumanReadable("ru", jtime.DurationOfDays(tc.days)); got != tc.want {
			t.Errorf("ru %d days: %q, want %q", tc.days, got, tc.want)
		}
	}
}

func TestTimePointNames(t *testing.T) {
	if MonthFromString("Sept") != jtime.September || MonthFromString("decembre") != jtime.December ||
		DayOfWeekFromString("SAT") != jtime.Saturday {
		t.Error("names from their prefixes")
	}
	for _, tc := range []struct {
		value, want string
	}{
		{"xyz", "java.lang.IllegalArgumentException"},
		{"ja", "java.lang.StringIndexOutOfBoundsException: Range [0, 3) out of bounds for length 2"},
	} {
		if got := panicMessage(t, func() { MonthFromString(tc.value) }); got != tc.want {
			t.Errorf("%q: %q, want %q", tc.value, got, tc.want)
		}
	}
	if got := TimePointOfStartOfDayYMD(2020, 1, 6).AbsoluteDayNum(); got != 18267 {
		t.Errorf("absolute day %d", got)
	}
	if got := TimePointOfStartOfDayYMD(1969, 12, 31).AbsoluteDayNum(); got != -1 {
		t.Errorf("absolute day %d", got)
	}
	if got := TimePointOfStartOfDayYMD(2020, 12, 1).Quarter(); got != "Q4" {
		t.Errorf("quarter %q", got)
	}
	if got := MonthRegexString(); got != "JAN[a-z]*|FEB[a-z]*|MAR[a-z]*|APR[a-z]*|MAY[a-z]*|JUN[a-z]*|JUL[a-z]*|AUG[a-z]*|SEP[a-z]*|OCT[a-z]*|NOV[a-z]*|DEC[a-z]*" {
		t.Errorf("%q", got)
	}
	if got := DayOfWeekRegexString(); got != "MON[a-z]*|TUE[a-z]*|WED[a-z]*|THU[a-z]*|FRI[a-z]*|SAT[a-z]*|SUN[a-z]*" {
		t.Errorf("%q", got)
	}
	for _, s := range []string{"September", "SEPT", "sep"} {
		if !MonthUbrex().Match(s).ExactMatch() {
			t.Errorf("month %q", s)
		}
	}
	for _, s := range []string{"Saturday", "SAT", "sat"} {
		if !DayOfWeekUbrex().Match(s).ExactMatch() {
			t.Errorf("day %q", s)
		}
	}
	if PrintScaleFromString("Weekly") != PrintScaleDaily || PrintScaleFromString("weekly") != PrintScaleWeekly ||
		PrintScaleYearly.DefaultScale() != 1.0/60 {
		t.Error("print scale")
	}
	strategy := WeekNumberStrategy{jtime.Monday, 4}
	if got := TimePointOfStartOfDayYMD(2024, 12, 30).WeekOfYear(strategy); got != 53 {
		t.Errorf("week %d", got)
	}
}
