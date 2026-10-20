# Spec Delta

## Purpose

Supports PlantUML v1.2026.8 state diagrams under `@startuml`, with the same syntax and Graphviz-based layout.

## ADDED Requirements

### Requirement: States

The system SHALL support states declared with `state`, display names with `as` (`state "Long name" as Code` and `state Code as "Long name"`), description lines added with `state Code : text` or `Code : text`, stereotypes, tags, URLs, background colors, and line colors and styles (`##[dashed]red`). A state named only in a transition SHALL be created as a plain state. Each state SHALL be drawn as upstream draws it, and the diagram type SHALL be state.

#### Scenario: States and transitions

- **WHEN** a diagram contains `[*] --> Idle`, `Idle --> Running : start` and `Running --> [*]`
- **THEN** the start is drawn as a filled circle, the end as a circled dot, `Idle` and `Running` as rounded boxes, an arrow labeled `start` goes from `Idle` to `Running`, and the diagram type is state

#### Scenario: Display names and descriptions

- **WHEN** a diagram contains `state Idle : waiting for input`, `Idle : no job`, `state "Waiting for input" as Wait` and `Idle --> Wait`
- **THEN** `Idle` is drawn with the lines `waiting for input` and `no job` below its name, and `Wait` is drawn with the name `Waiting for input`

#### Scenario: State colors

- **WHEN** a diagram contains `state Error #pink ##[dashed]red` and `[*] --> Error`
- **THEN** `Error` is drawn with a pink background and a dashed red border

### Requirement: Pseudo states

The system SHALL support upstream's pseudo states: `[*]` as the start (on the left of a transition) or the end (on the right) of the diagram or of the enclosing composite state, the history states `[H]` and `[H*]` of the enclosing state and `Code[H]` and `Code[H*]` of a named state, the synchronization bars `==name==`, and the states declared with the stereotypes `<<choice>>`, `<<fork>>`, `<<join>>`, `<<start>>`, `<<end>>`, `<<history>>` and `<<history*>>`.

#### Scenario: Choice, fork and join

- **WHEN** a diagram contains `state c <<choice>>`, `state f <<fork>>`, `state j <<join>>`, `[*] --> f`, `f --> A`, `f --> B`, `A --> j`, `B --> j`, `j --> c`, `c --> Done : [ok]` and `c --> A : [retry]`
- **THEN** `c` is drawn as a diamond, `f` and `j` as black bars, and the transitions leave `c` with the labels `[ok]` and `[retry]`

#### Scenario: History state

- **WHEN** a diagram contains a `state Active {` block with `[*] --> Running` and `Running --> Waiting`, closed by `}`, followed by `Active --> Paused : pause` and `Paused --> Active[H] : resume`
- **THEN** a history state drawn as a circle with `H` is placed inside `Active`, and the transition labeled `resume` goes from `Paused` to it

#### Scenario: Synchronization bar

- **WHEN** a diagram contains `state A`, `state B`, `A --> ==sync==`, `B --> ==sync==` and `==sync== --> C`
- **THEN** one black bar is drawn for `sync`, with transitions from `A` and `B` into it and one from it to `C`

### Requirement: Transitions

The system SHALL support upstream's transition syntax between states: `->` and `-->`, the reversed form (`B <-- A`), direction hints (`-left->`, `-up->`, ...), length through repeated dashes, colors and line styles in brackets (`-[#blue,dashed]->`), labels, stereotypes, the `x` start and the `o` end. With the `node` label style, set by `-[node]->` or `skinparam stateDiagramEdgeLabelStyle node`, the label of a transition SHALL be drawn as a node between the two states, as upstream draws it, and the output SHALL be the same on every render.

#### Scenario: Direction, style and label

- **WHEN** a diagram contains `state A`, `A -left-> B : back` and `A -[#blue,dashed]-> C`
- **THEN** `B` is placed to the left of `A` with an arrow labeled `back` from `A` to `B`, and a dashed blue arrow goes from `A` to `C`

#### Scenario: Reversed transition

- **WHEN** a diagram contains `state A`, `state B` and `B <-- A : back`
- **THEN** an arrow labeled `back` goes from `A` to `B`

#### Scenario: Transition label as a node

- **WHEN** a diagram contains `skinparam stateDiagramEdgeLabelStyle node`, `state A` and `A --> B : go`, and it is rendered twice
- **THEN** `go` is drawn as a node between `A` and `B`, with a line from `A` to it and an arrow from it to `B`, and both renders are identical

### Requirement: Composite states

The system SHALL support composite states declared with a block (`state Code {` ... `}`, or closed by `end state`), with the same name, color, stereotype and URL options as states, nested composite states, states named with dots (`state A.x` declares `x` inside `A`), and `frame` groups. A composite state that has no transition crossing its border SHALL be laid out on its own and drawn as one state of its parent, as upstream does.

#### Scenario: Composite state with its own start and end

- **WHEN** a diagram contains a `state Active {` block with `[*] --> Running` and `Running --> [*]`, closed by `}`, followed by `[*] --> Active` and `Active --> [*]`
- **THEN** `Running` is drawn inside `Active` between a start and an end of its own, and the start and end of the diagram are drawn outside `Active`

#### Scenario: Dotted state names

- **WHEN** a diagram contains `state A.x`, `state A.y` and `x --> y`
- **THEN** `x` and `y` are drawn inside the composite state `A`, with an arrow from `x` to `y`

### Requirement: Concurrent regions

The system SHALL support concurrent regions inside a composite state, separated by lines of dashes (`--`), which stack the regions vertically, or of pipes (`||`), which place them side by side, each drawn with upstream's dashed separator. A transition from a state inside a region to a state outside of it SHALL be reported as an error, as upstream does.

#### Scenario: Stacked regions

- **WHEN** a diagram contains a `state Active {` block with `[*] --> A1`, `--` and `[*] --> B1`, closed by `}`
- **THEN** `A1` and `B1` are drawn in two regions of `Active`, one above the other, separated by a dashed horizontal line

#### Scenario: Side by side regions

- **WHEN** a diagram contains a `state Active {` block with `[*] --> A1`, `||` and `[*] --> B1`, closed by `}`
- **THEN** `A1` and `B1` are drawn in two regions of `Active`, side by side, separated by a dashed vertical line

#### Scenario: Transition out of a region

- **WHEN** a diagram contains a `state Active {` block with `state A`, `--` and `state B`, closed by `}`, followed by `state Other` and `B --> Other`
- **THEN** the system reports the error `The state B cannot be used here.` on the line of `B --> Other`

### Requirement: Entry and exit points

The system SHALL support states with the stereotypes `<<entryPoint>>`, `<<exitPoint>>`, `<<inputPin>>`, `<<outputPin>>`, `<<expansionInput>>` and `<<expansionOutput>>` declared inside a composite state, drawn on the border of that state as upstream draws them, and transitions to and from them.

#### Scenario: Entry and exit points

- **WHEN** a diagram contains a `state Somp {` block with `state entry1 <<entryPoint>>`, `state exit1 <<exitPoint>>`, `state Inner`, `entry1 --> Inner` and `Inner --> exit1`, closed by `}`, followed by `[*] --> entry1` and `exit1 --> [*]`
- **THEN** `entry1` and `exit1` are drawn on the border of `Somp`, and the transitions pass through them

### Requirement: Notes, URLs and visibility

The system SHALL support notes attached to states and to transitions, floating notes, URLs on states (`[[url]]` and `url for X is [[url]]`), the `hide`, `show` and `remove` commands with the same scoping as upstream, and `hide empty description`, which draws the states without description lines with their name only.

#### Scenario: Note on a state

- **WHEN** a diagram contains `state Idle` and `note right of Idle : waiting`
- **THEN** a note labeled `waiting` is drawn to the right of `Idle`, with a pointer towards `Idle`

#### Scenario: Hide empty description

- **WHEN** a diagram contains `hide empty description`, `state A`, `state B : busy` and `A --> B`
- **THEN** `A` is drawn with its name only, without a separator line, and `B` with its name and the line `busy`

### Requirement: Map and JSON elements in state diagrams

The system SHALL support the `map` and `json` elements in state diagrams, as upstream does.

#### Scenario: Map element in a state diagram

- **WHEN** a diagram contains `state App`, a `map Config {` block with the line `port => 8080` closed by `}`, and `App --> Config`
- **THEN** `Config` is drawn as a map with the row `port` and `8080`, linked from `App`

### Requirement: State diagram direction

The system SHALL support `left to right direction` and `top to bottom direction` in state diagrams, with the same effect on the layout as upstream.

#### Scenario: Left to right states

- **WHEN** a diagram contains `left to right direction`, `[*] --> Idle` and `Idle --> Busy`
- **THEN** `Busy` is placed to the right of `Idle`

### Requirement: State output formats

State diagrams SHALL render to SVG and PNG. TXT and UTXT output SHALL report that the format is not supported for state diagrams, as it does for class and description diagrams.

#### Scenario: TXT for a state diagram

- **WHEN** a state diagram is rendered as TXT
- **THEN** the system reports that TXT output is not supported for state diagrams and produces no output file

### Requirement: State graph layout dependency

State diagrams SHALL be laid out through the graph-layout capability and SHALL report the graph-layout errors when Graphviz `dot` is unavailable or fails.

#### Scenario: State diagram without dot

- **WHEN** no `dot` executable can be found and a state diagram is rendered
- **THEN** the system produces the error diagram stating that Graphviz `dot` was not found
