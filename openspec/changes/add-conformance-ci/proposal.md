# Proposal

## Why

The conformance suite is the gate for every diagram type, but it runs only when someone starts `make conformance` by hand, and even then it never fails: the target passes an empty `-require`, and `testdata/conformance/supported.yaml` does not exist, so failures only show up in the report. The foundation design names a linux/amd64 CI run as the authoritative conformance run, and the CI workflow has none.

Follow-up changes (description diagrams, Smetana) will touch shared code such as `abel`, `svek` and `klimt`, where a regression in sequence, class or activity diagrams would go unnoticed. A local run on 2026-10-17 with every group required has 0 failures, so the gate can start green.

## What Changes

- A `conformance` job in `.github/workflows/ci.yml`, next to `test`, on every pull request and every push to `main`, on a linux/amd64 runner.
- The job builds the reference image (`tools/reference/Dockerfile`) with a layer cache. It caches the Vega corpus and the uncommitted references (`.cache/reference/vega`, `.cache/reference-png`) and renders them with Java only when their inputs change.
- When the job renders references, it fails if the committed curated references in `testdata/conformance/reference/` differ from the new render.
- The Go implementation renders the cases inside the reference image, so both implementations use the same Graphviz 14.0.1. A new `GO_IN_IMAGE=1` option of `make conformance` does this, in CI and on local linux/amd64 hosts. `make conformance` without it works as today.
- Every group is required (`REQUIRE=all`): any failing case that is not a known difference fails the job. That includes a drawing where the "not supported yet" error is expected.
- The job publishes `conformance-report.md` as a run artifact and in the run summary, whether it passes or fails.
- The README describes the CI gate and how to reproduce it locally.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `upstream-compatibility`: adds a requirement that the conformance suite runs in CI for every pull request and every push to `main`, with the Go outputs produced in the reference container, every group required and the report published.

## Impact

- **CI:** a second job in `.github/workflows/ci.yml`. The first run without a cache builds the image (about 133 MB) and renders all references with Java. Later runs restore both from the GitHub Actions cache. The job uses private-repository CI minutes on every pull request.
- **Makefile:** `conformance` gets the optional `GO_IN_IMAGE` mode. The default mode does not change.
- **Pull requests:** CI fails for a pull request that breaks a conformance case until the case is fixed or listed in `known-differences.yaml` with a reason. This also applies to follow-up changes that add a diagram type.
- **Product code:** none. The CLI, library, server and WASM build do not change.
- **arm64 hosts (Apple Silicon):** `GO_IN_IMAGE=1` does not work there, because Go binaries crash under amd64 emulation. Local runs on those hosts stay indicative, as the foundation design already states.
