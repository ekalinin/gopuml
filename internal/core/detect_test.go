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
		// Spec scenarios: state and legacy activity syntax.
		{"@startuml\n[*] --> Idle\n@enduml\n", "state", 1},
		{"@startuml\n(*) --> \"First step\"\n@enduml\n", "legacy activity", 1},

		{"@startuml\nstate Idle\nIdle --> Busy\n@enduml\n", "state", 1},
		{"@startuml\nIdle --> Busy\nBusy --> [*]\n@enduml\n", "state", 2},
		{"@startuml\n\"First\" --> (*)\n@enduml\n", "legacy activity", 1},
		{"@startuml\nUser -> (Login)\n@enduml\n", "description", 1},
		{"@startuml\n:Admin: --> (Manage)\n@enduml\n", "description", 1},
		{"@startuml\nusecase Login\n@enduml\n", "description", 1},
		{"@startuml\nrectangle \"PlantUML at its best\" as leaf42\n@enduml\n", "description", 1},
		{"@startuml\n[Web] --> [Db]\n@enduml\n", "description", 1},
		{"@startuml\narchimate #Business \"Shop\" as shop\n@enduml\n", "description", 1},
		{"@startuml\nrobust \"Web\" as WB\n@0\nWB is Idle\n@enduml\n", "timing", 1},
		{"@startuml\nclock clk with period 1\n@enduml\n", "timing", 1},

		// Syntax that sequence, class or activity diagrams also use.
		{"@startuml\nactor User\nUser -> Bob : hi\n@enduml\n", "", 0},
		{"@startuml\nnode Server {\n}\n@enduml\n", "", 0},
		{"@startuml\nrectangle Lane {\n:step;\n}\n@enduml\n", "", 0},
		{"@startuml\nA --> B\n@enduml\n", "", 0},
		{"@startuml\n:Read input;\n@enduml\n", "", 0},
		// Activity actions whose text has two colons are not actors.
		{"@startuml\n:Step 1: read input;\n@enduml\n", "", 0},
		{"@startuml\n:Step 1: read input; <<input>>\n@enduml\n", "", 0},
		{"@startuml\n:Admin:\n@enduml\n", "description", 1},
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
	// No in-scope factory accepts state syntax, so the block is reported as
	// a state diagram, not as a generic syntax error.
	factories := []Factory{fakeFactory{name: "sequence", typ: Sequence, consumed: 1}}
	d, ok := CreateDiagram(NewRenderContext(), srcLines("@startuml\n[*] --> Idle\n@enduml\n"), nil, factories).(*ErrorDiagram)
	if !ok || d.Error.Message() != "state diagrams are not supported yet" {
		t.Fatalf("got %#v", d)
	}
	// A factory that accepts the block wins over detection.
	factories = []Factory{fakeFactory{name: "sequence", typ: Sequence, accept: "-->"}}
	if _, ok := CreateDiagram(NewRenderContext(), srcLines("@startuml\n[*] --> Idle\n@enduml\n"), nil, factories).(okDiagram); !ok {
		t.Fatal("an accepting factory must win")
	}
}

func TestDescriptionBeforeActivity(t *testing.T) {
	for _, tt := range []struct {
		src  string
		want bool
		line int
	}{
		{"@startuml\nrectangle \"PlantUML at its best\" as leaf42\n@enduml\n", true, 1},
		{"@startuml\ntitle T\ncard A {\n  rectangle B\n}\n@enduml\n", true, 2},
		{"@startuml\nrectangle Lane {\n:step;\n}\n@enduml\n", false, 0},
		{"@startuml\n:step;\n@enduml\n", false, 0},
		{"@startuml\n@enduml\n", false, 0},
	} {
		lines := BuildBlocks(tt.src, "string", nil)[0].Lines
		d := detectDescriptionOnly(NewUmlSource(lines, lines, true))
		if (d != nil) != tt.want {
			t.Errorf("%q: detected %v, want %v", tt.src, d != nil, tt.want)
			continue
		}
		if ed, ok := d.(*ErrorDiagram); ok && (ed.Position() != tt.line || !ed.NotSupportedYet) {
			t.Errorf("%q: line %d, want %d", tt.src, ed.Position(), tt.line)
		}
	}
}
