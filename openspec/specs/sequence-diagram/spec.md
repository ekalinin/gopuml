# sequence-diagram Specification

## Purpose

Supports PlantUML v1.2026.9 sequence diagrams under `@startuml`, with the same syntax and the layout of upstream's Teoz engine, its only sequence engine.

## Requirements

### Requirement: Participants

The system SHALL support the participant kinds:

- `participant`, `actor`, `boundary`, `control`, `entity`, `database`, `collections` and `queue`
- aliases with `as` and quoted display names
- per-participant colors, stereotypes and `order` values
- `box`...`end box` grouping
- implicit declaration on first use in a message
- a `MinimumWidth` style wider than the name, with the name placed in the widened box by the `HorizontalAlignment` style (centered when it is not set)

#### Scenario: Declared participant kinds

- **WHEN** a diagram declares `actor User` and `database DB` and contains `User -> DB : query`
- **THEN** `User` is drawn as an actor, `DB` as a database, and an arrow labeled `query` runs between them

#### Scenario: Implicit participant

- **WHEN** a diagram contains only `Alice -> Bob : hi`
- **THEN** participants `Alice` and `Bob` are drawn in that order

#### Scenario: Name aligned in a widened box

- **WHEN** a diagram contains a style that sets `MinimumWidth 100` and `HorizontalAlignment right` for `participant` under `sequenceDiagram`, `participant "A" as A` and `A -> B : Message`
- **THEN** the box of `A` is widened to the minimum width and the name `A` is drawn on the right side of the box

### Requirement: Messages

The system SHALL support every message arrow form upstream supports for sequence diagrams, including:

- solid and dotted lines
- open, half and lost-message heads (`->>`, `-\`, `-/`, `->x`)
- bidirectional arrows and `o` end circles
- self messages, incoming and outgoing messages (`[->`, `->]`) and short arrows (`?->`)
- colored arrows (`-[#red]>`)
- multi-line labels with creole markup

#### Scenario: Dotted reply

- **WHEN** a diagram contains `Alice -> Bob : request` and `Bob --> Alice : response`
- **THEN** the request is drawn as a solid arrow and the response as a dotted arrow

#### Scenario: Self message

- **WHEN** a diagram contains `Alice -> Alice : think`
- **THEN** a looping arrow from `Alice` back to `Alice` labeled `think` is drawn

### Requirement: Autonumbering

The system SHALL support `autonumber` with optional start, increment and format string, and `autonumber stop` and `autonumber resume`.

#### Scenario: Custom start and format

- **WHEN** a diagram contains `autonumber 10 5 "<b>[000]"` followed by two messages
- **THEN** the messages are labeled `[010]` and `[015]` in bold

### Requirement: Lifelines and activation

The system SHALL support `activate`, `deactivate`, `destroy` and `create`, the inline shortcuts `++`, `--`, `**` and `!!`, and `return`, including nested activations and activation colors.

#### Scenario: Activation with return

- **WHEN** a diagram contains `Alice -> Bob ++ : call` and `return done`
- **THEN** an activation bar is drawn on `Bob`, and a dotted arrow labeled `done` returns to `Alice` and ends the activation

### Requirement: Grouping fragments

The system SHALL support `alt`/`else`, `opt`, `loop`, `par`, `break`, `critical` and `group` fragments closed by `end`, including nesting and fragment labels. The gap between a fragment's frame and the lifeline that follows it on the right SHALL be the same as the gap between the lifeline before it and the frame on the left, as upstream lays it out.

#### Scenario: Alternative fragment

- **WHEN** a diagram contains `alt success`, a message, `else failure`, a message and `end`
- **THEN** a fragment labeled `alt` is drawn with two sections labeled `success` and `failure`

#### Scenario: Right margin of a fragment

- **WHEN** a diagram contains `x -> a : start`, a `group test` holding `a <- a`, `a -> b : 1`, `b -> c` and `c -> c`, then `end` and `c -> d : next`
- **THEN** the gap between the frame's right edge and the lifeline of `d` equals the gap between the lifeline of `x` and the frame's left edge

### Requirement: Notes

The system SHALL support `note left`, `note right` and `note over` (one or more participants), `note across`, the `hnote` and `rnote` shapes, multi-line notes, and notes attached to messages. A note that starts with `&` SHALL be drawn on the same row as the previous one, and notes joined this way on the same participant and position SHALL be stacked one below another. A note that starts with `/` SHALL be drawn on the same row as the previous one, with the warning `The leading '/' note-merge syntax is deprecated; use '&' to place notes on the same row instead.` A `note top` or `note bottom` after `end` or after a `ref` SHALL produce upstream's warning: after `end` the note is kept, after a `ref` it is ignored.

#### Scenario: Note over two participants

- **WHEN** a diagram contains `note over Alice, Bob : shared`
- **THEN** a note labeled `shared` spans the lifelines of `Alice` and `Bob`

#### Scenario: Stacked notes of one participant

- **WHEN** a diagram contains `A -> B : hello`, `note over A : first note`, `& note over A : second note` and `& note over A : third note`
- **THEN** the three notes are drawn over `A` one below another, and none of them overlaps another

#### Scenario: Deprecated note merge

- **WHEN** a diagram contains `Alice -> Bob : hi`, `note over Alice : first` and `/ note over Bob : second`
- **THEN** both notes are drawn on the same row, and the warning `The leading '/' note-merge syntax is deprecated; use '&' to place notes on the same row instead.` is shown above the diagram

#### Scenario: Note position after a fragment

- **WHEN** a diagram contains `group g`, `A -> B`, `end` and `note top : y`
- **THEN** the note `y` is drawn, and the warning `'note top' is not supported after 'end': use 'note left' or 'note right'` is shown above the diagram

### Requirement: Dividers, delays, spacing and references

The system SHALL support dividers (`== text ==`), delays (`...` and `... text ...`), spacing (`|||` and `||N||`) and references (`ref over A, B : text`).

#### Scenario: Divider

- **WHEN** a diagram contains `== Initialization ==` between two messages
- **THEN** a horizontal divider labeled `Initialization` is drawn across all lifelines between the messages

### Requirement: Footbox and pages

The system SHALL support `hide footbox`, and `newpage` with an optional title. Each page SHALL be rendered as a separate image, and the caller SHALL be able to select a page by index.

#### Scenario: Two pages

- **WHEN** a diagram contains messages, then `newpage Part 2`, then more messages
- **THEN** the diagram produces two images, and the second is titled `Part 2`

### Requirement: Teoz engine layout

The system SHALL lay out sequence diagrams the same way as upstream's Teoz engine, the only sequence engine in v1.2026.9. `!pragma teoz` SHALL be accepted with any value and SHALL NOT change the layout, as upstream does.

#### Scenario: Default engine

- **WHEN** a sequence diagram without pragmas is rendered
- **THEN** the layout matches upstream's Teoz output within the parity tolerance defined by upstream-compatibility

#### Scenario: Teoz pragma has no effect

- **WHEN** a sequence diagram contains `!pragma teoz false`
- **THEN** the diagram renders with the Teoz layout and no error is reported

### Requirement: Text rendering

The system SHALL render sequence diagrams as TXT and UTXT text art, following the output-formats capability.

#### Scenario: UTXT output

- **WHEN** `Alice -> Bob : hello` is rendered as UTXT
- **THEN** the output uses Unicode box-drawing characters for the participant boxes and lifelines

### Requirement: SVG element groups

In SVG output, the system SHALL wrap the lifeline, the head and the tail of each participant, and the arrow of each message, in a group with upstream's attributes:

- lifeline, head and tail: the class `participant-lifeline`, `participant participant-head` or `participant participant-tail`, the participant's name (`data-qualified-name`) and uid (`data-entity-uid`), the id `<uid>-lifeline`, `<uid>-head` or `<uid>-tail`, and the source line (`data-source-line`)
- message: the class `message`, the uids of both participants (`data-entity-1`, `data-entity-2`), the message id (`msg1`, `msg2`, ... in source order) and the source line

#### Scenario: Groups of a message

- **WHEN** `Alice -> Bob : hi` is rendered to SVG
- **THEN** the lifeline of `Alice` is wrapped in `<g class="participant-lifeline" data-qualified-name="Alice" data-entity-uid="part1" id="part1-lifeline" data-source-line="1">`, its head and tail in groups of class `participant participant-head` (id `part1-head`) and `participant participant-tail` (id `part1-tail`), and the arrow in `<g class="message" data-entity-1="part1" data-entity-2="part2" id="msg1" data-source-line="1">`
