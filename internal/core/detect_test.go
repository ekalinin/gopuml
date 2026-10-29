// SPDX-License-Identifier: MIT
// Provenance: original

package core

import (
	"strings"
	"testing"
)

func TestDetectUnsupportedUmlTypes(t *testing.T) {
	tests := []struct {
		src  string
		want string // type name in the message, "" for no detection
		line int    // 0-based position of the reported line
	}{
		// Spec scenarios: state and legacy activity syntax are not reported
		// (these diagrams are supported), archimate syntax is.
		{"@startuml\n[*] --> Idle\n@enduml\n", "", 0},
		{"@startuml\n(*) --> \"First step\"\n@enduml\n", "", 0},
		{"@startuml\narchimate #Business \"Customer\" as C\n@enduml\n", "archimate", 1},

		{"@startuml\nstate Idle\nIdle --> Busy\n@enduml\n", "", 0},
		{"@startuml\nIdle --> Busy\nBusy --> [*]\n@enduml\n", "", 0},
		{"@startuml\n\"First\" --> (*)\n@enduml\n", "", 0},
		{"@startuml\narchimate #Business \"Shop\" as shop\n@enduml\n", "archimate", 1},

		// Syntax that sequence, class, description or activity diagrams
		// parse: their factories report it.
		{"@startuml\nUser -> (Login)\n@enduml\n", "", 0},
		{"@startuml\nusecase Login\n@enduml\n", "", 0},
		{"@startuml\n[Web] --> [Db]\n@enduml\n", "", 0},
		{"@startuml\nactor User\nUser -> Bob : hi\n@enduml\n", "", 0},
		{"@startuml\nnode Server {\n}\n@enduml\n", "", 0},
		{"@startuml\nrectangle Lane {\n:step;\n}\n@enduml\n", "", 0},
		{"@startuml\nA --> B\n@enduml\n", "", 0},
		{"@startuml\n:Read input;\n@enduml\n", "", 0},
		// Activity actions whose text has two colons are not actors.
		{"@startuml\n:Step 1: read input;\n@enduml\n", "", 0},
		{"@startuml\n:Step 1: read input; <<input>>\n@enduml\n", "", 0},
		{"@startuml\n:Admin:\n@enduml\n", "", 0},
	}
	for _, tt := range tests {
		lines := BuildBlocks(tt.src, "string", nil)[0].Lines
		d := detectUnsupportedUmlType(NewUmlSource(lines, lines, true))
		if tt.want == "" {
			if d != nil {
				t.Errorf("%q: unexpected detection %q", tt.src, d.(*ErrorDiagram).Error.Error)
			}
			continue
		}
		e, ok := d.(*ErrorDiagram)
		if !ok {
			t.Errorf("%q: no detection, want %s", tt.src, tt.want)
			continue
		}
		if !strings.HasPrefix(e.Error.Error, tt.want) || !strings.HasSuffix(e.Error.Error, " diagrams are not supported yet") || e.Error.Position() != tt.line {
			t.Errorf("%q: %q at %d, want %s at %d", tt.src, e.Error.Error, e.Error.Position(), tt.want, tt.line)
		}
	}
}

func TestCreateDiagramReportsUnsupportedUmlType(t *testing.T) {
	// No in-scope factory accepts archimate syntax, so the block is reported
	// as an archimate diagram, not as a generic syntax error.
	src := "@startuml\narchimate #Business \"Customer\" as C\n@enduml\n"
	factories := []Factory{fakeFactory{name: "sequence", typ: Sequence, consumed: 1}}
	d, ok := CreateDiagram(NewRenderContext(), srcLines(src), nil, factories).(*ErrorDiagram)
	if !ok || d.Error.Message() != "archimate diagrams are not supported yet" {
		t.Fatalf("got %#v", d)
	}
	// A factory that accepts the block wins over detection.
	factories = []Factory{fakeFactory{name: "sequence", typ: Sequence, accept: "archimate"}}
	if _, ok := CreateDiagram(NewRenderContext(), srcLines(src), nil, factories).(okDiagram); !ok {
		t.Fatal("an accepting factory must win")
	}
}
