// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"errors"
	"io"
	"math"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/command/regex"
	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// expectPanic runs f and returns the Java exception class it panics with.
func expectPanic(t *testing.T, f func()) (class, message string) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("no panic")
		}
		re, ok := r.(*color.RuntimeError)
		if !ok {
			t.Fatalf("panic %v (%T), want *color.RuntimeError", r, r)
		}
		class, message = re.Class, re.Message
	}()
	f()
	return
}

func TestUStroke(t *testing.T) {
	s := NewUStroke(7, 3, 2)
	if s.String() != "7.0-3.0-2.0" || s.OnlyThickness() != UStrokeWithThickness(2) {
		t.Errorf("stroke %v / %v", s, s.OnlyThickness())
	}
	if UStrokeSimple() != NewUStroke(0, 0, 1) || UStrokeSimple().DasharraySvg() != nil {
		t.Error("simple stroke")
	}
	if d := s.DasharraySvg(); len(d) != 2 || d[0] != 7 || d[1] != 3 {
		t.Errorf("dasharray %v", d)
	}
	if p, ok := s.DashTikz(); !ok || p != "on 7.0pt off 3.0pt" {
		t.Errorf("tikz %q %v", p, ok)
	}
	// Java's Double.hashCode of the three values summed.
	if h := UStrokeSimple().HashCode(); h != javaDoubleHashCode(1) {
		t.Errorf("hash %d", h)
	}
}

func TestUTranslate(t *testing.T) {
	a := NewUTranslate(3, 4)
	b := UTranslateDx(-1).Compose(UTranslateDy(2))
	if c := a.Compose(b); c.Dx() != 2 || c.Dy() != 6 {
		t.Errorf("compose %v", c)
	}
	if r := a.Reverse(); r.Dx() != -3 || r.Dy() != -4 {
		t.Errorf("reverse %v", r)
	}
	if s := a.Sym(); s.Dx() != 4 || s.Dy() != 3 {
		t.Errorf("sym %v", s)
	}
	if m := a.Scaled(2).MultiplyBy(0.5); m != a {
		t.Errorf("scale %v", m)
	}
	if !a.IsAlmostSame(NewUTranslate(3, 9)) || a.IsAlmostSame(NewUTranslate(1, 9)) {
		t.Error("IsAlmostSame uses an or of the two offsets")
	}
	r := a.Apply(geom.NewXRectangle2D(1, 1, 5, 6))
	if r.X != 4 || r.Y != 5 || r.Width != 5 || r.Height != 6 {
		t.Errorf("apply %v", r)
	}
	if a.String() != "translate dx=3.0 dy=4.0" {
		t.Errorf("string %q", a)
	}
}

// testGraphic is a minimal output graphic on AbstractUGraphic, recording
// the driver calls.
type testGraphic struct {
	AbstractUGraphic[*[]string]
	hidden bool
}

func newTestGraphic(log *[]string) *testGraphic {
	g := &testGraphic{AbstractUGraphic: NewAbstractUGraphic[*[]string](fakeBounder{})}
	g.SetSelf(g)
	g.Copy(color.White, color.IdentityMapper, log)
	g.register()
	return g
}

// register registers the drivers, which read the clip of their graphic as
// upstream's drivers read their ClipContainer.
func (g *testGraphic) register() {
	RegisterDriver(&g.AbstractUGraphic, UDriverFunc[*URectangle, *[]string](
		func(r *URectangle, x, y float64, mapper *color.ColorMapper, param UParam, o *[]string) {
			clip := ""
			if c := g.Clip(); c != nil {
				clip = " " + c.String()
			}
			*o = append(*o, "rect@"+d2s(x)+","+d2s(y)+" "+hc(param.Color())+"/"+hc(param.Backcolor())+" "+
				param.Stroke().String()+clip)
		}))
	IgnoreShape[*UText](&g.AbstractUGraphic)
}

func (g *testGraphic) CopyUGraphic() CommonUGraphic {
	c := &testGraphic{AbstractUGraphic: NewAbstractUGraphic[*[]string](fakeBounder{}), hidden: g.hidden}
	c.SetSelf(c)
	c.CopyFrom(&g.AbstractUGraphic)
	c.register()
	return c
}

func (g *testGraphic) WriteToStream(io.Writer, string, int) error { return nil }

func TestAbstractUGraphic(t *testing.T) {
	var log []string
	g := newTestGraphic(&log)
	ug := g.Apply(NewUTranslate(10, 20)).Apply(color.Red).Apply(Bg(color.Blue)).Apply(NewUStroke(1, 2, 3))
	ug.Apply(NewUTranslate(1, 1)).Draw(URectangleBuild(5, 5))
	ug.Apply(CopyForegroundColorToBackgroundColor{}).Draw(URectangleBuild(5, 5))
	ug.Apply(UHiddenHidden).Draw(URectangleBuild(5, 5)) // hidden: bounds only
	ug.Draw(UTextBuild("ignored", testFC(10)))
	ug.Draw(NewUComment("c"))
	ug.Draw(NewUEmpty(1, 1))
	want := []string{"rect@11.0,21.0 #FF0000/#0000FF 1.0-2.0-3.0", "rect@10.0,20.0 #FF0000/#FF0000 1.0-2.0-3.0"}
	if strings.Join(log, "|") != strings.Join(want, "|") {
		t.Errorf("log %q", log)
	}
	// The original graphic is unchanged by Apply.
	if g.TranslateX() != 0 || g.Param().Color() != color.None() {
		t.Error("Apply changed the receiver")
	}
	// The bounds are shared by the derived graphics.
	// The hidden rectangle and the text without driver count too: the text
	// ends at 10 + 7 * 10 * 0.6.
	if g.MaxX() != 52 || g.MaxY() != 25 {
		t.Errorf("bounds %v %v", g.MaxX(), g.MaxY())
	}
	class, msg := expectPanic(t, func() { ug.Draw(UEllipseBuild(2, 2)) })
	if class != unsupportedOperation || !strings.Contains(msg, "UEllipse") {
		t.Errorf("missing driver: %s %s", class, msg)
	}
	clipped := g.Apply(NewUTranslate(100, 100)).Apply(NewUClip(0, 0, 5, 5))
	if c := clipped.(*testGraphic).Clip(); c == nil || c.X() != 100 || c.Width() != 5 {
		t.Errorf("clip %v", c)
	}
	log = nil
	clipped.Draw(URectangleBuild(1, 1))
	if len(log) != 1 || !strings.HasSuffix(log[0], "CLIP x=100.0 y=100.0 w=5.0 h=5.0") {
		t.Errorf("the drivers of a copy see its clip: %q", log)
	}
}

func TestUGraphicNull(t *testing.T) {
	g := NewUGraphicNull()
	g.Draw(NewUEmpty(1, 1)) // ignored before the drivers
	class, _ := expectPanic(t, func() { g.Apply(NewUTranslate(1, 1)).Draw(URectangleBuild(1, 1)) })
	if class != unsupportedOperation {
		t.Error(class)
	}
	if _, ok := g.StringBounder().(font.StringBounderAwt); !ok {
		t.Errorf("bounder %T", g.StringBounder())
	}
}

func TestLimitFinderApply(t *testing.T) {
	lf := LimitFinderCreate(fakeBounder{}, false)
	if mm := lf.MinMax(); mm != geom.MinMaxEmpty(true) {
		t.Errorf("empty finder %v", mm)
	}
	ug := lf.Apply(NewUStroke(1, 1, 1)).Apply(UHiddenHidden).Apply(AntiAliasingOff).Apply(color.Red).
		Apply(Bg(color.Red)).Apply(CopyForegroundColorToBackgroundColor{})
	ug.Apply(NewUTranslate(2, 3)).Draw(NewUEmpty(4, 5))
	if mm := lf.MinMax(); mmString(mm) != "2.0,3.0,6.0,8.0" {
		t.Errorf("minmax %s", mmString(mm))
	}
	class, _ := expectPanic(t, func() { lf.Apply(UPatternHorizontalStripe) })
	if class != unsupportedOperation {
		t.Error(class)
	}
	class, _ = expectPanic(t, func() { TextLimitFinderCreate(fakeBounder{}, false).Apply(UHiddenHidden) })
	if class != unsupportedOperation {
		t.Error(class)
	}
	class, _ = expectPanic(t, func() { lf.Draw(UHorizontalLineInfinite(1, 0, 0, '-')) })
	if class != unsupportedOperation {
		t.Error("LimitFinder does not know UHorizontalLine:", class)
	}
	if lf.ColorMapper() != color.IdentityMapper {
		t.Error("identity mapper")
	}
}

func TestUGraphicFilterAndDispatch(t *testing.T) {
	var log []string
	rec := newRecUG(&log)
	f := NewUGraphicFilter(rec, (*ULine)(nil))
	f.Draw(NewULine(1, 2))
	f.Draw(URectangleBuild(1, 1))
	if len(log) != 1 || !strings.HasSuffix(log[0], "line 1.0,2.0") {
		t.Errorf("filter %q", log)
	}
	class, _ := expectPanic(t, func() { f.Apply(NewUTranslate(1, 1)) })
	if class != unsupportedOperation {
		t.Error(class)
	}
	log = nil
	d := NewUGraphicDispatchDrawable(rec).Apply(NewUTranslate(5, 0))
	d.Draw(TextBlockUtilsWithMargin(raw(10, "x"), 1, 1))
	if len(log) != 2 || !strings.Contains(log[0], "empty") || !strings.Contains(log[1], "text 'x'") {
		t.Errorf("dispatch %q", log)
	}
}

func TestTextBlockMemoized(t *testing.T) {
	calls := 0
	var m TextBlockMemoized
	slow := func(font.StringBounder) geom.XDimension2D { calls++; return geom.NewXDimension2D(1, 2) }
	m.Dimension(fakeBounder{}, slow)
	m.Dimension(fakeBounder{}, slow)
	if calls != 1 {
		t.Errorf("calls %d", calls)
	}
	m.Dimension(font.StringBounderAwt{}, slow)
	if calls != 2 {
		t.Errorf("a new bounder type recomputes: %d", calls)
	}
	m.InvalidateDimensionCache()
	m.Dimension(font.StringBounderAwt{}, slow)
	if calls != 3 {
		t.Errorf("after invalidation: %d", calls)
	}
}

func TestTextBlockUtilsEmpty(t *testing.T) {
	sb := fakeBounder{}
	e := TextBlockUtilsEmptyTextBlock()
	b := raw(10, "x")
	if TextBlockUtilsMergeLR(e, b, geom.VerticalAlignmentTop) != b || TextBlockUtilsMergeTB(b, e,
		geom.HorizontalAlignmentLeft) != b {
		t.Error("merging with EMPTY_TEXT_BLOCK returns the other block")
	}
	if !TextBlockUtilsIsEmpty(nil, sb) || !TextBlockUtilsIsEmpty(TextBlockUtilsEmpty(0, 0), sb) ||
		TextBlockUtilsIsEmpty(b, sb) {
		t.Error("IsEmpty")
	}
	if TextBlockUtilsWithMargin(b, 0, 0) != b {
		t.Error("no margin returns the block")
	}
	class, _ := expectPanic(t, func() { b.MinMax(sb) })
	if class != unsupportedOperation {
		t.Error(class)
	}
	class, _ = expectPanic(t, func() { NewTextBlockVertical([]TextBlock{b}, geom.HorizontalAlignmentLeft) })
	if class != illegalArgument {
		t.Error(class)
	}
}

func TestShapesPanics(t *testing.T) {
	for name, f := range map[string]func(){
		"rect width":  func() { URectangleBuild(0, 1) },
		"rect height": func() { URectangleBuild(1, 0) },
		"empty":       func() { NewUEmpty(0, 5) },
		"motif":       func() { NewUMotif("x!") },
		"hline style": func() { UHorizontalLineInfinite(1, 0, 0, 0).Stroke() },
	} {
		t.Run(name, func(t *testing.T) { expectPanic(t, f) })
	}
}

func TestPorts(t *testing.T) {
	p := NewPorts()
	p.Add("x", 2, 10, 5)
	p.Add("x", 1, 0, 0) // lower score ignored
	g := p.AllPortGeometry()
	if len(g) != 1 || g[0].Position() != 10 || g[0].LastY() != 15 || g[0].ID() != EncodePortNameToID("x") {
		t.Errorf("ports %v", g)
	}
	if EncodePortNameToID("x") != "p9dd4e461268c8034f5c8564e155c67a6" {
		t.Errorf("md5 id %s", EncodePortNameToID("x"))
	}
}

func TestPragmaAndOptions(t *testing.T) {
	p := PragmaCreateEmpty()
	p.Define("svek-trace", "", false)
	p.Define("Teoz", "on", true)
	p.Define("unknown", "x", true)
	p.Define("svg_font", "", false)
	if v, ok := p.Value(PragmaSvekTrace); !ok || v != "true" || !p.IsTrue(PragmaTeoz) {
		t.Errorf("default value %q %v", v, ok)
	}
	if !p.IsDefine(PragmaSvgFont) {
		t.Error("a pragma without value is defined")
	}
	if _, ok := p.Value(PragmaSvgFont); ok {
		t.Error("a pragma without value has no value")
	}
	p.Undefine(PragmaTeoz)
	if p.IsDefine(PragmaTeoz) {
		t.Error("undefine")
	}
	w := NewWarning("a", "b")
	p.AddWarning(w)
	p.AddWarning(NewWarning("a", "b"))
	p.AddWarning(nil)
	if len(p.Warnings()) != 1 || p.Warnings()[0].AsSingleLine() != "a\nb" {
		t.Errorf("warnings %v", p.Warnings())
	}
	o := NewConfigurationStore[OptionKey]()
	o.Define(OptionHandwritten, "TRUE", true)
	o.Define(OptionLanguage, "", false)
	if !o.IsTrue(OptionHandwritten) || !o.IsDefine(OptionLanguage) {
		t.Error("options")
	}
	o.Undefine(OptionHandwritten)
	if o.IsDefine(OptionHandwritten) {
		t.Error("undefine option")
	}
	if k, ok := OptionKeyLazyFrom("svg-desc"); !ok || k != OptionSvgDesc {
		t.Error("lazy option")
	}
}

func TestSkinParamValueCacheQuirk(t *testing.T) {
	sp := newTestSkinParam()
	if _, ok := sp.Value("dpi"); ok {
		t.Fatal("dpi is not set")
	}
	sp.SetParam("dpi", "120")
	// As upstream, the missing value read before SetParam stays cached.
	if _, ok := sp.Value("dpi"); ok {
		t.Error("the value cache is not cleared by SetParam")
	}
	if sp.Dpi() != 96 {
		t.Errorf("dpi %d", sp.Dpi())
	}
	sp2 := newTestSkinParam()
	sp2.SetParam("dpi", "120")
	if sp2.Dpi() != 120 {
		t.Errorf("dpi %d", sp2.Dpi())
	}
}

func TestSkinParamSizeHackOrder(t *testing.T) {
	sp := newTestSkinParam()
	sp.SetSvgSize("ab", "X")
	sp.SetSvgSize("a", "Y")
	sp.SetSvgSize("<U+00E9>", "e")
	// HashMap order: "a" (bucket 1) is replaced before "ab" (bucket 1 too,
	// but inserted first), then "é".
	keys := javaHashMapOrder(sp.svgCharSizes.keys)
	got := sp.TransformStringForSizeHack("abé")
	want := "abé"
	for _, k := range keys {
		want = strings.ReplaceAll(want, k, sp.svgCharSizes.values[k])
	}
	if got != want {
		t.Errorf("%q != %q", got, want)
	}
}

func TestSkinParamDelegators(t *testing.T) {
	sp := newTestSkinParam()
	sp.SetParam("noteBackgroundColor", "red")
	b := NewSkinParamBackcolored(sp, color.Green, color.Blue, false)
	if hc(b.HtmlColor(ColorParamNoteBackground, nil, false)) != "#00FF00" || hc(b.BackgroundColor()) != "#0000FF" {
		t.Error("backcolored")
	}
	b.ForceColor(ColorParamNoteBorder, color.Gray)
	if hc(b.HtmlColor(ColorParamNoteBorder, nil, false)) != "#808080" {
		t.Error("forced")
	}
	colors, err := color.NewColors("#back:yellow;text:navy;shadowing:false", color.Instance(), color.TypeBack)
	if err != nil {
		t.Fatal(err)
	}
	c := ColorsMute(colors, sp)
	if hc(c.HtmlColor(ColorParamNoteBackground, nil, false)) != "#FFFF00" ||
		hc(c.FontHtmlColor(nil, FontParamNote)) != "#000080" || c.Shadowing(nil) {
		t.Error("colors")
	}
	if hc(c.HtmlColor(ColorParamHyperlink, nil, false)) != "null" {
		t.Error("parameters without color type are delegated")
	}
	// Upstream's shadowingForNote recurses forever; here it delegates.
	if c.ShadowingForNote(nil) != sp.ShadowingForNote(nil) {
		t.Error("shadowingForNote")
	}
	if hc(NewSkinParamForecolored(sp, color.Red).HtmlColor(ColorParamBackground, nil, false)) != "#FF0000" {
		t.Error("forecolored")
	}
	if hc(NewSkinParamForceColor(sp, ColorParamArrowHead, color.Red).HtmlColor(ColorParamArrowHead, nil,
		false)) != "#FF0000" {
		t.Error("force color")
	}
	r := NewSkinParamBackcoloredReference(sp, color.Red, nil)
	if hc(r.HtmlColor(ColorParamSequenceReferenceHeaderBackground, nil, false)) != "#FF0000" ||
		r.HtmlColor(ColorParamSequenceReferenceBackground, nil, false) != nil {
		t.Error("reference")
	}
}

func TestColorsHelpers(t *testing.T) {
	sp := newTestSkinParam()
	sp.SetParam("classBackgroundColor<<foo>>", "pink")
	colors, err := ColorsApplyStereotype(color.EmptyColors(), StereotypeBuild("<<foo>>"), sp,
		ColorParamClassBackground)
	if err != nil {
		t.Fatal(err)
	}
	if hc(colors.Color(color.TypeBack)) != "#FFC0CB" {
		t.Errorf("stereotype color %s", hc(colors.Color(color.TypeBack)))
	}
	dashed := color.EmptyColors().AddLegacyStroke("dashed")
	if s := ColorsSpecificLineStroke(dashed); s == nil || *s != NewUStroke(7, 7, 1) {
		t.Errorf("dashed %v", s)
	}
	if ColorsSpecificLineStroke(color.EmptyColors()) != nil {
		t.Error("no line style")
	}
	if ColorsMuteStroke(color.EmptyColors().AddLegacyStroke("bold"), UStrokeSimple()) != UStrokeWithThickness(2) {
		t.Error("bold")
	}
	if ColorsMuteStroke(color.EmptyColors().AddLegacyStroke("hidden"), UStrokeSimple()) != UStrokeSimple() {
		t.Error("hidden keeps the stroke")
	}
}

func TestStyleParserErrors(t *testing.T) {
	if _, err := NewStyleParser().Parse(command.SingleString("root { FontSize 12 } {")); err == nil {
		t.Error("an open bracket without selector is an error")
	} else {
		var pe *StyleParsingError
		if !errors.As(err, &pe) || pe.Message != "Invalid open bracket" {
			t.Errorf("error %v", err)
		}
	}
	if _, err := NewStyleParser().Parse(command.SingleString("root { FontSize : : }")); err == nil {
		t.Error("bad definition")
	}
	// A key at the end of the text makes upstream throw NullPointerException.
	class, _ := expectPanic(t, func() { NewStyleParser().ParseSingleLine("foo") })
	if class != nullPointer {
		t.Error(class)
	}
	styles, err := NewStyleParser().Parse(command.NewBlocLines(nil))
	if err != nil || styles != nil {
		t.Error("empty sheet")
	}
}

func TestStyleLoader(t *testing.T) {
	if _, err := StyleLoaderLoadSkin(nil, "missing.skin"); err == nil {
		t.Error("missing skin")
	} else {
		var ns *NoStyleAvailableError
		if !errors.As(err, &ns) {
			t.Errorf("error %T", err)
		}
	}
	strict := mustLoadSkin(t, "strictuml.skin")
	if missing := StyleLoaderMissingRootProperties(strict); len(missing) != 8 {
		t.Errorf("strictuml is a fragment: %v", missing)
	}
	ctx := core.NewRenderContext()
	ctx.FS = fstest.MapFS{
		"dir/plantuml.skin": {Data: []byte("root {\n FontSize 99\n}\n")},
		"dir/legacy.skin":   {Data: []byte("skinparam foo bar\n")},
	}
	ctx.CurrentDir = "dir"
	b, err := StyleLoaderLoadSkin(ctx, "plantuml.skin")
	if err != nil {
		t.Fatal(err)
	}
	if v := StyleSignatureBasicOf(SNameRoot).MergedStyle(b).Value(PNameFontSize).AsString(); v != "99" {
		t.Errorf("a local skin wins: %q", v)
	}
	if _, err := StyleLoaderLoadSkin(ctx, "legacy.skin"); err == nil {
		t.Error("a file without style")
	}
	// A skin parameter set falls back on the default skin.
	sp := newTestSkinParam()
	sp.SetDefaultSkin("missing.skin")
	if sp.CurrentStyleBuilder() == nil {
		t.Error("fallback")
	}
}

func TestLoadBlocLines(t *testing.T) {
	got := loadBlocLines("a\r\nb\rc\n\nd", nil).Strings()
	if strings.Join(got, "|") != "a|b|c||d" {
		t.Errorf("%q", got)
	}
	if n := loadBlocLines("a\n", nil).Size(); n != 1 {
		t.Errorf("trailing newline: %d", n)
	}
}

// styleTestDiagram is a diagram for the style commands.
type styleTestDiagram struct {
	command.Base
	sp *SkinParam
}

func (*styleTestDiagram) Description() string     { return "test" }
func (d *styleTestDiagram) SkinParam() ISkinParam { return d.sp }

func TestStyleCommands(t *testing.T) {
	ctx := core.NewRenderContext()
	ctx.FS = fstest.MapFS{"mystyle.css": {Data: []byte("note {\n BackGroundColor lime\n}\n")}}
	d := &styleTestDiagram{sp: SkinParamCreate(nil, ctx, core.Sequence, PragmaCreateEmpty(),
		NewConfigurationStore[OptionKey]())}
	note := StyleSignatureBasicOf(SNameRoot, SNameElement, SNameNote)
	back := func() string {
		return d.sp.CurrentStyleBuilder().MergedStyle(note).Value(PNameBackGroundColor).AsString()
	}

	single := NewCommandStyleSingleLineCSS[*styleTestDiagram]()
	line := command.SingleString("<style>note{BackGroundColor:red}</style>")
	if single.IsValid(line) != command.OK {
		t.Fatal("single line")
	}
	if r := single.Execute(d, line, command.One); !r.IsOK() || back() != "red" {
		t.Errorf("single %v %q", r.Err(), back())
	}

	multi := NewCommandStyleMultilinesCSS[*styleTestDiagram]()
	block := command.NewBlocLines(nil).Add(command.SingleString("<style>").At(0))
	if multi.IsValid(block) != command.OKPartial {
		t.Error("partial")
	}
	for _, l := range []string{"' a comment", "note {", "  BackGroundColor blue", "}", "</style>"} {
		block = block.Add(command.SingleString(l).At(0))
	}
	if multi.IsValid(block) != command.OK {
		t.Fatal("multi")
	}
	if r := multi.Execute(d, block, command.One); !r.IsOK() || back() != "blue" {
		t.Errorf("multi %v %q", r.Err(), back())
	}
	bad := command.NewBlocLines(nil)
	for _, l := range []string{"<style>", "{", "</style>"} {
		bad = bad.Add(command.SingleString(l).At(0))
	}
	if r := multi.Execute(d, bad, command.One); r.IsOK() || r.Err() != "Error in style definition: Invalid open bracket" {
		t.Errorf("bad sheet %q", r.Err())
	}

	imp := NewCommandStyleImport[*styleTestDiagram]()
	for _, c := range []struct{ line, err, back string }{
		{"<style file=mystyle.css>", "", "lime"},
		{"<style file=\"rose.skin\">", "", "#FBFB77"},
		{"<style file=missing.css>", "Cannot read: missing.css", ""},
		{"<style file=https://example.com/x.css>", "Cannot read: https://example.com/x.css (remote style sheets are " +
			"not supported yet)", ""},
	} {
		l := command.SingleString(c.line)
		if imp.IsValid(l) != command.OK {
			t.Fatalf("import %q", c.line)
		}
		r := imp.Execute(d, l, command.One)
		if r.Err() != c.err {
			t.Errorf("%q: %q", c.line, r.Err())
		}
		if c.back != "" && back() != c.back {
			t.Errorf("%q: back %q", c.line, back())
		}
	}
}

func TestNotPortedPathsPanic(t *testing.T) {
	sp := newTestSkinParam()
	for text, class := range map[string]string{"<math>x</math>": "AtomMath", "<latex>x</latex>": "AtomMath",
		"<qrcode:x>": "createQrcode", "<img:file.png>": "AtomImg", "<img:http://x/y.png>": "AtomImg",
		"<img:data:image/png;spm lib/1>": "AtomImg"} {
		_, msg := expectPanic(t, func() { DisplayCreate(text).Create(testFC(12), geom.HorizontalAlignmentLeft, sp) })
		if !strings.Contains(msg, class) {
			t.Errorf("%s: message %q", text, msg)
		}
	}
	_, msg := expectPanic(t, func() {
		DisplayCreate("{{", "a -> b", "}}").Create(testFC(12), geom.HorizontalAlignmentLeft, sp)
	})
	if !strings.Contains(msg, "EmbeddedDiagram") {
		t.Errorf("message %q", msg)
	}
	_, msg = expectPanic(t, func() {
		DisplayCreate("<latex>", "x", "</latex>").Create(testFC(12), geom.HorizontalAlignmentLeft, sp).
			CalculateDimension(fakeBounder{})
	})
	if !strings.Contains(msg, "AtomMath") {
		t.Errorf("message %q", msg)
	}
	_, msg = expectPanic(t, func() { SpriteImageFromInternal("archimate/actor") })
	if !strings.Contains(msg, "SpriteImage") {
		t.Errorf("message %q", msg)
	}
	if SpriteImageFromInternal("nothing") != nil {
		t.Error("unknown sprites are nil")
	}
	_, msg = expectPanic(t, func() { NewTextBlockInEllipse(raw(10, "x"), fakeBounder{}) })
	if !strings.Contains(msg, "Footprint") {
		t.Errorf("message %q", msg)
	}
	_, msg = expectPanic(t, func() { NewUbrexURLBuilder("", URLModeStrict) })
	if !strings.Contains(msg, "UbrexUrlBuilder") {
		t.Errorf("message %q", msg)
	}
}

func TestJavaHelpers(t *testing.T) {
	for s, want := range map[string]float64{"1.5": 1.5, " 2 ": 2, "1e2": 100, "3d": 3, "-0.5": -0.5, ".5": .5,
		"5.": 5} {
		if got := javaParseDouble(s); got != want {
			t.Errorf("%q: %v", s, got)
		}
	}
	for _, s := range []string{"", ".", "0x10", "1_0", "Inf", "1.2.3"} {
		if _, ok := javaTryParseDouble(s); ok {
			t.Errorf("%q parses", s)
		}
	}
	if !math.IsNaN(javaParseDouble("NaN")) {
		t.Error("NaN")
	}
	class, _ := expectPanic(t, func() { javaParseInt("99999999999") })
	if class != numberFormat {
		t.Error(class)
	}
	if javaLen("a😀") != 3 || javaCharAt("😀", 1) != 0xDE00 {
		t.Error("UTF-16")
	}
	if javaStringHashCode("polygenelubricants") != math.MinInt32 {
		t.Errorf("hash %d", javaStringHashCode("polygenelubricants"))
	}
	if !javaEqualsIgnoreCase("ı", "I") || javaEqualsIgnoreCase("a", "b") {
		t.Error("equalsIgnoreCase")
	}
	if got := strings.Join(javaTokenizeWithDelims("a\tb\t\tc", "\t"), "|"); got != "a|\t|b|\t|\t|c" {
		t.Errorf("tokens %q", got)
	}
	if got := javaSplit("a::b::", "::"); len(got) != 2 {
		t.Errorf("split %q", got)
	}
}

func TestSmallTypes(t *testing.T) {
	if p, ok := PackageStyleFromString("rect"); !ok || p != PackageStyleRectangle {
		t.Error("rect")
	}
	if _, ok := PackageStyleFromString("nope"); ok {
		t.Error("unknown package style")
	}
	if c, ok := ConditionStyleFromString("foo1"); !ok || c != ConditionStyleInsideDiamond {
		t.Error("condition style")
	}
	if DirectionLazzyValueOf("top") != DirectionUp || DirectionFromChar('<') != DirectionLeft ||
		DirectionRight.Inv() != DirectionLeft || DirectionUp.Clockwise() != DirectionRight {
		t.Error("direction")
	}
	if d, ok := DirectionFromVector(geom.NewXPoint2D(0, 0), geom.NewXPoint2D(0, 5)); !ok || d != DirectionDown {
		t.Error("vector")
	}
	if MarginsUniform(2).Merge(NewMargins(1, 3, 0, 9)).String() != "MARGIN[2.0,3.0,2.0,9.0]" {
		t.Error("margins")
	}
	if NewStereotag("tag").String() != "$tag" {
		t.Error("stereotag")
	}
	expectPanic(t, func() { NewStereotag("$bad") })
	if StereotypePatternRemoveChevronBrackets("<<x>>") != "x" || StereotypePatternRemoveChevronBrackets("y") != "y" {
		t.Error("chevrons")
	}
	if got := StereostylesBuild("<<<a>>> x <<<b>>> <<<a>>>").StyleNames(); strings.Join(got, ",") != "a,b" {
		t.Errorf("stereostyles %v", got)
	}
	if v, ok := VisibilityOf("-x", true); ok {
		t.Errorf("too short: %v", v)
	}
	if v, ok := VisibilityOf("+field", false); !ok || v != VisibilityPublicMethod {
		t.Error("public method")
	}
	if got := VisibilityReplaceVisibilityModifierByUnicodeChar("#x y", true); got != string(charProtectedField)+"x y" {
		t.Errorf("unicode %q", got)
	}
	if TikzFontDistortionFromValue("2;-3", true).String() != "2.0;-3.0" ||
		TikzFontDistortionFromValue("x;1", true).String() != "1.2;4.0" {
		t.Error("tikz")
	}
	if a, ok := AlignmentParamNoteTextAlignment, true; !ok || a.DefaultValue() != geom.HorizontalAlignmentLeft {
		t.Error("alignment param")
	}
	if PaddingParamBox.SkinName() != "boxPadding" || CornerParamCard.RoundKey() != "cardroundcorner" {
		t.Error("params")
	}
	if NewSplitParam(color.Red, nil, 0).ExternalMargin() != 1 {
		t.Error("split param")
	}
	if ArrowsAsTo(ArrowsTriangle{}, DirectionLeft).String() != "asToLeft" {
		t.Error("arrows")
	}
	if got := len(SkinParamPossibleValues()); got < 300 {
		t.Errorf("possible values %d", got)
	}
}

func TestImages(t *testing.T) {
	svg := NewUImageSvg(`<?xml version="1.0"?><svg xmlns="x" style="width:10px;background:#FFF;" `+
		`viewBox="0 0 12.2 7"><g><rect/></g></svg>`, 2)
	if svg.Width() != 26 || svg.Height() != 14 {
		t.Errorf("svg size %v %v", svg.Width(), svg.Height())
	}
	if got := svg.Svg(false); !strings.HasPrefix(got, `<svg><g><rect fill="#FFF" style="width:10px;background:#FFF;" `+
		`width="13" height="7"/> <rect/>`) {
		t.Errorf("svg %q", got)
	}
	noBox := NewUImageSvg(`<svg width="30" height="40"></svg>`, 1)
	if noBox.Width() != 30 || noBox.Height() != 40 {
		t.Error("attributes")
	}
	expectPanic(t, func() { NewUImageSvg("<svg></svg>", 1).Width() })
	tikz := NewUImageTikz("\\scalebox{0.5}{\n\\begin{tikzpicture}\n\\href{x}{y}\n", 10, 20)
	if tikz.Width() != 5 || tikz.Height() != 10 || !tikz.HasURL() {
		t.Error("tikz")
	}
	img := NewPortableImageARGB(2, 1)
	img.SetRGB(0, 0, int32(-0x1000000)) // opaque black
	img.SetRGB(1, 0, 0x00FFFFFF)        // transparent white
	u := NewUImage(NewPixelImage(img, AffineTransformBilinear))
	if u.Width() != 1 || u.Height() != 0 {
		t.Errorf("image %d %d", u.Width(), u.Height())
	}
	muted := u.MuteColor(color.NewXColor(255, 0, 0)).Image(1)
	// As upstream, the opaque new color plus the alpha overflows: 0xFFFF0000
	// + 0xFF000000 is 0xFEFF0000.
	if muted.RGB(0, 0) != int32(-0x1010000) || muted.RGB(1, 0) != 0x00FFFFFF {
		t.Errorf("mute %x %x", muted.RGB(0, 0), muted.RGB(1, 0))
	}
	if tr := u.MuteTransparentColor(color.NewXColor(0, 0, 255)).Image(1); tr.RGB(1, 0) != color.NewXColor(0, 0,
		255).RGB() {
		t.Errorf("transparent %x", tr.RGB(1, 0))
	}
	if !strings.HasPrefix(img.ToPngDataURL(), "data:image/png;base64,iVBOR") {
		t.Error("png")
	}
	_, msg := expectPanic(t, func() { u.Scale(2).Image(1) })
	if !strings.Contains(msg, "PortableImageAwt") {
		t.Errorf("scale %q", msg)
	}
}

func TestConcurrentStyles(t *testing.T) {
	var wg sync.WaitGroup
	results := make([]string, 8)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sp := newTestSkinParam()
			sp.SetParam("participantBackgroundColor", "red")
			style := StyleSignatureBasicOf(SNameRoot, SNameElement, SNameSequenceDiagram, SNameParticipant).
				MergedStyle(sp.CurrentStyleBuilder())
			b := mustLoadSkin(t, "plantuml.skin")
			shared := StyleSignatureBasicOf(SNameRoot, SNameDocument)
			var wg2 sync.WaitGroup
			for range 4 {
				wg2.Add(1)
				go func() {
					defer wg2.Done()
					b.MergedStyle(shared)
				}()
			}
			wg2.Wait()
			results[i] = styleValuesString(style)
		}(i)
	}
	wg.Wait()
	for _, r := range results[1:] {
		if r != results[0] {
			t.Fatal("concurrent renders differ")
		}
	}
	if !strings.Contains(results[0], "BackGroundColor=red") {
		t.Error(results[0])
	}
}

func TestReviewFixes(t *testing.T) {
	// HashMap doubles its table when a bucket gets a 9th element below 64
	// slots (values from java.util.HashSet on JDK 25).
	keys := []string{"k17", "k309", "k370", "k408", "k480", "k507", "k590", "k606", "k1000"}
	if got := strings.Join(javaHashMapOrder(keys), ","); got != "k309,k408,k507,k606,k17,k370,k480,k590,k1000" {
		t.Errorf("crowded bucket order %s", got)
	}
	// Double.parseDouble accepts hexadecimal literals and rejects suffixed
	// NaN and Infinity.
	if v, ok := javaTryParseDouble("0x1.8p1d"); !ok || v != 3 {
		t.Errorf("hex %v %v", v, ok)
	}
	for _, s := range []string{"NaNd", "Infinityf", "-Infinityd", "0x1.8"} {
		if _, ok := javaTryParseDouble(s); ok {
			t.Errorf("%q parses", s)
		}
	}
	if javaLower("ΟΔΟΣ ΣΑ") != "οδος σα" {
		t.Errorf("final sigma %q", javaLower("ΟΔΟΣ ΣΑ"))
	}
	if javaCharLower('𐐀') != '𐐀' || javaStringCompare("￿", "😀") <= 0 {
		t.Error("UTF-16 helpers")
	}
	re := regex.MustCompile(javaDots(`^a.b$`))
	if !re.MatchString("axb") || re.MatchString("a\rb") || re.MatchString("a b") {
		t.Error("java dot")
	}

	// Delegators force inPackageTitle to false.
	sp := newTestSkinParam()
	if sp.Font(nil, true, FontParamNote).FontFace() != font.BoldFace() {
		t.Error("package title is bold")
	}
	if NewSkinParamForecolored(sp, color.Red).Font(nil, true, FontParamNote).FontFace() != font.NormalFace() {
		t.Error("delegator forces inPackageTitle false")
	}
	// A null arrow direction falls through "direction" to the default.
	sp.SetParam("noteTextAlignment", "direction")
	if a := sp.HorizontalAlignment(AlignmentParamNoteTextAlignment, ArrowDirectionNull, false,
		geom.HorizontalAlignmentCenter); a != geom.HorizontalAlignmentCenter {
		t.Errorf("null direction %v", a)
	}

	// Paths built from another share its curves, as upstream's mutable
	// curve objects.
	p := NewDotPath().AddCurve4(geom.NewXPoint2D(0, 0), geom.NewXPoint2D(1, 0), geom.NewXPoint2D(2, 0),
		geom.NewXPoint2D(3, 0))
	q := p.AddCurve(geom.NewXPoint2D(4, 0), geom.NewXPoint2D(5, 0), geom.NewXPoint2D(6, 0))
	q.MoveStartPoint(0, 2)
	if p.StartPoint().Y != 2 {
		t.Error("shared first curve")
	}
	c := q.Copy()
	c.MoveEndPoint(0, 5)
	if q.EndPoint().Y != 0 {
		t.Error("Copy is deep")
	}

	// The null display panics where upstream uses its null list.
	for name, f := range map[string]func(){
		"add":     func() { DisplayNULL().Add(Text("x")) },
		"replace": func() { DisplayNULL().Replace("a", "b") },
		"get":     func() { DisplayNULL().Get(0) },
		"equals":  func() { DisplayNULL().Equals(DisplayEmpty()) },
	} {
		t.Run(name, func(t *testing.T) {
			if class, _ := expectPanic(t, f); class != nullPointer {
				t.Error(class)
			}
		})
	}
	if DisplayEmpty().EqualsLike(DisplayNULL()) || DisplayEmpty().Equals(DisplayNULL()) {
		t.Error("an empty display is not the null display")
	}
	if class, _ := expectPanic(t, func() { DisplayEmpty().AppendFirstLine("x") }); class != indexOutOfBounds {
		t.Error(class)
	}
	if class, _ := expectPanic(t, func() { DisplayCreate("a").SubList(1, 0) }); class != illegalArgument {
		t.Error(class)
	}
	if class, _ := expectPanic(t, func() { Bg(nil) }); class != nullPointer {
		t.Error(class)
	}

	// Dark-only values have no regular text.
	b := mustLoadSkin(t, "plantuml.skin")
	styles, err := NewStyleParserCounter(b).Parse(command.SingleString("@media dark { root { LineStyle 2-2 } }"))
	if err != nil {
		t.Fatal(err)
	}
	root := StyleSignatureBasicOf(SNameRoot).MergedStyle(b.MuteStyle(styles))
	if class, _ := expectPanic(t, func() { root.Stroke() }); class != nullPointer {
		t.Error(class)
	}
}
