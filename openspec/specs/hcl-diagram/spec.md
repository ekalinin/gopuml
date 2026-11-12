# hcl-diagram Specification

## Purpose
Supports PlantUML v1.2026.9 HCL diagrams under `@starthcl`: the subset of HCL that upstream's parser reads, drawn as JSON diagrams with the `yamlDiagram` styles, and the image for HCL that it cannot read.

## Requirements

### Requirement: HCL diagram blocks

The system SHALL read `@starthcl` ... `@endhcl` blocks as upstream does. The `<style>` blocks and the lines taken out before the data SHALL be handled as in JSON diagrams, except that the `<style>` blocks, `title` and `skin` SHALL NOT be applied, as upstream does, and the other lines, after preprocessing, SHALL be the HCL text. A line whose first character other than a space is `#` SHALL be dropped, and the other lines SHALL be joined without line breaks, so that only a space, a bracket, a comma, `=` or `:` ends an unquoted value, as upstream does. The system SHALL read the HCL that upstream's parser reads: blocks, each a header of names and quoted labels followed by `{` and `}`, whose body holds attributes written `name = value` or `name : value`; values that are quoted strings, unquoted words and numbers, lists (`[...]`), objects (`{...}`) and function calls (`name(arguments)`). Every value SHALL be drawn as its text, `true` and `false` included. A function call without arguments SHALL be the text `name()`, and a call with arguments an object with the key `name()` whose value is the list of the arguments. A source of one block SHALL be drawn as the body of that block, and a source of several blocks as an object whose keys are the block headers, labels in quotes, as upstream does. The data SHALL be drawn as a JSON diagram, with the `yamlDiagram` styles and without highlighted rows, as upstream does. The diagram SHALL have one page and the description `(HCL)`, and the syntax check SHALL report it as `HCL`.

#### Scenario: Single block

- **WHEN** a `@starthcl` block contains `resource "aws_instance" "web" {`, `  ami           = "ami-a1b2c3d4"`, `  instance_type = "t2.micro"`, `  count         = 2`, `  enabled       = true` and `}`
- **THEN** the diagram is one box with the rows `ami` `ami-a1b2c3d4`, `instance_type` `t2.micro`, `count` `2` and `enabled` `true`, without a row for the block header, as upstream does

#### Scenario: Two blocks

- **WHEN** a `@starthcl` block contains `variable "region" {`, `  default = "us-east-1"`, `}`, `provider aws {`, `  region = "x"` and `}`
- **THEN** the first box has the rows `variable "region"` and `provider aws`, linked to a box with the row `default` `us-east-1` and to a box with the row `region` `x`

#### Scenario: Lists, objects and function calls

- **WHEN** a `@starthcl` block contains `module "m" {`, `  tags = ["a", "b"]`, `  settings = {`, `    level = "high"`, `  }`, `  name = lower("ABC", "d")`, `  empty = list()` and `}`
- **THEN** the box has the rows `tags`, `settings`, `name` and `empty` `list()`; `tags` is linked to a box with the rows `a` and `b`, `settings` to a box with the row `level` `high`, and `name` to a box with the row `lower()`, which is linked to a box with the rows `ABC` and `d`

#### Scenario: Attribute with a colon

- **WHEN** a `@starthcl` block contains `a {`, `  b : "c"`, `  d = 1.5` and `}`
- **THEN** the box has the rows `b` `c` and `d` `1.5`

#### Scenario: Comment line

- **WHEN** a `@starthcl` block contains `a {`, `  # a comment line`, `  b = "c"` and `}`
- **THEN** the box has the row `b` `c`

#### Scenario: Style and title ignored

- **WHEN** a `@starthcl` block contains `<style>`, `yamlDiagram { node { BackGroundColor #ffeecc } }`, `</style>`, `title Ignored`, `a {`, `  b = "c"` and `}`
- **THEN** the box has the default background `#F1F1F1`, and no title is drawn, as upstream does

#### Scenario: Syntax check of an HCL diagram

- **WHEN** the syntax check reads a `@starthcl` block that contains `a {`, `  b = "c"` and `}`
- **THEN** it reports the type `HCL` and the description `(HCL)`

### Requirement: HCL that cannot be read

When upstream's parser rejects the HCL text, the system SHALL draw the image of the text `Your data does not sound like HCL data`, without an error, as upstream does: for an attribute outside a block, a nested block written without `=`, a `#` comment after a value, and attributes on lines without indentation after an unquoted value, among others.

#### Scenario: Attribute outside a block

- **WHEN** a `@starthcl` block contains `a = "b"`
- **THEN** the diagram is the text `Your data does not sound like HCL data`, without an error

#### Scenario: Nested block without an equals sign

- **WHEN** a `@starthcl` block contains `module "m" {`, `  nested {`, `    x = "y"`, `  }` and `}`
- **THEN** the diagram is the text `Your data does not sound like HCL data`, without an error, as upstream does

#### Scenario: Comment after a value

- **WHEN** a `@starthcl` block contains `a {`, `  b = "c" # trailing` and `}`
- **THEN** the diagram is the text `Your data does not sound like HCL data`, without an error, as upstream does

#### Scenario: Attributes without indentation

- **WHEN** a `@starthcl` block contains `a {`, `b = "c"`, `d = 1`, `e = 2` and `}`
- **THEN** the diagram is the text `Your data does not sound like HCL data`, without an error, as `1` and `e` are read as one value, as upstream does
