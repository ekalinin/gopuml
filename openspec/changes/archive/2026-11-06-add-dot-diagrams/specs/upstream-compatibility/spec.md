# Spec Delta

## MODIFIED Requirements

### Requirement: Conformance corpus

The conformance corpus SHALL contain the `.puml` inputs from upstream's Vega test suite at v1.2026.9, with the zip archives these inputs import, plus curated cases for every requirement scenario of the diagram type, preprocessor, styling and output-format capabilities. Each case SHALL record its diagram family and type. A curated case MAY name a security profile with one of upstream's profile names; upstream renders its reference and this implementation renders the case in that profile. A case that names no profile SHALL be rendered in upstream's default profile, `LEGACY`.

#### Scenario: Corpus coverage

- **WHEN** the conformance suite runs
- **THEN** every corpus case is either executed or reported as belonging to a diagram type that is not supported yet

#### Scenario: Case rendered in a security profile

- **WHEN** a curated case names the profile `INTERNET_WITH_DOTSVG` and the references and the outputs of this implementation are rendered
- **THEN** upstream renders the reference of the case with `PLANTUML_SECURITY_PROFILE` set to `INTERNET_WITH_DOTSVG`, this implementation renders the case in that profile, and the cases that name no profile are rendered in `LEGACY` by both
