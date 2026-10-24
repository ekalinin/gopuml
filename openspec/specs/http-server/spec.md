# http-server Specification

## Purpose

Serves rendered diagrams over HTTP using the URL schemes of PlantUML v1.2026.9's built-in server (`--http-server`) and plantuml-server, so that existing URLs, editors and integrations can point to this implementation.

## Requirements

### Requirement: Starting the server

The system SHALL start an HTTP server from the CLI flag `--http-server[:port]` and from the library. The default port SHALL be 8080.

#### Scenario: Default port

- **WHEN** the CLI runs `plantuml --http-server`
- **THEN** the server accepts rendering requests on port 8080

### Requirement: GET rendering

The server SHALL render encoded source from `GET` requests on these paths, with and without a leading `/plantuml` prefix:

- `/png/<encoded>` and `/img/<encoded>` as PNG
- `/svg/<encoded>` as SVG
- `/txt/<encoded>` as TXT
- `/utxt/<encoded>` as UTXT

Paths with a `d` prefix (`/dpng/`, `/dimg/`, `/dsvg/`) SHALL render in dark mode, as plantuml-server does. Responses SHALL use the format's MIME type and include:

- `X-PlantUML-Diagram-Width`, `X-PlantUML-Diagram-Height` and `X-PlantUML-Diagram-Description`, as upstream's built-in server sends them
- `Access-Control-Allow-Origin: *`

The server SHALL answer both upstream servers' routes: the `/plantuml` prefix, `/utxt` and `/serverinfo` come from upstream's built-in server, and `/img` and the `d`-prefixed routes come from plantuml-server.

#### Scenario: SVG by URL

- **WHEN** a client requests `GET /svg/<encoded>` for an encoded valid sequence diagram
- **THEN** the response has status 200, content type `image/svg+xml`, the SVG body, and the diagram width and height headers

#### Scenario: Dark mode URL

- **WHEN** a client requests `GET /dsvg/<encoded>`
- **THEN** the SVG is rendered in dark mode

### Requirement: POST rendering

The server SHALL accept:

- `POST /render` with a JSON body `{"source": "...", "options": [...]}`, where `options` holds CLI flags, as upstream's built-in server does
- `POST /png`, `POST /svg` and `POST /txt` with raw source text in the body, as plantuml-server does

Source without a `@start` line SHALL be wrapped in `@startuml`/`@enduml`. A request body that cannot be parsed SHALL get status 400.

#### Scenario: JSON render request

- **WHEN** a client posts `{"source": "Alice -> Bob", "options": ["--svg"]}` to `/render`
- **THEN** the response contains an SVG showing a message from `Alice` to `Bob`

#### Scenario: Invalid JSON

- **WHEN** a client posts `{not json` to `/render`
- **THEN** the response status is 400

### Requirement: Error responses

For a diagram with a syntax error, the server SHALL return the error image with status 400, as both upstream servers do; `POST /render` SHALL return it with status 200, as upstream's built-in server does. It SHALL add the `X-PlantUML-Diagram-Error` and `X-PlantUML-Diagram-Error-Line` headers, with the line number of upstream's built-in server. Encoded text that cannot be decoded SHALL get status 200 with upstream's error image about a bad URL, as upstream's built-in server does. A rendering failure such as a Graphviz crash SHALL get status 503, as upstream's built-in server does.

#### Scenario: Syntax error by URL

- **WHEN** a client requests `GET /svg/<encoded>` for a diagram with a syntax error
- **THEN** the response has status 400 and contains the error image, and the error headers name the message and line

#### Scenario: Undecodable URL

- **WHEN** a client requests `GET /svg/@@not-valid@@`
- **THEN** the response status is 200 and the body is upstream's error image about a bad URL

### Requirement: Utility endpoints

The server SHALL provide:

- `GET /check/<encoded>`: returns the syntax check result as text
- `GET /coder/<encoded>`: returns the decoded source as text
- `POST /coder` with source text: returns the encoded text
- `GET /language`: returns the language keywords
- `GET /serverinfo`: returns JSON with the version and supported formats

Each endpoint SHALL use the same response format as the upstream server that defines it.

#### Scenario: Encode through the coder endpoint

- **WHEN** a client posts `Alice -> Bob` to `/coder`
- **THEN** the response body is encoded text that decodes back to that source

### Requirement: Server security defaults

The server SHALL use the `INTERNET` security profile unless `PLANTUML_SECURITY_PROFILE` is set. It SHALL reject request bodies and decoded sources larger than a configurable limit (default 1 MiB) with status 413. It SHALL stop rendering requests that exceed a configurable time limit (default 60 seconds) and return status 503.

#### Scenario: Local file include over HTTP

- **WHEN** no security profile is set and a client requests a diagram containing `!include /etc/passwd`
- **THEN** the include is denied and the file content does not appear in the response

#### Scenario: Oversized source

- **WHEN** a client posts 2 MiB of source to `/svg`
- **THEN** the response status is 413
