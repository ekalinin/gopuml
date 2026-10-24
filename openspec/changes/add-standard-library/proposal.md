# Proposal

## Why

gopuml bundles no standard library: `embeddedStdlib` in `internal/tim/stdlib.go` is empty, so every `!include <library/path>` fails with "library ... is not bundled" where upstream draws the diagram (issue #60, for example `!include <C4/C4_Container>` and `!include <awslib/AWSCommon>`). The libraries also need sprites, which gopuml cannot draw: any defined sprite referenced in a text crashes the render with "not ported yet: klimt/sprite/SpriteMonochrome (drawing of sprites)". C4's `Person()` draws the `<$person>` sprite of the C4 library, so bundling the files alone would not render the issue's example.

## What Changes

- Bundle the standard library of the plantuml-mit 1.2026.9 jar (`stdlib/`): for each bundled library, the Brotli-compressed channel files (`info.spm`, `puml.spm`, `json.spm`, `sprite.spm`, `svg.spm`, `image.spm`) embedded unchanged, and the `home.spm` list of names. A new tool imports them from the pinned jar, and CI checks that they match it. 25 of the 34 entries of 1.2026.9 total 8.7 MB.
- Bundle only libraries whose license is verified to allow redistribution, as the themes do (design D14 of the foundation change, user decision). `adaml` (GPL-3.0) is excluded. `classy`, `classy-c4`, `cloudogu`, `edgy`, `elastic`, `gcp`, `osa` and `osa2` have no license file in their source repositories: each is bundled only if a license is found in its README, its files or upstream's history, and excluded otherwise. NOTICE lists the bundled libraries with their attributions and licenses, and the excluded ones with the reason.
- Port upstream's Brotli decoder (`net/sourceforge/plantuml/brotli`), which reads the channel files. No new dependency.
- Port `Stdlib` and `SpmChannel` in full: libraries are matched in lower case, the `link` key of `info.spm` makes a name an alias of another library (`awslib` of `awslib14`, `bootstrap` of `bootstrap1.13.1`, `material` and `material2` of `material2.1.19`, `material7` of `material7.4.47`), and each channel is decoded on first use and kept.
- Resolve `!include <lib/path>` (and `!include_many`, `!include_once`, `!includesub`) from the bundled library, as upstream does. A file missing from a bundled library reports `cannot include <lib/path>`. A library that upstream does not have either reports upstream's error. A library that this build excludes keeps the "not bundled" error.
- `-stdlib`, `%get_all_stdlib()` and `%get_stdlib()` list the bundled libraries with their version and source; `%load_json` reads the JSON resources and `!theme <name> from <lib/dir>` the themes of a library.
- Draw sprites (user decision: all sprite kinds): port the drawing of `SpriteMonochrome`, `SpriteColor`, `SpriteImage` and `SpriteSvg`, the SVG sprite parsers `SvgNanoParser` (the default) and `SvgSaxParser` (`!pragma svgparser sax`) with the emoji helpers they use, the stdlib sprites `sprite $name #lib#sprite` (`StdlibSprite`) and `sprite name :lib:sprite` (`StdlibSpriteSvg`), and the stdlib images `<img data:image/png;spm lib/N>` (`FutureImage`). Sprites defined in a diagram also start to draw.
- The WebAssembly module embeds the same standard library (user decision), so its output stays identical to the native build's. `plantuml.wasm` grows by about 8.7 MB, from 37.8 MB.
- Curated conformance cases for the new scenarios, among them C4 container and context diagrams, an AWS icon, a tupadr3 icon, a material7 SVG icon, the sprite kinds and the include errors. The `preprocessor/library-not-bundled` case moves from C4 to `adaml`.

Out of scope:

- The internal sprites of the jar's `/sprites/` folder (the archimate icons that the `archimate` keyword draws, issue #35). They keep reporting "not ported yet".
- Emoji (`<:name:>`) and OpenIconic (`<&name>`) references, which stay literal text, although emoji drawing upstream uses the same SVG parser.
- Diagram features that the macros of a library produce and that gopuml does not support, for example the `archimate` keyword (#35). Such diagrams report what they report today for these features.
- `@startsprite` and `listsprites` (#55), and `-encodesprite`.

## Capabilities

### New Capabilities

- `standard-library`: the bundled libraries of upstream 1.2026.9, the license exclusions, library aliases, and the metadata and resources that `-stdlib`, `%get_all_stdlib`, `%get_stdlib`, `%load_json` and stdlib themes read
- `sprites`: drawing of sprites defined in a diagram (monochrome, color, image and SVG), the choice of SVG parser, and the sprites and images of the standard library

### Modified Capabilities

- `preprocessor`: "Standard library includes" resolves includes from the bundled libraries, with upstream's errors for a missing file and a missing library and the "not bundled" error for an excluded library
- `styling`: "Creole text markup" draws sprite references (`<$name>`); OpenIconic and emoji references stay literal text
- `wasm-build`: "Sandboxed resource access" also resolves standard library includes, from the embedded library, as upstream allows them in the `SANDBOX` profile

## Impact

- **Code:** new `internal/brotli` (decoder) and `internal/stdlib` (embedded files, channel reader, exclusions); `internal/tim` (`stdlib.go`, `stdlib_info.go`, the include check of `tcontext.go`); `internal/uml` (`sprites.go` and `sprite_decode.go`: stdlib sprite commands, drawing); `internal/klimt` (sprite drawing, SVG parsers and emoji helpers under file prefixes, the stdlib images of `AtomImg`).
- **Data:** about 8.7 MB of binary files under `internal/stdlib/`, committed so that `go install` and `go:embed` find them. The CLI binary and the WebAssembly module each grow by about 8.7 MB.
- **Tooling:** a new import tool for the standard library (`make stdlib`) and a test that compares the bundled files with the unpacked sources jar in CI; new curated cases and references (`make reference`).
- **Docs:** NOTICE (bundled and excluded libraries), README (standard library support).
- **Compatibility:** `!include <lib/...>` of a bundled library now succeeds, and an unknown library reports upstream's "Fatal parsing error" instead of "not bundled". Defined sprites draw instead of crashing. `-stdlib` and `%get_all_stdlib()` are no longer empty.
- **Dependencies:** none new.
