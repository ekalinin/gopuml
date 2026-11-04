# Design

## Context

See proposal.md (Why) for the motivation. The engine is already a seam: `internal/svek` lays a DOT graph out through the `dot.Engine` interface (`Layout(ctx, dot) ([]byte, error)`, the SVG of `dot -Tsvg`). The engines that tell their Graphviz version implement `svek.VersionedEngine`. The engines that do not get `modernGraphvizVersion` (14.0.1). The svek package documentation already says that "a pure Go engine can replace it".

Upstream classes involved, and what Go has for them:

| Upstream (plantuml-mit 1.2026.9) | Go today | After this change |
|---|---|---|
| `dot/GraphvizRuntimeEnvironment`, `AbstractGraphviz`, `GraphvizLinux`, `GraphvizWindowsOld` (find `dot`: explicit, `GRAPHVIZ_DOT`, `PATH`, standard locations) | `dot.Config`, `dot.SearchDotExe`, `dot.Config.Find`, `dot.Config.NewExec` in `internal/dot/graphvizruntimeenvironment.go` | unchanged; called only when a `dot` is named |
| `dot/Graphviz`, `GraphvizVersionFinder`, `svek/GraphvizImageBuilder` (run `dot -Tsvg`, `dot -V`) | `dot.Exec` in `internal/dot/graphviz.go` | unchanged (the named `dot`) |
| `dot/GraphvizUtils`, `cli/OptionPrint.printCheckGraphviz` (`-testdot`) | `dot.CheckGraphviz`, `dot.AddDotStatus` in `internal/dot/graphvizutils.go` | a branch for gogv when no `dot` is named |
| `atmp/CucaDiagram` choosing `CucaDiagramFileMakerSmetana` when `isUseSmetana()` or `dotIsAvailable() == false`, otherwise `CucaDiagramFileMakerSvek` | `svek.Layout.resolve` in `internal/svek/cucadiagramfilemaker.go`: always the dotexec engine, or not found | the engine of `dot.Config.NewEngine`: gogv, or `dot.Exec` for a named `dot` |
| `sdot/*`, `smetana/*`, `h/*`, `gen/*` (Smetana, a Java port of Graphviz used in process) | not ported | not ported; gogv has its role as the in-process engine (D2) |
| `elk/*`, `vizjs/*` | not ported | not ported |

What is missing is not upstream code but the in-process engine: `github.com/ekalinin/gogv` exposes `gogv.Render(ctx, src, format) ([]byte, []gogv.Warning, error)`, which writes the bytes of `dot -T<format>`, is safe for concurrent use, stops on context cancellation, and builds for `GOOS=js`. It follows Graphviz 16.1.0, and its `dot -V` prints `dot - graphviz version 16.1.0 (gogv 0.2.0)`, the version of gogv without its leading `v`, or `devel` for a local build. gogv is public since v0.1.0. v0.2.0 exports that Graphviz version as `gogv.GraphvizVersion` (`"16.1.0"`), and its outputs are byte for byte those of v0.1.0.

What was measured before writing this design:

- The conformance suite, run with gogv's `dot` command as `GRAPHVIZ_DOT` (before the release, then with v0.1.0 and with v0.2.0, with the same results): 12 failures, all in graph-based types (activity 1, chen_eer 4, class 1, description 1, state 5). The other types are unchanged. The list is in proposal.md (Impact).
- For those 12 cases, gogv printed no warning: they are layout differences (node order, x positions) between gogv and Graphviz 14.0.1, not unsupported features. For example, `chen-eer-diagram/entities-relationship-and-attributes` is 4.65in wide with gogv and 3.34in with Graphviz 14.0.1. Graphviz 15.0.0 lays out the graphs of 10 of these cases with the same `-Tplain` size as 14.0.1, so the differences come from gogv, not from the Graphviz version.
- Over the whole corpus (298 DOT graphs), gogv's only warning is `node ..., port P unrecognized` (8 times), which Graphviz prints too. The graphs use `shape=rect|circle|ellipse|plaintext|diamond|point|octagon`, `style=rounded|solid|invis` and `rankdir=LR`, all supported by gogv. gopuml can also write `splines=ortho` (`skinparam linetype ortho`), which no case uses and gogv does not support yet.
- The reference image sets `ENV GRAPHVIZ_DOT=/usr/local/bin/dot`, and `tools/refgen` renders this implementation in process, so with `GO_IN_IMAGE=1` this implementation sees that variable.

## Goals / Non-Goals

**Goals:**

- Graph-based diagrams render without an installed Graphviz, natively and in WASM, with the layout of gogv.
- One choice of the engine, in `internal/dot`, used by the native and the WASM builds, the CLI check and the library.
- The errors of a gogv layout reuse the existing typed errors (`ErrGraphvizNotFound`, `ErrGraphvizTimeout`, `ErrGraphvizFailed` and their `dot` counterparts), so no API changes.
- The conformance run measures gogv, not the Graphviz of the reference image.

**Non-Goals:**

- Porting upstream's Smetana (`net/sourceforge/plantuml/sdot`, `smetana`, `h`, `gen`): gogv has its role as the in-process engine, and a second Graphviz port inside gopuml would duplicate it. The 38 cases with `!pragma layout smetana` stay known differences, now against gogv.
- Porting ELK (`net/sourceforge/plantuml/elk`) and VizJs (`net/sourceforge/plantuml/vizjs`): the pragmas keep using the engine of the render, as today.
- Removing the `PATH` and standard-location search of `dot.SearchDotExe`: it stays the port of `GraphvizRuntimeEnvironment`, but the engine choice no longer reaches it (D2).
- Fixing gogv's layout differences inside gopuml: they are fixed in gogv, or listed as known differences (D9).
- Passing gogv's warnings on (D6).
- `make bench`: it passes one `dot` to both implementations through `GRAPHVIZ_DOT`, so it keeps comparing them with that `dot`.

## Decisions

### D1. The gogv engine in `internal/dot/gogv.go`

A new file `internal/dot/gogv.go` (`Provenance: original`) defines `type Gogv struct { Timeout time.Duration; Render func(ctx context.Context, src, format string) ([]byte, []gogv.Warning, error) }`. A nil `Render` means `gogv.Render`, as the nil functions of `dot.Host` mean the operating system; tests set it to fake a failure. `Gogv` implements `dot.Engine` (`Layout` calls `Render(ctx, dot, "svg")`) and `svek.VersionedEngine` (`GraphvizVersion` returns `NewGraphvizVersion` of its version line, D4). Like `Exec`, it holds no mutable state and is safe for concurrent use.

Alternative: put the engine in `internal/svek`. Rejected: `internal/dot` already holds the engines, their errors and the check, and svek only uses `dot.Engine`.

### D2. The engine choice: `dot.Config.NewEngine`

`Config.NewEngine() (Engine, error)` in `internal/dot/graphvizruntimeenvironment.go` returns `c.NewExec()` when `GetenvGraphvizDot(c.DotPath, c.Host)` is not empty (a `dot` is named by `--dot-path`, `Options.DotPath` or `GRAPHVIZ_DOT`), with the `*NotFoundError` of a missing named `dot` as today, and `&Gogv{Timeout: c.Timeout}` otherwise. `svek.Layout.resolve` calls it instead of `NewExec`. In the WASM build (`execSupported` false), `resolve` uses `&dot.Gogv{...}` without asking `Config`, because a program cannot run there even when the JavaScript host passes a `GRAPHVIZ_DOT` in the environment. The version cache keyed by the executable stays for `*dot.Exec` only.

This keeps upstream's rule for a named `dot` (`GRAPHVIZ_DOT` is used even when no file exists there) and puts gogv where upstream puts Smetana when `dotIsAvailable()` is false, except that gogv is also used when a `dot` is on `PATH`, as the user asked.

Alternative: run `dot` when one is found on `PATH` and gogv only otherwise, as upstream does with Smetana. Rejected by the user: with Graphviz installed, gogv would never run, and the output would depend on the host.
Alternative: gogv only, removing the dotexec engine. Rejected by the user: there would be no way left to get Graphviz's own layout, or `splines=ortho`.

### D3. Timeout, cancellation, failures and panics

`Gogv.Layout` runs `Render` under `context.WithTimeout(ctx, timeout)`, with `DefaultTimeout` (30 seconds) when `Timeout` is zero or negative, as `Exec`. When the deadline of its own timeout stops the layout (the parent context is not done), it returns `&TimeoutError{Path: "gogv", Timeout: timeout}`. When the parent context is done, it returns the context's error, wrapped as `Exec` wraps it. A `recover` turns a panic of gogv into `&FailureError{Path: "gogv", ExitCode: -1, Err: fmt.Errorf("panic: %v", v)}`. An error of gogv becomes `&FailureError{Path: "gogv", ExitCode: -1, Err: err}`, and output without `<svg` becomes `&FailureError{Path: "gogv", Reason: "output is not SVG"}`, as for `Exec`. The error images and the typed errors of the root package (`graphvizError` in `plantuml.go`) then work unchanged, and a failing diagram does not stop the others.

Alternative: run `Render` in a goroutine and abandon it at the deadline. Rejected: the goroutine would keep running and holding memory. gogv already returns when its context is done.

### D4. The version line, from gogv

gogv v0.2.0 exports the Graphviz version it follows as `gogv.GraphvizVersion` (`"16.1.0"`), the constant that its own `dot -V` and the comment of its SVG output use. gopuml writes the version line as gogv's `dot -V` does: `dot - graphviz version ` + `gogv.GraphvizVersion` + ` (gogv ` + module version without its leading `v` + `)`, so `dot - graphviz version 16.1.0 (gogv 0.2.0)` with v0.2.0. The module version comes from `runtime/debug.ReadBuildInfo` (the dependency `github.com/ekalinin/gogv`), or is `devel` when it is not there, as in gogv's `dot -V`. `NewGraphvizVersion` of that line gives 1601. Like 1400 for 14.0.1, it triggers none of the version workarounds (`<= 228`, `== 230`, `239` or `240`), so the DOT that gopuml writes does not change.

Alternative: a gopuml constant `"16.1.0"`. Rejected: it would go stale silently when gogv follows a later Graphviz.
Alternative: lay out a graph and read the version from the `<!-- Generated by graphviz version ... -->` comment of the SVG. Rejected: it depends on a comment and costs a layout.

### D5. `--check-graphviz` without a named `dot`

`AddDotStatus` takes the gogv branch when `GetenvGraphvizDot(c.DotPath, c.Host)` is empty. In the INSECURE profile it prints `The environment variable GRAPHVIZ_DOT has not been set`, then `Dot executable is built in (gogv)` instead of a path. Then, in upstream's wording, `GraphViz: <version line of D4>`, the version check, and the test layout of `digraph foo { test; }` through `Gogv`, with upstream's messages for empty and unreadable output and `Installation seems OK. File generation OK` on success (exit status 0). With a named `dot`, the check is unchanged.

Alternative: report that Graphviz is not installed, with exit status 16, as upstream does without `dot`. Rejected: graph-based diagrams render in that case, so the check would report a failure that does not exist.

### D6. gogv's warnings are dropped

`Gogv.Layout` ignores the warnings of `Render`, as `Exec` ignores what `dot` writes on stderr when it succeeds and as upstream does. They name Graphviz features that gogv does not support yet. The corpus has none apart from the `port P unrecognized` that Graphviz prints too (see Context).

Alternative: print them on stderr in the CLI. Rejected: upstream prints nothing there, the CLI cases compare stderr, and the library has no channel for warnings.

### D7. The conformance run lays this implementation out with gogv

`tools/refgen`, when it renders this implementation (`-impl go`), unsets `GRAPHVIZ_DOT` in its own process before rendering. This covers the cases rendered in process, the CLI cases (whose environment is `os.Environ()` plus the case's `env`, so a case can still name a `dot`) and the server cases. The Makefile keeps `GO_IN_IMAGE=1` and CI keeps its command. On any host, `make conformance REQUIRE=all` then gives the results of CI for this implementation, because it no longer depends on the host's Graphviz. The README's arm64 wrapper is removed. A test of refgen checks that a render of `graph-layout/class-layout` with `GRAPHVIZ_DOT` pointing to a fake `dot` that fails does not run it.

Alternative: `-e GRAPHVIZ_DOT=` in the Makefile's `REFGEN_GO`. Rejected: it covers only `GO_IN_IMAGE=1`, and a developer's `GRAPHVIZ_DOT` would still leak into native runs.

### D8. WASM and the playground

In the WASM build, graph-based diagrams render through D2. `cmd/plantuml-wasm/main.go` drops the override of the message of `ErrGraphvizNotFound` ("graph layout is not available in this build"), which can no longer happen there, and keeps the error kind `graph-layout-unavailable` in `errorKinds`, which is part of the JavaScript API. The WASM tests of class, description, state, legacy activity and Chen EER diagrams compare the SVG with the native library's output, as those of packet and gantt diagrams do. The playground test renders a class diagram. The playground of the page runs the WebAssembly module of the release it names, so its note in `site/index.html` and the README's sentence about the playground change with the next release, as in add-chen-eer-diagrams.

### D9. The gogv release, and the 12 differing cases

gogv v0.2.0 is public, exports `GraphvizVersion` (D4) and is tagged. gopuml requires it in `go.mod`, and `NOTICE` names gogv among the linked Go code. Its outputs are those of v0.1.0, so the 12 cases fail with it as before. For the 12 cases, each difference is looked at in gogv (gopuml's DOT of the case, laid out by gogv and by Graphviz 14.0.1, compared with `-Tplain`). A difference that gogv can fix gets fixed there in a v0.2.x release, and gopuml moves to that tag. The cases that still differ get an entry in `known-differences.yaml`: "laid out by gogv, whose node placement differs from Graphviz 14.0.1 for this graph (<what differs>)". The reason of the 38 Smetana entries changes from "this release with dot" to "this release with gogv".

Alternative: merge with all 12 cases as known differences and fix gogv later. Not chosen as the default, since the fixes are in a project the user maintains. It stays the fallback for differences that gogv cannot fix before the merge.

### D10. Cases first, and the cases whose tests change

`TestScenarioCoverage` reads the delta specs of this change, so the `.go.yaml` cases of the new scenarios come first, pointing to their tests in the packages of D1 to D8:

- `graph-layout/named-dot-layout` ("Layout with the named dot"): root package, a fake `dot` named by `GRAPHVIZ_DOT` that records that it ran
- `graph-layout/named-dot-missing` ("Named dot is missing"): root package, the existing `TestRenderMissingGraphviz` (`Options.DotPath` naming a missing file)
- `graph-layout/builtin-timeout` ("Built-in layout timeout"): `internal/dot`, `Gogv` with a `Render` that waits for its context
- `graph-layout/builtin-failure` ("Built-in engine fails on one diagram"): `internal/svek`, two class diagrams exported through `svek.Layout{Engine: &dot.Gogv{...}}` whose `Render` panics on the first call only, as the other engine tests of `layout_test.go` do
- `graph-layout/check-named-dot` ("Check of a named dot"): `internal/cli`, the existing `TestRunCheckGraphviz`
- `upstream-compatibility/graph-cases-builtin-engine` ("Graph cases laid out by the built-in engine"): `tools/refgen` (D7)

The existing cases keep their scenarios, but their tests change with the behavior: `class-without-dot` (`TestClassDiagramWithoutDot`, which now checks a drawing), `dot-found-on-path` (`TestNewEngineIgnoresDotOnPath`), `smetana-pragma` (`TestSmetanaPragmaUsesEngineOfRender`), `library-api/missing-graphviz` and `cancelled-context` (with gogv and with a named fake `dot`), the five `wasm-build/*-diagram-error` cases (renamed without `-error`, with tests that compare the SVG), and `landing-page/playground`. Their `why` texts are updated. No `@startuml` case changes, so the golden dumps stay.

## Risks / Trade-offs

- [gogv lays out some graphs differently from Graphviz 14.0.1, so up to 12 more cases become known differences] → D9 fixes them in gogv first. Each one left is named with its reason, and a named `dot` gives upstream's layout.
- [`skinparam linetype ortho` draws splines with gogv, without a warning] → listed as out of scope in proposal.md. The README's Graphviz section says that a named `dot` draws orthogonal lines.
- [gopuml depends on an external module] → gogv v0.2.0 is public and tagged, and `go.mod` names that tag, not a pseudo-version of a commit.
- [A gogv layout that does not check its context often keeps running after the timeout] → the timeout test of D10 measures it. If gogv returns late, the fix goes into gogv, as for D9.
- [The WASM module grows] → gogv's own WASM build is 4.6 MB with the Go runtime, against 51 MB for gopuml's today. The size printed by `make wasm` is reported in the pull request.
- [Users with `GRAPHVIZ_DOT` set in their environment, as in the reference image, keep getting Graphviz] → intended (a named `dot`), and stated in the README.

## Migration Plan

1. gogv: public, `GraphvizVersion` exported and tagged (done in v0.2.0); the fixes of D9 in a v0.2.x release, if any.
2. gopuml: this change on a `feat/` branch, with v0.2.0, or the later tag of step 1, in `go.mod`.
3. Rollback: revert the pull request. Users who want Graphviz's own layout without a rollback set `GRAPHVIZ_DOT` or `--dot-path`.
