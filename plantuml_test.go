// SPDX-License-Identifier: MIT
// Provenance: original

package gopuml

import (
	"context"
	"errors"
	"slices"
	"testing"
	"testing/fstest"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	// Spec scenario "Encode and decode".
	for _, src := range []string{
		"@startuml\nAlice -> Bob : hello\n@enduml",
		"@startuml\nclass Café\nnote \"Ünïcödé ✓\" as N\n@enduml",
	} {
		got, err := Decode(Encode(src))
		if err != nil || got != src {
			t.Errorf("Decode(Encode(%q)) = %q, %v", src, got, err)
		}
	}
	if _, err := Decode("~1@@@"); !errors.Is(err, ErrDecode) {
		t.Errorf("invalid text: %v, want ErrDecode", err)
	}
}

func TestInspectNamesAndStatus(t *testing.T) {
	src := "@startuml login\nAlice -> Bob\n@enduml\n@startwbs\n* root\n@endwbs\n"
	blocks, err := Inspect(context.Background(), src, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 2 {
		t.Fatalf("%d blocks, want 2", len(blocks))
	}
	if b := blocks[0]; b.Name != "login" || b.Family != "uml" {
		t.Errorf("first block: name %q family %q", b.Name, b.Family)
	}
	if b := blocks[1]; b.Name != "" || b.Family != "wbs" || !errors.Is(b.Err, ErrUnsupportedDiagram) || b.Type != "" {
		t.Errorf("second block: name %q family %q type %q err %v", b.Name, b.Family, b.Type, b.Err)
	}
	if ue, ok := errors.AsType[*UnsupportedError](blocks[1].Err); !ok || ue.Line != 4 || ue.Message != "wbs diagrams are not supported yet" {
		t.Errorf("second block error: %#v", blocks[1].Err)
	}
	if got := blocks[0].Source(); len(got) != 3 || got[0] != "@startuml login" || got[2] != "@enduml" {
		t.Errorf("Source() = %q", got)
	}
	if got, _ := Decode(blocks[0].Encoded()); got != "@startuml\nAlice -> Bob\n@enduml" {
		t.Errorf("Encoded() decodes to %q", got)
	}
}

func TestRenderSelectionErrors(t *testing.T) {
	ctx := context.Background()
	src := "@startuml\nAlice -> Bob\n@enduml\n"
	if _, err := Render(ctx, src, &Options{Diagram: 1}); !errors.Is(err, ErrNoDiagram) {
		t.Errorf("diagram out of range: %v", err)
	}
	if _, err := Render(ctx, src, &Options{Page: 1}); !errors.Is(err, ErrNoDiagram) {
		t.Errorf("page out of range: %v", err)
	}
	if _, err := Render(ctx, "no diagram here", nil); !errors.Is(err, ErrNoDiagram) {
		t.Errorf("no block: %v", err)
	}
	if _, err := Render(ctx, src, &Options{SecurityProfile: "nope"}); err == nil {
		t.Error("unknown security profile accepted")
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := Render(cancelled, src, nil); !errors.Is(err, context.Canceled) {
		t.Errorf("cancelled context: %v", err)
	}
}

func TestSyntaxErrorIs(t *testing.T) {
	var err error = &SyntaxError{Line: 3, Message: "Syntax Error?"}
	var se *SyntaxError
	if !errors.Is(err, ErrSyntax) || !errors.As(err, &se) || se.Line != 3 {
		t.Errorf("errors.Is/As failed for %v", err)
	}
	if err.Error() != "plantuml: line 3: Syntax Error?" {
		t.Errorf("Error() = %q", err.Error())
	}
}

func TestRenderIncludesFromFS(t *testing.T) {
	// Spec scenario "Includes from a supplied file system". The include is
	// resolved from the in-memory files; none of them exists on disk.
	fsys := fstest.MapFS{"style.iuml": {Data: []byte("Bob -> Carol : from fs\n")}}
	src := "@startuml\nAlice -> Bob\n!include style.iuml\n@enduml\n"
	for _, profile := range []string{"", "SANDBOX"} {
		opts := &Options{FS: fsys, SecurityProfile: profile}
		blocks, err := Inspect(context.Background(), src, opts)
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"@startuml", "Alice -> Bob", "Bob -> Carol : from fs", "@enduml"}
		if got := blocks[0].Source(); !slices.Equal(got, want) {
			t.Errorf("profile %q: preprocessed %q, want %q", profile, got, want)
		}
		// Drawing is checked once sequence diagrams render (task 8.8).
		if _, err := Render(context.Background(), src, opts); err != nil && !errors.Is(err, ErrUnsupportedDiagram) {
			t.Errorf("profile %q: %v", profile, err)
		}
	}
	// Without supplied files, SANDBOX reads nothing from disk.
	blocks, err := Inspect(context.Background(), src, &Options{SecurityProfile: "SANDBOX"})
	if err != nil {
		t.Fatal(err)
	}
	if !errors.Is(blocks[0].Err, ErrSyntax) {
		t.Errorf("sandbox without files: %v", blocks[0].Err)
	}
}
