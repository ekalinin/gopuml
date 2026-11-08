# Spec Delta

## MODIFIED Requirements

### Requirement: Bundled libraries

The system SHALL bundle the libraries of the standard library of PlantUML v1.2026.9, with the content that the plantuml-mit 1.2026.9 jar ships, except the libraries whose license does not allow redistribution or cannot be verified. `adaml` (GPL-3.0) SHALL NOT be bundled. NOTICE SHALL list every bundled library with its source and license, and every excluded library with the reason. Library names SHALL be matched in lower case where upstream lower-cases them (include paths, sprite and image references) and as given where upstream does not (the arguments of `%get_stdlib`).

#### Scenario: Library name in another case

- **WHEN** a block contains `!include <c4/C4_Container>` and `Container(api, "API", "Go")`
- **THEN** the diagram renders the container `API` with the technology `[Go]`, as with `!include <C4/C4_Container>`

#### Scenario: Excluded library not listed

- **WHEN** a block evaluates `%get_all_stdlib()`
- **THEN** the result is a JSON array of the bundled library names in name order, which contains `c4` and `awslib14` and does not contain `adaml`

#### Scenario: Archimate library

- **WHEN** a block contains `!include <archimate/Archimate>`, `Business_Actor(a, "Customer")`, `Business_Service(s, "Sales")` and `Rel_Serving(s, a, "serves")`
- **THEN** the diagram draws `Customer` and `Sales` filled with `#FFFFCC`, with the business-actor and business-service icons in their upper right corner, and an arrow labeled `serves` from `Sales` to `Customer`, as upstream does
