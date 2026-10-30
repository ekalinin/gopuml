# salt-diagram Specification

## Purpose
Supports PlantUML v1.2026.9 salt diagrams under `@startsalt`, with the same widgets, groups, tables, trees, tabs, menus, scroll areas, border layout, images, sprites, errors and colors as upstream.

## Requirements

### Requirement: Salt diagram blocks

The system SHALL render `@startsalt` ... `@endsalt` blocks as upstream's salt diagrams. The diagram type SHALL be salt and the description `(Salt)`. A salt diagram SHALL be drawn as upstream draws it: the widgets of its top-level group in black and grays, with upstream's fixed font and a margin of 5. The common commands of upstream's salt diagrams SHALL be supported: title, header, footer, caption, legend, mainframe, `skin`, `skinparam`, `<style>`, sprites, `scale`, pragmas, and the other commands that upstream adds to salt diagrams. As upstream does, a line that matches a common command SHALL be read as that command, even inside a group, and `skinparam` and `<style>` SHALL change the background and the annotations only, not the widgets. Salt diagrams SHALL NOT need Graphviz. A first line `salt` SHALL fail with the message `This is not needed anymore. (Assumed diagram type: salt)`. A salt diagram whose drawing fails in upstream SHALL fail to draw: rendering it SHALL return an error and no drawing, in every format.

#### Scenario: Login form

- **WHEN** a `@startsalt` block contains `{`, `Login    | "MyName   "`, `Password | "****     "`, `[Cancel] | [  OK   ]` and `}`
- **THEN** three rows of two columns are drawn without lines between the cells: `Login` and `Password` in the first column, their text fields in the second, and the buttons `Cancel` and `OK` in the last row, the diagram type is salt and the description is `(Salt)`

#### Scenario: Title and background

- **WHEN** a salt diagram contains `skinparam backgroundColor #EEEEFF`, `title Login form`, `legend`, `Mockup` and `endlegend`, and a group with `[OK]`
- **THEN** the image is filled with `#EEEEFF`, the title `Login form` is drawn above the group and the legend `Mockup` below it, and the button `OK` is filled with `#EEE` and outlined in black, as without the skin parameter

#### Scenario: Scaled salt diagram

- **WHEN** a salt diagram contains `scale 2` and a group with `[OK]`
- **THEN** the drawing is twice as large as without `scale 2`, as upstream draws it

#### Scenario: Salt diagram without Graphviz

- **WHEN** a salt diagram is rendered and `dot` cannot be found
- **THEN** the diagram renders without an error

#### Scenario: Needless salt keyword

- **WHEN** a salt diagram contains `salt`, `{`, `[OK]` and `}`
- **THEN** the diagram fails at the line `salt` with the message `This is not needed anymore. (Assumed diagram type: salt)`

#### Scenario: Row read as a command

- **WHEN** a salt diagram contains `{`, `Title | "text"`, `Name | "me"` and `}`
- **THEN** the title `| "text"` is drawn above the diagram, and the group has only the row of `Name`

#### Scenario: Empty salt diagram

- **WHEN** a block contains only `@startsalt` and `@endsalt`
- **THEN** the system renders upstream's welcome diagram with the message `Empty description (Assumed diagram type: salt)`

#### Scenario: Content outside a group

- **WHEN** a salt diagram contains only `hello`, and is rendered as SVG and as PNG
- **THEN** rendering returns an error and no drawing, in both formats

#### Scenario: Unclosed group

- **WHEN** a salt diagram contains `{` and `[OK]`, without `}`
- **THEN** rendering returns an error and no drawing

### Requirement: Widgets

The system SHALL read each cell of a group as upstream does, trying in this order: a menu (`{*`), a tree (`{` then `T`), tabs (`{/`), a separator, a text field (a cell that starts and ends with `"`), a checked checkbox (`[X]` then its label), an unchecked checkbox (`[]` or `[ ]` then its label), a button (a cell that starts with `[` and ends with `]`, but not with `[[`), a droplist (a cell that starts and ends with `^`), a checked radio button (`(X)` then its label), an unchecked radio button (`()` or `( )` then its label), a pixel image, a reused image (`<<name>>`), plain text for any other cell that does not start with `{` or `}`, and a nested group, a scroll area or a border layout otherwise. The labels SHALL be creole text, as upstream draws them: bold, italics, sprites (`<$name>`) and links (`[[url label]]`), with OpenIconic references (`<&name>`) drawn as the styling capability requires. `\n` in a cell SHALL be drawn as is. Buttons, text fields and droplists SHALL be at least 8 pixels wide per character of their text, spaces included, where an icon counts as two characters and a creole tag as none, as upstream counts them.

#### Scenario: Widgets

- **WHEN** a group contains, one per line, `Just plain text`, `[This is my button]`, `()  Unchecked radio`, `(X) Checked radio`, `[]  Unchecked box`, `[X] Checked box`, `"Enter text here   "` and `^This is a droplist^`
- **THEN** they are drawn one below the other: the text, a rounded button filled with `#EEE`, an empty and a filled radio button, an empty and a checked box, the text `Enter text here` with a line under it, and a box filled with `#EEE` with the text `This is a droplist` and a down arrow on its right

#### Scenario: Text field width

- **WHEN** a group contains `"Enter text here   "`
- **THEN** the line under the text field is 147 pixels long: 18 characters of 8 pixels, plus 3, as upstream draws it

#### Scenario: Creole in a cell

- **WHEN** a group contains `<b>Bold</b> | <i>Italic</i> | [[http://plantuml.com PlantUML]]`
- **THEN** `Bold` is drawn in bold, `Italic` in italics, and `PlantUML` as a link to `http://plantuml.com`

#### Scenario: Line break in a cell

- **WHEN** a group contains `first\nsecond | x`
- **THEN** the first cell is drawn on one line as `first\nsecond`, as upstream draws it

### Requirement: Groups and tables

The system SHALL read a group from `{` to `}`, as upstream does: each line SHALL be a row, `|` SHALL separate its cells, and a group MAY be a cell of another group. Each column SHALL be as wide as its widest cell and each row as tall as its tallest cell, as upstream lays them out. The opening of a group SHALL select its lines: `{` none, `{+` an outline, `{#` an outline and the lines between all cells, `{!` the vertical lines on both sides of each column, `{-` the horizontal lines above and below each row, and `{^"title"` an outline with the title drawn in its top border on a white background. A cell `*` SHALL extend the cell before it over its column. A cell `.` SHALL be empty. `\|`, `\{` and `\}` SHALL be drawn as `|`, `{` and `}` without splitting the cell or starting a group, as upstream does. The lines of a group SHALL be drawn in upstream's order.

#### Scenario: Nested group

- **WHEN** a salt diagram contains `{#`, `Name | "text"`, `Modifiers: | { (X) public | () default }` and `}`
- **THEN** the radio buttons `public` and `default` are drawn side by side in the second cell of the second row, without lines between them, and the table has lines around and between its cells

#### Scenario: All table lines

- **WHEN** a salt diagram contains `{#`, `a | b`, `c | d` and `}`
- **THEN** the four cells are drawn with lines around and between them

#### Scenario: Outline

- **WHEN** a salt diagram contains `{+`, `a | b`, `c | d` and `}`
- **THEN** the four cells are drawn with a line around them and no lines between them

#### Scenario: Vertical lines

- **WHEN** a salt diagram contains `{!`, `a | b`, `c | d` and `}`
- **THEN** vertical lines are drawn on the left of, between and on the right of the two columns, and no horizontal lines are drawn

#### Scenario: Horizontal lines

- **WHEN** a salt diagram contains `{-`, `a | b`, `c | d` and `}`
- **THEN** horizontal lines are drawn above, between and below the two rows, and no vertical lines are drawn

#### Scenario: Group box with a title

- **WHEN** a salt diagram contains `{^"My group"`, `Login | "me"` and `}`
- **THEN** the cells are drawn in an outline, and the title `My group` is drawn on a white background over the top line of the outline, near its left end

#### Scenario: Spanning cell

- **WHEN** a salt diagram contains `{#`, `a | b | c`, `long text here | * | x` and `}`
- **THEN** `long text here` spans the first two columns of the second row, without a line between them, and `x` is in the third column

#### Scenario: Empty cell

- **WHEN** a salt diagram contains `{#`, `a | . | b` and `}`
- **THEN** the row has three cells, the middle one empty

#### Scenario: Escaped pipe and braces

- **WHEN** a salt diagram contains `{#`, `Ratio | 3\|4`, `Braces | a\{b\}c` and `}`
- **THEN** the table has two columns, with the cells `3|4` and `a{b}c` in the second one

### Requirement: Separators

The system SHALL draw a cell that starts and ends with `..`, `==`, `~~` or `--` as a separator across its cell, in `#AAA`: dotted for `..`, double for `==`, 1.5 pixels thick for `~~`, and a single line for `--`, as upstream draws them.

#### Scenario: Separators

- **WHEN** a group contains, one per line, `a`, `..`, `==`, `~~`, `--` and `b`
- **THEN** between `a` and `b`, a dotted line, a double line, a thick line and a single line are drawn in `#AAA`, as wide as the group

### Requirement: Droplists

The system SHALL draw a droplist `^text^` as a box filled with `#EEE`, with its text and a down arrow on its right. A droplist with more items after its text, separated by `^` (`^head^^item 1^^item 2^` or `^head^ item 1^ item 2^`), SHALL be open: its items SHALL be drawn in a box filled with `#EEE` under it, after the other widgets, over the cells that follow and without moving them, as upstream draws it. The image SHALL be large enough to show an open list that goes past the cells, as upstream's image is.

#### Scenario: Open droplist

- **WHEN** a salt diagram contains `{`, `^Open droplist^^ item 1^^ item 2^`, `Next row here | more | stuff` and `}`
- **THEN** the list with `item 1` and `item 2` is drawn under the droplist, over the row `Next row here`, and the row is drawn at the same place as with a closed droplist

#### Scenario: Wide open droplist

- **WHEN** a salt diagram contains `{`, `^Small^^ this is a very very very very long item that is wider than the head^` and `}`
- **THEN** the image is wide enough to show the whole long item, as upstream's image is

### Requirement: Trees

The system SHALL draw `{T` ... `}` as upstream's tree: each line SHALL be an entry, at the level given by its leading `+` characters, indented by 10 pixels per level and linked to its parent by gray (`#888`) lines. `T` followed by `#`, `!`, `-`, `+` or `^` SHALL draw the lines of a tree table, in gray, as the same characters do for a group. The cells after the first one of an entry SHALL be widgets, drawn in columns as wide as their widest cell. A tree MAY be the top-level group of the diagram. As upstream does, the widgets of a tree SHALL be drawn in the first pass only, so an open droplist in a tree SHALL NOT show its list.

#### Scenario: Tree

- **WHEN** a salt diagram contains `{`, `{T`, `+ World`, `++ America`, `+++ Canada`, `++ Europe`, `}` and `}`
- **THEN** `World`, `America`, `Canada` and `Europe` are drawn one below the other, indented by their levels, and each entry is linked to its parent by gray lines

#### Scenario: Tree table

- **WHEN** a salt diagram contains `{T#`, `+ World | "Earth"`, `++ America | [Visit]` and `}`
- **THEN** the entries are drawn in the first column, the text field and the button in the second, and gray lines are drawn around and between the cells

#### Scenario: Top-level tree

- **WHEN** a salt diagram contains `{T`, `+ a`, `++ b` and `}`, without an outer group
- **THEN** the tree with `a` and its child `b` is drawn

#### Scenario: Open droplist in a tree

- **WHEN** a salt diagram contains `{T#`, `+ Name | ^Pick^^ a^^ b^` and `}`
- **THEN** the droplist `Pick` is drawn and its items `a` and `b` are not, as upstream draws it

### Requirement: Tabs

The system SHALL draw `{/` ... `}` as upstream's tabs: the cells of its first line side by side, each outlined on its top and sides, with a short line at the bottom between the tabs. When a line of the tabs ends before `}`, the tabs SHALL be drawn one below the other.

#### Scenario: Tabs

- **WHEN** a salt diagram contains `{`, `{/ <b>General | Fullscreen | Behavior }` and `}`
- **THEN** the tabs `General`, in bold, `Fullscreen` and `Behavior` are drawn side by side, each outlined on its top and sides

#### Scenario: Vertical tabs

- **WHEN** a salt diagram contains `{`, `{/ General`, `Fullscreen`, `Behavior }` and `}`
- **THEN** the tabs `General`, `Fullscreen` and `Behavior` are drawn one below the other

### Requirement: Menus

The system SHALL draw `{*` ... `}` as upstream's menu bar: the cells of its first line SHALL be the entries of the bar, drawn on `#DDD` across the cell. On each following line, the first cell SHALL name an entry and the other cells SHALL be its items, where `-` is a separator line. The items of an entry SHALL be drawn under it in a list filled with `#DDD`, after the other widgets, over the cells that follow. When several entries have items, their lists SHALL be drawn in the order of the entries. An entry with items SHALL be drawn on a `#BBB` background when the diagram also has an open droplist, as upstream draws it because it measures the drawing first. An unknown entry name SHALL make the drawing fail.

#### Scenario: Menu with items

- **WHEN** a salt diagram contains `{`, `{* File | Edit | Source | Refactor`, ` Refactor | New | Open File | - | Close | Close All }`, `.` and `}`
- **THEN** the bar with `File`, `Edit`, `Source` and `Refactor` is drawn on `#DDD`, and the list with `New`, `Open File`, a line, `Close` and `Close All` is drawn on `#DDD` under `Refactor`, over the row that follows, and `Refactor` has no background of its own

#### Scenario: Highlighted menu entry

- **WHEN** a salt diagram contains `{`, `{* File | Edit`, ` File | New | Open }`, `^Pick^^ a^^ b^` and `}`
- **THEN** the entry `File` is drawn on a `#BBB` background, as upstream draws it

#### Scenario: Several menus with items

- **WHEN** a salt diagram contains `{`, `{* File | Edit`, ` File | New`, ` Edit | Undo }` and `}`
- **THEN** the list of `File` is drawn before the list of `Edit`

#### Scenario: Unknown menu

- **WHEN** a salt diagram contains `{* File | Edit` and ` View | Zoom }`
- **THEN** rendering returns an error and no drawing

### Requirement: Scroll areas

The system SHALL draw `{S` ... `}` as an outlined group with a vertical scroll bar on its right and a horizontal scroll bar below it, `{SI` with the vertical bar only and `{S-` with the horizontal bar only. Each bar SHALL be 15 pixels wide, with an arrow at each end, and SHALL be drawn in both drawing passes, as upstream draws it.

#### Scenario: Scroll area

- **WHEN** a salt diagram contains `{S`, `Message`, `.`, `.` and `}`
- **THEN** the cells are drawn in an outline, with a vertical bar on its right and a horizontal bar below it, each with two arrows, and each bar is drawn twice, as upstream draws it

#### Scenario: Vertical scroll bar

- **WHEN** a salt diagram contains `{SI`, `Message` and `}`
- **THEN** the cells are drawn in an outline with a vertical bar on its right and no horizontal bar

#### Scenario: Horizontal scroll bar

- **WHEN** a salt diagram contains `{S-`, `Message` and `}`
- **THEN** the cells are drawn in an outline with a horizontal bar below it and no vertical bar

### Requirement: Border layout

The system SHALL read a group whose first cell is `N=`, `S=`, `E=` or `W=` as upstream's border layout: each of `N=`, `S=`, `E=`, `W=` and `C=` SHALL be followed by the widget or group of that position, drawn at the top, the bottom, the right, the left and the center.

#### Scenario: Border layout

- **WHEN** a salt diagram contains `{`, `N={ [Top] }`, `W={ [Left] }`, `C={ [Center] }`, `E={ [Right] }`, `S={ [Bottom] }` and `}`
- **THEN** `Top` is drawn at the top, `Bottom` at the bottom, and `Left`, `Center` and `Right` between them, from left to right

### Requirement: Pixel images

The system SHALL read the lines from `<<name` or `<<` to `>>` as a pixel image, as upstream does: each `X` SHALL be drawn as a black pixel, and the other characters SHALL be left blank. `<<name>>` SHALL draw the image of that name again. An unknown name SHALL make the drawing fail.

#### Scenario: Pixel image

- **WHEN** a group contains `<<smile`, `XX.XX`, `.....`, `X...X`, `.XXX.`, `>>` and `<<smile>>`
- **THEN** the image is drawn twice, one below the other, with a black pixel for each `X`

#### Scenario: Unknown image

- **WHEN** a salt diagram contains `{`, `<<nothere>>` and `}`
- **THEN** rendering returns an error and no drawing

### Requirement: Sprites in salt diagrams

The system SHALL draw the sprites that a salt diagram defines (`sprite $name ...`) in its widgets with `<$name>`. As upstream does, these sprites SHALL be kept for the widgets only, so the title and the other annotations SHALL NOT draw them.

#### Scenario: Sprite in a cell

- **WHEN** a salt diagram contains `sprite $dot [4x4/16] {`, `0FF0`, `FFFF`, `FFFF`, `0FF0`, `}`, `{`, `<$dot> Ready` and `}`
- **THEN** the sprite is drawn before the text `Ready`

#### Scenario: Sprite in the title

- **WHEN** a salt diagram defines the sprite `$dot` as above and contains `title <$dot> T` and a group with `x`
- **THEN** the title `T` is drawn without the sprite, as upstream draws it

### Requirement: Salt dark colors

With dark mode, the system SHALL draw salt diagrams in upstream's dark colors: white instead of black and black instead of white, and the dark grays that upstream pairs with each gray.

#### Scenario: Dark mode

- **WHEN** a salt diagram with a group with `[OK] | ^Pick^` and `..` is rendered as SVG with `--dark-mode`
- **THEN** the output equals upstream's with `-darkmode`: the button and the droplist are filled with `#222` and outlined in white, and the separator is drawn in `#666`

### Requirement: Salt output formats

Salt diagrams SHALL render to SVG and PNG. TXT and UTXT SHALL report that the format is not supported for salt diagrams and SHALL NOT produce output, as for the other diagram types without a text output of their own.

#### Scenario: PNG salt diagram

- **WHEN** a salt diagram with a group with `Login | "MyName   "` and `[Cancel] | [  OK   ]` is rendered as PNG
- **THEN** the PNG image shows the text, the text field and the buttons, as upstream's PNG does

#### Scenario: TXT salt diagram

- **WHEN** a salt diagram with a group with `[OK]` is rendered as TXT
- **THEN** the system reports that TXT output is not supported for salt diagrams and produces no output

### Requirement: Embedded salt diagrams

A `{{salt` ... `}}` block in creole text SHALL be drawn as that salt diagram, as upstream draws it. A `{{` ... `}}` block whose first line is `salt` SHALL be drawn as upstream's error image `This looks like a salt diagram. Please use @startsalt instead of @startuml.`.

#### Scenario: Salt diagram in a note

- **WHEN** a sequence diagram has a note that contains `{{salt`, `{`, `[OK]`, `}` and `}}`
- **THEN** the note contains the drawing of the salt diagram with the button `OK`, and the rest of the diagram renders normally

#### Scenario: Salt diagram in an activity

- **WHEN** an activity diagram contains `start`, `:{{salt`, `{`, `Login | "user"`, `[OK]`, `}`, `}};` and `stop`
- **THEN** the box of the activity contains the drawing of the salt diagram, with `Login`, its text field and the button `OK`

#### Scenario: Old embedded form

- **WHEN** a sequence diagram has a note that contains `{{`, `salt`, `{`, `[OK]`, `}` and `}}`
- **THEN** the note contains upstream's error image with the message `This looks like a salt diagram. Please use @startsalt instead of @startuml.`
