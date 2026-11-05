# Design

## Context

See proposal.md - Why for the measurements. The four costs come from Go translations of upstream classes, where the Go code either does work that upstream skips or keeps state per diagram that upstream keeps per process:

| Upstream | Go today | What differs |
|---|---|---|
| `code/CompressionZlib.java`: a new `Deflater(COMPRESSION_LEVEL, true)` per call | `internal/code/code.go`, `deflateRaw`: a new `flate.Writer` at `flate.BestCompression` per call | Java's `Deflater` wraps native zlib; Go's `flate.NewWriter` allocates and clears about 1 MB of tables each time. |
| `regex/RegexLeaf.java`: `if (TeaVM.a()) assert internalSlowCount() == count`, a Java assertion that does not run unless the JVM enables assertions | `internal/command/regex/composed.go`, `NewNamedLeaf`: transforms and compiles the pattern on every call to check `NumSubexp`, then drops the compiled regexp | The check runs on every call, including calls made while commands execute (`color.SimpleColorID`, `timing.addState`, `sequencediagram.executeNote`, `klimt.URLBuilderOptional`). |
| `version/IteratorCounter2Impl.java`: `trace` is an `ArrayList` that `next()` appends to and `cloneMe()` copies | `internal/command/bloclines.go`, `iterator`: `trace []text.StringLocated`, appended in `next()`, copied by `slices.Clone` in `clone()` | Nothing; the Go port copies the Java structure. Since `pos` starts at 0 and only `next()` advances it, the trace is always `data[:pos]`. |
| `skin/SkinParam.java`: `cacheCleanForKey`, a `static` `LinkedHashMap` in access order, bounded to `CACHE_CLEAN_FOR_KEY_MAX = 2000` entries, under `synchronized` | `internal/klimt/skin_skinparam.go`: `SkinParam.cacheCleanForKey map[string][]string`, one per diagram | The process-wide LRU is missing; each diagram starts with an empty cache. |

Constraints:

- Every output stays byte-identical, including the encoded source of `<?plantuml-src ...?>` and the URL encoding.
- The library is safe for concurrent use (`library-api`: Concurrency and cancellation), the command line renders files on several goroutines (`--threads`), and the WASM build runs on one thread. Process-wide state must work in all three.
- The HTTP server renders untrusted sources for a long time: a process-wide cache keyed by source text must be bounded.

The change depends on `benchmark-pull-requests`, which lands first: its `BenchmarkRender` renders the curated cases per diagram type, `make bench-compare BASE=<ref>` compares the checked-out commit with another one, and its `benchmark` job does the same for the pull request against `main`. Allocation counts, which the decisions below reduce, do not suffer from the bias of CPU profiles on macOS, where the profiler attributes time to system calls that do not cost it.

## Goals / Non-Goals

**Goals:**

- Remove the four costs without changing any output or any exported API.
- Show the effect of each decision with `make bench-compare` and the `benchmark` job.

**Non-Goals:**

- Porting `assert`'s production behavior: `NewNamedLeaf` keeps its check (D2), because it catches porting mistakes in the tests at the cost of one compile per pattern.
- Changing `cleanForKeySlow`: its regular expressions stay as upstream's. Replacing them with string functions would have to reproduce the leftmost-match semantics of the alternations (`objectclassarrow`) and is not needed once the results are cached.
- Changing the iterator API: `clone`, `copyStateFrom` and the trace passed to `newError` keep their meaning.
- The costs listed under Out of scope in proposal.md.

## Decisions

### D1. Reuse deflate writers through a sync.Pool

`deflateRaw` takes a `*flate.Writer` from a package-level `sync.Pool`, calls `Reset` on a local `bytes.Buffer`, writes, closes and puts the writer back. `Reset` makes the writer equivalent to a new one with the same level, so the compressed bytes do not change. The buffer stays local: the returned slice is never shared with a pooled object. The pool's `New` creates the writer at `flate.BestCompression`, as today.

Alternative: one package-level writer behind a mutex. It would serialize the encodings of concurrent renders.

Alternative: a faster compression level. It changes the encoded text of every SVG and of the URL encoding.

### D2. Check the groups of a leaf pattern once per process

`NewNamedLeaf` looks the pattern up in a package-level `sync.Map` from pattern to group count. On a miss it transforms and compiles the pattern as today and stores `NumSubexp()`; a pattern that does not compile still panics with the same message and is not stored. On every call it compares the count with the declared one and panics with the same message on a mismatch. The patterns come from the Go code (no leaf is built from source text), so the map stays small.

Alternative: drop the check, as upstream's production build does. A wrong count would then only show as a wrong group name at match time.

Alternative: check every leaf in a test that walks the command lists. Leaves built while commands execute are not reachable from the command lists.

### D3. Derive the iterator trace from the source lines

The `trace` field of `iterator` is removed; a `trace()` method returns `it.data[:it.pos:it.pos]`. The full slice expression clips the capacity, so a consumer that appends to the trace cannot write into the source lines. `next()` no longer appends, and `clone()` copies `data` and `pos` only. The five `newError` calls of `factory.go` pass `it.trace()`. The consumers of `core.ErrorDiagram.Trace` (`core/diagram.go`, `errdiagram/psystemerror.go`, `svek/cucadiagramfilemaker.go`) only read it.

Alternative: keep the field and share the backing array between clones instead of copying it. It works only while at most one clone appends at a time, an invariant that `candidate` happens to keep today and that nothing enforces.

### D4. A process-wide LRU of normalized skinparam keys, as upstream

`cleanForKey` uses a package-level cache in `internal/klimt/skin_skinparam.go`: a map from key to `container/list` element, in access order, bounded to `cacheCleanForKeyMax = 2000` entries, under a `sync.Mutex`, as upstream's `LinkedHashMap` with `removeEldestEntry`. A hit moves the entry to the back; an insertion past 2000 entries drops the front. `cleanForKeySlow` runs outside the lock, so concurrent renders do not wait on its regular expressions; two goroutines that miss the same key compute the same result. The cached slices are shared and must not be modified: both callers (`SetParam`, `Value`) only range over them. The `cacheCleanForKey` field of `SkinParam` and its initialization in `SkinParamCreate` are removed.

Alternative: an unbounded `sync.Map`. Keys come from the diagram source (a stereotype is part of the key), so a long-running server would grow without limit.

Alternative: stop caching once the cache is full. Upstream's comment rejects it: a long-running server would stay stuck on the first keys it saw.

## Risks / Trade-offs

- [A pooled writer carries state into the next encoding] → `Reset` before each use; a unit test of `internal/code` compares the encoding of several sources, in sequence and in parallel, with the output of a fresh `flate.NewWriter`.
- [State shared across renders changes an output] → a test of the root package renders `Alice -> Bob : hello`, then a C4 diagram with skinparams, then the first diagram again, and expects the two renders of the first diagram to be byte-identical; `TestRenderConcurrent` covers parallel renders. The conformance suite renders every case with the new code.
- [A caller modifies a shared slice of the skinparam cache] → the cache's comment says the slices are read-only, as `Values()` already says of its map; both callers only range over them.
- [Lock contention with `--threads auto`] → the lock covers only the map and list operations, not `cleanForKeySlow`.
- [The LRU drops a key a diagram still needs] → the key is computed again; `cleanForKeySlow` is a pure function, so the output does not depend on what the cache holds.
- [Process memory] → up to 2000 short key lists and the idle writers of the pool, which the garbage collector drops between cycles.

## Migration Plan

Lands after `benchmark-pull-requests`. No API, output or configuration changes. Each decision is one commit, so a regression found later can be reverted alone.
