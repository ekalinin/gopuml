# Proposal

## Why

The implementation supports six diagram types (sequence, class and object, use case, component and deployment, state, and activity in both syntaxes) with 0 conformance failures, but there is no release: users have to clone the repository and build it with Go 1.26.

## What Changes

- Version v0.1.0, the first release.
- A GitHub Actions workflow that runs on a pushed `v*` tag: it runs the lint and the tests, then GoReleaser publishes a GitHub Release with:
  - the CLI (`plantuml`) for linux, macOS and windows on amd64 and arm64, built without cgo, with `LICENSE`, `NOTICE` and `README.md` in each archive
  - the WASM bundle (`plantuml.wasm`, `wasm_exec.js`, `plantuml.mjs`), as `make wasm` produces it
  - a checksums file for every archive
  - the release notes of the tag
- The version of a release build is the tag: `plantuml --version` prints `gopuml v0.1.0 (PlantUML 1.2026.8 compatible)`.
- A `.goreleaser.yaml` checked with `goreleaser check`, and a snapshot build that produces the same archives locally without publishing.
- A reproducible comparison of this implementation with the upstream jar, on the same machine and with the same `dot`: the start time (one process rendering a small diagram) and the rendering time (one process rendering the curated cases of each supported type). Its results go into the release notes.
- Release notes for v0.1.0 in `docs/releases/v0.1.0.md`: the supported diagram types and formats, what is not supported yet, the Graphviz requirement, the comparison, and the conformance results.
- The README explains how to install a release.

Out of scope:

- a landing page
- packages for Homebrew, Scoop, Linux distributions or container registries
- signing or notarizing the binaries

## Capabilities

### New Capabilities

- `release`: tagged releases with prebuilt CLI and WASM artifacts, the version of a release build, the release gate, release notes, and the comparison with upstream

### Modified Capabilities

None.

## Impact

- **New files:** `.goreleaser.yaml`, `.github/workflows/release.yml`, `docs/releases/v0.1.0.md`, and a comparison tool under `tools/`.
- **Product code:** none. The version stamping already exists (`internal/version.Version`, set with `-ldflags -X`).
- **CI:** the release workflow runs only on tags. The `test` and `conformance` jobs of `ci.yml` do not change.
- **Publishing:** pushing the `v0.1.0` tag publishes the release. The repository is private today, so the release is visible only to people with access to it until the repository is made public.
- **Dependencies:** GoReleaser runs in CI through its GitHub Action; it is not a module dependency.
