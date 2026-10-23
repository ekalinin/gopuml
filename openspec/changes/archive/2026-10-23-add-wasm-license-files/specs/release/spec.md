# Spec Delta

## MODIFIED Requirements

### Requirement: Release artifacts

Pushing a tag of the form `vMAJOR.MINOR.PATCH` SHALL publish a GitHub Release of that tag with an archive of the CLI (`plantuml`) for each of linux, macOS and windows on amd64 and arm64, built without cgo, each archive containing `LICENSE`, `NOTICE` and `README.md`; an archive of the WASM bundle (`plantuml.wasm`, `wasm_exec.js` and `plantuml.mjs`) containing `LICENSE` and `NOTICE`; and a checksums file with the SHA-256 of every archive. The same archives SHALL be produced by a local snapshot build that publishes nothing.

#### Scenario: Release of a tag

- **WHEN** the tag `v0.1.0` is pushed
- **THEN** a GitHub Release `v0.1.0` is published with six CLI archives (linux, macOS and windows, each on amd64 and arm64), the WASM archive and the checksums file

#### Scenario: License files of the WASM archive

- **WHEN** the archives of a release are built
- **THEN** the WASM archive contains `plantuml.wasm`, `wasm_exec.js`, `plantuml.mjs`, `LICENSE` and `NOTICE`
