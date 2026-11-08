# Proposal

## Why

Renders still allocate a new deflate compressor of about 1 MB in two cases. No issue tracks this. No conformance case fails or changes: the output of every render stays byte-identical.

- **Pool misses for SVG.** The `sync.Pool` of `code.deflateRaw` (change `reduce-render-allocations`, D1) misses in about 2-10% of the renders of `BenchmarkRender`. A writer that the render goroutine puts back sits in the private slot of one P, which a `Get` on another P cannot see, and the garbage collector drops the idle writers. These misses are about 12% of the allocated bytes of `class`, 16% of `timing` and 2.5% of `sequence`. The `sync.Pool` of the PNG data of image sprites (`klimt.ToPngDataURL`, change `add-archimate-diagrams`, D8) misses the same way.
- **No pool for PNG output.** It creates two zlib writers for each render, one for the image data (`png.WriteWithMetadata`) and one for the iTXt chunk of the source (`itxtChunk`). A PNG render of `Alice -> Bob : hello` allocates about 4.3 MB, of which about 1.4 MB are these writers; its SVG render allocates 0.76 MB.

## What Changes

- **A bounded list of reusable values.** A new package holds up to four values, takes them back after use and keeps them across garbage collections, unlike `sync.Pool`.
- **Reuse of every compressor.** Each deflate compressor of a render comes from such a list and goes back to it:
  - the writer of `<?plantuml-src ...?>` (`internal/code`);
  - the PNG encoder of the PNG data of image sprites (`internal/klimt`);
  - the PNG encoder of PNG output and the zlib writer of its iTXt chunk (`internal/png`).
- **Same output.** Every value is reset before use with the same compression level, so the compressed bytes do not change.
- **A local benchmark of PNG output**, which CI does not run, to measure the PNG writers.

Out of scope:

- PNG output in the benchmark of CI (`BenchmarkRender`, `tools/benchcmp`), tracked by issue #100.
- A comparison of `tools/benchcmp` limited to the cases that both commits have.
- The deflate readers that decode PNG files and the standard library, and the other allocations of PNG output (the raster, `quantify555`).
- A faster compression level, which would change the encoded source and the PNG bytes.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None. The change sets `skip_specs: true`: no requirement changes, every output stays the same.

## Impact

**Code:** a new package `internal/freelist`, `internal/code` (deflate writer), `internal/klimt` (PNG encoder of image sprites), `internal/png` (PNG encoder and iTXt zlib writer), and benchmarks.

**Conformance:** no case, reference, report count or golden dump changes. The suite must give the same results as before.

**Documentation:** none.

**Compatibility:** no API change. A long-running process keeps up to four idle compressors of each kind, about 16 MB in all, instead of letting the garbage collector drop them.
