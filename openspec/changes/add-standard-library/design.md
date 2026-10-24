# Design

## Context

See proposal.md for the motivation. The current state that shapes the approach:

- Upstream does not ship `.puml` files. Each library is a folder `stdlib/<name>/` of channel files, `info.spm`, `puml.spm`, `json.spm`, `sprite.spm`, `svg.spm` and `image.spm`, each one Brotli stream (`SpmChannel`). `info.spm` holds `key=value` lines. The other channels start with a count and hold Java `DataOutputStream` records: `puml` and `json` a name (`writeUTF`) and bytes, `sprite` a name, a width, a height and packed 4-bit gray levels, `svg` a name and an SVG text, `image` a color table and then a width, a height and 2-byte color indices per image. `stdlib/home.spm` lists the folder names. The jar also has `.sha1` files and a `README.md`, which no code reads.
- In 1.2026.9, `home.spm` lists 34 names. `awslib`, `bootstrap`, `material`, `material2` and `material7` have only an `info.spm` whose `link` key names another folder. The channel files total 8.9 MB (237 MB decoded). The largest decoded channel is `ibm/image.spm`, 74 MB. The `stdlib/` files of the sources jar are byte-identical to those of the binary jar, and CI already unpacks the sources jar into `.cache/upstream/src` for `make provenance`.
- The library files refer to the other channels: `sprite $person #c4#c4_context:person` (`CommandSpriteStdlib`), `sprite mdiAccount :material7.4.47:accountuser/account:mdiAccount` (`CommandSpriteStdlibSvg`), `<img data:image/png;spm awslib14/0{scale=...}>` (`AtomImg`), `%load_json('<office/all.json>')` and `!theme C4_blue from <C4/themes>`. Some include other libraries: `azure` includes `archimate`, `c4` includes `office`, `domainstory` includes `material2.1.19` and `k8s` includes `c4`. All of them are libraries with a verified license.
- gopuml already ports the `Stdlib` index in `internal/tim/stdlib.go` (lookups, metadata in `HashMap` order, `%get_stdlib`, `%get_all_stdlib`, `-stdlib`, the stdlib `PathSystem` folders, themes and JSON), over an in-memory map that `NewStdlib` builds for tests. `embeddedStdlib` is empty, and `tcontext.go` reports "not bundled" for any library missing from it.
- No sprite is drawn: `SpriteDef.AsTextBlock` in `internal/uml/sprite_decode.go` panics, although the sprite commands already decode every kind. The stdlib sprite commands of `internal/uml/sprites.go` always throw upstream's missing-library error. `AtomImgCreate` panics for `data:image/png;spm`. An undefined `<$name>` is drawn as literal text, where upstream draws nothing.
- Upstream draws a monochrome, color or image sprite as a `UImage`, which the SVG driver writes as a PNG data URL encoded with Go's `image/png` (a documented exception of `internal/klimt/svg`). Inline SVG sprites go through `SvgSpriteParserFactory` (`SvgNanoParser` by default, `SvgSaxParser` with `!pragma svgparser sax`) and draw shapes; SVG files go through `SpriteSvg`, a `UImageSvg` that the PNG driver ignores. The stdlib SVG channel always uses the factory without a pragma, so Nano. The parsers use `emoji.ColorResolver`, `emoji.GrayLevelRange`, `emoji.UGraphicWithScale` and `openiconic.SvgPath`, none of which is ported.
- Package graph: `internal/tim` imports neither `internal/klimt` nor `internal/uml`, and neither of them imports `internal/tim`. The sprite commands (`uml`) and `AtomImg` (`klimt`) need the stdlib as much as the preprocessor does.
- The foundation design D14 bundles themes only when their license is verified; ten themes are excluded and listed in NOTICE. Release archives carry `LICENSE`, `NOTICE` and `README.md` only.
- Repo-level license check of the 26 source repositories of 1.2026.9 (`gh api`, followed by the license files):

| Libraries | Source repository license |
|---|---|
| `archimate`, `aws`, `azure`, `bootstrap`, `bootstrap1.12.1`, `bootstrap1.13.1`, `c4`, `cloudinsight`, `domainstory`, `eip`, `k8s`, `logos`, `office`, `tupadr3` | MIT |
| `kubernetes` | Apache-2.0 |
| `awslib`, `awslib10`, `awslib14`, `awslib20` | icons CC BY-ND 2.0, code MIT (`LICENSE`, `LICENSE-CODE`) |
| `ibm` | CC BY-ND 2.0 |
| `material`, `material2`, `material2.1.19`, `material7`, `material7.4.47` | Pictogrammers Free License: icons Apache-2.0, code MIT |
| `adaml` | GPL-3.0 (also `license=GPL` in its `info.spm`) |
| `classy`, `classy-c4`, `cloudogu`, `edgy`, `elastic`, `gcp`, `osa`, `osa2` | no license file; the READMEs read so far state none for the PlantUML files (`osa` cites a share-alike license for the original icons, `cloudogu` a trademark notice) |

## Goals / Non-Goals

**Goals:**

- The libraries render as upstream renders them, through the curated cases and the conformance criteria that apply to every case (texts, size within 5%, SSIM).
- The bundled bytes are upstream's, checked in CI, so that a later upstream update is an import and a review of the check's output.
- No new dependency, and the existing Go names kept where the behavior they name does not change.

**Non-Goals:**

- Reducing the memory of decoded channels below upstream's (D7).
- Encoding images exactly as Java's ImageIO does: the PNG data in SVG output stays Go's encoding, as for every raster image today.

## Decisions

### D1. Embed the channel files unchanged

- Layout: `internal/stdlib/<name>/<channel>.spm` for each bundled library and `internal/stdlib/home.spm`, embedded with `//go:embed`. The `.sha1` files and `README.md` are not imported.
- `home.spm` keeps the lines of upstream's file, in its order, without the excluded names, so that the ported `getAllFolderNames` reads it unchanged.
- A tool `tools/stdlib` imports the files from the unpacked sources jar (`-src .cache/upstream/src`, after `tools/upstream/fetch.sh`), skipping the excluded libraries, behind a `make stdlib` target.
- The check is a test of `internal/stdlib`, `TestUpstreamStdlib`, as `TestUpstreamThemes` checks the themes: it reads `.cache/upstream/src/stdlib`, skips when the sources are not unpacked, and fails when a bundled file differs, an upstream folder is neither bundled nor excluded, an excluded folder is bundled, or `home.spm` differs from the filtered upstream list. CI's `make test` runs it, since CI fetches the sources before testing. The comparison is a function over two `fs.FS`, so a unit test covers each failure with in-memory files.
- Alternatives: decode at import time and store gzip, readable with the Go standard library (rejected: 19.2 MB instead of 8.7 MB, and no longer upstream's bytes to check); fetch the library at build or run time (rejected: `go install` and `go:embed` need the files in the module, and a runtime fetch breaks offline use and the sandbox).

### D2. License exclusions

- `internal/stdlib` holds the excluded names with their reason, the single list that the import tool skips, the check enforces and the include check reads (D5). `adaml` is excluded for GPL-3.0.
- `classy`, `classy-c4`, `cloudogu`, `edgy`, `elastic`, `gcp`, `osa` and `osa2` are checked as the themes were: their repository's README and files, the library's own files in the jar, and upstream's history (the commits that added them). A library is bundled when a license that allows redistribution is found, and excluded with the reason otherwise.
- CC BY-ND 2.0 (`awslib*`, `ibm`) allows verbatim redistribution with attribution. The files are shipped as PlantUML distributes them, and NOTICE gives the attribution and the license URI.
- NOTICE gains a "Bundled standard library" section, after "Bundled themes": per library, its name, version, source and license, with the texts that a binary distribution must carry (the MIT notices, the Apache-2.0 notice, the CC license URI), and the excluded libraries with their reason. License texts that NOTICE does not reproduce go to `internal/stdlib/LICENSE-*.txt`, as for the themes. "Excluded upstream components" names the excluded libraries.
- Alternatives (asked): bundling every library but `adaml`, or all 34 as upstream. Rejected by the user for consistency with D14.

### D3. Port upstream's Brotli decoder

- `internal/brotli` ports `net/sourceforge/plantuml/brotli` (`Decode`, `State`, `BitReader`, `Huffman`, `Context`, `Transform`, `Dictionary`, `DictionaryData`, `Utils`, `BrotliInputStream`), with `plantuml-mit 1.2026.9` provenance headers, as an `io.Reader` over a byte slice. The dictionary keeps upstream's encoded form, decoded once.
- Checked by unit tests that decode every bundled channel and compare it with the `brotli` CLI's output (a test fixture generated once), and by the conformance cases.
- Alternatives: `github.com/andybalholm/brotli` (rejected: a fourth dependency and a NOTICE entry, where the project translates the MIT sources); recompression (D1).

### D4. A leaf package for the library data

- `internal/stdlib` ports `Stdlib` (`retrieve` with `link`, the process-wide map of libraries), `SpmChannel` and the record formats, and returns raw data: info lines, `puml` and `json` bytes by lower-case name, a monochrome sprite as width, height and packed levels, an SVG sprite as text, an image as width, height and ARGB pixels (the decoding of `FutureImage.getNow`). It imports only `internal/brotli`, so `tim`, `uml` and `klimt` can all import it.
- A missing library or channel returns an error that carries Java's message (`java.io.FileNotFoundException: stdlib/<name>/info.spm (No such file or directory)`). Callers throw it as their upstream class does, as `retrieve`, `stdlibNotBundled` and `getJSONResource` do today.
- `tim.Stdlib` keeps its type, its methods and `NewStdlib` for tests (minimal diff). `embeddedStdlib` becomes an index over `internal/stdlib`; `parseStdlibInfo` keeps the metadata in `HashMap` order for `%get_stdlib`, and `getPumlResource` returns nil for a library without metadata, as upstream does.
- Alternative: move all of `tim.Stdlib` into `internal/stdlib` (rejected: changes every caller in `internal/tim` and the test helpers for no behavior change).

### D5. Includes and their errors

In `tcontext.go`, a `<library/path>` include of an excluded library reports `notBundledMessage` (unchanged text). Any other library goes to `retrieve`: a library that upstream does not ship throws `UncheckedIOException`, which the existing path reports as `Fatal parsing error`; a missing file in a bundled library reports `cannot include <library/path>`. The sprite commands and `%get_stdlib` treat an excluded library as missing, as upstream would without it.

### D6. Sprite drawing

- The drawing ports go into `internal/klimt` with the `sprite_` prefix (D2 of the foundation design): `SpriteMonochrome.asTextBlock` (gray levels mapped between the back color and the font or forced color, as a `UImage`), `SpriteColor`, `SpriteImage`, `SpriteSvg` (a `UImageSvg`), `StdlibSprite` (16-level monochrome from packed data) and `StdlibSpriteSvg` (parsed on first use).
- `uml.SpriteDef` keeps the decoding and its kinds; `AsTextBlock` delegates to the klimt drawing. Inline SVG sprites are parsed when the command runs, with the diagram's pragma and MD5 map, as `CommandSpriteSvg` and `CommandSpriteSvgMultiline` do, and SVG file sprites keep `SpriteSvg`, so the PNG output ignores them as upstream does.
- The SVG parsers form a cycle with `klimt` (they draw klimt shapes; `klimt.sprite` and `Stdlib` create them), so they join `internal/klimt` under prefixes: `svgparser_` (`ISvgSpriteParser`, `SvgSpriteParserFactory`, `SvgNanoParser`, `SpriteSvgNanoParser`, `SvgSaxParser`), `emoji_` (`ColorResolver`, `GrayLevelRange`, `UGraphicWithScale`) and `openiconic_` (`SvgPath`, `SvgCommand*`, `SvgPosition`, `Movement`). The OpenIconic icon set and emoji drawing stay out (proposal).
- `SvgSaxParser` reads SVG with Go's `encoding/xml` tokenizer in place of Java's SAX parser, reporting the same elements and attributes for well-formed SVG. Malformed SVG follows upstream's handling of a SAX error.
- The stdlib commands of `sprites.go` read `internal/stdlib`. `AtomImgCreate` draws `data:image/png;spm library/N` from it, with `ERROR DATA IMAGE` and `ERROR <exception>` texts as upstream prints them (a missing library, a library without images, a number that does not parse, an index out of range).
- An undefined `<$name>` draws nothing, as upstream: the literal-text fallback of the creole parser is removed.
- Checked against upstream through Java golden harnesses of the drawing calls (as `KlimtGolden` does for other shapes) for the sprite kinds and both parsers, and through the curated cases.

### D7. Decoding on first use, kept for the process

- Each (library, channel) is decoded once, on first use, under a `sync.Once`, and kept for the process lifetime, like upstream's static map and the per-channel maps of `Stdlib`. Concurrent renders of the HTTP server and `--threads` share it.
- Alternative: drop decoded channels after each render (rejected: a C4 diagram would decode the `c4` and `office` channels again on every request; the worst case of keeping everything is upstream's, 237 MB if every channel of every bundled library is used).

### D8. WebAssembly

The module embeds the same files (user decision), so `cmd/plantuml-wasm` needs no change. Standard library includes do not go through the host file access, so the `SANDBOX` profile allows them, as upstream does. The module size, measured in CI and not gated (foundation D12), grows from 37.8 MB to about 46.5 MB; the landing page loads the larger module.

### D9. Conformance cases

- A curated `.puml` case for each scenario that matches upstream, rendered by `make reference`: the library include, the case-insensitive name, the missing file and the missing library, the alias include and version, `%get_stdlib("c4", "version")`, the JSON resource, the library theme, the sprite kinds, both SVG parsers, the undefined sprite, the stdlib sprites, the stdlib image, the sprite of a missing library, and the sprite in a note.
- A `.cli.yaml` case for the include in the sandbox.
- `.go.yaml` cases where gopuml differs from upstream on purpose or the scenario is not a render: the excluded library (`preprocessor/library-not-bundled`, moved from C4 to `adaml`), `%get_all_stdlib()` and `-stdlib` (no excluded names), the modified-file check of `TestUpstreamStdlib`, and the WASM standard library include (as `render-matches-native`).
- The color sprite data comes from upstream's `-encodesprite color` of a 16 by 16 PNG, committed with the case.

## Risks / Trade-offs

- [Memory: drawing one `ibm` image decodes and keeps the 74 MB image channel, as upstream does] → Accepted for parity (D7). A later change can keep only the decoded images that renders use, if the server or the WASM module needs it.
- [Latency of the first use of a large channel] → The decoder runs once per channel; the tests measure the decoding of the largest channels and the result is recorded in the task.
- [Repository and binary size: +8.7 MB of binary files in git, each upstream update adds more history; the CLI binaries and the WASM module grow by 8.7 MB] → Accepted (user decision for WASM). The files change only with an upstream update.
- [A license found later to be incompatible] → The excluded list and NOTICE are the single place to change; the check enforces the list.
- [Library macros that need features gopuml lacks, such as the `archimate` keyword (#35)] → Out of scope (proposal); such diagrams report what they report today, and the curated cases use libraries whose output gopuml draws.
- [`encoding/xml` and Java's SAX parser disagree on edge cases (entities, DTDs)] → The SAX parser runs only with `!pragma svgparser sax`; the golden harness covers the SVG features of the stdlib.

## Migration Plan

No migration: libraries that failed now render. Diagrams that relied on the "not bundled" error for a library that is now bundled, or on literal text for an undefined `<$name>`, change output, as listed in the proposal's compatibility notes. Rollback is a revert.

## Open Questions

- Which of `classy`, `classy-c4`, `cloudogu`, `edgy`, `elastic`, `gcp`, `osa` and `osa2` have a verifiable license (D2). The answer changes the excluded list and NOTICE, not the specs or the tasks.
