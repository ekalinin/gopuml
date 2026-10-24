# Spec Delta

## MODIFIED Requirements

### Requirement: Sandboxed resource access

The WASM build SHALL run in the `SANDBOX` security profile by default. It SHALL resolve local includes only from sources supplied by the host page, through the render options, and standard library includes (`!include <library/path>`) from the standard library embedded in the module, which is the same as the native build's.

#### Scenario: Include supplied by the host

- **WHEN** the host passes a file named `style.iuml` in the render options and the source contains `!include style.iuml`
- **THEN** the include is resolved from the supplied file

#### Scenario: Standard library include

- **WHEN** the host supplies no files and the source is a sequence diagram that contains `!include <tupadr3/font-awesome/server>` and `Alice -> Bob : <$server> hi`
- **THEN** the render succeeds, and its SVG output is identical to the native build's output for the same source
