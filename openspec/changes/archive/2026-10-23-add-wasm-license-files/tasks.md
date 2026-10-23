# Tasks

`TestScenarioCoverage` fails from the proposal until the cases of the three new scenarios exist (task 2.2). Each case is added with its test.

## 1. Release archive

- [x] 1.1 List `LICENSE` and `NOTICE` in the `wasm` archive of `.goreleaser.yaml` (D1), make `TestReleaseConfig` require them with two mutations (the WASM archive without `LICENSE`, without `NOTICE`), and add "release: License files of the WASM archive" to the case `release/release-artifacts.go.yaml`. Verify that the test passes and fails for each mutation, that `goreleaser check` passes, and that the WASM archive of `goreleaser release --snapshot --clean` lists exactly `plantuml.wasm`, `wasm_exec.js`, `plantuml.mjs`, `LICENSE` and `NOTICE` at its root

## 2. Site

- [x] 2.1 Add the `-root` flag to `tools/site` and copy `LICENSE` and `NOTICE` from it into `<out>/wasm/` in both modes, reading them before anything is written (D2). Add `TestLicenseFiles` and the case `landing-page/license-files.go.yaml` for "landing-page: License files on the site". Verify that the tests of `tools/site` pass, that a root without `NOTICE` writes nothing, and that `make site` writes `dist/site/wasm/LICENSE` and `dist/site/wasm/NOTICE` equal to the files of the repository
- [x] 2.2 Add `LICENSE` and `NOTICE` to the push paths of `.github/workflows/pages.yml` (D3), make `TestPagesWorkflow` require them with a mutation that removes `NOTICE`, and add "landing-page: Change of the notices" to the case `landing-page/deployment.go.yaml`. Verify that the test passes and fails for the mutation, that `actionlint` reports nothing, and that `TestScenarioCoverage` passes

## 3. Documentation

- [x] 3.1 Say in the installation section of `README.md` that the WASM archive contains `LICENSE` and `NOTICE`. Verify that `TestReleaseNotes`, `TestLandingPage` and `TestLandingRelease` pass

## 4. Checks and publication

- [x] 4.1 Run `make lint`, `make test`, `make provenance` and `openspec validate add-wasm-license-files --strict`. Verify that all pass
- [x] 4.2 Push the branch, with the user's approval, and open a pull request. Verify that the `test` and `conformance` CI jobs pass
- [x] 4.3 After the merge, verify that the `pages` workflow deploys the site again and that `https://ekalinin.github.io/gopuml/wasm/LICENSE` and `https://ekalinin.github.io/gopuml/wasm/NOTICE` equal the files on `main`
