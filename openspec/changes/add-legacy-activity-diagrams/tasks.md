# Tasks

## 1. Conformance cases first

`TestScenarioCoverage` also reads the delta specs of active changes, so the cases come first and `make test` passes on the branch from the start (design D7). Until task 4.1, the new legacy activity cases are reported as not supported yet, and `diagram-source/legacy-activity-before-state` fails, because it is drawn as a state diagram. From task 4.1 to group 5, the legacy activity cases are compared and fail, because the activity image is not ported yet.

- [x] 1.1 Add curated cases under `testdata/conformance/cases/`: one `.puml` per compared scenario of `legacy-activity-diagram` under `legacy-activity-diagram/` (with `type: activity`), and `diagram-source/legacy-activity-before-state` for "diagram-source: Legacy activity before state". Render their references with the reference image (`go run ./tools/refgen -suite curated -png`) and regenerate the golden dumps of `internal/sequencediagram`, `internal/classdiagram` and `internal/activitydiagram3` (`go test -tags upstreamgen -run TestGenerateUpstream <package>`). Verify that the new references and dumps are written and that the other committed references and dumps do not change
- [x] 1.2 Add `.go.yaml` cases for the scenarios that are not compared with Java (D7): "TXT for a legacy activity diagram" (`TestLegacyActivityTxtUnsupported` in `internal/activitydiagram`, task 4.1), "Legacy activity diagram without dot" (`TestLegacyActivityWithoutDot` in `internal/activitydiagram`, task 5.2) and "Legacy activity diagram in WASM" (`TestWasmLegacyActivityDiagram` in `cmd/plantuml-wasm`, task 6.1). Verify that `TestScenarioCoverage` passes and `make test` passes

## 2. UBrex commands

- [x] 2.1 Export from `internal/klimt` the UBrex expression building, matching, `exactMatch` and capture lookups, port the builder parts the commands use (`UBrexConcat`, `UBrexLeaf` with `end`, `spaceZeroOrMore` and `spaceOneOrMore`, `UBrexNamed`, `UBrexOptional`, `UBrexOr`), and let `regex.Result` be backed by UBrex captures through an interface of `internal/command/regex` (D3). Verify with unit tests: a builder expression matches and captures as the same expression written as a string, `Get` and `Lazy` of a capture-backed result follow upstream's `RegexResult(UMatcher)`, and `TestGoldenUbrex` and the creole tests pass
- [x] 2.2 Add `uml.UBrexSingle` (upstream's `UBrexSingleLineCommand2`), and `UBrexCommandFootboxIgnored`, `UBrexCommandRankDir` and `UBrexCommandHideShow2` in `internal/abel` (D3). Verify with unit tests that a UBrex command accepts only exact matches of the trimmed line, and that the three commands accept upstream's forms (`hide footbox`, `left to right direction`, `hide <<stereo>>`, `show Name`) and set the rank direction and the hide and show rules
- [x] 2.3 Make the UBrex color parser return upstream's UBrex part, moving it from `internal/klimt/color` into `internal/klimt` (D3), and move its unit test to that form. Verify that `make test` passes

## 3. Legacy activity diagram

- [x] 3.1 Port `ActivityDiagram` and `ConditionalContext` into `internal/activitydiagram` (D1, D2): `startIf`, `endif`, `getStart`, `getEnd`, the `updateLasts` hook, `createNote`, `createInnerActivity`, `concurrentActivity`, the DOT settings and the description. Verify with unit tests: the start is created once, `getEnd` with and without a suffix gives different ends, branches nest and close, and notes do not change the last activity
- [x] 3.2 Port `CommandLinkActivity` and `CommandLinkLongActivity`, with upstream's `StringUtils.manageArrowForCuca` and `getArrowDirection` where they are missing. Verify with parser unit tests built from upstream's patterns: the start and the ends `(*)`, `(*top)` and `(*1)`, quoted activities with and without `as`, bare names, bars, the omitted source, labels in brackets, dotted bodies, direction hints, bracket styles, `-*->`, URLs, stereotypes, colors, `in Partition`, `--> {`, and activities on several lines
- [x] 3.3 Port `UBrexCommandIf`, `UBrexCommandElse`, `UBrexCommandEndif`, `UBrexCommandPartition` and `UBrexCommandEndPartition`. Verify with parser unit tests: every `if` form (quoted label with and without `then`, with `as`, unquoted label, from the start, a bar, a quoted or bare activity), the branch links of `else`, nested branches, the errors of `else` and `endif` without `if`, partitions closed by `}` and by `end partition`, with a color and a stereotype
- [x] 3.4 Port `CommandFactoryNoteActivity` into `internal/activitydiagram` (D1). Verify with parser unit tests: single-line and multi-line notes in the four positions, with a color and a URL, linked to the last activity
- [x] 3.5 Port `ActivityDiagramFactory` with upstream's command list, and the export (D6). Verify with parser unit tests that a diagram with two branches, bars, labels and ends parses with upstream's description ("(10 activities)"), that the corpus case `vega/nonreg/simple/A0004`, read from `.cache/vega` when it is fetched (the Vega inputs are not committed), parses with "(12 activities)", and that `make test` passes

## 4. Registration and detection

- [x] 4.1 Register `activitydiagram.Factory` in `engine.Factories` between the class and description factories, and remove the `legacy activity` entry of `unsupportedTypes` (D4). Update `TestDetectUnsupportedUmlTypes` (no error for the legacy inputs) and `TestEmbeddedUnsupportedType` (`robust "Web" as WB`), and the `why` of `diagram-source/unsupported-uml-types.go.yaml` and `styling/embedded-unsupported-type.go.yaml`. Add `TestLegacyActivityTxtUnsupported` (TXT and UTXT return `UnsupportedFormatError` naming activity diagrams, with no output). Verify that it passes, that `make test` passes, and that in `make conformance` the legacy activity cases are compared instead of reported as not supported yet, while the sequence, class, description, state and activity (new syntax) groups keep their results

## 5. Activity layout and images

- [x] 5.1 Port `EntityImageActivity` into `internal/svek`, replacing its stub (D5). Verify that the curated legacy activity cases and `diagram-source/legacy-activity-before-state` pass in `make conformance`, and that `vega/nonreg/simple/A0004` passes
- [x] 5.2 Port `CucaDiagramSimplifierActivity`, `GroupMakerActivity` and `InnerActivity` into `internal/svek`, and run the simplifier in `svek.TextBlock` after the layout engine is resolved, skipped without `dot` (D5). Add `TestLegacyActivityWithoutDot` (a legacy activity diagram with a missing `dot` gives the Graphviz-not-found error diagram). Verify that it passes, that an inner activity renders with the activities upstream draws, and update the "not ported yet" list of `internal/svek/doc.go`

## 6. WASM and documentation

- [x] 6.1 Add `TestWasmLegacyActivityDiagram`, which renders a legacy activity diagram in the WASM build and gets the "graph layout is not available in this build" error with the error image. Verify that `go test ./cmd/plantuml-wasm` passes
- [x] 6.2 Verify that `make conformance REQUIRE=all` reports 0 failures and no known differences that pass
- [x] 6.3 Verify that every test named in the `.go.yaml` cases of 1.2 exists in its package (`GoCase.TestExists`)
- [x] 6.4 Update `README.md`: legacy activity diagrams in the table of supported types with SVG and PNG (needs `dot`), legacy activity removed from the list of unsupported types, the WASM and Graphviz notes, and the case and known-difference counts. Verify that the README's counts match the conformance report of 6.2
- [ ] 6.5 Push the branch, with the user's approval, and open a pull request. Verify that the `test` and `conformance` CI jobs pass, including the committed-reference check
- [x] 6.6 Run `openspec validate add-legacy-activity-diagrams --strict` and `make provenance`. Verify that both pass
