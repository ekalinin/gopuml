# Spec Delta

## MODIFIED Requirements

### Requirement: Encoding

The system SHALL encode diagram source as upstream does:

1. remove the `@startuml`/`@enduml` lines of a `@startuml` source, as upstream's source compressor does. Only the lowercase forms are removed: a source written `@startUML`...`@endUML` keeps these lines
2. encode the text as UTF-8
3. compress with deflate
4. write the result in PlantUML's 64-character alphabet `0-9`, `A-Z`, `a-z`, `-`, `_`

Upstream SHALL decode the encoded text back to the same diagram. Byte-identical encoded strings are not required.

#### Scenario: Round trip through upstream

- **WHEN** the source `@startuml`, `Alice -> Bob : hi`, `@enduml` is encoded by this system and decoded by upstream v1.2026.9
- **THEN** upstream produces the same diagram source

#### Scenario: Start and end lines in another case

- **WHEN** the source `@startUML`, `Alice -> Bob : hi`, `@endUML` is encoded by this system and decoded by upstream v1.2026.9
- **THEN** upstream produces the source `@startUML`, `Alice -> Bob : hi`, `@endUML`

#### Scenario: URL-safe output

- **WHEN** any source is encoded
- **THEN** the result contains only characters from the alphabet `0-9A-Za-z-_`
