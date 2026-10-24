# Spec Delta

## MODIFIED Requirements

### Requirement: Flag compatibility

The CLI SHALL accept the upstream v1.2026.9 flag names, their short aliases and their deprecated single-dash aliases (for example `--svg`, `-tsvg` and `-svg`), matched case-insensitively as upstream does. This change SHALL support:

- formats: `--svg`, `--png`, `--txt`, `--utxt`, `--preproc`, `-f`/`--format`
- pipe mode: `-p`/`--pipe`, `-pipedelimitor`, `-pipenostderr`, `--pipe-image-index`
- input and output: `-o`/`--output-dir`, `--charset`, `-x`/`--exclude`, `--overwrite`, `--ignore-startuml-filename`, `-useseparatorminus`, `-filedir`, `-filename`
- preprocessing and styling: `-D`/`--define`, `-I`/`--include`, `-P`/`--pragma`, `-S`/`--skinparam`, `--theme`, `--config`, `--dark-mode`
- error handling: `--check-syntax`, `-syntax`, `--stop-on-error`, `--check-before-run`, `--no-error-image`
- Graphviz: `--graphviz-timeout`, `--dot-path`, `--check-graphviz`
- utilities: `--threads`, `--encode-url`, `--decode-url`, `--extract-source`, `--disable-metadata`, `--skip-fresh`, `--list-keywords`, `-stdlib`, `--http-server`
- information: `--version`, `-h`/`--help`, `-license`
- logging: `--verbose`, `--duration`, `-headless`

#### Scenario: Deprecated alias

- **WHEN** the CLI runs `plantuml -tsvg diagram.puml`
- **THEN** it behaves exactly like `plantuml --svg diagram.puml`

#### Scenario: Case-insensitive flag

- **WHEN** the CLI runs `plantuml -TSVG diagram.puml`
- **THEN** it produces `diagram.svg`

#### Scenario: Unknown flag

- **WHEN** the CLI runs `plantuml --foo diagram.puml`
- **THEN** it ignores `--foo`, as upstream does, writes `diagram.png` and exits with status 0

### Requirement: Utility commands

The CLI SHALL support the following utility commands:

- `--encode-url`: prints the encoded text of each diagram
- `--decode-url <text>`: prints the decoded source wrapped in a further `@startuml`/`@enduml` pair, as upstream does
- `--extract-source`: prints the source embedded in PNG or SVG files, framed as upstream frames it (a dashed line, the file name, a blank line, the source, then for PNG a blank line and the upstream version, and a dashed line)
- `--disable-metadata`: omits source metadata
- `--skip-fresh`: skips a diagram when the existing output already embeds the same source
- `--check-graphviz`: runs the graph-layout check
- `--threads <n|auto>`: renders files in parallel without changing any output
- `--list-keywords`: prints the language keywords
- `-stdlib`: lists the bundled standard libraries
- `--http-server[:port]`: starts the HTTP server
- `--version`: prints a version string that includes the targeted upstream version `1.2026.9`

#### Scenario: Extract source from PNG

- **WHEN** a PNG produced by this CLI is passed with `--extract-source`
- **THEN** the CLI prints the diagram source that produced it

#### Scenario: Skip fresh output

- **WHEN** `diagram.svg` was produced from the current `diagram.puml` and the CLI runs `plantuml --svg --skip-fresh diagram.puml`
- **THEN** `diagram.svg` is not rewritten

#### Scenario: Parallel rendering

- **WHEN** the same set of files is rendered with `--threads 1` and with `--threads auto`
- **THEN** both runs produce byte-identical output files
