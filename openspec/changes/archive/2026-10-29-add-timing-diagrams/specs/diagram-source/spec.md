# Spec Delta

## MODIFIED Requirements

### Requirement: Diagram type detection under @startuml

For `@startuml` blocks, the system SHALL try the supported diagram types in the same relative order as upstream (sequence, then class, then legacy activity, then description, then state, then activity in the new syntax, then timing) and SHALL use the first type that parses the whole block without error. The supported types SHALL be sequence, class, legacy activity, description (use case, component and deployment), state, activity (new syntax) and timing. Legacy activity diagrams and activity diagrams of the new syntax both have the diagram type activity.

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

#### Scenario: Timing diagram detected

- **WHEN** a `@startuml` block contains `robust "Web" as WB`, `@0` and `WB is Idle`
- **THEN** the system does not report that timing diagrams are not supported yet, and the diagram type is timing

### Requirement: Explicit errors for unsupported @startuml types

When no supported type parses a `@startuml` block, the system SHALL report a syntax error. When the block uses syntax that belongs to a `@startuml` diagram type that is not supported yet (archimate), the error message SHALL name that diagram type and state that it is not supported yet. Syntax of a supported type, such as state, legacy activity or timing, SHALL NOT be reported as not supported yet.

#### Scenario: State diagram syntax

- **WHEN** a `@startuml` block contains `[*] --> Idle`
- **THEN** the system does not report that state diagrams are not supported yet, and the diagram type is state

#### Scenario: Legacy activity syntax

- **WHEN** a `@startuml` block contains `(*) --> "First step"`
- **THEN** the system does not report that legacy activity diagrams are not supported yet, and the diagram type is activity

#### Scenario: Archimate syntax

- **WHEN** a `@startuml` block contains `archimate #Business "Customer" as C`
- **THEN** the system reports an error stating that archimate diagrams are not supported yet

#### Scenario: Timing syntax error

- **WHEN** a `@startuml` block contains `robust "Web" as WB`, `@0`, `WB is Idle` and `foo bar baz`
- **THEN** the diagram fails at the line `foo bar baz` with the message `Syntax Error? (Assumed diagram type: timing)`, and the system does not report that timing diagrams are not supported yet
