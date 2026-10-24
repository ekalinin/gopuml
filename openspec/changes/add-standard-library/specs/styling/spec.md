# Spec Delta

## MODIFIED Requirements

### Requirement: Creole text markup

The system SHALL render creole markup in every text that upstream renders with creole:

- bold `**`, italic `//`, monospace `""`, strike `--`, underline `__`, wave `~~`
- bulleted `*` and numbered `#` lists, tables with `|=` header cells, horizontal rules
- the HTML-like tags `<b>`, `<i>`, `<u>`, `<s>`, `<color>`, `<back>`, `<size>`, `<font>`
- links `[[url label]]`, the `~` escape, and Unicode escapes such as `<U+00E9>` and `&#233;`
- sprite references (`<$name>`), drawn as the sprites capability specifies

OpenIconic (`<&name>`) and emoji (`<:name:>`) references SHALL render as their literal text until the follow-up change that adds them.

#### Scenario: Bold and colored text

- **WHEN** a note contains `**important** <color:red>now</color>`
- **THEN** `important` is drawn bold and `now` is drawn in red

#### Scenario: Table in a note

- **WHEN** a note contains the lines `|= Key |= Value |` and `| a | 1 |`
- **THEN** the note renders a two-column table with a bold header row

#### Scenario: Icon reference before icons are supported

- **WHEN** a note contains `<&check> done`
- **THEN** the note renders the literal text `<&check> done`

#### Scenario: Sprite reference in a note

- **WHEN** a sequence diagram defines `sprite $foo [4x4/16]` with the rows `FFFF`, `F00F`, `F00F` and `FFFF`, and a note contains `<$foo> done`
- **THEN** the note draws the sprite followed by `done`, as upstream does
