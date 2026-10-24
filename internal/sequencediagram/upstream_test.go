// SPDX-License-Identifier: MIT
// Provenance: original

package sequencediagram

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// The differential tests compare the model built by this package with the
// model upstream's SequenceDiagramFactory builds, both dumped as text (see
// testdata/upstream/Harness.java and dump_test.go). The golden dumps are
// generated with the reference jar (upstreamgen_test.go).

const (
	casesDir        = "testdata/upstream/cases"
	goldenFile      = "testdata/upstream/upstream.txt"
	conformanceRoot = "../../testdata/conformance/cases"
	vegaRoot        = "../../.cache/vega"
	vegaGoldenFile  = "../../.cache/sequence-upstream/vega.txt"
)

// committedCases are the cases whose golden dumps are committed: the cases
// of testdata/upstream/cases and the @startuml conformance cases.
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
// with the reason. They are limitations of other packages.
var knownDifferences = map[string]string{
	// Upstream's RegexConcat.match first checks a "fox signature": the
	// characters a pattern needs. A "[%s]+" leaf needs a space, tab, line
	// end, form feed or no-break space, so a vertical tab alone (which
	// Java's \s accepts) makes the command reject the line. The command
	// framework (internal/command/regex) does not emulate this prefilter.
	"whitespace/vertical tab participant (escaped)": "RegexConcat fox signature prefilter",
	// The preprocessor does not bundle themes yet (task 7.3).
	"conformance/styling/theme-cerulean.puml": "themes not bundled",
}

func compareDumps(t *testing.T, cases []testCase, golden map[string][]string) {
	t.Helper()
	matches, known := 0, 0
	for _, c := range cases {
		want, ok := golden[c.name]
		if !ok {
			t.Errorf("%s: no golden dump; regenerate with -tags upstreamgen", c.name)
			continue
		}
		got := dumpCase(c)
		reason, isKnown := knownDifferences[c.name]
		if slices.Equal(got, want) {
			if isKnown {
				t.Errorf("%s: now matches upstream; remove it from knownDifferences", c.name)
			}
			matches++
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
	t.Logf("%d of %d cases match, %d known differences", matches, len(cases), known)
}

func TestUpstream(t *testing.T) {
	compareDumps(t, committedCases(t), readDumps(t, goldenFile))
}

func TestUpstreamVega(t *testing.T) {
	if _, err := os.Stat(vegaGoldenFile); err != nil {
		t.Skip("no Vega golden dumps; run tools/corpus/fetch-vega.sh and the upstreamgen test")
	}
	compareDumps(t, vegaCases(t), readDumps(t, vegaGoldenFile))
}
