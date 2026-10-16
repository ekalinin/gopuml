// SPDX-License-Identifier: MIT
// Provenance: original

package uml

import (
	"errors"
	"math"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/command/regex"
	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// panicMessage runs f and returns the Java exception it throws, or "".
func panicMessage(f func()) (msg string) {
	defer func() {
		if r := recover(); r != nil {
			var re *color.RuntimeError
			if err, ok := r.(error); ok && errors.As(err, &re) {
				msg = re.Error()
				return
			}
			panic(r)
		}
	}()
	f()
	return ""
}

func TestJavaParseInt(t *testing.T) {
	for in, want := range map[string]int{"0": 0, "12": 12, "-0012": -12, "+5": 5, "2147483647": 2147483647} {
		if got := JavaParseInt(in); got != want {
			t.Errorf("JavaParseInt(%q) = %d, want %d", in, got, want)
		}
	}
	for in, want := range map[string]string{
		"99999999999": `java.lang.NumberFormatException: For input string: "99999999999"`,
		"":            `java.lang.NumberFormatException: For input string: ""`,
		"0x1":         `java.lang.NumberFormatException: For input string: "0x1"`,
		"-":           `java.lang.NumberFormatException: For input string: "-"`,
	} {
		if got := panicMessage(func() { JavaParseInt(in) }); got != want {
			t.Errorf("JavaParseInt(%q) threw %q, want %q", in, got, want)
		}
	}
}

func TestJavaParseDouble(t *testing.T) {
	for in, want := range map[string]float64{"1.": 1, ".5": 0.5, "00012": 12, "99999999999999999999999": 1e23} {
		if got := JavaParseDouble(in); got != want {
			t.Errorf("JavaParseDouble(%q) = %v, want %v", in, got, want)
		}
	}
	for in, want := range map[string]string{
		"1.2.3": "java.lang.NumberFormatException: multiple points",
		"..":    "java.lang.NumberFormatException: multiple points",
		".":     `java.lang.NumberFormatException: For input string: "."`,
	} {
		if got := panicMessage(func() { JavaParseDouble(in) }); got != want {
			t.Errorf("JavaParseDouble(%q) threw %q, want %q", in, got, want)
		}
	}
}

func TestStringHelpers(t *testing.T) {
	if got := JavaSplit("ab", ""); !slices.Equal(got, []string{"a", "", "b"}) {
		t.Errorf("JavaSplit = %q", got)
	}
	if got := JavaSplit("abc", ""); !slices.Equal(got, []string{"abc"}) {
		t.Errorf("JavaSplit without separator = %q", got)
	}
	if got := JavaSplit("", ""); len(got) != 0 {
		t.Errorf("JavaSplit of the separator = %q", got)
	}
	// '@' separates names, as upstream's pattern does not accept it.
	if got := SplitComma(` A@b, "Long C" ,d.e `); !slices.Equal(got, []string{"A", "b", "Long C", "d.e"}) {
		t.Errorf("SplitComma = %q", got)
	}
	if !EqualsIgnoreCase("Hide", "hIDE") || EqualsIgnoreCase("hide", "hid") {
		t.Error("EqualsIgnoreCase")
	}
	if !IsEmpty("", false) || !IsEmpty(" \t\r\n\x00", true) || IsEmpty(" x", true) {
		t.Error("IsEmpty")
	}
	if JavaLength("a😀") != 3 {
		t.Error("JavaLength counts UTF-16 units")
	}
	if got := JavaDots(`a.b[.]\.c`); got != `a`+javaDot+`b[.]\.c` {
		t.Errorf("JavaDots = %q", got)
	}
}

func TestScaleFactor(t *testing.T) {
	tests := []struct {
		s    Scale
		want float64
	}{
		{Scale{Kind: ScaleSimple, A: 1.5}, 1.5},
		{Scale{Kind: ScaleSimple, A: 10}, 4},
		{Scale{Kind: ScaleSimple, A: -1}, 1},
		{Scale{Kind: ScaleWidth, A: 50}, 0.5},
		{Scale{Kind: ScaleHeight, A: 400}, 2},
		{Scale{Kind: ScaleMaxWidth, A: 400}, 1},
		{Scale{Kind: ScaleMaxHeight, A: 100}, 0.5},
		{Scale{Kind: ScaleWidthAndHeight, A: 300, B: 100}, 0.5},
		{Scale{Kind: ScaleMaxWidthAndHeight, A: 300, B: 400}, 1},
	}
	for _, tt := range tests {
		if got := tt.s.Factor(100, 200); got != tt.want {
			t.Errorf("%v.Factor = %v, want %v", &tt.s, got, tt.want)
		}
	}
	if got := (&Scale{Kind: ScaleWidth, A: 0}).Factor(0, 0); !math.IsNaN(got) {
		t.Errorf("0/0 gives NaN as in Java, got %v", got)
	}
	if got := (&Scale{Kind: ScaleWidthAndHeight, A: 3, B: 4}).String(); got != "ScaleWidthAndHeight(3,4)" {
		t.Errorf("String = %q", got)
	}
}

func TestEntityPortions(t *testing.T) {
	var s EntityPortions
	s = s.With(PortionMember.AsSet())
	if !s.Contains(PortionField) || !s.Contains(PortionMethod) || s.Contains(PortionStereotype) {
		t.Errorf("member is fields and methods: %v", s.List())
	}
	s = s.Without(PortionMethod.AsSet()).With(PortionStereotype.AsSet())
	if got := s.List(); !slices.Equal(got, []EntityPortion{PortionField, PortionStereotype}) {
		t.Errorf("List = %v", got)
	}
	if GenderPortion("Stereotypes") != PortionStereotype || GenderPortion("circle") != PortionCircledCharacter {
		t.Error("GenderPortion")
	}
}

func source(lines ...string) *core.UmlSource {
	var sl []text.StringLocated
	loc := text.NewLineLocation("test", nil)
	for _, l := range lines {
		loc = loc.OneLineRead()
		sl = append(sl, text.NewStringLocated(l, loc))
	}
	return core.NewUmlSource(sl, sl, true)
}

func TestPatchBase64(t *testing.T) {
	src := source("@startuml", "a <img data:image/png;base64,QUJD+/=={scale=1}> b data:image/png;base64,", "@enduml")
	patched, md5map := PatchBase64(src)
	line := patched.Lines()[1]
	// MD5 of "QUJD+/==" and of "".
	want := "a <img data:image/png;md5,faede9840fe28375764d5bc0ccb021b2{scale=1}> b data:image/png;md5,d41d8cd98f00b204e9800998ecf8427e"
	if line.S != want {
		t.Errorf("patched line\n got %q\nwant %q", line.S, want)
	}
	if line.Location.Position != 1 || md5map["faede9840fe28375764d5bc0ccb021b2"] != "QUJD+/==" {
		t.Errorf("location %d md5map %v", line.Location.Position, md5map)
	}
	unchanged := source("@startuml", "A -> B", "@enduml")
	if p, m := PatchBase64(unchanged); p != unchanged || len(m) != 0 {
		t.Error("a source without images is kept")
	}
}

// testDiagram is a minimal UML diagram for the common commands.
type testDiagram struct {
	Base
	hideEmpty bool
	gender    string
}

func (*testDiagram) Description() string              { return "(test)" }
func (d *testDiagram) SetHideEmptyDescription(h bool) { d.hideEmpty = h }
func (d *testDiagram) HideShowByGender(arg *regex.Result) command.Result {
	d.gender = arg.Str("COMMAND", 0) + " " + arg.Str("PORTION", 0)
	return command.Ok()
}

func testFactory() *Factory[*testDiagram] {
	return NewFactory(core.Class, func() []command.Command[*testDiagram] {
		return AddCommonCommands1[*testDiagram](nil)
	}, func(ctx *core.RenderContext, src *core.UmlSource, md5map map[string]string) *testDiagram {
		return &testDiagram{Base: NewBase(ctx, src, core.Class, md5map)}
	})
}

func parse(t *testing.T, f *Factory[*testDiagram], lines ...string) core.Diagram {
	t.Helper()
	src := source(append(append([]string{"@startuml"}, lines...), "@enduml")...)
	return f.CreateSystem(core.NewRenderContext(), src)
}

func TestCommonCommands(t *testing.T) {
	f := testFactory()
	r := parse(t, f,
		"title The title",
		"caption",
		"  cap",
		"end caption",
		"legend top left",
		"  leg",
		"endlegend",
		"left header head",
		"footer",
		" foot",
		"endfooter",
		"mainframe frame",
		"scale max 300x200",
		"page 2x3",
		"skinparam ArrowColor red",
		"skinparam participant {",
		"  FontSize 17",
		"}",
		"set separator ::",
		"!pragma teoz true",
		"hide empty description",
		"hide stereotype",
		"sprite $s {",
		"0F0F",
		"F0F0",
		"}",
		"",
	)
	d, ok := r.(*testDiagram)
	if !ok {
		e := r.(*core.ErrorDiagram)
		t.Fatalf("not parsed: %s at %d", e.Error.Error, e.Error.Position())
	}
	if got := d.Title().Display().String(); got != "[The title]" || d.Title().LineLocation().Position != 1 {
		t.Errorf("title %s", got)
	}
	if got := d.Caption().Display().String(); got != "[cap]" {
		t.Errorf("caption %s", got)
	}
	if lg := d.Legend(); lg.Display().String() != "[leg]" || lg.HorizontalAlignment() != geom.HorizontalAlignmentLeft ||
		lg.VerticalAlignment() != geom.VerticalAlignmentTop {
		t.Errorf("legend %s %v %v", lg.Display(), lg.HorizontalAlignment(), lg.VerticalAlignment())
	}
	if h := d.Header(); h.Display().String() != "[head]" || h.HorizontalAlignment() != geom.HorizontalAlignmentLeft {
		t.Errorf("header %s %v", h.Display(), h.HorizontalAlignment())
	}
	if ft := d.Footer(); ft.Display().String() != "[foot]" || ft.VerticalAlignment() != VerticalAlignmentNull {
		t.Errorf("footer %s", ft.Display())
	}
	if mf := d.MainFrame(); mf.Display().String() != "[frame]" || mf.HorizontalAlignment() != klimt.HorizontalAlignmentNull {
		t.Errorf("mainframe %s", mf.Display())
	}
	if s := d.Scale(); s == nil || s.Kind != ScaleMaxWidthAndHeight || s.A != 300 || s.B != 200 {
		t.Errorf("scale %v", s)
	}
	if h, v := d.SplitPages(); h != 2 || v != 3 {
		t.Errorf("pages %d %d", h, v)
	}
	values := d.Skin().Values()
	if values["arrowcolor"] != "red" || values["participantfontsize"] != "17" || !d.IsSkinParamUsed() {
		t.Errorf("skinparams %v", values)
	}
	if sep, ok := d.NamespaceSeparator(); !ok || sep != "::" {
		t.Errorf("separator %q", sep)
	}
	if !d.Pragma().IsDefine(klimt.PragmaTeoz) || !d.hideEmpty || d.gender != "hide stereotype" {
		t.Errorf("pragma, hide empty description or gender not applied")
	}
	if sp, ok := d.Skin().Sprite("s").(*SpriteDef); !ok || sp.Width != 4 || sp.Height != 2 || sp.GrayLevel != 16 {
		t.Errorf("sprite %#v", d.Skin().Sprite("s"))
	}
}

func TestCommonCommandErrors(t *testing.T) {
	f := testFactory()
	tests := []struct {
		lines []string
		msg   string
		line  int
	}{
		{[]string{"scale 0"}, "Scale cannot be zero", 1},
		{[]string{"title", "end title"}, "No title defined", 1},
		{[]string{"page 0x1"}, "Argument must be positive", 1},
		{[]string{"skin nosuch"}, "Cannot find style nosuch", 1},
		{[]string{"skin strictuml"}, "Incomplete style strictuml: root does not define [FontName, FontSize, FontStyle, FontColor, LineColor, LineThickness, BackGroundColor, HorizontalAlignment]", 1},
		{[]string{"!pragma layout"}, `You should send a mail to plantuml@gmail.com with this log (V1.2026.8) java.lang.NullPointerException: Cannot invoke "String.equalsIgnoreCase(String)" because "value" is null`, 1},
		{[]string{"sprite $s [2x2/3] {", "00", "}"}, "Only 4, 8 or 16 graylevel are allowed.", 1},
		{[]string{"sprite $s {", "}"}, "No sprite defined.", 1},
		{[]string{"scale 1.2.3"}, "You should send a mail to plantuml@gmail.com with this log (V1.2026.8) java.lang.NumberFormatException: multiple points", 1},
	}
	for _, tt := range tests {
		e, ok := parse(t, f, tt.lines...).(*core.ErrorDiagram)
		if !ok {
			t.Errorf("%q: not an error", tt.lines)
			continue
		}
		if e.Error.Error != tt.msg || e.Error.Position() != tt.line {
			t.Errorf("%q: %q at %d, want %q at %d", tt.lines, e.Error.Error, e.Error.Position(), tt.msg, tt.line)
		}
	}
}

func TestMultilineShapes(t *testing.T) {
	lines := func(s ...string) command.BlocLines {
		var sl []text.StringLocated
		for _, l := range s {
			sl = append(sl, text.NewStringLocated(l, nil))
		}
		return command.NewBlocLines(sl)
	}
	c := &Multilines2[*testDiagram]{
		Start:        regex.NewConcat(regex.Start(), Lit("block"), regex.SpaceZeroOrMore(), Lit(`\{`), regex.End()),
		Strategy:     RemoveStartingQuote,
		TrimEnd:      TrimBoth,
		End:          Compile(`^\}$`),
		FinalBracket: true,
	}
	for _, tt := range []struct {
		in   []string
		want command.Control
	}{
		{[]string{"block {"}, command.OKPartial},
		{[]string{"block"}, command.OKPartial},
		{[]string{"block", "{", "x", "}"}, command.OK},
		{[]string{"block {", "' comment", "}"}, command.OK},
		{[]string{"other {"}, command.NotOK},
	} {
		if got := c.IsValid(lines(tt.in...)); got != tt.want {
			t.Errorf("Multilines2.IsValid(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
	b := commandSkinParamMultilines[*testDiagram]()
	if got := b.IsValid(lines("skinparam {", "a {", "b c", "}", "}")); got != command.OK {
		t.Errorf("bracket block = %v", got)
	}
	if got := b.IsValid(lines("skinparam {", "a {", "b c", "}")); got != command.OKPartial {
		t.Errorf("open bracket block = %v", got)
	}
	if got := b.IsValid(lines("skinparam {", "{{ bad")); got != command.NotOK {
		t.Errorf("inconsistent bracket block = %v", got)
	}
	if got := strings.Join(TrimSmart(lines("legend", "    a", "      b", "  c"), 1).Strings(), "|"); got != "legend|a|  b|c" {
		t.Errorf("TrimSmart = %q", got)
	}
	if got := panicMessage(func() { SubExtract(lines("x"), 1, 1) }); got != "java.lang.IllegalArgumentException: fromIndex(1) > toIndex(0)" {
		t.Errorf("SubExtract threw %q", got)
	}
}

func TestDecodeSprites(t *testing.T) {
	s := buildGraySprite(16, -1, -1, []string{"0F", "8"})
	if s.Width != 2 || s.Height != 2 || !slices.Equal(s.Gray, []int32{0, 15, 8, 0}) {
		t.Errorf("gray16 %+v", s)
	}
	if got := panicMessage(func() { buildGraySprite(16, -1, -1, []string{"0G"}) }); got != "java.lang.IllegalArgumentException: level=-1 grayLevel=16" {
		t.Errorf("bad level threw %q", got)
	}
	s = buildGraySprite(8, 1, 2, []string{"z"})
	if !slices.Equal(s.Gray, []int32{61 / 8, 61 % 8}) {
		t.Errorf("gray8 %v", s.Gray)
	}
	c := buildColorSprite([]string{"..zz"})
	if c.Width != 2 || c.Color[0] != 0 || c.Color[1] != 0xffffff {
		t.Errorf("color %x", c.Color)
	}
	if got := panicMessage(func() { base64Decode("abc") }); got != "java.lang.IllegalArgumentException: Length of Base64 encoded input string is not a multiple of 4." {
		t.Errorf("base64 threw %q", got)
	}
	if got := string(base64Decode("QUJD")); got != "ABC" {
		t.Errorf("base64 = %q", got)
	}
	if got := panicMessage(func() { newGrid(1<<20, 1<<20) }); got != "java.lang.OutOfMemoryError: Java heap space" {
		t.Errorf("huge sprite threw %q", got)
	}
}

func TestFactoryConcurrent(t *testing.T) {
	f := testFactory()
	var wg sync.WaitGroup
	for i := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			title := strings.Repeat("t", i+1)
			d, ok := parse(t, f, "title "+title, "skinparam ArrowColor red").(*testDiagram)
			if !ok || d.Title().Display().String() != "["+title+"]" {
				t.Errorf("parallel parse %d", i)
			}
		}()
	}
	wg.Wait()
}
