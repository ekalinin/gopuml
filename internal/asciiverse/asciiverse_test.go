// SPDX-License-Identifier: MIT
// Provenance: original

package asciiverse

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
)

// export returns the text of a plan.
func export(t *testing.T, plan *InfinitePlan) string {
	t.Helper()
	var buf bytes.Buffer
	if err := plan.ExportTxt(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

// draw draws a block on a new plan of a format and returns the text.
func draw(t *testing.T, format font.FileFormat, block AsciiBlock) string {
	t.Helper()
	plan := NewInfinitePlan(format)
	block.AsciiDraw(plan)
	return export(t, plan)
}

// lines joins lines with a final newline, as ExportTxt writes them.
func lines(l ...string) string { return strings.Join(l, "\n") + "\n" }

// javaException returns the class of the Java exception a function panics
// with, "" if it does not panic.
func javaException(f func()) (class string) {
	defer func() {
		if r := recover(); r != nil {
			var e *color.RuntimeError
			if err, ok := r.(error); ok && errors.As(err, &e) {
				class = e.Class
			} else {
				class = "other panic"
			}
		}
	}()
	f()
	return ""
}

func TestADimension2D(t *testing.T) {
	d := NewADimension2D(3, 4)
	if d.String() != "[3,4]" {
		t.Errorf("String = %q", d.String())
	}
	if got := d.Delta(2); got != NewADimension2D(5, 6) {
		t.Errorf("Delta = %v", got)
	}
	if got := d.DeltaWH(0, 0); got != d {
		t.Errorf("DeltaWH(0, 0) = %v", got)
	}
	if got := d.MergeTB(NewADimension2D(5, 1)); got != NewADimension2D(5, 5) {
		t.Errorf("MergeTB = %v", got)
	}
	if got := d.MergeLR(NewADimension2D(5, 1)); got != NewADimension2D(8, 4) {
		t.Errorf("MergeLR = %v", got)
	}
	if got := d.AtLeast(10, 1); got != NewADimension2D(10, 4) {
		t.Errorf("AtLeast = %v", got)
	}
	if got := ADimension2DMax(d, NewADimension2D(1, 9)); got != NewADimension2D(3, 9) {
		t.Errorf("ADimension2DMax = %v", got)
	}
	if got := d.WithWidth(7); got != NewADimension2D(7, 4) {
		t.Errorf("WithWidth = %v", got)
	}
	for _, f := range []func(){
		func() { NewADimension2D(-1, 0) },
		func() { NewADimension2D(0, -1) },
		func() { d.DeltaWH(-4, 0) },
	} {
		if got := javaException(f); got != illegalArgument {
			t.Errorf("negative size: got %q, want %s", got, illegalArgument)
		}
	}
}

func TestATranslate(t *testing.T) {
	a := NewATranslate(1, 2).Compose(ATranslateDx(3))
	if a.Dx() != 4 || a.Dy() != 2 || a.String() != "translate dx=4 dy=2" {
		t.Errorf("Compose = %v", a)
	}
	if r := a.Reverse(); r != NewATranslate(-4, -2) {
		t.Errorf("Reverse = %v", r)
	}
	if !a.IsAlmostSame(ATranslateDy(2)) || a.IsAlmostSame(ATranslateNone()) {
		t.Error("IsAlmostSame")
	}
}

func TestInfiniteString(t *testing.T) {
	var s InfiniteString
	if s.CharAt(5) != ' ' || s.CharAt(-5) != ' ' || s.LeftmostPosition() != 0 {
		t.Fatal("empty line")
	}
	s.SetStringAt(-2, "ab😀c")
	// "ab" at -2 and -1, the surrogate pair at 0 and 1, "c" at 2
	if s.CharAt(-2) != 'a' || s.CharAt(-1) != 'b' || s.CharAt(2) != 'c' || s.LeftmostPosition() != -2 {
		t.Errorf("SetStringAt across 0: %q", s.String(-2))
	}
	if got := s.String(-4); got != "  ab😀c" {
		t.Errorf("String(-4) = %q", got)
	}
	if got := s.String(0); got != "😀c" {
		t.Errorf("String(0) = %q", got)
	}
	if got := len(s.Chars(0)); got != 3 {
		t.Errorf("Chars(0) has %d UTF-16 units, want 3", got)
	}
}

func TestInfinitePlanExportAlignsNegativeColumns(t *testing.T) {
	plan := NewInfinitePlan(font.FormatATXT)
	plan.Move(2, 0).DrawString("x")
	plan.Move(-3, 1).DrawString("y")
	plan.Move(0, 3).DrawChar('z')
	// Every line starts at the leftmost column; the cells a line reserved
	// on the left read as spaces
	want := lines("     x", "y  ", "   ", "   z")
	if got := export(t, plan); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestInfinitePlanMoveAndCharAt(t *testing.T) {
	plan := NewInfinitePlan(font.FormatATXT)
	moved := plan.Move(3, 1).Move(1, 1)
	moved.DrawChar('#')
	if plan.CharAt(4, 2) != '#' || moved.CharAt(0, 0) != '#' || moved.CharAt(0, -5) != ' ' {
		t.Error("CharAt does not see the moved drawing")
	}
	if got := javaException(func() { plan.Move(0, -1).DrawChar('x') }); got != "java.lang.IndexOutOfBoundsException" {
		t.Errorf("negative row: got %q", got)
	}
}

func TestInfinitePlanLoneSurrogateIsQuestionMark(t *testing.T) {
	plan := NewInfinitePlan(font.FormatUTXT)
	plan.DrawChar(0xd83d)
	plan.Move(1, 0).DrawChar('a')
	plan.Move(2, 0).DrawChar(0xde00)
	if got := export(t, plan); got != "?a?\n" {
		t.Errorf("got %q", got)
	}
}

func TestInfinitePlanMuteChar(t *testing.T) {
	plan := NewInfinitePlan(font.FormatATXT)
	plan.DrawString("-a")
	plan.MuteLifelineBelowBox()
	plan.Move(1, 0).MuteLifelineBelowBox()
	plan.Move(2, 0).MuteLifelineAboveBox()
	if got := export(t, plan); got != "+a|\n" {
		t.Errorf("got %q", got)
	}
	if got := javaException(func() { plan.MuteChar('a') }); got != illegalArgument {
		t.Errorf("odd values: got %q", got)
	}
	uplan := NewInfinitePlan(font.FormatUTXT)
	uplan.DrawChar(uplan.HLineChar())
	uplan.Move(1, 0).DrawChar(uplan.HLineChar())
	uplan.MuteLifelineBelowBox()
	uplan.Move(1, 0).MuteLifelineAboveBox()
	if got := export(t, uplan); got != "┬┴\n" {
		t.Errorf("UTXT got %q", got)
	}
}

func TestInfinitePlanDrawBox(t *testing.T) {
	box := func(format font.FileFormat, w, h int) string {
		plan := NewInfinitePlan(format)
		plan.DrawBox(w, h)
		return export(t, plan)
	}
	if got, want := box(font.FormatATXT, 4, 3), lines(",--.", "|  |", "`--'"); got != want {
		t.Errorf("ATXT got %q, want %q", got, want)
	}
	if got, want := box(font.FormatUTXT, 3, 3), lines("┌─┐", "│ │", "└─┘"); got != want {
		t.Errorf("UTXT got %q, want %q", got, want)
	}
	if got, want := box(font.FormatATXT, 1, 1), lines(","); got != want {
		t.Errorf("1x1 got %q, want %q", got, want)
	}
	if got := javaException(func() { box(font.FormatATXT, 0, 3) }); got != illegalArgument {
		t.Errorf("empty box: got %q", got)
	}
}

func TestInfinitePlanDottedLine(t *testing.T) {
	plan := NewInfinitePlan(font.FormatATXT)
	plan.DrawHLineDotted(1, 6, 0, true)
	plan.DrawHLineDotted(1, 3, 1, false)
	if got, want := export(t, plan), lines(" - - -", " ---"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDisplayBlock(t *testing.T) {
	d := klimt.DisplayCreate("ab", "日本語x")
	block := DisplayBlock(d)
	if got := block.AsciiDimension(); got != NewADimension2D(4, 2) {
		t.Errorf("AsciiDimension = %v", got)
	}
	if got := DisplayBlock(klimt.DisplayNULL()).AsciiDimension(); got != NewADimension2D(0, 0) {
		t.Errorf("null display = %v", got)
	}
	if got, want := draw(t, font.FormatATXT, block), lines("ab", "日本語x"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	withNull := klimt.DisplayCreateSeq([]klimt.CharSequence{nil})
	if got := javaException(func() { DisplayBlock(withNull).AsciiDimension() }); got != nullPointer {
		t.Errorf("null line: got %q", got)
	}
}

func TestMarginLR(t *testing.T) {
	block := MarginLR(DisplayBlock(klimt.DisplayCreate("ab")), 1, 2)
	if got := block.AsciiDimension(); got != NewADimension2D(5, 1) {
		t.Errorf("AsciiDimension = %v", got)
	}
	if got := draw(t, font.FormatATXT, block); got != " ab\n" {
		t.Errorf("got %q", got)
	}
	empty := MarginLR(DisplayBlock(klimt.DisplayCreate("")), 1, 1)
	if got := empty.AsciiDimension(); got != NewADimension2D(0, 1) {
		t.Errorf("empty block grew: %v", got)
	}
	if got := javaException(func() { MarginLR(block, -1, 0) }); got != illegalArgument {
		t.Errorf("negative margin: got %q", got)
	}
}

func TestAsciiBlockStackTB(t *testing.T) {
	top := DisplayBlock(klimt.DisplayCreate("T"))
	bottom := DisplayBlock(klimt.DisplayCreate("bottom"))
	stack := NewAsciiBlockStackTB(top, bottom)
	if got := stack.AsciiDimension(); got != NewADimension2D(6, 3) {
		t.Errorf("AsciiDimension = %v", got)
	}
	if got, want := draw(t, font.FormatATXT, stack), lines("  T", "", "bottom"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	noTop := NewAsciiBlockStackTB(DisplayBlock(klimt.DisplayCreate("")), bottom)
	if got := noTop.AsciiDimension(); got != NewADimension2D(6, 1) {
		t.Errorf("empty top reserved rows: %v", got)
	}
}

func TestANote(t *testing.T) {
	// The text of a note on a message: one blank column before, three after
	text := MarginLR(DisplayBlock(klimt.DisplayCreate("hi")), 1, 3)
	note := NewANote(text)
	if got := note.AsciiDimension(); got != NewADimension2D(8, 3) {
		t.Errorf("AsciiDimension = %v", got)
	}
	if got, want := draw(t, font.FormatATXT, note), lines(",----!. ", "| hi |_\\", "`------'"); got != want {
		t.Errorf("ATXT got %q, want %q", got, want)
	}
	if got, want := draw(t, font.FormatUTXT, note), lines("╔══════╗", "║ hi  ░║", "╚══════╝"); got != want {
		t.Errorf("UTXT got %q, want %q", got, want)
	}
	// The text is drawn last, over the fold
	plain := NewANote(DisplayBlock(klimt.DisplayCreate("hello")))
	if got, want := draw(t, font.FormatATXT, plain), lines(",---!. ", "|hello\\", "`-----'"); got != want {
		t.Errorf("ATXT got %q, want %q", got, want)
	}
	if got, want := draw(t, font.FormatUTXT, plain), lines("╔═════╗", "║hello║", "╚═════╝"); got != want {
		t.Errorf("UTXT got %q, want %q", got, want)
	}
	// The minimum size, for an empty text
	if got := NewANote(DisplayBlock(klimt.DisplayCreate())).AsciiDimension(); got != NewADimension2D(5, 3) {
		t.Errorf("minimum = %v", got)
	}
	plan := NewInfinitePlan(font.FormatATXT)
	if _, ok := plan.CreateNoteBox(text).(*ANote); !ok {
		t.Error("CreateNoteBox does not return an ANote")
	}
}

func TestAGroupFrame(t *testing.T) {
	dim := NewADimension2D(14, 5)
	frame := NewAGroupFrame(dim, "alt x", true, true)
	if frame.BodyRowOffset() != 3 {
		t.Errorf("BodyRowOffset = %d", frame.BodyRowOffset())
	}
	want := lines(
		",------------.",
		"! alt x /    !",
		"!______/     !",
		"!            !",
		"~~~~~~~~~~~~~~")
	if got := draw(t, font.FormatATXT, frame); got != want {
		t.Errorf("ATXT got\n%s\nwant\n%s", got, want)
	}
	want = lines(
		"╔═══════╤════╗",
		"║ alt x │    ║",
		"╟───────┘    ║",
		"║       │    ║",
		"╚═══════╧════╝")
	if got := draw(t, font.FormatUTXT, frame); got != want {
		t.Errorf("UTXT got\n%s\nwant\n%s", got, want)
	}
	// A title that does not fit: no tab
	narrow := NewAGroupFrame(NewADimension2D(8, 3), "too long", true, true)
	if narrow.BodyRowOffset() != 1 {
		t.Errorf("narrow BodyRowOffset = %d", narrow.BodyRowOffset())
	}
	// A partition: the title centered on the top border, a plain bottom
	partition := NewAGroupFrame(NewADimension2D(10, 3), "p1", true, false)
	want = lines(",-- p1 --.", "!        !", "`--------'")
	if got := draw(t, font.FormatATXT, partition); got != want {
		t.Errorf("partition got\n%s\nwant\n%s", got, want)
	}
	if got := NewAGroupFrame(dim, "", false, true).BodyRowOffset(); got != 1 {
		t.Errorf("null title BodyRowOffset = %d", got)
	}
}

func TestAElseSeparator(t *testing.T) {
	if got, want := draw(t, font.FormatATXT, NewAElseSeparator(12, "no", true)), " - no  - -\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	// A label that does not fit is not drawn
	if got, want := draw(t, font.FormatUTXT, NewAElseSeparator(7, "long", true)), " ─ ─ ─\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := NewAElseSeparator(9, "", false).AsciiDimension(); got != NewADimension2D(9, 1) {
		t.Errorf("AsciiDimension = %v", got)
	}
}

// TestATable checks the table of a Gantt diagram's text output against the
// reference jar's output for the same cells.
func TestATable(t *testing.T) {
	table := NewATable(3, 4).WithLeftPadding(1).WithRightPadding(1)
	rows := [][]string{
		{"Task", "Start", "End", "Duration"},
		{"Prototype design", "Day 1", "Day 13", "13 days"},
		{"Test", "Day 14", "Day 17", "4 days"},
	}
	for r, row := range rows {
		for c, cell := range row {
			table.SetCellContent(r, c, klimt.DisplayCreate(cell))
		}
	}
	want := lines(
		",------------------+--------+--------+----------.",
		"| Task             | Start  | End    | Duration |",
		"+------------------+--------+--------+----------+",
		"| Prototype design | Day 1  | Day 13 | 13 days  |",
		"+------------------+--------+--------+----------+",
		"| Test             | Day 14 | Day 17 | 4 days   |",
		"`------------------+--------+--------+----------'")
	if got := draw(t, font.FormatATXT, table); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
	if got := table.AsciiDimension(); got != NewADimension2D(49, 7) {
		t.Errorf("AsciiDimension = %v", got)
	}
}

func TestATableMergedAndTallCells(t *testing.T) {
	table := NewATable(2, 2)
	table.SetCellContent(0, 0, klimt.DisplayCreate("wide"))
	table.MergeCells(0, 0, 1, 2)
	table.SetCellContent(1, 0, klimt.DisplayCreate("a", "b"))
	table.SetCellContent(1, 1, klimt.DisplayCreate("c"))
	want := lines(
		",----+-.",
		"|wide  |",
		"+----+-+",
		"|a   |c|",
		"|    | |",
		"`----+-'")
	if got := draw(t, font.FormatATXT, table); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
	if got := table.CellContent(5, 5); got != klimt.DisplayNULL() {
		t.Errorf("missing cell = %v", got)
	}
}

func TestAsciiSnakeIsUnsupported(t *testing.T) {
	if got := javaException(func() { AsciiSnakeCreate() }); got != unsupportedOperation {
		t.Errorf("got %q", got)
	}
}

func TestThrowUnsupported(t *testing.T) {
	defer func() {
		r := recover()
		e, ok := r.(*color.RuntimeError)
		if !ok || e.Error() != "java.lang.UnsupportedOperationException: some.Class" {
			t.Errorf("got %v", r)
		}
	}()
	ThrowUnsupported("some.Class")
}
