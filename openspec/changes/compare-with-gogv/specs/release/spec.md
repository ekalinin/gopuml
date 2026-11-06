# Spec Delta

## MODIFIED Requirements

### Requirement: Comparison with upstream

A command SHALL compare this implementation with the upstream jar on the same machine, in three columns: the upstream jar with a Graphviz `dot`, this implementation with the same `dot`, and this implementation with its built-in layout (gogv), run with no `dot` named. For each column, it SHALL measure the start time, as the median time of separate processes that each render a small sequence diagram, and the rendering time, as the median time of one process that renders the curated cases of each supported diagram type. It SHALL print the results as a table with one column for each, with the machine, the Java, Go and Graphviz versions, the gogv version of the command line under test, and the number of runs.

#### Scenario: Comparison table

- **WHEN** the comparison runs with the CLI of this implementation and the upstream jar
- **THEN** it prints a table whose columns are `PlantUML 1.2026.9 (Java)`, `gopuml (dot)` and `gopuml (gogv)`, with the start time and the rendering time of each supported diagram type in each column, and the environment of the run, which names the gogv version of the CLI

#### Scenario: Layout of each column

- **WHEN** the comparison runs with `-dot` naming a Graphviz `dot`, while `GRAPHVIZ_DOT` names another `dot` in its environment
- **THEN** the upstream jar and the CLI of the `gopuml (dot)` column run with `GRAPHVIZ_DOT` set to the `dot` of `-dot`, and the CLI of the `gopuml (gogv)` column runs without `GRAPHVIZ_DOT`
