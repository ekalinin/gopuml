# Proposal

## Why

Nothing measures the cost of rendering when the code changes. The conformance suite catches a changed output, but a pull request that doubles the allocations of every render passes CI unnoticed: the profiling behind `reduce-render-allocations` found 1 MB of deflate tables allocated per render, a cost no check would have reported. `make bench` compares a release with the upstream jar and needs Java and a `dot`, so it does not run on pull requests. No issue tracks this, and no conformance case fails or is reported as not supported because of it.

## What Changes

- A Go benchmark, `BenchmarkRender`, renders the curated cases of each diagram type to SVG and reports `ns/op`, `B/op` and `allocs/op` per type. It comes from design D5 of `reduce-render-allocations`, which this change lands before and which no longer adds it.
- A command, `tools/benchcmp`, compares the benchmark of the checked-out commit with that of a base commit on the same machine: it builds both, runs them in alternating rounds over the same cases, and writes a Markdown table of the medians and their changes per diagram type. `make bench-compare BASE=<ref>` runs it locally.
- The comparison fails when the median `B/op` or `allocs/op` of a diagram type grows by more than 5%. Changes of `ns/op` are shown but do not fail it, since the time varies between runs of shared CI runners.
- CI runs the comparison on every pull request, against the commit the pull request merges into, in a `benchmark` job of `ci.yml`. The table goes to the run summary and is attached to the run as an artifact, also when the job fails, as the conformance report does.

Out of scope:

- The comparison with the upstream jar, which stays with `make bench` for releases.
- The start time of the command line and the PNG, TXT and UTXT outputs.
- A history of results across pull requests, comments on pull requests, and a way to mark a regression as accepted: the ruleset of `main` requires no status check, so a deliberate regression is merged with the failing job and explained in the pull request.

## Capabilities

### New Capabilities

- `render-performance`: the render benchmark, the comparison of a commit with a base commit, and the allocation regressions that fail the comparison of a pull request

### Modified Capabilities

None.

## Impact

**Code:** the root package (`BenchmarkRender` and its test), `tools/benchcmp` (new), `Makefile` (`bench-compare`), `.github/workflows/ci.yml` (`benchmark` job).

**Conformance:** new `.go.yaml` cases under `testdata/conformance/cases/render-performance/`. No reference, report count or golden dump changes.

**Documentation:** a "Performance check" section in README.md, after "Conformance suite". site/index.html does not change.

**Compatibility:** no change to the library, the command line or any output. Each pull request runs one more CI job of a few minutes. `reduce-render-allocations` loses its design D5 and its manual baseline tasks, and is measured by this job instead.
