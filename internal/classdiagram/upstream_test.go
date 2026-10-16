// SPDX-License-Identifier: MIT
// Provenance: original

package classdiagram

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// The differential tests compare the model built by this package with the
// model upstream's ClassDiagramFactory builds, both dumped as text (see
// testdata/upstream/Harness.java and dump_test.go). The golden dumps are
// generated with the reference jar (upstreamgen_test.go).

const (
	casesDir        = "testdata/upstream/cases"
	goldenFile      = "testdata/upstream/upstream.txt"
	conformanceRoot = "../../testdata/conformance/cases"
	vegaRoot        = "../../.cache/vega"
	vegaGoldenFile  = "../../.cache/class-upstream/vega.txt"
)

// committedCases are the cases whose golden dumps are committed: the cases
// of testdata/upstream/cases, the generated link cases and the @startuml
// conformance cases.
func committedCases(t testing.TB) []testCase {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(casesDir, "*.txt"))
	if err != nil {
		t.Fatal(err)
	}
	var all []testCase
	for _, f := range files {
		cases, err := readCases(f)
		if err != nil {
			t.Fatal(err)
		}
		for i := range cases {
			cases[i].name = strings.TrimSuffix(filepath.Base(f), ".txt") + "/" + cases[i].name
		}
		all = append(all, cases...)
	}
	all = append(all, linkComboCases()...)
	all = append(all, declComboCases()...)
	conf, err := corpusCases(conformanceRoot, "conformance/")
	if err != nil {
		t.Fatal(err)
	}
	return append(all, conf...)
}

// vegaCases are the @startuml sources of upstream's Vega suite, fetched in
// .cache (not committed).
func vegaCases(t testing.TB) []testCase {
	t.Helper()
	cases, err := corpusCases(vegaRoot, "vega/")
	if err != nil {
		t.Fatal(err)
	}
	return cases
}

// readDumps splits a dump file into the lines of each case, by name.
func readDumps(t testing.TB, path string) map[string][]string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	result := map[string][]string{}
	var name string
	for _, line := range strings.Split(strings.TrimSuffix(string(data), "\n"), "\n") {
		if n, ok := strings.CutPrefix(line, "=== "); ok {
			name = n
		}
		result[name] = append(result[name], line)
	}
	return result
}

// knownDifferences are the cases where the Go model differs from upstream's,
// with the reason.
var knownDifferences = map[string]string{
	// Upstream's middle circle leaf has the lookahead (?=[-=.~]); the RE2
	// rewrite (see commandLinkClass) does not accept the '~' that starts a
	// second name written with the separator '~'.
	"linkcombos/864 A -0~B": "middle circle lookahead before '~'",
	// The preprocessor does not bundle themes yet (task 7.3).
	"conformance/styling/theme-cerulean.puml": "themes not bundled",
}

// isClassDump reports whether upstream parsed the case as a class diagram
// (the corpus also holds other diagram types, which give errors).
func isClassDump(dump []string) bool {
	return slices.ContainsFunc(dump, func(s string) bool {
		return s == "result ClassDiagram" || s == "result NewpagedDiagram"
	})
}

func compareDumps(t *testing.T, cases []testCase, golden map[string][]string) {
	t.Helper()
	matches, known, classes, classMatches := 0, 0, 0, 0
	for _, c := range cases {
		want, ok := golden[c.name]
		if !ok {
			t.Errorf("%s: no golden dump; regenerate with -tags upstreamgen", c.name)
			continue
		}
		isClass := isClassDump(want)
		if isClass {
			classes++
		}
		got := dumpCase(c)
		reason, isKnown := knownDifferences[c.name]
		if slices.Equal(got, want) {
			if isKnown {
				t.Errorf("%s: now matches upstream; remove it from knownDifferences", c.name)
			}
			matches++
			if isClass {
				classMatches++
			}
			continue
		}
		if isKnown {
			t.Logf("%s: known difference (%s)", c.name, reason)
			known++
			continue
		}
		for i := 0; i < max(len(got), len(want)); i++ {
			var g, w string
			if i < len(got) {
				g = got[i]
			}
			if i < len(want) {
				w = want[i]
			}
			if g != w {
				t.Errorf("%s: line %d\n go:   %s\n java: %s", c.name, i, g, w)
				break
			}
		}
	}
	t.Logf("%d of %d cases match (%d of %d class diagrams), %d known differences", matches, len(cases),
		classMatches, classes, known)
}

func TestUpstream(t *testing.T) {
	compareDumps(t, committedCases(t), readDumps(t, goldenFile))
}

// TestUpstreamVega compares the @startuml cases of the Vega suite: the
// class diagrams, and the errors of the class factory on the other types.
func TestUpstreamVega(t *testing.T) {
	if _, err := os.Stat(vegaGoldenFile); err != nil {
		t.Skip("no Vega golden dumps; run tools/corpus/fetch-vega.sh and the upstreamgen test")
	}
	compareDumps(t, vegaCases(t), readDumps(t, vegaGoldenFile))
}
