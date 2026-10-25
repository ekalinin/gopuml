# Spec Delta

## MODIFIED Requirements

### Requirement: Packet output formats

Packet diagrams SHALL render to SVG and PNG. TXT and UTXT output SHALL be rejected with the error that names the format and the diagram type, without output, as for the other diagram types that have no text output (every type except sequence and gantt).

#### Scenario: PNG packet diagram

- **WHEN** a packet diagram with `0-7: Source Port` and `8-15: Destination Port` is rendered as PNG
- **THEN** the PNG image shows the scale and the two fields, as upstream's PNG does

#### Scenario: TXT for a packet diagram

- **WHEN** a packet diagram is rendered as TXT
- **THEN** the system reports that TXT output is not supported for packet diagrams and produces no output
