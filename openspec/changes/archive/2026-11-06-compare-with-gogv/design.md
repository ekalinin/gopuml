# Design

## Context

The change ports no upstream code: `tools/bench`, the landing page and their tests are original.

- `tools/bench` has two implementations (`impl`: a name and a command). `bench.command` runs each with `os.Environ()` plus `GRAPHVIZ_DOT=<dot>`, where the `dot` comes from `-dot`, else `GRAPHVIZ_DOT`, else `PATH`. The table header and the `- Exit status of <name>, <type>` notes use the names. Its test, `TestComparisonTable`, runs fake shell commands that log `$GRAPHVIZ_DOT`, and gives each command its durations by the base name of the command.
- gopuml uses Graphviz `dot` only when `--dot-path`, `GRAPHVIZ_DOT` or `Options.DotPath` names one (`internal/dot`, `Config.NewEngine`). With none named, it lays graphs out with gogv; a `dot` on `PATH` is not used. `tools/refgen` gets gogv the same way, by removing `GRAPHVIZ_DOT` (`useBuiltInLayout`).
- The CLI is built with gogv as a module dependency, which its build info records (`go version -m` gives `github.com/ekalinin/gogv v0.2.0`).
- The landing page draws the comparison as lifelines (`site/style.css`, `.lifelines`): one participant per column (`td.java`, `td.go`, with the colors `--java` and `--go` in light and dark, and a hatch for `go` in forced colors), each cell 34% wide, with a bar of `--t / --max`. `checkLandingRelease` compares every row of the notes' table, header included, with the rows of the page, whatever the number of columns, and checks the bars against the times.
- The curated type `activity` holds cases in the new syntax and legacy cases, which are laid out as graphs.

## Goals / Non-Goals

**Goals:**

- One `make bench` run measures the three columns on the same machine, in the same order for every row.
- The next release's preparation only replaces numbers: the page already styles a third column.

**Non-Goals:**

- Changing how the jar runs, or what the rows measure.
- A choice of the columns by flag: the three always run.

## Decisions

### D1. Three implementations, each with its own environment

`impl` gets the environment of its runs. The jar and `gopuml (dot)` run with `os.Environ()` without any `GRAPHVIZ_DOT`, plus `GRAPHVIZ_DOT=<dot>`; `gopuml (gogv)` runs with `os.Environ()` without any `GRAPHVIZ_DOT`, and with no `--dot-path`. The columns are, in this order, `PlantUML 1.2026.9 (Java)`, `gopuml (dot)` and `gopuml (gogv)`: the first two measure what the two columns of the earlier releases measured, and the third is added. The notes of a failing run name the column (`- Exit status of gopuml (gogv), class: ...`).

Alternative: name the `dot` of `gopuml (dot)` with `--dot-path`. The jar reads `GRAPHVIZ_DOT`, so one variable for both keeps "the same `dot`" a single setting.

### D2. Every row has the three columns

The start time and every diagram type run the three columns, with no list of the types that use a graph layout. For types without one, the two gopuml columns measure the same work, and their near-equal times show that the choice of the layout costs nothing there.

Alternative: run `gopuml (gogv)` only for graph types and show a dash elsewhere. It needs a list of graph types in `tools/bench`, and `activity` mixes both kinds of cases.

### D3. The gogv version comes from the build info of the CLI

The environment gets `- gogv: <version>`, the version of the module `github.com/ekalinin/gogv` in the build info of the CLI under test (`debug/buildinfo`), or the version of its replacement when `go.mod` replaces it, without the `v` of its tag, as gogv prints its own version (`gogv 0.2.0`): the checks of the landing page read every `vX.Y.Z` on the page as a release of gopuml. When the build info cannot be read or has no gogv, the line is `- gogv: unknown`, as for the fake CLI of the test, which is a shell script. The line follows `- Graphviz:`.

Alternative: parse `(gogv 0.2.0)` from the output of `--check-graphviz`. That ties the tool to a message text and runs the CLI once more.

### D4. A third participant on the page

`site/style.css` styles `td.gogv` as it styles `td.java` and `td.go`, with a color `--gogv` in light and dark and a forced-colors pattern different from the solid `java` bar and the hatched `go` bar. Its hue is the nearest cool hue to `--go`, as both columns are gopuml, that the palette validator still tells apart from it: a violet (`#6a4fc9` light, `#8a74ee` dark). Teals and blues next to `--go` fail its normal-vision floor (ΔE below 15). A table with a `td.gogv` cell (`.lifelines:has(td.gogv)`) gives each participant cell a width that leaves room for three columns; the two-column table of v0.4.0 keeps its 34%. The page's HTML does not change now: the preparation of the next release writes the three-column table, with `class="gogv"` on the cells of the third column.

Alternative: a third color unrelated to `--go`. The two gopuml columns would read as unrelated implementations.

### D5. The checks of the scenarios

- "release: Layout of each column": `TestComparisonLayout` in `tools/bench` runs the fake commands with `GRAPHVIZ_DOT` set to another `dot` than `-dot`, and checks from their log that the jar and `gopuml (dot)` saw the `dot` of `-dot` and that `gopuml (gogv)` saw an empty `GRAPHVIZ_DOT`. `TestComparisonTable` checks the three columns, the gogv line and the notes per column; the fake durations are keyed by the column, which the fake CLI cannot tell from its base name, so the test's `timed` reads the column from the environment of the command.
- "landing-page: Comparison with three columns": `TestLandingComparisonColumns` in the root package builds a page, a README and notes with a three-column comparison, checks that `checkLandingRelease` accepts them, and that it fails when the third column of one row is removed from the page.

## Risks / Trade-offs

- [The third column adds runs of the CLI] → each gopuml run of a row takes tens to hundreds of milliseconds against close to a second for the jar, so a run of `make bench` grows by a few seconds.
- [The colors of the page are checked by eye] → the verification of D4 takes screenshots of a local preview with a three-column table, in light and dark.
- [A preview table could be committed by mistake] → the preview is built in a temporary directory from `make site` output, and `site/index.html` stays unchanged in the diff of the pull request.
