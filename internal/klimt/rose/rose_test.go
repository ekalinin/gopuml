// SPDX-License-Identifier: MIT
// Provenance: original

package rose

import (
	"math"
	"strings"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
)

// The behavior that the golden file does not reach: Java nullability,
// panics, enum helpers and the Java semantics of the small helpers.

var (
	_ klimt.Component      = (*ComponentRoseParticipant)(nil)
	_ klimt.Component      = (*ComponentRoseActor)(nil)
	_ klimt.Component      = (*ComponentRoseBoundary)(nil)
	_ klimt.Component      = (*ComponentRoseControl)(nil)
	_ klimt.Component      = (*ComponentRoseEntity)(nil)
	_ klimt.Component      = (*ComponentRoseDatabase)(nil)
	_ klimt.Component      = (*ComponentRoseQueue)(nil)
	_ klimt.Component      = (*ComponentRoseNote)(nil)
	_ klimt.Component      = (*ComponentRoseNoteBox)(nil)
	_ klimt.Component      = (*ComponentRoseNoteHexagonal)(nil)
	_ klimt.Component      = (*ComponentRoseGroupingHeader)(nil)
	_ klimt.Component      = (*ComponentRoseGroupingElse)(nil)
	_ klimt.Component      = (*ComponentRoseLine)(nil)
	_ klimt.Component      = (*ComponentRoseActiveLine)(nil)
	_ klimt.Component      = (*ComponentRoseDelayLine)(nil)
	_ klimt.Component      = (*ComponentRoseDelayText)(nil)
	_ klimt.Component      = (*ComponentRoseDestroy)(nil)
	_ klimt.Component      = (*ComponentRoseNewpage)(nil)
	_ klimt.Component      = (*ComponentRoseDivider)(nil)
	_ klimt.Component      = (*ComponentRoseReference)(nil)
	_ klimt.Component      = (*ComponentRoseEnglober)(nil)
	_ klimt.ArrowComponent = (*ComponentRoseArrow)(nil)
	_ klimt.ArrowComponent = (*ComponentRoseSelfArrow)(nil)
)

// panicClass returns the Java class of the exception f panics with, "" when
// it does not panic.
func panicClass(t *testing.T, f func()) (class string) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(*color.RuntimeError)
			if !ok {
				t.Fatalf("panic %v is not a RuntimeError", r)
			}
			class = e.Class
		}
	}()
	f()
	return ""
}

func TestNotePosition(t *testing.T) {
	var names []string
	for _, p := range NotePositionValues() {
		names = append(names, p.String())
		if NotePositionValueOf(p.String()) != p {
			t.Error(p)
		}
	}
	if got := strings.Join(names, ","); got != "LEFT,RIGHT,OVER,OVER_SEVERAL,BOTTOM,TOP" {
		t.Error(got)
	}
	if got := NotePosition(9).String(); got != "NotePosition(9)" {
		t.Error(got)
	}
	if got := NotePositionDefaultLeft("ignored", false); got != NotePositionLeft {
		t.Error(got)
	}
	if got := NotePositionDefaultLeft("over_several", true); got != NotePositionOverSeveral {
		t.Error(got)
	}
	if c := panicClass(t, func() { NotePositionValueOf("left") }); c != illegalArgument {
		t.Error(c)
	}
}

func TestJavaHelpers(t *testing.T) {
	for _, c := range []struct {
		in   float64
		want int
	}{{math.NaN(), 0}, {3e10, 2147483647}, {-3e10, -2147483648}, {-2.7, -2}, {2.7, 2}} {
		if got := javaD2I(c.in); got != c.want {
			t.Errorf("javaD2I(%v) = %d", c.in, got)
		}
	}
	for _, c := range []struct {
		a, b string
		want bool
	}{{"first", "first", true}, {"first", "FIRST", true}, {"first", "Fırst", true}, {"first", "firs", false},
		{"direction", "direct1on", false}, {"first", "ﬁrst", false}} {
		if got := equalsIgnoreCase(c.a, c.b); got != c.want {
			t.Errorf("equalsIgnoreCase(%q, %q) = %v", c.a, c.b, got)
		}
	}
}

func TestComponentEdgeCases(t *testing.T) {
	sp := newSkinParam()
	group := seq(sp, klimt.SNameGroup)
	b := fakeBounder{}

	space := NewComponentRoseGroupingSpace(7)
	if c := panicClass(t, func() { space.UsedStyles() }); c != unsupportedOperation {
		t.Error(c)
	}
	if c := panicClass(t, func() { space.StyleSignature() }); c != unsupportedOperation {
		t.Error(c)
	}
	var log []string
	space.DrawU(newRec(&log, b), klimt.AreaCreate(10, 10), klimt.NewSimpleContext2D(false))
	if len(log) != 0 || space.PreferredDimension(b).String() != "[0.0,7.0]" {
		t.Error(log, space.PreferredDimension(b))
	}

	// A null comment gives the empty label (upstream's null check).
	elseNull := NewComponentRoseGroupingElse(false, group, nil, sp)
	if w := elseNull.PreferredWidth(b); w != 10 {
		t.Error(w)
	}
	header := NewComponentRoseGroupingHeader(false, group, seq(sp, klimt.SNameGroupHeader),
		klimt.DisplayCreateSeq([]klimt.CharSequence{nil, nil}), sp)
	if w := header.PreferredWidth(b); w != 45 {
		t.Error(w)
	}

	arrow := NewRose().CreateComponentArrow([]*klimt.Style{seq(sp, klimt.SNameArrow)},
		klimt.ArrowConfigurationWithDirectionNormal(), sp, klimt.DisplayCreate("x")).(*ComponentRoseArrow)
	if arrow.ArrowDeltaX() != 10 || arrow.ArrowDeltaY() != 4 {
		t.Error(arrow.ArrowDeltaX(), arrow.ArrowDeltaY())
	}

	// Null styles throw NullPointerException upstream.
	if c := panicClass(t, func() { NewRose().CreateComponentNewPage(nil, sp) }); c != nullPointer {
		t.Error(c)
	}
	if c := panicClass(t, func() {
		NewRose().CreateComponentNote(nil, klimt.ComponentTypeNote, sp, klimt.DisplayCreate("x"), nil)
	}); c != nullPointer {
		t.Error(c)
	}
}
