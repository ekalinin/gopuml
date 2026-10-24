# Proposal

## Why

Packet diagrams (`@startpacketdiag`) fail today with "packet diagrams are not supported yet". They are the largest group of unsupported cases in the conformance corpus: 62 of the 105 cases reported as not supported yet are under `vega/packet/`. Upstream's `packetdiag` package is small (11 files: the diagram, the factory, 3 drawing classes and 7 commands), needs no Graphviz, and draws with parts that are already ported (`uml.Base`, `uml.ExportTitled`, the common commands, styles, creole text, the stencil and the SVG and PNG exporters).

## What Changes

- Packet diagrams under `@startpacketdiag`, as upstream's `PacketDiagram`:
  - the optional `packetdiag {` ... `}` wrapper
  - fields declared as a bit (`N: text`), a bit range (`A-B: text`) or a field placed after the previous one (`* text`), with the `len`, `height` and `rotate` attributes in brackets. Other attributes (`label`, `color`, `description`, `numbered`, `colheight`, ...) are accepted and ignored, as upstream does, and so is `rotate`, which upstream parses but does not draw
  - the directives `colwidth`, `node_height`, `scale_direction` (`ltr`, `rtl`), `scale_interval` and `same_height`
  - upstream's layout: the bit scale with its numbered and full ticks, rows of `colwidth` bits, fields split across rows, labels shortened with `...` when they do not fit, and the row width reduced to the width of the first row
  - the common commands of upstream's `addCommonCommands1` (title, header, footer, caption, legend, `skin`, `skinparam`, `<style>`, ...)
  - upstream's errors: a syntax error with "Assumed diagram type: packet", and the crashes upstream has for an empty or reversed bit range (`4-3`, `* A [len = 0]`, `4-2`)
- Family detection: `packetdiag` becomes a supported family next to `uml`. The "not supported yet" error for `@startpacketdiag` goes away.
- Packet diagrams render in SVG and PNG. TXT and UTXT are rejected with the error that names the format and the diagram type, as for the other types except sequence.
- The WASM build renders packet diagrams fully, because they need no graph layout.
- Packet diagrams can be embedded in creole text with `{{packetdiag` ... `}}`, which the embedded diagram code already supports for every family the engine renders.
- Conformance: curated cases for every new scenario. The 62 Vega packet cases are compared instead of being reported as not supported yet.
- The README and the landing page list packet diagrams in the table of supported diagrams, and the README's conformance counts are updated. The playground note of the landing page does not change: the page uses the WASM module of v0.1.0, which has no packet diagrams. The release that ships them updates it.

Out of scope:

- TXT and UTXT output of packet diagrams
- `nwdiag`, which shares the attribute syntax and `VerticalLine` with `packetdiag` but is a separate diagram type
- the attributes upstream ignores, which stay ignored

## Capabilities

### New Capabilities

- `packet-diagram`: packet diagrams under `@startpacketdiag`, with upstream's syntax, layout and errors

### Modified Capabilities

- `diagram-source`: `packetdiag` becomes a supported family next to `uml`
- `wasm-build`: packet diagrams join sequence and activity diagrams as diagrams with a built-in layout that render fully in the WASM build

## Impact

- **Code:** a new `internal/packetdiag` package (`PacketDiagram`, `PacketDiagramFactory`, `PacketBlock`, `PacketIndicator`, the 7 commands, and the two helpers it uses from `nwdiag`: `NwDiagram.toSet` and `VerticalLine`), the factory list in `internal/engine`, and the supported families in `internal/core`.
- **Runtime dependency:** none. Packet diagrams do not need Graphviz.
- **Conformance:** new curated cases and committed references under `testdata/conformance/`. The not-supported-yet count of the report drops from 105 to 43, unless some packet cases end up in `known-differences.yaml`.
- **Documentation:** `README.md` and `site/index.html` (the table of supported diagrams, which `landing_test.go` keeps equal).
- **Compatibility:** inputs that failed with "packet diagrams are not supported yet" now render.
