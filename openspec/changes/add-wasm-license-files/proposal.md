# Proposal

## Why

`plantuml.wasm` contains this project's MIT code and the code it translates from PlantUML (MIT), the Go runtime and `golang.org/x/text` (BSD-3-Clause), `regexp2` (MIT), the DejaVu fonts, the bundled themes (MIT and Apache-2.0) and CLDR data (Unicode License v3). These licenses ask for their notices to go with every copy. The CLI archives of a release contain `LICENSE` and `NOTICE`, but the WASM archive contains only `plantuml.wasm`, `wasm_exec.js` and `plantuml.mjs`, and the landing page serves the module without them.

## What Changes

- The WASM archive of a release contains `LICENSE` and `NOTICE`, as the CLI archives do. This applies from the next release; the published v0.1.0 archive does not change.
- The site that `tools/site` assembles has `wasm/LICENSE` and `wasm/NOTICE` next to the module, copied from the repository, both in the deployed site and in the local preview (`make site`). A push to `main` that changes `LICENSE` or `NOTICE` deploys the site again.
- The README says that the WASM archive contains `LICENSE` and `NOTICE`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `release`: the WASM archive of a release contains `LICENSE` and `NOTICE`
- `landing-page`: the site serves `LICENSE` and `NOTICE` next to the WebAssembly module

## Impact

- **Changed files:** `.goreleaser.yaml`, `release_test.go` (`TestReleaseConfig`), `tools/site` and its tests, `.github/workflows/pages.yml` and `TestPagesWorkflow`, `README.md`.
- **Product code:** none.
- **Releases:** the next release's WASM archive has two more files. v0.1.0 stays as published.
- **Site:** the next deployment of the landing page adds two files under `wasm/`.
