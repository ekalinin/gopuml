// SPDX-License-Identifier: MIT
// Provenance: original

// Command benchcmp compares the render benchmark (BenchmarkRender of the
// root package) of the checked-out commit with that of a base commit, on the
// same machine, and fails when the allocations of a diagram type grow by
// more than a threshold.
//
// It adds a detached git worktree of the base under .cache/benchcmp/base,
// builds the test binaries of the root package of the base and of the
// checked-out commit, and runs both -count times in alternating order (base
// first in even rounds), with the checked-out commit as working directory so
// that both render its curated cases. It writes their outputs to
// .cache/benchcmp/base.txt and head.txt, and the table of the medians to
// .cache/benchcmp/report.md and stdout. The exit status is 1 when the
// median B/op or allocs/op of a diagram type is more than -threshold
// percent above the base, and 2 when the checked-out commit cannot be built
// or run. A base that cannot be built or run is noted in the table and does
// not fail the comparison.
//
// Usage: go run ./tools/benchcmp -base <ref> [-count n] [-threshold percent]
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// benchArgs run BenchmarkRender alone, with its allocations.
var benchArgs = []string{"-test.run", "^$", "-test.bench", "^BenchmarkRender$", "-test.benchmem"}

type benchcmp struct {
	base      string
	count     int
	threshold float64
	// run runs cmd and returns its combined output; tests replace it.
	run func(cmd *exec.Cmd) ([]byte, error)
}

func main() {
	var b benchcmp
	flag.StringVar(&b.base, "base", "", "base commit, as a git revision (required)")
	flag.IntVar(&b.count, "count", 6, "rounds of each side")
	flag.Float64Var(&b.threshold, "threshold", defaultThreshold, "increase of B/op or allocs/op, in percent, that fails a diagram type")
	flag.Parse()
	if b.base == "" || b.count < 1 || flag.NArg() > 0 {
		flag.Usage()
		os.Exit(2)
	}
	os.Exit(b.main(os.Stdout, os.Stderr))
}

// side is one of the two compared commits.
type side struct {
	bin    string
	out    bytes.Buffer
	failed bool
}

// main runs the comparison and returns the exit status.
func (b *benchcmp) main(stdout, stderr io.Writer) int {
	if b.run == nil {
		b.run = func(cmd *exec.Cmd) ([]byte, error) { return cmd.CombinedOutput() }
	}
	fail := func(format string, args ...any) int {
		fmt.Fprintf(stderr, "benchcmp: "+format+"\n", args...)
		return 2
	}
	out, err := b.exec("", "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return fail("not in a git checkout: %v\n%s", err, out)
	}
	root := strings.TrimSpace(string(out))
	dir := filepath.Join(root, ".cache", "benchcmp")
	worktree := filepath.Join(dir, "base")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fail("%v", err)
	}
	b.removeWorktree(root, worktree)
	if out, err := b.exec(root, "git", "worktree", "add", "--detach", worktree, b.base); err != nil {
		return fail("worktree of %s: %v\n%s", b.base, err, out)
	}
	defer b.removeWorktree(root, worktree)
	h := header{base: fmt.Sprintf("`%s` (%s)", b.base, b.shortID(root, b.base)), head: b.shortID(root, "HEAD"),
		rounds: b.count, threshold: b.threshold}

	base := &side{bin: filepath.Join(dir, "base.test")}
	head := &side{bin: filepath.Join(dir, "head.test")}
	var notes []string
	if out, err := b.exec(worktree, "go", "test", "-c", "-o", base.bin, "."); err != nil {
		base.failed = true
		notes = append(notes, failureNote("The base could not be built", err, out))
	}
	if out, err := b.exec(root, "go", "test", "-c", "-o", head.bin, "."); err != nil {
		return fail("building the commit: %v\n%s", err, out)
	}
	for round := range b.count {
		order := []*side{base, head}
		if round%2 == 1 {
			order = []*side{head, base}
		}
		for _, s := range order {
			if s.failed {
				continue
			}
			out, err := b.exec(root, s.bin, benchArgs...)
			s.out.Write(out)
			if err == nil {
				continue
			}
			if s == head {
				return fail("running the commit's benchmark: %v\n%s", err, out)
			}
			s.failed = true
			notes = append(notes, failureNote(fmt.Sprintf("The base run of round %d failed", round+1), err, out))
		}
	}

	baseResults, err := parseResults(bytes.NewReader(base.out.Bytes()))
	if err != nil {
		return fail("%v", err)
	}
	headResults, err := parseResults(bytes.NewReader(head.out.Bytes()))
	if err != nil {
		return fail("%v", err)
	}
	rows := compare(baseResults, headResults, b.threshold)
	var report bytes.Buffer
	writeMarkdown(&report, h, rows, notes)
	for name, data := range map[string][]byte{"base.txt": base.out.Bytes(), "head.txt": head.out.Bytes(), "report.md": report.Bytes()} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			return fail("%v", err)
		}
	}
	stdout.Write(report.Bytes())
	if len(regressions(rows)) > 0 {
		return 1
	}
	return 0
}

func (b *benchcmp) exec(dir, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return b.run(cmd)
}

// removeWorktree removes the worktree of the base, if any, with its files.
func (b *benchcmp) removeWorktree(root, worktree string) {
	b.exec(root, "git", "worktree", "remove", "--force", worktree)
	os.RemoveAll(worktree)
	b.exec(root, "git", "worktree", "prune")
}

// shortID returns the abbreviated commit id of rev, or "?".
func (b *benchcmp) shortID(root, rev string) string {
	out, err := b.exec(root, "git", "rev-parse", "--short", rev)
	if err != nil {
		return "?"
	}
	return strings.TrimSpace(string(out))
}

// failureNote is a note of the report on a failing base, with the last
// lines of its output.
func failureNote(what string, err error, out []byte) string {
	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	lines = lines[max(0, len(lines)-10):]
	return fmt.Sprintf("%s (%v); its last lines:\n\n```\n%s\n```", what, err, strings.Join(lines, "\n"))
}
