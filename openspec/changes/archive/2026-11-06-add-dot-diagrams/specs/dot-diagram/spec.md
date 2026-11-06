# Spec Delta

## Purpose

Supports PlantUML v1.2026.9 dot diagrams under `@startdot`: a Graphviz graph that the layout engine draws itself, with upstream's block rules, errors, security profile check, merged layout engine messages and failure images.

## ADDED Requirements

### Requirement: Dot diagram blocks

The system SHALL read `@startdot` ... `@enddot` blocks as upstream does. The `skinparam` and `!pragma` lines and the empty lines that directly follow the start line SHALL be dropped, and so SHALL the empty lines that follow them. The next line SHALL be a Graphviz graph header, alone on its line: an optional `strict`, `graph` or `digraph`, an optional name (an identifier, a number or a quoted string) and `{`. That line and every following line up to the end line, after preprocessing, SHALL be the DOT source of the diagram. A block without content SHALL fail at its end line with the message `Empty description`, and a block whose first line is not a graph header SHALL fail at that line with the message `Syntax Error?`, as upstream does. The diagram SHALL have one page and the description `(Dot)`.

#### Scenario: Directed graph

- **WHEN** the security profile is `INTERNET_WITH_DOTSVG` and a `@startdot` block contains `digraph G {`, `a -> b`, `b -> c [label="x"]` and `}`
- **THEN** the SVG shows `a`, `b` and `c` as ellipses from top to bottom, with arrows from `a` to `b` and from `b` to `c`, the second labeled `x`

#### Scenario: Strict undirected graph with a quoted name

- **WHEN** the security profile is `INTERNET_WITH_DOTSVG` and a `@startdot` block contains `strict graph "my g" {`, `a -- b`, `a -- b` and `}`
- **THEN** the SVG shows `a` above `b` and one line between them, without arrowheads

#### Scenario: Unnamed graph

- **WHEN** the security profile is `INTERNET_WITH_DOTSVG` and a `@startdot` block contains `digraph {`, `a -> b` and `}`
- **THEN** the SVG shows `a` above `b` with an arrow from `a` to `b`

#### Scenario: Skin parameter before the graph

- **WHEN** the security profile is `INTERNET_WITH_DOTSVG` and a `@startdot` block contains `skinparam monochrome true`, `digraph G {`, `a -> b` and `}`
- **THEN** the SVG shows the graph with `a` and `b`, as the `skinparam` line is dropped, as upstream does

#### Scenario: Preprocessor in a dot block

- **WHEN** the security profile is `INTERNET_WITH_DOTSVG` and a `@startdot` block contains `!define NODE alpha`, `digraph G {`, `' a comment`, `NODE -> b` and `}`
- **THEN** the SVG shows `alpha` above `b` with an arrow between them, and the comment is not part of the DOT source

#### Scenario: Empty dot block

- **WHEN** a `@startdot` block is followed directly by `@enddot`
- **THEN** the diagram fails at line 2 with the message `Empty description (Assumed diagram type: dot)`

#### Scenario: Block without a graph header

- **WHEN** a `@startdot` block contains `foo -> bar`
- **THEN** the diagram fails at line 2 with the message `Syntax Error? (Assumed diagram type: dot)`

### Requirement: SVG of dot diagrams and the security profile

In the `INSECURE` and `INTERNET_WITH_DOTSVG` security profiles, the SVG of a dot diagram SHALL be the output of the layout engine for the SVG format, written unchanged and without source metadata. In every other profile (`SANDBOX`, `ALLOWLIST`, `INTERNET` and `LEGACY`), the SVG SHALL be upstream's suppressed-feature document, which shows the text `This feature has been suppressed` as a link to `https://github.com/plantuml/plantuml/issues/2495`, and the layout engine SHALL NOT run, as upstream does. The diagram SHALL then not report an error.

#### Scenario: SVG suppressed by default

- **WHEN** no security profile is set and a `@startdot` block contains `digraph G {`, `a -> b` and `}`
- **THEN** the SVG is the suppressed-feature document with the text `This feature has been suppressed`, and shows no graph

#### Scenario: SVG suppressed in the INTERNET profile

- **WHEN** the security profile is `INTERNET` and a `@startdot` block contains `digraph G {`, `a -> b` and `}`
- **THEN** the SVG is the suppressed-feature document with the text `This feature has been suppressed`

#### Scenario: Suppressed before the layout

- **WHEN** no security profile is set and a `@startdot` block contains `digraph G {`, `a ->` and `}`, which is not valid DOT
- **THEN** the SVG is the suppressed-feature document, without an error, as the layout engine does not run

#### Scenario: SVG in the INSECURE profile

- **WHEN** the security profile is `INSECURE` and a `@startdot` block contains `digraph G {`, `a -> b` and `}`
- **THEN** the SVG shows `a` above `b` with an arrow from `a` to `b`

### Requirement: Layout engine of dot diagrams

The system SHALL lay dot diagrams out with the layout engine of the render: the built-in engine (gogv), or the `dot` named by the caller or by `GRAPHVIZ_DOT`, which SHALL run with `-T<format>` and the DOT source on its standard input. As upstream merges the standard error of `dot` into its output, the output SHALL be what the layout engine writes on its standard output and standard error, even when it is not an SVG document, and the diagram SHALL NOT report an error for it, as upstream does. With the built-in engine, the output SHALL hold the warnings that Graphviz writes for the same DOT source, each as a `Warning: <text>` line before the document, and the error of invalid DOT as an `Error: <stdin>: <message>` line, worded as Graphviz words them. The built-in engine's warnings about Graphviz features that it does not support yet SHALL NOT be written.

#### Scenario: Laid out by the built-in engine

- **WHEN** no `dot` is named, the security profile is `INTERNET_WITH_DOTSVG` and a `@startdot` block contains `digraph G {`, `a -> b` and `}`
- **THEN** the SVG is the document that the built-in engine writes for `dot -Tsvg`, whose comment names gogv

#### Scenario: Laid out by the named dot

- **WHEN** `GRAPHVIZ_DOT` names a working `dot`, the security profile is `INTERNET_WITH_DOTSVG` and a `@startdot` block contains `digraph G {`, `a -> b` and `}`
- **THEN** that `dot` runs with `-Tsvg`, and the SVG is its output, unchanged

#### Scenario: Syntax error in the DOT source

- **WHEN** the security profile is `INTERNET_WITH_DOTSVG` and a `@startdot` block contains `digraph G {`, `a ->` and `}`
- **THEN** the output is the text `Error: <stdin>: syntax error in line 3 near '}'`, which is not an SVG document, and the diagram reports no error, as upstream does

#### Scenario: Graphviz warning in the output

- **WHEN** the security profile is `INTERNET_WITH_DOTSVG` and a `@startdot` block contains `digraph G {`, `a -> 1b` and `}`
- **THEN** the output starts with the line `Warning: syntax ambiguity - badly delimited number '1b' in line 2 of <stdin> splits into two tokens`, followed by the SVG document of the graph, as upstream does

#### Scenario: Feature not supported by the built-in engine

- **WHEN** no `dot` is named, the security profile is `INTERNET_WITH_DOTSVG` and a `@startdot` block contains `digraph G {`, `a [shape=record]` and `}`
- **THEN** the output is the SVG document of the built-in engine, without a warning line

### Requirement: PNG of dot diagrams

The PNG of a dot diagram SHALL NOT depend on the security profile. With a named `dot`, the PNG SHALL be the output of that `dot` for `-Tpng`, written unchanged. The built-in engine does not write PNG, so without a named `dot` the PNG SHALL be upstream's error image with the text `GraphViz has crashed`, without source metadata and with an error status, as upstream draws it when its Graphviz has no PNG renderer.

#### Scenario: PNG without Graphviz

- **WHEN** no `dot` is named and a `@startdot` block that contains `digraph G {`, `a -> b` and `}` is rendered as PNG
- **THEN** the PNG shows `GraphViz has crashed`, and the command line writes it and exits with status 200

#### Scenario: PNG with the named dot

- **WHEN** `GRAPHVIZ_DOT` names a `dot` that writes PNG, and a `@startdot` block that contains `digraph G {`, `a -> b` and `}` is rendered as PNG
- **THEN** that `dot` runs with `-Tpng`, and the PNG is its output, unchanged

### Requirement: Layout engine failures in dot diagrams

When the named `dot` cannot be used, the image of a dot diagram SHALL show `There is an issue with your Dot/Graphviz installation`. When the layout engine stops without output, fails to run or runs longer than the timeout, the image SHALL show `GraphViz has crashed`. Both images SHALL have no source metadata and an error status, as upstream's. In the profiles that suppress the SVG, the SVG SHALL be the suppressed-feature document, as the layout engine does not run.

#### Scenario: Named dot is missing

- **WHEN** `GRAPHVIZ_DOT` is set to `/missing/dot`, no file exists there, the security profile is `INTERNET_WITH_DOTSVG`, and a `@startdot` block that contains `digraph G {`, `a -> b` and `}` is rendered as SVG
- **THEN** the image shows `There is an issue with your Dot/Graphviz installation`, and the diagram has an error status

#### Scenario: Layout timeout in a dot diagram

- **WHEN** the timeout is set to 1 second, the security profile is `INTERNET_WITH_DOTSVG`, the named `dot` does not finish within 1 second, and a `@startdot` block that contains `digraph G {`, `a -> b` and `}` is rendered as SVG
- **THEN** the `dot` process is stopped, the image shows `GraphViz has crashed`, and the diagram has an error status

### Requirement: Other outputs of dot diagrams

The system SHALL report that TXT and UTXT output is not supported for dot diagrams, and produce no output. The syntax check SHALL report a dot diagram as upstream does: the type `OTHER` and the description `(Dot)`.

#### Scenario: TXT dot diagram

- **WHEN** a `@startdot` block that contains `digraph G {`, `a -> b` and `}` is rendered as TXT
- **THEN** the system reports that TXT output is not supported for dot diagrams and produces no output

#### Scenario: Syntax check of a dot diagram

- **WHEN** the command line runs with `-syntax` and reads a `@startdot` block that contains `digraph G {`, `a -> b` and `}` from its standard input
- **THEN** it prints `OTHER` and `(Dot)`, and exits with status 0
