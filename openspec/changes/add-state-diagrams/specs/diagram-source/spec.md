# Spec Delta

## MODIFIED Requirements

### Requirement: Diagram type detection under @startuml

For `@startuml` blocks, the system SHALL try the supported diagram types in the same relative order as upstream (sequence, then class, then description, then state, then activity) and SHALL use the first type that parses the whole block without error. The supported types SHALL be sequence, class, description (use case, component and deployment), state and activity (new syntax).

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

#### Scenario: Activity diagram detected

- **WHEN** a `@startuml` block contains `start`, `:Read input;` and `stop`
- **THEN** the diagram type is activity

### Requirement: Explicit errors for unsupported @startuml types

When no supported type parses a `@startuml` block, the system SHALL report a syntax error. When the block uses syntax that belongs to a `@startuml` diagram type that is not supported yet (archimate, timing, legacy activity), the error message SHALL name that diagram type and state that it is not supported yet. Syntax of a supported type, such as state, SHALL NOT be reported as not supported yet.

#### Scenario: State diagram syntax

- **WHEN** a `@startuml` block contains `[*] --> Idle`
- **THEN** the system does not report that state diagrams are not supported yet, and the diagram type is state

#### Scenario: Legacy activity syntax

- **WHEN** a `@startuml` block contains `(*) --> "First step"`
- **THEN** the system reports an error stating that legacy activity diagrams are not supported yet

#### Scenario: Archimate syntax

- **WHEN** a `@startuml` block contains `archimate #Business "Customer" as C`
- **THEN** the system reports an error stating that archimate diagrams are not supported yet
