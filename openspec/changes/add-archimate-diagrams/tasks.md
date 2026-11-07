# Tasks

## 1. Conformance cases first

`TestScenarioCoverage` also reads the delta specs of active changes, so the cases come first and `make test` passes on the branch from the start (design D6). Until tasks 2.2 and 3.2, the new cases fail in the conformance suite: the sprite cases crash in `SpriteImageFromInternal`, and the archimate cases are reported as not supported yet.

- [x] 1.1 Add the curated cases of design D6, with the inputs of their scenarios and the front matter of their neighbors:
  - under `testdata/conformance/cases/description-diagram/`: `archimate-element`, `archimate-element-updated`, `archimate-multiline`, `archimate-group`, `archimate-unknown-icon`, `archimate-without-color` and `archimate-multiline-twice`, for the scenarios of "Archimate elements";
  - under `testdata/conformance/cases/sprites/`: `internal-svg-sprite`, `internal-svg-sprite-sax`, `internal-png-sprite`, `internal-sprite-stereotype`, `internal-sprite-command` and `missing-internal-sprite`, for the scenarios of "Internal sprites", and `internal-sprites-all` without scenarios, with one note line `<$archimate/<name>> <name>` per file of `.cache/upstream/src/sprites/archimate/`, in name order;
  - `testdata/conformance/cases/standard-library/archimate-library`, for "standard-library: Archimate library", with `!include <archimate/Archimate>`, `Business_Actor(a, "Customer")`, `Business_Service(s, "Sales")` and `Rel_Serving(s, a, "serves")`.

  Verify that `TestScenarioCoverage` finds a case for every scenario of the delta specs.
- [x] 1.2 Render the references of the new cases with `make reference` (design D6), and regenerate the golden dumps of `internal/sequencediagram`, `internal/classdiagram` and `internal/activitydiagram3` (`go test -tags upstreamgen -run TestGenerateUpstream ./internal/<package>`). Verify that the references show what the scenarios state:
  - the `#FFFFCC` element with the business-actor icon for `archimate-element`;
  - one element `Updated` for `archimate-element-updated`;
  - `«nope»` for `archimate-unknown-icon`;
  - `Syntax Error? (Assumed diagram type: sequence)` for `archimate-without-color`;
  - `Already exists c (Assumed diagram type: component)` for `archimate-multiline-twice`;
  - `No such internal sprite: archimate/nope (Assumed diagram type: sequence)` for `missing-internal-sprite`;
  - the same icon paths in `internal-svg-sprite` and `internal-svg-sprite-sax`.

  Verify that the other committed references and golden dumps do not change. Some tests of `make test` fail on a new case before the task that implements it: the error images of `TestDiagramReferences` in `internal/errdiagram` for the two archimate error cases, and the sequence golden dump tests for `sprite $a jar:archimate/actor`, which panics today. Skip that case in that test until the task that implements it, with a comment naming the task. Verify that `make test` passes
- [x] 1.3 Update the `why` of the `.go.yaml` cases `diagram-source/unsupported-uml-types` and `styling/embedded-unsupported-type` (design D5). The first: upstream renders archimate and the state and legacy activity syntax as diagrams, while this implementation must not report them as not supported yet. The second: upstream draws the embedded JSON diagram, while this implementation draws its "not supported yet" error image. Verify that `TestScenarioCoverage` and `GoCase.TestExists` pass

## 2. Internal sprites

- [x] 2.1 Copy the 139 files of `.cache/upstream/src/sprites/archimate/` unchanged into `internal/klimt/sprites/archimate/` (design D1, D3). Add `TestUpstreamInternalSprites`, which compares the embedded tree with `.cache/upstream/src/sprites/` through a function that returns the problems and skips when the sources are not unpacked. Add a unit test of that function on two `fstest.MapFS`: one changed byte names the file, a missing file and an extra file are reported, identical trees give no problem. Add the NOTICE paragraph on the sprites under "Bundled assets", with Archi's copyright line and the MIT License. Verify that `go test ./internal/klimt` passes with the unpacked sources
- [x] 2.2 Port `SpriteImage.fromInternal` into `internal/klimt/sprite_internal.go` and delete `sprite_internalnames.go` (design D1, D2). Replace the panic expectation of `klimt_test.go` with tests of the new function:
  - `SpriteImageFromInternal("archimate/actor")` is an `ISvgSpriteParser`;
  - `"archimate/access"` is a `*SpriteImage` of 16 by 16 pixels, and so are the two PNG files with a color map;
  - every embedded name gives a non-nil sprite;
  - `"archimate/nope"` and `"../stdlib/home"` give nil;
  - `"archimate/actor.svg"` panics with `IllegalArgumentException`.

  Remove the skip of task 1.2 from the sequence golden dump tests. Verify that `make test` passes, and that `rectangle "Customer" <<$archimate/business-actor>> as c` and `Alice -> Bob : <$archimate/actor> hi` render as SVG and PNG without "drawing crashed"

## 3. Archimate commands

- [x] 3.1 Port `CommandArchimate`, `CommandArchimateMultilines` and `CommandArchimatePackage` into `internal/descdiagram/commands_archimate.go`, and register them in `factory.go` after `abel.CommandCreateJSONSingleLine` and before `commandCreateDomain` (design D1, D4). Verify with unit tests on parsed diagrams:
  - `archimate #Business "Customer" <<business-actor>> as c` gives one description leaf `c` with the display `Customer`, the symbol `ARCHIMATE`, the back color `#FFFFCC` and a stereotype whose sprite name is `archimate/business-actor`;
  - `archimate #Business Customer`, `archimate #Business "Display" as d`, `archimate #Business d as "Display"` and `archimate #Business Display as d` give the codes and displays of upstream's four name forms;
  - `archimate #Business "Old" as u` followed by `archimate #Application "Updated" <<application-component>> as u` gives one leaf with the display `Updated`, the back color `#C2F0FF` and the application-component stereotype;
  - `archimate #Business m <<business-actor>> [`, `first line`, `second line`, `]` gives a leaf with the symbol `RECTANGLE` and the two lines as its display, and the same block twice for `c` fails at the second `]` with `Already exists c`;
  - `archimate #Technology "Domain" <<technology-node>> as d {`, `archimate #Application "App" as a`, `}` gives a group `d` with the symbol `ARCHIMATE` and the back color `#C9FFC9`, which contains `a`.

  Verify that `go test ./internal/descdiagram` passes
- [x] 3.2 Empty `unsupportedTypes` in `internal/core/detect.go`, remove `patterns` and the `regex` import, and give the file the header `Provenance: original` (design D5). Update the tests:
  - `TestDetectUnsupportedUmlTypes` expects the description type for `archimate #Business "Customer" as C`, keeps its state and legacy activity inputs, and checks the error of a test entry that it sets in `unsupportedTypes`;
  - `TestEmbeddedUnsupportedType` uses `{{json`, `{"a": 1}`, `}}` and expects "json diagrams are not supported yet".

  Remove the skips of task 1.2 from `TestDiagramReferences`. Verify that `archimate "Customer" as c` fails at that line with `Syntax Error? (Assumed diagram type: sequence)`, that the error images of `archimate-without-color` and `archimate-multiline-twice` match their references, and that `make test` passes
- [x] 3.3 Make the new cases pass in `make conformance REQUIRE=all GO_IN_IMAGE=1` (design D6). On arm64 hosts, run the native `make conformance REQUIRE=all`, which gives the results of CI. Fix the differences in `internal/klimt`, `internal/descdiagram` or `internal/svek`. List a case that still differs in `testdata/conformance/known-differences.yaml`, with its reason, only when its fix is outside this change. Verify:
  - the report has 0 failures and no known differences that pass;
  - no archimate case is reported as not supported yet, and the not-supported-yet count stays at 5;
  - the results of the other diagram types do not change

## 4. Documentation and checks

- [x] 4.1 Update the documentation (design D7):
  - in `README.md` and `site/index.html`: the row "Use case, component, deployment and archimate" of both tables of supported diagrams, and "Other diagram types (JSON, YAML, and so on)";
  - the README's case, known-difference and not-supported-yet counts, which follow `conformance-report.md` of task 3.3;
  - the comment of `commands()` in `internal/descdiagram/factory.go`;
  - the sprite paragraph of `internal/klimt/HANDOFF.md`.

  Verify that `TestLandingPage` and `TestReleaseNotes` pass
- [ ] 4.2 Verify that `make lint`, `make provenance` and `make wasm` pass, and that the WASM build renders `Alice -> Bob : <$archimate/actor> hi` with the actor icon, as the native build does
- [ ] 4.3 Push the branch, with the user's approval, and open a pull request that closes issue #35. Verify that the `test` and `conformance` CI jobs pass, including the committed-reference check, and that the benchmark report shows no regression of the other diagram types
- [ ] 4.4 Run `openspec validate add-archimate-diagrams --strict`. Verify that the change is reported as valid
