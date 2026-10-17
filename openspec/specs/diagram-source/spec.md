# diagram-source Specification

## Purpose

Defines how diagram sources are found in input text, split into diagram blocks, assigned a diagram type and reported when invalid, matching PlantUML v1.2026.8.

## Requirements

### Requirement: Diagram block extraction

The system SHALL extract every diagram block that starts with a line beginning with `@start<family>` or `\start<family>` and ends with an `@end...` or `\end...` line. As upstream requires, the `start` and `end` keywords SHALL be lower case, while the family part (for example `uml` or `MindMap`) SHALL match case-insensitively. Text outside blocks SHALL be ignored, and blocks SHALL be processed in the order they appear in the input.

#### Scenario: Two blocks in one input

- **WHEN** the input contains a `@startuml`...`@enduml` block followed by a second `@startuml`...`@enduml` block
- **THEN** the system produces two diagrams in input order

#### Scenario: Diagram embedded in other text

- **WHEN** a Java source file contains a `@startuml`...`@enduml` block inside a comment
- **THEN** the system extracts that block and ignores the surrounding Java code

#### Scenario: No block in input

- **WHEN** the input contains no line starting with `@start` or `\start`
- **THEN** the system reports that no diagram was found and produces no diagram

#### Scenario: Upper-case keyword

- **WHEN** the input contains `@STARTUML`, `Alice -> Bob` and `@ENDUML`
- **THEN** no diagram block is found

#### Scenario: Mixed-case family

- **WHEN** the input contains `@startUML`, `Alice -> Bob` and `@endUML`
- **THEN** one sequence diagram block is found

### Requirement: Diagram name from the start line

The system SHALL expose the text that follows the start tag on the start line as the diagram name, normalized as upstream does: surrounding quotes and spaces are removed, the text is cut at the first comma, and a three-letter extension and a `file://` prefix are removed. Text containing `<`, `>` or `|` SHALL produce no name, and so SHALL a start line with nothing after the tag.

#### Scenario: Named diagram

- **WHEN** a block starts with `@startuml login-flow`
- **THEN** the diagram name is `login-flow`

#### Scenario: Unnamed diagram

- **WHEN** a block starts with `@startuml` and nothing else
- **THEN** the diagram has no name

#### Scenario: Name normalization

- **WHEN** a block starts with `@startuml "login-flow.png"`
- **THEN** the diagram name is `login-flow`

### Requirement: Diagram family detection

The system SHALL select the diagram family from the start tag. In this change the only supported family SHALL be `uml`. Every other start tag that PlantUML v1.2026.8 recognizes (for example `@startmindmap`, `@startgantt`, `@startjson`) SHALL produce an error diagram that names the family and states it is not supported yet. A start tag that upstream does not recognize, including `@startditaa`, which the MIT variant lacks, SHALL produce upstream's "Diagram not supported by this release of PlantUML" diagram.

#### Scenario: Supported family

- **WHEN** a block starts with `@startuml`
- **THEN** the system detects the diagram type within the `uml` family

#### Scenario: Recognized but unsupported family

- **WHEN** a block starts with `@startmindmap`
- **THEN** the system produces an error diagram stating that `mindmap` diagrams are not supported yet

#### Scenario: Unknown directive

- **WHEN** a block starts with `@startfoo`
- **THEN** the system produces upstream's unsupported diagram, which names the directive `@startfoo`

### Requirement: Empty diagram

An `@startuml` block without content SHALL render upstream's welcome diagram.

#### Scenario: Welcome screen

- **WHEN** the input contains only `@startuml` and `@enduml`
- **THEN** the system renders upstream's welcome diagram, which starts with the text `Welcome to PlantUML!`

### Requirement: Diagram type detection under @startuml

For `@startuml` blocks, the system SHALL try the supported diagram types in the same relative order as upstream (sequence, then class, then activity) and SHALL use the first type that parses the whole block without error. In this change the supported types SHALL be sequence, class and activity (new syntax).

#### Scenario: Sequence diagram detected

- **WHEN** a `@startuml` block contains only `Alice -> Bob : hello`
- **THEN** the diagram type is sequence

#### Scenario: Class diagram detected

- **WHEN** a `@startuml` block contains `class Order` and `Order --> Customer`
- **THEN** the diagram type is class

#### Scenario: Activity diagram detected

- **WHEN** a `@startuml` block contains `start`, `:Read input;` and `stop`
- **THEN** the diagram type is activity

### Requirement: Explicit errors for unsupported @startuml types

When no supported type parses a `@startuml` block, the system SHALL report a syntax error. When the block uses syntax that belongs to a `@startuml` diagram type that this change does not support (use case, component, deployment, archimate, state, timing, legacy activity), the error message SHALL name that diagram type and state that it is not supported yet.

#### Scenario: State diagram syntax

- **WHEN** a `@startuml` block contains `[*] --> Idle`
- **THEN** the system reports an error stating that state diagrams are not supported yet

#### Scenario: Legacy activity syntax

- **WHEN** a `@startuml` block contains `(*) --> "First step"`
- **THEN** the system reports an error stating that legacy activity diagrams are not supported yet

### Requirement: Syntax error reporting

When a block has a syntax error, the system SHALL report the line number that upstream reports for the same input, counting the lines of the whole input, including text before the block and preprocessor lines, together with the text of the offending line and an error message. When no supported type matched and no unsupported type was identified, the message SHALL state the assumed diagram type in the upstream form `Syntax Error? (Assumed diagram type: sequence)`.

#### Scenario: Invalid sequence line

- **WHEN** a `@startuml` block contains `Alice -> Bob : hi` followed by the line `Alice -> -> Bob`
- **THEN** the system reports a syntax error with the same line number upstream reports, and the error includes the text `Alice -> -> Bob`

### Requirement: Error images

When a diagram has an error, rendering SHALL produce an error image in the requested output format. The image SHALL show the source lines up to the error, mark the offending line, and include the error message, unless the caller disables error images. For sources shorter than five lines, the image SHALL also show upstream's welcome text. The image SHALL NOT contain the sponsor messages upstream adds during some minutes of each hour.

#### Scenario: Error image in SVG

- **WHEN** an invalid diagram is rendered to SVG
- **THEN** the output is a valid SVG document that contains the error message and the offending line text

#### Scenario: Error images disabled

- **WHEN** an invalid diagram is rendered with error images disabled
- **THEN** no image is produced and the error is still reported to the caller
