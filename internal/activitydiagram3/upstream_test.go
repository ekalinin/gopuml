// SPDX-License-Identifier: MIT
// Provenance: original

package activitydiagram3

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// The differential tests compare the instruction tree built by this
// package with the tree upstream's ActivityDiagramFactory3 builds, both
// dumped as text (see testdata/upstream/Harness.java and dump_test.go). The
// golden dumps are generated with the reference jar (upstreamgen_test.go).

const (
	casesDir        = "testdata/upstream/cases"
	goldenFile      = "testdata/upstream/upstream.txt"
	conformanceRoot = "../../testdata/conformance/cases"
	vegaRoot        = "../../.cache/vega"
	vegaGoldenFile  = "../../.cache/activity-upstream/vega.txt"
)

// committedCases are the cases whose golden dumps are committed: the cases
// of testdata/upstream/cases, the generated combinations and the
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
	all = append(all, generatedCases()...)
	conf, err := corpusCases(filepath.Join(conformanceRoot, "activity-diagram"), "conformance/")
	if err != nil {
		t.Fatal(err)
	}
	all = append(all, conf...)
	// The other conformance cases are compared when upstream parses them as
	// activity diagrams (Harness.java checks with the whole PSystemBuilder).
	others, err := corpusCases(conformanceRoot, "allconf/")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range others {
		if !strings.HasPrefix(c.name, "allconf/activity-diagram/") {
			all = append(all, c)
		}
	}
	return all
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
	"misc/vertical tab swimlane keyword (escaped)": "RegexConcat fox signature prefilter",
	// Java's CASE_INSENSITIVE folds ASCII letters only, RE2's (?i) uses
	// Unicode simple folding: the Kelvin sign (U+212A) matches "k" and the
	// long s (U+017F) matches "s" in Go only.
	"misc/kelvin kill (escaped)": "RE2 Unicode case folding",
	"misc/long s (escaped)":      "RE2 Unicode case folding",
}

func compareDumps(t *testing.T, cases []testCase, golden map[string][]string) {
	t.Helper()
	matches, known, skipped := 0, 0, 0
	for _, c := range cases {
		want, ok := golden[c.name]
		if !ok {
			t.Errorf("%s: no golden dump; regenerate with -tags upstreamgen", c.name)
			continue
		}
		if len(want) == 2 && want[1] == "not activity3" {
			skipped++
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
	t.Logf("%d of %d cases match, %d known differences, %d not activity diagrams upstream", matches,
		len(cases)-skipped, known, skipped)
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
