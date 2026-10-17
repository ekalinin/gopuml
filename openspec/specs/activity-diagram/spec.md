# activity-diagram Specification

## Purpose

Supports PlantUML v1.2026.8 activity diagrams in the new (v3) syntax under `@startuml`, with the same syntax and layout as upstream, which lays out activity diagrams without Graphviz.

## Requirements

### Requirement: Actions and terminals

The system SHALL support `start`, `stop`, `end`, `kill` and `detach`, actions written as `:text;` (including multi-line text and creole), and the box shapes upstream selects with a stereotype after the action, such as `<<input>>`, `<<output>>`, `<<procedure>>`, `<<load>>`, `<<save>>`, `<<continuous>>`, `<<task>>`, `<<object>>`, `<<objectSignal>>`, `<<trigger>>`, `<<sendSignal>>`, `<<acceptEvent>>` and `<<timeEvent>>`.

#### Scenario: Simple flow

- **WHEN** a diagram contains `start`, `:Read input;`, `:Process;` and `stop`
- **THEN** a start node, two action boxes and a stop node are drawn top to bottom and connected by arrows

#### Scenario: SDL shape

- **WHEN** a diagram contains `:Receive order; <<input>>`
- **THEN** the action is drawn with upstream's input shape, a box notched on the right

### Requirement: Conditionals

The system SHALL support `if (cond) then (label)`, `elseif`, `else` and `endif`, including nesting, and `switch`/`case`/`endswitch`.

#### Scenario: If with else

- **WHEN** a diagram contains `if (valid?) then (yes)`, `:Save;`, `else (no)`, `:Reject;` and `endif`
- **THEN** a decision diamond labeled `valid?` is drawn with branches labeled `yes` and `no`, merging after the branches

#### Scenario: Switch

- **WHEN** a diagram contains a `switch (kind)` with three `case` branches
- **THEN** a decision with three labeled branches is drawn, merging at `endswitch`

### Requirement: Loops

The system SHALL support `repeat`...`repeat while (cond) is (label)` with optional `backward` actions, `while (cond) is (label)`...`endwhile (label)`, and `break` inside loops.

#### Scenario: While loop

- **WHEN** a diagram contains `while (more?) is (yes)`, `:Read;` and `endwhile (no)`
- **THEN** a loop with the condition `more?` is drawn, with a back edge to the condition and an exit labeled `no`

### Requirement: Parallel processing

The system SHALL support `fork`, `fork again`, `end fork` and `end merge`, and `split`, `split again` and `end split`.

#### Scenario: Fork

- **WHEN** a diagram contains `fork`, `:A;`, `fork again`, `:B;` and `end fork`
- **THEN** `A` and `B` are drawn side by side between two synchronization bars

### Requirement: Swimlanes and grouping

The system SHALL support swimlanes (`|Lane|` and `|#color|Lane|`), `partition` blocks, `group` blocks, and the other grouping blocks upstream supports in activity diagrams.

#### Scenario: Two swimlanes

- **WHEN** a diagram contains `|Customer|`, `:Order;`, `|Shop|` and `:Ship;`
- **THEN** two lanes titled `Customer` and `Shop` are drawn, with `Order` in the first lane and `Ship` in the second

### Requirement: Arrows, connectors and notes

The system SHALL support arrow labels (`-> label;`), arrow colors and styles (`-[#red,dashed]->`), connectors (`(A)`), `note left`/`note right` on actions, and `floating note`.

#### Scenario: Labeled arrow

- **WHEN** a diagram contains `:A;`, `-> next;` and `:B;`
- **THEN** the arrow from `A` to `B` is labeled `next`

### Requirement: Built-in layout

The system SHALL lay out activity diagrams the same way as upstream's activity layout, without Graphviz. The result SHALL match upstream within the parity tolerance defined by upstream-compatibility.

#### Scenario: Rendering without Graphviz

- **WHEN** no `dot` executable is available and an activity diagram is rendered
- **THEN** the diagram renders without error
