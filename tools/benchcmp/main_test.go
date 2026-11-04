// SPDX-License-Identifier: MIT
// Provenance: original

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// benchLine is a line of go test -bench -benchmem for the type.
func benchLine(typ string, ns, bytes, allocs float64) string {
	return fmt.Sprintf("BenchmarkRender/%s-4   \t      10\t %s ns/op\t %s B/op\t %s allocs/op\n",
		typ, number(ns), number(bytes), number(allocs))
}

// compareTexts compares two outputs of go test -bench as the command
// compares base.txt and head.txt, and returns the report and the regressed
// types.
func compareTexts(t *testing.T, base, head string) (string, []string) {
	t.Helper()
	b, err := parseResults(strings.NewReader(base))
	if err != nil {
		t.Fatal(err)
	}
	h, err := parseResults(strings.NewReader(head))
	if err != nil {
		t.Fatal(err)
	}
	rows := compare(b, h, defaultThreshold)
	var sb strings.Builder
	writeMarkdown(&sb, header{base: "`main` (abc1234)", head: "def5678", rounds: 1, threshold: defaultThreshold}, rows, nil)
	return sb.String(), regressions(rows)
}

// tableRow returns the line of the type in the report.
func tableRow(t *testing.T, report, typ string) string {
	t.Helper()
	for line := range strings.SplitSeq(report, "\n") {
		if strings.HasPrefix(line, "| "+typ+" |") {
			return line
		}
	}
	t.Fatalf("no row for %s in\n%s", typ, report)
	return ""
}

func TestParseResults(t *testing.T) {
	out := `goos: linux
goarch: amd64
pkg: github.com/ekalinin/gopuml
BenchmarkRender/sequence-4  12  98765432 ns/op  2000000 B/op  20000 allocs/op
BenchmarkRender/class 5 1 ns/op 2 B/op 3 allocs/op
BenchmarkRender/chen_eer-12 3 4 ns/op 5 B/op 6 allocs/op
BenchmarkRender/nomem-4 10 7 ns/op
BenchmarkOther-4 1 8 ns/op 9 B/op 10 allocs/op
PASS
ok  	github.com/ekalinin/gopuml	1.234s
`
	got, err := parseResults(strings.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	want := results{
		"sequence": {{98765432, 2000000, 20000}},
		"class":    {{1, 2, 3}},
		"chen_eer": {{4, 5, 6}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMedianAndSpread(t *testing.T) {
	// Even count: the mean of the two middle values.
	s := summarize([]metrics{{30, 3, 1}, {10, 1, 1}, {20, 2, 1}, {100, 10, 1}})
	if s.n != 4 || s.ns != 25 || s.bytes != 2.5 || s.allocs != 1 {
		t.Errorf("summary %+v", s)
	}
	if s.nsLow != -60 || s.nsHigh != 300 {
		t.Errorf("spread %v, %v, want -60, +300", s.nsLow, s.nsHigh)
	}
	// Odd count: the middle value.
	if s := summarize([]metrics{{3, 0, 0}, {1, 0, 0}, {2, 0, 0}}); s.ns != 2 {
		t.Errorf("median %v, want 2", s.ns)
	}
}

func TestComparisonTable(t *testing.T) {
	// Spec scenario "Comparison table".
	report, regressed := compareTexts(t,
		benchLine("sequence", 1000000, 2000000, 20000),
		benchLine("sequence", 900000, 1000000, 12000))
	want := "| sequence | 1000000 (+0.0%, +0.0%) | 900000 (+0.0%, +0.0%) | -10.0% | 2000000 | 1000000 | -50.0% | 20000 | 12000 | -40.0% |  |"
	if got := tableRow(t, report, "sequence"); got != want {
		t.Errorf("row\n got: %s\nwant: %s", got, want)
	}
	if len(regressed) != 0 || !strings.Contains(report, "No allocation regression above +5.0%.") {
		t.Errorf("regressions %v in\n%s", regressed, report)
	}
}

func TestTypeWithoutBaseResult(t *testing.T) {
	// Spec scenario "Type without a base result": a type the base does not
	// know, and a base without the benchmark.
	head := benchLine("sequence", 1, 2, 3) + benchLine("timing", 1, 2, 3)
	report, regressed := compareTexts(t, benchLine("sequence", 1, 2, 3), head)
	want := "| timing |  | 1 (+0.0%, +0.0%) |  |  | 2 |  |  | 3 |  | no base result |"
	if got := tableRow(t, report, "timing"); got != want {
		t.Errorf("row\n got: %s\nwant: %s", got, want)
	}
	if len(regressed) != 0 {
		t.Errorf("regressions %v", regressed)
	}
	report, regressed = compareTexts(t, "PASS\nok  \tgithub.com/ekalinin/gopuml\t0.5s\n", head)
	for _, typ := range []string{"sequence", "timing"} {
		if row := tableRow(t, report, typ); !strings.HasSuffix(row, "| no base result |") {
			t.Errorf("row %s", row)
		}
	}
	if len(regressed) != 0 {
		t.Errorf("regressions %v", regressed)
	}
}

func TestAllocationRegression(t *testing.T) {
	// Spec scenario "Allocation regression".
	report, regressed := compareTexts(t,
		benchLine("class", 1000000, 2000000, 20000)+benchLine("sequence", 1, 2000000, 20000),
		benchLine("class", 1000000, 2200000, 20000)+benchLine("sequence", 1, 2000000, 21001))
	if row := tableRow(t, report, "class"); !strings.Contains(row, "| +10.0% |") || !strings.HasSuffix(row, "| regression |") {
		t.Errorf("row %s", row)
	}
	if !slices.Equal(regressed, []string{"class", "sequence"}) {
		t.Errorf("regressions %v, want class and sequence", regressed)
	}
	if !strings.Contains(report, "Allocation regressions above +5.0%: class, sequence.") {
		t.Errorf("no regression line in\n%s", report)
	}
	// Exactly +5.0% is not above the threshold.
	_, regressed = compareTexts(t, benchLine("class", 1, 2000000, 20000), benchLine("class", 1, 2100000, 21000))
	if len(regressed) != 0 {
		t.Errorf("+5.0%%: regressions %v", regressed)
	}
}

func TestSlowerWithoutMoreAllocations(t *testing.T) {
	// Spec scenario "Slower without more allocations".
	report, regressed := compareTexts(t,
		benchLine("class", 1000000, 2000000, 20000),
		benchLine("class", 1200000, 2080000, 20800))
	want := "| class | 1000000 (+0.0%, +0.0%) | 1200000 (+0.0%, +0.0%) | +20.0% | 2000000 | 2080000 | +4.0% | 20000 | 20800 | +4.0% |  |"
	if got := tableRow(t, report, "class"); got != want {
		t.Errorf("row\n got: %s\nwant: %s", got, want)
	}
	if len(regressed) != 0 {
		t.Errorf("regressions %v", regressed)
	}
}

// fakeBenchcmp returns a benchcmp whose commands are recorded in calls (with
// the checkout's root written <root>) and answered by answer, except git's.
func fakeBenchcmp(t *testing.T, count int, answer func(name, dir string) (string, error)) (b *benchcmp, root string, calls *[]string) {
	root = t.TempDir()
	calls = new([]string)
	b = &benchcmp{base: "main", count: count, threshold: defaultThreshold, run: func(cmd *exec.Cmd) ([]byte, error) {
		name := filepath.Base(cmd.Args[0])
		*calls = append(*calls, strings.ReplaceAll(name+" "+strings.Join(cmd.Args[1:], " ")+" @"+cmd.Dir, root, "<root>"))
		switch {
		case name == "git" && slices.Contains(cmd.Args, "--show-toplevel"):
			return []byte(root + "\n"), nil
		case name == "git" && cmd.Args[1] == "rev-parse":
			return []byte("abc1234\n"), nil
		case name == "git":
			return nil, nil
		}
		out, err := answer(name, cmd.Dir)
		return []byte(out), err
	}}
	return b, root, calls
}

// runsOf returns the names of the benchmark binaries run, in order.
func runsOf(calls []string) []string {
	var out []string
	for _, c := range calls {
		if name, _, _ := strings.Cut(c, " "); strings.HasSuffix(name, ".test") {
			out = append(out, c)
		}
	}
	return out
}

func TestRunAlternatesRoundsInTheRoot(t *testing.T) {
	b, root, calls := fakeBenchcmp(t, 3, func(name, dir string) (string, error) {
		if name == "go" {
			return "", nil
		}
		return benchLine("sequence", 1, 2, 3), nil
	})
	var stdout, stderr strings.Builder
	if code := b.main(&stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	args := strings.Join(benchArgs, " ")
	var want []string
	for _, bin := range []string{"base", "head", "head", "base", "base", "head"} {
		want = append(want, fmt.Sprintf("%s.test %s @<root>", bin, args))
	}
	if got := runsOf(*calls); !slices.Equal(got, want) {
		t.Errorf("runs\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	for _, c := range []string{
		"git worktree add --detach <root>/.cache/benchcmp/base main @<root>",
		"go test -c -o <root>/.cache/benchcmp/base.test . @<root>/.cache/benchcmp/base",
		"go test -c -o <root>/.cache/benchcmp/head.test . @<root>",
	} {
		if !slices.Contains(*calls, c) {
			t.Errorf("no call %q in\n%s", c, strings.Join(*calls, "\n"))
		}
	}
	if last := (*calls)[len(*calls)-2]; last != "git worktree remove --force <root>/.cache/benchcmp/base @<root>" {
		t.Errorf("worktree not removed at the end: %s", last)
	}
	data, err := os.ReadFile(filepath.Join(root, ".cache", "benchcmp", "base.txt"))
	if err != nil || strings.Count(string(data), "BenchmarkRender/sequence") != 3 {
		t.Errorf("base.txt %q, %v", data, err)
	}
	report, err := os.ReadFile(filepath.Join(root, ".cache", "benchcmp", "report.md"))
	if err != nil || string(report) != stdout.String() || !strings.Contains(stdout.String(), "Base: `main` (abc1234). Commit: abc1234. Rounds: 3.") {
		t.Errorf("report.md %q, %v; stdout %q", report, err, stdout.String())
	}
}

func TestRunExitsOneOnRegression(t *testing.T) {
	b, _, _ := fakeBenchcmp(t, 1, func(name, dir string) (string, error) {
		switch name {
		case "base.test":
			return benchLine("class", 1, 100, 3), nil
		case "head.test":
			return benchLine("class", 1, 106, 3), nil
		}
		return "", nil
	})
	var stdout, stderr strings.Builder
	if code := b.main(&stdout, &stderr); code != 1 || !strings.Contains(stdout.String(), "| regression |") {
		t.Errorf("exit %d, report\n%s", code, stdout.String())
	}
}

func TestRunWithFailingBase(t *testing.T) {
	b, _, calls := fakeBenchcmp(t, 3, func(name, dir string) (string, error) {
		switch name {
		case "base.test":
			return benchLine("sequence", 1, 2, 3) + "--- FAIL: BenchmarkRender/timing\nFAIL\n", errors.New("exit status 1")
		case "head.test":
			return benchLine("sequence", 1, 2, 3) + benchLine("timing", 1, 2, 3), nil
		}
		return "", nil
	})
	var stdout, stderr strings.Builder
	if code := b.main(&stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	report := stdout.String()
	if row := tableRow(t, report, "sequence"); !strings.Contains(row, "| +0.0% |") {
		t.Errorf("sequence row %s", row)
	}
	if row := tableRow(t, report, "timing"); !strings.HasSuffix(row, "| no base result |") {
		t.Errorf("timing row %s", row)
	}
	if !strings.Contains(report, "The base run of round 1 failed (exit status 1); its last lines:") ||
		!strings.Contains(report, "--- FAIL: BenchmarkRender/timing") {
		t.Errorf("no note on the failing base in\n%s", report)
	}
	runs := runsOf(*calls)
	if n := len(runs); n != 4 {
		t.Errorf("%d runs, want 1 of the base and 3 of the commit:\n%s", n, strings.Join(runs, "\n"))
	}
}

func TestRunWithFailingCommit(t *testing.T) {
	for _, failing := range []string{"build", "run"} {
		t.Run(failing, func(t *testing.T) {
			b, _, calls := fakeBenchcmp(t, 2, func(name, dir string) (string, error) {
				switch {
				case failing == "build" && name == "go" && !strings.HasSuffix(dir, filepath.Join(".cache", "benchcmp", "base")):
					return "compile error", errors.New("exit status 1")
				case failing == "run" && name == "head.test":
					return "panic", errors.New("exit status 2")
				}
				return benchLine("sequence", 1, 2, 3), nil
			})
			var stdout, stderr strings.Builder
			if code := b.main(&stdout, &stderr); code != 2 {
				t.Errorf("exit %d, want 2", code)
			}
			if !strings.Contains(stderr.String(), "benchcmp: ") {
				t.Errorf("stderr %q", stderr.String())
			}
			if last := (*calls)[len(*calls)-2]; last != "git worktree remove --force <root>/.cache/benchcmp/base @<root>" {
				t.Errorf("worktree not removed after the failure: %s", last)
			}
		})
	}
}

func TestBenchmarkWorkflow(t *testing.T) {
	t.Skip("task 4.1")
}
