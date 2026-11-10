// SPDX-License-Identifier: MIT
// Provenance: original

package jsondiagram

import (
	"math"
	"testing"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// drawn is a shape and the translation it is drawn at.
type drawn struct {
	shape klimt.UShape
	at    klimt.UTranslate
}

// recorder is a UGraphic that records the shapes drawn on it.
type recorder struct {
	*klimt.LimitFinder
	shapes *[]drawn
}

func newRecorder() recorder {
	return recorder{LimitFinder: klimt.LimitFinderCreate(font.NewStringBounderSvg(nil), false), shapes: new([]drawn)}
}

func (r recorder) Apply(change klimt.UChange) klimt.UGraphic {
	return recorder{LimitFinder: r.LimitFinder.Apply(change).(*klimt.LimitFinder), shapes: r.shapes}
}

func (r recorder) Draw(shape klimt.UShape) {
	*r.shapes = append(*r.shapes, drawn{shape: shape, at: r.Translate()})
	r.LimitFinder.Draw(shape)
}

// samePoints reports whether the coordinates are the same, to 1e-9.
func samePoints(got, want []float64) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if math.Abs(got[i]-want[i]) > 1e-9 {
			return false
		}
	}
	return true
}

// checkPath checks the segments of a path: their types and coordinates.
func checkPath(t *testing.T, name string, shape klimt.UShape, types []geom.USegmentType, coords [][]float64) {
	t.Helper()
	path, ok := shape.(*klimt.UPath)
	if !ok {
		t.Fatalf("%s: drawn %T, want a path", name, shape)
	}
	segments := path.Segments()
	if len(segments) != len(types) {
		t.Fatalf("%s: %d segments %v, want %d", name, len(segments), segments, len(types))
	}
	for i, seg := range segments {
		if seg.SegmentType() != types[i] || !samePoints(seg.Coord(), coords[i]) {
			t.Errorf("%s: segment %d is %v, want %v %v", name, i, seg, types[i], coords[i])
		}
	}
}

func TestMirror(t *testing.T) {
	if got := (Mirror{max: 100}).invAndXYSwitch(geom.NewXPoint2D(10, 30)); got != geom.NewXPoint2D(70, 10) {
		t.Errorf("invAndXYSwitch(10, 30) = %v, want (70, 10)", got)
	}
}

func TestArrow(t *testing.T) {
	r := newRecorder()
	arrow{p1: geom.NewXPoint2D(0, 0), p2: geom.NewXPoint2D(10, 0)}.drawArrow(r)
	if len(*r.shapes) != 1 {
		t.Fatalf("%d shapes, want one path", len(*r.shapes))
	}
	checkPath(t, "arrow", (*r.shapes)[0].shape,
		[]geom.USegmentType{geom.SegMoveTo, geom.SegLineTo, geom.SegLineTo, geom.SegLineTo, geom.SegLineTo},
		[][]float64{{0, 4}, {3, 0}, {0, -4}, {10, 0}, {0, 4}})
}

func TestJSONCurve(t *testing.T) {
	c, err := newJSONCurve("e,0,35 0,0 0,10 0,20 0,30", Mirror{max: 100}, 13)
	if err != nil {
		t.Fatal(err)
	}
	r := newRecorder()
	c.drawCurve(color.Black, r)
	if len(*r.shapes) != 2 {
		t.Fatalf("%d shapes, want the curve and the head", len(*r.shapes))
	}
	checkPath(t, "curve", (*r.shapes)[0].shape,
		[]geom.USegmentType{geom.SegMoveTo, geom.SegLineTo, geom.SegCubicTo},
		[][]float64{{113, 0}, {100, 0}, {90, 0, 80, 0, 70, 0}})
	r = newRecorder()
	c.drawSpot(r)
	if spot := (*r.shapes)[0]; spot.at.Dx() != 110 || spot.at.Dy() != -3 {
		t.Errorf("spot drawn at %v, want (110, -3), 13 points before the first point", spot.at)
	}
}

func TestNewJSONCurveErrors(t *testing.T) {
	for _, pos := range []string{"e,1,1 0,0 1,1 2,2 3,3;e,1,1 0,0 1,1 2,2 3,3", "0,0 1,1", "e,1 0,0 1,1 2,2 3,3"} {
		if _, err := newJSONCurve(pos, Mirror{max: 10}, 13); err == nil {
			t.Errorf("%q: no error", pos)
		}
	}
	c, err := newJSONCurve("e,0,0 0,0 1,1 2,2 3,3", Mirror{max: 10}, 13)
	if err != nil || c.ep != nil {
		t.Errorf("end point at the origin: %v, error %v, want no end point", c, err)
	}
}
