// SPDX-License-Identifier: MIT
// Provenance: original

package mindmap

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/uml"
)

// panicClass returns the Java exception that f throws, "" for none.
func panicClass(f func()) (class string) {
	defer func() {
		if p := recover(); p != nil {
			var re *color.RuntimeError
			if err, ok := p.(error); ok && errors.As(err, &re) {
				class = re.Class
			} else {
				class = fmt.Sprint(p)
			}
		}
	}()
	f()
	return ""
}

func stripesString(f *stripeFrontier) string {
	s := ""
	for _, st := range f.stripes {
		s += fmt.Sprintf("[%v %v %v]", st.x1, st.x2, st.value)
	}
	return s
}

func TestStripeFrontier(t *testing.T) {
	f := newStripeFrontier()
	if !f.isEmpty() {
		t.Fatal("a new frontier is not empty")
	}
	f.addSegment(0, 10, 5)
	f.addSegment(20, 30, 5)
	if f.isEmpty() {
		t.Fatal("the frontier is empty after two segments")
	}
	// A segment over three stripes is split at their starts.
	f.addSegment(5, 25, 8)
	m := math.MaxFloat64
	want := fmt.Sprintf("[%v 0 %v][0 5 5][5 10 8][10 20 8][20 25 8][25 30 5][30 %v %v]", -m, -m, m, -m)
	if got := stripesString(f); got != want {
		t.Errorf("stripes %s, want %s", got, want)
	}
	// A segment below the frontier changes nothing.
	f.addSegment(0, 10, 3)
	if got := stripesString(f); got != want {
		t.Errorf("after a lower segment: stripes %s, want %s", got, want)
	}
	for _, tc := range []struct{ x1, x2, want float64 }{
		{3, 22, 8}, {0, 5, 5}, {26, 28, 5}, {31, 40, -m}, {-5, 0, -m},
	} {
		if got := f.getContact(tc.x1, tc.x2); got != tc.want {
			t.Errorf("getContact(%v, %v) = %v, want %v", tc.x1, tc.x2, got, tc.want)
		}
	}
	if got := panicClass(func() { f.addSegment(5, 5, 1) }); got != uml.IllegalArgumentException {
		t.Errorf("empty segment: %q, want %s", got, uml.IllegalArgumentException)
	}
	if got := panicClass(func() { newStripe(2, 1, 0) }); got != uml.IllegalArgumentException {
		t.Errorf("reversed stripe: %q, want %s", got, uml.IllegalArgumentException)
	}
}

func TestTetris(t *testing.T) {
	// The second tee goes below the first: its phalanx touches the bottom
	// of the first phalanx.
	tr := newTetris()
	tr.add(newSymetricalTee(10, 20, 0, 0))
	tr.add(newSymetricalTee(10, 20, 0, 0))
	if y0, y1 := tr.elements[0].y, tr.elements[1].y; y0 != 0 || y1 != 10 {
		t.Errorf("placed at %v and %v, want 0 and 10", y0, y1)
	}
	// balance centers the tees on 0.
	tr.balance()
	if y0, y1 := tr.elements[0].y, tr.elements[1].y; y0 != -5 || y1 != 5 {
		t.Errorf("balanced at %v and %v, want -5 and 5", y0, y1)
	}
	if h, w := tr.height(), tr.width(); h != 20 || w != 20 {
		t.Errorf("height %v width %v, want 20 and 20", h, w)
	}
	if got := panicClass(tr.balance); got != uml.IllegalStateException {
		t.Errorf("second balance: %q, want %s", got, uml.IllegalStateException)
	}

	// The lower of the two placements is kept: the nail of the second tee
	// meets the nail of the first one (at 15) lower than its phalanx meets
	// the first phalanx (at 5).
	tr = newTetris()
	tr.add(newSymetricalTee(10, 20, 30, 20))
	tr.add(newSymetricalTee(10, 10, 10, 30))
	if y := tr.elements[1].y; y != 20 {
		t.Errorf("second tee at %v, want 20", y)
	}
	tr.balance()
	if h, w := tr.height(), tr.width(); h != 40 || w != 40 {
		t.Errorf("height %v width %v, want 40 and 40", h, w)
	}

	// An empty tetris has no height and no width.
	tr = newTetris()
	tr.balance()
	if h, w := tr.height(), tr.width(); h != 0 || w != 0 {
		t.Errorf("empty: height %v width %v", h, w)
	}
}
