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
	}
	texts := [][]string{{"label"}, {"first", "--", "second", "==", "third"}, {"a", "--title--", "b", "..", "c"},
		{"first", "----", "second"}}
	noteStyle := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameSequenceDiagram,
		klimt.SNameNote).MergedStyle(sp.CurrentStyleBuilder())
	stereos := []string{"", "<<stereo>>"}
	symbols := []struct {
		name   string
		symbol USymbol
	}{{"DATABASE", Database}, {"QUEUE", Queue}}
	for _, s := range symbols {
		var snames []string
		for _, n := range s.symbol.SNames() {
			snames = append(snames, n.String())
		}
		lines = append(lines, s.name+"\tsnames "+javaList(snames)+" supp="+
			strconv.Itoa(s.symbol.SuppHeightBecauseOfShape())+","+strconv.Itoa(s.symbol.SuppWidthBecauseOfShape()))
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
						for _, kind := range []string{"small", "big"} {
							var log []string
							step(&log, func() {
								var tb klimt.TextBlock
								if kind == "small" {
									tb = s.symbol.AsSmall(nil, label, stereo, fashion, geom.HorizontalAlignmentCenter)
								} else {
									tb = s.symbol.AsBig(label, geom.HorizontalAlignmentLeft, stereo, 150, 90, fashion,
										geom.HorizontalAlignmentCenter)
								}
								dim := tb.CalculateDimension(b.bounder)
								log = append(log, "dim "+d2s(dim.Width())+"x"+d2s(dim.Height()))
								tb.DrawU(newRec(&log, b.bounder).Apply(klimt.NewUTranslate(3, 4)))
							})
							prefix := s.name + "/" + kind + "/" + strconv.Itoa(ti) + "/" + strconv.Itoa(si) + "/" +
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
	compareGolden(t, "symbol.golden", lines)
}
