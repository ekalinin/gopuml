# Tasks

No output changes, so no conformance case, reference or golden dump is added or changed. The branch is `refactor/reuse-compressors` (the branch types of AGENTS.md have no `perf/`). The local benchmark of PNG output lands first, so that it records the cost before the change; each later group measures its own effect after its commit.

## 1. Local benchmark of PNG output

- [x] 1.1 Add `BenchmarkRenderSmallPNG` to `render_test.go` (design D6). It renders `@startuml`, `Alice -> Bob : hello`, `@enduml` as PNG with `b.ReportAllocs()`, after one unmeasured render, as `BenchmarkRender` does, with a comment that CI does not compare it (issue #100). Verify:
  - `go test -run '^$' -bench '^BenchmarkRenderSmallPNG$' -benchmem .` reports about 4.3 MB per operation;
  - `go test -list '^BenchmarkRender$' .` lists `BenchmarkRender` alone, so `tools/benchcmp` does not run the new benchmark.

  Commit as `chore(gopuml): measure the allocations of a small PNG render`

## 2. Bounded list of reusable values

- [x] 2.1 Create `internal/freelist` with `doc.go` and `freelist.go` (`Provenance: original`): `List[T any]`, `New[T](n int, create func() T) *List[T]`, `Get() T` and `Put(v T)` on a buffered channel, neither of which blocks (design D1). Add `freelist_test.go` with the tests of design D5:
  - `New(4, func() *int { return new(1) }).Get()` points to 1, and a list without `create` returns nil;
  - a pointer put back is the one the next `Get` returns;
  - after four `Put` calls, a fifth value is dropped, and four `Get` calls return the first four;
  - a pointer put back is returned after two `runtime.GC()` calls;
  - 8 goroutines that each get and put 1000 times end with at most four values in the list.

  Verify that `go test -race ./internal/freelist` passes and that `var _ png.EncoderBufferPool = (*freelist.List[*png.EncoderBuffer])(nil)` compiles in the test. Commit as `feat(freelist): add a bounded list of reusable values`

## 3. Writer of the encoded source (`internal/code`)

- [x] 3.1 Replace the `sync.Pool` `deflaters` with a `freelist.List[*flate.Writer]` of capacity 4, whose `create` makes a writer at `flate.BestCompression`, and keep the steps of `deflateRaw` (design D2). Verify:
  - `go test -race ./internal/code` passes, `TestDeflateRawReusesWriters` included;
  - `TestEncodeDecodeRoundTrip` of the root package passes.

  Commit as `perf(code): keep the deflate writers across garbage collections`, then verify that `make bench-compare BASE=HEAD^1` exits 0, with `-5.0%` or lower for the `B/op` of `class`, and no type above `+1.0%` in `B/op` or `allocs/op`. Done: `B/op` fell for every type, by 16.6% for `class`, 17.0% for `activity` and 18.7% for `chen_eer`, but by 2.8% only for `timing`, which this task first expected at `-5.0%` or lower from a memory profile; `allocs/op` stayed within 0.1%

## 4. PNG data of image sprites (`internal/klimt`)

- [x] 4.1 Give `pngDataURLEncoder` a `freelist.List[*png.EncoderBuffer]` of capacity 4 without `create` as its `BufferPool`, and remove `pngBufferPool` (design D3). Add a test of `ToPngDataURL`: for a 16 by 16 image with a gradient, a 1 by 1 transparent image and a 200 by 100 image, in turn, and for the same images on 8 goroutines, the text equals `data:image/png;base64,` followed by `png.Encode` of the same pixels. Verify that `go test -race -run 'ToPngDataURL|SpriteImageFromInternal' ./internal/klimt` passes. Commit as `perf(klimt): keep the PNG encoder buffers of image data across garbage collections`, then verify that `make bench-compare BASE=HEAD^1` exits 0 with no type above `+1.0%` in `B/op` or `allocs/op`

## 5. PNG output (`internal/png`)

- [ ] 5.1 Encode `WriteWithMetadata` with a package-level `png.Encoder` at `png.DefaultCompression` whose `BufferPool` is a `freelist.List[*png.EncoderBuffer]` of capacity 4. Make `itxtChunk` take a `*zlib.Writer` from a `freelist.List[*zlib.Writer]` of capacity 4, whose `create` is `zlib.NewWriter(nil)`, `Reset` it on its buffer, write, close and put it back (design D4). Add to `pngio_test.go` (design D5): for the three images of task 4.1 with the metadata `@startuml\nAlice -> Bob : hello\n@enduml` and dpi 96 and 192, in turn and on 8 goroutines, `WriteWithMetadata` writes the same bytes as the same chunks built with a new `png.Encoder` and a new `zlib.NewWriter`. Verify:
  - `go test -race ./internal/png` passes;
  - `go test -run '^$' -bench '^BenchmarkRenderSmallPNG$' -benchmem .` reports 2.9 MB or less per operation (about 4.3 MB in task 1.1).

  Commit as `perf(png): reuse the PNG encoder and the zlib writer of the metadata`

## 6. Verification

- [ ] 6.1 Run `make bench-compare BASE=main`. Verify that it exits 0, that `B/op` falls for `class`, `activity` and `chen_eer` by 5% or more, and that no type grows above `+1.0%` in `B/op` or `allocs/op`
- [ ] 6.2 Run `make test` and `go test -race . ./internal/freelist ./internal/code ./internal/klimt ./internal/png`. Verify that both pass, `TestRenderAfterOtherRenders` and `TestRenderConcurrent` included
- [ ] 6.3 Run `make conformance REQUIRE=all GO_IN_IMAGE=1`; on arm64 hosts, the native `make conformance REQUIRE=all`, which gives the results of CI. Verify 0 failures, no known difference that now passes, and the same counts of passed cases, known differences and unsupported cases per diagram type and per format (PNG included) as before the change
- [ ] 6.4 Run `make wasm`. Verify that it builds `dist/wasm/plantuml.wasm`, and that `go test ./cmd/plantuml-wasm` passes (the lists run on one thread there)

## 7. Documentation, checks and pull request

- [ ] 7.1 Check README.md and site/index.html: no count, table or timing changes. Verify that `TestLandingPage` and `TestReleaseNotes` pass
- [ ] 7.2 Run `make lint` and `make provenance`. Verify that both pass, with the header of `internal/freelist`
- [ ] 7.3 With the user's approval, push the branch and open a pull request whose description has the conformance result of task 6.3, the `B/op` of `BenchmarkRenderSmallPNG` before and after, and a link to issue #100. Verify that its `test`, `conformance` and `benchmark` jobs pass
- [ ] 7.4 Run `openspec validate reuse-compressors --strict`. Verify that it passes
