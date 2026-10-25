# packet-diagram Specification

## Purpose
Supports PlantUML v1.2026.9 packet diagrams under `@startpacketdiag`, with the same syntax, layout and errors as upstream.

## Requirements

### Requirement: Packet diagram blocks

The system SHALL render `@startpacketdiag` ... `@endpacketdiag` blocks as upstream's packet diagrams. The fields and directives MAY be wrapped in `packetdiag {` ... `}` or `{` ... `}`, which SHALL not change the drawing. The diagram type SHALL be packet and the description `Packet Diagram`. The common commands of upstream's packet diagrams (title, header, footer, caption, legend, `skin`, `skinparam`, `<style>` and the other commands of upstream's `addCommonCommands1`) SHALL be supported. Packet diagrams SHALL NOT need Graphviz.

#### Scenario: Packet header

- **WHEN** a `@startpacketdiag` block contains `packetdiag {`, `0-7: Source Port`, `8-15: Destination Port`, `16-31: Sequence Number` and `}`
- **THEN** a bit scale numbered `0`, `8` and `16` is drawn above two rows, the first with the boxes `Source Port` and `Destination Port` side by side and the second with the box `Sequence Number` across the whole row, the diagram type is packet and the description is `Packet Diagram`

#### Scenario: Fields without a wrapper

- **WHEN** a `@startpacketdiag` block contains `0-7: Source Port`, `8-15: Destination Port` and `16-31: Sequence Number` without `packetdiag {`
- **THEN** the drawing is the same as with the wrapper

#### Scenario: Skin and title

- **WHEN** a `@startpacketdiag` block contains `skin rose`, `title TCP header`, `0-7: Source Port` and `8-15: Destination Port`
- **THEN** the title `TCP header` is drawn above the scale, and the boxes are drawn with the colors of the `rose` skin

#### Scenario: Packet diagram without Graphviz

- **WHEN** a packet diagram is rendered and `dot` cannot be found
- **THEN** the diagram renders without an error

### Requirement: Fields

The system SHALL support upstream's field declarations: `N: text` for a field that starts at bit `N`, `A-B: text` for the bits `A` to `B`, and `* text` for a field that starts after the last declared field, or at bit 0 when it is the first. A field SHALL be `B - A + 1` bits wide for a range, and as wide as its `len` attribute (1 by default) otherwise. The attributes SHALL be written in brackets after the text as `name = value` pairs separated by commas, with optionally quoted values. The `height` attribute SHALL multiply the height of the field (1 by default). The other attributes, `rotate` included, SHALL be accepted and ignored, as upstream does. Fields SHALL be drawn in declaration order, each as wide as its bits, so that, as upstream does, the bit numbers of a field only give its width.

#### Scenario: Fields placed after the previous one

- **WHEN** a packet diagram contains `* A [len = 4]` and `* B`
- **THEN** `A` is drawn 4 bits wide and `B` 1 bit wide to its right

#### Scenario: Field height

- **WHEN** a packet diagram contains `0-3: A [height = 3]`, `4-7: B` and `8-15: C`
- **THEN** `A` is drawn three times as high as `B` and `C`, as upstream draws it

#### Scenario: Ignored attributes

- **WHEN** a packet diagram contains `1-10: server [color="palegreen", textcolor="blue", linecolor="red"]` and `11-20: database [rotate = 270]`
- **THEN** both fields are drawn as they are without attributes

#### Scenario: Sparse bit numbers

- **WHEN** a packet diagram contains `2: A` and `5: B`
- **THEN** `A` and `B` are drawn next to each other, without a gap for the bits between them

### Requirement: Packet layout

The system SHALL lay out packet diagrams as upstream does. The fields SHALL fill rows of `colwidth` bits (16 by default). A field that does not fit in the rest of a row SHALL continue on the next rows, with its text repeated in each part. When the first row is narrower than `colwidth`, `colwidth` SHALL become the width of the first row. A bit SHALL be three times the font size wide and 32 pixels high, and a field SHALL be at least as high as its text with margins of 10 pixels. A text that does not fit in its field SHALL be shortened character by character and end with `...`. The scale above the rows SHALL have a tick for each bit position from 0 to `colwidth`, a full tick every quarter of `colwidth` (every bit when `colwidth` is less than 4) and a short tick otherwise, and the number of the bit position every `scale_interval` bits (half of `colwidth` by default).

#### Scenario: Field split across rows

- **WHEN** a packet diagram contains `colwidth = 8` and `0-15: Long field`
- **THEN** the scale is numbered `0`, `4` and `8`, and `Long field` is drawn in two rows of 8 bits, each part with the text `Long field`

#### Scenario: Shortened text

- **WHEN** a packet diagram contains `0: Checksum` and `1: B`
- **THEN** the first field shows `C...` and the second `B`

#### Scenario: Narrow first row

- **WHEN** a packet diagram contains `* A [len = 4]` and `* B`
- **THEN** the scale goes from bit 0 to bit 5, with only `0` numbered

### Requirement: Packet directives

The system SHALL support upstream's directives, each written as `name = value` with an optional `;`: `colwidth` (1 to 3 digits; 0 is ignored), `node_height` (the height of a bit in pixels, 1 to 3 digits), `scale_direction` (`ltr` or `rtl`), `scale_interval` (a positive number of bits, limited to `colwidth`; 0 is ignored) and `same_height` (`true` or `false`, which makes every field of a row as high as the highest one). With `scale_direction = rtl`, the scale SHALL be numbered from `colwidth` down to 0 and the fields of each row SHALL be drawn from right to left.

#### Scenario: Right to left

- **WHEN** a packet diagram contains `scale_direction = rtl`, `0-7: A` and `8-15: B`
- **THEN** the scale is numbered `16`, `8` and `0` from left to right, and `B` is drawn to the left of `A`

#### Scenario: Scale interval

- **WHEN** a packet diagram contains `scale_interval = 4` and `0-15: A`
- **THEN** the scale is numbered `0`, `4`, `8`, `12` and `16`

#### Scenario: Same height

- **WHEN** a packet diagram contains `same_height = true`, `0-3: A [height = 3]`, `4-7: B` and `8-15: C`
- **THEN** `A`, `B` and `C` are drawn with the same height, the height of `A`

#### Scenario: Node height

- **WHEN** a packet diagram contains `node_height = 80`, `0-7: A` and `8-15: B`
- **THEN** the fields are drawn 80 pixels high, as upstream draws them

### Requirement: Packet errors

The system SHALL report upstream's errors for packet diagrams. A line that is neither a field, a directive nor a common command SHALL be a syntax error on that line with the message `Syntax Error? (Assumed diagram type: packet)`. A diagram whose first row is 0 bits wide, such as one with the only field `4-3: x` or `* A [len = 0]`, SHALL produce upstream's error diagram with the message `Fatal crash error, you should send a mail to plantuml@gmail.com with: java.lang.ArithmeticException: / by zero` on the last line of the block. A field with a reversed range, such as `4-2: x`, SHALL make the rendering of the diagram fail with an error instead of a drawing, where upstream fails to draw it and writes a crash report.

#### Scenario: Syntax error

- **WHEN** a packet diagram contains `packetdiag {`, `foo: Source Port` and `}`
- **THEN** the system reports a syntax error on the line `foo: Source Port` with the message `Syntax Error? (Assumed diagram type: packet)`

#### Scenario: Empty range

- **WHEN** a packet diagram contains only `4-3: x`
- **THEN** the system produces an error diagram with the message `Fatal crash error, you should send a mail to plantuml@gmail.com with: java.lang.ArithmeticException: / by zero` on the `@endpacketdiag` line, as upstream does

#### Scenario: Reversed range

- **WHEN** a packet diagram contains only `4-2: x`
- **THEN** rendering the diagram returns an error and no drawing

### Requirement: Packet output formats

Packet diagrams SHALL render to SVG and PNG. TXT and UTXT output SHALL be rejected with the error that names the format and the diagram type, without output, as for the other diagram types that have no text output (every type except sequence and gantt).

#### Scenario: PNG packet diagram

- **WHEN** a packet diagram with `0-7: Source Port` and `8-15: Destination Port` is rendered as PNG
- **THEN** the PNG image shows the scale and the two fields, as upstream's PNG does

#### Scenario: TXT for a packet diagram

- **WHEN** a packet diagram is rendered as TXT
- **THEN** the system reports that TXT output is not supported for packet diagrams and produces no output

### Requirement: Embedded packet diagrams

A `{{packetdiag` ... `}}` block in creole text SHALL be drawn as the packet diagram it contains, as upstream draws it.

#### Scenario: Packet diagram in a note

- **WHEN** a sequence diagram has a note that contains `{{packetdiag`, `0-7: Source Port`, `8-15: Destination Port` and `}}`
- **THEN** the note contains the drawing of the packet diagram, and the rest of the diagram renders normally
