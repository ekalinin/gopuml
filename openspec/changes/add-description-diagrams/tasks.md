# Tasks

## 1. Conformance cases first

`TestScenarioCoverage` also reads the delta specs of active changes, so the cases come first and `make test` passes on the branch from the start. Until group 5, the new embedded-diagram cases fail in `make conformance`. Until task 3.6, the new description cases are reported as not supported yet, except the Problem Frames case, which fails with a syntax error until tasks 3.6 and 4.2.

- [x] 1.1 Add curated cases under `testdata/conformance/cases/`: one `.puml` per compared scenario of `description-diagram`, `diagram-source/description-detected`, and the styling scenarios "Use case diagram in a note" and "Sequence diagram in an activity" (design D7). Render their references with `make reference`, regenerate the golden dumps of `internal/sequencediagram`, `internal/classdiagram` and `internal/activitydiagram3`, which cover every case of the corpus (`go test -tags upstreamgen -run TestGenerateUpstream <package>`), and list the two embedded-diagram cases in the `exportExceptions` of the sequence and activity tests until task 5.2 (the activity layout test skips those exceptions). Verify that the new references and dumps are written and that the other committed references and dumps do not change
- [x] 1.2 Add `.go.yaml` cases for the scenarios that are not compared with Java (D7): "TXT for a description diagram" (`TestDescriptionTxtUnsupported`, task 3.5), "Description diagram without dot" (`TestDescriptionWithoutDot`, task 3.5), "Description diagram in WASM" (`TestWasmDescriptionDiagram`, task 6.1) and "Embedded diagram of an unsupported type" (`TestEmbeddedUnsupportedType`, task 5.2), and add "Archimate syntax" to `diagram-source/unsupported-uml-types.go.yaml`. Verify that `TestScenarioCoverage` passes and `make test` passes

## 2. Shared code moves

- [x] 2.1 Move `Labels`, `StringWithArrow` and their helpers from `internal/classdiagram/commands_link.go` to `internal/descdiagram`, exported, and make `internal/classdiagram` use them from there (D1). No logic changes. Verify that `make test` passes and that the sequence, class and activity groups of `make conformance` have the same results as on `main`
- [x] 2.2 Move the `map` and `json` commands from `internal/classdiagram/objectdiagram_commands.go` to `internal/abel`, generic over `abel.Diagram` instead of typed on `*ClassDiagram` (D1). Verify as in 2.1
- [x] 2.3 Add a `CleanID` hook to `abel.CucaHooks` and make `CucaDiagram.CleanID` call it, so that the shared commands use it (D2). Verify with a unit test that a diagram with an overriding hook gets its IDs cleaned by the note and package commands, and that `make test` passes

## 3. Description diagram

- [x] 3.1 Port `DescriptionDiagram` (`cleanId`, `isUsecase`, `makeDiagramReady`, `checkFinalError`) into `internal/descdiagram`. Verify with unit tests of `cleanId` on the short forms `:Actor:`, `(Use case)`, `[Component]` and `() Interface`
- [x] 3.2 Port `CommandCreateElementFull` into `internal/descdiagram` (regex, `executeArg` with the usecase, actor, component, port and circle logic, `isForbidden`), using the patterns of `internal/abel/descdiagram_command_createelement.go`. Verify with parser unit tests built from upstream's patterns: every keyword, `actor/` and `usecase/`, the short forms, `as` aliases, stereotypes and colors
- [x] 3.3 Port `CommandLinkElement` (regex, `getLinkType`, `getDirection`, `executeArg`, `getDummy`) into `internal/descdiagram`. Verify with parser unit tests of heads, line styles, direction hints, length, labels and bracket colors
- [ ] 3.4 Port `CommandCreateDomain` into `internal/descdiagram`. Verify with parser unit tests of `domain` with each of the five stereotypes and of `requirement`
- [ ] 3.5 Port `DescriptionDiagramFactory` with upstream's command list without the three archimate commands, and the description export with SVG and PNG, and TXT and UTXT rejected (D5). Add `TestDescriptionTxtUnsupported` (TXT returns `UnsupportedFormatError`) and `TestDescriptionWithoutDot` (a use case diagram with a missing `dot` gives the Graphviz-not-found error diagram). Verify that both tests pass
- [ ] 3.6 Register `descdiagram.Factory` in `engine.Factories` between the class and activity factories, remove `detectDescriptionOnly` and its call, and reduce the description entry of `unsupportedTypes` to the archimate pattern with the message "archimate diagrams are not supported yet" (D3). Extend `TestDetectUnsupportedUmlTypes` with `archimate #Business "Customer" as C`. Verify that `make test` passes and that the description cases of `make conformance` are compared instead of reported as not supported yet

## 4. Entity images

- [ ] 4.1 Port `EntityImagePort`, `AbstractEntityImageBorder` and `FrontierCalculator` into `internal/svek`, replacing the stubs in `generalimagebuilder.go` and `cluster.go` (D4). Verify that the Vega case `svg/interactive/SVG0005_Svek` and the curated "Linked ports" case pass in `make conformance`
- [ ] 4.2 Port `EntityImageDomain`, `EntityImageDesignedDomain`, `EntityImageMachine`, `EntityImageRequirement` and `BoxedCharacter` into `internal/svek` (D4). Verify that the curated "Machine and requirement" case passes in `make conformance`

## 5. Embedded diagrams

- [ ] 5.1 Add the embedded-diagram renderer to `core.RenderContext` and set it in the engine: it parses the lines as one block with the same render context and renders it as SVG or PNG (D6). Verify with a unit test that renders a nested sequence diagram to both formats and returns its size
- [ ] 5.2 Port upstream's `EmbeddedDiagram` atom (size from the rendered image, `UImageSvg` for SVG output without the `<?plantuml ...?>` instruction, `UImage` decoded from PNG otherwise, rendered once), make `klimt.SkinParam` and the skin parameter wrappers implement `EmbeddedDiagramMaker`, add `TestEmbeddedUnsupportedType`, remove the two embedded-diagram cases of 1.1 and the `tikz` cases from the `exportExceptions` of the sequence and activity tests, and update `internal/klimt/HANDOFF.md`. Verify that a sequence note with `{{ Bob -> Carol }}` renders to SVG and PNG instead of crashing, that `TestEmbeddedUnsupportedType` passes, and that the curated styling cases of 1.1 pass in `make conformance`

## 6. WASM, known differences and documentation

- [ ] 6.1 Add `TestWasmDescriptionDiagram`, which renders a deployment diagram in the WASM build and gets the "graph layout is not available in this build" error with the error image. Verify that `go test ./cmd/plantuml-wasm` passes
- [ ] 6.2 Remove the 5 `tikz` entries from `testdata/conformance/known-differences.yaml`. Verify that `make conformance REQUIRE=all` reports 0 failures, with the 10 Vega description cases and the 5 `tikz` cases passing
- [ ] 6.3 Verify that every test named in the `.go.yaml` cases of 1.2 exists in its package (`GoCase.TestExists`)
- [ ] 6.4 Update `README.md`: description diagrams (use case, component, deployment) in the table of supported types with SVG and PNG, archimate in the list of unsupported types, and embedded diagrams. Verify that the README's counts match the conformance report of 6.2
- [ ] 6.5 Push the branch and open a pull request. Verify that the `test` and `conformance` CI jobs pass, including the committed-reference check
- [ ] 6.6 Run `openspec validate add-description-diagrams --strict`. Verify that the change is reported as valid
