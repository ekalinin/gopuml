# Spec Delta

## Purpose

Provides an importable Go API to render PlantUML source to images and text, inspect diagrams and handle errors, so that Go programs can embed the renderer without running a separate process.

## ADDED Requirements

### Requirement: Render source

The library SHALL render PlantUML source text to a requested format and return the output bytes together with the diagram width, height, description, and the page and diagram counts. By default the first diagram and first page SHALL be rendered. The caller SHALL be able to select another diagram and page by index.

#### Scenario: Render to SVG

- **WHEN** a Go program renders `@startuml`, `Alice -> Bob`, `@enduml` to SVG
- **THEN** it receives a well-formed SVG document and the diagram width and height

#### Scenario: Select the second diagram

- **WHEN** a Go program renders source that contains two blocks and selects diagram index 1
- **THEN** it receives the output of the second block

### Requirement: Render options

The library SHALL accept, per call:

- output format, metadata on or off, error images on or off
- preprocessor definitions, include paths, config lines, theme, dark mode
- security profile, Graphviz `dot` path and timeout
- a base directory or a file system abstraction used to resolve includes

Options SHALL apply only to that call.

#### Scenario: Includes from a supplied file system

- **WHEN** a Go program renders source that contains `!include style.iuml` and supplies an in-memory file system holding `style.iuml`
- **THEN** the include is resolved from the supplied file system and no disk access is made

### Requirement: Typed errors

The library SHALL report failures as Go error values that callers can tell apart with `errors.Is` or `errors.As`:

- syntax errors, carrying the line number and message
- no diagram found
- unsupported diagram type
- unsupported format for a diagram type
- Graphviz not found
- Graphviz timeout or failure
- access denied by the security profile
- decoding errors

When error images are enabled, a syntax error SHALL be returned together with the error image bytes.

#### Scenario: Syntax error details

- **WHEN** a Go program renders source with a syntax error on a line
- **THEN** the returned error is a syntax error value exposing that line number and the message, and the error image bytes are returned

#### Scenario: Missing Graphviz

- **WHEN** a Go program renders a class diagram and `dot` cannot be found
- **THEN** the returned error matches the Graphviz-not-found error

### Requirement: Diagram inspection

The library SHALL list the blocks of a source with each block's name, family, detected diagram type and syntax status, without rendering images.

#### Scenario: List blocks

- **WHEN** a Go program inspects source that contains a sequence block named `login` and a class block
- **THEN** it receives two entries: the sequence block named `login`, and the class block with no name

### Requirement: Concurrency and cancellation

The library SHALL be safe for concurrent use from multiple goroutines, and concurrent renders SHALL NOT affect each other's output. Every render SHALL honor `context.Context` cancellation and deadlines, including stopping a running `dot` process.

#### Scenario: Concurrent renders

- **WHEN** two goroutines render different diagrams with different themes at the same time
- **THEN** each output is identical to rendering that diagram alone

#### Scenario: Cancelled context

- **WHEN** the context of a class diagram render is cancelled while `dot` is running
- **THEN** the `dot` process is stopped and the render returns the context error

### Requirement: Encoding helpers

The library SHALL expose the url-encoding functions to encode source to PlantUML's URL text format and decode it back.

#### Scenario: Encode and decode

- **WHEN** a Go program encodes a source and decodes the result
- **THEN** it gets back the original source
