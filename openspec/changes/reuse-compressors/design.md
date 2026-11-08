# Design

## Context

See proposal.md - Why for the measurements. Upstream relies on Java's deflate, which wraps native zlib, and the Go translations allocate the tables of Go's `compress/flate` instead:

| Upstream | Go today | What differs |
|---|---|---|
| `code/CompressionZlib.java`: a new `Deflater(COMPRESSION_LEVEL, true)` per call | `internal/code/code.go`, `deflateRaw`: a `flate.Writer` at `flate.BestCompression` from a `sync.Pool` (`deflaters`) | Some calls miss the pool and allocate about 1 MB. |
| `klimt/drawing/svg/SvgGraphics.java`: `SImageIO.write(image, "png", baos)` for the data of an image | `internal/klimt/atmp_pixelimage.go`, `PortableImageARGB.ToPngDataURL`: `pngDataURLEncoder`, a `png.Encoder` whose `BufferPool` is a `sync.Pool` (`pngBufferPool`) | The same misses, at about 1.1 MB each. |
| `png/PngIO.java` and `png/PngIOMetadata.java`: ImageIO's PNG writer, with the iTXt chunk of the source in its metadata | `internal/png/pngio.go`: `WriteWithMetadata` encodes with a new `png.Encoder` per call, and `itxtChunk` compresses the source with a new `zlib.NewWriter` per call | No reuse at all: about 1.1 MB and 1.0 MB per PNG render. |

Measurements behind the decisions:

- A new `flate.Writer` at `BestCompression` allocates about 1.0 MB, at `DefaultCompression` 0.75 MB, a zlib writer about 1.1 MB once written to, and `png.Encode` of a 16 by 16 image about 1.1 MB.
- In `BenchmarkRender`, the garbage collector runs between about 3 and 16 times per pass over a diagram type. Memory profiles count about 49 new deflate writers per 713 `class` renders and 220 per 2108 `timing` renders.
- `sync.Pool` keeps a value that is put back in the private slot of the current P. A `Get` on another P cannot take it, and a collection moves the pool to its victim cache, which the next collection empties.

Constraints:

- Every output stays byte-identical, including the encoded source of `<?plantuml-src ...?>`, the URL encoding, the PNG data of image sprites in SVG and the bytes of PNG output.
- The library is safe for concurrent use, the command line renders files on several goroutines (`--threads`), and the WASM build runs on one thread.
- The HTTP server runs for a long time, so what the process keeps must be bounded.

## Goals / Non-Goals

**Goals:**

- Reuse the four compressors of a render across renders and garbage collections, without changing any output or any exported API.
- Show the effect on SVG with `make bench-compare`, and on PNG with a local benchmark.

**Non-Goals:**

- Porting anything from upstream: the change touches only how the Go translations allocate.
- PNG output in `BenchmarkRender` and in the `benchmark` job (issue #100).
- The deflate readers of PNG decoding and of the standard library, which these profiles do not rank high.
- Another compression level, which would change the encoded source and the PNG bytes.

## Decisions

### D1. A bounded list of reusable values

A new package `internal/freelist` (`Provenance: original`) holds `List[T any]`, a buffered channel of capacity n and an optional function that creates a value:

- `New[T](n int, create func() T) *List[T]`;
- `Get() T` receives a value when the channel holds one, and otherwise returns `create()`, or the zero value without `create`;
- `Put(v T)` sends the value when the channel has room and drops it otherwise.

The channel is referenced by a package-level variable of each user, so its values survive garbage collections. `Get` and `Put` never block, and a value is held by one caller at a time. With `Get() T` and `Put(T)`, a `*List[*png.EncoderBuffer]` without `create` is a `png.EncoderBufferPool` as it is: the encoder allocates its buffer when `Get` returns nil.

Each list holds up to four values. A process that renders more than four diagrams at a time allocates the extra compressors and drops them, as the pools do today.

Alternative: `sync.Pool`, as today. Its private slots per P and its victim cache cause the measured misses.

Alternative: a slice behind a `sync.Mutex`, as the spike of `add-archimate-diagrams` did. It behaves the same, with more code and a lock.

Alternative: a capacity of `runtime.GOMAXPROCS(0)`. A server on many cores would keep more idle compressors, and four already cover the command line's default of one thread and its common `--threads` values.

### D2. The writer of the encoded source

`internal/code` replaces its `sync.Pool` `deflaters` with a `freelist.List[*flate.Writer]` whose `create` makes a writer at `flate.BestCompression`. `deflateRaw` keeps its steps: `Get`, `Reset` on a local buffer, `Write`, `Close`, `Put`. `TestDeflateRawReusesWriters`, which compares the output with a fresh writer in turn and in parallel, keeps covering it.

### D3. The PNG data of image sprites

`internal/klimt` gives `pngDataURLEncoder` a `freelist.List[*png.EncoderBuffer]` without `create` as its `BufferPool`, and removes the `pngBufferPool` type. The encoder keeps the default compression level.

### D4. PNG output

`internal/png` encodes `WriteWithMetadata` with a package-level `png.Encoder` at `png.DefaultCompression`, whose `BufferPool` is a `freelist.List[*png.EncoderBuffer]`. `itxtChunk` takes a `*zlib.Writer` from a `freelist.List[*zlib.Writer]`, whose `create` is `zlib.NewWriter(nil)`, at the default level as today. It calls `Reset` on its buffer, writes, closes and puts the writer back. `Reset` keeps the level, so the bytes do not change.

### D5. Tests of reuse

- `internal/freelist`:
  - an empty list returns a created value, or the zero value without `create`;
  - a value put back is returned by the next `Get`;
  - a fifth `Put` into a full list of four is dropped;
  - a value put back is still returned after two `runtime.GC()` calls, which empty a `sync.Pool`;
  - 8 goroutines that get and put values run with `-race`.
- `internal/png`:
  - `WriteWithMetadata` of several images in turn gives the same bytes as an encoding with a new `png.Encoder` and a new zlib writer;
  - the same holds on 8 goroutines.
- `internal/klimt`: `ToPngDataURL` of several images in turn and on 8 goroutines gives the same text as `png.Encode`.
- `TestRenderAfterOtherRenders` and `TestRenderConcurrent` of the root package keep checking that a render does not depend on earlier or parallel renders.

### D6. Local benchmark of PNG output

`render_test.go` gets `BenchmarkRenderSmallPNG`, which renders `@startuml`, `Alice -> Bob : hello`, `@enduml` as PNG with its allocations. `tools/benchcmp` runs `^BenchmarkRender$` only, so CI does not compare it; it measures D4 locally until issue #100 adds PNG to the comparison. Before this change it reports about 4.3 MB per render.

Alternative: a benchmark of `WriteWithMetadata` in `internal/png`. It would leave out the PNG data of image sprites and the rest of a PNG render, which the numbers of proposal.md are about.

## Risks / Trade-offs

- [A reused value carries state into the next use] → each value is reset before use; the tests of D5 compare the output with fresh compressors, in turn and in parallel, and the conformance suite renders every case with the new code.
- [Process memory] → up to four idle values in each of the four lists, about 1 MB each, so about 16 MB at most in a process that has rendered four diagrams at a time. The command line exits after its renders. The WASM build, on one thread, keeps one value per list.
- [More than four renders at a time] → the extra compressors are allocated and dropped, as before the change.
- [`BenchmarkRender` does not show the gain on PNG] → `BenchmarkRenderSmallPNG` (D6) measures it locally; issue #100 adds PNG to the comparison of CI.

## Migration Plan

No API, output or configuration changes. The list comes first, then each user in its own commit (D2, D3, D4), so a regression found later can be reverted alone.
