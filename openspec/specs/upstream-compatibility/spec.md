# upstream-compatibility Specification

## Purpose

Defines what "compatible with PlantUML" means for this implementation: the pinned upstream version, the reference renders, the conformance corpus, the pass criteria, and when a diagram type may be declared supported.

## Requirements

### Requirement: Pinned upstream reference

Compatibility SHALL be measured against PlantUML v1.2026.9 in its MIT variant, run in a pinned container image with a pinned Java runtime, Graphviz version and font set. References SHALL be generated with upstream's time-based error-image decorations disabled, and SHALL be regenerated only when the pinned versions change.

#### Scenario: Reproducible references

- **WHEN** the reference renders are generated twice from the same corpus with the same container image
- **THEN** both runs produce identical reference files

### Requirement: Conformance corpus

The conformance corpus SHALL contain the `.puml` inputs from upstream's Vega test suite at v1.2026.9, with the zip archives these inputs import, plus curated cases for every requirement scenario of the diagram type, preprocessor, styling and output-format capabilities. Each case SHALL record its diagram family and type.

#### Scenario: Corpus coverage

- **WHEN** the conformance suite runs
- **THEN** every corpus case is either executed or reported as belonging to a diagram type that is not supported yet

### Requirement: Visual parity criteria

A corpus case SHALL pass when all of the following hold against its reference render:

- Both implementations agree on success or failure. For failures, they report the same error line.
- The SVG output contains the same text strings as the reference, in any order. Version strings of the two implementations are excluded.
- The output width and height are each within 5% of the reference.
- The rasterized output has a structural similarity (SSIM, grayscale) of at least 0.90 against the rasterized reference at the reference size.

The thresholds SHALL only be tightened over time, never loosened.

#### Scenario: Passing case

- **WHEN** a sequence diagram case renders with the same texts, a size within 5% and SSIM 0.96 against its reference
- **THEN** the case passes

#### Scenario: Missing text fails

- **WHEN** a case's SVG output lacks a message label that the reference contains
- **THEN** the case fails regardless of its similarity score

### Requirement: Text output parity

For TXT, UTXT and PREPROC outputs, a case SHALL pass only when the output is identical to the reference text, apart from trailing whitespace at line ends.

#### Scenario: ASCII art matches

- **WHEN** a sequence diagram case is rendered as TXT
- **THEN** the output equals the upstream TXT reference line by line, ignoring trailing spaces

### Requirement: CLI and server behavior parity

The conformance suite SHALL also run CLI cases (exit codes, output file names, pipe mode, syntax check output) and HTTP cases (status codes, headers, endpoint bodies) against the upstream CLI and built-in server in the reference container, and against the pinned plantuml-server image for routes only plantuml-server provides. It SHALL compare observable results for the flags and endpoints this change supports.

#### Scenario: Exit code parity

- **WHEN** the CLI case for a missing input file runs against both implementations
- **THEN** both exit with status 50

### Requirement: Declaring a diagram type supported

A diagram type, output format or engine SHALL be declared supported only when every one of its corpus cases passes or is listed in a known-differences file with a written justification. The conformance report SHALL show the pass rate per diagram type and format.

#### Scenario: Known difference

- **WHEN** a class case that uses `!pragma layout smetana` fails because the Smetana layout is not ported, and it is listed in the known-differences file with that reason
- **THEN** the class type can still be declared supported, and the report lists the case as a known difference

### Requirement: Continuous conformance

The conformance suite SHALL run on linux/amd64 for every pull request and every push to the main branch. In that run:

- The outputs of this implementation SHALL be produced in the pinned reference container, so that both implementations use the same Graphviz version.
- Every conformance group SHALL be required: a failing case that is not listed in the known-differences file SHALL fail the run.
- When the run renders references, it SHALL fail if the committed references differ from the new render.
- The run SHALL publish the conformance report, whether it passes or fails.

#### Scenario: Regression fails the run

- **WHEN** a pull request changes the implementation so that a sequence case no longer matches its reference
- **THEN** the conformance run of the pull request fails and names the case

#### Scenario: Drawing an unsupported type fails the run

- **WHEN** a change makes a diagram type that is not supported yet produce a drawing instead of the "not supported yet" error
- **THEN** the conformance run fails

#### Scenario: Known differences do not fail the run

- **WHEN** the only failing cases are listed in the known-differences file
- **THEN** the conformance run passes, and the report lists those cases as known differences

#### Scenario: Stale committed reference fails the run

- **WHEN** the run renders the references and a committed curated reference differs from the new render
- **THEN** the conformance run fails and names the differing reference files

#### Scenario: Report published on failure

- **WHEN** the conformance run fails
- **THEN** the conformance report of that run is still available from the run
