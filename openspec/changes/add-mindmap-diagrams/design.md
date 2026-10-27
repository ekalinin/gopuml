# Design

## Context

See proposal.md for the motivation. The current state that shapes the approach:

- Upstream's `mindmap` package has 18 classes and no subpackage: the diagram and its factory, five commands (`CommandMindMapOrgmode`, `CommandMindMapOrgmodeMultiline`, `CommandMindMapPlus`, `CommandMindMapRoot`, `CommandMindMapDirection`), the model (`Idea`, `IdeaShape`, `Branch`, `MindMap`) and the layout (`Finger`, `FingerImpl`, `Tetris`, `StripeFrontier`, `Stripe`, `SymetricalTee`, `SymetricalTeePositioned`). It imports nothing from `wbs`. Outside the package it uses `FtileBoxOld.createMindMap` of `activitydiagram3`, `CommandRankDir`, `CommonCommands.addCommonCommands1`, the style queries `StyleQueries.MINDMAPDIAG_*`, `StyleBuilder.getMergedStyleSpecial`, `SkinParamColors`, `TextBlockMemoized` and `TextBlockUtils`. The `TeaVM.a()` branches only hold Java `assert` statements, which are off without `-ea`.
- `MindMapDiagram` is a `TitledDiagram` that draws through `getTextBlock`, like `PacketDiagram` and `GanttDiagram`, with the default margins. Its constructor sets the rank direction of its skin parameters to left to right. It keeps a list of `MindMap`s, drawn one below the other, and its text block is 10 wider than the widest one.
- The Go side already has `uml.Base`, `uml.NewFactory`, `uml.AddCommonCommands1`, `uml.Multilines2` with `RemoveStartingQuote` and `TrimBoth`, `abel.CommandRankDir`, `klimt.DirectionLazzyValueOf`, `Display.EndingStereotype` and `RemoveEndingStereotype`, `klimt.NewSkinParamColors`, `klimt.NewSheetBlock1Style`, `klimt.NewSheetBlock2`, `klimt.TextBlockMemoized`, `klimt.TextBlockUtilsWithMargin4`, `StyleSignatureBasic.AddStereotypeOf`, `AddLevel` and `AddStar`, `StyleBuilder.MergedStyleSpecial` with the ancestor rank of 1.2026.9, the SNames `mindmapDiagram`, `node`, `rootNode`, `leafNode`, `boxless` and `arrow`, the `mindmapDiagram` sections of the bundled skins, and `uml/chrome.go`, which maps `core.MindMap` to the `mindmapDiagram` style. `MergedStyleSpecial` has no caller outside `internal/klimt` yet.
- `internal/activitydiagram3` has `FtileBox`, `BoxStyle` and `LinkRendering`, but not `FtileBoxOld`. `FtileBoxOld` differs from `FtileBox`: its sheet block takes the padding of the style instead of the skin parameters, its stencil has no padding, its sheet uses the alignment of the style without `getDefaultTextAlignment`, and it draws the text at `MinimumWidth` or more.
- The command factory turns a panic of a command into upstream's "You should send a mail to plantuml@gmail.com with this log" error, and `uml.Throw` gives the Java exception names. `getSmartLevel` throws `UnsupportedOperationException` for markers it cannot read.
- `core.CheckFamily` reports every family that is not in `supportedTypes` (`Sequence`, `Class`, `Activity`, `Packet`, `Gantt`) as not supported yet. `core.FindStartTypes` already maps `@startmindmap` to `core.MindMap`. Several tests use `@startmindmap` as their example of a family that is not supported yet: `TestCheckFamily` (`internal/core`, the case `diagram-source/unsupported-family`), `TestNotSupportedYet` (`internal/errdiagram`) and `TestInspectNamesAndStatus` (root package).
- The engine rejects TXT and UTXT for every type other than sequence and gantt. Upstream draws a mindmap as TXT with the labels at their positions.
- `TestScenarioCoverage` reads the delta specs of active changes, so every scenario of this change needs a case once the specs are committed. `TestDiagramReferences` of `internal/errdiagram` draws the error image of every curated case whose reference is an error.

## Goals / Non-Goals

**Goals:**

- Port the whole `mindmap` package file by file, with the SVG byte-identical to upstream's apart from the encoded source.
- Keep upstream's quirks: the scenarios of "Indentation" pin the ones users can see.
- Keep the results of every other diagram type unchanged. The shared code that changes is the list of supported families, the factory list and a new file in `internal/activitydiagram3`.

**Non-Goals:**

- WBS diagrams and `FtileBoxOld.createWbs`, which only WBS uses. The WBS change adds them.
- TXT and UTXT output.
- Porting the `explainArg` and `explainNow` methods of the commands, which only upstream's `@Explain` tooling calls, and the `TeaVM.a()` asserts.

## Decisions

### D1. Package layout

`internal/mindmap` holds the whole package, with upstream's grouping: `diagram.go` (`MindMapDiagram`), `factory.go`, `commands.go` (the five commands), `idea.go` (`Idea`, `IdeaShape`), `branch.go`, `mindmap.go` (`MindMap`), `finger.go` (`Finger`, `FingerImpl`), `tetris.go` (`Tetris`, `Stripe`, `StripeFrontier`, `SymetricalTee`, `SymetricalTeePositioned`) and `export.go`. Each file has the `Provenance` lines of the Java files it ports. `IdeaShape` is exported (`mindmap.IdeaShape`), so that the WBS change can import it as upstream's `wbs` does.

`Finger` is upstream's interface, implemented only by `FingerImpl`. Go keeps the struct alone, because `Branch.getX12` casts the interface to `FingerImpl` anyway.

Alternative: one package for mindmap and WBS diagrams. They are two upstream packages with their own factories, and WBS is not part of this change.

### D2. FtileBoxOld in activitydiagram3

`FtileBoxOld` goes into `internal/activitydiagram3/ftile_vertical_ftileboxold.go`, next to `FtileBox`, with `FtileBoxOldCreateMindMap(style, skinParam, label)` as upstream's `createMindMap`. It reuses `BoxStyle`, `LinkRendering`, `RainbowFromStyle`, `SkinParamColors` and the sheet blocks, with upstream's differences from `FtileBox` (Context). `internal/mindmap` imports `internal/activitydiagram3`, as upstream's `mindmap` imports `activitydiagram3.ftile.vertical`. There is no import cycle.

Alternative: a copy of `FtileBoxOld` in `internal/mindmap`. WBS needs it too (`createWbs`), and it would leave two ports of one Java file.

### D3. MindMapDiagram on top of uml.Base

- `MindMapDiagram` embeds `uml.Base` with the type `core.MindMap`, as `PacketDiagram` does, so the SVG has `data-diagram-type="MINDMAP"` and the annotations, skins and styles come from the shared code. `Description` returns `MindMap`. The constructor sets the rank direction of the skin parameters to left to right and adds the first `MindMap`, as upstream's constructor does.
- `ExportDiagram` renders with `uml.ExportTitled` and upstream's `getTextBlock`, as `internal/packetdiag/export.go` does.
- `MindMap` embeds `klimt.TextBlockMemoized` for upstream's `calculateDimensionSlow`. The finger of a branch is built once, on the first call to the dimension or the drawing, and the `Tetris` of a finger once, with the string bounder of its first call, as upstream caches them.

### D4. Parsing

- The factory adds the commands in upstream's order: `uml.AddCommonCommands1`, `abel.CommandRankDir`, then the multiline org-mode command, the org-mode command, the root command, the `+` and `-` command and the direction command. The direction command matches any first token, as upstream's `FirstTokens.ANYTHING`, and comes last, so that `top to bottom direction` is read by `CommandRankDir`.
- The commands are `command.SingleLine` and `uml.Multilines2` with upstream's patterns and first tokens (`*` and `#`, `+` and `-`, `0`, and none for the direction command). The org-mode and `+` commands set `NoTrim`, as upstream builds them with `doTrim` set to `false`, so that the spaces and tabs before the markers reach `getSmartLevel`. The multiline command has upstream's end pattern `^(.*);\s*(\<\<(.+)\>\>)?$`.
- `getSmartLevel` keeps upstream's rules, including the first marker it saw, with Java's `trim` (characters up to the space) instead of `strings.TrimSpace`. Markers that it cannot read panic with `uml.Throw(uml.UnsupportedOperationException, "type=<...>[...]")`, which the command factory turns into upstream's crash message.
- Colors are read with the color set of the skin parameters. An unknown color fails the command with `No such color`, as the other diagram types do.
- `MindMap.addIdeaInternal` keeps upstream's rules: the first node becomes the root, the first depth after the root sets the multiplier, `Bad indentation` and `error42L`, and `isFull`, which starts a new `MindMap` for a node at depth zero. The `NoStyleAvailableException` branch is not ported: upstream throws it when no skin file can be loaded, and the Go skins are bundled.

### D5. Styles

- The style queries are built inline with `klimt.StyleSignatureBasicOf(SNameRoot, SNameElement, SNameMindmapDiagram, SNameNode, ...)`, `AddStereotypeOf` and `AddLevel`, the Go form of upstream's `StyleQueries.MINDMAPDIAG_*` with `withStereotype` and `addLevel`.
- `Idea.getStyle` keeps upstream's cascade: the style of the node at rank 0, then the starred style (`AddStar`) of each ancestor's query at the node's level, with a rank that drops by one per ancestor, merged with `MergeOverwriteExistingValue`. Each idea keeps the style builder of the time it was created, as upstream does, so a `<style>` after a node does not change it.

### D6. Layout as upstream

- `Tetris`, `StripeFrontier` and `SymetricalTeePositioned` keep Java's order of operations on `float64`. `-Double.MAX_VALUE` and `Double.MAX_VALUE` are `-math.MaxFloat64` and `math.MaxFloat64`.
- Upstream's `TreeSet<Stripe>` is ordered by `(int) Math.signum(x1 - other.x1)`. Go keeps the stripes in a slice sorted by their start, which iterates in the same order, because no two stripes of a frontier share a start.
- `Stripe` and `StripeFrontier.addSegment` panic with `uml.Throw(uml.IllegalArgumentException, "")` for an empty segment, and `collisionning` and `contains` with `UnsupportedOperationException`, as upstream throws. `System.err` output is not ported.
- `FingerImpl` keeps upstream's link path (a line, a cubic curve and a line, with the offsets 10 and 25 from left to right and 3 and 10 from top to bottom), the margins of the phalanx (the style's margin for a box, 3 and 1 for text without a box) and `getX2`, which adds 30 (left to right) or 5 (top to bottom) to the margin.

### D7. Registration, formats and WASM

- `mindmap.Factory` (`uml.NewFactory(core.MindMap, ...)`) goes into `engine.Factories` after `activitydiagram3.Factory` and before `packetdiag.Factory`, as `MindMapDiagramFactory` comes before `PacketDiagramFactory` and `GanttDiagramFactory` in upstream's `PSystemBuilder`.
- `core.MindMap` joins `supportedTypes`. `TestCheckFamily` gets a `@startmindmap` check, and its unsupported family, `TestNotSupportedYet` and `TestInspectNamesAndStatus` use `@startwbs` instead of `@startmindmap`. The `why` of `diagram-source/unsupported-family.go.yaml` names `@startwbs`.
- The TXT check of the engine does not change, so TXT and UTXT stay rejected for mindmap diagrams.
- The WASM build needs no special case, because mindmap diagrams do not use Graphviz.

### D8. Documentation

- The README and `site/index.html` get the row `Mindmap | yes | yes | no` in the table of supported diagrams. `mindmap` is removed from the sentence that lists the types that are not supported yet, in both. The Graphviz note of the page and the last sentence of the README's Graphviz section name mindmap diagrams among the diagrams that do not need `dot`. The README's conformance counts follow the report.
- The playground note of the page and the README's sentence about the playground do not change: they follow the WASM module of the release on the page.

### D9. Conformance

- Curated cases go under `testdata/conformance/cases/mindmap-diagram/` with `family: mindmap` and `type: mindmap`, one per compared scenario, plus `diagram-source/mindmap-detected` and an embedded case (`family: uml`, `type: sequence`). The PNG scenario uses `formats: [svg, png]`, the others `[svg]`. Their references come from `make reference`, and CI renders them again and checks them against the committed files.
- The embedded case is an `@startuml` case, so the golden dumps of `internal/sequencediagram`, `internal/classdiagram` and `internal/activitydiagram3` are regenerated, as for the gantt change.
- Scenarios that are not compared with Java get `.go.yaml` cases: "Mindmap diagram without Graphviz" (environment), "TXT for a mindmap diagram" (upstream draws TXT, this change rejects it) and "Mindmap diagram in WASM" (environment).
- The cases come before the code. Until the factory is registered, the mindmap cases report "not supported yet", which the runner counts as unsupported, `TestDiagramReferences` skips the curated mindmap error cases, and the embedded case fails in `make conformance`.
- The 4 Vega mindmap cases are compared once the factory is registered. A case that still differs at the end is listed in `known-differences.yaml` with its reason.

### D10. Port order

The layers are ported bottom up, each with unit tests: `FtileBoxOld`, the packing (`Stripe`, `StripeFrontier`, `SymetricalTee`, `Tetris`), the model (`IdeaShape`, `Idea`, `Branch`, `MindMap.addIdeaInternal`) and the diagram, the commands and the factory, then the drawing (`FingerImpl`, `MindMap`, the export).

## Risks / Trade-offs

- **[First caller of the style cascade]** `MergedStyleSpecial` and the ancestor rank of starred styles have no caller outside `internal/klimt` yet. → The scenarios "Style of the descendants" and "Nearest ancestor" and the Vega case `mindmap-deep-ancestor-cascade-with-stereotype` check them against upstream. A difference there is fixed in `internal/klimt`, with the style golden tests of `internal/klimt` kept unchanged.
- **[FtileBoxOld versus FtileBox]** A port that reuses `FtileBox` would draw the text with the padding of the skin parameters. → D2 ports `FtileBoxOld` as its own type, and the scenarios "Minimum width and alignment" and "Title and style" and the Vega case `mindmap_02_horizontal_alignment` check its padding, minimum width and alignment.
- **[Floating point]** The packing compares and adds `double` values. → D6 keeps Java's order of operations, and the Vega cases with several levels check it.
- **[Tests that use mindmap as the unsupported example]** They would fail once mindmap is registered. → D7 moves them to `@startwbs`, which stays unsupported until the WBS change, as the gantt change did for its check in `internal/core/builder_test.go`.

## Migration Plan

Nothing to migrate. Inputs that failed with "mindmap diagrams are not supported yet" now render. Rollback means reverting the change.
