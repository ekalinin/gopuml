# Spec Delta

## MODIFIED Requirements

### Requirement: Diagram family detection

The system SHALL select the diagram family from the start tag. The supported families SHALL be `uml`, `packetdiag`, `gantt` (`@startgantt` and `@startproject`), `mindmap`, `wbs` and `salt`. Every other start tag that PlantUML v1.2026.9 recognizes (for example `@startjson`, `@startnwdiag`) SHALL produce an error diagram that names the family and states it is not supported yet. A start tag that upstream does not recognize, including `@startditaa`, which the MIT variant lacks, SHALL produce upstream's "Diagram not supported by this release of PlantUML" diagram.

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

#### Scenario: Salt family

- **WHEN** a block starts with `@startsalt` and contains `{`, `[OK]` and `}`
- **THEN** the system does not report that salt diagrams are not supported yet, and the diagram type is salt

#### Scenario: Recognized but unsupported family

- **WHEN** a block starts with `@startjson`
- **THEN** the system produces an error diagram stating that `json` diagrams are not supported yet

#### Scenario: Unknown directive

- **WHEN** a block starts with `@startfoo`
- **THEN** the system produces upstream's unsupported diagram, which names the directive `@startfoo`
