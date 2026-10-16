// SPDX-License-Identifier: MIT
// Provenance: original

package svek

import (
	"math"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// A fragment of the SVG written by dot -Tsvg (Graphviz 14.0.1) for a graph
// of two nodes and an edge.
const dotFragment = `<svg width="62pt" height="141pt" viewBox="0.00 0.00 62.00 141.00">
<g id="graph0" class="graph" transform="scale(1 1) rotate(0) translate(4 137)">
<!-- sh0002 -->
<g id="node1" class="node">
<title>sh0002</title>
<polygon fill="none" stroke="#000002" points="47.58,-133 6,-133 6,-85 47.58,-85 47.58,-133"/>
</g>
<!-- sh0002&#45;&gt;sh0003 -->
<g id="edge1" class="edge">
<title>sh0002&#45;&gt;sh0003</title>
<path fill="none" stroke="#000004" d="M26.79,-84.74C26.79,-67.06 26.79,-47.87 26.79,-30.21"/>
<polygon fill="#000004" stroke="#000004" points="30.29,-30.72 26.79,-20.72 23.29,-30.72 30.29,-30.72"/>
</g>
</svg>`

func TestSvgResultParsing(t *testing.T) {
	svg := newSvgResult(dotFragment, newYDelta(141))

	idx := svg.indexOf("<title>sh0002</title>", 0)
	points := svg.substring(idx).extractList(svgPointsEquals)
	if len(points) != 5 {
		t.Fatalf("points %v", points)
	}
	if minPt := svekGetMinXY(points); minPt != geom.NewXPoint2D(6, 8) {
		t.Errorf("min %v, want (6,8)", minPt)
	}
	if maxPt := svekGetMaxXY(points); maxPt != geom.NewXPoint2D(47.58, 56) {
		t.Errorf("max %v, want (47.58,56)", maxPt)
	}

	// The stroke of the edge is found before its fill.
	if got, want := svg.indexFromColor(4), indexFrom(dotFragment, `stroke="#000004"`, 0); got != want {
		t.Errorf("indexFromColor %d, want %d", got, want)
	}
	if svg.indexFromColor(9) != -1 {
		t.Error("an absent color is found")
	}

	idx = svg.indexOf("d=\"", svg.indexFromColor(4))
	end := svg.indexOf("\"", idx+3)
	path := svg.substring2(idx+3, end)
	if !path.isPathConsistent() {
		t.Fatal("path not consistent")
	}
	dp := path.toDotPath()
	if len(dp.Beziers()) != 1 {
		t.Fatalf("beziers %v", dp.Beziers())
	}
	height, y1, y2 := 141.0, -84.74, -30.21
	if s, e := dp.StartPoint(), dp.EndPoint(); s != geom.NewXPoint2D(26.79, y1+height) ||
		e != geom.NewXPoint2D(26.79, y2+height) {
		t.Errorf("path from %v to %v", s, e)
	}

	// The polygons after the edge path: the arrow head, then nothing.
	it := svg.substring(end).pointsWithThisColor(4)
	head := it.next()
	if len(head) != 4 || !it.hasNext() {
		t.Fatalf("head %v", head)
	}
	if rest := it.next(); len(rest) != 0 || it.hasNext() {
		t.Errorf("after the head: %v", rest)
	}
	if none := svg.pointsWithThisColor(9); len(none.next()) != 0 || none.hasNext() {
		t.Error("iterator of an absent color is not empty")
	}

	// A list with a number Java cannot read gives no point.
	if bad := newSvgResult(`points="1,2 x,3"`, newYDelta(0)).extractList(svgPointsEquals); bad != nil {
		t.Errorf("bad list %v", bad)
	}
}

func TestSvekUtils(t *testing.T) {
	if got := svekGetValue(`<ellipse cx="12.5" cy="-3" rx="2"/>`, 0, "cy"); got != -3 {
		t.Errorf("getValue %v", got)
	}
	for _, tc := range []struct {
		pixel float64
		want  string
	}{{72, "1.000000"}, {35, "0.486111"}, {41.577, "0.577458"}, {0, "0.000000"}} {
		if got := PixelToInches(tc.pixel); got != tc.want {
			t.Errorf("PixelToInches(%v) = %s, want %s", tc.pixel, got, tc.want)
		}
	}
	if compareDouble(math.Copysign(0, -1), 0) != -1 || compareDouble(math.NaN(), 1) != 1 || compareDouble(1, 2) != -1 {
		t.Error("compareDouble")
	}
}

func TestLineOfSegments(t *testing.T) {
	var los lineOfSegments
	los.addSegment(0, 10)
	los.addSegment(5, 15)
	los.addSegment(30, 40)
	got := los.solveOverlaps()
	// The first two overlap by 5: the segments after the first move right,
	// then all move left to keep the mean.
	want := []float64{-10.0 / 3, 10 - 10.0/3, 35 - 10.0/3}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-9 {
			t.Errorf("segment %d starts at %v, want %v", i, got[i], want[i])
		}
	}
}
