# Design

## Context

See proposal.md for the motivation. The current state that shapes the approach:

- Upstream's `directdot` package has `PSystemDotFactory`, a `PSystemBasicFactory` that reads the block line by line, `PSystemDot`, a `DirectOsDiagram` whose `exportDiagramNow` writes what the layout engine returns, and `CounterOutputStream`, which counts the bytes written. `PSystemDot` asks `GraphvizRuntimeEnvironment.createForSystemDot` for a Graphviz and runs it with `AbstractGraphviz.createFile3`, whose `ProcessRunner` merges the standard error of `dot` into its output (`redirectErrorStream(true)`) and reports `TERMINATED_OK` whatever the exit status of `dot`. `SecurityProfile.allowDotSvg` is true for `INSECURE` and `INTERNET_WITH_DOTSVG` only.
- What Go already has:
  - the header pattern of `PSystemDotFactory` (`graphvizDotHeader` in `internal/core/builder.go`, whose `Provenance` already names `PSystemDotFactory.java`), used only by the "Please use @startdot" check of `@startuml` blocks
  - `core.Dot`, its start tag `@startdot` and its name `dot` (`HumanReadableName`), and `core.CheckFamily`, which reports every family outside `supportedTypes` as not supported yet
  - `command.Factory`, the port of `PSystemCommandFactory` and `PSystemAbstractFactory`, with `command.EmptyDescription`
  - `UmlSource` (`internal/core/umlsource.go`) without `removeInitialNoise` and `isStartDef`
  - `dot.Exec` with `createFile3(ctx, dot, redirectErrorStream, types...)` and the `processRunner` that keeps upstream's states, `dot.Gogv`, which calls `gogv.Render` for `svg` only, the `dot.Engine` interface (`Layout`, SVG only), and `svek.Layout`, whose `resolve` chooses the engine of a render: the named `dot` when the platform can run programs (`execSupported`), gogv otherwise
  - `security.Profile` with `InternetWithDotSVG`, without `allowDotSvg`
  - `klimt.GraphicStringsCreateForError`, `export.TextBlockExporter` with its status and seed, and `export.ImageDataSimple`
  - `engine.ContextDrawableDiagram`, which lets a diagram write its pages to an `io.Writer`, and the engine's check that rejects TXT and UTXT for `core.TypedDiagram`s other than sequence, gantt and WBS
- What is missing: `PSystemBasicFactory`, `PSystemDotFactory` itself, `PSystemDot`, `removeInitialNoise`, `allowDotSvg`, and a way to run the layout engine of a render for a format and get its merged output.
- gogv v0.2.0 writes `svg` and the text formats of `dot`, not `png` (ekalinin/gogv#33). It returns its warnings apart, each telling whether it is about a Graphviz feature it does not support yet (`NotSupportedYet`), and a `*gogv.SyntaxError` worded as Graphviz words it. Its own `dot` command writes the warnings as `Warning: <text>` lines and the error as `Error: <stdin>: <message>`, as Graphviz does.
- The conformance harness renders every curated case in upstream's default profile, `LEGACY`: a case cannot name a profile today. `ReferenceRender` reads all the cases of a suite in one JVM, and upstream reads `PLANTUML_SECURITY_PROFILE` once and keeps the profile in a static field (`SecurityUtils.current`).
- In the reference image, Graphviz 14.0.1 has no PNG renderer. Upstream's PNG of a dot diagram there depends on timing: `dot -Tpng` exits at once, and upstream either fails to write the DOT source to it (the "GraphViz has crashed" image, status 503) or writes it before `dot` exits and then writes `dot`'s `Format: "png" not recognized. Use one of: ...` text as the PNG (status 0). Three runs in a row gave the image once and the text twice.

## Goals / Non-Goals

**Goals:**

- Port `PSystemBasicFactory`, `PSystemDotFactory`, `PSystemDot` and `CounterOutputStream`, with upstream's output for every profile, engine outcome and error.
- Compare the layout of gogv with the layout of upstream's Graphviz for dot diagrams, through cases rendered in `INTERNET_WITH_DOTSVG`.
- Keep the results of every other diagram type unchanged. The shared code that changes is the factory list, the supported families, the TXT check of the engine, and additions to `core`, `security`, `dot`, `svek` and the conformance harness that the other diagrams do not call.

**Non-Goals:**

- `GraphvizJs` and the factory hooks of `GraphvizRuntimeEnvironment.createForSystemDot` (`createWithFactory`, `useVizJs`): the Go build has gogv instead of viz.js, and nothing installs a factory.
- `PSystemBasicFactory.explain`, which throws upstream ("wip"): `@Explain` is reported as not supported yet before the factories run.
- The statistics of `DirectOsDiagram.exportDiagram` (`ENABLE_STATS`), which no Go code collects.
- The formats other than SVG and PNG, see proposal.md.

## Decisions

### D1. Package layout

- `internal/directdot` holds the package, with upstream's name and a `Provenance` line for each Java file it ports: `doc.go`, `factory.go` (`PSystemDotFactory`) and `psystemdot.go` (`PSystemDot`, `CounterOutputStream`, and `DirectOsDiagram`, whose only behavior is to call `exportDiagramNow`).
- `internal/command/basicfactory.go` ports `PSystemBasicFactory` as `BasicFactory`, with the `InitDiagram` and `ExecuteLine` functions of the factory that uses it. Its diagrams are `core.Diagram`s, not a type parameter, so that "no diagram yet" is a plain nil and not a nil pointer in an interface. Upstream's math and LaTeX factories build on it too.
- `internal/core/umlsource.go` gets `RemoveInitialNoise` and `IsStartDef` (`UmlSource.removeInitialNoise`, `isNoise` and `isStartDef`). The header pattern stays in `internal/core/builder.go` and is exported as `IsGraphvizDotHeader` (`PSystemDotFactory.isGraphvizDotHeader`), which the misuse check and the factory use: `directdot` imports `core`, so `core` cannot use a pattern of `directdot`.
- `internal/security/security.go` gets `Profile.AllowDotSVG` (`SecurityProfile.allowDotSvg`).
- `internal/dot` gets the `Renderer` interface, `Exec.RenderFormat` in `graphviz.go` and `Gogv.RenderFormat` in `gogv.go` (D4). `internal/svek/cucadiagramfilemaker.go` gets `Layout.RenderFormat` (D4).

Alternative: the factory and the diagram in `internal/core`, next to the header pattern. Upstream has a package of its own, and `core` holds the types shared by all diagrams, so a `directdot` package keeps the port where its `Provenance` lines point.

### D2. Reading the block

- `BasicFactory.CreateSystem` follows upstream: it calls `RemoveInitialNoise` (the lines directly after the start line that are empty or start with `skinparam `, `skinparamlocked ` or `!pragma ` are dropped), skips the empty lines that come first, and calls `ExecuteLine` for each line up to the end line. At the end line, a source of two lines that does not start with `@startdef` fails with `command.EmptyDescription` at the end line, as `buildEmptyError`. A nil diagram from `ExecuteLine` fails with `Syntax Error?`, a syntax error of score 0 at that line with the lines read so far as the trace. Both errors carry the factory's diagram type, so the message ends with `(Assumed diagram type: dot)`. A block of blank lines that are not empty (`@startdot`, `   `, `@enddot`) reaches its end line without a diagram: upstream then throws an `IllegalStateException` while merging the errors, and `BasicFactory` returns nil, which the engine reports as an unsupported diagram.
- `PSystemDotFactory` keeps the DOT source in the diagram, not in the factory: upstream's `data` field is state of the factory, which Go factories cannot have, as concurrent renders share them. `InitDiagram` returns nil. `ExecuteLine` returns a new diagram with the line and `\n` when there is no diagram yet and the line is a graph header, nil when there is no diagram, and otherwise a diagram whose source is the previous one plus the line and `\n`. The output is upstream's.

### D3. The diagram

- `PSystemDot` holds the DOT source, the security profile of the render (`rc.Security.Profile`, or `security.ParseProfile(rc.SecurityProfile)`, as the engine reads it), the layout of the render (`svek.LayoutFromContext`) and the seed of the source (upstream's `seed()`, the hash of the source lines that `uml.Base.Seed` computes).
- Its description is `(Dot)` and it has one page. It is not a `core.TypedDiagram`, as upstream's `DirectOsDiagram` is not a `TitledDiagram`: the syntax check prints `OTHER`, as upstream, and the conformance meta records the type `OTHER`, as the reference does.
- `ExportDiagramContext` follows `exportDiagramNow`:
  1. For SVG, when the profile does not allow it, it writes upstream's suppressed-feature document, byte for byte, and returns an image of size 0 with status 0, as `ImageDataSimple.ok()`.
  2. It asks `Layout.RenderFormat` for the format, `svg` or `png` (upstream's lower-case format name). Other formats return `export.NotImplementedError`, which the engine reports as not drawable.
  3. A `*dot.NotFoundError` gives the error image `There is an issue with your Dot/Graphviz installation`. A canceled context returns the context's error, as the graph diagrams do. Any other error, or an empty output, gives the error image `GraphViz has crashed` (upstream's `counter.getLength() == 0 || state.differs(TERMINATED_OK)`).
  4. Otherwise it writes the output unchanged, without metadata, and returns status 0.
- The error images go through `export.NewBuilder(klimt.GraphicStringsCreateForError(lines, false), fo, false)` with the seed and status 503 (upstream's `FileImageData.CRASH`), without metadata, as upstream's `TextBlockExporter.builder` without `.metadata(...)`. The color mapper of the format option applies, so `--dark-mode` darkens them as upstream; it does not change the output of the layout engine, which upstream writes as is.
- `CounterOutputStream` is not a type of its own: `Layout.RenderFormat` returns the whole output, whose length is the count. Its `Provenance` line goes in `psystemdot.go`.

### D4. Running the layout engine for a format

- `dot.Renderer` is `RenderFormat(ctx, src, format string) ([]byte, error)` (the name `Render` is taken by the field of `dot.Gogv` that tests use to replace gogv): the output of the engine for `-T<format>`, with its standard error merged, as upstream's `createFile3` writes it.
- `Exec.RenderFormat` runs `createFile3(ctx, src, true, format)`. A terminated `dot` gives its merged output and no error, whatever its exit status, as upstream. A timeout gives a `*TimeoutError`, a failure to run it a `*NotFoundError` (missing file) or a `*FailureError`, and a canceled context the context's error, as `Exec.Layout`.
- `Gogv.RenderFormat` calls `gogv.Render` with the timeout and the panic recovery of `Gogv.Layout`, and writes what gogv's `dot` would write on its merged output: a `Warning: <text>` line for each warning that is not `NotSupportedYet`, then the output, then `Error: <stdin>: <message>` for a `*gogv.SyntaxError`, which is not an error of `Render`. A `*gogv.FormatError` (`png`) and the other errors give a `*FailureError`, and a timeout a `*TimeoutError`.
- `svek.Layout.RenderFormat` chooses the engine as `resolve` does: the injected `Engine` when it is a `dot.Renderer` (tests), the named `dot` through `Config.NewEngine` when `execSupported`, with its `*NotFoundError` returned as is, and gogv otherwise, so the WASM build uses gogv.

Alternative: add `RenderFormat` to `dot.Engine`. Every fake engine of the svek tests would have to implement it, and the graph diagrams never call it, so a separate interface keeps them unchanged.

Alternative: return gogv's warnings apart and let `PSystemDot` write them. Merging is what `dot` does with its standard error, so the engine owns it, and `Exec.RenderFormat` and `Gogv.RenderFormat` then return the same kind of output.

### D5. PNG without a named dot

Without a named `dot`, gogv returns a `*gogv.FormatError` for `png`, and the diagram draws the `GraphViz has crashed` image with status 503, so the command line exits with 200. This is one of the two outputs of upstream with the Graphviz of the reference image (see Context). The other, `dot`'s `Format: "png" not recognized` text written as a PNG with status 0, is not reproduced: it depends on timing, and a PNG file that holds text is not an image.

Alternative: write gogv's `FormatError` text as `dot` writes it, which upstream gives in most runs. It would give a "PNG" that holds text and a success status, so the crash image, which upstream also gives, is kept.

### D6. Security profile

`Profile.AllowDotSVG` is true for `Insecure` and `InternetWithDotSVG`. The defaults stay: `LEGACY` for the CLI and the library, `INTERNET` for the server, `SANDBOX` for the WASM build. PNG does not depend on the profile, as upstream checks it for SVG only.

### D7. Registration

- `directdot.Factory` goes into `engine.Factories` after `salt.Factory` and before `mindmap.Factory`, as `PSystemDotFactory` comes after `PSystemSaltFactory` and before `MindMapDiagramFactory` in upstream's `PSystemBuilder`.
- `core.Dot` joins `supportedTypes`. `TestCheckFamily` gets a `@startdot` check.
- `engine.Block.Render` rejects TXT and UTXT for a `*directdot.PSystemDot` with `UnsupportedFormatError{Type: core.Dot}` (`TXT output is not supported for dot diagrams`), as `PSystemDot` is not a `core.TypedDiagram`, which the current check needs.

### D8. Security profile of conformance cases

- `conformance.Case` gets `Profile` (`profile:` in the front matter), which `ParseCase` checks with `security.ParseProfile`.
- With the Java implementation, `refgen` writes one manifest per profile in `renderDiagrams`, and `run.sh` runs `ReferenceRender` once per manifest, the named profiles with `PLANTUML_SECURITY_PROFILE=<profile>` set for that JVM only. A JVM per profile is needed as upstream keeps the first profile it reads.
- With the Go implementation, `renderCaseGo` sets `rc.SecurityProfile` to the case's profile, and so does `TestDiagramReferences` of `internal/errdiagram`.
- The Vega suite has no profile. CLI cases keep setting `PLANTUML_SECURITY_PROFILE` through `env`.

Alternative: CLI cases with `env: {PLANTUML_SECURITY_PROFILE: INTERNET_WITH_DOTSVG}`. They compare the SVG texts only, not the size and the similarity, so they would not check gogv's layout against Graphviz's.

### D9. Conformance

- Curated cases go under `testdata/conformance/cases/dot-diagram/` with `family: dot` and `type: dot`, one per compared scenario of `dot-diagram`, with the profile its scenario names: `INTERNET_WITH_DOTSVG` for the drawn graphs, `INSECURE` and `INTERNET` for the profile scenarios, none for the others. "Syntax check of a dot diagram" is a CLI case (`-syntax`, standard input, `compare: [exit, files, stdout]`). `diagram-source/dot-detected` covers "diagram-source: Dot family".
- Two cases without scenarios compare the examples of upstream's dot page in `INTERNET_WITH_DOTSVG`: `digraph foo {`, ` PlantUML -> Dot [label=use];`, `}`; and the graph with `node [style=rounded]`, `shape=box`, `fillcolor=yellow, style="rounded,filled", shape=diamond` and `shape=record, label="{ a | b | c }"`, which gogv draws with its defaults and which is listed in `known-differences.yaml` with the gogv features it lacks.
- `.go.yaml` cases cover the scenarios that are not compared with Java: the named `dot` and the missing, slow or PNG-writing `dot` (environment), the built-in engine and its dropped warnings (the reference uses Graphviz), the syntax error and the Graphviz warning (the output is not an SVG document, which the comparison of SVG texts needs), the PNG without a named `dot` (D5: upstream's output is not reproducible), TXT (upstream asks `dot` for `-Tatxt`), WASM, and "upstream-compatibility: Case rendered in a security profile" (behavior of the harness).
- No `@startuml` case is added or changed, so the golden dumps stay as they are.
- The cases come before the code. Until the factory is registered, the dot cases report "not supported yet", which the runner counts as unsupported, and `TestDiagramReferences` skips the two dot error cases.

### D10. Documentation

- The README and `site/index.html` get the row `Dot | yes, in the INTERNET_WITH_DOTSVG and INSECURE profiles | with a named dot | no` at the end of the table of supported diagrams.
- The README's Graphviz section and the Graphviz note of the page say that dot diagrams are drawn by the layout engine as upstream does, that their SVG needs the `INTERNET_WITH_DOTSVG` or `INSECURE` profile and is upstream's suppressed-feature notice otherwise, and that their PNG needs a named `dot`, as gogv does not write PNG yet.
- The README's conformance counts follow the report. The conformance tally of the page and its playground note follow the release on the page, so they do not change.

### D11. Port order

Cases first, with the `profile` of the harness that their references need. Then `RemoveInitialNoise`, `IsGraphvizDotHeader` and `AllowDotSVG`, the `RenderFormat` of the engines and of `svek.Layout`, `BasicFactory`, the `directdot` package, and the registration.

## Risks / Trade-offs

- **[Layout of gogv]** gogv follows Graphviz 16.1.0 and the references come from Graphviz 14.0.1. → The edges of the simple graphs of the scenarios differ by hundredths of a point, within the tolerance. A case that differs more is listed in `known-differences.yaml` with its reason.
- **[Output that is not SVG]** With a Graphviz warning or invalid DOT, the SVG output starts with or is a line of text, as upstream's. → Kept as upstream does. These outputs are checked by unit tests, as the harness compares SVG texts.
- **[Reproducible references]** Only the PNG of dot diagrams depends on timing upstream. → No compared case asks for it (D5, D9). The SVG cases do not run `dot` for PNG.
- **[Shared code]** The factory list, the supported families and the TXT check of the engine change for every diagram. → Only `@startdot` blocks reach the new factory, the TXT check adds a case for `PSystemDot` only, and the results of the other types must stay the same in `make conformance`.
- **[Harness]** The reference run starts one more JVM per profile. → Three profiles, a few seconds per run.

## Migration Plan

Nothing to migrate. Inputs that failed with "dot diagrams are not supported yet" now give upstream's output. Rollback means reverting the change.
