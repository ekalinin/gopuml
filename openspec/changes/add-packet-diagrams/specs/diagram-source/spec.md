# Spec Delta

## MODIFIED Requirements

### Requirement: Diagram family detection

The system SHALL select the diagram family from the start tag. The supported families SHALL be `uml` and `packetdiag`. Every other start tag that PlantUML v1.2026.8 recognizes (for example `@startmindmap`, `@startgantt`, `@startjson`) SHALL produce an error diagram that names the family and states it is not supported yet. A start tag that upstream does not recognize, including `@startditaa`, which the MIT variant lacks, SHALL produce upstream's "Diagram not supported by this release of PlantUML" diagram.

#### Scenario: Supported family

- **WHEN** a block starts with `@startuml`
- **THEN** the system detects the diagram type within the `uml` family

#### Scenario: Packet family

- **WHEN** a block starts with `@startpacketdiag` and contains `0-7: Source Port`
- **THEN** the system does not report that packet diagrams are not supported yet, and the diagram type is packet

#### Scenario: Recognized but unsupported family

- **WHEN** a block starts with `@startmindmap`
- **THEN** the system produces an error diagram stating that `mindmap` diagrams are not supported yet

#### Scenario: Unknown directive

- **WHEN** a block starts with `@startfoo`
- **THEN** the system produces upstream's unsupported diagram, which names the directive `@startfoo`
