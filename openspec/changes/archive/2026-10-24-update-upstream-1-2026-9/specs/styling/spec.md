# Spec Delta

## MODIFIED Requirements

### Requirement: skinparam

The system SHALL support `skinparam` lines and `skinparam` blocks with the parameter names upstream supports for the diagram types in this change. Parameters SHALL be mapped onto the style system the same way as upstream. Unknown parameter names SHALL be ignored without error, as upstream does. `skinparam handwritten` and `skinparam ParticipantPadding` SHALL produce upstream's warnings, and `skinparam padding` SHALL produce none.

#### Scenario: Background color

- **WHEN** a diagram contains `skinparam backgroundColor #EEEEEE`
- **THEN** the rendered diagram background is `#EEEEEE`

#### Scenario: Element-specific parameter

- **WHEN** a sequence diagram contains `skinparam sequenceArrowColor red`
- **THEN** message arrows are drawn in red

#### Scenario: Ignored participant padding

- **WHEN** a sequence diagram contains `skinparam ParticipantPadding 40` and `A -> B`
- **THEN** the warning lines `'skinparam ParticipantPadding' is ignored: use a style instead`, `<style>`, `participant {`, `  Padding 40`, `}` and `</style>` are shown above the diagram

#### Scenario: Padding without warning

- **WHEN** a sequence diagram contains `skinparam padding 5` and `A -> B`
- **THEN** no warning is shown

### Requirement: Style blocks

The system SHALL support `<style>`...`</style>` blocks with upstream's selector rules: element selectors, diagram selectors (for example `sequenceDiagram`), stereotype selectors (`.name`), nesting, and the upstream property names. When `skinparam` and `<style>` set the same property, the result SHALL match upstream precedence. The stereotype label of a class SHALL take its font from the style of `stereotype` under `class` in `classDiagram`, as upstream does.

#### Scenario: Stereotype selector

- **WHEN** a class diagram contains a style rule `.important { BackgroundColor pink }` and `class Order <<important>>`
- **THEN** the `Order` class is filled pink and other classes keep the default fill

#### Scenario: Class stereotype label

- **WHEN** a class diagram contains a style rule `root { FontColor white }` and `class HelloService <<Service>>`
- **THEN** the stereotype label `«Service»` is drawn in white

### Requirement: Built-in themes

The system SHALL bundle every theme that upstream v1.2026.9 ships, subject to the license check in the design. It SHALL apply a theme from `!theme <name>`, from `!theme <name> from <directory>` (a local theme file), and from the `--theme` CLI flag. An unknown theme name SHALL produce an error at the directive line. A local theme file that does not exist SHALL produce the error `Cannot load theme <name> in <directory>` at the directive line.

#### Scenario: Bundled theme

- **WHEN** a diagram contains `!theme cerulean`
- **THEN** the diagram renders with the cerulean theme colors

#### Scenario: Unknown theme

- **WHEN** a diagram contains `!theme does-not-exist`
- **THEN** the system reports an error at that line naming the theme

#### Scenario: Missing local theme

- **WHEN** a diagram contains `!theme mine from themes`, and the directory `themes` holds no `puml-theme-mine.puml`
- **THEN** the system reports an error at that line with the message `Cannot load theme mine in themes`
