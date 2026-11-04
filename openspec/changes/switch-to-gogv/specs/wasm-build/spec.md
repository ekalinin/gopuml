# Spec Delta

## MODIFIED Requirements

### Requirement: Diagram support without Graphviz

In the WASM build, every supported diagram type SHALL render fully: the diagram types with a built-in layout (sequence, activity in the new syntax, packet, gantt, mindmap, WBS, timing and salt), and the graph-based diagrams (class, description, state, legacy activity and Chen EER), which the built-in layout engine (gogv) lays out as in the native build without a named `dot`.

#### Scenario: Class diagram in WASM

- **WHEN** the WASM render function is called with a class diagram with `class A`, `class B` and `A --> B`, and format `svg`
- **THEN** it returns the SVG document without an error, byte-identical to the native library's output for the same input

#### Scenario: Description diagram in WASM

- **WHEN** the WASM render function is called with a deployment diagram with `node Server`, `database Orders` and `Server --> Orders`, and format `svg`
- **THEN** it returns the SVG document without an error, byte-identical to the native library's output for the same input

#### Scenario: State diagram in WASM

- **WHEN** the WASM render function is called with a state diagram with `[*] --> Idle`, `state Active {`, `[*] --> Running`, `}` and `Idle --> Active`, and format `svg`
- **THEN** it returns the SVG document without an error, byte-identical to the native library's output for the same input

#### Scenario: Legacy activity diagram in WASM

- **WHEN** the WASM render function is called with a legacy activity diagram with `(*) --> "First step"` and `"First step" --> (*)`, and format `svg`
- **THEN** it returns the SVG document without an error, byte-identical to the native library's output for the same input

#### Scenario: Chen diagram in WASM

- **WHEN** the WASM render function is called with a Chen EER diagram with `entity CUSTOMER {`, `Number <<key>>`, `}`, `relationship RENTED_TO {`, `}` and `RENTED_TO -1- CUSTOMER`, and format `svg`
- **THEN** it returns the SVG document without an error, byte-identical to the native library's output for the same input

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
