# Design

## Context

See proposal.md for the motivation. The current state that shapes the approach:

- Upstream's `packetdiag` package has `PacketDiagram` (with the nested `PacketItem` and `ScaleDirection`), `PacketDiagramFactory`, `PacketBlock`, `PacketIndicator` and the commands `CommandPacketDiagStart`, `CommandPacketDiagEnd`, `CommandColWidth`, `CommandNodeHeight`, `CommandScaleDirection`, `CommandScaleInterval`, `CommandSameHeight` and `CommandNumRange`. It also uses two classes of `nwdiag`: `NwDiagram.toSet` (the `name = value` attribute parser) and `VerticalLine`.
- `PacketDiagram` is a `TitledDiagram` that draws through `getTextBlock`, like `ActivityDiagram3`. The Go side already has what that needs: `uml.Base`, `uml.NewFactory`, `uml.AddCommonCommands1`, `uml.ExportTitled`, `klimt.SNamePacketdiagDiagram` (also used by `uml/chrome.go` for the title and other annotations), `klimt.UGraphicStencilCreateDim`, `URectangle.Rounded`, `UPath.ArcTo`, `TextBlockMemoized`, `TextBlockUtilsMergeTB`, and the `uml.Throw` helpers with the Java exception names (`uml.ArithmeticException`, ...).
- `core.FindStartTypes` already maps `@startpacketdiag` to `core.Packet`, and `HumanReadableName` gives `packet`. `core.CheckFamily` reports every family other than those of `supportedTypes` (`Sequence`, `Class`, `Activity`) as not supported yet.
- `core.CreateDiagram` gives each factory only the blocks whose types include its type, so a packet factory sees `@startpacketdiag` blocks only, and the `@startuml` factories never see them.
- The embedded diagram code (`klimt/creole_legacy.go`) already accepts `{{packetdiag` and renders it as a `@startpacketdiag` block through the engine.
- The engine rejects TXT and UTXT for every type other than sequence (`UnsupportedFormatError`).
- In the 62 Vega packet cases, upstream fails in three ways: a parse-time `ArithmeticException` in `adjustIndicators` (`R07.02`, `R11.01`), which `PSystemBuilder` turns into a "Fatal crash error" error diagram whose line is the last line of the block; a drawing-time exception for a negative width (`R07.01`, `error-with-exception`), which upstream writes as a crash report with status 503; and a syntax error (`R23.01`). The runner already accepts any failure of the implementation when the reference is a crash report.
- `core.createSafely` turns a panic of a factory into the "Fatal crash error" error diagram with the whole block as its trace. No corpus case has reached it yet.
- `TestDiagramReferences` of `internal/errdiagram` draws the error image of every curated case whose reference is an error, so a curated packet error case fails `make test` while the factory is not registered.
- `landing_test.go` requires the table of supported diagrams of `site/index.html` to have the rows of the README's table.

## Goals / Non-Goals

**Goals:**

- Port `packetdiag` file by file, with the drawing byte-identical to upstream's SVG, apart from the encoded source.
- Keep the results of every other diagram type unchanged. The only shared code that changes is the list of supported families and the factory list.

**Non-Goals:**

- Porting `nwdiag`. Only the two helpers that `packetdiag` calls are ported.
- A crash report image. Go has none for any diagram type: a drawing that panics returns an error.

## Decisions

### D1. Package layout mirrors upstream

`internal/packetdiag` holds `diagram.go` (`PacketDiagram`, `PacketItem`, `ScaleDirection`), `factory.go`, `block.go` (`PacketBlock`), `indicator.go` (`PacketIndicator`), `export.go`, `commands.go` (the start and end commands and the five directives), `commands_numrange.go` (`CommandNumRange` and its attribute helpers) and `nwdiag.go` (`toSet` and `VerticalLine`), each with the `Provenance` lines of the Java files it ports. The `nwdiag` helpers are unexported.

Alternative: an `internal/nwdiag` package for `toSet` and `VerticalLine`. It would be a package for two helpers with one user. They move there when `nwdiag` is ported.

### D2. PacketDiagram on top of uml.Base

- `PacketDiagram` embeds `uml.Base` with the type `core.Packet`, as `ActivityDiagram3` does, so the SVG has `data-diagram-type="PACKET"` and the annotations, skins, styles and margins come from the shared code. `Description` returns `Packet Diagram`.
- `CheckFinalError` runs upstream's `build` (default scale interval, `adjustLayout`, `adjustColWidth`, `adjustIndicators`, `adjustBitWidth`) after the commands, as upstream does. The diagram style is computed there and kept, so `skin` and `<style>` apply.
- `ExportDiagram` calls `uml.ExportTitled` with upstream's `getTextBlock`, as `internal/activitydiagram3/export.go` does.
- `rotate` is parsed into `PacketItem.textRotation` and not drawn, and `PacketBlock` keeps upstream's unused `openerStroke` and open edges. The port keeps them so that a later upstream fix maps onto the same code.

### D3. Java behavior that the output depends on

- `adjustIndicators` checks the divisors before `%` and panics with `uml.Throw(uml.ArithmeticException, "/ by zero")`, because Go's own panic text is `runtime error: integer divide by zero`. `IntStream.limit` with a negative size panics with `uml.IllegalArgumentException` and the size as the message, as Java's does.
- Bit numbers, widths and heights are `int32`, so that the sums of `CommandNumRange` overflow as Java's `int` does. The attribute values are parsed with Java's `Integer.parseInt` rules (sign, digits, 32-bit range), and a value that does not parse gives the default, as upstream's `getIntAttribute`.
- The text shortening of `PacketBlock` cuts the label by UTF-16 code units, as `String.substring` does.
- A negative width or height panics in `geom.NewXDimension2D`, where Java throws. `uml.ExportTitled` turns the panic into an error, so a reversed range fails at drawing time, as upstream does.

### D4. Registration and detection

- `packetdiag.Factory` (`uml.NewFactory(core.Packet, ...)`) goes into `engine.Factories` after `activitydiagram3.Factory`, which is upstream's `PSystemBuilder` order without the factories that are not ported.
- `core.Packet` joins `supportedTypes`, so `CheckFamily` lets `@startpacketdiag` blocks reach the factory. `TestCheckFamily` gets a `@startpacketdiag` check, and keeps `@startmindmap` as its unsupported family.

### D5. Errors through the shared crash handling

The `ArithmeticException` of D3 happens in `CheckFinalError`, inside `core.createSafely`, which builds the "Fatal crash error" diagram with the whole block as its trace, as upstream's `PSystemBuilder` does. `TestVegaReferences` of `internal/errdiagram` already draws the images of `R07.02` and `R11.01` from the message and line of their references and matches them, but no corpus case has reached `createSafely` itself. The port checks that `createSafely` gives the same message and line for these two cases and the curated "Empty range" case, and fixes `createSafely` if it does not. Such a fix changes only the error diagrams of factory panics.

### D6. Formats and WASM

TXT and UTXT need no new code: the engine rejects them before the export, as for class, description and state diagrams. The WASM build needs no special case either, because packet diagrams do not use Graphviz.

### D7. Documentation

- The README and `site/index.html` get the row `Packet | yes | yes | no` in the table of supported diagrams, and the Graphviz note of the page names packet diagrams among the diagrams that do not need `dot`. The README's Graphviz sentence lists only the diagrams that need `dot`, so it does not change. The README's conformance counts follow the report.
- The playground note of the page and the README's sentence about the playground stay as they are, because the page runs the WASM module of v0.1.0, which has no packet diagrams. The release that ships packet diagrams updates them.

### D8. Conformance

- Curated cases go under `testdata/conformance/cases/packet-diagram/` with `family: packetdiag` and `type: packet`, one per compared scenario, plus `diagram-source/packet-detected` and an embedded case (`family: uml`, `type: sequence`). Their references come from `make reference`, and CI renders them again and checks them against the committed files.
- The embedded case is an `@startuml` case, so the golden dumps of `internal/sequencediagram`, `internal/classdiagram` and `internal/activitydiagram3` are regenerated. The `@startpacketdiag` cases are not part of those dumps.
- Scenarios that are not compared with Java get `.go.yaml` cases: "Packet diagram without Graphviz" (environment), "Reversed range" (upstream's crash report has the date, the machine and the Java version, so it cannot be a committed reference), "TXT for a packet diagram" (upstream draws TXT) and "Packet diagram in WASM" (environment).
- The cases come before the code, because `TestScenarioCoverage` reads the delta specs of active changes. Until the factory is registered, the new packet cases report "not supported yet", which the runner counts as unsupported, not failed, and `TestDiagramReferences` skips the two curated packet error cases. The embedded case is a sequence case, which the runner never counts as unsupported, so it fails in `make conformance` until then.
- The 62 Vega packet cases are compared once the factory is registered. A case that still differs at the end is listed in `known-differences.yaml` with its reason.

## Risks / Trade-offs

- **[Text shortening depends on measurement]** The shortening measures each shorter label, so a measurement difference changes where the text is cut. → It uses the same string bounder and font metrics as the other diagram types. The Vega cases with shortened labels (`R04.01`, `R04.02`, `R05.01` to `R05.05`, `R16.01`) and the "Shortened text" case check it.
- **[createSafely not compared before]** Its message and line have not been checked against upstream until now. → D5: the two Vega references and the curated "Empty range" case compare them.
- **[Drawing crash without a crash report]** A reversed range returns an error where upstream draws a crash report. → The runner accepts this for crash references. The HTTP server answers 503, as upstream's built-in server does for a crash.
- **[Page ahead of the release]** The page lists packet diagrams while its WASM module (v0.1.0) has none. → Decided with the user for this change. The playground note does not claim packet diagrams render in the browser.
- **[Supported families]** `supportedTypes` lets every `@startpacketdiag` block through to the factory. → Only the packet factory has the `Packet` type, so no other diagram type sees these blocks.

## Migration Plan

Nothing to migrate. Inputs that failed with "packet diagrams are not supported yet" now render. Rollback means reverting the change.
