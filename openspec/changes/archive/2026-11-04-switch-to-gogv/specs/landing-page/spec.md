# Spec Delta

## MODIFIED Requirements

### Requirement: Playground

The page SHALL provide a playground: an editor whose source is rendered as SVG in the browser by the WebAssembly module of the release the page names, without a server. The rendering SHALL follow the edits of the source. The playground SHALL offer example diagrams, SHALL show the error message and its line when the source has an error, SHALL state which diagram types render in the browser, and SHALL let the visitor download the rendered SVG.

#### Scenario: Sequence diagram

- **WHEN** a visitor types `@startuml`, `Alice -> Bob : hello` and `@enduml` in the editor
- **THEN** the playground shows the rendered sequence diagram as SVG

#### Scenario: Syntax error

- **WHEN** the source has a syntax error on its third line
- **THEN** the playground shows the error message with line 3, together with the error image

#### Scenario: Graph-based diagram

- **WHEN** a visitor types `@startuml`, `class A`, `class B`, `A --> B` and `@enduml` in the editor
- **THEN** the playground shows the rendered class diagram as SVG, with `B` below `A` and an arrow from `A` to `B`

#### Scenario: Example diagram

- **WHEN** a visitor selects an example diagram
- **THEN** its source replaces the source of the editor and is rendered

#### Scenario: Download

- **WHEN** a visitor downloads the result after a successful rendering
- **THEN** the browser saves the rendered SVG document
