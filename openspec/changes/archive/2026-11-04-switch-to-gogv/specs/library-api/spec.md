# Spec Delta

## MODIFIED Requirements

### Requirement: Typed errors

The library SHALL report failures as Go error values that callers can tell apart with `errors.Is` or `errors.As`:

- syntax errors, carrying the line number and message
- no diagram found
- unsupported diagram type
- unsupported format for a diagram type
- Graphviz not found, for a named `dot` that cannot be found
- Graphviz timeout or failure, of a named `dot` or of the built-in layout engine
- access denied by the security profile
- decoding errors

When error images are enabled, a syntax error SHALL be returned together with the error image bytes.

#### Scenario: Syntax error details

- **WHEN** a Go program renders source with a syntax error on a line
- **THEN** the returned error is a syntax error value exposing that line number and the message, and the error image bytes are returned

#### Scenario: Missing Graphviz

- **WHEN** a Go program renders a class diagram with `class A`, `class B` and `A --> B`, and the `dot` path option names a file that does not exist
- **THEN** the returned error matches the Graphviz-not-found error

### Requirement: Concurrency and cancellation

The library SHALL be safe for concurrent use from multiple goroutines, and concurrent renders SHALL NOT affect each other's output. Every render SHALL honor `context.Context` cancellation and deadlines, including stopping a running layout: a layout of the built-in engine, or a running named `dot` process.

#### Scenario: Concurrent renders

- **WHEN** two goroutines render different diagrams with different themes at the same time
- **THEN** each output is identical to rendering that diagram alone

#### Scenario: Cancelled context

- **WHEN** the context of a class diagram render is cancelled while the layout runs, with or without a named `dot`
- **THEN** the layout is stopped, a named `dot` process with it, and the render returns the context error
