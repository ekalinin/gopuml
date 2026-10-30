# Spec Delta

## MODIFIED Requirements

### Requirement: Diagram support without Graphviz

In the WASM build, diagram types with a built-in layout (sequence, activity in the new syntax, packet, gantt, mindmap, WBS, timing and salt) SHALL render fully. Graph-based diagrams (class, description, state and legacy activity) SHALL return the Graphviz-not-available error until the pure Go layout engine is added.

#### Scenario: Class diagram in WASM

- **WHEN** the WASM render function is called with a class diagram
- **THEN** it returns an error stating that graph layout is not available in this build, together with the error image

#### Scenario: Description diagram in WASM

- **WHEN** the WASM render function is called with a deployment diagram
- **THEN** it returns an error stating that graph layout is not available in this build, together with the error image

#### Scenario: State diagram in WASM

- **WHEN** the WASM render function is called with a state diagram
- **THEN** it returns an error stating that graph layout is not available in this build, together with the error image

#### Scenario: Legacy activity diagram in WASM

- **WHEN** the WASM render function is called with a legacy activity diagram
- **THEN** it returns an error stating that graph layout is not available in this build, together with the error image

#### Scenario: Packet diagram in WASM

- **WHEN** the WASM render function is called with a packet diagram and format `svg`
- **THEN** it returns the SVG document without an error, byte-identical to the native library's output for the same input

#### Scenario: Gantt diagram in WASM

- **WHEN** the WASM render function is called with a gantt diagram and format `svg`
- **THEN** it returns the SVG document without an error, byte-identical to the native library's output for the same input

#### Scenario: Mindmap diagram in WASM

- **WHEN** the WASM render function is called with a mindmap diagram and format `svg`
- **THEN** it returns the SVG document without an error, byte-identical to the native library's output for the same input

#### Scenario: WBS diagram in WASM

- **WHEN** the WASM render function is called with a WBS diagram and format `svg`
- **THEN** it returns the SVG document without an error, byte-identical to the native library's output for the same input

#### Scenario: Timing diagram in WASM

- **WHEN** the WASM render function is called with a timing diagram and format `svg`
- **THEN** it returns the SVG document without an error, byte-identical to the native library's output for the same input

#### Scenario: Salt diagram in WASM

- **WHEN** the WASM render function is called with a salt diagram and format `svg`
- **THEN** it returns the SVG document without an error, byte-identical to the native library's output for the same input
