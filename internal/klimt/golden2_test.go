// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/command/regex"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// The golden files compared here are written by
// testdata/javagen/KlimtGolden2.java.

// recUG is the harness's Rec: a UGraphic that records what is drawn.
type recUG struct {
	log    *[]string
	t      UTranslate
	color  color.HColor
	back   color.HColor
	stroke UStroke
}

var _ UGraphic = (*recUG)(nil)

func newRecUG(log *[]string) *recUG {
	return &recUG{log: log, color: color.Black, back: color.None(), stroke: UStrokeSimple()}
}

func (r *recUG) StringBounder() font.StringBounder { return fakeBounder{} }

type recParam struct{ r *recUG }

func (p recParam) Color() color.HColor     { return p.r.color }
func (p recParam) Backcolor() color.HColor { return p.r.back }
func (p recParam) Stroke() UStroke         { return p.r.stroke }
func (p recParam) IsHidden() bool          { return false }
func (p recParam) Pattern() UPattern       { return UPatternFull }

func (r *recUG) Param() UParam { return recParam{r} }

func (r *recUG) Draw(shape UShape) {
	*r.log = append(*r.log, "@"+d2s(r.t.Dx())+","+d2s(r.t.Dy())+" c="+hc(r.color)+" b="+hc(r.back)+" s="+
		r.stroke.String()+" "+describeShape(shape))
}

func (r *recUG) Apply(change UChange) UGraphic {
	c := *r
	switch ch := change.(type) {
	case UTranslate:
		c.t = ch.Compose(r.t)
	case color.HColor:
		c.color = ch
	case UBackground:
		c.back = ch.BackColor()
	case UStroke:
		c.stroke = ch
	default:
		*r.log = append(*r.log, "apply ?")
		return r
	}
	return &c
}

func (r *recUG) ColorMapper() *color.ColorMapper { return color.IdentityMapper }
func (r *recUG) StartURL(url *URL)               { *r.log = append(*r.log, "startUrl "+url.URL()) }
func (r *recUG) CloseURL()                       { *r.log = append(*r.log, "closeUrl") }
func (r *recUG) StartGroup(group *UGroup) {
	*r.log = append(*r.log, "startGroup "+groupString(group))
}
func (r *recUG) CloseGroup()                                { *r.log = append(*r.log, "closeGroup") }
func (r *recUG) FlushUg()                                   {}
func (r *recUG) MatchesProperty(string) bool                { return false }
func (r *recUG) DefaultBackground() color.HColor            { return color.White }
func (r *recUG) WriteToStream(io.Writer, string, int) error { return nil }

func groupString(g *UGroup) string {
	var parts []string
	for _, e := range g.Entries() {
		parts = append(parts, e.Key.String()+"="+e.Value)
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func segmentsString(p *UPath) string {
	var segs []string
	for _, s := range p.Segments() {
		segs = append(segs, s.String())
	}
	return javaList(segs)
}

func describeShape(shape UShape) string {
	switch s := shape.(type) {
	case *URectangle:
		return "rect " + d2s(s.Width()) + "," + d2s(s.Height()) + " r=" + d2s(s.Rx()) + "," + d2s(s.Ry()) + " sh=" +
			d2s(s.DeltaShadow())
	case *UEllipse:
		return "ellipse " + d2s(s.Width()) + "," + d2s(s.Height()) + " " + d2s(s.Start()) + "," + d2s(s.Extend()) +
			" sh=" + d2s(s.DeltaShadow())
	case *ULine:
		return "line " + d2s(s.DX()) + "," + d2s(s.DY())
	case *UPolygon:
		var pts []string
		for _, pt := range s.Points() {
			pts = append(pts, d2s(pt.X)+":"+d2s(pt.Y))
		}
		return "polygon " + javaList(pts) + " sh=" + d2s(s.DeltaShadow())
	case *UPath:
		return "path " + segmentsString(s) + " sh=" + d2s(s.DeltaShadow())
	case *UText:
		return "text '" + s.Text() + "' " + uf(s.FontConfiguration().Font()) + " " + hc(s.FontConfiguration().Color())
	case *DotPath:
		return "dotpath " + s.String()
	case *UEmpty:
		return "empty " + d2s(s.Width()) + "," + d2s(s.Height())
	case *UCenteredCharacter:
		return "char " + string(s.Char()) + " " + uf(s.Font())
	case *UHorizontalLine:
		return "hline"
	case *SpecialText:
		return "SpecialText"
	}
	return "?"
}

func drawingLines(out *[]string, name string, d func(ug UGraphic)) {
	var log []string
	func() {
		defer func() {
			if r := recover(); r != nil {
				if re, ok := r.(*color.RuntimeError); ok {
					log = append(log, "ERR "+re.Class[strings.LastIndexByte(re.Class, '.')+1:])
					return
				}
				panic(r)
			}
		}()
		d(newRecUG(&log))
	}()
	for _, l := range log {
		*out = append(*out, name+"\t"+l)
	}
}

type testGuideLine struct{}

func (testGuideLine) ArrowDirection() Direction       { return DirectionRight }
func (testGuideLine) ArrowDirectionInRadian() float64 { return 0.7 }

func TestGoldenDrawing(t *testing.T) {
	var out []string
	add := func(name string, d func(ug UGraphic)) { drawingLines(&out, name, d) }
	add("raw", func(ug UGraphic) { raw(14, "Hello", "Tab\there", "").DrawU(ug) })
	add("error", func(ug UGraphic) { GraphicStringsCreateForError([]string{"Syntax", "line 2"}, false).DrawU(ug) })
	add("bordered", func(ug UGraphic) {
		TextBlockUtilsBordered(raw(12, "note"), strokePtr(UStrokeWithThickness(1.5)), color.Red, color.MyYellow, 5,
			ClockwiseTopRightBottomLeftOf(2, 3, 4, 5), "id").DrawU(ug)
	})
	add("borderednoback", func(ug UGraphic) {
		TextBlockUtilsBordered(raw(12, "n"), strokePtr(UStrokeWithThickness(0)), color.Red, color.White, 0,
			ClockwiseSame(1), "").DrawU(ug)
	})
	add("vertical", func(ug UGraphic) {
		TextBlockUtilsMergeTB(TextBlockUtilsAddBackcolor(raw(12, "colored"), color.Red), raw(10, "wider plain"),
			geom.HorizontalAlignmentCenter).DrawU(ug)
	})
	add("horizontal", func(ug UGraphic) {
		TextBlockUtilsMergeLR(raw(20, "Big"), raw(10, "small"), geom.VerticalAlignmentCenter).DrawU(ug)
	})
	add("margin", func(ug UGraphic) { TextBlockUtilsWithMargin(raw(12, "m"), 3, 4).DrawU(ug) })
	add("minwidth", func(ug UGraphic) {
		TextBlockUtilsWithMinWidth(raw(12, "x"), 60, geom.HorizontalAlignmentCenter).DrawU(ug)
	})
	add("linebefore", func(ug UGraphic) {
		NewTextBlockLineBefore(1.5, raw(12, "after"), '=', nil).DrawU(
			UGraphicStencilCreateDim(ug, geom.NewXDimension2D(80, 30)))
	})
	add("linebeforedotted", func(ug UGraphic) {
		NewTextBlockLineBefore(1.5, raw(12, "after"), '.', nil).DrawU(
			UGraphicStencilCreateDim(ug, geom.NewXDimension2D(80, 30)).Apply(NewUTranslate(5, 7)))
	})
	add("linebeforetitle", func(ug UGraphic) {
		NewTextBlockLineBefore(1.5, raw(12, "after"), '-', raw(10, "Title")).DrawU(
			UGraphicStencilCreateDim(ug, geom.NewXDimension2D(100, 30)))
	})
	add("linebeforedefault", func(ug UGraphic) {
		NewTextBlockLineBefore(2.5, raw(12, "after"), 0, nil).DrawU(
			UGraphicStencilCreateDim(ug, geom.NewXDimension2D(50, 30)))
	})
	add("linebeforenostencil", func(ug UGraphic) { NewTextBlockLineBefore(2.5, raw(12, "after"), '_', nil).DrawU(ug) })
	add("bigframe", func(ug UGraphic) {
		NewBigFrame(raw(12, "frame"), raw(12, "line one", "line two"), ClockwiseSame(3),
			NewFashion(color.White, color.Black).WithShadow(2).WithCorner(4, 0)).DrawU(ug)
	})
	add("bigframelongtitle", func(ug UGraphic) {
		NewBigFrame(raw(12, "a much longer frame title"), raw(12, "x"), ClockwiseNone(),
			NewFashion(nil, color.Blue)).DrawU(ug)
	})
	for p := PackageStyle(0); p < packageStyleCount; p++ {
		add("package "+p.String(), func(ug UGraphic) {
			title := geom.NewXDimension2D(20, 10)
			p.DrawU(ug, geom.NewXDimension2D(60, 40), &title, true)
		})
	}
	add("package frame notitle", func(ug UGraphic) {
		PackageStyleFrame.DrawU(ug, geom.NewXDimension2D(60, 40), nil, false)
	})
	for v := VisibilityModifier(0); v < visibilityCount; v++ {
		add("visibility "+v.String(), func(ug UGraphic) { v.UBlock(10, color.Black, color.Green, true).DrawU(ug) })
	}
	add("visibility drawable", func(ug UGraphic) { VisibilityPrivateField.UDrawable(8, color.Red, nil).DrawU(ug) })
	add("circled", func(ug UGraphic) {
		NewCircledCharacter('C', 9, font.Monospace(17), color.Green, color.Black, color.Black).DrawU(ug)
	})
	for d := DirectionRight; d <= DirectionUp; d++ {
		add("arrow "+d.String(), func(ug UGraphic) { NewTextBlockArrow(d, testFC(13)).DrawU(ug) })
	}
	add("arrow2", func(ug UGraphic) { NewTextBlockArrow2(testGuideLine{}, testFC(14)).DrawU(ug) })
	add("padder", func(ug UGraphic) {
		PadderNone().WithMargin(ClockwiseSame(2)).WithPadding(ClockwiseMargin1Margin2(3, 4)).
			WithBackgroundColor(color.Red).WithRoundCorner(6).Apply(raw(12, "padded")).DrawU(ug)
	})
	add("motif", func(ug UGraphic) { NewUMotif("AAGGHH").DrawHorizontal(ug, 2, 3, 2) })
	add("motifrect", func(ug UGraphic) { ug.Draw(NewUMotifInts(-2, 0, 0, 2, 2, 0).Rectangle(20, 10)) })
	add("diagonal", func(ug UGraphic) { ug.Draw(URectangleBuild(30, 20).DiagonalCorner(5)) })
	add("halfrounded", func(ug UGraphic) { ug.Draw(URectangleBuild(30, 20).HalfRounded(8)) })
	add("fashion", func(ug UGraphic) {
		NewFashion(color.Red, nil).WithStroke(NewUStroke(3, 4, 2)).Apply(ug).Draw(URectangleBuild(5, 5))
	})
	add("url", func(ug UGraphic) {
		TextBlockWithURLOf(raw(12, "link"), NewURL("http://x.org", "", false, "", false)).DrawU(ug)
	})
	add("style", func(ug UGraphic) {
		b := mustLoadSkin(t, "plantuml.skin")
		s := StyleSignatureBasicOf(SNameRoot, SNameDocument, SNameLegend).MergedStyle(b)
		s.ApplyStrokeAndLineColor(ug, color.Instance()).Draw(NewULine(10, 0))
		s.SymbolContext(color.Instance(), nil).Apply(ug).Draw(UEllipseBuild(4, 4))
	})
	add("rotate", func(ug UGraphic) {
		p := NewUPolygon()
		p.AddPoint(0, 0)
		p.AddPoint(10, 0)
		p.AddPoint(5, 7)
		pi := math.Pi
		p.Rotate(pi / 6)
		ug.Draw(p)
		ug.Draw(NewULine(10, 3).Rotate(1.1))
		path := UPathNone()
		path.MoveTo(1, 2)
		path.LineTo(8, 3)
		ug.Draw(path.Rotate(-0.4))
		ug.Apply(NewUTranslate(3, 4).Rotate(2.5)).Draw(NewUEmpty(1, 1))
	})
	compareGolden(t, "drawing.golden", out)
}

func TestGoldenHashOrder(t *testing.T) {
	var out []string
	for _, n := range []int{3, 12, 13, 30, 70} {
		var keys []string
		for i := 0; i < n; i++ {
			keys = append(keys, EncodePortNameToID("port"+itoa(i)))
		}
		out = append(out, "ports"+itoa(n)+"\t"+strings.Join(keys, ",")+"\t"+strings.Join(javaHashMapOrder(keys), ","))
	}
	words := []string{"participant", "actor", "boundary", "a", "b", "Aa", "BB", "zz", "<U+00E9>", "x", "é", "日本",
		"ab", "ba", "polygenelubricants", "GydZG_", "DESIGNING WORKHOUSES"}
	out = append(out, "words\t"+strings.Join(words, ",")+"\t"+strings.Join(javaHashMapOrder(words), ","))
	var pts []geom.XPoint2D
	var ins, hashes = []string{}, []int32{}
	for i := 0; i < 20; i++ {
		p := geom.NewXPoint2D(float64(i)*1.5-7, float64((i*7)%11)-0.25)
		pts = append(pts, p)
		ins = append(ins, d2s(p.X)+":"+d2s(p.Y))
		hashes = append(hashes, p.HashCode())
	}
	var ord []string
	for _, idx := range javaHashOrder(hashes) {
		ord = append(ord, d2s(pts[idx].X)+":"+d2s(pts[idx].Y))
	}
	out = append(out, "points\t"+strings.Join(ins, ",")+"\t"+strings.Join(ord, ","))
	dp := NewDotPath().AddCurve4(geom.NewXPoint2D(0, 0), geom.NewXPoint2D(10, -30), geom.NewXPoint2D(40, 50),
		geom.NewXPoint2D(60, 5))
	var sample []string
	for _, p := range dp.Sample() {
		sample = append(sample, d2s(p.X)+":"+d2s(p.Y))
	}
	out = append(out, "sample\t\t"+strings.Join(sample, ","))
	ports := NewPorts()
	ports.Add("a", 1, 10, 5)
	ports.Add("b", 2, 3, 5)
	ports.Add("c", 1, 10, 7)
	ports.Add("a", 0, 1, 1)
	ports.Add("d", 3, 10, 2)
	more := NewPorts()
	more.Add("a", 5, 20, 1)
	more.Add("e", 1, -2, 1)
	ports.AddThis(more.TranslateY(4))
	var geoms []string
	for _, g := range ports.AllPortGeometry() {
		geoms = append(geoms, g.String())
	}
	out = append(out, "portsgeom\t\t"+javaList(geoms)+" "+ports.String())
	compareGolden(t, "hashorder.golden", out)
}

func dispString(d *Display) string {
	var parts []string
	for _, cs := range d.AsList() {
		parts = append(parts, cs.String())
	}
	return itoa(d.Size()) + ":" + strings.Join(parts, "|") + ":" + alignString(d.NaturalHorizontalAlignment())
}

func TestGoldenDisplayOps(t *testing.T) {
	d := DisplayCreate("first : type", "+field", "<<stereo>>", "page %page% of %lastpage%")
	var splits []string
	for _, part := range DisplayCreate("a--b", "c", "d--e--f").SplitMultiline(mustCompileJava("--")) {
		splits = append(splits, dispString(part))
	}
	out := []string{
		"create\t" + dispString(d),
		"underlined\t" + dispString(d.Underlined()),
		"underlinedName\t" + dispString(d.UnderlinedName()),
		"guillemet\t" + dispString(DisplayCreate("+private <<a>> <<b>>", "x <<c>>").ManageGuillemet(true)),
		"guillemetnovis\t" + dispString(DisplayCreate("+private <<a>>", "x").ManageGuillemet(false)),
		"withPage\t" + dispString(d.WithPage(3, 7)),
		"replace\t" + dispString(d.Replace("e", "E")),
		"removeEndingStereotype\t" + dispString(DisplayCreate("a", "name <<st>>").RemoveEndingStereotype()),
		"endingStereotype\t" + DisplayCreate("name << st >>").EndingStereotype().String(),
		"addFirst\t" + dispString(d.AddFirst(Text("zero"))),
		"appendFirstLine\t" + dispString(d.AppendFirstLine(">>")),
		"addGeneric\t" + dispString(DisplayCreate("List").AddGeneric("T")),
		"addGenericEmpty\t" + dispString(DisplayEmpty().AddGeneric("T")),
		"addAll\t" + dispString(d.AddAll(DisplayCreate("x", "y"))),
		"subList\t" + dispString(d.SubList(1, 3)),
		"splitMultiline\t" + strings.Join(splits, " ; "),
		"replaceBackslashT\t" + strings.ReplaceAll(dispString(DisplayCreate(`a\tb`).ReplaceBackslashT()), "\t", "<TAB>"),
		"contentWidth\t" + itoa(DisplayCreate("ab", "日本語x", "😀😀").ContentWidth()),
		"isWhite\t" + boolStr(DisplayCreate("   ").IsWhite()) + "," + boolStr(DisplayCreate(" ", "").IsWhite()) + "," +
			boolStr(DisplayEmpty().IsWhite()),
		"guidelines\t" + boolStr(DisplayHasSeveralGuideLines(`a\n< b`)) + "," +
			boolStr(DisplayHasSeveralGuideLines("a < b")) + "," + boolStr(DisplayCreate("x", "y >").HasSeveralGuideLines()),
		"equals\t" + boolStr(d.Equals(DisplayCreate("first : type", "+field", "<<stereo>>",
			"page %page% of %lastpage%"))) + "," + itoa(int(d.HashCode())) + "," + itoa(int(DisplayNULL().HashCode())) +
			"," + boolStr(DisplayIsNull(DisplayNULL())),
		"tooltip\t" + d.ToTooltipText() + "," + DisplayEmpty().ToTooltipText(),
		"foo\t" + dispString(DisplayCreateFoo([]text.StringLocated{text.NewStringLocated("a", nil),
			text.NewStringLocated("}}", nil), text.NewStringLocated("", nil)})),
	}
	compareGolden(t, "displayops.golden", out)
}

func fcString(f *FontConfiguration) string {
	attrs := f.Attributes()
	var parts []string
	for _, k := range attrs.Keys() {
		parts = append(parts, k+"="+attrs.Attributes()[k])
	}
	return uf(f.Font()) + " " + hc(f.Color()) + " " + hc(f.ExtendedColor()) + " " + itoa(f.Space()) + " " +
		boolStr(f.ContainsStyle(font.FontStyleUnderline)) + " " + strokeString(f.UnderlineStroke()) + " " +
		itoa(f.TabSize()) + " {" + strings.Join(parts, ", ") + "}"
}

func TestGoldenFontConfig(t *testing.T) {
	base := NewFontConfigurationTab(font.Build("Courier", font.BoldFace(), 13), color.Red, color.Blue,
		strokePtr(UStrokeWithThickness(2)), 4)
	mono := font.Monospace(10)
	out := []string{
		"base\t" + fcString(base),
		"italic\t" + fcString(base.Italic()),
		"unbold\t" + fcString(base.Unbold()),
		"plain\t" + fcString(base.Add(font.FontStylePlain)),
		"bigger\t" + fcString(base.Bigger(2.7)),
		"smaller\t" + fcString(base.Bigger(-20)),
		"changeSize\t" + fcString(base.ChangeSize(9.9)),
		"exposant\t" + fcString(base.ChangeFontPosition(font.FontPositionExposant)),
		"indice\t" + fcString(base.ChangeFontPosition(font.FontPositionIndice)),
		"hyperlink\t" + fcString(base.Hyperlink()),
		"wave\t" + fcString(base.Wave(color.Green)),
		"family\t" + fcString(base.ChangeFamily("Serif")),
		"forceFontNull\t" + fcString(base.ForceFont(nil, nil)),
		"forceFont\t" + fcString(base.ForceFont(&mono, color.Gray)),
		"reset\t" + fcString(base.Italic().Bigger(5).ChangeColor(color.Green).ResetFont()),
		"attrs\t" + fcString(base.ChangeAttributes(SvgAttributesBuild(`a=1 b="two words" c = x`))),
		"faceweight\t" + fcString(base.WithFontFace(font.NormalFace().WithWeight(300))),
		"equals\t" + boolStr(base.Equals(base.Italic().Unitalic())) + "," + boolStr(base.Equals(base.Italic())),
		"blackblue\t" + fcString(FontConfigurationBlackBlueTrue(font.Serif(11))),
	}
	compareGolden(t, "fontconfig.golden", out)
}

func TestGoldenStrokes(t *testing.T) {
	root := StyleSignatureBasicOf(SNameRoot).MergedStyle(mustLoadSkin(t, "plantuml.skin"))
	var out []string
	for _, ls := range []string{"", "7;7", "1-3", "x", "5", "-", "2,8", " 3 - 4 ", "1.5;2.5;9", ";;", "1e2-3"} {
		out = append(out, "linestyle\t"+ls+"\t"+item(func() string {
			return root.EventuallyOverrideString(PNameLineStyle, ls).Stroke().String()
		}))
	}
	out = append(out, "overridestroke\t\t"+root.EventuallyOverrideStroke(strokePtr(NewUStroke(4, 2, 3))).Stroke().String())
	for _, p := range []string{"5", "1 2", "1 2 3", "1 2 3 4", "1 2 3 4 5", " 7 ", "", "a", "1.5", "99999999999",
		"  1   2  "} {
		out = append(out, "padding\t"+p+"\t"+ClockwiseRead(p).String())
	}
	for _, l := range []string{"100", "-20", "auto", "AUTO", "", "12px", "007"} {
		out = append(out, "linebreak\t"+l+"\t"+d2s(NewLineBreakStrategy(l).MaxWidth())+","+
			boolStr(NewLineBreakStrategy(l).IsAuto()))
	}
	out = append(out, "linebreak\tnull\t"+d2s(LineBreakStrategyNone.MaxWidth())+","+boolStr(LineBreakStrategyNone.IsAuto()))
	for _, v := range []string{"12", "1.5", "abc", "a1b2", "", "  bold ", "italic", "600", "bolder", "true", "TRUE",
		"center", "1.2.3"} {
		val := root.EventuallyOverrideString(PNameFontSize, v).Value(PNameFontSize)
		out = append(out, "value\t"+v+"\t"+item(func() string { return itoa(val.AsInt()) })+","+
			item(func() string { return itoa(val.AsIntButMinusOneIfError()) })+","+
			item(func() string { return d2s(val.AsDouble()) })+","+item(func() string { return boolStr(val.AsBoolean()) })+
			","+item(func() string { return val.AsFontFace().String() })+","+
			item(func() string { return alignString(val.AsHorizontalAlignment()) })+","+
			item(func() string { return val.AsColor(color.Instance()).AsString() }))
	}
	compareGolden(t, "strokes.golden", out)
}

func TestGoldenDotPath(t *testing.T) {
	p := NewDotPath().AddCurve4(geom.NewXPoint2D(0, 0), geom.NewXPoint2D(10, 0), geom.NewXPoint2D(20, 5),
		geom.NewXPoint2D(30, 5)).AddCurve(geom.NewXPoint2D(40, 5), geom.NewXPoint2D(50, 20), geom.NewXPoint2D(60, 40))
	m := p.Middle()
	moved := p.Copy()
	moved.MoveStartPoint(2, 3)
	moved.MoveEndPoint(-1, 4)
	movedFar := p.Copy()
	movedFar.MoveStartPoint(40, 1)
	delta := p.Copy()
	delta.MoveDelta(5, -5)
	head := geom.NewRectangleArea(50, 15, 70, 50)
	tail := geom.NewRectangleArea(-5, -5, 12, 3)
	ortho := NewDotPath().AddCurve4(geom.NewXPoint2D(0, 0), geom.NewXPoint2D(0, 0), geom.NewXPoint2D(0, 20),
		geom.NewXPoint2D(0, 20)).AddCurve(geom.NewXPoint2D(0, 20), geom.NewXPoint2D(30, 20), geom.NewXPoint2D(30, 20)).
		AddCurve(geom.NewXPoint2D(30, 20), geom.NewXPoint2D(30, 50), geom.NewXPoint2D(30, 50))
	ortho.MuteToRoundOrthogonalPaths(6)
	up := p.ToUPath()
	out := []string{
		"path\t" + p.String(),
		"middle\t" + d2s(m.X()) + "," + d2s(m.Y()) + "," + d2s(m.Angle()),
		"angles\t" + d2s(p.StartAngle()) + "," + d2s(p.EndAngle()),
		"minmax\t" + p.MinMax().String(),
		"mindist\t" + d2s(p.MinDist(geom.NewXPoint2D(25, -3))),
		"reverse\t" + p.Reverse().String(),
		"upath\t" + segmentsString(up) + " " + d2s(up.MinX()) + "," + d2s(up.MinY()) + "," + d2s(up.MaxX()) + "," +
			d2s(up.MaxY()),
		"isline\t" + boolStr(p.IsLine()) + "," + boolStr(NewDotPath().AddCurve4(geom.NewXPoint2D(0, 0),
			geom.NewXPoint2D(1, 1), geom.NewXPoint2D(2, 2), geom.NewXPoint2D(3, 3)).IsLine()),
		"moved\t" + moved.String(),
		"movedfar\t" + movedFar.String(),
		"delta\t" + delta.String(),
		"compound\t" + p.SimulateCompound(&head, &tail).String(),
		"rounded\t" + ortho.String(),
	}
	compareGolden(t, "dotpath.golden", out)
}

func lineString(l geom.XLine2D, ok bool) string {
	if !ok {
		return "null"
	}
	return d2s(l.X1) + "," + d2s(l.Y1) + "," + d2s(l.X2) + "," + d2s(l.Y2)
}

func TestGoldenGeometry(t *testing.T) {
	var out []string
	e := NewUEllipse(30, 20, 0, 0)
	for _, y := range []float64{0, 3, 10, 17.5, 20} {
		out = append(out, "ellipsex\t"+d2s(y)+"\t"+d2s(e.StartingX(y))+","+d2s(e.EndingX(y)))
	}
	pi := math.Pi
	for _, a := range []float64{0, 0.3, 1, pi / 2, 2.5, -1.2, 10} {
		p := e.PointAtAngle(a)
		out = append(out, "ellipseangle\t"+d2s(a)+"\t"+d2s(p.X)+","+d2s(p.Y))
	}
	for _, a := range []float64{0.1, 0.7, 1.3, 2.9, -0.5, 4.4} {
		l := NewULine(13, 7).Rotate(a)
		tr := NewUTranslate(5, -2).Rotate(a)
		out = append(out, "rotate\t"+d2s(a)+"\t"+d2s(l.DX())+","+d2s(l.DY())+","+d2s(tr.Dx())+","+d2s(tr.Dy()))
	}
	out = append(out, "linelength\t\t"+d2s(NewULine(3, 4).Length())+","+d2s(NewULine(0.1, 0.2).Length()))
	at := NewXAffineTransformIdentity()
	at.Translate(3, 4)
	at.Rotate(0.9, 2, 1)
	at.Scale(1.5, 0.5)
	at.Concatenate(XAffineTransformRotateInstance(-0.3))
	out = append(out, "affine\t\t"+at.String())
	tp := at.Transform(geom.NewXPoint2D(7, -3))
	out = append(out, "affinept\t\t"+d2s(tp.X)+","+d2s(tp.Y))
	path := UPathNone()
	path.MoveTo(1, 1)
	path.ArcToPoint(geom.NewXPoint2D(9, 9), 4, 0, 1)
	path.LineTo(2, 8)
	ap := path.Affine(at, 0.25, 2)
	out = append(out, "affinepath\t\t"+segmentsString(ap)+" "+d2s(ap.MinX())+","+d2s(ap.MinY())+","+d2s(ap.MaxX())+
		","+d2s(ap.MaxY()))
	clip := NewUClip(0, 0, 10, 20)
	r := clip.ClippedRectangle(geom.NewXRectangle2D(5, 5, 20, 20))
	out = append(out, "clip\t\t"+lineString(clip.ClippedLine(geom.NewXLine2D(-5, 3, 15, 3)))+" "+
		lineString(clip.ClippedLine(geom.NewXLine2D(4, -5, 4, 30)))+" "+
		lineString(clip.ClippedLine(geom.NewXLine2D(-5, -5, 15, 30)))+" "+
		lineString(clip.ClippedLine(geom.NewXLine2D(-5, -5, -5, 30)))+" "+
		lineString(clip.ClippedLine(geom.NewXLine2D(2, 3, 4, 5)))+" "+
		lineString(clip.ClippedLine(geom.NewXLine2D(3, 3, 30, 3)))+" "+
		d2s(r.X)+","+d2s(r.Y)+","+d2s(r.Width)+","+d2s(r.Height))
	g := NewUGroup()
	g.Put(UGroupTitle, "a/b<c> d-e_f")
	g.Put(UGroupClass, "x😀y")
	out = append(out, "ugroup\t\t"+groupString(g))
	compareGolden(t, "geometry.golden", out)
}

func TestGoldenStereoStyles(t *testing.T) {
	b := mustLoadSkin(t, "plantuml.skin")
	data, err := os.ReadFile(filepath.Join(goldenDir, "user_style.css"))
	if err != nil {
		t.Fatal(err)
	}
	lines := loadBlocLines(string(data), nil).Lines()
	parsed, err := NewStyleParserCounter(b).Parse(command.NewBlocLines(lines[1 : len(lines)-1]))
	if err != nil {
		t.Fatal(err)
	}
	user := b.MuteStyle(parsed)
	part := StyleSignatureBasicOf(SNameRoot, SNameElement, SNameSequenceDiagram, SNameParticipant)
	var out []string
	for _, st := range []string{"<<foo>>", "<<bar>>", "<<foo>><<bar>>", "<<other>>", "<<(C,red)>>"} {
		stereo := StereotypeBuild(st)
		out = append(out, "with\t"+st+"\t"+styleValuesString(part.WithTOBECHANGED(stereo).MergedStyle(user)))
		out = append(out, "itself\t"+st+"\t"+styleValuesString(part.ForStereotypeItself(stereo).MergedStyle(user)))
		var sts []string
		for _, s := range stereo.Styles(user) {
			sts = append(sts, styleValuesString(s))
		}
		out = append(out, "getstyles\t"+st+"\t"+javaList(sts))
	}
	out = append(out, "special\t\t"+styleValuesString(user.MergedStyleSpecial(part, 5000)))
	set := color.Instance()
	for _, g := range []string{"<<#red>>", "<<##blue>>", "<<###green>>", "<<##[dashed]orange>>",
		"<<#back:yellow;line:navy>>", "<<BackGroundColor:pink;FontColor:teal;>>", "<<foo>> <<#AABBCC>>",
		"<<##[bold]>>"} {
		group := StereogroupBuild(g)
		colors, err := group.InnerColors(set)
		if err != nil {
			t.Fatal(err)
		}
		muted, err := group.Mute(part.MergedStyle(user), set)
		if err != nil {
			t.Fatal(err)
		}
		hcolor, err := group.HColor(part, PNameLineColor, user, set)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, "group\t"+g+"\t"+javaList(group.Labels())+" "+hc(colors.Color(color.TypeBack))+","+
			hc(colors.Color(color.TypeLine))+","+hc(colors.Color(color.TypeText))+","+colors.LineStyle().String()+" "+
			hc(muted.Value(PNameBackGroundColor).AsColor(set))+","+hc(muted.Value(PNameLineColor).AsColor(set))+","+
			hc(hcolor))
	}
	compareGolden(t, "stereostyles.golden", out)
}

func mustCompileJava(p string) *regexp.Regexp { return regex.MustCompile(p) }

func unescParserInput(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
			if s[i] == 'n' {
				sb.WriteByte('\n')
			} else {
				sb.WriteByte(s[i])
			}
			continue
		}
		sb.WriteByte(s[i])
	}
	return sb.String()
}

func TestGoldenParser(t *testing.T) {
	var out []string
	for _, line := range readLines(t, "parser_inputs.txt") {
		r := item(func() string {
			var lines []text.StringLocated
			for _, l := range strings.Split(unescParserInput(line), "\n") {
				lines = append(lines, text.NewStringLocated(l, nil))
			}
			styles, err := NewStyleParser().Parse(command.NewBlocLines(lines))
			if err != nil {
				return "ERR StyleParsingException"
			}
			var parts []string
			for _, st := range styles {
				parts = append(parts, canonical(st.Signature())+" "+styleValuesString(st))
			}
			return strings.Join(parts, " | ")
		})
		out = append(out, line+"\t"+r)
	}
	compareGolden(t, "parser.golden", out)
}

func pixString(img PortableImage) string {
	var l []string
	for i := 0; i < img.Width(); i++ {
		l = append(l, strconv.FormatUint(uint64(uint32(img.RGB(i, 0))), 16))
	}
	return strings.Join(l, ",")
}

func TestGoldenPixels(t *testing.T) {
	values := []uint32{0xFF102030, 0x80FFFFFF, 0x00000000, 0xFF000000, 0xFFFFFFFF, 0xFF7F7F7F, 0x40102030}
	img := NewPortableImageARGB(len(values), 1)
	for i, v := range values {
		img.SetRGB(i, 0, int32(v))
	}
	p := NewPixelImage(img, AffineTransformBilinear)
	out := []string{
		"source\t" + pixString(p.Image()),
		"monochrome\t" + pixString(p.Monochrome().Image()),
		"mutecolor\t" + pixString(p.MuteColor(color.NewXColor(10, 200, 30)).Image()),
		"mutetransparent\t" + pixString(p.MuteTransparentColor(color.NewXColor(1, 2, 3)).Image()),
	}
	compareGolden(t, "pixels.golden", out)
}

func TestSmallHelpers(t *testing.T) {
	if ParserScale("x scale=1.5", true, 1) != 1.5 || ParserScale("", false, 2) != 2 || ParserScale("*3", true, 1) != 3 {
		t.Error("scale")
	}
	if c, ok := ParserColor("color:#AABBCC x"); !ok || c != "#AABBCC" {
		t.Error("color")
	}
	if !ParserIsTreeStart("|_ x") || ParserIsTreeStart("|") {
		t.Error("tree")
	}
	for name, a := range map[string]Arrows{"regular": ArrowsRegular{}, "triangle": ArrowsTriangle{}} {
		for d := DirectionRight; d <= DirectionUp; d++ {
			if ArrowsAsTo(a, d).Point(1) != geom.NewXPoint2D(0, 0) && ArrowsAsTo(a, d).Point(2) != geom.NewXPoint2D(0, 0) {
				t.Errorf("%s %v: the tip is at the origin", name, d)
			}
		}
	}
	var log []string
	rec := newRecUG(&log)
	if b, ok := ChangeBack(rec.Apply(color.Red)).(UBackground); !ok || b.BackColor() != color.Red {
		t.Error("ChangeBack")
	}
	at := XAffineTransformScaleInstance(2, 3)
	at.Concatenate(XAffineTransformTranslateInstance(1, 1))
	if at.ScaleX() != 2 || at.ScaleY() != 3 || at.TranslateX() != 2 || at.TranslateY() != 3 {
		t.Error(at)
	}
	d := DisplayCreateSeq([]CharSequence{Text("a"), StereotypeBuild("<<s>>")})
	style := NewStyle(StyleSignatureBasicEmpty(), map[PName]Value{PNameShowStereotype: ValueImplRegularPriority("false", 1)})
	if d.WithoutStereotypeIfNeeded(style).ShowStereotype() || !d.ShowStereotype() || d.StereotypeIfAny() == nil {
		t.Error("show stereotype")
	}
	if DisplayNULL().WithoutStereotypeIfNeeded(style) != DisplayNULL() {
		t.Error("null display")
	}
}
