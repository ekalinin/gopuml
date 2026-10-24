# Spec Delta

## Purpose

Bundles the standard library of PlantUML v1.2026.9 (the `stdlib/` folder of the plantuml-mit jar), subject to a license check, and provides its files, JSON resources, themes and metadata as upstream does.

## ADDED Requirements

### Requirement: Bundled libraries

The system SHALL bundle the libraries of the standard library of PlantUML v1.2026.9, with the content that the plantuml-mit 1.2026.9 jar ships, except the libraries whose license does not allow redistribution or cannot be verified. `adaml` (GPL-3.0) SHALL NOT be bundled. NOTICE SHALL list every bundled library with its source and license, and every excluded library with the reason. Library names SHALL be matched in lower case where upstream lower-cases them (include paths, sprite and image references) and as given where upstream does not (the arguments of `%get_stdlib`).

#### Scenario: Library name in another case

- **WHEN** a block contains `!include <c4/C4_Container>` and `Container(api, "API", "Go")`
- **THEN** the diagram renders the container `API` with the technology `[Go]`, as with `!include <C4/C4_Container>`

#### Scenario: Excluded library not listed

- **WHEN** a block evaluates `%get_all_stdlib()`
- **THEN** the result is a JSON array of the bundled library names in name order, which contains `c4` and `awslib14` and does not contain `adaml`

### Requirement: Bundled files match the pinned upstream jar

Every bundled file SHALL be byte-identical to the file of the same path under `stdlib/` in the plantuml-mit 1.2026.9 jar, except the list of library names, which SHALL hold the names of the upstream list without the excluded libraries, in the upstream order. The project SHALL provide a check, run in continuous integration, that fails when a bundled file differs from the pinned jar, when a library of the pinned jar is neither bundled nor excluded, or when an excluded library is bundled.

#### Scenario: Modified library file

- **WHEN** a byte of a bundled channel file of `c4` is changed and the check runs against the unpacked 1.2026.9 sources jar
- **THEN** the check fails and names the file

### Requirement: Library aliases

A library whose metadata declares a `link` SHALL resolve, for includes, sprites, images, resources and metadata, to the linked library, as upstream does: `awslib` to `awslib14`, `bootstrap` to `bootstrap1.13.1`, `material` and `material2` to `material2.1.19`, and `material7` to `material7.4.47`.

#### Scenario: Include through an alias

- **WHEN** a block contains `!include <awslib/AWSCommon>`, `!include <awslib/Compute/EC2>` and `EC2(web, "Web", "x")`
- **THEN** the diagram renders the EC2 icon of `awslib14` with the label `Web`, as upstream does

#### Scenario: Version of an alias

- **WHEN** a block evaluates `%get_stdlib("awslib", "version")`
- **THEN** the result is `14.0.0`

### Requirement: Library metadata

The system SHALL report the metadata of the bundled libraries as upstream does, without the excluded libraries:

- `-stdlib` prints, for each bundled library in name order, its name, `Version <version>`, `Delivered by <source>` and a line holding one space
- `%get_all_stdlib()` returns the JSON array of the library names, and `%get_all_stdlib(x)` an object with the name, version and source of each library
- `%get_stdlib()` returns the metadata of every library, `%get_stdlib(name)` the metadata of one library, and `%get_stdlib(name, key)` one value, or the empty string when the key is missing

A name that `%get_stdlib` does not find (an excluded library, a library that upstream does not ship, or a name in another case than the library folder) SHALL produce upstream's error for a missing library.

#### Scenario: Library list

- **WHEN** the CLI runs `plantuml -stdlib`
- **THEN** the output starts with the lines `archimate`, `Version 3.2.2`, `Delivered by https://github.com/plantuml-stdlib/Archimate-PlantUML` and ` `, lists `c4` with `Version 2.14.0`, has no `adaml` entry, and the CLI exits with status 0

#### Scenario: Version of a library

- **WHEN** a block evaluates `%get_stdlib("c4", "version")`
- **THEN** the result is `2.14.0`

### Requirement: Library resources

The system SHALL read the JSON resources of a bundled library with `%load_json("<library/file.json>")`, and the themes of a bundled library with `!theme <name> from <library/directory>`, as upstream does.

#### Scenario: JSON resource of a library

- **WHEN** a block contains `!$o = %load_json("<office/all.json>")` and a message labeled `%size($o)`
- **THEN** the message label is `10`

#### Scenario: Theme of a library

- **WHEN** a block contains `!include <C4/C4_Container>`, `!theme C4_blue from <C4/themes>` and `Container(api, "API", "Go")`
- **THEN** the diagram renders with the colors of the `C4_blue` theme of the C4 library, as upstream does
