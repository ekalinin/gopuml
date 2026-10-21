# Proposal

## Why

Activity diagrams in the legacy syntax (`(*) --> "First step"`, `if ... endif`, `===bar===`) fail today with "legacy activity diagrams are not supported yet", although many existing diagrams still use that syntax. The state diagram change also left a detection gap: upstream tries its legacy activity parser before description and state diagrams, so a block such as `A --> ==sync==` with `==sync== --> C` is a legacy activity diagram upstream, and this implementation now draws it as a state diagram.

Most of the drawing is ported already: the start and end circles, branches, synchronization bars, notes, partitions as clusters, octagon nodes and the Graphviz pipeline. What is missing is the diagram, its factory and commands, the activity image and the separate layout of inner activities. Five of the upstream commands are written with upstream's UBrex expressions, whose engine is ported for creole, but not the parts that commands need.

## What Changes

- Legacy activity diagrams under `@startuml`, as upstream's `ActivityDiagram`:
  - the start `(*)` and the ends `(*)`, `(*top)` and `(*1)`, activities written as quoted text (on one or several lines), with codes (`"Long name" as L`) or as bare names, stereotypes and background colors, and `in Partition` to put the target of a transition in a partition
  - transitions from an activity or from the last one (`--> "Next"`), with labels in brackets, dotted bodies, direction hints, bracket styles, URLs, and `-*->` for a transition without layout constraint
  - branches with `if "label" then` ... `else` ... `endif`, synchronization bars `===name===`, partitions (`partition Name {` ... `}` or `end partition`), notes on the last activity (`note left: text` and multi-line notes), notes on links, and inner activities (`--> {` ... `}`)
  - `skinparam activityShape octagon`
- Diagram type detection follows upstream's order: sequence, class, legacy activity, description, state, then activity (new syntax). The "not supported yet" error for legacy activity diagrams goes away, and the blocks upstream reads as legacy activity diagrams are no longer drawn as state diagrams.
- Legacy activity diagrams need Graphviz `dot`. In the WASM build they return the Graphviz-not-available error, like class, description and state diagrams.
- The styling scenario that shows the error image of an embedded unsupported type uses the timing syntax.
- Conformance: curated cases for every new scenario, the Vega legacy activity case `vega/nonreg/simple/A0004` compared, and the README updated.

Out of scope:

- TXT and UTXT output of legacy activity diagrams, which is rejected as it is for class, description and state diagrams
- upstream's `CommandInnerConcurrent`, which upstream does not register
- the Smetana layout

## Capabilities

### New Capabilities

- `legacy-activity-diagram`: activity diagrams in the legacy syntax under `@startuml`, with upstream's syntax and Graphviz-based layout

### Modified Capabilities

- `diagram-source`: legacy activity becomes a supported `@startuml` type in upstream's detection order, and leaves the list of types reported as not supported yet
- `styling`: the example of an embedded diagram of an unsupported type uses the timing syntax
- `wasm-build`: legacy activity diagrams join the graph-based diagrams that report that graph layout is not available
- `graph-layout`: only activity diagrams in the new syntax render without `dot`; legacy activity diagrams need it

## Impact

- **Code:** a new `internal/activitydiagram` package (diagram, conditional context, factory, the link, branch, partition and note commands); the UBrex builder and an exported matcher in `internal/klimt`, `regex.Result` backed by UBrex captures, a UBrex single-line command in `internal/uml` and the UBrex variants of `footbox`, `direction` and `hide`/`show` in `internal/abel`, and the UBrex color parser as a UBrex part; `EntityImageActivity`, `InnerActivity`, `GroupMakerActivity` and `CucaDiagramSimplifierActivity` in `internal/svek`; the factory list in `internal/engine` and the detectors in `internal/core`.
- **Runtime dependency:** legacy activity diagrams need Graphviz `dot`.
- **Conformance:** new curated cases and committed references under `testdata/conformance/`, and regenerated golden dumps.
- **Compatibility:** inputs that failed with "legacy activity diagrams are not supported yet" now render. Blocks that both the legacy activity and the state parser accept become legacy activity diagrams, as upstream.
