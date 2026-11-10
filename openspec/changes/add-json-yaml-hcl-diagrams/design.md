# Design

## Context

See proposal.md for the motivation. The current state that shapes the approach:

- Upstream classes and their state in Go:
  - `jsondiagram`: `JsonDiagramFactory`, `StyleExtractor`, `JsonDiagram`, `TextBlockJson`, `SmetanaForJson`, `JsonCurve`, `Arrow` and `Mirror`. None is ported.
  - `yaml`: `YamlDiagramFactory` and `Highlighted`. None is ported. `SimpleYamlParser` and `YamlLines` are not called: the factory's call of `SimpleYamlParser` is commented out.
  - `yaml.parser`: `YamlParser`, `YamlBuilder`, `YamlLine`, `YamlLineType`, `Monomorph`, `MonomorphType`, `MonomorphToJson` and `YamlSyntaxException`, without regular expressions. None is ported. `YamlParser` reads its lines through `utils.Peeker` and `utils.PeekerUtils`, which Go does not have.
  - `hcl`: `HclDiagramFactory`, `HclParser`, `HclSource`, `HclTerm` and `SymbolType`. None is ported.
  - `jsondiagram` imports `yaml.Highlighted`, and `yaml` imports `jsondiagram.JsonDiagram` and `StyleExtractor`: the two packages form a cycle. `hcl` imports both, `yaml.parser` imports neither.
- What Go already has:
  - `internal/json`, the port of upstream's `json` package: `json.Parse`, `json.Value` with `IsObject`, `IsArray`, `IsString`, `IsNumber`, `IsBoolean`, `IsTrue` and `IsNull`, `*json.Object` with its members in source order, `*json.Array`, and numbers that keep their source text (`1.5e3`)
  - `uml.Base` (`TitledDiagram`) with `LoadSkin`, `SetTitle` and the scale, `uml.ExportTitled`, `uml.AddCommonScaleCommands` and `uml.DisplayPositionedSingle`
  - in `internal/klimt`: `SNameJsonDiagram` and `SNameYamlDiagram`, the `yamlDiagram,jsonDiagram` styles of `skin/plantuml.skin`, `StyleBuilder`, `Style.ApplyStrokeAndLineColor`, `FontConfiguration`, `WrapWidth` and `HorizontalAlignment`, `StyleLoaderParseStyleText`, `SkinParam.MuteStyle` and `SetDefaultSkin`, `StyleLoaderMissingRootProperties`, `Display.Create0` and `DisplayWithNewlines`, `TextBlockUtilsWithMargin` and `TextBlockUtilsMinMax`, `URectangle`, `ULine`, `UPath`, `UEllipse`, `UTranslate.Sym`, `FontConfigurationBlackBlueTrue`, `font.Monospace` and `StereotypeBuild`
  - `core.JSON`, `core.YAML` and `core.HCL` with their start tags, and `core.CheckFamily`, which reports every family outside `supportedTypes` as not supported yet
  - the engine's check that rejects TXT and UTXT for the `core.TypedDiagram`s other than sequence, gantt and WBS, and `RenderContext.EmbeddedDiagram`, which renders `{{` blocks with the factories of the engine
  - `dot.Gogv`, whose `Layout` returns gogv's SVG and whose `RenderFormat` merges gogv's warnings into the output, as `dot` does on its standard error
  - `HorizontalAlignment.draw` only as the unexported `drawAligned` of `internal/svek`
- What is missing: the packages above, `Peeker`, `HorizontalAlignment.draw` outside `svek`, and gogv's output in a format other than SVG without the warnings. `UGraphicHandwritten` is ported for no diagram type.
- `SmetanaForJson` lays the boxes out with upstream's Smetana port, never with Graphviz, whatever `!pragma layout` and `GRAPHVIZ_DOT` say. It builds a top-to-bottom graph whose nodes have their width and height swapped, with `shape=record` labels `{_dim_<h>_<w>_|{<P0>_dim_<h0>_<w>_|...}}` (an object) or `<P0>_dim_<h0>_<w>_|...` (an array), and edges `tailport=P<i>`, then turns the result a quarter (`UTranslate.sym()` and `Mirror`) so that the drawing goes from left to right. `_dim_` sets the size of a record field in numbers, a hack of Smetana (`Macro.hackInitDimensionFromLabel`) that Graphviz does not have.
- gogv v0.2.0 has no `record` shape and supports neither `margin` nor `fixedsize` nor `shape=plain`, so a `plaintext` node is padded by 8 points on the left and right and 4 at the top and bottom. It supports labels that are tables of empty cells with `FIXEDSIZE`, `WIDTH`, `HEIGHT` and `PORT`, `tailport`, and the `json0` output, which has the center, width and height of every node and, for every edge, its points and end point (`pos="e,<x>,<y> <x1>,<y1> ..."`).
- In a spike, the graph of `vega/nonreg/json/style-highlight-with-style-tags` and a tree of 42 nodes (an array of 20 objects with 2 items each) were given to gogv with the tables of D3. The boxes were within 3 points of upstream's, the drawings within 0.5% of upstream's size, and the order of the boxes was the same. The arrows started 4 to 5 points inside their box, and the ports of the first and last rows moved 4 points inwards.
- `PSystemBuilder` registers `JsonDiagramFactory` after `TimingDiagramFactory`, then `YamlDiagramFactory` and `HclDiagramFactory`, and `ChenEerDiagramFactory` after them.

## Goals / Non-Goals

**Goals:**

- Port the `jsondiagram`, `yaml`, `yaml.parser` and `hcl` packages, with upstream's output for every scenario of the specs.
- Lay the boxes out with gogv within the size and similarity thresholds of the conformance suite.
- Keep the results of every other diagram type unchanged. The shared code that changes is the factory list, the supported families and a new method of `dot.Gogv`.

**Non-Goals:**

- `UGraphicHandwritten`: no diagram type of this release draws by hand, so the drawing ignores `skinparam handwritten true`.
- `UGraphicTxt`, which upstream's TXT and UTXT of these diagrams use: the follow-up that adds the TXT of class and activity diagrams needs it too.
- Upstream's Smetana port (`gen/` and `smetana/`, about 40,000 lines), which only these diagrams, the git diagrams and `!pragma layout smetana` use.
- `SimpleYamlParser` and `YamlLines`, which upstream does not call.
- `PSystemBuilderJs`, the factory list of upstream's TeaVM build.
- The dimension cache of `JsonDiagram.calculateDimension` per `StringBounder` class (`@PerformanceIssue`): the Go diagram lays its boxes out once per export (D4).

## Decisions

### D1. Package layout

- `internal/jsondiagram` holds `jsondiagram` and, as the cycle requires, the two classes of `yaml` (AGENTS.md); `doc.go` explains the prefixes: a file without a prefix ports `jsondiagram`, a `yaml_` file ports `yaml`.
  - `factory.go`: `JsonDiagramFactory`; `styleextractor.go`: `StyleExtractor`; `diagram.go`: `JsonDiagram`; `textblockjson.go`: `TextBlockJson`; `smetanaforjson.go`: `SmetanaForJson` (D3); `jsoncurve.go`: `JsonCurve`; `arrow.go`: `Arrow`; `mirror.go`: `Mirror`
  - `yaml_factory.go`: `YamlDiagramFactory`; `yaml_highlighted.go`: `Highlighted`
- `internal/yamlparser` holds `yaml.parser`: `parser.go` (`YamlParser`), `builder.go` (`YamlBuilder`), `line.go` (`YamlLine`), `linetype.go` (`YamlLineType`), `monomorph.go` (`Monomorph`, `MonomorphType`), `monomorphtojson.go` (`MonomorphToJson`), `syntaxexception.go` (`YamlSyntaxException`), and `peeker.go`, unexported, with the `Provenance` lines of `utils/Peeker.java` and `utils/PeekerUtils.java`.
- `internal/hcl` holds `hcl`: `factory.go` (`HclDiagramFactory`), `parser.go` (`HclParser`), `source.go` (`HclSource`), `term.go` (`HclTerm`, `SymbolType`).
- `jsondiagram` imports `yamlparser`, and `hcl` imports `jsondiagram`, which exports what `HclDiagramFactory` uses: the diagram constructor, `StyleExtractor` and `Highlighted`.

Alternative: one package for the four upstream packages. AGENTS.md merges only packages that import each other in a cycle; `yaml.parser` and `hcl` keep packages of their own, where their `Provenance` lines point.

### D2. Factories and the diagram

- Each factory is a value that implements `core.Factory` with its diagram type, as upstream's `PSystemAbstractFactory`, without commands. `CreateSystem` follows upstream: `StyleExtractor` reads the lines of the source; the first line (the start line) and the last (the end line) are skipped; the JSON factory drops the `#` lines that are not highlight paths and parses the rest with `json.Parse`, the YAML factory reads highlight paths and gives the rest to `yamlparser`, and the HCL factory gives every line to `HclSource`. A parse that fails gives a nil root.
- The parsers return errors where upstream throws: `YamlSyntaxException`, the `IllegalStateException`s and `UnsupportedOperationException`s of `YamlParser`, `Monomorph` and `HclParser`, and the end of the input in the middle of an HCL block, where upstream's iterator throws `NoSuchElementException`. Upstream's YAML and HCL factories catch every `Exception`, so every error gives a nil root. The JSON factory catches only `ParseException`, which is the error of `json.Parse`.
- `JsonDiagram` embeds `uml.Base` with the type `core.JSON`, `core.YAML` or `core.HCL`, so it is a `core.TypedDiagram`: the syntax check reports the type and the description (`(Json)`, `(Yaml)`, `(HCL)`), and the engine rejects TXT and UTXT with `TXT output is not supported for json diagrams` without a change. It wraps a scalar root in an array of one item, and an empty object or array root in an array of one empty string, as upstream's constructor does.
- `scale`: the constructor runs the commands of `uml.AddCommonScaleCommands` on the `scale` line, as upstream. `title`: the JSON and YAML factories set it with `uml.DisplayPositionedSingle`, centered. Styles: `applyStyles` loads the `skin` with the checks of upstream, then mutes the style blocks with `StyleLoaderParseStyleText`; its errors are dropped, as upstream logs them. The HCL factory calls neither, as upstream's code for them is commented out.
- `StyleExtractor` records `skinparam handwritten true`, which the drawing ignores (Non-Goals).

### D3. Layout through gogv

`SmetanaForJson` keeps upstream's walk of the values (`manageOneNode`): it creates the node of a box, then for each row whose value is an object or an array the subtree of that value and the edge to it. It writes this graph as DOT for gogv, in the order of creation:

- the graph: `digraph g {`, without attributes, as Smetana's defaults are Graphviz's
- a node `N<i>` per box: `shape=plaintext`, `width` the height of the box and `height` its width, in inches, as upstream swaps them, and a table label of empty cells (`BORDER`, `CELLBORDER`, `CELLSPACING` and `CELLPADDING` 0). Along the rows, the cells of the ports `P0` ... `P<n-1>` have the heights of the rows, rounded so that their sum follows the sum of the heights (`round(sum of the first k+1) - round(sum of the first k)`). For an object, a first row has one cell of `COLSPAN` n and of the width of the key column; the port cells are as wide as the value column. For an array, the port cells are as wide as the box.
- compensation of the padding of `plaintext`: the first and the last port cells are 8 points shorter, and a single port cell 16; for an object, the key and value cells are 4 points narrower, and for an array the port cells 8. The padded node then has the size of upstream's record, so the ranks and the gaps are Smetana's. All the compensation is in one function, which goes away when gogv supports `margin` (Risks).
- a box without rows (an empty object or array that is a value) has `label=""` and the size of upstream's empty node: 15 points along the rows and 16 across, measured on upstream's output in the spike (Open Questions)
- an edge per link: `N<a> -> N<b> [tailport=P<i>, arrowsize=.75, arrowtail=none, arrowhead=normal]`, written when upstream creates it, after the subtree of its value

The layout is read from gogv's `json0` output with `encoding/json`: the `pos`, `width` and `height` of each node, and for each edge, found by its tail, head and tail port, the points and end point of its `pos`. An edge without exactly one spline is an error, as upstream's `JsonCurve` throws. From there, the drawing follows upstream: `Mirror` of the largest top of a node, the position of each box `(x - w/2, mirror(y + h/2))` turned by `Sym`, and `JsonCurve` and `Arrow` on the turned points.

Alternative: ask gogv for `margin=0` or `shape=plain` first (a gogv change and release before this change). The compensation gives the same ranks and gaps now, and the gogv feature can replace it later.

Alternative: port Smetana for these diagrams. It is about 40,000 lines of generated C-like Java for one layout that gogv already does within a few points.

Alternative: read gogv's SVG, as svek does. The `json0` output has the end points of the edges, which upstream's arrows need, and no node needs to be found by its color.

### D4. Engine of the layout

- The layout always runs gogv, also when `--dot-path`, the library option or `GRAPHVIZ_DOT` names a `dot`, as upstream never runs Graphviz for these diagrams; the drawing then does not depend on the host. The timeout is the one of the render (`RenderContext.DotTimeout`).
- `dot.Gogv` gets `Output(ctx, src, format) ([]byte, error)`: gogv's output for the format, without its warnings, with the timeout and the panic recovery of `Layout`. The diagram calls it with `json0`.
- `JsonDiagram` implements `ExportDiagramContext`: it lays the boxes out first, then draws them through `uml.ExportTitled`. A canceled context returns the context's error, and a layout that fails or times out returns its error, as the graph diagrams do; upstream's in-process Smetana has no such failure. The dimension of the diagram is the extent of its drawing (`TextBlockUtilsMinMax` with `initToZero`), as upstream's `getMinMax`.

Alternative: the engine of `svek.LayoutFromContext`, which runs a named `dot`. The output would then depend on the Graphviz of the host, where upstream's does not.

### D5. Drawing

- `TextBlockJson` is ported with its rows, its minimum size of 30 by 15 points, and its texts built with `Display.Create0` without creole, with margins of 5 and 2. `HorizontalAlignment.draw` is an unexported function of `jsondiagram` with the `Provenance` line of `HorizontalAlignment.java`, as `svek`'s `drawAligned` is unexported.
- The image for a nil root is `Your data does not sound like <TYPE> data` (`JSON`, `YAML` or `HCL`, the name of the diagram type), in `font.Monospace(14)` with `FontConfigurationBlackBlueTrue` and margins of 5 and 2.
- The style queries `NODE`, `NODE_HEADER`, `NODE_HIGHLIGHT`, `NODE_HEADER_HIGHLIGHT`, `NODE_SEPARATOR` and `ARROW` of `StyleQueries`, with `jsonDiagram` for a JSON diagram and `yamlDiagram` for the others, are unexported helpers of `jsondiagram`, as the other diagram packages keep their own.

### D6. Highlighted rows

`Highlighted` is ported with its pattern (`^([^<>]+)(<<.*>>)?$`, valid in Go's `regexp`), `upOneLevel` and `isKeyHighlight`, and its stereotype from `klimt.StereotypeBuild`. Quotes are removed from each key as upstream's `eventuallyRemoveStartingAndEndingDoubleQuote` does.

### D7. Registration

- `jsondiagram.Factory`, `jsondiagram.YamlFactory` and `hcl.Factory` go into `engine.Factories` after `timing.Factory` and before `cheneer.Factory`, in upstream's order.
- `core.JSON`, `core.YAML` and `core.HCL` join `supportedTypes`. `TestCheckFamily` and `TestCreateDiagramFamilies` in `internal/core` take `@startnwdiag` as the unsupported family, and `TestEmbeddedUnsupportedType` in `internal/engine` takes `{{nwdiag`.
- Embedded `{{json` and `{{yaml` blocks need no code: `RenderContext.EmbeddedDiagram` renders them with the engine's factories.

### D8. Conformance

- Curated cases go under `testdata/conformance/cases/json-diagram/`, `yaml-diagram/` and `hcl-diagram/`, with `family` and `type` `json`, `yaml` or `hcl` and the front matter of their neighbors, one per compared scenario, with the source lines of the scenario. The wide-array case writes its twenty objects out. The syntax checks are CLI cases (`-syntax`, standard input, `compare: [exit, files, stdout]`), as `dot-diagram/syntax-check`.
- `diagram-source/json-detected`, `yaml-detected` and `hcl-detected` cover the new family scenarios. `styling/embedded-json-in-note` and `styling/embedded-yaml-in-note` cover the embedded diagrams; they are `@startuml` cases, so the golden dumps are regenerated.
- `.go.yaml` cases cover what is not compared with upstream: "json-diagram: TXT of a JSON diagram" (upstream draws TXT), the three WASM scenarios, and the updated `diagram-source/unsupported-family` and `styling/embedded-unsupported-type` (their `why` names `@startnwdiag` and `{{nwdiag`).
- The report groups the cases as `json`, `yaml` and `hcl`, the types of their references. `vega/nonreg/json/style-highlight-with-style-tags` is compared and must pass; `vega/nonreg/yaml/theme-with-empty-style-block` stays a known difference for the amiga theme.
- A case that fails the thresholds is fixed in the layout of D3. It is listed in `known-differences.yaml` only with the measured difference as its reason, and with `resolvedBy` naming the gogv issue on `margin` when that is the cause.
- The cases come first. Until the factories are registered, they report "not supported yet", which the runner counts apart.

### D9. Documentation

- The README and `site/index.html` get the rows `JSON`, `YAML` and `HCL`, each `yes | yes | no`, after `Dot` in the table of supported diagrams. The sentence on the other diagram types names nwdiag and EBNF instead of JSON and YAML. The note of the page on the layout without Graphviz names JSON, YAML and HCL among the diagrams that gogv lays out.
- The README's conformance counts follow the report.

### D10. Port order

Cases first (D8). Then `dot.Gogv.Output`; then `jsondiagram`: `Highlighted`, `StyleExtractor`, `TextBlockJson`, `Mirror`, `Arrow`, `JsonCurve`, `SmetanaForJson`, `JsonDiagram` and the JSON factory, registered with `core.JSON`; then `yamlparser` and the YAML factory; then `hcl`; then the documentation.

## Risks / Trade-offs

- **[Layout of other trees]** The spike covered two trees; Smetana and gogv may differ more on others. → Every curated case checks the thresholds, and the wide-array case checks a column of twenty boxes. The gogv issue on `margin` removes the compensation and the moved ports.
- **[Arrows]** The arrows start 4 to 5 points inside their box, and the ports of the first and last rows are 4 points off, as the cells of the ports are inside the padding. → Kept: the spike stayed within the thresholds. A case that fails because of it is a known difference resolved by the gogv issue.
- **[Rounded cells]** Table cells have whole sizes, upstream's record fields do not. → The cumulative rounding keeps each row within half a point of its place.
- **[Errors as values]** A parser path that panics where upstream throws would crash the render, where upstream draws the "does not sound like" image. → The parsers return errors for every exception of upstream, and the unit tests feed them the failing inputs of the specs.
- **[Shared code]** The factory list, the supported families and `dot.Gogv` change. → Only `@startjson`, `@startyaml`, `@starthcl` and the embedded blocks reach the new code, and `make conformance` must give the other types the same results.
- **[Benchmark]** The new cases are error images on `main` and drawings on the branch, so the benchmark of the pull request compares them with different work. → Compare the cases that exist on `main` before calling a difference a regression.

## Migration Plan

Nothing to migrate. Inputs that failed with "... diagrams are not supported yet" now draw upstream's diagram. Rollback means reverting the change.

## Open Questions

- The exact size of upstream's empty node, measured as 15 by 16 points in the spike, is checked on the references of the empty object and array cases; it changes a constant of D3, not the approach.
