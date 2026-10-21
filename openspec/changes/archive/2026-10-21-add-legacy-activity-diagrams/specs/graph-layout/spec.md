# Spec Delta

## MODIFIED Requirements

### Requirement: Missing dot

When `dot` cannot be found, rendering a graph-based diagram SHALL produce an error diagram stating that Graphviz `dot` was not found and listing the ways to configure it. Diagram types that do not need `dot` (sequence, and activity in the new syntax) SHALL render normally.

#### Scenario: Class diagram without Graphviz

- **WHEN** no `dot` executable can be found and a class diagram is rendered
- **THEN** the system reports an error stating that Graphviz `dot` was not found, and produces an error image

#### Scenario: Sequence diagram without Graphviz

- **WHEN** no `dot` executable can be found and a sequence diagram is rendered
- **THEN** the diagram renders without error
