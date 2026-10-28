# output-formats Specification

## Purpose

Defines the output formats this change produces (SVG, PNG, TXT/UTXT and PREPROC), how the diagram source is embedded as metadata, and the guarantee that output is deterministic.

## Requirements

### Requirement: SVG output

The system SHALL render any supported diagram to a standalone SVG document with width, height and viewBox set from the diagram size. Creole links SHALL become SVG hyperlinks. Unless metadata is disabled, the SVG SHALL embed the encoded source in a `<?plantuml-src ...?>` processing instruction, as upstream does.

#### Scenario: SVG with metadata

- **WHEN** a sequence diagram is rendered to SVG with default options
- **THEN** the output is a well-formed SVG document that contains a `<?plantuml-src` processing instruction whose decoded value is the diagram source

#### Scenario: Link in SVG

- **WHEN** a class diagram contains `class Api [[https://example.com]]`
- **THEN** the SVG wraps the `Api` element in a hyperlink to `https://example.com`

### Requirement: PNG output

The system SHALL render any supported diagram to a PNG image. Unless metadata is disabled, the PNG SHALL store the source and the upstream version in an iTXt chunk with the keyword `plantuml`, as upstream does. `skinparam dpi` SHALL scale the pixel size the same way as upstream.

#### Scenario: PNG with metadata

- **WHEN** a class diagram is rendered to PNG with default options
- **THEN** the output is a valid PNG with a `plantuml` text chunk containing the diagram source

#### Scenario: DPI scaling

- **WHEN** a diagram rendered at the default DPI is 200 pixels wide and the same diagram is rendered with `skinparam dpi 192`
- **THEN** the second image is twice as wide within one pixel, as in upstream's output

### Requirement: Text output

The system SHALL render sequence diagrams as ASCII art (TXT, file suffix `.atxt`) and as Unicode box-drawing art (UTXT, file suffix `.utxt`), with the same layout rules as upstream. The system SHALL render gantt diagrams as upstream's table of tasks, drawn in ASCII for both TXT and UTXT, as upstream does. The system SHALL render WBS diagrams as upstream's boxes of the root and of its children, drawn in ASCII for both TXT and UTXT, as upstream does.

#### Scenario: ASCII sequence diagram

- **WHEN** `Alice -> Bob : hello` is rendered as TXT
- **THEN** the output is plain ASCII text that shows the participants `Alice` and `Bob` and an arrow labeled `hello`

#### Scenario: ASCII gantt table

- **WHEN** a gantt diagram with `Project starts 2020-01-06` and `[Build] lasts 3 days` is rendered as TXT
- **THEN** the output is an ASCII table with the header row `Task`, `Start`, `End` and `Duration` and the row `Build`, `Jan 6`, `Jan 8` and `3 days`, as upstream's TXT output

#### Scenario: ASCII WBS diagram

- **WHEN** a WBS diagram with `* Project`, `** Planning`, `*** Scope` and `** Build` is rendered as TXT
- **THEN** the output is ASCII text with the boxes of `Project`, `Planning` and `Build`, each box 4 columns to the right of and 4 lines below the previous one, without `Scope`, as upstream's TXT output

### Requirement: Preprocessed source output

The system SHALL output the preprocessed source of each block as text (PREPROC, file suffix `.preproc`), matching upstream's PREPROC output for the same input.

#### Scenario: Preprocessed output

- **WHEN** a block containing `!$x = "B"` and `A -> $x` is output as PREPROC
- **THEN** the output contains the line `A -> B` and no preprocessor directives

### Requirement: Unsupported format for a diagram type

When the requested format is not available for a diagram type in this change (for example TXT for a class diagram, which upstream supports and a follow-up change adds, or any format outside this change), the system SHALL report an error naming the format and the diagram type, and SHALL NOT produce an empty or partial file.

#### Scenario: TXT for a class diagram

- **WHEN** a class diagram is rendered as TXT
- **THEN** the system reports that TXT output is not supported for class diagrams and produces no output file

### Requirement: Disabling metadata

The system SHALL omit embedded source metadata from SVG and PNG output when the caller disables metadata.

#### Scenario: No metadata

- **WHEN** a diagram is rendered to SVG with metadata disabled
- **THEN** the SVG contains no `<?plantuml-src` processing instruction

### Requirement: Deterministic output

For the same input, options, and Graphviz version, the system SHALL produce byte-identical output across runs, machines and operating systems. Output SHALL NOT depend on fonts installed on the host, the current time or the locale. Values that are time-dependent or random by definition, such as those of `%date`, `%now` and `%random`, are exempt.

#### Scenario: Same output on two machines

- **WHEN** the same sequence diagram is rendered to SVG on macOS and on Linux
- **THEN** the two SVG files are byte-identical

### Requirement: Dark mode

The system SHALL render diagrams in dark mode when the caller requests it, transforming colors the same way as upstream's dark-mode color mapper.

#### Scenario: Dark background

- **WHEN** an unstyled diagram is rendered in dark mode
- **THEN** the background is dark and text is light, matching upstream dark-mode output within the parity tolerance
