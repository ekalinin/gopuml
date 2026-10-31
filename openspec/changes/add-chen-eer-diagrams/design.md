# Design

## Context

See proposal.md for the motivation. The current state that shapes the approach:

- Upstream's `cheneer` package has `ChenEerDiagram` (an `AbstractEntityDiagram` with a stack of owners), `ChenEerDiagramFactory` and six commands: `CommandCreateEntity`, `CommandCreateAttribute`, `CommandAssociate`, `CommandEndGroup`, `CommandSimpleSubclass` and `CommandMultiSubclass`. The factory adds `CommonCommands.addCommonCommands1` and `CommandRankDir` first. `ChenEerDiagram` overrides neither `cleanId`, `updateLasts`, `makeDiagramReady` nor `checkFinalError`.
- The drawing is upstream's svek: `GeneralImageBuilder` maps the leaf types `CHEN_ENTITY`, `CHEN_RELATIONSHIP`, `CHEN_ATTRIBUTE` and `CHEN_CIRCLE` to `EntityImageChenEntity`, `EntityImageChenRelationship`, `EntityImageChenAttribute` and `EntityImageChenCircle`. Each image opens a group with only `UGroupType.ID`, so the SVG has a bare `<g>` around each element, as the Vega references show, and not the `<g class="entity" ...>` of the other images. The images take their styles from `StyleQueries.CHENEER_*`: `root`, `element`, `chenEerDiagram` and `chenEntity`, `chenRelationship`, `chenAttribute` or `circle`, plus `title` for the names, with the stereotype.
- The Go side already has `abel.AbstractEntityDiagram` (its `DotStrings` and its description `(N entities)`), `InitCucaDiagram` (whose nil hooks are upstream's defaults), `QuarkInContext`, `ReallyCreateLeaf`, `Entity.MuteToTypeSymbol`, `SetStereotype` and `SetStereostyle`, `abel.NewLink`, `LinkArgBuild`, `Link.SetPortMembers`, `CucaDiagram.PortID`, `abel.CommandRankDir`, `uml.AddCommonCommands1`, the color parsers (`color.SimpleColor`), `klimt.DisplayWithNewlinesString`, `klimt.DisplayNULL`, the leaf types, the SNames `chenEerDiagram`, `chenEntity`, `chenRelationship`, `chenAttribute` and `circle`, `LinkType.GoBold`, `WithMiddleSubset` and `WithMiddleSuperset` with their extremities in svek edges, the shape types `RECTANGLE`, `DIAMOND` and `OVAL` in the DOT output, `IEntityImageMargin` and `IEntityImageMarginLine`, `FontConfiguration.Underline`, `NewUStroke`, and the SVG writer, which ignores `UGroupID`.
- `GeneralImageBuilder` (`internal/svek/generalimagebuilder.go`) panics with `notPorted("EntityImageChen")` for the four Chen leaf types.
- Upstream's `TextBlockInEllipse` is ported twice: `klimt.TextBlockInContainingEllipse` in `internal/klimt/svek_footprint.go`, which use cases and requirements use, and the stub `klimt.TextBlockInEllipse` in `internal/klimt/shape_textblockmisc.go`, which panics with "not ported" and is called only by its panic check in `klimt_test.go`.
- `uml.DiagramTypeStyleName` maps no style name to `core.ChenEER` and returns `activityDiagram`, as upstream's `DiagramType.getStyleName`.
- `core.FindStartTypes` already maps `@startchen` to `core.ChenEER`, and `core.ChenEER.HumanReadableName()` is `chen_eer`. `core.CheckFamily` reports every family outside `supportedTypes` as not supported yet. The engine rejects TXT and UTXT for every type other than sequence, gantt and WBS.
- The class, description, state and legacy activity packages each have their own `export.go` with `ExportDiagram`, `ExportDiagramContext`, `ExportDiagramLayout` and `abortedError`, built on `svek.TextBlock` and `uml.ExportTitled`.
- `TestScenarioCoverage` reads the delta specs of active changes, so every scenario of this change needs a case once the specs are committed. `TestDiagramReferences` of `internal/errdiagram` draws the error image of every curated case whose reference is an error.

## Goals / Non-Goals

**Goals:**

- Port the `cheneer` package and the four Chen entity images, with the SVG byte-identical to upstream's apart from the encoded source.
- Keep upstream's quirks: the scenarios "Line read as an attribute", "Bold only from the left", "Attribute colors", "Union", "Unclosed block", "Dotted names" and "Title and legend" pin the ones users can see.
- Keep the results of every other diagram type unchanged. The shared code that changes is the Chen branch of `GeneralImageBuilder`, the removal of an unused stub, the list of supported families and the factory list.

**Non-Goals:**

- Porting the `explainArg` methods, which only upstream's `@Explain` tooling calls, the `mandatoryFirstTokensFast` lists, which only speed up upstream's command matching, and `getStyleQuery`, which no svek code calls.
- Giving Chen diagrams a style name of their own for the title and the legend (see `uml.DiagramTypeStyleName` in Context).

## Decisions

### D1. Package layout

`internal/cheneer` holds the package, with upstream's names and a `Provenance` line for each Java file it ports: `doc.go`, `diagram.go` (`ChenEerDiagram`), `factory.go` (`ChenEerDiagramFactory`), `commands_entity.go` (`CommandCreateEntity`, `CommandEndGroup`), `commands_attribute.go` (`CommandCreateAttribute`), `commands_link.go` (`CommandAssociate`, `CommandSimpleSubclass`, `CommandMultiSubclass`) and `export.go`. The name is upstream's package name, as `salt` and `packetdiag`.

Alternative: add the commands to `internal/descdiagram`, which also builds on `AbstractEntityDiagram`. Chen diagrams have their own family, diagram type and factory upstream, so a package of their own keeps the description diagrams unchanged.

### D2. The diagram

- `ChenEerDiagram` embeds `abel.AbstractEntityDiagram` and calls `InitCucaDiagram` with `core.ChenEER` and nil hooks, which are upstream's `cleanId`, `updateLasts` and `isHideEmptyDescriptionForState` of `CucaDiagram`. The SVG has `data-diagram-type="CHEN_EER"`, and the description comes from `AbstractEntityDiagram`, so it counts the attributes and the circles (the Vega case `ChenMovieAlias_Test` has `(30 entities)`).
- The stack of owners is a slice of `*abel.Entity` with `PushOwner`, `PopOwner` (false on an empty stack) and `PeekOwner` (nil on an empty stack). It does not use the current group of `CucaDiagram`, as upstream's comment explains: the attributes are linked to their owner, not nested in it.
- The diagram adds no `MakeDiagramReady` or `CheckFinalError` of its own, as upstream. An owner left on the stack at the end is not an error (scenario "Unclosed block").

### D3. The factory and the commands

- `Factory` is `uml.NewFactory(core.ChenEER, commands, NewChenEerDiagram)`. The commands come in upstream's order: `uml.AddCommonCommands1`, `abel.CommandRankDir`, then entity, attribute, associate, end group, simple subclass and multiple subclass. There is no `newpage`, as upstream.
- The patterns are upstream's, written with the macros of `internal/command/regex` (`%pLN`, `%g`, `%s`). Because the attribute pattern accepts spaces and `:`, it matches every line that the common commands and the entity command leave, apart from the links and `}`, which it cannot match. Nothing is added to prevent it (scenario "Line read as an attribute").
- The entity command reads the type, builds the quark with `QuarkInContext(true, CleanID(code))`, creates the leaf with `DisplayWithNewlines` of the display name or of the quark's name, or calls `MuteToTypeSymbol(type, nil)` on an existing one and fails with `Bad name` when it returns false. It sets the stereotype and the stereostyle, the colors with `color.SimpleColor(color.TypeBack)`, and pushes the entity.
- The attribute command fails with upstream's message when the stack is empty. The id is `owner.Name() + "/" + CleanID(trimmed code)`, built with `QuarkInContext(true, id)`, and an existing quark fails with `Attribute already exists`. The display is the display name or the short id. The colors use `color.SimpleColor(color.TypeLine)`, on the attribute and on its link, a `NONE`/`NONE` link from the attribute to its owner with `LinkArgBuild(DisplayNULL(), 2)`. A trailing `{` pushes the attribute.
- The associate command and the simple subclass command look up both names with `QuarkInContext(true, CleanID(name))` and fail with `No such entity: <name>` when a quark has no entity. The bold line comes from the first participation character only (scenario "Bold only from the left"). The cardinality goes through `DisplayWithNewlinesString`, so that a missing one is `Display.NULL`, as upstream's `getWithNewlines(null)`. The link length is 3, the port members are `PortID` of both names, kept although Chen names cannot hold `::`, and the colors use `color.SimpleColor(color.TypeLine)`.
- The multiple subclass command keeps upstream's order of side effects, so that the UIDs of the entities and links (`ent0001`, `lnk4`), which the SVG shows, follow upstream's: it creates the circle first, with the quark `superclass + "/" + symbol + subclasses + "/center"` built with `QuarkInContext(false, ...)` from the raw text between the braces, spaces included (the Vega references have path ids such as `PERSON/U CUSTOMER, DIRECTOR /center`), then looks up the superclass, then adds the link to the superclass (length 2) and one link per subclass (length 3), splitting the subclasses on `,` and trimming them. The colors use `color.SimpleColor(color.TypeBack)` on the circle and on every link.
- Each failure returns `command.Error` with upstream's message, so the shared error diagram adds `(Assumed diagram type: chen_eer)`.

### D4. The entity images

- `internal/svek/image_entityimagechen.go` ports the four images, with a `Provenance` line for each Java file. Upstream repeats `hasStereotype` and `applyColor` in each class; the Go file has them once, as unexported helpers.
- The styles are `klimt.StyleSignatureBasicOf(SNameRoot, SNameElement, SNameChenEerDiagram, <name>)`, plus `SNameTitle` for the names, with `WithTOBECHANGED(stereotype)` and `MergedStyle`, as `image_entityimagestate.go` ports upstream's `StyleQuery.withStereotype`. The circle uses `SNameCircle`, as upstream.
- Each image opens a `UGroup` with only `UGroupID` and the quark's `StringPoint` (upstream's `toStringPoint`), which the SVG writer ignores, giving the bare `<g>` of the references.
- The geometry is upstream's: the entity is its name plus `2*IEntityImageMargin + 2*IEntityImageMarginLine` with the name at `IEntityImageMargin + IEntityImageMarginLine`; the relationship is a diamond of aspect ratio 2 whose diagonal is `(width + 2*height)/sqrt(5) + 2*IEntityImageMargin`, with `math.Sqrt` and Java's order of operations; the attribute takes the dimension of `klimt.NewTextBlockInContainingEllipse` around its name and draws a plain ellipse of that size with the name at y 6; the circle is 25 by 25. `<<weak>>` and `<<multi>>` add the shape inset by 3 on each side, `<<identifying>>` the diamond inset by 10 and 5, `<<derived>>` replaces the stroke with `NewUStroke(10, 10, thickness)`, and `<<key>>` underlines the name.
- The shape types are `RECTANGLE`, `DIAMOND`, `OVAL` and `OVAL`, which the DOT output already handles.
- `GeneralImageBuilder` returns the four images instead of the `notPorted` panic.

### D5. Removal of the `TextBlockInEllipse` stub

- `internal/klimt/shape_textblockmisc.go` loses the type `TextBlockInEllipse`, `NewTextBlockInEllipse`, its two methods and its `Provenance` line for `TextBlockInEllipse.java`, which stays in `svek_footprint.go`.
- `klimt_test.go` loses the panic check of `NewTextBlockInEllipse`.
- The comment of `TextBlockInContainingEllipse` in `svek_footprint.go` stops naming the stub.
- `notPorted` stays: other code of `klimt` still calls it.

Alternative: rename `TextBlockInContainingEllipse` to `TextBlockInEllipse`, upstream's name. It would change the use case and requirement code for no change in behavior, so the working name stays.

### D6. Export

`export.go` follows `internal/descdiagram/export.go`: `ExportDiagram`, `ExportDiagramContext` and `ExportDiagramLayout`, with `svek.TextBlock` over `DotStrings` and the layout of the render context, `uml.ExportTitled`, `abortedError` for a cancelled context, and `export.NotImplementedError` for TXT, UTXT and XMI. The engine already rejects TXT and UTXT before the export, with the message `TXT output is not supported for chen_eer diagrams`.

Alternative: move the export of the CucaDiagram packages into a shared helper. It would change four packages for no change in behavior, so the Chen package keeps its own copy, as each of them does.

### D7. Styles

Nothing changes in the styling code. `uml.DiagramTypeStyleName` keeps returning `activityDiagram` for `core.ChenEER` (scenario "Title and legend"), and the skin keeps no Chen entries, so the elements take the default style of `element` (scenario "Entities, relationship and attributes").

### D8. Registration and WASM

- `cheneer.Factory` goes at the end of `engine.Factories`, after `timing.Factory`, as `ChenEerDiagramFactory` comes after every factory of the Go list in upstream's `PSystemBuilder`.
- `core.ChenEER` joins `supportedTypes`. `TestCheckFamily` gets a `@startchen` check.
- The WASM build needs no special case: the layout of the render context already returns the Graphviz-not-available error for the svek diagrams.

### D9. Documentation

- The README and `site/index.html` get the row `Chen EER | yes (needs dot) | yes (needs dot) | no` at the end of the table of supported diagrams.
- The README's introduction and its Graphviz section, and the Graphviz note of the page, name Chen EER diagrams among the diagrams that need `dot`. The README's WASM note names them among the diagrams that report that graph layout is not available.
- The README's conformance counts follow the report. The sentence on the types that are not supported yet does not change, as it does not name Chen diagrams.
- The conformance tally of the page and its playground note follow the release on the page, so they do not change.

### D10. Conformance

- Curated cases go under `testdata/conformance/cases/chen-eer-diagram/` with `family: chen` and `type: chen_eer`, the type of the Vega Chen cases in the report, one per compared scenario, plus `diagram-source/chen-detected`. The PNG scenario uses `formats: [svg, png]`, the others `[svg]`. "Chen diagram in a note" is an `@startuml` case (`family: uml`, `type: sequence`), so the golden dumps of `internal/sequencediagram`, `internal/classdiagram` and `internal/activitydiagram3`, which cover every `@startuml` case, are regenerated. "Dark mode" is a CLI case, `chen-eer-diagram/dark-mode.cli.yaml`, with `--svg --dark-mode` and `compare: [exit, files]`. Their references come from `make reference`, and CI renders them again and checks them against the committed files.
- Scenarios that are not compared with Java get `.go.yaml` cases: "Chen diagram without dot" (environment), "TXT Chen diagram" (upstream draws its generic text output) and "wasm-build: Chen diagram in WASM" (environment).
- The cases come before the code. Until the factory is registered, the Chen cases report "not supported yet", which the runner counts as unsupported, `TestDiagramReferences` skips the curated Chen error cases, and the embedded case fails in `make conformance`.
- The 4 Vega Chen cases are compared once the factory is registered. A case that still differs at the end is listed in `known-differences.yaml` with its reason.

### D11. Port order

Cases first, then the removal of the stub, the package with its commands, the images and the export, then the registration.

## Risks / Trade-offs

- **[UIDs and path ids]** The SVG shows the UIDs of the entities and links and the names of both ends in the path ids, so the order in which the commands create them, and the raw text of the subclass groups, must be upstream's. → D3 keeps upstream's order of side effects and the raw text, and the Vega cases `ChenMovieExtended_Test` and `ChenMovieAlias_Test`, which use every command, check them.
- **[Floating point in the diamond]** The size of a relationship goes through `sqrt(5)` and a division. → The code keeps Java's order of operations on `float64`, and `ChenRankDir_Test` checks the points of the diamond.
- **[Size of the attributes]** The ellipse of an attribute comes from the footprint of its name, which depends on how the text is drawn. → `TextBlockInContainingEllipse` already gives upstream's sizes for use cases, and the Vega cases check attributes with underlined, plain and composite names.
- **[Graphviz version]** The layout depends on the version of `dot`. → As for the other svek diagrams, the references come from the reference image with its Graphviz, and `make conformance REQUIRE=all GO_IN_IMAGE=1` renders this implementation in the same image. Native runs on arm64 hosts use the host's Graphviz and are indicative only.
- **[Shared svek code]** `GeneralImageBuilder` changes. → Only the Chen branch changes, and the results of the other types must stay the same in `make conformance`.
- **[Removal of an exported name]** `klimt.TextBlockInEllipse` goes away. → `klimt` is an internal package, and the only caller is the removed panic check, so the build shows any other use.

## Migration Plan

Nothing to migrate. Inputs that failed with "chen_eer diagrams are not supported yet" now render, or fail with upstream's error. Rollback means reverting the change.
