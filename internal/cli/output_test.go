// SPDX-License-Identifier: MIT
// Provenance: original

package cli

import (
	"path/filepath"
	"slices"
	"testing"
)

type block struct {
	name    string
	hasName bool
	pages   int
}

func names(n *OutputNamer, blocks ...block) []string {
	var out []string
	for _, b := range blocks {
		out = append(out, n.Files(b.name, b.hasName, b.pages)...)
	}
	return out
}

func TestOutputNames(t *testing.T) {
	unnamed := block{pages: 1}
	login := block{name: "login", hasName: true, pages: 1}
	tests := []struct {
		name   string
		namer  OutputNamer
		blocks []block
		want   []string
	}{
		{"two diagrams in one file", OutputNamer{Source: "flows.puml", OutputDir: ".", Suffix: ".svg", Separator: "_"},
			[]block{unnamed, unnamed}, []string{"flows.svg", "flows_001.svg"}},
		{"named diagram", OutputNamer{Source: "flows.puml", OutputDir: ".", Suffix: ".svg", Separator: "_"},
			[]block{login}, []string{"login.svg"}},
		{"named and unnamed blocks", OutputNamer{Source: "flows.puml", OutputDir: ".", Suffix: ".svg", Separator: "_"},
			[]block{unnamed, login, unnamed}, []string{"flows.svg", "login.svg", "flows_001.svg"}},
		{"duplicate names overwrite", OutputNamer{Source: "flows.puml", OutputDir: ".", Suffix: ".svg", Separator: "_"},
			[]block{login, login}, []string{"login.svg", "login.svg"}},
		{"separator minus", OutputNamer{Source: "flows.puml", OutputDir: ".", Suffix: ".png", Separator: "-"},
			[]block{unnamed, unnamed}, []string{"flows.png", "flows-001.png"}},
		{"ignore names", OutputNamer{Source: "flows.puml", OutputDir: ".", Suffix: ".svg", Separator: "_", IgnoreNames: true},
			[]block{login, unnamed}, []string{"flows.svg", "flows_001.svg"}},
		{"pages", OutputNamer{Source: "flows.puml", OutputDir: ".", Suffix: ".svg", Separator: "_"},
			[]block{{pages: 2}, unnamed, {name: "seq", hasName: true, pages: 2}},
			[]string{"flows.svg", "flows_001.svg", "flows_002.svg", "seq.svg", "seq_001.svg"}},
		{"source without extension", OutputNamer{Source: "README", OutputDir: ".", Suffix: ".atxt", Separator: "_"},
			[]block{unnamed}, []string{"README.atxt"}},
		{"name with dot", OutputNamer{Source: "a.puml", OutputDir: ".", Suffix: ".svg", Separator: "_"},
			[]block{{name: "my.diagram", hasName: true, pages: 1}}, []string{"my.svg"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := tt.namer
			if got := names(&n, tt.blocks...); !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOutputNameDirectory(t *testing.T) {
	dir := t.TempDir()
	n := OutputNamer{Source: filepath.Join(dir, "flows.puml"), OutputDir: dir, Suffix: ".svg", Separator: "_"}
	got := n.Files("out/", true, 1)
	if want := filepath.Join(dir, "out", "flows.svg"); len(got) != 1 || got[0] != want {
		t.Errorf("got %v, want %s", got, want)
	}
}

func TestResolveOutputDir(t *testing.T) {
	if got := ResolveOutputDir("docs/a.puml", ""); got != "docs" {
		t.Errorf("default: %q", got)
	}
	// Spec scenario "Relative output directory": relative to the source.
	if got := ResolveOutputDir("docs/a.puml", "out"); got != filepath.Join("docs", "out") {
		t.Errorf("relative: %q", got)
	}
	if got := ResolveOutputDir("docs/a.puml", "/tmp/x"); got != "/tmp/x" {
		t.Errorf("absolute: %q", got)
	}
	if got := ResolveOutputDir("docs/a.puml", "../out"); got != "out" {
		t.Errorf("parent: %q", got)
	}
}
