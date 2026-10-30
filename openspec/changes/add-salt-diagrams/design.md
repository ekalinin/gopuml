# Design

## Context

See proposal.md for the motivation. The current state that shapes the approach:

- Upstream's `salt` package has 62 files: `PSystemSalt`, `PSystemSaltFactory`, `CommandAnything`, `SaltDictionary`, the tokenizer (`DataSource`, `DataSourceImpl`, `Terminated`, `Terminator`), the layout helpers (`Cell`, `Positionner2`), 19 element factories under `factory/` and 32 element classes under `element/`. `Position` is not used. `element/Skeleton2` is used only by `klimt.creole.atom.AtomTree`, and is already ported in `internal/klimt/creole_atom.go`. The `TeaVM.a()` branches only hold Java `assert` statements, which are off without `-ea`.
- All the work of upstream happens in `getTextBlock`, at drawing time: the lines that `CommandAnything` collected are filtered by `manageSprite`, cut into tokens by `DataSourceImpl`, turned into an element tree by the factories, measured and drawn. A parse error can only come from the commands; every other failure throws while drawing.
- `PSystemSalt` is a `TitledDiagram`. Its factory adds `addCommonCommands2`, `addCommonScaleCommands`, `addTitleCommands` and `CommandAnything`, which matches any line without trimming it. Its margins are 5. Its text block draws the element tree twice, with the z-index 0 then 1, in `HColors.BLACK`, and its back color is the background of the skin parameters.
- `SaltDictionary` is the `ISkinSimple` of every widget: no skin values, no padding, Guillemet `GUILLEMET`, the monospaced family of `Parser`, a tab size of 8, 96 dpi, an empty pragma and the `CreoleParser` sheet. It holds the pixel images by name and the sprites, because `PSystemSalt` overrides `addSprite`.
- The colors of the widgets are `HColors.BLACK.withDark(WHITE)`, `WHITE.withDark(BLACK)` and the grays `#8/#8`, `#A/#6`, `#B/#5`, `#D/#3` and `#E/#2`. The fonts are `UFontFactory.byDefault(n)`, which is `sansSerif(12)` for every `n`.
- The Go side already has `uml.Base`, `uml.NewFactory`, `uml.AddCommonCommands2`, `uml.AddCommonScaleCommands`, `uml.AddTitleCommands`, `uml.ExportTitled` (which returns a panic while drawing as an error, without output), `uml.Throw` with the Java exception classes, the creole engine (`klimt.DisplayCreate`, `DisplayWithNewlines`, `Display.Create`), `klimt.ISkinSimple`, `klimt.FontConfigurationBlackBlueTrue`, `font.ByDefault` (which keeps upstream's `sansSerif(12)`), `color.HColor.WithDark`, `HColorSet.GetColorOrWhite`, `klimt.ChangeBack`, `URectangle.Rounded`, `UPath`, `UEllipse`, `UPolygon`, `UPixel` with its SVG driver, `NewUStroke`, `klimt.TextBlockUtilsMinMax` over `LimitFinder`, `geom.XDimension2D.AtLeast`, `klimt.JavaHashOrder`, `klimt.JawsBlockE1Newline`, `jfmt.ParseDouble`, the sprite commands, the SName `saltDiagram` and `uml/chrome.go`, which maps `core.Salt` to it. `klimt` has `javaTokenize` and `javaTokenizeWithDelims` (Java's `StringTokenizer`), unexported.
- `Splitter.purgeAllTag` and its `htmlTag` pattern are not ported: the comment of `internal/klimt/creole_command.go` says upstream does not use them, which salt does.
- The sprite commands of `internal/uml/sprites.go` call `d.Titled().AddSprite(...)`, so a diagram cannot keep its sprites itself.
- `core.CheckFamily` reports every family that is not in `supportedTypes` as not supported yet. `core.FindStartTypes` already maps `@startsalt` to `core.Salt`, `internal/core/builder.go` already has the error `This looks like a salt diagram. Please use @startsalt instead of @startuml.`, and the creole engine already knows `{{salt`. No test uses `@startsalt` as its example of a family that is not supported yet.
- The engine rejects TXT and UTXT for every type other than sequence, gantt and WBS, so salt diagrams get the unsupported-format error without a change.
- `TestScenarioCoverage` reads the delta specs of active changes, so every scenario of this change needs a case once the specs are committed. `TestDiagramReferences` of `internal/errdiagram` draws the error image of every curated case whose reference is an error.

## Goals / Non-Goals

**Goals:**

- Port the whole `salt` package file by file, with the SVG byte-identical to upstream's apart from the encoded source.
- Keep upstream's quirks: the scenarios "Row read as a command", "Line break in a cell", "Open droplist in a tree", "Highlighted menu entry", "Scroll area" and "Sprite in the title" pin the ones users can see.
- Keep the results of every other diagram type unchanged. The shared code that changes is the list of supported families, the factory list, the sprite commands and `klimt` (two exported tokenizers and `purgeAllTag`).

**Non-Goals:**

- Porting the `explainArg` method of `CommandAnything`, which only upstream's `@Explain` tooling calls, the `TeaVM.a()` asserts, the `Log` calls, `Position`, and the deprecated constructor of `PSystemSalt`.
- Reproducing upstream's order of the open lists of several menus (D9).

## Decisions

### D1. Package layout

`internal/salt` holds the whole package, with upstream's names and a `Provenance` line for each Java file it ports, grouped by layer: `diagram.go` (`PSystemSalt`), `dictionary.go` (`SaltDictionary`), `factory.go` (`PSystemSaltFactory`, `CommandAnything`), `datasource.go` (`DataSource`, `DataSourceImpl`, `Terminated`, `Terminator`), `positionner.go` (`Cell`, `Positionner2`), `factory_complex.go` (`ElementFactory`, `AbstractElementFactoryComplex`, the factories of groups, scroll areas, border layouts, trees, tabs and menus, `ScrollStrategy`), `factory_simple.go` (the other factories), `element.go` (`Element`, `AbstractElement`, `AbstractElementText`, `ElementEmpty`, `WrappedElement`), `element_widgets.go` (`ElementText`, `ElementButton`, `ElementTextField`, `ElementRadioCheckbox`, `ElementDroplist`, `ElementLine`, `ElementImage`), `pyramid.go` (`ElementPyramid`, `ElementPyramidScrolled`, `TableStrategy`, `LeftFirst`, `TopFirst`), `grid.go` (`Grid`, `Grid2`, `Segment`), `border.go`, `tree.go` (`ElementTree`, `ElementTreeEntry`, `Skeleton`, `ListWidth`), `tabbar.go`, `menu.go` (`ElementMenuBar`, `ElementMenuEntry`, `ElementMenuPopup`) and `export.go`. `Skeleton2` stays in `internal/klimt`, where creole trees use it.

Alternative: subpackages `element` and `factory`, as upstream. The factories build the elements and the complex factories hold each other, so one package keeps the Go code free of interfaces that exist only to break import cycles.

### D2. PSystemSalt and its dictionary

- `PSystemSalt` embeds `uml.Base` with the type `core.Salt`, so the SVG has `data-diagram-type="SALT"` and the annotations, skins and styles come from the shared code. `Description` returns `(Salt)` and `DefaultMargins` returns 5 on each side.
- `add` splits a line on `klimt.JawsBlockE1Newline`, as upstream's `StringLocated.expandsNewline`.
- `SaltDictionary` implements `klimt.ISkinSimple` with upstream's values (see Context), its `Sheet` and `SheetStereo` return the creole parser, and `CopyAllFrom` and `Values` panic with `uml.Throw(uml.UnsupportedOperationException, "")`, as upstream throws. Its map of elements and its map of sprites are Go maps, because upstream only reads them by name.
- `ExportDiagram` renders with `uml.ExportTitled`, as `internal/wbs/export.go` does. The text block is built by `getTextBlock` on each call, as upstream builds a new element tree each time.

### D3. Sprites of the diagram

`internal/uml` gets an optional interface, `SpriteAdder`, with `AddSprite(name string, sprite klimt.Sprite)`. The sprite commands call it when the diagram implements it, and `d.Titled().AddSprite` otherwise, as `HideEmptyDescriptionSetter` already does for `hide empty description`. `PSystemSalt` implements it by storing the sprite in its dictionary only, upstream's `addSprite` override, so the title does not see it (scenario "Sprite in the title").

Alternative: have `SaltDictionary.Sprite` read the sprites of `Base`. The widgets would draw them, but the title would too, which upstream does not do.

### D4. Parsing

- The factory adds the commands in upstream's order: `uml.AddCommonCommands2`, `uml.AddCommonScaleCommands`, `uml.AddTitleCommands`, then `CommandAnything`. `CommandAnything` is a `command.SingleLine` with the pattern `^(.*)$`, any first token, as upstream's `FirstTokens.ANYTHING`, and `NoTrim`, as upstream builds it with `doTrim` set to `false`. It fails with `This is not needed anymore.` when the diagram has no line yet and the trimmed line is `salt`.
- Because the common commands come first, a line such as `Title | "text"` is a title, as upstream reads it (scenario "Row read as a command"). Nothing is done to prevent it.
- `manageSprite` runs at drawing time, as upstream: it drops `hide stereotype` and the lines that start with `skinparam `, reads `scale <n>` with `jfmt.ParseDouble` and sets the scale, and runs the multiline sprite command on a `sprite $` block. The common commands catch these lines before `CommandAnything` in most inputs, but the code is ported so that the rest behaves as upstream's.

### D5. Tokenizer

- `DataSourceImpl` replaces `\|`, `\{` and `\}` with upstream's private-use placeholders `U+E000`, `U+E001` and `U+E002`, cuts each line with Java's `StringTokenizer(s, "|}", true)`, trims each token with `text.Trin`, splits it with the pattern `\{(?:[-+^#!*/]|S-|SI|S)?`, sets the terminators as upstream does, and puts the characters back in each token. Go's `regexp` takes the leftmost alternative first, as Java does, so `{SI` stays one token.
- `klimt` exports `JavaTokenize` and `JavaTokenizeWithDelims`, which `internal/salt` uses for `DataSourceImpl` and `ElementDroplist`. Their callers in `klimt` do not change.
- `peek` past the end panics with `uml.ThrowIndexOutOfBounds`, which an unclosed group reaches, as upstream's `ArrayList.get` throws.

### D6. Element factories

- `ElementFactory` is a Go interface with `Ready` and `Create`. `AbstractElementFactoryComplex` holds its factories in a slice, in the order they were added, and `getNextElement` returns the first ready one, or panics with `uml.Throw(uml.IllegalStateException, <token>)`.
- `createElement` keeps upstream's wiring: one shared tree factory, the complex factories in the order group, scroll area, border layout, tree, each with the simple factories in upstream's order and then all the complex factories. No ready factory panics with `uml.Throw(uml.IllegalArgumentException, "")` (scenario "Content outside a group").
- The `ready` checks keep upstream's tests: Java's `String.matches` becomes an anchored Go pattern (`^[NSEW]=$`, `^<<\w+$`, `^<<\w+>>$`), and the tree marker check is shared by the group and tree factories.
- An unknown image name panics with `uml.ThrowNPE` from `Objects.requireNonNull`, an unknown menu name with `uml.Throw(uml.IllegalArgumentException, "")`, a `*` at the start of a group with the `NullPointerException` of `Cell.mergeLeft`, and an unknown position of a border layout with `uml.Throw(uml.IllegalStateException, "")`.

### D7. Elements

- `Element` is a Go interface with `PreferredDimension(stringBounder, x, y)`, `DrawU(ug, zIndex, dimToUse)` and `MayDrawBeyondPreferredDimension`. `AbstractElement` gives the colors of the context, built with `GetColorOrWhite` and `WithDark`, and `FontConfigurationBlackBlueTrue`.
- `AbstractElementText` counts the characters of a button, a text field or a droplist as upstream: each `<&name>` becomes `00`, then `purgeAllTag` removes the creole tags, then the length is Java's `length`, in UTF-16 code units. The width is the larger of the text width and 8 pixels per character.
- `klimt` gets `SplitterPurgeAllTag` and the `htmlTag` pattern, built from the Splitter patterns as upstream builds it, with the `Provenance` line of `Splitter.java` and a comment that salt uses them. Upstream runs it with `String.replaceAll`, which compiles a plain Java pattern: case-sensitive, and with `%s`, `%q` and `%g` as plain characters. `internal/command/regex` gets `TransformJava`, the rewrite of the Java-only escapes that `Transform` does after expanding the macros, to compile it the same way. The comment of `creole_command.go` stops saying that `purgeAllTag` is not used.
- The widgets keep upstream's geometry and strokes: the button's stroke of 2.5, rounded by 10, the text field's line and ticks, the radio button's ellipses and the checkbox's polygon with `ChangeBack`, the droplist's arrow, and the separators (`NewUStroke(1, 2, 1)` for `..`, two lines for `==`, a stroke of 1.5 for `~~`).
- A tree draws its widgets with a nil `dimToUse`, as upstream passes `null`. A separator there dereferences it: Go panics with the helpful `NullPointerException` message of upstream with `uml.ThrowNPE`.

### D8. Groups, tables and their lines

- `Positionner2` keeps its elements in a slice of pairs in insertion order, upstream's `LinkedHashMap`, and `ElementPyramid` reads the cell of an element from it. `init` sorts the cells with `sort.SliceStable` by `LeftFirst` and `TopFirst`, as Java's stable `Collections.sort`, and keeps upstream's `ensureColWidth` and `ensureRowHeight`.
- `Grid` keeps the horizontal and the vertical segments in two sets in insertion order, without duplicates, and draws them in the order of `klimt.JavaHashOrder` over `Segment.hashCode` (`row * 47 + col`), upstream's `HashSet` order. The title is drawn on a white rectangle, 6 pixels from the left, as upstream does.
- `Grid2` draws the lines of tree tables in the order of upstream's lists.

### D9. Scroll areas, border layouts, trees, tabs and menus

- `ElementPyramidScrolled` draws its bars on each pass, as upstream's `drawU` does not check the z-index (scenario "Scroll area").
- `ElementTree` and `ElementTreeEntry` draw at the z-index 0 only, so an open droplist in a tree does not show its list (scenario "Open droplist in a tree"), and keep upstream's skeleton and widths.
- `ElementMenuBar` keeps its lists in insertion order. Upstream keeps them in a `HashMap` keyed by `ElementMenuEntry`, which has no `hashCode`, so their order follows Java's identity hash codes and is not stable across runtimes: the drawing order of several lists cannot be pinned by a reference. The scenario "Several menus with items" is a Go test.
- `ElementMenuBar` sets the background `#BBB` of the entries with items during the pass 1, as upstream does, and the entries keep it: the next drawing of the same element tree draws them with it. This is what the dry run of an open droplist triggers (scenario "Highlighted menu entry").

### D10. Two passes and the canvas

`getTextBlock` builds the element tree, measures it with the default string bounder of the file format option, and, when `MayDrawBeyondPreferredDimension` is true somewhere in the tree, draws both passes on `klimt.TextBlockUtilsMinMax(..., true)` and grows the size with `AtLeast` to the maximum it found, as upstream's `withActualExtent`. The text block draws the pass 0 then the pass 1 with that size, and its dimension is that size. The drawing of the output goes through this text block once, as upstream's, so the entries of menus are highlighted only after a dry run.

### D11. Errors

- The parse errors are the ones of the commands: `This is not needed anymore.`, and those of the common commands.
- Every failure while drawing panics with `uml.Throw` and the Java class and message of upstream's exception (D5, D6, D7). `uml.ExportTitled` returns it as an error, without output, as for gantt, WBS and timing diagrams. Upstream draws a crash report with a random quote and the Java stack instead, which cannot be a reference.

### D12. Registration and WASM

- `salt.Factory` (`uml.NewFactory(core.Salt, ...)`) goes into `engine.Factories` after `activitydiagram3.Factory` and before `mindmap.Factory`, as `PSystemSaltFactory` comes after `ActivityDiagramFactory3` and before `MindMapDiagramFactory` in upstream's `PSystemBuilder`.
- `core.Salt` joins `supportedTypes`. `TestCheckFamily` gets a `@startsalt` check.
- The text check of the engine does not change: TXT and UTXT of salt diagrams return the unsupported-format error.
- The WASM build needs no special case, because salt diagrams do not use Graphviz.

### D13. Documentation

- The README and `site/index.html` get the row `Salt | yes | yes | no` in the table of supported diagrams. The Graphviz note of the page and the last sentence of the README's Graphviz section name salt diagrams among the diagrams that do not need `dot`. The sentence on the types that are not supported yet names archimate and JSON instead of archimate and salt. The README's conformance counts follow the report.
- The conformance tally of the page follows the release on the page, and the playground note of the page and the README's sentence about the playground follow the WASM module of that release: they do not change.

### D14. Conformance

- Curated cases go under `testdata/conformance/cases/salt-diagram/` with `family: salt` and `type: salt`, one per compared scenario, plus `diagram-source/salt-detected`. The PNG scenario uses `formats: [svg, png]`, the others `[svg]`. The embedded cases are `@startuml` cases (`family: uml`, `type: sequence` for the notes, `type: activity` for the activity). "Dark mode" is a CLI case, `salt-diagram/dark-mode.cli.yaml`, with `--svg --dark-mode` and `compare: [exit, files]`, as `output-formats/dark-mode`. Their references come from `make reference`, and CI renders them again and checks them against the committed files.
- The embedded cases are `@startuml` cases, so the golden dumps of `internal/sequencediagram`, `internal/classdiagram` and `internal/activitydiagram3` are regenerated, as for the WBS change.
- Scenarios that are not compared with Java get `.go.yaml` cases: "Salt diagram without Graphviz" (environment), "Content outside a group", "Unclosed group", "Unknown image" and "Unknown menu" (upstream draws a crash report), "Several menus with items" (D9), "TXT salt diagram" (upstream draws its generic text output) and "Salt diagram in WASM" (environment).
- The cases come before the code. Until the factory is registered, the salt cases report "not supported yet", which the runner counts as unsupported, `TestDiagramReferences` skips the curated salt error case ("Needless salt keyword"), and the embedded cases fail in `make conformance`.
- The 10 Vega salt cases are compared once the factory is registered. A case that still differs at the end is listed in `known-differences.yaml` with its reason.

### D15. Port order

The layers are ported bottom up, each with unit tests: the `klimt` helpers and `SpriteAdder`, the tokenizer and the positioner, the elements, the factories, the diagram, the factory and the export, then the registration.

## Risks / Trade-offs

- **[Order of the lines of a table]** The SVG lists the lines of a group in the order of upstream's `HashSet`, which depends on the capacity of the set as much as on the hash codes. → `klimt.JavaHashOrder` already models the capacity of a default `HashSet`. The scenarios "All table lines", "Outline", "Vertical lines", "Horizontal lines", "Spanning cell" and the Vega tables, which have up to four rows, check it.
- **[Number of drawings]** The highlight of the menu entries depends on how many times the element tree is drawn. A Go export that draws the text block more often than upstream would highlight entries that upstream does not. → The scenarios "Menu with items" (no highlight) and "Highlighted menu entry" (highlight after the dry run) check both cases against upstream.
- **[Order of the menu lists]** Several menus with items differ from upstream in the order of their lists in the SVG, when upstream's identity hash codes put them in another order. → The order is documented in the spec, the scenario is a Go test, and a compared case with several lists, if one is ever added, goes into `known-differences.yaml`.
- **[The dry run of open droplists]** Upstream measures the canvas of an open droplist with `LimitFinder`, which its comment says is not always pixel-identical to the preferred dimension. → The same calls are ported, so the same rounding applies. The three Vega cases of `group2882` and the scenario "Wide open droplist" check the size of the image.
- **[Character count]** The width of buttons, text fields and droplists depends on Java's `length` after `purgeAllTag`. → `length` is counted in UTF-16 code units, and the scenario "Text field width" pins the rule. `purgeAllTag` follows upstream's `htmlTag` pattern, which the Vega case `brace_escape` checks with links.
- **[Creole in cells]** The cells go through the shared creole engine with `SaltDictionary` as the skin, whose values differ from those of the diagrams ported so far (no skin values, upstream's Guillemet, an empty pragma). → The scenario "Creole in a cell" and the Vega cases with links check the drawing. A difference is fixed in `SaltDictionary` or in `internal/klimt`, with the creole golden tests unchanged.
- **[Shared sprite commands]** The sprite commands change for every diagram. → Only `PSystemSalt` implements `SpriteAdder`, and the other diagrams keep calling `Base.AddSprite`. The sprite cases of the other types stay unchanged.
- **[Floating point]** The layout adds and compares `double` values. → The drawing keeps Java's order of operations on `float64`, and the Vega tables and tree tables check it.

## Migration Plan

Nothing to migrate. Inputs that failed with "salt diagrams are not supported yet" now render, or fail with upstream's error. Rollback means reverting the change.
