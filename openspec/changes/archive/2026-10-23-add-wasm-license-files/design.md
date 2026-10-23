# Design

## Context

See proposal.md for the motivation. The current state:

- `.goreleaser.yaml` has two archives. `cli` lists `LICENSE`, `NOTICE` and `README.md` in its `files`; `wasm` lists only the staged `wasm_exec.js` and `cmd/plantuml-wasm/plantuml.mjs` (with `strip_parent`), besides the binary `plantuml.wasm`.
- `TestReleaseConfig` (`release_test.go`) parses `.goreleaser.yaml` and requires exactly these two files in the WASM archive, with mutations that must make the check fail.
- `tools/site` copies `site/` into the output directory and writes the three module files into `<out>/wasm/`, from a local build (`-wasm`) or from the WASM archive of the release the page names (`-download`). It writes nothing when a check fails.
- `.github/workflows/pages.yml` deploys on a push to `main` that changes `site/**`, `tools/site/**` or the workflow; `TestPagesWorkflow` checks these paths.
- The published v0.1.0 WASM archive has no `LICENSE` or `NOTICE`, and the page already links to `LICENSE` and `NOTICE` on `main`.

## Goals / Non-Goals

**Goals:**

- Every copy of the module that the project distributes, in a release archive or on the site, comes with `LICENSE` and `NOTICE`.

**Non-Goals:**

- Republishing v0.1.0: its archive and checksums stay as published.
- `README.md` in the WASM archive: the module has no command line to document.

## Decisions

### D1. Archive files

The `wasm` archive lists `LICENSE` and `NOTICE` as plain paths, as the `cli` archive does, so they are at the root of the archive next to the module. `TestReleaseConfig` requires the staged `wasm_exec.js`, `plantuml.mjs`, `LICENSE` and `NOTICE`, and gets two more mutations: the WASM archive without `LICENSE`, and without `NOTICE`.

### D2. Source of the license files on the site

`tools/site` copies `LICENSE` and `NOTICE` from the repository into `<out>/wasm/`, in both modes, with a `-root` flag (default `.`, the repository root from which `make site` and the workflow run). They are read before anything is written, so a missing file fails the run with no output, as the other checks do.

The repository is the source rather than the release archive: the v0.1.0 archive has no license files, the page already links to the files on `main`, and the deployed site comes from `main`.

Alternative: take them from the archive of the release when it has them, and from the repository otherwise. That gives the same file two sources, and v0.1.0 needs the repository anyway.

### D3. Redeployment when the notices change

The push paths of `pages.yml` include `LICENSE` and `NOTICE`, so the site follows a change of either on `main`. `TestPagesWorkflow` requires them, with a mutation that removes `NOTICE`.

### D4. Tests and cases

- `tools/site`: `TestLicenseFiles` assembles a site in each mode from a temporary repository root and checks that `wasm/LICENSE` and `wasm/NOTICE` equal its files, and that a root without `NOTICE` writes nothing.
- The case `release/release-artifacts.go.yaml` (`TestReleaseConfig`) also covers "release: License files of the WASM archive".
- New cases in `testdata/conformance/cases/landing-page/`: `license-files.go.yaml` (`TestLicenseFiles`) for "landing-page: License files on the site", and the case of `TestPagesWorkflow` also covers "landing-page: Change of the notices".

## Risks / Trade-offs

- **[NOTICE of the site and of the release can differ]** The site serves the module of the release with the `NOTICE` of `main`. → `NOTICE` only grows with the project (it gained the Go code section after v0.1.0); a release with new licensed parts updates it in the same pull request, so the `NOTICE` of `main` covers the module of the latest release.
- **[GoReleaser behavior]** A file listed twice or a path pattern could change the archive layout. → A snapshot build is checked by hand: the WASM archive lists exactly the five files at its root.
