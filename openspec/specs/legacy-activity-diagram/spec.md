# legacy-activity-diagram Specification

## Purpose

Supports PlantUML v1.2026.8 activity diagrams in the legacy syntax (`(*) --> "First step"`) under `@startuml`, with the same syntax and Graphviz-based layout.

## Requirements

### Requirement: Activities and transitions

The system SHALL support upstream's legacy activity syntax: the start `(*)`, the ends `(*)`, `(*top)` and numbered ends such as `(*1)`, activities written as quoted text (on one line or on several lines up to the closing quote), with a code (`"Long name" as L`) or as bare names, stereotypes and background colors on activities, and transitions from a given activity or, when the source is omitted, from the last activity. A transition SHALL support a label in brackets (`-->[label]`), a dotted body (`..>`), direction hints (`-left->`, `-up->`, ...), styles and colors in brackets, a URL, `-*->` for a transition without layout constraint, and `in Partition` after its target. Activities SHALL be drawn as rounded boxes, and the diagram type SHALL be activity.

#### Scenario: Start, activities and end

- **WHEN** a diagram contains `(*) --> "First step"`, `"First step" --> "Second step"` and `"Second step" --> (*)`
- **THEN** the start is drawn as a filled circle, `First step` and `Second step` as rounded boxes, the end as a circled dot, arrows link them in that order, and the diagram type is activity

#### Scenario: Transitions from the last activity

- **WHEN** a diagram contains `(*) --> "Check order"`, `--> "Pay"` and `--> (*)`
- **THEN** arrows go from the start to `Check order`, from `Check order` to `Pay` and from `Pay` to the end

#### Scenario: Activity with a code

- **WHEN** a diagram contains `(*) --> "Long name" as L` and `L --> (*)`
- **THEN** one activity labeled `Long name` is drawn, with an arrow from the start to it and one from it to the end

#### Scenario: Activity on several lines

- **WHEN** a diagram contains `(*) --> "First line`, then `second line"`, then `--> (*)`
- **THEN** one activity is drawn with the lines `First line` and `second line`

#### Scenario: Transition styles

- **WHEN** a diagram contains `(*) -left-> "A"`, `"A" ..> "B"` and `"B" -->[label] (*)`
- **THEN** `A` is placed to the left of the start, a dotted arrow goes from `A` to `B`, and an arrow labeled `label` goes from `B` to the end

#### Scenario: Activity colors and stereotypes

- **WHEN** a diagram contains `(*) --> "Pay" <<important>> #pink` and `--> (*)`
- **THEN** `Pay` is drawn with a pink background

### Requirement: Branches

The system SHALL support branches written `if "label" then` ... `else` ... `endif` (and the other `if` forms upstream accepts, including a code with `as`), linked from the last activity or from an activity, the start or a bar written before `if`, nested branches, and transitions with labels inside each part. A branch SHALL be drawn as a diamond, with its label on the transition into it.

#### Scenario: Branch with two outcomes

- **WHEN** a diagram contains `(*) --> "Check"`, `if "valid?" then`, `-->[yes] "Ship"`, `else`, `-->[no] "Reject"` and `endif`
- **THEN** a diamond is drawn after `Check`, the transition into it is labeled `valid?`, and transitions labeled `yes` and `no` go from it to `Ship` and `Reject`

### Requirement: Synchronization bars

The system SHALL support synchronization bars written `===name===`, as the source or the target of a transition, drawn as black bars.

#### Scenario: Fork and join

- **WHEN** a diagram contains `(*) --> ===B1===`, `===B1=== --> "Pack"`, `===B1=== --> "Bill"`, `"Pack" --> ===B2===`, `"Bill" --> ===B2===` and `===B2=== --> (*)`
- **THEN** two bars are drawn, the first with arrows to `Pack` and `Bill` and the second with arrows from them

### Requirement: Partitions

The system SHALL support partitions declared with `partition Name {` ... `}` or `partition Name` ... `end partition`, with a color and a stereotype, and activities put in a partition with `in Name`. A partition SHALL be drawn as a box titled with its name around its activities.

#### Scenario: Partition block

- **WHEN** a diagram contains `partition Payment {`, `(*) --> "Pay"`, `"Pay" --> (*)` and `}`
- **THEN** `Pay` is drawn inside a box titled `Payment`

#### Scenario: Activity in a partition

- **WHEN** a diagram contains `(*) --> "Pay" in Payment` and `"Pay" --> (*)`
- **THEN** `Pay` is drawn inside a box titled `Payment`

### Requirement: Notes

The system SHALL support notes on the last activity, on one line (`note right: text`) or on several lines up to `end note`, on the left, right, top or bottom, with a color, and notes on transitions.

#### Scenario: Note on one line

- **WHEN** a diagram contains `(*) --> "Pay"`, `note right: by card` and `"Pay" --> (*)`
- **THEN** a note labeled `by card` is drawn to the right of `Pay`, linked to it

#### Scenario: Note on several lines

- **WHEN** a diagram contains `(*) --> "Pay"`, `note left`, `first line`, `second line`, `end note` and `"Pay" --> (*)`
- **THEN** a note with the lines `first line` and `second line` is drawn to the left of `Pay`, linked to it

### Requirement: Octagon activities

The system SHALL draw activities as octagons when `skinparam activityShape octagon` is set, as upstream does.

#### Scenario: Octagon shape

- **WHEN** a diagram contains `skinparam activityShape octagon`, `(*) --> "Pay"` and `--> (*)`
- **THEN** `Pay` is drawn as an octagon

### Requirement: Legacy activity output formats

Legacy activity diagrams SHALL render to SVG and PNG. TXT and UTXT output SHALL report that the format is not supported for activity diagrams, as it does for the activity diagrams of the new syntax.

#### Scenario: TXT for a legacy activity diagram

- **WHEN** a legacy activity diagram is rendered as TXT
- **THEN** the system reports that TXT output is not supported for activity diagrams and produces no output file

### Requirement: Legacy activity graph layout dependency

Legacy activity diagrams SHALL be laid out through the graph-layout capability and SHALL report the graph-layout errors when Graphviz `dot` is unavailable or fails.

#### Scenario: Legacy activity diagram without dot

- **WHEN** no `dot` executable can be found and a legacy activity diagram is rendered
- **THEN** the system produces the error diagram stating that Graphviz `dot` was not found
