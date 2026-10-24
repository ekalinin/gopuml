# Spec Delta

## MODIFIED Requirements

### Requirement: Relations

The system SHALL support upstream's relation syntax:

- extension, composition, aggregation, dependency, realization and plain links, solid or dotted
- the heads `#`, `x`, `}`, `+` and `^`
- labels, cardinalities on both ends, label arrows (`: label >`)
- roles on both ends (`A "1"/"-owner" --> "*"/"+items" B`), drawn with the field visibility icon of a leading visibility character and placed so that they overlap neither the line nor a class, as upstream places them
- direction hints (`-up->`, `-left->`, ...) and length through repeated dashes
- `[hidden]`, and colors and line styles in brackets (`-[#red,dashed]->`)
- association classes (`(A, B) .. C`)

#### Scenario: Composition with cardinalities

- **WHEN** a diagram contains `Order "1" *-- "many" Line : contains`
- **THEN** a composition is drawn from `Order` to `Line`, with cardinality `1` near `Order`, `many` near `Line`, and the label `contains`

#### Scenario: Direction hint

- **WHEN** a diagram contains `A -left-> B`
- **THEN** `B` is placed to the left of `A`

#### Scenario: Roles with visibility

- **WHEN** a diagram contains `class A`, `class B` and `A "1"/"-owner" --> "*"/"+items" B`
- **THEN** the role `owner` is drawn near `A` with the private field icon, the role `items` near `B` with the public field icon, and neither role overlaps the arrow or a class

### Requirement: Packages and namespaces

The system SHALL support `package` and `namespace` blocks, nesting, package display styles (for example `<<Node>>`, `<<Folder>>`, `<<Frame>>`, `<<Cloud>>`, `<<Database>>`, `<<Rectangle>>`), `set separator` (including a single backslash), and `together { }` grouping. A package SHALL also be declared with its code before a quoted display name (`package code as "Display"`).

#### Scenario: Nested package

- **WHEN** a diagram contains a `package shop {` block holding a `package model {` block that declares `class Order`, each closed by `}` on its own line
- **THEN** `Order` is drawn inside `model`, which is drawn inside `shop`

#### Scenario: Backslash separator

- **WHEN** a diagram contains `set separator \` and `class App\Entity\A`
- **THEN** `A` is drawn inside `Entity`, which is drawn inside `App`

#### Scenario: Package code before its display name

- **WHEN** a diagram contains `+package uid as "Hello" <<Frame>> {`, `class World` and `}`
- **THEN** a frame titled `Hello` is drawn around `World`
