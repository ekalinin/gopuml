// SPDX-License-Identifier: MIT
// Provenance: original

package conformance

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const goCaseSuffix = ".go.yaml"

// GoCase maps spec scenarios that cannot be compared with Java PlantUML (Go
// API behavior, intended differences from upstream, environment-dependent
// behavior) to the Go test that verifies them. It is stored as
// cases/<capability>/<name>.go.yaml.
type GoCase struct {
	ID        string   `yaml:"id"`
	Scenarios []string `yaml:"scenarios"`
	// Package is the Go package directory, relative to the repository root.
	Package string `yaml:"package"`
	// Test is the name of the test function.
	Test string `yaml:"test"`
	// Why explains why the scenario is not compared with Java PlantUML.
	Why string `yaml:"why"`

	Path string `yaml:"-"`
}

// ParseGoCase parses one .go.yaml file.
func ParseGoCase(relPath string, data []byte) (*GoCase, error) {
	var c GoCase
	if err := decodeStrict(relPath, data, &c); err != nil {
		return nil, err
	}
	c.Path = relPath
	err := errors.Join(checkID(c.ID, relPath, goCaseSuffix), checkScenarios(c.Scenarios))
	if len(c.Scenarios) == 0 {
		err = errors.Join(err, errors.New("at least one scenario is required"))
	}
	if c.Package == "" || !strings.HasPrefix(c.Test, "Test") || strings.TrimSpace(c.Why) == "" {
		err = errors.Join(err, errors.New("package, a Test* function name and why are required"))
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", relPath, err)
	}
	return &c, nil
}

// LoadGoCases reads every .go.yaml case below root, sorted by id.
func LoadGoCases(fsys fs.FS, root string) ([]*GoCase, error) {
	cases, err := loadBySuffix(fsys, root, goCaseSuffix, ParseGoCase)
	slices.SortFunc(cases, func(a, b *GoCase) int { return strings.Compare(a.ID, b.ID) })
	return cases, err
}

// TestExists reports whether the test function of the case is defined in the
// _test.go files of its package. repoRoot is the repository root.
func (c *GoCase) TestExists(repoRoot string) bool {
	files, _ := filepath.Glob(filepath.Join(repoRoot, filepath.FromSlash(c.Package), "*_test.go"))
	needle := []byte("func " + c.Test + "(")
	for _, f := range files {
		if data, err := os.ReadFile(f); err == nil && bytes.Contains(data, needle) {
			return true
		}
	}
	return false
}

// SpecScenarios returns "<capability>: <scenario>" for every scenario of the
// spec files, where the capability is the directory holding spec.md.
func SpecScenarios(specFiles []string) ([]string, error) {
	var out []string
	for _, f := range specFiles {
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		capability := filepath.Base(filepath.Dir(f))
		sc := bufio.NewScanner(bytes.NewReader(data))
		for sc.Scan() {
			if name, ok := strings.CutPrefix(sc.Text(), "#### Scenario:"); ok {
				out = append(out, capability+": "+strings.TrimSpace(name))
			}
		}
	}
	slices.Sort(out)
	return slices.Compact(out), nil
}

// SpecFiles finds the spec files of the project: the main specs and the
// delta specs of active (not archived) changes.
func SpecFiles(repoRoot string) ([]string, error) {
	main, err := filepath.Glob(filepath.Join(repoRoot, "openspec/specs/*/spec.md"))
	if err != nil {
		return nil, err
	}
	changes, err := filepath.Glob(filepath.Join(repoRoot, "openspec/changes/*/specs/*/spec.md"))
	if err != nil {
		return nil, err
	}
	return append(main, changes...), nil
}

// Coverage compares spec scenarios with the scenarios that cases reference.
// It returns the scenarios no case covers and the references that match no
// scenario.
func Coverage(scenarios []string, refs map[string][]string) (uncovered, unknown []string) {
	covered := map[string]bool{}
	for _, list := range refs {
		for _, s := range list {
			covered[s] = true
		}
	}
	for _, s := range scenarios {
		if !covered[s] {
			uncovered = append(uncovered, s)
		}
	}
	for id, list := range refs {
		for _, s := range list {
			if !slices.Contains(scenarios, s) {
				unknown = append(unknown, id+" -> "+s)
			}
		}
	}
	slices.Sort(unknown)
	return uncovered, unknown
}

// CaseScenarios loads every case kind below root and returns the scenarios
// each case id references.
func CaseScenarios(fsys fs.FS, root string) (map[string][]string, error) {
	refs := map[string][]string{}
	cases, err := LoadCases(fsys, root)
	if err != nil {
		return nil, err
	}
	for _, c := range cases {
		refs[c.ID] = c.Scenarios
	}
	cli, err := LoadCLICases(fsys, root)
	if err != nil {
		return nil, err
	}
	for _, c := range cli {
		refs[c.ID] = c.Scenarios
	}
	httpCases, err := LoadHTTPCases(fsys, root)
	if err != nil {
		return nil, err
	}
	for _, c := range httpCases {
		refs[c.ID] = c.Scenarios
	}
	goCases, err := LoadGoCases(fsys, root)
	if err != nil {
		return nil, err
	}
	for _, c := range goCases {
		refs[c.ID] = c.Scenarios
	}
	return refs, nil
}
