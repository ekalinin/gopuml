# Spec Delta

## MODIFIED Requirements

### Requirement: Description elements

The system SHALL support every element keyword upstream accepts in description diagrams: `actor`, `usecase`, `component`, `node`, `database`, `artifact`, `rectangle`, `card`, `cloud`, `folder`, `frame`, `file`, `package`, `queue`, `stack`, `storage`, `agent`, `person`, `boundary`, `control`, `entity`, `interface`, `circle`, `collections`, `hexagon`, `label`, `process`, `action` and `archimate` (see "Archimate elements"). It SHALL also support the business variants `actor/` and `usecase/`, the short forms `:Actor:`, `(Use case)`, `[Component]` and `() Interface`, display names with `as` aliases, stereotypes, colors, and element bodies spanning several lines in `[` ... `]`. Each element SHALL be drawn with upstream's symbol for its keyword, and the diagram type SHALL be description.

#### Scenario: Actor and use case

- **WHEN** a diagram contains `actor User`, `usecase Login` and `User --> Login`
- **THEN** `User` is drawn as a stick figure, `Login` as an ellipse, an arrow goes from `User` to `Login`, and the diagram type is description

#### Scenario: Short forms

- **WHEN** a diagram contains `:Admin: --> (Reset password)` and `[API] --> () REST`
- **THEN** `Admin` is drawn as an actor, `Reset password` as a use case, `API` as a component and `REST` as an interface circle

#### Scenario: Deployment symbols

- **WHEN** a diagram contains `node Server`, `database Orders` and `Server --> Orders`
- **THEN** `Server` is drawn with the node symbol and `Orders` with the database symbol

#### Scenario: Element body on several lines

- **WHEN** a diagram contains `rectangle Summary [`, the lines `first line` and `second line`, and `]`
- **THEN** one rectangle is drawn with both lines as its text

## ADDED Requirements

### Requirement: Archimate elements

The system SHALL support upstream's three `archimate` commands in description diagrams, as upstream does:

- `archimate #color <name> <<icon>>` SHALL create an element drawn with the archimate symbol. The name takes the forms of the other element keywords: a name or a quoted name, `"Display" as code`, `code as "Display"` and `Display as code`. When an element of that name exists, the command SHALL update its display, its icon and its color instead.
- `archimate #color code <<icon>> [`, followed by lines and a line ending with `]`, SHALL create an element whose text is the lines between the brackets, drawn with the rectangle symbol, as upstream does. It SHALL fail with `Already exists <code>` when an element of that name exists.
- `archimate #color <name> <<icon>> {` SHALL open a group drawn with the archimate symbol, closed by `}`.

The color and the icon are optional. The color is a background color, including upstream's ArchiMate colors `#Business`, `#Application`, `#Motivation`, `#Strategy`, `#Technology`, `#Physical` and `#Implementation`. Upstream's pattern wants spaces both before and after the optional color, so a command with one space after `archimate` and no color SHALL be a syntax error, as upstream does. The icon `<<name>>` is made of letters, digits, `_` and `-`. It SHALL become the stereotype `<<$archimate/name>>`, which draws the internal sprite `archimate/name` as the stereotype of the element (see the sprites capability). When the jar has no such sprite, it SHALL draw the stereotype `<<name>>` instead, as upstream does.

#### Scenario: Archimate element

- **WHEN** a diagram contains `archimate #Business "Customer" <<business-actor>> as c`
- **THEN** `Customer` is drawn with the archimate symbol, a rectangle filled with `#FFFFCC`, with the business-actor icon above its name, and the diagram type is description

#### Scenario: Archimate element without a color

- **WHEN** a diagram contains `archimate "Customer" as c`
- **THEN** the diagram fails at that line with `Syntax Error? (Assumed diagram type: sequence)`, as upstream does

#### Scenario: Archimate element updated

- **WHEN** a diagram contains `archimate #Business "Old" as u` and `archimate #Application "Updated" <<application-component>> as u`
- **THEN** one element is drawn, `Updated`, filled with `#C2F0FF` and with the application-component icon

#### Scenario: Archimate element on several lines

- **WHEN** a diagram contains `archimate #Business m <<business-actor>> [`, the lines `first line` and `second line`, and `]`
- **THEN** one element is drawn with the rectangle symbol, filled with `#FFFFCC`, with the business-actor icon above both lines, as upstream does

#### Scenario: Archimate element on several lines defined twice

- **WHEN** a diagram contains `archimate #Business c [`, `first`, `]`, `archimate #Business c [`, `second` and `]`
- **THEN** the diagram fails at the second `]` with `Already exists c (Assumed diagram type: component)`

#### Scenario: Archimate group

- **WHEN** a diagram contains `archimate #Technology "Domain" <<technology-node>> as d {`, `archimate #Application "App" as a` and `}`
- **THEN** `Domain` is drawn as a group filled with `#C9FFC9`, with the technology-node icon above its name, around the element `App`

#### Scenario: Archimate icon without a sprite

- **WHEN** a diagram contains `archimate #Business "Unknown icon" <<nope>> as n`
- **THEN** the element draws the stereotype `«nope»` above `Unknown icon`, as upstream does
