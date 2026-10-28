// SPDX-License-Identifier: MIT
// Provenance: original

package core

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/text"
)

// okDiagram is a successfully parsed diagram in tests.
type okDiagram struct{ by string }

func (okDiagram) Description() string { return "(ok)" }

// fakeFactory accepts sources containing accept, and otherwise fails after
// consuming consumed lines with the given score.
type fakeFactory struct {
	name     string
	typ      DiagramType
	accept   string
	consumed int
	score    int
	panics   bool
	seen     *[]string
}

func (f fakeFactory) DiagramType() DiagramType { return f.typ }

func (f fakeFactory) CreateSystem(_ *RenderContext, src *UmlSource) Diagram {
	if f.seen != nil {
		*f.seen = append(*f.seen, f.name)
	}
	if f.panics {
		panic("boom")
	}
	for _, l := range src.Lines() {
		if f.accept != "" && strings.Contains(l.S, f.accept) {
			return okDiagram{by: f.name}
		}
	}
	line := src.Lines()[min(f.consumed, len(src.Lines())-1)]
	return &ErrorDiagram{
		Source: src.Lines(),
		Trace:  src.Lines()[:f.consumed],
		Error:  ErrorUml{Type: SyntaxError, Error: "Syntax Error? " + f.name, Score: f.score, Line: &line},
	}
}

func srcLines(s string) []text.StringLocated {
	return BuildBlocks(s, "string", nil)[0].Lines
}

func TestCreateDiagramFirstSuccessWins(t *testing.T) {
	var seen []string
	factories := []Factory{
		fakeFactory{name: "sequence", typ: Sequence, accept: "->", seen: &seen},
		fakeFactory{name: "class", typ: Class, accept: "class", seen: &seen},
		fakeFactory{name: "mindmap", typ: MindMap, accept: "*", seen: &seen},
	}
	d := CreateDiagram(NewRenderContext(), srcLines("@startuml\nclass A\n@enduml\n"), nil, factories)
	if ok, isOK := d.(okDiagram); !isOK || ok.by != "class" {
		t.Fatalf("got %#v, want the class factory's diagram", d)
	}
	if strings.Join(seen, ",") != "sequence,class" {
		t.Errorf("factories tried: %v (mindmap must be skipped, nothing after class)", seen)
	}
}

func TestCreateDiagramBestErrorWins(t *testing.T) {
	factories := []Factory{
		fakeFactory{name: "sequence", typ: Sequence, consumed: 1},
		fakeFactory{name: "class", typ: Class, consumed: 2},
		fakeFactory{name: "activity", typ: Activity, consumed: 2},
	}
	d, ok := CreateDiagram(NewRenderContext(), srcLines("@startuml\na\nb\nc\n@enduml\n"), nil, factories).(*ErrorDiagram)
	if !ok {
		t.Fatalf("want an error diagram")
	}
	// class and activity tie on score 20: the first one is kept.
	if d.Error.Error != "Syntax Error? class" || d.Error.Position() != 2 {
		t.Errorf("error %q at %d, want class at 2", d.Error.Error, d.Error.Position())
	}

	factories = append(factories, fakeFactory{name: "timing", typ: Timing, consumed: 1, score: 15})
	d = CreateDiagram(NewRenderContext(), srcLines("@startuml\na\nb\nc\n@enduml\n"), nil, factories).(*ErrorDiagram)
	if d.Error.Error != "Syntax Error? timing" {
		t.Errorf("error %q, want timing (score 25)", d.Error.Error)
	}
}

func TestCreateDiagramPanic(t *testing.T) {
	d, ok := CreateDiagram(NewRenderContext(), srcLines("@startuml\nA\n@enduml\n"), nil, []Factory{fakeFactory{typ: Sequence, panics: true}}).(*ErrorDiagram)
	if !ok || d.Error.Type != ExecutionError || !strings.HasPrefix(d.Error.Error, "Fatal crash error") || !strings.Contains(d.Error.Error, "boom") {
		t.Fatalf("got %#v", d)
	}
}

func TestCreateDiagramBasicChecks(t *testing.T) {
	factories := []Factory{fakeFactory{typ: Sequence, accept: "{"}}
	tests := []struct {
		src  string
		want string
	}{
		{"@startuml\ndigraph G {\n  a -> b\n}\n@enduml\n", "This looks like a DOT diagram. Please use @startdot instead of @startuml. (Assumed diagram type: sequence)"},
		{"@startuml\n  strict graph \"x y\" {\n}\n@enduml\n", "This looks like a DOT diagram."},
		{"@startuml\nditaa\n@enduml\n", "This looks like a DITAA diagram."},
		{"@startuml\n salt \n@enduml\n", "This looks like a salt diagram."},
		{"@startuml\nnwdiag {\n}\n@enduml\n", "This looks like a network diagram."},
	}
	for _, tt := range tests {
		d, ok := CreateDiagram(NewRenderContext(), srcLines(tt.src), nil, factories).(*ErrorDiagram)
		if !ok || !strings.HasPrefix(d.Error.Message(), tt.want) || d.Error.Position() != 1 {
			t.Errorf("%q: got %#v", tt.src, d)
		}
	}
}

func TestCreateDiagramFamilies(t *testing.T) {
	if _, ok := CreateDiagram(NewRenderContext(), srcLines("@startfoo\nx\n@endfoo\n"), nil, nil).(*UnsupportedDiagram); !ok {
		t.Error("@startfoo must be upstream's unsupported diagram")
	}
	if d, ok := CreateDiagram(NewRenderContext(), srcLines("@startjson\n{}\n@endjson\n"), nil, nil).(*ErrorDiagram); !ok || d.Error.Error != "json diagrams are not supported yet" {
		t.Errorf("@startjson: got %#v", d)
	}
	if d, ok := CreateDiagram(NewRenderContext(), srcLines("@startexplainuml\nA -> B\n@enduml\n"), nil, nil).(*ErrorDiagram); !ok || !strings.Contains(d.Error.Error, "explain") {
		t.Errorf("@startexplain: got %#v", d)
	}
	if _, ok := CreateDiagram(NewRenderContext(), srcLines("@startuml\nA\n@enduml\n"), nil, nil).(*UnsupportedDiagram); !ok {
		t.Error("no applicable factory must give the unsupported diagram")
	}
}

func TestUmlSourceJoinsBackslashLines(t *testing.T) {
	lines := []text.StringLocated{text.NewStringLocated("@startuml", nil), text.NewStringLocated(`A -> \`, nil), text.NewStringLocated("B", nil)}
	if got := NewUmlSource(lines, nil, true).Lines(); len(got) != 2 || got[1].S != "A -> B" {
		t.Errorf("joined = %v", got)
	}
	if got := NewUmlSource(lines, nil, false).Lines(); len(got) != 3 {
		t.Errorf("not joined = %v", got)
	}
}
