# graph-layout Specification

## Purpose

Lays out graph-based diagrams with the built-in layout engine gogv, a Go implementation of Graphviz `dot`, or with the external Graphviz `dot` program when one is named, including how a named `dot` is found, timeouts and failure reporting.

## Requirements

### Requirement: Graphviz dot discovery

The system SHALL run an external Graphviz `dot` only when one is named, in this order:

1. an explicit path from the caller (the `--dot-path` CLI flag or the library option)
2. the `GRAPHVIZ_DOT` environment variable

When `GRAPHVIZ_DOT` is set, that path SHALL be used even if no file exists there, as upstream does. When no `dot` is named, the system SHALL lay graphs out with its built-in layout engine (gogv) and SHALL NOT search the directories in `PATH` or the standard locations, unlike upstream, which runs the `dot` it finds there.

#### Scenario: Environment variable wins over PATH

- **WHEN** `GRAPHVIZ_DOT` points to `/custom/dot` and another `dot` is on `PATH`
- **THEN** the system runs `/custom/dot`

#### Scenario: Found on PATH

- **WHEN** no explicit path and no `GRAPHVIZ_DOT` are set and `dot` is on `PATH`
- **THEN** the system lays the graph out with the built-in engine and does not run the `dot` found on `PATH`

#### Scenario: Environment variable points to a missing file

- **WHEN** `GRAPHVIZ_DOT` points to `/missing/dot` and another `dot` is on `PATH`
- **THEN** the system treats `dot` as not found, does not run the one on `PATH` and does not fall back to the built-in engine

### Requirement: Graph layout through dot

For graph-based diagrams, the system SHALL measure every node and label itself, have the layout engine of the render compute node positions and edge routes from the DOT graph, and draw the result itself. The layout engine SHALL be the built-in engine, which lays out the DOT language as Graphviz `dot` does without an installed Graphviz, or the named `dot`. With the built-in engine, the arrangement of nodes and edges SHALL match upstream's arrangement with the Graphviz of the reference image within the parity tolerance defined by upstream-compatibility, apart from the cases listed as known differences. With a named `dot` of the same Graphviz version, it SHALL match upstream's arrangement within that tolerance.

#### Scenario: Class diagram layout

- **WHEN** a class diagram with `class A`, `class B` and `A --> B` is rendered
- **THEN** `B` is placed below `A` and the arrow runs from `A` to `B`, as in upstream's output

#### Scenario: Layout with the named dot

- **WHEN** `GRAPHVIZ_DOT` names a working `dot` and a class diagram with `class A`, `class B` and `A --> B` is rendered
- **THEN** that `dot` lays the graph out, and the built-in engine is not used

### Requirement: Missing dot

When a named `dot` cannot be found, rendering a graph-based diagram SHALL produce an error diagram stating that Graphviz `dot` was not found and listing the ways to configure it. When no `dot` is named, graph-based diagrams SHALL render with the built-in engine, whether Graphviz is installed or not. Diagram types that do not need a graph layout (sequence, and activity in the new syntax) SHALL render normally.

#### Scenario: Class diagram without Graphviz

- **WHEN** no `dot` is named, no Graphviz is installed, and a class diagram with `class A`, `class B` and `A --> B` is rendered
- **THEN** the diagram is drawn without error, laid out by the built-in engine, as upstream also draws it, with its Smetana port

#### Scenario: Sequence diagram without Graphviz

- **WHEN** no `dot` executable can be found and a sequence diagram is rendered
- **THEN** the diagram renders without error

#### Scenario: Named dot is missing

- **WHEN** the `dot` path is set to `/missing/dot` (by `--dot-path` or the library option), no file exists there, and a class diagram with `class A`, `class B` and `A --> B` is rendered
- **THEN** the system reports an error stating that Graphviz `dot` was not found, and produces an error image

### Requirement: dot timeout

The system SHALL stop a layout that runs longer than the configured timeout, whether a named `dot` process or a layout of the built-in engine runs it, and report a timeout error for that diagram. The default timeout SHALL be 30 seconds, as in upstream, and SHALL be configurable in seconds (the `--graphviz-timeout` CLI flag or the library option).

#### Scenario: Timeout exceeded

- **WHEN** the timeout is set to 1 second and the named `dot` does not finish within 1 second
- **THEN** the `dot` process is terminated and the diagram reports a Graphviz timeout error

#### Scenario: Built-in layout timeout

- **WHEN** the timeout is set to 1 second and the built-in engine does not finish the layout of a class diagram within 1 second
- **THEN** the layout is stopped and the diagram reports a Graphviz timeout error

### Requirement: dot failure

When the layout engine fails, that is when a named `dot` exits with an error or returns output the system cannot use, or when the built-in engine returns an error or stops on an internal error, the system SHALL report an error for that diagram that includes the engine's error message, and SHALL continue with other diagrams.

#### Scenario: dot crashes on one diagram

- **WHEN** the named `dot` fails on the first of two class diagrams in one input
- **THEN** the first diagram reports a Graphviz error and the second diagram still renders

#### Scenario: Built-in engine fails on one diagram

- **WHEN** the built-in engine stops with an internal error on the first of two class diagrams in one input
- **THEN** the first diagram reports a Graphviz error with the engine's message, and the second diagram still renders

### Requirement: Graphviz check

The system SHALL provide a check that prints the Graphviz version and whether a test layout succeeds, in upstream's wording (`Installation seems OK. File generation OK` on success). Without a named `dot`, the check SHALL report the built-in engine as `GraphViz: dot - graphviz version <Graphviz version it follows> (gogv <gogv version>)` and lay the test graph out with it. With a named `dot`, the check SHALL report and run that `dot`, as upstream does. The check SHALL exit with status 0 when the test succeeds and non-zero otherwise.

#### Scenario: Working installation

- **WHEN** no `dot` is named and the check runs
- **THEN** the report shows `GraphViz: dot - graphviz version 16.1.0 (gogv <version>)`, with the version of the gogv module without its leading `v` (`gogv 0.2.0` for v0.2.0), and `Installation seems OK. File generation OK`, and the exit status is 0

#### Scenario: Check of a named dot

- **WHEN** `GRAPHVIZ_DOT` names a working Graphviz `dot` and the check runs
- **THEN** the report shows the version that this `dot` prints for `dot -V` and `Installation seems OK. File generation OK`, and the exit status is 0

### Requirement: Layout engine pragmas

The system SHALL accept `!pragma layout smetana`, `!pragma layout elk` and `!pragma layout vizjs` and render with the layout engine of the render (the built-in engine, or the named `dot`) until the corresponding engines are implemented.

#### Scenario: Smetana pragma before the port

- **WHEN** a class diagram contains `!pragma layout smetana`, `class A`, `class B` and `A --> B`, and no `dot` is named
- **THEN** the diagram renders using the built-in engine without error
