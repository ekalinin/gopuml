// SPDX-License-Identifier: MIT
// Provenance: original

package main

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// ciWorkflow is the part of a GitHub Actions workflow that
// checkBenchmarkJob reads.
type ciWorkflow struct {
	On   map[string]any    `yaml:"on"`
	Jobs map[string]*ciJob `yaml:"jobs"`
}

type ciJob struct {
	If    string   `yaml:"if"`
	Steps []ciStep `yaml:"steps"`
}

type ciStep struct {
	If   string         `yaml:"if"`
	Uses string         `yaml:"uses"`
	Run  string         `yaml:"run"`
	With map[string]any `yaml:"with"`
}

// checkBenchmarkJob checks the contract of the benchmark job of the CI
// workflow (render-performance: Comparison with a base commit).
func checkBenchmarkJob(w *ciWorkflow) error {
	var errs []error
	if _, ok := w.On["pull_request"]; !ok {
		errs = append(errs, errors.New("the workflow does not run on pull requests"))
	}
	job, ok := w.Jobs["benchmark"]
	if !ok {
		return errors.Join(append(errs, errors.New("no benchmark job"))...)
	}
	if job.If != "github.event_name == 'pull_request'" {
		errs = append(errs, errors.New("the job does not run on pull requests only"))
	}
	find := func(what string, match func(s ciStep) bool) int {
		i := slices.IndexFunc(job.Steps, match)
		if i < 0 {
			errs = append(errs, errors.New("no step that "+what))
		}
		return i
	}
	// The table steps also run when the comparison failed.
	afterFailure := func(s ciStep) bool {
		cond := strings.TrimSuffix(strings.TrimPrefix(s.If, "${{ "), " }}")
		return cond == "!cancelled()" || cond == "always()"
	}
	checkout := find("checks out the merge commit with its first parent", func(s ciStep) bool {
		return strings.HasPrefix(s.Uses, "actions/checkout@") && s.With["fetch-depth"] == 2
	})
	compare := find("compares with the commit the pull request merges into", func(s ciStep) bool {
		return strings.Contains(s.Run, "make bench-compare BASE=HEAD^1")
	})
	summary := find("adds the table to the run summary after a failure too", func(s ciStep) bool {
		return afterFailure(s) && strings.Contains(s.Run, ".cache/benchcmp/report.md") && strings.Contains(s.Run, "$GITHUB_STEP_SUMMARY")
	})
	upload := find("attaches the table to the run after a failure too", func(s ciStep) bool {
		path, _ := s.With["path"].(string)
		return afterFailure(s) && strings.HasPrefix(s.Uses, "actions/upload-artifact@") &&
			s.With["name"] == "benchmark-report" && strings.Contains(path, ".cache/benchcmp/report.md")
	})
	if len(errs) == 0 && !(checkout < compare && compare < summary && compare < upload) {
		errs = append(errs, errors.New("the steps are not in the order checkout, comparison, table"))
	}
	return errors.Join(errs...)
}

func TestBenchmarkWorkflow(t *testing.T) {
	// Spec scenarios "Pull request compared with its base" and "Table
	// published on failure".
	data, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	parse := func() *ciWorkflow {
		var w ciWorkflow
		if err := yaml.Unmarshal(data, &w); err != nil {
			t.Fatal(err)
		}
		return &w
	}
	if err := checkBenchmarkJob(parse()); err != nil {
		t.Fatal(err)
	}
	// stepOf returns the step of the benchmark job that match selects.
	stepOf := func(w *ciWorkflow, match func(s ciStep) bool) *ciStep {
		steps := w.Jobs["benchmark"].Steps
		i := slices.IndexFunc(steps, match)
		if i < 0 {
			t.Fatal("no such step")
		}
		return &steps[i]
	}
	isCheckout := func(s ciStep) bool { return strings.HasPrefix(s.Uses, "actions/checkout@") }
	isCompare := func(s ciStep) bool { return strings.Contains(s.Run, "bench-compare") }
	isSummary := func(s ciStep) bool { return strings.Contains(s.Run, "$GITHUB_STEP_SUMMARY") }
	isUpload := func(s ciStep) bool { return strings.HasPrefix(s.Uses, "actions/upload-artifact@") }
	for _, m := range []struct {
		name   string
		mutate func(w *ciWorkflow)
	}{
		{"on pushes too", func(w *ciWorkflow) { w.Jobs["benchmark"].If = "" }},
		{"no fetch-depth: 2", func(w *ciWorkflow) { delete(stepOf(w, isCheckout).With, "fetch-depth") }},
		{"base main", func(w *ciWorkflow) {
			s := stepOf(w, isCompare)
			s.Run = strings.ReplaceAll(s.Run, "BASE=HEAD^1", "BASE=main")
		}},
		{"summary skipped after a failure", func(w *ciWorkflow) { stepOf(w, isSummary).If = "success()" }},
		{"upload skipped after a failure", func(w *ciWorkflow) { stepOf(w, isUpload).If = "success()" }},
	} {
		w := parse()
		m.mutate(w)
		if checkBenchmarkJob(w) == nil {
			t.Errorf("%s: the check passes", m.name)
		}
	}
}
