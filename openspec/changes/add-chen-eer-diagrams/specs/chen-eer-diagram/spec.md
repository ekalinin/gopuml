# Spec Delta

## Purpose

Supports PlantUML v1.2026.9 Chen EER diagrams under `@startchen`, with the same entities, relationships, attributes, associations, subclasses, stereotypes, colors, styles, errors and Graphviz-based layout as upstream.

## ADDED Requirements

### Requirement: Chen diagram blocks

The system SHALL render `@startchen` ... `@endchen` blocks as upstream's Chen EER diagrams, laid out with Graphviz `dot` as upstream lays them out. The diagram type SHALL be chen_eer and the description `(N entities)`, where N counts the entities, the relationships, the attributes and the circles of the subclass groups. The common commands of upstream's Chen diagrams SHALL be supported: title, header, footer, caption, legend, mainframe, `skin`, `skinparam`, `<style>`, sprites, `scale`, pragmas and the common hide commands. A block opened with `{` and never closed SHALL NOT be an error, as upstream does. A `}` that closes no block SHALL fail with the message `Unbalanced brackets`.

#### Scenario: Entities, relationship and attributes

- **WHEN** a `@startchen` block contains `entity CUSTOMER {`, `Number`, `}`, `entity MOVIE {`, `Code`, `}`, `relationship RENTED_TO {`, `Date`, `}`, `RENTED_TO -1- CUSTOMER` and `RENTED_TO -N- MOVIE`
- **THEN** `CUSTOMER` and `MOVIE` are drawn as rectangles, `RENTED_TO` as a diamond, `Number`, `Code` and `Date` as ellipses linked to their owners, the links from `RENTED_TO` are labeled `1` and `N`, the diagram type is chen_eer and the description is `(6 entities)`

#### Scenario: Chen diagram without dot

- **WHEN** no `dot` executable can be found and a Chen diagram is rendered
- **THEN** the system produces the error diagram stating that Graphviz `dot` was not found

#### Scenario: Unclosed block

- **WHEN** a Chen diagram contains `entity A {` and `Name`, without `}`
- **THEN** `A` is drawn with the attribute `Name` linked to it, without an error

#### Scenario: Unbalanced brackets

- **WHEN** a Chen diagram contains `entity A {`, `}` and `}`
- **THEN** the diagram fails at the second `}` with the message `Unbalanced brackets (Assumed diagram type: chen_eer)`

### Requirement: Entities and relationships

The system SHALL support entities (`entity CODE {`) and relationships (`relationship CODE {`), each followed by the attributes of its block, with an optional display name (`entity "Name" as CODE {`), an optional stereotype and an optional color, as upstream reads them: the main color is the background color, and `line:<color>` sets the border. An entity SHALL be drawn as a rectangle, a double rectangle with the stereotype `<<weak>>`. A relationship SHALL be drawn as a diamond, a double diamond with the stereotype `<<identifying>>`. A name may contain dots, and SHALL then be drawn as is, without a package. Declaring an entity again with the same type SHALL add the attributes of the new block to it. Declaring it again with another type SHALL fail with the message `Bad name`.

#### Scenario: Display names

- **WHEN** a Chen diagram contains `entity "Director" as DIRECTOR {`, `}`, `relationship "was-rented-to" as RENTED_TO {`, `}` and `RENTED_TO -1- DIRECTOR`
- **THEN** the rectangle is labeled `Director` and the diamond `was-rented-to`

#### Scenario: Weak entity and identifying relationship

- **WHEN** a Chen diagram contains `entity PARENT {`, `}`, `entity CHILD <<weak>> {`, `}`, `relationship PARENT_OF <<identifying>> {`, `}`, `PARENT_OF -1- PARENT` and `PARENT_OF -N- CHILD`
- **THEN** `PARENT` is drawn as a single rectangle, `CHILD` as a double rectangle and `PARENT_OF` as a double diamond

#### Scenario: Entity and relationship colors

- **WHEN** a Chen diagram contains `entity PERSON #pink;line:red {`, `}`, `relationship OWNS #lightblue {`, `}` and `OWNS -1- PERSON`
- **THEN** `PERSON` is filled with pink and outlined in red, and `OWNS` is filled with light blue and outlined in the default color

#### Scenario: Entity declared again

- **WHEN** a Chen diagram contains `entity A {`, `x`, `}`, `entity A {`, `y` and `}`
- **THEN** one rectangle `A` is drawn, with the attributes `x` and `y` linked to it

#### Scenario: Relationship with the name of an entity

- **WHEN** a Chen diagram contains `entity A {`, `}`, `relationship A {` and `}`
- **THEN** the diagram fails at the line `relationship A {` with the message `Bad name (Assumed diagram type: chen_eer)`

#### Scenario: Dotted names

- **WHEN** a Chen diagram contains `entity shop.Customer {`, `first.name` and `}`
- **THEN** a rectangle labeled `shop.Customer` is drawn with an ellipse labeled `first.name` linked to it, and no package is drawn

### Requirement: Attributes

The system SHALL read every line of a block that no other command matches as an attribute of the entity, the relationship or the composite attribute that the block belongs to, as upstream does: the name may contain letters, digits, spaces, `_`, `.` and `:`, and may be preceded by a display name (`"No." as Number`) and followed by a stereotype, a color and `{`, which opens the block of a composite attribute. An attribute SHALL be drawn as an ellipse around its name, linked to its owner. The stereotype `<<key>>` SHALL underline the name, `<<multi>>` SHALL draw a double ellipse and `<<derived>>` a dashed ellipse. The main color of an attribute SHALL be the line color, which also colors its link. The same attribute name SHALL be allowed in different owners. An attribute outside any block SHALL fail with the message `Attribute must be inside an entity, relationship or another attribute`, and an attribute declared twice in the same owner with the message `Attribute already exists`.

#### Scenario: Composite attribute

- **WHEN** a Chen diagram contains `entity DIRECTOR {`, `Name {`, `Fname`, `Lname`, `}` and `}`
- **THEN** `Name` is linked to `DIRECTOR`, and `Fname` and `Lname` are linked to `Name`

#### Scenario: Key, multi-valued and derived attributes

- **WHEN** a Chen diagram contains `entity CUSTOMER {`, `Number <<key>>`, `Name <<multi>>`, `Bonus <<derived>>` and `}`
- **THEN** `Number` is underlined in a single ellipse, `Name` is drawn in a double ellipse, and `Bonus` in a dashed ellipse

#### Scenario: Line read as an attribute

- **WHEN** a Chen diagram contains `entity A {`, `class Foo`, `Born : DATE` and `}`
- **THEN** two attributes labeled `class Foo` and `Born : DATE` are linked to `A`

#### Scenario: Attribute colors

- **WHEN** a Chen diagram contains `entity PERSON {`, `Name #lime;line:orange` and `}`
- **THEN** the ellipse of `Name` is outlined in orange and filled with the default color, and its link to `PERSON` is orange

#### Scenario: Same attribute in two entities

- **WHEN** a Chen diagram contains `entity DIRECTOR {`, `Number`, `}`, `entity CUSTOMER {`, `Number` and `}`
- **THEN** two ellipses labeled `Number` are drawn, one linked to `DIRECTOR` and one to `CUSTOMER`

#### Scenario: Attribute outside a block

- **WHEN** a Chen diagram contains `entity A {`, `}` and `Name`
- **THEN** the diagram fails at the line `Name` with the message `Attribute must be inside an entity, relationship or another attribute (Assumed diagram type: chen_eer)`

#### Scenario: Attribute declared twice

- **WHEN** a Chen diagram contains `entity A {`, `Name`, `Name` and `}`
- **THEN** the diagram fails at the second `Name` with the message `Attribute already exists (Assumed diagram type: chen_eer)`

### Requirement: Associations

The system SHALL support associations between two entities or relationships, written `A -<cardinality>- B`, where the cardinality is a word (`1`, `N`) or a pair (`(N,M)`, with optional spaces after the comma) and may be omitted. The cardinality SHALL be drawn as the label of the link. A `=` instead of the first `-` SHALL draw the link bold, as upstream does; the character after the cardinality SHALL NOT change the line. The main color of an association SHALL be the line color. An association with a name that is not an entity or a relationship SHALL fail with the message `No such entity: <name>`.

#### Scenario: Cardinalities

- **WHEN** a Chen diagram contains the entities `CUSTOMER` and `DIRECTOR`, the relationship `RENTED_TO`, `RENTED_TO -1- CUSTOMER` and `RENTED_TO -(N, M)- DIRECTOR`
- **THEN** the link to `CUSTOMER` is labeled `1` and the link to `DIRECTOR` is labeled `(N, M)`

#### Scenario: Total participation

- **WHEN** a Chen diagram contains the entity `CHILD`, the relationship `PARENT_OF` and `PARENT_OF =N= CHILD`
- **THEN** the link from `PARENT_OF` to `CHILD` is bold and labeled `N`

#### Scenario: Bold only from the left

- **WHEN** a Chen diagram contains the entities `A`, `B` and `C`, `A -N= B` and `A =N- C`
- **THEN** the link from `A` to `B` is drawn with a normal line and the link from `A` to `C` is bold

#### Scenario: Association color

- **WHEN** a Chen diagram contains the entity `PERSON`, the relationship `OWNS` and `OWNS -1- PERSON #lime`
- **THEN** the link from `OWNS` to `PERSON` is lime

#### Scenario: Association with an unknown entity

- **WHEN** a Chen diagram contains `entity A {`, `}` and `A -1- B`
- **THEN** the diagram fails at the line `A -1- B` with the message `No such entity: B (Assumed diagram type: chen_eer)`

### Requirement: Subclasses

The system SHALL support subclass links between two entities, `A ->- B` with the superset sign in the middle of the link and `A -<- B` with the subset sign, bold when the first character is `=`. The system SHALL support subclass groups, `A ->- <symbol> { B, C }` with the symbol `d`, `o` or `U`, drawn as a circle with the symbol, linked to `A` and to each subclass: with `d` and `o`, each link to a subclass SHALL have the superset sign; with `U`, the link to `A` SHALL have the superset sign instead. Only the link to `A` SHALL be bold when the first character is `=`. The main color of a subclass link SHALL be the line color, and that of a subclass group the background color of its circle, with `line:<color>` coloring the circle's border and every link of the group. A subclass link or group with a name that is not an entity or a relationship SHALL fail with the message `No such entity: <name>`.

#### Scenario: Subset and superset

- **WHEN** a Chen diagram contains the entities `CUSTOMER`, `PARENT` and `MEMBER`, `CUSTOMER ->- PARENT` and `MEMBER -<- CUSTOMER`
- **THEN** the link from `CUSTOMER` to `PARENT` has the superset sign in its middle, and the link from `MEMBER` to `CUSTOMER` has the subset sign

#### Scenario: Disjoint subclass group

- **WHEN** a Chen diagram contains the entities `CHILD`, `TODDLER` and `TEEN` and `CHILD =>= d { TODDLER, TEEN }`
- **THEN** a circle with `d` is linked to `CHILD` by a bold line without a sign, and to `TODDLER` and `TEEN` by lines with the superset sign

#### Scenario: Union

- **WHEN** a Chen diagram contains the entities `PERSON`, `CUSTOMER` and `DIRECTOR` and `PERSON ->- U { CUSTOMER, DIRECTOR } #lime;line:green`
- **THEN** a circle with `U`, filled with lime and outlined in green, is linked to `PERSON` by a green line with the superset sign, and to `CUSTOMER` and `DIRECTOR` by green lines without a sign

#### Scenario: Subclass group with an unknown entity

- **WHEN** a Chen diagram contains `entity A {`, `}`, `entity B {`, `}` and `A ->- d { B, C }`
- **THEN** the diagram fails at the line `A ->- d { B, C }` with the message `No such entity: C (Assumed diagram type: chen_eer)`

### Requirement: Chen diagram direction

The system SHALL support `left to right direction` and `top to bottom direction` in Chen diagrams, with the same effect on the layout as upstream.

#### Scenario: Left to right Chen diagram

- **WHEN** a Chen diagram contains `left to right direction`, the entities `Person` and `Location`, the relationship `Birthplace`, `Person -N- Birthplace` and `Birthplace -1- Location`
- **THEN** `Person`, `Birthplace` and `Location` are placed from left to right

### Requirement: Chen styles

The system SHALL style the elements of Chen diagrams with the `chenEerDiagram` styles, as upstream does: `chenEntity` for the entities, `chenRelationship` for the relationships, `chenAttribute` for the attributes, `circle` for the circles of the subclass groups, `title` under each of them for their names, and `.name` for the elements with the stereotype `name`. The title, the legend and the other annotations SHALL take the styles of `activityDiagram`, as upstream maps no style name of its own to Chen diagrams. A Chen diagram SHALL take the default style of the elements where no style is set: a light gray background, a dark gray border of 0.5 and black names.

#### Scenario: Element styles

- **WHEN** a Chen diagram contains a `<style>` block with `chenEerDiagram { chenEntity { BackGroundColor #FFFF00 } chenRelationship { BackGroundColor #00FFFF } chenAttribute { LineColor #0000FF } circle { BackGroundColor #FF00FF } }`, an entity `A` with the attribute `x`, a relationship `R`, an entity `S`, `R -1- A` and `A ->- d { S }`
- **THEN** `A` and `S` are filled with yellow, `R` with cyan, the circle `d` with magenta, and the ellipse of `x` is outlined in blue

#### Scenario: Stereotype style

- **WHEN** a Chen diagram contains a `<style>` block with `chenEerDiagram { .hot { BackGroundColor #FF0000 } }`, `entity H <<hot>> {`, `}`, `entity A {` and `}`
- **THEN** `H` is filled with red and `A` keeps the default background

#### Scenario: Title and legend

- **WHEN** a Chen diagram contains a `<style>` block with `activityDiagram { legend { BackGroundColor #FFFF00 } }` and `chenEerDiagram { legend { BackGroundColor #00FFFF } }`, `title Movies`, an entity `A`, `legend`, `Rentals` and `endlegend`
- **THEN** the title `Movies` is drawn above the diagram, and the legend `Rentals` is drawn below it and filled with yellow

### Requirement: Chen output formats

Chen diagrams SHALL render to SVG and PNG, and with the dark mode option in upstream's dark colors. TXT and UTXT SHALL report that the format is not supported for chen_eer diagrams and SHALL NOT produce output, as for class and description diagrams.

#### Scenario: PNG Chen diagram

- **WHEN** a Chen diagram with the entity `CUSTOMER`, its attribute `Number <<key>>`, the relationship `RENTED_TO` and `RENTED_TO -1- CUSTOMER` is rendered as PNG
- **THEN** the PNG image shows the rectangle, the ellipse, the diamond and the labeled link, as upstream's PNG does

#### Scenario: TXT Chen diagram

- **WHEN** a Chen diagram with the entity `CUSTOMER` is rendered as TXT
- **THEN** the system reports that TXT output is not supported for chen_eer diagrams and produces no output

#### Scenario: Dark mode

- **WHEN** a Chen diagram with the entity `CUSTOMER`, its attribute `Number <<key>>`, the relationship `RENTED_TO` and `RENTED_TO -1- CUSTOMER` is rendered as SVG with the dark mode option
- **THEN** the image has upstream's dark colors: a `#1B1B1B` background, the shapes filled with `#313139` and outlined in `#E7E7E7`, and white names

### Requirement: Embedded Chen diagrams

A `{{chen` ... `}}` block in creole text SHALL be drawn as the Chen diagram of its lines, as upstream draws it.

#### Scenario: Chen diagram in a note

- **WHEN** a sequence diagram has a note that contains `{{chen`, `entity CUSTOMER {`, `Number <<key>>`, `}`, `entity MOVIE {`, `}`, `CUSTOMER -N- MOVIE` and `}}`
- **THEN** the note contains the drawing of the Chen diagram, and the rest of the diagram renders normally
