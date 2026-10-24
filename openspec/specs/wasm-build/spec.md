# wasm-build Specification

## Purpose

Provides a WebAssembly build of the renderer that runs in browsers and other JavaScript hosts without a server, a JVM or Graphviz.

## Requirements

### Requirement: WebAssembly artifact

The project SHALL build a WebAssembly module from the same sources as the CLI, together with a JavaScript loader that runs in current browsers and in Node.js.

#### Scenario: Build produces the module

- **WHEN** the WASM build target runs
- **THEN** it produces the `.wasm` module and its JavaScript loader

### Requirement: JavaScript rendering API

The WASM build SHALL expose a JavaScript function that takes source text and options (format, diagram index, page index, dark mode, metadata) and returns the rendered SVG or PNG bytes together with the width, height and error details. The same input SHALL produce output identical to the native build's output.

#### Scenario: Render in a browser

- **WHEN** a web page calls the render function with a sequence diagram and format `svg`
- **THEN** it receives an SVG document byte-identical to the native library's output for the same input

### Requirement: Diagram support without Graphviz

In the WASM build, diagram types with a built-in layout (sequence, activity in the new syntax, and packet) SHALL render fully. Graph-based diagrams (class, description, state and legacy activity) SHALL return the Graphviz-not-available error until the pure Go layout engine is added.

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

### Requirement: Sandboxed resource access

The WASM build SHALL run in the `SANDBOX` security profile by default. It SHALL resolve includes only from sources supplied by the host page, through the render options.

#### Scenario: Include supplied by the host

- **WHEN** the host passes a file named `style.iuml` in the render options and the source contains `!include style.iuml`
- **THEN** the include is resolved from the supplied file
