# sequence-diagram Specification

## Purpose

Supports PlantUML v1.2026.8 sequence diagrams under `@startuml`, with the same syntax and the layout of upstream's Teoz engine, its only sequence engine.

## Requirements

### Requirement: Participants

The system SHALL support the participant kinds:

- `participant`, `actor`, `boundary`, `control`, `entity`, `database`, `collections` and `queue`
- aliases with `as` and quoted display names
- per-participant colors, stereotypes and `order` values
- `box`...`end box` grouping
- implicit declaration on first use in a message

#### Scenario: Declared participant kinds

- **WHEN** a diagram declares `actor User` and `database DB` and contains `User -> DB : query`
- **THEN** `User` is drawn as an actor, `DB` as a database, and an arrow labeled `query` runs between them

#### Scenario: Implicit participant

- **WHEN** a diagram contains only `Alice -> Bob : hi`
- **THEN** participants `Alice` and `Bob` are drawn in that order

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

The system SHALL support `alt`/`else`, `opt`, `loop`, `par`, `break`, `critical` and `group` fragments closed by `end`, including nesting and fragment labels.

#### Scenario: Alternative fragment

- **WHEN** a diagram contains `alt success`, a message, `else failure`, a message and `end`
- **THEN** a fragment labeled `alt` is drawn with two sections labeled `success` and `failure`

### Requirement: Notes

The system SHALL support `note left`, `note right` and `note over` (one or more participants), `note across`, the `hnote` and `rnote` shapes, multi-line notes, and notes attached to messages.

#### Scenario: Note over two participants

- **WHEN** a diagram contains `note over Alice, Bob : shared`
- **THEN** a note labeled `shared` spans the lifelines of `Alice` and `Bob`

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

The system SHALL lay out sequence diagrams the same way as upstream's Teoz engine, the only sequence engine in v1.2026.8. `!pragma teoz` SHALL be accepted with any value and SHALL NOT change the layout, as upstream does.

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
