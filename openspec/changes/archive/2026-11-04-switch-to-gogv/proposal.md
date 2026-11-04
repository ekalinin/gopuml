# Proposal

## Why

Class, object, description, state, legacy activity and Chen EER diagrams need an installed Graphviz `dot` today. Without it they render only upstream's "Cannot find Graphviz" error image. The WebAssembly build and the playground cannot draw them at all: the five `wasm-build/*-diagram-error` cases and `landing-page/playground` check that error. gogv (`github.com/ekalinin/gogv`) is a Go implementation of `dot`, without cgo and without an installed Graphviz, and it already lays out the DOT that gopuml v0.3.0 generates for its conformance corpus. No conformance case fails today. Rendered with gogv v0.2.0 as `dot`, 12 of the 972 cases fail (listed under Impact), and the other cases keep their results. The change closes issue #80.

## What Changes

- gogv becomes the layout engine of graph-based diagrams. The DOT graph that `internal/svek` writes is laid out in the process by `gogv.Render`, which returns the SVG that `dot -Tsvg` returns, so no Graphviz installation is needed.
- The external `dot` runs only when it is named: by `--dot-path`, the `GRAPHVIZ_DOT` environment variable or `Options.DotPath` of the library. `dot` is no longer searched for on `PATH` and in the standard install directories. A named `dot` that does not exist still gives the "Cannot find Graphviz" error image, without falling back to gogv.
- `--graphviz-timeout` and `Options.DotTimeout` also limit a gogv layout (30 seconds by default). A layout stopped by the timeout reports the Graphviz timeout error, and cancelling the context of a render stops the layout.
- A gogv layout that fails or panics reports the Graphviz failure error with gogv's message for that diagram, and the other diagrams of the input still render.
- `--check-graphviz` without a named `dot` checks gogv: it prints `GraphViz: dot - graphviz version 16.1.0 (gogv <version>)`, lays out the test graph and prints `Installation seems OK. File generation OK`, with exit status 0.
- `!pragma layout smetana`, `!pragma layout elk` and `!pragma layout vizjs` are laid out with gogv, or with the named `dot`.
- WASM: class, description, state, legacy activity and Chen EER diagrams render with gogv instead of returning the "graph layout is not available in this build" error, and so they do in the playground of the landing page.
- Conformance: the outputs of this implementation are laid out with gogv in CI too. `tools/refgen` removes the `GRAPHVIZ_DOT` of the reference image from the environment of this implementation, unless a case sets it. The cases that differ because gogv lays a graph out differently from Graphviz 14.0.1 are fixed in gogv, or listed in `known-differences.yaml` with that reason.
- Dependency: `github.com/ekalinin/gogv` v0.2.0 in `go.mod`, the first release that exports the Graphviz version it follows (`gogv.GraphvizVersion`), or a later v0.2.x release that fixes some of the 12 cases before this change merges.

Out of scope:

- Upstream's search for `dot` on `PATH` and in the standard install directories when no `dot` is named. gopuml uses gogv instead, as upstream uses its Smetana port when it finds no `dot`.
- Upstream's Smetana, ELK and VizJs engines: the layout pragmas keep using the engine of the render.
- `skinparam linetype ortho` with gogv: gogv does not support `splines=ortho` yet and draws splines instead. A named `dot` draws orthogonal lines, as upstream does.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `graph-layout`: gogv lays out graph-based diagrams unless a `dot` is named; discovery, the missing-`dot` error, the timeout, failures, the check and the layout pragmas follow from that
- `wasm-build`: graph-based diagrams render in the WASM build instead of returning the Graphviz-not-available error
- `landing-page`: the playground renders graph-based diagrams
- `library-api`: the Graphviz-not-found error comes only from a named `dot`, and cancellation stops a gogv layout too
- `upstream-compatibility`: the conformance run lays out the outputs of this implementation with gogv, not with the Graphviz of the reference image

## Impact

- **Code:**
  - `internal/dot`: a gogv engine next to the dotexec engine, the choice between them, and the check of `--check-graphviz` for gogv
  - `internal/svek`: the engine of a layout comes from that choice, in the native and the WASM builds
  - `internal/cli`, the root package and `cmd/plantuml-wasm`: comments and tests; no API is added or removed
  - the tests that skip when no `dot` is installed (`internal/svek`, `internal/classdiagram`, `internal/statediagram`, `internal/cheneer`, the root package) run with gogv
  - `tools/refgen`: the environment of the implementation under test
- **Runtime dependency:** Graphviz `dot` is no longer needed, and is used only when named. The Go module `github.com/ekalinin/gogv` v0.2.0 (MIT) is added; the WASM module grows by its code (gogv's own WASM build of its `dot` command is 4.6 MB, Go runtime included; gopuml's is 51 MB today).
- **Conformance:** the 12 cases that fail with gogv: `chen-eer-diagram/entities-relationship-and-attributes`, `state-diagram/choice-fork-join`, `state-diagram/history-style`, `vega/migration/ChenMovie_Test`, `ChenMovieAlias_Test`, `ChenMovieExtended_Test`, `vega/nonreg/group2681/history_style_all_forms`, `vega/nonreg/simple/A0004`, `vega/nonreg/simple/ComponentExtraArrows_0001`, `vega/scxml/state2`, `vega/scxml/state3` and `vega/svg/interactive/SVG0006_Svek`. Each one is fixed in gogv or becomes a known difference, so the known-difference count of the report may rise from 42 up to 54. The reasons of the 38 Smetana known differences name gogv instead of `dot`. New `.go.yaml` cases cover the new scenarios, and the WASM cases check a rendering instead of an error. No `@startuml` case changes, so the golden dumps stay.
- **Documentation:** `README.md` and `site/index.html` (the "needs `dot`" column of the tables of supported diagrams, the Graphviz notes, the WASM and playground notes, and the conformance section, whose arm64 `GRAPHVIZ_DOT` wrapper is no longer needed), and `NOTICE` (gogv among the linked Go code).
- **Compatibility:** without a named `dot`, graph-based diagrams are laid out by gogv even when Graphviz is installed, so their arrangement can differ from upstream's within the parity tolerance. Users who need upstream's exact arrangement, or `skinparam linetype ortho`, set `GRAPHVIZ_DOT` or `--dot-path`. A `dot` on `PATH` is no longer used. `--check-graphviz` succeeds without Graphviz.
