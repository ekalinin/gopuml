# Tasks

## 1. Release configuration

`TestScenarioCoverage` fails from the proposal until the last `.go.yaml` case of this change exists (task 4.1). Each case is added with its test, because `TestGoCaseTestsExist` requires the test of every case to exist.

- [ ] 1.1 Add `.goreleaser.yaml` (D1, D2), with the staging directory of `wasm_exec.js` under `.cache/`. Verify that `goreleaser check` passes, and that `goreleaser release --snapshot --clean` produces the six CLI archives with `LICENSE`, `NOTICE` and `README.md`, the WASM archive with its three files, and `checksums.txt`, and that the darwin/arm64 binary prints the snapshot version with `--version`
- [ ] 1.2 Add `TestReleaseConfig` in the root package (D4): it parses `.goreleaser.yaml` and checks the six CLI targets without cgo, the version flag, the archive files, the WASM archive and the checksums. Add the case `release/release-artifacts.go.yaml` for "release: Release of a tag". Verify that the test passes and fails when a target or an archive file is removed
- [ ] 1.3 Add `TestReleaseVersion` in `cmd/plantuml`: it builds the CLI with the version flag of `.goreleaser.yaml` set to `v0.1.0` and checks `--version`. Add the case `release/release-version.go.yaml` for "release: Version of v0.1.0". Verify that the test passes

## 2. Release workflow

- [ ] 2.1 Add `.github/workflows/release.yml` (D3) and `TestReleaseWorkflow` in the root package: tags `v*.*.*` only, `contents: write`, full history, upstream sources, `make lint` and `make test` before the pinned GoReleaser action, and `--release-notes docs/releases/<tag>.md`. Add the case `release/release-gate.go.yaml` for "release: Failing tests". Verify that the test passes and fails when the tests step is removed or moved after GoReleaser

## 3. Comparison with upstream

- [ ] 3.1 Add `tools/bench` and the `bench` target of the Makefile (D6), with a unit test `TestComparisonTable` that runs the tool with fake commands and checks the medians, the table and the environment lines. Add the case `release/comparison.go.yaml` for "release: Comparison table". Verify that the test passes and that `make bench` runs on this machine
- [ ] 3.2 Run `make bench` on this machine with 10 start runs and 5 rendering runs, and keep its output for the notes. Verify that both implementations render every case without error in the runs, or record the cases that fail in both

## 4. Release notes and documentation

- [ ] 4.1 Write `docs/releases/v0.1.0.md` (D5) with the output of 3.2 and the conformance results of a `make conformance REQUIRE=all` run, and add `TestReleaseNotes` in the root package. Add the case `release/release-notes.go.yaml` for "release: Notes of v0.1.0". Verify that the test passes, and that `TestScenarioCoverage` and `make test` pass
- [ ] 4.2 Add an installation section to `README.md`: the archives of a release, the checksums, `go install github.com/ekalinin/gopuml/cmd/plantuml@v0.1.0` (with `GOPRIVATE` while the repository is private), and the quarantine attribute on macOS. Verify that the commands of the section work with the snapshot archives of 1.1

## 5. Publication

- [ ] 5.1 Run `make lint`, `make test`, `make provenance`, `goreleaser check` and `openspec validate add-release --strict`. Verify that all pass
- [ ] 5.2 Push the branch, with the user's approval, and open a pull request. Verify that the `test` and `conformance` CI jobs pass
- [ ] 5.3 After the merge, create the tag `v0.1.0` on `main` and push it, with the user's approval (D7). Verify that the release workflow passes, that the release has the six CLI archives, the WASM archive, `checksums.txt` and the notes of `docs/releases/v0.1.0.md`, that the checksums match the downloaded archives, and that a downloaded binary prints `gopuml v0.1.0 (PlantUML 1.2026.8 compatible)`
