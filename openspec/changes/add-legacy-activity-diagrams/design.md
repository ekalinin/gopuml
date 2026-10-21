# Design

## Context

See proposal.md for the motivation. The current state that shapes the approach:

- Upstream's `activitydiagram` package has `ActivityDiagram`, `ConditionalContext`, `ActivityDiagramFactory` and the commands `CommandLinkActivity`, `CommandLinkLongActivity` (a multi-line activity), `UBrexCommandIf`, `UBrexCommandElse`, `UBrexCommandEndif`, `UBrexCommandPartition`, `UBrexCommandEndPartition` and `CommandInnerConcurrent`, which upstream does not register. The factory also uses `CommandFactoryNoteActivity` (notes on the last activity), the notes on links, the common commands, and the UBrex variants `UBrexCommandFootboxIgnored`, `UBrexCommandRankDir` and `UBrexCommandHideShow2`.
- Upstream's `ActivityDiagram` extends `CucaDiagram` directly, not `AbstractEntityDiagram`: it has no namespace separator, its own DOT settings and description ("(N activities)"), and it overrides `updateLasts` to remember the last activity and the last branch.
- The UBrex engine (`com.plantuml.ubrex`) is ported, unexported, in `internal/klimt` for the creole commands. Its builder (`com.plantuml.ubrex.builder`: `UBrexConcat`, `UBrexLeaf`, `UBrexNamed`, `UBrexOptional`, `UBrexOr`, ...), upstream's `UBrexSingleLineCommand2`, and the reading of UBrex captures through `RegexResult` (`get` reads the values of a key, `getLazzy` the values of the first key with a prefix) are not. `color.UBrexColorParser` translates upstream's UBrex color pattern into a regex leaf; its only user is the partition command.
- In `internal/svek`, `createEntityImageBlock` panics with "not ported yet" for `EntityImageActivity`, and `svek.TextBlock` for `CucaDiagramSimplifierActivity`. `GroupMakerActivity` and `InnerActivity` are not ported. The octagon shape, start and end circles, branches, synchronization bars, notes and clusters are.
- `internal/core/detect.go` reports `(*)` syntax as "legacy activity diagrams are not supported yet". The diagram-source scenario "Legacy activity syntax", the styling scenario "Embedded diagram of an unsupported type", `TestDetectUnsupportedUmlTypes`, `TestEmbeddedUnsupportedType` and their `.go.yaml` cases use that syntax as their unsupported type.
- The activity golden tests (`internal/activitydiagram3`) treat the conformance cases under `cases/activity-diagram/` as activity diagrams of the new syntax.
- The corpus has one legacy activity case, `vega/nonreg/simple/A0004`, reported as not supported yet.
- Upstream logs an `IllegalStateException` for the links of an inner activity (`--> {` ... `}`) and draws the activities without those links.

## Goals / Non-Goals

**Goals:**

- Port legacy activity diagrams file by file from upstream, keeping the UBrex commands in UBrex rather than translating them.
- Keep the results of the other diagram types unchanged, now that the legacy parser runs before the description and state parsers.

**Non-Goals:**

- TXT, UTXT and the other formats upstream draws through `CucaDiagramTxtMaker`.
- `CommandInnerConcurrent`, which upstream does not register.
- The parts of the UBrex builder that the ported commands do not use.

## Decisions

### D1. Package layout mirrors upstream

`internal/activitydiagram` holds `ActivityDiagram`, `ConditionalContext`, the factory, the export and the commands of `activitydiagram/command`, one Go file per Java file where practical. `CommandFactoryNoteActivity` (upstream's `command/note`) goes there too, because it is typed on `ActivityDiagram`, which `internal/abel` cannot import. The package name follows upstream's (`activitydiagram` next to `activitydiagram3`).

### D2. ActivityDiagram on CucaDiagram

`ActivityDiagram` embeds `abel.CucaDiagram` with hooks that hold a pointer to the diagram (as state diagrams, design D2 of `add-state-diagrams`): `UpdateLasts` records the last activity and the last branch, and the other hooks keep their defaults. It has no namespace separator, and its own `DotStrings` and `Description`.

### D3. UBrex commands

- The UBrex engine stays in `internal/klimt`, which exports what commands need: building an expression from builder parts, matching a line, and the matcher's `exactMatch` and capture lookups. Upstream's builder is ported there for the parts the commands use (`UBrexConcat`, `UBrexLeaf` with `end`, `spaceZeroOrMore` and `spaceOneOrMore`, `UBrexNamed`, `UBrexOptional`, `UBrexOr`).
- `regex.Result` can be backed by UBrex captures, through an interface of two lookups defined in `internal/command/regex` and implemented by the klimt matcher: `Get` returns the values of a key, and `Lazy` those of the first key with a prefix, as upstream's `RegexResult(UMatcher)`. `internal/command/regex` does not import `internal/klimt`, which imports it.
- `uml.UBrexSingle` is upstream's `UBrexSingleLineCommand2`: an exact match of the trimmed line, executed with a capture-backed `regex.Result`, with the `Passes` and `Forbidden` options of `command.SingleLine`.
- The UBrex variants of shared commands are ported next to their regex versions in `internal/abel`: `UBrexCommandFootboxIgnored`, `UBrexCommandRankDir` and `UBrexCommandHideShow2`.
- `color.UBrexColorParser` returns upstream's UBrex part instead of the regex leaf, since the partition command, its only user, is now a UBrex command. It moves into `internal/klimt` (`klimt.UBrexColorParser`), because a UBrex part is a type of `internal/klimt`, which imports `internal/klimt/color`, as `klimt.ColorsColorOf` and the other color helpers that need style types did.

Alternatives:

- Translating the UBrex patterns into regular expressions, as the color parser did. UBrex matches like a PEG, without backtracking, so a translated pattern can accept other lines, and the `if` pattern is long enough that the difference would be hard to see.
- Moving the engine into its own package. It would be cleaner, but the move changes every creole file that uses it, for no change of behavior.

### D4. Registration and detection

- `activitydiagram.Factory` goes into `engine.Factories` between `classdiagram.Factory` and `descdiagram.Factory`, which gives upstream's order of `PSystemBuilder`.
- The `legacy activity` entry of `unsupportedTypes` is removed. A block with `(*)` that the legacy parser rejects gets the parser's error, as upstream.
- As in the state change (design D9 of `add-state-diagrams`), the scenario "diagram-source: Legacy activity syntax" keeps its name and now states that the syntax is not reported as not supported yet, so its `.go.yaml` case keeps its reference. `TestDetectUnsupportedUmlTypes` checks that the detector returns nothing for the legacy inputs. `TestEmbeddedUnsupportedType` embeds `robust "Web" as WB` and expects the "timing diagrams are not supported yet" image.

### D5. Activity images and inner activities

- `EntityImageActivity` is ported into `internal/svek` and replaces its stub: a rounded box, or the octagon of its node when `activityShape` is `octagon`.
- `CucaDiagramSimplifierActivity`, `GroupMakerActivity` and `InnerActivity` are ported for inner activities, and the simplifier replaces its stub in `svek.TextBlock`, after the layout engine is resolved and skipped without `dot`, as the state simplifier (design D6 of `add-state-diagrams`). `GroupMakerActivity` uses the group hierarchy of `dotData` with its own `InnerGroupHierarchy`.
- Inner activities have no spec scenario: upstream logs an exception for their links and draws them without these links. `graphvizImageBuilder.addLink` already skips such links in the same way.

### D6. Export and formats

The export follows `internal/statediagram/export.go`. TXT and UTXT are rejected by the engine before the export, as for every type other than sequence. The WASM build needs no special case.

### D7. Conformance

- Curated cases go under `testdata/conformance/cases/legacy-activity-diagram/`, not under `cases/activity-diagram/`, whose cases the activity golden tests expect to be in the new syntax. `diagram-source/legacy-activity-before-state` covers the new detection scenario.
- Scenarios that are not compared with Java get `.go.yaml` cases: "TXT for a legacy activity diagram", "Legacy activity diagram without dot" and "Legacy activity diagram in WASM".
- The cases come before the code, because `TestScenarioCoverage` reads the delta specs of active changes. Until the factory is registered, the new cases are reported as not supported yet, except `diagram-source/legacy-activity-before-state`, which is drawn as a state diagram and fails until then.
- `vega/nonreg/simple/A0004` is compared once the factory is registered.

## Risks / Trade-offs

- **[Detection order]** The legacy parser now runs before the description and state parsers, so a block that it accepts becomes a legacy activity diagram. → That is upstream's order, so the references expect it. CI requires every group, so a description, state or activity case that changes type fails the run.
- **[Shared UBrex code]** Exporting the engine and adding capture-backed results touches the creole commands and `regex.Result`. → The engine does not change; its golden test (`TestGoldenUbrex`) and the creole tests must stay green.
- **[Color parser]** `color.UBrexColorParser` changes form and package. → Its only user is the partition command; its unit test moves with it, to the UBrex form.
- **[Inner activities]** They are ported without a scenario, because upstream's own output for them is broken. → The code follows upstream; a later change can add a scenario if upstream fixes them.

## Migration Plan

Nothing to migrate. Inputs that failed with "legacy activity diagrams are not supported yet" now render. Blocks without `(*)` that both the legacy activity and the state syntax accept become legacy activity diagrams, as upstream. Rollback means reverting the change.
