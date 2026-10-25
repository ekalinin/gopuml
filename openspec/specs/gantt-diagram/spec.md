# gantt-diagram Specification

## Purpose
Supports PlantUML v1.2026.9 gantt diagrams under `@startgantt` and `@startproject`, with the same syntax, scheduling, drawing, TXT table and errors as upstream.

## Requirements

### Requirement: Gantt diagram blocks

The system SHALL render `@startgantt` ... `@endgantt` and `@startproject` ... `@endproject` blocks as upstream's gantt diagrams. The diagram type SHALL be gantt and the description `(Gantt)`. A gantt diagram SHALL be drawn as upstream draws it: the table of tasks on the left, the calendar above the tasks and repeated below them, the tasks as bars or diamonds with their names, the arrows between tasks, and the resources below the tasks. The common commands of upstream's gantt diagrams SHALL be supported: title, header, footer, caption, legend, mainframe, `skin`, `skinparam`, `<style>`, sprites, `scale`, pragmas, and the other commands that upstream adds to gantt diagrams. Gantt diagrams SHALL NOT need Graphviz.

#### Scenario: Project plan

- **WHEN** a `@startgantt` block contains `Project starts 2020-01-06`, `[Design] lasts 5 days`, `[Build] starts at [Design]'s end` and `[Build] lasts 3 days`
- **THEN** a daily calendar from January 6, 2020 is drawn with a row of months, a row of days of the week and a row of day numbers, `Design` and `Build` are drawn as bars with their names, an arrow goes from the end of `Design` to the start of `Build`, the table on the left gives the start, end and duration of each task, the calendar is repeated below the tasks, the diagram type is gantt and the description is `(Gantt)`

#### Scenario: Project start tag

- **WHEN** a `@startproject` ... `@endproject` block contains `Project starts 2020-01-06` and `[A] lasts 2 days`
- **THEN** the diagram is drawn as the same lines in a `@startgantt` block

#### Scenario: Title and style

- **WHEN** a gantt diagram contains `title Release plan` and a `<style>` block with `ganttDiagram { task { BackGroundColor #33b2e2 } arrow { LineColor red } }`, and two tasks linked by `starts at`
- **THEN** the title `Release plan` is drawn above the diagram, the bars are filled with `#33b2e2` and the arrow is red

#### Scenario: Scaled gantt diagram

- **WHEN** a gantt diagram contains `scale 2`
- **THEN** the drawing is twice as large as without `scale 2`, as upstream draws it

#### Scenario: Gantt diagram without Graphviz

- **WHEN** a gantt diagram is rendered and `dot` cannot be found
- **THEN** the diagram renders without an error

### Requirement: Sentences

The system SHALL parse upstream's sentences: a subject, then one or more verb phrases joined by ` and `. The words SHALL match without regard to the case of ASCII letters. Words SHALL be separated by spaces, and a tab SHALL NOT count as a space. The words that upstream accepts between a verb and its complement (`on`, `for`, `the`, `at`, and `from` after some verbs) MAY be written there. Text after the last verb phrase SHALL be ignored, as upstream does, unless it starts with ` and `. A line that no command accepts, for example because its subject or its first verb phrase is not recognized, SHALL be a syntax error with the message `Syntax Error? (Assumed diagram type: gantt)`. A verb phrase after ` and ` that is not recognized SHALL fail the line with the message `No matching verb phrase (Assumed diagram type: gantt)`, after the verb phrases before it have been applied. `then` before a new task SHALL start it at the end of the last task created, with an arrow. `it` SHALL be the last task used as a subject, and `she`, `he` and `they` the last resource used as a subject.

#### Scenario: Chained verb phrases

- **WHEN** a gantt diagram contains `[Design] lasts 3 days` and `[Build] starts at [Design]'s end and lasts 2 days`
- **THEN** `Build` starts at the end of `Design` and lasts 2 days

#### Scenario: Case of the words

- **WHEN** a gantt diagram contains `PROJECT STARTS the 6th of january 2020` and `[A] Lasts 2 Days`
- **THEN** the diagram is drawn as with `Project starts 2020-01-06` and `[A] lasts 2 days`

#### Scenario: Ignored trailing text

- **WHEN** a gantt diagram contains `[A] lasts 5 days foo bar`
- **THEN** `A` lasts 5 days and no error is reported

#### Scenario: No matching verb phrase

- **WHEN** a gantt diagram contains `[A] lasts 3 days and foo`
- **THEN** the system reports the error `No matching verb phrase (Assumed diagram type: gantt)` on that line

#### Scenario: Gantt syntax error

- **WHEN** a gantt diagram contains `[A] foo`
- **THEN** the system reports a syntax error on that line with the message `Syntax Error? (Assumed diagram type: gantt)`

#### Scenario: Then and it

- **WHEN** a gantt diagram contains `[A] lasts 3 days`, `then [B] lasts 2 days` and `it is colored in red`
- **THEN** `B` starts at the end of `A`, an arrow goes from `A` to `B`, and `B` is red

### Requirement: Tasks

The system SHALL support upstream's tasks. A task SHALL be created the first time `[name]` is used, and `[name] as [id]` SHALL create a task drawn as `name` that later lines name `[id]`. Task ids SHALL be case-sensitive. `lasts` and `requires` SHALL give the work of a task in days, weeks, months, hours, minutes or seconds, several of them joined by `and` or commas, where a week is as many days as are open in a week when the line is read and a month is 30 days. The work of a task SHALL be done on the open days of its calendar only, so that a task spans the closed days it crosses. A task SHALL start on the project start unless a start is given. `starts` and `ends` SHALL take a date, a date with a time of day, `D+N` or `N days after start`, and the end date SHALL be the last day of the task. `happens` SHALL make a milestone, drawn as a diamond, on a date or at the start or end of another task. `occurs from [A] to [B]` SHALL make a task from the start of `A` to the end of `B`. A task SHALL also support a stereotype after its name, `is colored in` (fill and border color), `is colored for completion from` ... `to` ..., `is N% completed` (the part not done drawn in the undone color), `is deleted`, `pauses on` (dates, ranges of dates and days of the week), `links to [[url]]`, `is displayed as` and `displays on same row as`.

#### Scenario: Weeks of open days

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `saturday are closed`, `sunday are closed` and `[A] lasts 2 weeks`
- **THEN** `A` is drawn from January 6 to January 17, 2020, across the closed weekend, and the closed days are drawn in the closed color

#### Scenario: Start and end dates

- **WHEN** a gantt diagram contains `Project starts 2020-01-06` and `[A] as [T1] starts on 2020-01-08 and ends on 2020-01-21`
- **THEN** a task named `A` is drawn from January 8 to January 21, 2020, both included

#### Scenario: Times of day

- **WHEN** a gantt diagram contains `Project starts 2024-09-22`, `[Task 1] starts 2024-09-22 10:30:46` and `[Task 1] ends 2024-09-24T16:30:47`
- **THEN** `Task 1` is drawn from 10:30:46 on September 22 to 16:30:47 on September 24, as upstream draws it

#### Scenario: Relative dates

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `[A] starts D+3 and lasts 2 days` and `[B] starts 5 days after start and lasts 1 day`
- **THEN** `A` is drawn on January 9 and 10, 2020 and `B` on January 11

#### Scenario: Milestones

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `[A] lasts 4 days`, `[M1] happens at [A]'s end` and `[M2] happens 2020-01-15`
- **THEN** `M1` is drawn as a diamond on January 9, 2020, the last day of `A`, and `M2` as a diamond on January 15

#### Scenario: Task between milestones

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `[P_start] happens 2020-01-08`, `[P_end] happens 2020-01-13` and `[Design] occurs from [P_start] to [P_end]`
- **THEN** `Design` is drawn from January 8 to January 13, 2020

#### Scenario: Alias and colors

- **WHEN** a gantt diagram contains `[Prototype design] as [P] lasts 5 days`, `[P] is colored in Lavender/LightBlue` and `[Prototype design] lasts 2 days`
- **THEN** a task drawn as `Prototype design` lasts 5 days with a lavender fill and a light blue border, and the last line creates a second task `Prototype design` that lasts 2 days, because its id is not `P`

#### Scenario: Same row

- **WHEN** a gantt diagram contains `[T1] lasts 10 days`, `[M1] happens at [T1]'s end` and `[M1] displays on same row as [T1]`
- **THEN** `M1` is drawn on the row of `T1`

#### Scenario: Completion

- **WHEN** a gantt diagram contains `[A] lasts 4 days and is 25% completed`
- **THEN** the first quarter of the bar of `A` is drawn in the task color and the rest in the undone color

#### Scenario: Completion colors

- **WHEN** a gantt diagram contains `[A] lasts 4 days and is 50% completed` and `[A] is colored for completion from red to green`
- **THEN** `A` is drawn in the color halfway between red and green, as upstream computes it

#### Scenario: Deleted task

- **WHEN** a gantt diagram contains `[A] lasts 2 days` and `[A] is deleted`
- **THEN** `A` is drawn white with a black border

#### Scenario: Pauses

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `[A] lasts 3 days` and `[A] pauses on 2020-01-07`
- **THEN** `A` is drawn from January 6 to January 9, 2020, interrupted on January 7

#### Scenario: Display name and link

- **WHEN** a gantt diagram contains `[C] lasts 2 days and links to [[https://example.com]]` and `[D] lasts 2 days and is displayed as Build step`
- **THEN** the bar of `C` is a link to `https://example.com` and the bar of `D` is labeled `Build step`

### Requirement: Constraints

The system SHALL support upstream's constraints between tasks. `starts` and `ends` SHALL take the start or end of another task, written `[X]'s start` or `[X]'s end`, optionally after `at` or `with`, or after a delay of days or weeks, optionally working days, and `before` or `after`. A delay in working days SHALL skip the closed days. A start or end given by another task SHALL draw an arrow from that task, and `with` a color, and optionally `dotted`, `bold` or `dashed`, then `link`, SHALL give the arrow its color and style. `[A] -> [B]` SHALL start `B` at the end of `A` and draw an arrow, creating the tasks that do not exist. `A -> B` and `A -[style]-> B` between task ids SHALL do the same for existing tasks, with the style of the arrow.

#### Scenario: Delays

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `saturday are closed`, `sunday are closed`, `[A] lasts 5 days`, `[B] starts 2 days after [A]'s end and lasts 1 day` and `[C] starts 2 working days after [A]'s end and lasts 1 day`
- **THEN** `B` is drawn on January 13, 2020, `C` on January 14, and arrows go from `A` to `B` and from `A` to `C`

#### Scenario: Arrow between tasks

- **WHEN** a gantt diagram contains `[A] lasts 3 days`, `[B] lasts 2 days` and `[A] -> [B]`
- **THEN** `B` starts at the end of `A` and an arrow goes from `A` to `B`

#### Scenario: Arrow styles

- **WHEN** a gantt diagram contains `[A] as [A1] lasts 3 days`, `[B] as [B1] starts at [A1]'s end with red dotted link`, `[B1] lasts 2 days`, `[C] as [C1] lasts 2 days` and `B1 -[#blue,dashed]-> C1`
- **THEN** a red dotted arrow goes from `A` to `B`, a blue dashed arrow goes from `B` to `C`, and `C` starts at the end of `B`

#### Scenario: End at another task's end

- **WHEN** a gantt diagram contains `[A] lasts 3 days`, `[B] starts at [A]'s end and lasts 4 days` and `[C] lasts 3 days and ends at [B]'s end`
- **THEN** `C` ends on the last day of `B` and an arrow goes from `B` to `C`

### Requirement: Calendar

The system SHALL support upstream's calendar. The project start SHALL be set with `Project starts` and a date before the first task. Days of the week SHALL be closed and opened with `are closed` and `are opened`, and dates and ranges of dates (`A to B`, `A and N days`, `then N days`, `D+N`) with `is` or `are` `closed` and `opened`, globally or `for [task]`, except that `are closed for [task]` after a day of the week closes it for every task, as upstream does. Days of the week, dates and ranges of dates SHALL be colored with `colored in`, and ranges of dates SHALL be named with `are named [name]`, which draws the name in the calendar and makes `[name]'s start` and `[name]'s end` available to tasks. `today is` a date SHALL set the current day and `today is colored in` SHALL color it. Without a date, the current day SHALL be the date on which the diagram is rendered, so that the drawing depends on that date, as upstream's does. `separator just before`, `separator just after` and `separator just at` SHALL draw a vertical separator. `from H:M to H:M are working hours` SHALL limit the work of a day to the working hours declared so far and draw the calendar with the hours of each day. `hide closed` SHALL leave the closed days out of a daily calendar. Without `Project starts`, the calendar SHALL be numbered by day from 1, as upstream's.

#### Scenario: Closed and opened days

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `saturday are closed`, `sunday are closed`, `2020-01-11 is opened` and `[A] lasts 8 days`
- **THEN** `A` is drawn from January 6 to January 14, 2020, working on Saturday, January 11, and the closed days are drawn in the closed color

#### Scenario: Closed range and colored days

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `2020-01-08 to 2020-01-09 are closed`, `2020-01-13 to 2020-01-14 are colored in salmon`, `friday are colored in lightblue` and `[A] lasts 10 days`
- **THEN** January 8 and 9, 2020 are closed, January 13 and 14 are salmon, the Fridays are light blue, and `A` does not work on the closed days

#### Scenario: Named days

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `2020-01-06 and 4 days are colored in lightblue and are named [Sprint 1]`, `then 5 days are colored in lightgreen and are named [Sprint 2]` and `[Release] lasts 2 days and starts at [Sprint 2]'s start`
- **THEN** the days of each sprint are colored and named in the calendar, as upstream draws them, and `Release` starts on January 10, 2020, the first day of `Sprint 2` as upstream counts it

#### Scenario: Day opened for one task

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `saturday are closed`, `sunday are closed`, `saturday are open for [Foo]`, `[Bar] lasts 6 days` and `[Foo] lasts 6 days`
- **THEN** `Foo` works on Saturday, January 11, 2020 and ends on that day, and `Bar` ends on Monday, January 13

#### Scenario: Today

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `today is 2020-01-08 and is colored in #AAF` and `[A] lasts 5 days`
- **THEN** the column of January 8, 2020 is drawn in `#AAF`

#### Scenario: Current day

- **WHEN** a gantt diagram contains `today is colored in #AAF` without a date, and is rendered on a known date
- **THEN** the column of that date is drawn in `#AAF`

#### Scenario: Vertical separators

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `[A] lasts 3 days`, `separator just before 2020-01-08` and `separator just at [A]'s end`
- **THEN** vertical separators are drawn before January 8 and before January 9, 2020, as upstream draws them

#### Scenario: Working hours

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `from 9:00 to 17:00 are working hours` and `[A] lasts 2 days`
- **THEN** a day of work is 8 hours, `A` is drawn from January 6 to 17:00 on January 7, 2020, and the calendar is drawn with the working hours of each day, as upstream draws it

#### Scenario: Hide closed days

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `saturday are closed`, `sunday are closed`, `hide closed` and `[A] lasts 8 days`
- **THEN** the calendar has no column for the weekend, and `A` is drawn across the open days only

#### Scenario: Calendar without a project start

- **WHEN** a gantt diagram contains `[A] lasts 5 days` and `[B] lasts 15 days` and no `Project starts`
- **THEN** the calendar is numbered by day from 1, as upstream draws it, and the table gives the start and end of each task as `Day N`

### Requirement: Resources

The system SHALL support upstream's resources. `on {name}` or `on {name:P%}` after a task SHALL assign the resource to the task at `P` percent (100 by default), which makes the work take `100 / P` times longer, and the name and the percentage SHALL be drawn after the name of the task. `{name} is off` and `{name} is on` with dates, ranges of dates or days of the week SHALL close and open days for the tasks of that resource. Below the tasks, each resource SHALL be drawn with its load per day, as upstream draws it. `hide resources names` and `hide resources footbox` SHALL hide the names of the resources after the tasks and the resources below the tasks.

#### Scenario: Resource load

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `[A] on {Alice} lasts 2 days` and `[B] on {Bob:50%} lasts 2 days`
- **THEN** `B` is drawn from January 6 to January 9, 2020, the labels are `A {Alice}` and `B {Bob:50%}`, and below the tasks `Alice` has the load `100` on two days and `Bob` the load `50` on four days

#### Scenario: Resource days off

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `{Alice} is off on monday` and `[A] on {Alice} lasts 3 days`
- **THEN** `A` is drawn from January 7 to January 9, 2020

#### Scenario: Hidden resources

- **WHEN** a gantt diagram contains `[A] on {Alice} lasts 2 days`, `hide resources names` and `hide resources footbox`
- **THEN** `A` is labeled `A` and no resource is drawn below the tasks

### Requirement: Groups, separators and notes

The system SHALL support upstream's groups, separators and notes. `group [name]` ... `end group` SHALL draw a group bracket from the start of its first task to the end of its last task, with its name. `-- text --` and `--` SHALL draw a horizontal separator across the diagram, with the text when there is one. `note bottom` ... `end note` SHALL attach a note to the last task and draw it below the task, moving the following rows down when they would overlap it, as upstream does.

#### Scenario: Group

- **WHEN** a gantt diagram contains `group [Phase 1]`, `[A] lasts 2 days`, `then [B] lasts 3 days` and `end group`
- **THEN** a bracket named `Phase 1` is drawn above `A` and `B`, from the start of `A` to the end of `B`

#### Scenario: Horizontal separator

- **WHEN** a gantt diagram contains `[A] lasts 2 days`, `-- Phase 2 --` and `[B] lasts 3 days`
- **THEN** a horizontal line with the text `Phase 2` is drawn between `A` and `B`

#### Scenario: Note

- **WHEN** a gantt diagram contains `[A] lasts 5 days`, `note bottom`, `Long very very very Long Long note`, `end note` and `[B] lasts 3 days`
- **THEN** the note is drawn below `A`, and `B` is drawn below the note

### Requirement: Time scale

The system SHALL support upstream's time scales: `printscale`, `projectscale` or `ganttscale` followed by `daily` (the default), `weekly`, `monthly`, `quarterly` or `yearly`, with the options `with calendar date` and `with week numbering from N` for weeks and `zoom Z`. The weekly scale SHALL number the weeks as Java's week of year does, with the first day of the week and the minimal number of days of the first week given by `week starts on DAY and must have at least N days` (Monday and 4 by default). `print between DATE and DATE` SHALL draw only the days between the two dates, without the tasks outside them. `hide footbox` SHALL not repeat the calendar below the tasks.

#### Scenario: Weekly scale

- **WHEN** a gantt diagram contains `Project starts 2024-12-23`, `printscale weekly` and `[A] lasts 20 days`
- **THEN** the weeks are numbered `52`, `53` and `2`, as Java's week of year numbers them, below the months `Dec` and `Jan`

#### Scenario: Weekly scale with dates

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `week starts on sunday and must have at least 1 day`, `printscale weekly with calendar date` and `[A] lasts 20 days`
- **THEN** the weeks start on Sunday and are labeled with the day of the month of their first day, `12` and `19`

#### Scenario: Week numbering from a number

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `printscale weekly with week numbering from 1` and `[A] lasts 20 days`
- **THEN** the weeks are numbered `1`, `2` and `3`

#### Scenario: Monthly scale

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `printscale monthly` and `[A] lasts 90 days`
- **THEN** the calendar has a row with the year `2020` and a row with the months, as upstream draws them

#### Scenario: Quarterly scale

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `printscale quarterly` and `[A] lasts 200 days`
- **THEN** the calendar has a row with the year `2020` and a row with the quarters `Q1`, `Q2` and `Q3`

#### Scenario: Yearly scale

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `printscale yearly` and `[A] lasts 490 days`
- **THEN** the calendar has a row with the year `2020`, and the label `2021` is left out because the part of 2021 that the task covers is too narrow for it, as upstream draws it

#### Scenario: Zoom

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `printscale daily zoom 2` and `[A] lasts 5 days`
- **THEN** the days are twice as wide as without `zoom 2`

#### Scenario: Print between

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `print between 2020-01-10 and 2020-01-20`, `[A] lasts 3 days` and `[B] starts 2020-01-12 and lasts 3 days`
- **THEN** the calendar goes from January 10 to January 20, 2020, and only `B` is drawn

#### Scenario: Hide footbox

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `hide footbox` and `[A] lasts 3 days`
- **THEN** the calendar is drawn above the tasks only

### Requirement: Task table

The system SHALL draw upstream's table of tasks on the left of the calendar, with the columns `Start`, `End` and `Duration` by default, in upstream's language. `show column` and `hide column` followed by `task`, `start`, `end` or `duration` SHALL show and hide a column. The columns SHALL keep upstream's order (task, start, end, duration) whatever the order of the lines.

#### Scenario: Columns

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `[A] lasts 3 days`, `hide column duration` and `show column task`
- **THEN** the table has the columns `Task`, `Start` and `End`, in that order

### Requirement: Language

`language` followed by a language code SHALL set the language of the calendar and the table: the names of the months and the days of the week, the dates of the table and the durations. The languages SHALL be those of upstream (`de`, `es`, `fr`, `ja`, `ko`, `ru`, `zh`), and any other code SHALL give English, as upstream does.

#### Scenario: German

- **WHEN** a gantt diagram contains `language de`, `Project starts 2024-09-22`, `[Task 1] starts 2024-09-22T10:30:46` and `[Task 1] ends 2024-09-24 16:30:47`
- **THEN** the months and days of the week of the calendar are in German, and the table has the columns `Start`, `Ende` and `Dauer`, as upstream draws them

### Requirement: Gantt errors

The system SHALL report upstream's errors for gantt diagrams, with upstream's messages and lines, each message followed by ` (Assumed diagram type: gantt)`. The errors SHALL include the syntax errors and the unrecognized verb phrases of the sentences, the errors of the commands (for example `No such task Nope`, `No group to be closed`, `No starting date for the project` and `Starting point must be set before task definition`), the check that a diagram does not last more than 20 years, and the Java exceptions that upstream reports with `You should send a mail to plantuml@gmail.com with this log (V1.2026.9)` and the exception, such as an invalid date. A command that upstream ignores with a warning SHALL be ignored with the same warning. When drawing a diagram fails, where upstream draws a crash report inside the diagram, the rendering SHALL fail with an error instead of a drawing.

#### Scenario: No starting date

- **WHEN** a gantt diagram contains only `[A] starts 2020-01-10`
- **THEN** the system reports the error `No starting date for the project (Assumed diagram type: gantt)` on that line

#### Scenario: Late project start

- **WHEN** a gantt diagram contains `[A] lasts 3 days` and then `Project starts 2020-01-06`
- **THEN** the system reports the error `Starting point must be set before task definition (Assumed diagram type: gantt)` on the line `Project starts 2020-01-06`

#### Scenario: Unknown task

- **WHEN** a gantt diagram contains only `[A] starts at [Nope]'s end`
- **THEN** the system reports the error `No such task Nope (Assumed diagram type: gantt)` on that line

#### Scenario: No group to close

- **WHEN** a gantt diagram contains `Project starts 2020-01-06` and `end group`
- **THEN** the system reports the error `No group to be closed (Assumed diagram type: gantt)` on the line `end group`

#### Scenario: Twenty years

- **WHEN** a gantt diagram contains `Project starts 2000-01-01` and `[A] lasts 7302 days`
- **THEN** the system reports the error `Gantt diagrams cannot last more than 20 years (Assumed diagram type: gantt)` on the line `Project starts 2000-01-01`, as upstream does

#### Scenario: Invalid date

- **WHEN** a gantt diagram contains `Project starts 2020-02-30`
- **THEN** the system reports the error `You should send a mail to plantuml@gmail.com with this log (V1.2026.9) java.time.DateTimeException: Invalid date 'FEBRUARY 30' (Assumed diagram type: gantt)` on that line

#### Scenario: Labels on a column

- **WHEN** a gantt diagram contains `labels on first column and left aligned` and `[T] lasts 3 days`
- **THEN** the line is ignored, and the diagram shows upstream's warning `'labels on first column' is no longer supported: this line is ignored`

#### Scenario: Drawing failure

- **WHEN** a gantt diagram contains `Project starts 2020-01-06`, `group [G]`, `[A] on {Alice} lasts 2 days` and `end group`
- **THEN** rendering the diagram returns an error and no drawing

### Requirement: Gantt output formats

Gantt diagrams SHALL render to SVG and PNG. TXT and UTXT SHALL render upstream's table of tasks in ASCII, without the title and the other annotations: a header row with the task, start, end and duration, in the language of the diagram, and a row per task and milestone, without the groups and the separators. The start and the end SHALL be dates as upstream writes them, or `Day N` without `Project starts`, the end being the day before the end of the task, as upstream computes it, and the duration SHALL be the time from the start to the end of the task, closed days included.

#### Scenario: PNG gantt diagram

- **WHEN** a gantt diagram with `Project starts 2020-01-06`, `[A] lasts 3 days` and `then [B] lasts 2 days` is rendered as PNG
- **THEN** the PNG image shows the calendar, the two tasks and the arrow, as upstream's PNG does

#### Scenario: UTXT gantt table

- **WHEN** a gantt diagram with `Project starts 2020-01-06`, `[Build] lasts 3 days` and `then [Test] lasts 2 days` is rendered as UTXT
- **THEN** the output is the same ASCII table as for TXT, with the rows `Build`, `Jan 6`, `Jan 8`, `3 days` and `Test`, `Jan 9`, `Jan 10`, `2 days`

#### Scenario: Table rows

- **WHEN** a gantt diagram with `Project starts 2020-01-06`, `group [G]`, `[A] lasts 2 days`, `end group`, `-- Next --` and `[M] happens at [A]'s end` is rendered as TXT
- **THEN** the table has the rows `A` and `M` only

#### Scenario: Table in French

- **WHEN** a gantt diagram with `language fr`, `Project starts 2020-01-06` and `[A] lasts 9 days` is rendered as TXT
- **THEN** the header row is `Tâche`, `Début`, `Fin` and `Durée` and the row of `A` is `6 janv.`, `14 janv.` and `9 jours`

#### Scenario: Table without a project start

- **WHEN** a gantt diagram with `[A] lasts 2 days` and `[B] on {Bob:50%} lasts 2 days` is rendered as TXT
- **THEN** the row of `A` is `Day 1`, `Day 2` and `2 days` and the row of `B` is `Day 1`, `Day 4` and `4 days`

#### Scenario: Table with working hours

- **WHEN** a gantt diagram with `Project starts 2020-01-06`, `from 9:00 to 17:00 are working hours` and `[A] lasts 1 day` is rendered as TXT
- **THEN** the row of `A` is `Jan 6`, `Jan 5` and `17 hours`, as upstream writes it

### Requirement: Embedded gantt diagrams

A `{{gantt` ... `}}` block in creole text SHALL be drawn as the gantt diagram it contains, as upstream draws it.

#### Scenario: Gantt diagram in a note

- **WHEN** a sequence diagram has a note that contains `{{gantt`, `[A] lasts 2 days` and `}}`
- **THEN** the note contains the drawing of the gantt diagram, and the rest of the diagram renders normally
