# Spec Delta

## MODIFIED Requirements

### Requirement: Embedded diagrams

The system SHALL render a `{{` ... `}}` block in creole text (notes, activity labels, class bodies and the other texts upstream renders with creole) as upstream's `EmbeddedDiagram` does: the lines between the markers form a diagram of the type named after `{{` (`uml` when no type is named), which is rendered and drawn in place at its own size. SVG output SHALL embed the nested diagram as SVG and PNG output as an image, as upstream does. Nested `{{` blocks SHALL be supported. A nested diagram of a type that is not supported yet SHALL be drawn as its "not supported yet" error image.

#### Scenario: Use case diagram in a note

- **WHEN** a note of a sequence diagram contains `{{`, `actor User`, `usecase Login`, `User --> Login` and `}}`
- **THEN** the note contains the use case diagram, with `User` drawn as an actor and `Login` as a use case

#### Scenario: Sequence diagram in an activity

- **WHEN** an activity label contains `{{`, `Alice -> Bob : hello` and `}}`
- **THEN** the activity box contains the sequence diagram with the message `hello`

#### Scenario: Embedded diagram of an unsupported type

- **WHEN** a note contains `{{`, `robust "Web" as WB` and `}}`
- **THEN** the note contains the error image stating that timing diagrams are not supported yet, and the rest of the diagram renders normally
