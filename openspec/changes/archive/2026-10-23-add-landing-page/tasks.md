# Tasks

`TestScenarioCoverage` fails from the proposal until the last `.go.yaml` case of this change exists (task 4.1). Each case is added in `testdata/conformance/cases/landing-page/` with its test, because `TestGoCaseTestsExist` requires the test of every case to exist.

## 1. Page

- [x] 1.1 Write `site/index.html` and `site/style.css` (D1, D2): the presentation, the supported diagrams table of the README, the installation of `v0.1.0` (archives, `checksums.txt`, `go install`), the comparison table with its method and environment from `docs/releases/v0.1.0.md`, the conformance results, the license with links to `LICENSE` and `NOTICE`, the links to the repository, the releases and pkg.go.dev, the `gopuml-release` meta tag, and the playground markup with its `<noscript>` message. Add `TestLandingPage` in the root package and the case `landing-page/presentation.go.yaml` for "landing-page: Visit the page", "landing-page: JavaScript disabled", "landing-page: Supported diagram types" and "landing-page: Installation". Verify that the test passes, and that the page opened from the file with JavaScript disabled shows every section except the playground
- [x] 1.2 Add `TestLandingRelease` in the root package (D2, D7): the meta tag, the installation commands and the README name the same release, its notes exist, and every row of the comparison table of the notes is in the page with the same cells; with mutations of the page (another tag, a changed cell). Add the case `landing-page/release.go.yaml` for "landing-page: Comparison of v0.1.0", "landing-page: Page behind the README" and "landing-page: Comparison out of date". Verify that the test passes and that each mutation makes the check fail

- [x] 1.3 Add a download with `curl` beside the GitHub CLI in the installation of the page (the archive and `checksums.txt` from `https://github.com/ekalinin/gopuml/releases/download/<tag>/`), and check its commands in `TestLandingPage`. Verify that the test passes and fails without them, and that the commands, run against a local copy of the v0.1.0 assets under the same path, download the files that `shasum -c` and `tar` accept

## 2. Site assembly

- [x] 2.1 Add `tools/site` with the `-wasm` mode (D4) and the `site` target of the Makefile (`make wasm`, then the tool with `-wasm dist/wasm` into `dist/site`). Add `TestLocalBuild` and the case `landing-page/preview.go.yaml` for "landing-page: Preview". Verify that the test passes, and that `make site` writes `dist/site/index.html` and `dist/site/wasm/` with `plantuml.wasm`, `wasm_exec.js` and `plantuml.mjs`
- [x] 2.2 Add the `-download` mode (D4): `checksums.txt` and the WASM archive of the release named by the page, the SHA-256 check, the unpacking into `wasm/`, the notice and `deploy=false` on a 404, `deploy=true` otherwise. Add `TestDownloadRelease`, `TestDownloadMismatch` and `TestDownloadUnpublished` with `httptest` servers, and the cases `landing-page/checksum.go.yaml` for "landing-page: Checksum mismatch" and `landing-page/unpublished.go.yaml` for "landing-page: Release not published yet". Verify that the tests pass, that a mismatch and a 404 leave no output, and that a run against `https://github.com/ekalinin/gopuml/releases/download` prints the notice while the repository is private

## 3. Playground

- [x] 3.1 Write `site/playground.mjs` (D6): the controller created with a `load` function (activation, rendering as SVG with only the latest result kept, examples of sequence and new-syntax activity diagrams, the error with its line, the download as `diagram.svg`). Add `cmd/plantuml-wasm/playground_test.go` with `TestPlayground` (real loader: every example renders without error, a sequence source gives SVG, a syntax error on line 3 gives line 3 and an error image, a class diagram gives the graph layout error, the download is an `image/svg+xml` file named `diagram.svg`) and `TestPlaygroundLoading` (fake loaders: no load before activation, a rejected load gives the error state), both skipping without `node`. Add the cases `landing-page/playground.go.yaml` for "landing-page: Sequence diagram", "landing-page: Syntax error", "landing-page: Graph-based diagram", "landing-page: Example diagram" and "landing-page: Download", and `landing-page/loading.go.yaml` for "landing-page: Reading the top of the page" and "landing-page: Module not available". Verify that both tests pass with `node` installed
- [x] 3.2 Write `site/page.mjs` (D6): the lazy loading on visibility or focus, the loading and failure messages, the rendering 300 ms after the last edit, the SVG in an `<img>` from a `blob:` URL, the error image with the message and the line, the example selector, the download link, and the note on the types that render in the browser. Verify by hand in a browser on `make site` served with `python3 -m http.server -d dist/site`: the network panel shows no WASM request before the playground is reached, the result follows the edits, a syntax error shows its line, a class diagram shows the graph layout error, an example replaces the source, the download saves `diagram.svg`, and with `dist/site/wasm/plantuml.wasm` removed the playground shows that it could not be loaded while the rest of the page works

## 4. Deployment

- [x] 4.1 Add `.github/workflows/pages.yml` (D5) and `TestPagesWorkflow` in the root package, with mutations (no `workflow_run`, no success condition, a `pull_request` trigger, no `id-token: write`, a deployment step not conditional on `deploy=true`) that must make the check fail. Add the case `landing-page/deployment.go.yaml` for "landing-page: Change of the page". Verify that the test passes, that each mutation makes the check fail, and that `TestScenarioCoverage` passes

## 5. Documentation

- [x] 5.1 Add a link to the page and its playground near the top of `README.md`. Verify that `TestReleaseNotes`, `TestLandingPage` and `TestLandingRelease` pass

## 6. Checks and publication

- [x] 6.1 Run `make lint`, `make test` (with `node` installed), `make provenance` and `openspec validate add-landing-page --strict`. Verify that all pass and that the playground tests ran, not skipped, in `go test -v ./cmd/plantuml-wasm`
- [x] 6.2 Push the branch, with the user's approval, and open a pull request. Verify that the `test` and `conformance` CI jobs pass
- [x] 6.3 After the merge and after the user has made the repository public and enabled GitHub Pages with "GitHub Actions" as the source, run the Pages workflow by hand, with the user's approval. Verify that it deploys, that `https://ekalinin.github.io/gopuml/` shows the page and that its playground renders a sequence diagram, record whether `plantuml.wasm` is served with `content-encoding: gzip` (design, Open Questions), and set the page as the homepage of the repository with the user's approval
