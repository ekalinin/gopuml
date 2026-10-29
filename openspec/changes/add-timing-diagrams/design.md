# Design

## Context

See proposal.md for the motivation. The current state that shapes the approach:

- Upstream's `timingdiagram` package has 66 files in three directories: the diagram, its factory, the players, the times and the ruler at the top, 22 commands and `TimeTickBuilder` in `command/`, and the panels of each type of player, `PlayerFrame`, `TimeArrow` and the shapes in `graphic/`. They depend on each other in a cycle: `Player.panels()` builds the panels of `graphic/`, which read `TimingRuler`, `ChangeState` and `TimeConstraint`. The `TeaVM.a()` branches only hold Java `assert` statements, which are off without `-ea`.
- `TimingDiagram` is a `TitledDiagram`, read from `@startuml`. Its factory comes after `GanttDiagramFactory` in upstream's `PSystemBuilder`. Its `getTextBlock` returns a block with a margin of 10. Its drawing calls `ruler.ensureNotEmpty()` after the dimension was computed. It has no `exportTxt`, so upstream's TXT falls back to the generic text output, as for class diagrams.
- The times are `BigDecimal`s. A `TimeTick` is compared with `compareTo`, which ignores the scale, and has no `equals`, so `getLabel` of a manual axis compares ticks by identity. The labels use `toPlainString`, `longValue` (truncation), `longValueExact`, `String.format`, a lenient `GregorianCalendar` in GMT for dates and `SimpleDateFormat(pattern, Locale.US)` in the JVM's default time zone. The analog values use `Double.parseDouble`, Java's `"" + double` and a `DecimalFormat` built by `DeduceFormat`.
- The Go side already has `uml.Base`, `uml.NewFactory`, `uml.AddCommonCommands1`, `uml.Single`, `uml.Multilines2` with `RemoveStartingQuote` and `TrimBoth`, `regex.Result.Lazy`, `abel.CommandFootboxIgnored`, `abel.Position`, `klimt.StereotypePatternOptional`, `klimt.StereotypeBuild`, `color.SimpleColor`, the `Colors` parsing, `klimt.NewOpale`, `decoration.WithLinkType` with `ApplyStyle` and `SetSpecificColor`, `klimt.ArrowConfiguration`, `klimt.FontConfigurationFromStyle` and `FontConfigurationFromSkinParam` with `FontParamTiming`, `klimt.TextBlockUtilsWithMargin`, `klimt.JavaHashMapOrder`, the shapes (`UPath`, `ULine`, `URectangle`, `UPolygon`), `jmath.Atan2`, `Sin` and `Cos`, `jfmt.DoubleToString`, `jtime.DateOf`, `DateOfEpochDay` and `EpochDay`, `uml.Throw` and `uml.ThrowNPE`, `uml.ExportTitled`, the SNames of timing diagrams (`timingDiagram`, `robust`, `concise`, `rectangle`, `clock`, `binary`, `analog`, `arrow`, `constraintArrow`, `highlight`, `note`, `timegrid`, `timeline`), the `timingDiagram` sections of the bundled skins, and `uml/chrome.go`, which maps `core.Timing` to the `timingDiagram` style.
- Missing on the Go side: `BigDecimal`, a `DecimalFormat` for doubles (`internal/sequencediagram/decimalformat.go` formats `int64` only), a full `Double.parseDouble` (`uml.JavaParseDouble` reads digits and dots only), and an exported `SimpleDateFormat`: `internal/tim/javadate.go` reproduces it with every pattern letter and its errors, and `internal/tim/javatimezone.go` the zones of the reference JVM, but both are unexported. `CommandArrow.getColorOrStylePattern` and `applyStyle` exist only as unexported `colorOrStylePattern` and `applyStyle` in `internal/sequencediagram/commands_arrow.go`, and `LINE_STYLE` as `descdiagram.LineStyle`, while `classdiagram`, `sequencediagram` and `activitydiagram3` each keep their own copy of it.
- `core.detectUnsupportedUmlType` reports archimate and timing syntax as not supported yet when no parser accepts a `@startuml` block. `TestDetectUnsupportedUmlTypes` and `TestCreateDiagramReportsUnsupportedUmlType` (`internal/core`) and `TestEmbeddedUnsupportedType` (`internal/engine`) use timing syntax as their example of a type that is not supported yet. `core.CreateDiagram` keeps the best-scored error of the factories when none accepts the block.
- The engine rejects TXT and UTXT for every type other than sequence, gantt and WBS.
- `TestScenarioCoverage` reads the delta specs of active changes, so every scenario of this change needs a case once the specs are committed. `TestDiagramReferences` of `internal/errdiagram` draws the error image of every curated case whose reference is an error. The golden dumps of `internal/sequencediagram`, `internal/classdiagram` and `internal/activitydiagram3` cover every curated `@startuml` case, so each new timing case adds to them.
- The reference image renders in the time zone of the container (UTC). The Go side renders `%date` in `time.Local` when no location is given.

## Goals / Non-Goals

**Goals:**

- Port the whole `timingdiagram` package file by file, with the SVG byte-identical to upstream's apart from the encoded source.
- Keep upstream's quirks: the scenarios of "Times", "State changes", "Messages" and "Notes" pin the ones users can see.
- Keep the results of every other diagram type unchanged. The shared code that changes is the factory list, the "not supported yet" detection, an exported form of `SimpleDateFormat` in `internal/tim`, and additions to `internal/jmath` and `internal/jfmt`.

**Non-Goals:**

- Porting the `explainArg` and `explainNow` methods of the commands and `TimeTickBuilder.describeTime`, which only upstream's `@Explain` tooling calls, the `TeaVM.a()` asserts and the logs of `Logme`.
- The Julian calendar of `GregorianCalendar` before 1582-10-15: dates are computed in the proleptic Gregorian calendar.
- A time zone option for `use date format`: the default zone is used, as for `%date`.

## Decisions

### D1. One package

`internal/timing` holds the whole package, with upstream's names and a `Provenance` line for each Java file it ports, grouped by layer: `diagram.go` (`TimingDiagram`, `Clocks`), `factory.go`, `timetick.go` (`TimeTick`, `TimingFormat`, `DeduceFormat`, `TimeAxisStategy`), `ruler.go` (`TimingRuler`), `player.go` (`Player`, `AbstractStatePlayer`, `PlayerRobust`, `PlayerConcise`, `PlayerRectangle`), `player_clock.go`, `player_binary.go`, `player_analog.go` (with `TimeSeries`), `changestate.go` (`ChangeState`, `TickInPlayer`, `TimingStyle`), `timemessage.go`, `timeconstraint.go`, `highlight.go`, `note.go` (`TimingNote`), `commands_player.go`, `commands_state.go`, `commands_time.go` (with `TimeTickBuilder`), `commands_annotation.go` (messages, constraints, notes, highlights), `panels.go` (`Panels`, `PanelsNoLeft`), `panels_state.go` (`PanelsState`, `PanelsConcise`, `PanelsRectangle` and their shapes), `panels_robust.go`, `panels_binary.go`, `panels_clock.go`, `panels_analog.go`, `playerframe.go`, `timearrow.go` (`TimeArrow`, `IntricatedPoint`) and `export.go`.

Alternative: subpackages `command` and `graphic`, as upstream. The players build their panels and the panels read the model, so Go would need interfaces to break the cycle.

### D2. BigDecimal in internal/jmath

`jmath.BigDecimal` is an unscaled `big.Int` and a scale, with the parts of `java.math.BigDecimal` that timing diagrams use: the constructors from a string (sign, digits and an optional fraction, with the JDK's `NumberFormatException` messages, such as `Character array contains more than one decimal point.`), from an `int64` and from a `float64` (exact, as `new BigDecimal(double)`), `ZERO`, `valueOf(long)`, `add` (the larger scale), `multiply` (the sum of the scales), `compareTo` (which ignores the scale), `signum`, `doubleValue` (correctly rounded), `longValue` (truncated to 64 bits), `longValueExact` (with the JDK's `ArithmeticException`) and `toPlainString`. It panics with `*color.RuntimeError`, as `internal/jtime` does. Its file has `Provenance: original`.

Alternatives: `big.Float` and `big.Rat` lose the scale, which `toPlainString` prints (`1.50`) and the first tick of a time keeps. A new package for one type was rejected because `jmath` is already the Java-compatible math library.

### D3. DecimalFormat and parseDouble in internal/jfmt

- `jfmt.DecimalFormat` formats a `float64` like `java.text.DecimalFormat` with the patterns that `DeduceFormat.getDecimalFormat` builds: an optional `+` prefix, optional grouping by 3, the decimal and grouping separators, the minimum and maximum fraction digits, and `HALF_EVEN` rounding, with `-0` for a negative value that rounds to zero. `internal/sequencediagram/decimalformat.go` does not change.
- `jfmt.ParseDouble` is `Double.parseDouble`: it trims the characters up to the space, accepts Java's grammar (a sign, `NaN`, `Infinity`, decimal and hexadecimal floats with an exponent, and the `f`, `F`, `d` and `D` suffixes) and returns an error for the rest, such as Go's `inf` or `1_0`. `PlayerAnalog` reads a value that fails as 0, as upstream does. `uml.JavaParseDouble` does not change.
- The package comment says that `jfmt` also parses numbers like Java.

### D4. SimpleDateFormat from internal/tim

`internal/tim` gets an exported `DateFormat`: `NewDateFormat(pattern)` compiles a pattern with the existing `newJavaDateFormat` and returns its error, and `Format(t, loc)` formats with the existing code in the zone of `zoneOfLocation(loc)`. `CommandUseDateFormat` fails with `Bad date format` when `NewDateFormat` fails, and `TimingFormat` formats in `time.Local`, the Go form of the JVM's default zone, as `%date` does without a location. The unexported names of `tim` do not change.

Alternative: a copy of `javaDateFormat` in `internal/timing`, which would duplicate 300 lines and the zone table.

### D5. Times and the ruler

- `TimeTick` holds a `jmath.BigDecimal` and a `TimingFormat`. The ruler keeps its times in a sorted slice with `compareTo` and keeps the first tick of a time, as upstream's `TreeSet`. A tick has identity: the manual axis labels a tick with a name only when the named tick is the same `*TimeTick`, and reads the names in `klimt.JavaHashMapOrder`, as upstream's `HashMap`.
- `TimeTickBuilder` keeps upstream's order of the alternatives and its rules: a relative time is relative to the current time, or absolute when there is none, `:code+delta` is a `DECIMAL` tick, `clk*N` ignores the offset of the clock, an hour is `3600h+60m+s` without range checks, and a date is the epoch second of `y/m/d` at 00:00 GMT, normalized leniently (a month above 12 and a day above the end of the month roll over), from `jtime.DateOf` of the first day of the month plus the days.
- `TimingFormat` keeps upstream's formats: `DECIMAL` with `toPlainString` and `longValue`, `HOUR` with Java's `int` arithmetic, `DATE` as `MM/dd` from `jtime.DateOfEpochDay`, and a `SimpleDateFormat` (D4). The ruler uses the format of the last tick that is not `DECIMAL`. `longValueExact` panics with upstream's `ArithmeticException` on a fractional time.
- `TimingRuler` keeps upstream's arithmetic: the tick unit from the highest common factor of the truncated times (`-1` without one), the limit of 4000 pixels with `jmath.Round`, the forced unit and interval of `scale ... as ... pixels`, `ensureNotEmpty` run by the drawing after the dimension was computed, and the `IllegalStateException` of an axis without a time at or above zero, thrown with `uml.Throw`. The positions keep Java's order of operations on `float64`.

### D6. Model

- `TimingDiagram` embeds `uml.Base` with the type `core.Timing`, as `WBSDiagram` does, so the annotations, skins, styles and pragmas come from the shared code. `Description` returns `(Timing Diagram)`.
- The players are kept in declaration order with a map from the code; a second declaration of a code replaces the player at its first place, as upstream's `LinkedHashMap`. The clocks are kept by code. `lastPlayer` is set by the robust, concise and rectangle declarations and by `@<player>` only, and `compactByDefault` is read when a player is created.
- `AbstractStatePlayer` keeps its changes sorted by time with the first change of a time kept, `PlayerBinary` and `PlayerAnalog` keep the last value of a time, as upstream's `TreeSet` and `TreeMap`. The `has` labels apply when a state is set. The methods that upstream does not implement for a type of player (`defineState`, `setState` and `createConstraint` of clocks, `defineState` of binary and analog players) panic with `uml.Throw(uml.UnsupportedOperationException, "")`, which the command factory turns into upstream's crash message.
- `TimeMessage` embeds `decoration.WithLinkType`, and its stroke is always the stroke of the `arrow` style, as upstream's override of `getUStroke`. `TimeArrow` takes the `*TimeMessage` for the stroke.
- The style queries are built inline with `klimt.StyleSignatureBasicOf(SNameRoot, SNameElement, SNameTimingDiagram, ...)`, the Go form of upstream's `StyleQueries.TIMINGDIAG_*`, with `AddSName` and `AddStereotypeOf`.
- The panels of a player are built on their first use and kept, as upstream's lazy `panels()`.

### D7. Parsing

- The factory adds the commands in upstream's order: `uml.AddCommonCommands1`, `abel.CommandFootboxIgnored`, then the 22 commands of `TimingDiagramFactory`. The commands are `uml.Single` with upstream's patterns and first tokens, case-insensitive, and the long note is `uml.Multilines2` with `RemoveStartingQuote`, `TrimBoth` and the end pattern `^end[%s]?note$`.
- `LINE_STYLE`, `getColorOrStylePattern` and `applyStyle` of `CommandArrow` are copied into `commands_annotation.go` with the `Provenance` lines of `CommandArrow.java` and `CommandLinkElement.java`, as the other diagram packages keep their own copies. The messages use `WithLinkType.ApplyStyle`, and the constraints the copied `applyStyle`.
- The groups are read with `Result.Lazy` where upstream calls `getLazzy`. The stereotype before `as` wins over the one after the code, as upstream does.
- Upstream's errors are returned with their messages (`Unknown "<code>"`, `Unknown <code>`, `Missing @ line before this`, `No such participant <code>`, `What time?`, `Unknown time label`, `You have to provide a participant`, `No such element: <code>`, `Bad value`, `Bad date format`) and an unknown color with `No such color`. The `NullPointerException`s that upstream throws while parsing are thrown with `uml.ThrowNPE` and the JDK's helpful messages, read from the reference outputs: `Cannot invoke "net.sourceforge.plantuml.timingdiagram.TimeTick.getTime()" because "result" is null` for an unknown named time with an offset, and `Cannot invoke "java.lang.Comparable.compareTo(Object)" because "k1" is null` for an unknown named time in a player block.

### D8. Drawing as upstream

- The text block of the diagram computes its dimension from the players and the ruler, and its drawing follows upstream's `drawInternal`: the background of the players, the borders unless `mode compact`, the back of the highlights, the vertical lines of the ticks, then for each player its separator, its title and its panels, then the time axis, the messages and the lines and captions of the highlights. The SVG elements come in that order.
- The panels keep upstream's constants and quirks: the `int` division of the step of a robust player, the levels of the declared states in reverse order, the notes that only concise, rectangle and binary players draw, the bottom notes of state players placed without the height of the top comment, the clock waveform from time 0, capped at 1000 periods, the hatching of two states every 5 pixels, and the projection of a message on a binary player at its high level.
- `TimeArrow` keeps upstream's choice of the shortest of the four segments (a tie goes to the later one) and draws the head with `jmath.Atan2`, `Sin` and `Cos`. `TimeConstraint` keeps its two forms (heads outward when the distance is above 20, inward otherwise). The notes use `klimt.NewOpale`. The caption of a highlight uses `FontParamTiming`, as upstream does.
- A panic while drawing, such as the `IllegalStateException` of the ruler, the `IndexOutOfBoundsException` of a constraint on a robust player without changes or the `ArithmeticException` of `ticks num on multiple 0`, is returned as an error without output by `uml.ExportTitled`.

### D9. Registration and detection

- `timing.Factory` (`uml.NewFactory(core.Timing, ...)`) goes into `engine.Factories` after `gantt.Factory`, as `TimingDiagramFactory` comes after `GanttDiagramFactory` in upstream's `PSystemBuilder`. Its errors join the best-scored error of `core.CreateDiagram`, as upstream's do.
- The timing entry and the `Provenance` lines of the five timing commands leave `internal/core/detect.go`, which keeps archimate. `TestDetectUnsupportedUmlTypes` loses its timing rows, and `TestCreateDiagramReportsUnsupportedUmlType` and `TestEmbeddedUnsupportedType` use `archimate #Business "Customer" as C`, with the message `archimate diagrams are not supported yet`. The `why` of `styling/embedded-unsupported-type.go.yaml` names the embedded archimate diagram.
- The engine's check of the text formats does not change, so TXT and UTXT of a timing diagram return `UnsupportedFormatError`.
- The WASM build needs no special case, because timing diagrams do not use Graphviz.

### D10. Documentation

- The README and `site/index.html` get the row `Timing | yes | yes | no` in the table of supported diagrams. The Graphviz note of the page and the last sentence of the README's Graphviz section name timing diagrams among the diagrams that do not need `dot`. The sentence on the types that are not supported yet names `archimate, salt` instead of `archimate, timing`. The README's conformance counts follow the report.
- The conformance tally and the playground note of the page and the README's sentence about the playground follow the release on the page: they do not change.

### D11. Conformance

- Curated cases go under `testdata/conformance/cases/timing-diagram/` with `family: uml` and `type: timing`, one per compared scenario, plus `diagram-source/timing-detected`, `diagram-source/timing-syntax-error` and the embedded case (`type: sequence`). The PNG scenario uses `formats: [svg, png]`, the others `[svg]`. Their references come from `make reference`, and CI renders them again and checks them against the committed files. The golden dumps of `internal/sequencediagram`, `internal/classdiagram` and `internal/activitydiagram3` are regenerated for the new `@startuml` cases.
- The date scenarios use dates on the 15th of a month, at 00:00 GMT, and the pattern `MMM yyyy`, so their labels are the same in every time zone of the host.
- Scenarios that are not compared with Java get `.go.yaml` cases: "Timing diagram without Graphviz" (environment), "Drawing failure" (upstream draws a crash report with a random quote and the Java stack, this change returns an error), "TXT timing diagram" (upstream draws its generic text output, this change reports an unsupported format) and "Timing diagram in WASM" (environment).
- The cases come before the code. Until the factory is registered, the timing cases report "not supported yet", which the runner counts as unsupported, `TestDiagramReferences` skips the curated timing error cases, and the embedded case fails in `make conformance`.
- The 3 Vega timing cases are compared once the factory is registered. A case that still differs at the end is listed in `known-differences.yaml` with its reason.

### D12. Port order

The layers are ported bottom up, each with unit tests: the JDK parts (`BigDecimal`, `DecimalFormat`, `ParseDouble`, `DateFormat`), the times and the ruler, the model and the players, the commands and the factory, the panels of the state players, the other panels, then the messages, constraints, highlights, notes and the export, and the registration.

## Risks / Trade-offs

- **[Time zone of `use date format`]** Upstream formats in the JVM's default zone, which is UTC in the reference image, and Go in `time.Local`, which is the zone of the host. → The curated cases use a pattern and dates whose labels do not depend on the zone (D11). A diagram rendered on a host outside UTC with a pattern that shows hours can differ from the reference image, as upstream run on that host would.
- **[DecimalFormat rounding]** `DecimalFormat` rounds `HALF_EVEN` from the binary value of the double, which differs from rounding its shortest decimal form for values like 2.675. → `jfmt.DecimalFormat` rounds the exact value, and its unit tests check ties and near-ties against the jar.
- **[BigDecimal scale]** The first tick of a time keeps its scale and its format, and the labels of a manual axis print it. → `TimeTick` keeps the `BigDecimal` it was parsed from, and the scenarios "Decimal times", "Manual time axis" and "Decimal clock period" check the labels against upstream.
- **[New errors for broken `@startuml` blocks]** The timing parser now takes part in the best-scored error of a `@startuml` block that no parser accepts, as upstream's does, where Go reported timing syntax as not supported yet. → `make conformance` checks every error case of the other types, and the scenario "Timing syntax error" pins the timing one.
- **[Size of a diagram without times]** Upstream computes the size of the diagram before `ensureNotEmpty` changes the ruler, so the image is wider than its drawing. → The Go drawing calls `ensureNotEmpty` at the same point, and the scenario "Timing diagram without times" checks the size against upstream.
- **[Messages of the JDK]** The crash messages of the commands (D7) come from the JVM of the reference image. → They are read from the committed references, and the scenarios "Unknown named time with an offset" and "State of a clock" check them.
- **[Floating point]** The layout adds and compares `double` values. → The drawing keeps Java's order of operations on `float64`, and the Vega case with a decimal period and offset checks it.
- **[Order of the archives]** The delta spec of `wasm-build` contains the requirement of `add-wbs-diagrams` with timing added. Archiving this change before that one would let the WBS archive replace it without timing. → `add-mindmap-diagrams` and `add-wbs-diagrams` are archived first.

## Migration Plan

Nothing to migrate. Inputs that failed with "timing diagrams are not supported yet" now render, or fail with upstream's error. Rollback means reverting the change.
