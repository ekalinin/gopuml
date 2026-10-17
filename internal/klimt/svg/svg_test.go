// SPDX-License-Identifier: MIT
// Provenance: original

package svg

import (
	"bytes"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/ekalinin/gopuml/internal/code"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// mustPanic runs f and returns the *color.RuntimeError it panics with.
func mustPanic(t *testing.T, f func()) *color.RuntimeError {
	t.Helper()
	var got *color.RuntimeError
	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("no panic")
			}
			e, ok := r.(error)
			if !ok || !errors.As(e, &got) {
				t.Fatalf("panic %v is not a RuntimeError", r)
			}
		}()
		f()
	}()
	return got
}

func newGraphic(format font.FileFormat) *UGraphicSvg {
	return Build(SvgOptionBasic().WithBackcolor(color.White), false, 42, font.NewStringBounderSvg(nil), format)
}

func write(t *testing.T, ug klimt.UGraphic, metadata string) string {
	t.Helper()
	var buf bytes.Buffer
	if err := ug.WriteToStream(&buf, metadata, 96); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestInvalidStatesPanic(t *testing.T) {
	g := NewSvgGraphics(1, SvgOptionBasic())
	if e := mustPanic(t, func() { g.CloseLink() }); e.Class != illegalState {
		t.Errorf("close link: %v", e)
	}
	if e := mustPanic(t, func() { g.CloseGroup() }); e.Class != illegalState {
		t.Errorf("close group: %v", e)
	}
	if e := mustPanic(t, func() { g.StartGroup(nil) }); e.Class != illegalArgument {
		t.Errorf("empty group: %v", e)
	}
	w := NewXMLWriter(0)
	if e := mustPanic(t, func() { w.Attribute("a", "b") }); e.Class != illegalState {
		t.Errorf("attribute: %v", e)
	}
	if e := mustPanic(t, func() { w.EndElement() }); e.Class != illegalState {
		t.Errorf("end element: %v", e)
	}
	w.StartElement("x")
	if e := mustPanic(t, func() { w.XML() }); e.Message != "Some elements were left open: [x]" {
		t.Errorf("open: %v", e)
	}
	if e := mustPanic(t, func() { (&XMLDocument{}).ToXML(0) }); e.Message != "No root element set." {
		t.Errorf("no root: %v", e)
	}
	// Upstream's UGraphicSvg needs a background.
	if e := mustPanic(t, func() {
		Build(SvgOptionBasic(), false, 42, font.NewStringBounderSvg(nil), font.FormatSVG)
	}); e.Class != nullPointer {
		t.Errorf("null background: %v", e)
	}
}

func TestNotPortedPanics(t *testing.T) {
	g := newGraphic(font.FormatSVG)
	e := mustPanic(t, func() { g.Draw(klimt.NewUCenteredCharacter('\u263A', font.Monospace(17))) })
	if !strings.Contains(e.Message, "not ported yet: net.sourceforge.plantuml.klimt.font.UnusedSpace") {
		t.Errorf("centered character: %v", e)
	}
	asPath := Build(SvgOptionBasic().WithBackcolor(color.White), true, 42, font.NewStringBounderSvg(nil),
		font.FormatSVG)
	fc := klimt.NewFontConfiguration(font.SansSerif(12), color.Black, color.Blue, nil)
	e = mustPanic(t, func() { asPath.Draw(klimt.UTextBuild("x", fc)) })
	if !strings.Contains(e.Message, "not ported yet: java.awt.font.TextLayout") {
		t.Errorf("text as path: %v", e)
	}
	// A clipped text is not drawn, so it does not panic.
	asPath.Apply(klimt.NewUClip(100, 100, 1, 1)).Draw(klimt.UTextBuild("x", fc))
}

// Upstream throws NullPointerException for an underlined text without
// hyperlink underline stroke ("skinparam hyperlinkUnderline false"); here
// the text has no underline (design D16).
func TestUnderlineWithoutStroke(t *testing.T) {
	g := newGraphic(font.FormatSVG)
	fc := klimt.NewFontConfiguration(font.SansSerif(12), color.Black, color.Blue, nil).Underline()
	g.Draw(klimt.UTextBuild("abc", fc))
	out := write(t, g, "")
	if strings.Contains(out, "underline") || !strings.Contains(out, ">abc</text>") {
		t.Errorf("output %s", out)
	}
}

// Upstream decodes <U+XXXX> in link titles with appendReplacement, which
// throws for '$' and '\' and parses an int that overflows; here they are
// characters (design D16).
func TestLinkTitleEscapes(t *testing.T) {
	tests := []struct{ title, want string }{
		{"a<U+0024>b", "a$b"},
		{"a<U+005C>b", `a\b`},
		{"<U+263a>", "☺"},
		{"<U+D83D><U+DE00>!", "😀!"},
		{"<U+D83D>x", "?x"},
		{"<U+100000041>", "A"},
		{`one\ntwo`, "one\ntwo"},
		{"plain", "plain"},
	}
	for _, tt := range tests {
		if got := (linkData{title: tt.title}).xlinkTitle(); got != tt.want {
			t.Errorf("xlinkTitle(%q) = %q, want %q", tt.title, got, tt.want)
		}
	}
}

// The SVG metadata scenario: the plantuml-src instruction decodes to the
// diagram source.
func TestMetadataDecodes(t *testing.T) {
	source := "@startuml\nAlice -> Bob : héllo & <bye>\n@enduml"
	g := newGraphic(font.FormatSVG)
	g.Apply(color.Black).Draw(klimt.URectangleBuild(5, 5))
	out := write(t, g, source+"\n1.2026.8")
	start := strings.Index(out, "<?plantuml-src ")
	if start < 0 {
		t.Fatalf("no metadata in %s", out)
	}
	rest := out[start+len("<?plantuml-src "):]
	encoded := rest[:strings.Index(rest, "?>")]
	decoded, err := code.Decode(encoded)
	if err != nil || decoded != source {
		t.Errorf("decoded %q, %v", decoded, err)
	}
	if out2 := write(t, newGraphic(font.FormatSVG), ""); strings.Contains(out2, "plantuml-src") {
		t.Errorf("metadata without metadata: %s", out2)
	}
}

func TestMatchesProperty(t *testing.T) {
	g := newGraphic(font.FormatSVG)
	if !g.MatchesProperty("svg") || !g.MatchesProperty("SVG") || g.MatchesProperty("SPECIALTXT") {
		t.Error("MatchesProperty")
	}
	if g.SvgGraphics() != g.GraphicObject() {
		t.Error("SvgGraphics")
	}
}

func TestAssetText(t *testing.T) {
	for _, name := range []string{"default.min.css", "default.min.js", "sequencediagram.min.css",
		"sequencediagram.min.js"} {
		s, ok := assetText(name)
		if !ok || !strings.HasSuffix(s, "\n") || strings.Contains(s, "\r") {
			t.Errorf("asset %s: %v", name, ok)
		}
	}
	if _, ok := assetText("missing.css"); ok {
		t.Error("missing asset")
	}
	// A missing asset leaves the document without style or script.
	g := NewSvgGraphics(1, SvgOptionBasic().WithInteractive("missing"))
	var buf bytes.Buffer
	if err := g.CreateXML(&buf); err != nil || strings.Contains(buf.String(), "<style") {
		t.Errorf("missing interactive asset: %s %v", buf.String(), err)
	}
}

func TestJavaHelpers(t *testing.T) {
	if javaLen("a😀é") != 4 || javaLen("") != 0 {
		t.Error("javaLen")
	}
	if !isJavaBlank(" \t\n\x0B\f\r") || isJavaBlank("  ") || !isJavaBlank("") {
		t.Error("isJavaBlank")
	}
	if !javaEqualsIgnoreCase("MonoSpaced", "monospaced") || javaEqualsIgnoreCase("a", "ab") {
		t.Error("javaEqualsIgnoreCase")
	}
	if seedString(-9223372036854775808) != "-1y2p0ij32e8e8" || seedString(42) != "16" || seedString(-42) != "16" {
		t.Error("seedString")
	}
	if shortenColor("#AABBCC") != "#ABC" || shortenColor("#AABBCD") != "#AABBCD" || shortenColor("none") != "none" {
		t.Error("shortenColor")
	}
	if trimZeros("1.500") != "1.5" || trimZeros("2.000") != "2" || trimZeros("-0.000") != "-0" || trimZeros("30") != "30" {
		t.Error("trimZeros")
	}
}

// Graphics share nothing: concurrent renders give the same output.
func TestConcurrentRenders(t *testing.T) {
	render := func() string {
		g := newGraphic(font.FormatSVG)
		ug := g.Apply(color.Red).Apply(klimt.Bg(color.Gradient(color.White, color.Blue, '|')))
		ug.Draw(klimt.URectangleBuild(10, 20))
		fc := klimt.NewFontConfiguration(font.SansSerif(12), color.Black, color.Blue, nil)
		ug.Apply(klimt.NewUTranslate(3, 15)).Draw(klimt.UTextBuild("concurrent text", fc))
		return write(t, g, "@startuml\na\n@enduml")
	}
	want := render()
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			if got := render(); got != want {
				t.Errorf("concurrent render differs")
			}
		})
	}
	wg.Wait()
}

func TestSvgOptionTitle(t *testing.T) {
	o := SvgOptionBasic().WithTitle(klimt.DisplayEmpty())
	if _, ok := o.Title(); ok {
		t.Error("empty display gives a title")
	}
	o.WithTitle(klimt.DisplayCreate("first", "second"))
	if title, ok := o.Title(); !ok || title != "first\nsecond" {
		t.Errorf("title %q %v", title, ok)
	}
	store := klimt.NewConfigurationStore[klimt.OptionKey]()
	store.Define(klimt.OptionSvgTitle, "option", true)
	store.Define(klimt.OptionSvgDesc, "", false)
	o.WithConfigurationStore(store)
	if title, _ := o.Title(); title != "option" {
		t.Errorf("option title %q", title)
	}
	if _, ok := o.Desc(); ok {
		t.Error("description without value")
	}
	if got := o.WithRootAttribute("a", "1").WithRootAttribute("b", "2").WithRootAttribute("a", "3").
		RootAttributes(); len(got) != 2 || got[0] != [2]string{"a", "3"} {
		t.Errorf("root attributes %v", got)
	}
}
