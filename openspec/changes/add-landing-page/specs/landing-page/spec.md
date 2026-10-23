# Spec Delta

## Purpose

Publishes a landing page of gopuml on GitHub Pages that presents the project, compares it with upstream PlantUML, and lets visitors render diagrams in the browser with the WebAssembly module of a release.

## ADDED Requirements

### Requirement: Published landing page

The project SHALL publish a landing page at `https://ekalinin.github.io/gopuml/` as static files, without a server-side component. The content of the page other than the playground SHALL be readable without JavaScript.

#### Scenario: Visit the page

- **WHEN** a visitor opens `https://ekalinin.github.io/gopuml/`
- **THEN** the page shows the name of the project, what it is, and links to the repository and the releases

#### Scenario: JavaScript disabled

- **WHEN** a visitor opens the page with JavaScript disabled
- **THEN** every section except the playground is readable

### Requirement: Project presentation

The page SHALL describe the four forms of gopuml (a command line that replaces `java -jar plantuml.jar`, a Go library, an HTTP server, a WebAssembly module), the supported diagram types with their output formats and Graphviz requirement, the installation of the release it names (the release archives with their checksums, and `go install`), the conformance results, and the license. It SHALL link to the repository, the releases and the Go package documentation.

#### Scenario: Supported diagram types

- **WHEN** a visitor reads the supported diagram types on the page
- **THEN** they are the rows of the table of supported diagrams of the README, with the same formats and the same Graphviz requirement

#### Scenario: Installation

- **WHEN** a visitor reads the installation section
- **THEN** it names the release the page names, and shows how to download an archive with its checksums file and how to install the command line with `go install`

### Requirement: Comparison with upstream on the page

The page SHALL show the comparison of gopuml with upstream PlantUML from the notes of the release it names: the start time and the rendering time of each supported diagram type for both implementations, how they are measured, the machine, the Java, Go and Graphviz versions, the number of runs, and that the numbers depend on the machine.

#### Scenario: Comparison of v0.1.0

- **WHEN** the page names the release `v0.1.0`
- **THEN** its comparison table has the values of the table in `docs/releases/v0.1.0.md`, with the environment of that run

### Requirement: Consistency with the release

The page SHALL name exactly one release, the same release as the installation section of the README, and that release SHALL have notes in `docs/releases/<tag>.md`. The tests of the repository SHALL fail when the page names another release than the README, or when its comparison table differs from the table in the notes of that release.

#### Scenario: Page behind the README

- **WHEN** the README names `v0.2.0` and the page still names `v0.1.0`
- **THEN** the tests of the repository fail

#### Scenario: Comparison out of date

- **WHEN** a value of the comparison table on the page differs from the table in the notes of the release it names
- **THEN** the tests of the repository fail

### Requirement: Playground

The page SHALL provide a playground: an editor whose source is rendered as SVG in the browser by the WebAssembly module of the release the page names, without a server. The rendering SHALL follow the edits of the source. The playground SHALL offer example diagrams, SHALL show the error message and its line when the source has an error, SHALL state which diagram types render in the browser, and SHALL let the visitor download the rendered SVG.

#### Scenario: Sequence diagram

- **WHEN** a visitor types `@startuml`, `Alice -> Bob : hello` and `@enduml` in the editor
- **THEN** the playground shows the rendered sequence diagram as SVG

#### Scenario: Syntax error

- **WHEN** the source has a syntax error on its third line
- **THEN** the playground shows the error message with line 3, together with the error image

#### Scenario: Graph-based diagram

- **WHEN** a visitor renders a class diagram
- **THEN** the playground shows the error of the WASM build that graph layout is not available, together with the error image

#### Scenario: Example diagram

- **WHEN** a visitor selects an example diagram
- **THEN** its source replaces the source of the editor and is rendered

#### Scenario: Download

- **WHEN** a visitor downloads the result after a successful rendering
- **THEN** the browser saves the rendered SVG document

### Requirement: Loading of the module

The page SHALL NOT download the WebAssembly module until the playground becomes visible or the visitor interacts with it. While the module loads, the playground SHALL show that it is loading. When the module cannot be loaded, the playground SHALL show an error, and the rest of the page SHALL stay usable.

#### Scenario: Reading the top of the page

- **WHEN** a visitor reads the top of the page and does not reach the playground
- **THEN** the browser has not requested the WebAssembly module

#### Scenario: Module not available

- **WHEN** the request for the WebAssembly module fails
- **THEN** the playground shows that it could not be loaded, and the other sections of the page are unaffected

### Requirement: Deployment

A workflow SHALL assemble the site from the sources of the page and the WASM archive of the release the page names, and deploy it to GitHub Pages. It SHALL run on a push to `main` that changes the page or the workflow, after a successful run of the release workflow, and by hand. It SHALL verify the SHA-256 of the WASM archive against the checksums file of that release, and deploy nothing when they differ. When the release the page names is not published yet, it SHALL deploy nothing and finish without failing, with a notice.

#### Scenario: Change of the page

- **WHEN** a commit that changes the page is pushed to `main`, and the release the page names is published
- **THEN** the site is deployed with that commit's page and that release's WebAssembly module

#### Scenario: Checksum mismatch

- **WHEN** the SHA-256 of the downloaded WASM archive differs from the checksums file of the release
- **THEN** the workflow fails and the published site does not change

#### Scenario: Release not published yet

- **WHEN** a commit that makes the page name `v0.2.0` is pushed to `main` before the release `v0.2.0` is published
- **THEN** the workflow deploys nothing and succeeds with a notice, and the site is deployed after the release workflow of `v0.2.0` succeeds

### Requirement: Local preview

A build target SHALL assemble the same site locally from the sources of the page and a local WebAssembly build, without publishing anything, so that it can be served by any static file server.

#### Scenario: Preview

- **WHEN** the target runs and its output directory is served by a static file server
- **THEN** the page and its playground work as on GitHub Pages, with the local WebAssembly build
