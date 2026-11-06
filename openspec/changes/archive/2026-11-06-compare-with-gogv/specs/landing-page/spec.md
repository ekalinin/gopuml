# Spec Delta

## MODIFIED Requirements

### Requirement: Comparison with upstream on the page

The page SHALL show the comparison of gopuml with upstream PlantUML from the notes of the release it names: every column of its table, each with its own participant and bar color, the start time and the rendering time of each supported diagram type in each column, how they are measured, the environment of the run as the notes list it (the machine, the versions and the number of runs), and that the numbers depend on the machine.

#### Scenario: Comparison of v0.1.0

- **WHEN** the page names the release `v0.1.0`
- **THEN** its comparison table has the values of the table in `docs/releases/v0.1.0.md`, with the environment of that run

#### Scenario: Comparison with three columns

- **WHEN** the notes of the release the page names have a comparison table with the columns `PlantUML 1.2026.9 (Java)`, `gopuml (dot)` and `gopuml (gogv)`
- **THEN** the page shows the three columns with the values of that table, and the tests of the repository fail when one of the columns is missing on the page
