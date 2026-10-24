# Spec Delta

## MODIFIED Requirements

### Requirement: Standard library includes

The system SHALL resolve `!include <library/path>`, and the other include directives with a `<library/path>` argument, against the bundled standard library, using the same library and path naming as upstream: names are matched in lower case, the `.puml` extension is optional, and an include inside a library file resolves within that library. Standard library includes SHALL be allowed in every security profile, as upstream allows them. A path that a bundled library does not contain SHALL produce the error `cannot include <library/path>` at the directive line. A library that upstream does not ship SHALL produce upstream's error for a missing library, `Fatal parsing error`, at the directive line. A library that upstream ships and this build does not bundle SHALL produce an error at the directive line that names the library and states it is not bundled.

#### Scenario: Library include

- **WHEN** a block contains `!include <C4/C4_Container>`, `Person(user, "User")`, `Container(api, "API", "Go")` and `Rel(user, api, "Uses")`
- **THEN** the diagram renders as upstream renders it: the person `User` with the C4 person icon, the container `API` with the technology `[Go]`, and the relation `Uses`

#### Scenario: Library include in the sandbox

- **WHEN** `PLANTUML_SECURITY_PROFILE=SANDBOX` and the CLI renders a file that contains `!include <C4/C4_Context>` and `System(sys, "System")`
- **THEN** the include succeeds and the diagram renders the system `System`

#### Scenario: Missing file in a library

- **WHEN** a block contains `!include <C4/Nope>`
- **THEN** the system reports an error at that line with the message `cannot include <C4/Nope>`

#### Scenario: Missing library

- **WHEN** a block contains `!include <foo/bar>`
- **THEN** the system reports an error at that line with the message `Fatal parsing error`

#### Scenario: Library not bundled

- **WHEN** a block contains `!include <adaml/AdaML>` and the adaml library is not bundled
- **THEN** the system reports an error at that line stating that the `adaml` library is not bundled
