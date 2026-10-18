# styling Specification

## Purpose

Controls the appearance of rendered diagrams the same way PlantUML v1.2026.8 does: `skinparam`, the `<style>` system, built-in themes, colors, creole text markup and common diagram annotations.

## Requirements

### Requirement: Default style

The system SHALL render diagrams with the same default style values as upstream's default style sheet (for example the default font family `SansSerif` and font size 14) when the source sets no styling.

#### Scenario: Unstyled diagram

- **WHEN** a sequence diagram without any styling directives is rendered
- **THEN** the colors, line widths and font sizes match the upstream default style within the parity tolerance defined by upstream-compatibility

### Requirement: skinparam

The system SHALL support `skinparam` lines and `skinparam` blocks with the parameter names upstream supports for the diagram types in this change. Parameters SHALL be mapped onto the style system the same way as upstream. Unknown parameter names SHALL be ignored without error, as upstream does.

#### Scenario: Background color

- **WHEN** a diagram contains `skinparam backgroundColor #EEEEEE`
- **THEN** the rendered diagram background is `#EEEEEE`

#### Scenario: Element-specific parameter

- **WHEN** a sequence diagram contains `skinparam sequenceArrowColor red`
- **THEN** message arrows are drawn in red

### Requirement: Style blocks

The system SHALL support `<style>`...`</style>` blocks with upstream's selector rules: element selectors, diagram selectors (for example `sequenceDiagram`), stereotype selectors (`.name`), nesting, and the upstream property names. When `skinparam` and `<style>` set the same property, the result SHALL match upstream precedence.

#### Scenario: Stereotype selector

- **WHEN** a class diagram contains a style rule `.important { BackgroundColor pink }` and `class Order <<important>>`
- **THEN** the `Order` class is filled pink and other classes keep the default fill

### Requirement: Built-in themes

The system SHALL bundle every theme that upstream v1.2026.8 ships, subject to the license check in the design. It SHALL apply a theme from `!theme <name>`, from `!theme <name> from <directory>` (a local theme file), and from the `--theme` CLI flag. An unknown theme name SHALL produce an error at the directive line.

#### Scenario: Bundled theme

- **WHEN** a diagram contains `!theme cerulean`
- **THEN** the diagram renders with the cerulean theme colors

#### Scenario: Unknown theme

- **WHEN** a diagram contains `!theme does-not-exist`
- **THEN** the system reports an error at that line naming the theme

### Requirement: Colors

The system SHALL accept the color syntax upstream accepts: HTML color names, `#RGB`, `#RRGGBB` and `#RRGGBBAA` hex values, and the gradient forms `#c1/c2`, `#c1|c2`, `#c1-c2` and `#c1\c2`.

#### Scenario: Named and hex colors

- **WHEN** a class diagram contains `class A #LightBlue` and `class B #f00`
- **THEN** `A` is filled with `#ADD8E6` and `B` with `#FF0000`

#### Scenario: Gradient fill

- **WHEN** a class diagram contains `class A #white/blue`
- **THEN** `A` is filled with a gradient from white to blue in the upstream direction for `/`

### Requirement: Creole text markup

The system SHALL render creole markup in every text that upstream renders with creole:

- bold `**`, italic `//`, monospace `""`, strike `--`, underline `__`, wave `~~`
- bulleted `*` and numbered `#` lists, tables with `|=` header cells, horizontal rules
- the HTML-like tags `<b>`, `<i>`, `<u>`, `<s>`, `<color>`, `<back>`, `<size>`, `<font>`
- links `[[url label]]`, the `~` escape, and Unicode escapes such as `<U+00E9>` and `&#233;`

OpenIconic (`<&name>`), emoji (`<:name:>`) and sprite (`<$name>`) references SHALL render as their literal text until the follow-up change that adds them.

#### Scenario: Bold and colored text

- **WHEN** a note contains `**important** <color:red>now</color>`
- **THEN** `important` is drawn bold and `now` is drawn in red

#### Scenario: Table in a note

- **WHEN** a note contains the lines `|= Key |= Value |` and `| a | 1 |`
- **THEN** the note renders a two-column table with a bold header row

#### Scenario: Icon reference before icons are supported

- **WHEN** a note contains `<&check> done`
- **THEN** the note renders the literal text `<&check> done`

### Requirement: Common diagram annotations

The system SHALL support `title`, `header`, `footer`, `caption`, `legend`/`endlegend` with their position options, `scale`, and `mainframe` for the diagram types in this change, with the same placement as upstream.

#### Scenario: Title and footer

- **WHEN** a diagram contains `title Checkout` and `footer page 1`
- **THEN** `Checkout` is drawn above the diagram and `page 1` below it

#### Scenario: Scale factor

- **WHEN** a diagram contains `scale 2`
- **THEN** the SVG width and height are twice those of the same diagram without `scale`, and the PNG width and height are twice those values within one pixel

### Requirement: Embedded diagrams

The system SHALL render a `{{` ... `}}` block in creole text (notes, activity labels, class bodies and the other texts upstream renders with creole) as upstream's `EmbeddedDiagram` does: the lines between the markers form a diagram of the type named after `{{` (`uml` when no type is named), which is rendered and drawn in place at its own size. SVG output SHALL embed the nested diagram as SVG and PNG output as an image, as upstream does. Nested `{{` blocks SHALL be supported. A nested diagram of a type that is not supported yet SHALL be drawn as its "not supported yet" error image.

#### Scenario: Use case diagram in a note

- **WHEN** a note of a sequence diagram contains `{{`, `actor User`, `usecase Login`, `User --> Login` and `}}`
- **THEN** the note contains the use case diagram, with `User` drawn as an actor and `Login` as a use case

#### Scenario: Sequence diagram in an activity

- **WHEN** an activity label contains `{{`, `Alice -> Bob : hello` and `}}`
- **THEN** the activity box contains the sequence diagram with the message `hello`

#### Scenario: Embedded diagram of an unsupported type

- **WHEN** a note contains `{{`, `[*] --> Idle` and `}}`
- **THEN** the note contains the error image stating that state diagrams are not supported yet, and the rest of the diagram renders normally
