// SPDX-License-Identifier: MIT
// Provenance: original

package activitydiagram

import (
	"testing"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

func noteCommands() []cmd {
	return []cmd{commandLinkLongActivity(), commandNoteActivitySingle(), commandNoteActivityMulti(),
		commandLinkActivity()}
}

// noteLink returns the note of body and its link.
func noteLink(t *testing.T, body ...string) (*abel.Entity, *abel.Link) {
	t.Helper()
	d := mustParse(t, noteCommands, body...)
	for _, l := range d.Links() {
		for _, e := range []*abel.Entity{l.Entity1(), l.Entity2()} {
			if e.LeafType() == abel.LeafTypeNote {
				return e, l
			}
		}
	}
	t.Fatalf("%q: no note", body)
	return nil, nil
}

// TestNotePositions checks the four positions of single-line and
// multi-line notes: the link goes from the activity to the note on the
// right and at the bottom, from the note to the activity otherwise, and is
// longer at the top and the bottom.
func TestNotePositions(t *testing.T) {
	for _, c := range []struct {
		position     string
		noteIsSecond bool
		length       int
	}{{"right", true, 1}, {"left", false, 1}, {"bottom", true, 2}, {"top", false, 2}} {
		for _, body := range [][]string{
			{`(*) --> "Pay"`, "note " + c.position + ": by card"},
			{`(*) --> "Pay"`, "note " + c.position, "by card", "end note"},
		} {
			note, l := noteLink(t, body...)
			activity := l.Entity1()
			if !c.noteIsSecond {
				activity = l.Entity2()
			}
			if (l.Entity2() == note) != c.noteIsSecond || activity.Name() != "Pay" || l.Length() != c.length {
				t.Errorf("%q: link %s length %d", body, ends(l), l.Length())
			}
			if displayOf(note) != "by card" {
				t.Errorf("%q: note %q", body, displayOf(note))
			}
		}
	}
}

// TestNoteColorAndURL checks the color of a note and the URL on the first
// line of a multi-line note.
func TestNoteColorAndURL(t *testing.T) {
	note, _ := noteLink(t, `(*) --> "Pay"`, "note left #pink", "[[http://example.com]]", "by card", "end note")
	if note.Colors().Color(color.TypeBack) == nil {
		t.Error("the note has no color")
	}
	if note.URL99() == nil || displayOf(note) != "by card" {
		t.Errorf("url %v, note %q", note.URL99(), displayOf(note))
	}
}

// TestNoteWithoutActivity checks that a note before any activity is
// attached to the start.
func TestNoteWithoutActivity(t *testing.T) {
	_, l := noteLink(t, "note right: first", `(*) --> "Pay"`)
	if l.Entity1().LeafType() != abel.LeafTypeCircleStart {
		t.Errorf("link %s", ends(l))
	}
}
