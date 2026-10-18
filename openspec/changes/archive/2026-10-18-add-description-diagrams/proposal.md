# Proposal

## Why

Use case, component and deployment diagrams (upstream's description diagrams) fail today with "description diagrams are not supported yet". Most of their drawing is already ported, because class diagrams with `allowmixing` use the description element images, every `USymbol` and the package clusters. What is missing is the diagram itself, its factory, two partly ported commands, ports and the Problem Frames elements.

Five known differences (the Vega `tikz` cases) wait for this change, but they also need embedded diagrams (`{{ }}`). Those are not ported: today every `{{` in a note, an activity label or a class body crashes the render, even when the embedded diagram is a supported sequence diagram.

## What Changes

- Description diagrams under `@startuml`, as upstream's `DescriptionDiagram`:
  - every element keyword upstream accepts (`actor`, `usecase`, `component`, `node`, `database`, `artifact`, `rectangle`, `card`, `cloud`, `folder`, `frame`, `file`, `queue`, `stack`, `storage`, `agent`, `person`, `boundary`, `control`, `entity`, `interface`, `circle`, `collections`, `hexagon`, `label`, `process`, `action`), the business variants `actor/` and `usecase/`, the short forms `(Use case)`, `:Actor:`, `[Component]` and `() Interface`, and multiline element bodies
  - links with upstream's syntax: heads, line styles, direction hints, length, labels and colors
  - containers drawn with a symbol (`node`, `folder`, `frame`, `cloud`, `database`, `rectangle`, `package` and the others upstream allows), nesting, `together`, notes, URLs, `hide`/`show`/`remove`, directions, and the `map` and `json` elements
  - ports: `port`, `portin` and `portout`
  - the Problem Frames elements of `CommandCreateDomain` (domains, designed domains, machines and requirements)
- Diagram type detection follows upstream's order: sequence, class, description, then activity (new syntax). The "not supported yet" errors for use case, component and deployment diagrams go away. Archimate keeps its "not supported yet" error.
- Embedded diagrams: `{{ ... }}` in creole text renders the nested diagram as upstream's `EmbeddedDiagram` does, for every diagram type this implementation supports. A nested diagram of an unsupported type renders its "not supported yet" error image in place.
- In the WASM build, description diagrams return the Graphviz-not-available error, like class diagrams.
- Conformance: curated cases for every new scenario. The 10 Vega description cases and the 5 `tikz` embed cases are compared, and those 5 entries leave `known-differences.yaml`.
- The README lists description diagrams as supported.

Out of scope:

- archimate elements, which need sprites for their icons
- sprites, OpenIconic icons and stdlib libraries in stereotypes or labels (`<<$name>>`, `<$name>`, `<&name>`, `!include <awslib/...>`), which keep their current behavior
- TXT and UTXT output of description diagrams, which is rejected as it is for class diagrams
- XMI and GraphML output

## Capabilities

### New Capabilities

- `description-diagram`: use case, component and deployment diagrams under `@startuml`, with upstream's syntax and Graphviz-based layout

### Modified Capabilities

- `diagram-source`: description becomes a supported `@startuml` type in upstream's detection order, and use case, component and deployment leave the list of types reported as not supported yet
- `styling`: creole text renders embedded diagrams (`{{ ... }}`)
- `wasm-build`: description diagrams join class diagrams as graph-based diagrams that report that graph layout is not available

## Impact

- **Code:** a new `internal/descdiagram` package (diagram, factory, the element, link and domain commands), the remaining entity images (ports and Problem Frames elements) and `FrontierCalculator` in `internal/svek`, the embedded diagram maker between `internal/klimt` and `internal/engine`, the factory list in `internal/engine`, and the detectors in `internal/core`. Shared pieces now inside `internal/classdiagram` move out: `Labels` and `StringWithArrow` to `internal/descdiagram`, and the `map` and `json` commands to `internal/abel`.
- **Runtime dependency:** description diagrams need Graphviz `dot`, like class diagrams.
- **Conformance:** new curated cases and committed references under `testdata/conformance/`. CI renders the references again and checks them against the committed ones.
- **Compatibility:** inputs that failed with "not supported yet" now render. Inputs with `{{ }}` that crashed now render.
