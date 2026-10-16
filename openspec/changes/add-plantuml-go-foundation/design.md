# Design

## Context

- The repository is empty apart from OpenSpec scaffolding. See proposal.md for motivation and scope.
- Local toolchain:
  - Go 1.27.1 (darwin/arm64)
  - Graphviz 15.0.0
  - Docker 29.7.1
  - no Java runtime
- Upstream target: PlantUML v1.2026.8 (released 2026-09-05), MIT variant. Upstream facts that shape the approach:
  - **Size:** ~3,300 Java files and ~447k lines. Per-package line counts for this change's scope:
    - drawing layer `klimt`: 47.7k
    - activity v3: 33k
    - sequence, including Teoz: 24.7k
    - Graphviz bridge `svek`: 22.5k
    - preprocessor `tim`: 13.9k
    - command framework: 10.9k
    - skin: 8.9k
    - style: 6.1k
    - class diagram: 4k
  - **Translation baseline:** the MIT sources are published as `net.sourceforge.plantuml:plantuml-mit:1.2026.8:sources` on Maven Central. They are produced from the main tree with the `sjpp` preprocessor and `__MIT__` define, which drops ditaa, jcckit, sudoku, zxing and webp.
  - **Upstream design patterns:**
    - Parsing is line-oriented regex commands, tried by diagram factories in a fixed order (`PSystemBuilder`).
    - Graph diagrams write a DOT string, run `dot -Tsvg`, parse coordinates back from the SVG, and draw everything themselves.
    - Many process-wide singletons (`GlobalConfig`, the file system's current dir, the security profile).
  - **Text measurement** uses Java2D `FontMetrics` on the logical font `SansSerif`. That resolves to DejaVu Sans on Linux with `fonts-dejavu`. Upstream's own Dockerfile uses `eclipse-temurin:21-jre-jammy`, `fonts-dejavu` and Graphviz 14.0.1 built from source.
  - **Themes:** 44 bundled themes. Their license headers: 31 MIT, 3 Apache-2.0, 1 GPL-3+ (`sunlust`). The other 9 have no license value: `_none_` has no header, and 8 (amiga, blueprint, carbon-gray, crt-amber, crt-green, mimeograph, mono, plain) have an empty `license:` field.
  - **Test suite:** upstream's Vega suite has ~400 `.puml` inputs with SVG goldens. Those goldens use a fixed width table and Smetana, so they do not match upstream's default CLI output.

## Goals / Non-Goals

**Goals:**

- A Go codebase whose structure maps to upstream's, so later diagram types and upstream fixes can be ported file by file
- No process-wide mutable state: every render is self-contained (needed for library-api concurrency and the HTTP server)
- One drawing abstraction with SVG and raster drivers, so every diagram type gets SVG and PNG at once
- Deterministic output without host fonts (output-formats spec)
- A conformance pipeline that every later follow-up change reuses unchanged

**Non-Goals:**

- Matching upstream's internal Java APIs or class names one to one; only behavior is the contract
- Performance targets. Go is expected to be faster than the JVM, but no budget is set in this change.
- Byte-identical SVG/PNG against upstream, and matching the Vega goldens (both need Smetana and the fixed width table; possible later)
- Supporting Graphviz versions older than upstream supports

## Decisions

### D1. Translate from the published MIT sources jar, file by file

- Translate Java to idiomatic Go by hand (AI-assisted), keeping upstream's package boundaries and algorithms.
- Each Go file starts with a header naming the upstream Java file(s) it derives from and keeping the upstream copyright notice, as MIT requires.
- The baseline is the Maven Central sources jar, verified by SHA-256, never the GPL main tree. Nothing from the excluded directories is read.

Alternatives considered:

- An automatic Java-to-Go transpiler: output is unidiomatic, and the inheritance, overloading and generics produce unmaintainable code at this scale.
- A clean-room rewrite: rejected by the user. Parity would be much slower to reach.

### D2. Module and package layout mirrors upstream

- **Module path:** `github.com/ekalinin/plantuml-in-go`. This is an assumption based on the GitHub account and directory name, and is cheap to change before the first release.
- **Go version:** the `go.mod` directive is `go 1.26`, so the two supported Go releases build it.
- **Binary:** `plantuml`.

| Go path | Upstream source | Purpose |
|---|---|---|
| `/` (package `plantuml`) | `SourceStringReader`, API | Public library API (library-api spec) |
| `cmd/plantuml` | `Run`, `cli/*` | CLI |
| `cmd/plantuml-wasm` | TeaVM entry points | WASM entry point (`GOOS=js GOARCH=wasm`) |
| `internal/core` | `BlockUml`, `PSystemBuilder`, `core/DiagramType` | Block extraction, factory order, error diagrams |
| `internal/tim` | `tim`, `preproc`, `preproc2` | Preprocessor |
| `internal/command` | `command`, `regex`, `com/plantuml/ubrex` | Regex command framework |
| `internal/klimt` | `klimt`, `klimt.shape`, `klimt.drawing`, `klimt.creole`, `klimt.sprite`, `style`, `skin`, `stereo`, `url` | Drawing model, creole, styling (one package: see below) |
| `internal/klimt/color`, `internal/klimt/geom`, `internal/klimt/font` | `klimt.color`, `klimt.geom`, `klimt.font` | Colors, geometry, text measurement |
| `internal/klimt/svg`, `internal/klimt/g2d` | `klimt.drawing.svg`, `klimt.drawing.g2d` | Output drivers |
| `internal/klimt/rose` | `skin.rose` | Skin components (participants, arrows, notes, groupings) |
| `internal/uml` | `AbstractPSystem`, `TitledDiagram`, `UmlDiagram`, `command` (common commands), `DiagramChromeFactory` | Shared diagram base, common commands, annotations |
| `internal/export`, `internal/png` | `core.TextBlockExporter`, `png` | Exporting a drawing to a format; PNG chunks |
| `internal/engine` | `SourceStringReader`, `BlockUml` | Parsing and drawing of blocks, used by the root package, CLI, server and WASM |
| `internal/errdiagram` | `error`, `PSystemWelcome` | Error, unsupported and welcome images |
| `internal/svek`, `internal/svek/extremity`, `internal/dot` | `svek` (+ `svek.image`), `svek.extremity`, `dot` | DOT generation, entity images, arrow heads, `dot` process, SVG coordinate parsing |
| `internal/decoration`, `internal/decoration/symbol` | `decoration`, `decoration.symbol` | Link types and decorations, USymbols |
| `internal/sequencediagram` | `sequencediagram` (+ `teoz`, `graphic`) | Sequence diagrams |
| `internal/classdiagram`, `internal/abel` | `classdiagram` (+ `objectdiagram`); `abel` (+ `cucadiagram`) | Class and object diagrams; the shared entity model |
| `internal/activitydiagram3` | `activitydiagram3` (+ `ftile`) | Activity diagrams |
| `internal/asciiverse`, `internal/asciiart` | `asciiverse`, `asciiart` | TXT/UTXT rendering |
| `internal/themes`, `internal/version` | `themes/`, `version` | Bundled themes; version and license text |
| `internal/code` | `code` | URL text encoding (re-exported by the root package) |
| `internal/security` | `security` | Security profiles, file and URL access |
| `internal/server` | `picoweb` + plantuml-server routes | HTTP server |
| `internal/conformance`, `tools/reference` | — | Conformance harness and reference image |

Keeping upstream names makes every Go file traceable to its source and keeps follow-up ports mechanical.

Java packages may reference each other in cycles; Go packages may not. Where upstream packages form such a cycle, they become one Go package, with file names prefixed by the original package:

- `klimt`, `klimt.shape`, `klimt.drawing`, `klimt.creole` (with `atom`, `command`, `legacy`), `klimt.sprite`, `style` (with `parser`), `skin`, `stereo` and `url` form one cycle and become `internal/klimt` (for example `shape_textblock.go`, `style_stylebuilder.go`, `skin_skinparam.go`). The leaf packages `klimt.color`, `klimt.geom` and `klimt.font` stay separate (`internal/klimt/color`, `internal/klimt/geom`, `internal/klimt/font`), and so do output drivers that nothing else imports (for example `internal/klimt/svg`).
- `tim`, `tim.expression`, `tim.iterator` and `tim.builtin` become `internal/tim`.
- A small type that an outer package needs from an inner one (for example `svek.Ports`, used by shapes) moves into the inner package instead of creating a cycle.

### D3. Replace singletons with an explicit render context

- Upstream global state (`GlobalConfig`, `FileSystem` current directory, `SecurityProfile`, pragma flags, the Graphviz path and timeout) becomes fields of a `RenderContext`.
- The context is created per render call and passed down explicitly.
- Static caches that are read-only after initialization (themes, font metrics, keyword lists) stay package-level and are loaded lazily with `sync.Once`.

This is what makes the library-api concurrency requirement and parallel `--threads` rendering possible without locks around rendering.

Alternative considered: keep globals guarded by a mutex. That serializes all rendering and breaks concurrent renders with different options.

### D4. Regular expressions: RE2 for commands, regexp2 only for user regexes

- Upstream command patterns are translated to Go `regexp` (RE2).
- The few patterns with lookarounds or backreferences are rewritten by hand, each covered by a unit test built from upstream's pattern and sample lines.
- Upstream's own `ubrex` engine is ported as Go code where in-scope commands use it.
- User-supplied regexes in preprocessor builtins (for example `%splitstr_regex`) use `github.com/dlclark/regexp2`, the closest pure Go match to Java regex semantics, with a match timeout.

Alternative considered: regexp2 everywhere. It backtracks, so ReDoS is possible on the HTTP server, and it is slower for the thousands of command matches per diagram.

### D5. Font metrics from the reference JVM, embedded as data

- A tool in the reference container uses Java2D `FontMetrics` with fractional metrics, exactly as upstream measures, to dump per-codepoint advance widths, ascent, descent and line height for these fonts:
  - DejaVu Sans: regular, bold, italic, bold-italic
  - DejaVu Sans Mono: the monospaced font
  - DejaVu Serif: the `Serif` logical font
- The result is a compact embedded table.
- Text width is the sum of advances scaled linearly by font size, matching Java's unkerned `getStringBounds`. Unknown codepoints use the fallback advance Java reports for the font.
- PNG rendering embeds the DejaVu TTF files (Bitstream Vera/DejaVu license, compatible with MIT bundling) for glyph outlines.
- Host fonts are never read. That gives deterministic output on every OS and in WASM.
- Font families named in diagrams (for example `skinparam defaultFontName Arial`) map to the closest embedded family. The mapping is recorded in the known-differences file.

Alternatives considered:

- Upstream's `UnicodeFontWidthSansSerif` table: deterministic, but differs from upstream's default output.
- Reading host fonts: not deterministic, and impossible in WASM.

### D6. One drawing model, two drivers

- Port `klimt`'s `UGraphic` model (shapes, strokes, colors, gradients, text, links, translations) as the single drawing API that diagram code uses.
- **SVG driver:** a port of `klimt/drawing/svg`. It keeps upstream's element order, attribute names and number formatting rules, so SVG output stays structurally close and the text-parity check is meaningful.
- **Raster driver:** a port of the `g2d` driver onto an in-house rasterizer that follows Java2D's rules: exact fills, Marlin-like sampling and stroking (joins, caps, miter limit, dashes), Java's 8-bit compositing, and text drawn from the embedded DejaVu fonts (unhinted outlines, placed on whole pixels). It uses only the standard library, works in WASM and is deterministic.
- **PNG metadata:** the `plantuml` text chunk is written by our own PNG chunk writer.

Alternatives considered:

- Rasterizing our own SVG: needs an SVG renderer, and none in pure Go supports everything upstream emits.
- Building on `golang.org/x/image/vector` directly: means writing stroking, dashing and gradients ourselves.
- `github.com/tdewolff/canvas` (the first choice): it pulls LGPL-2.1 code (fribidi) and an unlicensed package (scanx) into every binary, and it was only needed for stroking and dashing, which the in-house stroker now does with the same results.
- `fogleman/gg`: unmaintained.

### D7. Graph layout behind an engine interface; `dot` process first

- `svek` keeps upstream's pipeline:
  1. measure nodes
  2. generate the DOT string (same `DotStringFactory` logic)
  3. run `dot -Tsvg`
  4. parse node positions and edge paths from the returned SVG, as upstream does
  5. draw with `klimt`
- Steps 2-4 sit behind an `Engine` interface: `Layout(ctx, graph) (positions, error)`.
- **`dotexec` engine** (this change):
  - runs `dot` with `exec.CommandContext`, which gives the timeout and context cancellation from the graph-layout and library-api specs
  - finds `dot` in the order given in the graph-layout spec
  - reads `dot`'s stderr for error messages
- **Smetana engine** (follow-up change): a separate Go module, `github.com/ekalinin/plantuml-in-go/smetana`, in a subdirectory with its own `go.mod` and EPL-1.0 `LICENSE`. That keeps the EPL file-level copyleft out of the MIT module. The main module only imports it behind a build tag or an optional registration.

Alternatives considered:

- `goccy/go-graphviz` (Graphviz compiled to WASM, run by wazero): a large binary, it does not run inside our own WASM build, and it is a newer Graphviz than the reference.
- Pure Go Sugiyama libraries: layouts differ from `dot`, so parity fails.

### D8. Error reporting and detection follow upstream exactly

- Port `PSystemBuilder`'s factory order and error merging, including the "Assumed diagram type" logic, restricted to the in-scope factories.
- For out-of-scope `@startuml` types, a lightweight detector recognizes their characteristic commands. These are the first commands of upstream's state, description, object, timing and legacy activity factories. The detector turns a would-be syntax error into the "not supported yet" error from the diagram-source spec.
- Each follow-up change replaces a detector with the real factory.

### D9. Security layer wraps all I/O

- Port upstream's `security` package (`SFile`, `SURL`, profiles, allowlist).
- All file and URL access from the preprocessor and themes goes through an `Access` interface over `fs.FS` plus an HTTP client.
- **Default profile:** `LEGACY` for the CLI and library, `INTERNET` for the server, `SANDBOX` for WASM.
- **Library and WASM:** callers can supply their own `fs.FS`.
- **Error messages:** follow upstream's rule of hiding local paths outside the input.

### D10. URL encoding

- Raw deflate (`compress/flate`, best compression) plus upstream's 64-character alphabet.
- A port of upstream's legacy Huffman decoder, and the hex, gzip and zip forms with stdlib packages.
- Encoded strings may differ byte-wise from Java's `Deflater`. The url-encoding spec only requires that upstream can decode them.

### D11. HTTP server on `net/http`

- One `http.Handler` implements upstream's `--http-server` routes plus the plantuml-server routes from the http-server spec.
- Body size limit (1 MiB) and per-request timeout (60 s) via `http.MaxBytesReader` and a context deadline. The timeout also stops `dot`.
- The server reuses the library API, so its output is identical to CLI output.

### D12. WASM via `GOOS=js GOARCH=wasm`

- `cmd/plantuml-wasm` exports `plantuml.render(source, options)` through `syscall/js` and returns a Promise.
- The loader is Go's `wasm_exec.js` plus a thin ES module wrapper that runs in browsers and Node.js.
- Graph diagrams return the Graphviz-not-available error until the Smetana module exists.
- Binary size is measured and reported in CI but not gated.

Alternatives considered:

- `wasip1`: needs a WASI shim in the browser.
- TinyGo: incomplete support for `reflect` and `regexp`. Can be evaluated later for size.

### D13. Reference renders and conformance harness

**Reference image** (`tools/reference/Dockerfile`):

- base image `eclipse-temurin:21-jre-jammy`, pinned by digest
- `fonts-dejavu`
- Graphviz 14.0.1 copied from upstream's official image `plantuml/plantuml:1.2026.8` (pinned by digest), which builds it from source on the same base image. Compiling it in our image needs amd64 emulation on arm64 hosts, where the compiler crashes at random
- `plantuml-mit-1.2026.8.jar` from Maven Central, verified by SHA-256
- `resvg`, pinned, used as a single neutral SVG rasterizer
- the font-metrics dump tool from D5

**Corpus:**

- Vega inputs are fetched at tag `v1.2026.8` into a cache by a script, not committed. That avoids redistributing upstream test files.
- Curated cases live in `testdata/conformance/cases/<capability>/<case>.puml`, with expected format and type in a YAML header.

**References:**

- generated by `make reference`. A Java driver (`tools/reference/ReferenceRender.java`) uses the jar as a library in one JVM and records per-case metadata (blocks, diagram type, error line and messages, sizes). PREPROC comes from the CLI in batch mode
- for curated cases, SVG, TXT, UTXT, PREPROC and metadata are committed under `testdata/conformance/reference/`, as reviewable text
- Vega references go to `.cache/reference/vega/` and are not committed, because they embed the Vega sources
- PNG references are regenerated on demand from the pinned image, which is deterministic
- upstream crash reports (render status 503) show a random quote and picture, so only their metadata is recorded

**Comparison** (`go test -tags conformance ./internal/conformance`). Outputs are produced in the reference image, so both sides use the same Graphviz. The comparison itself runs on the host, with the pinned resvg and the DejaVu fonts copied from the image (`tools/reference/fetch-tools.sh`), because Go amd64 binaries crash under QEMU emulation on arm64 hosts:

1. Success or failure must agree, and the error line must match.
2. The SVG `<text>` contents must match as a multiset after whitespace normalization.
3. Width and height must be within 5%.
4. SSIM must be ≥ 0.90:
   - for SVG cases, both SVGs are rasterized by `resvg` and compared
   - for PNG cases, the Go PNG is compared with the Java PNG
5. Text formats must be equal apart from trailing whitespace.

**CLI and HTTP cases** run the same scenario against both binaries, or both servers, inside the container and compare exit codes, file names, status codes, headers and bodies. Routes that only plantuml-server provides are recorded against the pinned `plantuml/plantuml-server:v1.2026.8` image.

**Output:**

- `testdata/conformance/known-differences.yaml` lists case ID, reason, and the follow-up change that resolves it
- the report is written to `conformance-report.md`, with pass rates per diagram type and format

Alternatives considered:

- Vega goldens as-is: they encode Smetana and the fixed width table, not upstream's default output.
- Pixel diff thresholds instead of SSIM: SSIM is less sensitive to sub-pixel anti-aliasing shifts caused by font rasterizers.

### D14. Bundled assets and licensing files

- Themes are embedded with `embed.FS`. `sunlust` is excluded because it is GPL-3+.
- Themes without a clear license line are included only after checking upstream history and `LICENSES.md`. Otherwise they are excluded and listed in NOTICE.
- `LICENSE` is MIT, with upstream's copyright line plus the project's.
- `NOTICE` lists:
  - the upstream origin and version
  - attributions for the Apache-2.0 themes
  - the DejaVu font license
  - the excluded components
- Stdlib bundling is out of scope. The preprocessor's `<lib/...>` resolution is implemented against an empty embedded index, so the "not bundled" error from the preprocessor spec is real behavior.

### D15. Versioning

- `--version` prints `plantuml-go <semver> (PlantUML 1.2026.8 compatible)`.
- `%version` returns `1.2026.8`, because diagrams use it to branch on upstream features.

### D16. Upstream bugs: fix them, follow upstream elsewhere

Checking the specs against upstream v1.2026.8 found behavior that is a defect rather than a design choice. Where upstream crashes, hangs or ignores one of its own options, this implementation fixes it. Everything else follows upstream, even where it looks surprising.

Fixed:

| Situation | Upstream v1.2026.8 (MIT build) | This implementation |
|---|---|---|
| CLI without arguments | tries to open the GUI, HeadlessException, exit 1 | usage, exit 0 |
| `--pdf` and other flags outside this change | crashes, leaves a 0-byte PDF, exit 0 | error naming the flag, exit 1 |
| `dot` timeout | not enforced | 30 s default, process stopped |
| `dot` failure | crash report without `dot`'s message | error that includes `dot`'s message |
| `--stop-on-error` | only works together with `--threads` | stops at the first error |
| Large request bodies and slow renders | accepted without limit | 413 above 1 MiB, 503 after 60 s |
| `!pragma layout elk` | crashes (no ELK classes in the MIT build) | rendered with `dot` |

Followed as upstream: for example, an undecodable URL returns status 200 with upstream's error image, `!include_once` includes a file again without error, unknown CLI flags are ignored, and `--check-syntax` prints only "Some diagram description contains errors".

Intentionally omitted: upstream adds sponsor messages to error images during some minutes of each hour. This implementation never adds them, and references are generated with them disabled.

### D17. Deterministic floating point

The output-formats spec requires byte-identical output on every OS and architecture. Go may fuse `a*b+c` into one multiply-add on arm64 (and amd64 v3), including inside the standard `math` package, so the same code can give different bits on a Mac and on a Linux server.

- Products that feed an addition are written `float64(a*b)` in layout and drawing code; an explicit conversion forbids fusion.
- Elementary functions (atan2, atan, sin, cos, exp, log, pow, ...) come from an internal Java-math package, a port of fdlibm (the library behind Java's StrictMath) with the same guards. Results then match Java's StrictMath and are identical on every platform.
- fdlibm is under Sun's permissive license ("Permission to use, copy, modify, and distribute this software is freely granted, provided that this notice is preserved"). Its files carry a `Provenance: fdlibm` header with the notice, and NOTICE lists it.

Alternative considered: Go's `math` package: not deterministic across architectures, and differs from Java in the last bit for about a third of atan2 inputs.

## Risks / Trade-offs

- **[Scale]** About 150k Java lines are in scope for this change alone, so there is a risk of a long period with nothing usable. → Port in vertical slices. Sequence diagrams to SVG through the library and CLI come first, then the other diagram types and interfaces. Only code reachable from in-scope features is translated. Every slice is gated by conformance cases.
- **[Java vs Go semantics]**
  - `double` formatting, integer overflow, `HashMap` iteration order and string handling (UTF-16 indices) differ between Java and Go.
  - → Port upstream's number formatting helpers exactly.
  - → Use insertion-ordered maps where Java uses `LinkedHashMap`.
  - → Audit `HashMap` iterations that reach output and record any unavoidable differences.
  - → Index strings by rune where Java indexes by `char`, with tests on non-BMP text.
- **[Font metric drift]** Metrics extracted from one JVM and font build may differ from users' Java setups. → Parity is defined against the pinned reference only. That's no worse than upstream itself, whose output already varies by OS.
- **[amd64 emulation on arm64 hosts]** Under QEMU, the JVM works, but the compiler and Go binaries crash at random. → References come from the pinned amd64 image, and comparison runs natively on the host. The authoritative conformance run is on linux/amd64 (CI), where the Go implementation runs in the reference image without emulation. On arm64 hosts, the Go implementation runs natively and graph diagrams use the host's Graphviz, so their results there are indicative only.
- **[Graphviz version drift]** Users' `dot` versions (for example 15.0.0 locally) give different layouts from the reference 14.0.1. → Conformance runs only in the pinned image. The dot SVG parser is tested against both 14.0.1 and the latest release. Upstream behaves the same way.
- **[RE2 gaps]** A missed lookaround translation silently changes parsing. → Each translated command pattern gets table-driven tests from upstream's samples. Conformance error-line checks catch regressions.
- **[License contamination]** Code from GPL-only parts could slip in. → Translate only from the MIT sources jar. A CI check requires every Go file under `internal/` to have a provenance header naming an MIT-jar path or "original". Excluded package names are blocked.
- **[SSIM threshold]** 0.90 may hide real differences or flag harmless ones. → The text and size checks catch semantic errors independently. The threshold is reviewed after the first diagram type and can only be tightened.
- **[WASM size]** Embedded fonts, themes and Go runtime could exceed 20 MB. → Measure in CI. Optimize later (lazy theme loading, compression, TinyGo evaluation) without spec changes.
- **[Upstream moves on]** New upstream releases add behavior. → The target is pinned. Bumping the version is its own change that regenerates references and reviews diffs.

## Migration Plan

This is a new project, so there is nothing to migrate. Rollout order, matching tasks.md:

1. Project skeleton, licensing files, reference image, conformance harness
2. Core pipeline: block extraction, preprocessor, URL encoding, security, drawing model, SVG driver, font metrics, styling
3. Sequence diagrams (Teoz) with SVG and TXT/UTXT, then library API and CLI basics. First usable slice.
4. Activity diagrams
5. Graph layout (svek and the `dot` engine) and class diagrams
6. Raster driver and PNG parity
7. HTTP server
8. WASM build
9. Full CLI flag coverage, conformance gating and the support declaration per diagram type

Users replace `java -jar plantuml.jar` with `plantuml` per diagram type, as the conformance report declares types supported. Rolling back means switching back to the Java jar. Input files do not change.

## Open Questions

- Do the themes with no license line have a stated license elsewhere in upstream history? This decides only whether each such theme is bundled now or later. The bundling mechanism is the same either way.
