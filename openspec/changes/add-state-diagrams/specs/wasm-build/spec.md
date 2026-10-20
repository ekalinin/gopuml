# Spec Delta

## MODIFIED Requirements

### Requirement: Diagram support without Graphviz

In the WASM build, diagram types with a built-in layout (sequence and activity) SHALL render fully. Graph-based diagrams (class, description and state) SHALL return the Graphviz-not-available error until the pure Go layout engine is added.

#### Scenario: Class diagram in WASM

- **WHEN** the WASM render function is called with a class diagram
- **THEN** it returns an error stating that graph layout is not available in this build, together with the error image

#### Scenario: Description diagram in WASM

- **WHEN** the WASM render function is called with a deployment diagram
- **THEN** it returns an error stating that graph layout is not available in this build, together with the error image

#### Scenario: State diagram in WASM

- **WHEN** the WASM render function is called with a state diagram
- **THEN** it returns an error stating that graph layout is not available in this build, together with the error image
