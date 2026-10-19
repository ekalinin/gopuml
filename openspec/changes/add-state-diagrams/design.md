# Design

## Context

See proposal.md for the motivation. The current state that shapes the approach:

- Upstream's `statediagram` package has 11 files: `StateDiagram`, `StateDiagramFactory`, and the commands `CommandCreateState`, `CommandLinkStateCommon` with `CommandLinkState` and `CommandLinkStateReverse`, `CommandCreatePackageState`, `CommandCreatePackage2` (`frame`), `CommandEndState`, `CommandAddField`, `CommandConcurrentState` and `CommandHideEmptyDescription`. The last one is already ported with the common commands (`uml.HideEmptyDescriptionSetter`).
- The entity model in `internal/abel` already has what states need: the state leaf types (`STATE`, `STATE_CHOICE`, `STATE_FORK_JOIN`, `PSEUDO_STATE`, `DEEP_HISTORY`, `STATE_CONCURRENT`, ...), the `STATE` and `CONCURRENT_STATE` group types, `SetConcurrentSeparator`, `UniqueSequence2`, `IsAutarkic`, `OverrideImage`, `EntityUtilsIsPureInnerLink12`, the entity positions of entry and exit points, and the `IsHideEmptyDescriptionForState` hook.
- `internal/svek` has the state parts of the cluster code (the state header with its description lines, the state style signatures, `drawSwinLinesState`), `FrontierCalculator` and `manageEntryExitPoint`. It panics with "not ported yet" for `EntityImageState`, `EntityImageCircleStart`, `EntityImageCircleEnd`, `EntityImageBranch`, `EntityImageSynchroBar`, `EntityImagePseudoState`, `EntityImageDeepHistory`, `EntityImageTransitionLabel`, `Cluster.drawUState` and `CucaDiagramSimplifierState`. `GroupMakerState`, `InnerStateAutonom`, `ConcurrentStates`, `EntityImageState2`, `EntityImageStateEmptyDescription`, `EntityImageStateBorder` and `EntityImageStateCommon` are not ported at all.
- `dotData` takes the children groups from the diagram, while upstream's `DotData` takes them from a `GroupHierarchy`, which `GroupMakerState` replaces for the composite state it lays out.
- The command parser runs the passes a diagram requires (`RequiredPasses`), and `command.SingleLine` has a `Passes` field, but every diagram uses one pass only. Upstream's state diagrams use three: states in all passes, transitions in the second, notes on transitions in the second and notes on states in the third. The shared note commands of `internal/abel` are always eligible for the first pass only.
- `klimt.Stereogroup` cannot return a leaf type, because `internal/klimt` is below `internal/abel`.
- `internal/core/detect.go` reports state syntax (`^state`, `[*]`, `[H]`) as "state diagrams are not supported yet" after every factory fails. `TestEmbeddedUnsupportedType` and the styling scenario "Embedded diagram of an unsupported type" use state syntax as their unsupported type.
- The engine already rejects TXT and UTXT for every diagram type other than sequence (`UnsupportedFormatError`).
- 19 of the 38 Vega state cases set `!pragma layout smetana`. The Vega inputs are never committed (`tools/corpus/fetch-vega.sh`).
- Upstream's `node` label style (`createTransitionWithIntermediateNode`) names the label node with `System.currentTimeMillis()`. The name appears in the SVG comments and path ids, so upstream's output changes from one run to the next.

## Goals / Non-Goals

**Goals:**

- Port state diagrams file by file from upstream, reusing the shared commands and the svek pipeline that class and description diagrams use.
- Keep the sequence, class, description and activity conformance results unchanged while the shared code changes (note commands, `dotData`, the order of the layout steps in `svek.TextBlock`).

**Non-Goals:**

- SCXML, TXT and UTXT output (`ScxmlStateDiagramStandard`, `CucaDiagramTxtMaker`).
- The Smetana counterparts of the state layout (`CucaDiagramSimplifierStateSmetana`, `GroupMakerStateSmetana`).
- The Kermor layout pragma, which stays "not ported yet" for every diagram type.
- The legacy activity diagram, including a recognizer of its syntax (see Risks).

## Decisions

### D1. Package layout mirrors upstream

`internal/statediagram` holds `StateDiagram`, the factory, the export and the commands of `statediagram/command`, one Go file per Java file where practical, as `internal/descdiagram` does. `Stereogroup.getLeafType` becomes an unexported function of `internal/statediagram`, its only user, because `klimt.Stereogroup` cannot import `internal/abel`.

Alternative: a function in `internal/abel` next to the leaf types. Nothing else needs it, so it stays with the commands.

### D2. StateDiagram on top of AbstractEntityDiagram

- `StateDiagram` embeds `abel.AbstractEntityDiagram`, sets the namespace separator to `.`, and passes hooks that read its `hideEmptyDescription` field for `IsHideEmptyDescriptionForState`. As in description diagrams (description design D2), the hooks are a separate value holding a pointer to the diagram, not the diagram itself. `CleanID` keeps the default, because upstream's `StateDiagram` does not override `cleanId`.
- Upstream's `endGroup` override, which also closes an open concurrent region, becomes `StateDiagram.EndGroup`. Only the state commands call it (`CommandEndState`), so no hook is needed: the shared commands of the state factory do not close groups.
- `CheckFinalError` ports upstream's check of transitions out of a concurrent region. Otherwise it reports no error, as upstream's `AbstractDiagram.checkFinalError`, which `StateDiagram` calls: unlike description diagrams, state diagrams do not pack packages or apply the single strategy.

### D3. Three parser passes

- `StateDiagram.RequiredPasses` returns the first, second and third pass, as upstream's `getRequiredPass`. The single-line state commands set `command.SingleLine.Passes` to upstream's `isEligibleFor`.
- The shared note commands get their pass from a small wrapper in `internal/command`, which overrides `EligibleFor` of a command: the state factory wraps the notes on states for the third pass and the notes on transitions for the second, as upstream's `CommandFactoryNoteOnEntity("state", ..., ParserPass.THREE)` and `CommandFactoryNoteOnLink(ParserPass.TWO)`. The class, description and activity factories do not change.
- State diagrams are the first diagrams that use more than one pass, so the pass loop of `command.Factory` gets unit tests with state input: a transition before the declaration of the composite state that holds one of its states, and a note on a state declared after the note.

Alternative: a pass parameter on the four note constructors of `internal/abel`, which is upstream's shape. It changes every note command call site of the class, description and activity factories for no behavior change, against the minimal diff.

### D4. Registration and detection

- `statediagram.Factory` goes into `engine.Factories` between `descdiagram.Factory` and `activitydiagram3.Factory`, which gives upstream's order of `PSystemBuilder` without the factories that are not ported.
- The `state` entry of `unsupportedTypes` is removed. A block with state syntax that the state factory rejects now gets the state factory's error, as upstream.

### D5. State images

`EntityImageStateCommon`, `EntityImageState`, `EntityImageState2`, `EntityImageStateEmptyDescription`, `EntityImageStateBorder`, `EntityImageCircleStart`, `EntityImageCircleEnd`, `EntityImageBranch`, `EntityImageSynchroBar`, `EntityImagePseudoState`, `EntityImageDeepHistory` and `EntityImageTransitionLabel` are ported into `internal/svek` and replace the stubs of `createEntityImageBlock`. `EntityImageStateBorder` takes the cluster of its parent state from the `Bibliotekon`, as `EntityImagePort` does. `Cluster.drawUState` replaces its stub.

### D6. Separate layout of composite states

- `CucaDiagramSimplifierState`, `GroupMakerState`, `InnerStateAutonom` and `ConcurrentStates` are ported into `internal/svek`. The simplifier replaces each autarkic composite state, deepest first, by a leaf drawn with the image of its own Graphviz run, as upstream.
- `dotData` gets a group hierarchy, upstream's `GroupHierarchy`: the diagram's children groups by default, and for `GroupMakerState` the children groups of its composite state without the concurrent regions. The class and description layouts keep the default.
- `svek.TextBlock` resolves the layout engine before the simplifier and passes it on, so the nested runs use the same `dot`, timeout and context as the main run. When `dot` is not found, the simplifier is skipped and the main run returns the not-found error image, which is the image upstream ends with. When a nested run fails or times out, the layout stops with that run's error and its error image, as for the main run.

Alternative: run the simplifier where it is today and resolve the engine again inside it. The engine would be resolved several times, and a cancelled context could be seen by one run and not by another.

### D7. Deterministic label nodes

The `node` label style is ported, but the name of the label node ends with a sequence number of the diagram instead of `System.currentTimeMillis()`. The output is then the same on every render, as the output-formats spec requires ("Deterministic output"). Upstream's output cannot be compared, so the scenario "Transition label as a node" is checked by a Go test.

### D8. Export and formats

The state export follows `internal/descdiagram/export.go`: the generic `svek.TextBlock` of the entity diagram, for SVG and PNG. TXT and UTXT are rejected by the engine with `UnsupportedFormatError` before the export is called, as for class and description diagrams, so the scenario "TXT for a state diagram" needs no new code. The WASM build needs no special case: `cmd/plantuml-wasm` turns `ErrGraphvizNotFound` into the "graph layout is not available" error for every diagram type.

### D9. Unsupported-type examples

- `TestEmbeddedUnsupportedType` embeds `(*) --> "First step"` and expects the "legacy activity diagrams are not supported yet" image, following the modified styling scenario. The `why` of `styling/embedded-unsupported-type.go.yaml` names legacy activity instead of state.
- The scenario "diagram-source: State diagram syntax" keeps its name and now states that state syntax is not reported as not supported yet. OpenSpec does not let a modified requirement drop a scenario, and keeping the name lets `diagram-source/unsupported-uml-types.go.yaml` keep its reference. `TestDetectUnsupportedUmlTypes` checks that the detector returns nothing for the state inputs, and the `why` of the case says that state syntax is not reported. `TestCreateDiagramReportsUnsupportedUmlType` (in `internal/core/detect_test.go`) uses the timing syntax instead of the state syntax.

Alternative: remove the requirement and add it again under another name without the scenario. It renames a requirement only to drop one scenario, and the case would have to change in the archive commit, when the old scenario leaves the main spec.

### D10. Conformance

- Curated cases go under `testdata/conformance/cases/state-diagram/`, one per compared scenario, plus `diagram-source/state-detected`. Their references come from `make reference`. CI renders them again on the cache miss and checks them against the committed files.
- Scenarios that are not compared with Java get `.go.yaml` cases: "Transition label as a node" (D7), "TXT for a state diagram" (upstream draws TXT), "State diagram without dot" and "State diagram in WASM" (environment).
- The cases come before the code, because `TestScenarioCoverage` reads the delta specs of active changes. Until the factory is registered, the new state cases report "not supported yet", which the runner counts as unsupported, not failed.
- Each `vega/state/` case that differs from its reference is listed in `known-differences.yaml` with the reason of the 4 `group2712` cases and `resolvedBy: add-smetana-layout`. A case that matches anyway is not listed, because the report asks for the removal of known differences that pass. Because the Vega inputs are not committed, the port of those 19 cases is checked once instead: with the pragma removed, each is rendered by the upstream jar and by this implementation with the same `dot`, and the SVGs must match apart from the encoded source.

## Risks / Trade-offs

- **[Legacy activity before state]** Upstream tries the legacy activity factory before the state factory. A block without `(*)` that both accept, such as `A --> ==sync==` and `==sync== --> C`, or `A --> B` and `B --> A[H]`, is a legacy activity diagram upstream and becomes a state diagram here. Before this change it was a syntax error. → Upstream's boundary depends on every line of the block (adding `note right of A` makes the same block a state diagram), so only a port of the legacy activity grammar can follow it. Legacy activity diagrams almost always use `(*)`, which the state factory rejects, so they keep their "not supported yet" error. The legacy activity cases of the conformance suite show any change.
- **[First multi-pass diagram]** The pass loop of `command.Factory` was ported but never run with more than one pass. → Unit tests of D3, and the conformance cases with transitions before declarations.
- **[Shared code changes]** The note wrapper, the group hierarchy of `dotData` and the new order of the steps in `svek.TextBlock` touch the class and description paths. → They change no behavior for those types. The class and description unit tests and conformance groups must stay green, and CI requires every group.
- **[Smetana cases checked once]** The 19 `vega/state/` cases are compared without their pragma only during this change, not in CI. → The curated cases cover composite states and both kinds of regions in CI. `add-smetana-layout` brings the Vega cases back into the comparison.
- **[Nested layout cost]** Each autarkic composite state runs `dot` once more. → As upstream. The runs share one resolved engine and the context's timeout and cancellation.

## Migration Plan

Nothing to migrate. Inputs that failed with "state diagrams are not supported yet" now render. Rollback means reverting the change.
