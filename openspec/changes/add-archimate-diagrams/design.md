# Design

## Context

See proposal.md for the motivation. The current state that shapes the approach:

- Upstream ports:
  - `CommandArchimate` and `CommandArchimatePackage` (`SingleLineCommand2<DescriptionDiagram>`) and `CommandArchimateMultilines` (`CommandMultilines2<AbstractEntityDiagram>`), in `descdiagram/command`. `DescriptionDiagramFactory` registers them after `CommandCreateJsonSingleLine` and before `CommandCreateDomain`. No other factory registers them.
  - `SpriteImage.fromInternal`, which reads `/sprites/<name>.svg` from the jar and parses it with `SvgSpriteParserFactory.create(String)`, that is the Nano parser without a pragma. Otherwise it reads `/sprites/<name>.png` as a `SpriteImage`. It returns null for a missing name and for any exception. Names ending with `.png` or `.svg` throw `IllegalArgumentException`.
- What Go already has:
  - in `internal/descdiagram`: the factory (`factory.go`, whose comment says the archimate commands are left out), `commandCreateElementFull` (`commands_createelement.go`) and `commandCreateDomain`
  - in `internal/abel`: the element patterns `ElementCode`, `ElementCodeWithQuote`, `ElementDisplay` and `ElementDisplayWithoutQuote`, `CommandCreateElementMultilines`, `SpottedStereotype`, `CucaDiagram.QuarkInContext`, `ReallyCreateLeaf` and `GotoGroup`, and the symbols `USymbolArchimate` and `USymbolRectangle`
  - in `internal/klimt`: `StereotypePatternOptionalArchimate`, the stereotype label of an `archimate/` sprite, `SkinParam.Sprite`, which falls back to `SpriteImageFromInternal`, `SvgSpriteParserFactoryCreateDefault`, `NewSpriteImage` and `PortableImageFromImage`
  - the ArchiMate colors in `internal/klimt/color`, the style name `archimate`, the skin parameters `archimateBorder` and the archimate corner, and the archimate symbol in `internal/decoration/symbol`
  - in `internal/uml`: the `sprite $name jar:<name>` branch of the sprite command, which calls `SpriteImageFromInternal` and reports `No such internal sprite: <name>` for nil
- What is missing:
  - the sprite files: `SpriteImageFromInternal` panics with "not ported yet" for the 139 names of `sprite_internalnames.go` and returns nil for the others
  - the three commands
- `internal/core/detect.go` lists one unsupported `@startuml` type, archimate, with the pattern `^archimate[%s]`. Its `Provenance` names `CommandArchimate.java`.
- The unpacked sources jar has the sprites in `.cache/upstream/src/sprites/archimate/`: 116 SVG files (206 KB) and 23 PNG files (5 KB), 16 by 16 pixels, two of them with a 1-bit color map. Upstream's file headers credit them to Archi (https://www.archimatetool.com), whose repository is MIT-licensed (Copyright (c) 2013-2026 Phillip Beauvoir, Jean-Baptiste Sarrodie, The Open Group).
- A spike rendered each of the 116 SVG files as an inline `sprite` with the ported parser: all 116 draw without error.

## Goals / Non-Goals

**Goals:**

- Port `SpriteImage.fromInternal` with the jar's sprites, and the three archimate commands, with upstream's output and errors.
- Compare every bundled sprite with upstream's drawing, in SVG and PNG.
- Keep the results of the other diagram types unchanged.

**Non-Goals:**

- The `explainArg` and `explainNow` methods of the three commands: they feed `@Explain`, which is reported as not supported yet before the factories run.
- The `TeaVM.isTeaVM()` branch of `fromInternal`, which returns null in upstream's JavaScript build: the WASM build is not TeaVM and has the sprites.
- Jar resources outside `sprites/`. Upstream's class loader may resolve a `jar:` name with `..` to another resource of the jar. Here such a name is not a valid embedded path and gives nil, so the sprite command reports `No such internal sprite`.
- `PSystemListArchimateSpritesFactory` (issue #57).

## Decisions

### D1. Package layout

- `internal/klimt/sprites/archimate/` holds the 139 files of the jar, copied unchanged. `internal/klimt/sprite_internal.go` embeds them (`//go:embed sprites`) and holds `SpriteImageFromInternal`, with the `Provenance` of `klimt/sprite/SpriteImage.java`. The function moves there from `sprite_sprite.go`. `sprite_internalnames.go` is deleted: the embedded files replace the list of names.
- `internal/descdiagram/commands_archimate.go` ports the three commands as `commandArchimate`, `commandArchimateMultilines` and `commandArchimatePackage`, with a `Provenance` line for each Java file. `factory.go` registers them in upstream's order, after `abel.CommandCreateJSONSingleLine` and before `commandCreateDomain`, and its comment no longer leaves them out.
- `internal/core/detect.go` keeps the detection with an empty list (D5).

Alternative: `CommandArchimateMultilines` in `internal/abel`, generic over `Diagram`, next to `CommandCreateElementMultilines`, since upstream types it with `AbstractEntityDiagram`. Only the description factory registers it, so it stays with the other two in `internal/descdiagram`.

### D2. Internal sprite lookup

`SpriteImageFromInternal(name)` follows `fromInternal`:

1. A name ending with `.png` or `.svg` panics with `IllegalArgumentException`, as today.
2. `sprites/<name>.svg`, when embedded, gives `SvgSpriteParserFactoryCreateDefault` of its text. That is the Nano parser, whatever `!pragma svgparser` says, since upstream passes no pragma.
3. Otherwise `sprites/<name>.png`, when embedded, is decoded with `image.Decode` and gives `NewSpriteImage(PortableImageFromImage(img))`. A file that does not decode gives nil, as upstream's `catch (Throwable)`.
4. Otherwise nil.

Each call reads and parses the file, as upstream does. The embedded files are in memory and the sprites are small.

Alternative: a cache of the parsed sprites by name. Upstream has none, and a sprite shared by concurrent renders would have to be safe for concurrent drawing. A cache can come later if `make bench-compare` shows a cost.

### D3. Bundled files match the jar

`TestUpstreamInternalSprites` in `internal/klimt`, as `TestUpstreamStdlib` and `TestUpstreamThemes`, compares the embedded `sprites/` tree with `.cache/upstream/src/sprites/`. The same names and the same bytes are expected. The test skips when the sources are not unpacked, and CI unpacks them before the tests. NOTICE gets a paragraph under "Bundled assets" naming `internal/klimt/sprites/`, its origin (the `sprites/` folder of the plantuml-mit jar), Archi and its MIT License with the copyright line. Archi has no NOTICE file to reproduce.

### D4. The archimate commands

- `commandArchimate` builds upstream's pattern:
  - `archimate`, one or more spaces, `color.SimpleColor(color.TypeBack)`'s optional color, one or more spaces;
  - the four name forms (`CODE1` to `CODE4`, `DISPLAY2` to `DISPLAY4`, with `StereotypePatternOptionalArchimate` before `as`);
  - the optional `STEREOTYPE`, then the end of the line.

  The two space runs around an optional color keep upstream's rule that one space without a color is a syntax error. Execution follows `executeArg`:
  - `QuarkInContext(true, CleanID(code))`;
  - the display without quotes, or the quark's name;
  - a new leaf of type description with `USymbolArchimate` when the quark has none;
  - `SetDisplay`, then the icon: the chevrons removed and `SpottedStereotype` of `<<$archimate/<icon>>>`;
  - `SetColors` of the simple color.
- `commandArchimateMultilines` is a `uml.Multilines2` with `uml.RemoveStartingQuote` and both ends trimmed. Its start pattern:
  - `archimate`, the two space runs around the color, `CODE` (`[%pLN_.]+`), the archimate stereotype;
  - the optional URL, `color.Exp1()`, `[` and the rest of the line.

  Its end pattern is `^(.*)\]$`. Execution follows `executeNow`:
  - the lines trimmed, `QuarkInContext(false, ...)`, and `Already exists <name>` when the quark has an entity;
  - a leaf with `USymbolRectangle`, whose display is the lines between the first and the last (`SubExtract(lines, 1, 1)`);
  - the icon and the simple color of the first line.

  The URL, the text after `[` and the text before `]` are parsed and not used, as upstream.
- `commandArchimatePackage` has the single-line pattern followed by optional spaces and `{`. Execution follows `executeArg`:
  - `GotoGroup` with `GroupTypePackage` and `USymbolArchimate`, whose error is returned;
  - then `SetDisplay`, the icon and the colors of the current group.

### D5. The unsupported @startuml types

`unsupportedTypes` becomes an empty list, with a comment that no `@startuml` type is reported as not supported yet in this release. The `patterns` helper and the `regex` import, which only the archimate entry used, are removed, so that the linter does not report unused code. The header of `detect.go` becomes `Provenance: original`, since no upstream pattern is left in it. `TestDetectUnsupportedUmlTypes` keeps its state and legacy activity inputs, expects the description type for the archimate input, and checks the mechanism with a test entry that it sets in `unsupportedTypes` for its duration. `TestEmbeddedUnsupportedType` uses `{{json`, `{"a": 1}`, `}}` and expects "json diagrams are not supported yet". The `why` of both `.go.yaml` cases is updated.

Alternative: removing the detection and its requirement. The user chose to keep the mechanism: the diff is smaller, and the mechanism can serve a later `@startuml` type.

### D6. Conformance cases

- **description-diagram:** one curated case per scenario of "Archimate elements", with `type: description` and `formats: [svg, png]`:
  - `archimate-element`, `archimate-element-updated`, `archimate-multiline`, `archimate-group` and `archimate-unknown-icon`;
  - the two error cases, with `formats: [svg]`: `archimate-without-color` (`type: sequence`, the assumed type) and `archimate-multiline-twice` (`type: description`).
- **sprites:** one case per scenario of "Internal sprites":
  - `internal-svg-sprite`, `internal-svg-sprite-sax`, `internal-png-sprite` and `internal-sprite-command`, sequence diagrams with `formats: [svg, png]`;
  - `internal-sprite-stereotype`, a description diagram;
  - `missing-internal-sprite`, with `formats: [svg]`.
- **standard-library:** `archimate-library`, a description diagram with `formats: [svg, png]`.
- **All sprites:** `sprites/internal-sprites-all`, without scenarios and with `formats: [svg, png]`. It is a note of a sequence diagram with one line `<$archimate/<name>> <name>` per bundled sprite, so that every file is compared with upstream's drawing, including the two PNG files with a color map.
- **Golden dumps:** the new sequence cases are `@startuml` cases, so the golden dumps of `internal/sequencediagram`, `internal/classdiagram` and `internal/activitydiagram3` are regenerated.

### D7. Documentation

- The row "Use case, component and deployment" of the tables of supported diagrams in `README.md` and `site/index.html` becomes "Use case, component, deployment and archimate". `TestLandingPage` compares the two tables. `TestReleaseNotes`, which looks for the rows of the release notes in the README, maps the old wording of the row to the new one, as it does for the other rows reworded since a release.
- The sentence "Other diagram types (archimate, JSON, and so on)" of both files becomes "Other diagram types (JSON, YAML, and so on)".
- The README's counts follow `conformance-report.md`.
- `internal/klimt/HANDOFF.md` no longer says that the jar's sprites are not bundled.

Alternative: leaving the tables as they are, since upstream counts archimate as part of description diagrams. A reader looking for archimate would then not find it in the table.

## Risks / Trade-offs

- [The ported Nano parser or the PNG decoding draws an internal sprite differently from upstream] → `sprites/internal-sprites-all` compares every file in SVG and PNG. A sprite that still differs after a fix of the parser goes to `known-differences.yaml` with its reason.
- [The binary and the WASM module grow by about 211 KB] → small next to the module's 52 MB. The files stay as the jar has them, as the stdlib files do, so that D3 compares them byte for byte.
- [Parsing an SVG sprite at each lookup costs time in large archimate diagrams] → upstream does the same. `make bench-compare` on the pull request shows the cost of the other diagram types, which do not look up internal sprites unless they reference one.
- [The empty list leaves `detectUnsupportedUmlTypes` without a production entry] → its test sets a test entry, so a broken mechanism still fails.
