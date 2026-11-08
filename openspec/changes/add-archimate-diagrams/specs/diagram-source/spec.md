# Spec Delta

## MODIFIED Requirements

### Requirement: Explicit errors for unsupported @startuml types

When no supported type parses a `@startuml` block, the system SHALL report a syntax error. When the block uses syntax that belongs to a `@startuml` diagram type that is not supported yet, the error message SHALL name that diagram type and state that it is not supported yet. In this release, no `@startuml` diagram type is reported as not supported yet. Syntax of a supported type, such as state, legacy activity, timing or archimate, SHALL NOT be reported as not supported yet.

#### Scenario: State diagram syntax

- **WHEN** a `@startuml` block contains `[*] --> Idle`
- **THEN** the system does not report that state diagrams are not supported yet, and the diagram type is state

#### Scenario: Legacy activity syntax

- **WHEN** a `@startuml` block contains `(*) --> "First step"`
- **THEN** the system does not report that legacy activity diagrams are not supported yet, and the diagram type is activity

#### Scenario: Archimate syntax

- **WHEN** a `@startuml` block contains `archimate #Business "Customer" as C`
- **THEN** the system does not report that archimate diagrams are not supported yet, and the diagram type is description

#### Scenario: Timing syntax error

- **WHEN** a `@startuml` block contains `robust "Web" as WB`, `@0`, `WB is Idle` and `foo bar baz`
- **THEN** the diagram fails at the line `foo bar baz` with the message `Syntax Error? (Assumed diagram type: timing)`, and the system does not report that timing diagrams are not supported yet
