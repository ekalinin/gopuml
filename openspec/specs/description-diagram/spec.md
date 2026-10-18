# description-diagram Specification

## Purpose

Supports PlantUML v1.2026.8 description diagrams (use case, component and deployment diagrams) under `@startuml`, with the same syntax and Graphviz-based layout.

## Requirements

### Requirement: Description elements

The system SHALL support every element keyword upstream accepts in description diagrams except `archimate`: `actor`, `usecase`, `component`, `node`, `database`, `artifact`, `rectangle`, `card`, `cloud`, `folder`, `frame`, `file`, `package`, `queue`, `stack`, `storage`, `agent`, `person`, `boundary`, `control`, `entity`, `interface`, `circle`, `collections`, `hexagon`, `label`, `process` and `action`. It SHALL also support the business variants `actor/` and `usecase/`, the short forms `:Actor:`, `(Use case)`, `[Component]` and `() Interface`, display names with `as` aliases, stereotypes, colors, and element bodies spanning several lines in `[` ... `]`. Each element SHALL be drawn with upstream's symbol for its keyword, and the diagram type SHALL be description.

#### Scenario: Actor and use case

- **WHEN** a diagram contains `actor User`, `usecase Login` and `User --> Login`
- **THEN** `User` is drawn as a stick figure, `Login` as an ellipse, an arrow goes from `User` to `Login`, and the diagram type is description

#### Scenario: Short forms

- **WHEN** a diagram contains `:Admin: --> (Reset password)` and `[API] --> () REST`
- **THEN** `Admin` is drawn as an actor, `Reset password` as a use case, `API` as a component and `REST` as an interface circle

#### Scenario: Deployment symbols

- **WHEN** a diagram contains `node Server`, `database Orders` and `Server --> Orders`
- **THEN** `Server` is drawn with the node symbol and `Orders` with the database symbol

#### Scenario: Element body on several lines

- **WHEN** a diagram contains `rectangle Summary [`, the lines `first line` and `second line`, and `]`
- **THEN** one rectangle is drawn with both lines as its text

### Requirement: Description links

The system SHALL support upstream's link syntax between description elements: solid and dotted lines, the heads upstream accepts on either end, direction hints (`-up->`, `-left->`, ...), length through repeated dashes or dots, labels, and colors and line styles in brackets (`-[#red,dashed]->`).

#### Scenario: Dotted link with direction and label

- **WHEN** a diagram contains `component A`, `component B` and `A .left.> B : uses`
- **THEN** a dotted arrow labeled `uses` goes from `A` to `B`, and `B` is placed to the left of `A`

### Requirement: Containers

The system SHALL support containers declared with a block (`node Server {` ... `}`) for the keywords upstream allows as containers, drawn with the container's symbol, nested containers, and `together { }` grouping.

#### Scenario: Component inside a node

- **WHEN** a diagram contains a `node Server {` block that declares `component Web`, closed by `}` on its own line
- **THEN** `Web` is drawn inside `Server`, which is drawn with the node symbol

### Requirement: Ports

The system SHALL support `port`, `portin` and `portout` elements declared inside a component, drawn on the border of the component as upstream draws them, and links between ports.

#### Scenario: Linked ports

- **WHEN** a diagram declares `component C1 {` with `portout p1`, and `component C2 {` with `portin p2`, each closed by `}`, followed by `p1 --> p2`
- **THEN** `p1` is drawn on the border of `C1`, `p2` on the border of `C2`, and an arrow goes from `p1` to `p2`

### Requirement: Notes, URLs and visibility

The system SHALL support notes attached to description elements and to links, floating notes, URLs on elements (`[[url]]` and `url for X is [[url]]`), and the `hide`, `show` and `remove` commands with the same scoping as upstream.

#### Scenario: Note attached to an actor

- **WHEN** a diagram contains `actor User` and `note right of User : signs in`
- **THEN** a note labeled `signs in` is drawn to the right of `User`, with a pointer towards `User`

### Requirement: Map and JSON elements

The system SHALL support the `map` and `json` elements in description diagrams, as upstream does.

#### Scenario: Map element

- **WHEN** a diagram contains `component App`, a `map Config {` block with the line `port => 8080` closed by `}`, and `App --> Config`
- **THEN** `Config` is drawn as a map with the row `port` and `8080`, linked from `App`

### Requirement: Problem Frames elements

The system SHALL support upstream's `domain` and `requirement` elements (`domain "Name" as D`, `requirement "Name" as R`), including the domain stereotypes `<<Machine>>`, `<<Causal>>`, `<<Designed>>`, `<<Lexical>>` and `<<Biddable>>`, each drawn with upstream's shape.

#### Scenario: Machine and requirement

- **WHEN** a diagram contains `domain "Controller" as M <<Machine>>`, `requirement "Keep the temperature" as R` and `M -- R`
- **THEN** `Controller` is drawn as a machine domain, `Keep the temperature` as a requirement, and a line connects them

### Requirement: Description diagram direction

The system SHALL support `left to right direction` and `top to bottom direction` in description diagrams, with the same effect on the layout as upstream.

#### Scenario: Left to right use cases

- **WHEN** a diagram contains `left to right direction`, `actor User` and `User --> (Login)`
- **THEN** `Login` is placed to the right of `User`

### Requirement: Description output formats

Description diagrams SHALL render to SVG and PNG. TXT and UTXT output SHALL report that the format is not supported, naming the diagram type with upstream's name for it ("component"), as it does for class diagrams.

#### Scenario: TXT for a description diagram

- **WHEN** a description diagram is rendered as TXT
- **THEN** the system reports that TXT output is not supported for component diagrams and produces no output file

### Requirement: Description graph layout dependency

Description diagrams SHALL be laid out through the graph-layout capability and SHALL report the graph-layout errors when Graphviz `dot` is unavailable or fails.

#### Scenario: Description diagram without dot

- **WHEN** no `dot` executable can be found and a use case diagram is rendered
- **THEN** the system produces the error diagram stating that Graphviz `dot` was not found
