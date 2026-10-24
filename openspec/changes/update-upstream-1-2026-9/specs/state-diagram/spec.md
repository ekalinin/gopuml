# Spec Delta

## MODIFIED Requirements

### Requirement: Pseudo states

The system SHALL support upstream's pseudo states: `[*]` as the start (on the left of a transition) or the end (on the right) of the diagram or of the enclosing composite state, the history states `[H]` and `[H*]` of the enclosing state and `Code[H]` and `Code[H*]` of a named state, the synchronization bars `==name==`, and the states declared with the stereotypes `<<choice>>`, `<<fork>>`, `<<join>>`, `<<start>>`, `<<end>>`, `<<junction>>`, `<<history>>` and `<<history*>>`. A junction SHALL be drawn as a filled circle smaller than the start. History states SHALL take the style of the `.history` stereotype selector, and deep history states also that of `.deepHistory`, and the colors given to the state, as upstream does.

#### Scenario: Choice, fork and join

- **WHEN** a diagram contains `state c <<choice>>`, `state f <<fork>>`, `state j <<join>>`, `[*] --> f`, `f --> A`, `f --> B`, `A --> j`, `B --> j`, `j --> c`, `c --> Done : [ok]` and `c --> A : [retry]`
- **THEN** `c` is drawn as a diamond, `f` and `j` as black bars, and the transitions leave `c` with the labels `[ok]` and `[retry]`

#### Scenario: History state

- **WHEN** a diagram contains a `state Active {` block with `[*] --> Running` and `Running --> Waiting`, closed by `}`, followed by `Active --> Paused : pause` and `Paused --> Active[H] : resume`
- **THEN** a history state drawn as a circle with `H` is placed inside `Active`, and the transition labeled `resume` goes from `Paused` to it

#### Scenario: Synchronization bar

- **WHEN** a diagram contains `state A`, `state B`, `A --> ==sync==`, `B --> ==sync==` and `==sync== --> C`
- **THEN** one black bar is drawn for `sync`, with transitions from `A` and `B` into it and one from it to `C`

#### Scenario: Junction

- **WHEN** a diagram contains `state j <<junction>>`, `[*] --> j`, `j --> A : [x > 0]` and `j --> B : [x <= 0]`
- **THEN** `j` is drawn as a filled circle 12 wide, smaller than the start circle, and the transitions leave it with the labels `[x > 0]` and `[x <= 0]`

#### Scenario: Styled history states

- **WHEN** a diagram contains a style that sets `BackGroundColor palegreen` for `.history` under `stateDiagram`, `state A`, `state B`, `state hh <<history>>`, `state hd <<history*>>`, `[H] --> A`, `[H*] --> B`, `hh --> A` and `hd --> B`
- **THEN** the four history circles are filled palegreen

### Requirement: Transitions

The system SHALL support upstream's transition syntax between states: `->` and `-->`, the reversed form (`B <-- A`), direction hints (`-left->`, `-up->`, ...), length through repeated dashes, colors and line styles in brackets (`-[#blue,dashed]->`), labels, stereotypes, the `x` start and the `o` end. With the `node` label style, set by `-[node]->` or `skinparam stateDiagramEdgeLabelStyle node`, the label of a transition SHALL be drawn as a node between the two states, as upstream draws it, and the output SHALL be the same on every render. When `hide` or `remove` applies to a state at either end of such a transition, the label node SHALL be hidden or removed too.

#### Scenario: Direction, style and label

- **WHEN** a diagram contains `state A`, `A -left-> B : back` and `A -[#blue,dashed]-> C`
- **THEN** `B` is placed to the left of `A` with an arrow labeled `back` from `A` to `B`, and a dashed blue arrow goes from `A` to `C`

#### Scenario: Reversed transition

- **WHEN** a diagram contains `state A`, `state B` and `B <-- A : back`
- **THEN** an arrow labeled `back` goes from `A` to `B`

#### Scenario: Transition label as a node

- **WHEN** a diagram contains `skinparam stateDiagramEdgeLabelStyle node`, `state A` and `A --> B : go`, and it is rendered twice
- **THEN** `go` is drawn as a node between `A` and `B`, with a line from `A` to it and an arrow from it to `B`, and both renders are identical

#### Scenario: Removed state with a label node

- **WHEN** a diagram contains `skinparam stateDiagramEdgeLabelStyle node`, `state Idle`, `state Charging`, `state Faulted $fault`, `Idle --> Charging : start`, `Charging --> Faulted : fault` and `remove $fault`
- **THEN** neither `Faulted` nor the label node `fault` is drawn, and the label node `start` is drawn between `Idle` and `Charging`
