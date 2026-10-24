# Spec Delta

## MODIFIED Requirements

### Requirement: Version of a release build

The CLI of a release build SHALL report the tag of the release as its version.

#### Scenario: Version of v0.1.0

- **WHEN** the CLI built with the tag `v0.1.0` runs with `--version`
- **THEN** it prints `gopuml v0.1.0 (PlantUML 1.2026.9 compatible)`
