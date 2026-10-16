// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"strconv"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// The golden file compared here, svek.golden, is written by
// internal/klimt/rose/testdata/javagen/RoseGolden.java (svek section): the
// svek shapes moved to this package for the Rose components.

func TestGoldenSvekShapes(t *testing.T) {
	var lines []string
	sp := newTestSkinParam()
	fc := creoleFC(13)
	fashions := []*Fashion{
		NewFashion(color.MyYellow, color.MyRed),
		NewFashion(color.White, color.Blue).WithShadow(3).WithStroke(UStrokeWithThickness(2)),
	}
	type direction struct {
		d  Direction
		ok bool
	}
	dirs := []direction{{}, {DirectionLeft, true}, {DirectionRight, true}, {DirectionUp, true}, {DirectionDown, true}}
	points := [][4]float64{{10, 12, -25, 8}, {200, -40, 90, 120}, {-30, 300, 40, -20}}
	for _, b := range creoleBounders {
		text := DisplayCreate("note", "second <b>line</b>").Create(fc, geom.HorizontalAlignmentLeft, sp)
		for _, shadow := range []float64{0, 3} {
			for _, rc := range []float64{0, 8} {
				for _, withStroke := range []bool{false, true} {
					for _, withLink := range []bool{false, true} {
						for di, dir := range dirs {
							for pi, pt := range points {
								if (!withLink || di == 0) && pi > 0 {
									continue
								}
								var log []string
								func() {
									defer recoverJava(&log)
									var stroke *UStroke
									if withStroke {
										stroke = strokePtr(UStrokeWithThickness(2))
									}
									o := NewOpale(shadow, color.MyRed, color.MyYellow, text, withLink, stroke)
									o.SetRoundCorner(rc)
									if dir.ok {
										o.SetOpale(dir.d, geom.NewXPoint2D(pt[0], pt[1]), geom.NewXPoint2D(pt[2], pt[3]))
									}
									dim := o.CalculateDimension(b.bounder)
									log = append(log, "dim "+d2s(dim.Width())+"x"+d2s(dim.Height())+" m="+
										strconv.Itoa(o.MarginX1()))
									o.DrawU(newCreoleRec(&log, b.bounder))
								}()
								dirName := "null"
								if dir.ok {
									dirName = dir.d.String()
								}
								prefix := "opale/" + d2s(shadow) + "/" + d2s(rc) + "/" + strconv.FormatBool(withStroke) +
									"/" + strconv.FormatBool(withLink) + "/" + dirName + "/" + strconv.Itoa(pi) + "\t" +
									b.name + "\t"
								for _, l := range log {
									lines = append(lines, prefix+l)
								}
							}
						}
					}
				}
			}
		}
	}
	for _, rc := range []float64{0, 6, 10} {
		lines = append(lines, "opalecorner/"+d2s(rc)+"\t"+describeCreoleShape(OpaleCorner(55.5, rc)))
		lines = append(lines, "opalepolygon/"+d2s(rc)+"\t"+describeCreoleShape(OpalePolygonNormal(55.5, 33.25, rc)))
	}
	for fi, fashion := range fashions {
		for _, kind := range []string{"control", "boundary", "entity"} {
			var log []string
			func() {
				defer recoverJava(&log)
				var tb TextBlock
				switch kind {
				case "control":
					tb = NewControl(fashion)
				case "boundary":
					tb = NewBoundary(fashion)
				default:
					tb = NewEntityDomain(fashion)
				}
				dim := tb.CalculateDimension(fakeBounder{})
				log = append(log, "dim "+d2s(dim.Width())+"x"+d2s(dim.Height()))
				tb.DrawU(newCreoleRec(&log, fakeBounder{}).Apply(NewUTranslate(2, 3)))
			}()
			for _, l := range log {
				lines = append(lines, kind+"/"+strconv.Itoa(fi)+"\t"+l)
			}
		}
	}
	compareGolden(t, "svek.golden", lines)
}
