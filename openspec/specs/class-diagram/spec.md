# class-diagram Specification

## Purpose

Supports PlantUML v1.2026.8 class diagrams under `@startuml`, with the same syntax and Graphviz-based layout.

## Requirements

### Requirement: Class-like elements

The system SHALL support every element keyword upstream accepts in class diagrams (including `class`, `abstract class`, `interface`, `enum`, `annotation`, `entity`, `exception`, `metaclass`, `protocol`, `struct`, `stereotype`, `circle` and `diamond`). The object-diagram elements `object`, `map` and `json`, which upstream parses as class diagrams, SHALL be supported too. Each kind SHALL get upstream's header spot, generic parameters (`class Box<T>`), stereotypes with optional custom spots (`<< (S,#FF7700) Singleton >>`), and per-element colors.

#### Scenario: Interface and enum

- **WHEN** a diagram declares `interface Shape` and an `enum Color` whose body lists `RED` and `GREEN` on separate lines
- **THEN** `Shape` is drawn with the interface spot, and `Color` with the enum spot and the values `RED` and `GREEN`

#### Scenario: Object in a class diagram

- **WHEN** a diagram declares `object Alice` and `Alice : age = 30`
- **THEN** an object box titled `Alice` with the field `age = 30` is drawn, and the diagram type is class

#### Scenario: Generic class with stereotype

- **WHEN** a diagram declares `class Repo<T> <<service>>`
- **THEN** the class header shows `Repo`, the generic parameter `T` and the stereotype `service`

### Requirement: Members

The system SHALL support fields and methods declared in a `{...}` body or with `Element : member` lines, visibility markers (`-`, `#`, `~`, `+`) drawn as upstream's icons, the `{static}`, `{abstract}`, `{field}` and `{method}` modifiers, and the separators `--`, `..`, `==` and `__` with optional titles.

#### Scenario: Visibility and static members

- **WHEN** a class body contains `-id : int` and `{static} +create() : Order`
- **THEN** `id` is drawn with the private icon and `create` with the public icon, underlined as static

### Requirement: Relations

The system SHALL support upstream's relation syntax:

- extension, composition, aggregation, dependency, realization and plain links, solid or dotted
- the heads `#`, `x`, `}`, `+` and `^`
- labels, cardinalities on both ends, label arrows (`: label >`)
- direction hints (`-up->`, `-left->`, ...) and length through repeated dashes
- `[hidden]`, and colors and line styles in brackets (`-[#red,dashed]->`)
- association classes (`(A, B) .. C`)

#### Scenario: Composition with cardinalities

- **WHEN** a diagram contains `Order "1" *-- "many" Line : contains`
- **THEN** a composition is drawn from `Order` to `Line`, with cardinality `1` near `Order`, `many` near `Line`, and the label `contains`

#### Scenario: Direction hint

- **WHEN** a diagram contains `A -left-> B`
- **THEN** `B` is placed to the left of `A`

### Requirement: Packages and namespaces

The system SHALL support `package` and `namespace` blocks, nesting, package display styles (for example `<<Node>>`, `<<Folder>>`, `<<Frame>>`, `<<Cloud>>`, `<<Database>>`, `<<Rectangle>>`), `set separator`, and `together { }` grouping.

#### Scenario: Nested package

- **WHEN** a diagram contains a `package shop {` block holding a `package model {` block that declares `class Order`, each closed by `}` on its own line
- **THEN** `Order` is drawn inside `model`, which is drawn inside `shop`

### Requirement: Notes

The system SHALL support notes attached to elements (`note left of`, `note right of`, `note top of`, `note bottom of`), drawn as upstream's note shape with a pointer towards the element, floating notes (`note "text" as N`) linked to elements, and notes on links (`note on link`).

#### Scenario: Note attached to a class

- **WHEN** a diagram contains `class A` and `note right of A : important`
- **THEN** a note labeled `important` is drawn to the right of `A`, with a pointer towards `A`

### Requirement: Hide, show and remove

The system SHALL support `hide` and `show` commands for members, empty members, circles (spots), stereotypes and elements, including `hide empty members`, `hide circle`, `hide <<stereotype>>`, `hide @unlinked` and `remove` with the same scoping as upstream.

#### Scenario: Hide empty members

- **WHEN** a diagram contains `class A` with no members and `hide empty members`
- **THEN** `A` is drawn as a header box with no member compartments

### Requirement: Diagram direction

The system SHALL support `left to right direction` and `top to bottom direction`, with the same effect on the layout as upstream.

#### Scenario: Left to right

- **WHEN** a diagram contains `left to right direction` and `A --> B`
- **THEN** `B` is placed to the right of `A`

### Requirement: Graph layout dependency

Class diagrams SHALL be laid out through the graph-layout capability and SHALL report the graph-layout errors when Graphviz `dot` is unavailable or fails.

#### Scenario: Rendering requires dot

- **WHEN** a class diagram is rendered and `dot` is available
- **THEN** the node positions come from `dot`, as in upstream's output
