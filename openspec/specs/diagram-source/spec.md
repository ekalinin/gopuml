# diagram-source Specification

## Purpose

Defines how diagram sources are found in input text, split into diagram blocks, assigned a diagram type and reported when invalid, matching PlantUML v1.2026.9.

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

The system SHALL select the diagram family from the start tag. The supported families SHALL be `uml`, `packetdiag`, `gantt` (`@startgantt` and `@startproject`), `mindmap` and `wbs`. Every other start tag that PlantUML v1.2026.9 recognizes (for example `@startjson`, `@startsalt`) SHALL produce an error diagram that names the family and states it is not supported yet. A start tag that upstream does not recognize, including `@startditaa`, which the MIT variant lacks, SHALL produce upstream's "Diagram not supported by this release of PlantUML" diagram.

#### Scenario: Supported family

- **WHEN** a block starts with `@startuml`
- **THEN** the system detects the diagram type within the `uml` family

#### Scenario: Packet family

- **WHEN** a block starts with `@startpacketdiag` and contains `0-7: Source Port`
- **THEN** the system does not report that packet diagrams are not supported yet, and the diagram type is packet

#### Scenario: Gantt family

- **WHEN** a block starts with `@startgantt` and contains `[Build] lasts 3 days`
- **THEN** the system does not report that gantt diagrams are not supported yet, and the diagram type is gantt

#### Scenario: Mindmap family

- **WHEN** a block starts with `@startmindmap` and contains `* root`
- **THEN** the system does not report that mindmap diagrams are not supported yet, and the diagram type is mindmap

#### Scenario: WBS family

- **WHEN** a block starts with `@startwbs` and contains `* root`
- **THEN** the system does not report that wbs diagrams are not supported yet, and the diagram type is wbs

#### Scenario: Recognized but unsupported family

- **WHEN** a block starts with `@startjson`
- **THEN** the system produces an error diagram stating that `json` diagrams are not supported yet

#### Scenario: Unknown directive

- **WHEN** a block starts with `@startfoo`
- **THEN** the system produces upstream's unsupported diagram, which names the directive `@startfoo`

### Requirement: Empty diagram

An `@startuml` block without content SHALL render upstream's welcome diagram.

#### Scenario: Welcome screen

- **WHEN** the input contains only `@startuml` and `@enduml`
- **THEN** the system renders upstream's welcome diagram, which starts with the text `Welcome to PlantUML!`

### Requirement: Diagram type detection under @startuml

For `@startuml` blocks, the system SHALL try the supported diagram types in the same relative order as upstream (sequence, then class, then legacy activity, then description, then state, then activity in the new syntax) and SHALL use the first type that parses the whole block without error. The supported types SHALL be sequence, class, legacy activity, description (use case, component and deployment), state and activity (new syntax). Legacy activity diagrams and activity diagrams of the new syntax both have the diagram type activity.

#### Scenario: Sequence diagram detected

- **WHEN** a `@startuml` block contains only `Alice -> Bob : hello`
- **THEN** the diagram type is sequence

#### Scenario: Class diagram detected

- **WHEN** a `@startuml` block contains `class Order` and `Order --> Customer`
- **THEN** the diagram type is class

#### Scenario: Description diagram detected

- **WHEN** a `@startuml` block contains `component Web` and `node Server`
- **THEN** the diagram type is description

#### Scenario: State diagram detected

- **WHEN** a `@startuml` block contains `[*] --> Idle` and `Idle --> [*]`
- **THEN** the diagram type is state

#### Scenario: Legacy activity before state

- **WHEN** a `@startuml` block contains `A --> ==sync==` and `==sync== --> C`, which both the legacy activity and the state syntax accept
- **THEN** the diagram type is activity, and `sync` is drawn as a synchronization bar of a legacy activity diagram

#### Scenario: Activity diagram detected

- **WHEN** a `@startuml` block contains `start`, `:Read input;` and `stop`
- **THEN** the diagram type is activity

### Requirement: Explicit errors for unsupported @startuml types

When no supported type parses a `@startuml` block, the system SHALL report a syntax error. When the block uses syntax that belongs to a `@startuml` diagram type that is not supported yet (archimate, timing), the error message SHALL name that diagram type and state that it is not supported yet. Syntax of a supported type, such as state or legacy activity, SHALL NOT be reported as not supported yet.

#### Scenario: State diagram syntax

- **WHEN** a `@startuml` block contains `[*] --> Idle`
- **THEN** the system does not report that state diagrams are not supported yet, and the diagram type is state

#### Scenario: Legacy activity syntax

- **WHEN** a `@startuml` block contains `(*) --> "First step"`
- **THEN** the system does not report that legacy activity diagrams are not supported yet, and the diagram type is activity

#### Scenario: Archimate syntax

- **WHEN** a `@startuml` block contains `archimate #Business "Customer" as C`
- **THEN** the system reports an error stating that archimate diagrams are not supported yet

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
