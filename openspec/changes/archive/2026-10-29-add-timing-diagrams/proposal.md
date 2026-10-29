# Proposal

## Why

Timing diagrams fail today with "timing diagrams are not supported yet" (issue #33): an `@startuml` block with timing commands (`robust`, `concise`, `clock`, `binary`, `analog`, `@<time>`) is drawn as an error image (`internal/core/detect.go`). The three Vega timing cases of the conformance report are not supported yet because of it. Upstream's `timingdiagram` package (66 files) needs no Graphviz: it lays the players out itself. Its parts outside the package are already ported (`uml.Base`, the common commands, `Opale`, the style engine, `WithLinkType`, `ArrowConfiguration`), apart from the JDK behavior it relies on: `java.math.BigDecimal` for the times, `DecimalFormat` for the values of analog players, `SimpleDateFormat` for `use date format` and `Double.parseDouble`.

## What Changes

- Timing diagrams under `@startuml`, as upstream's `TimingDiagram`, with the whole `timingdiagram` package ported:
  - the players: `robust`, `concise`, `rectangle`, `clock` (with `period`, `pulse` and `offset`), `binary` and `analog` (with `between ... and ...`), with their quoted names, aliases, stereotypes, background colors and `compact`
  - the times: `@<time>` blocks with absolute, relative (`@+50`), decimal, named (`@0 as :start`, `@:start+50`), clock (`@clk*2`), hour (`@0:15:00`) and date (`@2024/07/15`) times, and `@<player>` blocks with `<time> is <state>` lines
  - the state changes (`WB is Idle`, `{hidden}`, `{...}`, `{-}`, `{A,B}`, colors and comments), the initial states, and the state labels (`WB has A,B,C`, `WB has "Long name" as L`)
  - the messages (`WU -> WB : URL`, with times and colors), the constraints (`WB@0 <-> @50 : {50 ms lag}`), the highlights (`highlight 0 to 50 #Gold : Peak`) and the notes on players, on one line or several
  - the time axis: `hide time-axis`, `manual time-axis`, `scale 100 as 50 pixels` (with units), `use date format`, `mode compact`, `<player> ticks num on multiple <n>` and `<player> is <n> pixels height`
  - upstream's errors (`Unknown "X"`, `Missing @ line before this`, `No such participant X`, `What time?`, `Unknown time label`, `You have to provide a participant`, `No such element: X`, `Bad value`, `Bad date format`, `No such color`), the crash messages of the commands that upstream cannot run (a state on a clock, an unknown named time with an offset), and upstream's quirks
  - the styles of `timingDiagram` (`robust`, `concise`, `rectangle`, `clock`, `binary`, `analog`, `arrow`, `constraintArrow`, `highlight`, `note`, `timeline`, `timegrid` and stereotypes) and the common commands (title, header, footer, caption, legend, mainframe, `skin`, `skinparam`, `<style>`, sprites, `scale`, `hide footbox`, ...)
- The JDK behavior that timing diagrams need: `BigDecimal` in `internal/jmath`, `DecimalFormat` for doubles and `Double.parseDouble` in `internal/jfmt`, and the `SimpleDateFormat` of `internal/tim` made available to other packages.
- Detection under `@startuml`: timing becomes a supported type, after activity in the new syntax. The "not supported yet" detection keeps archimate only.
- Timing diagrams render in SVG and PNG. TXT and UTXT stay unsupported, as for class diagrams.
- The WASM build renders timing diagrams fully, because they need no graph layout.
- Timing diagrams can be embedded in creole text with `{{` ... `}}`. The example of an embedded diagram of a type that is not supported yet becomes an archimate diagram.
- Conformance: curated cases for every new scenario. The 3 Vega timing cases are compared instead of being reported as not supported yet.
- The README and the landing page list timing diagrams in the table of supported diagrams and among the diagrams that do not need Graphviz, and stop naming them among the types that are not supported yet. The README's conformance counts are updated.

Out of scope:

- The crash image that upstream draws when the drawing of a timing diagram throws (for example when every time is below zero). It has a random quote and a Java stack trace, so it cannot be reproduced. The rendering fails with an error instead, as for gantt and WBS diagrams.
- Upstream's generic text output of the diagrams that have no text output of their own, which TXT and UTXT of timing diagrams fall back to, as class diagrams do.
- Dates before 1582-10-15, which upstream's `GregorianCalendar` computes in the Julian calendar.

## Capabilities

### New Capabilities

- `timing-diagram`: timing diagrams under `@startuml`, with upstream's players, times, state changes, messages, constraints, highlights, notes, time axis, styles and errors

### Modified Capabilities

- `diagram-source`: timing becomes a supported `@startuml` type, tried after activity in the new syntax, and the types that are reported as not supported yet are reduced to archimate
- `styling`: the scenario of an embedded diagram of a type that is not supported yet uses an archimate diagram instead of a timing diagram
- `wasm-build`: timing diagrams join sequence, activity, packet, gantt, mindmap and WBS diagrams as diagrams with a built-in layout that render fully in the WASM build

## Impact

- **Code:**
  - a new `internal/timing` package: the diagram, the factory, the commands, the times and the ruler, the players and their model, and the drawing (`Panels*`, `PlayerFrame`, `TimeArrow`, the shapes, the constraints, the highlights and the notes)
  - `internal/jmath`: `BigDecimal`
  - `internal/jfmt`: `DecimalFormat` for doubles and `ParseDouble`
  - `internal/tim`: an exported form of its `SimpleDateFormat`
  - `internal/engine`: the factory list
  - `internal/core`: the timing entry of the "not supported yet" detection goes away
  - tests that use timing syntax as their example of a type that is not supported yet (`internal/core`, `internal/engine`) switch to archimate
- **Runtime dependency:** none. Timing diagrams do not need Graphviz.
- **Conformance:** new curated cases and committed references under `testdata/conformance/`. The not-supported-yet count of the report drops from 22 to 19, unless some timing cases end up in `known-differences.yaml`. A broken `@startuml` block now gets the error of the timing parser when it scores best, as upstream does, so the error cases of the other types are checked again.
- **Documentation:** `README.md` and `site/index.html` (the table of supported diagrams, the Graphviz notes and the sentence on the types that are not supported yet).
- **Ordering:** the delta spec of `wasm-build` starts from the requirement of `add-wbs-diagrams`, so `add-mindmap-diagrams` and `add-wbs-diagrams` are archived first.
- **Compatibility:** inputs that failed with "timing diagrams are not supported yet" now render, or fail with upstream's error.
