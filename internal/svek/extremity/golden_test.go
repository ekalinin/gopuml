// SPDX-License-Identifier: MIT
// Provenance: original

package extremity

import (
	"strconv"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/decoration"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// The golden file compared here is written by
// testdata/javagen/ExtremityGolden.java; the loops below follow it.

var backs = []color.HColor{color.White, color.MyYellow, nil}

var triples = [][6]float64{
	{96, 120, 100, 130, 104, 120},
	{104, 50, 100, 40, 96, 50},
	{50, 96, 60, 100, 50, 104},
	{50, 104, 40, 100, 50, 96},
	{10.5, 20.25, 17.3, 31.7, 19.1, 18.4},
	{5, 5, 5, 5, 5, 5},
	{0, 0, 3.5, 7.25, 0.0005, 10},
	{200.123, 12.5, 190.75, 14.125, 199.5, 21.875},
	{30, 30, 31, 31, 30.5, 30},
}

// angles are the harness's ANGLES, computed step by step as javac folds
// them.
var angles = []float64{0, pi / 2, pi, -pi / 4, 1.234, 3 * pi / 2, pi/2 + 0.0005, 2*pi - 0.0001, -2.5, 0.75}

var points = [][2]float64{{0, 0}, {33.5, 47.25}}

func sidePtr(s geom.Side) *geom.Side { return &s }

var allSides = []*geom.Side{nil, sidePtr(geom.SideNorth), sidePtr(geom.SideEast), sidePtr(geom.SideSouth),
	sidePtr(geom.SideWest)}

func sidesOf(d decoration.LinkDecor) []*geom.Side {
	if d == decoration.LinkDecorCrowfoot {
		return allSides
	}
	return []*geom.Side{nil, sidePtr(geom.SideSouth)}
}

func sideName(s *geom.Side) string {
	if s == nil {
		return "null"
	}
	return s.String()
}

// javaName returns the simple Java class name of a factory or decoration.
func javaName(v any) string {
	switch x := v.(type) {
	case nil:
		return "null"
	case interface{ factoryName() string }:
		return x.factoryName()
	case *MiddleFactoryCircle:
		return "MiddleFactoryCircle"
	case *MiddleFactoryCircleCircled:
		return "MiddleFactoryCircleCircled"
	case *MiddleFactorySubset:
		return "MiddleFactorySubset"
	case *Arrow:
		return "ExtremityArrow"
	case *arrowAndCircle:
		return "ExtremityArrowAndCircle"
	case *circle:
		return "ExtremityCircle"
	case *circleConnect:
		return "ExtremityCircleConnect"
	case *circleCross:
		return "ExtremityCircleCross"
	case *circleCrowfoot:
		return "ExtremityCircleCrowfoot"
	case *circleLine:
		return "ExtremityCircleLine"
	case *crowfoot:
		return "ExtremityCrowfoot"
	case *diamond:
		return "ExtremityDiamond"
	case *doubleLine:
		return "ExtremityDoubleLine"
	case *extends:
		return "ExtremityExtends"
	case *redefines:
		return "Redefines"
	case *definedBy:
		return "DefinedBy"
	case *halfArrow:
		return "ExtremityHalfArrow"
	case *lineCrowfoot:
		return "ExtremityLineCrowfoot"
	case *notNavigable:
		return "ExtremityNotNavigable"
	case *other:
		return "ExtremityOther"
	case *parenthesis:
		return "ExtremityParenthesis"
	case *parenthesis2:
		return "ExtremityParenthesis2"
	case *plus:
		return "ExtremityPlus"
	case *square:
		return "ExtremitySquare"
	case *StateLine1:
		return "ExtremityStateLine1"
	case *StateLine2:
		return "ExtremityStateLine2"
	case *triangle:
		return "ExtremityTriangle"
	case *middleCircle:
		return "MiddleCircle"
	case *middleCircleCircled:
		return "MiddleCircleCircled"
	case *middleSubset:
		return "MiddleSubset"
	}
	return "?"
}

func (f abstractFactory) factoryName() string { return f.name }

func pt(p *geom.XPoint2D) string {
	if p == nil {
		return "null"
	}
	return d2s(p.X) + ":" + d2s(p.Y)
}

// states are the harness's graphics states: default, colored with a thick
// stroke, without color.
func states(log *[]string) []*rec {
	return []*rec{
		{log: log, color: color.Black, back: color.None(), stroke: klimt.UStrokeSimple()},
		{log: log, color: color.MyRed, back: color.Blue, stroke: klimt.UStrokeWithThickness(2.5)},
		{log: log, color: nil, back: color.None(), stroke: klimt.UStrokeSimple()},
	}
}

func record(lines *[]string, prefix string, make func() klimt.UDrawable) {
	var log []string
	var u klimt.UDrawable
	step(&log, func() {
		u = make()
		log = append(log, "class "+javaName(u))
	})
	if e, ok := u.(Extremity); ok {
		step(&log, func() { log = append(log, "somePoint "+pt(e.SomePoint())) })
		step(&log, func() { log = append(log, "length "+d2s(e.DecorationLength())) })
		step(&log, func() {
			kal := e.DeltaForKal()
			log = append(log, "kal "+d2s(kal.Dx())+","+d2s(kal.Dy()))
		})
	}
	if u != nil {
		for k, r := range states(&log) {
			log = append(log, "state "+strconv.Itoa(k))
			step(&log, func() { u.DrawU(r.Apply(klimt.NewUTranslate(3, 4))) })
			if a, ok := u.(*Arrow); ok {
				step(&log, func() { a.DrawLineIfTransparent(r) })
			}
		}
	}
	for _, l := range log {
		*lines = append(*lines, prefix+"\t"+l)
	}
}

func recordFactory(lines *[]string, prefix string, f Factory, sides []*geom.Side) {
	for _, s := range sides {
		for i, t := range triples {
			p0, p1, p2 := geom.NewXPoint2D(t[0], t[1]), geom.NewXPoint2D(t[2], t[3]), geom.NewXPoint2D(t[4], t[5])
			record(lines, prefix+"/tbr"+strconv.Itoa(i)+"/"+sideName(s), func() klimt.UDrawable {
				return f.CreateTBRDrawableLegacy(p0, p1, p2, s)
			})
		}
		for j, p := range points {
			for i, angle := range angles {
				p0 := geom.NewXPoint2D(p[0], p[1])
				record(lines, prefix+"/u"+strconv.Itoa(j)+"_"+strconv.Itoa(i)+"/"+sideName(s),
					func() klimt.UDrawable { return f.CreateUDrawable(p0, angle, s) })
			}
		}
	}
}

func TestGoldenExtremities(t *testing.T) {
	var lines []string
	for d := decoration.LinkDecorNone; d <= decoration.LinkDecorSquareToBeRemoved; d++ {
		for _, back := range backs {
			legacy := FactoryLegacy(d, back)
			complete := FactoryComplete(d, back)
			lines = append(lines, "decor "+d.String()+" back="+hc(back)+" legacy="+javaName(legacy)+" complete="+
				javaName(complete))
			if legacy != nil {
				recordFactory(&lines, d.String()+"/"+hc(back)+"/legacy", legacy, sidesOf(d))
			}
			if complete != nil && d == decoration.LinkDecorExtends {
				recordFactory(&lines, d.String()+"/"+hc(back)+"/complete", complete, sidesOf(d))
			}
		}
	}
	for _, back := range backs {
		recordFactory(&lines, "extends/"+hc(back), NewFactoryExtends(back), allSides)
	}

	middleColors := [][2]color.HColor{{color.White, color.White}, {color.MyYellow, color.Blue}, {nil, nil}}
	for d := decoration.LinkMiddleDecorNone; d <= decoration.LinkMiddleDecorSuperset; d++ {
		for _, c := range middleColors {
			prefix := "middle " + d.String() + "/" + hc(c[0]) + "/" + hc(c[1])
			var log []string
			var f MiddleFactory
			step(&log, func() {
				f = MiddleFactoryFor(d, c[0], c[1])
				log = append(log, "factory "+javaName(f))
			})
			for _, l := range log {
				lines = append(lines, prefix+"\t"+l)
			}
			if f == nil {
				continue
			}
			for i, a := range angles {
				angle := a * 57
				record(&lines, prefix+"/"+strconv.Itoa(i), func() klimt.UDrawable { return f.CreateUDrawable(angle) })
			}
		}
	}

	for i, a := range angles {
		c := geom.NewXPoint2D(12.5, 7.75)
		n := strconv.Itoa(i)
		record(&lines, "arrow/"+n, func() klimt.UDrawable { return NewArrow(c, a) })
		record(&lines, "arrowcenter/"+n, func() klimt.UDrawable { return NewArrowCenter(c, a, geom.NewXPoint2D(20, 30)) })
		record(&lines, "arrowshort/"+n, func() klimt.UDrawable { return NewArrowCenter(c, a, geom.NewXPoint2D(13, 8)) })
		record(&lines, "stateline1/"+n, func() klimt.UDrawable { return NewStateLine1(a, c) })
		record(&lines, "stateline2/"+n, func() klimt.UDrawable { return NewStateLine2(a, c) })
		record(&lines, "parenthesis2/"+n, func() klimt.UDrawable {
			return newParenthesis2(c, a, geom.NewXPoint2D(40.5, 3.25))
		})
		record(&lines, "halfarrow/"+n, func() klimt.UDrawable { return newHalfArrow(c, a, -1) })
	}
	poly := klimt.NewUPolygon()
	poly.AddPoint(1, 2)
	poly.AddPoint(5.5, 7.25)
	poly.AddPoint(0, 9)
	record(&lines, "other", func() klimt.UDrawable { return NewOther(poly) })
	record(&lines, "otherempty", func() klimt.UDrawable { return NewOther(klimt.NewUPolygon()) })
	record(&lines, "triangle/null", func() klimt.UDrawable {
		return newTriangle(geom.NewXPoint2D(3, 4), 0.5, false, nil, 8, 3, 8)
	})

	compareGolden(t, "extremity.golden.gz", lines)
}

func TestAbstractFactoryCreateUDrawable(t *testing.T) {
	// Every factory overrides createUDrawable; upstream's default throws
	// UnsupportedOperationException naming the class.
	var log []string
	step(&log, func() { abstractFactory{"ExtremityFactoryFoo"}.CreateUDrawable(geom.NewXPoint2D(0, 0), 0, nil) })
	want := "ERR UnsupportedOperationException class net.sourceforge.plantuml.svek.extremity.ExtremityFactoryFoo"
	if len(log) != 1 || log[0] != want {
		t.Errorf("got %q, want %q", log, want)
	}
}
