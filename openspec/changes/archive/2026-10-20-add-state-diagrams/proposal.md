# Proposal

## Why

State diagrams fail today with "state diagrams are not supported yet", and they are the most widely used of the diagram types that are still missing. Much of what they need is already ported: the entity model has the state leaf and group types and the concurrent separators, `internal/svek` has the state headers of composite states, the entry and exit points (`FrontierCalculator`) and the Graphviz pipeline, and the command parser has the multi-pass loop. What is missing is the diagram itself, its factory and commands, the state images, and the separate layout of composite states.

The conformance corpus has 38 state cases, all reported as not supported yet. 19 of them (every case under `vega/state/`) set `!pragma layout smetana`, which this implementation lays out with `dot`, so they can only match upstream once the Smetana layout is ported.

## What Changes

- State diagrams under `@startuml`, as upstream's `StateDiagram`:
  - states declared with `state`, display names with `as`, description lines (`state X : text` and `X : text`), stereotypes, tags, URLs, background colors, and line colors and styles (`##[dashed]red`)
  - pseudo states: `[*]` as the start or end of the diagram or of a composite state, the history states `[H]`, `[H*]`, `X[H]` and `X[H*]`, the synchronization bars `==name==`, and the stereotypes `<<choice>>`, `<<fork>>`, `<<join>>`, `<<start>>`, `<<end>>`, `<<history>>` and `<<history*>>`
  - entry and exit points and pins (`<<entryPoint>>`, `<<exitPoint>>`, `<<inputPin>>`, `<<outputPin>>`, `<<expansionInput>>`, `<<expansionOutput>>`) drawn on the border of their composite state
  - transitions with upstream's syntax: `->` and `-->`, the reversed form `<--`, direction hints, line styles and colors in brackets, length, labels, stereotypes, the `x` start and the `o` end, and the `node` label style (`-[node]->` or `skinparam stateDiagramEdgeLabelStyle node`)
  - composite states (`state X {` ... `}` or `end state`), nesting, dotted names (`state A.x`), `frame` groups, and concurrent regions separated by `--` or `||`, with upstream's error when a state inside a concurrent region is linked outside of it
  - notes on states and transitions, floating notes, URLs, `hide`/`show`/`remove`, `hide empty description`, directions, and the `map` and `json` elements
- Diagram type detection follows upstream's order: sequence, class, description, state, then activity (new syntax). The "not supported yet" error for state diagrams goes away.
- In the WASM build, state diagrams return the Graphviz-not-available error, like class and description diagrams.
- State diagrams can be embedded with `{{` ... `}}`. The styling scenario that shows the error image of an embedded unsupported type uses the legacy activity syntax instead of the state syntax.
- Conformance: curated cases for every new scenario. The 19 Vega state cases without the Smetana pragma are compared and pass. Those of the 19 with it that differ are listed in `known-differences.yaml`, resolved by `add-smetana-layout`, like the 4 class cases that already set it.
- The README lists state diagrams as supported.

Out of scope:

- SCXML output (`--scxml`), which stays rejected as an unknown flag
- TXT and UTXT output of state diagrams, which is rejected as it is for class and description diagrams
- the Smetana layout, which the `vega/state/` cases need to match upstream
- sprites, OpenIconic icons and stdlib libraries in stereotypes or labels, which keep their current behavior

## Capabilities

### New Capabilities

- `state-diagram`: state diagrams under `@startuml`, with upstream's syntax and Graphviz-based layout

### Modified Capabilities

- `diagram-source`: state becomes a supported `@startuml` type in upstream's detection order, and leaves the list of types reported as not supported yet
- `styling`: the example of an embedded diagram of an unsupported type uses the legacy activity syntax, because state diagrams are now supported
- `wasm-build`: state diagrams join class and description diagrams as graph-based diagrams that report that graph layout is not available

## Impact

- **Code:** a new `internal/statediagram` package (diagram, factory and the commands of `statediagram/command`), the state images and the state layout in `internal/svek` (`EntityImageState` and its variants, the circles, branches, bars, history states and transition labels, `CucaDiagramSimplifierState`, `GroupMakerState`, `InnerStateAutonom`, `ConcurrentStates`, `Cluster.drawUState`), a parser pass parameter for the shared note commands in `internal/abel`, the factory list in `internal/engine`, and the detectors in `internal/core`.
- **Runtime dependency:** state diagrams need Graphviz `dot`, like class and description diagrams.
- **Conformance:** new curated cases and committed references under `testdata/conformance/`, and up to 19 new entries in `known-differences.yaml`. CI renders the references again and checks them against the committed ones.
- **Compatibility:** inputs that failed with "state diagrams are not supported yet" now render.
