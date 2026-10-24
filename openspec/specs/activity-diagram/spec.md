# activity-diagram Specification

## Purpose

Supports PlantUML v1.2026.9 activity diagrams in the new (v3) syntax under `@startuml`, with the same syntax and layout as upstream, which lays out activity diagrams without Graphviz.

## Requirements

### Requirement: Actions and terminals

The system SHALL support `start`, `stop`, `end`, `kill` and `detach`, actions written as `:text;` (including multi-line text and creole), and the box shapes upstream selects with a stereotype after the action, such as `<<input>>`, `<<output>>`, `<<procedure>>`, `<<load>>`, `<<save>>`, `<<continuous>>`, `<<task>>`, `<<object>>`, `<<objectSignal>>`, `<<trigger>>`, `<<sendSignal>>`, `<<acceptEvent>>` and `<<timeEvent>>`. A color or a stereotype written before the `:` of a one-line action (`#red:X;`, `<<foo>>:X;`) SHALL be ignored with upstream's warning, and the lines that follow SHALL be parsed as usual.

#### Scenario: Simple flow

- **WHEN** a diagram contains `start`, `:Read input;`, `:Process;` and `stop`
- **THEN** a start node, two action boxes and a stop node are drawn top to bottom and connected by arrows

#### Scenario: SDL shape

- **WHEN** a diagram contains `:Receive order; <<input>>`
- **THEN** the action is drawn with upstream's input shape, a box notched on the right

#### Scenario: Color before an action

- **WHEN** a diagram contains `#red:X;` and `:Y;`
- **THEN** the actions `X` and `Y` are drawn with the default fill, and the warning `This syntax is deprecated and the color is ignored: write ':X; <<#red>>' instead of '#red:X;'` is shown above the diagram

### Requirement: Conditionals

The system SHALL support `if (cond) then (label)`, `elseif`, `else` and `endif`, including nesting, and `switch`/`case`/`endswitch`. A color in a stereotype at the end of an `if` or `elseif` line, including the `is` form (`if (c?) is (v) then <<#palegreen>>`), SHALL fill the diamond. A color written before `if`, `elseif` or `switch` (`#red:if (c?) then (yes)`) SHALL be ignored with the warning `This syntax is deprecated, you must add <<#red>> at the end of the line`.

#### Scenario: If with else

- **WHEN** a diagram contains `if (valid?) then (yes)`, `:Save;`, `else (no)`, `:Reject;` and `endif`
- **THEN** a decision diamond labeled `valid?` is drawn with branches labeled `yes` and `no`, merging after the branches

#### Scenario: Switch

- **WHEN** a diagram contains a `switch (kind)` with three `case` branches
- **THEN** a decision with three labeled branches is drawn, merging at `endswitch`

#### Scenario: Stereotype color of a condition

- **WHEN** a diagram contains `if (c?) is (v) then <<#palegreen>>`, `:A;`, `else (no)`, `:B;` and `endif`
- **THEN** the diamond labeled `c?` is filled palegreen

#### Scenario: Color before a condition

- **WHEN** a diagram contains `#red:if (c?) then (yes)`, `:A;`, `else (no)`, `:B;` and `endif`
- **THEN** the diamond labeled `c?` is drawn with the default fill, and the warning `This syntax is deprecated, you must add <<#red>> at the end of the line` is shown above the diagram

### Requirement: Loops

The system SHALL support `repeat`...`repeat while (cond) is (label)` with optional `backward` actions, `while (cond) is (label)`...`endwhile (label)`, and `break` inside loops. A color written before `repeat` or `while` SHALL be ignored with upstream's warning.

#### Scenario: While loop

- **WHEN** a diagram contains `while (more?) is (yes)`, `:Read;` and `endwhile (no)`
- **THEN** a loop with the condition `more?` is drawn, with a back edge to the condition and an exit labeled `no`

#### Scenario: Color before repeat

- **WHEN** a diagram contains `#red:repeat`, `:A;` and `repeat while (more?)`
- **THEN** the loop is drawn without the color, and the warning `This syntax is deprecated and the color is ignored: write 'repeat <<#red>>' instead of '#red:repeat'` is shown above the diagram

### Requirement: Parallel processing

The system SHALL support `fork`, `fork again`, `end fork` and `end merge`, and `split`, `split again` and `end split`.

#### Scenario: Fork

- **WHEN** a diagram contains `fork`, `:A;`, `fork again`, `:B;` and `end fork`
- **THEN** `A` and `B` are drawn side by side between two synchronization bars

### Requirement: Swimlanes and grouping

The system SHALL support swimlanes (`|Lane|` and `|#color|Lane|`), `partition` blocks, `group` blocks, and the other grouping blocks upstream supports in activity diagrams. A grouping block opened without `{` or closed by `end group` SHALL produce upstream's warning.

#### Scenario: Two swimlanes

- **WHEN** a diagram contains `|Customer|`, `:Order;`, `|Shop|` and `:Ship;`
- **THEN** two lanes titled `Customer` and `Shop` are drawn, with `Order` in the first lane and `Ship` in the second

#### Scenario: Partition without a bracket

- **WHEN** a diagram contains `partition "My part"`, `:A;` and `}`
- **THEN** the partition `My part` is drawn around `A`, and the warning `Missing '{': write 'partition "My part" {' and close it with '}'` is shown above the diagram

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
