# Spec Delta

## MODIFIED Requirements

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
