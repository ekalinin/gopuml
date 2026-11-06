// SPDX-License-Identifier: MIT
// Provenance: original

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ekalinin/gopuml/internal/conformance"
)

func TestUndecoratedWait(t *testing.T) {
	at := func(min, sec int) time.Time { return time.Date(2026, 10, 23, 12, min, sec, 0, time.UTC) }
	// India is 5:30 ahead of UTC: 12:45 there is 07:15 UTC, a decorated
	// minute, while minute 45 is not one. Upstream counts minutes from the
	// Unix time, as UTC.
	india := time.FixedZone("IST", 5*3600+30*60)
	window := 20 * time.Second
	for _, tc := range []struct {
		name string
		now  time.Time
		want time.Duration
	}{
		{"plain minute", at(16, 0), 0},
		{"window ends before a decorated minute", at(14, 30), 0},
		{"window reaches minute 15", at(14, 50), 70 * time.Second},
		{"in minute 1", at(1, 10), 50 * time.Second},
		{"in minute 13", at(13, 0), time.Minute},
		{"window reaches minute 30", at(29, 45), 75 * time.Second},
		{"across the hour", at(59, 50), 0},
		{"window reaches minute 1 of the next hour", at(0, 50), 70 * time.Second},
		{"Unix minutes, not local ones", time.Date(2026, 10, 23, 12, 45, 0, 0, india), time.Minute},
	} {
		if got := undecoratedWait(tc.now, window); got != tc.want {
			t.Errorf("%s: undecoratedWait(%s) = %s, want %s", tc.name, tc.now.Format(time.TimeOnly+" MST"), got, tc.want)
		}
	}
}

func TestGoImplementationUnsetsGraphvizDot(t *testing.T) {
	// Spec upstream-compatibility "Graph cases laid out by the built-in
	// engine": with GRAPHVIZ_DOT set, as in the reference image, the Go
	// implementation still lays the case out with its built-in engine.
	if runtime.GOOS == "windows" {
		t.Skip("the fake dot is a shell script")
	}
	cases, err := conformance.LoadCases(os.DirFS(filepath.Join("..", "..", "testdata", "conformance")), "cases")
	if err != nil {
		t.Fatal(err)
	}
	var c *conformance.Case
	for _, cc := range cases {
		if cc.ID == "graph-layout/class-layout" {
			c = cc
		}
	}
	if c == nil {
		t.Fatal("no case graph-layout/class-layout")
	}
	dir := t.TempDir()
	marker := filepath.Join(dir, "ran")
	fakeDot := filepath.Join(dir, "dot")
	if err := os.WriteFile(fakeDot, []byte("#!/bin/sh\ntouch "+marker+"\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "class-layout.puml"), []byte(c.Source), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GRAPHVIZ_DOT", fakeDot)
	useBuiltInLayout()
	g := &generator{root: dir, impl: "go"}
	if err := g.renderDiagramsGo([]diagramJob{{c: c, file: "class-layout.puml", outRoot: "out"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Error("the dot of GRAPHVIZ_DOT ran")
	}
	svg, err := os.ReadFile(filepath.Join(dir, "out", "graph-layout", "class-layout", "block0-page0.svg"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{">A</text>", ">B</text>"} {
		if !strings.Contains(string(svg), want) {
			t.Errorf("the SVG has no %s", want)
		}
	}
}

func TestReferenceScriptProfiles(t *testing.T) {
	// Spec upstream-compatibility "Case rendered in a security profile":
	// upstream renders the cases of each profile in its own JVM, with
	// PLANTUML_SECURITY_PROFILE set, and the others in LEGACY.
	plain := &conformance.Case{ID: "a/plain", Formats: []string{"svg"}}
	dotsvg := &conformance.Case{ID: "a/dotsvg", Formats: []string{"svg"}, Profile: "INTERNET_WITH_DOTSVG"}
	g := &generator{impl: "java"}
	files, _ := g.referenceRun("scratch", []diagramJob{
		{c: plain, file: "scratch/a/plain.puml", outRoot: "out"},
		{c: dotsvg, file: "scratch/a/dotsvg.puml", outRoot: "out"},
	})
	if got, want := string(files[filepath.Join("scratch", "manifest.tsv")]), "a/plain\t/work/scratch/a/plain.puml\tsvg\n"; got != want {
		t.Errorf("manifest.tsv = %q, want %q", got, want)
	}
	if got, want := string(files[filepath.Join("scratch", "manifest-INTERNET_WITH_DOTSVG.tsv")]), "a/dotsvg\t/work/scratch/a/dotsvg.puml\tsvg\n"; got != want {
		t.Errorf("manifest-INTERNET_WITH_DOTSVG.tsv = %q, want %q", got, want)
	}
	want := "set -e\n" +
		"java -cp " + jar + ":/opt/tools ReferenceRender /work/scratch/manifest.tsv /work/out\n" +
		"PLANTUML_SECURITY_PROFILE=INTERNET_WITH_DOTSVG java -cp " + jar + ":/opt/tools ReferenceRender /work/scratch/manifest-INTERNET_WITH_DOTSVG.tsv /work/out\n"
	if got := string(files[filepath.Join("scratch", "run.sh")]); got != want {
		t.Errorf("run.sh = %q, want %q", got, want)
	}
	if len(files) != 3 {
		t.Errorf("%d files, want the two manifests and run.sh", len(files))
	}
}

func TestGoImplementationProfile(t *testing.T) {
	// The Go implementation renders a case in the profile it names: SANDBOX
	// denies the include that LEGACY reads.
	dir := t.TempDir()
	src := "@startuml\n!include part.iuml\n@enduml\n"
	if err := os.WriteFile(filepath.Join(dir, "part.iuml"), []byte("Alice -> Bob\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"legacy", "sandbox"} {
		if err := os.WriteFile(filepath.Join(dir, name+".puml"), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	g := &generator{root: dir, impl: "go"}
	err := g.renderDiagramsGo([]diagramJob{
		{c: &conformance.Case{ID: "p/legacy", Formats: []string{"svg"}}, file: "legacy.puml", outRoot: "out"},
		{c: &conformance.Case{ID: "p/sandbox", Formats: []string{"svg"}, Profile: "SANDBOX"}, file: "sandbox.puml", outRoot: "out"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for id, want := range map[string]string{"p/legacy": "SEQUENCE", "p/sandbox": "ERROR"} {
		var meta conformance.Meta
		data, err := os.ReadFile(filepath.Join(dir, "out", filepath.FromSlash(id), "meta.json"))
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, &meta); err != nil {
			t.Fatal(err)
		}
		if len(meta.Blocks) != 1 || meta.Blocks[0].Type != want {
			t.Errorf("%s: blocks %+v, want one %s block", id, meta.Blocks, want)
		}
	}
}
