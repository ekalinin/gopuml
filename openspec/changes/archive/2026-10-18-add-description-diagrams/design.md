# Design

## Context

See proposal.md for the motivation. The current state that shapes the approach:

- Class diagrams with `allowmixing` already draw description elements, so these are ported: `svek/image_entityimagedescription.go`, all 29 `USymbol` classes (`internal/decoration/symbol`, registry in `internal/abel/usymbol.go`), the clusters with a symbol, the extremities, and the generic commands (`CommandPackageWithUSymbol`, `CommandTogether`, notes, `CommandUrl`, `CommandHideShow2`, `CommandRemoveRestore`, `CommandRankDir`, the multiline element commands).
- Partly ported: `CommandCreateElementFull` (only its constants and `ExistsWithBadType3`, in `internal/abel/descdiagram_command_createelement.go`) and `CommandLinkElement` (only `LINE_STYLE`).
- Upstream's `descdiagram/command/Labels.java` and `StringWithArrow.java` are ported privately inside `internal/classdiagram/commands_link.go`. The `map` and `json` commands (`objectdiagram/command/*`) are ported in `internal/classdiagram/objectdiagram_commands.go` and are tied to `*ClassDiagram`.
- Missing: `DescriptionDiagramFactory`, `DescriptionDiagram`, `CommandCreateDomain`, `EntityImagePort` with `AbstractEntityImageBorder`, `FrontierCalculator` (stubbed in `svek/cluster.go`), and the Problem Frames images (`EntityImageDomain`, `EntityImageDesignedDomain`, `EntityImageMachine`, `EntityImageRequirement`, `BoxedCharacter`, stubbed in `svek/generalimagebuilder.go`).
- `(*abel.CucaDiagram).CleanID` is a plain method, and the shared commands call it directly, so an override in a description diagram would be skipped.
- `internal/core/detect.go` has two detectors for description syntax: an `unsupportedTypes` entry (run after every factory fails) and `detectDescriptionOnly` (run before the activity factory, because upstream's description factory comes first).
- Embedded diagrams: `klimt` collects the `{{` lines and then needs the `ISkinSimple` of the text to implement `klimt.EmbeddedDiagramMaker`. Nothing implements it, so every `{{` panics with "not ported yet". `SkinParam` already holds the `*core.RenderContext` (upstream's `PathSystem`). The SVG driver draws `UImageSvg`, and both drivers draw `UImage`.
- The 10 Vega description cases are compared as SVG or PREPROC, including the `graphml` and `xmi` ones, which fall back to SVG.

## Goals / Non-Goals

**Goals:**

- Port description diagrams file by file from upstream, reusing what class diagrams already ported.
- Embedded diagrams for every type the engine supports, through the render context, without a global registry.
- Keep the sequence, class and activity conformance results unchanged while code moves out of `internal/classdiagram`.

**Non-Goals:**

- Archimate commands, sprites, OpenIconic, emoji and the stdlib (see the proposal).
- TXT, UTXT, XMI and GraphML output (`CucaDiagramTxtMaker`, `XmiDescriptionDiagram*`, the GraphML exporter).
- Embedded diagrams in LaTeX/TikZ output, which this implementation does not produce.

## Decisions

### D1. Package layout mirrors upstream

- `internal/descdiagram` holds `DescriptionDiagram`, `DescriptionDiagramFactory`, `CommandCreateDomain` and the commands of `descdiagram/command` (`CommandCreateElementFull`, `CommandLinkElement`, `Labels`, `StringWithArrow`), one Go file per Java file where practical, as the other diagram packages do.
- `Labels` and `StringWithArrow` move from `internal/classdiagram/commands_link.go` to `internal/descdiagram`, exported. `internal/classdiagram` then imports `internal/descdiagram`, the same direction as upstream (`classdiagram.command` imports `descdiagram.command`).
- The `map` and `json` commands move from `internal/classdiagram/objectdiagram_commands.go` to `internal/abel/objectdiagram_commands.go`, generic over `abel.Diagram` instead of typed on `*ClassDiagram`, as upstream types them on `AbstractEntityDiagram`. `internal/abel` already holds the other commands that class and description diagrams share (notes, packages, element creation). The `object` commands and `AbstractClassOrObjectDiagram` stay in `internal/classdiagram`, because they are class-only.
- `CommandCreateElementFull` goes into `internal/descdiagram`, using the patterns and `ExistsWithBadType3` already in `internal/abel/descdiagram_command_createelement.go`. Class diagrams with `allowmixing` use their own `CommandCreateElementFull2` (`internal/classdiagram/commands_element.go`), so they do not depend on it.

Alternative: export `Labels` and the object commands from `internal/classdiagram` and let `internal/descdiagram` import it. The diff would be about the same size, but the dependency would run against upstream's, and a description package would depend on the class package.

### D2. A CleanID hook

`abel.CucaHooks` (`internal/abel/atmp_cucadiagram.go`) gets a `CleanID` hook, and `CucaDiagram.CleanID` calls it, so every command that calls `CleanID` gets the override without a change at the call site. The previous body becomes `DefaultCleanID`, upstream's `super.cleanId`. The hooks are a separate value, not the diagram itself: a diagram type that embeds `CucaDiagram` already has a promoted `CleanID`, which would call itself. `DescriptionDiagram` sets it to upstream's `cleanId`, which strips the short-form markers (`:Actor:`, `(Use case)`, `[Component]`, `()`). Class diagrams keep the current behavior through the default.

### D3. Registration and detection

- `descdiagram.Factory` goes into `engine.Factories` between `classdiagram.Factory` and `activitydiagram3.Factory`, which gives upstream's order of `PSystemBuilder` without the factories that are not ported.
- `detectDescriptionOnly` and its call in `core.CreateDiagram` are removed: the real factory now comes before the activity factory.
- The description entry of `unsupportedTypes` becomes an `archimate` entry with only the `^archimate[%s]` pattern, so that archimate input reports "archimate diagrams are not supported yet". The three archimate commands are not registered, so a block with an archimate line fails in the description factory and reaches that detector.

### D4. Ports and Problem Frames images

`EntityImagePort`, `AbstractEntityImageBorder` and `FrontierCalculator` are ported into `internal/svek`, replacing the stubs in `generalimagebuilder.go` and `cluster.go`. `EntityImageDomain`, `EntityImageDesignedDomain`, `EntityImageMachine`, `EntityImageRequirement` and `BoxedCharacter` are ported the same way. `ClusterDotString` already handles `hasPort`.

### D5. Export and formats

The description export follows `internal/classdiagram/export.go`: the generic `svek.TextBlock` of the entity diagram, for SVG and PNG. TXT and UTXT return the engine's `UnsupportedFormatError`, as class diagrams do. The WASM build needs no special case: `cmd/plantuml-wasm` turns `ErrGraphvizNotFound` into the "graph layout is not available" error for every diagram, whatever its type. The scenario "Description diagram in WASM" checks this.

### D6. Embedded diagrams through the render context

- `core.RenderContext` gets an embedded-diagram renderer, a function the engine sets when it creates the context. It parses the given lines as one block with a copy of the render context, as upstream builds a `BlockUml`: without the definitions, configuration lines and pragmas of the command line (`Defines.createEmpty()`), and starting with the skinparam values of the embedding diagram (upstream's `Previous`, applied in `uml.NewBase`). It renders the block as SVG or PNG and returns the image. The engine sets the renderer on a copy of the caller's context in `engine.Parse`, so concurrent renders do not share a modified context.
- `klimt.SkinParam` implements `EmbeddedDiagramMaker` with that renderer, taken from its `pathSystem` field. The skin parameter wrappers (`SkinParamDelegator` and the others) forward the call to the wrapped skin parameters.
- The atom is upstream's `EmbeddedDiagram`: its size comes from the rendered image. Its drawing uses `UImageSvg` when the output is SVG (with the `<?plantuml ...?>` processing instruction removed, as upstream does) and `UImage` otherwise. The rendering happens once per atom and is reused.
- A nested diagram of an unsupported type renders the error image that the engine produces for it, so no special case is needed.
- Without a renderer in the context (a `SkinParam` built with a nil context, as in some unit tests), the atom keeps today's "not ported yet" panic.

Alternatives:

- A package-level hook in `klimt` set by `internal/engine` at init. It would be a global singleton, which the foundation design replaced with the explicit render context (foundation D3).
- Importing the engine from `klimt`. That is an import cycle.

### D7. Conformance

- Curated cases go under `testdata/conformance/cases/description-diagram/`, one per scenario, plus the new scenarios of `diagram-source` and `styling`. Their references come from `make reference`. CI renders them again on the cache miss and checks them against the committed files.
- Scenarios that are not compared with Java get `.go.yaml` cases: "TXT for a description diagram" (upstream draws TXT, this implementation rejects it), "Description diagram without dot", "Description diagram in WASM", and "Embedded diagram of an unsupported type" (upstream supports state diagrams).
- The cases come before the code: `TestScenarioCoverage` reads the delta specs of active changes, so without the cases `make test` would fail on the branch. Until the factory is registered, the new description cases report "not supported yet", which the runner counts as unsupported, not failed.
- The 5 `tikz` entries leave `known-differences.yaml`. CI then requires the 10 Vega description cases and the 5 `tikz` cases to pass, together with every other group.

## Risks / Trade-offs

- **[Detection order]** With the description factory before the activity factory, a block that both factories accept now becomes a description diagram. → That is upstream's order, so the references expect it. The activity and Vega cases in CI show any change.
- **[Moving shared code]** Moving `Labels`, `StringWithArrow` and the object commands can change class diagram behavior. → The moves change no logic. The class unit tests and the class conformance cases must stay green, and CI requires them.
- **[Sprites in deployment diagrams]** Stereotypes such as `<<$server>>` and `<$icon>` references crash the drawing today, because sprite drawing is not ported. Deployment diagrams use them often. → Out of scope, and unchanged by this change. The error is explicit, not a wrong drawing.
- **[Embedded rendering cost]** A nested diagram is parsed and rendered once per atom, including its own `dot` run for graph diagrams. → As upstream. The atom caches its image, so measuring and drawing do not render twice.
- **[PNG in PNG]** For PNG output, the nested diagram is rendered as PNG and decoded into a `UImage`. → Upstream does the same (`getImage` reads the PNG it exported). The PNG is decoded with `image/png` into a `PortableImage`, as `internal/errdiagram/logo.go` does for the logo.

## Migration Plan

Nothing to migrate. Inputs that failed with "not supported yet" or crashed on `{{` now render. Rollback means reverting the change.
