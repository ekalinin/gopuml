# preprocessor Specification

## Purpose

Provides PlantUML v1.2026.8 compatible preprocessing of each diagram block before parsing: variables, macros, control flow, functions, builtins, JSON data, file inclusion and access control.

## Requirements

### Requirement: Preprocessing before parsing

The system SHALL run the preprocessor on every diagram block before diagram type detection. Diagram parsing SHALL see only the preprocessed lines.

#### Scenario: Variable used in a message

- **WHEN** a block contains `!$user = "Alice"` and `$user -> Bob : hi`
- **THEN** the parsed diagram contains a message from `Alice` to `Bob`

### Requirement: Variables and legacy defines

The system SHALL support `!$name = value` assignment, conditional assignment with `?=`, `!local` and `!global` scope modifiers, and the legacy `!define`, `!definelong`/`!enddefinelong` and `!undef` macros with the same substitution rules as upstream.

#### Scenario: Conditional assignment keeps the existing value

- **WHEN** a block contains `!$color = "red"` followed by `!$color ?= "blue"`
- **THEN** `$color` evaluates to `red`

#### Scenario: Legacy macro with arguments

- **WHEN** a block contains `!define SEND(a,b) a -> b` and the line `SEND(Alice,Bob)`
- **THEN** the preprocessed line is `Alice -> Bob`

### Requirement: Conditionals and loops

The system SHALL support `!if`/`!elseif`/`!else`/`!endif`, `!ifdef`/`!ifndef`, `!while`/`!endwhile` and `!foreach`/`!endfor`, with expressions evaluated by the same operator rules as upstream.

#### Scenario: If-else branch

- **WHEN** a block contains `!$n = 2`, `!if $n > 1`, `A -> B : many`, `!else`, `A -> B : one`, `!endif`
- **THEN** only the message `many` is in the preprocessed output

#### Scenario: While loop

- **WHEN** a block contains a `!while $i < 3` loop that emits `A -> B : $i` and increments `$i` from 0
- **THEN** the preprocessed output contains the three messages `0`, `1` and `2` in order

### Requirement: Procedures and functions

The system SHALL support `!procedure` and `!function` definitions with `!return`, default argument values, keyword arguments, `!unquoted` and `!final` modifiers, and calling them from any line.

#### Scenario: Procedure call with default argument

- **WHEN** a block defines `!procedure $msg($to, $text="ping")` emitting `A -> $to : $text` and calls `$msg("B")`
- **THEN** the preprocessed output contains `A -> B : ping`

#### Scenario: Function return value

- **WHEN** a block defines `!function $double($x)` returning `$x * 2` and contains `A -> B : $double(21)`
- **THEN** the preprocessed output contains `A -> B : 42`

### Requirement: Builtin functions

The system SHALL provide every builtin function that PlantUML v1.2026.8 provides (for example `%strlen`, `%upper`, `%substr`, `%dec2hex`, `%date`, `%splitstr`, `%json_set`, `%darken`, `%version`), with the same argument handling and results. `%version` SHALL return the upstream version that this implementation targets.

#### Scenario: String builtins

- **WHEN** a block evaluates `%strlen("abc")` and `%upper("abc")`
- **THEN** the results are `3` and `ABC`

#### Scenario: Numeric conversion builtin

- **WHEN** a block evaluates `%dec2hex(255)`
- **THEN** the result is `ff`

### Requirement: JSON data

The system SHALL support JSON literals as variable values, member and index access (for example `$data.items[0].name`), `%load_json` and the `%json_*` builtins, with the same results as upstream.

#### Scenario: JSON member access

- **WHEN** a block contains `!$data = { "name": "api", "ports": [80, 443] }` and `A -> B : $data.ports[1]`
- **THEN** the preprocessed output contains `A -> B : 443`

### Requirement: File inclusion

The system SHALL support `!include` (resolved relative to the including file, then the include path), `!include_many`, `!include_once`, `!includesub` with `!startsub`/`!endsub`, `!includedef`, and `!import` of zip archives. As in upstream v1.2026.8, each of `!include`, `!include_many` and `!include_once` SHALL include the file every time the directive appears. An included file that contains `@startuml`...`@enduml` SHALL contribute only the lines of its first block, unless a block id is selected with `!include file.puml!id`.

#### Scenario: Include relative to the including file

- **WHEN** `diagrams/main.puml` contains `!include common/style.iuml` and `diagrams/common/style.iuml` exists
- **THEN** the lines of `diagrams/common/style.iuml` are preprocessed in place of the directive

#### Scenario: Include once

- **WHEN** a block contains `!include_once a.iuml` twice
- **THEN** the lines of `a.iuml` are included twice and no error is reported

#### Scenario: Missing include file

- **WHEN** a block contains `!include missing.iuml` and the file does not exist
- **THEN** the system reports an error at that line with the message `cannot include missing.iuml`

### Requirement: Security profiles

The system SHALL restrict file and URL access by the security profile given in the `PLANTUML_SECURITY_PROFILE` environment variable, accepting the upstream values `SANDBOX`, `ALLOWLIST`, `INTERNET`, `INTERNET_WITH_DOTSVG`, `LEGACY` and `UNSECURE`. When the variable is not set, the CLI and library SHALL use `LEGACY` and the HTTP server SHALL use `INTERNET`. Paths under `/etc/`, `/dev/`, `/boot/`, `/proc/` and `/sys/`, and paths starting with `//`, SHALL be denied in every profile except `UNSECURE`, as upstream does. A denied access SHALL be reported at the directive line as `cannot include <path>` and SHALL NOT reveal other local file system paths.

#### Scenario: Sandbox blocks local files

- **WHEN** `PLANTUML_SECURITY_PROFILE=SANDBOX` and a block contains `!include common.iuml`, and `common.iuml` exists next to the source
- **THEN** the include is denied and reported as an error at that line

#### Scenario: System paths are denied by default

- **WHEN** no security profile is set and a block contains `!include /etc/passwd`
- **THEN** the include is denied and reported as `cannot include /etc/passwd` at that line

#### Scenario: Legacy allows local files

- **WHEN** no security profile is set and the CLI renders a file containing `!include common.iuml` that exists next to it
- **THEN** the include succeeds

#### Scenario: URL include under the internet profile

- **WHEN** `PLANTUML_SECURITY_PROFILE=INTERNET` and a block contains `!include https://example.com/style.iuml`
- **THEN** the system fetches and includes the URL content

### Requirement: Standard library includes

The system SHALL resolve `!include <library/path>` against an embedded standard library, using the same library and path naming as upstream. A library that this build does not bundle SHALL produce an error at the directive line that names the library and states it is not bundled.

#### Scenario: Library not bundled

- **WHEN** a block contains `!include <C4/C4_Container>` and the C4 library is not bundled
- **THEN** the system reports an error at that line stating that the `C4` library is not bundled

### Requirement: Assertions, logging and diagnostics

The system SHALL support `!assert` (failing assertions are reported as an error at that line with the assertion message), `!log` (writes to the diagnostic log, not to the diagram) and `!dump_memory`.

#### Scenario: Failing assertion

- **WHEN** a block contains `!assert 1 == 2 : "numbers differ"`
- **THEN** the system reports an error at that line containing `numbers differ`

### Requirement: Pragmas and options

The system SHALL accept `!pragma` and `!option` directives and pass their values to diagram processing, with the same names and effects as upstream for the features this change implements. Pragmas for features outside this change SHALL be accepted and ignored.

#### Scenario: Layout pragma for an unsupported engine

- **WHEN** a class diagram contains `!pragma layout elk`
- **THEN** the diagram renders with the default layout engine and no error is reported

### Requirement: Preprocessor errors

The system SHALL report undefined functions in expressions, wrong argument counts, unbalanced control directives and invalid expressions as errors at the offending line. As upstream does, a function call in plain diagram text that names no defined function SHALL be left unchanged.

#### Scenario: Undefined function

- **WHEN** a block contains `!$x = $missing(1)` and `$missing` is not defined
- **THEN** the system reports an error at that line with the message `Unknown built-in function $missing`

#### Scenario: Undefined function in diagram text

- **WHEN** a block contains `A -> B : $missing(1)` and `$missing` is not defined
- **THEN** the message label is the literal text `$missing(1)` and no error is reported
