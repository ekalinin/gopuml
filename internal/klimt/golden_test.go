// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/jfmt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// The golden files of testdata/golden are written by the upstream jar
// (testdata/javagen/KlimtGolden.java). Each test computes the same lines
// and compares them.

const goldenDir = "testdata/golden"

func readLines(t *testing.T, name string) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(goldenDir, name))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func compareGolden(t *testing.T, name string, got []string) {
	t.Helper()
	want := readLines(t, name)
	if len(got) != len(want) {
		t.Errorf("%s: %d lines, want %d", name, len(got), len(want))
	}
	failures := 0
	for i := 0; i < min(len(got), len(want)); i++ {
		if got[i] != want[i] {
			failures++
			if failures <= 10 {
				t.Errorf("%s line %d:\n got: %s\nwant: %s", name, i+1, got[i], want[i])
			}
		}
	}
	if failures > 10 {
		t.Errorf("%s: %d more differences", name, failures-10)
	}
}

// esc is the harness's esc: tabs, newlines, backslashes, control and
// private-use characters escaped.
func esc(s string) string {
	var sb strings.Builder
	for _, c := range s {
		switch {
		case c == '\t':
			sb.WriteString(`\t`)
		case c == '\n':
			sb.WriteString(`\n`)
		case c == '\\':
			sb.WriteString(`\\`)
		case c < 32 || c >= 0xE000 && c <= 0xF8FF:
			fmt.Fprintf(&sb, `\u%04X`, c)
		default:
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

func javaList(items []string) string { return "[" + strings.Join(items, ", ") + "]" }

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func d2s(d float64) string { return jfmt.DoubleToString(d) }

func signatureOf(spec string) StyleSignatureBasic {
	sig := StyleSignatureBasicEmpty()
	for _, tok := range strings.Fields(spec) {
		switch {
		case strings.HasPrefix(tok, "@"):
			sig = sig.AddStereotype(tok[1:])
		case strings.HasPrefix(tok, "#"):
			sig = sig.AddLevel(javaParseInt(tok[1:]))
		case tok == "*":
			sig = sig.AddStar()
		default:
			found := false
			for i, n := range snameNames {
				if n == tok {
					sig = sig.AddSName(SName(i))
					found = true
				}
			}
			if !found {
				panic("unknown SName " + tok)
			}
		}
	}
	return sig
}

func canonical(sig StyleSignatureBasic) string {
	var names []string
	for _, n := range sig.Key().SNames() {
		names = append(names, n.String())
	}
	st := sig.Stereotypes()
	slices.Sort(st)
	s := strings.Join(names, ".") + javaList(st)
	if sig.Key().Level() != -1 {
		s += "#" + itoa(sig.Key().Level())
	}
	if sig.Key().IsStared() {
		s += "*"
	}
	return s
}

func styleValuesString(style *Style) string {
	if style == nil {
		return "null"
	}
	var parts []string
	for p := PName(0); p < pnameCount; p++ {
		if style.HasValue(p) {
			parts = append(parts, p.String()+"="+style.Value(p).String())
		}
	}
	return strings.Join(parts, ";")
}

// item is the harness's item: the value, or "ERR" and the simple name of
// the Java exception.
func item(f func() string) (result string) {
	defer func() {
		if r := recover(); r != nil {
			if re, ok := r.(*color.RuntimeError); ok {
				result = "ERR " + re.Class[strings.LastIndexByte(re.Class, '.')+1:]
				return
			}
			result = fmt.Sprintf("ERR go:%v", r)
		}
	}()
	return f()
}

func uf(f font.UFont) string {
	return f.Family("", font.ContextSVG) + "/" + f.FontFace().String() + "/" + itoa(f.Size())
}

func alignString(a geom.HorizontalAlignment) string {
	if a == HorizontalAlignmentNull {
		return "null"
	}
	return a.String()
}

func strokeString(s *UStroke) string {
	if s == nil {
		return "null"
	}
	return s.String()
}

func derived(style *Style, set *color.HColorSet) string {
	if style == nil {
		return "null"
	}
	var sb strings.Builder
	sb.WriteString("font=" + item(func() string { return uf(style.UFont()) }))
	sb.WriteString(" stroke=" + item(func() string { return style.Stroke().String() }))
	sb.WriteString(" padding=" + item(func() string { return style.Padding().String() }))
	sb.WriteString(" margin=" + item(func() string { return style.Margin().String() }))
	sb.WriteString(" align=" + item(func() string { return alignString(style.HorizontalAlignment()) }))
	sb.WriteString(" shadow=" + item(func() string { return d2s(style.Shadowing()) }))
	sb.WriteString(" wrap=" + item(func() string { return style.WrapWidth().String() }))
	sb.WriteString(" back=" + item(func() string { return style.Value(PNameBackGroundColor).AsColor(set).AsString() }))
	sb.WriteString(" line=" + item(func() string { return style.Value(PNameLineColor).AsColor(set).AsString() }))
	sb.WriteString(" fontcolor=" + item(func() string { return style.Value(PNameFontColor).AsColor(set).AsString() }))
	sb.WriteString(" round=" + item(func() string { return d2s(style.Value(PNameRoundCorner).AsDouble()) }))
	sb.WriteString(" fontsizeint=" + item(func() string { return itoa(style.Value(PNameFontSize).AsInt()) }))
	sb.WriteString(" fashion=" + item(func() string {
		f := style.SymbolContext(set, nil)
		return f.Stroke().String() + "/" + d2s(f.DeltaShadow()) + "/" + d2s(f.RoundCorner()) + "/" +
			d2s(f.DiagonalCorner())
	}))
	sb.WriteString(" fc=" + item(func() string {
		fc := style.FontConfiguration(set, nil)
		return uf(fc.Font()) + "/" + fc.Color().AsString() + "/" + strokeString(fc.UnderlineStroke())
	}))
	return sb.String()
}

func checkStyles(t *testing.T, name string, builder *StyleBuilder) {
	t.Helper()
	set := color.Instance()
	var out, out2 []string
	for _, spec := range readLines(t, "signatures.txt") {
		sig := signatureOf(spec)
		style := builder.MergedStyle(sig)
		out = append(out, spec+"\t"+canonical(sig)+"\t"+styleValuesString(style))
		out2 = append(out2, spec+"\t"+derived(style, set))
	}
	compareGolden(t, name+".golden", out)
	compareGolden(t, name+"_derived.golden", out2)
}

func mustLoadSkin(t *testing.T, name string) *StyleBuilder {
	t.Helper()
	b, err := StyleLoaderLoadSkin(nil, name)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestGoldenStyleDefault(t *testing.T) {
	checkStyles(t, "style_default", mustLoadSkin(t, "plantuml.skin"))
}

func TestGoldenStyleOtherSkins(t *testing.T) {
	for _, skin := range []string{"rose", "debug", "sonyxperiadev"} {
		t.Run(skin, func(t *testing.T) { checkStyles(t, "style_"+skin, mustLoadSkin(t, skin+".skin")) })
	}
}

func TestGoldenStyleUser(t *testing.T) {
	user := mustLoadSkin(t, "plantuml.skin")
	data, err := os.ReadFile(filepath.Join(goldenDir, "user_style.css"))
	if err != nil {
		t.Fatal(err)
	}
	lines := loadBlocLines(string(data), nil).Lines()
	parsed, err := NewStyleParserCounter(user).Parse(command.NewBlocLines(lines[1 : len(lines)-1]))
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, s := range parsed {
		out = append(out, canonical(s.Signature())+"\t"+styleValuesString(s))
	}
	compareGolden(t, "style_user_parsed.golden", out)
	checkStyles(t, "style_user", user.MuteStyle(parsed))
}

func TestGoldenStyleSingleLine(t *testing.T) {
	var out []string
	for _, s := range []string{"BackGroundColor:red;", "FontSize: 12; LineColor: blue;",
		"BackgroundColor:#FF0000;LineThickness:2;", "FontColor : #ABC ; FontStyle : italic ;"} {
		st, err := NewStyleParser().ParseSingleLine(s)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, esc(s)+"\t"+styleValuesString(st))
	}
	compareGolden(t, "style_singleline.golden", out)
}

func newTestSkinParam() *SkinParam {
	return SkinParamCreate(map[string]string{}, nil, core.Sequence, PragmaCreateEmpty(),
		NewConfigurationStore[OptionKey]())
}

func TestGoldenFromSkinparam(t *testing.T) {
	var out []string
	for _, line := range readLines(t, "skinparams.txt") {
		key, value, _ := strings.Cut(line, "\t")
		for _, k := range cleanForKeySlow(key) {
			conv := NewFromSkinparamToStyle(k)
			conv.ConvertNow(value, &AutomaticCounterBasic{})
			var parts []string
			for _, s := range conv.Styles() {
				parts = append(parts, canonical(s.Signature())+" "+styleValuesString(s))
			}
			out = append(out, esc(line)+"\t"+k+"\t"+strings.Join(parts, " | "))
		}
	}
	compareGolden(t, "fromskinparam.golden", out)
}

func hc(c color.HColor) string {
	if c == nil {
		return "null"
	}
	return c.AsString()
}

func TestGoldenSkinParam(t *testing.T) {
	sp := newTestSkinParam()
	for _, line := range readLines(t, "skinparams.txt") {
		key, value, _ := strings.Cut(line, "\t")
		sp.SetParam(key, value)
	}
	builder := sp.CurrentStyleBuilder()
	checkStyles(t, "style_skinparam", builder)
	foo := StereotypeBuild("<<foo>>")
	opt := func(s string, ok bool) string {
		if !ok {
			return "null"
		}
		return s
	}
	var missing []string
	for _, p := range StyleLoaderMissingRootProperties(builder) {
		missing = append(missing, p.String())
	}
	out := []string{
		"value participantbackgroundcolor\t" + opt(sp.Value("participantbackgroundcolor")),
		"value missing\t" + opt(sp.Value("missing")),
		"hyperlink\t" + hc(sp.HyperlinkColor()),
		"background\t" + hc(sp.BackgroundColor()),
		"html activityBackground\t" + hc(sp.HtmlColor(ColorParamActivityBackground, nil, false)),
		"html noteBackground foo\t" + hc(sp.HtmlColor(ColorParamNoteBackground, foo, false)),
		"font note\t" + uf(sp.Font(nil, false, FontParamNote)),
		"font note foo\t" + uf(sp.Font(foo, false, FontParamNote)),
		"font package\t" + uf(sp.Font(nil, true, FontParamPackage)),
		"font circled\t" + uf(sp.Font(nil, false, FontParamCircledCharacter)),
		"fontcolor note\t" + hc(sp.FontHtmlColor(nil, FontParamNote)),
		"fontcolor footer\t" + hc(sp.FontHtmlColor(nil, FontParamFooter)),
		"thickness arrow\t" + strokeString(sp.Thickness(LineParamArrow, nil)),
		"thickness legendBorder\t" + strokeString(sp.Thickness(LineParamLegendBorder, nil)),
		"thickness noteBorder\t" + strokeString(sp.Thickness(LineParamNoteBorder, nil)),
		"roundcorner default\t" + d2s(sp.RoundCorner(CornerParamDefault, nil)),
		"roundcorner titleBorder\t" + d2s(sp.RoundCorner(CornerParamTitleBorder, nil)),
		"diagonal rectangle\t" + d2s(sp.DiagonalCorner(CornerParamRectangle, nil)),
		"shadowing\t" + boolStr(sp.Shadowing(nil)),
		"shadowingForNote\t" + boolStr(sp.ShadowingForNote(nil)),
		"strictuml\t" + boolStr(sp.StrictUmlStyle()),
		"underline\t" + strokeString(sp.UseUnderlineForHyperlink()),
		"circledradius\t" + itoa(sp.CircledCharacterRadius()),
		"align seqmsg\t" + alignString(sp.HorizontalAlignment(AlignmentParamSequenceMessageAlignment,
			ArrowDirectionLeftToRightNormal, false, HorizontalAlignmentNull)),
		"align seqmsgtext\t" + alignString(sp.HorizontalAlignment(AlignmentParamSequenceMessageTextAlignment,
			ArrowDirectionRightToLeftReverse, false, HorizontalAlignmentNull)),
		"align note\t" + alignString(sp.HorizontalAlignment(AlignmentParamNoteTextAlignment,
			ArrowDirectionLeftToRightNormal, false, HorizontalAlignmentNull)),
		"padding\t" + sp.Padding().String(),
		"dpi\t" + itoa(sp.Dpi()),
		"tabsize\t" + itoa(sp.TabSize()),
		"guillemet\t" + sp.Guillemet().ManageGuillemet("<<x>>"),
		"forceUnderline\t" + boolStr(sp.ForceSequenceParticipantUnderlined()),
		"padder\t" + boolStr(sp.SequenceDiagramPadder() == PadderNone()),
		"splitparam\t" + boolStr(sp.SplitParam().IsSet()),
		"componentstyle\t" + sp.ComponentStyle().String(),
		"packagestyle\t" + sp.PackageStyle().String(),
		"conditionstyle\t" + sp.ConditionStyle().String(),
		"maxmessagesize\t" + sp.MaxMessageSize().String(),
		"lengthadjust\t" + sp.LengthAdjust().String(),
		"missingroot\t" + javaList(missing),
	}
	compareGolden(t, "skinparam.golden", out)
}

func TestGoldenDisplay(t *testing.T) {
	var out []string
	for _, line := range readLines(t, "display_inputs.txt") {
		d := DisplayWithNewlines(PragmaCreateEmpty(), line)
		var parts []string
		for _, cs := range d.AsList() {
			parts = append(parts, esc(cs.String()))
		}
		out = append(out, esc(line)+"\t"+itoa(d.Size())+"\t"+strings.Join(parts, "|")+"\t"+
			alignString(d.NaturalHorizontalAlignment())+"\t"+esc(javaList(DisplayWithNewlines3(line, true)))+"\t"+
			boolStr(d.HasSeveralGuideLines())+"\t"+boolStr(d.HasURL())+"\t"+boolStr(d.IsWhite()))
	}
	compareGolden(t, "display.golden", out)
}

func TestGoldenStereotypes(t *testing.T) {
	var out []string
	for _, line := range readLines(t, "stereotypes.txt") {
		simple := StereotypeBuild(line)
		complex := item(func() string {
			c, err := StereotypeBuildSpotted(line, 6, font.Monospace(17), color.Instance())
			if err != nil {
				return "ERR NoSuchColorException"
			}
			return c.String() + "|" + itoa(int(c.Character())) + "|" + hc(c.HtmlColor()) + "|" +
				javaList(c.StyleNames()) + "|" + javaList(c.Labels(GuillemetGuillemet()))
		})
		label := func(g *Guillemet) string { return simple.labelOrNull(g) }
		packageStyle := "null"
		if p, ok := simple.PackageStyle(); ok {
			packageStyle = p.String()
		}
		out = append(out, esc(line)+"\t"+simple.String()+"\t"+label(GuillemetGuillemet())+"\t"+
			label(GuillemetDoubleComparator())+"\t"+javaList(simple.Labels(GuillemetNone()))+"\t"+
			javaList(simple.MultipleLabels())+"\t"+javaList(simple.StyleNames())+"\t"+packageStyle+"\t"+
			boolStr(simple.IsWithOOSymbol())+"\t"+boolStr(simple.IsSpotted())+"\t"+complex+"\t"+
			canonical(StyleSignatureBasicOf(SNameRoot, SNameElement).AddStereotypeOf(simple)))
	}
	compareGolden(t, "stereotypes.golden", out)
}

func urlString(u *URL) string {
	if u == nil {
		return "null"
	}
	return u.URL() + "|" + u.Tooltip() + "|" + u.Label()
}

func TestGoldenURLs(t *testing.T) {
	var out []string
	for _, line := range readLines(t, "urls.txt") {
		out = append(out, line+"\t"+urlString(NewURLBuilder("", false, URLModeAnywhere).URL(line))+"\t"+
			urlString(NewURLBuilder("", false, URLModeStrict).URL(line))+"\t"+
			urlString(NewURLBuilder("http://top/", true, URLModeAnywhere).URL(line)))
	}
	compareGolden(t, "urls.golden", out)
}

func TestGoldenCleanForKey(t *testing.T) {
	var out []string
	for _, line := range readLines(t, "cleanforkey_inputs.txt") {
		out = append(out, line+"\t"+javaList(cleanForKeySlow(line)))
	}
	compareGolden(t, "cleanforkey.golden", out)
}

// fakeBounder is the harness's StringBounder: 0.6 em per UTF-16 unit and a
// height of 1.25 em.
type fakeBounder struct{ font.StringBounderDefaults }

func (fakeBounder) CalculateDimension(f font.UFont, text string) font.Dimension {
	return font.Dimension{Width: float64(javaLen(text)) * f.Size2D() * 0.6, Height: f.Size2D() * 1.25}
}

func (fakeBounder) FileFormat() font.FileFormat { return font.FormatPNG }

func testFC(size int) *FontConfiguration {
	return NewFontConfiguration(font.SansSerif(size), color.Black, color.Blue, strokePtr(UStrokeSimple()))
}

func raw(size int, lines ...string) TextBlock { return NewTextBlockRaw(lines, testFC(size)) }

func mmString(m geom.MinMax) string {
	return d2s(m.MinX()) + "," + d2s(m.MinY()) + "," + d2s(m.MaxX()) + "," + d2s(m.MaxY())
}

func blockString(tb TextBlock) string {
	sb := fakeBounder{}
	d := tb.CalculateDimension(sb)
	return d2s(d.Width()) + "x" + d2s(d.Height()) + "\t" + mmString(TextBlockUtilsMinMax(tb, sb, false)) + "\t" +
		mmString(TextBlockUtilsMinMax(tb, sb, true))
}

// shapesBlock is the harness's ShapesBlock.
type shapesBlock struct{ TextBlockDefaults }

func (shapesBlock) DrawU(ug UGraphic) {
	r := URectangleBuild(10, 20)
	r.SetDeltaShadow(3)
	ug.Apply(NewUTranslate(5, 5)).Draw(r)
	ug.Apply(NewUTranslate(-3, 2)).Draw(UEllipseBuild(8, 8))
	p := NewUPolygon()
	p.AddPoint(0, 0)
	p.AddPoint(15, 4)
	p.AddPoint(3, 9)
	ug.Apply(NewUTranslate(40, 1)).Draw(p)
	ug.Apply(NewUTranslate(7, 30)).Draw(NewULine(12, -5))
	path := UPathNone()
	path.MoveTo(0, 0)
	path.LineTo(10, 3)
	path.ArcTo(5, 5, 0, 0, 1, 20, 8)
	path.CubicTo(25, 10, 30, 12, 31, -4)
	ug.Apply(NewUTranslate(50, 20)).Draw(path)
	dp := NewDotPath().AddCurve4(geom.NewXPoint2D(0, 0), geom.NewXPoint2D(5, -8), geom.NewXPoint2D(9, 12),
		geom.NewXPoint2D(14, 3))
	ug.Apply(NewUTranslate(60, 40)).Draw(dp)
	ug.Apply(NewUTranslate(2, 50)).Draw(UTextBuild("text", testFC(10)))
	ug.Apply(NewUTranslate(80, 60)).Draw(NewUEmpty(4, 6))
	clipped := ug.Apply(NewUClip(0, 0, 70, 70))
	clipped.Apply(NewUTranslate(65, 65)).Draw(URectangleBuild(20, 20))
	clipped.Apply(NewUTranslate(10, 10)).Draw(URectangleBuild(5, 5))
}

func (shapesBlock) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(1, 1)
}

func TestGoldenTextBlocks(t *testing.T) {
	t1 := raw(14, "Hello", "World!", "")
	out := []string{
		"raw\t" + blockString(t1),
		"error\t" + blockString(GraphicStringsCreateForError([]string{"Syntax error", "line 2"}, true)),
		"mergeTB\t" + blockString(TextBlockUtilsMergeTB(t1, TextBlockUtilsEmpty(30, 40),
			geom.HorizontalAlignmentCenter)),
		"mergeTBright\t" + blockString(TextBlockUtilsMergeTB(raw(10, "a"), raw(12, "bbbbbb"),
			geom.HorizontalAlignmentRight)),
		"mergeLR\t" + blockString(TextBlockUtilsWithMargin(
			TextBlockUtilsMergeLR(raw(13, "A"), raw(11, "BBB\tC"), geom.VerticalAlignmentBottom), 5, 7)),
		"mergeLRcenter\t" + blockString(TextBlockUtilsMergeLR(raw(20, "big"), raw(8, "s"),
			geom.VerticalAlignmentCenter)),
		"bordered\t" + blockString(TextBlockUtilsBordered(raw(12, "note"), strokePtr(UStrokeWithThickness(1.5)),
			color.Red, color.White, 5, ClockwiseTopRightBottomLeftOf(2, 3, 4, 5), "id")),
		"minwidth\t" + blockString(TextBlockUtilsWithMinWidth(raw(12, "x"), 100, geom.HorizontalAlignmentRight)),
		"minwidthcenter\t" + blockString(TextBlockUtilsWithMinWidth(raw(12, "xyz"), 50,
			geom.HorizontalAlignmentCenter)),
		"margin4\t" + blockString(TextBlockUtilsWithMargin4(raw(12, "m"), 1, 2, 3, 4)),
		"bigframe\t" + blockString(NewBigFrame(raw(12, "frame"), raw(12, "line one", "line two"),
			ClockwiseSame(3), NewFashion(color.White, color.Black).WithShadow(2))),
		"bigframeempty\t" + blockString(NewBigFrame(TextBlockUtilsEmpty(0, 0), raw(12, "content"), ClockwiseNone(),
			NewFashion(nil, color.Black))),
		"shapes\t" + blockString(shapesBlock{}),
		"padder\t" + blockString(PadderNone().WithMargin(ClockwiseSame(2)).
			WithPadding(ClockwiseMargin1Margin2(3, 4)).WithBorderColor(color.Red).Apply(raw(12, "padded"))),
		"visibility\t" + blockString(VisibilityProtectedMethod.UBlock(10, color.Black, color.Red, true)),
		"visibilityfield\t" + blockString(VisibilityPackagePrivateField.UBlock(9, color.Black, nil, false)),
		"circled\t" + blockString(NewCircledCharacter('C', 9, font.Monospace(17), color.Green, color.Black,
			color.Black)),
		"arrow\t" + blockString(NewTextBlockArrow(DirectionRight, testFC(13))),
		"arrowup\t" + blockString(NewTextBlockArrow(DirectionUp, testFC(9))),
		"backcolor\t" + blockString(TextBlockUtilsMergeTB(TextBlockUtilsAddBackcolor(raw(12, "colored"), color.Red),
			raw(12, "plain text"), geom.HorizontalAlignmentLeft)),
	}
	tlf := TextLimitFinderCreate(fakeBounder{}, false)
	t1.DrawU(tlf)
	out = append(out, "textlimitfinder\t"+d2s(tlf.MinX())+","+d2s(tlf.MinY())+","+d2s(tlf.MaxX())+","+
		d2s(tlf.MaxY()))
	compareGolden(t, "textblock.golden", out)
}
