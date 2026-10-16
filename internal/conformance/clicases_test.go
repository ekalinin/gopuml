// SPDX-License-Identifier: MIT
// Provenance: original

package conformance

import (
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

func TestParseCLICase(t *testing.T) {
	data := `id: cli/two-diagrams
scenarios: ["cli: Two diagrams in one file"]
args: ["--svg", "flows.puml"]
files:
  flows.puml: |
    @startuml
    A -> B
    @enduml
`
	c, err := ParseCLICase("cli/two-diagrams.cli.yaml", []byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(c.Args, []string{"--svg", "flows.puml"}) {
		t.Errorf("args = %v", c.Args)
	}
	if !slices.Equal(c.Compare, []string{"exit", "files"}) {
		t.Errorf("default compare = %v", c.Compare)
	}
	if !strings.Contains(c.Files["flows.puml"], "A -> B") {
		t.Errorf("files = %v", c.Files)
	}
}

func TestParseCLICaseErrors(t *testing.T) {
	tests := []struct{ name, path, data, wantErr string }{
		{"id mismatch", "cli/a.cli.yaml", "id: cli/b\n", "does not match"},
		{"unknown compare", "cli/a.cli.yaml", "id: cli/a\ncompare: [exit, color]\n", `unknown compare value "color"`},
		{"escaping file", "cli/a.cli.yaml", "id: cli/a\nfiles:\n  ../x.puml: x\n", "relative path inside"},
		{"absolute file", "cli/a.cli.yaml", "id: cli/a\nfiles:\n  /tmp/x.puml: x\n", "relative path inside"},
		{"unknown field", "cli/a.cli.yaml", "id: cli/a\nargv: []\n", "invalid YAML"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCLICase(tt.path, []byte(tt.data))
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestParseHTTPCase(t *testing.T) {
	data := `id: http-server/svg-by-url
server: builtin
method: GET
path: /svg/{encoded}
source: "Alice -> Bob"
`
	c, err := ParseHTTPCase("http-server/svg-by-url.http.yaml", []byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(c.Compare, []string{"status", "headers", "body"}) {
		t.Errorf("default compare = %v", c.Compare)
	}
}

func TestParseHTTPCaseErrors(t *testing.T) {
	base := "id: http-server/a\nserver: builtin\nmethod: GET\npath: /svg/x\n"
	tests := []struct{ name, data, wantErr string }{
		{"bad server", strings.Replace(base, "builtin", "tomcat", 1), "server must be"},
		{"relative path", strings.Replace(base, "/svg/x", "svg/x", 1), "absolute path"},
		{"missing method", strings.Replace(base, "method: GET\n", "", 1), "method and an absolute path"},
		{"encoded without source", strings.Replace(base, "/svg/x", "/svg/{encoded}", 1), "source is empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseHTTPCase("http-server/a.http.yaml", []byte(tt.data))
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoadCLIAndHTTPCases(t *testing.T) {
	fsys := fstest.MapFS{
		"cases/cli/b.cli.yaml":          {Data: []byte("id: cli/b\n")},
		"cases/cli/a.cli.yaml":          {Data: []byte("id: cli/a\n")},
		"cases/http-server/x.http.yaml": {Data: []byte("id: http-server/x\nserver: builtin\nmethod: GET\npath: /serverinfo\n")},
		"cases/sequence-diagram/s.puml": {Data: []byte(validCase)},
	}
	cli, err := LoadCLICases(fsys, "cases")
	if err != nil {
		t.Fatal(err)
	}
	if len(cli) != 2 || cli[0].ID != "cli/a" || cli[1].ID != "cli/b" {
		t.Errorf("cli cases = %v", cli)
	}
	httpCases, err := LoadHTTPCases(fsys, "cases")
	if err != nil {
		t.Fatal(err)
	}
	if len(httpCases) != 1 || httpCases[0].ID != "http-server/x" {
		t.Errorf("http cases = %v", httpCases)
	}
}
