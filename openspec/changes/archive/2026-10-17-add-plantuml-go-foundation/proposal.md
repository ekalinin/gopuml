# Proposal

## Why

The goal is a 1:1 replacement of PlantUML (https://github.com/plantuml/plantuml) written in Go. Upstream v1.2026.8 is about 3,300 Java files and ~447k lines covering ~30 diagram families, so a port can't land in one step. It needs a shared foundation (input language, styling, rendering, layout, interfaces, conformance testing) that every diagram type builds on, plus a staged roadmap for the rest.

## What Changes

- New Go module (MIT license) that translates the `plantuml-mit` variant of PlantUML v1.2026.8 and keeps the upstream copyright notices. The pinned upstream version is the compatibility target for this change.
- Core pipeline shared by all diagram types:
  - extracting `@start`/`@end` blocks and detecting the diagram type
  - syntax errors and error images
  - the preprocessor: variables, conditionals, procedures and functions, builtins, includes, themes
  - styling: `skinparam`, `<style>`, built-in themes, creole text markup
- Output formats: SVG, PNG, PREPROC, and TXT/UTXT for sequence diagrams. SVG and PNG embed the source as metadata, as upstream does.
- Graph layout through the external Graphviz `dot` binary (upstream's default), behind a layout abstraction so that a pure Go port of Smetana can replace it later.
- First diagram types:
  - sequence, using the Teoz engine (the only sequence engine in v1.2026.8)
  - class
  - activity, new syntax
- Interfaces:
  - a CLI compatible with `plantuml.jar` (flags, file naming, exit codes)
  - a Go library API
  - an HTTP server with upstream's `--http-server` routes and the plantuml-server URL scheme
  - a WASM build
- PlantUML URL text encoding and decoding: deflate plus PlantUML's base64 alphabet, and the `~h` hex form.
- Where upstream v1.2026.8 crashes, hangs or ignores one of its own options, this implementation fixes the behavior (listed in design.md). Everywhere else it follows upstream, including behavior that may look surprising.
- A conformance harness that compares Go output against reference renders from pinned Java PlantUML v1.2026.8 running in Docker. The corpus is upstream's Vega test inputs plus curated per-diagram cases.
- Out of scope for this change, planned as follow-up changes, one per item or group:
  - Diagram types under `@startuml`: use case/component/deployment/archimate (description diagrams), state, object, timing, legacy activity
  - Diagram families: mindmap, wbs, gantt, json/yaml/hcl, salt, nwdiag/packetdiag, ebnf, regex, chen, files, board, git, wire, chart, flow, bpm, math/latex, dot, creole, def, sprites
  - Assets: sprites, OpenIconic icons, emoji, and the bundled stdlib (C4, AWS, Azure, ...)
  - Output formats: PDF, EPS, LaTeX/TikZ, XMI, SCXML, GraphML, HTML, VDX, BASE64, Braille PNG, OBFUSCATE, and TXT/UTXT for class and activity diagrams
  - Engines: the pure Go Smetana layout port (fallback when `dot` is missing, and needed for WASM graph layout) and ELK layout
- Not planned:
  - features the MIT variant removes: ditaa, jcckit, sudoku, zxing QR codes, the webp decoder
  - JVM-only tooling: `--gui`, `--ftp-server`, `--clipboard*`, statistics flags
  - byte-identical output (the compatibility target is visual parity)

## Capabilities

### New Capabilities

- `diagram-source`: finding `@start`/`@end` blocks in input, detecting the diagram type, reporting syntax errors and rendering error images
- `preprocessor`: upstream-compatible preprocessing (variables, legacy defines, conditionals, loops, procedures, functions, builtins, JSON, includes, security profiles)
- `styling`: `skinparam`, the `<style>` system, built-in themes, colors, and creole text markup
- `output-formats`: SVG, PNG, TXT/UTXT, PREPROC output, embedded source metadata, deterministic rendering
- `graph-layout`: layout of graph-based diagrams through Graphviz `dot`, including discovery, timeout and error handling
- `sequence-diagram`: sequence diagrams under `@startuml`
- `class-diagram`: class diagrams under `@startuml`
- `activity-diagram`: activity diagrams (new syntax) under `@startuml`
- `cli`: a command line interface compatible with `plantuml.jar`
- `library-api`: a Go package API for rendering diagrams from source
- `http-server`: an HTTP rendering server compatible with upstream URL schemes
- `url-encoding`: encoding and decoding diagram source in PlantUML's URL text format
- `wasm-build`: a WebAssembly build for rendering in browsers and other JS hosts
- `upstream-compatibility`: the parity contract with PlantUML v1.2026.8 and how it is measured

### Modified Capabilities

None. This is a greenfield project with no existing specs.

## Impact

- **Code:** a new Go module in this repository with the CLI, library packages, the WASM entry point, a test corpus and the reference-render tooling.
- **Runtime dependency:** Graphviz `dot` is needed for class diagrams until the pure Go Smetana port lands. Sequence and activity diagrams don't need it.
- **Development dependency:** Docker, to produce reference renders from Java PlantUML v1.2026.8. There's no local Java runtime here.
- **Licensing:**
  - MIT, with upstream copyright and attribution notices.
  - Bundled assets such as themes need a license check before they ship.
  - The future Smetana port must remain a separate EPL-1.0 module.
- **Compatibility:** users can swap `plantuml.jar` for the Go binary for the diagram types and formats in this change. Unsupported diagram types and formats fail with explicit errors and never produce silent wrong output.
