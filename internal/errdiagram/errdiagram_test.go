// SPDX-License-Identifier: MIT
// Provenance: original

package errdiagram_test

import (
	"bytes"
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/conformance"
	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/engine"
	"github.com/ekalinin/plantuml-in-go/internal/errdiagram"
	"github.com/ekalinin/plantuml-in-go/internal/export"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/text"
)

type otherDiagram struct{}

func (otherDiagram) Description() string { return "(other)" }

func TestHandles(t *testing.T) {
	for _, d := range []core.Diagram{&core.ErrorDiagram{}, &core.UnsupportedDiagram{}, &errdiagram.Welcome{}} {
		if !errdiagram.Handles(d) {
			t.Errorf("Handles(%T) = false", d)
		}
	}
	if errdiagram.Handles(otherDiagram{}) || errdiagram.Handles(nil) {
		t.Error("Handles accepts another diagram")
	}
	var buf bytes.Buffer
	if _, err := errdiagram.Export(&buf, otherDiagram{}, errdiagram.Source{},
		export.NewFileFormatOption(font.FormatSVG)); !errors.Is(err, errdiagram.ErrUnhandled) {
		t.Errorf("Export of another diagram: %v", err)
	}
	if buf.Len() != 0 {
		t.Error("Export of another diagram wrote output")
	}
}

func TestStatus(t *testing.T) {
	// FileImageData.ERROR and CRASH.
	if errdiagram.StatusError != 400 || errdiagram.StatusCrash != 503 {
		t.Error("status codes differ from upstream's")
	}
}

func lines(desc string, ss ...string) []text.StringLocated {
	loc := text.NewLineLocation(desc, nil)
	var out []text.StringLocated
	for _, s := range ss {
		loc = loc.OneLineRead()
		out = append(out, text.NewStringLocated(s, loc))
	}
	return out
}

func TestWelcomeFactory(t *testing.T) {
	f := errdiagram.WelcomeFactory{}
	if f.DiagramType() != core.Sequence {
		t.Errorf("DiagramType = %v", f.DiagramType())
	}
	two := lines("string", "@startuml", "@enduml")
	d := f.CreateSystem(core.NewRenderContext(), core.NewUmlSource(two, two, true))
	if d == nil || d.Description() != "(Empty)" {
		t.Fatalf("CreateSystem of an empty block = %v", d)
	}
	three := lines("string", "@startuml", "' comment", "@enduml")
	if d := f.CreateSystem(core.NewRenderContext(), core.NewUmlSource(three, three, true)); d != nil {
		t.Errorf("CreateSystem of three lines = %v, want nil", d)
	}
	// Goes first, as upstream's PSystemWelcomeFactory: the sequence factory
	// would report an empty description.
	factories := append([]core.Factory{f}, engine.Factories...)
	if d := core.CreateDiagram(core.NewRenderContext(), two, two, factories); d.Description() != "(Empty)" {
		t.Errorf("CreateDiagram = %T %s", d, d.Description())
	}
}

// The "not supported yet" errors of this implementation are drawn as
// upstream's error image.
func TestNotSupportedYet(t *testing.T) {
	rc := core.NewRenderContext()
	blocks, err := engine.Parse(context.Background(), rc, engine.Source{Text: "@startmindmap\n* a\n** b\n@endmindmap\n"})
	if err != nil {
		t.Fatal(err)
	}
	d, ok := blocks[0].Diagram.(*core.ErrorDiagram)
	if !ok || !d.NotSupportedYet {
		t.Fatalf("diagram %T, want a not supported yet error", blocks[0].Diagram)
	}
	var buf bytes.Buffer
	data, err := errdiagram.Export(&buf, d, errdiagram.Source{Data: blocks[0].Data, Raw: blocks[0].Raw},
		export.NewFileFormatOption(font.FormatSVG))
	if err != nil {
		t.Fatal(err)
	}
	if data.Status() != errdiagram.StatusError {
		t.Errorf("status %d", data.Status())
	}
	texts, err := conformance.SVGTexts(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(texts, "|")
	// Four lines: the welcome help goes on top.
	want := "Welcome to PlantUML!|"
	if !strings.HasPrefix(got, want) {
		t.Errorf("texts %q do not start with %q", got, want)
	}
	// Then the location of the last line of the trace, the trace and the
	// message.
	want = "|[From string (line " + strconv.Itoa(d.Position()+1) + ") ]"
	for _, l := range d.Trace {
		want += "|" + l.S
	}
	want += "|mindmap diagrams are not supported yet"
	if !strings.HasSuffix(got, want) {
		t.Errorf("texts %q do not end with %q", got, want)
	}
}

func TestNotImplementedFormat(t *testing.T) {
	d := &core.UnsupportedDiagram{Directive: "@startfoo"}
	src := errdiagram.Source{Data: lines("string", "@startfoo", "@endfoo")}
	var buf bytes.Buffer
	if _, err := errdiagram.Export(&buf, d, src, export.NewFileFormatOption(font.FormatEPS)); !errors.Is(err,
		export.ErrNotImplemented) {
		t.Errorf("EPS: %v", err)
	}
	if _, err := errdiagram.ExportNoValidStart(&buf, export.NewFileFormatOption(font.FormatEPS)); !errors.Is(err,
		export.ErrNotImplemented) {
		t.Errorf("EPS: %v", err)
	}
}

// Images drawn in parallel share no state (design D3): run with -race.
func TestParallel(t *testing.T) {
	src := lines("string", "@startuml", "Alice -> Bob : hi", "Alice -> -> Bob", "@enduml")
	last := src[2]
	seq := core.Sequence
	diagrams := []core.Diagram{
		&core.ErrorDiagram{Source: src, Trace: src[:3],
			Error: core.ErrorUml{Type: core.SyntaxError, Error: "Syntax Error?", Line: &last, AssumedType: &seq}},
		&core.UnsupportedDiagram{Directive: "@startuml"},
		errdiagram.WelcomeFactory{}.CreateSystem(core.NewRenderContext(), core.NewUmlSource(src[:2], src[:2], true)),
	}
	want := make([][]byte, len(diagrams))
	for i, d := range diagrams {
		var buf bytes.Buffer
		if _, err := errdiagram.Export(&buf, d, errdiagram.Source{Data: src}, export.NewFileFormatOption(font.FormatSVG)); err != nil {
			t.Fatal(err)
		}
		want[i] = buf.Bytes()
	}
	var wg sync.WaitGroup
	for range 8 {
		for i, d := range diagrams {
			wg.Go(func() {
				var buf bytes.Buffer
				if _, err := errdiagram.Export(&buf, d, errdiagram.Source{Data: src},
					export.NewFileFormatOption(font.FormatSVG)); err != nil {
					t.Error(err)
				}
				if !bytes.Equal(buf.Bytes(), want[i]) {
					t.Errorf("diagram %d differs in parallel", i)
				}
			})
		}
	}
	wg.Wait()
}
