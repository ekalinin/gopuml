# Proposal

## Why

Dot diagrams (`@startdot`) fail today with "dot diagrams are not supported yet" (issue #52). Neither the Vega corpus nor the curated cases have a `@startdot` case, so the report counts no failing or unsupported dot case: this change adds the first ones. Upstream's `directdot` package is small (a factory and a diagram that hands the block to `dot -T<format>` and writes what `dot` returns), and the layout engines it needs already exist: gogv, which writes the SVG of `dot -Tsvg` itself, and the dotexec engine for a named `dot`.

## What Changes

- Dot diagrams under `@startdot`, as upstream's `PSystemDotFactory` and `PSystemDot`:
  - the block is a Graphviz graph: after the `skinparam` and `!pragma` lines that directly follow the start line, which are dropped, and the empty lines, the first line is a graph header (`[strict] [di]graph [ID] {`), and that line and the following ones, preprocessed, are the DOT source
  - upstream's errors: `Empty description` for a block without content and `Syntax Error?` for a block that does not start with a graph header
  - the SVG is the output of the layout engine for `-Tsvg`, written unchanged and without source metadata, in the `INSECURE` and `INTERNET_WITH_DOTSVG` security profiles. In the other profiles, including `LEGACY`, the default of the CLI and the library, and `INTERNET`, the default of the server, the SVG is upstream's "This feature has been suppressed" document, without a layout
  - the layout engine is the one of the render: gogv, or the `dot` named by `--dot-path`, the library option or `GRAPHVIZ_DOT`. As upstream merges what `dot` writes on standard error into the output, the output holds the warnings that Graphviz writes (`Warning: ...`) and, for invalid DOT, its error message (`Error: <stdin>: syntax error in line 3 near '}'`). gogv's warnings about Graphviz features it does not support yet are not written
  - the PNG is the output of a named `dot` for `-Tpng`. gogv does not write PNG (ekalinin/gogv#33), so without a named `dot` the PNG is upstream's "GraphViz has crashed" image, as upstream draws it with a Graphviz that has no PNG renderer
  - a named `dot` that cannot be used gives upstream's "There is an issue with your Dot/Graphviz installation" image, and a layout that fails without output or times out gives the "GraphViz has crashed" image, both with an error status
  - the syntax check reports a dot diagram as `OTHER` with the description `(Dot)`, as upstream
- Family detection: `dot` becomes a supported family. The "not supported yet" error for `@startdot` goes away.
- TXT and UTXT of dot diagrams stay unsupported, as for the other diagram types without a text output.
- In the WASM build, which runs in the `SANDBOX` profile, the SVG of a dot diagram is the suppressed-feature document, as in the native build with that profile.
- The conformance harness lets a case name the security profile that upstream and this implementation render it with, so that the layout of dot diagrams can be compared with upstream's in `INTERNET_WITH_DOTSVG`.
- Conformance: curated cases for every compared scenario, and `.go.yaml` cases for the scenarios that depend on the environment or on output that is not SVG. Cases whose graphs use Graphviz features that gogv does not support yet, such as `shape=record`, `fillcolor` and `style=filled`, are listed in `known-differences.yaml`.
- The README and the landing page list dot diagrams in the table of supported diagrams, with the profile that the SVG needs and the named `dot` that the PNG needs. The README's conformance counts are updated.

Out of scope:

- The output formats other than SVG and PNG that upstream asks `dot` for (`dot -Teps` and others).
- TXT and UTXT, which upstream asks `dot` for as `-Tatxt` and `-Tutxt` and which always give its "GraphViz has crashed" text.
- A PNG writer in gogv, tracked by ekalinin/gogv#33.
- The Graphviz features that gogv does not support yet, which are gogv's to add.
- The other output of upstream for `dot -Tpng` without a PNG renderer: when `dot` has not exited yet as upstream writes the DOT source, upstream writes `dot`'s `Format: "png" not recognized` text as the PNG, with a success status (see design).

## Capabilities

### New Capabilities

- `dot-diagram`: dot diagrams under `@startdot`, with upstream's block rules, errors, security profile check for SVG, layout engine output and its merged messages, PNG output, layout failures and syntax check

### Modified Capabilities

- `diagram-source`: `dot` becomes a supported family next to `uml`, `packetdiag`, `gantt`, `mindmap`, `wbs`, `salt` and `chen`
- `wasm-build`: dot diagrams render in the WASM build as in the native build with the `SANDBOX` profile, that is as the suppressed-feature document for SVG
- `upstream-compatibility`: a conformance case can name the security profile that both implementations render it with

## Impact

- **Code:**
  - a new `internal/directdot` package: the factory and the diagram
  - `internal/command`: a port of `PSystemBasicFactory`, the base of factories that read a block line by line without commands
  - `internal/core`: `UmlSource.removeInitialNoise`, the exported check of a Graphviz graph header, and the supported families
  - `internal/security`: `allowDotSvg` of the profiles
  - `internal/dot` and `internal/svek`: running the layout engine of the render for a format, with `dot`'s standard error merged into the output
  - `internal/engine`: the factory list, and the TXT and UTXT check for dot diagrams
  - `internal/conformance` and `tools/refgen`: the `profile` of a case
- **Runtime dependency:** none. Without a named `dot`, gogv lays the graphs out; a named `dot` is needed for PNG.
- **Conformance:** new curated cases, CLI cases and `.go.yaml` cases, with committed references under `testdata/conformance/`; references rendered in `INTERNET_WITH_DOTSVG`, `INSECURE` and `INTERNET` for the cases that name these profiles. The not-supported-yet count of the report stays at 5. New entries in `known-differences.yaml` for the cases whose graphs use features that gogv does not support yet. No `@startuml` case changes, so the golden dumps stay as they are.
- **Documentation:** `README.md` and `site/index.html` (the table of supported diagrams, with a note on the security profile and the named `dot`), and the README's conformance counts.
- **Compatibility:** inputs that failed with "dot diagrams are not supported yet" now give upstream's output: the suppressed-feature SVG in the default profiles of the CLI, the library, the server and the WASM build, the layout in `INSECURE` and `INTERNET_WITH_DOTSVG`, or upstream's errors.
