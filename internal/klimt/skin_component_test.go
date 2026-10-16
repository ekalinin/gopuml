// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"strconv"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// The golden file compared here is written by
// testdata/javagen/CreoleGolden.java (skin section).

func arrowString(a *ArrowConfiguration) string {
	b := strconv.FormatBool
	return a.Name() + " dir=" + a.ArrowDirection().String() + " self=" + b(a.IsSelfArrow()) + " dotted=" +
		b(a.IsDotted()) + " hidden=" + b(a.IsHidden()) + " head=" + a.Head().String() + " async=" + b(a.IsAsync1()) +
		"," + b(a.IsAsync2()) + " part=" + a.Part().String() + " incl=" + strconv.Itoa(a.Inclination1()) + "," +
		strconv.Itoa(a.Inclination2()) + " rev=" + b(a.IsReverseDefine()) + " deco=" + a.Decoration1().String() +
		"," + a.Decoration2().String() + " d1=" + a.Dressing1().Head().String() + "/" + a.Dressing1().Part().String() +
		" d2=" + a.Dressing2().Head().String() + "/" + a.Dressing2().Part().String() + " color=" + hc(a.Color())
}

// testComponent is the harness's TestComponent.
type testComponent struct {
	AbstractTextualComponent
	log *[]string
}

func (c *testComponent) DrawInternalU(ug UGraphic, area *Area) {
	*c.log = append(*c.log, "internal "+area.String())
	c.TextBlock().DrawU(ug)
}

func (c *testComponent) DrawBackgroundInternalU(ug UGraphic, area *Area) {
	*c.log = append(*c.log, "background "+area.String())
	ug.Draw(URectangleBuildDim(area.DimensionToUse()))
}

func (c *testComponent) PaddingX() float64 { return 3 }

func (c *testComponent) PreferredWidth(stringBounder font.StringBounder) float64 {
	return c.TextWidth(stringBounder) + 1
}

func (c *testComponent) PreferredHeight(stringBounder font.StringBounder) float64 {
	return c.TextHeight(stringBounder) + 2
}

func (c *testComponent) sizes(sb font.StringBounder) string {
	return "pure=" + d2s(c.PureTextWidth(sb)) + " text=" + d2s(c.TextWidth(sb)) + "x" + d2s(c.TextHeight(sb)) +
		" pad=" + d2s(c.OldPaddingX1()) + "," + d2s(c.OldPaddingX2()) + "," + d2s(c.OldPaddingY()) + " pref=" +
		c.PreferredDimension(sb).String()
}

func TestGoldenSkinComponents(t *testing.T) {
	var out []string
	normal := ArrowConfigurationWithDirectionNormal()
	arrows := []*ArrowConfiguration{
		normal,
		ArrowConfigurationWithDirectionBoth(),
		ArrowConfigurationWithDirectionSelf(false),
		ArrowConfigurationWithDirectionSelf(true),
		ArrowConfigurationWithDirectionReverse(),
		normal.Self(),
		normal.WithBody(ArrowBodyDotted),
		normal.WithBody(ArrowBodyHidden).WithHead(ArrowHeadAsync),
		normal.Reverse().WithHead(ArrowHeadCrossx),
		normal.WithHead1(ArrowHeadAsync),
		normal.WithHead2(ArrowHeadNone),
		normal.WithHead2(ArrowHeadNone).WithHead1(ArrowHeadNone),
		normal.WithPart(ArrowPartTopPart),
		normal.WithHead2(ArrowHeadNone).WithPart(ArrowPartBottomPart),
		normal.WithDecoration1(ArrowDecorationCircle).WithDecoration2(ArrowDecorationCircle).Reverse(),
		normal.WithColor(color.Red).WithThickness(3),
		normal.ReverseDefine().ReverseDefine().ReverseDefine(),
		normal.WithInclination(4),
		normal.WithInclination(4).WithHead2(ArrowHeadCrossx).WithHead1(ArrowHeadNormal),
		normal.WithInclination(-2).Reverse().WithHead(ArrowHeadAsync),
		ArrowConfigurationWithDirectionBoth().WithInclination(5).WithHead1(ArrowHeadCrossx),
	}
	for _, a := range arrows {
		out = append(out, "arrow\t"+arrowString(a))
	}

	builder := mustLoadSkin(t, "plantuml.skin")
	arrowStyle := StyleSignatureBasicOf(SNameRoot, SNameElement, SNameSequenceDiagram, SNameArrow).MergedStyle(builder)
	noteStyle := StyleSignatureBasicOf(SNameRoot, SNameElement, SNameSequenceDiagram, SNameNote).MergedStyle(builder)
	for idx, a := range []*ArrowConfiguration{normal, normal.WithBody(ArrowBodyDotted), normal.WithThickness(2.5),
		normal.WithBody(ArrowBodyDotted).WithThickness(1.5)} {
		var log []string
		ug := newCreoleRec(&log, fakeBounder{})
		a.ApplyStroke(ug, arrowStyle).Draw(NewULine(10, 0))
		a.ApplyStroke(ug, noteStyle).Draw(NewULine(0, 10))
		a.ApplyStrokeLegacy(ug).Draw(NewULine(1, 1))
		a.ApplyThicknessOnly(ug).Draw(NewULine(2, 2))
		ArrowConfigurationStroke(ug, 5, 3, 2).Draw(NewULine(3, 3))
		for _, l := range log {
			out = append(out, "stroke "+strconv.Itoa(idx)+"\t"+l)
		}
	}

	area := AreaCreate(30.5, 12)
	areaLine := func() string {
		return "area\t" + area.String() + " " + area.DimensionToUse().String() + " " + d2s(area.DeltaX1()) + " " +
			strconv.Itoa(area.Level()) + " " + d2s(area.LiveDeltaSize()) + " " + d2s(area.TextDeltaX())
	}
	out = append(out, areaLine())
	area.SetDeltaX1(2.5)
	area.SetLevel(3)
	area.SetLiveDeltaSize(4.25)
	area.SetTextDeltaX(-1)
	out = append(out, areaLine())

	fashions := []*Fashion{
		NewFashion(color.MyYellow, color.MyRed),
		NewFashion(color.White, color.Blue).WithShadow(3).WithStroke(UStrokeWithThickness(2)),
		NewFashion(nil, color.Black).WithStroke(NewUStroke(2, 3, 0.5)),
	}
	for _, style := range []ActorStyle{ActorStyleStickman, ActorStyleStickmanBusiness, ActorStyleAwesome,
		ActorStyleHollow} {
		for i, f := range fashions {
			var log []string
			tb := style.TextBlock(f)
			dim := tb.CalculateDimension(fakeBounder{})
			log = append(log, "dim "+d2s(dim.Width())+"x"+d2s(dim.Height()))
			tb.DrawU(newCreoleRec(&log, fakeBounder{}))
			for _, l := range log {
				out = append(out, "actor "+style.String()+" "+strconv.Itoa(i)+"\t"+l)
			}
		}
	}
	for _, thickness := range []float32{1.5, 0.3} {
		var log []string
		ci := NewCircleInterface(color.MyYellow, color.MyRed, 10.5, thickness)
		log = append(log, "pref "+d2s(ci.PreferredWidth(fakeBounder{}))+"x"+d2s(ci.PreferredHeight(fakeBounder{})))
		ci.DrawU(newCreoleRec(&log, fakeBounder{}))
		name := strconv.FormatFloat(float64(thickness), 'f', -1, 32)
		for _, l := range log {
			out = append(out, "circle "+name+"\t"+l)
		}
	}

	for _, typ := range ComponentTypeValues() {
		var log []string
		func() {
			defer recoverJava(&log)
			st := typ.StyleSignature().MergedStyle(builder)
			javaString := func(name PName) string {
				if s, ok := valueString(st.Value(name)); ok {
					return s
				}
				return "null"
			}
			log = append(log, javaString(PNameBackGroundColor)+" "+javaString(PNameLineColor)+" "+
				javaString(PNameFontSize)+" "+javaString(PNameLineThickness))
		}()
		out = append(out, "componenttype\t"+typ.String()+" "+strconv.FormatBool(typ.IsArrow())+" "+log[0])
	}

	padding := ClockwiseTopRightBottomLeftOf(2, 7, 3, 5)
	type componentCase struct {
		name     string
		display  *Display
		label    string
		hasLabel bool
		enhanced bool
	}
	cases := []componentCase{
		{name: "label", label: `Some **label**\nsecond`, hasLabel: true},
		{name: "emptylabel", label: "", hasLabel: true},
		{name: "nulllabel"},
		{name: "display", display: DisplayCreate("plain", "<b>bold</b>")},
		{name: "enhanced", display: DisplayCreate("head", "--", "after line", "==title==", "end"), enhanced: true},
		{name: "enhancedplain", display: DisplayCreate("one", "two"), enhanced: true},
		{name: "stereo", display: DisplayCreate("text").Add(StereotypeBuild("<<st>>"))},
	}
	for _, c := range cases {
		for _, b := range creoleBounders {
			var log []string
			func() {
				defer recoverJava(&log)
				comp := &testComponent{log: &log}
				if c.display == nil {
					comp.InitAbstractTextualComponentLabel(comp, noteStyle, LineBreakStrategyNone, padding,
						newTestSkinParam(), c.label, c.hasLabel)
				} else {
					comp.InitAbstractTextualComponent(comp, noteStyle, noteStyle, LineBreakStrategyNone, padding,
						newTestSkinParam(), c.display, c.enhanced)
				}
				log = append(log, comp.sizes(b.bounder))
				ug := newCreoleRec(&log, b.bounder)
				comp.DrawU(ug, AreaCreate(40, 20), NewSimpleContext2D(true))
				comp.DrawU(ug, AreaCreate(40, 20), NewSimpleContext2D(false))
			}()
			for _, l := range log {
				out = append(out, "component "+c.name+"\t"+b.name+"\t"+l)
			}
		}
	}
	compareGolden(t, "creole_skin.golden", out)
}

func TestSkinComponentInterfaces(t *testing.T) {
	var _ Component = (*testComponent)(nil)
	var _ Context2D = NewSimpleContext2D(false)
	comp := &testComponent{}
	class, _ := expectPanic(t, func() { comp.UsedStyles() })
	if class != unsupportedOperation {
		t.Error(class)
	}
	class, _ = expectPanic(t, func() { comp.StyleSignature() })
	if class != unsupportedOperation {
		t.Error(class)
	}
	class, _ = expectPanic(t, func() { ActorStyle(9).TextBlock(NewFashion(nil, nil)) })
	if class != illegalState {
		t.Error(class)
	}
	if ArrowDressingCreate().Name() != "NONE" || ArrowBody(7).String() != "ArrowBody(7)" {
		t.Error("names")
	}
	root := StyleSignatureBasicOf(SNameRoot).MergedStyle(mustLoadSkin(t, "plantuml.skin"))
	body := BodyFactoryCreate3(DisplayCreate("a"), newTestSkinParam(), geom.HorizontalAlignmentLeft, testFC(10),
		LineBreakStrategyNone, root)
	if w := body.CalculateDimension(fakeBounder{}).Width(); w != 6 {
		t.Error(w)
	}
}
