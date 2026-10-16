// SPDX-License-Identifier: MIT
// Provenance: original

package plantuml

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

const seqSource = "@startuml\nAlice -> Bob : hello\n@enduml\n"

func TestRenderSVG(t *testing.T) {
	// Spec scenario "Render to SVG".
	res, err := Render(context.Background(), seqSource, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := xml.Unmarshal(res.Data, new(struct{})); err != nil {
		t.Fatalf("not well-formed XML: %v", err)
	}
	if !bytes.HasPrefix(res.Data, []byte("<svg ")) || !bytes.Contains(res.Data, []byte(">hello</text>")) {
		t.Errorf("unexpected SVG: %.200s", res.Data)
	}
	// The size upstream reports for this diagram (X-PlantUML-Diagram-Width
	// and -Height of the reference server); the SVG declares 120x130.
	if res.Width != 119 || res.Height != 129 {
		t.Errorf("size %dx%d, want 119x129", res.Width, res.Height)
	}
	if res.Description != "(2 participants)" || res.Pages != 1 || res.Diagrams != 1 {
		t.Errorf("metadata %+v", res)
	}
}

func TestRenderSelectDiagram(t *testing.T) {
	// Spec scenario "Select the second diagram".
	src := seqSource + "@startuml\nCarol -> Dave : second\nnewpage\nDave -> Carol : page 2\n@enduml\n"
	res, err := Render(context.Background(), src, &Options{Diagram: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(res.Data, []byte(">second</text>")) || bytes.Contains(res.Data, []byte(">hello</text>")) {
		t.Errorf("diagram 1 not selected: %.300s", res.Data)
	}
	if res.Diagrams != 2 || res.Pages != 2 {
		t.Errorf("diagrams %d pages %d, want 2 and 2", res.Diagrams, res.Pages)
	}
	res, err = Render(context.Background(), src, &Options{Diagram: 1, Page: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(res.Data, []byte(">page 2</text>")) {
		t.Errorf("page 1 not selected: %.300s", res.Data)
	}
}

func TestRenderSyntaxErrorDetails(t *testing.T) {
	// Spec scenario "Syntax error details".
	src := "@startuml\nAlice -> Bob\nfoo bar baz\n@enduml\n"
	res, err := Render(context.Background(), src, nil)
	var se *SyntaxError
	if !errors.As(err, &se) || !errors.Is(err, ErrSyntax) {
		t.Fatalf("error %v, want a *SyntaxError", err)
	}
	if se.Line != 3 || se.Message != "Syntax Error? (Assumed diagram type: sequence)" {
		t.Errorf("syntax error %+v", se)
	}
	if res == nil || !bytes.HasPrefix(res.Data, []byte("<svg ")) || !bytes.Contains(res.Data, []byte("foo bar baz")) {
		t.Fatalf("no error image returned")
	}
	// The size upstream reports for this error image.
	if res.Width != 591 || res.Height != 340 {
		t.Errorf("error image %dx%d, want 591x340", res.Width, res.Height)
	}
	res, err = Render(context.Background(), src, &Options{NoErrorImage: true})
	if res != nil || !errors.As(err, &se) {
		t.Errorf("NoErrorImage: result %v, error %v", res, err)
	}
}

func TestInspectBlocks(t *testing.T) {
	// Spec scenario "List blocks".
	src := "@startuml login\nAlice -> Bob : hi\n@enduml\n@startuml\nclass A\nclass B\nA --> B\n@enduml\n"
	blocks, err := Inspect(context.Background(), src, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 2 {
		t.Fatalf("%d blocks", len(blocks))
	}
	if b := blocks[0]; b.Name != "login" || b.Type != "SEQUENCE" || b.Family != "uml" || b.Err != nil {
		t.Errorf("first block %+v", b)
	}
	if b := blocks[1]; b.Name != "" || b.Type != "CLASS" || b.Err != nil {
		t.Errorf("second block %+v", b)
	}
}

func TestRenderConcurrent(t *testing.T) {
	// Spec scenario "Concurrent renders": parallel renders with different
	// styling give the same output as each render alone. Themes are not
	// bundled yet (task 7.3), so the styling comes from config lines.
	type job struct {
		src  string
		opts *Options
	}
	var jobs []job
	for i, color := range []string{"red", "blue", "green", "orange"} {
		src := fmt.Sprintf("@startuml\nparticipant P%d\nP%d -> Q : msg %d\n@enduml\n", i, i, i)
		jobs = append(jobs, job{src, &Options{Config: []string{"skinparam participantBackgroundColor " + color}}})
	}
	want := make([][]byte, len(jobs))
	for i, j := range jobs {
		res, err := Render(context.Background(), j.src, j.opts)
		if err != nil {
			t.Fatal(err)
		}
		want[i] = res.Data
	}
	var wg sync.WaitGroup
	for round := range 8 {
		for i, j := range jobs {
			wg.Add(1)
			go func() {
				defer wg.Done()
				res, err := Render(context.Background(), j.src, j.opts)
				if err != nil {
					t.Error(err)
					return
				}
				if !bytes.Equal(res.Data, want[i]) {
					t.Errorf("round %d job %d: output differs from the render alone", round, i)
				}
			}()
		}
	}
	wg.Wait()
	if !strings.Contains(string(want[0]), `fill="#F00"`) {
		t.Errorf("the config line was not applied")
	}
}
