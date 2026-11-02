# AGENTS.md

gopuml translates PlantUML v1.2026.9 (the MIT variant, `plantuml-mit`) into
Go. Upstream is the reference: where the two disagree, this implementation is
wrong, unless `testdata/conformance/known-differences.yaml` lists the case.

## Porting

- Read the Java before writing Go. `tools/upstream/fetch.sh` unpacks the
  upstream sources into `.cache/upstream/src`.
- Every Go file names its source in its header: `// Provenance: original` for
  new code, or the upstream file it translates. The comment of
  `tools/provenance/main.go` gives the format, and `make provenance` checks
  it. Translate only from `plantuml-mit`.
- Find the Go counterpart of a Java class through these headers:
  `git grep -l 'Provenance: .*/PSystemSalt.java'`.
- Upstream packages that import each other in a cycle become one Go package,
  whose `doc.go` explains the file prefixes (see `internal/svek/doc.go`).

## Correctness is conformance

- Before you call a rendering change done, run the conformance suite as the
  "Conformance suite" section of `README.md` describes, including on arm64
  hosts.
- An entry of `known-differences.yaml` names the case and its reason. The
  pass thresholds (spec `upstream-compatibility`) only get tighter.
- When you add or change a `@startuml` conformance case, regenerate the golden
  dumps of `internal/sequencediagram`, `internal/classdiagram` and
  `internal/activitydiagram3`:
  `go test -tags upstreamgen -run TestGenerateUpstream ./internal/<package>`.

## Kinds of change

- A parity fix makes a supported diagram match upstream. Send it as a pull
  request with the fix and either a regression case or a removed entry of
  `known-differences.yaml`. A regression case goes under
  `testdata/conformance/cases/<capability>/`, with the front matter of its
  neighbors and without `scenarios:`. Render its reference with
  `make reference` and commit both.
- New behavior (a diagram type, a flag, an endpoint, a library API) starts as
  an OpenSpec change in `openspec/changes/` (`openspec init` sets up your
  tool), or as an issue.
- `TestScenarioCoverage` wants a case for every scenario of `openspec/specs`
  and of the delta specs of active changes, so write the cases together with
  the spec.

## Done

A change is ready for review when:

- `make lint test provenance wasm` pass, as in the `test` job of CI;
- a rendering change passes the conformance suite;
- the README's counts of cases, known differences and unsupported cases match
  `conformance-report.md`, and its table of supported diagrams matches
  `site/index.html` (`TestLandingPage`, `TestReleaseNotes`).

## Git and style

- Commits: `type(scope): subject`, with `feat`, `fix`, `docs`, `refactor`,
  `perf` or `chore`, and the package or area as the scope (`svek`, `cheneer`,
  `wasm`, `openspec`).
- Branches start with the type: `feat/`, `fix/`, `docs/`, `refactor/`,
  `chore/`. One change per branch and pull request.
- Headers in sentence case; a spaced hyphen ` - ` where a dash is needed.
