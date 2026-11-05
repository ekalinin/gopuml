# Design

## Context

No upstream class is ported: upstream has no harness that measures its rendering, and the comparison with it stays in `tools/bench` (`release: Comparison with upstream`). What the repository already has:

- `tools/bench` (`Provenance: original`) times the CLI of a release and the upstream jar on the curated cases, with a `timed` function that its tests replace. Its structure is the model for the new command.
- `internal/conformance.LoadCases` reads the curated cases with their `Type`; the fixtures of a case are in `<name>.files` next to it, and `tools/refgen` copies them next to the source before rendering.
- `.github/workflows/ci.yml` has the jobs `test` and `conformance`, runs on every pull request and every push to `main`, and has only `contents: read`. The conformance job adds `conformance-report.md` to `$GITHUB_STEP_SUMMARY` and uploads it with `if: ${{ !cancelled() }}`, and `TestCIWorkflow` checks that contract by parsing the workflow and mutating it.
- The ruleset of `main` forbids deletion and force pushes, and requires no status check.
- A pull request runs on its merge commit (`refs/pull/<n>/merge`), whose first parent is the commit of the base branch it merges into.

## Goals / Non-Goals

**Goals:**

- One command gives the same table in CI and on a developer's machine.
- Base and pull request run on the same runner, over the same cases, so that the table compares code, not machines or corpora.
- Allocation regressions fail the job; time is shown for the reviewer.

**Non-Goals:**

- Statistical tests on `ns/op`, such as benchstat's: the time does not fail the job, so a median with the spread of the rounds is enough to read it. The raw outputs are attached to the run in the format of `go test -bench`, so benchstat can be run on them locally.
- Benchmarks of internal packages: the table is per diagram type, through the public `Render`.
- A new module dependency: the benchmark lines are parsed by the command itself.

## Decisions

### D1. BenchmarkRender in the root package

`render_test.go` gets `BenchmarkRender` and a helper that groups the curated cases by `Type` (every type of the curated cases, `unsupported` and `welcome` included). Each sub-benchmark, named after the type, renders every source of the type to SVG with `b.ReportAllocs()` and `b.Loop()`. A case with fixtures renders with its `<name>.files` directory as `Options.Dir` and `<name>.puml` as `Options.Filename`, as the CLI renders the copied case in `tools/refgen`. `TestRenderBenchmarkTypes` checks, with the same helper, that the groups hold every type of the curated cases with the number of cases of each, and renders the first source of each group without a Go panic. The benchmark itself takes about a second per type, too long for `make test`.

This is design D5 of `reduce-render-allocations`, moved here.

Alternative: one sub-benchmark per case. The table would have about 480 rows, and a type with many small cases would hide in it.

### D2. One command for CI and local runs: tools/benchcmp

`go run ./tools/benchcmp -base <ref>` (`Provenance: original`, in the style of `tools/bench`):

1. adds a detached `git worktree` of `<ref>` under `.cache/benchcmp/base`, and removes it when it ends;
2. builds the test binary of the root package of the base and of the checked-out commit (`go test -c -o .cache/benchcmp/{base,head}.test .`);
3. runs both binaries `-count` times (6 by default) with `-test.run '^$' -test.bench '^BenchmarkRender$' -test.benchmem`, appending their outputs to `.cache/benchcmp/base.txt` and `head.txt`;
4. writes the table to `.cache/benchcmp/report.md` and to stdout, and exits with status 1 on a regression (D4).

`make bench-compare BASE=main` runs it locally, `BASE` defaulting to `main`. The commands it runs go through a function that the tests replace, as `tools/bench` replaces `timed`.

Alternative: shell steps in the workflow. The same comparison could then not be run locally, and its parsing and threshold could not be tested.

Alternative: `benchstat` (`golang.org/x/perf/cmd/benchstat`) for the table. It needs a download in CI, prints no Markdown, and knows no threshold, so the gate would still need code of its own.

### D3. Same machine, same cases, alternating rounds

Both binaries run with the checked-out commit as their working directory, so the base benchmark reads the cases of the commit: a pull request that adds a sequence case adds it to both sides, and the `sequence` row compares code only. The rounds alternate the order (base first in even rounds, head first in odd ones), so that a runner that slows down during the job affects both sides alike.

A base run that fails (for example, a base benchmark that cannot read a case format the pull request changed) does not fail the comparison: the results it printed before failing are kept, the types without a result are shown without a change, and the table notes the failure with the last lines of its output. A base that has no `BenchmarkRender` prints no result, and every type is shown without a change. A build or run failure of the checked-out commit fails the command with status 2.

Alternative: run each binary on the cases of its own commit. A pull request that adds cases would then show a regression of its type.

### D4. Medians and the 5% allocation gate

For each type and each metric, the table takes the median of the rounds (the mean of the two middle values for an even count) and the change `(head - base) / base`, written as `%+.1f%%`. A type whose median `B/op` or `allocs/op` changes by more than `+5.0%` is a regression: its row ends with `regression`, a line under the table names the regressed types, and the command exits with status 1. `ns/op` is shown with the spread of the rounds of each side (`min-max` around the median, in percent) and never fails. The threshold is a constant of `tools/benchcmp` with a `-threshold` flag.

Alternative: gate on `ns/op` too. GitHub's shared runners vary by 5-15% between runs, so the job would fail without a code change.

Alternative: a tighter threshold, such as 2%. The allocations of a render are stable, but `sync.Pool` and map growth depend on the garbage collector's timing; 5% leaves room for that until the job has a history of results.

### D5. The benchmark job of ci.yml

A `benchmark` job in `.github/workflows/ci.yml`, with `if: github.event_name == 'pull_request'` and `timeout-minutes: 20`:

- `actions/checkout@v5` with `fetch-depth: 2`, so that `HEAD^1`, the commit the pull request merges into, is in the clone;
- `actions/setup-go@v6` with `go-version-file: go.mod`;
- a step that runs `make bench-compare BASE=HEAD^1`;
- a step with `if: ${{ !cancelled() }}` that appends `.cache/benchcmp/report.md` to `$GITHUB_STEP_SUMMARY` when it exists;
- `actions/upload-artifact@v7` with `if: ${{ !cancelled() }}`, the name `benchmark-report`, and `report.md`, `base.txt` and `head.txt`.

It runs next to `test` and `conformance`, with the workflow's `contents: read`. `TestBenchmarkWorkflow` in `tools/benchcmp` parses `ci.yml`, checks these steps, and checks that mutations fail the check: no `fetch-depth: 2`, another base than `HEAD^1`, a summary or upload step with `if: success()`.

Alternative: a separate workflow file. It would duplicate the triggers and the Go setup, while the job belongs to the same pull request checks.

Alternative: run it on pushes to `main` too, comparing each merge with the previous commit. A regression would then show after it merged, as a failing `main`.

### D6. A deliberate regression is merged with the failing job

A pull request that needs more allocations (a feature that draws what was an error image before, for example) keeps a failing `benchmark` job, and its description explains the increase. The ruleset of `main` requires no status check, so the merge is not blocked, and the next pull requests compare with the new base.

Alternative: a label that turns the failure into a warning. The job would need `labeled` and `unlabeled` triggers, which in `ci.yml` would run `test` and `conformance` again on each label change.

### D7. This change lands before reduce-render-allocations

`reduce-render-allocations` drops its design D5 and its tasks that add the benchmark and record a baseline by hand. Its pull request is then the first one measured by this job, and its table shows the effect of each of its commits against `main`.

## Risks / Trade-offs

- [Allocations vary by more than 5% between rounds without a code change] → medians of 6 rounds; the raw outputs are attached to the run; the threshold is one constant and a flag, to be raised if a job fails without a cause.
- [The job adds minutes to each pull request] → about 12 types × 1 s × 12 runs plus two builds, 4 to 5 minutes, in parallel with the conformance job; `-count` lowers it if needed.
- [A pull request changes `tools/benchcmp` or the benchmark so that the job passes] → the head's command and benchmark run, as the head's harness runs the conformance suite; the diff of those files is reviewed like any other.
- [The worktree stays behind after an interrupted run] → it is under `.cache` (ignored), and the next run removes a leftover `.cache/benchcmp/base` worktree before adding it again.
- [The first pull request, this change's own, has no base benchmark] → D3: every type is shown without a change, and the job passes.

## Migration Plan

Land this change before `reduce-render-allocations`, whose plan already follows D7. Rolling back means removing the job from `ci.yml`; nothing else depends on it.
