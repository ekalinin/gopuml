# Design

## Context

See proposal.md for the motivation. The current state that shapes the approach:

- Upstream's `gantt` package has 167 files in 12 subpackages (`core`, `lang`, `ulang`, `command`, `data`, `draw`, `draw/header`, `timescale`, `time`, `solver`, `ngm`, `ngm/math`). They depend on each other in cycles (`core.TaskImpl` uses `GanttDiagram`, `time.TimePoint` uses `core.PrintScale`, `data.TaskDrawRegistryData` uses `draw`, which uses `data`), except `ngm` and `ngm/math`, which only depend on each other and `java.time`. Several classes are never used (`PiecewiseTimeProjection`, `PiecewiseTimeType`, `PixelInstant`, `HeaderPiece`, `NGMTaskFixedDuration`, `time/Instant`, `ImpossibleSolvingException`, which nothing throws), and the `TeaVM.a()` branches only hold Java `assert` statements, which are off without `-ea`.
- `GanttDiagram` is a `TitledDiagram` that draws through `getTextBlock`, like `PacketDiagram`, with no margins by default. Its layout runs once in the constructor of `GanttDiagramMainBlock`, which uses `Real` positions for the rows. The Go side already has `uml.Base`, `uml.NewFactory`, `uml.AddTitleCommands`, `uml.AddCommonCommands2`, `uml.AddCommonScaleCommands`, `uml.ExportTitled`, `klimt.SNameGanttDiagram` and every SName gantt uses, the `ganttDiagram` sections of the bundled skins, `klimt.NewOpale`, `Real` and `Segment` in `internal/sequencediagram`, `asciiverse.ATable` and `InfinitePlan`, the link styles of `internal/decoration`, `jmath.Round` and the `uml.Throw` helpers with the Java exception names.
- The natural-language commands match with UBrex. `internal/klimt` has the UBrex engine and the builders `UBrexLeaf`, `UBrexConcat`, `UBrexNamed`, `UBrexOptional` and `UBrexOr`, but not `UBrexOneOrMore`, `UBrexZeroOrMore` and `UBrexUpto`, which gantt's `Words` and `lang` use.
- Gantt uses `java.time` throughout: `LocalDate`, `LocalDateTime` (with `MIN` and `MAX`), `LocalTime` (with `MAX`), `Duration` (`getSeconds`, `toMillis`, `toString`), `DayOfWeek`, `Month`, `YearMonth`, `WeekFields.weekOfYear()` and the `DateTimeException` messages of invalid dates. The Go side has no `java.time` emulation: `internal/tim/javadate.go` reproduces `java.util.Calendar` for `%date`, with other semantics.
- The output depends on Java details that a direct rewrite would lose. The solver keeps the two attributes set last (start, end, load), and a new start never moves a task earlier than a start that is still kept. The start, end and plan of a task are cached and not invalidated by later calendar lines, and `addWorkingTimeRange` does not clear the cache of the calendar. The work of a task is frozen when its line is read (a week is the number of open days of a week at that time). The load integration stops with `IllegalStateException` after 9999 segments. Pixel positions come from `long` milliseconds times a `double` cell width. Week numbers follow `WeekFields`, not ISO weeks.
- `today` without a date uses `LocalDate.now()` in the JVM's default time zone. Nothing else in gantt reads the clock.
- `GanttDiagramMainBlock.drawU` catches every exception and draws a crash image with a random quote, a random icon and the Java stack trace on top of what is drawn. Exceptions in the layout go to `UgDiagram`'s crash report.
- The engine rejects TXT and UTXT for every type other than sequence (`internal/engine/engine.go`). Upstream's gantt TXT is a 4-column `ATable` drawn on an `InfinitePlan` in ATXT for both TXT and UTXT, without the title and the other annotations.
- `core.CheckFamily` reports every family that is not in `supportedTypes` (`Sequence`, `Class`, `Activity`, `Packet`) as not supported yet. `core.FindStartTypes` already maps `@startgantt` and `@startproject` to `core.Gantt`, and `uml/chrome.go` maps `core.Gantt` to the `ganttDiagram` style.
- `TestScenarioCoverage` reads the delta specs of active changes, so every scenario of this change needs a case once the specs are committed. `TestDiagramReferences` of `internal/errdiagram` draws the error image of every curated case whose reference is an error.

## Goals / Non-Goals

**Goals:**

- Port the whole `gantt` package file by file, with the SVG byte-identical to upstream's apart from the encoded source, and the TXT output identical.
- Keep upstream's quirks and bugs: the scenarios pin the ones users can see.
- Keep the results of every other diagram type unchanged. The shared code that changes is the list of supported families, the factory list, the TXT check of the engine and the new UBrex builders.

**Non-Goals:**

- The crash image of upstream. A drawing that panics returns an error, as for the other diagram types.
- A clock in the public API. Only the tests of the gantt package replace the clock.
- Porting the unused classes listed in Context, and the `TeaVM.a()` asserts.

## Decisions

### D1. Package layout

`internal/gantt` holds every subpackage of upstream's `gantt` except `ngm`, in one Go package, because the Java subpackages depend on each other in cycles. The files keep upstream's grouping with a prefix: `diagram.go`, `factory.go`, `commands*.go` (the regex commands and `NaturalGanttCommand`), `ulang.go`, `lang_*.go` (subjects, complements, words, verbs and `DayPattern`), `core_*.go` (tasks, resources, moments, time ranges), `calendar.go` (`OpenClose`, `DaysAsDates`, `DayCalendarData`), `constraint.go`, `solver.go`, `time*.go` (`TimePoint`, `PrintScale`, `TimePointFormat`, `MonthUtils`, `DayOfWeekUtils`, `WeekNumberStrategy`), `i18n.go` (`GanttI18n` and upstream's `utils/I18nTimeData`, which only gantt uses), `data_*.go`, `timescale.go`, `header_*.go`, `draw_*.go`, `mainblock.go`, `tasktable.go`, `export.go` and `txt.go`. Each file has the `Provenance` lines of the Java files it ports.

`internal/gantt/ngm` holds `ngm` and `ngm/math` (`Load`, `NGMAllocation`, `NGMTaskFixedTotalEffort`, `Fraction`, `Segment`, the piecewise constants and `LoadIntegrator`). It is a leaf package, so it can be tested on its own. The working hours reach it as pairs of `jtime.LocalTime`, without upstream's `PValue` marker.

Alternative: one Go package per Java subpackage. The cycles make it impossible without interfaces that upstream does not have.

### D2. java.time in `internal/jtime`

A new package `internal/jtime` (`Provenance: original`, next to `jmath` and `jfmt`) reproduces the `java.time` behavior that gantt uses:

- `LocalDate` as a proleptic Gregorian epoch day, `LocalTime` and `LocalDateTime` with nanoseconds, so that `LocalTime.MAX`, `LocalDateTime.MIN` and `LocalDateTime.MAX` behave as Java's
- `Duration` with Java's `getSeconds` (floor), `toMillis` (an `ArithmeticException` on overflow), `plus`, `multipliedBy`, `between`, `isNegative` and `toString` (`PT25H1M1S`, `PT-10H`)
- `DayOfWeek` (Monday is 1) and `Month` with their enum names, `YearMonth`, and `ChronoUnit.DAYS.between`
- `WeekFields.of(firstDayOfWeek, minimalDays).weekOfYear()`, with the JDK's formula, which gives week 0 and week 53 where ISO weeks differ
- the validation of `LocalDate.of`, `LocalTime.of` and `LocalDateTime.of`, which panics with `uml.Throw` and the JDK's `DateTimeException` messages (`Invalid date 'FEBRUARY 30'`, `Invalid value for MonthOfYear (valid values 1 - 12): 13`, ...). `uml` gets the exception name `java.time.DateTimeException`

Alternative: Go's `time.Time` in UTC. `time.Date` turns February 30 into March 1 instead of failing, `LocalDateTime.MIN` and `MAX` are outside its range, and its ISO weeks are not Java's week of year.

### D3. The current day

`internal/gantt` reads the current date through an unexported package variable (`now`, `time.Now` by default, in the local time zone, as the JVM's default zone). The test of the scenario "Current day" replaces it. The compared cases always give `today` a date, so their references do not depend on the day they are rendered.

Alternative: a clock in `core.RenderContext` or the public options. Only `today` needs it, and the preprocessor's clock (`tim.Options.Now`) is not exposed either.

### D4. GanttDiagram on top of uml.Base

- `GanttDiagram` embeds `uml.Base` with the type `core.Gantt`, as `PacketDiagram` does, so the SVG has `data-diagram-type="GANTT"` and the annotations, skins and styles come from the shared code. `Description` returns `(Gantt)`, and a `DefaultMargins` method returns no margins, as upstream's `getDefaultMargins`.
- `CheckFinalError` runs upstream's `checkFinalError` (`initMinMax` and the 20-year check). A panic there becomes the "Fatal crash error" diagram through `core.createSafely`, as upstream's `PSystemBuilder` does.
- `ExportDiagram` renders TXT and UTXT with `exportTxt` (D9), and the other formats with `uml.ExportTitled` and upstream's `getTextBlock`. The text block gets the string bounder of the file format option first, as `internal/sequencediagram/export.go` does, and runs the layout once.
- Styles are built inline with `klimt.StyleSignatureBasicOf(SNameRoot, SNameElement, SNameGanttDiagram, ...)`, the Go form of upstream's `StyleQueries.GANTTDIAG`. Tasks, separators and constraints keep the style builder of the time they were created, as upstream does, so a `<style>` after a task does not change it.

### D5. Parsing

- The factory adds the commands in upstream's order: `uml.AddTitleCommands`, `uml.AddCommonCommands2`, `uml.AddCommonScaleCommands`, the nine `NaturalGanttCommand`s, then the 18 regex and UBrex commands.
- `NaturalGanttCommand`, `ParsedSentence`, `VerbPhraseMatcher` and `VerbPhraseAction` are ported on the UBrex engine of `internal/klimt`. `UBrexOneOrMore`, `UBrexZeroOrMore` and `UBrexUpto` are added to `internal/klimt/ubrex_builder.go` with the `Provenance` lines of their Java files. UBrex's semantics give upstream's behavior without extra code: ordered choice, possessive repetition, lazy `Upto`, case folding of ASCII letters only, `〴s` for the space character only, `〴d` with the underscore, and a match of a prefix of the line.
- The regex commands use `uml.Single` and the multiline note command uses the multiline helpers of `internal/uml`, with upstream's patterns.
- Numbers are parsed with `uml.JavaParseInt` and `uml.JavaParseDouble`, so that `1_0` and overflows give upstream's `NumberFormatException`. Java `int` arithmetic (weeks times open days, months times 30) uses `int32`, so that it wraps as Java's does.
- A panic in a command becomes upstream's "You should send a mail to plantuml@gmail.com with this log" error through the command factory, as for the other diagram types.

### D6. Model and scheduling as upstream

- The tasks and the resources are kept in insertion order, as upstream's `LinkedHashMap`s, because the drawing, the TXT rows, `then`, notes and the resource footer depend on it. Upstream's `HashMap`s and `HashSet`s are only read by key or reduced to a minimum or a maximum, so Go maps give the same result.
- `Solver` keeps an ordered list of at most two attributes with upstream's rule for a new start.
- The caches of `TaskImpl` (start, end, plan) and of `OpenClose` are ported with the same invalidation, including the cases where upstream does not invalidate them.
- `Fraction` uses `int64`, where Go's wraparound and truncated division are Java's. `LoadIntegrator` and the segment iteration keep the order of the operations, the truncation of `wholePart` and the limit of 9999 segments.
- Positions keep Java's order of operations: milliseconds times the cell width divided by 86,400,000, the cell width as the font size times 1.6 times the scale, and `jmath.Round` where Java calls `Math.round`. `int` divisions stay integer divisions.

### D7. Shared code reuse

- `Real` and `Segment` are used from `internal/sequencediagram`, which gantt imports. There is no import cycle, and gantt only uses the exported functions (`RealUtilsCreateOrigin`, `AddFixed`, `AddAtLeast`, `CurrentValue`, `CompileNow`, `NewSegment`, ...). The comment of `real.go` that says it only serves Teoz is updated.
- Notes use `klimt.NewOpale`, arrows the link styles of `internal/decoration`, links `klimt`'s URL builder, the TXT table `internal/asciiverse`, and the resource names `font.Serif`.

Alternative: move `Real` and `Segment` to a new package, as upstream's `net.sourceforge.plantuml.real`. It would change the sequence diagram code for no change of behavior.

### D8. Errors and crashes

- The errors of the commands and of `checkFinalError` go through the shared code, which adds ` (Assumed diagram type: gantt)` and the line, as for packet diagrams. The curated error cases check the lines (scenarios of "Gantt errors").
- `CommandLabelOnColumn` adds upstream's warning with `uml.Base.AddWarning`.
- A panic while drawing, such as upstream's `ClassCastException` for a resource in a group, is recovered by `uml.ExportTitled` and returned as an error, where upstream draws its crash image. The runner of the conformance suite does not compare such a case, so it is a `.go.yaml` case.

### D9. Registration, formats and WASM

- `gantt.Factory` (`uml.NewFactory(core.Gantt, ...)`) goes into `engine.Factories` after `packetdiag.Factory`, as in upstream's `PSystemBuilder`.
- `core.Gantt` joins `supportedTypes`. `TestCheckFamily` gets a `@startgantt` check and keeps `@startmindmap` as its unsupported family. The check of `internal/core/builder_test.go` that expects gantt diagrams to be not supported yet uses `@startwbs` instead.
- The TXT check of the engine accepts `core.Gantt` with `core.Sequence`. `exportTxt` draws the `ATable` on an ATXT `InfinitePlan` for TXT and UTXT, as upstream does.
- The WASM build needs no special case, because gantt diagrams do not use Graphviz.

### D10. Documentation

- The README and `site/index.html` get the row `Gantt | yes | yes | yes` in the table of supported diagrams. `gantt` is removed from the sentence that lists the types that are not supported yet, in both. The Graphviz note of the page and the last sentence of the README's Graphviz section name gantt diagrams among the diagrams that do not need `dot`. The README's conformance counts follow the report.
- The playground note of the page and the README's sentence about the playground do not change: they follow the WASM module of the release on the page, and the release that ships gantt diagrams updates them.

### D11. Conformance

- Curated cases go under `testdata/conformance/cases/gantt-diagram/` with `family: gantt` and `type: gantt`, one per compared scenario, plus `diagram-source/gantt-detected`, the TXT case of `output-formats` and an embedded case (`family: uml`, `type: sequence`). The TXT and UTXT scenarios use `formats: [txt]` or `[utxt]`, the PNG scenario `[svg, png]`, the others `[svg]`. Their references come from `make reference`, and CI renders them again and checks them against the committed files.
- The embedded case is an `@startuml` case, so the golden dumps of `internal/sequencediagram`, `internal/classdiagram` and `internal/activitydiagram3` are regenerated, as for the packet change.
- Scenarios that are not compared with Java get `.go.yaml` cases: "Gantt diagram without Graphviz" (environment), "Current day" (depends on the date), "Drawing failure" (upstream's crash image cannot be a reference) and "Gantt diagram in WASM" (environment).
- The cases come before the code. Until the factory is registered, the gantt cases report "not supported yet", which the runner counts as unsupported, `TestDiagramReferences` skips the curated gantt error cases, and the embedded case fails in `make conformance`.
- The 17 Vega gantt cases are compared once the factory is registered. A case that still differs at the end is listed in `known-differences.yaml` with its reason.

### D12. Port order

The layers are ported bottom up, each with unit tests from upstream's behavior (checked with the jar's TXT output, which gives the dates of every task): `jtime`, the UBrex builders, `ngm`, the time and calendar model, the tasks and the solver, the sentences and the commands, the TXT export, then the drawing (time scales, headers, task drawings, table, main block). The TXT export comes before the drawing because it checks the scheduling against upstream without the drawing.

## Risks / Trade-offs

- **[Size]** The port is about as large as `activitydiagram3`. → The bottom-up order of D12 and the TXT cases check each layer before the drawing, and the tasks are split by layer.
- **[Floating point]** The positions come from large millisecond values times a `double`. → D6 keeps Java's order of operations. The Vega cases with long tasks (`gantt5`, 490 days) and with working hours (`hms2`, `hms3`) check it.
- **[Upstream quirks fixed by mistake]** A natural Go rewrite would fix the caches, the solver rule or the TXT end date. → D6 ports them as they are, and the scenarios "Table with working hours", "Alias and colors", "Weekly scale" and "Named days" pin them.
- **[Time zone of the current day]** Go's local zone may differ from a JVM's default zone on the same machine. → Only `today` without a date uses it, and its output depends on the day anyway.
- **[Drawing crash without a crash image]** Inputs that make upstream draw a crash image return an error. → The same holds for the other diagram types, and the HTTP server answers with its error status, as it does for a crash of any other type.
- **[UBrex builders]** The three new builders are used for the first time. → They are ported from upstream with unit tests, and every subject and complement gets parser tests built from upstream's patterns.

## Migration Plan

Nothing to migrate. Inputs that failed with "gantt diagrams are not supported yet" now render. Rollback means reverting the change.
