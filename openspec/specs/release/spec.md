# release Specification

## Purpose

Publishes tagged releases of gopuml with prebuilt CLI and WASM artifacts, release notes, and a reproducible comparison with upstream PlantUML.

## Requirements

### Requirement: Release artifacts

Pushing a tag of the form `vMAJOR.MINOR.PATCH` SHALL publish a GitHub Release of that tag with an archive of the CLI (`plantuml`) for each of linux, macOS and windows on amd64 and arm64, built without cgo, each archive containing `LICENSE`, `NOTICE` and `README.md`; an archive of the WASM bundle (`plantuml.wasm`, `wasm_exec.js` and `plantuml.mjs`); and a checksums file with the SHA-256 of every archive. The same archives SHALL be produced by a local snapshot build that publishes nothing.

#### Scenario: Release of a tag

- **WHEN** the tag `v0.1.0` is pushed
- **THEN** a GitHub Release `v0.1.0` is published with six CLI archives (linux, macOS and windows, each on amd64 and arm64), the WASM archive and the checksums file

### Requirement: Version of a release build

The CLI of a release build SHALL report the tag of the release as its version.

#### Scenario: Version of v0.1.0

- **WHEN** the CLI built for the release `v0.1.0` runs with `--version`
- **THEN** it prints `gopuml v0.1.0 (PlantUML 1.2026.8 compatible)`

### Requirement: Release gate

The release workflow SHALL run on pushed `v*` tags only, and SHALL run the lint and the tests of the repository before publishing. When they fail, nothing SHALL be published.

#### Scenario: Failing tests

- **WHEN** a `v*` tag is pushed on a commit whose tests fail
- **THEN** the workflow fails before publishing, and no release is created

### Requirement: Release notes

Each release SHALL have release notes in `docs/releases/<tag>.md`, used as the description of the GitHub Release. They SHALL list the supported diagram types and output formats, the types that are not supported yet, the Graphviz requirement, the conformance results, and the comparison with upstream.

#### Scenario: Notes of v0.1.0

- **WHEN** the release `v0.1.0` is published
- **THEN** its description is `docs/releases/v0.1.0.md`, which lists the supported diagram types with their formats, the conformance results and the comparison with upstream

### Requirement: Comparison with upstream

A command SHALL compare this implementation with the upstream jar on the same machine, with the same `dot` for both: the start time, as the median time of separate processes that each render a small sequence diagram, and the rendering time, as the median time of one process that renders the curated cases of each supported diagram type. It SHALL print the results as a table, with the machine, the Java, Go and Graphviz versions, and the number of runs.

#### Scenario: Comparison table

- **WHEN** the comparison runs with the CLI of this implementation and the upstream jar
- **THEN** it prints a table with the start time and the rendering time of each supported diagram type for both, and the environment of the run
