# Design

## Context

See proposal.md for the motivation. The current state that shapes the approach:

- The release v0.1.0 has `gopuml_0.1.0_wasm.tar.gz` (`plantuml.wasm`, `wasm_exec.js`, `plantuml.mjs`, about 10 MB) and `checksums.txt`. `plantuml.wasm` is about 36 MB uncompressed.
- `plantuml.mjs` exports `load(wasmURL)`, which resolves to `{ render(source, options) }`. `render` returns `data` (a `Uint8Array`), `width`, `height`, `description`, `pages`, `diagrams` and `error` (`null`, or `{ kind, message, line }`); for an error, `data` is the error image. The module runs in the `SANDBOX` profile, and graph-based types return the "graph layout is not available" error.
- `make wasm` builds the same three files into `dist/wasm` from the working tree. `cmd/plantuml-wasm/wasm_test.go` builds the module and runs it with Node.js, and skips when `node` is not installed.
- The README names the release in its installation section (`gh release download v0.1.0`, `plantuml@v0.1.0`). `TestReleaseNotes` collects these tags and checks that each has `docs/releases/<tag>.md`. `docs/releases/v0.1.0.md` has the `make bench` table, its method and its environment.
- `release_test.go` checks the release workflow by parsing it and by mutating it: each mutation (no tests, wrong trigger, wrong permission, ...) must make the check fail.
- `TestScenarioCoverage` requires a case for every scenario of the active changes, and `TestGoCaseTestsExist` requires the test that every `.go.yaml` case names.
- The release is created by GoReleaser with `GITHUB_TOKEN`. Events caused by `GITHUB_TOKEN` do not start other workflows, except `workflow_dispatch` and `repository_dispatch`, so a `release: published` trigger would not run after a release.
- On the free plan, GitHub Pages is available only for public repositories, and release assets of a public repository download from `https://github.com/ekalinin/gopuml/releases/download/<tag>/<file>` without a token.

## Goals / Non-Goals

**Goals:**

- One release on the page, the same in its text, its comparison and its playground, kept consistent with the README and the notes by tests.
- No build step for the page itself: the files in `site/` are the published files.
- The playground runs exactly the module users download from the release.

**Non-Goals:**

- Third-party assets or services on the page: no web fonts, CDNs, analytics or trackers.
- Running the lint and the tests in the Pages workflow: the files it deploys come from `main`, which the CI checks on every pull request.

## Decisions

### D1. Files of the page

`site/` holds the published files:

- `index.html`: every section of the page in static HTML (presentation, supported types, installation, comparison, conformance, license, playground). The playground markup has a `<noscript>` message.
- `style.css`: the styles, with the system font stack and light and dark color schemes (`prefers-color-scheme`).
- `playground.mjs`: the playground logic without the DOM (loading, rendering, examples, download), so that Node.js can test it (D6).
- `page.mjs`: the DOM glue of the playground (editor, example selector, result, error, download link, lazy loading).

The JavaScript and CSS files start with the `SPDX-License-Identifier: MIT` and `Provenance: original` comments of the project. The page links to `LICENSE` and `NOTICE` in the repository.

Alternative: a site generator (Hugo, Jekyll). It would add a tool and a build step for one page.

### D2. One release named by the page

`index.html` names its release once in `<meta name="gopuml-release" content="v0.1.0">`, and in the text of its installation commands and comparison. Root-package tests (D7) check that the meta tag, the installation commands and the README name the same tag, that `docs/releases/<tag>.md` exists, and that every row of the comparison table of the notes is in the table of the page with the same cells. Every new release updates the README, the notes and the page in the same pull request, before the tag is pushed.

Alternative: generate the page from the README and the notes. It needs a template step, and the files in `site/` would no longer be the published files; the tests give the same guarantee.

### D3. Module of the release, not of `main`

The site uses the WASM archive of the release the page names, not a build of `main`. The playground then runs what the installation section offers, under that release's name, and `wasm_exec.js` comes from the same Go toolchain as `plantuml.wasm`, because both are in one archive. The archive is checked against `checksums.txt` of the release.

Alternative: `make wasm` in the workflow. It would run unreleased code under the release's name.

### D4. Site assembly tool

`tools/site` is an original Go command that writes the site into an output directory:

- It copies `site/` and reads the release from the meta tag of `index.html`.
- With `-wasm <dir>`, it copies `plantuml.wasm`, `wasm_exec.js` and `plantuml.mjs` from a local build into `<out>/wasm/`. `make site` runs `make wasm` and then the tool with `-wasm dist/wasm` and the output `dist/site`.
- With `-download <base URL>`, it downloads `checksums.txt` and `gopuml_<version>_wasm.tar.gz` of the release, compares the SHA-256 of the archive with its line in `checksums.txt`, and unpacks the three files into `<out>/wasm/`. On a mismatch it fails and leaves no output.
- When `checksums.txt` of the release answers 404, the release is not published: the tool prints a `::notice::` line, writes no output, writes `deploy=false` to the file that `GITHUB_OUTPUT` names (when set), and exits with status 0. Otherwise it writes `deploy=true`.

It uses only the standard library (`net/http`, `crypto/sha256`, `archive/tar`, `compress/gzip`), and its tests use `httptest` servers for a published release, a mismatch and a missing release. While the repository is private, the download answers 404, so the workflow deploys nothing instead of failing.

Alternative: a shell script with `gh`, `curl` and `sha256sum`. It would be tested only by running it against GitHub.

### D5. Pages workflow

`.github/workflows/pages.yml`:

- Triggers: `push` to `main` with paths `site/**`, `tools/site/**` and the workflow file; `workflow_run` of the `release` workflow with type `completed`, the job running only when its conclusion is `success`; `workflow_dispatch`. `workflow_run` is used because the release, created with `GITHUB_TOKEN`, does not start a `release: published` workflow.
- Permissions: `contents: read`, `pages: write`, `id-token: write`. Concurrency group `pages`, without cancelling a deployment in progress.
- One job in the `github-pages` environment: checkout, `setup-go` from `go.mod`, `go run ./tools/site -download https://github.com/ekalinin/gopuml/releases/download -out _site`, then `actions/configure-pages`, `actions/upload-pages-artifact` (path `_site`) and `actions/deploy-pages`, these three only when the tool wrote `deploy=true`.

### D6. Playground

- Lazy loading: `page.mjs` calls `load()` of `wasm/plantuml.mjs` only when an `IntersectionObserver` reports the playground section as visible, or when the editor gets the focus, whichever comes first. Until then the page requests no WASM file. While the module loads, the result area shows a loading message; when `load()` rejects, it shows that the playground could not be loaded.
- Rendering: the source is rendered as SVG 300 ms after the last edit; a result that arrives after a newer edit is dropped.
- Display: the SVG is shown in an `<img>` from a `blob:` URL of type `image/svg+xml`, not inserted into the page. Scripts and event attributes of an SVG do not run in an `<img>`, so a source cannot run code in the page; links inside diagrams are not clickable in the playground.
- Errors: for `result.error`, the error image is shown with the message and `line N`.
- Examples: a selector with sequence and new-syntax activity examples, which all render in the WASM build. The first sequence example is the initial source. A note under the editor lists the types that render in the browser and says that the others need Graphviz.
- Download: a link with the `download` attribute (`diagram.svg`) on the `blob:` URL of the last successful SVG.
- `playground.mjs` exports a controller created with a `load` function, so that tests can pass the real loader or a fake one.

### D7. Tests and cases

- `landing_test.go` (root package):
  - `TestLandingPage`: the sections of D1 are in the static HTML; the rows of the supported diagrams table of the README are in the table of the page; the installation shows the release archives, `checksums.txt` and `go install`; the links to the repository, the releases and pkg.go.dev.
  - `TestLandingRelease`: the release checks of D2, with mutations of the page (another tag, a changed cell of the comparison) that must make the check fail.
  - `TestPagesWorkflow`: the workflow of D5, parsed and checked with mutations (no `workflow_run`, no success condition, `pull_request` trigger, missing `id-token: write`, deployment not conditional on the tool output), as `TestReleaseWorkflow` does.
- `tools/site`: tests of a local build, a published release, a checksum mismatch and a missing release.
- `cmd/plantuml-wasm/playground_test.go`: builds the module as `wasm_test.go` does, and runs a Node.js script on `site/playground.mjs` with the real loader (every example renders without error; a sequence source gives SVG; a syntax error on line 3 gives line 3 and an error image; a class diagram gives the graph layout error; the download is an `image/svg+xml` file named `diagram.svg`) and with fake loaders (no load before activation; a rejected load gives the error state). It skips when `node` is not installed, as `wasm_test.go` does; `ubuntu-latest` runners have Node.js.
- One `.go.yaml` case per scenario of the spec in `testdata/conformance/cases/landing-page/`, naming these tests. The DOM glue of `page.mjs` (the observer, the debounce, the `<img>`) is checked by hand in a browser.

## Risks / Trade-offs

- **[Size of the module]** The playground downloads about 36 MB uncompressed. → It loads only when the playground is reached (D6); the transferred size is checked after the first deployment (see Open Questions).
- **[Page out of date after a release]** A new release could leave the page on the previous one. → The page and the README must name the same release (D2), and the README already changes with each release.
- **[Failed deployment on the merge of a release bump]** The page names a release that is published only after the merge. → The tool skips the deployment with a notice, and the `workflow_run` trigger deploys after the release workflow succeeds.
- **[Node.js not installed locally]** The playground tests skip. → CI runs them; the skip is visible in `go test -v`.
- **[Repository settings]** Pages has to be enabled with "GitHub Actions" as the source, which the workflow cannot do. → A task for the user, after the repository is made public.

## Migration Plan

1. Merge the pull request of this change. While the repository is private, the workflow deploys nothing (D4).
2. The user makes the repository public, and enables GitHub Pages with "GitHub Actions" as the source (`gh api -X POST repos/ekalinin/gopuml/pages -f build_type=workflow`).
3. Run the Pages workflow by hand, check the page and the playground at `https://ekalinin.github.io/gopuml/`, and set it as the homepage of the repository.

Rollback: disable GitHub Pages in the settings, or revert the commit and run the workflow again.

## Open Questions

- Does GitHub Pages compress `plantuml.wasm` (a `content-encoding: gzip` response)? It is checked after the first deployment. If it does not, a follow-up can serve a gzip copy and decompress it with `DecompressionStream`, without changing the spec.
