# Spec Delta

## Purpose

Draws sprites as PlantUML v1.2026.9 does: the sprites that a diagram defines with `sprite`, and the sprites and images of the standard library, in SVG and PNG output.

## ADDED Requirements

### Requirement: Sprites defined in a diagram

The system SHALL draw a sprite defined with `sprite` (monochrome sprites with 4, 8 or 16 gray levels, plain or compressed with `z`, color sprites with `/color`, image files, SVG files, and SVG given inline) wherever upstream draws a sprite reference, at the size, scale and colors that upstream uses, in SVG and PNG output. A monochrome, color or image sprite SHALL be drawn as a raster image (in SVG output, an `image` element with PNG data), and an inline SVG sprite as vector shapes. An SVG file sprite SHALL be embedded as an SVG image in SVG output and not drawn in PNG output, as upstream does. A reference to a sprite that the diagram does not define SHALL draw nothing, as upstream does.

#### Scenario: Monochrome sprite

- **WHEN** a sequence diagram defines `sprite $foo [4x4/16]` with the rows `FFFF`, `F00F`, `F00F` and `FFFF`, and contains `Alice -> Bob : <$foo> hi`
- **THEN** the message label draws a 4 by 4 pixel image before `hi`, as upstream does

#### Scenario: Color sprite

- **WHEN** a sequence diagram defines a sprite with `[16x16/color]` and 16 rows of 16 colors, each written as two characters of upstream's 4096-color palette, and a message label references it
- **THEN** the message label draws the sprite in its colors, as upstream does

#### Scenario: SVG sprite

- **WHEN** a sequence diagram defines `sprite foo <svg viewBox="0 0 16 16"><circle cx="8" cy="8" r="8" fill="#FF0000"/></svg>` and contains `Alice -> Bob : <$foo> hi`
- **THEN** the message label draws a red circle of radius 8 before `hi` (in SVG output, an `ellipse` element), as upstream does

#### Scenario: Undefined sprite

- **WHEN** a sequence diagram defines no sprite and contains `Alice -> Bob : <$nope> hi`
- **THEN** the message label draws only `hi`

### Requirement: SVG sprite parser

The system SHALL parse inline SVG sprites with upstream's default SVG sprite parser, and with upstream's SAX parser when the diagram contains `!pragma svgparser sax`, so that each pragma value draws the shapes that upstream draws with it. The SVG sprites of the standard library SHALL always use the default parser, as upstream does.

#### Scenario: SAX parser

- **WHEN** a sequence diagram contains `!pragma svgparser sax`, defines `sprite foo <svg viewBox="0 0 16 16"><circle cx="8" cy="8" r="8" fill="#FF0000"/></svg>` and contains `Alice -> Bob : <$foo> hi`
- **THEN** the message label draws the red circle as a path of arcs before `hi`, as upstream does

### Requirement: Standard library sprites

The system SHALL define a sprite from the standard library with `sprite $name #library#sprite` (a monochrome sprite of the library) and `sprite name :library:sprite` (an SVG sprite of the library), as the files of the bundled libraries do. A library that is not bundled SHALL produce upstream's error for a missing library at the command line.

#### Scenario: Icon of a library

- **WHEN** a sequence diagram contains `!include <tupadr3/font-awesome/server>` and `Alice -> Bob : <$server> hi`
- **THEN** the message label draws the 48 by 48 pixel server icon before `hi`, as upstream does

#### Scenario: SVG icon of a library

- **WHEN** a sequence diagram contains `!include <material7/AccountUser/account>` and `Alice -> Bob : <$mdiAccount> hi`
- **THEN** the message label draws the account icon as vector paths before `hi`, as upstream does

#### Scenario: Sprite of a missing library

- **WHEN** a sequence diagram contains `sprite $x #foo#bar`
- **THEN** the system reports an error at that line containing `java.io.UncheckedIOException: java.io.FileNotFoundException: stdlib/foo/info.spm (No such file or directory)`

### Requirement: Standard library images

The system SHALL draw `<img data:image/png;spm library/N>` in creole text as image `N` of the library's images, with the `{scale=...}` option, as upstream does. A reference without `/` SHALL draw the text `ERROR DATA IMAGE`, and a missing library or image number SHALL draw upstream's `ERROR` text.

#### Scenario: Image of a library

- **WHEN** a sequence diagram contains `Alice -> Bob : <img data:image/png;spm awslib14/0> hi`
- **THEN** the message label draws image 0 of `awslib14` (the Analytics icon) before `hi`, as upstream does
