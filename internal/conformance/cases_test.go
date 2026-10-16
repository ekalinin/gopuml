// SPDX-License-Identifier: MIT
// Provenance: original

package conformance

import (
	"strings"
	"testing"
	"testing/fstest"
)

const validCase = `---
id: sequence-diagram/self-message
family: uml
type: sequence
formats: [svg, txt]
scenarios:
  - "sequence-diagram: Self message"
---
@startuml
Alice -> Alice : think
@enduml
`

func TestParseCase(t *testing.T) {
	c, err := ParseCase("sequence-diagram/self-message.puml", []byte(validCase))
	if err != nil {
		t.Fatal(err)
	}
	if c.ID != "sequence-diagram/self-message" || c.Family != "uml" || c.Type != "sequence" {
		t.Errorf("unexpected header: %+v", c)
	}
	if strings.Join(c.Formats, ",") != "svg,txt" {
		t.Errorf("formats = %v", c.Formats)
	}
	if len(c.Scenarios) != 1 || c.Scenarios[0] != "sequence-diagram: Self message" {
		t.Errorf("scenarios = %v", c.Scenarios)
	}
	if c.Source != "@startuml\nAlice -> Alice : think\n@enduml\n" {
		t.Errorf("source = %q", c.Source)
	}
	if c.Capability() != "sequence-diagram" {
		t.Errorf("capability = %q", c.Capability())
	}
}

func TestParseCaseCRLF(t *testing.T) {
	data := strings.ReplaceAll(validCase, "\n", "\r\n")
	c, err := ParseCase("sequence-diagram/self-message.puml", []byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(c.Source, "\r") {
		t.Errorf("source keeps CR characters: %q", c.Source)
	}
}

func TestParseCaseErrors(t *testing.T) {
	replace := func(old, new string) string { return strings.Replace(validCase, old, new, 1) }
	tests := []struct {
		name    string
		path    string
		data    string
		wantErr string
	}{
		{"no header", "sequence-diagram/self-message.puml", "@startuml\nA -> B\n@enduml\n", "missing YAML header"},
		{"unterminated header", "sequence-diagram/self-message.puml", "---\nid: x\n@startuml\n", "unterminated YAML header"},
		{"empty header", "sequence-diagram/self-message.puml", "---\n---\n@startuml\nA -> B\n@enduml\n", "missing id"},
		{"id mismatch", "sequence-diagram/other.puml", validCase, "does not match the file location"},
		{"missing family", "sequence-diagram/self-message.puml", replace("family: uml\n", ""), "missing family"},
		{"missing type", "sequence-diagram/self-message.puml", replace("type: sequence\n", ""), "missing type"},
		{"missing formats", "sequence-diagram/self-message.puml", replace("formats: [svg, txt]\n", ""), "missing formats"},
		{"unknown format", "sequence-diagram/self-message.puml", replace("[svg, txt]", "[svg, pdf]"), `unknown format "pdf"`},
		{"unknown field", "sequence-diagram/self-message.puml", replace("type: sequence\n", "type: sequence\ncolour: red\n"), "invalid YAML header"},
		{"bad scenario ref", "sequence-diagram/self-message.puml", replace(`"sequence-diagram: Self message"`, `"Self message"`), "is not in the form"},
		{"empty source", "sequence-diagram/self-message.puml", validCase[:strings.Index(validCase, "@startuml")], "empty diagram source"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCase(tt.path, []byte(tt.data))
			if err == nil {
				t.Fatalf("expected error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error %q does not contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoadCases(t *testing.T) {
	second := strings.NewReplacer(
		"sequence-diagram/self-message", "cli/missing-file",
		"type: sequence", "type: none",
	).Replace(validCase)
	fsys := fstest.MapFS{
		"cases/sequence-diagram/self-message.puml": {Data: []byte(validCase)},
		"cases/cli/missing-file.puml":              {Data: []byte(second)},
		"cases/README.md":                          {Data: []byte("ignored")},
	}
	cases, err := LoadCases(fsys, "cases")
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) != 2 {
		t.Fatalf("loaded %d cases, want 2", len(cases))
	}
	if cases[0].ID != "cli/missing-file" || cases[1].ID != "sequence-diagram/self-message" {
		t.Errorf("cases not sorted by id: %s, %s", cases[0].ID, cases[1].ID)
	}
}

func TestLoadCasesReportsBadFile(t *testing.T) {
	fsys := fstest.MapFS{
		"cases/sequence-diagram/broken.puml": {Data: []byte("@startuml\n@enduml\n")},
	}
	if _, err := LoadCases(fsys, "cases"); err == nil || !strings.Contains(err.Error(), "broken.puml") {
		t.Fatalf("expected an error naming broken.puml, got %v", err)
	}
}
