# Proposal

## Why

Every render allocates far more than its diagram needs, and in the command line the garbage collector takes 35-55% of the CPU time: rendering the 91 curated sequence cases takes 0.40 s of user time with the default `GOGC` and 0.26 s with `GOGC=off` (timing: 0.19 s and 0.09 s, gantt: 0.14 s and 0.06 s). Profiles of renders through `Render` (Linux arm64, curated cases) trace most of it to four places:

- `code.deflateRaw` creates a new `flate.Writer` at the best compression level, about 1 MB of tables, for each `<?plantuml-src ...?>` metadata: 55-61% of the allocated bytes of sequence, class, timing and gantt renders, and 5.5% of the CPU time of sequence renders.
- `regex.NewNamedLeaf` compiles its pattern only to check its number of groups, also while commands execute (a timing state, a note, a link, a URL): 6-7% of the allocated bytes of class and timing renders. Upstream makes this check with a Java `assert`, which does not run by default.
- The line iterator of `internal/command` keeps a copy of the lines it has read (the trace of an error), and clones it for each multi-line command it tries: 27% of the allocated bytes of the `preprocessor/stdlib-include` case (C4).
- `SkinParam.cleanForKey` caches the normalized skinparam keys per diagram, so each diagram that sets many skinparams normalizes them again with regular expressions: 8-10% of the CPU time of the description cases and of the C4 case. Upstream caches them once per process (an LRU of 2000 keys).

No issue tracks this, and no conformance case fails or changes: the output of every render stays byte-identical.

## What Changes

- The deflate compressor of `internal/code` is reused across encodings instead of being created for each one. The encoded text does not change.
- `regex.NewNamedLeaf` checks the number of groups of a pattern once per process, and still panics on a wrong count.
- The trace of the command iterator is the prefix of the source lines it has read, not a copy: reading a line and trying a multi-line command no longer copy lines.
- The normalized skinparam keys are cached once per process in an LRU of 2000 keys, as upstream does.

This change lands after `benchmark-pull-requests`: its `BenchmarkRender` and its `benchmark` job measure this change against `main`.

Out of scope:

- The other costs the profiles show: the trie of the standard preprocessor functions built for each render (16% of the CPU time of sequence renders), the 14 drivers registered by each copy of the SVG graphic, the quadratic matching of multi-line command blocks, and a `GOGC` setting of the command line.
- The zlib writer of the iTXt chunk of PNG outputs, which these profiles did not measure.
- The modernizations reported by `go fix`, for a separate change.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None. The change sets `skip_specs: true`: no requirement changes, every output stays the same.

## Impact

**Code:** `internal/code` (deflate), `internal/command/regex` (leaf check), `internal/command` (iterator trace), `internal/klimt` (skinparam key cache), and a test of the root package.

**Conformance:** no case, reference, report count or golden dump changes. The suite must give the same results as before.

**Documentation:** none. The timings of `site/index.html` are those of the v0.3.0 release.

**Compatibility:** no API change. A process keeps up to 2000 normalized skinparam keys, as upstream does, and idle deflate compressors until the garbage collector drops them.
