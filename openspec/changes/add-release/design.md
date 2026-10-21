# Design

## Context

See proposal.md for the motivation. The current state that shapes the approach:

- The module has no cgo and one dependency (`go.yaml.in/yaml/v3`), so the CLI cross-compiles with `CGO_ENABLED=0`.
- `internal/version.Version` is set with `-ldflags -X`; without it, `Semver` takes the module version from the build information (`go install ...@v0.1.0` gives `v0.1.0`), else `devel`. `--version` prints `gopuml <version> (PlantUML 1.2026.8 compatible)`.
- `make wasm` builds `plantuml.wasm` with `-trimpath` and copies `wasm_exec.js` from the Go root and `cmd/plantuml-wasm/plantuml.mjs` into `WASM_OUT` (`dist/wasm` by default, which GoReleaser also uses for its output).
- `.github/workflows/ci.yml` has a `test` job (upstream sources, lint, test, provenance, WASM build) and a `conformance` job, on pull requests and pushes to `main`. `internal/conformance/ci_test.go` checks the contract of that workflow by parsing it.
- The repository is private; `LICENSE` and `NOTICE` are at its root.
- GoReleaser 2.16 is installed on the development machine, not in CI images.
- The upstream jar is in `.cache/upstream/plantuml-mit-1.2026.8.jar` after `tools/upstream/fetch.sh`; both implementations look `dot` up through `GRAPHVIZ_DOT`.
- `TestScenarioCoverage` requires a case for every scenario of the active changes, and `TestGoCaseTestsExist` requires the test of every `.go.yaml` case to exist.

## Goals / Non-Goals

**Goals:**

- Publish v0.1.0 from a tag with nothing built by hand.
- Keep the release reproducible locally (`goreleaser check`, snapshot build) before the tag is pushed.
- A comparison with upstream that anyone with the jar, Java and Graphviz can run again.

**Non-Goals:**

- Package managers, container images, signing and notarization (see the proposal).
- Running the conformance suite in the release workflow (see D3).

## Decisions

### D1. GoReleaser configuration

`.goreleaser.yaml` (GoReleaser v2):

- A `cli` build of `./cmd/plantuml`, binary `plantuml`, `CGO_ENABLED=0`, `-trimpath`, `-s -w`, for linux, darwin and windows on amd64 and arm64.
- A `wasm` build of `./cmd/plantuml-wasm` for `js/wasm`, binary `plantuml.wasm`, with `-trimpath`, as `make wasm`.
- A `before` hook copies `wasm_exec.js` from the Go root that builds the release into a staging directory under `.cache/`, so that it matches the compiler of `plantuml.wasm`.
- Archives `gopuml_<version>_<os>_<arch>` (`tar.gz`, `zip` on windows) with the binary, `LICENSE`, `NOTICE` and `README.md`, and `gopuml_<version>_wasm.tar.gz` with `plantuml.wasm`, `wasm_exec.js` and `plantuml.mjs`.
- `checksums.txt` with the SHA-256 of every archive.
- No generated changelog: the description of the release is the notes file (D5).

The archives keep the binary name `plantuml`, which the README already uses, and name the project `gopuml`, the module name.

Alternative: building the WASM bundle with `make wasm` in a hook and uploading its files as extra release files. They would be three loose files outside the checksums of the archives.

### D2. Version stamping

The `cli` build sets `-X github.com/ekalinin/gopuml/internal/version.Version={{ .Tag }}`. The tag keeps its `v`, as the module versions that `Semver` reads otherwise, so `--version` prints the same thing for a release archive and for `go install ...@v0.1.0`. No product code changes.

### D3. Release workflow

`.github/workflows/release.yml` runs on pushed tags matching `v*.*.*`, with `contents: write`. It checks out the full history (GoReleaser needs the tag), sets up Go from `go.mod`, fetches the upstream sources as the `test` job does, runs `make lint` and `make test`, and then runs GoReleaser through its pinned GitHub Action with `release --clean --release-notes docs/releases/<tag>.md`. A failing step stops the job before GoReleaser, so nothing is published.

The conformance suite does not run there: it needs the reference image and takes minutes, and a tag is pushed on a commit of `main`, where the `conformance` job of `ci.yml` already ran when the commit was pushed.

Alternative: making the release workflow depend on a successful CI run of the tagged commit through `workflow_run`. It couples two workflows for no stronger guarantee than the tag being on `main`.

### D4. Checked configuration

As `ci_test.go` does for the CI workflow, root-package tests parse `.goreleaser.yaml` and `release.yml` and check their contract: the six CLI targets without cgo, the version flag, the archive contents, the WASM archive, the checksums, the tag trigger, the lint and tests before GoReleaser, and the release notes argument. A test in `cmd/plantuml` builds the CLI with the version flag and checks `--version`. `goreleaser check` and a snapshot build are run by hand before tagging, because GoReleaser is not installed in the `test` job.

### D5. Release notes

`docs/releases/<tag>.md` holds the notes of each release, written by hand. The notes of v0.1.0 list the supported diagram types and formats from the README table, the types not supported yet, the Graphviz requirement, the conformance results (cases, failures, known differences), the comparison (D6), the installation, and the quarantine attribute of unsigned macOS binaries. A root-package test checks that the notes of each release named in the repository exist and contain these sections.

### D6. Comparison tool

`tools/bench` is an original Go command, run by `make bench`:

- It takes the CLI binary, the jar, the `dot` to use (passed to both through `GRAPHVIZ_DOT`) and the number of runs.
- Start time: separate processes that each render a small sequence diagram with `-tsvg -pipe`; the median of the runs.
- Rendering time: for each supported type, the bodies (from `@startuml`) of its curated cases are copied to a temporary directory, and one process renders them all with `-tsvg -o <dir>`; the median of the runs, with the time per diagram.
- It prints a Markdown table, and the operating system, CPU, Java, Go and Graphviz versions and the number of runs.

The front matter of the cases is removed so that both implementations read the same diagrams. The unit test runs the tool with fake commands, so that it does not depend on Java or Graphviz.

Alternative: `hyperfine`. It measures separate processes well, but it is not installed here or in CI, and the rendering time per type needs the case selection anyway.

### D7. Tag and publication

The tag `v0.1.0` is created on `main` once the pull request of this change is merged, and pushed only with the user's approval, because pushing it publishes the release.

## Risks / Trade-offs

- **[Private repository]** The release is visible only to people with access to the repository. → Stated in the notes of the change; making the repository public is the user's decision.
- **[Unsigned macOS binaries]** Gatekeeper blocks binaries downloaded with a browser. → The notes explain `xattr -d com.apple.quarantine plantuml`.
- **[Comparison numbers]** They depend on the machine, and the JVM warms up over the cases of one process. → The tool prints the environment and the number of runs, and the notes describe the method.
- **[GoReleaser changes]** A new GoReleaser version could change the archives. → The GitHub Action pins the major version, and D4 checks the configuration.
- **[A wrong tag]** A tag on a commit with failing tests fails the workflow before publishing. A published release can be deleted with its tag by hand.

## Migration Plan

Nothing to migrate. After the merge, the tag is pushed with the user's approval, the workflow publishes the release, and the archives are checked: checksums, and `--version` of a downloaded binary.
