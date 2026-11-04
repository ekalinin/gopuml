# render-performance Specification

## Purpose

Measures the cost of rendering the curated cases per diagram type, and compares every pull request with the commit it merges into, so that a change that makes rendering allocate more is noticed before it merges.

## Requirements

### Requirement: Render benchmark

The repository SHALL provide a Go benchmark that, for each diagram type of the curated conformance cases, renders the source of every case of that type to SVG, and reports the time, the allocated bytes and the number of allocations of one pass over the type. A case with fixtures SHALL render with its fixtures, as the conformance suite renders it.

#### Scenario: One result per diagram type

- **WHEN** `go test -run '^$' -bench BenchmarkRender -benchmem .` runs in the repository
- **THEN** it prints one result for each diagram type of the curated cases, among them `sequence`, `class` and `timing`, each with `ns/op`, `B/op` and `allocs/op`

### Requirement: Comparison with a base commit

A command SHALL compare the render benchmark of the checked-out commit with that of a base commit, on the same machine: it SHALL build both, run them in alternating rounds over the curated cases of the checked-out commit, and write a Markdown table with, for each diagram type, the median `ns/op`, `B/op` and `allocs/op` of the base and of the commit and their change in percent. CI SHALL run the comparison for every pull request, with the commit the pull request merges into as the base, and SHALL add the table to the run summary and attach it to the run, whether the comparison passes or fails.

#### Scenario: Comparison table

- **WHEN** the base gives `sequence` medians of 1000000 `ns/op`, 2000000 `B/op` and 20000 `allocs/op`, and the commit gives 900000 `ns/op`, 1000000 `B/op` and 12000 `allocs/op`
- **THEN** the row of `sequence` shows these medians with the changes `-10.0%`, `-50.0%` and `-40.0%`

#### Scenario: Type without a base result

- **WHEN** the base has no result for a diagram type, because it has no render benchmark or no case of that type
- **THEN** the table shows the results of the commit for that type without a change, and that type does not fail the comparison

#### Scenario: Pull request compared with its base

- **WHEN** a pull request is opened or updated
- **THEN** CI runs the comparison with the commit the pull request merges into as the base

#### Scenario: Table published on failure

- **WHEN** the comparison of a pull request fails
- **THEN** its table is still in the run summary and attached to the run

### Requirement: Allocation regressions fail the comparison

The comparison SHALL fail when the median `B/op` or the median `allocs/op` of a diagram type is more than 5% above that of the base, and the table SHALL mark each such type. A change of `ns/op` SHALL NOT fail the comparison.

#### Scenario: Allocation regression

- **WHEN** the median `B/op` of `class` is 2000000 on the base and 2200000 on the commit
- **THEN** the comparison fails and the table marks `class` as a regression

#### Scenario: Slower without more allocations

- **WHEN** the median `ns/op` of `class` is 20% above that of the base, and its `B/op` and `allocs/op` are less than 5% above
- **THEN** the comparison passes and the table shows the `+20.0%` change of `ns/op`
