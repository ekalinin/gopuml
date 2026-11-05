# Tasks

No output changes, so no conformance case, reference or golden dump is added or changed. `BenchmarkRender` and `make bench-compare` come from `benchmark-pull-requests`, which merges first. The regression test lands first here, so that every later group is checked by it; each group measures its own effect with `make bench-compare BASE=HEAD^1` after its commit.

## 1. Branch and regression test

- [x] 1.1 Once `benchmark-pull-requests` is merged, create the branch `refactor/reduce-render-allocations` from `main` (the branch types of AGENTS.md have no `perf/`); the untracked files of this change move with it. Verify that `git log -1 --format=%s` shows the head of `main`, that `git status --short` lists only `openspec/changes/reduce-render-allocations/` besides the files that were untracked before, and that `go test -run '^$' -bench BenchmarkRender -benchtime 1x .` and `make -n bench-compare` work on the branch
- [x] 1.2 Add `TestRenderAfterOtherRenders` to `render_test.go` (design Risks): it renders `@startuml`, `Alice -> Bob : hello`, `@enduml`, then `@startuml`, `!include <C4/C4_Container>`, `Person(user, "User")`, `Container(api, "API", "Go")`, `Rel(user, api, "Uses")`, `@enduml` twice, then the first source again, and expects the two SVG outputs of the first source to be byte-identical, `<?plantuml-src ...?>` included, and the two outputs of the C4 source too. Verify that it passes on the current code, with `go test -race -run 'TestRenderAfterOtherRenders|TestRenderConcurrent' .`, and that `make test` passes. Commit as `chore(gopuml): check that a render does not depend on earlier renders`

## 2. Deflate writers (`internal/code`)

- [x] 2.1 Make `deflateRaw` take its `*flate.Writer` from a package-level `sync.Pool` whose `New` creates it at `flate.BestCompression`, `Reset` it on a local `bytes.Buffer`, and put it back after `Close` (design D1). Add to `code_test.go` a test that compares `deflateRaw` with a fresh `flate.NewWriter(..., flate.BestCompression)` for `@startuml\nAlice -> Bob : hello\n@enduml`, for a 100 KB source and for these sources encoded in turn, and the same on 8 goroutines in parallel. Verify that `go test -race ./internal/code` and `TestEncodeDecodeRoundTrip` of the root package pass. Commit as `perf(code): reuse deflate writers`, then verify that `make bench-compare BASE=HEAD^1` exits 0 and shows `-50.0%` or lower for the `B/op` of `sequence`

## 3. Leaf pattern check (`internal/command/regex`)

- [x] 3.1 Make `NewNamedLeaf` look up the group count of its pattern in a package-level `sync.Map`, compute and store it on a miss, and compare it with `count` on every call (design D2). Add to `composed_test.go`: `NewNamedLeaf(1, "X", "(a)")` called twice returns two leaves with the pattern `(a)`; `NewNamedLeaf(2, "X", "(a)")` panics with `regex: leaf "(a)" has 1 groups, declared 2` after the correct call has stored the pattern; `NewNamedLeaf(0, "", "(")` panics on two calls in a row, with the same message as before the change. Verify that `go test -race ./internal/command/...` passes. Commit as `perf(regex): check the groups of a leaf pattern once`, then verify that `make bench-compare BASE=HEAD^1` exits 0 and shows a negative change for the `allocs/op` of `class` and `timing`

## 4. Iterator trace (`internal/command`)

- [x] 4.1 Remove the `trace` field of `iterator` and add a `trace()` method that returns `it.data[:it.pos:it.pos]`; `next()` no longer appends, `clone()` copies `data` and `pos`, and the five `newError` calls of `factory.go` pass `it.trace()` (design D3). Add to `factory_test.go`: over the lines `a`, `b`, `c`, two `next()` calls give a trace of `a`, `b`; a clone advanced by one more line has the trace `a`, `b`, `c` while the original keeps `a`, `b`; appending `x` to a trace leaves the third line `c`. Verify that `go test ./internal/command ./internal/errdiagram ./internal/sequencediagram ./internal/classdiagram ./internal/activitydiagram3` passes (the golden error images and dumps are unchanged). Commit as `perf(command): derive the iterator trace from the source lines`, then verify that `make bench-compare BASE=HEAD^1` exits 0 and shows a negative change for the `B/op` of `description` (which holds the C4 cases)

## 5. Skinparam key cache (`internal/klimt`)

- [ ] 5.1 Replace `SkinParam.cacheCleanForKey` with a package-level LRU in `skin_skinparam.go`: a map to `container/list` elements in access order, `cacheCleanForKeyMax = 2000`, a `sync.Mutex` around the map and list operations, `cleanForKeySlow` computed outside the lock, and a comment that the cached slices are read-only (design D4). Remove the field and its initialization in `SkinParamCreate`. Add to `klimt_test.go`: for each line of `testdata/golden/cleanforkey_inputs.txt`, two calls of `cleanForKey` return `cleanForKeySlow` of the line; after the keys `k0` to `k1999` and a lookup of `k0`, adding `k2000` drops `k1` and keeps `k0`, with 2000 entries; 8 goroutines that each look up 1000 keys out of 3000 get `cleanForKeySlow` of each. Verify that `go test -race ./internal/klimt` passes, `TestGoldenCleanForKey` included. Commit as `perf(klimt): cache normalized skinparam keys per process, as upstream`, then verify that `make bench-compare BASE=HEAD^1` exits 0 and that the `ns/op` of `description` falls by more than its spread

## 6. Verification

- [ ] 6.1 Run `make bench-compare BASE=main`. Verify that it exits 0, that `B/op` falls for every type, by more than half for `sequence` (`-50.0%` or lower), and that no type's `ns/op` grows beyond its spread
- [ ] 6.2 Run `make test` and `go test -race . ./internal/code ./internal/command/... ./internal/klimt`. Verify that both pass, `TestRenderAfterOtherRenders` and `TestRenderConcurrent` included
- [ ] 6.3 Run `make conformance REQUIRE=all` natively with `GRAPHVIZ_DOT` unset (the layout is gogv's, as in CI, which runs it with `GO_IN_IMAGE=1`). Verify 0 failures, no known difference that now passes, and the same counts of passed cases, known differences and unsupported cases per diagram type as `conformance-report.md`
- [ ] 6.4 Run `make wasm`. Verify that it builds `dist/wasm/plantuml.wasm` (the pool and the mutex run on one thread there)

## 7. Documentation, checks and pull request

- [ ] 7.1 Check README.md and site/index.html: no count, table or timing changes (the timings of the site are those of v0.3.0). Verify that `TestLandingPage` and `TestReleaseNotes` pass
- [ ] 7.2 Run `make lint` and `make provenance`. Verify that both pass
- [ ] 7.3 With the user's approval, push the branch and open a pull request whose description has the conformance result of task 6.3; the benchmark table is in the summary of its `benchmark` job. Verify that its `test`, `conformance` and `benchmark` jobs pass, and that the summary of `benchmark` shows `-50.0%` or lower for the `B/op` of `sequence`
- [ ] 7.4 Run `openspec validate reduce-render-allocations --strict`. Verify that it passes
