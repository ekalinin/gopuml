# Spec Delta

## ADDED Requirements

### Requirement: Internal sprites

The system SHALL bundle the sprites of upstream's jar resources: the 116 SVG and 23 PNG files of `sprites/archimate/` in the plantuml-mit 1.2026.9 jar. It SHALL use them where upstream does:

- in a sprite reference `<$archimate/name>` that the diagram does not define;
- in a sprite stereotype `<<$archimate/name>>`;
- in the icon of an archimate element (see the description-diagram capability);
- in `sprite $name jar:archimate/name`.

An SVG sprite of the jar SHALL be drawn as vector shapes with upstream's default SVG sprite parser, whatever `!pragma svgparser` selects, in SVG and PNG output. A PNG sprite of the jar SHALL be drawn as a raster image. A reference to a name that has no sprite in the jar SHALL draw nothing, as for an undefined sprite, and `sprite $name jar:<name>` SHALL fail with `No such internal sprite: <name>`.

#### Scenario: Internal SVG sprite

- **WHEN** a sequence diagram contains `Alice -> Bob : <$archimate/actor> hi`
- **THEN** the message label draws the actor icon as vector paths before `hi`, as upstream does

#### Scenario: Internal SVG sprite with the SAX parser

- **WHEN** a sequence diagram contains `!pragma svgparser sax` and `Alice -> Bob : <$archimate/actor> hi`
- **THEN** the message label draws the actor icon with the same paths as without the pragma, as upstream does

#### Scenario: Internal PNG sprite

- **WHEN** a sequence diagram contains `Alice -> Bob : <$archimate/access> hi`
- **THEN** the message label draws the 16 by 16 pixel access icon before `hi`, as upstream does

#### Scenario: Internal sprite as a stereotype

- **WHEN** a diagram contains `rectangle "Customer" <<$archimate/business-actor>> as c`
- **THEN** the rectangle draws the business-actor icon above `Customer`, as upstream does

#### Scenario: Internal sprite of a sprite command

- **WHEN** a sequence diagram contains `sprite $a jar:archimate/actor` and `Alice -> Bob : <$a> hi`
- **THEN** the message label draws the actor icon before `hi`, as upstream does

#### Scenario: Missing internal sprite

- **WHEN** a sequence diagram contains `sprite $a jar:archimate/nope` and `Alice -> Bob : <$a> hi`
- **THEN** the diagram fails at the line `sprite $a jar:archimate/nope` with `No such internal sprite: archimate/nope (Assumed diagram type: sequence)`
