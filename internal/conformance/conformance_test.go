// SPDX-License-Identifier: MIT
// Provenance: original

//go:build conformance

package conformance

import (
	"bytes"
	"errors"
	"flag"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// The conformance suite compares on the host, with the pinned resvg and the
// fonts of the reference image (tools/reference/fetch-tools.sh):
//
//	make conformance IMPL=java
var (
	flagRoot    = flag.String("root", "../..", "repository root")
	flagActual  = flag.String("actual", "", "output tree of the implementation under test")
	flagImpl    = flag.String("impl", "go", "name of the implementation under test")
	flagReport  = flag.String("report", "", "report file (default <root>/conformance-report.md)")
	flagResvg   = flag.String("resvg", "", "resvg executable (default: resvg on PATH)")
	flagFonts   = flag.String("fonts", "", "directory of the fonts used by resvg")
	flagRequire = flag.String("require", "", `groups that must fully pass: "all", a comma-separated list, or empty to read testdata/conformance/supported.yaml`)
)

func TestConformance(t *testing.T) {
	if *flagActual == "" {
		t.Fatal("-actual is required")
	}
	root := *flagRoot
	known, err := LoadKnownDifferences(filepath.Join(root, "testdata/conformance/known-differences.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	r := &Runner{
		Reference: func(id string) string {
			if strings.HasPrefix(id, VegaPrefix) {
				return filepath.Join(root, ".cache/reference", filepath.FromSlash(id))
			}
			return filepath.Join(root, "testdata/conformance/reference", filepath.FromSlash(id))
		},
		Actual:     *flagActual,
		Rasterizer: Resvg{Path: *flagResvg, FontsDir: *flagFonts},
		Known:      known,
	}
	if dirExists(filepath.Join(root, ".cache/reference-png")) {
		r.PNGReference, r.PNGActual = filepath.Join(root, ".cache/reference-png"), *flagActual+"-png"
	}

	fsys := os.DirFS(filepath.Join(root, "testdata/conformance"))
	var cases []*Case
	var cliCases []*CLICase
	var httpCases []*HTTPCase
	if dirExists(filepath.Join(root, "testdata/conformance/cases")) {
		cases, err = LoadCases(fsys, "cases")
		if err != nil {
			t.Fatal(err)
		}
		if cliCases, err = LoadCLICases(fsys, "cases"); err != nil {
			t.Fatal(err)
		}
		if httpCases, err = LoadHTTPCases(fsys, "cases"); err != nil {
			t.Fatal(err)
		}
	}
	if vega := filepath.Join(root, ".cache/vega"); dirExists(vega) {
		vc, err := LoadVegaCases(os.DirFS(vega), ".")
		if err != nil {
			t.Fatal(err)
		}
		cases = append(cases, vc...)
	}

	var results []CaseResult
	for _, c := range cases {
		results = append(results, r.CompareDiagram(c))
	}
	for _, c := range cliCases {
		results = append(results, r.CompareCLI(c))
	}
	for _, c := range httpCases {
		results = append(results, r.CompareHTTP(c))
	}

	var report bytes.Buffer
	if err := WriteReport(&report, *flagImpl, results); err != nil {
		t.Fatal(err)
	}
	reportFile := *flagReport
	if reportFile == "" {
		reportFile = filepath.Join(root, "conformance-report.md")
	}
	if err := os.WriteFile(reportFile, report.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	groups, _ := Summarize(results)
	for _, s := range groups {
		t.Logf("%-14s %4d cases, %4d passed, %3d known, %3d failed", s.Name, s.Cases, s.Passed, s.Known, s.Failed)
	}

	required, err := requiredGroups(root)
	if err != nil {
		t.Fatal(err)
	}
	for i := range results {
		res := &results[i]
		if res.Status() == StatusFail && (required["all"] || required[res.Group]) {
			t.Errorf("%s (%s): %s", res.ID, res.Group, strings.Join(allFailures(res), "; "))
		}
	}
}

// requiredGroups returns the groups whose failures fail the test.
func requiredGroups(root string) (map[string]bool, error) {
	var names []string
	if *flagRequire != "" {
		names = strings.Split(*flagRequire, ",")
	} else {
		data, err := os.ReadFile(filepath.Join(root, "testdata/conformance/supported.yaml"))
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		if err := yaml.Unmarshal(data, &names); err != nil {
			return nil, err
		}
	}
	out := map[string]bool{}
	for _, n := range names {
		if n = strings.TrimSpace(n); n != "" {
			out[n] = true
		}
	}
	return out, nil
}

func dirExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

// TestReferencesReproducible renders the curated cases twice with the
// reference image and checks that both runs produce identical files:
//
//	make reference-check
func TestReferencesReproducible(t *testing.T) {
	root, err := filepath.Abs(*flagRoot)
	if err != nil {
		t.Fatal(err)
	}
	render := func(out string) string {
		t.Helper()
		cmd := exec.Command("go", "run", "./tools/refgen", "-suite", "curated", "-out", out)
		cmd.Dir = root
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("refgen: %v\n%s", err, output)
		}
		return filepath.Join(root, out)
	}
	a, b := render(".cache/conformance/repro-a"), render(".cache/conformance/repro-b")
	files := 0
	err = filepath.WalkDir(a, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(a, p)
		want, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		got, err := os.ReadFile(filepath.Join(b, rel))
		if err != nil {
			return err
		}
		if !bytes.Equal(want, got) {
			t.Errorf("%s differs between runs", rel)
		}
		files++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if files == 0 {
		t.Fatal("no files rendered")
	}
	t.Logf("%d files identical", files)
}
