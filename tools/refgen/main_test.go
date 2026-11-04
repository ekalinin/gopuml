// SPDX-License-Identifier: MIT
// Provenance: original

package main

import (
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
