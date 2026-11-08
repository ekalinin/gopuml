// SPDX-License-Identifier: MIT
// Provenance: original

package core

import (
	"regexp"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/command/regex"
)

// withUnsupportedType sets a test entry, the type foo of the lines starting
// with "foo ", in unsupportedTypes for the duration of a test, since no
// @startuml type is reported as not supported yet in this release.
func withUnsupportedType(t *testing.T) {
	t.Helper()
	saved := unsupportedTypes
	unsupportedTypes = []unsupportedType{{"foo", []*regexp.Regexp{regex.MustCompile(`^foo[%s]`)}}}
	t.Cleanup(func() { unsupportedTypes = saved })
}

func TestDetectUnsupportedUmlTypes(t *testing.T) {
	withUnsupportedType(t)
	tests := []struct {
		src  string
		want string // type name in the message, "" for no detection
		line int    // 0-based position of the reported line
	}{
		// Spec scenarios: state, legacy activity and archimate syntax are not
		// reported (these diagrams are supported), the syntax of the test
		// entry is.
		{"@startuml\n[*] --> Idle\n@enduml\n", "", 0},
		{"@startuml\n(*) --> \"First step\"\n@enduml\n", "", 0},
		{"@startuml\narchimate #Business \"Customer\" as C\n@enduml\n", "", 0},
		{"@startuml\nfoo bar\n@enduml\n", "foo", 1},

		{"@startuml\nstate Idle\nIdle --> Busy\n@enduml\n", "", 0},
		{"@startuml\nIdle --> Busy\nBusy --> [*]\n@enduml\n", "", 0},
		{"@startuml\n\"First\" --> (*)\n@enduml\n", "", 0},
		{"@startuml\nAlice -> Bob\nfoo bar\n@enduml\n", "foo", 2},

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
	withUnsupportedType(t)
	// No in-scope factory accepts the syntax of the test entry, so the block
	// is reported as a foo diagram, not as a generic syntax error.
	src := "@startuml\nfoo bar\n@enduml\n"
	factories := []Factory{fakeFactory{name: "sequence", typ: Sequence, consumed: 1}}
	d, ok := CreateDiagram(NewRenderContext(), srcLines(src), nil, factories).(*ErrorDiagram)
	if !ok || d.Error.Message() != "foo diagrams are not supported yet" {
		t.Fatalf("got %#v", d)
	}
	// A factory that accepts the block wins over detection.
	factories = []Factory{fakeFactory{name: "sequence", typ: Sequence, accept: "foo"}}
	if _, ok := CreateDiagram(NewRenderContext(), srcLines(src), nil, factories).(okDiagram); !ok {
		t.Fatal("an accepting factory must win")
	}
}
