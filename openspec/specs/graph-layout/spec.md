# graph-layout Specification

## Purpose

Lays out graph-based diagrams (class diagrams in this change) with the external Graphviz `dot` program, as PlantUML does by default, including how `dot` is found, timeouts and failure reporting.

## Requirements

### Requirement: Graphviz dot discovery

The system SHALL locate the `dot` executable in this order, as upstream does:

1. an explicit path from the caller (the `--dot-path` CLI flag or the library option)
2. the `GRAPHVIZ_DOT` environment variable
3. the directories in `PATH`
4. the standard locations `/usr/local/bin/dot`, `/opt/homebrew/bin/dot`, `/opt/homebrew/opt/graphviz/bin/dot`, `/usr/bin/dot` and `/opt/local/bin/dot` (`dot.exe` in the usual install directories on Windows)

When `GRAPHVIZ_DOT` is set, that path SHALL be used even if no file exists there, as upstream does.

#### Scenario: Environment variable wins over PATH

- **WHEN** `GRAPHVIZ_DOT` points to `/custom/dot` and another `dot` is on `PATH`
- **THEN** the system runs `/custom/dot`

#### Scenario: Found on PATH

- **WHEN** no explicit path and no `GRAPHVIZ_DOT` are set and `dot` is on `PATH`
- **THEN** the system runs the `dot` found on `PATH`

#### Scenario: Environment variable points to a missing file

- **WHEN** `GRAPHVIZ_DOT` points to `/missing/dot` and another `dot` is on `PATH`
- **THEN** the system treats `dot` as not found and does not run the one on `PATH`

### Requirement: Graph layout through dot

For graph-based diagrams, the system SHALL measure every node and label itself, have `dot` compute node positions and edge routes, and draw the result itself. For the same input and Graphviz version, the arrangement of nodes and edges SHALL match upstream's arrangement within the parity tolerance defined by upstream-compatibility.

#### Scenario: Class diagram layout

- **WHEN** a class diagram with classes `A`, `B` and the relation `A --> B` is rendered
- **THEN** `B` is placed below `A` and the arrow runs from `A` to `B`, as in upstream's output

### Requirement: Missing dot

When `dot` cannot be found, rendering a graph-based diagram SHALL produce an error diagram stating that Graphviz `dot` was not found and listing the ways to configure it. Diagram types that do not need `dot` (sequence, activity) SHALL render normally.

#### Scenario: Class diagram without Graphviz

- **WHEN** no `dot` executable can be found and a class diagram is rendered
- **THEN** the system reports an error stating that Graphviz `dot` was not found, and produces an error image

#### Scenario: Sequence diagram without Graphviz

- **WHEN** no `dot` executable can be found and a sequence diagram is rendered
- **THEN** the diagram renders without error

### Requirement: dot timeout

The system SHALL stop a `dot` process that runs longer than the configured timeout and report a timeout error for that diagram. The default timeout SHALL be 30 seconds, as in upstream, and SHALL be configurable in seconds (the `--graphviz-timeout` CLI flag or the library option).

#### Scenario: Timeout exceeded

- **WHEN** the timeout is set to 1 second and `dot` does not finish within 1 second
- **THEN** the `dot` process is terminated and the diagram reports a Graphviz timeout error

### Requirement: dot failure

When `dot` exits with an error or returns output the system cannot use, the system SHALL report an error for that diagram that includes `dot`'s error message, and SHALL continue with other diagrams.

#### Scenario: dot crashes on one diagram

- **WHEN** `dot` fails on the first of two class diagrams in one input
- **THEN** the first diagram reports a Graphviz error and the second diagram still renders

### Requirement: Graphviz check

The system SHALL provide a check that prints the Graphviz version and whether a test layout succeeds, in upstream's wording (`Installation seems OK. File generation OK` on success). The check SHALL exit with status 0 when the test succeeds and non-zero otherwise.

#### Scenario: Working installation

- **WHEN** a working `dot` is installed and the check runs
- **THEN** the report shows the Graphviz version and `Installation seems OK. File generation OK`, and the exit status is 0

### Requirement: Layout engine pragmas

The system SHALL accept `!pragma layout smetana`, `!pragma layout elk` and `!pragma layout vizjs` and render with `dot` until the corresponding engines are implemented.

#### Scenario: Smetana pragma before the port

- **WHEN** a class diagram contains `!pragma layout smetana` and `dot` is available
- **THEN** the diagram renders using `dot` without error
