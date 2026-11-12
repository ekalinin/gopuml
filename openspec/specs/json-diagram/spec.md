# json-diagram Specification

## Purpose
Supports PlantUML v1.2026.9 JSON diagrams under `@startjson`, and the drawing that YAML and HCL diagrams share with them: the boxes of objects and arrays, their layout, highlighted rows, styles, title and scale, the image for data that cannot be read, and the output formats.

## Requirements

### Requirement: JSON diagram blocks

The system SHALL read `@startjson` ... `@endjson` blocks as upstream does. Before the data, the `<style>` ... `</style>` blocks and the lines that start with `title `, `scale `, `skin `, `skinparam `, `!pragma `, `!assume ` or `hide ` SHALL be taken out of the data; after the first data line, only the `<style>` blocks SHALL be. A line that starts with `#highlight ` SHALL be a highlight path, another line that starts with `#` SHALL be dropped, and the other lines, after preprocessing, SHALL be the JSON text. The diagram SHALL have one page and the description `(Json)`, and the syntax check SHALL report it as `JSON`, as upstream does.

Every object and every array SHALL be drawn as a box of rows: an object has one row per member, with the key in the header font in a first column and the value in a second column, and an array has one row per item, in one column. A string SHALL be drawn as its text, a number as written in the source, `true` as `☑ true`, `false` as `☐ false` and `null` as `␀`. A member or item whose value is an object or an array SHALL show an empty value in its row and SHALL be linked to the box of that value by an arrow that starts at a dot on the row. A box without rows, for an empty object or array that is a value, SHALL be drawn 30 by 15 points. A root that is a string, a number, a boolean or `null` SHALL be drawn as an array of one item, and an empty root object or array as an array of one empty item, as upstream does.

#### Scenario: Object with nested values

- **WHEN** a `@startjson` block contains `{"firstName": "John", "age": 27, "address": {"city": "New York"}, "phones": ["212 555-1234", "646 555-4567"]}`
- **THEN** the first box has the rows `firstName` `John`, `age` `27`, `address` and `phones`, the row `address` is linked to a box with the row `city` `New York`, and the row `phones` to a box with the rows `212 555-1234` and `646 555-4567`

#### Scenario: Values of every kind

- **WHEN** a `@startjson` block contains `{"s": "text", "n": 42, "f": 1.5e3, "t": true, "no": false, "z": null}`
- **THEN** the box has the values `text`, `42`, `1.5e3`, `☑ true`, `☐ false` and `␀`

#### Scenario: Empty object and array values

- **WHEN** a `@startjson` block contains `{"e": {}, "ea": []}`
- **THEN** the rows `e` and `ea` are each linked to an empty box of 30 by 15 points

#### Scenario: Root value

- **WHEN** a `@startjson` block contains `"hello"`
- **THEN** the diagram is one box with the row `hello`

#### Scenario: Empty root object

- **WHEN** a `@startjson` block contains `{}`
- **THEN** the diagram is one box with one empty row, as upstream does

#### Scenario: Line starting with a hash

- **WHEN** a `@startjson` block contains `#foo bar` and `{"a": 1}`
- **THEN** the `#foo bar` line is dropped and the box has the row `a` `1`, as upstream does

#### Scenario: Syntax check of a JSON diagram

- **WHEN** the syntax check reads a `@startjson` block that contains `{"a": 1}`
- **THEN** it reports the type `JSON` and the description `(Json)`

### Requirement: Data that cannot be read

When the data of a JSON diagram is not JSON, or the block has no data, the system SHALL draw upstream's image of the text `Your data does not sound like JSON data` in a monospaced font, and SHALL NOT report an error, as upstream does. A YAML diagram SHALL draw `YAML` and an HCL diagram `HCL` in place of `JSON`.

#### Scenario: Invalid JSON

- **WHEN** a `@startjson` block contains `{"a": }`
- **THEN** the diagram is the text `Your data does not sound like JSON data`, without an error

#### Scenario: Empty JSON block

- **WHEN** a `@startjson` block is followed directly by `@endjson`
- **THEN** the diagram is the text `Your data does not sound like JSON data`, without an error

### Requirement: Layout of JSON diagrams

The system SHALL place the boxes as upstream places them: the root box on the left, each nested box in the next column to the right of the box that links to it, and the boxes linked from one box in the order of its rows, from top to bottom. Each arrow SHALL be a curve from the dot on its row to the left side of the nested box, ending with an arrowhead.

#### Scenario: Deep nesting

- **WHEN** a `@startjson` block contains `{"a": {"b": {"c": {"d": 1}}}}`
- **THEN** the diagram has four boxes from left to right, with the rows `a`, `b`, `c` and `d` `1`, each linked to the next

#### Scenario: Wide array

- **WHEN** a `@startjson` block contains `{"items": [` followed by twenty objects `{"id": <n>, "name": "item number <n>", "tags": ["x", "y"]}` for `<n>` from 1 to 20, and `]}`
- **THEN** the root box is linked to the box of `items`, whose twenty rows are linked to the twenty object boxes in one column, from top to bottom, and each object box is linked to its box of `x` and `y` in the next column

### Requirement: Highlighted rows

The system SHALL draw the rows that a highlight path names with the background of the `highlight` style, as upstream does. A path is one or more keys, quoted and separated by `/`: its last key names a row of the box that its other keys lead to from the root, an array item being named by its index. A `*` key SHALL match any key, and a `**` first key SHALL match the box at any depth. A path that ends with a stereotype (`<<name>>`) SHALL use the `highlight` style of that stereotype. YAML diagrams SHALL have highlight paths too; HCL diagrams SHALL NOT, as upstream does.

#### Scenario: Highlighted key

- **WHEN** a `@startjson` block contains `#highlight "lastName"` and `{"firstName": "John", "lastName": "Smith"}`
- **THEN** the row `lastName` has the background `#CCFF02` and the row `firstName` has none

#### Scenario: Highlight paths with wildcards

- **WHEN** a `@startjson` block contains `#highlight "a" / "b"`, `#highlight "*" / "c"`, `#highlight "**" / "d"` and `{"a": {"b": 2, "x": 3}, "k": {"c": 4, "d": 5}, "m": {"n": {"d": 6}}}`
- **THEN** the rows `b` of `a`, `c` and `d` of `k`, and `d` of `n` have the background `#CCFF02`, and the row `x` has none

#### Scenario: Highlight with a stereotype

- **WHEN** a `@startjson` block contains `<style>`, `.h1 { BackGroundColor lightblue }`, `</style>`, `#highlight "a" <<h1>>` and `{"a": 1, "b": 2}`
- **THEN** the row `a` has the background `#ADD8E6`

### Requirement: Styles of JSON diagrams

The system SHALL draw JSON diagrams with the `jsonDiagram` styles of upstream: `node` for the boxes and their values (background, line, corners, font), its `header` for the keys, its `separator` for the lines between rows and columns, its `highlight` for highlighted rows, and `arrow` for the arrows and their dots. YAML and HCL diagrams SHALL use the `yamlDiagram` styles instead, as upstream does. By default, the boxes SHALL have the background `#F1F1F1`, black lines of 1.5 points and rounded corners, the keys SHALL be bold, and the arrows SHALL be black lines of 1 point dashed `3,3`. Themes, `title`, `scale` and `skin` SHALL apply as upstream applies them.

#### Scenario: Unstyled JSON diagram

- **WHEN** a `@startjson` block contains `{"a": {"b": 1}}`
- **THEN** the two boxes have the background `#F1F1F1`, black lines of 1.5 points and rounded corners, the keys `a` and `b` are bold, and the arrow is a black dashed line

#### Scenario: Node and arrow styles

- **WHEN** a `@startjson` block contains `<style>`, `jsonDiagram {`, `node { BackGroundColor #eeeeee; LineColor blue }`, `arrow { LineColor red; LineThickness 2 }`, `}`, `</style>` and `{"a": {"b": 1}}`
- **THEN** the boxes have the background `#EEEEEE` and blue lines, and the arrow is red and 2 points thick

#### Scenario: Theme on a JSON diagram

- **WHEN** a `@startjson` block contains `!theme cerulean` and `{"a": {"b": 1}}`
- **THEN** the boxes have the gradient background and the `#2683B9` lines of the cerulean theme

#### Scenario: Title and scale

- **WHEN** a `@startjson` block contains `title My data`, `scale 2` and `{"a": 1}`
- **THEN** the title `My data` is drawn above the box, and the diagram is drawn at twice its size

### Requirement: Output formats of JSON, YAML and HCL diagrams

The system SHALL render JSON, YAML and HCL diagrams as SVG and as PNG. The TXT and UTXT formats SHALL be reported as not supported for these diagram types, with an error that names the format and the diagram type, as for the other diagram types without a text output.

#### Scenario: PNG of a JSON diagram

- **WHEN** a `@startjson` block that contains `{"a": {"b": [1, 2]}}` is rendered as PNG
- **THEN** the image shows the three boxes and their arrows, as the SVG does

#### Scenario: TXT of a JSON diagram

- **WHEN** a `@startjson` block that contains `{"a": 1}` is rendered as TXT
- **THEN** the system reports the error `TXT output is not supported for json diagrams` and writes no file
