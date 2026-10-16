// SPDX-License-Identifier: MIT
// Provenance: original

package symbol

import (
	"strconv"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// The golden file compared here is written by
// internal/klimt/rose/testdata/javagen/RoseGolden.java (symbol section).

// symbolNames are the symbols of upstream's USymbols, in the order of the
// harness.
var symbolNames = []string{"ACTION", "ACTOR_AWESOME", "ACTOR_HOLLOW", "ACTOR_STICKMAN",
	"ACTOR_STICKMAN_BUSINESS", "AGENT", "ARCHIMATE", "ARTIFACT", "BOUNDARY", "CARD", "CLOUD", "COLLECTIONS",
	"COMPONENT_RECTANGLE", "COMPONENT1", "COMPONENT2", "CONTROL", "DATABASE", "ENTITY_DOMAIN", "FILE",
	"FOLDER", "FRAME", "GROUP", "HEXAGON", "INTERFACE", "LABEL", "NODE", "PACKAGE", "PARTITION", "PERSON",
	"PROCESS", "QUEUE", "RECTANGLE", "STACK", "STORAGE", "USECASE", "USECASE_BUSINESS"}

func TestGoldenSymbols(t *testing.T) {
	var lines []string
	sp := klimt.SkinParamCreate(map[string]string{}, nil, core.Sequence, klimt.PragmaCreateEmpty(),
		klimt.NewConfigurationStore[klimt.OptionKey]())
	simple := klimt.UStrokeSimple()
	fc := klimt.NewFontConfiguration(font.SansSerif(13), color.Black, color.Blue, &simple)
	fcs := klimt.NewFontConfiguration(font.SansSerif(10), color.Red, color.Blue, &simple)
	fashions := []*klimt.Fashion{
		klimt.NewFashion(color.MyYellow, color.MyRed),
		klimt.NewFashion(color.White, color.Blue).WithShadow(3).WithStroke(klimt.UStrokeWithThickness(2)),
		klimt.NewFashion(color.MyYellow, color.MyRed).WithCorner(10, 0),
		klimt.NewFashion(color.White, color.Blue).WithShadow(3).WithCorner(0, 8),
	}
	texts := [][]string{{"label"}, {"first", "--", "second", "==", "third"}, {"a", "--title--", "b", "..", "c"},
		{"first", "----", "second"}}
	noteStyle := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameSequenceDiagram,
		klimt.SNameNote).MergedStyle(sp.CurrentStyleBuilder())
	stereos := []string{"", "<<stereo>>"}
	magnetic := [][2]float64{{5, 5}, {60, 3}, {60, -2}, {30, -1}, {120, 10}, {20, 0}}
	for _, name := range symbolNames {
		symbol := FromKey(name)
		if symbol == nil {
			t.Fatalf("FromKey(%q) is nil", name)
		}
		var snames []string
		for _, n := range symbol.SNames() {
			snames = append(snames, n.String())
		}
		lines = append(lines, name+"\tsnames "+javaList(snames)+" supp="+
			strconv.Itoa(symbol.SuppHeightBecauseOfShape())+","+strconv.Itoa(symbol.SuppWidthBecauseOfShape()))
		for ti, txt := range texts {
			for si, st := range stereos {
				for fi, fashion := range fashions {
					for _, b := range bounders {
						var label klimt.TextBlock
						if ti == 1 || ti == 2 {
							label = klimt.BodyFactoryCreate3(klimt.DisplayCreate(txt...), sp, geom.HorizontalAlignmentLeft,
								fc, klimt.LineBreakStrategyNone, noteStyle)
						} else {
							label = klimt.DisplayCreate(txt...).Create(fc, geom.HorizontalAlignmentLeft, sp)
						}
						var stereo klimt.TextBlock
						if st == "" {
							stereo = klimt.TextBlockUtilsEmpty(0, 0)
						} else {
							stereo = klimt.DisplayCreate(st).Create(fcs, geom.HorizontalAlignmentCenter, sp)
						}
						title := klimt.DisplayCreate("pkg").Create(fcs, geom.HorizontalAlignmentLeft, sp)
						more := fi == 0 && ti < 2
						withMagnetic := fi == 0 && ti == 0 && si == 1
						for _, kind := range []string{"small", "big", "smallr", "bigr", "bigl", "smalln"} {
							if len(kind) > 5 && !more {
								continue
							}
							if kind == "bigl" && (ti != 0 || fi != 0) {
								continue
							}
							var log []string
							step(&log, func() {
								var tb klimt.TextBlock
								switch kind {
								case "small":
									tb = symbol.AsSmall(title, label, stereo, fashion, geom.HorizontalAlignmentCenter)
								case "smallr":
									tb = symbol.AsSmall(title, label, stereo, fashion, geom.HorizontalAlignmentRight)
								case "smalln":
									tb = symbol.AsSmall(nil, label, stereo, fashion, geom.HorizontalAlignmentLeft)
								case "big":
									tb = symbol.AsBig(label, geom.HorizontalAlignmentLeft, stereo, 150, 90, fashion,
										geom.HorizontalAlignmentCenter)
								case "bigr":
									tb = symbol.AsBig(label, geom.HorizontalAlignmentRight, stereo, 150, 90, fashion,
										geom.HorizontalAlignmentRight)
								default:
									tb = symbol.AsBig(label, geom.HorizontalAlignmentCenter, stereo, 220, 140, fashion,
										geom.HorizontalAlignmentCenter)
								}
								dim := tb.CalculateDimension(b.bounder)
								log = append(log, "dim "+d2s(dim.Width())+"x"+d2s(dim.Height()))
								tb.DrawU(newRec(&log, b.bounder).Apply(klimt.NewUTranslate(3, 4)))
								if withMagnetic {
									for _, pt := range magnetic {
										force := tb.MagneticBorder().ForceAt(b.bounder, geom.NewXPoint2D(pt[0], pt[1]))
										log = append(log, "mb "+d2s(pt[0])+","+d2s(pt[1])+" -> "+d2s(force.Dx())+","+
											d2s(force.Dy()))
									}
								}
							})
							prefix := name + "/" + kind + "/" + strconv.Itoa(ti) + "/" + strconv.Itoa(si) + "/" +
								strconv.Itoa(fi) + "\t" + b.name + "\t"
							for _, l := range log {
								lines = append(lines, prefix+l)
							}
						}
					}
				}
			}
		}
	}
	compareGolden(t, "symbol.golden.gz", lines)
}

func TestStyleSymbols(t *testing.T) {
	for style, want := range map[klimt.PackageStyle]USymbol{
		klimt.PackageStyleNode: Node, klimt.PackageStyleCard: Card, klimt.PackageStyleDatabase: Database,
		klimt.PackageStyleCloud: Cloud, klimt.PackageStyleFrame: Frame, klimt.PackageStyleRectangle: Rectangle,
		klimt.PackageStyleFolder: Package, klimt.PackageStyleAgent: nil, klimt.PackageStyleStorage: nil,
		klimt.PackageStyleComponent1: nil, klimt.PackageStyleComponent2: nil, klimt.PackageStyleArtifact: nil,
	} {
		if got := PackageStyleToUSymbol(style); got != want {
			t.Errorf("PackageStyleToUSymbol(%v) = %v, want %v", style, got, want)
		}
	}
	for style, want := range map[klimt.ComponentStyle]USymbol{klimt.ComponentStyleUML1: Component1,
		klimt.ComponentStyleUML2: Component2, klimt.ComponentStyleRectangle: ComponentRectangle} {
		if got := ComponentStyleToUSymbol(style); got != want {
			t.Errorf("ComponentStyleToUSymbol(%v) = %v, want %v", style, got, want)
		}
	}
	for style, want := range map[klimt.ActorStyle]USymbol{klimt.ActorStyleStickman: ActorStickman,
		klimt.ActorStyleAwesome: ActorAwesome, klimt.ActorStyleHollow: ActorHollow,
		klimt.ActorStyleStickmanBusiness: ActorStickmanBusiness} {
		if got := ActorStyleToUSymbol(style); got != want {
			t.Errorf("ActorStyleToUSymbol(%v) = %v, want %v", style, got, want)
		}
	}
	if FromKey("") != nil || FromKey("UNKNOWN") != nil {
		t.Error("FromKey of an empty or unknown key is not nil")
	}
	// Symbols of the same class with other parameters are distinct, as
	// upstream's instances.
	distinct := []USymbol{Folder, Package, Frame, Group, Partition, Rectangle, Agent, Archimate,
		ComponentRectangle, Usecase, UsecaseBusiness, ActorStickman, ActorAwesome, ActorHollow}
	for i, a := range distinct {
		for j, b := range distinct {
			if i != j && a == b {
				t.Errorf("symbols %d and %d are equal", i, j)
			}
		}
	}
}

func TestJavaRandom(t *testing.T) {
	// Values of new java.util.Random(42).nextDouble() and of seed 0.
	r := newJavaRandom(42)
	want := []float64{0.7275636800328681, 0.6832234717598454, 0.30871945533265976}
	for i, w := range want {
		if got := r.nextDouble(); got != w {
			t.Errorf("Random(42) value %d = %v, want %v", i, got, w)
		}
	}
	if got := newJavaRandom(0).nextDouble(); got != 0.730967787376657 {
		t.Errorf("Random(0) = %v", got)
	}
}
