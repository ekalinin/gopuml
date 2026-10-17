// SPDX-License-Identifier: MIT
// Provenance: original

package core

import (
	"slices"
	"testing"

	"github.com/ekalinin/gopuml/internal/text"
)

func TestFindStartTypes(t *testing.T) {
	tests := []struct {
		line string
		want []DiagramType
	}{
		{"@startuml", umlTypes},
		{"  \\startuml x", umlTypes},
		{"@STARTUML", umlTypes}, // keyword matching folds ASCII case
		{"@startmindmap", []DiagramType{MindMap}},
		{"@StartMindMap", []DiagramType{MindMap}},
		{"@startproject", []DiagramType{Gantt}},
		{"@startgantt", []DiagramType{Gantt}},
		{"@startpacketdiag", []DiagramType{Packet}},
		{"@startchen", []DiagramType{ChenEER}},
		{"@startditaa", []DiagramType{Unknown}}, // not in the MIT variant
		{"@startfoo", []DiagramType{Unknown}},
		{"@start", nil},
		{"@sta", nil},
		{"startuml", nil},
		{"Alice -> Bob", nil},
		{"", nil},
	}
	for _, tt := range tests {
		if got := FindStartTypes(tt.line); !slices.Equal(got, tt.want) {
			t.Errorf("FindStartTypes(%q) = %v, want %v", tt.line, got, tt.want)
		}
	}
}

func TestDiagramTypeNames(t *testing.T) {
	if Sequence.String() != "SEQUENCE" || ChenEER.String() != "CHEN_EER" || Unknown.String() != "UNKNOWN" {
		t.Error("enum names do not match upstream")
	}
	if Description.HumanReadableName() != "component" || MindMap.HumanReadableName() != "mindmap" {
		t.Error("human readable names do not match upstream")
	}
	if !Sprites.IsLegacyUML() || MindMap.IsLegacyUML() {
		t.Error("IsLegacyUML")
	}
	seq := Sequence
	e := ErrorUml{Error: "Syntax Error?", AssumedType: &seq}
	if e.Message() != "Syntax Error? (Assumed diagram type: sequence)" {
		t.Errorf("message = %q", e.Message())
	}
}

func block(first string) *Block {
	return &Block{Lines: []text.StringLocated{text.NewStringLocated(first, text.NewLineLocation("s", nil).OneLineRead())}}
}

func TestCheckFamily(t *testing.T) {
	// Supported family.
	if d := CheckFamily(block("@startuml")); d != nil {
		t.Fatalf("@startuml: got %#v, want nil", d)
	}

	// Recognized but unsupported family: the error names the family.
	d, ok := CheckFamily(block("@startmindmap")).(*ErrorDiagram)
	if !ok || d.Error.Message() != "mindmap diagrams are not supported yet" {
		t.Fatalf("@startmindmap: got %#v", d)
	}
	if d.Description() != "(Error)" || d.Error.Position() != 0 {
		t.Errorf("description %q position %d", d.Description(), d.Error.Position())
	}

	// Unknown directive: upstream's unsupported diagram.
	u, ok := CheckFamily(block("@startfoo")).(*UnsupportedDiagram)
	if !ok || u.Directive != "@startfoo" || u.Description() != "(Unsupported)" {
		t.Fatalf("@startfoo: got %#v", u)
	}
	if lines := u.Lines("v", "MIT"); lines[2] != `Sorry, but the following directive ""@startfoo"" is not recognized.` {
		t.Errorf("unsupported text: %q", lines[2])
	}
}
