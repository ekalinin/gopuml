# Spec Delta

## Purpose

Encodes diagram source into PlantUML's URL text format and decodes it back, so that URLs and encoded strings work interchangeably with PlantUML v1.2026.8 and plantuml-server.

## ADDED Requirements

### Requirement: Encoding

The system SHALL encode diagram source as upstream does:

1. remove the `@startuml`/`@enduml` lines of a `@startuml` source, as upstream's source compressor does
2. encode the text as UTF-8
3. compress with deflate
4. write the result in PlantUML's 64-character alphabet `0-9`, `A-Z`, `a-z`, `-`, `_`

Upstream SHALL decode the encoded text back to the same diagram. Byte-identical encoded strings are not required.

#### Scenario: Round trip through upstream

- **WHEN** the source `@startuml`, `Alice -> Bob : hi`, `@enduml` is encoded by this system and decoded by upstream v1.2026.8
- **THEN** upstream produces the same diagram source

#### Scenario: URL-safe output

- **WHEN** any source is encoded
- **THEN** the result contains only characters from the alphabet `0-9A-Za-z-_`

### Requirement: Decoding

The system SHALL decode every form upstream decodes:

- unprefixed text: deflate data, falling back to the legacy Huffman form
- `~0` (deflate), `~1` (legacy Huffman), `~h` (hexadecimal, uncompressed), `~g` (gzip) and `~zip~` (zip)

When the decoded text does not start with `@start`, the system SHALL add `@startuml`/`@enduml` lines, as upstream does.

#### Scenario: Decode an upstream-encoded string

- **WHEN** a string produced by upstream's encoder is decoded
- **THEN** the result is the original diagram source

#### Scenario: Hex form

- **WHEN** `~h` followed by the hex bytes of `Alice -> Bob` is decoded
- **THEN** the result is a `@startuml` block containing `Alice -> Bob`

### Requirement: Invalid encoded text

When the text cannot be decoded in any supported form, the system SHALL return a decoding error and SHALL NOT return partial source.

#### Scenario: Corrupted text

- **WHEN** the string `@@not-valid@@` is decoded
- **THEN** a decoding error is returned
