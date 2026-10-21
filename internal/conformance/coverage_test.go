// SPDX-License-Identifier: MIT
// Provenance: original

package conformance

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

func TestParseGoCase(t *testing.T) {
	ok := "id: library-api/concurrent\nscenarios: [\"library-api: Concurrent renders\"]\npackage: .\ntest: TestRenderConcurrent\nwhy: Go API behavior\n"
	if _, err := ParseGoCase("library-api/concurrent.go.yaml", []byte(ok)); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{
		"id: library-api/concurrent\npackage: .\ntest: TestX\nwhy: x\n",                                    // no scenario
		"id: library-api/concurrent\nscenarios: [\"library-api: X\"]\npackage: .\ntest: RenderX\nwhy: x\n", // not a test name
		"id: library-api/concurrent\nscenarios: [\"library-api: X\"]\npackage: .\ntest: TestX\n",           // no why
		"id: library-api/other\nscenarios: [\"library-api: X\"]\npackage: .\ntest: TestX\nwhy: x\n",        // id mismatch
	} {
		if _, err := ParseGoCase("library-api/concurrent.go.yaml", []byte(bad)); err == nil {
			t.Errorf("expected an error for %q", bad)
		}
	}
}

func TestGoCaseTestExists(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pkg", "a_test.go"), []byte("package pkg\nfunc TestHere(t *testing.T) {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !(&GoCase{Package: "pkg", Test: "TestHere"}).TestExists(dir) {
		t.Error("TestHere should exist")
	}
	if (&GoCase{Package: "pkg", Test: "TestMissing"}).TestExists(dir) {
		t.Error("TestMissing should not exist")
	}
}

func TestSpecScenariosAndCoverage(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "cli", "spec.md")
	if err := os.MkdirAll(filepath.Dir(spec), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "### Requirement: X\n#### Scenario: Missing file\n- **WHEN** a\n#### Scenario: No arguments\n"
	if err := os.WriteFile(spec, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	scenarios, err := SpecScenarios([]string{spec})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(scenarios, []string{"cli: Missing file", "cli: No arguments"}) {
		t.Fatalf("scenarios = %q", scenarios)
	}
	uncovered, unknown := Coverage(scenarios, map[string][]string{
		"cli/missing-file": {"cli: Missing file"},
		"cli/typo":         {"cli: Mising file"},
	})
	if !slices.Equal(uncovered, []string{"cli: No arguments"}) {
		t.Errorf("uncovered = %q", uncovered)
	}
	if !slices.Equal(unknown, []string{"cli/typo -> cli: Mising file"}) {
		t.Errorf("unknown = %q", unknown)
	}
}

func TestCaseScenarios(t *testing.T) {
	fsys := fstest.MapFS{
		"cases/sequence-diagram/self-message.puml": {Data: []byte(validCase)},
		"cases/cli/missing.cli.yaml":               {Data: []byte("id: cli/missing\nscenarios: [\"cli: Missing file\"]\n")},
		"cases/library-api/c.go.yaml":              {Data: []byte("id: library-api/c\nscenarios: [\"library-api: Concurrent renders\"]\npackage: .\ntest: TestC\nwhy: Go API\n")},
	}
	refs, err := CaseScenarios(fsys, "cases")
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 3 || refs["cli/missing"][0] != "cli: Missing file" || refs["library-api/c"][0] != "library-api: Concurrent renders" {
		t.Fatalf("refs = %v", refs)
	}
}

// TestScenarioCoverage checks that every scenario of the project's specs is
// covered by at least one curated case, and that cases only reference
// existing scenarios.
func TestScenarioCoverage(t *testing.T) {
	root := filepath.Join("..", "..")
	files, err := SpecFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Skip("no spec files")
	}
	scenarios, err := SpecScenarios(files)
	if err != nil {
		t.Fatal(err)
	}
	refs, err := CaseScenarios(os.DirFS(filepath.Join(root, "testdata/conformance")), "cases")
	if err != nil {
		t.Fatal(err)
	}
	uncovered, unknown := Coverage(scenarios, refs)
	if len(unknown) > 0 {
		t.Errorf("cases reference unknown scenarios:\n  %s", strings.Join(unknown, "\n  "))
	}
	if len(uncovered) > 0 {
		t.Errorf("%d of %d scenarios have no case:\n  %s", len(uncovered), len(scenarios), strings.Join(uncovered, "\n  "))
	}
}

// TestGoCaseTestsExist checks that the test named by every .go.yaml case
// exists in its package, so that the scenarios these cases cover are
// tested.
func TestGoCaseTestsExist(t *testing.T) {
	root := filepath.Join("..", "..")
	cases, err := LoadGoCases(os.DirFS(filepath.Join(root, "testdata/conformance")), "cases")
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("no Go case")
	}
	for _, c := range cases {
		if !c.TestExists(root) {
			t.Errorf("%s: no %s in %s", c.ID, c.Test, c.Package)
		}
	}
}
