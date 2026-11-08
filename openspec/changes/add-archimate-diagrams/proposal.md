# Proposal

## Why

The `archimate` commands of description diagrams give "archimate diagrams are not supported yet" today (issue #35). Every element macro of the bundled `archimate` library expands to such a command (`Business_Actor(a, "Customer")` becomes `archimate #BUSINESS "Customer" <<business-actor>> as a`), so `!include <archimate/Archimate>` fails as well. Below the commands, the 139 sprites of upstream's jar (`sprites/archimate/`), which draw the ArchiMate icons, are not bundled. A diagram that references one fails with "drawing crashed: ... not ported yet: klimt/sprite/SpriteImage.fromInternal" even without the `archimate` keyword, for example `rectangle "Customer" <<$archimate/business-actor>> as c`. No curated case covers archimate yet, so the report counts no failing or unsupported archimate case: this change adds the first ones. The `.go.yaml` cases `diagram-source/unsupported-uml-types` and `styling/embedded-unsupported-type` check the "not supported yet" error that goes away.

## What Changes

- Internal sprites, as upstream's `SpriteImage.fromInternal`: the 116 SVG and 23 PNG sprites of the plantuml-mit 1.2026.9 jar are bundled. They serve:
  - a sprite reference that the diagram does not define (`<$archimate/actor>`);
  - a sprite stereotype (`<<$archimate/business-actor>>`);
  - `sprite $name jar:archimate/<name>`.
  - An SVG sprite is drawn as vector shapes with upstream's default SVG sprite parser, in SVG and PNG output. A PNG sprite is drawn as an image.
- The three `archimate` commands of description diagrams, as upstream's `CommandArchimate`, `CommandArchimateMultilines` and `CommandArchimatePackage`, registered in upstream's order:
  - `archimate #color "Display" <<icon>> as code` creates an element with the archimate symbol, or updates the element of that name. The icon selects the `archimate/<icon>` sprite.
  - `archimate #color code <<icon>> [` ... `]` creates an element whose text is the lines between the brackets, drawn with the rectangle symbol, and fails with `Already exists <code>` when the name exists.
  - `archimate #color "Display" <<icon>> as code {` opens a group with the archimate symbol.
  - Upstream's surprising rules are kept: its pattern wants spaces both before and after the optional color, so `archimate "Customer" as c`, with one space and no color, is a syntax error. An icon without a sprite draws its stereotype (`<<nope>>`).
- `!include <archimate/Archimate>` renders, as upstream does.
- Unsupported `@startuml` types: archimate is removed from the list of `@startuml` types reported as not supported yet. The detection stays, with an empty list.
- The embedded diagram of an unsupported type is checked with `{{json`, which is still reported as not supported yet.
- NOTICE names the ArchiMate sprites, which upstream takes from Archi (MIT License).
- The README and the landing page list archimate with the description diagrams and no longer name it among the unsupported types. The README's conformance counts are updated.

Out of scope:

- `listsprite` and `listsprites` (`PSystemListArchimateSpritesFactory`), the utility diagram that lists the internal sprites (issue #57).
- The `@Explain` texts of the three commands (`explainArg`, `explainNow`), which belong to explain diagrams (issue #36).
- The other diagram types still reported as not supported yet.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `description-diagram`: the element keywords include `archimate`, and a new requirement covers the archimate elements, their multiline form, their groups and their icons
- `sprites`: a new requirement covers the internal sprites of upstream's jar
- `diagram-source`: archimate is no longer reported as a `@startuml` type that is not supported yet
- `styling`: the scenario of an embedded diagram of an unsupported type uses `{{json`
- `standard-library`: a scenario renders the bundled `archimate` library

## Impact

- **Code:**
  - `internal/klimt`: the embedded sprites and the port of `SpriteImage.fromInternal`; the list of internal sprite names goes away
  - `internal/descdiagram`: the three commands and their registration in the factory
  - `internal/core`: the archimate entry of the unsupported `@startuml` types
  - the tests that expect the panic, the archimate "not supported yet" error or the embedded archimate error image (`internal/klimt`, `internal/core`, `internal/engine`)
- **Runtime dependency:** none. The sprites add about 211 KB to the binary and the WASM module.
- **Conformance:** new curated cases for the archimate elements, the internal sprites and the `archimate` library, with committed references. The `.go.yaml` cases `diagram-source/unsupported-uml-types` and `styling/embedded-unsupported-type` are updated. The sprite cases are `@startuml` sequence diagrams, so the golden dumps are regenerated. The not-supported-yet count of the report stays at 5.
- **Documentation:** `README.md` and `site/index.html` (the row of description diagrams and the sentence on the unsupported types), the README's conformance counts, `NOTICE`, and `internal/klimt/HANDOFF.md`.
- **Compatibility:** inputs that failed with "archimate diagrams are not supported yet" or "drawing crashed" now give upstream's drawing or upstream's errors.
