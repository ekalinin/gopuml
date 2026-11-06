# Proposal

## Why

Since `switch-to-gogv`, gopuml lays graphs out with gogv unless a `dot` is named, but `make bench` still passes one Graphviz `dot` to both implementations through `GRAPHVIZ_DOT`, as the release spec asks. The comparison of v0.4.0 therefore measures gopuml with Graphviz, which is not what a user without `GRAPHVIZ_DOT` runs, and it says nothing about the cost of gogv against Graphviz. No issue tracks this, and no conformance case fails or is reported as not supported because of it.

## What Changes

- `make bench` (`tools/bench`) compares three columns instead of two: the upstream jar with a Graphviz `dot`, gopuml with the same `dot`, and gopuml with gogv, its built-in layout, run without `GRAPHVIZ_DOT`. Every row has the three columns: the start time and the rendering time of each diagram type.
- The environment of the run also names the gogv version of the CLI under test.
- The landing page shows the comparison of the notes of the release it names with all of its columns, each with its own participant and bar color. The page of v0.4.0 keeps its two columns: the notes of the next release carry the three-column table, and the page follows them.

Out of scope:

- The notes and the page of v0.4.0, which stay as published.
- A dot diagram row in the comparison: in the default profile of the command line, dot diagrams render upstream's suppressed-feature notice, not a layout.
- The render benchmark of pull requests (`BenchmarkRender`, `make bench-compare`), which already renders with gogv.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `release`: the requirement "Comparison with upstream" compares the upstream jar with `dot`, gopuml with the same `dot` and gopuml with gogv, and names the gogv version in the environment.
- `landing-page`: the requirement "Comparison with upstream on the page" shows every column of the comparison of the notes, not two implementations.

## Impact

**Code:** `tools/bench` (the third column, the environment of each column, the gogv version) and its test; `site/style.css` (a third participant of the comparison) and `landing_test.go`, if the check of the page needs it.

**Conformance:** the `.go.yaml` cases `release/comparison` and `landing-page/release` name the new scenarios. No reference, report count or golden dump changes.

**Documentation:** the comment of the `bench` target in the `Makefile`. README.md and site/index.html do not change until the next release, whose preparation runs the new `make bench`.

**Compatibility:** no change to the library, the command line or any output. `make bench` runs gopuml once more per row, which adds a few seconds to a run.
