// SPDX-License-Identifier: MIT
// Provenance: original

package conformance

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// ciWorkflow is the part of a GitHub Actions workflow that
// checkConformanceJob reads.
type ciWorkflow struct {
	On   map[string]any `yaml:"on"`
	Jobs map[string]struct {
		Steps []ciStep `yaml:"steps"`
	} `yaml:"jobs"`
}

type ciStep struct {
	ID   string         `yaml:"id"`
	If   string         `yaml:"if"`
	Uses string         `yaml:"uses"`
	Run  string         `yaml:"run"`
	With map[string]any `yaml:"with"`
}

// checkConformanceJob checks the contract of the conformance job of the CI
// workflow (upstream-compatibility: Continuous conformance).
func checkConformanceJob(data []byte) error {
	var w ciWorkflow
	if err := yaml.Unmarshal(data, &w); err != nil {
		return err
	}
	var errs []error
	if _, ok := w.On["pull_request"]; !ok {
		errs = append(errs, errors.New("the workflow does not run on pull requests"))
	}
	push, _ := w.On["push"].(map[string]any)
	branches, _ := push["branches"].([]any)
	if !slices.Contains(branches, any("main")) {
		errs = append(errs, errors.New("the workflow does not run on pushes to main"))
	}
	job, ok := w.Jobs["conformance"]
	if !ok {
		return errors.Join(append(errs, errors.New("no conformance job"))...)
	}
	find := func(what string, match func(s ciStep) bool) int {
		i := slices.IndexFunc(job.Steps, match)
		if i < 0 {
			errs = append(errs, fmt.Errorf("no step that %s", what))
		}
		return i
	}
	cacheMissed := "steps.references.outputs.cache-hit != 'true'"
	// The report steps also run when an earlier step failed.
	afterFailure := func(s ciStep) bool {
		cond := strings.TrimSuffix(strings.TrimPrefix(s.If, "${{ "), " }}")
		return cond == "!cancelled()" || cond == "always()"
	}

	cache := find("caches the references with the committed references in its key", func(s ciStep) bool {
		key, _ := s.With["key"].(string)
		return s.ID == "references" && strings.HasPrefix(s.Uses, "actions/cache@") && strings.Contains(key, "'testdata/conformance/reference/**'")
	})
	render := find("renders the references on a cache miss", func(s ciStep) bool {
		return s.If == cacheMissed && strings.Contains(s.Run, "go run ./tools/refgen -image plantuml-go-reference -png")
	})
	check := find("fails when the render changed the committed references", func(s ciStep) bool {
		return s.If == cacheMissed && strings.Contains(s.Run, "git status --porcelain -- testdata/conformance/reference") && strings.Contains(s.Run, "exit 1")
	})
	run := find("runs the suite in the reference image with every group required", func(s ciStep) bool {
		return strings.Contains(s.Run, "make conformance REQUIRE=all GO_IN_IMAGE=1")
	})
	summary := find("adds the report to the run summary after a failure too", func(s ciStep) bool {
		return afterFailure(s) && strings.Contains(s.Run, "conformance-report.md") && strings.Contains(s.Run, "$GITHUB_STEP_SUMMARY")
	})
	upload := find("uploads the report after a failure too", func(s ciStep) bool {
		path, _ := s.With["path"].(string)
		return afterFailure(s) && strings.HasPrefix(s.Uses, "actions/upload-artifact@") && path == "conformance-report.md"
	})
	if len(errs) == 0 && !(cache < render && render < check && check < run && run < summary && run < upload) {
		errs = append(errs, errors.New("the steps are not in the order reference cache, render, check, conformance, report"))
	}
	return errors.Join(errs...)
}

func TestCIWorkflow(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := checkConformanceJob(data); err != nil {
		t.Fatal(err)
	}
	for _, m := range []struct{ name, old, new string }{
		{"not every group required", "REQUIRE=all ", ""},
		{"committed references not in the cache key", ", 'testdata/conformance/reference/**'", ""},
		{"no reference check", "git status --porcelain", "git diff --stat"},
		{"report skipped after a failure", "if: ${{ !cancelled() }}", "if: success()"},
	} {
		mutated := strings.ReplaceAll(string(data), m.old, m.new)
		if mutated == string(data) {
			t.Errorf("%s: %q is not in the workflow", m.name, m.old)
			continue
		}
		if checkConformanceJob([]byte(mutated)) == nil {
			t.Errorf("%s: the check passes", m.name)
		}
	}
}
