# Proposal

## Why

Gantt diagrams (`@startgantt`, `@startproject`) fail today with "gantt diagrams are not supported yet" (issue #28). They are the largest group of cases that are not supported yet: 17 of the 50 cases of the conformance report (the 16 cases of `vega/xgantt/` and `vega/nonreg/warnings/gantt_labels`). Upstream's `gantt` package (167 files) needs no Graphviz. Its export, styles, notes, text, `Real` positions and ASCII tables use parts that are already ported (`uml.Base`, `uml.ExportTitled`, the common commands, the UBrex engine, `Opale`, `sequencediagram.Real`, `asciiverse.ATable`).

## What Changes

- Gantt diagrams under `@startgantt` and `@startproject`, as upstream's `GanttDiagram`, with the whole `gantt` package ported:
  - the natural-language sentences: a subject (`Project`, `today`, a task, a resource, a day of the week, a date, a range of dates, `separator`, working hours) followed by verb phrases joined with ` and `, with `then`, `it` and `she`, `he`, `they`
  - tasks: durations (`lasts`, `requires`), starts and ends (dates, times of day, `D+N`, `N days after start`), milestones (`happens`), `occurs from [A] to [B]`, aliases (`as`), stereotypes, colors, completion, `is deleted`, pauses, links, `is displayed as` and `displays on same row as`
  - constraints between tasks and their arrows (`starts at [A]'s end`, delays in days, working days and weeks, `[A] -> [B]`, `with red dotted link`)
  - the calendar: closed and opened days of the week and dates, days opened for one task, colored and named days, `then N days`, `today`, vertical separators, working hours and `hide closed`
  - resources with load percentages, their days off and on, and the footer of resource loads
  - groups, horizontal separators and notes
  - the time scales `daily`, `weekly`, `monthly`, `quarterly` and `yearly`, with `zoom`, the week numbering options, `print between` and `hide footbox`
  - the columns of the task table, `language` (de, es, fr, ja, ko, ru, zh) and the common commands (title, header, footer, caption, legend, `skin`, `skinparam`, `<style>`, sprites, `scale`, ...)
  - upstream's errors and quirks, including the user-visible bugs of upstream's syntax (for example `works on`, which always fails), which this change keeps
- Family detection: `gantt` becomes a supported family. The "not supported yet" error for `@startgantt` and `@startproject` goes away.
- Gantt diagrams render in SVG and PNG. TXT and UTXT render upstream's table of tasks (task, start, end and duration), drawn in ASCII for both formats, as upstream does.
- The WASM build renders gantt diagrams fully, because they need no graph layout.
- Gantt diagrams can be embedded in creole text with `{{gantt` ... `}}`.
- Conformance: curated cases for every new scenario. The 17 Vega gantt cases are compared instead of being reported as not supported yet.
- The README and the landing page list gantt diagrams in the table of supported diagrams and stop naming them among the types that are not supported yet. The README's conformance counts are updated.

Out of scope:

- The crash image that upstream draws inside a gantt diagram when the drawing throws. It has a random quote, a random icon and a Java stack trace, so it cannot be reproduced. The rendering fails with an error instead, as for packet diagrams.
- Fixing upstream's bugs in the gantt syntax: the port keeps them.

## Capabilities

### New Capabilities

- `gantt-diagram`: gantt diagrams under `@startgantt` and `@startproject`, with upstream's syntax, scheduling, drawing, TXT table and errors

### Modified Capabilities

- `diagram-source`: `gantt` becomes a supported family next to `uml` and `packetdiag`
- `output-formats`: TXT and UTXT output covers gantt diagrams (upstream's table of tasks) in addition to sequence diagrams
- `wasm-build`: gantt diagrams join sequence, activity and packet diagrams as diagrams with a built-in layout that render fully in the WASM build
- `packet-diagram`: the rejection of TXT and UTXT is no longer described as the rule for every diagram type except sequence, because gantt diagrams have TXT too

## Impact

- **Code:**
  - a new `internal/gantt` package: the diagram, the factory, the commands and the natural-language subjects and complements, the model, the solver, the time scales and headers, the task drawings, the task table and the TXT export
  - a new `internal/gantt/ngm` package for upstream's load model (`ngm` and `ngm/math`)
  - a new `internal/jtime` package that reproduces the `java.time` behavior that gantt uses: dates, durations, week numbers and the `DateTimeException` messages
  - upstream's `I18nTimeData` tables
  - the missing UBrex builders in `internal/klimt` (`UBrexOneOrMore`, `UBrexZeroOrMore`, `UBrexUpto`)
  - `internal/engine`: the factory list, and TXT and UTXT allowed for gantt diagrams
  - `internal/core`: the supported families
- **Runtime dependency:** none. Gantt diagrams do not need Graphviz.
- **Conformance:** new curated cases and committed references under `testdata/conformance/`. The not-supported-yet count of the report drops from 50 to 33, unless some gantt cases end up in `known-differences.yaml`.
- **Documentation:** `README.md` and `site/index.html` (the table of supported diagrams, the sentence on the types that are not supported yet and the Graphviz note).
- **Compatibility:** inputs that failed with "gantt diagrams are not supported yet" now render.
