# Design

## Context

See proposal.md for the motivation. The current state that shapes the approach:

- Upstream's `wbs` package has 14 classes and no subpackage: the diagram and its factory, five commands (`CommandWBSItemNew` and `CommandWBSItemOld`, each in two modes, `CommandWBSItemMultilineNew`, `CommandWBSItemMultilineOld` and `CommandWBSLink`), the model (`WElement`, `WBSLink`) and the drawing (`ITF`, `ITFLeaf`, `ITFComposed`, `Fork`, `WBSTextBlock`). Outside the package it uses `mindmap.IdeaShape`, `FtileBoxOld.createWbs` and `setMinimumWidth`, `Direction.getWBSDirection`, `ExtremityArrow`, `XRectangle2D.intersect`, `ColorParser.simpleColor`, `StereotypePattern.optional`, the style queries `StyleQueries.WBSDIAG_*`, `StyleBuilder.getMergedStyleSpecial`, `SkinParamColors`, `TextBlockMemoized`, `TextBlockUtils` and, for the text output, `ATable` and `InfinitePlan` of `asciiverse`. The `TeaVM.a()` branches only hold Java `assert` statements, which are off without `-ea`.
- `WBSDiagram` is a `TitledDiagram` that is its own text block: `getTextBlock` returns the diagram, `drawU` draws a new `Fork` of the root at (10, 10) and the dimension is the one of a new `Fork`, 20 larger. Upstream builds a new `Fork`, with new `ITF`s, on each call. Its `exportTxt` draws the root and its children with `asciiverse` in ASCII, for TXT and UTXT alike, without the annotations.
- The nodes record their position while they are drawn, when the graphic is an `AbstractCommonUGraphic` (`ITFLeaf.drawU`, `ITFComposed.drawU`). `drawMe` draws the links after the tree, with the translation of the graphic undone, from these positions.
- The Go side already has `uml.Base`, `uml.NewFactory`, `uml.AddCommonCommands1`, `uml.Multilines2`, `command.SingleLine` with `NoTrim`, `regex.Result.Lazy` and `Get`, `mindmap.IdeaShape` with `IdeaShapePseudo` and `IdeaShapeFromDesc`, `klimt.DisplayWithNewlines`, `klimt.StereotypeBuild`, `klimt.StereotypePatternOptional`, `color.SimpleColor`, `klimt.NewSkinParamColors`, `klimt.TextBlockMemoized` with `InvalidateDimensionCache`, `klimt.TextBlockUtilsEmptyTextBlock`, `klimt.TextBlockUtilsWithMargin4`, `StyleBuilder.MergedStyleSpecial`, `Style.ApplyStrokeAndLineColor`, `extremity.NewArrow`, `geom.XRectangle2D.Intersect`, `geom.XLine2DLine`, `uml.Base.AddWarning` with the warning banner of `uml/chrome.go`, `asciiverse.ATable` with its paddings and `asciiverse.InfinitePlan` with `Move` and `ExportTxt`, the SName `wbsDiagram`, the `wbsDiagram` sections of the bundled skins, and `uml/chrome.go`, which maps `core.WBS` to the `wbsDiagram` style. `klimt.CommonUGraphic` is the Go form of `AbstractCommonUGraphic`: only the graphics that embed `AbstractCommonUGraphic` implement it, and `UGraphicNo`, which the limit finder embeds, does not.
- `internal/activitydiagram3/ftile_vertical_ftileboxold.go` ports `FtileBoxOld` with `createMindMap` only. Its comment says that `createWbs` and `setMinimumWidth` are left for WBS diagrams. `klimt.Direction` says that upstream's `getWBSDirection` belongs to the WBS diagram.
- `core.CheckFamily` reports every family that is not in `supportedTypes` as not supported yet. `core.FindStartTypes` already maps `@startwbs` to `core.WBS`. Several tests use `@startwbs` as their example of a family that is not supported yet: `TestCheckFamily` (`internal/core`, the case `diagram-source/unsupported-family`), a check of `internal/core/builder_test.go`, `TestNotSupportedYet` (`internal/errdiagram`) and `TestInspectNamesAndStatus` (root package).
- The engine rejects TXT and UTXT for every type other than sequence and gantt. `gantt.exportTxt` is the model of a text output on `asciiverse`: a panic is returned as an error and nothing is written.
- `uml.ExportTitled` returns a panic while drawing as an error, without output.
- `TestScenarioCoverage` reads the delta specs of active changes, so every scenario of this change needs a case once the specs are committed. `TestDiagramReferences` of `internal/errdiagram` draws the error image of every curated case whose reference is an error.

## Goals / Non-Goals

**Goals:**

- Port the whole `wbs` package file by file, with the SVG byte-identical to upstream's apart from the encoded source, and the TXT and UTXT identical.
- Keep upstream's quirks: the scenarios of "Sides", "Tree structure" and "Links" pin the ones users can see.
- Keep the results of every other diagram type unchanged. The shared code that changes is the list of supported families, the factory list, the check of the text formats in the engine and `FtileBoxOld`.

**Non-Goals:**

- Porting the `explainArg` and `explainNow` methods of the commands, which only upstream's `@Explain` tooling calls, the `TeaVM.a()` asserts and the `NoStyleAvailableException` branches, which upstream reaches only when no skin file can be loaded (the Go skins are bundled).

## Decisions

### D1. Package layout

`internal/wbs` holds the whole package, with upstream's grouping: `diagram.go` (`WBSDiagram`: the tree, the aliases and the links), `factory.go`, `commands.go` (the five commands, `getWBSDirection` and `orderWarning`), `welement.go`, `wbslink.go`, `itf.go` (`ITF`, `ITFLeaf`, `ITFComposed`), `fork.go`, `textblock.go` (`WBSTextBlock`), `export.go` and `txt.go`. Each file has the `Provenance` lines of the Java files it ports. `getWBSDirection` goes into `commands.go`, with the `Provenance` line of `utils/Direction.java`, because `klimt.Direction` leaves it to the WBS diagram. The node shape is `mindmap.IdeaShape`, as upstream's `wbs` imports `mindmap.IdeaShape`.

Alternative: one package for mindmap and WBS diagrams. They are two upstream packages with their own factories, and they share only `IdeaShape` and `FtileBoxOld`.

### D2. FtileBoxOld for WBS

`FtileBoxOldCreateWbs(style, skinParam, label)` is upstream's `createWbs`. It returns `*FtileBoxOld`, because `ITFLeaf` and `ITFComposed` call `SetMinimumWidth` on it. `SetMinimumWidth` keeps the larger of the two widths and empties the geometry cache, as upstream's `setMinimumWidth` does with `invalidateGeometryCache`. `FtileBoxOldCreateMindMap` does not change. The comment of the type stops saying that these methods are not ported.

### D3. WBSDiagram on top of uml.Base

- `WBSDiagram` embeds `uml.Base` with the type `core.WBS`, as `MindMapDiagram` does, so the SVG has `data-diagram-type="WBS"` and the annotations, skins, styles and warnings come from the shared code. `Description` returns `Work Breakdown Structure`.
- `TextBlock` returns a text block over the diagram, upstream's `getTextBlock` returning `this`: its `DrawU` is upstream's `drawU` and `drawMe`, its dimension is the dimension of the `Fork` plus 20. Like upstream, each call builds a new `Fork` of the root, so the auto width and the caches of the text blocks belong to one call.
- `ExportDiagram` renders TXT and UTXT with `exportTxt` (D7), and the other formats with `uml.ExportTitled`, as `internal/mindmap/export.go` does.
- A diagram without nodes builds its `Fork` with no root. Upstream throws a `NullPointerException` there. Go panics with `uml.Throw(uml.NullPointerException, ...)`, which `uml.ExportTitled` and `exportTxt` return as an error, without output.

### D4. Parsing

- The factory adds the commands in upstream's order: `uml.AddCommonCommands1`, `CommandWBSItemNew` in mode 1 (the quoted label with `as`) and mode 0, `CommandWBSItemMultilineNew`, `CommandWBSItemOld` in mode 1 and mode 0, `CommandWBSItemMultilineOld`, then `CommandWBSLink`. The link command matches any first token, as upstream's `FirstTokens.ANYTHING`, and comes last.
- The item commands are `command.SingleLine` with upstream's patterns and the first tokens `*`, `+` and `-`. They set `NoTrim`, as upstream builds them with `doTrim` set to `false`, so that the spaces and tabs before the markers reach `getSmartLevel`. The link command trims, as upstream's default. The multiline commands are `uml.Multilines2` with `RemoveStartingQuote`, `TrimBoth` and upstream's end pattern `^(.*);\s*(\<\<(.+)\>\>)?$`, and match their first line trimmed, as upstream does.
- The two orders of `_` and `<` or `>`, and of the color and the alias, are upstream's `RegexOr`s with the groups `SHAPE_1` and `SHAPE_2`, `DIRECTION_1` and `DIRECTION_2`, `BACKCOLOR_1` and `BACKCOLOR_2`, and `CODE_1` and `CODE_2`, read with `Result.Lazy`. The label is read with `Result.Get`, so that a missing label (upstream's `null`) differs from an empty one: a missing label with `_` makes a pseudo node, and without `_` fails with `Missing label for WBS node.`.
- The old commands add `orderWarning` with `AddWarning` before they add the node, as upstream does, so the warning stays when the node fails.
- `addIdea` removes the stereotype with upstream's `patternStereotype` (`^\s*(.*?)\s*(\<\<\s*(.*)\s*\>\>)\s*$`) and passes its group 2 to `klimt.StereotypeBuild`, and builds the label with `klimt.DisplayWithNewlines` and the pragma of the diagram. This is not `Display.EndingStereotype` of the mindmap diagram.
- Colors are read with the color set of the skin parameters. An unknown color fails the command with `No such color`, as the other diagram types do.
- The link command reads its color with `color.SimpleColor(color.TypeLine)` and its stereotype with `klimt.StereotypePatternOptional("STEREOTYPE")`, and does not use the label, as upstream does.

### D5. Model

- `getSmartLevel` keeps upstream's WBS rules, which differ from the mindmap's: the first node is always at level zero and records its markers, tabs become spaces, and the other rules use Java's `trim` (characters up to the space) instead of `strings.TrimSpace`. Markers that it cannot read panic with `uml.Throw(uml.UnsupportedOperationException, "type=<...>[...]")`, which the command factory turns into upstream's crash message.
- `add` keeps upstream's rules: a child of the last node at its level plus one, a child of the ancestor `getParentOfLast` at a level up to the last one, `Bad tree structure` otherwise, and `WBS diagram only accepts one root first declared.` for a second node at level zero. Aliases go into a map from the alias to the node.
- `WElement.createElement` keeps upstream's quirk: a node on the left side at level one is put first in the right children of its parent and also added to its left children. The `Fork` of the root reads the right children only, so such a node is drawn first among the children of the root.
- The style queries are built inline with `klimt.StyleSignatureBasicOf(SNameRoot, SNameElement, SNameWbsDiagram, SNameNode, ...)`, `AddStereotypeOf` and `AddLevel`, the Go form of upstream's `StyleQueries.WBSDIAG_*` with `withStereotype` and `addLevel`. `WElement.getStyle` keeps the cascade of the starred styles of the ancestors with the ancestor rank, as the mindmap's `Idea.getStyle` does. Each element keeps the style builder of the time it was created.
- `link` returns `No such node <alias>` for an unknown alias, and takes the line color of `WBSDIAG_ARROW` with the stereotype of the link when the link has no color.

### D6. Drawing as upstream

- `ITF` is a Go interface with upstream's methods (`T1`, `T2`, `F1`, `F2`, `MainBoxWidth` and `SetForcedMinWidth`), implemented by `ITFLeaf` and `ITFComposed`. `Fork` and `ITFComposed` embed `WBSTextBlock`, which embeds `klimt.TextBlockMemoized`. Their auto width runs once per instance, from the first call to the dimension or the drawing, and empties the cache of the dimension, as upstream's `applyAutoWidth` does.
- `ITFLeaf` and `ITFComposed` record the position of their node when the graphic implements `klimt.CommonUGraphic`, from `Common().Translate()`, the Go form of upstream's `instanceof AbstractCommonUGraphic`. The limit finder does not implement it, so, as upstream, the positions come from the drawing of the output.
- `drawMe` draws the `Fork`, then, when the graphic implements `klimt.CommonUGraphic`, undoes its translation and draws the links.
- `WBSLink.drawU` keeps upstream's geometry: the line between the centers of the two rectangles, cut at the border of each with `XRectangle2D.Intersect`, drawn as a `ULine` from the first cut (upstream's `XLine2D.drawU`), with `extremity.NewArrow` at the second cut and the angle of the line. A line that does not cross a rectangle is upstream's `null`, on which upstream throws a `NullPointerException`: Go panics with `uml.Throw(uml.NullPointerException, ...)`.
- `WBSTextBlock.drawLine` applies the stroke and the line color of `WBSDIAG_ARROW` at the level of the node and with its stereotype, and draws from the smaller x to the larger, as upstream does. `Fork` keeps upstream's offsets (20 between the children of the root, 40 between the root and its children) and `ITFComposed` its offset of 10 from the vertical line and the bottom margin of its style (0 for a pseudo node).

### D7. Text output

- `exportTxt` draws on an `asciiverse.InfinitePlan` in ASCII, for TXT and UTXT alike, as upstream's `exportTxt` with `FileFormat.ATXT`, and writes it with `ExportTxt`, as `gantt.exportTxt` does. A panic is returned as an error and nothing is written.
- `Fork.asciiDraw` draws the label of the root in an `ATable` of one cell with a padding of 1, then, for each child of the root, moves the plan by 4 and 4 and draws the child: an `ITFComposed` with a padding of 7 and an `ITFLeaf` with a padding of 3. `ITFComposed.asciiDraw` does not draw the children of its node, as upstream's.
- The text check of the engine lets `core.WBS` through, next to `core.Sequence` and `core.Gantt`.

### D8. Registration and WASM

- `wbs.Factory` (`uml.NewFactory(core.WBS, ...)`) goes into `engine.Factories` after `mindmap.Factory` and before `packetdiag.Factory`, as `WBSDiagramFactory` comes after `MindMapDiagramFactory` and before `PacketDiagramFactory` in upstream's `PSystemBuilder`.
- `core.WBS` joins `supportedTypes`. `TestCheckFamily` gets a `@startwbs` check, and its unsupported family, the check of `internal/core/builder_test.go`, `TestNotSupportedYet` and `TestInspectNamesAndStatus` use `@startjson` instead of `@startwbs`, with the message `json diagrams are not supported yet`. The `why` of `diagram-source/unsupported-family.go.yaml` names `@startjson`.
- The WASM build needs no special case, because WBS diagrams do not use Graphviz.

### D9. Documentation

- The README and `site/index.html` get the row `WBS | yes | yes | yes` in the table of supported diagrams. The Graphviz note of the page and the last sentence of the README's Graphviz section name WBS diagrams among the diagrams that do not need `dot`. The README's conformance counts follow the report.
- The conformance tally of the page follows the release on the page (v0.2.0), and the playground note of the page and the README's sentence about the playground follow the WASM module of that release: they do not change.

### D10. Conformance

- Curated cases go under `testdata/conformance/cases/wbs-diagram/` with `family: wbs` and `type: wbs`, one per compared scenario, plus `diagram-source/wbs-detected`, `output-formats/wbs-txt` for "ASCII WBS diagram" and an embedded case (`family: uml`, `type: sequence`). The PNG scenario uses `formats: [svg, png]`, the UTXT scenario `[utxt]`, the TXT scenarios `[txt]`, the others `[svg]`. Their references come from `make reference`, and CI renders them again and checks them against the committed files.
- The embedded case is an `@startuml` case, so the golden dumps of `internal/sequencediagram`, `internal/classdiagram` and `internal/activitydiagram3` are regenerated, as for the mindmap change.
- Scenarios that are not compared with Java get `.go.yaml` cases: "WBS diagram without Graphviz" (environment), "WBS diagram without nodes" (upstream draws a crash report with the Java stack, this change returns an error, as for the gantt drawing failure) and "WBS diagram in WASM" (environment).
- The cases come before the code. Until the factory is registered, the WBS cases report "not supported yet", which the runner counts as unsupported, `TestDiagramReferences` skips the curated WBS error cases, and the embedded case fails in `make conformance`.
- The 7 Vega WBS cases are compared once the factory is registered. A case that still differs at the end is listed in `known-differences.yaml` with its reason.

### D11. Port order

The layers are ported bottom up, each with unit tests: `FtileBoxOld`, the model (`WElement`, `getSmartLevel`, `add`, `link`) and the diagram, the commands and the factory, the drawing (`WBSTextBlock`, `ITFLeaf`, `ITFComposed`, `Fork`, `WBSLink`, the export), the text output, then the registration.

## Risks / Trade-offs

- **[Positions from the translation of the graphic]** The links use the positions that the nodes record from `Common().Translate()`. A Go graphic in the drawing path that implements `klimt.CommonUGraphic` where upstream's graphic is not an `AbstractCommonUGraphic`, or the reverse, would drop or move the links. → The scenarios "Link between aliased nodes", "Scaled WBS diagram" (a link under `scale`) and "Old order" (a link below the warning banner) check the positions against upstream. A difference is fixed in `internal/wbs` or in the graphic that differs from upstream.
- **[Shared FtileBoxOld]** `SetMinimumWidth` empties the geometry cache of a type that mindmap diagrams use. → Mindmap diagrams never call it. The mindmap cases and the golden dumps of `internal/activitydiagram3` stay unchanged.
- **[Second caller of the style cascade]** `MergedStyleSpecial` and the ancestor rank have one caller so far, the mindmap diagram. → The scenarios "Style of the descendants" and "Nearest ancestor" and the Vega cases `star-and-plain-same-stereotype` and `wbs-deep-ancestor-cascade-with-stereotype` check them against upstream. A difference there is fixed in `internal/klimt`, with its style golden tests unchanged.
- **[Floating point]** The layout adds and compares `double` values. → The drawing keeps Java's order of operations on `float64`, and the Vega cases with several levels and `autowidth` check it.
- **[Tests that use wbs as the unsupported example]** They would fail once WBS is registered. → D8 moves them to `@startjson`, which stays unsupported until the JSON diagrams change (#39).
- **[Order of the archives]** The delta specs of `diagram-source` and `wasm-build` contain the requirements of `add-mindmap-diagrams` with WBS added. Archiving this change before that one would let the mindmap archive replace them without WBS. → `add-mindmap-diagrams` is archived first.

## Migration Plan

Nothing to migrate. Inputs that failed with "wbs diagrams are not supported yet" now render. Rollback means reverting the change.
