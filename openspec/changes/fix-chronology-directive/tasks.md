# Tasks

## 1. Conformance cases first

`TestScenarioCoverage` also reads the delta specs of active changes, so the case comes first and `make test` passes on the branch from the start. Until task 2.1, `make conformance` reports the new case as not supported yet, as this implementation still draws "chronology diagrams are not supported yet".

- [x] 1.1 Add the curated case `testdata/conformance/cases/diagram-source/chronology-directive.puml` for "diagram-source: Chronology directive", with the front matter of `diagram-source/unknown-directive` (`family: chronology`, `type: unsupported`, `formats: [svg]`) and the block `@startchronology`, `foo`, `@endchronology`. Render its reference with `make reference`. No `@startuml` case is added, so the golden dumps are not regenerated. Verify that the reference is upstream's unsupported diagram with the texts `Diagram not supported by this release of PlantUML` and `@startchronology`, that the other committed references do not change, and that `TestScenarioCoverage` and `make test` pass

## 2. Fix

- [x] 2.1 In `CheckFamily` of `internal/core/diagram.go`, return upstream's unsupported diagram (`UnsupportedDiagram` with the start line as its directive) for `Chronology`, as for `Unknown`, with a comment that upstream's MIT variant recognizes the start tag but registers no factory for it (`ChronologyDiagramFactory` is commented out in `PSystemBuilder`). Add a `@startchronology` check to `TestCheckFamily`: it gives an `*UnsupportedDiagram` whose directive is `@startchronology` and whose description is `(Unsupported)`. Verify that the test and `make test` pass, that `-syntax` with the block `@startchronology`, `foo`, `@endchronology` on standard input prints nothing and exits with status 200, as upstream, and that a sequence diagram with `Alice -> Bob` and a note that holds `{{chronology`, `foo`, `}}` embeds an image of 587 by 220, the size of upstream's
- [x] 2.2 Make `diagram-source/chronology-directive` pass in `make conformance REQUIRE=all GO_IN_IMAGE=1` (on arm64 hosts, the native `make conformance REQUIRE=all`, which gives the results of CI). Verify that the report has 0 failures and no known differences that pass, that its not-supported-yet count stays at 5, and that the results of the other cases do not change

## 3. Documentation and checks

- [x] 3.1 Update the README's case count, which follows `conformance-report.md` of task 2.2; the counts of known differences and of cases not supported yet do not change. Leave `site/index.html` unchanged: its conformance tally follows the release on the page. Verify that `TestLandingPage` and `TestReleaseNotes` pass
- [x] 3.2 Verify that `make lint` and `make provenance` pass
- [x] 3.3 Push the branch, with the user's approval, and open a pull request that closes issue #58. Verify that the `test` and `conformance` CI jobs pass, including the committed-reference check
- [x] 3.4 Run `openspec validate fix-chronology-directive --strict`. Verify that the change is reported as valid
