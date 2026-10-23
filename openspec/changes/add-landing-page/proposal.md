# Proposal

## Why

gopuml has a release (v0.1.0), but the project has no page of its own: the only entry points are the README and the release page on GitHub. The comparison of the start and rendering time with Java PlantUML is only in the release notes. The WebAssembly module can render diagrams in a browser, so the page can also let visitors try gopuml without installing anything.

## What Changes

- A landing page published on GitHub Pages at `https://ekalinin.github.io/gopuml/`, written in plain HTML, CSS and JavaScript in a new `site/` directory, without a site generator and without Node.js.
- The page presents the project: what it is (a command line that replaces `java -jar plantuml.jar`, a Go library, an HTTP server, a WebAssembly module), the supported diagram types and formats, the installation of the latest release, the conformance results, the license, and links to the repository, the releases and the Go package documentation.
- A comparison with upstream PlantUML on the page: the start time and the rendering time per diagram type, from the `make bench` table of the release notes of the release the page names, with the method and the environment of the run.
- A playground on the page: an editor where a visitor writes a diagram and sees it rendered as SVG by the WebAssembly module of the release, with example diagrams, error messages with their line, and a download of the result. Sequence diagrams and activity diagrams in the new syntax render; graph-based types show the "graph layout is not available" error of the WASM build.
- The page names one release, the same as the installation section of the README. Its installation, comparison and playground all refer to that release, and tests check that the page agrees with the README and with the notes of the release.
- A GitHub Actions workflow that assembles the site from `site/` and the WASM archive of the release the page names (verified against `checksums.txt`) and deploys it to GitHub Pages: on a push to `main` that changes the site, after a successful release workflow, and by hand.
- A `make site` target that assembles the same site locally from `make wasm`, for a preview.
- A link to the page in the README.

Out of scope:

- a custom domain
- new benchmark measurements: the page shows the table measured for the release
- a Graphviz build for the browser, so that graph-based diagrams render in the playground
- more pages than the landing (documentation, a blog)

## Capabilities

### New Capabilities

- `landing-page`: the published landing page, its content, the comparison with upstream, the playground, the deployment workflow, and the consistency of the page with the release it names

### Modified Capabilities

None.

## Impact

- **New files:** `site/` (the page, its styles and the playground script), `.github/workflows/pages.yml`, and root-package tests of the workflow and the page.
- **Changed files:** `Makefile` (`site` target), `README.md` (link to the page).
- **Product code:** none. The playground uses the existing `plantuml.mjs` API of the WASM archive.
- **CI:** a new workflow. `ci.yml` and `release.yml` do not change.
- **Repository settings:** GitHub Pages has to be enabled with "GitHub Actions" as the source. On the free plan, GitHub Pages works only for public repositories, so the page can be published only after the repository is made public.
- **Dependencies:** the official `actions/configure-pages`, `actions/upload-pages-artifact` and `actions/deploy-pages` actions in CI; no module dependency.
