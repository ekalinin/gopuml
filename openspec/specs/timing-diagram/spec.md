# timing-diagram Specification

## Purpose
Supports PlantUML v1.2026.9 timing diagrams under `@startuml`, with the same players, times, state changes, messages, constraints, highlights, notes, time axis, styles and errors as upstream.

## Requirements

### Requirement: Timing diagram blocks

The system SHALL render `@startuml` blocks that upstream reads as timing diagrams as upstream's timing diagrams. The diagram type SHALL be timing and the description `(Timing Diagram)`. A timing diagram SHALL be drawn as upstream draws it: the players one below the other in the order of their declarations, each with its title above its lane, and the time axis below the last player. The common commands of upstream's timing diagrams SHALL be supported: title, header, footer, caption, legend, mainframe, `skin`, `skinparam`, `<style>`, sprites, `scale`, pragmas, `hide footbox` and `show footbox`, which are read and change nothing, and the other commands that upstream adds to timing diagrams. Timing diagrams SHALL NOT need Graphviz. A timing diagram whose drawing fails in upstream SHALL fail to draw: rendering it SHALL return an error and no drawing, in every format.

#### Scenario: Timing diagram

- **WHEN** a `@startuml` block contains `robust "Web Browser" as WB`, `concise "Web User" as WU`, `@0`, `WU is Idle`, `WB is Idle`, `@100`, `WU is Waiting`, `WB is Processing`, `@300` and `WB is Waiting`
- **THEN** the lane of `Web Browser` is drawn above the lane of `Web User`, with the states `Idle`, `Processing` and `Waiting` of `Web Browser` and `Idle` and `Waiting` of `Web User`, the time axis below them has the labels `0`, `100` and `300`, the diagram type is timing and the description is `(Timing Diagram)`

#### Scenario: Scaled timing diagram

- **WHEN** a timing diagram contains `scale 2`, `concise "Web User" as WU`, `@0`, `WU is Idle`, `@100` and `WU is Waiting`
- **THEN** the drawing is twice as large as without `scale 2`, as upstream draws it

#### Scenario: Footbox commands

- **WHEN** a timing diagram contains `hide footbox`, `concise "Web User" as WU`, `@0`, `WU is Idle`, `@100` and `WU is Waiting`
- **THEN** the diagram is drawn as without `hide footbox`

#### Scenario: Timing diagram without players

- **WHEN** a `@startuml` block contains only `@0` and `@100`
- **THEN** the diagram is drawn with the time axis and its labels `0` and `100`, and without any lane

#### Scenario: Timing diagram without times

- **WHEN** a `@startuml` block contains only `concise "User" as U` and `U is Idle`
- **THEN** the lane of `User` is drawn with the state `Idle`, and the time axis has the label `0`, with the size of upstream's image

#### Scenario: Timing diagram without Graphviz

- **WHEN** a timing diagram is rendered and `dot` cannot be found
- **THEN** the diagram renders without an error

#### Scenario: Drawing failure

- **WHEN** a timing diagram contains `concise "User" as U`, `@-20`, `U is Idle`, `@-10` and `U is Busy`, every time being below zero, and is rendered as SVG and as PNG
- **THEN** rendering returns an error and no drawing, in both formats

### Requirement: Players

The system SHALL read the declarations of upstream's players: `robust`, `concise` and `rectangle` players (`robust "Web Browser" as WB`, or `robust WB` without a title), `clock` players (`clock "Clock" as C with period 50`), `binary` players (`binary "Enable" as EN`) and `analog` players (`analog "Vcore" as VDD`, whose quoted title is required), each optionally preceded by `compact` and followed by a stereotype (`<<name>>`). A `robust`, `concise` or `rectangle` declaration MAY end with a background color (`#lightblue`), which SHALL fill the whole lane of the player, its title included. An unknown color SHALL fail the line with the message `No such color (Assumed diagram type: timing)`. A robust player SHALL be drawn as a line that moves between the levels of its states, with the labels of the states on the left, a concise player as a ribbon with the label of each state inside it, and a rectangle player as a ribbon of rectangles, as upstream draws them.

#### Scenario: Robust player

- **WHEN** a timing diagram contains `robust "Web Browser" as WB`, `@0`, `WB is Idle`, `@100` and `WB is Busy`
- **THEN** the title `Web Browser` is drawn above the lane, and the line of the player moves from the level of `Idle` to the level of `Busy` at 100, with the labels `Idle` and `Busy` on the left

#### Scenario: Concise player

- **WHEN** a timing diagram contains `concise "Web User" as WU`, `@0`, `WU is Idle`, `@100` and `WU is Waiting`
- **THEN** a ribbon is drawn with `Idle` from 0 to 100 and `Waiting` after 100

#### Scenario: Rectangle player

- **WHEN** a timing diagram contains `rectangle "Bus" as B`, `@0`, `B is Idle`, `@50` and `B is Busy`
- **THEN** the title `Bus` is drawn above the lane, and the states `Idle` and `Busy` are drawn in rectangles, as upstream draws them

#### Scenario: Player without a title

- **WHEN** a timing diagram contains `robust WB`, `@0`, `WB is Idle`, `@100` and `WB is Busy`
- **THEN** the player is drawn without a title, with the states `Idle` and `Busy`

#### Scenario: Player background color

- **WHEN** a timing diagram contains `concise "Web User" as WU #lightblue`, `concise "Other" as O`, `@0`, `WU is Idle`, `O is A`, `@100` and `WU is Waiting`
- **THEN** a light blue (`#ADD8E6`) rectangle is drawn behind the lane of `Web User`, its title included, and not behind the lane of `Other`

#### Scenario: Unknown player color

- **WHEN** a timing diagram contains `concise "Web User" as WU #notacolor`
- **THEN** the diagram fails at that line with the message `No such color (Assumed diagram type: timing)`

### Requirement: Times

The system SHALL read upstream's times, after `@` in a time block (`@100`) or before `is` in a player block (`100 is Busy`): integers, which MAY be negative, decimals (`2.5`), times relative to the current time (`+50`), named times (`:start`, given by `@0 as :start`) with an optional offset (`:start+50`), multiples of the period of a clock (`clk*3`), hours (`0:15:00`) and dates (`2024/07/15`). A time block SHALL set the current time of the lines that follow it. A time that names an unknown named time SHALL fail the line with the message `What time? (Assumed diagram type: timing)`, and with upstream's crash message when it has an offset. The time axis SHALL label its ticks as upstream does: decimals without their fraction (`2.5` as `2`), hours as `h:mm:ss`, dates as `MM/dd`, and dates in the pattern of a previous `use date format "<pattern>"` line, formatted as upstream's `SimpleDateFormat` in the en_US locale and the default time zone. A pattern that upstream's `SimpleDateFormat` rejects SHALL fail the line with the message `Bad date format (Assumed diagram type: timing)`. Times below zero SHALL be placed on the axis, with a tick at 0, as upstream does.

#### Scenario: Relative times

- **WHEN** a timing diagram contains `concise "User" as U`, `@0`, `U is Idle`, `@+50`, `U is Busy`, `@+50` and `U is Done`
- **THEN** `Busy` starts at 50, `Done` at 100, and the time axis has the labels `0`, `50` and `100`

#### Scenario: Decimal times

- **WHEN** a timing diagram contains `concise "User" as U`, `@0`, `U is Idle`, `@2.5`, `U is Busy`, `@5` and `U is Done`
- **THEN** `Busy` starts at 2.5 and the time axis has the labels `0`, `2` and `5`, as upstream labels them

#### Scenario: Named times

- **WHEN** a timing diagram contains `concise "User" as U`, `@0 as :start`, `U is Idle`, `@100 as :end`, `U is Busy`, `@:start+50` and `U is Mid`
- **THEN** `Mid` starts at 50, between `Idle` and `Busy`, and the time axis has the labels `0`, `50` and `100`

#### Scenario: Unknown named time

- **WHEN** a timing diagram contains `concise "User" as U` and `@:start`
- **THEN** the diagram fails at the line `@:start` with the message `What time? (Assumed diagram type: timing)`

#### Scenario: Unknown named time with an offset

- **WHEN** a timing diagram contains `concise "User" as U` and `@:start+5`
- **THEN** the diagram fails at the line `@:start+5` with the message `You should send a mail to plantuml@gmail.com with this log (V1.2026.9) java.lang.NullPointerException: Cannot invoke "net.sourceforge.plantuml.timingdiagram.TimeTick.getTime()" because "result" is null (Assumed diagram type: timing)`

#### Scenario: Clock multiples

- **WHEN** a timing diagram contains `clock clk with period 10`, `concise "User" as U`, `@0`, `U is Idle`, `@clk*3` and `U is Busy`
- **THEN** `Busy` starts at 30, and the time axis has the labels `0`, `10` and `30`

#### Scenario: Hour times

- **WHEN** a timing diagram contains `concise "User" as U`, `@0:15:00`, `U is Idle`, `@1:30:00` and `U is Busy`
- **THEN** the time axis has the labels `0:15:00` and `1:30:00`

#### Scenario: Date times

- **WHEN** a timing diagram contains `concise "User" as U`, `@2024/07/15`, `U is Idle`, `@2024/07/18` and `U is Busy`
- **THEN** the time axis has the labels `07/15` and `07/18`

#### Scenario: Date format

- **WHEN** a timing diagram contains `use date format "MMM yyyy"`, `concise "User" as U`, `@2024/07/15`, `U is Idle`, `@2024/08/15` and `U is Busy`
- **THEN** the time axis has the labels `Jul 2024` and `Aug 2024`

#### Scenario: Bad date format

- **WHEN** a timing diagram contains `concise "User" as U` and `use date format "yyyy-qq"`
- **THEN** the diagram fails at the line `use date format "yyyy-qq"` with the message `Bad date format (Assumed diagram type: timing)`

#### Scenario: Times below zero

- **WHEN** a timing diagram contains `concise "User" as U`, `@-10`, `U is Idle`, `@10` and `U is Busy`
- **THEN** the time axis has the labels `-10`, `0` and `10`

### Requirement: State changes

The system SHALL change the state of a player with `<player> is <state>` at the current time, and with `<time> is <state>` in a player block that starts with `@<player>`, as upstream does. A state MAY be a name, a quoted label, `{hidden}` (nothing is drawn), `{...}` (a blank segment), `{-}` (a flat line) or two states (`{A,B}`), and MAY be followed by a color (`#pink`) and a comment (`: text`). A state given before the first time SHALL be the initial state of the player, drawn before its first change. When two changes of a robust, concise or rectangle player happen at the same time, the first one SHALL be kept, as upstream does. `<player> has A,B,C` SHALL declare the states of a robust player, which SHALL get a level each, even when no change uses them, and `<player> has "Long label" as L` SHALL give the state `L` the label `Long label`. An unknown player SHALL fail the line with the message `Unknown "<code>" (Assumed diagram type: timing)` for `is` and `Unknown <code> (Assumed diagram type: timing)` for `has`. `@<code>` with an unknown player SHALL fail with the message `No such participant <code> (Assumed diagram type: timing)`. A `<time> is <state>` line before any player block or robust, concise or rectangle declaration SHALL fail with the message `Missing @ line before this (Assumed diagram type: timing)`, because, as upstream does, the declarations of clock, binary and analog players do not start a player block. A state change of a clock player SHALL fail the line with upstream's crash message `You should send a mail to plantuml@gmail.com with this log (V1.2026.9) java.lang.UnsupportedOperationException (Assumed diagram type: timing)`.

#### Scenario: Player block

- **WHEN** a timing diagram contains `robust "Web Browser" as WB`, `concise "Web User" as WU`, `@WB`, `0 is idle`, `+200 is Proc.`, `+100 is Waiting`, `@WU`, `0 is Waiting` and `+500 is ok`
- **THEN** `Web Browser` is `idle` from 0, `Proc.` from 200 and `Waiting` from 300, `Web User` is `Waiting` from 0 and `ok` from 500, and the time axis has the labels `0`, `200`, `300` and `500`

#### Scenario: Initial state

- **WHEN** a timing diagram contains `concise "U" as U`, `U is Start`, `@10` and `U is A`
- **THEN** `Start` is drawn before 10 and `A` after it, and the time axis has the label `10` only

#### Scenario: Two changes at the same time

- **WHEN** a timing diagram contains `concise "U" as U`, `@10`, `U is A`, `U is B`, `@20` and `U is C`
- **THEN** `A` is drawn from 10 to 20 and `B` is not drawn

#### Scenario: Special states

- **WHEN** a timing diagram contains `concise "U" as U`, `@0`, `U is A`, `@10`, `U is {hidden}`, `@20`, `U is {...}`, `@30`, `U is {-}`, `@40` and `U is D`
- **THEN** `A` is drawn from 0 to 10, nothing from 10 to 20, a blank segment from 20 to 30, a flat line from 30 to 40 and `D` after 40, as upstream draws them

#### Scenario: State color and comment

- **WHEN** a timing diagram contains `concise "Web User" as WU`, `@0`, `WU is Idle #pink`, `@100` and `WU is Waiting : slow`
- **THEN** the segment of `Idle` is filled with pink (`#FFC0CB`) and the comment `slow` is drawn next to `Waiting`

#### Scenario: Declared states

- **WHEN** a timing diagram contains `robust "DNS" as DNS`, `DNS has Idle,Resolving,Done`, `concise "U" as U`, `U has "Long waiting state" as W`, `@0`, `DNS is Idle`, `U is W`, `@100` and `DNS is Done`
- **THEN** `DNS` has the levels of `Done`, `Resolving` and `Idle`, with their labels on the left, `Resolving` included, and the state of `U` is labeled `Long waiting state`

#### Scenario: Unknown player

- **WHEN** a timing diagram contains `concise "User" as U`, `@0` and `X is Idle`
- **THEN** the diagram fails at the line `X is Idle` with the message `Unknown "X" (Assumed diagram type: timing)`

#### Scenario: Unknown player in a state declaration

- **WHEN** a timing diagram contains `concise "User" as U` and `X has A,B`
- **THEN** the diagram fails at the line `X has A,B` with the message `Unknown X (Assumed diagram type: timing)`

#### Scenario: Unknown player block

- **WHEN** a timing diagram contains `concise "User" as U` and `@X`
- **THEN** the diagram fails at the line `@X` with the message `No such participant X (Assumed diagram type: timing)`

#### Scenario: Change without a player block

- **WHEN** a timing diagram contains `binary "B" as B` and `0 is high`
- **THEN** the diagram fails at the line `0 is high` with the message `Missing @ line before this (Assumed diagram type: timing)`

#### Scenario: State of a clock

- **WHEN** a timing diagram contains `clock clk with period 10` and `clk is high`
- **THEN** the diagram fails at the line `clk is high` with the message `You should send a mail to plantuml@gmail.com with this log (V1.2026.9) java.lang.UnsupportedOperationException (Assumed diagram type: timing)`

### Requirement: Clocks

The system SHALL draw a clock player as upstream draws it: a square wave that starts at time 0, low until the `offset`, then, in each `period`, high for the `pulse` (half the period without a pulse) and low for the rest of the period, up to the end of the time axis. The period of a clock SHALL be a time of the time axis. Decimal periods, pulses and offsets SHALL be supported.

#### Scenario: Clock

- **WHEN** a timing diagram contains `clock "Clock" as C with period 50`, `concise "User" as U`, `@0`, `U is Idle`, `@100` and `U is Busy`
- **THEN** the lane of `Clock` is a square wave that is high from 0 to 25 and from 50 to 75 and low from 25 to 50 and from 75 to 100, and the time axis has the labels `0`, `50` and `100`

#### Scenario: Clock pulse and offset

- **WHEN** a timing diagram contains `clock "Clock" as C with period 50 pulse 15 offset 10`, `concise "User" as U`, `@0`, `U is Idle`, `@100` and `U is Busy`
- **THEN** the lane of `Clock` is low from 0 to 10, high from 10 to 25, low from 25 to 60 and high from 60 to 75

#### Scenario: Decimal clock period

- **WHEN** a timing diagram contains `clock "MyClock" as Clk with period 12.5 offset 0.1`, `binary "State" as B`, `@0`, `B is low`, `@10.5` and `B is high`
- **THEN** the time axis has the labels `0`, `10` and `12`, as upstream labels them

### Requirement: Binary and analog players

The system SHALL draw a binary player as a line at its high level for the values `high` and `1`, in any case, and at its low level for every other value, as upstream does. The system SHALL draw an analog player as a line through its values, with the lowest and the highest value on the left (with zero among them) in Java's format of doubles (`5.0`). `between <min> and <max>` SHALL give the bounds of an analog player, and the format of its labels, which SHALL have the fraction digits of the bounds, as upstream deduces them. `<player> ticks num on multiple <n>` SHALL label the multiples of `n` between the bounds, and `<player> is <n> pixels height` SHALL set the height of the lane of the player. A value of an analog player that is not a number SHALL be read as 0, as upstream does.

#### Scenario: Binary player

- **WHEN** a timing diagram contains `binary "Enable" as EN`, `@0`, `EN is low`, `@10`, `EN is high`, `@20`, `EN is foo`, `@30` and `EN is 1`
- **THEN** the line of `Enable` is low from 0 to 10, high from 10 to 20, low from 20 to 30 and high after 30

#### Scenario: Analog player

- **WHEN** a timing diagram contains `analog "Vcore" as VDD`, `@0`, `VDD is 0`, `@100`, `VDD is 2.5`, `@200` and `VDD is 5`
- **THEN** a line is drawn through the three values, with the labels `0.0` and `5.0` on the left

#### Scenario: Analog bounds and ticks

- **WHEN** a timing diagram contains `analog "Vcore" between 0 and 5.25 as VDD`, `VDD ticks num on multiple 1`, `VDD is 200 pixels height`, `@0`, `VDD is 0`, `@100`, `VDD is 2.5`, `@200` and `VDD is 5`
- **THEN** the lane of `Vcore` is taller than without `VDD is 200 pixels height`, and has the labels `0`, `1`, `2`, `3`, `4` and `5` on the left

#### Scenario: Analog value that is not a number

- **WHEN** a timing diagram contains `analog "V" as V`, `@0`, `V is 3`, `@10`, `V is abc`, `@20` and `V is 4`
- **THEN** the line goes down to 0 at 10, and the labels on the left are `0.0` and `4.0`

### Requirement: Messages

The system SHALL draw a message `<player1> -> <player2> : <label>` as upstream does: an arrow from the lane of the first player to the lane of the second, at the current time, with its label. Each player MAY be followed by a time (`WU@0 -> WB@100`), and a relative time of either player SHALL be relative to the current time. A color MAY be given between the dashes and `>` (`-[#red]>`). A message without a time SHALL NOT be drawn when there is no current time yet, as upstream does. An unknown player SHALL fail the line with the message `No such element: <code> (Assumed diagram type: timing)`.

#### Scenario: Message between players

- **WHEN** a timing diagram contains `robust "Web Browser" as WB`, `concise "Web User" as WU`, `@0`, `WU is Idle`, `WB is Idle`, `@100`, `WU is Waiting`, `WB is Processing`, `@300`, `WB is Waiting` and `WU -> WB : URL`
- **THEN** an arrow labeled `URL` is drawn from `Web User` to `Web Browser` at 300

#### Scenario: Message with times

- **WHEN** a timing diagram has the players and states of "Message between players" and contains `WU@0 -> WB@100 : URL`
- **THEN** an arrow labeled `URL` is drawn from `Web User` at 0 to `Web Browser` at 100

#### Scenario: Relative message time

- **WHEN** a timing diagram contains `concise "A" as A`, `concise "B" as B`, `@0`, `A is x`, `B is y`, `@100`, `A is z` and `A@0 -> B@+50 : m`
- **THEN** the arrow labeled `m` ends at 150, as with `A@0 -> B@150 : m`

#### Scenario: Message color

- **WHEN** a timing diagram has the players and states of "Message between players" and contains `WU@0 -[#red]> WB@100 : URL`
- **THEN** the arrow labeled `URL` is red

#### Scenario: Message without a current time

- **WHEN** a timing diagram contains `clock "Sample Clock" as sample with period 10`, `binary "Threshold Reached" as threshold` and `threshold -> sample : notify`, without any time
- **THEN** the two players are drawn and no arrow is drawn

#### Scenario: Message to an unknown player

- **WHEN** a timing diagram has the players and states of "Message between players" and contains `WU -> X : URL`
- **THEN** the diagram fails at the line `WU -> X : URL` with the message `No such element: X (Assumed diagram type: timing)`

### Requirement: Constraints

The system SHALL draw a constraint `<player>@<time1> <-> @<time2> : <label>` as upstream does: a double arrow in the lane of the player, from the first time to the second, with its label above it. A relative second time SHALL be relative to the first time. Without a player, the constraint SHALL apply to the player of the current player block, and without one it SHALL fail the line with the message `You have to provide a participant (Assumed diagram type: timing)`. A color MAY be given in the arrow (`<-[#red]->`). An unknown named time SHALL fail the line with the message `Unknown time label (Assumed diagram type: timing)`.

#### Scenario: Constraint

- **WHEN** a timing diagram has the players and states of "Message between players" and contains `WB@0 <-> @50 : {50 ms lag}`
- **THEN** a double arrow from 0 to 50 is drawn in the lane of `Web Browser`, with the label `{50 ms lag}`

#### Scenario: Constraint relative to its start

- **WHEN** a timing diagram has the players and states of "Message between players" and contains `WB@100 <-> @+150 : {150 ms}`
- **THEN** a double arrow from 100 to 250 is drawn in the lane of `Web Browser`, with the label `{150 ms}`

#### Scenario: Constraint color

- **WHEN** a timing diagram has the players and states of "Message between players" and contains `WB@0 <-[#red]-> @50 : lag`
- **THEN** the double arrow labeled `lag` is red

#### Scenario: Constraint without a player

- **WHEN** a timing diagram contains `binary "B" as B` and `@0 <-> @50 : lag`
- **THEN** the diagram fails at the line `@0 <-> @50 : lag` with the message `You have to provide a participant (Assumed diagram type: timing)`

#### Scenario: Constraint with an unknown named time

- **WHEN** a timing diagram has the players and states of "Message between players" and contains `WB@:x <-> @50 : lag`
- **THEN** the diagram fails at the line `WB@:x <-> @50 : lag` with the message `Unknown time label (Assumed diagram type: timing)`

### Requirement: Highlights

The system SHALL draw `highlight <time1> to <time2> <color> : <caption>` as upstream does: an area over every lane between the two times, filled with the background color, between two lines of the line color (`#Gold;line:DimGrey`), with the caption at the top of the area.

#### Scenario: Highlight

- **WHEN** a timing diagram has the players and states of "Message between players" and contains `highlight 100 to 200 #Gold;line:DimGrey : Peak`
- **THEN** an area filled with gold (`#FFD700`) is drawn from 100 to 200 over both lanes, between two dim grey lines, with the caption `Peak`

### Requirement: Notes

The system SHALL draw `note top of <player> : <text>` and `note bottom of <player> : <text>` as upstream does: a note above or below the lane of the player, at the current time. A note MAY span several lines, from `note top of <player>` to `end note`. As upstream does, notes SHALL be drawn for concise, rectangle and binary players only, and SHALL NOT be drawn for the other players. An unknown player SHALL fail the line with the message `Unknown "<code>" (Assumed diagram type: timing)`.

#### Scenario: Notes on a player

- **WHEN** a timing diagram has the players and states of "Message between players" and contains `@100`, `note top of WU : first` and `note bottom of WU : second`
- **THEN** the note `first` is drawn above the lane of `Web User` and the note `second` below it, at 100

#### Scenario: Note on several lines

- **WHEN** a timing diagram has the players and states of "Message between players" and contains `@100`, `note bottom of WU`, `line one`, `line two` and `end note`
- **THEN** one note with the lines `line one` and `line two` is drawn below the lane of `Web User`

#### Scenario: Note on a robust player

- **WHEN** a timing diagram has the players and states of "Message between players" and contains `@100` and `note top of WB : dropped`
- **THEN** the diagram is drawn without the note, as upstream draws it

#### Scenario: Note on an unknown player

- **WHEN** a timing diagram has the players and states of "Message between players" and contains `note top of X : x`
- **THEN** the diagram fails at the line `note top of X : x` with the message `Unknown "X" (Assumed diagram type: timing)`

### Requirement: Time axis

The system SHALL draw the time axis as upstream does. `hide time-axis` SHALL leave out the time axis. `manual time-axis` SHALL put a tick at each time of the diagram, labeled with its name when a named time gave it. `scale <n> as <m> pixels` SHALL put a tick every `n` time units, `m` pixels apart, where `n` MAY be followed by the unit `s`, `m` (60), `h` (3600), `d` (86400) or `y`, and a value of zero SHALL fail the line with the message `Bad value (Assumed diagram type: timing)`. `mode compact` SHALL draw the players declared after it, and `compact` the player it precedes, without the line above the lane and the line under the title, and `mode compact` SHALL also leave out the vertical borders of the diagram.

#### Scenario: Hidden time axis

- **WHEN** a timing diagram contains `hide time-axis`, `concise "Web User" as WU`, `@0`, `WU is Idle`, `@100`, `WU is Waiting`, `@300` and `WU is Done`
- **THEN** the lane of `Web User` is drawn, and no time axis and no time label are drawn

#### Scenario: Manual time axis

- **WHEN** a timing diagram contains `manual time-axis`, `concise "Web User" as WU`, `@0 as :start`, `WU is Idle`, `@100 as :mid`, `WU is Waiting`, `@300` and `WU is Done`
- **THEN** the time axis has the labels `start`, `mid` and `300`

#### Scenario: Scale in pixels

- **WHEN** a timing diagram contains `scale 100 as 120 pixels`, `concise "Web User" as WU`, `@0`, `WU is Idle`, `@100`, `WU is Waiting`, `@300` and `WU is Done`
- **THEN** the time axis has the labels `0`, `100`, `200`, `300` and `400`, 120 pixels apart

#### Scenario: Scale with a unit

- **WHEN** a timing diagram contains `scale 1 m as 60 pixels`, `concise "U" as U`, `@0`, `U is A`, `@120` and `U is B`
- **THEN** the time axis has the labels `0`, `60`, `120` and `180`, 60 pixels apart

#### Scenario: Bad scale

- **WHEN** a timing diagram contains `concise "U" as U` and `scale 0 as 50 pixels`
- **THEN** the diagram fails at the line `scale 0 as 50 pixels` with the message `Bad value (Assumed diagram type: timing)`

#### Scenario: Compact mode

- **WHEN** a timing diagram contains `mode compact`, `robust "Web Browser" as WB`, `concise "Web User" as WU`, `@0`, `WU is Idle`, `WB is Idle`, `@100`, `WU is Waiting`, `@300` and `WU is Done`
- **THEN** the lanes are drawn without the lines above them and under their titles, and the diagram has no vertical borders

#### Scenario: Compact player

- **WHEN** a timing diagram contains `compact concise "Web User" as WU`, `concise "Other" as O`, `@0`, `WU is Idle`, `O is A`, `@100` and `WU is Waiting`
- **THEN** `Web User` is drawn without the line above its lane and the line under its title, and `Other` with them

### Requirement: Timing styles

The system SHALL style timing diagrams with the `timingDiagram` styles, as upstream does: `robust`, `concise`, `rectangle`, `clock`, `binary` and `analog` for the lanes of each type of player, `arrow` for the messages, `constraintArrow` for the constraints, `highlight` for the highlights, `note` for the notes, `timeline` for the time axis, `timegrid` for the vertical lines of the ticks, and `.name` for the players and notes with the stereotype `name`.

#### Scenario: Title and style

- **WHEN** a timing diagram contains a `<style>` block with `timingDiagram { concise { BackGroundColor #33b2e2 } arrow { LineColor #FF0000 } }`, `title Plan`, `concise "Web User" as WU`, `concise "Server" as S`, `@0`, `WU is Idle`, `S is Idle`, `@100`, `WU is Waiting`, `S is Busy` and `WU -> S : req`
- **THEN** the title `Plan` is drawn above the diagram, the ribbons of both players are filled with `#33b2e2` and the arrow labeled `req` is red

#### Scenario: Stereotype style

- **WHEN** a timing diagram contains a `<style>` block with `timingDiagram { .hot { BackGroundColor #33b2e2 } }`, `concise "Web User" as WU <<hot>>`, `concise "Server" as S`, `@0`, `WU is Idle`, `S is Idle`, `@100` and `WU is Waiting`
- **THEN** the ribbon of `Web User` is filled with `#33b2e2`, and the ribbon of `Server` keeps the default color

### Requirement: Timing output formats

Timing diagrams SHALL render to SVG and PNG. TXT and UTXT SHALL report that the format is not supported for timing diagrams and SHALL NOT produce output, as for the other diagram types without a text output of their own.

#### Scenario: PNG timing diagram

- **WHEN** a timing diagram with `concise "Web User" as WU`, `@0`, `WU is Idle`, `@100` and `WU is Waiting` is rendered as PNG
- **THEN** the PNG image shows the lane and the time axis, as upstream's PNG does

#### Scenario: TXT timing diagram

- **WHEN** a timing diagram with `concise "Web User" as WU`, `@0` and `WU is Idle` is rendered as TXT
- **THEN** the system reports that TXT output is not supported for timing diagrams and produces no output

### Requirement: Embedded timing diagrams

A `{{` ... `}}` block in creole text whose lines form a timing diagram SHALL be drawn as that timing diagram, as upstream draws it.

#### Scenario: Timing diagram in a note

- **WHEN** a sequence diagram has a note that contains `{{`, `concise "Web User" as WU`, `@0`, `WU is Idle`, `@100`, `WU is Waiting` and `}}`
- **THEN** the note contains the drawing of the timing diagram, and the rest of the diagram renders normally
