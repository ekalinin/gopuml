# Spec Delta

## Purpose

Provides a `plantuml` command that can replace `java -jar plantuml.jar` v1.2026.8 for the diagram types and formats this implementation supports: the same flags, input selection, output naming, pipe mode and exit codes.

## ADDED Requirements

### Requirement: Flag compatibility

The CLI SHALL accept the upstream v1.2026.8 flag names, their short aliases and their deprecated single-dash aliases (for example `--svg`, `-tsvg` and `-svg`), matched case-insensitively as upstream does. This change SHALL support:

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

### Requirement: Recognized but unsupported flags

When an upstream flag that this build does not implement is given (for example `--pdf`, `--gui`, `--ftp-server`, `--encode-sprite`), the CLI SHALL print an error naming the flag and stating it is not supported, write no output files, and exit with status 1.

#### Scenario: PDF requested

- **WHEN** the CLI runs `plantuml --pdf diagram.puml`
- **THEN** it prints that `--pdf` is not supported, creates no file, and exits with status 1

#### Scenario: No arguments

- **WHEN** the CLI runs without any arguments
- **THEN** it prints the usage help and exits with status 0

### Requirement: Input selection

The CLI SHALL accept file paths, directories and wildcard patterns, as upstream does:

- A directory SHALL select, without recursion, the files whose names match `(?i)^.*\.(txt|tex|java|htm|html|c|h|cpp|apt|pu|puml|hpp|hh)$`.
- A pattern with `*` or `?` SHALL match within one directory.
- A pattern with `**` SHALL match recursively.
- `-x`/`--exclude` patterns SHALL remove matches.
- The selected files of each argument SHALL be processed in sorted order.

#### Scenario: Directory argument

- **WHEN** the CLI runs `plantuml docs/` and `docs/` contains `a.puml`, `b.txt`, `c.plantuml` and a subdirectory with `d.puml`
- **THEN** it processes `a.puml` and `b.txt` only

#### Scenario: Recursive pattern with exclusion

- **WHEN** the CLI runs `plantuml "**/*.puml" -x "**/draft/**"`
- **THEN** it processes every `.puml` file under the current directory except those under `draft` directories

### Requirement: Output location and naming

The CLI SHALL write outputs next to each source file by default, or in the `-o` directory. A relative `-o` path SHALL be resolved against the source file's directory, and the directory SHALL be created when missing.

Output files SHALL be named as follows:

- The base name SHALL be the diagram name from the start line, unless the diagram has no name or `--ignore-startuml-filename` is given. In those cases it SHALL be the source file name without its extension.
- The first output with a given base name SHALL be `<base><suffix>`, where the suffix is `.svg`, `.png`, `.atxt`, `.utxt` or `.preproc`.
- Each further unnamed diagram from the same source, and each further page of a diagram, SHALL be `<base>_001<suffix>`, `<base>_002<suffix>` and so on. With `-useseparatorminus` the separator SHALL be `-`.
- A later block with the same name as an earlier one SHALL overwrite the earlier block's file, as upstream does.

#### Scenario: Two diagrams in one file

- **WHEN** the CLI runs `plantuml --svg flows.puml` and `flows.puml` contains two unnamed blocks
- **THEN** it writes `flows.svg` and `flows_001.svg`

#### Scenario: Named diagram

- **WHEN** `flows.puml` contains `@startuml login` and the CLI runs `plantuml --svg flows.puml`
- **THEN** it writes `login.svg`

#### Scenario: Named and unnamed blocks

- **WHEN** `flows.puml` contains an unnamed block, a block named `login` and another unnamed block, and the CLI runs `plantuml --svg flows.puml`
- **THEN** it writes `flows.svg`, `login.svg` and `flows_001.svg`

#### Scenario: Relative output directory

- **WHEN** the CLI runs `plantuml -o out docs/a.puml`
- **THEN** it writes `docs/out/a.png` and creates `docs/out` if it does not exist

### Requirement: Default format

The CLI SHALL produce PNG when no format flag is given.

#### Scenario: No format flag

- **WHEN** the CLI runs `plantuml diagram.puml`
- **THEN** it writes `diagram.png`

### Requirement: Pipe mode

With `-p`/`--pipe`, the CLI SHALL read source from standard input and write the rendered output of each diagram to standard output, in order:

- Input without any `@start` line SHALL be wrapped in `@startuml`/`@enduml`. Empty input SHALL exit with status 50.
- With `-pipedelimitor <text>`, the text SHALL be written after each image, followed by a newline: directly after `</svg>` for SVG, and on its own line for text formats.
- `--pipe-image-index <n>` SHALL select page n of each diagram.
- The error image SHALL go to standard output and the error report to standard error. With `-pipenostderr`, the error report SHALL go to standard output instead of the error image.

#### Scenario: Pipe with unmarked input

- **WHEN** `Alice -> Bob` is piped to `plantuml --svg -p`
- **THEN** standard output contains one SVG document showing a message from `Alice` to `Bob`

#### Scenario: Delimiter between images

- **WHEN** two blocks are piped to `plantuml --svg -p -pipedelimitor ===`
- **THEN** standard output contains two SVG documents, each followed directly by `===` and a newline

### Requirement: Exit codes

The CLI SHALL exit with the upstream codes:

- `0` on success
- `50` when no input file is found
- `100` when no diagram is found in the inputs
- `200` when at least one diagram has an error

`50` SHALL apply only when no input file is found at all: a missing file is ignored when another input file exists. When several conditions apply, `200` SHALL take precedence.

#### Scenario: Missing file

- **WHEN** the CLI runs `plantuml does-not-exist.puml`
- **THEN** it exits with status 50

#### Scenario: Syntax error

- **WHEN** the CLI renders a file whose only diagram has a syntax error
- **THEN** it writes the error image and exits with status 200

#### Scenario: Missing file next to an existing one

- **WHEN** the CLI runs `plantuml nodiag.puml missing.puml`, `nodiag.puml` exists without a diagram and `missing.puml` does not exist
- **THEN** it exits with status 100

### Requirement: Error handling options

The CLI SHALL support:

- `--no-error-image`: no image is written for a failing diagram
- `--stop-on-error`: processing stops at the first diagram with an error
- `--check-before-run`: all inputs are checked first, and nothing is generated when any has an error

#### Scenario: Check before run

- **WHEN** the CLI runs `plantuml --check-before-run good.puml bad.puml` and `bad.puml` has a syntax error
- **THEN** no output file is written for either input and the exit status is 200

### Requirement: Syntax checking

`--check-syntax` SHALL check every input without writing images, print `Some diagram description contains errors` to standard error when a diagram has an error, and exit with the upstream exit codes. `-syntax` SHALL read source from standard input. For a valid diagram it SHALL print the diagram type name and description. For an invalid diagram it SHALL print `ERROR`, the 0-based error line and the error message, each on its own line, and exit with status 200.

#### Scenario: Syntax-only check

- **WHEN** the CLI runs `plantuml --check-syntax bad.puml` and `bad.puml` has an error on a line
- **THEN** no image is written, standard error contains `Some diagram description contains errors`, and the exit status is 200

#### Scenario: Syntax from standard input

- **WHEN** `Alice -> Bob` is piped to `plantuml -syntax`
- **THEN** standard output contains the diagram type `SEQUENCE` followed by the diagram description

#### Scenario: Syntax error from standard input

- **WHEN** `Alice -> Bob` followed by the line `Alice -> -> Bob` is piped to `plantuml -syntax`
- **THEN** standard output contains `ERROR`, then `2`, then `Syntax Error? (Assumed diagram type: sequence)`, and the exit status is 200

### Requirement: Preprocessing and configuration flags

The CLI SHALL apply:

- `-D`/`--define` as preprocessor definitions
- `-I`/`--include` as files included before each diagram
- `-P`/`--pragma` as pragmas
- `-S`/`--skinparam` as skin parameters
- `--theme` as a theme
- `--config <file>` lines as lines inserted at the start of each diagram
- `--charset` as the input encoding (UTF-8 by default)

#### Scenario: Define from the command line

- **WHEN** the CLI runs `plantuml -DUSER=Alice diagram.puml` and the diagram contains `USER -> Bob`
- **THEN** the rendered diagram shows a message from `Alice` to `Bob`

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
- `--version`: prints a version string that includes the targeted upstream version `1.2026.8`

#### Scenario: Extract source from PNG

- **WHEN** a PNG produced by this CLI is passed with `--extract-source`
- **THEN** the CLI prints the diagram source that produced it

#### Scenario: Skip fresh output

- **WHEN** `diagram.svg` was produced from the current `diagram.puml` and the CLI runs `plantuml --svg --skip-fresh diagram.puml`
- **THEN** `diagram.svg` is not rewritten

#### Scenario: Parallel rendering

- **WHEN** the same set of files is rendered with `--threads 1` and with `--threads auto`
- **THEN** both runs produce byte-identical output files
