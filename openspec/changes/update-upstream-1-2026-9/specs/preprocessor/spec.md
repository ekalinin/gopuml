# Spec Delta

## MODIFIED Requirements

### Requirement: Builtin functions

The system SHALL provide every builtin function that PlantUML v1.2026.9 provides (for example `%strlen`, `%upper`, `%substr`, `%dec2hex`, `%date`, `%splitstr`, `%json_set`, `%darken`, `%version`), with the same argument handling and results. `%version` SHALL return the upstream version that this implementation targets.

#### Scenario: String builtins

- **WHEN** a block evaluates `%strlen("abc")` and `%upper("abc")`
- **THEN** the results are `3` and `ABC`

#### Scenario: Numeric conversion builtin

- **WHEN** a block evaluates `%dec2hex(255)`
- **THEN** the result is `ff`

### Requirement: File inclusion

The system SHALL support `!include` (resolved relative to the including file, then the include path), `!include_many`, `!include_once`, `!includesub` with `!startsub`/`!endsub`, `!includedef`, and `!import` of zip archives, after which a path is also looked up in the imported archives, from their root. As upstream v1.2026.9 does for local files, `!include` SHALL skip a file that the block has already included with the same selector, `!include_once` of such a file SHALL report the error `This file has already been included`, and `!include_many` SHALL include the file every time. An included file that contains `@startuml`...`@enduml` SHALL contribute only the lines of its first block, unless a block is selected with `!include file.puml!N` (the index of the block, from 0) or `!include file.puml!ID` (the block whose start line declares `id=ID`). A selector that selects no block SHALL produce the error `cannot include <file>!<selector>`.

#### Scenario: Include relative to the including file

- **WHEN** `diagrams/main.puml` contains `!include common/style.iuml` and `diagrams/common/style.iuml` exists
- **THEN** the lines of `diagrams/common/style.iuml` are preprocessed in place of the directive

#### Scenario: Include once

- **WHEN** a block contains `!include_once a.iuml` twice
- **THEN** the lines of `a.iuml` are included once, and the second directive reports the error `This file has already been included`

#### Scenario: Include twice

- **WHEN** a block contains `!include a.iuml` twice
- **THEN** the lines of `a.iuml` are included once and no error is reported

#### Scenario: Include many

- **WHEN** a block contains `!include_many a.iuml` twice
- **THEN** the lines of `a.iuml` are included twice

#### Scenario: Selected block

- **WHEN** `multi.puml` holds two `@startuml`...`@enduml` blocks and a block contains `!include multi.puml!1`
- **THEN** only the lines of the second block of `multi.puml` are included

#### Scenario: Unusable selector

- **WHEN** `multi.puml` holds two `@startuml`...`@enduml` blocks and a block contains `!include multi.puml!5`
- **THEN** the system reports an error at that line with the message `cannot include multi.puml!5`

#### Scenario: Include from an imported archive

- **WHEN** `archive.zip` holds `lib/part.iuml`, and a block contains `!import archive.zip` and then `!include lib/part.iuml`
- **THEN** the lines of `lib/part.iuml` from the archive are preprocessed in place of the `!include` directive

#### Scenario: Missing include file

- **WHEN** a block contains `!include missing.iuml` and the file does not exist
- **THEN** the system reports an error at that line with the message `cannot include missing.iuml`

### Requirement: Pragmas and options

The system SHALL accept `!pragma` and `!option` directives and pass their values to diagram processing, with the same names and effects as upstream for the features this change implements. Pragmas for features outside this change SHALL be accepted and ignored. An `!option` with an unknown name or without a value SHALL be ignored with upstream's warning.

#### Scenario: Layout pragma for an unsupported engine

- **WHEN** a class diagram contains `!pragma layout elk`
- **THEN** the diagram renders with the default layout engine and no error is reported

#### Scenario: Unknown option

- **WHEN** a sequence diagram contains `!option foo bar` and `A -> B`
- **THEN** the diagram renders, and the warning lines `Unknown '!option foo': this line is ignored` and `Known options are: language, use_descriptive_names, handwritten, debug, svg_desc, svg_title` are shown above it
