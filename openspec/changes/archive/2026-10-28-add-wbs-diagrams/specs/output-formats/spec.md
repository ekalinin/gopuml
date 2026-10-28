# Spec Delta

## MODIFIED Requirements

### Requirement: Text output

The system SHALL render sequence diagrams as ASCII art (TXT, file suffix `.atxt`) and as Unicode box-drawing art (UTXT, file suffix `.utxt`), with the same layout rules as upstream. The system SHALL render gantt diagrams as upstream's table of tasks, drawn in ASCII for both TXT and UTXT, as upstream does. The system SHALL render WBS diagrams as upstream's boxes of the root and of its children, drawn in ASCII for both TXT and UTXT, as upstream does.

#### Scenario: ASCII sequence diagram

- **WHEN** `Alice -> Bob : hello` is rendered as TXT
- **THEN** the output is plain ASCII text that shows the participants `Alice` and `Bob` and an arrow labeled `hello`

#### Scenario: ASCII gantt table

- **WHEN** a gantt diagram with `Project starts 2020-01-06` and `[Build] lasts 3 days` is rendered as TXT
- **THEN** the output is an ASCII table with the header row `Task`, `Start`, `End` and `Duration` and the row `Build`, `Jan 6`, `Jan 8` and `3 days`, as upstream's TXT output

#### Scenario: ASCII WBS diagram

- **WHEN** a WBS diagram with `* Project`, `** Planning`, `*** Scope` and `** Build` is rendered as TXT
- **THEN** the output is ASCII text with the boxes of `Project`, `Planning` and `Build`, each box 4 columns to the right of and 4 lines below the previous one, without `Scope`, as upstream's TXT output
