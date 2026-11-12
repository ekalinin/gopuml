# yaml-diagram Specification

## Purpose
Supports PlantUML v1.2026.9 YAML diagrams under `@startyaml`: the subset of YAML that upstream's parser reads, drawn as JSON diagrams with the `yamlDiagram` styles, and the image for YAML that it cannot read.

## Requirements

### Requirement: YAML diagram blocks

The system SHALL read `@startyaml` ... `@endyaml` blocks as upstream does. The `<style>` blocks, the lines taken out before the data and the `#highlight` lines SHALL be handled as in JSON diagrams, and the other lines, after preprocessing, SHALL be the YAML text. The system SHALL read the YAML that upstream's parser reads: `key: value` mappings, nested by indentation; sequences of `- value` items and of mappings, including a mapping that starts on the line after a lone `-`; flow sequences (`key: [a, b]`); `key: |` block values, whose lines are kept with their line breaks; values continued on the next, more indented lines, joined with spaces; quoted values; keys without a value, which have an empty value; and `#` comments, alone on a line or after a value. The data SHALL be drawn as a JSON diagram, with the `yamlDiagram` styles, as upstream does. The diagram SHALL have one page and the description `(Yaml)`, and the syntax check SHALL report it as `YAML`.

#### Scenario: Mapping with nested values

- **WHEN** a `@startyaml` block contains `fruit: Apple`, `size: Large`, `color:`, `  - Red`, `  - Green`, `info:`, `  origin: Spain` and `  weight: 120`
- **THEN** the first box has the rows `fruit` `Apple`, `size` `Large`, `color` and `info`, the row `color` is linked to a box with the rows `Red` and `Green`, and the row `info` to a box with the rows `origin` `Spain` and `weight` `120`

#### Scenario: Sequence of mappings

- **WHEN** a `@startyaml` block contains `- name: Mark`, `  hr: 65`, `- name: Sammy` and `  hr: 63`
- **THEN** the first box has two rows, linked to a box with the rows `name` `Mark` and `hr` `65` and to a box with the rows `name` `Sammy` and `hr` `63`

#### Scenario: Mappings after lone dashes

- **WHEN** a `@startyaml` block contains `-`, `  name: Mark`, `-` and `  name: Sammy`
- **THEN** the first box has two rows, linked to a box with the row `name` `Mark` and to a box with the row `name` `Sammy`

#### Scenario: Flow sequence, empty and quoted values

- **WHEN** a `@startyaml` block contains `tags: [a, b, c]`, `empty:` and `quoted: "hello: world"`
- **THEN** the row `tags` is linked to a box with the rows `a`, `b` and `c`, the row `empty` has an empty value, and the row `quoted` has the value `hello: world`

#### Scenario: Block value

- **WHEN** a `@startyaml` block contains `text: |`, `  line one`, `  line two` and `next: value`
- **THEN** the row `text` holds `line one` and `line two` in one value on one row, and the row `next` has the value `value`, as upstream does

#### Scenario: Value continued on the next lines

- **WHEN** a `@startyaml` block contains `desc:`, `  first part`, `  second part` and `other: 1`
- **THEN** the row `desc` has the value `first part second part` and the row `other` the value `1`

#### Scenario: YAML comments

- **WHEN** a `@startyaml` block contains `# a comment`, `a: 1 # trailing` and `b: 2`
- **THEN** the box has the rows `a` `1` and `b` `2`

#### Scenario: Highlighted YAML key

- **WHEN** a `@startyaml` block contains `#highlight "info" / "origin"`, `info:`, `  origin: Spain` and `  weight: 120`
- **THEN** the row `origin` has the background `#CCFF02` and the row `weight` has none

#### Scenario: YAML styles and title

- **WHEN** a `@startyaml` block contains `<style>`, `yamlDiagram {`, `node { BackGroundColor #ffeecc }`, `}`, `</style>`, `title Fruits`, `a:` and `  b: 1`
- **THEN** the title `Fruits` is drawn above the boxes, and the boxes have the background `#FFEECC`

#### Scenario: Syntax check of a YAML diagram

- **WHEN** the syntax check reads a `@startyaml` block that contains `a: 1`
- **THEN** it reports the type `YAML` and the description `(Yaml)`

### Requirement: YAML that cannot be read

When upstream's parser rejects the YAML text, the system SHALL draw the image of the text `Your data does not sound like YAML data`, without an error, as upstream does: for a line of text without a key outside a value, and for a sequence item at the level of a mapping, among others.

#### Scenario: Text without a key

- **WHEN** a `@startyaml` block contains `just some text`
- **THEN** the diagram is the text `Your data does not sound like YAML data`, without an error

#### Scenario: Sequence item after a mapping

- **WHEN** a `@startyaml` block contains `tags: [a, b, c]` and `- x`
- **THEN** the diagram is the text `Your data does not sound like YAML data`, without an error, as upstream does
